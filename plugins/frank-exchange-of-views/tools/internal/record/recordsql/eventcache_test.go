package recordsql

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// logAct is a seat's log entry at a fixed clock, keyed by n; text is what tells two of them apart.
func logAct(t *testing.T, seat string, n int, text string) *recordpb.Event {
	t.Helper()
	ev := &recordpb.Event{
		SeatId: proto.String(seat),
		Ts:     proto.String("2026-01-01T00:00:00Z"),
		Key:    proto.String(fmt.Sprintf("%s:log:#%d", seat, n)),
	}
	if _, err := recordpb.SetBody(ev, &recordpb.Log{
		Text:   proto.String(text),
		Type:   recordpb.LogType_LOG_TYPE_FRICTION.Enum(),
		Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum(),
	}); err != nil {
		t.Fatal(err)
	}
	return ev
}

func registerAct(t *testing.T, seat string, n int) *recordpb.Event {
	t.Helper()
	ev := &recordpb.Event{
		SeatId: proto.String(seat),
		Ts:     proto.String("2026-01-01T00:00:00Z"),
		Key:    proto.String(fmt.Sprintf("%s:register:#%d", seat, n)),
	}
	if _, err := recordpb.SetBody(ev, &recordpb.Register{}); err != nil {
		t.Fatal(err)
	}
	return ev
}

func insertAll(t *testing.T, db *sql.DB, evs ...*recordpb.Event) {
	t.Helper()
	for _, ev := range evs {
		if _, err := Insert(db, ev); err != nil {
			t.Fatal(err)
		}
	}
}

// sameAsUncached holds Events(q) to the uncached read of the same querier, event for event and
// window for window, and returns the cached read.
func sameAsUncached(t *testing.T, q Querier) []*recordpb.Event {
	t.Helper()
	got, gotWs, err := Events(q)
	if err != nil {
		t.Fatal(err)
	}
	want, wantWs, err := eventsWhere(q, true, 0, ``)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) || len(gotWs) != len(wantWs) {
		t.Fatalf("the cached read holds %d events and %d windows, the uncached %d and %d", len(got), len(gotWs), len(want), len(wantWs))
	}
	for i := range want {
		if !proto.Equal(got[i], want[i]) {
			t.Errorf("event %d: cached %v, uncached %v", i, got[i], want[i])
		}
		if gotWs[i] != wantWs[i] {
			t.Errorf("window %d: cached %+v, uncached %+v", i, gotWs[i], wantWs[i])
		}
	}
	return got
}

func cachedLast(db *sql.DB) int64 {
	streamMu.Lock()
	defer streamMu.Unlock()
	return streams[db].last()
}

// A LATER READ IS THE CACHED STREAM AND ITS TAIL, and equals the read made whole.
func TestTheCachedStreamIsTheWholeReadAfterAppends(t *testing.T) {
	db := store(t)
	insertAll(t, db, registerAct(t, "red-lens-logic", 1), logAct(t, "red-lens-logic", 1, "one"))
	first := sameAsUncached(t, db)
	insertAll(t, db, logAct(t, "red-lens-logic", 2, "two"), registerAct(t, "red-chair", 1), logAct(t, "red-chair", 1, "three"))
	second := sameAsUncached(t, db)
	if len(second) != 5 {
		t.Fatalf("read %d events after five were written", len(second))
	}
	if second[0] != first[0] {
		t.Error("the second read decoded the first event again: it was not served from the stream")
	}
	if cachedLast(db) != 5 {
		t.Errorf("the handle's stream ends at %d, want 5", cachedLast(db))
	}
}

