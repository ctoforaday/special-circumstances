//go:build !linux

package recordtest

import "testing"

// LocksHeld reports the lock table unmeasured: it is read from /proc/self/fdinfo, which only Linux
// publishes. The rule it would measure — closing any descriptor to a file drops every POSIX lock
// the process holds on it — holds on macOS and the BSDs too, so a caller asserts its lock counts
// only when measured and runs every other assertion regardless (locks_linux.go).
func LocksHeld(t *testing.T, path string) (n int, measured bool) {
	t.Helper()
	return 0, false
}
