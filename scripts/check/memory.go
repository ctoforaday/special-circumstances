package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MEMORY THIS PROCESS CAN ACTUALLY GET, NOT THE SIZE OF A FILESYSTEM.
//
// tmpfs pages are RAM charged to the cgroup of whoever writes them, and a gate's heap is charged
// to the same cgroup. /dev/shm's free space says nothing about either: on 2026-11-23 a check run
// inside a daemon capped at memory.high=6G saw 7 GB free in /dev/shm, put 1.3 GB of test
// databases there beside a 1.2 GB leftover, and its deadcode gate (3.1 GB) then sat throttled in
// reclaim for 2h43m, stalling every other process in the daemon's cgroup. The headroom is the
// smallest of what the kernel says is available and what each cgroup from this process's up to
// the root leaves under its memory.high and memory.max.

// headroom is a reading of memory this process can still allocate. ok is false where nothing
// could be read; binding names the limit that set the number, for the printed reason.
type headroom struct {
	bytes   uint64
	binding string
	ok      bool
}

func (h headroom) String() string {
	if !h.ok {
		return "memory headroom not measured"
	}
	return fmt.Sprintf("%s leaves %d MB", h.binding, h.bytes>>20)
}

// readHeadroom reads /proc/meminfo's MemAvailable and the memory.high, memory.max and
// memory.current of every cgroup v2 ancestor named by /proc/self/cgroup. procRoot and cgRoot are
// "/proc" and "/sys/fs/cgroup" outside tests.
func readHeadroom(procRoot, cgRoot string) headroom {
	h := headroom{}
	if avail, ok := memAvailable(filepath.Join(procRoot, "meminfo")); ok {
		h = headroom{bytes: avail, binding: "MemAvailable", ok: true}
	}
	rel, ok := unifiedCgroup(filepath.Join(procRoot, "self", "cgroup"))
	if !ok {
		return h
	}
	for dir := filepath.Join(cgRoot, rel); ; dir = filepath.Dir(dir) {
		cur, curOK := readCgroupValue(filepath.Join(dir, "memory.current"))
		if curOK {
			for _, lim := range []string{"memory.high", "memory.max"} {
				limit, limOK := readCgroupValue(filepath.Join(dir, lim))
				if !limOK {
					continue
				}
				left := uint64(0)
				if limit > cur {
					left = limit - cur
				}
				if !h.ok || left < h.bytes {
					name, _ := filepath.Rel(cgRoot, dir)
					h = headroom{bytes: left, binding: "cgroup " + name + " " + lim, ok: true}
				}
			}
		}
		if filepath.Clean(dir) == filepath.Clean(cgRoot) || dir == filepath.Dir(dir) {
			break
		}
	}
	return h
}

// memAvailable returns MemAvailable from a meminfo file, in bytes.
func memAvailable(path string) (uint64, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) == 3 && f[0] == "MemAvailable:" && f[2] == "kB" {
			kb, err := strconv.ParseUint(f[1], 10, 64)
			if err != nil {
				return 0, false
			}
			return kb << 10, true
		}
	}
	return 0, false
}

// unifiedCgroup returns this process's cgroup v2 path from a /proc/<pid>/cgroup file: the line
// "0::<path>". A v1-only host has none.
func unifiedCgroup(path string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(b), "\n") {
		if rel, ok := strings.CutPrefix(line, "0::"); ok {
			return strings.TrimPrefix(rel, "/"), true
		}
	}
	return "", false
}

// readCgroupValue reads a cgroup memory file holding a byte count. "max" (no limit) and a missing
// file both report false: neither limits anything.
func readCgroupValue(path string) (uint64, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	v, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return v, err == nil
}
