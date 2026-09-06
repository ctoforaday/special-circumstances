package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN AFFORDANCE IS ON THE LIST AND DOES NOT BLOCK.
//
// This is the rule sitting.go states — a seat told it is unfinished by this view and cleared by
// every write path learns to trust neither surface — and it is why `Available` used to be a
// SECOND list. It never needed to be. The rule constrains what `complete` may be computed from,
// which is a property of one field; making it a property of a whole separate surface is what put
// a lens seat's entire real workload somewhere its completion check could not see.
//
// So the same guarantee, on one list: affordances appear, affordances carry Blocks:false, and
// `complete` reads the blocking items alone.
func TestAnAffordanceIsListedAndDoesNotBlock(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{
		recordtest.Event(t, "blue-respond-r1", 1, &recordpb.BlueEdit{Answers: proto.String("R1-2")}),
		// Both duties a blue seat owes on an empty board, discharged, so nothing blocks.
		recordtest.Event(t, "blue-respond-r1", 1, &recordpb.Log{}),
		recordtest.Event(t, "blue-respond-r1", 1, &recordpb.Revision{}),
	}}
	s := SittingOf(b.Events, workStatesOfBoardT(b), "blue", "blue-respond-r1")

	var afforded, blocking int
	for _, it := range s.Open {
		if it.Blocks {
			blocking++
		} else {
			afforded++
		}
	}
	if afforded == 0 {
		t.Fatalf("an edit with no manifest row afforded nothing on the work list: %v", hows(s.Open))
	}
	if blocking != 0 {
		t.Fatalf("nothing should block here; blocking items: %v", hows(s.Open))
	}
	if !s.Complete {
		t.Errorf("complete = false with %d afforded and 0 blocking — an affordance must not gate closure; "+
			"that is the whole constraint, and it is now a property of Item.Blocks rather than of a second list", afforded)
	}
	// And the other direction, which the two-list shape could not state at all: the seat is
	// clear to close AND still has work in front of it.
	if s.Complete && len(s.Open) == 0 {
		t.Error("complete with an EMPTY list — the affordances vanished, which is the state a lens seat read before stopping")
	}
}

// EVERY DERIVATION FIRES ON THE STATE IT CLAIMS TO WATCH.
//
// Two of these — the manifest receipt and the unmoved grade — cannot fire at the START of any probe
// board, because both need an act the SEAT performs mid-sitting. That makes them exactly the shape
// this repository keeps paying for: code whose only evidence of working is that nobody has seen it
// fail. A derivation that quietly matches nothing returns an empty affordance list, which reads
// precisely like a board with nothing to offer.
func TestEveryAffordanceDerivationFiresOnItsState(t *testing.T) {

	t.Run("manifest row missing after an edit", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{
			recordtest.Event(t, "blue-respond-r1", 0, &recordpb.BlueEdit{Answers: proto.String("R1-2")}),
		}}
		got := availableOf(b.Events, workStatesOfBoardT(b), "blue", "blue-respond-r1")
		if !mentions(got, "gap R1-2 was answered by an edit and carries no manifest row") {
			t.Fatalf("an edit answering R1-2 with no manifest row afforded nothing: %v", hows(got))
		}
		// And it stops once the receipt exists, or the line is a nag rather than a fact.
		b.Events = append(b.Events, recordtest.Event(t, "blue-respond-r1", 0, &recordpb.ManifestRow{GapId: proto.String("R1-2")}))
		if got := availableOf(b.Events, workStatesOfBoardT(b), "blue", "blue-respond-r1"); mentions(got, "gap R1-2 was answered by an edit and carries no manifest row") {
			t.Errorf("the manifest affordance survived its own discharge: %v", hows(got))
		}
	})

	t.Run("grade accepted and never moved", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{
			// THE FIXTURE IS NOW A REAL EXCHANGE, because the join demands one. The gap id lives
			// on the FILING and the verdict on the RULING, in different shards, and Motions()
			// pairs them on the motion id — so a lone motion-rule carrying a gap_id, which is what
			// this fixture used to be, describes a state the record cannot hold.
			recordtest.Event(t, "blue-respond-r1", 0, &recordpb.Motion{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("R1-1")}},
			}),
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.MotionRule{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_ACCEPTED},
			}),
		}}
		got := availableOf(b.Events, workStatesOfBoardT(b), "merge", "red-merge-r1")
		if !mentions(got, "gap R1-1 had a grade motion ACCEPTED and no regrade") {
			t.Fatalf("an accepted grade motion with no regrade afforded nothing: %v", hows(got))
		}
		b.Events = append(b.Events, recordtest.Event(t, "red-merge-r1", 0, &recordpb.Regrade{GapId: proto.String("R1-1")}))
		if got := availableOf(b.Events, workStatesOfBoardT(b), "merge", "red-merge-r1"); mentions(got, "gap R1-1 had a grade motion ACCEPTED and no regrade") {
			t.Errorf("the regrade affordance survived the regrade: %v", hows(got))
		}
	})

	// A REJECTED motion owes no regrade, and saying it does would be the unmeetable expectation
	// this package's own coverage gate exists to refuse.
	t.Run("a rejected motion affords no regrade", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{
			// THE FIXTURE IS NOW A REAL EXCHANGE, because the join demands one. The gap id lives
			// on the FILING and the verdict on the RULING, in different shards, and Motions()
			// pairs them on the motion id — so a lone motion-rule carrying a gap_id, which is what
			// this fixture used to be, describes a state the record cannot hold.
			recordtest.Event(t, "blue-respond-r1", 0, &recordpb.Motion{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("R1-1")}},
			}),
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.MotionRule{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
				Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED},
			}),
		}}
		if got := availableOf(b.Events, workStatesOfBoardT(b), "merge", "red-merge-r1"); mentions(got, "no regrade followed it") {
			t.Errorf("a REJECTED grade motion afforded a regrade: %v", hows(got))
		}
	})
}

