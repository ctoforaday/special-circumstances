package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A GAP ITS LENS CLOSED IS THE LENS'S, WHEN THE BENCH RULES IT AFTER. The archived wave-b run holds
// two gaps closed by red and then ruled closed by the bench on a docket motion that stood unruled
// (G-ede17021: close ev376, ruling ev432; G-77b3a089: ev574, ev609) — the shape universe-m25 had
// four of, whose outcome then said the bench closed them. Migrated, the board and the family name
// red as the closer and count none of them as the bench's, and the bench's word is the fate. The
// archive's pairs are lost acts; the fixture's G2 is the same shape on delivered sittings.
func TestAGapItsLensClosedBeforeTheBenchRuledIsTheLenss(t *testing.T) {
	// wave-b's two closes (ev376, ev574) and the bench's two rulings (ev432, ev609) stood on no
	// delivery the record holds: each is a lost act, so both gaps read open, and the shape is held on
	// a source whose lens and bench were delivered the report before they acted.
	res, dst := migrateArchive(t, "2026-10-12_is-91-prime-wave-b.tar.gz")
	lost := migrate.LostIDs(res)
	for _, id := range []int64{376, 432, 574, 609} {
		if !lost[id] {
			t.Errorf("wave-b source event %d is not listed lost; its gap's closing pair is held here only once it is", id)
		}
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"G-ede17021", "G-77b3a089"} {
		if g := fam.Gap(id); g == nil || !g.Open {
			t.Errorf("%s: both its closing acts are lost, and the migrated family reads it %+v", id, g)
		}
	}
	fres, _, dir := migrate.MigratedDeliveredBench(t)
	assertLensClosedBeforeBench(t, runtest.Open(t, dir), fres.IDs["G2"])
}

// assertLensClosedBeforeBench holds each gap, closed by its lens and then ruled closed by the bench,
// to the lens as its closer and the bench's word as its fate, on the family and the board.
func assertLensClosedBeforeBench(t *testing.T, dst record.Run, ruledAfter ...string) {
	t.Helper()
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
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
