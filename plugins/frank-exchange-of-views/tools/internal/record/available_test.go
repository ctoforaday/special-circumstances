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
