package reportvoice

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EVERY GUARDABLE SEAT ID IS REFUSED IN THE REPORT'S VOICE, AND THE LIST IS NOT TYPED.
//
// It was typed, and it guarded `judge-terminal` — a seat #1070 collapsed into `judge` — long
// after no dispatch could produce one. It also omitted `judge` and `frontier` on purpose, because
// both are ordinary English words, and nothing in a hand-written alternation distinguishes a
// considered omission from an oversight.
//
// This walks the roster rather than a list of its own, so there is no second list to forget to
// update. What it proves is that the generation actually REACHES every guardable id, which is the
// part a `for` loop can still get wrong.
func TestEverySeatIDIsRefusedInTheReportsVoice(t *testing.T) {
	for _, id := range seatclass.SeatIDs() {
		if seatclass.Seats[id].CommonWord {
			continue // guarded below, as prose that must stay clean
		}
		s := "The " + id + " concluded that the figure stands."
		if got := Refused(s); len(got) == 0 {
			t.Errorf("a report may say %q and the voice gate does not refuse it — that seat id names who acted in the run", s)
		}
	}
	// The lane shape, which is the one id the roster cannot enumerate.
	if got := Refused("blue-lane-2 drafted this section."); len(got) == 0 {
		t.Error("a lane seat id is not refused in the report's voice")
	}
	// THE OTHER HALF, AND IT IS THE ONE A GENERATED LIST COULD BREAK. An id that is also an
	// ordinary English word must stay OUT: adding the roster wholesale would flag subject prose,
	// and reportvoice's own corpus already carries "a sitting judge" as prose it must not touch.
	for phrase, id := range map[string]string{
		"a sitting judge heard the case":        "judge",
		"at the frontier of the field":          "frontier",
		"the operator of the plant was trained": "operator",
	} {
		if !seatclass.Seats[id].CommonWord {
			t.Errorf("%q is no longer marked CommonWord on the roster, so the voice gate now guards it by name", id)
		}
		if got := Refused(phrase); len(got) != 0 {
			t.Errorf("subject prose %q is refused because %q is guarded by name: %+v", phrase, id, got)
		}
	}
}

// A NAME NOTHING CAN PRODUCE IS NOT GUARDED. The stale alternation spent its whole life refusing
// `judge-terminal`; a gate that guards a retired id is spending its only signal on nothing.
func TestARetiredSeatIDIsNoLongerGuarded(t *testing.T) {
	if got := Refused("judge-terminal assembled the report."); len(got) != 0 {
		t.Errorf("`judge-terminal` is still refused, and no dispatch can produce it (#1070): %+v", got)
	}
}
