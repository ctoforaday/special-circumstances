package migrate_test

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CLOSURE THAT NAMED ITS OWN GAP AS AN ANCHOR MIGRATES WITHOUT IT. m30 recorded three: G-97b44713
// named itself and its successor, G-968365ad itself and two proofs, G-dc877265 itself and the
// citation it re-read. The write now refuses a gap as its own anchor; migrate carries each closure
// with the gap's id taken out of its anchors and every other anchor kept, in order. G-97b44713's is
// left out: its successor's mint closed it (TestAGapSupersededWhileOpenIsClosedByItsSuccessorsMint).
// A close stands only on a delivery in its lens's sitting, which m30's sittings do not hold.
func TestASelfAnchoredClosureMigratesWithoutItsOwnGap(t *testing.T) {
	// The translation is held on m30's closes that stand and, whatever m30's lens sittings lose, on a
	// delivered source's close of G2, which named G2 and G1.
	res, run := migrateArchive(t, "2026-10-19_is-91-prime-m30.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	// A close the manifest lists as lost carries nothing to the migrated record; every other one is
	// held to its anchors.
	lost := migrate.LostCloseGaps(t, "2026-10-19_is-91-prime-m30.tar.gz", res)
	want := map[string][]string{}
	for g, anchors := range map[string][]string{
		"G-968365ad": {"K-24a6af66", "K-7de24bab"},
		"G-dc877265": {"C-4f4b8cfa"},
	} {
		if !lost[g] {
			want[g] = anchors
		}
	}
	assertClosureAnchors(t, run, want)
	fres, _, dir := migrate.MigratedDeliveredBench(t)
	assertClosureAnchors(t, runtest.Open(t, dir), map[string][]string{fres.IDs["G2"]: {fres.IDs["G1"]}})
}

// assertClosureAnchors holds each named gap's closure on run to its anchors, in order, and every
// closure on run to carrying no anchor that is its own gap.
func assertClosureAnchors(t *testing.T, run record.Run, want map[string][]string) {
	t.Helper()
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range fam.Gaps {
		if slices.Contains(g.Closure.GetAnchors(), g.ID) {
			t.Errorf("%s's closure carries its own gap among its anchors %v", g.ID, g.Closure.GetAnchors())
		}
		w, ok := want[g.ID]
		if !ok {
			continue
		}
		delete(want, g.ID)
		if got := g.Closure.GetAnchors(); !slices.Equal(got, w) {
			t.Errorf("%s's closure carries anchors %v, want %v", g.ID, got, w)
		}
	}
	for id := range want {
		t.Errorf("the migrated board holds no closure of %s", id)
	}
}
