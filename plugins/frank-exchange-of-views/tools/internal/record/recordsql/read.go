package recordsql

import (
	"database/sql"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Events reads the whole record back, in the order the acts happened.
//
// # The ordering is the database's, not a merge's
//
// The shard format had one file per seat and merged them on read, sorting by timestamp because
// nothing else spanned the files. That sort IS the ordering hazard: two seats' clocks, one order,
// and a ruling that could land before its filing. Here `id` is assigned by the database at insert,
// so the order events are read in is the order they were RECORDED in — one sequence, no merge, and
// no second pass to repair it.
//
// # Anomalies are gone rather than relocated
//
// `ReadShard` returned a list of lines it could not parse: a torn write left a half line, and the
// reader's job was to drop it and say so. A transaction cannot leave a half event, so there is no
// anomaly to report and no honest-zero problem — the absence of anomalies is now a fact rather than
// a hope, which is why this signature has no place to put them.
func Events(db *sql.DB) ([]*recordpb.Event, error) {
	rows, err := db.Query(`SELECT id, seat_id, round, seq, nonce, ts, type, key FROM events ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("recordsql: reading the record: %w", err)
	}
	defer rows.Close()

	var out []*recordpb.Event
	var ids []int64
	for rows.Next() {
		var id int64
		var seatID, nonce, ts, typ string
		var round, seq int32
		var key *string
		if err := rows.Scan(&id, &seatID, &round, &seq, &nonce, &ts, &typ, &key); err != nil {
			return nil, err
		}
		t, ok := eventTypeOf(typ)
		if !ok {
			// A TYPE THE SCHEMA DOES NOT DECLARE IS NOT A ROW TO SKIP. The column has a foreign key
			// to nothing — `events.type` is the one enum written before its vocabulary is known —
			// so this is the one place a stored word can fail to resolve, and folding it into the
			// zero would replay the act as an event of no type at all.
			return nil, fmt.Errorf("recordsql: event %d has type %q, which the schema does not declare", id, typ)
		}
		ev := &recordpb.Event{
			SeatId: proto.String(seatID),
			Round:  proto.Int32(round),
			Seq:    proto.Int32(seq),
			Nonce:  proto.String(nonce),
			Ts:     proto.String(ts),
			Type:   &t,
		}
		if key != nil {
			ev.Key = key
		}
		out = append(out, ev)
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, ev := range out {
		if err := loadBody(db, ids[i], ev); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func eventTypeOf(word string) (recordpb.EventType, bool) {
	ed := recordpb.EventType(0).Descriptor()
	vd, ok := recordpb.BySpelling(ed, word)
	if !ok {
		return recordpb.EventType_EVENT_TYPE_UNSPECIFIED, false
	}
	return recordpb.EventType(vd.Number()), true
}

// loadBody fills an event's body from its table, by the same descriptor walk the write used.
//
// The message is minted BY THE EVENT — `NewField` on the oneof arm — rather than built dynamically
// and assigned. A dynamic message satisfies the descriptor and not the Go type, so it round-trips
// through the schema and then fails at the first `BodyAs[*recordpb.Mint]`, which is every reader.
// Asking the event for its own field gets the concrete type by construction.
func loadBody(db *sql.DB, id int64, ev *recordpb.Event) error {
	evm := ev.ProtoReflect()
	od := evm.Descriptor().Oneofs().ByName("body")
	arm := od.Fields().ByName(protoreflect.Name(recordpb.Word(ev.GetType())))
	if arm == nil || arm.Message() == nil {
		return fmt.Errorf("recordsql: no body message for %s", recordpb.Word(ev.GetType()))
	}
	msg := evm.NewField(arm).Message()
	md := arm.Message()

	if err := scanInto(db, TableName(md), id, msg); err != nil {
		return err
	}
	// Repeated fields come back as their own rows, in the order they were written.
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if !fd.IsList() {
			continue
		}
		vals, err := scanList(db, TableName(md)+"_"+string(fd.Name()), id)
		if err != nil {
			return err
		}
		if len(vals) == 0 {
			continue
		}
		l := msg.Mutable(fd).List()
		for _, v := range vals {
			l.Append(protoreflect.ValueOfString(v))
		}
	}
	// The oneof's message arms, found through the discriminator the write recorded.
	for i := 0; i < md.Oneofs().Len(); i++ {
		sub := md.Oneofs().Get(i)
		if sub.IsSynthetic() {
			continue
		}
		which, err := armOf(db, TableName(md), id, string(sub.Name())+"_case")
		if err != nil {
			return err
		}
		if which == "" {
			continue
		}
		fd := sub.Fields().ByName(protoreflect.Name(which))
		if fd == nil || fd.Message() == nil {
			continue
		}
		inner := msg.NewField(fd).Message()
		if err := scanInto(db, TableName(md)+"_"+which, id, inner); err != nil {
			return err
		}
		msg.Set(fd, protoreflect.ValueOfMessage(inner))
	}
	evm.Set(arm, protoreflect.ValueOfMessage(msg))
	return nil
}

// scanInto reads one row into a message, setting ONLY the columns that are non-NULL.
//
// A NULL column leaves the field UNSET rather than zero, which is the round trip the whole schema
// turns on: a mint with no severity comes back ungraded, not graded zero, and `Has` answers the
// same question after a write-read cycle that it answered before one.
func scanInto(db *sql.DB, table string, id int64, msg protoreflect.Message) error {
	md := msg.Descriptor()
	var fields []protoreflect.FieldDescriptor
	var names []string
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsList() || fd.Message() != nil {
			continue
		}
		fields = append(fields, fd)
		names = append(names, fmt.Sprintf("%q", fd.Name()))
	}
	if len(fields) == 0 {
		return nil
	}
	cols := make([]any, len(fields))
	for i := range cols {
		cols[i] = new(any)
	}
	q := "SELECT " + join(names, ", ") + fmt.Sprintf(" FROM %q WHERE \"event_id\" = ?", table)
	if err := db.QueryRow(q, id).Scan(cols...); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("recordsql: event %d has no %s body — an envelope with no body replays as a seat that did nothing", id, table)
		}
		return fmt.Errorf("recordsql: reading %s for event %d: %w", table, id, err)
	}
	for i, fd := range fields {
		v := *(cols[i].(*any))
		if v == nil {
			continue // absent, and absent is not zero
		}
		pv, err := protoValue(fd, v)
		if err != nil {
			return fmt.Errorf("recordsql: %s.%s: %w", table, fd.Name(), err)
		}
		msg.Set(fd, pv)
	}
	return nil
}

func scanList(db *sql.DB, table string, id int64) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT \"value\" FROM %q WHERE \"event_id\" = ? ORDER BY \"ord\"", table), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func armOf(db *sql.DB, table string, id int64, col string) (string, error) {
	var arm *string
	err := db.QueryRow(fmt.Sprintf("SELECT %q FROM %q WHERE \"event_id\" = ?", col, table), id).Scan(&arm)
	if err != nil || arm == nil {
		return "", nil
	}
	return *arm, nil
}

// protoValue turns a stored column back into a field value. An enum resolves through BySpelling, so
// a word the schema stopped declaring FAILS here rather than becoming the zero — which would read
// as a seat that never answered.
func protoValue(fd protoreflect.FieldDescriptor, v any) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.EnumKind:
		s, _ := v.(string)
		vd, ok := recordpb.BySpelling(fd.Enum(), s)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("%q is not a value of %s", s, fd.Enum().FullName())
		}
		return protoreflect.ValueOfEnum(vd.Number()), nil
	case protoreflect.BoolKind:
		if n, ok := v.(int64); ok {
			return protoreflect.ValueOfBool(n != 0), nil
		}
		b, _ := v.(bool)
		return protoreflect.ValueOfBool(b), nil
	case protoreflect.Int32Kind:
		n, _ := v.(int64)
		return protoreflect.ValueOfInt32(int32(n)), nil
	case protoreflect.Int64Kind:
		n, _ := v.(int64)
		return protoreflect.ValueOfInt64(n), nil
	case protoreflect.BytesKind:
		b, _ := v.([]byte)
		return protoreflect.ValueOfBytes(b), nil
	}
	switch s := v.(type) {
	case string:
		return protoreflect.ValueOfString(s), nil
	case []byte:
		return protoreflect.ValueOfString(string(s)), nil
	}
	return protoreflect.ValueOfString(fmt.Sprint(v)), nil
}

func join(xs []string, sep string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += sep
		}
		out += x
	}
	return out
}
