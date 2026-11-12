package fuzz

import (
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

// TestAForcedUnverifiedRunTypesOnlyMaterialSeverities: every severity a forced-UNVERIFIED run's drives
// type — a drawn mint's, a fixed one's, a regrade's — is a material grade, and a regrade still moves
// the axis. On other axes, and in a drawn run, the grade is any grade.
func TestAForcedUnverifiedRunTypesOnlyMaterialSeverities(t *testing.T) {
	forced := &runner{rng: newLockedRand(1), forceUnverified: true}
	drawn := &runner{rng: newLockedRand(1)}
	seen := map[string]bool{}
	for range 500 {
		for name, g := range map[string]string{
			"a drawn mint":             forced.severityGrade(),
			"a fixed mint":             forced.severityOr("low"),
			"a regrade":                forced.gradeOn("severity"),
			"a regrade off the medium": forced.gradeOnOtherThan("severity", "medium"),
		} {
			if !slices.Contains(materialGrades, g) {
				t.Fatalf("%s typed severity %q, which is not material", name, g)
			}
			if name == "a regrade off the medium" && g == "medium" {
				t.Fatal("a regrade typed the grade it moves from")
			}
		}
		seen[forced.gradeOn("impact")] = true
		seen[drawn.severityGrade()] = true
	}
	if drawn.severityOr("low") != "low" {
		t.Error("a drawn run's fixed severity is not the fixed one")
	}
	for _, g := range grades {
		if !seen[g] {
			t.Errorf("no draw typed %q on another axis or in a drawn run: those keep every grade", g)
		}
	}
}
