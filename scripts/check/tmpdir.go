package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// THE GATES' TEMP DIRECTORY IS MEMORY WHEN THE BOX HAS SOME TO SPARE.
//
// Every FEOV test database is created under TMPDIR, and SQLite keeps the record's write
// guarantee (synchronous FULL, by ruling) with an fsync per commit. On a disk that is most of
// the cost: one test that migrates every archived run took 47s with TMPDIR on disk and 15s on
// tmpfs, the same code. CI's Linux jobs already point TMPDIR at /dev/shm; this is the local
// gate doing what CI does. The fsyncs still happen — they land in RAM.
//
// MEASURED PEAK: a full `check` with TMPDIR on tmpfs held at most 1109 MB there (sampled every
// second, 2026-11-23, 65 gates), test databases, test homes and go's build work directories
// together.
//
// The floor applies to BOTH /dev/shm's free space and the memory this process can get
// (headroom, memory.go): a tmpfs page is RAM charged to this process's cgroup, so a /dev/shm with
// room in it says nothing about a cgroup near its memory.high.
const (
	memRoot = "/dev/shm"
	// memFloor is the free space below which the gates stay on disk: nearly three times the
	// measured peak, because other sessions on a shared box run their own suites beside this
	// one, and a full tmpfs fails every test that writes rather than slowing it.
	memFloor = 3 << 30
	// memStale is how old a directory with no owner record must be before it is reclaimed: one
	// made by a check built before the record existed, which may still be running in another
	// worktree. A directory with an owner record goes as soon as its owner is dead.
	memStale  = 6 * time.Hour
	memPrefix = "sc-check-"
	// ownerFile is the record in each directory naming the process that made it.
	ownerFile = "owner.json"
)

// owner names the process a check's directory belongs to. start is the process's start time
// from /proc/<pid>/stat, so a pid the kernel has since given to another process does not read
// as the owner still running.
type owner struct {
	PID   int    `json:"pid"`
	Start string `json:"start"`
}

// selfOwner is this process, or false where its start time cannot be read (no /proc).
func selfOwner() (owner, bool) {
	start, ok := procStart("/proc", os.Getpid())
	return owner{PID: os.Getpid(), Start: start}, ok
}

// procStart returns field 22 (starttime) of /proc/<pid>/stat. The command name, field 2, is in
// parentheses and may hold spaces, so the fields are counted from after its closing one.
func procStart(procRoot string, pid int) (string, bool) {
	b, err := os.ReadFile(filepath.Join(procRoot, strconv.Itoa(pid), "stat"))
	if err != nil {
		return "", false
	}
	i := strings.LastIndexByte(string(b), ')')
	if i < 0 {
		return "", false
	}
	f := strings.Fields(string(b[i+1:]))
	const startAfterComm = 22 - 3 // field 3 is the first after the command name
	if len(f) <= startAfterComm {
		return "", false
	}
	return f[startAfterComm], true
}

// ownerState says whether the process a directory's owner record names is still running.
type ownerState int

const (
	ownerUnknown ownerState = iota // no record, or no way to check it: age decides
	ownerAlive
	ownerDead
)

func readOwner(dir, procRoot string) ownerState {
	b, err := os.ReadFile(filepath.Join(dir, ownerFile))
	if err != nil {
		return ownerUnknown
	}
	var o owner
	if json.Unmarshal(b, &o) != nil || o.PID <= 0 || o.Start == "" {
		return ownerUnknown
	}
	if _, err := os.Stat(procRoot); err != nil {
		return ownerUnknown
	}
	start, ok := procStart(procRoot, o.PID)
	if ok && start == o.Start {
		return ownerAlive
	}
	return ownerDead
}

// tmpChoice is where the gates' TMPDIR goes, and why — printed with the report, because a gate
// whose timing moved because its disk did is otherwise a mystery.
type tmpChoice struct {
	dir    string // empty: leave TMPDIR as it is
	reason string
}

