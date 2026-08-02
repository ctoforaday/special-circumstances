package report

import (
	"strings"
	"testing"
)

// The load-bearing no-leak guarantee: every finding-anchor "[^f-<id>]" is stripped, and
// blue's own claim footnotes ("[^L1]") are left untouched. Assemble applies this once at
// the blue read, so every downstream lift is marker-free by construction.
func TestStripFindingMarkers(t *testing.T) {
	in := "A claim.[^f-abc] Another one[^L1], and a third[^f-9f2a1c].\n"
	got := StripFindingMarkers(in)
	want := "A claim. Another one[^L1], and a third.\n"
	if got != want {
		t.Errorf("StripFindingMarkers = %q, want %q", got, want)
	}
	if strings.Contains(got, "[^f-") {
		t.Error("a raw finding-marker survived the strip — it would leak into the shipped report")
	}
	// A [^food]-style label (begins f, not f-) is a CLAIM label and is NOT stripped.
	if got := StripFindingMarkers("Tasty.[^food]"); got != "Tasty.[^food]" {
		t.Errorf("[^food] wrongly stripped: %q", got)
	}
}
