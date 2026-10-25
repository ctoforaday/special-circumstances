package fuzz

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A SECOND SWEEP REFUSES, NAMING THE FIRST. This process holds the lock; a gated, local run of
// this test binary then exits 1 in TestMain, before any test runs, naming this process's pid,
// start time and worktree. The user cache moves through the variable os.UserCacheDir reads on
// each OS: XDG_CACHE_HOME on Linux, HOME on macOS, LocalAppData on Windows.
func TestASecondSweepRefusesNamingTheFirst(t *testing.T) {
	root := t.TempDir()
	cache := root
	if runtime.GOOS == "darwin" {
		cache = filepath.Join(root, "Library", "Caches")
	}
	path := filepath.Join(cache, "feov", "release-fuzz.lock")
	held, err := takeSweepLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })

	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(envWithout("GITHUB_ACTIONS", "FEOV_RELEASE_GATE", "XDG_CACHE_HOME", "HOME", "LocalAppData", "LOCALAPPDATA"),
		"FEOV_RELEASE_GATE=1", "XDG_CACHE_HOME="+root, "HOME="+root, "LocalAppData="+root)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("the second sweep did not refuse while this process held the lock (%v); stderr:\n%s", err, stderr.String())
	}
	got := stderr.String()
	for _, want := range []string{
		"another sweep holds " + path,
		fmt.Sprintf("pid=%d ", os.Getpid()),
		"started=",
		"worktree=",
		"One release fuzz invocation per machine at a time",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, got)
		}
	}
}

// envWithout is this process's environment less every variable named (case-insensitively, as
// Windows reads them).
func envWithout(names ...string) []string {
	var env []string
	for _, kv := range os.Environ() {
		keep := true
		for _, n := range names {
			if len(kv) > len(n) && kv[len(n)] == '=' && strings.EqualFold(kv[:len(n)], n) {
				keep = false
			}
		}
		if keep {
			env = append(env, kv)
		}
	}
	return env
}

// NO LOCK WHERE NONE APPLIES: a run with the gate unset is not a sweep, and a GitHub Actions
// runner is one job's machine.
func TestTheSweepLockIsLocalAndGatedOnly(t *testing.T) {
	for _, c := range []struct{ gate, actions string }{{"", ""}, {"", "true"}, {"1", "true"}} {
		t.Setenv("FEOV_RELEASE_GATE", c.gate)
		t.Setenv("GITHUB_ACTIONS", c.actions)
		f, err := holdSweepLockIfLocal()
		if err != nil || f != nil {
			t.Errorf("gate=%q GITHUB_ACTIONS=%q: took a lock (%v, %v), want none", c.gate, c.actions, f, err)
		}
	}
}
