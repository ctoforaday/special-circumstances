package runlive

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// THE MARKER'S SHAPE IS A CROSS-MODULE CONTRACT, AND THIS IS THE WRITER'S END OF IT.
//
// .claude/run-live.json is written here and read by prosthetic-conscience's guards, which
// live in a different module and therefore cannot import this type — they restate it, and a
// restatement is a copy that can go stale. It did. #529 made this file a list; this
// package's own doc records that the change broke a private decoder inside THIS module, and
// the decoder one module out was not migrated at all, so `json.Unmarshal` handed the guards
// a marker with every field zero. They kept warning, with nothing in the warning: "a
// research run is LIVE (, started )". A guard that still speaks reads as a guard that still
// works.
//
// So the shape is pinned as bytes, in a file both modules hold a copy of:
//
//	this test          — the writer PRODUCES exactly testdata/run-live.golden.json
//	pc's runlive test  — its reader UNDERSTANDS the same bytes
//	scripts/check      — the two committed copies are byte-identical
//
// Change the shape and this test fails FIRST, before the readers are consulted at all. Take
// the new golden and the parity gate fails until the other module's copy is updated; update
// that and the reader's test fails until the reader is migrated. The chain ends where the
// original defect began, which is the only place it could have been caught.
func TestWriterProducesTheGoldenShape(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC)

	// Two runs, because the list shape is the whole point of #529 and a one-element list
	// hides the separator; the second omits runId/scriptPath so the `omitempty` fields are
	// exercised in both directions.
	WriteRunLiveMarker(dir, "research/2026-07-18_alpha", []string{"ideas/backlog.md", "research/old"}, t0, "run-alpha", "scripts/run.mjs")
	WriteRunLiveMarker(dir, "research/2026-07-18_beta", []string{"research/old", "ideas/other.md"}, t0.Add(90*time.Minute), "", "")

	got, err := os.ReadFile(filepath.Join(dir, ".claude", "run-live.json"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join("testdata", "run-live.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("the marker's bytes moved.\n\n--- got ---\n%s\n--- testdata/run-live.golden.json ---\n%s\n"+
			"If this change is intended, update BOTH copies of the golden — this one and\n"+
			"plugins/prosthetic-conscience/tools/internal/runlive/testdata/run-live.golden.json —\n"+
			"and migrate that module's reader. scripts/check gates the two copies, and its\n"+
			"reader test will fail until it understands the new shape. That failure chain is\n"+
			"deliberate: the last time this shape moved, nothing failed at all.", got, want)
	}
}

// The upsert is what makes the file a set of open runs rather than a log: a second write for
// the SAME run replaces its row instead of appending a duplicate. Without it, a run that
// re-announces itself would appear twice and a single capture would leave one behind — a
// marker that says a finished run is still live.
func TestWritingTheSameRunTwiceReplacesItsRow(t *testing.T) {
	dir := t.TempDir()
	t0 := time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC)
	WriteRunLiveMarker(dir, "research/x", []string{"a"}, t0, "", "")
	WriteRunLiveMarker(dir, "research/x", []string{"a", "b"}, t0.Add(time.Hour), "", "")

	runs := ReadRunLive(dir)
	if len(runs) != 1 {
		t.Fatalf("re-writing one run must not append a second row, got %d: %+v", len(runs), runs)
	}
	if len(runs[0].PinnedPaths) != 2 {
		t.Errorf("the row must carry the LATEST write, got %+v", runs[0])
	}
}
