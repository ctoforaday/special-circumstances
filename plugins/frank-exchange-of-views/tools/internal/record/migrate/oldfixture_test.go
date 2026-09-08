package migrate_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// oldRecord hand-authors a FORMER-vocabulary record: the 2026-09-02 era shape, with the
// retired words as their own body tables. This is deliberately not built by any current
// writer — the current writer cannot write these shapes, which is the whole situation the
// package exists for.
type oldRecord struct {
	dir string // the run directory; the database lives under records/
	db  *sql.DB
}

func newOldRecord(t *testing.T) *oldRecord {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, ddl := range []string{
		`CREATE TABLE "events" ("id" INTEGER PRIMARY KEY, "seat_id" TEXT NOT NULL, "round" INTEGER NOT NULL, "ts" TEXT NOT NULL, "type" TEXT NOT NULL, "key" TEXT)`,
		`CREATE TABLE "enum_event_type" ("value" TEXT PRIMARY KEY, "means" TEXT NOT NULL)`,
		`CREATE TABLE "register" ("event_id" INTEGER PRIMARY KEY, "tool_version" TEXT)`,
		`CREATE TABLE "friction" ("event_id" INTEGER PRIMARY KEY, "text" TEXT, "kind" TEXT, "estopped_by" TEXT)`,
		`CREATE TABLE "friction_none" ("event_id" INTEGER PRIMARY KEY, "text" TEXT)`,
		`CREATE TABLE "opinion" ("event_id" INTEGER PRIMARY KEY, "gap_id" TEXT, "disposition" TEXT, "principle" TEXT, "tension" TEXT, "review_flag" TEXT, "rationale" TEXT, "settled" TEXT, "reopens_on" TEXT, "final" INTEGER)`,
	} {
		if _, err := db.Exec(ddl); err != nil {
			t.Fatalf("old fixture DDL: %v", err)
		}
	}
	for _, w := range []string{"register", "friction", "friction_none", "opinion"} {
		if _, err := db.Exec(`INSERT INTO "enum_event_type" ("value", "means") VALUES (?, 'former vocabulary')`, w); err != nil {
			t.Fatal(err)
		}
	}
	return &oldRecord{dir: dir, db: db}
}

func (o *oldRecord) event(t *testing.T, seat string, round int, ts, word string) int64 {
	t.Helper()
	res, err := o.db.Exec(`INSERT INTO "events" ("seat_id", "round", "ts", "type") VALUES (?, ?, ?, ?)`,
		seat, round, ts, word)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func (o *oldRecord) row(t *testing.T, table string, cols string, args ...any) {
	t.Helper()
	marks := ""
	for range args {
		if marks != "" {
			marks += ", "
		}
		marks += "?"
	}
	if _, err := o.db.Exec(`INSERT INTO "`+table+`" (`+cols+`) VALUES (`+marks+`)`, args...); err != nil {
		t.Fatalf("old fixture %s row: %v", table, err)
	}
}

func (o *oldRecord) records() string { return filepath.Join(o.dir, "records") }
