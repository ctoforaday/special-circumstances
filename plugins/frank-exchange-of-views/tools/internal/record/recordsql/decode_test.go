package recordsql

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// fillShape is how a fixture body is populated.
type fillShape int

const (
	// fillEvery sets every field: each scalar to a non-zero value, each list to three elements, and
	// the oneof arm the event was built for.
	fillEvery fillShape = iota
	// fillZero sets every field with presence to its zero — "", 0, false — which must read back SET,
	// not absent. An enum's zero has no word to store, so an enum stays unset, or takes a word where
	// the column is NOT NULL.
	fillZero
	// fillRequired sets only the fields the schema makes NOT NULL, which must read back as the only
	// fields set: NULL is absent, not zero.
	fillRequired
)

// fillBody populates m per shape, choosing arm k of each real oneof; n makes every value unique to
// this event, so no UNIQUE column refuses a second fixture.
func fillBody(m protoreflect.Message, shape fillShape, k, n int) {
	md := m.Descriptor()
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			if shape == fillRequired || od.Fields().Get(k%od.Fields().Len()) != fd {
				continue
			}
		}
		if shape == fillRequired && !Opts(fd).GetRequired() {
			continue
		}
		switch {
		case fd.IsList():
			if shape != fillEvery {
				continue
			}
			l := m.Mutable(fd).List()
			for e := 0; e < 3; e++ {
				l.Append(fixtureValue(fd, fillEvery, n*10+e))
			}
		case fd.Message() != nil:
			inner := m.Mutable(fd).Message()
			fillBody(inner, shape, k, n)
		default:
			s := shape
			if shape == fillZero && fd.Kind() == protoreflect.EnumKind {
				if !Opts(fd).GetRequired() {
					continue
				}
				s = fillEvery
			}
			if s == fillRequired || (s == fillZero && Opts(fd).GetUnique()) {
				s = fillEvery // a UNIQUE column holds one zero, and every fixture writes one
			}
			m.Set(fd, fixtureValue(fd, s, n))
		}
	}
}

// fixtureValue is a value of fd's kind: non-zero and distinct per n, or the zero.
func fixtureValue(fd protoreflect.FieldDescriptor, shape fillShape, n int) protoreflect.Value {
	zero := shape == fillZero
	switch fd.Kind() {
	case protoreflect.StringKind:
		if zero {
			return protoreflect.ValueOfString("")
		}
		return protoreflect.ValueOfString(fmt.Sprintf("%s-%d", fd.Name(), n))
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(!zero)
	case protoreflect.Int32Kind:
		if zero {
			return protoreflect.ValueOfInt32(0)
		}
		return protoreflect.ValueOfInt32(int32(n + 1))
	case protoreflect.Int64Kind:
		if zero {
			return protoreflect.ValueOfInt64(0)
		}
		return protoreflect.ValueOfInt64(int64(n) + 1<<40)
	case protoreflect.BytesKind:
		if zero {
			return protoreflect.ValueOfBytes([]byte{})
		}
		return protoreflect.ValueOfBytes([]byte(fmt.Sprint(n)))
	case protoreflect.EnumKind:
		// Every non-zero value in turn, across the fixtures.
		vals := fd.Enum().Values()
		return protoreflect.ValueOfEnum(vals.Get(1 + n%(vals.Len()-1)).Number())
	}
	panic(fmt.Sprintf("%s: no fixture for kind %s", fd.FullName(), fd.Kind()))
}

// maxArms is the most arms any real oneof of md has, so k = 0..maxArms-1 reaches every arm.
func maxArms(md protoreflect.MessageDescriptor) int {
	most := 1
	for i := 0; i < md.Oneofs().Len(); i++ {
		if od := md.Oneofs().Get(i); !od.IsSynthetic() && od.Fields().Len() > most {
			most = od.Fields().Len()
		}
	}
	return most
}

