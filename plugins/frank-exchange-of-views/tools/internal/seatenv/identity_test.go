package seatenv

import "testing"

// bound is the resolver a caller supplies: which seat this agent registered as. Nothing sets an
// environment variable for it any more — FEOV_SEAT had readers and no writer, so these tests were
// exercising a branch no real run could reach. The value now comes off the record, and a test
// supplies it directly.
func bound(seat string) func() (string, error) {
	return func() (string, error) { return seat, nil }
}

// A DISAGREEING --seat-id IS REFUSED, not obeyed and not silently overridden.
//
// Attribution is the one fact a seat must not be able to get wrong: an event filed under the
// wrong seat is credited to the wrong party, and found_by, estoppel and every parity check read
// it. Obeying the flag reinstates the typo; overriding it silently makes a seat's own argument
// vanish, which is the failure where a seat spends a round arguing with a value it never chose.
func TestASeatIDDisagreeingWithTheBoundIdentityIsRefused(t *testing.T) {
	if _, err := ResolveSeat("red-lens-logic", bound("red-lens-evidence")); err == nil {
		t.Fatal("a --seat-id naming a DIFFERENT seat was accepted; attribution would silently land on the wrong lens")
	}
	// Agreement is not a conflict, and the flag alone still works before this agent has registered.
	if s, err := ResolveSeat("red-lens-evidence", bound("red-lens-evidence")); err != nil || s.ID != "red-lens-evidence" {
		t.Errorf("the same value from both sources must resolve, got %+v %v", s, err)
	}
	if s, err := ResolveSeat("red-lens-evidence", bound("")); err != nil || s.ID != "red-lens-evidence" {
		t.Errorf("an unregistered agent must still be able to identify itself by flag, got %+v %v", s, err)
	}
}

// A LOOKUP FAILURE IS NOT AN UNREGISTERED AGENT, and folding the two together is how this
// mechanism would acquire the defect it was built to remove: a record nobody can read would
// present as a seat that has not registered, and the seat would be sent to register a second
// time against a record that is not there.
func TestAnUnreadableBindingIsAnErrorRatherThanAnAbsence(t *testing.T) {
	boom := func() (string, error) { return "", errRead }
	if _, err := ResolveSeat("red-lens-evidence", boom); err == nil {
		t.Fatal("a record that could not be read resolved as an agent that simply has not registered")
	}
}

type readErr struct{}

func (readErr) Error() string { return "the record could not be read" }

var errRead = readErr{}

// (A Seat no longer carries a round at all, so "unknown is not zero" has nothing left to guard
// here. The phantom-archive bug it named, #327 — a terminal seat's closure filed as round 0 —
// cannot recur: the record stamps the EPOCH at the write, which is always defined and is 0 only
// before any chair has sat. See events_w."epoch" and TestATerminalSeatIsStampedWithTheEpochItActsIn.)
