package fuzz

import (
	"slices"
	"testing"
)

// TestAForcedRunDrawsNoReadingThatOwesALaterMint holds the forced runs' bound (findsFresh) at the
// readings of a citation: a run forced to a verdict never reads one unreachable, which is a failed
// check whose gap its reader owes and whose remedy readies blue, and past its first sitting reads
// none as contradicting, which owes a gap. A drawn run draws every reading.
func TestAForcedRunDrawsNoReadingThatOwesALaterMint(t *testing.T) {
	const draws = 2000
	for _, c := range []struct {
		name  string
		force func(*runner)
	}{
		{"forced UNVERIFIED", func(r *runner) { r.forceUnverified = true }},
		{"forced VERIFIED", func(r *runner) { r.forceVerified = true }},
	} {
		t.Run(c.name, func(t *testing.T) {
			for _, sitting := range []int{1, 2} {
				r := &runner{rng: newLockedRand(1), chairRegisters: sitting}
				c.force(r)
				for range draws {
					if got := r.citationReading("red-lens-evidence"); got == "unreachable" ||
						sitting > 1 && (got == "refutes" || got == "absent") {
						t.Fatalf("chair sitting %d: a drawn reading came back %q", sitting, got)
					}
					if got := r.owedCitationReading(); got != "supports" {
						t.Fatalf("chair sitting %d: an owed check came back %q", sitting, got)
					}
				}
			}
		})
	}
	t.Run("drawn", func(t *testing.T) {
		r := &runner{rng: newLockedRand(1), chairRegisters: 2}
		drawn, owed := map[string]bool{}, map[string]bool{}
		for range draws {
			drawn[r.citationReading("red-lens-evidence")] = true
			owed[r.owedCitationReading()] = true
		}
		for _, o := range verifyOutcomes {
			if !drawn[o] {
				t.Errorf("a drawn run never read a source %q in %d draws", o, draws)
			}
		}
		if !owed["unreachable"] || !owed["supports"] {
			t.Errorf("a drawn run's owed checks came back %v, want supports and unreachable", owed)
		}
	})
}

// TestAForcedRunLeavesNoInquiryOpen: a run forced to a verdict records a settling fate on every
// inquiry it proposes, since an open one readies blue-respond on every plan.
func TestAForcedRunLeavesNoInquiryOpen(t *testing.T) {
	for _, r := range []*runner{
		{rng: newLockedRand(1), forceUnverified: true},
		{rng: newLockedRand(1), forceVerified: true},
	} {
		for range 200 {
			if st := r.firstInquiryFate(); !slices.Contains(settlingFates, st) {
				t.Fatalf("a forced run's proposal recorded %q first, which leaves it open", st)
			}
		}
	}
}
