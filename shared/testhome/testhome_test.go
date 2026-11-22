package testhome

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// lockedTree plants what Go's module cache looks like — a directory without write permission
// holding a read-only file — under root, and returns the directory.
func lockedTree(t *testing.T, root string) string {
	t.Helper()
	locked := filepath.Join(root, "go", "pkg", "mod", "example.com", "m@v1.0.0")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "go.mod"), []byte("module m\n"), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { // if the code under test fails, do not leak the fixture as well
		_ = os.Chmod(locked, 0o755)
		_ = os.RemoveAll(root)
	})
	return locked
}

// THE HOME IS GONE AFTER ITS UNDO, even when a test left a read-only tree in it.
func TestTheUndoRemovesAReadOnlyTree(t *testing.T) {
	for _, k := range append([]string{"GOMODCACHE", "GOCACHE"}, Keys...) {
		t.Setenv(k, os.Getenv(k))
	}
	dir, undo, err := New("testhome-test-")
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range Keys {
		if got := os.Getenv(k); got != dir {
			t.Errorf("%s = %q, want the temporary home %q", k, got, dir)
		}
	}
	lockedTree(t, dir)
	if err := undo(); err != nil {
		t.Fatalf("the undo reported: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("%s is still there after its undo (stat: %v)", dir, err)
	}
}

// A `go` STARTED UNDER THE HOME RESOLVES ITS CACHES OUTSIDE IT, so the module cache never lands
// in the sandbox in the first place.
func TestAGoChildKeepsItsCachesOutOfTheHome(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go on PATH")
	}
	for _, k := range append([]string{"GOMODCACHE", "GOCACHE"}, Keys...) {
		t.Setenv(k, os.Getenv(k))
	}
	_ = os.Unsetenv("GOMODCACHE")
	_ = os.Unsetenv("GOCACHE")
	dir, undo, err := New("testhome-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := undo(); err != nil {
			t.Error(err)
		}
	})
	out, err := exec.Command("go", "env", "GOMODCACHE", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if rel, err := filepath.Rel(dir, p); err == nil && filepath.IsLocal(rel) {
			t.Errorf("a go child resolves %s, inside the temporary home %s", p, dir)
		}
	}
}

// A cache the caller set is theirs: New leaves it, and the undo does not unset it.
func TestACacheTheCallerSetIsLeftAlone(t *testing.T) {
	own := t.TempDir()
	t.Setenv("GOMODCACHE", own)
	for _, k := range Keys {
		t.Setenv(k, os.Getenv(k))
	}
	_, undo, err := New("testhome-test-")
	if err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("GOMODCACHE"); got != own {
		t.Errorf("GOMODCACHE = %q under the home, want the caller's %q", got, own)
	}
	if err := undo(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("GOMODCACHE"); got != own {
		t.Errorf("GOMODCACHE = %q after the undo, want the caller's %q", got, own)
	}
}

func TestRemoveTreeReportsNothingForAMissingDirectory(t *testing.T) {
	if err := RemoveTree(filepath.Join(t.TempDir(), "absent")); err != nil {
		t.Fatal(err)
	}
}

// NOTHING A `go` STARTED UNDER THE HOME LEAVES BEHIND COMES BACK AFTER THE REMOVAL. A go command
// with no telemetry mode in its config directory starts a detached child that writes there after
// the parent exits, and re-created homes removed right after it.
func TestAGoChildLeavesNothingThatOutlivesTheRemoval(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go on PATH")
	}
	for _, k := range append([]string{"GOMODCACHE", "GOCACHE", "XDG_CONFIG_HOME"}, Keys...) {
		t.Setenv(k, os.Getenv(k))
	}
	_ = os.Unsetenv("XDG_CONFIG_HOME")
	var dirs []string
	for range 3 {
		dir, undo, err := New("testhome-test-")
		if err != nil {
			t.Fatal(err)
		}
		dirs = append(dirs, dir)
		if out, err := exec.Command("go", "env", "GOMODCACHE").CombinedOutput(); err != nil {
			t.Fatalf("go env: %v\n%s", err, out)
		}
		if err := undo(); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(3 * time.Second)
	for _, d := range dirs {
		if _, err := os.Stat(d); !os.IsNotExist(err) {
			t.Errorf("%s is back after its removal (stat: %v): something the go command started outlived it", d, err)
			_ = RemoveTree(d)
		}
	}
}

// fakeM stands in for *testing.M: its Run is the suite.
type fakeM func() int

func (f fakeM) Run() int { return f() }

// RUN REMOVES THE HOME AND FAILS THE PACKAGE ON WHAT IT COULD NOT DO. A green suite with a
// failed check after it exits non-zero: a warning on a passing package is printed to nobody.
func TestRunRemovesTheHomeAndFailsOnACheck(t *testing.T) {
	for _, k := range append([]string{"GOMODCACHE", "GOCACHE"}, Keys...) {
		t.Setenv(k, os.Getenv(k))
	}
	var home string
	suite := fakeM(func() int {
		home = os.Getenv("HOME")
		lockedTree(t, home)
		return 0
	})
	if code := Run(suite, "testhome-test-"); code != 0 {
		t.Fatalf("Run = %d over a passing suite", code)
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("the home %s survived Run (stat: %v)", home, err)
	}
	failing := func() error { return os.ErrInvalid }
	if code := Run(fakeM(func() int { return 0 }), "testhome-test-", failing); code == 0 {
		t.Error("Run = 0 with a post-suite check that failed")
	}
	if code := Run(fakeM(func() int { return 3 }), "testhome-test-"); code != 3 {
		t.Errorf("Run = %d, want the suite's own failing code 3", code)
	}
}
