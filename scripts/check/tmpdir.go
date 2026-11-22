package main

import (
	"fmt"
	"os"
	"path/filepath"
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
// MEASURED PEAK: a full `check` with TMPDIR on tmpfs held at most 595 MB there (sampled every
// second, 2026-10-14), test databases and go's build work directories together.
const (
	memRoot = "/dev/shm"
	// memFloor is the free space below which the gates stay on disk: several times the
	// measured peak, because other sessions on a shared box run their own suites beside this
	// one, and a full tmpfs fails every test that writes rather than slowing it.
	memFloor = 3 << 30
	// memStale is how old a directory a previous check left behind must be before it is
	// reclaimed. A check that is killed runs no cleanup, and tmpfs is RAM until reboot; the
	// longest full run measured is under an hour.
	memStale  = 6 * time.Hour
	memPrefix = "sc-check-"
)

// tmpChoice is where the gates' TMPDIR goes, and why — printed with the report, because a gate
// whose timing moved because its disk did is otherwise a mystery.
type tmpChoice struct {
	dir    string // empty: leave TMPDIR as it is
	reason string
}

// chooseTmp decides from what it is told; the caller does the looking. callerSet is the
// TMPDIR the developer chose, if any — a choice that is theirs to make.
func chooseTmp(callerSet string, rootExists bool, free uint64) tmpChoice {
	switch {
	case callerSet != "":
		return tmpChoice{reason: "TMPDIR=" + callerSet + ", as set by the caller"}
	case !rootExists:
		return tmpChoice{reason: "no " + memRoot + " on this system; the gates use the default temp directory"}
	case free < memFloor:
		return tmpChoice{reason: fmt.Sprintf("%s has %d MB free, under the %d MB this run leaves for others; the gates stay on disk",
			memRoot, free>>20, uint64(memFloor)>>20)}
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
	c, errs := reapThenChoose(memRoot, os.Getenv("TMPDIR"), time.Now(), freeBytes)
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, "check: reaping a stale temp directory:", err)
	}
	if c.dir == "" {
		return c.reason, none
	}
	dir, err := os.MkdirTemp(c.dir, memPrefix)
	if err != nil {
		return fmt.Sprintf("could not make a directory in %s (%v); the gates use the default temp directory", c.dir, err), none
	}
	if err := os.Setenv("TMPDIR", dir); err != nil {
		_ = testhome.RemoveTree(dir)
		return fmt.Sprintf("could not set TMPDIR (%v); the gates use the default temp directory", err), none
	}
	return c.reason + " — " + dir, func() error { return testhome.RemoveTree(dir) }
}

// reapThenChoose reaps root, when it exists, and only then measures its free space and chooses.
func reapThenChoose(root, callerSet string, now time.Time, free func(string) uint64) (tmpChoice, []error) {
	_, statErr := os.Stat(root)
	var errs []error
	if statErr == nil {
		errs = reapStale(root, now)
	}
	return chooseTmp(callerSet, statErr == nil, free(root)), errs
}

// reapStale removes what killed checks left in root and returns what it could not remove. Only
// this tool's own prefix, only past memStale: a directory that young may belong to a check still
// running in another session. A gate can leave a read-only tree (Go's module cache), so the removal
// restores write permission where a plain RemoveAll fails.
func reapStale(root string, now time.Time) []error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var errs []error
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), memPrefix) {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) < memStale {
			continue
		}
		if err := testhome.RemoveTree(filepath.Join(root, e.Name())); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
