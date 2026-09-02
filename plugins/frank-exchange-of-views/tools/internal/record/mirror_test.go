package record

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// THE KEY IS FROZEN. These twelve characters name a directory that already exists on
// developers' disks, written by every `merge verdict` since the mirror shipped. Any change to
// how it is derived — a filepath.Clean, a case fold, a different hash, os.UserCacheDir instead
// of os.UserHomeDir — renames every mirror, which does not fail: the writer creates a fresh one
// under the new name and the purge reports 0 against the old, both of them silent.
//
// The literals are computed independently of this package, so the test cannot agree with a
// broken implementation by sharing its mistake.
func TestTheMirrorKeyIsFrozen(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	for _, c := range []struct{ runDir, key string }{
		{"/home/dev/research/2026-08-22_is-7-prime", "466a16c6d549"},
		// NOT case-folded — recordroot.go's sibling key lowercases on Windows and this one
		// must not, or the two names diverge from what is already on disk.
		{"/Home/Dev/Research/Run", "5b8311c87439"},
		// NOT cleaned — a trailing separator is a different key, because it was always a
		// different key. The writer passes run.Dir() through verbatim.
		{"/home/dev/research/run/", "a573112119e6"},
	} {
		got, err := MirrorDir(c.runDir)
		if err != nil {
			t.Fatal(err)
		}
		if base := filepath.Base(got); base != c.key {
			t.Errorf("MirrorDir(%q) keyed %q, want %q — every mirror already on disk is now "+
				"unreachable to both the writer and the purge", c.runDir, base, c.key)
		}
	}
}

// The root is under the user cache and NOT under TMPDIR: a temp purge must not void the sole
// recovery path for a record that is not committed yet.
func TestTheMirrorRootIsTheUserCache(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	root, err := MirrorRoot()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".cache", "feov", "run-mirror"); root != want {
		t.Errorf("MirrorRoot() = %q, want %q", root, want)
	}
	dir, err := MirrorDir("/any/run")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(dir) != root {
		t.Errorf("MirrorDir landed in %q, outside the root %q", filepath.Dir(dir), root)
	}
}

// AGE IS MODIFICATION TIME, which is the property that makes this safe to run while other runs
// are live. A mirror is rewritten every round, so an active run's is fresh however old the run
// is; what ages out is the orphan of a run that stopped.
func TestPurgeStaleMirrorsTakesOrphansAndLeavesLiveRuns(t *testing.T) {
	root := t.TempDir()
	for _, n := range []string{"orphan", "live"} {
		if err := os.Mkdir(filepath.Join(root, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(filepath.Join(root, "orphan"), old, old); err != nil {
		t.Fatal(err)
	}

	if n := PurgeStaleMirrors(root, time.Now(), 30); n != 1 {
		t.Fatalf("purged %d, want 1", n)
	}
	if _, err := os.Stat(filepath.Join(root, "orphan")); !os.IsNotExist(err) {
		t.Error("the orphan survived the purge")
	}
	if _, err := os.Stat(filepath.Join(root, "live")); err != nil {
		t.Errorf("the live run's mirror was purged: %v", err)
	}
}

// A MISSING ROOT RETURNS THE SAME ZERO AS A CLEAN ONE, so the caller has to say which it got.
// This pins the behaviour the callers' reporting is written against rather than leaving it to
// be rediscovered.
func TestPurgeStaleMirrorsReturnsZeroWhenThereIsNoRoot(t *testing.T) {
	if n := PurgeStaleMirrors(filepath.Join(t.TempDir(), "never-created"), time.Now(), 30); n != 0 {
		t.Errorf("purged %d against a root that does not exist, want 0", n)
	}
}
