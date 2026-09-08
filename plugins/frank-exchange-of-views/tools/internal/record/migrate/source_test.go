package migrate_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// THE WAL IS THE RECORD'S TAIL, AND READING WITHOUT IT IS THE FAILING ARM. Measured on the
// real archive during the plan audit: record.db alone read CLEAN at 691 events while 122
// more — most of the friction events the migration exists to translate — lived only in the
// 3.1 MB -wal beside it. Same bytes as a shorter honest record, which is why the adapter's
// contract is the file SET.
func TestReadingWithoutTheWALIsShorterAndTheAdapterIsNot(t *testing.T) {
	old := newOldRecord(t)
	if _, err := old.db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		id := old.event(t, "blue-lane-1", 1, ts1, "friction")
		old.row(t, "friction", `"event_id", "text", "kind"`, id, "entry", "friction")
	}
	// COPY WHILE THE WRITER IS STILL OPEN — closing checkpoints the WAL, and the archived
	// runs were tarred from live directories, which is exactly how the real one kept its
	// tail in the -wal.
	if _, err := os.Stat(filepath.Join(old.records(), "record.db-wal")); err != nil {
		t.Skipf("no WAL materialized (%v) — the driver checkpointed early and this arm cannot be staged", err)
	}

	// The failing arm: the database alone.
	alone := t.TempDir()
	copyFile(t, filepath.Join(old.records(), "record.db"), filepath.Join(alone, "record.db"))
	db, err := sql.Open("sqlite", filepath.Join(alone, "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var short int
	if err := db.QueryRow(`SELECT count(*) FROM "events"`).Scan(&short); err != nil {
		t.Fatal(err)
	}
	if short >= 20 {
		t.Skipf("the WAL held no tail (%d events without it) — nothing for this arm to show", short)
	}

	// The adapter: the file set.
	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 20 {
		t.Fatalf("the adapter read %d events; the record holds 20 (the db alone showed %d) — the file-set contract is broken", len(evs), short)
	}
}

// The originals are never modified: reading through the adapter leaves the source bytes
// exactly as found, because a bare open can checkpoint a WAL and the archive is the raw
// record.
func TestTheSourceBytesAreLeftAsFound(t *testing.T) {
	old := newOldRecord(t)
	id := old.event(t, "blue-lane-1", 1, ts1, "friction_none")
	old.row(t, "friction_none", `"event_id", "text"`, id, "clean")

	before := dirSizes(t, old.records())
	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := src.Events(); err != nil {
		t.Fatal(err)
	}
	src.Close()
	after := dirSizes(t, old.records())
	for name, n := range before {
		if after[name] != n {
			t.Errorf("%s changed size %d -> %d — the adapter touched the original", name, n, after[name])
		}
	}
}

// A row-bearing table no rule can place is a refusal: rows nobody can attribute are events
// nobody replays, and folding them into a clean run is the exact silence this package exists
// to refuse.
func TestARowBearingUnclassifiableTableRefuses(t *testing.T) {
	old := newOldRecord(t)
	old.event(t, "blue-lane-1", 1, ts1, "friction_none")
	old.row(t, "friction_none", `"event_id", "text"`, 1, "clean")
	if _, err := old.db.Exec(`CREATE TABLE "mystery" ("x" TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := old.db.Exec(`INSERT INTO "mystery" ("x") VALUES ('?')`); err != nil {
		t.Fatal(err)
	}
	src, err := migrate.OpenSQLite(old.records(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	if _, err := src.Events(); err == nil {
		t.Fatal("a row-bearing unclassifiable table read clean — those rows just vanished")
	}
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	b, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func dirSizes(t *testing.T, dir string) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		st, err := e.Info()
		if err != nil {
			t.Fatal(err)
		}
		out[e.Name()] = st.Size()
	}
	return out
}
