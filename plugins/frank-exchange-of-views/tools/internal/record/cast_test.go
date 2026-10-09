package record

import (
	"encoding/json"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The cast is the record's, written once under harness; membership admits a seat and its
// petition sitting, and "no cast" is told apart from "not in it".
func TestTheCastIsReadOffTheRecordAndAdmitsItsMembers(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	if cast, err := CastOf(run); err != nil || cast != nil {
		t.Fatalf("an empty record has no cast: (%v, %v)", cast, err)
	}
	if member, has, err := InCast(run, "red-chair"); err != nil || member || has {
		t.Fatalf("no cast: (%v, %v, %v), want (false, false, nil)", member, has, err)
	}
	recordtest.Seed(t, runDir, recordtest.At(t, "harness", "harness:cast:#1", &recordpb.Cast{
		SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge"},
	}))
	cast, err := CastOf(run)
	if err != nil || len(cast) != 5 || cast[0] != "red-lens-evidence" || cast[4] != "judge" {
		t.Fatalf("CastOf = (%v, %v)", cast, err)
	}
	for seat, want := range map[string]bool{
		"red-chair":               true,
		"judge":                   true, // a petition sitting by a cast seat
		"red-lens-voice":          false,
		"judge-petition-frontier": false, // frontier is not in this cast
	} {
		member, has, err := InCast(run, seat)
		if err != nil || !has || member != want {
			t.Errorf("InCast(%q) = (%v, %v, %v), want member=%v with a cast present", seat, member, has, err, want)
		}
	}
}

// A LANE THE CAST DOES NOT SEAT IS REFUSED AT THE WRITE (#1153b).
//
// `lane_seat_ids` is what makes a lane's role and tier a FIELD rather than something recovered from
// the shape of its id. It is a SUBSET of the cast, and a lane named outside it would be dispatchable
// at `register` and absent from the run's own membership at the same moment — two gates reading one
// record and disagreeing. The write refuses it instead, which is what makes the field trustworthy
// to the tier join and the coverage audit that now read it.
func TestACastNamingALaneItDoesNotSeatIsRefused(t *testing.T) {
	dir := newRun(t)
	_, err := Append(Identity{Run: mustRun(t, dir), SeatID: HarnessSeat},
		castBody([]string{"red-chair", "blue-lane-1"}, []string{"blue-lane-9"}))
	if err == nil {
		t.Fatal("a cast naming a lane outside its own seats was accepted")
	}
	for _, want := range []string{"blue-lane-9", "SUBSET of the cast"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q:\n%v", want, err)
		}
	}
	// The honest case still lands: a lane the cast seats.
	seats, lanes := CastFor(nil, 2)
	if _, err := Append(Identity{Run: mustRun(t, newRun(t)), SeatID: HarnessSeat},
		castBody(seats, lanes)); err != nil {
		t.Errorf("a cast naming its own lanes was refused: %v", err)
	}
}

// A GAP'S BACKING RIDES THE GAP (#1153 follow-up).
//
// A lens auditing a claim asks "is this backed, and did anyone check?" — a property of the
// sentence in front of it, not of the run. Measured across three runs, `show evidence` was called
// 99 times against a projection holding ~10 events: a seat polling a whole table for one fact
// about one gap.
//
// THE SENTENCE IS THE ONE HOLDING THE GAP'S ANCHOR, and only evidence backs it: the gap's own
// anchor, another gap's and a finding's stand in that sentence and are not listed. An about gap
// whose about_ref is a gap lists nothing.
func TestAGapCarriesWhatBacksItAndWhetherAnyoneChecked(t *testing.T) {
	cited := anchor.Token("C-a1b2c3d4")
	unchecked := anchor.Token("C-99887766")
	was := reportRenderer
	t.Cleanup(func() { reportRenderer = was })
	reportRenderer = func(string, bool, []ReportOp) (string, error) {
		return "# R\n\nA sentence " + cited + " and another" + unchecked + anchor.Token("F-0b0b0b0b") + anchor.Token("G-00000001") + anchor.Token("G-00000009") + ". Next one.\n", nil
	}
	dir := newRun(t)
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "red-lens-evidence"}, &recordpb.Mint{
		GapId: proto.String("G-00000001"), Class: proto.String("c"),
		Problem:     proto.String("the claim rests on one source"),
		Location:    proto.String("A sentence " + cited + " and another " + unchecked),
		RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("chk"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH),
		Impact:     recordtest.P(recordpb.Grade_GRADE_HIGH),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "red-lens-evidence"}, &recordpb.Mint{
		GapId: proto.String("G-00000002"), Class: proto.String("c"), Problem: proto.String("about the first gap"),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String("G-00000001"),
		RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("chk"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH),
		Impact:     recordtest.P(recordpb.Grade_GRADE_HIGH),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "red-lens-evidence"}, &recordpb.Verify{
		Anchor: proto.String("C-a1b2c3d4"), Claim: proto.String("the claim"),
		Text:       proto.String("the source states it at the leaf"),
		Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
		Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
	}); err != nil {
		t.Fatal(err)
	}
	// THROUGH THE CALL A SEAT MAKES, not through the oracle's entry point. This test drove
	// WorkJSONOfRun and passed while the seat-facing renderer still assembled with a nil backing —
	// the feature reached the consistency oracle and no seat at all.
	b, err := WorkJSONBytes(mustRun(t, dir), "lens", "red-lens-evidence")
	if err != nil {
		t.Fatal(err)
	}
	var w WorkJSON
	if err := json.Unmarshal(b, &w); err != nil {
		t.Fatal(err)
	}
	if len(w.Open) != 2 || len(w.Open[1].Backing) != 0 {
		t.Fatalf("want the two open gaps, the about gap backed by nothing: %+v", w.Open)
	}
	by := map[string]GapBackingJSON{}
	for _, b := range w.Open[0].Backing {
		by[b.Anchor] = b
	}
	if len(by) != 2 {
		t.Fatalf("backing = %+v, want the two citation anchors in the gap's sentence and nothing else", w.Open[0].Backing)
	}
	if g := by["C-a1b2c3d4"]; g.Outcome == "" || g.Kind != "citation" || g.VerifiedBy != "red-lens-evidence" {
		t.Errorf("a verified anchor lost its verification: %+v", g)
	}
	// UNVERIFIED IS AN ENTRY. "nobody looked" and "no anchor here" must not be the same bytes.
	if g, ok := by["C-99887766"]; !ok || g.Outcome != "" {
		t.Errorf("an unchecked anchor is missing or claims an outcome: %+v", g)
	}
	// THE BOARD CARRIES THE SAME BACKING. It declared the field and never filled it, so every board
	// read showed `backing: null` beside a help page promising the anchors behind each open gap.
	board, err := BoardJSONOfRun(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(board.Open) != 2 || len(board.Open[0].Backing) != 2 {
		t.Fatalf("the board's open gap does not carry both anchors: %+v", board.Open)
	}
	for _, b := range board.Open[0].Backing {
		if b != by[b.Anchor] {
			t.Errorf("the board and the work list disagree on %s: board %+v, work %+v", b.Anchor, b, by[b.Anchor])
		}
	}
}

