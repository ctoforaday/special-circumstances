package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A GAP ITS LENS CLOSED IS THE LENS'S, WHEN THE BENCH RULES IT AFTER. The archived wave-b run holds
// two gaps closed by red and then ruled closed by the bench on a docket motion that stood unruled
// (G-ede17021: close ev376, ruling ev432; G-77b3a089: ev574, ev609) — the shape universe-m25 had
// four of, whose outcome then said the bench closed them. Migrated, the board and the family name
// red as the closer and count none of them as the bench's, and the bench's word is the fate.
func TestAGapItsLensClosedBeforeTheBenchRuledIsTheLenss(t *testing.T) {
	_, dst := migrateArchive(t, "2026-10-12_is-91-prime-wave-b.tar.gz")
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	ruledAfter := []string{"G-ede17021", "G-77b3a089"}
	for _, id := range ruledAfter {
		g := fam.Gap(id)
		if g == nil {
			t.Fatalf("%s is not in the migrated family", id)
		}
		if g.Open || g.Closure == nil || g.BenchClosure == nil {
			t.Fatalf("%s: open=%v close=%v ruling=%v, want closed by both arms", id, g.Open, g.Closure != nil, g.BenchClosure != nil)
		}
		if g.ClosedByBench {
			t.Errorf("%s: its lens closed it before the bench ruled, and the family names the bench its closer", id)
		}
		if !g.BenchClosedLast {
			t.Errorf("%s: the bench ruled after its lens closed it, and the family says the bench's act is not the later", id)
		}
		if got, want := g.ClosureReason(), recordpb.Word(g.BenchClosure.GetDisposition()); got != want {
			t.Errorf("%s: fate %q, want the bench's later word %q", id, got, want)
		}
	}
	bj, err := record.BoardJSONOfRun(dst)
	if err != nil {
		t.Fatal(err)
	}
	byBench := 0
	for _, g := range bj.Closed {
		if g.ClosedByBench {
			byBench++
		}
		for _, id := range ruledAfter {
			if g.ID == id && g.ClosedByBench {
				t.Errorf("board: %s reads closed_by_bench", id)
			}
		}
	}
	if bj.Counts.ClosedByBench != byBench {
		t.Errorf("board: counts.closed_by_bench = %d, and %d closed gaps read closed_by_bench", bj.Counts.ClosedByBench, byBench)
	}
}
