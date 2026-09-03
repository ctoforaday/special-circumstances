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
//
// # Bodies load per TABLE, not per event
//
// This read is always the WHOLE record, and every body-table row belongs to it — `event_id` is a
// primary key referencing `events(id)`, so a full scan of each table touched by the record reads
// exactly the rows the per-event queries used to fetch one at a time. That turns a query count
// proportional to the number of EVENTS (one body row, one per list field, two per oneof arm — the
// N+1 every projection paid on every render) into one proportional to the number of TABLES the
// record's event types use, which the schema bounds. Validation still walks the events in record
// order, so the first refusal an interleaved read would have raised is the same refusal this one
// raises.
func Events(db *sql.DB) ([]*recordpb.Event, error) {
	rows, err := db.Query(`SELECT id, seat_id, round, ts, type, key FROM events ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("recordsql: reading the record: %w", err)
	}
	defer rows.Close()

	var out []*recordpb.Event
	var ids []int64
	for rows.Next() {
		var id int64
		var seatID, ts, typ string
		var round int32
		var key *string
		if err := rows.Scan(&id, &seatID, &round, &ts, &typ, &key); err != nil {
			return nil, err
		}
		t, ok := eventTypeOf(typ)
		if !ok {
			// A TYPE THE SCHEMA DOES NOT DECLARE IS NOT A ROW TO SKIP. `events.type` references
			// `enum_event_type` now, so a word outside the vocabulary cannot be stored — reaching
			// here means the record disagrees with its own schema, and folding that into the zero
			// would replay the act as an event of no type at all.
			return nil, fmt.Errorf("recordsql: event %d has type %q, which the schema does not declare", id, typ)
		}
		ev := &recordpb.Event{
			SeatId: proto.String(seatID),
			Round:  proto.Int32(round),
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
	if err := loadBodies(db, ids, out); err != nil {
		return nil, err
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

// tableBatch is one body table's whole contribution to the read: its descriptor walk done once,
// its rows, lists and arms each fetched by one scan.
type tableBatch struct {
	md      protoreflect.MessageDescriptor
	table   string
	scalars []protoreflect.FieldDescriptor // non-list, non-message: the row's columns
	oneofs  []protoreflect.OneofDescriptor // non-synthetic: each carries a `_case` discriminator
	rows    map[int64][]any                // scalar values, then one case word per oneof, aligned
	lists   map[string]map[int64][]string  // list-field name → event → values in ord order
	arms    map[string]*armBatch           // arm-field name → its table's scan
}

type armBatch struct {
	scalars []protoreflect.FieldDescriptor
	rows    map[int64][]any
}

// loadBodies fills every event's body from its table, by the same descriptor walk the write used.
//
// The message is minted BY THE EVENT — `NewField` on the oneof arm — rather than built dynamically
// and assigned. A dynamic message satisfies the descriptor and not the Go type, so it round-trips
// through the schema and then fails at the first `BodyAs[*recordpb.Mint]`, which is every reader.
// Asking the event for its own field gets the concrete type by construction.
func loadBodies(db *sql.DB, ids []int64, evs []*recordpb.Event) error {
	// The event's own arm on the envelope, resolved per event; the batches, resolved per table.
	arms := make([]protoreflect.FieldDescriptor, len(evs))
	msgs := make([]protoreflect.Message, len(evs))
	batches := map[string]*tableBatch{}
	for i, ev := range evs {
		evm := ev.ProtoReflect()
		od := evm.Descriptor().Oneofs().ByName("body")
		arm := od.Fields().ByName(protoreflect.Name(recordpb.Word(ev.GetType())))
		if arm == nil || arm.Message() == nil {
			return fmt.Errorf("recordsql: no body message for %s", recordpb.Word(ev.GetType()))
		}
		arms[i] = arm
		msgs[i] = evm.NewField(arm).Message()
		md := arm.Message()
		if _, ok := batches[TableName(md)]; !ok {
			batches[TableName(md)] = newTableBatch(md)
		}
	}
	for _, b := range batches {
		if err := b.load(db); err != nil {
			return err
		}
	}
	// Validation walks the events in record order, so which refusal fires first does not depend
	// on which table happened to load first.
	for i, ev := range evs {
		b := batches[TableName(arms[i].Message())]
		if err := b.fill(ids[i], msgs[i]); err != nil {
			return err
		}
		ev.ProtoReflect().Set(arms[i], protoreflect.ValueOfMessage(msgs[i]))
	}
	return nil
}

func newTableBatch(md protoreflect.MessageDescriptor) *tableBatch {
	b := &tableBatch{md: md, table: TableName(md)}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsList() || fd.Message() != nil {
			continue
		}
		b.scalars = append(b.scalars, fd)
	}
	for i := 0; i < md.Oneofs().Len(); i++ {
		if od := md.Oneofs().Get(i); !od.IsSynthetic() {
			b.oneofs = append(b.oneofs, od)
		}
	}
	return b
}

func (b *tableBatch) load(db *sql.DB) error {
	var cols []string
	for _, fd := range b.scalars {
		cols = append(cols, fmt.Sprintf("%q", fd.Name()))
	}
	for _, od := range b.oneofs {
		cols = append(cols, fmt.Sprintf("%q", string(od.Name())+"_case"))
	}
	var err error
	if b.rows, err = scanTable(db, b.table, cols); err != nil {
		return err
	}
	b.lists = map[string]map[int64][]string{}
	for i := 0; i < b.md.Fields().Len(); i++ {
		fd := b.md.Fields().Get(i)
		if !fd.IsList() {
			continue
		}
		vals, err := scanLists(db, b.table+"_"+string(fd.Name()))
		if err != nil {
			return err
		}
		b.lists[string(fd.Name())] = vals
	}
	// Only the arms some event in the record actually filed get a scan — the discriminators
	// just read say which those are.
	b.arms = map[string]*armBatch{}
	for oi, od := range b.oneofs {
		for _, row := range b.rows {
			which := wordAt(row, len(b.scalars)+oi)
			if which == "" {
				continue
			}
			fd := od.Fields().ByName(protoreflect.Name(which))
			if fd == nil || fd.Message() == nil {
				continue
			}
			if _, ok := b.arms[which]; ok {
				continue
			}
			a := &armBatch{}
			amd := fd.Message()
			var acols []string
			for i := 0; i < amd.Fields().Len(); i++ {
				afd := amd.Fields().Get(i)
				if afd.IsList() || afd.Message() != nil {
					continue
				}
				a.scalars = append(a.scalars, afd)
				acols = append(acols, fmt.Sprintf("%q", afd.Name()))
			}
			if a.rows, err = scanTable(db, b.table+"_"+which, acols); err != nil {
				return err
			}
			b.arms[which] = a
		}
	}
	return nil
}

// fill assembles one event's body from the batch, setting ONLY the columns that were non-NULL.
//
// A NULL column leaves the field UNSET rather than zero, which is the round trip the whole schema
// turns on: a mint with no severity comes back ungraded, not graded zero, and `Has` answers the
// same question after a write-read cycle that it answered before one.
func (b *tableBatch) fill(id int64, msg protoreflect.Message) error {
	row, ok := b.rows[id]
	if !ok && len(b.scalars) > 0 {
		return fmt.Errorf("recordsql: event %d has no %s body — an envelope with no body replays as a seat that did nothing", id, b.table)
	}
	for i, fd := range b.scalars {
		v := row[i]
		if v == nil {
			continue // absent, and absent is not zero
		}
		pv, err := protoValue(fd, v)
		if err != nil {
			return fmt.Errorf("recordsql: %s.%s: %w", b.table, fd.Name(), err)
		}
		msg.Set(fd, pv)
	}
	// Repeated fields came back as their own rows, in the order they were written.
	for name, vals := range b.lists {
		v := vals[id]
		if len(v) == 0 {
			continue
		}
		l := msg.Mutable(b.md.Fields().ByName(protoreflect.Name(name))).List()
		for _, s := range v {
			l.Append(protoreflect.ValueOfString(s))
		}
	}
	// The oneof's message arms, found through the discriminator the write recorded.
	for oi, od := range b.oneofs {
		which := wordAt(row, len(b.scalars)+oi)
		if which == "" {
			continue
		}
		fd := od.Fields().ByName(protoreflect.Name(which))
		if fd == nil || fd.Message() == nil {
			continue
		}
		a := b.arms[which]
		arow, ok := a.rows[id]
		if !ok && len(a.scalars) > 0 {
			return fmt.Errorf("recordsql: event %d has no %s body — an envelope with no body replays as a seat that did nothing", id, b.table+"_"+which)
		}
		inner := msg.NewField(fd).Message()
		for i, afd := range a.scalars {
			v := arow[i]
			if v == nil {
				continue
			}
			pv, err := protoValue(afd, v)
			if err != nil {
				return fmt.Errorf("recordsql: %s.%s: %w", b.table+"_"+which, afd.Name(), err)
			}
			inner.Set(afd, pv)
		}
		msg.Set(fd, protoreflect.ValueOfMessage(inner))
	}
	return nil
}

// scanTable reads a whole body table into memory, keyed by event. Every row belongs to this read:
// `event_id` is a primary key referencing `events(id)`, and the read is always the whole record.
func scanTable(db *sql.DB, table string, cols []string) (map[int64][]any, error) {
	out := map[int64][]any{}
	if len(cols) == 0 {
		return out, nil
	}
	q := `SELECT "event_id", ` + join(cols, ", ") + fmt.Sprintf(" FROM %q", table)
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("recordsql: reading %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		targets := make([]any, len(cols)+1)
		var id int64
		targets[0] = &id
		for i := 1; i < len(targets); i++ {
			targets[i] = new(any)
		}
		if err := rows.Scan(targets...); err != nil {
			return nil, fmt.Errorf("recordsql: reading %s: %w", table, err)
		}
		vals := make([]any, len(cols))
		for i := range vals {
			vals[i] = *(targets[i+1].(*any))
		}
		out[id] = vals
	}
	return out, rows.Err()
}

func scanLists(db *sql.DB, table string) (map[int64][]string, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT \"event_id\", \"value\" FROM %q ORDER BY \"event_id\", \"ord\"", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][]string{}
	for rows.Next() {
		var id int64
		var v string
		if err := rows.Scan(&id, &v); err != nil {
			return nil, err
		}
		out[id] = append(out[id], v)
	}
	return out, rows.Err()
}

// wordAt is a discriminator column read back: TEXT, or NULL when the oneof was never filed.
func wordAt(row []any, i int) string {
	if i >= len(row) {
		return ""
	}
	switch v := row[i].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return ""
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
