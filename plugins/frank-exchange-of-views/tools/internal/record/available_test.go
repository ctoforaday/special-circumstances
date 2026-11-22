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
	b := loadedFamilyT(t, nil, []*Event{
		dispatchBlue(t, "G1"), registers(t, "blue-respond"),
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		// What a blue-respond sitting owes for a gap it found open, discharged, so nothing blocks.
		recordtest.Event(t, "blue-respond", seatLog("nothing in the way")),
		recordtest.Event(t, "blue-respond", &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("what the bench is asked to weigh")}),
	})
	s := SittingOf(b.Events, b.At.IDs(b.Events), b.At, workStatesOfFamilyT(b), "blue", "blue-respond", nil)

	var afforded, blocking int
	for _, it := range s.Open {
		if it.Blocks {
			blocking++
		} else {
			afforded++
		}
	}
	if afforded == 0 {
		t.Fatalf("the sitting afforded nothing on the work list: %v", hows(s.Open))
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

// inquiryAt builds one `inquiry` event, so a test can place an inquiry at a status in a round.
func inquiryAt(t *testing.T, id, status string) *Event {
	t.Helper()
	st, ok := InquiryStatusOf(status)
	if !ok {
		t.Fatalf("%q is not an inquiry status", status)
	}
	return recordtest.Event(t, "blue-respond", &recordpb.Inquiry{
		InquiryId: proto.String(id),
		Status:    &st,
		Question:  proto.String("a question"),
	})
}

// chairSits is what moves the epoch: a red-chair register on the record (plans/roundless.md
// §III.A.0). A fixture that wants an event "in epoch 2" puts two of these before it.
func chairSits(t *testing.T) *Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Register{})
}

// A REAFFIRMED INQUIRY STOPS NAGGING; A NEGLECTED ONE DOES NOT.
//
// Recording an inquiry as `underway` this round with what it learned — the enum's own definition
// of that status, "you are working the question and it is still open" — settles it for the round,
// and an inquiry untouched since round 0 is stale. The fates that mean stop (`declined`,
// `abandoned`, `deferred`) settle it too: the channel expresses carrying on as well as giving up.
func TestAnUnderwayInquiryReaffirmedThisRoundIsNotStale(t *testing.T) {
	b := loadedFamilyT(t, nil, []*Event{
		// Epoch 0: the base phase, before any chair has sat.
		inquiryAt(t, "Q1", "proposed"),
		inquiryAt(t, "Q1", "underway"),
		inquiryAt(t, "Q2", "underway"),
		inquiryAt(t, "Q3", "underway"), // never revisited
		inquiryAt(t, "Q4", "deferred"),
		inquiryAt(t, "Q5", "abandoned"),
		inquiryAt(t, "Q7", "confirmed"), // followed to its end in epoch 0, never touched again
		// The chair sits twice: everything below is in epoch 2, the current one.
		chairSits(t),
		chairSits(t),
		inquiryAt(t, "Q2", "underway"), // reaffirmed in the current epoch
		inquiryAt(t, "Q6", "proposed"), // undecided, and `proposed` owes a fate whenever asked
		inquiryAt(t, "Z", "underway"),  // carries the epoch forward
	})

	stale := map[string]bool{}
	for _, a := range StaleInquiriesOf(b.Events, b.At) {
		stale[a.ID] = true
	}

	if stale["Q2"] {
		t.Error("Q2 was reaffirmed as `underway` in the current epoch and is still reported as owing a decision — " +
			"recording exactly what the enum asks for must settle the inquiry, or the only way to clear it is to abandon it")
	}
	if !stale["Q3"] {
		t.Error("Q3 has sat at `underway` since epoch 0 and is NOT reported — that is the neglect this exists to catch")
	}
	if !stale["Q6"] {
		t.Error("Q6 is `proposed` — the enum says it \"owes a fate\", with no round condition")
	}
	// CONFIRMED IS A SETTLED FATE FOR AN INQUIRY THAT WORKED. Without it an inquiry could only stop
	// nagging by dying.
	for _, settled := range []string{"Q4", "Q5", "Q7"} {
		if stale[settled] {
			t.Errorf("%s is at a settled fate and is reported as owing a decision — `deferred` in particular is a "+
				"DECISION (worth taking, not by this run), not an omission", settled)
		}
	}
}

