package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE ARCHIVE THAT HELD THE OLD WORD. is-7-prime-run2's bench deferred gap R1-1 under the
// disposition's earlier spelling, "carried", through the shard era's opinion. An opinion is written as
// no event (EventSchema 80), so its seven opinions are the manifest's stated residue, no docket ruling
// is on the migrated record, R1-1 reads open as the opinion found it, and verify holds.
// TestAnIdentityRowHoldingTheOldDispositionMigrates holds the old word's respelling.
func TestIs7PrimeRun2Archive(t *testing.T) {
	res, dst := migrateArchive(t, "2026-08-22_is-7-prime-run2.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	if len(res.Dockets.Opinions) != 7 || res.In["opinion"] != 7 {
		t.Fatalf("%d opinion(s) in, %v stated as no event; want 7 and 7", res.In["opinion"], res.Dockets.Opinions)
	}
	gap := res.IDs["R1-1"]
	if gap == "" {
		t.Fatalf("R1-1 is not in the id table: %v", res.IDs)
	}
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range fam.Events {
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetDocket() != nil {
			t.Errorf("the migrated record holds docket ruling %s", e.GetKey())
		}
	}
	if g := fam.Gap(gap); g == nil || !g.Open {
		t.Errorf("R1-1 (%s) does not read open with its opinion gone", gap)
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
	bodies, err := migrate.Entries().Translate(old)
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
