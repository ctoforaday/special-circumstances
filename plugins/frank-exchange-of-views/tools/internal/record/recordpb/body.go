package recordpb

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SetBody puts a typed body on an event and stamps the matching EventType, deriving BOTH from the
// schema rather than from anything the caller says.
//
// THIS IS WHAT REPLACES `Append(id, "motion-rule", payload)`. The old signature carried the event
// type as a STRING beside a map, so the two could disagree and nothing could refuse: a typo in the
// type wrote a well-formed event of a type no reader switches on, and a payload built for one type
// under another's name wrote a complete-looking event whose fields no reader would ever ask for.
// Both failures were silent at the write and invisible at the read. Here the body IS the type —
// there is no second argument to get wrong.
//
// The correspondence it relies on (oneof field `motion_rule` <-> `EVENT_TYPE_MOTION_RULE`) is the
// one TestEveryEventTypeHasABodyAndViceVersa pins as SETS at build time. This function checks the
// pair per event through EventTypeOf, which is the per-write half that test's header names.
func SetBody(ev *Event, body proto.Message) (EventType, error) {
	if ev == nil {
		return EventType_EVENT_TYPE_UNSPECIFIED, fmt.Errorf("recordpb: SetBody on a nil event")
	}
	if body == nil {
		return EventType_EVENT_TYPE_UNSPECIFIED, fmt.Errorf("recordpb: an event with no body is not an event — every EventType names a body message")
	}

	want := body.ProtoReflect().Descriptor()
	md := ev.ProtoReflect().Descriptor()
	od := md.Oneofs().ByName("body")
	if od == nil {
		return EventType_EVENT_TYPE_UNSPECIFIED, fmt.Errorf("recordpb: Event has no `body` oneof")
	}

	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if fd.Message() == nil || fd.Message().FullName() != want.FullName() {
			continue
		}
		// The arm's name is the type's word. An arm with no type must fail loudly here rather
		// than write an UNSPECIFIED type nothing switches on.
		typ, ok := EventTypeOf(string(fd.Name()))
		if !ok {
			return EventType_EVENT_TYPE_UNSPECIFIED, fmt.Errorf(
				"recordpb: the `body` oneof carries %s but EventType has no EVENT_TYPE_%s — nothing could stamp "+
					"that type, so every reader switching on `type` would skip the body silently",
				fd.Name(), strings.ToUpper(string(fd.Name())))
		}
		ev.ProtoReflect().Set(fd, protoreflect.ValueOfMessage(body.ProtoReflect()))
		ev.Type = typ.Enum()
		return typ, nil
	}

	// A message that is not a body at all. Naming what WOULD have worked matters more here than
	// anywhere else in the write path: this is the refusal a seat's author meets when they reach
	// for a message the record does not carry as an event.
	return EventType_EVENT_TYPE_UNSPECIFIED, fmt.Errorf(
		"recordpb: %s is not an event body — the `body` oneof carries %d messages, and an event is exactly one of them",
		want.Name(), od.Fields().Len())
}

// EventTypeOf resolves an event type's WORD — `motion_rule` — to its EventType, exactly: the zero's
// spelling, a louder spelling and a hyphenated one all resolve nothing.
//
// IT IS THE ONE WORD → TYPE LOOKUP. The word is three things at once — the spelling of the enum
// value, the name of its arm of the `body` oneof, and the text `events.type` stores — and a reader
// starting from any of them calls this. The correspondence is the schema's, not a hand-kept map: a
// map would be a further carrier of a fact two already hold, and the one most likely to be
// forgotten when a body is added.
func EventTypeOf(word string) (EventType, bool) {
	vd, ok := BySpelling(EventType(0).Descriptor(), word)
	if !ok {
		return EventType_EVENT_TYPE_UNSPECIFIED, false
	}
	return EventType(vd.Number()), true
}

// BodyField resolves an event type's WORD to the arm of Event's `body` oneof that carries it — the
// one place the word ↔ body pairing is declared, and the one word → body lookup. The arm's
// `Message()` is the body's descriptor; `NewField` on an event mints the concrete body.
//
// A field of the envelope that is no arm (`seat_id`) resolves nothing, as does a word the schema
// does not declare.
func BodyField(word string) (protoreflect.FieldDescriptor, bool) {
	od := (&Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		return nil, false
	}
	fd := od.Fields().ByName(protoreflect.Name(word))
	if fd == nil || fd.Message() == nil {
		return nil, false
	}
	return fd, true
}

// BodyAs reads the body and asserts it is exactly T, keeping NO BODY and WRONG TYPE
// distinguishable from each other by returning false for both while never returning a usable
// zero value that a caller might act on.
//
// Placed centrally because all ten converting agents reach for this same shape — read the body,
// assert the message, proceed. Ten private copies of it would be the duplication this migration
// exists to remove, reproduced one level up.
func BodyAs[T proto.Message](ev *Event) (T, bool) {
	var zero T
	body, ok := Body(ev)
	if !ok {
		return zero, false
	}
	typed, ok := body.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}

// Body returns the event's body message, and whether it had one.
//
// A no-match returns false rather than a zero-valued message: an event with no body and an event
// whose body is empty are different facts, and the reader must be able to tell them apart.
func Body(ev *Event) (proto.Message, bool) {
	if ev == nil {
		return nil, false
	}
	md := ev.ProtoReflect().Descriptor()
	od := md.Oneofs().ByName("body")
	if od == nil {
		return nil, false
	}
	fd := ev.ProtoReflect().WhichOneof(od)
	if fd == nil {
		return nil, false
	}
	return ev.ProtoReflect().Get(fd).Message().Interface(), true
}
