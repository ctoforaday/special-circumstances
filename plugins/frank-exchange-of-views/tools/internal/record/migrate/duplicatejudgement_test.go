package migrate_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A MINT'S distinct_from LIST IS NOT CARRIED, AND THE MINT IS. m28 holds six rows of it; every mint
// comes forward, nothing is refused, and the migrated record verifies.
func TestADistinctFromListIsNotCarried(t *testing.T) {
	const name = "2026-10-18_is-91-prime-m28.tar.gz"
	runDir := t.TempDir()
	untar(t, archivePath(t, name), runDir)
	db, err := sql.Open("sqlite", "file:"+filepath.Join(runDir, "records", "record.db")+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	var rows int
	err = db.QueryRow(`SELECT count(*) FROM "mint_distinct_from"`).Scan(&rows)
	db.Close()
	if err != nil {
		t.Fatalf("reading the source's distinct_from rows: %v", err)
	}
	// AN ASSERTION OVER NOTHING IS NOT AN ASSERTION: a source with no rows exercises no drop.
	if rows == 0 {
		t.Fatalf("%s holds no distinct_from row, so the drop was never exercised — pick an archive that has them", name)
	}
	res, dst := migrateArchive(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	if res.In["mint"] == 0 || res.In["mint"] != res.Out["mint"] {
		t.Errorf("mint census: in=%d out=%d — dropping the list dropped a mint", res.In["mint"], res.Out["mint"])
	}
	assertVerifies(t, dst)
}

// A CREDIT COMES FORWARD AS NO EVENT. No archive holds one, so the source is built here: a lens's
// register and its credit of an open gap with a finding. The credit is counted in and not out, and
// nothing is refused.
func TestACreditComesForwardAsNoEvent(t *testing.T) {
	old := newOldRecord(t)
	for _, ddl := range []string{
		`CREATE TABLE "credit" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "found_by" TEXT)`,
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('credit', 'former vocabulary')`,
	} {
		if _, err := old.db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	id := old.event(t, "red-lens-evidence", 1, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version"`, id, "abc1234")
	id = old.event(t, "red-lens-evidence", 1, ts2, "credit")
	old.row(t, "credit", `"event_id", "gap_id", "found_by"`, id, "G-0000beef", "F-0000f00d")
	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals on a fixture built to land: %+v", res.Refusals)
	}
	if res.In["credit"] != 1 || res.Out["credit"] != 0 {
		t.Errorf("credit census: in=%d out=%d, want 1 and 0", res.In["credit"], res.Out["credit"])
	}
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if e.GetSeatId() == "red-lens-evidence" && e.GetType().String() != "EVENT_TYPE_REGISTER" {
			t.Errorf("the lens's credit came forward as %s", e.GetType())
		}
	}
}
