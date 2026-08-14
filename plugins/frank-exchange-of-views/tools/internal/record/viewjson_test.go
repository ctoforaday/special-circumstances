package record

import (
	"encoding/json"
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
	// The markdown debate render is cross-checked against these same events in the view
	// package's tests (TestMarkdownDebateChangelogInquiryAndCitations); the render moved
	// out of this package, so the JSON view is pinned here and the markdown there.
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

// TestWorklistIsOpenOnlyLeanAndClosedIndexHasNoProse pins the merge's shrinking working set:
// OPEN gaps only in a lean shape (grades + class + location + a TRUNCATED problem synopsis +
// found_by, but NOT required_fix/acceptance_check), and closed gaps collapsed to a prose-free
// {id, location, class} index. This is the once-per-turn read the full board is not.
func TestWorklistIsOpenOnlyLeanAndClosedIndexHasNoProse(t *testing.T) {
	runDir := t.TempDir()
	m := "red-merge-r1"
	longProblem := strings.Repeat("word ", 60) // ~300 chars, well over the 140-rune synopsis budget
	writeShard(t, runDir, m, "aaaaaaaa", []Event{
		ev(m, "aaaaaaaa", 0, 1, "mint", m+":mint:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("problem", longProblem).Set("location", "§open").
			Set("class", "correctness").Set("required_fix", "SECRET_FIX_PROSE").
			Set("acceptance_check", "SECRET_CHECK_PROSE").Set("severity", "high").
			Set("found_by", []string{"L1-F1"})),
		ev(m, "aaaaaaaa", 1, 1, "mint", m+":mint:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("problem", "a closed problem").Set("location", "§closed").
			Set("class", "citation").Set("required_fix", "fix").Set("acceptance_check", "chk")),
		ev(m, "aaaaaaaa", 2, 1, "close", m+":close:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("class", "resolved")),
	})

	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	w := WorklistJSONOf(b)

	if len(w.Open) != 1 || w.Open[0].ID != "R1-1" {
		t.Fatalf("worklist Open = %+v, want the single open gap R1-1", w.Open)
	}
	if len(w.ClosedIndex) != 1 || w.ClosedIndex[0].ID != "R1-2" {
		t.Fatalf("worklist ClosedIndex = %+v, want the single closed gap R1-2", w.ClosedIndex)
	}
	if w.Counts.Open != 1 || w.Counts.Closed != 1 {
		t.Errorf("counts = %+v, want open 1 closed 1", w.Counts)
	}
	// The open gap's problem is TRUNCATED to the synopsis budget, and the full-prose fields
	// (required_fix, acceptance_check) are absent from the JSON entirely.
	if r := []rune(w.Open[0].ProblemSynopsis); len(r) > synopsisLimit+1 { // +1 for the ellipsis
		t.Errorf("problem synopsis is %d runes, want <= %d + ellipsis", len(r), synopsisLimit)
	}
	if !strings.HasSuffix(w.Open[0].ProblemSynopsis, "…") {
		t.Errorf("a truncated synopsis must be marked with an ellipsis: %q", w.Open[0].ProblemSynopsis)
	}
	blob, _ := json.Marshal(w)
	for _, secret := range []string{"SECRET_FIX_PROSE", "SECRET_CHECK_PROSE", "a closed problem"} {
		if strings.Contains(string(blob), secret) {
			t.Errorf("the worklist must not carry prose %q — it is lean by design:\n%s", secret, blob)
		}
	}
	// The lean open gap keeps its grades, class, location and found_by.
	if w.Open[0].Severity != "high" || w.Open[0].Class != "correctness" || w.Open[0].Location != "§open" {
		t.Errorf("open gap lost a lean field: %+v", w.Open[0])
	}
	if len(w.Open[0].FoundBy) != 1 || w.Open[0].FoundBy[0] != "L1-F1" {
		t.Errorf("open gap lost found_by: %+v", w.Open[0].FoundBy)
	}
	// The closed index carries id/location/class only — no problem prose.
	if w.ClosedIndex[0].Location != "§closed" || w.ClosedIndex[0].Class != "citation" {
		t.Errorf("closed index lost id/location/class: %+v", w.ClosedIndex[0])
	}
}

// TestBoardJSONFlattensMintWithoutDuplicating pins the board de-duplication: a gap's lineage
// (found_by, supersedes) used to live ONLY inside a nested `mint`
// object that ALSO re-stated every top-level field, so each gap carried its prose twice. They are
// now first-class and the nested object is gone — one copy, and nothing a reader needs is buried.
func TestBoardJSONFlattensMintWithoutDuplicating(t *testing.T) {
	runDir := t.TempDir()
	m := "red-merge-r1"
	writeShard(t, runDir, m, "aaaaaaaa", []Event{
		ev(m, "aaaaaaaa", 0, 1, "mint", m+":mint:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("problem", "an open problem").Set("location", "§1").
			Set("required_fix", "do the thing").Set("acceptance_check", "run the check").
			Set("severity", "high").Set("existence", "verified").
			Set("found_by", []string{"L1-F1", "L5-F3"}).Set("supersedes", []string{"R0-9"})),
	})
	b, err := BoardJSONBytes(runDir)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	// The lineage + leaf-check fields are promoted to the top level (their only home now).
	// Tokens, not a compact substring — BoardJSONBytes pretty-prints.
	for _, want := range []string{`"found_by"`, `"L1-F1"`, `"L5-F3"`, `"supersedes"`, `"R0-9"`} {
		if !strings.Contains(s, want) {
			t.Errorf("board gap is missing top-level %s:\n%s", want, s)
		}
	}
	// The redundant nested `mint` object is gone, so the prose appears exactly ONCE.
	if strings.Contains(s, `"mint":`) {
		t.Error("the board still carries the redundant nested `mint` object")
	}
	if n := strings.Count(s, "an open problem"); n != 1 {
		t.Errorf("the problem prose appears %d times, want exactly 1 (the duplication this removed)", n)
	}
}

// UNCREDITED FINDINGS, not undisposed observations (#327).
//
// The old metric counted `observe` events with no `dispose` fate. Both verbs are retired, so it
// would now be PERMANENTLY ZERO — a dead detector and a clean board printing the same number,
// which is the exact shape this codebase keeps finding. A finding is addressed by coalescence,
// so the honest question is whether it was ever credited in a gap's found_by.
func TestUncreditedFindingsCountsFindingsNoGapCredits(t *testing.T) {
	runDir := t.TempDir()
	s := "red-lens-r1-L1"
	m := "red-merge-r1"
	writeShard(t, runDir, s, "aaaaaaaa", []Event{
		ev(s, "aaaaaaaa", 0, 1, "finding", s+":finding:L1-F1", NewPayload().Set("label", "L1-F1").Set("text", "credited")),
		ev(s, "aaaaaaaa", 1, 1, "finding", s+":finding:L1-F2", NewPayload().Set("label", "L1-F2").Set("text", "never credited")),
	})
	writeShard(t, runDir, m, "bbbbbbbb", []Event{
		ev(m, "bbbbbbbb", 0, 1, "mint", m+":mint:k", NewPayload().Set("gap_id", "R1-1").Set("found_by", []string{"L1-F1"})),
	})
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	bj := BoardJSONOf(b)
	if bj.Counts.UncreditedFindings != 1 {
		t.Errorf("exactly one finding is credited by no gap; got %d uncredited", bj.Counts.UncreditedFindings)
	}
	for _, o := range bj.Observations {
		want := o.Label == "L1-F1"
		if o.Credited != want {
			t.Errorf("%s: credited=%v, want %v", o.Label, o.Credited, want)
		}
	}
}
