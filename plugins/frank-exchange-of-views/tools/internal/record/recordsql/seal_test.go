package recordsql

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A SEALED RECORD READS THE SAME WITH AND WITHOUT ITS WRITE-AHEAD LOG, and an unsealed one does
// not. That difference is the whole defect, so the test reproduces it rather than asserting that
// Seal returned nil.
//
// The reader here opens the database `immutable=1`, which is what tells SQLite the file cannot
// change and to skip the log entirely. That is not a contrived flag: it is the natural choice for
// an archived run that genuinely cannot change, it is what a `cp record.db` out of a tarball
// amounts to, and it is what an agent auditing the 2026-09-20 run actually used — reading 186
// events where there were 301, no outcome row where the run was `verified`, and filing the
// difference as a release blocker.
func TestAnUnsealedRecordReadsShortAndSealingFixesIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "record.db")

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(path) })
	if _, err := db.Exec(`CREATE TABLE ev (id INTEGER PRIMARY KEY, body TEXT)`); err != nil {
		t.Fatal(err)
	}
	// Enough rows that the log is certain to hold some of them rather than fitting in one page.
	const rows = 400
	for i := 0; i < rows; i++ {
		if _, err := db.Exec(`INSERT INTO ev (body) VALUES (?)`, "event body padded out so the log grows"); err != nil {
			t.Fatal(err)
		}
	}

	// The log must actually be carrying data, or the rest of this test proves nothing: a run whose
	// log happened to be empty would pass every assertion below with Seal deleted.
	n, err := SealedSize(path)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("the write-ahead log is empty before sealing, so this test cannot tell a seal from a no-op")
	}

	withoutLog := countVia(t, "file:"+path+"?mode=ro&immutable=1")
	if withoutLog == rows {
		t.Fatalf("an UNSEALED record already reads complete (%d rows) — the defect this guards is not reproduced here, "+
			"so a passing seal would prove nothing", withoutLog)
	}

	if err := Seal(path); err != nil {
		t.Fatal(err)
	}

	if n, err := SealedSize(path); err != nil {
		t.Fatal(err)
	} else if n != 0 {
		t.Errorf("after sealing, %d bytes remain in the write-ahead log — the record is still two files", n)
	}
	if got := countVia(t, "file:"+path+"?mode=ro&immutable=1"); got != rows {
		t.Errorf("a SEALED record still reads short without its log: %d rows, want %d", got, rows)
	}
}

// SealedSize reports zero for a record that never had a log, rather than an error a caller would
// have to special-case at every site.
func TestSealedSizeIsZeroWhenThereIsNoLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "record.db")
	if err := os.WriteFile(path, []byte("not a database, and not read as one"), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := SealedSize(path)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("SealedSize = %d for a record with no log, want 0", n)
	}
}

// countVia opens the database through its own handle — NOT the package cache — because the cached
// handle is the writer, and a writer sees its own uncommitted log regardless of how the reader
// would. The point is to read the file the way a later, separate process reads it.
func countVia(t *testing.T, dsn string) int {
	t.Helper()
	db, err := sql.Open(driverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM ev`).Scan(&n); err != nil {
		// A database whose table is not visible at all is the extreme of the same defect.
		return -1
	}
	return n
}

// A LEASED CONNECTION IS REFUSED, NOT WAITED ON OR CLOSED UNDER. The pool's one connection, leased,
// would leave the seal waiting forever; and database/sql closes an in-use connection only when it is
// returned, so its descriptor would stay open across the byte read whose close then drops its locks.
// The refusal leaves the handle cached and usable; once the lease ends the read goes through. The
// call runs under a deadline, so a missing refusal fails here rather than at the suite's timeout.
func TestSealedBytesRefusesWhileAConnectionIsLeased(t *testing.T) {
	path := filepath.Join(t.TempDir(), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(path) })
	if _, err := db.Exec(`CREATE TABLE ev (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	conn, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := SealedBytes(path); done <- err }()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "still in use") {
			t.Fatalf("SealedBytes with a leased connection = %v, want a refusal naming the connection in use", err)
		}
	case <-time.After(20 * time.Second):
		_ = conn.Close()
		t.Fatal("SealedBytes with a leased connection did not return: it waited on the lease instead of refusing")
	}
	if again, err := Open(path); err != nil || again != db {
		t.Fatalf("the refusal released the cached handle (reopen = %p, %v; was %p)", again, err, db)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := SealedBytes(path)
	if err != nil || len(b) == 0 {
		t.Fatalf("SealedBytes after the lease ended = %d bytes, %v", len(b), err)
	}
}
