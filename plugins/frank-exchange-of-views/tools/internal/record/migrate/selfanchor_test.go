package migrate_test

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A CLOSURE THAT NAMED ITS OWN GAP AS AN ANCHOR MIGRATES WITHOUT IT. m30 recorded three: G-97b44713
// named itself and its successor, G-968365ad itself and two proofs, G-dc877265 itself and the
// citation it re-read. The write now refuses a gap as its own anchor; migrate carries each closure
// with the gap's id taken out of its anchors and every other anchor kept, in order.
func TestASelfAnchoredClosureMigratesWithoutItsOwnGap(t *testing.T) {
	res, run := migrateArchive(t, "2026-10-19_is-91-prime-m30.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"G-97b44713": {"G-968365ad"},
		"G-968365ad": {"P-24a6af66", "P-7de24bab"},
		"G-dc877265": {"C-4f4b8cfa"},
	}
	for _, g := range fam.Gaps {
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
