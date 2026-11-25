package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestChooseTmp(t *testing.T) {
	plenty := uint64(memFloor) * 2
	roomy := headroom{bytes: plenty, binding: "MemAvailable", ok: true}
	for _, c := range []struct {
		name      string
		callerSet string
		exists    bool
		free      uint64
		mem       headroom
		wantDir   string
		wantWhy   string
	}{
		{"memory when there is room", "", true, plenty, roomy, memRoot, "memory"},
		{"the caller's choice stands", "/somewhere/else", true, plenty, roomy, "", "as set by the caller"},
		{"no tmpfs on this system", "", false, plenty, roomy, "", "no " + memRoot},
		{"too little free for a shared box", "", true, memFloor - 1, roomy, "", "stay on disk"},
		{"free space unreadable", "", true, 0, roomy, "", "stay on disk"},
		// 2026-11-23: /dev/shm had 7 GB free inside a cgroup at its memory.high.
		{"tmpfs roomy, cgroup near its limit", "", true, plenty,
			headroom{bytes: 100 << 20, binding: "cgroup app.slice/agy.service memory.high", ok: true}, "", "memory.high leaves 100 MB"},
		{"memory headroom unreadable", "", true, plenty, headroom{}, "", "could not be read"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := chooseTmp(c.callerSet, c.exists, c.free, c.mem)
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
	if errs := reapStale(root, filepath.Join(root, "no-proc"), now); len(errs) != 0 {
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

	c, errs := reapThenChoose(root, "", filepath.Join(root, "no-proc"), now, func(string) uint64 { return memFloor - 1 },
		func() headroom { return headroom{bytes: memFloor * 2, binding: "MemAvailable", ok: true} })
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

// A DIRECTORY WHOSE OWNER IS DEAD GOES AT ANY AGE. On 2026-11-23 a killed check's 1.2 GB sat in
// /dev/shm 3.5 hours old, under the age rule, while the next run's gate starved beside it. A live
// owner's stays; a pid the kernel has given to another process (start time differs) is dead; a
// directory with no record keeps the age rule, since an older check may own it.
func TestReapStaleJudgesAnOwnedDirectoryByItsOwner(t *testing.T) {
	root := t.TempDir()
	proc := filepath.Join(root, "proc")
	stat := func(pid int, start string) {
		d := filepath.Join(proc, strconv.Itoa(pid))
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		// The command name holds a space and a parenthesis, as /proc allows.
		line := strconv.Itoa(pid) + " (go run) x) S 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 " + start + " 99 100\n"
		if err := os.WriteFile(filepath.Join(d, "stat"), []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	shm := filepath.Join(root, "shm")
	mk := func(name string, o *owner) string {
		p := filepath.Join(shm, memPrefix+name)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if o != nil {
			b, _ := json.Marshal(o)
			if err := os.WriteFile(filepath.Join(p, ownerFile), b, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		return p
	}
	stat(100, "5000")
	stat(200, "9999")
	alive := mk("alive", &owner{PID: 100, Start: "5000"})
	dead := mk("dead", &owner{PID: 300, Start: "7000"})
	reused := mk("reused", &owner{PID: 200, Start: "1234"})
	unowned := mk("unowned", nil)

	if errs := reapStale(shm, proc, time.Now()); len(errs) != 0 {
		t.Fatalf("reapStale reported %v", errs)
	}
	for _, c := range []struct {
		dir  string
		gone bool
	}{{alive, false}, {dead, true}, {reused, true}, {unowned, false}} {
		_, err := os.Stat(c.dir)
		if gone := os.IsNotExist(err); gone != c.gone {
			t.Errorf("%s: removed=%v, want %v", filepath.Base(c.dir), gone, c.gone)
		}
	}
}

// The record a run writes names a process the reap reads as alive: the start time written and
// the one read back are the same field of the same file.
func TestWriteOwnerRecordsALiveOwner(t *testing.T) {
	if _, err := os.Stat("/proc/self/stat"); err != nil {
		t.Skip("no /proc on this system: directories carry no owner record and age decides")
	}
	dir, err := makeRunDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got := readOwner(dir, "/proc"); got != ownerAlive {
		t.Errorf("readOwner = %v for this process's own record, want ownerAlive", got)
	}
}
