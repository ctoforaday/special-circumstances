package testbuild

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// THE SANDBOX HOME IS GONE AFTER ITS UNDO, even when a test left a read-only tree in it. Go
// writes its module cache that way — directories without write permission — and the undo's
// RemoveAll failed on it with the error discarded: one leaked directory per test binary, 30 GB
// of them on 2026-11-22 before the disk filled.
func TestTheSandboxHomeIsRemovedEvenWhenReadOnly(t *testing.T) {
	// The undo puts HOME back to the developer's; put back Main's sandbox after this test.
	t.Setenv(homeKey(), os.Getenv(homeKey()))
	t.Setenv("GOMODCACHE", os.Getenv("GOMODCACHE"))
	t.Setenv("GOCACHE", os.Getenv("GOCACHE"))

	undo, err := sandboxHome()
	if err != nil {
		t.Fatal(err)
	}
	h := os.Getenv(homeKey())
	locked := filepath.Join(h, "go", "pkg", "mod", "example.com", "m@v1.0.0")
	if err := os.MkdirAll(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locked, "go.mod"), []byte("module m\n"), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { // if the undo fails, do not leak the fixture as well
		_ = os.Chmod(locked, 0o755)
		_ = os.RemoveAll(h)
	})

	if err := undo(); err != nil {
		t.Fatalf("the undo reported: %v", err)
	}
	if _, err := os.Stat(h); !os.IsNotExist(err) {
		t.Fatalf("the sandbox HOME %s is still there after its undo (stat: %v)", h, err)
	}
}

// A `go` STARTED UNDER THE SANDBOX USES THE REAL MODULE CACHE, so nothing read-only lands in the
// sandbox at all. nativehelpertest's build inherits this process's environment; before the
// caches were pinned it resolved GOPATH under the sandbox and downloaded the module graph there.
func TestAGoChildKeepsTheModuleCacheOutOfTheSandbox(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go on PATH")
	}
	h := os.Getenv(homeKey())
	out, err := exec.Command("go", "env", "GOMODCACHE", "GOCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range strings.Fields(string(out)) {
		if rel, err := filepath.Rel(h, p); err == nil && !strings.HasPrefix(rel, "..") {
			t.Errorf("a go child resolves %s, inside the sandbox HOME %s", p, h)
		}
	}
}
