package telecli

import (
	"fmt"
	"os"
	"testing"
)

// testRuntimeDir is the runtime directory every test in this package sees as $XDG_RUNTIME_DIR.
var testRuntimeDir string

// TestMain points $XDG_RUNTIME_DIR at a directory of the suite's own before any test runs: every
// verb prunes the Antigravity server records under it, and a test that builds its Env from
// Defaults() would otherwise prune the host's records with an unreleased build.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "telecli-runtime-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	testRuntimeDir = dir
	os.Setenv("XDG_RUNTIME_DIR", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestDefaultsNeverReachTheHostRuntimeDir(t *testing.T) {
	if got := Defaults().RuntimeDir; got != testRuntimeDir || got == "" {
		t.Errorf("Defaults().RuntimeDir = %q, want the suite's own %q", got, testRuntimeDir)
	}
}