// chooseTmp decides from what it is told; the caller does the looking. callerSet is the
// TMPDIR the developer chose, if any — a choice that is theirs to make.
func chooseTmp(callerSet string, rootExists bool, free uint64, mem headroom) tmpChoice {
	switch {
	case callerSet != "":
		return tmpChoice{reason: "TMPDIR=" + callerSet + ", as set by the caller"}
	case !rootExists:
		return tmpChoice{reason: "no " + memRoot + " on this system; the gates use the default temp directory"}
	case free < memFloor:
		return tmpChoice{reason: fmt.Sprintf("%s has %d MB free, under the %d MB this run leaves for others; the gates stay on disk",
			memRoot, free>>20, uint64(memFloor)>>20)}
	case !mem.ok:
		return tmpChoice{reason: "the memory this run can get could not be read; the gates stay on disk"}
	case mem.bytes < memFloor:
		return tmpChoice{reason: fmt.Sprintf("%s, under the %d MB this run leaves for others; the gates stay on disk",
			mem, uint64(memFloor)>>20)}
	}
	return tmpChoice{dir: memRoot, reason: "memory (" + memRoot + "): test databases fsync to RAM, as in CI"}
}

// useMemoryTmp applies the choice: a fresh directory under memRoot becomes TMPDIR for every gate
// this process starts. The returned func removes it, and must run before the process exits.
//
// What killed checks left behind is reaped FIRST, whatever the choice: the reap is what frees the
// space the floor measures, so reaping only above the floor never reaps a box the leftovers hold
// just under it — 28 directories, 4.9 GB of RAM, 9 to 12 days old, on 2026-11-22. A directory the
// reap cannot remove is printed; the release's error is the run's own directory surviving.
func useMemoryTmp() (string, func() error) {
	none := func() error { return nil }
	c, errs := reapThenChoose(memRoot, os.Getenv("TMPDIR"), "/proc", time.Now(), freeBytes,
		func() headroom { return readHeadroom("/proc", "/sys/fs/cgroup") })
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, "check: reaping a stale temp directory:", err)
	}
	if c.dir == "" {
		return c.reason, none
	}
	dir, err := makeRunDir(c.dir)
	if err != nil {
		return fmt.Sprintf("%v; the gates use the default temp directory", err), none
	}
	if err := os.Setenv("TMPDIR", dir); err != nil {
		_ = testhome.RemoveTree(dir)
		return fmt.Sprintf("could not set TMPDIR (%v); the gates use the default temp directory", err), none
	}
	return c.reason + " — " + dir, func() error { return testhome.RemoveTree(dir) }
}

// makeRunDir makes this run's directory under root and records this process as its owner, so a
// later run removes it as soon as this one is dead rather than hours later.
func makeRunDir(root string) (string, error) {
	dir, err := os.MkdirTemp(root, memPrefix)
	if err != nil {
		return "", fmt.Errorf("could not make a directory in %s (%v)", root, err)
	}
	if err := writeOwner(dir); err != nil {
		_ = testhome.RemoveTree(dir)
		return "", fmt.Errorf("could not record this run as the owner of %s (%v)", dir, err)
	}
	return dir, nil
}

// writeOwner records this process as dir's owner. Where no start time can be read the directory
// carries no record, and a later reap judges it by age.
func writeOwner(dir string) error {
	o, ok := selfOwner()
	if !ok {
		return nil
	}
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ownerFile), b, 0o600)
}

// reapThenChoose reaps root, when it exists, and only then measures its free space and the
// memory this process can get, and chooses.
func reapThenChoose(root, callerSet, procRoot string, now time.Time, free func(string) uint64, mem func() headroom) (tmpChoice, []error) {
	_, statErr := os.Stat(root)
	var errs []error
	if statErr == nil {
		errs = reapStale(root, procRoot, now)
	}
	return chooseTmp(callerSet, statErr == nil, free(root), mem()), errs
}

// reapStale removes what killed checks left in root and returns what it could not remove. Only
// this tool's own prefix. A directory whose owner record names a dead process goes at any age; a
// live owner's stays; one with no record goes past memStale, since a directory that young may
// belong to a check still running in another session. A gate can leave a read-only tree (Go's
// module cache), so the removal restores write permission where a plain RemoveAll fails.
func reapStale(root, procRoot string, now time.Time) []error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var errs []error
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), memPrefix) {
			continue
		}
		dir := filepath.Join(root, e.Name())
		switch readOwner(dir, procRoot) {
		case ownerAlive:
			continue
		case ownerUnknown:
			info, err := e.Info()
			if err != nil || now.Sub(info.ModTime()) < memStale {
				continue
			}
		}
		if err := testhome.RemoveTree(dir); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
