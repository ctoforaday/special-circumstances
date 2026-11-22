// Package hooktest keeps a test run out of the developer's own state.
//
// The failure record (internal/hookfailures) resolves its path from XDG_STATE_HOME or the home
// directory, so a test that drives a hook end to end writes to ~/.local/state unless something
// stops it — and one did, on the first run of this package's own wiring: a real record, with real
// entries, from a test fixture. The mtime cannot prove it either way afterwards, which is why the
// isolation is set for a WHOLE package here rather than trusted test by test.
//
// Test-only, and it ships with the plugin because the plugin's tests are what need it: a consumer
// who runs `go test ./...` in an installed tools/ tree must be protected by the same rule.
package hooktest

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Isolated runs a package's tests with every home-ish variable pointed at a temporary directory.
// Use it as the whole body of TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(hooktest.Isolated(m)) }
//
// A test that wants its own may still override with t.Setenv.
func Isolated(m *testing.M) int {
	dir, err := os.MkdirTemp("", "pc-hook-test-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "hooktest: no temp dir to isolate state in: %v\n", err)
		return 1
	}
	keepGoCaches()
	// USERPROFILE as well as HOME: os.UserHomeDir reads the first on Unix and the second on
	// Windows, and gray-area's suite already failed on the Windows leg for exactly this.
	for _, k := range []string{"HOME", "USERPROFILE", "XDG_STATE_HOME"} {
		if err := os.Setenv(k, dir); err != nil {
			fmt.Fprintf(os.Stderr, "hooktest: setting %s: %v\n", k, err)
			_ = removeTree(dir)
			return 1
		}
	}
	code := m.Run()
	// A directory this package made and could not remove fails the package: `go test` prints a
	// passing package's stderr only under -v, so a warning on a green run is read by nobody.
	// frank-exchange-of-views' testbuild leaked a read-only module cache on every run that way.
	if err := removeTree(dir); err != nil {
		fmt.Fprintf(os.Stderr, "hooktest: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	return code
}

// removeTree removes dir and everything under it, restoring write permission on every
// directory first when a plain RemoveAll fails — a read-only tree a test left behind still goes —
// and returns what survives that.
func removeTree(dir string) error {
	if err := os.RemoveAll(dir); err == nil {
		return nil
	}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			if info, ierr := d.Info(); ierr == nil {
				_ = os.Chmod(p, info.Mode().Perm()|0o700)
			}
		}
		return nil
	})
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("could not remove %s: %w", dir, err)
	}
	return nil
}

// keepGoCaches pins the Go build and module caches to where they are NOW, before the home
// directory moves.
//
// Both default to paths under the home directory, so a test that shells out to `go build` after
// the redirect was handed an EMPTY cache in a directory about to be deleted — and compiled the
// standard library and every dependency from nothing, on every run. Measured in internal/doctor:
// the test that builds every command for every published platform took 49.5s of this module's
// 62s suite, and 15.7s with the caches kept.
//
// They are not the state this package exists to protect. The build cache is content-addressed
// and safe under concurrent writers, and nothing a test asserts on is read from either. A value
// the caller already set is left alone, and a machine with no `go` on PATH is left as it was:
// a test that needs one says so itself.
func keepGoCaches() {
	out, err := exec.Command("go", "env", "GOCACHE", "GOMODCACHE").Output()
	if err != nil {
		return
	}
	paths := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, k := range []string{"GOCACHE", "GOMODCACHE"} {
		if i < len(paths) && strings.TrimSpace(paths[i]) != "" && os.Getenv(k) == "" {
			_ = os.Setenv(k, strings.TrimSpace(paths[i]))
		}
	}
}
