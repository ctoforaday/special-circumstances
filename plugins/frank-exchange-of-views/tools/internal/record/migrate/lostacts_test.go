package migrate_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
)

// AN ARCHIVED ACT THE DELIVERED-TEXT RULE REFUSES IS AN ACCEPTED LOSS, LISTED BY ACT (gblock
// 2026-11-14). The gap-audit run holds three closes and a spot-check whose sittings hold no recorded
// delivery: each is written as no event, listed in lost_acts with the rule's refusal, its word's
// reason in the accepted losses, and nothing else is refused.
func TestTheGapAuditRunLosesTheActsItsDeliveriesDoNotHold(t *testing.T) {
	m, _ := migrateArchive(t, gapAuditArchive)
	if len(m.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(m.Refusals), m.Refusals[0])
	}
	got := map[string]int{}
	for _, a := range m.LostActs {
		if !strings.Contains(a.Err, "stands only on report text delivered to you") {
			t.Errorf("lost act %d (%s) is listed for %q, not the delivered-text rule", a.OldID, a.Word, a.Err)
		}
		got[a.Word]++
	}
	want := map[string]int{"close": 3, "spot_check": 1}
	for w, n := range want {
		if got[w] != n {
			t.Errorf("lost %s acts = %d, want %d (all: %v)", w, got[w], n, got)
		}
	}
	assertLossesAccounted(t, m)
}

// A LINK ACROSS SEATS DROPS, SO A SUCCESSOR OF A GAP A LOST CLOSE LEFT OPEN STANDS AS ITS OWN GAP.
// m32's lost close of the computation lens's G-11c8fa75 (source 344) leaves it open; the logic
// lens's mint naming it as ancestor (source 448) is another seat's, so the link drops and the mint
// stands as G-cc9fe904, its own gap. Nothing on it depends on the lost close: the mint, its
// dispatches (457, 458, 488, 489) and its regrade (470) are written, and the run loses only the
// closes and the spot-check the delivered-text rule refuses, and its four claim retires that carry
// no anchor (P12b: a retire takes anchors and nothing else) — 284, 285 and 476 named none, and 377
// named only a finding's, which leaves every list it stood in.
func TestM32sCrossSeatSuccessorOfALostCloseStandsAsItsOwnGap(t *testing.T) {
	m, run := migrateArchive(t, "2026-10-24_is-91-prime-m32.tar.gz")
	if len(m.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(m.Refusals), m.Refusals[0])
	}
	lost := map[int64]bool{}
	var retires []int64
	for _, a := range m.LostActs {
		lost[a.OldID] = true
		switch {
		case a.Word == "retire" && strings.HasPrefix(a.Err, "this claim retire carries no anchor"):
			retires = append(retires, a.OldID)
		case a.Word != "close" && a.Word != "spot_check":
			t.Errorf("lost act %d is a %s (%q); only closes, spot-checks and anchorless claim retires are lost", a.OldID, a.Word, a.Err)
		}
	}
	if !slices.Equal(retires, []int64{284, 285, 377, 476}) {
		t.Errorf("the anchorless claim retires lost are source %v, want [284 285 377 476]", retires)
	}
	if !lost[344] {
		t.Errorf("source 344 (the close of G-11c8fa75) is not listed as lost")
	}
	for _, id := range []int64{448, 457, 458, 470, 488, 489} {
		if lost[id] {
			t.Errorf("source %d is listed as lost; the successor's link drops across seats and its acts stand", id)
		}
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	if fam.Gap("G-cc9fe904") == nil {
		t.Fatal("G-cc9fe904 is not on the migrated record")
	}
	assertLossesAccounted(t, m)
}

// assertLossesAccounted holds the manifest's counts to its list: every lost act's word has its reason
// among the accepted losses, and is one fewer event written than read.
func assertLossesAccounted(t *testing.T, m *migration.Manifest) {
	t.Helper()
	lost := map[string]int{}
	for _, a := range m.LostActs {
		lost[a.Word]++
		if m.Accepted[a.Word] == "" {
			t.Errorf("lost %s act %d has no accepted-loss reason for its word", a.Word, a.OldID)
		}
	}
	for w, n := range lost {
		if m.In[w]-m.Out[w] < n {
			t.Errorf("%s: %d read, %d written, but %d listed lost", w, m.In[w], m.Out[w], n)
		}
	}
}
