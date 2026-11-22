package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestChooseTmp(t *testing.T) {
	plenty := uint64(memFloor) * 2
	for _, c := range []struct {
		name      string
		callerSet string
		exists    bool
		free      uint64
		wantDir   string
		wantWhy   string
	}{
		{"memory when there is room", "", true, plenty, memRoot, "memory"},
		{"the caller's choice stands", "/somewhere/else", true, plenty, "", "as set by the caller"},
		{"no tmpfs on this system", "", false, plenty, "", "no " + memRoot},
		{"too little free for a shared box", "", true, memFloor - 1, "", "stay on disk"},
		{"free space unreadable", "", true, 0, "", "stay on disk"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := chooseTmp(c.callerSet, c.exists, c.free)
			if got.dir != c.wantDir || !strings.Contains(got.reason, c.wantWhy) {
				t.Errorf("chooseTmp = %+v, want dir %q and a reason saying %q", got, c.wantDir, c.wantWhy)
			}
		})
	}
}

// Only this tool's own leftovers, and only old ones: a young directory may be another session's
// check, still running.
func TestReapStaleTakesOnlyOldDirectoriesOfItsOwn(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	mk := func(name string, age time.Duration) string {
		p := filepath.Join(root, name)
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, now.Add(-age), now.Add(-age)); err != nil {
			t.Fatal(err)
		}
		return p
	}
	old := mk(memPrefix+"old", memStale+time.Minute)
	young := mk(memPrefix+"young", memStale-time.Minute)
	foreign := mk("someone-elses", memStale*10)
	if errs := reapStale(root, now); len(errs) != 0 {
		t.Fatalf("reapStale reported %v", errs)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("a stale directory of this tool's was left in RAM")
	}
	for _, p := range []string{young, foreign} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s was removed: %v", filepath.Base(p), err)
		}
	}
}

// THE REAP RUNS BEFORE THE FLOOR IS MEASURED, AND TAKES A READ-ONLY TREE. Reaping only above the
// floor never reaped a /dev/shm the leftovers held just under it (3057 MB free against 3072), and a
// gate's leftover holds Go's module cache, whose directories a plain RemoveAll cannot empty.
func TestTheReapRunsUnderTheFloorAndTakesReadOnlyTrees(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	stale := filepath.Join(root, memPrefix+"killed")
	locked := filepath.Join(stale, "go", "pkg", "mod", "m@v1.0.0")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "go.mod"), nil, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	old := now.Add(-memStale - time.Hour)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}

	c, errs := reapThenChoose(root, "", now, func(string) uint64 { return memFloor - 1 })
	if len(errs) != 0 {
		t.Fatalf("the reap reported %v", errs)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("a stale directory holding a read-only tree survived a reap under the floor (stat: %v)", err)
	}
	if c.dir != "" {
		t.Errorf("chose %q with the root under the floor, want the default temp directory", c.dir)
	}
}