// A FILE REPLACED AT THE PATH IS READ WHOLE. The replacement holds the same events at the same ids,
// the last of them identical, and a different text at the first: only the file's identity tells the
// two apart, so a stream keyed on the path serves the old first event with the new file's tail.
func TestAFileReplacedAtThePathIsReadWhole(t *testing.T) {
	dir := tmpRun(t)
	path := filepath.Join(dir, "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// Every query on a fresh connection, so the handle reads whichever file the path names.
	db.SetMaxIdleConns(0)
	insertAll(t, db, logAct(t, "red-lens-logic", 1, "old"), logAct(t, "red-lens-logic", 2, "same"))
	sameAsUncached(t, db)

	other := filepath.Join(tmpRun(t), "record.db")
	odb, err := Open(other)
	if err != nil {
		t.Fatal(err)
	}
	insertAll(t, odb, logAct(t, "red-lens-logic", 1, "new"), logAct(t, "red-lens-logic", 2, "same"), logAct(t, "red-lens-logic", 3, "tail"))
	if err := Close(other); err != nil {
		t.Fatal(err)
	}
	for _, sfx := range []string{"-wal", "-shm"} {
		if err := os.Remove(path + sfx); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	if err := os.Rename(other, path); err != nil {
		t.Fatal(err)
	}
	got := sameAsUncached(t, db)
	if len(got) != 3 || got[0].GetLog().GetText() != "new" {
		t.Fatalf("after the replace the handle read %d events, the first %q", len(got), got[0].GetLog().GetText())
	}
}

// A FILE REWRITTEN IN PLACE KEEPS ITS IDENTITY, and the last cached row, asked for again, tells the
// handle its stream is of another record: a copy over the path writes into the same inode.
func TestAFileRewrittenInPlaceIsReadWhole(t *testing.T) {
	dir := tmpRun(t)
	path := filepath.Join(dir, "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxIdleConns(0)
	insertAll(t, db, logAct(t, "red-lens-logic", 1, "old"), logAct(t, "red-lens-logic", 2, "old"))
	sameAsUncached(t, db)

	other := filepath.Join(tmpRun(t), "record.db")
	odb, err := Open(other)
	if err != nil {
		t.Fatal(err)
	}
	insertAll(t, odb, logAct(t, "red-chair", 1, "new"), logAct(t, "red-chair", 2, "new"), logAct(t, "red-chair", 3, "new"))
	if err := Close(other); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	for _, sfx := range []string{"-wal", "-shm"} {
		if err := os.Remove(path + sfx); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("the rewrite made a new file; this test is of one that keeps its identity")
	}
	got := sameAsUncached(t, db)
	if len(got) != 3 || got[0].GetSeatId() != "red-chair" {
		t.Fatalf("after the rewrite the handle read %d events, the first by %s", len(got), got[0].GetSeatId())
	}
}

// A WRITE TRANSACTION READS ITS OWN EVENTS, AND THE HANDLE'S STREAM TAKES NONE OF THEM: they may
// roll back. What was committed when it began is published; a rollback leaves the handle's stream
// as it was, and a commit is read by the next read as a tail.
func TestAWriteTransactionSeesItsOwnEventsAndPublishesNone(t *testing.T) {
	db := store(t)
	insertAll(t, db, logAct(t, "red-chair", 1, "a"), logAct(t, "red-chair", 2, "b"))

	// Nothing read yet: the transaction's read publishes what was committed before it.
	tx, err := BeginWrite(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InsertTx(tx.Tx, logAct(t, "red-chair", 3, "uncommitted")); err != nil {
		t.Fatal(err)
	}
	if got := sameAsUncached(t, tx); len(got) != 3 {
		t.Fatalf("the transaction read %d events, want its own third", len(got))
	}
	if cachedLast(db) != 2 {
		t.Errorf("the handle's stream ends at %d after a write transaction's read, want the committed 2", cachedLast(db))
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if got := sameAsUncached(t, db); len(got) != 2 {
		t.Fatalf("after the rollback the handle read %d events, want 2", len(got))
	}

	// A second read in one transaction extends the first through it.
	tx, err = BeginWrite(db)
	if err != nil {
		t.Fatal(err)
	}
	sameAsUncached(t, tx)
	if _, err := InsertTx(tx.Tx, logAct(t, "red-chair", 3, "c")); err != nil {
		t.Fatal(err)
	}
	if got := sameAsUncached(t, tx); len(got) != 3 || got[2].GetLog().GetText() != "c" {
		t.Fatalf("the transaction's second read holds %d events", len(got))
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if got := sameAsUncached(t, db); len(got) != 3 {
		t.Fatalf("after the commit the handle read %d events, want 3", len(got))
	}
	if cachedLast(db) != 3 {
		t.Errorf("the handle's stream ends at %d after the commit's read, want 3", cachedLast(db))
	}
}

// A ROLLBACK TO A SAVEPOINT FREES ITS IDS, so the transaction's earlier read is no prefix of its next:
// the act written after it takes the rolled-back act's id, with the same seat, clock and type.
func TestARolledBackSavepointIsNotReadAgain(t *testing.T) {
	db := store(t)
	insertAll(t, db, logAct(t, "red-chair", 1, "a"))
	tx, err := BeginWrite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`SAVEPOINT act`); err != nil {
		t.Fatal(err)
	}
	if _, err := InsertTx(tx.Tx, logAct(t, "red-chair", 2, "refused")); err != nil {
		t.Fatal(err)
	}
	sameAsUncached(t, tx)
	if _, err := tx.Exec(`ROLLBACK TO act`); err != nil {
		t.Fatal(err)
	}
	tx.ForgetReads()
	if _, err := InsertTx(tx.Tx, logAct(t, "red-chair", 2, "landed")); err != nil {
		t.Fatal(err)
	}
	if got := sameAsUncached(t, tx); got[1].GetLog().GetText() != "landed" {
		t.Fatalf("the read after the rollback holds %q at the reused id", got[1].GetLog().GetText())
	}
}

// A READ-ONLY TRANSACTION'S READ IS COMMITTED, and the handle's stream takes it.
func TestAReadTransactionPublishes(t *testing.T) {
	db := store(t)
	insertAll(t, db, logAct(t, "red-chair", 1, "a"), logAct(t, "red-chair", 2, "b"))
	tx, err := BeginRead(db)
	if err != nil {
		t.Fatal(err)
	}
	sameAsUncached(t, tx)
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if cachedLast(db) != 2 {
		t.Errorf("the handle's stream ends at %d after a read transaction, want 2", cachedLast(db))
	}
}

// A NARROWED READ NEITHER FILLS THE STREAM NOR IS SERVED FROM IT, and a handle Open did not return,
// or a bare transaction, keeps none.
func TestANarrowedOrUncachedReadKeepsNoStream(t *testing.T) {
	db := store(t)
	insertAll(t, db, registerAct(t, "red-chair", 1), logAct(t, "red-chair", 1, "a"))
	if _, _, err := EventsOfTypes(db, "log"); err != nil {
		t.Fatal(err)
	}
	if _, err := EventAt(db, 1); err != nil {
		t.Fatal(err)
	}
	if cachedLast(db) != 0 {
		t.Fatal("a narrowed read filled the handle's stream")
	}
	sameAsUncached(t, db)
	insertAll(t, db, logAct(t, "red-chair", 2, "b"))
	logs, _, err := EventsOfTypes(db, "log")
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("the narrowed read after an append holds %d logs, want 2", len(logs))
	}

	raw, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := Events(raw); err != nil {
		t.Fatal(err)
	}
	if err := raw.Rollback(); err != nil {
		t.Fatal(err)
	}
	if cachedLast(db) != 2 {
		t.Errorf("a bare transaction's read moved the handle's stream to %d", cachedLast(db))
	}

	sep, err := OpenSeparate(filepath.Join(filepath.Dir(pathOfForTest(t, db)), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sep.Close()
	sameAsUncached(t, sep)
	streamMu.Lock()
	_, kept := streams[sep]
	streamMu.Unlock()
	if kept {
		t.Error("a handle Open did not return keeps a stream")
	}
}

func pathOfForTest(t *testing.T, db *sql.DB) string {
	t.Helper()
	streamMu.Lock()
	defer streamMu.Unlock()
	p, ok := pathOf[db]
	if !ok {
		t.Fatal("the handle Open returned has no path")
	}
	return p
}
