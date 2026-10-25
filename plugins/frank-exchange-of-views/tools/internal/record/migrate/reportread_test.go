package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// readFixture is an old record holding a lens's register, the base ingest and a report read whose head
// is head: no archive holds a report read yet, so the source is built here.
func readFixture(t *testing.T, head func(ingest int64) int64) (*oldRecord, int64) {
	t.Helper()
	old := newOldRecord(t)
	for _, ddl := range []string{
		`CREATE TABLE "base_ingest" ("event_id" INTEGER PRIMARY KEY, "text" TEXT)`,
		`CREATE TABLE "report_read" ("event_id" INTEGER PRIMARY KEY, "head" INTEGER)`,
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('base_ingest', 'current vocabulary')`,
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('report_read', 'current vocabulary')`,
	} {
		if _, err := old.db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	id := old.event(t, "red-lens-evidence", 1, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version"`, id, "abc1234")
	ingest := old.event(t, "blue-synthesize", 1, ts2, "base_ingest")
	old.row(t, "base_ingest", `"event_id", "text"`, ingest, "# report\n\nA sentence.\n")
	id = old.event(t, "red-lens-evidence", 1, ts3, "report_read")
	old.row(t, "report_read", `"event_id", "head"`, id, head(ingest))
	return old, ingest
}

// A MIGRATED REPORT READ IS SERVED AT THE HEAD ITS SOURCE WAS. Its head is an events.id — the lens's
// pin — and the replay renumbers, the cast it synthesizes landing first, so a head carried as written
// names the event before the ingest.
func TestAMigratedReportReadIsServedAtTheHeadItsSourceWas(t *testing.T) {
	old, ingest := readFixture(t, func(ingest int64) int64 { return ingest })
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
	m, err := record.MergedEvents(dst)
	if err != nil {
		t.Fatal(err)
	}
	ids := m.At.IDs(m.Events)
	var head, base int64
	for i, e := range m.Events {
		if r, ok := recordpb.BodyAs[*recordpb.ReportRead](e); ok {
			head = r.GetHead()
		}
		if e.GetType() == recordpb.EventType_EVENT_TYPE_BASE_INGEST {
			base = ids[i]
		}
	}
	if base == ingest {
		t.Fatal("the replay did not renumber, so nothing here exercised the translation")
	}
	if head != base {
		t.Errorf("the migrated read was served at event %d; its source's head was written as %d", head, base)
	}
}

// A head naming an event the replay wrote nothing for is refused by name, as a pin is.
func TestAReportReadServedAtNothingMigratedIsRefused(t *testing.T) {
	old, _ := readFixture(t, func(int64) int64 { return 99 })
	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	res, err := migrate.Replay(src, migrate.Entries(), runtest.New(t, recordtest.TmpRun(t)), migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, "served at event 99") {
		t.Fatalf("a report read naming no migrated event must be refused by name: %+v", res.Refusals)
	}
}
