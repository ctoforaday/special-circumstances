package record

import (
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
func TestAGapCarriesWhatBacksItAndWhetherAnyoneChecked(t *testing.T) {
	cited := anchor.Token("c-a1b2c3d4")
	unchecked := anchor.Token("c-99887766")
	dir := newRun(t)
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "red-lens-evidence"}, &recordpb.Mint{
		GapId: proto.String("G1"), Class: proto.String("c"),
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
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "red-lens-evidence"}, &recordpb.Verify{
		Anchor: proto.String("c-a1b2c3d4"), Claim: proto.String("the claim"),
		Text:       proto.String("the source states it at the leaf"),
		Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
		Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
	}); err != nil {
		t.Fatal(err)
	}
	w, err := WorkJSONOfRun(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Open) != 1 {
		t.Fatalf("want the one open gap: %+v", w.Open)
	}
	by := map[string]GapBackingJSON{}
	for _, b := range w.Open[0].Backing {
		by[b.Anchor] = b
	}
	if len(by) != 2 {
		t.Fatalf("backing = %+v, want both anchors in the gap's location", w.Open[0].Backing)
	}
	if g := by["c-a1b2c3d4"]; g.Outcome == "" || g.Kind != "citation" || g.VerifiedBy != "red-lens-evidence" {
		t.Errorf("a verified anchor lost its verification: %+v", g)
	}
	// UNVERIFIED IS AN ENTRY. "nobody looked" and "no anchor here" must not be the same bytes.
	if g, ok := by["c-99887766"]; !ok || g.Outcome != "" {
		t.Errorf("an unchecked anchor is missing or claims an outcome: %+v", g)
	}
}