// OWNERSHIP IS ANSWERED, NOT DERIVED (gblock's ruling: work is self-sufficient for any job that is
// not reading the whole document).
//
// Every one of universe-m10's eight refusals was a seat acting on a gap another seat minted, and the
// information was on the list the whole time — as `found_by: ["computation-F1"]`, a finding label
// whose owning seat you recover by decoding a prefix and then recalling that only the originator may
// close. Two inferences, at the moment of acting, ten calls into a sitting.
//
// THROUGH THE SEAT PATH, because `yours_to_close` is a fact ABOUT THE READER and the oracle's
// entry point has no reader. A test through WorkJSONOfRun would assert the false case twice and
// pass — which is how #1162 shipped a field that reached no seat.
func TestTheWorkListAnswersWhetherAGapIsYoursToClose(t *testing.T) {
	dir := newRun(t)
	// Two DIFFERENT defects: the mint screens the open board, and one defect minted twice is refused.
	problem := map[string]string{
		"G-00000001": "the retention figure contradicts the stated universal",
		"G-00000002": "the survey sample is too small for the precision claimed",
	}
	mint := func(seat, gap string) {
		if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: seat}, &recordpb.Mint{
			GapId: proto.String(gap), Class: proto.String("c"),
			Problem:     proto.String(problem[gap]),
			Location:    proto.String("A sentence"),
			RequiredFix: proto.String("qualify it"), AcceptanceCheck: proto.String("the sentence names the range"),
			CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:   recordtest.P(recordpb.Grade_GRADE_HIGH),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Impact:     recordtest.P(recordpb.Grade_GRADE_HIGH),
		}); err != nil {
			t.Fatal(err)
		}
	}
	mint("red-lens-evidence", "G-00000001")
	mint("red-lens-logic", "G-00000002")

	for _, reader := range []string{"red-lens-evidence", "red-lens-logic"} {
		b, err := WorkJSONBytes(mustRun(t, dir), "lens", reader)
		if err != nil {
			t.Fatal(err)
		}
		var w WorkJSON
		if err := json.Unmarshal(b, &w); err != nil {
			t.Fatal(err)
		}
		if len(w.Open) != 2 {
			t.Fatalf("want both gaps on the list (a lens sees the open set): %+v", w.Open)
		}
		for _, g := range w.Open {
			mine := g.MintedBy == reader
			if g.YoursToClose != mine {
				t.Errorf("%s reading %s: yours_to_close=%v but minted_by=%q — the field must answer for THIS reader",
					reader, g.ID, g.YoursToClose, g.MintedBy)
			}
			// The prose that made a seat leave this list for the board.
			if g.AcceptanceCheck == "" || g.RequiredFix == "" || g.Problem == "" {
				t.Errorf("%s reading %s: the gap still withholds what it takes to act: check=%q fix=%q problem=%q",
					reader, g.ID, g.AcceptanceCheck, g.RequiredFix, g.Problem)
			}
		}
	}
}
