package hooktest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This package's own tests run isolated, so what they observe is what every package that uses
// Isolated gets.
func TestMain(m *testing.M) { os.Exit(Isolated(m)) }

func goEnv(t *testing.T, key string) string {
	t.Helper()
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		t.Skipf("no go on PATH to ask (%v) — the cache half of the isolation was NOT checked here", err)
	}
	return strings.TrimSpace(string(out))
}

func TestHomeIsATemporaryDirectory(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(filepath.Base(home), "pc-hook-test-") {
		t.Errorf("the home directory is %q — a test in an isolated package would write to the developer's own state", home)
	}
	if got := os.Getenv("XDG_STATE_HOME"); got != home {
		t.Errorf("XDG_STATE_HOME = %q, want the isolated home %q", got, home)
	}
}

// THE REDIRECT MUST NOT TAKE THE GO CACHES WITH IT. Both default to paths under the home
// directory; left to follow it, every `go build` a test runs starts from an empty cache in a
// directory that is deleted when the binary exits, and compiles the standard library again.
// Nothing fails when that happens — the suite is only slow — so the cost is asserted here.
func TestAGoSubprocessStillSeesTheRealCaches(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"GOCACHE", "GOMODCACHE"} {
		got := goEnv(t, key)
		if got == "" {
			t.Errorf("go env %s is empty", key)
		}
		if strings.HasPrefix(got, home) {
			t.Errorf("go env %s = %q, inside the isolated home %q — a go subprocess gets an empty cache it rebuilds on every run", key, got, home)
		}
	}
}

// A value the caller chose is the caller's: CI and a sandboxed box both set these on purpose.
func TestACacheTheCallerSetIsLeftAlone(t *testing.T) {
	chosen := filepath.Join(t.TempDir(), "chosen")
	t.Setenv("GOCACHE", chosen)
	keepGoCaches()
	if got := os.Getenv("GOCACHE"); got != chosen {
		t.Errorf("GOCACHE = %q after keepGoCaches, want the caller's %q", got, chosen)
	}
}
