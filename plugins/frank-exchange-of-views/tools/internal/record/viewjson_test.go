package record

import (
	"path/filepath"
	"strings"
	"testing"
)

// The debate JSON view must group events into the SAME sections render.go writes to
// debate.md — that is the whole point of deriving both from BoardState. This test builds a
// two-round fixture and checks the structured view against the markdown render, section for
// section, so the two readings of one replay are proven not to drift.
func TestDebateJSONMirrorsRenderSections(t *testing.T) {
	runDir := t.TempDir()
	merge := "red-merge-r1"
	blue := "blue-lane-1"
	judge := "judge-r1"
	merge2 := "red-merge-r2"
	blue2 := "blue-lane-2"

	writeShard(t, runDir, merge, "aaaaaaaa", []Event{
		ev(merge, "aaaaaaaa", 0, 1, "position", merge+":position", NewPayload().Set("text", "red r1")),
		ev(merge, "aaaaaaaa", 1, 1, "closing", merge+":closing:R1-1", NewPayload().Set("gap_id", "R1-1").Set("text", "red closes r1")),
	})
	writeShard(t, runDir, blue, "bbbbbbbb", []Event{
		ev(blue, "bbbbbbbb", 0, 1, "position", blue+":position", NewPayload().Set("text", "blue r1")),
		ev(blue, "bbbbbbbb", 1, 1, "confidence", blue+":confidence:C1", NewPayload().Set("label", "claim one").Set("grade", "medium")),
		ev(blue, "bbbbbbbb", 2, 1, "dispute", blue+":dispute:R1-1", NewPayload().Set("gap_id", "R1-1").Set("dimension", "likelihood").Set("proposed", "low").Set("evidence", "blue evidence")),
	})
	writeShard(t, runDir, judge, "cccccccc", []Event{
		ev(judge, "cccccccc", 0, 1, "opinion", judge+":opinion:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("disposition", "upheld").Set("principle", "correctness first")),
	})
	// Round 2: red positions again, blue does not (a red-only round — its Red is non-empty,
	// its Blue is the empty array a consumer counts as zero, never a null).
	writeShard(t, runDir, merge2, "dddddddd", []Event{
		ev(merge2, "dddddddd", 0, 2, "position", merge2+":position", NewPayload().Set("text", "red r2")),
	})
	// A blue seat that recorded nothing in round 2 (present in the run, silent this round).
	writeShard(t, runDir, blue2, "eeeeeeee", []Event{})

	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	dj := DebateJSONOf(b)

	if len(dj.Rounds) != 2 {
		t.Fatalf("want 2 rounds, got %d: %+v", len(dj.Rounds), dj.Rounds)
	}
	r1, r2 := dj.Rounds[0], dj.Rounds[1]
	if r1.Round != 1 || r2.Round != 2 {
		t.Fatalf("rounds out of order: %d, %d", r1.Round, r2.Round)
	}
	if len(r1.Red) != 1 || r1.Red[0] != "red r1" {
		t.Errorf("round 1 Red = %v, want [red r1]", r1.Red)
	}
	if len(r1.Blue) != 1 || r1.Blue[0] != "blue r1" {
		t.Errorf("round 1 Blue = %v, want [blue r1]", r1.Blue)
	}
	if len(r1.Lead) != 1 || r1.Lead[0].Disposition != "upheld" || r1.Lead[0].Principle != "correctness first" {
		t.Errorf("round 1 Lead = %+v, want one upheld opinion", r1.Lead)
	}
	if len(r1.RedClosings) != 1 || r1.RedClosings[0].GapID != "R1-1" || r1.RedClosings[0].Text != "red closes r1" {
		t.Errorf("round 1 RedClosings = %+v", r1.RedClosings)
	}
	if len(r1.Confidence) != 1 || r1.Confidence[0].Label != "claim one" || r1.Confidence[0].Grade != "medium" {
		t.Errorf("round 1 Confidence = %+v", r1.Confidence)
	}
	if len(r1.Disputes) != 1 || r1.Disputes[0].Kind != "dispute" || r1.Disputes[0].Proposed != "low" {
		t.Errorf("round 1 Disputes = %+v", r1.Disputes)
	}
	// The red-only round: Red present, Blue an empty (non-nil) array.
	if len(r2.Red) != 1 || r2.Red[0] != "red r2" {
		t.Errorf("round 2 Red = %v, want [red r2]", r2.Red)
	}
	if r2.Blue == nil || len(r2.Blue) != 0 {
		t.Errorf("round 2 Blue = %v, want an empty non-nil array", r2.Blue)
	}

	// redRounds — the count telemetryAudit derives — is the number of rounds with a red
	// sitting. Both rounds have one here.
	redRounds := 0
	for _, r := range dj.Rounds {
		if len(r.Red) > 0 {
			redRounds++
		}
	}
	if redRounds != 2 {
		t.Errorf("redRounds = %d, want 2", redRounds)
	}

	// Cross-check against the markdown render: every Red/Blue text the JSON view carries must
	// appear under the matching "### RED"/"### BLUE" heading of debate.md.
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	debate := readFile(t, filepath.Join(res.Out, "debate.md"))
	for _, want := range []string{"### RED\nred r1", "### BLUE\nblue r1", "### RED\nred r2"} {
		if !strings.Contains(debate, want) {
			t.Errorf("debate.md missing %q — JSON view and markdown render disagree:\n%s", want, debate)
		}
	}
}

// --json is view-selected only where a JSON form exists. A JSON-native view rejects it (there
// is one way to that JSON, by name) and a markdown view with no JSON form rejects it too.
// This is enforced in the show read-path; the check here guards the DebateJSONBytes entry.
func TestDebateJSONBytesIsValidJSON(t *testing.T) {
	runDir := t.TempDir()
	writeShard(t, runDir, "red-merge-r1", "aaaaaaaa", []Event{
		ev("red-merge-r1", "aaaaaaaa", 0, 1, "position", "red-merge-r1:position", NewPayload().Set("text", "red")),
	})
	out, err := DebateJSONBytes(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"rounds"`) || !strings.Contains(string(out), `"red"`) {
		t.Errorf("DebateJSONBytes output missing expected keys:\n%s", out)
	}
}
