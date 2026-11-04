package recordsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE DECODED STREAM IS KEPT PER HANDLE, AND A LATER READ LOADS ONLY THE EVENTS PAST IT (#1632).
//
// # Why a tail read is the whole read
//
// A decoded event never changes after it is written. The row is append-only (the events triggers
// refuse UPDATE and DELETE); its body rows are inserted in the same transaction as the row and no
// shipped writer updates or deletes one. Its Window is columns of the row — sitting_id, sitting and
// epoch, stamped by sittingOf at the insert — plus the sitting's owner, which is the opening event's
// seat (and its sitting_open body), an event at or before this one and as immutable. Nothing a later
// write does reaches an earlier row's decoded form. And `id` is the rowid of a table that is never
// deleted from, assigned under the write lock: a committed id is above every id committed before it.
// So the record past the last id this process decoded is exactly `WHERE id > last`, and the stream
// read whole is the cached prefix followed by that tail.
//
// # The key is the handle and the file it reads
//
// A stream is kept for a handle Open returned, beside the identity of the file at its path (device
// and inode, or the Windows volume and file index — os.SameFile). A file replaced at the path — a
// migration's fresh record moved into place, a test's copy — is another file, and its stream is read
// whole. The path alone would serve the old file's events with the new file's tail. The handle is
// part of the key because each handle is ONE connection (openUncached): every read on it is
// serialised, and each sees a snapshot no older than the one before it, so a prefix this handle read
// is present in every later read through it — the property the tail read stands on. A second handle
// on the same file promises nothing of the kind, and keeps a stream of its own.
//
// Before a tail read, the last cached row is asked for by id and compared: a handle whose connection
// was reopened onto a different file at the same path is caught there too.
//
// # What is never cached
//
// A narrowed read (EventsOfTypes, EventAt) neither fills the stream nor is served from it: it reads
// its own rows. A handle Open did not return (OpenSeparate, a test's own), and a bare *sql.Tx, read
// uncached. A transaction begun by BeginRead or BeginWrite carries its handle (Tx) and extends the
// handle's stream through itself.
//
// # A write transaction sees its own events, and publishes none of them
//
// A write transaction's read is the handle's stream plus its own tail, read through the transaction,
// so the events it has inserted and not committed are in it. Those may still roll back, so the
// handle's stream takes only the rows at or below the highest id committed when the transaction
// began (committedMax, read under the write lock BEGIN takes). Within the transaction a second read
// extends from the first (Tx.own): a migration's Replay validates every act against the record so
// far, and each validation reads only the acts since the last.

// stream is one handle's decoded record: the file it was read from, and the events with their
// windows, in id order. Its slices are never written after it is stored — a reader gets copies.
type stream struct {
	file os.FileInfo
	evs  []*recordpb.Event
	ws   []Window
}

func (s *stream) last() int64 {
	if s == nil || len(s.ws) == 0 {
		return 0
	}
	return s.ws[len(s.ws)-1].ID
}

var (
	streamMu sync.Mutex
	// pathOf is the database path of each handle Open returned; a handle not in it is not cached.
	pathOf  = map[*sql.DB]string{}
	streams = map[*sql.DB]*stream{}
)

// forgetHandle drops a handle's stream and path, as its handle is closed. The caller holds openMu.
func forgetHandle(db *sql.DB) {
	streamMu.Lock()
	defer streamMu.Unlock()
	delete(pathOf, db)
	delete(streams, db)
}

// rememberHandle records the path a cached handle reads. The caller holds openMu.
func rememberHandle(db *sql.DB, path string) {
	streamMu.Lock()
	defer streamMu.Unlock()
	pathOf[db] = path
}

// Tx is a transaction begun on a handle Open returned, carrying that handle, so a read of the whole
// record through it starts from the handle's stream rather than from the first event.
type Tx struct {
	*sql.Tx
	db *sql.DB
	// writes is a transaction begun by BeginWrite. Its reads may hold its own uncommitted events.
	writes bool
	// committedMax is the highest event id committed when a write transaction began: the most of
	// what it reads that the handle's stream may take.
	committedMax int64
	// own is the last stream read through this transaction; a later read in it extends from here.
	own *stream
}

// BeginRead opens a read-only transaction on db: one snapshot, every read of which is committed.
func BeginRead(db *sql.DB) (*Tx, error) {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, db: db}, nil
}

