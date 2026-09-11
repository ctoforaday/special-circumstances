package catalogue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A SCAN THAT STOPS EARLY REPORTS UNKNOWN, never the mode it had seen so far: the line past the cap
// may hold a newer prompt in a different mode, and resuming in the older one would be a state the
// session had left.
func TestPermissionModeIsUnknownWhenTheReadStopsEarly(t *testing.T) {
	old := permissionModeLineCap
	permissionModeLineCap = 1024
	t.Cleanup(func() { permissionModeLineCap = old })

	path := filepath.Join(t.TempDir(), "s.jsonl")
	body := `{"type":"user","permissionMode":"plan","message":{"role":"user","content":"hi"}}` + "\n" +
		`{"type":"user","permissionMode":"auto","message":{"role":"user","content":"` + strings.Repeat("x", 4096) + `"}}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := PermissionMode(path); got != "" {
		t.Errorf("PermissionMode = %q after a line past the cap, want \"\" (unknown)", got)
	}

	// The same file under the real cap reads through to the newest mode, so the case above fails
	// for the reason it names.
	permissionModeLineCap = old
	if got := PermissionMode(path); got != "auto" {
		t.Errorf("PermissionMode = %q under the real cap, want \"auto\"", got)
	}
}
