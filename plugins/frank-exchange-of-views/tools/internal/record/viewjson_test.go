package record

import (
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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

	writeShard(t, runDir, []*Event{
		recordtest.At(t, merge, 1, merge+":position", &recordpb.Position{Text: proto.String("red r1")}),
		// The gap must exist before anything speaks about it: `closing.gap_id` and the bench's
		// opinion are both foreign keys onto the mint.
		recordtest.At(t, merge, 1, merge+":mint:R1-1", &recordpb.Mint{
			GapId: proto.String("R1-1"), Class: proto.String("overclaim"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
		recordtest.At(t, merge, 1, merge+":closing:R1-1", &recordpb.Closing{GapId: proto.String("R1-1"), Text: proto.String("red closes r1")}),
	})
	writeShard(t, runDir, []*Event{
		recordtest.At(t, blue, 1, blue+":position", &recordpb.Position{Text: proto.String("blue r1")}),
		// The `confidence` event type went with the per-claim confidence grades; blue's round-1
		// section is carried by its position, which is what this test reads.
	})
	writeShard(t, runDir, []*Event{
		recordtest.At(t, judge, 1, judge+":opinion:R1-1", &recordpb.Opinion{GapId: proto.String("R1-1"), Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CLOSED), Principle: proto.String("correctness first")}),
	})
	// Round 2: red positions again, blue does not (a red-only round — its Red is non-empty,
	// its Blue is the empty array a consumer counts as zero, never a null).
	writeShard(t, runDir, []*Event{
		recordtest.At(t, merge2, 2, merge2+":position", &recordpb.Position{Text: proto.String("red r2")}),
	})
	// A blue seat that recorded nothing in round 2 is simply absent from the record — there is no
	// empty shard file to write, which is what the call here used to produce.

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
	if len(r1.Lead) != 1 || r1.Lead[0].Disposition != "closed" || r1.Lead[0].Principle != "correctness first" {
		t.Errorf("round 1 Lead = %+v, want one closed opinion", r1.Lead)
	}
	if len(r1.RedClosings) != 1 || r1.RedClosings[0].GapID != "R1-1" || r1.RedClosings[0].Text != "red closes r1" {
		t.Errorf("round 1 RedClosings = %+v", r1.RedClosings)
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
	writeShard(t, runDir, []*Event{
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:position", &recordpb.Position{Text: proto.String("red")}),
	})
	out, err := DebateJSONBytes(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"rounds"`) || !strings.Contains(string(out), `"red"`) {
		t.Errorf("DebateJSONBytes output missing expected keys:\n%s", out)
	}
}

// TestWorkIsOpenOnlyLeanAndClosedIndexHasNoProse pins the merge's shrinking working set:
// OPEN gaps only in a lean shape (grades + class + location + a TRUNCATED problem synopsis +
// found_by, but NOT required_fix/acceptance_check), and closed gaps collapsed to a prose-free
// {id, location, class} index. This is the once-per-turn read the full board is not.
func TestWorkIsOpenOnlyLeanAndClosedIndexHasNoProse(t *testing.T) {
	runDir := t.TempDir()
	m := "red-merge-r1"
	longProblem := strings.Repeat("word ", 60) // ~300 chars, well over the 140-rune synopsis budget
	writeShard(t, runDir, []*Event{
		recordtest.At(t, m, 1, m+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String(longProblem), Location: proto.String("§open"), RequiredFix: proto.String("SECRET_FIX_PROSE"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, m, 1, m+":mint:R1-2", &recordpb.Mint{GapId: proto.String("R1-2"), Class: proto.String("overclaim"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("a closed problem"), Location: proto.String("§closed"), RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("chk")}),
		recordtest.At(t, m, 1, m+":close:R1-2", &recordpb.Close{
			GapId: proto.String("R1-2"),
			// `class: "resolved"` before the schema — a key `merge close` never wrote (it writes
			// closure_class) carrying a word the enum never had. Untyped, it replayed as a closure
			// with NO class: a torn closure, which is the state this test is not about.
			ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_CLOSED),
		}),
	})

	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	w := WorkJSONOf(b)

	if len(w.Open) != 1 || w.Open[0].ID != "R1-1" {
		t.Fatalf("work list Open = %+v, want the single open gap R1-1", w.Open)
	}
	if len(w.ClosedIndex) != 1 || w.ClosedIndex[0].ID != "R1-2" {
		t.Fatalf("work list ClosedIndex = %+v, want the single closed gap R1-2", w.ClosedIndex)
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
			t.Errorf("the work list must not carry prose %q — it is lean by design:\n%s", secret, blob)
		}
	}
	// The lean open gap keeps its grades, class, location and found_by.
	if w.Open[0].Severity != recordpb.Grade_GRADE_HIGH || w.Open[0].Class != "correctness" || w.Open[0].Location != "§open" {
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
	writeShard(t, runDir, []*Event{
		recordtest.At(t, m, 1, m+":mint:R1-1", &recordpb.Mint{Problem: proto.String("an open problem"), Location: proto.String("§1"), AcceptanceCheck: proto.String("run the check"), Supersedes: []string{"R0-9"}}),
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
	writeShard(t, runDir, []*Event{
		recordtest.At(t, s, 1, s+":finding:L1-F1", &recordpb.Finding{Label: proto.String("L1-F1"), Text: proto.String("credited")}),
		recordtest.At(t, s, 1, s+":finding:L1-F2", &recordpb.Finding{Label: proto.String("L1-F2"), Text: proto.String("never credited")}),
	})
	writeShard(t, runDir, []*Event{
		recordtest.At(t, m, 1, m+":mint:k", &recordpb.Mint{GapId: proto.String("k"), FoundBy: []string{"L1-F1"}}),
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
