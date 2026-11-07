package fuzz

import (
	"maps"
	"slices"
	"testing"
)

// TestAForcedRunDrawsNoReadingThatOwesALaterMint holds the forced runs' bound (findsFresh) at the
// readings of a citation: a run forced to a verdict never reads one unreachable, which is a failed
// check whose gap its reader owes and whose remedy readies blue, and a lens that draws no fresh mint
// reads none as finding against, which mints a gap. A drawn run keeps every outcome it drew.
func TestAForcedRunDrawsNoReadingThatOwesALaterMint(t *testing.T) {
	for _, c := range []struct {
		name  string
		force func(*runner)
	}{
		{"forced UNVERIFIED", func(r *runner) { r.forceUnverified = true }},
		{"forced VERIFIED", func(r *runner) { r.forceVerified = true }},
		{"drawn", func(*runner) {}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := &runner{rng: newLockedRand(1)}
			c.force(r)
			for _, fresh := range []bool{true, false} {
				for _, drawn := range verifyOutcomes {
					got := r.citationReading(drawn, fresh)
					want := drawn
					switch {
					case (drawn == "refutes" || drawn == "absent") && !fresh:
						want = "supports"
					case drawn == "unreachable" && r.forcedVerdict():
						want = "supports"
					}
					if got != want {
						t.Errorf("fresh %v: drew %q, read %q, want %q", fresh, drawn, got, want)
					}
				}
			}
			owed := map[string]bool{}
			for range 2000 {
				owed[r.owedCitationReading()] = true
			}
			want := map[string]bool{"supports": true, "unreachable": !r.forcedVerdict()}
			for o, w := range want {
				if owed[o] != w {
					t.Errorf("an owed check read %q: %v, want %v (all: %v)", o, owed[o], w, owed)
				}
			}
		})
	}
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

// TestTheForcedFoundAgainstOwnerReadsNoRepeatCopyUnreachable holds foundAgainstSeed's bound at the two
// copies of the twice-held sentence: in the forced run, the owed check of either copy and a drawn
// unreachable reading of one read backing, and every other citation keeps its draw; a drawn run keeps
// every draw over the copies too, as does the forced run once the owner has made both readings it is
// held to. An unreachable reading of a copy has blue retire it, after which the chain's reading of the
// sentence reads nothing for the rest of the run.
func TestTheForcedFoundAgainstOwnerReadsNoRepeatCopyUnreachable(t *testing.T) {
	for _, c := range []struct {
		name         string
		forced, read bool
	}{{"forced found-against", true, false}, {"forced found-against, both readings made", true, true}, {"drawn", false, false}} {
		t.Run(c.name, func(t *testing.T) {
			r := &runner{rng: newLockedRand(1), forceFoundAgainst: c.forced, repeatedCites: [2]string{"c-1", "c-2"},
				foundAgainstChained: c.read, foundAgainstRepeatedRead: c.read}
			for _, anchorID := range []string{"c-1", "c-2", "c-3", ""} {
				copy := c.forced && !c.read && (anchorID == "c-1" || anchorID == "c-2")
				owed := map[string]bool{}
				for range 2000 {
					owed[r.owedReadingOf(anchorID)] = true
				}
				want := map[string]bool{"supports": true}
				if !copy {
					want["unreachable"] = true
				}
				if !maps.Equal(owed, want) {
					t.Errorf("%q: owed checks read %v, want %v", anchorID, owed, want)
				}
				for _, drawn := range verifyOutcomes {
					for _, fresh := range []bool{true, false} {
						want := r.citationReading(drawn, fresh)
						if copy && drawn == "unreachable" {
							want = "supports"
						}
						if got := r.citationReadingOf(anchorID, drawn, fresh); got != want {
							t.Errorf("%q: drew %q (fresh %v), read %q, want %q", anchorID, drawn, fresh, got, want)
						}
					}
				}
			}
		})
	}
}
