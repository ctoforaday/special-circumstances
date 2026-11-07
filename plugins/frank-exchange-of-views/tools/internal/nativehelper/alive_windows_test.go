//go:build windows

package nativehelper

import "testing"

// assertGone is unmeasured on Windows: package os and syscall have no signal 0 there, and a
// liveness probe needs OpenProcess and GetExitCodeProcess. The unix legs carry the check; the
// kill itself is the same os.Process.Kill on every platform.
func assertGone(t *testing.T, pid int) {
	t.Helper()
	t.Logf("NOT MEASURED on Windows: whether helper process %d is gone", pid)
}
