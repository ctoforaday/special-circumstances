package migrate_test

import (
	"encoding/json"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED RUN'S ACCESS DATES DO NOT COME FORWARD, and its citations print none (EventSchema 73).
// A source's accessed date is when its copy was fetched: the fetch index knew it, and run-archive keeps
// no cache. The archive's cite access_date was the cite's own clock and its verify access_date a date
// the seat typed; neither is that time, so a migrated cite holds no fetched_at and every one of its
// citations arrives without it.
func TestAnArchivedRunsAccessDatesAreNotCarried(t *testing.T) {
	const archive = "2026-11-11_eiffel-tower-height.tar.gz"
	src := openRO(t, recordtest.RequireArchive(t, archive))
	cites := countOf(t, src, `SELECT count(*) FROM "cite"`)
	dated := countOf(t, src, `SELECT count(*) FROM "cite" WHERE "access_date" IS NOT NULL`)
	readingsDated := countOf(t, src, `SELECT count(*) FROM "verify" WHERE "access_date" IS NOT NULL`)
	if cites != 44 || dated != 44 || readingsDated != 1 {
		t.Fatalf("the archive holds %d cites, %d with an access_date, and %d dated readings; this test was written against 44, 44 and 1", cites, dated, readingsDated)
	}

	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	ddb := openRO(t, dst.Dir())
	if got := countOf(t, ddb, `SELECT count(*) FROM "cite"`); got != cites {
		t.Errorf("%d cites came forward of %d", got, cites)
	}
	if got := countOf(t, ddb, `SELECT count(*) FROM "cite" WHERE "fetched_at" IS NOT NULL`); got != 0 {
		t.Errorf("%d migrated cites carry a fetched_at the archive never recorded", got)
	}

	b, err := record.EvidenceJSONBytes(dst)
	if err != nil {
		t.Fatal(err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatal(err)
	}
	if len(ev.Sources) == 0 {
		t.Fatal("show evidence lists no source, so none was checked")
	}
	for _, s := range ev.Sources {
		if s.FetchedAt != "" || record.AccessedOn(s.FetchedAt) != "" {
			t.Errorf("source %s shows fetched_at %q", s.Anchor, s.FetchedAt)
		}
	}
}
