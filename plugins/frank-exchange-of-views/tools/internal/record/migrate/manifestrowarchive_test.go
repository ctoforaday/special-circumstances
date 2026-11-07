package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A REPAIR'S RECEIPT LEAVES AN ARCHIVED RUN, AND THE REPAIR STAYS (EventSchema 65). m19 and the mersenne smoke hold
// blue's prose receipts, each naming a gap blue's edit answered. Migrated, every receipt is gone —
// counted in and none out, with no refusal — while every edit arrives, still naming the gap it
// answers, and the record holds no table for the receipt at all.
func TestARepairsReceiptLeavesAnArchivedRun(t *testing.T) {
	for _, c := range []struct {
		archive  string
		receipts int
	}{
		{archive: "2026-10-10_is-91-prime-m19.tar.gz", receipts: 6},
		{archive: "2026-10-30_mersenne-31-prime-sum.tar.gz", receipts: 22},
	} {
		t.Run(c.archive, func(t *testing.T) {
			src := openRO(t, recordtest.RequireArchive(t, c.archive))
			if n := countOf(t, src, `SELECT count(*) FROM "manifest_row"`); n != c.receipts {
				t.Fatalf("the archive holds %d receipt(s), this test was written against %d", n, c.receipts)
			}
			edits := countOf(t, src, `SELECT count(*) FROM "blue_edit"`)
			answering := countOf(t, src, `SELECT count(*) FROM "blue_edit" WHERE "answers" IS NOT NULL AND "answers" != ''`)

			res, dst := migrateArchive(t, c.archive)
			if len(res.Refusals) != 0 {
				t.Fatalf("refusals: %+v", res.Refusals)
			}
			if res.In["manifest_row"] != c.receipts || res.Out["manifest_row"] != 0 {
				t.Errorf("receipts: %d in, %d out; want %d in and none out", res.In["manifest_row"], res.Out["manifest_row"], c.receipts)
			}
			ddb := openRO(t, dst.Dir())
			if n := countOf(t, ddb, `SELECT count(*) FROM sqlite_master WHERE "name" = 'manifest_row'`); n != 0 {
				t.Errorf("the migrated record holds a receipt table")
			}
			if n := countOf(t, ddb, `SELECT count(*) FROM "blue_edit"`); n != edits {
				t.Errorf("%d archived edit(s) arrived as %d", edits, n)
			}
			if n := countOf(t, ddb, `SELECT count(*) FROM "blue_edit" WHERE "answers" IS NOT NULL AND "answers" != ''`); n != answering {
				t.Errorf("%d archived edit(s) answering a gap arrived as %d", answering, n)
			}
		})
	}
}
