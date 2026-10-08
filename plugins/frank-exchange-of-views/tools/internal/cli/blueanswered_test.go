package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THROUGH THE CALL A SEAT MAKES (#1122).
//
// blueAnswers has its own table in internal/record, and a table is not a call site: #1162 shipped a
// projection field whose own test drove the oracle's entry point while the seat's renderer passed a
// nil, so the feature reached no seat and every gate was green. This drives `show work` as the
// MINTING LENS and asserts the item is in the bytes that seat is handed.
func TestTheLensWorkListSaysWhetherBlueAnswered(t *testing.T) {
	items := func(t *testing.T, runDir string) []string {
		t.Helper()
		out, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
		if err != nil {
			t.Fatal(err)
		}
		var w struct {
			Sitting record.SittingJSON `json:"sitting"`
		}
		if err := json.Unmarshal([]byte(out), &w); err != nil {
			t.Fatalf("the lens's work list did not parse: %v", err)
		}
		var got []string
		for _, it := range w.Sitting.Open {
			got = append(got, it.What)
		}
		return got
	}
	has := func(t *testing.T, got []string, want string) {
		t.Helper()
		for _, g := range got {
			if strings.Contains(g, want) {
				return
			}
		}
		t.Errorf("no item says %q; the lens's list was:\n  %s", want, strings.Join(got, "\n  "))
	}

	// BLUE ENGAGED AND ANSWERED. The chair dispatches blue onto the gap, blue sits and edits it.
	t.Run("answered", func(t *testing.T) {
		runDir, id := answeredBoard(t)
		dispatchBlueOnto(t, runDir, id)
		editAnswering(t, runDir, id)
		has(t, items(t, runDir), "has been ANSWERED")
	})

	// BLUE ENGAGED, ITS SITTING CLOSED, NO EDIT. The lens is told so rather than left to infer it
	// from an absence.
	t.Run("engaged and no answer", func(t *testing.T) {
		runDir, id := answeredBoard(t)
		dispatchBlueOnto(t, runDir, id)
		registerBlue(t, runDir)
		// Blue's sitting is closed by its agent's return, which the SubagentStop hook writes. Without
		// it the sitting is unresolved and the NOT MEASURED arm is the correct answer instead.
		closeBlueSitting(t, runDir)
		got := items(t, runDir)
		has(t, got, "NO edit answers it")
		for _, g := range got {
			if strings.Contains(g, "NOT MEASURED") {
				t.Errorf("blue's sitting is closed on the record, so nothing here is unmeasured: %q", g)
			}
		}
	})

	// BLUE IS STILL SITTING. "has not answered" and "may yet answer" are the same silence, and this
	// is the arm that keeps them apart.
	t.Run("blue is still sitting", func(t *testing.T) {
		runDir, id := answeredBoard(t)
		dispatchBlueOnto(t, runDir, id)
		registerBlue(t, runDir)
		got := items(t, runDir)
		has(t, got, "NOT MEASURED")
		for _, g := range got {
			if strings.Contains(g, "NO edit answers it") {
				t.Errorf("blue's sitting is open, so this is not a report that no answer came: %q", g)
			}
		}
	})

	// NEVER DISPATCHED ONTO. Silence: the gap waits on the chair, which is not news about blue.
	t.Run("blue never engaged", func(t *testing.T) {
		runDir, _ := answeredBoard(t)
		for _, g := range items(t, runDir) {
			for _, no := range []string{"has been ANSWERED", "NO edit answers it", "NOT MEASURED — a blue sitting"} {
				if strings.Contains(g, no) {
					t.Errorf("a gap blue was never dispatched onto produced %q", g)
				}
			}
		}
	})
}

// A GAP WHOSE SENTENCE BLUE CUT IS NOT UNANSWERED. Blue engaged on the gap and edited its sentence
// down to the bare anchor without naming the gap: the gap reads `gone`, and its lens is put to judge
// the report as it stands — never told that no edit answers it.
func TestBlueAnsweredReadsAGoneGapAsMoved(t *testing.T) {
	runDir, id := answeredBoard(t)
	dispatchBlueOnto(t, runDir, id)
	registerBlue(t, runDir)
	tok := "<!--gap:" + id + "-->"
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "CUT",
		"--quote", "The cost is high"+tok+" and rising.", "--new", tok,
		"--reason", "the sentence cannot stand"); err != nil {
		t.Fatalf("blue cutting the sentence: %v\n%s", err, out)
	}
	closeBlueSitting(t, runDir)
	out, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatal(err)
	}
	var w record.WorkJSON
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatal(err)
	}
	if len(w.Open) != 1 || w.Open[0].LocationState != record.LocationGone || len(w.Open[0].EditedSince) != 0 {
		t.Fatalf("the cut gap should read gone with no edit listed: %+v", w.Open)
	}
	var said bool
	for _, it := range w.Sitting.Open {
		if strings.Contains(it.What, "NO edit answers it") {
			t.Errorf("a gap whose sentence blue cut was reported as unanswered: %q", it.What)
		}
		said = said || strings.Contains(it.What, "gap "+id+" is open and its anchor is no longer in the report — judge whether the report as it stands")
	}
	if !said {
		t.Errorf("no item puts the gone gap to its lens to judge:\n%s", out)
	}
}

