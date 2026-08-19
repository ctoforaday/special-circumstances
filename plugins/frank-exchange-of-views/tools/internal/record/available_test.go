package record

import (
	"strings"
	"testing"
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
	b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
		{SeatID: "blue-respond-r1", Type: "blue_edit", Payload: NewPayload().Set("answers", "R1-2")},
		// Both duties a blue seat owes on an empty board, discharged, so nothing blocks.
		{SeatID: "blue-respond-r1", Type: "friction"},
		{SeatID: "blue-respond-r1", Type: "revision"},
	}}
	s := SittingOf(b, "blue", "blue-respond-r1")

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
		b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "blue-respond-r1", Type: "blue_edit", Payload: NewPayload().Set("answers", "R1-2")},
		}}
		got := AvailableOf(b, "blue", "blue-respond-r1")
		if !mentions(got, "gap R1-2 was answered by an edit and carries no manifest row") {
			t.Fatalf("an edit answering R1-2 with no manifest row afforded nothing: %v", hows(got))
		}
		// And it stops once the receipt exists, or the line is a nag rather than a fact.
		b.Events = append(b.Events, Event{SeatID: "blue-respond-r1", Type: "manifest-row", Payload: NewPayload().Set("gap_id", "R1-2")})
		if got := AvailableOf(b, "blue", "blue-respond-r1"); mentions(got, "gap R1-2 was answered by an edit and carries no manifest row") {
			t.Errorf("the manifest affordance survived its own discharge: %v", hows(got))
		}
	})

	t.Run("grade accepted and never moved", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "red-merge-r1", Type: "motion-rule", Payload: NewPayload().
				Set("subject", "grade").Set("as", "accepted").Set("gap_id", "R1-1")},
		}}
		got := AvailableOf(b, "merge", "red-merge-r1")
		if !mentions(got, "gap R1-1 had a grade motion ACCEPTED and no regrade") {
			t.Fatalf("an accepted grade motion with no regrade afforded nothing: %v", hows(got))
		}
		b.Events = append(b.Events, Event{SeatID: "red-merge-r1", Type: "regrade", Payload: NewPayload().Set("gap_id", "R1-1")})
		if got := AvailableOf(b, "merge", "red-merge-r1"); mentions(got, "gap R1-1 had a grade motion ACCEPTED and no regrade") {
			t.Errorf("the regrade affordance survived the regrade: %v", hows(got))
		}
	})

	// A REJECTED motion owes no regrade, and saying it does would be the unmeetable expectation
	// this package's own coverage gate exists to refuse.
	t.Run("a rejected motion affords no regrade", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "red-merge-r1", Type: "motion-rule", Payload: NewPayload().
				Set("subject", "grade").Set("as", "rejected").Set("gap_id", "R1-1")},
		}}
		if got := AvailableOf(b, "merge", "red-merge-r1"); mentions(got, "no regrade followed it") {
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
func inquiryAt(id, status string, round int) Event {
	return Event{
		Type:    "line-of-inquiry",
		Round:   round,
		SeatID:  "blue-respond-r" + string(rune('0'+round)),
		Payload: NewPayload().Set("inquiry_id", id).Set("status", status).Set("line", "a line"),
	}
}

// A REAFFIRMED LINE STOPS NAGGING; A NEGLECTED ONE DOES NOT. They used to be the same bytes.
//
// StaleInquiries and AvailableOf each carried `Status == "proposed" || Status == "pursued"` and
// nothing else, while the affordance's text said a line of inquiry "has no fate THIS ROUND" and
// StaleInquiries' own doc said "a line of inquiry still open LATE IN A RUN". Neither read `Inquiry.Round`,
// which was populated on every event.
//
// So blue moving a line to `pursued` this round with what it learned — the enum's own definition
// of that status, "you are following it, OR YOU FOLLOWED IT" — produced the identical line to a
// line untouched since round 0. The only statuses that DID clear it were `declined`, `abandoned`
// and `deferred`, all of which mean stop: the channel could express giving up and not carrying on.
func TestAPursuedInquiryReaffirmedThisRoundIsNotStale(t *testing.T) {
	b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
		inquiryAt("Q1", "proposed", 0),
		inquiryAt("Q1", "pursued", 0),
		inquiryAt("Q2", "pursued", 0),
		inquiryAt("Q2", "pursued", 2), // reaffirmed in the current round
		inquiryAt("Q3", "pursued", 0), // never revisited
		inquiryAt("Q4", "deferred", 0),
		inquiryAt("Q5", "abandoned", 0),
		inquiryAt("Q6", "proposed", 2), // undecided, and `proposed` owes a move whenever asked
		inquiryAt("Z", "pursued", 2),   // carries the round forward
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
