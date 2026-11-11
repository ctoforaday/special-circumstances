package recordsql

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"

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
// A read of the whole record takes every body-table row past the events it already holds —
// `event_id` is a primary key referencing `events(id)`, so a range scan of each table touched by
// those events reads exactly the rows the per-event queries used to fetch one at a time. That turns a query count
// proportional to the number of EVENTS (one body row, one per list field, two per oneof arm — the
// N+1 every projection paid on every render) into one proportional to the number of TABLES the
// record's event types use, which the schema bounds. Validation still walks the events in record
// order, so the first refusal an interleaved read would have raised is the same refusal this one
// raises.
//
// # Each event comes back with the window it sits in
//
// The second slice is aligned with the first by position: where each event sits on the record,
// read off the row's stored sitting_id, sitting and epoch and that sitting's owner in ONE statement. A reader that
// needs which sitting an act belongs to takes it from here rather than counting openings over the
// slice — the write path decided that once (sittingOf), and a second count is a second definition.
//
// One statement is one snapshot. Read as two on a pooled handle, a sitting a seat opened between
// them loads with its id and no owner, and the seat's latest sitting is then the one before.
//
// # A handle reads each event once
//
// On a handle Open returned, and through a transaction BeginRead or BeginWrite opened on one, the
// events decoded before are kept and only the rows past them are read (eventcache.go says why that
// read is the whole record). The events returned are shared with every later read on the handle: a
// caller reads them and never edits one.
func Events(db Querier) ([]*recordpb.Event, []Window, error) {
	return cachedEvents(db)
}

// Window is where an event sits on the record: its row, the sitting it belongs to, and that
// sitting's place on the record's two windows (plans/roundless.md §III.A.0).
//
// SittingID is the stored fact — events.sitting_id, stamped once at the write by sittingOf — and
// the rest is read beside it. OWNER is the seat whose sitting that is (the `sittings` view's owner:
// a register's seat, or the seat a hook bracket's body names). SITTING is the rank of that sitting
// among its owner's, and EPOCH is how many of the chair's sittings opened at or before this row.
// Both numbers are columns of the row, stamped by the write beside SittingID (sittingOf), and
// nothing in Go counts them again.
//
// A row in no sitting — the harness's own bookkeeping, a cast, a seat's act before it ever sat —
// has SittingID 0, Owner "" and Sitting 0. An OPENING is a row whose SittingID is its own ID.
type Window struct {
	ID        int64
	SittingID int64
	Owner     string
	Epoch     int
	Sitting   int
}

// Opens is the seat whose sitting this row opened, and whether it opened one.
func (w Window) Opens() (string, bool) {
	return w.Owner, w.ID != 0 && w.SittingID == w.ID
}

// EventsOfTypes is Events narrowed to the named type words, in the same record order — for a
// projection that renders one family of acts and has no business hauling the rest of the record
// through the loader to get them. The words are the schema's own spellings (recordpb.Word); a
// word the vocabulary does not hold simply matches nothing, exactly as it would in the stream.
//
// Each event comes back with its Window, as from Events: the projections that narrow are the ones
// that print an epoch or a sitting number, and the stored window carries both whatever else the
// narrowing left out. A slice holding no chair act still knows which of the chair's sittings each
// of its rows falls in.
//
// THAT LAST SENTENCE IS TRUE OF THE ARGUMENT AND WAS ASSUMED OF THE DATABASE, and the two are not
// the same claim. `WHERE type IN (…)` matches nothing when the caller asks for a word no event
// carries — the honest zero — AND when the DATABASE spells its events with words this schema does
// not declare. Measured on run-archive/2026-09-02_quadratic-formula, whose 35 `friction` events
// this schema now calls something else:
//
//	recordsql.Events(db)                 -> 0, `event 8 has type "friction", which the schema does not declare`
//	recordsql.EventsOfTypes(db, "log")   -> 0, nil
//
// The full replay already refuses, at the row (see eventsWhere below). The narrowed read never
// reaches the row: SQL filtered it out first, and the projection rendered `{"log": [], "counts":
// {"total": 0}}` with exit 0 over a run holding 35 of them — a clean board, on the artifact
// CLAUDE.md says every audit re-reads.
//
// So the vocabulary is checked before the narrowing, and this is not a version negotiation: there
// is no epoch to compare and nothing here knows what the words used to be. It asks the one question
// the data can answer on its own — does this record spell its events in words this schema declares
// — and the schema's own enum is the whole authority.
func EventsOfTypes(db Querier, words ...string) ([]*recordpb.Event, []Window, error) {
	if len(words) == 0 {
		return nil, nil, nil
	}
	if err := refuseUndeclaredTypes(db); err != nil {
		return nil, nil, err
	}
	marks := strings.TrimSuffix(strings.Repeat("?, ", len(words)), ", ")
	args := make([]any, len(words))
	for i, w := range words {
		args[i] = w
	}
	return eventsWhere(db, true, 0, ` WHERE e."type" IN (`+marks+`)`, args...)
}

