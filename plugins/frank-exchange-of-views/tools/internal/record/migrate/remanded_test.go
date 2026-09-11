package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE ARCHIVE THAT HOLDS THE OLD WORD. is-7-prime-run2's bench deferred gap R1-1 under the
// disposition's earlier spelling, "carried", through the shard era's opinion — the one archived
// deferral on the record. Migrated, it must read as the current word, and verify must hold.
func TestIs7PrimeRun2Archive(t *testing.T) {
	res, dst := migrateArchive(t, "2026-08-22_is-7-prime-run2.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	gap := res.GapIDs["R1-1"]
	if gap == "" {
		t.Fatalf("R1-1 is not in the gap-id table: %v", res.GapIDs)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	gapOf := map[string]string{}
	for _, e := range fam.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
			gapOf[m.GetMotionId()] = m.GetDocket().GetGapId()
		}
	}
	remanded := 0
	for _, e := range fam.Events {
		r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || r.GetDocket() == nil {
			continue
		}
		if r.GetDocket().GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
			remanded++
			if gapOf[r.GetMotionId()] != gap {
				t.Errorf("the remanded ruling answers %s, which docketed %q; want R1-1's %s", r.GetMotionId(), gapOf[r.GetMotionId()], gap)
			}
		}
	}
	if remanded != 1 {
		t.Errorf("%d remanded ruling(s) on the migrated record, want the archive's one", remanded)
	}
	assertVerifies(t, dst)
}

// The SQLite era's identity path takes the same respelling. No archived SQLite-era record holds
// a deferral (quadratic-formula's bench repaired and owed elsewhere; the is-91-prime runs
// repaired), so the row is synthetic: a docket ruling as that era's table stored it.
func TestAnIdentityRowHoldingTheOldDispositionMigrates(t *testing.T) {
	old := migrate.OldEvent{
		ID: 11, SeatID: "judge-r1", Round: 1, TS: ts1, Word: "motion_rule",
		Fields: map[string]any{"motion_id": "M1", "subject": "docket", "ruling_case": "docket"},
		Arms:   map[string]map[string]any{"docket": {"disposition": "carried", "principle": "correctness first"}},
	}
	dst := runtest.New(t, recordtest.TmpRun(t))
	bodies, err := (migrate.Registry{}).Translate(old, dst)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := bodies[0].(*recordpb.MotionRule)
	if !ok {
		t.Fatalf("body is %T, want MotionRule", bodies[0])
	}
	if got := r.GetDocket().GetDisposition(); got != recordpb.Disposition_DISPOSITION_REMANDED {
		t.Errorf("disposition %v, want REMANDED", got)
	}
}
