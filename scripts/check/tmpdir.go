package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
func useMemoryTmp() (string, func()) {
	_, statErr := os.Stat(memRoot)
	c := chooseTmp(os.Getenv("TMPDIR"), statErr == nil, freeBytes(memRoot))
	if c.dir == "" {
		return c.reason, func() {}
	}
	reapStale(c.dir, time.Now())
	dir, err := os.MkdirTemp(c.dir, memPrefix)
	if err != nil {
		return fmt.Sprintf("could not make a directory in %s (%v); the gates use the default temp directory", c.dir, err), func() {}
	}
	if err := os.Setenv("TMPDIR", dir); err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Sprintf("could not set TMPDIR (%v); the gates use the default temp directory", err), func() {}
	}
	return c.reason + " — " + dir, func() { _ = os.RemoveAll(dir) }
}

// reapStale removes what killed checks left in root. Only this tool's own prefix, only past
// memStale: a directory that young may belong to a check still running in another session.
func reapStale(root string, now time.Time) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), memPrefix) {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) < memStale {
			continue
		}
		_ = os.RemoveAll(filepath.Join(root, e.Name()))
	}
}
