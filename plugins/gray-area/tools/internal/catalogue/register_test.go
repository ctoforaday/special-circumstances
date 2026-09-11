package catalogue

import (
	"database/sql"
	"testing"
	"time"
)

type registered struct {
	projectDir, bridge string
	first, last        int64
	closed             sql.NullInt64
}

func registeredRow(t *testing.T, db *sql.DB, sid string) registered {
	t.Helper()
	var r registered
	if err := db.QueryRow(`SELECT project_dir, bridge_session_id, ingested_first, ingested_last, closed_at
		FROM v_session WHERE session_id = ?`, sid).Scan(&r.projectDir, &r.bridge, &r.first, &r.last, &r.closed); err != nil {
		t.Fatalf("no v_session row for %s: %v", sid, err)
	}
	return r
}

// REGISTRATION INSERTS, REOPENS, AND NEVER FORGETS A CLOUD ID. Each of the three is what a later
// `agents --lost` depends on: a row for a session with nothing ingested yet, a settled session
// made unsettled again by running, and the one id the transcript does not hold.
func TestRegisterSessionInsertsReopensAndKeepsTheCloudID(t *testing.T) {
	db := store(t)
	at := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	if err := RegisterSession(db, "S", "/p/-w", "session_01A", at); err != nil {
		t.Fatal(err)
	}
	r := registeredRow(t, db, "S")
	if r.projectDir != "/p/-w" || r.bridge != "session_01A" || r.first != at.Unix() || r.last != at.Unix() || r.closed.Valid {
		t.Fatalf("after insert: %+v", r)
	}

	// Closure settled it; the session then runs again.
	if _, err := db.Exec(`UPDATE session SET closed_at = 5 WHERE session_id = 'S'`); err != nil {
		t.Fatal(err)
	}
	if err := RegisterSession(db, "S", "", "", at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	r = registeredRow(t, db, "S")
	if r.closed.Valid {
		t.Error("a registration left closed_at set, so a resumed session stays settled and is never listed as lost")
	}
	if r.bridge != "session_01A" {
		t.Errorf("a registration with no cloud id erased the known one: %q", r.bridge)
	}
	if r.projectDir != "/p/-w" {
		t.Errorf("a registration with no directory erased the known one: %q", r.projectDir)
	}
	if r.first != at.Unix() || r.last != at.Unix() {
		t.Errorf("a repeat registration moved ingested_first/last to %d/%d — a registration reads nothing", r.first, r.last)
	}

	if err := RegisterSession(db, "S", "/p/-x", "session_02B", at.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if r = registeredRow(t, db, "S"); r.bridge != "session_02B" || r.projectDir != "/p/-x" {
		t.Errorf("a new id and directory were not recorded: %+v", r)
	}
}

func TestRegisterSessionRefusesAnEmptyID(t *testing.T) {
	db := store(t)
	if err := RegisterSession(db, "", "/p", "session_01A", time.Now()); err == nil {
		t.Error("a session with no id was registered")
	}
	var n int
	db.QueryRow(`SELECT count(*) FROM session`).Scan(&n)
	if n != 0 {
		t.Errorf("%d rows after a refused registration", n)
	}
}
