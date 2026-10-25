package fuzz

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// sweepLockHelperEnv names the lock file a helper process takes. Set only on the helper.
const sweepLockHelperEnv = "FEOV_SWEEP_LOCK_HELPER_PATH"

// TestSweepLockHelper is the second process of TestASecondSweepQueuesBehindTheFirst: it takes
// the lock at the path it is handed, says so on stdout, and holds it until stdin closes.
func TestSweepLockHelper(t *testing.T) {
	path := os.Getenv(sweepLockHelperEnv)
	if path == "" {
		t.Skip("helper process for TestASecondSweepQueuesBehindTheFirst")
	}
	f, err := acquireSweepLock(path, 20*time.Millisecond, os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fmt.Printf("ACQUIRED pid=%d\n", os.Getpid())
	_, _ = io.Copy(io.Discard, os.Stdin)
}

// lockProc is one helper process, its stdout and stderr read line by line.
type lockProc struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout chan string
	stderr chan string
}

func startLockHelper(t *testing.T, path string) *lockProc {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestSweepLockHelper$", "-test.count=1")
	// The helper must not take the REAL lock in TestMain: a local gated run holds it already, in
	// this very process, and the helper would queue behind its own parent.
	env := []string{sweepLockHelperEnv + "=" + path}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "FEOV_RELEASE_GATE=") && !strings.HasPrefix(kv, sweepLockHelperEnv+"=") {
			env = append(env, kv)
		}
	}
	cmd.Env = env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	errp, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	p := &lockProc{cmd: cmd, stdin: stdin, stdout: lines(out), stderr: lines(errp)}
	t.Cleanup(func() {
		_ = p.stdin.Close()
		_ = p.cmd.Wait()
	})
	return p
}

func lines(r io.Reader) chan string {
	ch := make(chan string, 64)
	go func() {
		s := bufio.NewScanner(r)
		for s.Scan() {
			ch <- s.Text()
		}
		close(ch)
	}()
	return ch
}

// await returns the first line on ch containing want, failing after the deadline.
func await(t *testing.T, ch chan string, want, what string) string {
	t.Helper()
	deadline := time.After(60 * time.Second)
	for {
		select {
		case l, ok := <-ch:
			if !ok {
				t.Fatalf("%s: stream closed before a line containing %q", what, want)
			}
			if strings.Contains(l, want) {
				return l
			}
		case <-deadline:
			t.Fatalf("%s: no line containing %q within 60s", what, want)
		}
	}
}

// TWO PROCESSES, ONE LOCK. The first takes it; the second waits, naming the first by pid, and
// does not take it until the first exits.
func TestASecondSweepQueuesBehindTheFirst(t *testing.T) {
	if !sweepLockSupported {
		t.Skip("the sweep lock is a flock: Linux and macOS only")
	}
	path := filepath.Join(t.TempDir(), "feov", "release-fuzz.lock")

	first := startLockHelper(t, path)
	await(t, first.stdout, "ACQUIRED", "first")
	firstPid := fmt.Sprintf("pid=%d ", first.cmd.Process.Pid)

	second := startLockHelper(t, path)
	waiting := await(t, second.stderr, "release fuzz: waiting for "+path, "second")
	for _, want := range []string{"held by " + firstPid, "started=", "worktree="} {
		if !strings.Contains(waiting, want) {
			t.Errorf("the waiting line does not name the holder (%q): %s", want, waiting)
		}
	}
	select {
	case l := <-second.stdout:
		t.Fatalf("the second process ran while the first held the lock: %q", l)
	case <-time.After(300 * time.Millisecond):
	}

	_ = first.stdin.Close()
	got := await(t, second.stdout, "ACQUIRED", "second")
	secondPid := fmt.Sprintf("pid=%d", second.cmd.Process.Pid)
	if !strings.Contains(got, secondPid) {
		t.Errorf("second reported %q, want %s", got, secondPid)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), secondPid+" ") {
		t.Errorf("the lock file names %q, want the second holder %s", b, secondPid)
	}
}

// NO LOCK WHERE NONE APPLIES: a run with the gate unset is not a sweep, and CI is one job's
// machine.
func TestTheSweepLockIsLocalAndGatedOnly(t *testing.T) {
	for _, c := range []struct{ gate, ci string }{{"", ""}, {"", "true"}, {"1", "true"}} {
		t.Setenv("FEOV_RELEASE_GATE", c.gate)
		t.Setenv("CI", c.ci)
		f, err := holdSweepLockIfLocal()
		if err != nil || f != nil {
			t.Errorf("gate=%q CI=%q: took a lock (%v, %v), want none", c.gate, c.ci, f, err)
		}
	}
}

// THE SWEEP ITSELF QUEUES: a gated, local run of this test binary waits in TestMain while
// another process holds the lock, and takes it once that process lets go. XDG_CACHE_HOME moves
// the user cache, which os.UserCacheDir reads on Linux only.
func TestAGatedLocalRunQueuesInTestMain(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("moves the user cache through XDG_CACHE_HOME, which only Linux reads")
	}
	cache := t.TempDir()
	path := filepath.Join(cache, "feov", "release-fuzz.lock")
	held, err := acquireSweepLock(path, time.Second, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()

	cmd := exec.Command(os.Args[0], "-test.run=^$")
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "CI=") && !strings.HasPrefix(kv, "XDG_CACHE_HOME=") && !strings.HasPrefix(kv, "FEOV_RELEASE_GATE=") {
			env = append(env, kv)
		}
	}
	cmd.Env = append(env, "FEOV_RELEASE_GATE=1", "XDG_CACHE_HOME="+cache)
	errp, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	stderr := lines(errp)
	waiting := await(t, stderr, "release fuzz: waiting for "+path, "the gated run")
	if want := fmt.Sprintf("held by pid=%d ", os.Getpid()); !strings.Contains(waiting, want) {
		t.Errorf("the waiting line does not name this process (%q): %s", want, waiting)
	}
	_ = held.Close()
	await(t, stderr, "release fuzz: took "+path, "the gated run")
	for range stderr {
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("the gated run failed after taking the lock: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("pid=%d ", cmd.Process.Pid); !strings.HasPrefix(string(b), want) {
		t.Errorf("the lock file names %q, want the gated run %s", b, want)
	}
}
