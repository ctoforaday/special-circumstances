package migrate_test

import (
	"strings"
	"testing"

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

// AN ACT REFUSED ONLY BECAUSE IT DEPENDS ON A LOST ACT IS THAT SAME LOSS. m32's lost close of
// G-11c8fa75 leaves it open, so the logic lens's mint superseding it (source 448) is refused, and
// every act on the gap that mint created (dispatches 457, 458, 488, 489 and regrade 470) names a gap
// the record never holds. Each is listed with the act it depended on.
func TestM32sActsOnALostCloseAreLostWithIt(t *testing.T) {
	m, _ := migrateArchive(t, "2026-10-24_is-91-prime-m32.tar.gz")
	if len(m.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(m.Refusals), m.Refusals[0])
	}
	byID := map[int64]migration.Refusal{}
	for _, a := range m.LostActs {
		byID[a.OldID] = a
	}
	for _, c := range []struct {
		id   int64
		word string
		on   string
	}{{448, "mint", "G-11c8fa75"}, {457, "dispatch", "G-cc9fe904"}, {458, "dispatch", "G-cc9fe904"},
		{470, "regrade", "G-cc9fe904"}, {488, "dispatch", "G-cc9fe904"}, {489, "dispatch", "G-cc9fe904"}} {
		a, ok := byID[c.id]
		switch {
		case !ok:
			t.Errorf("source %d (%s) is not listed as lost", c.id, c.word)
		case a.Word != c.word || !strings.Contains(a.Err, "depends on gap "+c.on) || !strings.Contains(a.Err, "source event"):
			t.Errorf("source %d is listed as %s: %q; want a %s depending on %s, naming the lost act", c.id, a.Word, a.Err, c.word, c.on)
		}
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
