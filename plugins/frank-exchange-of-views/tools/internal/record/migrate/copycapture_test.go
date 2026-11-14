package migrate_test

import (
	"encoding/json"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// AN ARCHIVED RUN'S ARCHIVE COPIES COME FORWARD WITHOUT A CAPTURE TIME (EventSchema 77). When the
// web archive took a copy is a fact the fetch index held for the run that fetched it, and run-archive
// keeps no cache; the archive's cites hold the snapshot's url and nothing that states its time. A
// migrated cite of an archive copy carries no copy_captured_at, and `show evidence` and evidence.md
// show none: that is the residue.
func TestAnArchivedRunsArchiveCopiesCarryNoCaptureTime(t *testing.T) {
	const archive = "2026-11-11_eiffel-tower-height.tar.gz"
	src := openRO(t, recordtest.RequireArchive(t, archive))
	cites := countOf(t, src, `SELECT count(*) FROM "cite"`)
	archived := countOf(t, src, `SELECT count(*) FROM "cite" WHERE "copy_url" LIKE 'https://web.archive.org/%'`)
	if cites != 44 || archived != 12 {
		t.Fatalf("the archive holds %d cites, %d of them of a web archive copy; this test was written against 44 and 12", cites, archived)
	}

	res, dst := migrateArchive(t, archive)
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	ddb := openRO(t, dst.Dir())
	if got := countOf(t, ddb, `SELECT count(*) FROM "cite" WHERE "copy_url" LIKE 'https://web.archive.org/%'`); got != archived {
		t.Errorf("%d cites of a web archive copy came forward of %d", got, archived)
	}
	if got := countOf(t, ddb, `SELECT count(*) FROM "cite" WHERE "copy_captured_at" IS NOT NULL`); got != 0 {
		t.Errorf("%d migrated cites carry a copy_captured_at the archive never recorded", got)
	}

	b, err := record.EvidenceJSONBytes(dst)
	if err != nil {
		t.Fatal(err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal(b, &ev); err != nil {
		t.Fatal(err)
	}
	for _, s := range ev.Sources {
		if s.CopyCapturedAt != "" {
			t.Errorf("source %s shows copy_captured_at %q", s.Anchor, s.CopyCapturedAt)
		}
	}
}
