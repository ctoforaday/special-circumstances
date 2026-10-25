//go:build !linux

package recordtest

import "testing"

// PosixLocksHeld skips the test: the POSIX lock table is read from /proc/locks, which only Linux
// publishes, and the descriptor-close rule it measures is Linux's (locks_linux.go).
func PosixLocksHeld(t *testing.T, path string) int {
	t.Helper()
	t.Skip("the POSIX lock table is read from /proc/locks, which only Linux has")
	return 0
}
