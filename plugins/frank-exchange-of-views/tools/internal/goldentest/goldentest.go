// Package goldentest is the shared golden-file helper for Go package tests: a rendered
// artifact is compared byte-for-byte against testdata/<name>.golden, and UPDATE_GOLDENS=1
// rewrites it. It is the framework for pinning a WHOLE deterministic artifact (a rendered
// HTML page, a composed report) so a format regression fails loudly, instead of scattering
// dozens of strings.Contains checks that each guard one line and miss everything between them.
//
// The difftest harness carries its own golden loop (its transcript format is bespoke); this
// is for the ordinary case — one function, one artifact, one golden file per case.
package goldentest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Update is set by UPDATE_GOLDENS=1: an update run rewrites every golden it touches, a normal
// run compares. Same env var the difftest harness and the mjs golden loop read, so ONE command
// regenerates every golden in the tree.
var Update = os.Getenv("UPDATE_GOLDENS") == "1"

// Assert compares got against testdata/<name>.golden. On UPDATE it writes the file and returns;
// otherwise a mismatch is a test failure naming the regenerate command. The artifact must be
// DETERMINISTIC — normalize any run-dir path, timestamp, or nonce in `got` before calling, or the
// golden will not reproduce.
func Assert(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if Update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden %q — record it with UPDATE_GOLDENS=1 go test ./... : %v", name, err)
	}
	// Normalize CRLF→LF on the golden read: git checks .golden files out with the platform's
	// line endings on some Windows runners (there is no *.golden eol rule), while the artifact
	// is rendered with LF. The difftest harness does the same, so the two golden loops agree.
	want := strings.ReplaceAll(string(wantBytes), "\r\n", "\n")
	if want != got {
		t.Errorf("%s differs from its golden. If this change is INTENTIONAL, regenerate with "+
			"UPDATE_GOLDENS=1 go test ./... and review the testdata diff on its own.\n"+
			"--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}
