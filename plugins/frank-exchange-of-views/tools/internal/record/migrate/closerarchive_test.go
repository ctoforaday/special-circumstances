package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A GAP ITS LENS CLOSED IS THE LENS'S, WHEN THE BENCH RULES IT AFTER. The archived wave-b run holds
// two gaps closed by red and then ruled closed by the bench on a docket the chair filed by hand
// (G-ede17021: close ev376, ruling ev432; G-77b3a089: ev574, ev609) — the shape universe-m25 had
// four of, whose outcome then said the bench closed them. A docket no dispatch filed is written as no
// event with its ruling (EventSchema 80, O2), so migrated, the board and the family name red as the
// closer, hold no bench closure for either, and the lens's word is the fate. The
// archive's pairs are lost acts; the fixture's G2 is the same shape on delivered sittings.
func TestAGapItsLensClosedBeforeTheBenchRuledIsTheLenss(t *testing.T) {
	// A close or ruling the manifest lists as lost (its seat's sitting holds no delivery) is written
	// as no event: a pair with a lost arm leaves its gap open if both are lost, and is not this
	// shape; a pair with neither lost is held to it. The fixture's G2 is the shape on delivered
	// sittings, so it is held whatever the archive's pairs lose.
	res, dst := migrateArchive(t, "2026-10-12_is-91-prime-wave-b.tar.gz")
	lost := migrate.LostIDs(res)
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	var ruledAfter []string
	for _, c := range []struct {
		gap          string
		close, ruled int64
	}{{"G-ede17021", 376, 432}, {"G-77b3a089", 574, 609}} {
		switch {
		case !lost[c.close] && !lost[c.ruled]:
			ruledAfter = append(ruledAfter, c.gap)
		case lost[c.close] && lost[c.ruled]:
			if g := fam.Gap(c.gap); g == nil || !g.Open {
				t.Errorf("%s: both its closing acts are lost, and the migrated family reads it %+v", c.gap, g)
			}
		}
	}
	assertLensClosedBeforeBench(t, dst, ruledAfter...)
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
		if g.Open || g.Closure == nil || g.BenchClosure != nil {
			t.Fatalf("%s: open=%v close=%v ruling=%v, want closed by its lens alone", id, g.Open, g.Closure != nil, g.BenchClosure != nil)
		}
		if g.ClosedByBench {
			t.Errorf("%s: its lens closed it, and the family names the bench its closer", id)
		}
		if got, want := g.ClosureReason(), recordpb.Word(g.Closure.GetClosureClass()); got != want {
			t.Errorf("%s: fate %q, want its lens's word %q", id, got, want)
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
