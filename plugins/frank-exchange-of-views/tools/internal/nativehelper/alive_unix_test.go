//go:build unix

package nativehelper

import (
	"syscall"
	"testing"
)

// assertGone fails the test when the process pid still exists. The client reaps what it kills, so
// a killed helper leaves no zombie that would answer signal 0.
func assertGone(t *testing.T, pid int) {
	t.Helper()
	if err := syscall.Kill(pid, 0); err == nil {
		t.Errorf("helper process %d is still running", pid)
	}
}
