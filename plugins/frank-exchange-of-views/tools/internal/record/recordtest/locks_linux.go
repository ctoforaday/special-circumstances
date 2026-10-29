package recordtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// LocksHeld counts the advisory locks THIS PROCESS holds on the file at path — POSIX and OFD
// alike — read from /proc/self/fdinfo, the kernel's own per-descriptor lock table. It is what
// another process's SQLite consults before it decides it is the last connection to a record.
// measured is true: this platform publishes the table.
//
// A lock count is the measurement because a dropped lock is invisible to the connection that lost
// it — every query through it still answers. A POSIX lock is dropped by any close of any descriptor
// on the file; an OFD lock (recordsql's ofd.go, the database file's locks on Linux) is not.
//
// fdinfo and not /proc/locks, because /proc/locks prints an OFD lock with pid -1: owned by a
// descriptor, not a process, so a row there cannot say whose it is. fdinfo lists, under each of
// this process's descriptors, the locks placed through it, so a lock is counted once. A
// descriptor is matched to path by stat — device AND inode — and stat opens nothing: an open of
// the record here would be the very close this measures.
func LocksHeld(t *testing.T, path string) (n int, measured bool) {
	t.Helper()
	want, err := os.Stat(path)
	if err != nil {
		t.Fatalf("recordtest: stat %s: %v", path, err)
	}
	fds, err := os.ReadDir("/proc/self/fdinfo")
	if err != nil {
		t.Fatalf("recordtest: %v", err)
	}
	for _, fd := range fds {
		fi, err := os.Stat(filepath.Join("/proc/self/fd", fd.Name()))
		if err != nil || !os.SameFile(fi, want) {
			continue // closed since the listing, or another file
		}
		info, err := os.ReadFile(filepath.Join("/proc/self/fdinfo", fd.Name()))
		if err != nil {
			continue // closed since the listing
		}
		// "lock:\t1: OFDLCK ADVISORY  READ -1 08:01:4567 1073741826 1073742335" — a blocked
		// waiter carries a "->" field after the ordinal and is not a held lock.
		for _, line := range strings.Split(string(info), "\n") {
			rest, ok := strings.CutPrefix(line, "lock:")
			if !ok {
				continue
			}
			fields := strings.Fields(rest)
			if len(fields) < 2 || fields[1] == "->" {
				continue
			}
			n++
		}
	}
	return n, true
}