// A duty says WHAT is owed, so these read `what`. They used to read `how` — an invocation this
// type no longer carries, because the help page is the only page that instructs.
func hows(ds []Item) []string {
	out := []string{}
	for _, d := range ds {
		out = append(out, d.What)
	}
	return out
}

func mentions(ds []Item, want string) bool {
	for _, d := range ds {
		if strings.Contains(d.What, want) {
			return true
		}
	}
	return false
}

// inquiryAt builds one `line of inquiry` event, so a test can place a line at a status in a round.
func inquiryAt(t *testing.T, id, status string, round int) *Event {
	t.Helper()
	st, ok := AvenueStatusOf(status)
	if !ok {
		t.Fatalf("%q is not a line-of-inquiry status", status)
	}
	return recordtest.Event(t, "blue-respond-r"+string(rune('0'+round)), round, &recordpb.Avenue{
		AvenueId: proto.String(id),
		Status:   &st,
		Line:     proto.String("a line"),
	})
}

// A REAFFIRMED LINE STOPS NAGGING; A NEGLECTED ONE DOES NOT. They used to be the same bytes.
//
// StaleInquiries and availableOf each carried `Status == "proposed" || Status == "pursued"` and
// nothing else, while the affordance's text said a line of inquiry "has no fate THIS ROUND" and
// StaleInquiries' own doc said "a line of inquiry still open LATE IN A RUN". Neither read `Inquiry.Round`,
// which was populated on every event.
//
// So blue moving a line to `pursued` this round with what it learned — the enum's own definition
// of that status, "you are following it, OR YOU FOLLOWED IT" — produced the identical line to a
// line untouched since round 0. The only statuses that DID clear it were `declined`, `abandoned`
// and `deferred`, all of which mean stop: the channel could express giving up and not carrying on.
func TestAPursuedInquiryReaffirmedThisRoundIsNotStale(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []*Event{
		inquiryAt(t, "Q1", "proposed", 0),
		inquiryAt(t, "Q1", "pursued", 0),
		inquiryAt(t, "Q2", "pursued", 0),
		inquiryAt(t, "Q2", "pursued", 2), // reaffirmed in the current round
		inquiryAt(t, "Q3", "pursued", 0), // never revisited
		inquiryAt(t, "Q4", "deferred", 0),
		inquiryAt(t, "Q5", "abandoned", 0),
		inquiryAt(t, "Q6", "proposed", 2), // undecided, and `proposed` owes a move whenever asked
		inquiryAt(t, "Z", "pursued", 2),   // carries the round forward
	}}

	stale := map[string]bool{}
	for _, a := range StaleInquiries(b) {
		stale[a.ID] = true
	}

	if stale["Q2"] {
		t.Error("A2 was reaffirmed as `pursued` in the current round and is still reported as owing a decision — " +
			"recording exactly what the enum asks for must settle the line, or the only way to clear it is to abandon it")
	}
	if !stale["Q3"] {
		t.Error("A3 has sat at `pursued` since round 0 and is NOT reported — that is the neglect this exists to catch")
	}
	if !stale["Q6"] {
		t.Error("A6 is `proposed` — the enum calls that \"the state that owes a move\", with no round condition")
	}
	for _, settled := range []string{"Q4", "Q5"} {
		if stale[settled] {
			t.Errorf("%s is at a settled fate and is reported as owing a decision — `deferred` in particular is a "+
				"DECISION (worth taking, not by this run), not an omission", settled)
		}
	}
}

