//go:build unix

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// holdPresenceLock opens (creating if needed) <agyDir>/presence/<conv>.lock and acquires an
// exclusive flock on it for the duration of the returned release func.
func holdPresenceLock(t *testing.T, agyDir, conv string) func() {
	t.Helper()
	dir := filepath.Join(agyDir, "presence")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(dir, conv+".lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		t.Fatalf("flock presence/%s.lock: %v", conv, err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
}
