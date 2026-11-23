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
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// dir is the isolated home of this test binary, "" outside Isolated.
var dir string

// Isolated runs a package's tests with every home-ish variable pointed at a temporary directory,
// removed at exit; a directory that survives fails the package (see shared/testhome). Use it as
// the whole body of TestMain:
//
//	func TestMain(m *testing.M) { os.Exit(hooktest.Isolated(m)) }
//
// A test that wants its own may still override with t.Setenv.
//
// The Go build and module caches are pinned to where they were before the redirect: both default
// to paths under the home directory, so a test that shells out to `go build` was handed an EMPTY
// cache in a directory about to be deleted — and compiled the standard library and every
// dependency from nothing, on every run. Measured in internal/doctor: the test that builds every
// command for every published platform took 49.5s of this module's 62s suite, and 15.7s with the
// caches kept.
func Isolated(m *testing.M) int {
	d, undo, err := testhome.New("pc-hook-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "hooktest:", err)
		return 1
	}
	dir = d
	code := m.Run()
	if err := undo(); err != nil {
		fmt.Fprintln(os.Stderr, "hooktest:", err)
		if code == 0 {
			code = 1
		}
	}
	return code
}

// Dir is the isolated home this test binary runs under: a place for what the whole binary shares
// (a binary built once) that the exit removes with everything else. A test's own t.Setenv of HOME
// does not move it.
func Dir() string { return dir }
