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
	defer os.RemoveAll(dir)
	// USERPROFILE as well as HOME: os.UserHomeDir reads the first on Unix and the second on
	// Windows, and gray-area's suite already failed on the Windows leg for exactly this.
	for _, k := range []string{"HOME", "USERPROFILE", "XDG_STATE_HOME"} {
		if err := os.Setenv(k, dir); err != nil {
			fmt.Fprintf(os.Stderr, "hooktest: setting %s: %v\n", k, err)
			return 1
		}
	}
	return m.Run()
}
