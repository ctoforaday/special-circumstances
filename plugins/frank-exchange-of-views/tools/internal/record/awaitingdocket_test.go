package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE VIEW SAYS WHICH OPEN GAPS THE BENCH HAS CARRIED, AND IT SAYS SO WITHOUT AN ORDER (#759).
//
// `carried` answers its motion — the gap returns only by being docketed again — so an open gap
// with a carry behind it is a different situation from one nobody has ever put before the bench,
// and the merge seat used to get the same sentence for both.
//
// FIVE GAPS, AND FOUR OF THEM ARE THE NEGATIVE CASES. A flag that is simply true for every open
// gap would pass any test that only checks the positive one, which is the shape #759 warned this
// measurement would take.
func TestAwaitingDocketIsSetOnlyByALiveCarry(t *testing.T) {
	run := mustRun(t, newRun(t))
	red := Identity{Run: run, SeatID: "red-merge-r1", Round: 1}
	judge := Identity{Run: run, SeatID: "judge-r1", Round: 1}
	app := func(id Identity, body proto.Message) {
		t.Helper()
		if _, err := Append(id, body); err != nil {
			t.Fatal(err)
		}
	}
	mint := func(id string) {
		app(red, &recordpb.Mint{GapId: proto.String(id), Class: proto.String("self-attestation"),
			Problem: proto.String("p " + id), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
			CheckKind:  recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
	}
	file := func(motionID, gapID string) {
		app(red, &recordpb.Motion{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:   proto.String("red cannot settle " + gapID),
			Filing:  &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}}})
	}
	carry := func(motionID, reopensOn string) {
		app(judge, &recordpb.MotionRule{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String("not this round"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED),
				Principle:   proto.String("thoroughness over speed"),
				Tension:     proto.String("cost against certainty"),
				ReviewFlag:  proto.String("none"),
				Settled:     proto.String("nothing yet — the gap survives"),
				ReopensOn:   proto.String(reopensOn)}}})
	}
	dispose := func(motionID string) {
		app(judge, &recordpb.MotionRule{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String("settled"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_NOT_A_DEFECT),
				Principle:   proto.String("correctness first"),
				Tension:     proto.String("cost against certainty"),
				ReviewFlag:  proto.String("none"),
				Settled:     proto.String("the claim as it stood may not be re-asserted"),
				Final:       proto.Bool(true)}}})
	}

	for _, id := range []string{"CARRIED", "DISPOSED", "REPENDING", "FRESH", "ORDER"} {
		mint(id)
	}
	// CARRIED: heard once, deferred, nothing pending. The one true case.
	file("M1", "CARRIED")
	carry("M1", "blue reporting what the stated direction found")
	// DISPOSED: carried in one sitting and settled in the next. The gap is closed, so it is not
	// awaiting anything — and this is the arm that would break if the flag ignored openness.
	file("M2", "DISPOSED")
	carry("M2", "the first condition")
	file("M3", "DISPOSED")
	dispose("M3")
	// REPENDING: carried, then docketed again and not yet answered. Already before the bench, so
	// asking the seat to file it would ask the same question twice.
	file("M4", "REPENDING")
	carry("M4", "the first condition")
	file("M5", "REPENDING")
	// FRESH: nobody has docketed it. Open, but not carried.
	//
	// ORDER: two carries, and the motion ids are chosen so that the LEXICOGRAPHIC latest is the
	// EARLIER event — "M10" sorts before "M9". A reader taking "the last ruling" off the motion
	// id reports the first condition; the record's own order reports the second.
	file("M9", "ORDER")
	carry("M9", "the FIRST condition, which was answered and superseded")
	file("M10", "ORDER")
	carry("M10", "the SECOND condition, which is the live one")

	evs, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	gaps, err := workGapStatesOfRun(run, evs.Events)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]WorkGapState{}
	for _, g := range gaps {
		by[g.ID] = g
	}
	for id, want := range map[string]bool{
		"CARRIED": true, "ORDER": true,
		"DISPOSED": false, "REPENDING": false, "FRESH": false,
	} {
		g, ok := by[id]
		if !ok {
			t.Fatalf("%s is not on the board at all", id)
		}
		if g.AwaitingDocket != want {
			t.Errorf("%s: awaiting_docket = %v, want %v (open=%v, reopens_on=%q)",
				id, g.AwaitingDocket, want, g.Open, g.DocketReopensOn)
		}
	}
	// AND THE CONDITION IS THE LIVE ONE. This is the half an order-free predicate cannot answer,
	// and the half a motion-id ordering gets wrong: `motion_rule.event_id` is the events primary
	// key, so the latest carry is the one the bench actually made last.
	if got := by["ORDER"].DocketReopensOn; !strings.Contains(got, "SECOND") {
		t.Errorf("ORDER: reopens_on = %q, want the SECOND condition — a reader ordering on the motion id "+
			"picks M9 over M10 and hands the seat a condition that was already answered", got)
	}
	if got := by["CARRIED"].DocketReopensOn; got != "blue reporting what the stated direction found" {
		t.Errorf("CARRIED: reopens_on = %q, want the bench's own words", got)
	}
	// A gap nobody carried must carry no condition — "" is the honest empty, not a missing read.
	if got := by["FRESH"].DocketReopensOn; got != "" {
		t.Errorf("FRESH was never carried and reports a reopens-on condition: %q", got)
	}
}
