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
	if _, err := ResolveSeat("red-lens-r1-L5", bound("red-lens-r1-L1"), nil); err == nil {
		t.Fatal("a --seat-id naming a DIFFERENT seat was accepted; attribution would silently land on the wrong lens")
	}
	// Agreement is not a conflict, and the flag alone still works before this agent has registered.
	if s, err := ResolveSeat("red-lens-r1-L1", bound("red-lens-r1-L1"), nil); err != nil || s.ID != "red-lens-r1-L1" {
		t.Errorf("the same value from both sources must resolve, got %+v %v", s, err)
	}
	if s, err := ResolveSeat("red-lens-r1-L1", bound(""), nil); err != nil || s.ID != "red-lens-r1-L1" {
		t.Errorf("an unregistered agent must still be able to identify itself by flag, got %+v %v", s, err)
	}
}

// A LOOKUP FAILURE IS NOT AN UNREGISTERED AGENT, and folding the two together is how this
// mechanism would acquire the defect it was built to remove: a record nobody can read would
// present as a seat that has not registered, and the seat would be sent to register a second
// time against a record that is not there.
func TestAnUnreadableBindingIsAnErrorRatherThanAnAbsence(t *testing.T) {
	boom := func() (string, error) { return "", errRead }
	if _, err := ResolveSeat("red-lens-r1-L1", boom, nil); err == nil {
		t.Fatal("a record that could not be read resolved as an agent that simply has not registered")
	}
}

type readErr struct{}

func (readErr) Error() string { return "the record could not be read" }

var errRead = readErr{}

// UNKNOWN IS NOT ROUND ZERO. Round 0 is synthesis — a real round with real events — so a caller
// that cannot know the round must not be handed 0.
//
// This is the phantom-archive bug in miniature (#327): `judge-terminal` carries no round, the
// regex returned 0, and a bench closure at run END looked like a closure BEFORE round 1. It made
// the W1.8 spot-check floor demand samples from rounds whose seats had done nothing wrong.
func TestAnUnknownRoundIsNotZero(t *testing.T) {
	s, err := ResolveSeat("judge-terminal", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.HasRound() {
		t.Errorf("a seat with no injected round and no inference must report UNKNOWN, got round %d", s.Round)
	}
	if s.Round == 0 {
		t.Error("unknown must not be spelled 0 — round 0 is synthesis, and conflating them is the phantom-archive bug")
	}
}

// The injected round is a FACT and wins over any inference from the id's shape.
func TestTheInjectedRoundWinsOverTheIDsShape(t *testing.T) {
	t.Setenv(RoundVar, "4")
	s, err := ResolveSeat("", bound("judge-terminal"), func(string) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}
	if s.Round != 4 {
		t.Errorf("the injected round must win, got %d — the seat id says nothing about the round here, which is exactly when guessing hurts", s.Round)
	}
}

// A malformed injected round is a DISPATCH bug and must be loud. Falling through to the regex
// would answer a question nobody asked correctly.
func TestAMalformedInjectedRoundIsRefused(t *testing.T) {
	for _, bad := range []string{"two", "-1"} {
		t.Setenv(RoundVar, bad)
		if _, err := ResolveSeat("", bound("red-merge-r2"), nil); err == nil {
			t.Errorf("FEOV_ROUND=%q was accepted; a malformed round is a dispatch bug, not something to guess around", bad)
		}
	}
}
