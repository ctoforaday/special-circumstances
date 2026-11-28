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

// A CARRIED GAP GETS ITS OWN ROW, AND A FRESH ONE NONE (#759). The chair files nothing on a gap — a
// proposal is its author's, and the dispatch dockets a gap at impasse — so a gap nobody docketed is
// offered to the chair by no row; the remanded one carries the bench's stated condition, in one row
// that does not block (sitting.go's open-gap row already refuses PASS over a material gap).
func TestACarriedGapReadsDifferentlyFromOneNobodyDocketed(t *testing.T) {
	gaps := []WorkGapState{
		{ID: "CARRIED", Open: true, Material: true, Remanded: true, remand: remandOwed, route: routeRemandOwed,
			DocketDirection: "blue reporting what the stated direction found"},
		{ID: "FRESH", Open: true, Material: true},
	}
	open := availableOf(nil, WindowIndex{}, gaps, "chair", "red-chair", false)
	var carried []Item
	for _, it := range open {
		if strings.Contains(it.What, "gap FRESH ") {
			t.Errorf("the chair is offered a row on a gap nobody docketed, and files nothing on it: %q", it.What)
		}
		if strings.Contains(it.What, "gap CARRIED ") {
			carried = append(carried, it)
		}
	}
	if len(carried) != 1 {
		t.Fatalf("the carried gap has %d rows on the work list, want exactly 1: %v", len(carried), hows(open))
	}
	if c := carried[0]; c.Blocks || !strings.Contains(c.What, "REMANDED it") || !strings.Contains(c.What, "blue reporting what the stated direction found") {
		t.Errorf("the carried gap's row blocks, or drops the remand or its stated condition: %+v", c)
	}
}
