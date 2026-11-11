package migrate

import (
	"database/sql"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A STORED PLAN'S READING BLOCKERS ARE CARRIED OR DROPPED (GAP 2). An archived `contradiction`
// blocker naming a reading of a citation becomes the reader's reading_debt on that citation; one
// naming a reading of a source its lens found, which names no citation, leaves the plan — the debt
// is keyed on a sentence the plan does not hold, the loss the plan's residue states — and the plan is
// re-keyed. A plan that permits no PASS and would be left holding no blocker is refused.
func TestAStoredPlansReadingBlockersAreCarriedOrDropped(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`CREATE TABLE "verify" ("event_id" INTEGER, "anchor" TEXT);
		INSERT INTO "verify" VALUES (20, 'C-1'), (21, NULL)`); err != nil {
		t.Fatal(err)
	}
	newID := map[int64]int64{10: 20, 11: 21}

	p := stubPlan(5)
	p.Blockers = []record.PlanBlocker{
		{Kind: archivedReadingBlocker, Subject: "10", Owner: "red-lens-evidence"},
		{Kind: archivedReadingBlocker, Subject: "11", Owner: "red-lens-logic"},
		{Kind: record.BlockerUnchecked, Subject: "C-2", Owner: "red-lens-evidence"},
	}
	pl := storedOf(p)
	dropped, err := TranslateReadingBlockers(pl, newID, db)
	if err != nil {
		t.Fatal(err)
	}
	if dropped != 1 {
		t.Errorf("dropped = %d, want the one found-source blocker counted", dropped)
	}
	got, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		t.Fatal(err)
	}
	want := []record.PlanBlocker{
		{Kind: record.BlockerReadingDebt, Subject: "C-1", Owner: "red-lens-evidence"},
		{Kind: record.BlockerUnchecked, Subject: "C-2", Owner: "red-lens-evidence"},
	}
	if !slices.Equal(got.Blockers, want) {
		t.Errorf("the migrated plan's blockers are %+v, want %+v", got.Blockers, want)
	}
	if k := got.WithKey().Key; got.Key != k {
		t.Errorf("the migrated plan's key is %q and its own is %q", got.Key, k)
	}

	only := stubPlan(5)
	only.Blockers = []record.PlanBlocker{{Kind: archivedReadingBlocker, Subject: "11", Owner: "red-lens-logic"}}
	_, err = TranslateReadingBlockers(storedOf(only), newID, db)
	if err == nil || !strings.Contains(err.Error(), "would hold no blocker while it permits no PASS") {
		t.Errorf("a plan left with no blocker and no PASS: %v, want the refusal naming it", err)
	}
}