// A `carried` RULING RE-OPENS THE FILING, AND THAT IS THE POINT OF `carried`.
//
// The docket affordance was keyed on "has this gap EVER been docketed", so it went silent
// permanently at the first filing. `carried` is a deferral: the bench answers the motion and
// deliberately keeps the gap alive with a stated condition for revisiting it, and the gap comes
// back as a FRESH filing next round. Under the old key the surface that would prompt that filing
// never fired again, so the one disposition the bench uses most often — 76 of 77 rulings on the
// measured corpus — had no route back to the bench.
//
// THE KEY IS PENDING, NOT EVER-FILED. Every disposition except `carried` closes the gap, so an
// OPEN gap whose docket motion is RULED is exactly the deferred one. An UNRULED motion still
// suppresses the affordance, because asking the bench the same question twice while it is thinking
// is not work.
func TestACarriedDocketRulingOffersTheGapBackToTheBench(t *testing.T) {
	// One motion per gap: G-carried is ruled and stays open; G-pending is filed and unruled;
	// G-fresh was never docketed at all. Only G-pending must be silent.
	file := func(motionID, gapID string) *Event {
		return recordtest.Event(t, "red-merge-r1", 1, &recordpb.Motion{
			MotionId: proto.String(motionID),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:    proto.String("red cannot settle " + gapID),
			Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}},
		})
	}
	b := &Board{
		GapOrder: []string{"G-carried", "G-pending", "G-fresh"},
		Gaps: map[string]*Gap{
			"G-carried": {ID: "G-carried", Open: true},
			"G-pending": {ID: "G-pending", Open: true},
			"G-fresh":   {ID: "G-fresh", Open: true},
		},
		Events: []*Event{
			file("M1", "G-carried"),
			recordtest.Event(t, "judge-r1", 1, &recordpb.MotionRule{
				MotionId: proto.String("M1"),
				Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
				Opinion:  proto.String("not this round"),
				Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
					Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED),
					ReopensOn:   proto.String("blue reporting what the stated direction found"),
				}},
			}),
			file("M2", "G-pending"),
		},
	}
	open := availableOf(b.Events, workStatesOfBoardT(b), "merge", "red-merge-r1")

	for _, want := range []string{"G-carried", "G-fresh"} {
		if !mentions(open, "gap "+want+" is open") {
			t.Errorf("%s is open and the bench is not being offered it: %v", want, hows(open))
		}
	}
	if mentions(open, "gap G-pending is open") {
		t.Error("G-pending's motion is filed and UNRULED — offering a second filing asks the bench the same question twice")
	}
	// AND IT IS STILL AN AFFORDANCE. A blocking version would make a seat that obeyed it FURTHER
	// from complete, because filing adds an unruled motion.
	for _, it := range open {
		if strings.Contains(it.What, "motion docket file") && it.Blocks {
			t.Errorf("the docket affordance blocks closure: %q", it.What)
		}
	}
}
