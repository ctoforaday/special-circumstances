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

// A MIGRATED STALE REFUSAL READS FROM THE EVENT ITS SOURCE READ FROM. Its `through` is an events.id,
// and the replay renumbers — the cast it synthesizes lands first — so a read point carried as written
// names the event before the one the refusal showed. No archive holds a stale refusal yet, so the
// source is built here: two registers and a refusal reading through the second.
func TestAMigratedStaleRefusalReadsFromTheEventItsSourceShowed(t *testing.T) {
	old := newOldRecord(t)
	for _, ddl := range []string{
		`CREATE TABLE "stale" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "through" INTEGER)`,
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('stale', 'current vocabulary')`,
	} {
		if _, err := old.db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	id := old.event(t, "red-lens-evidence", 1, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version"`, id, "abc1234")
	shown := old.event(t, "red-lens-logic", 1, ts2, "register")
	old.row(t, "register", `"event_id", "tool_version"`, shown, "abc1234")
	id = old.event(t, "red-lens-evidence", 1, ts3, "stale")
	old.row(t, "stale", `"event_id", "through"`, id, shown)

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
	var through, logicRegister int64
	for i, e := range m.Events {
		if s, ok := recordpb.BodyAs[*recordpb.Stale](e); ok {
			through = s.GetThrough()
		}
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && e.GetSeatId() == "red-lens-logic" {
			logicRegister = ids[i]
		}
	}
	if logicRegister == shown {
		t.Fatal("the replay did not renumber, so nothing here exercised the translation")
	}
	if through != logicRegister {
		t.Errorf("the migrated refusal reads through event %d; the event its source showed was written as %d", through, logicRegister)
	}
}

// A read point naming an event the replay wrote nothing for is refused by name, as a pin is.
func TestAStaleRefusalReadingFromNothingMigratedIsRefused(t *testing.T) {
	old := newOldRecord(t)
	for _, ddl := range []string{
		`CREATE TABLE "stale" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "through" INTEGER)`,
		`INSERT INTO "enum_event_type" ("value", "means") VALUES ('stale', 'current vocabulary')`,
	} {
		if _, err := old.db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	id := old.event(t, "red-lens-evidence", 1, ts1, "register")
	old.row(t, "register", `"event_id", "tool_version"`, id, "abc1234")
	id = old.event(t, "red-lens-evidence", 1, ts2, "stale")
	old.row(t, "stale", `"event_id", "through"`, id, 99)

	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	res, err := migrate.Replay(src, migrate.Entries(), runtest.New(t, recordtest.TmpRun(t)), migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 1 || !strings.Contains(res.Refusals[0].Err, "reads from event 99") {
		t.Fatalf("a read point naming no migrated event must be refused by name: %+v", res.Refusals)
	}
}