// EVERY BODY READS BACK AS IT WAS WRITTEN, through the generated decode.
//
// The arms come from Event's `body` oneof through its descriptor, so a body the generator has no
// reader for fails here as `no body message`, and the fields come from each body's descriptor, so a
// field the generator does not assign reads back unset and fails proto.Equal. The oracle is the
// written event itself: the write path stores what Has reports, NULL for the rest, and a decode that
// matches it returns exactly what was written — set-to-zero as set, unset as unset, lists in order.
//
// THE ORACLE IS THE EVENT AFTER THE INSERT, because InsertTx stamps a regrade's result grades onto
// the body it writes (stampRegradeResult), overwriting the fixture's: the event Insert was handed is
// the event the record holds. A fixture regrade names no minted gap, so its results are stamped
// absent here; TestARegradeIsStampedWithTheComputedGradesOverWhatItCarries reads populated ones back
// through the same decode.
//
// The fixtures are not valid acts — every field at once, references to rows that do not exist — so
// the connection writes them with CHECK and foreign-key enforcement off. This tests the decode; the
// constraints have tests of their own.
func TestEveryBodyRoundTripsThroughTheGeneratedDecode(t *testing.T) {
	db := store(t)
	written, ids := writeEveryBody(t, db)
	got, _, err := Events(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(written) {
		t.Fatalf("read %d events back, wrote %d", len(got), len(written))
	}
	for i := range written {
		if !proto.Equal(got[i], written[i]) {
			t.Errorf("event %d (%s) read back differently:\n  wrote %v\n  read  %v",
				ids[i], recordpb.Word(written[i].GetType()), written[i], got[i])
		}
	}
	// The read by one row goes through the same decode.
	for i := range written {
		one, err := EventAt(db, ids[i])
		if err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(one, written[i]) {
			t.Errorf("EventAt(%d) read back differently:\n  wrote %v\n  read  %v", ids[i], written[i], one)
		}
	}
}

// A WORD THE SCHEMA DOES NOT DECLARE IS REFUSED, not read as the enum's zero — which would be a seat
// that never answered — and the refusal names the column.
func TestTheDecodeRefusesAnUndeclaredEnumWord(t *testing.T) {
	db := store(t)
	mint := &recordpb.Mint{}
	fillBody(mint.ProtoReflect(), fillEvery, 0, 1)
	ev := event(t, 1, recordpb.EventType_EVENT_TYPE_MINT, mint)
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA ignore_check_constraints = ON`); err != nil {
		t.Fatal(err)
	}
	id, err := Insert(db, ev)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE "mint" SET "severity" = 'catastrophic' WHERE "event_id" = ?`, id); err != nil {
		t.Fatal(err)
	}
	_, _, err = Events(db)
	if err == nil {
		t.Fatal("a severity the schema does not declare read back without a refusal")
	}
	for _, want := range []string{"mint.severity", `"catastrophic"`, "is not a value of"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %s: %v", want, err)
		}
	}
}

// writeEveryBody writes, for every arm of Event's `body` oneof, one fixture per fillShape and per arm
// of the body's own oneofs, and returns them with their ids.
func writeEveryBody(t *testing.T, db *sql.DB) ([]*recordpb.Event, []int64) {
	t.Helper()
	for _, p := range []string{`PRAGMA foreign_keys = OFF`, `PRAGMA ignore_check_constraints = ON`} {
		if _, err := db.Exec(p); err != nil {
			t.Fatal(err)
		}
	}
	// AND WITH THE REFUSING TRIGGERS GONE: a refusal that reads another row (a docket ruling's filing)
	// is a trigger, which no pragma turns off.
	refusals, err := db.Query(`SELECT "name" FROM sqlite_master WHERE "type" = 'trigger' AND "sql" LIKE '%BEFORE INSERT%RAISE%'`)
	if err != nil {
		t.Fatal(err)
	}
	var drop []string
	for refusals.Next() {
		var name string
		if err := refusals.Scan(&name); err != nil {
			t.Fatal(err)
		}
		drop = append(drop, name)
	}
	refusals.Close()
	for _, name := range drop {
		if _, err := db.Exec(`DROP TRIGGER "` + name + `"`); err != nil {
			t.Fatal(err)
		}
	}
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		t.Fatal("Event has no `body` oneof")
	}
	var written []*recordpb.Event
	var ids []int64
	n := 0
	for i := 0; i < od.Fields().Len(); i++ {
		arm := od.Fields().Get(i)
		if _, ok := bodyTables[string(arm.Name())]; !ok {
			t.Errorf("the body arm %s has no generated reader — run `go -C scripts run ./protogen`", arm.Name())
		}
		for _, shape := range []fillShape{fillEvery, fillZero, fillRequired} {
			for k := 0; k < maxArms(arm.Message()); k++ {
				n++
				ev := &recordpb.Event{
					SeatId: proto.String("red-chair"),
					Ts:     proto.String("2026-01-01T00:00:00Z"),
					Key:    proto.String(fmt.Sprintf("fixture:%d", n)),
				}
				evm := ev.ProtoReflect()
				body := evm.NewField(arm).Message()
				fillBody(body, shape, k, n)
				evm.Set(arm, protoreflect.ValueOfMessage(body))
				typ, ok := recordpb.EventTypeOf(string(arm.Name()))
				if !ok {
					t.Fatalf("the body arm %s names no event type", arm.Name())
				}
				ev.Type = &typ
				id, err := Insert(db, ev)
				if err != nil {
					t.Fatalf("writing a %s fixture (shape %d, arm %d): %v", arm.Name(), shape, k, err)
				}
				written = append(written, ev)
				ids = append(ids, id)
				if shape == fillRequired {
					break // no oneof arm is set, so every k writes the same body
				}
			}
		}
	}

	return written, ids
}
