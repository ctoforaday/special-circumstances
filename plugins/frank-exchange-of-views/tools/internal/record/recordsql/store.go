package recordsql

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	// THE DRIVER BELONGS TO THE PACKAGE THAT OPENS THE DATABASE, not to its tests.
	//
	// It lived in schema_test.go, so database/sql had a registered "sqlite" driver throughout the
	// suite and none in the shipped binary. Every test passed and the first real `merge register`
	// failed with `unknown driver "sqlite"`. A blank import is invisible to the compiler's unused
	// check, which is exactly why the wrong file stayed good enough.
	_ "modernc.org/sqlite"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Open creates or opens a run's database and applies the derived schema to a new one.
//
// The pragmas are not defaults and each is load-bearing:
//
//   - foreign_keys is OFF by default in SQLite, for backwards compatibility with databases written
//     before it existed. Left off, every foreign key in this schema is decoration — including the
//     one that makes a ruling-before-its-filing unwritable, which is the reason the record is here
//     at all.
//   - journal_mode=WAL lets the readers a run has — the projections, the dashboard watcher — read
//     while a seat writes, which the shard files achieved by giving every seat its own file and
//     paid for with an ordering problem.
//   - busy_timeout keeps a concurrent seat waiting rather than failing. Seats are dispatched in
//     parallel and a lock contended for a few milliseconds is not an error to report to a human.
//
// dsnFor builds the DSN, and it uses net/url rather than escaping the path by hand.
//
// WE ARE THE ONES BUILDING A URI. `_txlock=immediate` is a DRIVER setting — it changes how BEGIN
// is issued and cannot be an Exec'd pragma — so the query-string form is not optional, and having
// chosen a URI, escaping the path into it is this function's obligation. It was `"file:" + path`,
// which is a filesystem path spliced into a grammar that reserves `?` and `#`: a run directory
// containing either was TRUNCATED there, silently, to a path that still opens. Two runs then share
// one database, or a run opens one that is not its own, and every write reports success.
//
// A HAND-ROLLED ESCAPE GOT THE THREE RESERVED CHARACTERS AND STOPPED THERE. url.URL escapes the
// same three and everything else the grammar needs, in the right order, and it also handles the
// one case percent-encoding cannot: a path beginning `//` would parse as an AUTHORITY, and
// url.URL emits an explicit empty authority (`file:////server/share`) so the path survives whole.
// RawQuery is written verbatim, which is what keeps `busy_timeout(5000)`'s parentheses intact.
func dsnFor(path string) string {
	u := url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "_txlock=immediate&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)",
	}
	return u.String()
}

