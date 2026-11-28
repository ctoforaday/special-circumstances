package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A RECORD FROM AN OLDER EPOCH IS REFUSED AT THE OPEN, AND MIGRATE IS THE WHOLE WAY OUT TO THIS
// BINARY'S VIEWS. The archived m20 run states epoch 24 and its `gap` view takes the EARLIEST closing
// act of each arm; this binary's takes the LAST and names it (`close_event`, `bench_close_event`),
// with the arm whose act ended the gap (`closed_by_bench`) and the arm that acted last
// (`bench_closed_last`). The fresh run states this binary's epoch, holds this binary's views, and
// reads.
func TestARecordWithAStaleGapViewIsRefusedAndMigrates(t *testing.T) {
	const archive = "2026-10-11_is-91-prime-m20.tar.gz"
	src := recordtest.RequireArchive(t, archive)
	epoch, stated := statedEpoch(t, src)
	if !stated {
		t.Fatalf("%s states no epoch", archive)
	}
	if epoch == record.EventSchema {
		t.Fatalf("%s states this binary's epoch (%d), so the open trusts it and nothing here is refused", archive, epoch)
	}
	run, err := record.NewRun(src)
	if err != nil {
		t.Fatal(err)
	}
	_, err = record.FamilyOf(run)
	if err == nil {
		t.Fatalf("%s was read unmigrated at epoch %d", archive, epoch)
	}
	if !strings.Contains(err.Error(), "`feov-migrate --from <runDir> --to <freshDir>`") {
		t.Errorf("the refusal does not name the way out: %v", err)
	}

	assertMigratedGapView(t, archive)
	// m20's closes are lost acts (no lens sitting of it holds a delivery), and the successor's mint
	// that closed its one other gap is that gap's Amend (Translation 7), so it holds no closed gap; the
	// closing-act columns are read on the gap-audit run, an older epoch too, whose lenses' closes
	// stand on their deliveries.
	if named := assertMigratedGapView(t, gapAuditArchive); named == 0 {
		t.Error("the gap-audit run names no closing act on its gap view, so nothing here read the closure columns")
	}
}

// assertMigratedGapView migrates the archive and holds its gap view to the family: the record states
// this binary's epoch and reads; every closed gap carries its reason and epoch; and the view names a
// closing act on every closed gap. It returns how many gaps the view names a closing act on; its
// caller holds that above zero where the archive closes a gap.
func assertMigratedGapView(t *testing.T, archive string) int {
	t.Helper()
	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if got, _ := statedEpoch(t, dst.Dir()); got != record.EventSchema {
		t.Errorf("%s: the migrated record states epoch %d, this binary writes %d", archive, got, record.EventSchema)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatalf("%s: the migrated record is refused: %v", archive, err)
	}
	closed := 0
	for _, g := range fam.Gaps {
		if g.Open {
			continue
		}
		closed++
		if !g.HasClosed || g.ClosureReason() == "" || g.ClosedEpoch == 0 {
			t.Errorf("%s: gap %s reads closed with has_closed %v, reason %q, closed_epoch %d", archive, g.ID, g.HasClosed, g.ClosureReason(), g.ClosedEpoch)
		}
	}
	db := openRO(t, dst.Dir())
	var named int
	if err := db.QueryRow(`SELECT count(*) FROM "gap" WHERE ("close_event" IS NOT NULL OR "bench_close_event" IS NOT NULL) AND "bench_closed_last" IS NOT NULL`).Scan(&named); err != nil {
		t.Fatalf("%s: the migrated record's gap view does not name its closing acts: %v", archive, err)
	}
	if named != closed {
		t.Errorf("%s: the gap view names a closing act on %d gaps and the family reads %d closed", archive, named, closed)
	}
	return named
}
