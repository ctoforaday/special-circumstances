package recordtest

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

// PosixLocksHeld counts the POSIX advisory locks THIS PROCESS holds on the file at path, read from
// /proc/locks — the kernel's own table, which is what another process's SQLite consults before it
// decides it is the last connection to a record. measured is true: this platform publishes the
// table.
//
// The defect it measures is POSIX fcntl semantics — closing any descriptor to a file drops every
// lock the process holds on it — and holds on macOS and the BSDs as much as here; only the table
// it is read from is Linux's (locks_other.go reports it unmeasured elsewhere). A lock count is the
// measurement because a dropped lock is invisible to the connection that lost it — every query
// through it still answers.
//
// A row matches on device AND inode: an inode number is unique only within its filesystem.
func PosixLocksHeld(t *testing.T, path string) (n int, measured bool) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("recordtest: stat %s: %v", path, err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("recordtest: no inode for %s", path)
	}
	major, minor := uint64(unix.Major(st.Dev)), uint64(unix.Minor(st.Dev))
	f, err := os.Open("/proc/locks")
	if err != nil {
		t.Fatalf("recordtest: %v", err)
	}
	defer f.Close()
	pid := strconv.Itoa(os.Getpid())
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// "1: POSIX  ADVISORY  READ 12345 08:01:4567 1073741826 1073742335" — major and minor in
		// hex, the inode in decimal. A blocked waiter carries a "->" field first and is not a held
		// lock.
		fields := strings.Fields(sc.Text())
		if len(fields) < 6 || fields[1] != "POSIX" || fields[4] != pid {
			continue
		}
		id := strings.Split(fields[5], ":")
		if len(id) != 3 {
			continue
		}
		maj, err1 := strconv.ParseUint(id[0], 16, 64)
		min, err2 := strconv.ParseUint(id[1], 16, 64)
		ino, err3 := strconv.ParseUint(id[2], 10, 64)
		if err1 == nil && err2 == nil && err3 == nil && maj == major && min == minor && ino == st.Ino {
			n++
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("recordtest: reading /proc/locks: %v", err)
	}
	return n, true
}
