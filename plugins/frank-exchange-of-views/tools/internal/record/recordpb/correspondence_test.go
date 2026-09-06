package recordpb_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	pb "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EventType AND THE `body` ONEOF ARE TWO CARRIERS OF ONE FACT, and this is the gate that makes
// that legal.
//
// The schema keeps `type` alongside the oneof deliberately: it is in every golden and every
// `switch e.Type`, and dropping it would rewrite ~29 switches into type assertions for no gain.
// facts-are-fields permits a second carrier only where something between them can say no —
// otherwise it is exactly the redundancy that drifts silently. Append checks the pair per event;
// this checks the SETS, at build time, so a body added without its enum value (or the reverse)
// fails here rather than at the first run that happens to write one.
//
// Derived from the descriptor, not from a hand-kept list: a list would be a third carrier.
func TestEveryEventTypeHasABodyAndViceVersa(t *testing.T) {
	md := (&pb.Event{}).ProtoReflect().Descriptor()

	od := md.Oneofs().ByName("body")
	if od == nil {
		t.Fatal("Event has no `body` oneof — the schema's central structure is missing")
	}

	// Oneof field names, normalized to the enum's spelling: `blue_edit` -> BLUE_EDIT.
	bodies := map[string]bool{}
	for i := 0; i < od.Fields().Len(); i++ {
		bodies[strings.ToUpper(string(od.Fields().Get(i).Name()))] = true
	}

	// Enum values, minus the UNSPECIFIED zero, which names no body by design.
	types := map[string]bool{}
	ed := pb.EventType(0).Descriptor().Values()
	for i := 0; i < ed.Len(); i++ {
		name := strings.TrimPrefix(string(ed.Get(i).Name()), "EVENT_TYPE_")
		if name == "UNSPECIFIED" {
			continue
		}
		types[name] = true
	}

	for name := range types {
		if !bodies[name] {
			t.Errorf("EventType has EVENT_TYPE_%s but the `body` oneof has no matching field — "+
				"an event of that type could be written with no body at all, which reads as a "+
				"complete event carrying nothing", name)
		}
	}
	for name := range bodies {
		if !types[name] {
			t.Errorf("the `body` oneof has %s but EventType has no EVENT_TYPE_%s — nothing could "+
				"stamp that type, so every reader switching on `type` would skip the body silently",
				strings.ToLower(name), name)
		}
	}

	// The census in plans/record-protobuf.md §II.1 said 32. If that number moves, the plan moves
	// with it — this is the assertion that makes the plan's census checkable rather than asserted.
	//
	// It is 31 since #681 Scope 2: `Opinion` was a body of its own, and the bench's disposition is
	// a docket MOTION's ruling now (`MotionRule.ruling.docket`), which is an arm of a body that
	// already existed. One fewer event type, one fewer body, and the pair still corresponds.
	const wantBodies = 31
	if len(bodies) != wantBodies {
		t.Errorf("the `body` oneof has %d fields, want %d — the event-type census in "+
			"plans/record-protobuf.md §II.1 and this schema must agree", len(bodies), wantBodies)
	}
}

// EVERY FIELD IS `optional`, WITHOUT EXCEPTION — the rule §II.1 states, enforced rather than
// trusted.
//
// proto3 without `optional` has implicit presence: an unset field and a zero-valued one marshal
// to the same bytes under EmitUnpopulated=false. That collapse is fatal in three measured places
// — `seq = 0` is the register event's real sequence number and difftest's rank key reads it;
// `schema_version` absent must differ from 0, which is what the read discriminates on; and
// `applied_verbatim`/`independent`/`none` carry meaning by PRESENCE, never being written false.
//
// A single forgotten `optional` reintroduces the plausible zero this schema exists to delete,
// and it does so invisibly: the code compiles and the field simply stops appearing.
func TestEveryScalarFieldHasExplicitPresence(t *testing.T) {
	var walk func(protoreflect.MessageDescriptor, map[string]bool)
	walk = func(md protoreflect.MessageDescriptor, seen map[string]bool) {
		if seen[string(md.FullName())] {
			return
		}
		seen[string(md.FullName())] = true

		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			// Repeated fields and map entries cannot carry presence and do not need it: an empty
			// list is written as an empty list, which is the distinction that matters.
			if fd.IsList() || fd.IsMap() {
				continue
			}
			// A oneof member has presence by construction — the case tells you.
			if fd.ContainingOneof() != nil {
				continue
			}
			if !fd.HasPresence() {
				t.Errorf("%s.%s has IMPLICIT presence — mark it `optional`. Without it, a zero "+
					"value and an unset field marshal identically under EmitUnpopulated=false, "+
					"so the field vanishes from the line and its absence becomes unreadable",
					md.FullName(), fd.Name())
			}
			if fd.Kind() == protoreflect.MessageKind {
				walk(fd.Message(), seen)
			}
		}
		for i := 0; i < md.Oneofs().Len(); i++ {
			od := md.Oneofs().Get(i)
			for j := 0; j < od.Fields().Len(); j++ {
				if fd := od.Fields().Get(j); fd.Kind() == protoreflect.MessageKind {
					walk(fd.Message(), seen)
				}
			}
		}
	}
	walk((&pb.Event{}).ProtoReflect().Descriptor(), map[string]bool{})
}