// A `carried` RULING RE-OPENS THE FILING, AND THAT IS THE POINT OF `carried`.
//
// The docket affordance was keyed on "has this gap EVER been docketed", so it went silent
// permanently at the first filing. `carried` is a deferral: the bench answers the docketing and
// deliberately keeps the gap alive with a stated condition for revisiting it, and the gap comes
// back as a FRESH proposal. Under the old key the surface that would prompt that filing
// never fired again, so the one disposition the bench uses most often — 76 of 77 rulings on the
// measured corpus — had no route back to the bench.
//
// THE KEY IS PENDING, NOT EVER-FILED. Every disposition except `carried` closes the gap, so an
// OPEN gap whose docketing is RULED is exactly the deferred one. A STANDING proposal still
// suppresses the affordance, because proposing a second act while the first is unanswered is
// refused (one standing proposal per gap).
func TestACarriedDocketRulingOffersTheGapBackToTheBench(t *testing.T) {
	// G-carried was docketed and the bench remanded it, so it stays open; G-pending carries a proposal
	// standing unanswered; G-fresh was never before the bench at all. Only G-pending must be silent.
	// All three are MATERIAL: the docket is offered only on a gap that holds PASS, and this test is
	// about which of three such gaps the offer stands on.
	b := loadedFamilyT(t, []*Gap{
		{ID: "G-carried", Open: true, Material: true},
		{ID: "G-pending", Open: true, Material: true},
		{ID: "G-fresh", Open: true, Material: true},
	},
		[]*Event{
			mintsGap(t, "red-lens-logic", "G-carried"),
			mintsGap(t, "red-lens-logic", "G-pending"),
			recordtest.Event(t, "red-chair", docketing("D1", "G-carried")),
			recordtest.Event(t, "judge", &recordpb.MotionRule{
				Opinion: proto.String("not this round"),
				Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
					DocketingId: proto.String("D1"),
					Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED),
					Direction:   proto.String("blue reporting what the stated direction found"),
					Principle:   proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s"),
				}},
			}),
			recordtest.Event(t, "red-chair", docketMotion("M2", "G-pending")),
		})
	open := availableOf(b.Events, b.At, workStatesOfFamilyT(b), "chair", "red-chair", false, nil)

	for _, want := range []string{"G-carried", "G-fresh"} {
		if !mentions(open, "gap "+want+" ") {
			t.Errorf("%s is open and the bench is not being offered it: %v", want, hows(open))
		}
	}
	if mentions(open, "gap G-pending with a docket motion") {
		t.Error("G-pending's motion is filed and UNRULED — offering a second filing asks the bench the same question twice")
	}
	// AND IT IS STILL AN AFFORDANCE. A blocking version would make a seat that obeyed it FURTHER
	// from complete, because filing adds an unruled motion.
	for _, it := range open {
		if strings.Contains(it.What, "with a docket motion") && it.Blocks {
			t.Errorf("the docket affordance blocks closure: %q", it.What)
		}
	}
}

// A CARRIED GAP GETS DIFFERENT WORDS, AND THAT DIFFERENCE IS THE WHOLE POINT (#759).
//
// The vacuous version of this test is "the chair sitting is incomplete and names the gap" — which
// passed before any of this existed, because the chair arm already blocks on every open material gap. So
// what is asserted here is the DISTINCTION: a carried gap and a never-docketed one must not read
// the same, the carried one must carry the bench's stated condition, and neither may add a second
// blocking row for a gap sitting.go already blocks on.
func TestACarriedGapReadsDifferentlyFromOneNobodyDocketed(t *testing.T) {
	gaps := []WorkGapState{
		{ID: "CARRIED", Open: true, Material: true, Remanded: true, remand: remandOwed, route: routeRemandOwed,
			DocketDirection: "blue reporting what the stated direction found"},
		{ID: "FRESH", Open: true, Material: true},
	}
	open := availableOf(nil, WindowIndex{}, gaps, "chair", "red-chair", false, nil)

	find := func(id string) string {
		t.Helper()
		for _, it := range open {
			if strings.Contains(it.What, "gap "+id+" ") {
				if it.Blocks {
					t.Errorf("%s: the docket affordance BLOCKS — sitting.go's open-gap row already "+
						"refuses PASS over a material gap, so this would be a second blocking row for one gap", id)
				}
				return it.What
			}
		}
		t.Fatalf("%s has no row at all: %v", id, hows(open))
		return ""
	}
	carried, fresh := find("CARRIED"), find("FRESH")

	// THE ANTI-VACUITY ASSERTION. Everything else here would still pass if both gaps got the
	// generic sentence; this is the line that fails when they do.
	if carried == fresh {
		t.Fatalf("a carried gap and a gap nobody has docketed read identically:\n  %q", carried)
	}
	if !strings.Contains(carried, "REMANDED it") {
		t.Errorf("the carried gap's row does not say the bench remanded it: %q", carried)
	}
	if !strings.Contains(carried, "blue reporting what the stated direction found") {
		t.Errorf("the carried gap's row drops the bench's stated condition, which is the substance "+
			"of the deferral: %q", carried)
	}
	// ONE ROW PER GAP. A second item naming the same gap is the duplicate #759 names explicitly.
	rows := 0
	for _, it := range open {
		if strings.Contains(it.What, "gap CARRIED ") {
			rows++
		}
	}
	if rows != 1 {
		t.Errorf("the carried gap has %d rows on the work list, want exactly 1: %v", rows, hows(open))
	}
	// And the never-docketed gap still gets the filing instruction it always did.
	if !strings.Contains(fresh, "with a docket motion") {
		t.Errorf("the undocketed gap lost its filing affordance: %q", fresh)
	}
}

// A DOCKETED GAP TAKES NO PROPOSAL (L38), so the chair is not offered one there: a docketing standing
// with no proposal suppresses the row as a standing proposal does.
func TestTheChairIsNotOfferedAProposalOnADocketedGap(t *testing.T) {
	b := loadedFamilyT(t, []*Gap{
		{ID: "G-docketed", Open: true, Material: true},
		{ID: "G-fresh", Open: true, Material: true},
	},
		[]*Event{
			mintsGap(t, "red-lens-logic", "G-docketed"),
			mintsGap(t, "red-lens-logic", "G-fresh"),
			recordtest.Event(t, "red-chair", docketing("D1", "G-docketed")),
		})
	open := availableOf(b.Events, b.At, workStatesOfFamilyT(b), "chair", "red-chair", false, nil)
	if mentions(open, "gap G-docketed with a docket motion") {
		t.Error("G-docketed has a docketing standing, and no proposal is filed on a docketed gap — the chair is offered one")
	}
	if !mentions(open, "gap G-fresh ") {
		t.Errorf("G-fresh is open and undocketed, and the chair is not offered a proposal on it: %v", hows(open))
	}
}
