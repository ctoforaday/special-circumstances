package record

import (
	"strings"
	"testing"
)

// THE DEFAULT IS THE SHIPPED SET, AND IT IS ASSERTED RATHER THAN INTENDED.
//
// available.go takes a real cost — an experiment knob inside a production path. The whole defence
// of that cost is that unset changes nothing, and a defence nobody checks is a comment. If this
// ever fails, every seat in every real run is reading a list somebody meant only for a probe.
func TestDutyArmDefaultIsTheShippedSet(t *testing.T) {
	t.Setenv(DutyArmEnv, "")
	if got := CurrentDutyArm(); got != DutyShipped {
		t.Fatalf("unset %s = %q, want %q — the shipped behaviour must be what a run gets without opting in", DutyArmEnv, got, DutyShipped)
	}
	// A TYPO MUST NOT EMPTY THE WORKLIST. `off` and "unrecognised" want opposite treatment: the
	// first is a deliberate floor arm, the second is a mistake, and silently returning the floor
	// for a mistake would hand a seat an empty duty list that reads exactly like a clean board.
	t.Setenv(DutyArmEnv, "availabel")
	if got := CurrentDutyArm(); got != DutyShipped {
		t.Errorf("a misspelled arm resolved to %q — it must fall back to shipped, never to off", got)
	}
}

// AND `Available` NEVER DECIDES WHETHER A SITTING IS FINISHED.
//
// This is the rule sitting.go states and the reason Available is a second list rather than more
// duties: a seat told it is unfinished by this view and cleared by every write path learns to
// trust neither surface.
func TestAvailableNeverGatesCompletion(t *testing.T) {
	t.Setenv(DutyArmEnv, string(DutyAvailable))
	b := &Board{Gaps: map[string]*Gap{}}
	s := SittingOf(b, "blue", "blue-respond-r1")
	s.Available = append(s.Available, Duty{What: "an affordance", How: "some verb"})
	// Complete is computed from Outstanding alone; appending affordances cannot move it.
	want := len(s.Outstanding) == 0
	if s.Complete != want {
		t.Fatalf("complete = %v with %d outstanding and %d available — completion must read Outstanding and nothing else",
			s.Complete, len(s.Outstanding), len(s.Available))
	}
}

// THE FLOOR ARM IS ACTUALLY EMPTY.
//
// A control that quietly still emits the shipped duties is not a control, and both arms would
// report the same behaviour for a reason that has nothing to do with the seat.
func TestOffArmEmitsNothing(t *testing.T) {
	t.Setenv(DutyArmEnv, string(DutyOff))
	b := &Board{Gaps: map[string]*Gap{}}
	for _, role := range []string{"blue", "merge", "bench", "lens"} {
		s := SittingOf(b, role, role+"-seat")
		if len(s.Outstanding) != 0 || len(s.Available) != 0 {
			t.Errorf("%s: off arm emitted %d outstanding and %d available — the floor arm must be empty or it is not a floor",
				role, len(s.Outstanding), len(s.Available))
		}
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
	t.Setenv(DutyArmEnv, string(DutyAvailable))

	t.Run("manifest row missing after an edit", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "blue-respond-r1", Type: "blue_edit", Payload: NewPayload().Set("answers", "R1-2")},
		}}
		got := AvailableOf(b, "blue", "blue-respond-r1")
		if !mentions(got, "manifest-row --id R1-2") {
			t.Fatalf("an edit answering R1-2 with no manifest row afforded nothing: %v", hows(got))
		}
		// And it stops once the receipt exists, or the line is a nag rather than a fact.
		b.Events = append(b.Events, Event{SeatID: "blue-respond-r1", Type: "manifest-row", Payload: NewPayload().Set("gap_id", "R1-2")})
		if got := AvailableOf(b, "blue", "blue-respond-r1"); mentions(got, "manifest-row --id R1-2") {
			t.Errorf("the manifest affordance survived its own discharge: %v", hows(got))
		}
	})

	t.Run("grade accepted and never moved", func(t *testing.T) {
		b := &Board{Gaps: map[string]*Gap{}, Events: []Event{
			{SeatID: "red-merge-r1", Type: "motion-rule", Payload: NewPayload().
				Set("subject", "grade").Set("as", "accepted").Set("gap_id", "R1-1")},
		}}
		got := AvailableOf(b, "merge", "red-merge-r1")
		if !mentions(got, "regrade --id R1-1") {
			t.Fatalf("an accepted grade motion with no regrade afforded nothing: %v", hows(got))
		}
		b.Events = append(b.Events, Event{SeatID: "red-merge-r1", Type: "regrade", Payload: NewPayload().Set("gap_id", "R1-1")})
		if got := AvailableOf(b, "merge", "red-merge-r1"); mentions(got, "regrade --id R1-1") {
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
		if got := AvailableOf(b, "merge", "red-merge-r1"); mentions(got, "regrade") {
			t.Errorf("a REJECTED grade motion afforded a regrade: %v", hows(got))
		}
	})
}

func hows(ds []Duty) []string {
	out := []string{}
	for _, d := range ds {
		out = append(out, d.How)
	}
	return out
}

func mentions(ds []Duty, want string) bool {
	for _, d := range ds {
		if strings.Contains(d.How, want) {
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
