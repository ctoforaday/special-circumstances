package catalogue

import (
	"context"
	"database/sql"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func newStore(t *testing.T) (string, *sql.DB) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "catalogue.db")
	db, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return p, db
}

// THE VIEWS ARE THE CONTRACT. Agents write their own SQL against these, so a renamed column
// breaks work outside this repository that no sweep can reach. This is the check that makes the
// rename fail HERE — #819's lesson, where 0 of 7 archives became unreadable because words were
// retired.
func TestViewColumnsAreTheContract(t *testing.T) {
	_, db := newStore(t)
	for view, want := range ViewColumns {
		rows, err := db.Query(`SELECT * FROM ` + view + ` LIMIT 0`)
		if err != nil {
			t.Errorf("%s: %v", view, err)
			continue
		}
		got, err := rows.Columns()
		rows.Close()
		if err != nil {
			t.Errorf("%s: %v", view, err)
			continue
		}
		if len(got) != len(want) {
			t.Errorf("%s columns = %v, want %v", view, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s columns = %v, want %v", view, got, want)
				break
			}
		}
	}
	// And the roster itself is pinned: a NEW view that nobody declared is as much a contract
	// change as a renamed column.
	rows, _ := db.Query(`SELECT name FROM sqlite_master WHERE type='view' ORDER BY name`)
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()
	var want []string
	for k := range ViewColumns {
		want = append(want, k)
	}
	sort.Strings(want)
	if len(names) != len(want) {
		t.Errorf("views on the contract = %v, declared = %v", names, want)
	}
}

// Writes are refused on the query path. Measured on modernc v1.57.0: mode=ro + query_only(1)
// refuse every mutation.
func TestWritesAreRefusedOnTheQueryPath(t *testing.T) {
	p, db := newStore(t)
	db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('S','/p',1,1)`)
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	for _, q := range []string{
		`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('X','/p',1,1)`,
		`UPDATE session SET project_dir='zzz'`,
		`DELETE FROM session`,
		`DROP TABLE session`,
		`CREATE TABLE evil(x)`,
	} {
		if _, err := ro.Exec(q); err == nil {
			t.Errorf("read-only connection ALLOWED %q", q)
		}
	}
	var n int
	if err := ro.QueryRow(`SELECT count(*) FROM v_session`).Scan(&n); err != nil || n != 1 {
		t.Errorf("a plain SELECT must still work: n=%d err=%v", n, err)
	}
}

// ATTACH is closed ON THE CONNECTION THE VERB USES, and this is the case that regresses
// silently: sqlite.Limit binds ONE connection, so a query issued through the pool can land on a
// connection the limit was never applied to. A test that only checks the pinned connection
// passes against an implementation that then queries through the pool.
func TestAttachIsRefusedOnThePinnedConnection(t *testing.T) {
	p, db := newStore(t)
	db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('S','/p',1,1)`)
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	ctx := context.Background()
	c, err := PinnedConn(ctx, ro)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.ExecContext(ctx, `ATTACH DATABASE '`+p+`' AS other`); err == nil {
		t.Error("ATTACH was allowed on the pinned connection — the limit did not bind")
	}
	// And an ordinary read still works through it.
	var n int
	if err := c.QueryRowContext(ctx, `SELECT count(*) FROM v_session`).Scan(&n); err != nil || n != 1 {
		t.Errorf("SELECT through the pinned connection: n=%d err=%v", n, err)
	}
}

// PRAGMA writable_schema=1 is a silent no-op under _defensive=1 — and unlike the ATTACH limit,
// this one holds pool-wide because it is applied at connection open.
func TestWritableSchemaIsANoOp(t *testing.T) {
	p, _ := newStore(t)
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	ro.Exec(`PRAGMA writable_schema=1`)
	var ws int
	if err := ro.QueryRow(`PRAGMA writable_schema`).Scan(&ws); err != nil {
		t.Fatal(err)
	}
	if ws != 0 {
		t.Errorf("writable_schema reads %d after being set; _defensive=1 did not hold", ws)
	}
}

// A runaway query is cancellable, so a bad SELECT cannot wedge the box.
func TestRunawayQueryIsCancelled(t *testing.T) {
	p, db := newStore(t)
	tx, _ := db.Begin()
	st, _ := tx.Prepare(`INSERT INTO act(session_id,seq,ts,tool,outcome) VALUES('S',?,0,'Bash','ok')`)
	for i := 0; i < 20000; i++ {
		st.Exec(i)
	}
	tx.Commit()
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := ro.QueryContext(ctx, `SELECT count(*) FROM act a, act b, act c`); err == nil {
		t.Error("a three-way cross join was not cancelled")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("cancellation took %v — the deadline is not being honoured", d)
	}
}

// A store written by a NEWER binary is refused, not read as this shape. The store is derived, so
// refusing costs a reprojection; guessing costs a store that answers wrongly while looking well.
func TestANewerShapeIsRefused(t *testing.T) {
	p, db := newStore(t)
	if _, err := db.Exec(`PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if _, err := Open(p); err == nil {
		t.Error("a store from a newer shape was opened rather than refused")
	}
}

// An absent store is an ERROR with a reason, never an empty result. "No rows" and "no store"
// are the same answer to a caller that cannot tell them apart.
func TestAnAbsentStoreIsRefusedNotEmpty(t *testing.T) {
	_, err := OpenRead(filepath.Join(t.TempDir(), "nothing.db"))
	if err == nil {
		t.Fatal("an absent store opened successfully; a caller would read zero rows as a clean board")
	}
}
