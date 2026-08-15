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
	"fmt"
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
		t.Errorf("%s differs from its golden.\n%s\n"+
			"To review this properly: `go run ./golden -review` from scripts/ — it verifies first, "+
			"shows each proposed diff, and accepts NOTHING unless you name it. "+
			"`UPDATE_GOLDENS=1` rewrites every golden without ever showing you this failure.",
			name, firstDivergence(want, got))
	}
}

// firstDivergence renders the mismatch as the first differing line plus its neighbours,
// rather than dumping both artifacts whole.
//
// THE DUMP WAS THE PROBLEM. Printing `got` and `want` in full turns one changed line in a
// rendered HTML page into hundreds of lines of identical output, and the reader's cheapest
// move becomes "regenerate and move on" — which is the review erosion the golden tooling
// exists to prevent, reintroduced by its own failure message. Showing WHERE they part makes
// reading the diff cheaper than skipping it.
func firstDivergence(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	n := len(w)
	if len(g) < n {
		n = len(g)
	}
	at := -1
	for i := 0; i < n; i++ {
		if w[i] != g[i] {
			at = i
			break
		}
	}
	if at == -1 {
		// Identical up to the shorter one: the difference is length alone, which a
		// line-by-line scan would otherwise report as "no difference found".
		return fmt.Sprintf("  identical for %d line(s), then the length differs: golden has %d line(s), got %d",
			n, len(w), len(g))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "  first difference at line %d:\n", at+1)
	for i := at - 2; i <= at+2; i++ {
		if i < 0 {
			continue
		}
		if i == at {
			if i < len(w) {
				fmt.Fprintf(&b, "  - golden %4d| %s\n", i+1, w[i])
			}
			if i < len(g) {
				fmt.Fprintf(&b, "  + got    %4d| %s\n", i+1, g[i])
			}
			continue
		}
		if i < len(w) && i < len(g) && w[i] == g[i] {
			fmt.Fprintf(&b, "    same   %4d| %s\n", i+1, w[i])
		}
	}
	fmt.Fprintf(&b, "  (%d golden line(s) vs %d got)", len(w), len(g))
	return b.String()
}
