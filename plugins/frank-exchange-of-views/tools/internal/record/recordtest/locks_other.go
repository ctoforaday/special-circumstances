//go:build !linux

package recordtest

import "testing"

// PosixLocksHeld reports the lock table unmeasured: it is read from /proc/locks, which only Linux
// publishes. The rule it would measure — closing any descriptor to a file drops every POSIX lock
// the process holds on it — is POSIX fcntl semantics and holds here too, so a caller asserts its
// lock counts only when measured and runs every other assertion regardless (locks_linux.go).
func PosixLocksHeld(t *testing.T, path string) (n int, measured bool) {
	t.Helper()
	return 0, false
}
