package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// twinDocket stages a gap with two docket motions standing on it (the record admits the second while
// the first stands) and the bench sitting the docket the chair convened it for, then build's acts.
func twinDocket(t *testing.T, build func(b *stage)) Run {
	t.Helper()
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).
		mint(evLens, "G1", "low").
		docketMotion("blue-respond", "M1", "G1").docketMotion("blue-respond", "M2", "G1").
		dispatch(2, benchSeat, "G1").benchRegister(recordpb.Occasion_OCCASION_DOCKET)
	build(b)
	return b.seed()
}

// benchItemsOn is the bench's blocking work-list items that name motion.
func benchItemsOn(t *testing.T, run Run, motion string) []Item {
	t.Helper()
	w, err := WorkOfSeat(run, "bench", benchSeat)
	if err != nil {
		t.Fatal(err)
	}
	var out []Item
	for _, it := range w.Sitting.Open {
		if it.Blocks && strings.HasPrefix(it.What, "motion "+motion+" ") {
			out = append(out, it)
		}
	}
	return out
}

// THE BENCH'S DOCKET SITTING OWES EVERY DOCKET MOTION STANDING ON A GAP IT IS CONVENED ON, AND ITS
// ENVELOPE IS REFUSED WHILE ONE STANDS (gblock's ruling, 2026-10-24). The plan keys the bench's
// sitting for a gap on its newest unruled filing, so a bench that rules one of two twins has sat for
// the other and no plan readies it again (#1557).
func TestTheBenchsDocketSittingIsHeldByEveryTwinMotion(t *testing.T) {
	t.Run("both twins are on the work list, each holding the sitting", func(t *testing.T) {
		run := twinDocket(t, func(*stage) {})
		for _, m := range []string{"M1", "M2"} {
			items := benchItemsOn(t, run, m)
			if len(items) != 1 {
				t.Fatalf("the bench's list holds %d blocking item(s) on %s, want 1", len(items), m)
			}
			if !strings.Contains(items[0].What, "your envelope is refused while it stands") || !strings.Contains(items[0].What, "G1") {
				t.Errorf("%s's item does not say it holds the sitting on G1: %q", m, items[0].What)
			}
		}
	})
	t.Run("ending with one twin unruled is refused, naming it", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.add(benchSeat, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT))
		})
		refusal, err := BenchEnvelopeRefusal(run, benchSeat)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(refusal, "M2 (on G1)") {
			t.Fatalf("the refusal does not name the standing twin M2: %q", refusal)
		}
		if strings.Contains(refusal, "M1") {
			t.Errorf("the refusal names M1, which is ruled: %q", refusal)
		}
		if !strings.Contains(refusal, "1 docket motion(s)") {
			t.Errorf("the refusal does not count the one standing motion: %q", refusal)
		}
	})
	t.Run("both unruled, the refusal names each", func(t *testing.T) {
		refusal, err := BenchEnvelopeRefusal(twinDocket(t, func(*stage) {}), benchSeat)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(refusal, "M1 (on G1), M2 (on G1)") {
			t.Errorf("the refusal does not name both twins: %q", refusal)
		}
	})
	t.Run("both ruled, the sitting ends", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.add(benchSeat, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).
				add(benchSeat, docketRuling("M2", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT))
		})
		if refusal, err := BenchEnvelopeRefusal(run, benchSeat); err != nil || refusal != "" {
			t.Errorf("both twins ruled, the envelope is refused (%v): %q", err, refusal)
		}
	})
	t.Run("a docket on a gap the sitting is not convened on does not hold it", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.add(benchSeat, docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).
				add(benchSeat, docketRuling("M2", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).
				mint(evLens, "G2", "low").docketMotion("blue-respond", "M3", "G2")
		})
		if refusal, err := BenchEnvelopeRefusal(run, benchSeat); err != nil || refusal != "" {
			t.Errorf("a motion on G2, which no row convened this sitting on, refuses the envelope (%v): %q", err, refusal)
		}
		items := benchItemsOn(t, run, "M3")
		if len(items) != 1 || strings.Contains(items[0].What, "envelope") {
			t.Errorf("M3 should hold the outcome and not the sitting: %+v", items)
		}
	})
	// The lens's row on G1 between the two sittings is no row of the bench's.
	t.Run("a later docket sitting is held by its own gaps, not an earlier sitting's", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) {
			b.mint(evLens, "G2", "low").docketMotion("blue-respond", "M3", "G2").
				dispatch(3, evLens, "G1").dispatch(3, benchSeat, "G2").benchRegister(recordpb.Occasion_OCCASION_DOCKET)
		})
		refusal, err := BenchEnvelopeRefusal(run, benchSeat)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(refusal, "M3 (on G2)") || strings.Contains(refusal, "G1") {
			t.Errorf("the sitting convened on G2 is held by %q, want M3 on G2 alone", refusal)
		}
	})
	t.Run("a petition sitting is not held by the docket", func(t *testing.T) {
		run := twinDocket(t, func(b *stage) { b.benchRegister(recordpb.Occasion_OCCASION_PETITION) })
		if refusal, err := BenchEnvelopeRefusal(run, benchSeat); err != nil || refusal != "" {
			t.Errorf("a petition sitting's envelope is refused over the docket (%v): %q", err, refusal)
		}
	})
	t.Run("no other seat's envelope is refused", func(t *testing.T) {
		if refusal, err := BenchEnvelopeRefusal(twinDocket(t, func(*stage) {}), "blue-respond"); err != nil || refusal != "" {
			t.Errorf("blue's envelope is refused over the bench's docket (%v): %q", err, refusal)
		}
	})
}
