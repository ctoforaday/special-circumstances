package recordtest

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// PosixLocksHeld counts the POSIX advisory locks THIS PROCESS holds on the file at path, read from
// /proc/locks — the kernel's own table, which is what another process's SQLite consults before it
// decides it is the last connection to a record.
//
// LINUX ONLY (locks_other.go skips the test elsewhere): the defect it measures is Linux's rule that closing
// any descriptor to a file drops every lock the process holds on it, and only Linux publishes the
// table. A lock count is the measurement because a dropped lock is invisible to the connection
// that lost it — every query through it still answers.
func PosixLocksHeld(t *testing.T, path string) int {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("recordtest: stat %s: %v", path, err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("recordtest: no inode for %s", path)
	}
	f, err := os.Open("/proc/locks")
	if err != nil {
		t.Fatalf("recordtest: %v", err)
	}
	defer f.Close()
	pid := strconv.Itoa(os.Getpid())
	ino := strconv.FormatUint(st.Ino, 10)
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// "1: POSIX  ADVISORY  READ 12345 08:01:4567 1073741826 1073742335"; a blocked waiter
		// carries a "->" field first and is not a held lock.
		fields := strings.Fields(sc.Text())
		if len(fields) < 6 || fields[1] != "POSIX" {
			continue
		}
		dev := strings.Split(fields[5], ":")
		if fields[4] == pid && len(dev) == 3 && dev[2] == ino {
			n++
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("recordtest: reading /proc/locks: %v", err)
	}
	return n
}
