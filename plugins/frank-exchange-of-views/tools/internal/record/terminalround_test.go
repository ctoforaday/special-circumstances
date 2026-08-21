package record

import "testing"

// #327/#396, END TO END: a bench act at run END is recorded AFTER the rounds, not before them.
//
// `judge-terminal` carries no `-r<N>`, the derivation answered 0, and 0 is synthesis — so the
// terminal closure sorted and rendered as though it had happened before round 1 ever ran. It put
// a phantom entry in the archive and made the W1.8 spot-check floor demand samples from rounds
// whose seats had done nothing wrong. It was found at 1 seed in 60, by luck.
//
// The round for a terminal seat is a FACT ON THE RECORD rather than something it has to be told:
// it acts after the last round, and every round's seats have already stamped theirs.
func TestATerminalSeatIsRecordedAfterTheRoundsRatherThanBeforeThem(t *testing.T) {
	runDir := newRun(t)
	for _, s := range []string{"red-merge-r1", "red-merge-r2", "red-merge-r3"} {
		if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: s, Round: RoundIn(runDir)(s)}); err != nil {
			t.Fatal(err)
		}
	}

	got := RoundIn(runDir)("judge-terminal")
	if got == 0 {
		t.Fatal("the terminal sitting is recorded at round 0 — synthesis — so every reader orders the run's LAST act before its first")
	}
	if got != 3 {
		t.Errorf("terminal round = %d, want 3: it acts after the last round on the record", got)
	}
}

// AND SYNTHESIS IS STILL ROUND 0, by rule rather than by accident. These seats are dispatched
// before the round loop, so 0 is exact — and the fix must not turn a correct 0 into an unknown.
func TestSynthesisSeatsAreRoundZeroByRule(t *testing.T) {
	runDir := newRun(t)
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r4", Round: 4}); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"frontier", "blue-synthesize", "blue-lane-2"} {
		if got := RoundIn(runDir)(s); got != 0 {
			t.Errorf("%s resolved to round %d; it runs in synthesis, which IS round 0 — and the record showing 4 must not drag it there", s, got)
		}
	}
}

// AN EMPTY RUN CANNOT ANSWER, and says so. Falling back to 0 here would rebuild the exact
// conflation this removes, one layer up: "no rounds have happened" would again be spelled the
// same as "this happened in synthesis".
func TestATerminalSeatOnAnEmptyRunIsUnknownRatherThanZero(t *testing.T) {
	if got := RoundIn(t.TempDir())("judge-terminal"); got != -1 {
		t.Errorf("terminal round on an empty run = %d, want -1 (unknown); 0 would claim it happened in synthesis", got)
	}
}
