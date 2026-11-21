package checkpoint

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// THE BUG THIS FILE EXISTS FOR, and it was measured before it was fixed.
//
// Six FileChanged events fired concurrently against a six-check note recorded
// FOUR records. Two were lost. Every hook process read the state, added its own
// key and wrote the whole map back, so whichever wrote last erased everything
// that landed after its own read. That is #165's "coverage stops advancing":
// records were never stopping, they were being overwritten by their siblings.
func TestConcurrentUpdatesDoNotLoseRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")

	const n = 24
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "check-" + strconv.Itoa(i)
			if err := UpdateRearm(path, func(s *RearmState) {
				s.Rearmed[key] = Rearm{Index: i, Check: key, By: "x.go", Event: "change", At: "t"}
			}); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("UpdateRearm: %v", err)
	}

	s, err := LoadRearm(os.ReadFile, path)
	if err != nil {
		t.Fatalf("state is unreadable after concurrent writes — it tore: %v", err)
	}
	if len(s.Rearmed) != n {
		t.Fatalf("kept %d of %d records; concurrent updates are erasing each other", len(s.Rearmed), n)
	}
}

// The other half of the same failure: a short write landing over a longer file
// leaves the previous document's tail behind. A live rearmed.json in this repo
// was found in exactly that state — a complete 287-byte document followed by the
// tail of a longer one, with five records stamped the same second.
func TestConcurrentUpdatesNeverLeaveATornFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")

	// Seed a large state, then race many small ones against it. Shrinking writes
	// are what tear.
	if err := UpdateRearm(path, func(s *RearmState) {
		for i := 0; i < 40; i++ {
			k := "seed-" + strconv.Itoa(i)
			s.Rearmed[k] = Rearm{Index: i, Check: k, By: "a/very/long/path/to/make/this/document/large.go", Event: "change", At: "t"}
		}
	}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = UpdateRearm(path, func(s *RearmState) {
				s.Rearmed = map[string]Rearm{"small": {Index: 1, Check: "small", At: "t"}}
			})
		}(i)
	}
	wg.Wait()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var probe map[string]any
	if err := json.Unmarshal(b, &probe); err != nil {
		t.Fatalf("file is not valid JSON after racing writes — it tore:\n%s", b)
	}
}

// holderEnv, set, turns this test binary into a lock holder: it takes the re-arm lock on the path
// the variable names, says "held" on stdout, and waits to be killed.
const holderEnv = "CHECKPOINT_TEST_REARM_HOLDER"

func TestMain(m *testing.M) {
	if path := os.Getenv(holderEnv); path != "" {
		if _, err := lockRearm(path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Println("held")
		time.Sleep(time.Hour)
		os.Exit(3)
	}
	os.Exit(m.Run())
}

// A HOOK KILLED WHILE HOLDING THE LOCK DOES NOT BLOCK THE NEXT ONE. The holder is another process,
// as a hook is: while it lives, an update is refused within the timeout and writes nothing; once
// it is killed, the operating system releases its lock and the next update takes it at once.
// Losing one re-arm is cheap; a lock that outlives its holder would lose every one after it.
func TestAKilledHolderDoesNotBlockTheNextUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")
	holder := exec.Command(os.Args[0], "-test.run=^$")
	holder.Env = append(os.Environ(), holderEnv+"="+path)
	holder.Stderr = os.Stderr
	out, err := holder.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := holder.Start(); err != nil {
		t.Fatal(err)
	}
	killed := false
	kill := func() {
		if !killed {
			killed = true
			_ = holder.Process.Kill()
			_ = holder.Wait()
		}
	}
	t.Cleanup(kill)
	ready := make(chan string, 1)
	go func() {
		line, _ := bufio.NewReader(out).ReadString('\n')
		ready <- strings.TrimSpace(line)
	}()
	select {
	case line := <-ready:
		if line != "held" {
			t.Fatalf("the holder process did not take the lock: said %q", line)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the holder process never said it held the lock")
	}

	add := func(key string) error {
		return UpdateRearm(path, func(s *RearmState) {
			s.Rearmed[key] = Rearm{Index: 1, Check: key, At: "t"}
		})
	}
	err = add("while-held")
	if err == nil {
		t.Fatal("an update went through while another process held the lock")
	}
	for _, want := range []string{"is not recorded", "nothing is damaged"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must carry %q: %v", want, err)
		}
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a refused update wrote the state: %v", err)
	}

	kill()
	start := time.Now()
	if err := add("after-kill"); err != nil {
		t.Fatalf("a killed holder's lock blocked the next update: %v", err)
	}
	if waited := time.Since(start); waited >= lockRearmTimeout {
		t.Errorf("the next update waited %s for a dead holder", waited)
	}
	s, err := LoadRearm(os.ReadFile, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Rearmed["after-kill"]; !ok {
		t.Errorf("the update after the kill was not recorded: %v", keysOf(s))
	}
}

// THE STATE'S DIRECTORY IS CREATED BEFORE ITS LOCK. The lock file lives beside the state, so a
// project with no .claude/checkpoints yet records its first event rather than failing to lock.
func TestAnUpdateCreatesTheStateDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude", "checkpoints", "rearmed.json")
	if err := UpdateRearm(path, func(s *RearmState) {
		s.Rearmed["k"] = Rearm{Index: 1, Check: "k", At: "t"}
	}); err != nil {
		t.Fatalf("the first update in a project with no state directory: %v", err)
	}
}

