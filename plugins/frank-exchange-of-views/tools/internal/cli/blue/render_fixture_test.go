package blue

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"os"
	"regexp"
	"testing"
)

// The fixture is a report-shaped document (title, TL;DR, sections, footnotes) whose prose is De
// Bello Gallico — obviously test data — carrying real-shaped immortal anchors: finding-markers
// (<!--fx:f-…-->) and citation anchors (<!--cite:c-…-->), several sitting mid-sentence and flush
// against a word, which is the common case (red anchors exactly the sentence blue is asked to
// repair). It is the standing guard that render/edit reproduces a realistic report byte-for-byte
// AND never loses a required reference — and a seed for edit fuzzing.
func loadFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/fixture_report.md")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return string(b)
}

var immortalAnchor = regexp.MustCompile(`<!--(?:fx:f|cite:c|proof:p)-[0-9a-f]+-->`)

// TestFixtureRenderReproducesEditPathAndKeepsEveryAnchor is the committed fidelity guard: a
// realistic sequence of edits (including one whose span sits flush before a finding-marker)
// reproduces byte-for-byte through Render, VerifyReproduction accepts it, and every immortal
// anchor present in the base is still present after the replay.
func TestFixtureRenderReproducesEditPathAndKeepsEveryAnchor(t *testing.T) {
	base := loadFixture(t)
	anchorsBefore := immortalAnchor.FindAllString(base, -1)
	if len(anchorsBefore) < 4 {
		t.Fatalf("fixture should carry several immortal anchors, found %d", len(anchorsBefore))
	}

	steps := []reportproj.Op{
		// plain edit, no anchor in span
		{Old: "Horum omnium fortissimi sunt Belgae", New: "Horum omnium fortissimi longe sunt Belgae"},
		// anchor-ADJACENT: a fragment that stops before the finding-marker without reaching it —
		// the words change, the anchor stays where it was placed.
		{Old: "legibus inter se", New: "legibus atque moribus inter se"},
		// anchor-CARRYING: the old span contains a citation anchor and --new carries it unchanged,
		// so it transits the edit — the case droppedMarker + AnchorsTransitUnchanged govern.
		{Old: "longissime absunt<!--cite:c-0badf00d-->", New: "longissime procul absunt<!--cite:c-0badf00d-->"},
	}

	viaEdit := base
	for i, s := range steps {
		next, err := planEdit(viaEdit, s.Old, s.New)
		if err != nil {
			t.Fatalf("planEdit step %d (%q): %v", i, s.Old, err)
		}
		viaEdit = next
	}

	rendered, err := reportproj.Render(base, steps)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if rendered != viaEdit {
		t.Fatalf("render drifted from the edit path on the fixture (edit %d bytes, render %d)", len(viaEdit), len(rendered))
	}
	if err := VerifyReproduction(viaEdit, rendered); err != nil {
		t.Fatalf("VerifyReproduction rejected an identical round-trip: %v", err)
	}

	// EVERY required reference survives. droppedMarker guards this per-edit; this asserts the
	// end-to-end replay result carries all of them.
	for _, a := range anchorsBefore {
		if !regexp.MustCompile(regexp.QuoteMeta(a)).MatchString(rendered) {
			t.Errorf("render LOST a required reference %q — the immortal-anchor guarantee is broken", a)
		}
	}
}
