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
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
