//go:build unix

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
)

// holdPresenceLock opens (creating if needed) <agyDir>/presence/<conv>.lock and acquires an
// exclusive lock on it for the duration of the returned release func.
func holdPresenceLock(t *testing.T, agyDir, conv string) func() {
	t.Helper()
	dir := filepath.Join(agyDir, "presence")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fl := flock.New(filepath.Join(dir, conv+".lock"), flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	ok, err := fl.TryLock()
	if err != nil || !ok {
		t.Fatalf("TryLock presence/%s.lock = (%v, %v)", conv, ok, err)
	}
	return func() {
		_ = fl.Unlock()
	}
}
