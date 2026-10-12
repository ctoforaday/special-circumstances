package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A RECORD AT AN EPOCH THIS BINARY HAS WRITTEN CAN STILL HOLD A VIEW IT NO LONGER DEFINES THAT WAY,
// because a view-only change does not move the epoch. The archived m20 run states epoch 24 and its
// `gap` view takes the EARLIEST closing act of each arm; this binary's takes the LAST and names it
// (`close_event`, `bench_close_event`, `closed_by_bench`). Such a record is refused at the open by
// the view it holds, and migrate is the whole way out: the fresh run states this binary's epoch,
// holds this binary's views, and reads.
func TestARecordWithAStaleGapViewIsRefusedAndMigrates(t *testing.T) {
	const archive = "2026-10-11_is-91-prime-m20.tar.gz"
	src := recordtest.RequireArchive(t, archive)
	epoch, stated := statedEpoch(t, src)
	if !stated {
		t.Fatalf("%s states no epoch", archive)
	}
	run, err := record.NewRun(src)
	if err != nil {
		t.Fatal(err)
	}
	_, err = record.FamilyOf(run)
	if err == nil {
		t.Fatalf("%s was read unmigrated: its gap view is the one its creating binary wrote", archive)
	}
	if !strings.Contains(err.Error(), "`--seat-id operator migrate --from <runDir> --to <freshDir>`") {
		t.Errorf("the refusal does not name the way out: %v", err)
	}
	// While the archive's epoch is this binary's, the view is the ONLY thing the open can refuse on.
	if epoch == record.EventSchema && !strings.Contains(err.Error(), `"gap"`) {
		t.Errorf("the record states this binary's epoch (%d) and the refusal does not name the gap view: %v", epoch, err)
	}

	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if got, _ := statedEpoch(t, dst.Dir()); got != record.EventSchema {
		t.Errorf("the migrated record states epoch %d, this binary writes %d", got, record.EventSchema)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatalf("the migrated record is refused: %v", err)
	}
	closed := 0
	for _, g := range fam.Gaps {
		if g.Open {
			continue
		}
		closed++
		if !g.HasClosed || g.ClosureReason() == "" || g.ClosedEpoch == 0 {
			t.Errorf("gap %s reads closed with has_closed %v, reason %q, closed_epoch %d", g.ID, g.HasClosed, g.ClosureReason(), g.ClosedEpoch)
		}
	}
	if closed == 0 {
		t.Error("the migrated record holds no closed gap, so nothing here read the gap view's closure")
	}
	db := openRO(t, dst.Dir())
	var named int
	if err := db.QueryRow(`SELECT count(*) FROM "gap" WHERE "close_event" IS NOT NULL OR "bench_close_event" IS NOT NULL`).Scan(&named); err != nil {
		t.Fatalf("the migrated record's gap view does not name its closing acts: %v", err)
	}
	if named != closed {
		t.Errorf("the gap view names a closing act on %d gaps and the family reads %d closed", named, closed)
	}
}
