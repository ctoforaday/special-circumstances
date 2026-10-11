package migrate_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE MIGRATED M19 OWES WHAT THE ARCHIVED ONE DOES. m19 is the first archive written at the epoch
// this binary reads, so its migration is a same-epoch one: nothing is refused, and the reopened
// citations and proof read off the migrated record are the three the archive holds
// (TestM19ReopenedEvidenceHoldsThePass). The anchors are compared and not the event ids: a
// migration may renumber rows, and that test's table is where a renumbering is recorded.
func TestTheMigratedM19HoldsTheSameReopenedAnchors(t *testing.T) {
	res, dst := migrateArchive(t, "2026-10-10_is-91-prime-m19.tar.gz")
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused migrating m19; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	raw, err := record.EvidenceJSONBytes(dst)
	if err != nil {
		t.Fatal(err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal(raw, &ev); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range ev.Reopened {
		got = append(got, r.Anchor)
		t.Logf("%s: edit %d, last check %d", r.Anchor, r.Edit, r.LastCheck)
	}
	want := []string{"C-706fda63", "C-4dae2507", "P-cd9f4c5a"}
	if !slices.Equal(got, want) {
		t.Errorf("the migrated m19's reopened anchors = %q, want %q", got, want)
	}
}
