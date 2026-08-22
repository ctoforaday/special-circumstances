// Package recordtest builds record events for fixtures.
//
// It exists because the migration turned every test's event literal into a shape no one wants to
// type: an identity, a typed body message, and a `body` oneof whose arm must agree with the
// EventType beside it. Written out at each of the ~45 fixture sites, that agreement is a fact
// stated twice, and the second copy is the one a hurried edit gets wrong — a fixture whose Type
// says `finding` while its body is a Mint tests a state the record cannot hold.
//
// So the type is DERIVED from the body here, by the same recordpb.SetBody the production write
// path uses. A fixture cannot express the disagreement, which is the point.
//
// It imports recordpb rather than record, so `internal/record`'s own tests can use it without an
// import cycle: record.Event is an alias of recordpb.Event, and both sides see the same type.
package recordtest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Event is the common fixture: a seat, a round, and what it recorded.
//
// It takes *testing.T and fails rather than returning an error. A fixture that could not be built
// is not a test condition — it is a broken test, and returning an error here would let a caller
// ignore it and assert against an empty event, which passes for the wrong reason.
func Event(t *testing.T, seatID string, round int, body proto.Message) *recordpb.Event {
	t.Helper()
	ev := &recordpb.Event{
		SeatId: proto.String(seatID),
		Round:  proto.Int32(int32(round)),
	}
	typ, err := recordpb.SetBody(ev, body)
	if err != nil {
		t.Fatalf("recordtest: %v", err)
	}
	ev.Type = &typ
	return ev
}

// At is Event with the shard coordinates a replay-ordering test needs. Most fixtures do not care
// about seq, nonce or key; the ones that are ABOUT ordering care about nothing else.
func At(t *testing.T, seatID string, round int, key string, body proto.Message) *recordpb.Event {
	t.Helper()
	ev := Event(t, seatID, round, body)
	if key != "" {
		ev.Key = proto.String(key)
	}
	return ev
}

// P is a pointer to a value, for the optional scalar and enum fields a fixture sets inline.
//
// Every field on these messages is optional, so every one is a pointer — and ABSENT IS NOT ZERO
// here: the zero of an enum is UNSPECIFIED, which the record reserves for "the seat never said".
// A fixture that wants to say "this gap is graded high" must therefore pass a pointer to
// Grade_GRADE_HIGH, and a fixture that wants to say "ungraded" passes nil. Writing that as a
// three-line closure at each site is how the distinction gets quietly collapsed.
func P[T any](v T) *T { return &v }

// Stamped sets an event's timestamp, for the fixtures that are ABOUT time — seat spans, ordering
// across shards, the parse-failure count. Most fixtures do not set one and must not: an invented
// stamp that happens to sort correctly hides an ordering bug rather than exposing it.
func Stamped(ev *recordpb.Event, ts string) *recordpb.Event {
	ev.Ts = proto.String(ts)
	return ev
}

// Seed writes events straight into a run's record, for fixtures that need a board to exist before
// the thing under test reads it.
//
// # Why fixtures need this at all
//
// They used to write a shard FILE — `events-<seat>-<nonce>.jsonl` — with a local helper copied
// into each test package. That worked because the storage was a file format anyone could produce.
// It is a database now, so a fixture that writes a file produces a run whose record is EMPTY, and
// a test asserting on an empty board passes for entirely the wrong reason. That failure is silent
// and it looks exactly like success, which is why the helper is here rather than re-copied.
//
// # Why not go through record.Append
//
// Append is the production write path: it validates, derives idempotency keys, and requires the
// seat to have registered. A fixture that wants a gap already closed would have to perform the
// whole debate to get one. Seed writes the state directly, which is what a fixture IS — and the
// constraints still apply, so a fixture that seeds a closure for a gap it never minted is refused
// by the foreign key rather than quietly producing a board nobody could have reached.
func Seed(t *testing.T, runDir string, evs ...*recordpb.Event) {
	t.Helper()
	dir := filepath.Join(runDir, "records")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("recordtest: %v", err)
	}
	db, err := recordsql.Open(filepath.Join(dir, "record.db"))
	if err != nil {
		t.Fatalf("recordtest: opening the run's record: %v", err)
	}
	defer db.Close()
	for i, ev := range evs {
		if ev.Ts == nil {
			// A STAMP EVERY EVENT, because `ts` is NOT NULL and several audits read it. The
			// spacing is one second per event so a fixture that does not care about time still
			// produces a plausible span rather than a burst — BackfillAudit reads exactly that
			// shape, and zero-width spans would make every seeded run look like narration.
			ev.Ts = proto.String(fmt.Sprintf("2026-01-01T00:%02d:%02dZ", i/60, i%60))
		}
		if _, err := recordsql.Insert(db, ev); err != nil {
			t.Fatalf("recordtest: seeding event %d (%s by %s): %v",
				i, recordpb.Word(ev.GetType()), ev.GetSeatId(), err)
		}
	}
}
