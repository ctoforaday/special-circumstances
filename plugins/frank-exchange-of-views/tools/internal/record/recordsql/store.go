package recordsql

import (
	"database/sql"
	"fmt"
	"strings"

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
func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='events'`).Scan(&n); err != nil {
		db.Close()
		return nil, fmt.Errorf("recordsql: reading %s: %w", path, err)
	}
	if n == 0 {
		schema, err := Schema()
		if err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.Exec(schema); err != nil {
			db.Close()
			return nil, fmt.Errorf("recordsql: applying the schema to %s: %w", path, err)
		}
	}
	return db, nil
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

	res, err := tx.Exec(
		`INSERT INTO events (seat_id, round, seq, nonce, ts, type, key) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ev.GetSeatId(), ev.GetRound(), ev.GetSeq(), ev.GetNonce(), ev.GetTs(),
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
	if err := insertBody(tx, id, body.ProtoReflect()); err != nil {
		return 0, err
	}
	return id, tx.Commit()
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
