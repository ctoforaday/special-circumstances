package fuzz

import "testing"

// A CONTRADICTION DRAWS ON THE LENS'S STOCK. A contradicting reading owes a gap the engine exempts
// from the budget and from one gap per class, so a lens that reads one at every sitting keeps the
// run from converging (readingOutcome). Past its stock, or past a forced run's first sitting, a lens
// reads none; within its stock it reads both kinds, so the raise and its answer stay driven.
func TestALensPastItsStockReadsNoContradiction(t *testing.T) {
	const seat = "red-lens-evidence"
	draw := func(r *runner) map[string]int {
		got := map[string]int{}
		for range 600 {
			got[r.readingOutcome(seat)]++
		}
		return got
	}
	for _, tc := range []struct {
		name string
		r    *runner
	}{
		{"a drawn run's lens at its stock", &runner{rng: newLockedRand(1), lensMints: map[string]int{seat: fuzzLensStock}}},
		{"a drawn run's lens past its stock", &runner{rng: newLockedRand(1), lensMints: map[string]int{seat: fuzzLensStock + 1}}},
		{"a forced run past its first sitting", &runner{rng: newLockedRand(1), forceVerified: true, chairRegisters: 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := draw(tc.r); got["refutes"]+got["absent"] > 0 {
				t.Errorf("drew %d refutes and %d absent readings, want none: %v", got["refutes"], got["absent"], got)
			}
		})
	}
	for _, tc := range []struct {
		name string
		r    *runner
	}{
		{"a drawn run's lens within its stock", &runner{rng: newLockedRand(1), lensMints: map[string]int{seat: fuzzLensStock - 1}}},
		{"a forced run at its first sitting", &runner{rng: newLockedRand(1), forceVerified: true, chairRegisters: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := draw(tc.r); got["refutes"] == 0 || got["absent"] == 0 {
				t.Errorf("drew %d refutes and %d absent readings, want both: %v", got["refutes"], got["absent"], got)
			}
		})
	}
}