func Open(path string) (*sql.DB, error) {
	// _txlock=immediate IS NOT A TUNING KNOB. IT IS THE DIFFERENCE BETWEEN WORKING AND NOT.
	//
	// A default `BEGIN` is DEFERRED: it takes no lock, acquires a READ lock at the first SELECT, and
	// tries to UPGRADE to a write lock at the first INSERT. The write path counts the seat's existing
	// events and then inserts, so it is exactly that shape — and SQLite deliberately does NOT apply
	// busy_timeout to an upgrade, because two readers both waiting to upgrade would deadlock. It
	// returns SQLITE_BUSY immediately instead.
	//
	// MEASURED, through the real binary: 8 concurrent seat processes writing 5 events each lost
	// roughly half of them to `database is locked (5) (SQLITE_BUSY)`. Not a slow path — a REFUSED
	// act, with the seat told its record failed. The shard layout had no contention by construction
	// (one file per seat), so this is a hazard the storage change INTRODUCED and had to answer.
	//
	// `immediate` takes the write lock at BEGIN, before any read. busy_timeout then applies, because
	// a writer waiting for a writer is a queue rather than a deadlock. It also fixes the correctness
	// half: a deferred transaction's count could be read from a snapshot another writer has already
	// moved past, which is what SQLITE_BUSY_SNAPSHOT exists to refuse.
	db, err := sql.Open("sqlite", dsnFor(path))
	if err != nil {
		return nil, err
	}
	// ONE CONNECTION. A pool is the wrong shape for a single-writer file database.
	//
	// `_txlock=immediate` takes the write lock at BEGIN, so two goroutines holding two pooled
	// connections collide in SQLITE_BUSY and then poll — on SQLite's default busy schedule, which
	// is a FIXED sequence ({1,2,5,10,15,20,25,25,25,50,50,100}ms, 100ms thereafter) with NO
	// jitter, so contending waiters wake together and retry together. With one connection they
	// queue on a Go mutex instead: no lock contention to resolve, no polling, no herd.
	//
	// It does NOT replace busy_timeout, which is still what handles the case this cannot touch —
	// a DIFFERENT PROCESS holding the write lock. Seats are separate processes; that is the
	// contention the 8-process test exists for.
	//
	// SAFE HERE because nothing queries the pool while a transaction from it is open: `Append`
	// validates (which reads the board) BEFORE insertNumbered begins its transaction, and
	// deriveKey counts through the tx rather than around it. With one connection, a db.Query
	// inside an open tx would deadlock on itself rather than merely contend — so that ordering is
	// now load-bearing, not incidental.
	db.SetMaxOpenConns(1)

	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='events'`).Scan(&n); err != nil {
		db.Close()
		return nil, fmt.Errorf("recordsql: reading %s: %w", path, err)
	}
	if n == 0 {
		if err := applySchema(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("recordsql: applying the schema to %s: %w", path, err)
		}
	}
	return db, nil
}

// applySchema creates the whole record in ONE transaction.
//
// IT WAS 171 STATEMENTS, EACH ITS OWN IMPLICIT TRANSACTION. SQLite autocommits any statement not
// already inside one, and every commit is an fsync — so creating a run directory paid 171 disk
// syncs to write a schema that is DERIVED and could be regenerated for free. Measured here:
// 499ms per fresh database, against 39ms for the same DDL in one transaction. Thirteen times, and
// the cost is paid by every test that opens a run (208 of them in internal/cli alone) as well as
// by every real run.
//
// NOTHING IS TRADED FOR IT. The obvious alternatives all weaken durability — `synchronous=off`
// measured 46ms and risks corruption, `synchronous=normal` 67ms and can drop recent commits — and
// this is FASTER than either while leaving `synchronous` at its default. One fsync instead of
// 171 is not a relaxed guarantee; it is the same guarantee, asked for once.
//
// SQLite's own forum states the mechanism: an autocommitted statement fsyncs on its own, so
// batching replaces one-fsync-per-statement with one fsync at COMMIT.
func applySchema(db *sql.DB) error {
	schema, err := Schema()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful Commit
	if _, err := tx.Exec(schema); err != nil {
		return err
	}
	if _, err := tx.Exec(ViewsDDL); err != nil {
		return err
	}
	return tx.Commit()
}

// Insert writes one event and its body inside a single transaction.
//
// ATOMIC BECAUSE A BODY WITHOUT ITS ENVELOPE IS NOT AN EVENT. The shard format could not express a
// half-written act — a torn line was simply dropped, and `ReadShard` counted it as an anomaly — so
// the failure mode moved rather than disappeared: an event with a row in `events` and none in its
// body table would replay as an act with no content, which reads as a seat that did nothing.
func Insert(db *sql.DB, ev *recordpb.Event) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // the commit below is what matters; a rollback after it is a no-op
	// THE ID COMES BACK FROM InsertTx, NOT FROM last_insert_rowid().
	//
	// The first split of this function asked SQLite for the last inserted rowid after InsertTx
	// returned — which is the id of the LAST row written, and InsertTx writes the envelope, then
	// the body, then a row per repeated value. For a mint with two `supersedes` entries it handed
	// back the id of a `mint_supersedes` row, and every caller then looked up an event that was
	// not there.
	id, err := InsertTx(tx, ev)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// InsertTx writes one event inside a transaction the CALLER owns.
//
// The write path needs that: an event's seq and its idempotency ordinal are counted from the rows
// already present, and a count in one transaction followed by an insert in another is a
// read-then-write with a gap in it. Here the count, the derivation and the insert are one atomic
// unit, which is a guarantee the shard layout got by giving each seat its own file and lost the
// moment two processes shared one.
func InsertTx(tx *sql.Tx, ev *recordpb.Event) (int64, error) {
	res, err := tx.Exec(
		`INSERT INTO events (seat_id, round, ts, type, key) VALUES (?, ?, ?, ?, ?)`,
		ev.GetSeatId(), ev.GetRound(), ev.GetTs(),
		recordpb.Word(ev.GetType()), nullable(ev.GetKey()),
	)
	if err != nil {
		return 0, fmt.Errorf("recordsql: recording the event: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	body, ok := recordpb.Body(ev)
	if !ok {
		return 0, fmt.Errorf("recordsql: event %d carries no body — every EventType names a body message", id)
	}
	return id, insertBody(tx, id, body.ProtoReflect())
}

// insertBody writes one message into its table, and recurses for the message arms of a oneof.
//
// It walks the SAME descriptors the DDL was derived from, so a field the schema has and the writer
// skips is not possible: both loops are the same loop over the same objects in the same process.
// A generated mapper would be a second derivation of one source, committed at a different moment
// from the schema, and the two could disagree.
func insertBody(tx *sql.Tx, eventID int64, m protoreflect.Message) error {
	md := m.Descriptor()
	table := TableName(md)

	cols := []string{`"event_id"`}
	vals := []any{eventID}
	var lists []protoreflect.FieldDescriptor
	var arms []protoreflect.FieldDescriptor

	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsList() {
			if m.Has(fd) {
				lists = append(lists, fd)
			}
			continue
		}
		if fd.Message() != nil {
			if m.Has(fd) {
				arms = append(arms, fd)
			}
			continue
		}
		if !m.Has(fd) {
			// ABSENT AND ZERO ARE DIFFERENT and the column is nullable for that reason. Writing a
			// zero here would record a grade the seat never gave, or a verdict for a run that
			// reached none.
			continue
		}
		cols = append(cols, fmt.Sprintf("%q", fd.Name()))
		vals = append(vals, value(fd, m.Get(fd)))
	}

	// The oneof's discriminator, so a reader knows which arm to look for without probing every
	// child table.
	for _, fd := range arms {
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			cols = append(cols, fmt.Sprintf("%q", string(od.Name())+"_case"))
			vals = append(vals, string(fd.Name()))
		}
	}

	stmt := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table,
		strings.Join(cols, ", "), strings.TrimSuffix(strings.Repeat("?, ", len(vals)), ", "))
	if _, err := tx.Exec(stmt, vals...); err != nil {
		return fmt.Errorf("recordsql: recording a %s: %w", table, err)
	}

	for _, fd := range lists {
		l := m.Get(fd).List()
		for i := 0; i < l.Len(); i++ {
			if _, err := tx.Exec(
				fmt.Sprintf("INSERT INTO %q (\"event_id\", \"ord\", \"value\") VALUES (?, ?, ?)", table+"_"+string(fd.Name())),
				eventID, i, l.Get(i).String(),
			); err != nil {
				return fmt.Errorf("recordsql: recording %s.%s: %w", table, fd.Name(), err)
			}
		}
	}
	for _, fd := range arms {
		if err := insertArm(tx, eventID, table+"_"+string(fd.Name()), m.Get(fd).Message()); err != nil {
			return err
		}
	}
	return nil
}

func insertArm(tx *sql.Tx, eventID int64, table string, m protoreflect.Message) error {
	md := m.Descriptor()
	cols := []string{`"event_id"`}
	vals := []any{eventID}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if !m.Has(fd) || fd.IsList() || fd.Message() != nil {
			continue
		}
		cols = append(cols, fmt.Sprintf("%q", fd.Name()))
		vals = append(vals, value(fd, m.Get(fd)))
	}
	stmt := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table,
		strings.Join(cols, ", "), strings.TrimSuffix(strings.Repeat("?, ", len(vals)), ", "))
	if _, err := tx.Exec(stmt, vals...); err != nil {
		return fmt.Errorf("recordsql: recording a %s: %w", table, err)
	}
	return nil
}

// value renders one field for SQL. An enum goes in as its WORD, which is what the vocabulary table
// holds and what a human reading the record with a plain SELECT expects to see.
func value(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.EnumKind:
		return recordpb.Word(enumAt{fd.Enum(), v.Enum()})
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return v.Int()
	case protoreflect.BytesKind:
		return v.Bytes()
	}
	return v.String()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type enumAt struct {
	ed protoreflect.EnumDescriptor
	n  protoreflect.EnumNumber
}

func (e enumAt) Descriptor() protoreflect.EnumDescriptor { return e.ed }
func (e enumAt) Type() protoreflect.EnumType             { return nil }
func (e enumAt) Number() protoreflect.EnumNumber         { return e.n }

var _ = proto.Marshal