// A LOCK THAT CANNOT BE TAKEN, OR A DIRECTORY THAT CANNOT BE MADE, IS A REFUSAL THAT SAYS SO: the
// event is not recorded and nothing is damaged. A regular file where the state's directory belongs
// fails both, on any platform.
func TestALockThatCannotBeTakenIsARefusal(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(blocker, "rearmed.json")
	_, lerr := lockRearm(path)
	uerr := UpdateRearm(path, func(s *RearmState) {})
	for _, c := range []struct {
		err  error
		says string
	}{{lerr, "could not take its lock"}, {uerr, "cannot create its directory"}} {
		if c.err == nil {
			t.Errorf("no refusal where %q was expected", c.says)
			continue
		}
		for _, want := range []string{c.says, "is not recorded", "nothing is damaged"} {
			if !strings.Contains(c.err.Error(), want) {
				t.Errorf("the refusal must carry %q: %v", want, c.err)
			}
		}
	}
}

// A RELEASE THAT FAILS IS REPORTED (#1553), never dropped: after a commit it is a LockLeftError,
// which the hook reads to keep reporting the commit, and after a refusal it joins the refusal.
func TestAFailedReleaseIsReported(t *testing.T) {
	refused := errors.New("refused")
	if err := withRelease("rearmed.json", nil, nil); err != nil {
		t.Errorf("a clean commit and release reported %v", err)
	}
	if err := withRelease("rearmed.json", refused, nil); err != refused {
		t.Errorf("a refusal with a clean release must be the refusal itself, got %v", err)
	}
	var left *LockLeftError
	if err := withRelease("rearmed.json", nil, os.ErrPermission); !errors.As(err, &left) {
		t.Errorf("a release that failed after a commit must be a LockLeftError, got %v", err)
	}
	err := withRelease("rearmed.json", refused, os.ErrPermission)
	if !errors.Is(err, refused) || !errors.Is(err, os.ErrPermission) || errors.As(err, &left) {
		t.Errorf("a release that failed after a refusal must join both and claim no commit, got %v", err)
	}
}

// A COMMITTED UPDATE WHOSE RELEASE FAILED SAYS BOTH HALVES: the event is recorded, and the lock is
// held until the hook exits. The hook reads the type to keep reporting the commit.
func TestALockLeftErrorSaysTheEventIsRecorded(t *testing.T) {
	var err error = &LockLeftError{Path: "rearmed.json", Err: os.ErrPermission}
	var left *LockLeftError
	if !errors.As(err, &left) || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("LockLeftError must be matchable and unwrap its cause: %v", err)
	}
	for _, want := range []string{"IS recorded", "could not be released", "when this hook exits"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must carry %q: %v", want, err)
		}
	}
}