// EventAt is the event stored at row id, with its body, and nil where no row has that id. It reads
// no window: a read by one row prints no sitting.
func EventAt(db Querier, id int64) (*recordpb.Event, error) {
	evs, _, err := eventsWhere(db, false, 0, ` WHERE e."id" = ?`, id)
	if err != nil || len(evs) == 0 {
		return nil, err
	}
	return evs[0], nil
}

// eventsWhere is the one read path. withWindows reads each row's Window beside it — the stored
// sitting_id, sitting and epoch, and the sitting's owner — in one statement; without it the read is
// the events table alone. A read by one key prints no window and does not ask for one.
//
// THE OWNER IS LOOKED UP BY KEY, NOT THROUGH THE `sittings` VIEW. The view's owner is the opening
// event's seat, or the seat its sitting_open body names; the join below asks exactly that of the one
// opening event each row's sitting_id names — a rowid SEARCH per row. Joined through the view, SQLite
// materialises every sitting first, a scan of the whole events table that a tail read of five rows
// paid in full. The migrate package's archived-run test holds each owner to the view's.
//
// after bounds the body tables' reads to rows past that event id, and where MUST select no event at
// or before it: a body row the bound leaves out is a body the event is then refused for lacking. The
// stream cache's tail read (eventcache.go) is the one caller that passes more than 0.
func eventsWhere(db Querier, withWindows bool, after int64, where string, args ...any) ([]*recordpb.Event, []Window, error) {
	q := `SELECT e."id", e."seat_id", e."ts", e."type", e."key" FROM "events" e` + where + ` ORDER BY e."id"`
	if withWindows {
		q = `SELECT e."id", e."seat_id", e."ts", e."type", e."key", e."epoch", e."sitting",
		            COALESCE(e."sitting_id", 0),
		            CASE WHEN s."id" IS NULL THEN '' ELSE COALESCE(NULLIF(o."seat_id", ''), s."seat_id") END
		     FROM "events" e
		     LEFT JOIN "events" s ON s."id" = e."sitting_id" AND s."sitting_id" = s."id"
		     LEFT JOIN "sitting_open" o ON o."event_id" = s."id"` + where + ` ORDER BY e."id"`
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("reading the record: %w", err)
	}
	defer rows.Close()

	var out []*recordpb.Event
	var windows []Window
	var ids []int64
	for rows.Next() {
		var id int64
		var seatID, ts, typ string
		var key *string
		var w Window
		dst := []any{&id, &seatID, &ts, &typ, &key}
		if withWindows {
			dst = append(dst, &w.Epoch, &w.Sitting, &w.SittingID, &w.Owner)
		}
		if err := rows.Scan(dst...); err != nil {
			return nil, nil, err
		}
		w.ID = id
		t, ok := recordpb.EventTypeOf(typ)
		if !ok {
			return nil, nil, fmt.Errorf("event %d has type %q, which the schema does not declare", id, typ)
		}
		ev := &recordpb.Event{
			SeatId: proto.String(seatID),
			Ts:     proto.String(ts),
			Type:   &t,
		}
		if key != nil {
			ev.Key = key
		}
		out = append(out, ev)
		ids = append(ids, id)
		if withWindows {
			windows = append(windows, w)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if err := loadBodies(db, ids, out, after); err != nil {
		return nil, nil, err
	}
	return out, windows, nil
}

// refuseUndeclaredTypes fails a record that spells its events in words this schema does not
// declare, naming them.
//
// IT IS ONE INDEXED SCAN AND IT IS ON THE NARROWED PATH ONLY. `events.type` is indexed and the
// distinct set is a few dozen rows, so the cost is a covering-index walk per narrowed read; the
// full replay does not pay it because eventsWhere already refuses the row itself, one event at a
// time, and reaching that row is precisely what narrowing prevents.
//
// The error says what is there, not what it should have been. There is no former-name table here
// and there will not be one: a reader who needs those events needs the binary that wrote them —
// or the migration, which is that knowledge quarantined in one package (internal/record/migrate)
// instead of a negotiation in every reader.
func refuseUndeclaredTypes(db Querier) error {
	rows, err := db.Query(`SELECT DISTINCT type FROM events`)
	if err != nil {
		return fmt.Errorf("reading the record's event types: %w", err)
	}
	defer rows.Close()
	var undeclared []string
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return err
		}
		if _, ok := recordpb.EventTypeOf(word); !ok {
			undeclared = append(undeclared, word)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(undeclared) == 0 {
		return nil
	}
	sort.Strings(undeclared)
	return fmt.Errorf("this record holds event types the schema does not declare (%s), so a "+
		"narrowed read of it cannot be trusted: the rows are there and SQL filters them out before "+
		"anything looks at them, which returns an empty projection and exit 0 — the same bytes as a "+
		"record that holds nothing. Read it with the binary that wrote it, or migrate it: "+
		"`feov-migrate --from <runDir> --to <freshDir>` — the plugin's operator binary, installed in its bin/ "+
		"beside feov-record — re-drives every event through the current write path into a fresh run "+
		"and says what it translated",
		strings.Join(undeclared, ", "))
}

// bodyTable is one body message's reader, generated by scripts/protogen into decode_gen.go: load
// reads its table, list tables and oneof arm tables whole, keyed by event, scanning each column
// straight into its typed Go field; attach puts one event's body on it.
//
// # Every identifier is table-qualified, so a missing column is an error
//
// SQLite reads a double-quoted identifier that names no column as a STRING LITERAL — `SELECT
// "exact_span" FROM "blue_edit"` succeeds and hands back the word as every row's value — but only
// where the name stands alone. A qualified `t."exact_span"` resolves as a column or fails to prepare
// with `no such column`. The generated statements qualify every column, so the read refuses a column
// the table lacks without asking the table for its columns first.
type bodyTable interface {
	load(q Querier, after int64) error
	attach(ev *recordpb.Event, id int64) error
}

// tableOfType is bodyTables keyed by event type rather than word, resolved once: the word is the
// arm's name and EventTypeOf is the one word → type lookup.
var tableOfType = sync.OnceValue(func() map[recordpb.EventType]func() bodyTable {
	out := make(map[recordpb.EventType]func() bodyTable, len(bodyTables))
	for word, mk := range bodyTables {
		if t, ok := recordpb.EventTypeOf(word); ok {
			out[t] = mk
		}
	}
	return out
})

// loadBodies fills every event's body from its table.
//
// Each table touched by the record is read once, from the row past after onward (0 is the whole
// table); validation then walks the events in record order, so which refusal fires first does not
// depend on which table loaded first.
func loadBodies(db Querier, ids []int64, evs []*recordpb.Event, after int64) error {
	readers := tableOfType()
	tables := map[recordpb.EventType]bodyTable{}
	for _, ev := range evs {
		t := ev.GetType()
		if _, ok := tables[t]; ok {
			continue
		}
		mk, ok := readers[t]
		if !ok {
			return fmt.Errorf("no body message for %s", recordpb.Word(t))
		}
		tables[t] = mk()
	}
	for _, t := range tables {
		if err := t.load(db, after); err != nil {
			return err
		}
	}
	for i, ev := range evs {
		if err := tables[ev.GetType()].attach(ev, ids[i]); err != nil {
			return err
		}
	}
	return nil
}

// scanRows runs one generated SELECT, scanning each row into targets and handing it to row.
func scanRows(q Querier, table, query string, after int64, targets []any, row func()) error {
	rows, err := q.Query(query, after)
	if err != nil {
		return fmt.Errorf("reading %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(targets...); err != nil {
			return fmt.Errorf("reading %s: %w", table, err)
		}
		row()
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading %s: %w", table, err)
	}
	return nil
}

// scanList runs one list table's SELECT, handing each element to add in the order it was written.
func scanList[V any](q Querier, table, query string, after int64, add func(id int64, v V)) error {
	var id int64
	var v V
	return scanRows(q, table, query, after, []any{&id, &v}, func() { add(id, v) })
}

// protoEnum is a generated enum type.
type protoEnum interface {
	~int32
	Descriptor() protoreflect.EnumDescriptor
}

// wordOf resolves a stored enum word through BySpelling, so a word the schema stopped declaring
// FAILS rather than becoming the zero — which would read as a seat that never answered. where is the
// column, named in the refusal.
func wordOf[E protoEnum](where, word string) (E, error) {
	var zero E
	ed := zero.Descriptor()
	vd, ok := recordpb.BySpelling(ed, word)
	if !ok {
		return zero, fmt.Errorf("%s: %q is not a value of %s", where, word, ed.FullName())
	}
	return E(vd.Number()), nil
}

// refuse records the first refusal of an event's row. It is returned when that event is attached,
// so a refused value in a row no read asks for refuses nothing.
func refuse(bad map[int64]error, id int64, err error) map[int64]error {
	if bad == nil {
		bad = map[int64]error{}
	}
	if _, ok := bad[id]; !ok {
		bad[id] = err
	}
	return bad
}

func noBody(id int64, table string) error {
	return fmt.Errorf("event %d has no %s body — an envelope with no body replays as a seat that did nothing", id, table)
}

// caseNamed reports whether some row's oneof discriminator names arm.
func caseNamed(cases map[int64]string, arm string) bool {
	for _, c := range cases {
		if c == arm {
			return true
		}
	}
	return false
}

// ref is a field with presence, set: NULL left it unset, and absent is not zero.
func ref[T any](v T) *T { return &v }

// queryRower is what both a *sql.DB and a *sql.Tx offer for a one-row question.
type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}

// Querier is what both a *sql.DB and a *sql.Tx offer a read: the record's handle, or one read
// transaction a projection asks every question of, so that all its answers come off one snapshot.
type Querier interface {
	queryRower
	rowsQuerier
}

// olderRunAdvice is the cause and the way out for a record behind this binary, given at the open
// (requireThisEpoch) — the one place an older binary is the cause.
// newerRunAdvice is olderRunAdvice's MIRROR, and it names a different remedy because it is a
// different mistake. The record is ahead of the binary, so there is nothing to migrate: the binary is
// the stale party and must be replaced by the one the run was set up with.
//
// MEASURED, and it cost a whole sitting. On universe-m11 a seat typed a BARE `feov-record` and
// reached the HOST's plugin cache — version 1.72.0 but event-schema epoch 8 — instead of the binary
// the run was set up with, at epoch 14. The write died on a raw `CHECK constraint failed: type (275)`,
// every later call was refused with "register is your first act and it has not happened", and that
// sitting made 100 tool calls and recorded NOTHING. A plugin VERSION is not an epoch: the two
// binaries carried the same version string.
//
// The epoch is compared once, at `setup`, which cannot see a seat invoking a different binary
// mid-run. This is the check that can.
const newerRunAdvice = "this record was written by a NEWER binary than this one, so there is nothing " +
	"to migrate — THIS BINARY is the stale party. A plugin version is not an event-schema epoch, so a " +
	"matching version proves nothing. Use the binary the run was set up with: the `feov-record` at the path " +
	"every seat's prompt names, the binDir setup printed, rather than whichever `feov-record` a bare command finds on PATH"

const olderRunAdvice = "it was created by an older binary than this one, " +
	"and a run's schema is fixed when its database is created (it is never altered in place). Migrate it: " +
	"`feov-migrate --from <runDir> --to <freshDir>` — the plugin's operator binary, installed in its bin/ " +
	"beside feov-record and writing this binary's epoch — replays every event through the current " +
	"write path into a fresh run, leaves the source untouched, and writes what it translated to " +
	"<freshDir>/inputs/migration.json; continue in <freshDir>"

// rowsQuerier is what both a *sql.DB and a *sql.Tx offer for a many-row question.
type rowsQuerier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}