// BeginWrite opens a write transaction on db. The DSN takes the write lock at BEGIN, so the highest
// committed id read here cannot move until the transaction ends.
func BeginWrite(db *sql.DB) (*Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	t := &Tx{Tx: tx, db: db, writes: true}
	if err := tx.QueryRow(`SELECT COALESCE(max("id"), 0) FROM "events"`).Scan(&t.committedMax); err != nil {
		tx.Rollback() //nolint:errcheck // the read's error is the one to report
		return nil, fmt.Errorf("reading the record's last event: %w", err)
	}
	return t, nil
}

// cachedEvents is Events: the whole record, from the handle's stream where q carries a cached handle.
func cachedEvents(q Querier) ([]*recordpb.Event, []Window, error) {
	var db *sql.DB
	var tx *Tx
	switch h := q.(type) {
	case *sql.DB:
		db = h
	case *Tx:
		db, tx = h.db, h
	}
	streamMu.Lock()
	path, cached := pathOf[db]
	var base *stream
	if cached {
		base = streams[db]
	}
	streamMu.Unlock()
	if !cached {
		return eventsWhere(q, true, 0, ``)
	}
	file, err := os.Stat(path)
	if err != nil {
		return eventsWhere(q, true, 0, ``)
	}
	// On Windows a FileInfo from os.Stat loads its file index lazily, from the path, at the first
	// SameFile. Loaded now, it names the file stat found rather than whatever the path names later.
	os.SameFile(file, file)
	if base != nil && !os.SameFile(base.file, file) {
		base = nil
	}
	// What this transaction read last is through its own snapshot, so it is a prefix of this read
	// whatever the path names now.
	if tx != nil && tx.own != nil && tx.own.last() >= base.last() {
		base = tx.own
	}
	if base != nil && len(base.evs) > 0 {
		same, err := lastRowMatches(q, base)
		if err != nil {
			return nil, nil, err
		}
		if !same {
			base = nil
		}
	}
	after := base.last()
	var tail []*recordpb.Event
	var tailWs []Window
	if after == 0 {
		tail, tailWs, err = eventsWhere(q, true, 0, ``)
	} else {
		tail, tailWs, err = eventsWhere(q, true, after, ` WHERE e."id" > ?`, after)
	}
	if err != nil {
		return nil, nil, err
	}
	read := &stream{file: file}
	if base != nil {
		read.evs = append(append(make([]*recordpb.Event, 0, len(base.evs)+len(tail)), base.evs...), tail...)
		read.ws = append(append(make([]Window, 0, len(base.ws)+len(tailWs)), base.ws...), tailWs...)
	} else {
		read.evs, read.ws = tail, tailWs
	}
	if tx != nil {
		tx.own = read
	}
	publish(db, read, tx)
	return append([]*recordpb.Event(nil), read.evs...), append([]Window(nil), read.ws...), nil
}

// lastRowMatches is whether the row at the stream's last id is still the event the stream holds
// there: the same seat, clock and type. A connection reopened onto another file at the path reads
// a different row, or none.
func lastRowMatches(q Querier, s *stream) (bool, error) {
	ev := s.evs[len(s.evs)-1]
	var seat, ts, typ string
	err := q.QueryRow(`SELECT "seat_id", "ts", "type" FROM "events" WHERE "id" = ?`, s.last()).Scan(&seat, &ts, &typ)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading the record: %w", err)
	}
	return seat == ev.GetSeatId() && ts == ev.GetTs() && typ == recordpb.Word(ev.GetType()), nil
}

// publish stores read as db's stream where it is longer than the one stored, or the stored one is of
// another file. A write transaction's read is cut at the id committed when it began.
func publish(db *sql.DB, read *stream, tx *Tx) {
	if tx != nil && tx.writes {
		n := len(read.ws)
		for n > 0 && read.ws[n-1].ID > tx.committedMax {
			n--
		}
		read = &stream{file: read.file, evs: read.evs[:n:n], ws: read.ws[:n:n]}
	}
	streamMu.Lock()
	defer streamMu.Unlock()
	if _, ok := pathOf[db]; !ok {
		return // closed while this read ran
	}
	if old := streams[db]; old != nil && os.SameFile(old.file, read.file) && old.last() >= read.last() {
		return
	}
	streams[db] = read
}

// ForgetReads drops the stream this transaction read. After a ROLLBACK TO a savepoint the ids of
// the rolled-back rows are free again, so what was read before it is not a prefix of what is read
// next.
func (t *Tx) ForgetReads() { t.own = nil }
