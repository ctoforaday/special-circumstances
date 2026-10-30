package compute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// OUTPUT DRIFT NAMES THE BYTE, not the rune: é and è share their first byte and part at the second.
func TestOutputDriftNamesTheFirstDifferingByte(t *testing.T) {
	if got, want := outputDrift("é", "è"), "output differs from byte 1 between runs"; got != want {
		t.Errorf("outputDrift(é, è) = %q, want %q", got, want)
	}
}

// AN EXCERPT IS VALID UTF-8 AND SAYS WHAT THE PROGRAM WROTE: valid text as written, cut on a rune
// boundary; text that is not UTF-8 Go-quoted, every byte kept.
func TestAnExcerptIsValidAndTruthful(t *testing.T) {
	for _, c := range []struct {
		in    string
		limit int
		want  string
	}{
		{"plain", 400, "plain"},
		{"aé", 2, "a…"},
		{"aéb", 3, "aé…"},
		{"\xff\x00x", 400, `"\xff\x00x"`},
		{"\xffabc", 2, `"\xffa"…`},
	} {
		got := Excerpt(c.in, c.limit)
		if got != c.want || !utf8.ValidString(got) {
			t.Errorf("Excerpt(%q, %d) = %q, want %q", c.in, c.limit, got, c.want)
		}
	}
}

// A BUNDLE THE STORE CANNOT READ SAYS WHY. Rerun carries the store's own error rather than calling
// every failure a missing bundle: here dir/ holds a link to a directory, which reads as no file.
func TestRerunCarriesTheStoresError(t *testing.T) {
	runDir := t.TempDir()
	dir := BundlePath(runDir, "abc")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(".", filepath.Join(dir, "loop")); err != nil {
		t.Skip("no symlinks here:", err)
	}
	_, err := Rerun(runDir, "abc", []string{"bash", "loop"})
	if err == nil || !strings.Contains(err.Error(), "is a directory") {
		t.Errorf("Rerun over an unreadable bundle: %v, want the store's own error", err)
	}
	_, err = Rerun(t.TempDir(), "abc", []string{"bash", "loop"})
	if err == nil || !strings.Contains(err.Error(), "not in the store") {
		t.Errorf("Rerun over a missing bundle: %v, want the store's refusal naming it missing", err)
	}
}