// dispatchBlueOnto has the chair record the dispatch that engages blue on the open gap. THE CHAIR'S
// OWN VERB, not a seeded row: which gaps a dispatch names is the chair's decision and the engagement
// this reads is exactly what that verb writes.
func dispatchBlueOnto(t *testing.T, runDir, gapID string) {
	t.Helper()
	out, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	if !strings.Contains(out, gapID) {
		t.Fatalf("the dispatch does not engage %s, so nothing below is about an engaged gap:\n%s", gapID, out)
	}
}

// editAnswering is blue sitting and answering the gap: old != new, which is what makes it movement
// rather than an act.
func editAnswering(t *testing.T, runDir, gapID string) {
	t.Helper()
	registerBlue(t, runDir)
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--answers", gapID,
		"--quote", "The cost is high<!--gap:"+gapID+"-->", "--new", "The cost is moderate<!--gap:"+gapID+"-->",
		"--reason", "the source gives a moderate figure"); err != nil {
		t.Fatalf("blue edit: %v\n%s", err, out)
	}
}

// closeBlueSitting closes blue's sitting the way the record allows a test to: blue's NEXT register.
// Named for what it DOES, not for the hook it stands in for.
// sittingCloser ends a sitting at whichever comes first, the opening of the seat's next sitting or
// its agent's stop, and here that opening is blue's own register — the stop is the SubagentStop
// hook's, and seeding one with a matching agent id would be asserting the hook's join rather than
// this projection's.
func closeBlueSitting(t *testing.T, runDir string) {
	t.Helper()
	registerBlue(t, runDir)
}

// answeredBoard is the smallest board these arms need: the cast the dispatch verb refuses without,
// a report with a sentence blue can edit, the chair and the minting lens registered, and one open
// gap. It mirrors stageEvents and differs in the one thing that matters here — the report has prose
// in it, because `blue edit` matches --quote against the report and a heading is refused.
func answeredBoard(t *testing.T) (runDir, gapID string) {
	t.Helper()
	runDir = newRun(t)
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:answered:%d", seat, n), body)
	}
	recordtest.Seed(t, runDir,
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# Findings\n\nThe cost is high and rising.\n")}),
		at("red-chair", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Register{}),
		at("red-lens-evidence", &recordpb.Mint{
			GapId: proto.String("G1"), Class: proto.String("c"),
			Problem:         proto.String("the figure rests on one source"),
			Location:        proto.String("The cost is high"),
			RequiredFix:     proto.String("qualify it"),
			AcceptanceCheck: proto.String("the sentence names the source's range"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:        recordtest.P(recordpb.Grade_GRADE_HIGH),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		at("red-lens-evidence", &recordpb.Anchor{Id: proto.String("G1"), Location: proto.String("The cost is high")}),
	)
	return runDir, "G1"
}
