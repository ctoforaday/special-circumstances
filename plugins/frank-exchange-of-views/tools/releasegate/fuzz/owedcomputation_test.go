package fuzz

import "testing"

// A FOLD KEPT BY AN UNCHECKED COMPUTATION IS COUNTED BY THE RECORD'S COMPUTATION IDS. The tally read
// the debt by its id prefix, and when the computation id's letter changed every such fold read as
// kept by something else: 407 folds kept by a debt, 0 by a computation, while the plans named
// unchecked K- ids. An unchecked citation, and an id the record holds no computation for, are not
// computations.
func TestAFoldOwingAComputationIsReadByTheRecordsIDs(t *testing.T) {
	computations := map[string]bool{"K-26b691db": true}
	for clause, want := range map[string]bool{
		"unchecked K-26b691db":                         true,
		"unchecked C-11111111, unchecked K-26b691db":   true,
		"unchecked C-11111111":                         false,
		"unchecked K-00000000":                         false,
		"blue's answer to G-26b691db":                  false,
		"red-lens-logic's finding K-26b691db, nothing": false,
	} {
		if got := owesAComputation(clause, computations); got != want {
			t.Errorf("%q owes a computation = %v, want %v", clause, got, want)
		}
	}
}
