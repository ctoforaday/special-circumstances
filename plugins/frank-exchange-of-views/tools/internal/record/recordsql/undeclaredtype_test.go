package recordsql

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

// A NARROWED READ OVER AN UNDECLARED VOCABULARY IS REFUSED, NOT ANSWERED WITH ZERO.
//
// The full replay already refuses at the row — eventsWhere errors on an `events.type` word the
// schema does not declare. The narrowed read never reaches the row, because `WHERE type IN (…)`
// filtered it out first, and an empty result is what a projection renders as a clean board.
//
// Measured on run-archive/2026-09-02_quadratic-formula, whose events are typed `friction`:
//
//	EventsOfTypes(db, "log") -> 0 rows, nil error   (35 such events in the file)
//
// The row is written here with the foreign key off, because the point is a record whose vocabulary
// this schema no longer holds and the enum table is exactly what refuses to store one.
func TestANarrowedReadOverAnUndeclaredTypeIsRefused(t *testing.T) {
	db := freshRecord(t)
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO "events" ("seat_id","round","ts","type","key") VALUES ('s',0,'t','friction','k')`); err != nil {
		t.Fatal(err)
	}

	evs, err := EventsOfTypes(db, "log")
	if err == nil {
		t.Fatalf("a narrowed read over an undeclared vocabulary returned %d events and no error — "+
			"that empty result is what renders as a clean board", len(evs))
	}
	// THE REFUSAL MUST NAME THE WORD IT FOUND. Without it the reader cannot tell an unreadable
	// record from a misspelled query, which are the two states this exists to separate.
	for _, want := range []string{"friction", "does not declare", "same bytes"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal is missing %q:\n%v", want, err)
		}
	}
}

// AN HONEST ZERO STAYS ZERO. A record whose vocabulary is entirely declared answers a narrowed read
// for a type it happens not to hold with no events and no error — which is the behaviour every
// projection depends on and the reason this cannot simply refuse an empty result.
func TestANarrowedReadForATypeTheRecordDoesNotHoldIsStillZero(t *testing.T) {
	db := freshRecord(t)
	if _, err := db.Exec(`INSERT INTO "events" ("seat_id","round","ts","type","key") VALUES ('s',0,'t','position','k')`); err != nil {
		t.Fatal(err)
	}
	evs, err := EventsOfTypes(db, "finding")
	if err != nil {
		t.Fatalf("a record with a wholly declared vocabulary was refused: %v", err)
	}
	if len(evs) != 0 {
		t.Errorf("a narrowed read for an absent type returned %d events", len(evs))
	}
}

// THE DECLARED CASE IS NOT REFUSED, which is the assertion that keeps this check off every run: the
// vocabulary this binary WRITES must be one it also accepts.
func TestANarrowedReadOverThisSchemasOwnVocabularyIsNotRefused(t *testing.T) {
	db := freshRecord(t)
	if _, err := db.Exec(`INSERT INTO "events" ("seat_id","round","ts","type","key") VALUES ('s',0,'t','position','k')`); err != nil {
		t.Fatal(err)
	}
	if err := refuseUndeclaredTypes(db); err != nil {
		t.Errorf("this schema refuses a record it wrote itself: %v", err)
	}
}

func freshRecord(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() { _ = CloseUnder(dir) })
	db, err := Open(filepath.Join(dir, "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	return db
}
