package recordsql

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// cacheOf is the statement cache on db's one connection.
func cacheOf(t *testing.T, db *sql.DB) *stmtCacheConn {
	t.Helper()
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var c *stmtCacheConn
	if err := conn.Raw(func(dc any) error {
		c = dc.(*stmtCacheConn)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return c
}

// keptStmt is the statement the cache keeps for query, or nil.
func keptStmt(c *stmtCacheConn, query string) *cachedStmt {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stmts[query]
}

func seedLogs(t *testing.T, db *sql.DB, n int) {
	t.Helper()
	for i := 1; i <= n; i++ {
		if _, err := Insert(db, event(t, int32(i), recordpb.EventType_EVENT_TYPE_LOG, logEntry())); err != nil {
			t.Fatal(err)
		}
	}
}

func logEntry() *recordpb.Log {
	return &recordpb.Log{Text: proto.String("x"), Type: recordpb.LogType_LOG_TYPE_FRICTION.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
}

const countEvents = `SELECT count(*) FROM "events"`

// A READ IS PREPARED ONCE ON ITS HANDLE, and asked again through the handle or through a transaction
// on it the kept statement answers — from the record as it is now, not as it was when prepared.
// A write and a PRAGMA are not kept.
func TestARecordReadIsPreparedOnceAndAnswersFromTheRecordAsItIsNow(t *testing.T) {
	db := store(t)
	seedLogs(t, db, 2)
	c := cacheOf(t, db)
	count := func(q Querier) int {
		var n int
		if err := q.QueryRow(countEvents).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := count(db); n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
	kept := keptStmt(c, countEvents)
	if kept == nil {
		t.Fatal("the read was not kept prepared")
	}
	if _, err := Insert(db, event(t, 3, recordpb.EventType_EVENT_TYPE_LOG, logEntry())); err != nil {
		t.Fatal(err)
	}
	if n := count(db); n != 3 {
		t.Fatalf("count after a write = %d, want 3 — the kept statement answered from a stale record", n)
	}
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if n := count(tx); n != 3 {
		t.Fatalf("count through a transaction = %d, want 3", n)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if got := keptStmt(c, countEvents); got != kept {
		t.Errorf("the read was prepared again (%p, first %p) — the cache did not keep it", got, kept)
	}
	if _, err := db.Exec(`PRAGMA user_version`); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	for q := range c.stmts {
		if !isCachedRead(q) {
			t.Errorf("the cache keeps %q, which is not a read", q)
		}
	}
	c.mu.Unlock()
}

// A KEPT STATEMENT IN USE IS NOT ASKED AGAIN. Inside one transaction, a read is asked while the rows
// of the same read are still open; each answer is whole, and the first is not rewound by the
// second.
func TestAReadAskedWhileItsRowsAreOpenAnswersBoth(t *testing.T) {
	db := store(t)
	seedLogs(t, db, 4)
	const ids = `SELECT "id" FROM "events" ORDER BY "id"`
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	outer, err := tx.Query(ids)
	if err != nil {
		t.Fatal(err)
	}
	var got []int64
	for outer.Next() {
		var id int64
		if err := outer.Scan(&id); err != nil {
			t.Fatal(err)
		}
		got = append(got, id)
		inner, err := tx.Query(ids)
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for inner.Next() {
			n++
		}
		if err := inner.Close(); err != nil {
			t.Fatal(err)
		}
		if n != 4 {
			t.Fatalf("the inner read returned %d rows, want 4", n)
		}
	}
	if err := outer.Close(); err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 || got[0] != 1 || got[3] != 4 {
		t.Errorf("the outer read returned %v, want ids 1..4 — the inner read rewound it", got)
	}
}

// CLOSING THE HANDLE FINALIZES ITS KEPT STATEMENTS, so the connection releases the file rather than
// lingering as a zombie that holds it (on Windows a held file cannot be removed), and a handle opened
// after it prepares its own.
func TestClosingTheHandleFinalizesItsKeptStatements(t *testing.T) {
	dir := tmpRun(t)
	path := filepath.Join(dir, "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	seedLogs(t, db, 1)
	var n int
	if err := db.QueryRow(countEvents).Scan(&n); err != nil {
		t.Fatal(err)
	}
	c := cacheOf(t, db)
	if keptStmt(c, countEvents) == nil {
		t.Fatal("the read was not kept prepared")
	}
	if err := Close(path); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	left := len(c.stmts)
	c.mu.Unlock()
	if left != 0 {
		t.Errorf("%d statement(s) outlived the handle's close", left)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(countEvents).Scan(&n); err != nil || n != 1 {
		t.Fatalf("a reopened handle reads %d, %v; want 1", n, err)
	}
	if err := Close(path); err != nil {
		t.Fatal(err)
	}
}
