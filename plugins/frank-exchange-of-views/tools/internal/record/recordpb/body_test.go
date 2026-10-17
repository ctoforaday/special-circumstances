package recordpb_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	pb "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE TYPE IS STAMPED FROM THE BODY, never taken from the caller. This is the property that
// retires `Append(id, "motion-rule", payload)`: with one argument there is no pair to disagree.
func TestSetBodyStampsTheTypeItsBodyNames(t *testing.T) {
	for _, tc := range []struct {
		body proto.Message
		want pb.EventType
	}{
		{&pb.Motion{}, pb.EventType_EVENT_TYPE_MOTION},
		{&pb.MotionRule{}, pb.EventType_EVENT_TYPE_MOTION_RULE},
		{&pb.Register{}, pb.EventType_EVENT_TYPE_REGISTER},
		{&pb.Mint{}, pb.EventType_EVENT_TYPE_MINT},
		{&pb.ClassNew{}, pb.EventType_EVENT_TYPE_CLASS_NEW},
		{&pb.Gate{}, pb.EventType_EVENT_TYPE_VERDICT},
	} {
		ev := &pb.Event{}
		got, err := pb.SetBody(ev, tc.body)
		if err != nil {
			t.Fatalf("%T: %v", tc.body, err)
		}
		if got != tc.want {
			t.Errorf("%T stamped %v, want %v", tc.body, got, tc.want)
		}
		if ev.GetType() != tc.want {
			t.Errorf("%T set ev.Type=%v, want %v — the returned type and the stamped one must be "+
				"the same fact", tc.body, ev.GetType(), tc.want)
		}
		body, ok := pb.Body(ev)
		if !ok {
			t.Fatalf("%T: Body() reports no body on an event that was just given one", tc.body)
		}
		if body.ProtoReflect().Descriptor().FullName() != tc.body.ProtoReflect().Descriptor().FullName() {
			t.Errorf("%T: Body() returned %s", tc.body, body.ProtoReflect().Descriptor().FullName())
		}
	}
}

// EVERY body message is reachable — not just the seven spot-checked above. A body added to the
// schema without its enum value must fail HERE, at build time, rather than at the first run that
// happens to write one.
func TestEveryBodyInTheOneofCanBeSet(t *testing.T) {
	od := (&pb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		t.Fatal("Event has no `body` oneof")
	}
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		msg := dynamicOf(t, string(fd.Message().FullName()))
		ev := &pb.Event{}
		if _, err := pb.SetBody(ev, msg); err != nil {
			t.Errorf("%s is a body field but SetBody refuses it: %v", fd.Name(), err)
		}
	}
}

// A NO-MATCH IS LOUD. A message that is not a body must be refused, not silently written as an
// event of no type — the plausible-zero failure this schema exists to remove.
func TestSetBodyRefusesAMessageThatIsNotABody(t *testing.T) {
	ev := &pb.Event{}
	_, err := pb.SetBody(ev, &pb.Event{}) // Event is not one of its own bodies
	if err == nil {
		t.Fatal("SetBody accepted a message that is not a body — an event of no type would be written")
	}
	if !strings.Contains(err.Error(), "is not an event body") {
		t.Errorf("the refusal must say what was wrong; got %q", err)
	}
	if ev.GetType() != pb.EventType_EVENT_TYPE_UNSPECIFIED {
		t.Error("a refused SetBody left a type stamped on the event")
	}
	if _, ok := pb.Body(ev); ok {
		t.Error("a refused SetBody left a body on the event")
	}
}

func TestSetBodyRefusesNoBodyAtAll(t *testing.T) {
	if _, err := pb.SetBody(&pb.Event{}, nil); err == nil {
		t.Fatal("SetBody accepted a nil body")
	}
	if _, err := pb.SetBody(nil, &pb.Motion{}); err == nil {
		t.Fatal("SetBody accepted a nil event")
	}
}

// Body() distinguishes ABSENT from EMPTY, which the map-shaped payload could not.
func TestBodyReportsAbsenceRatherThanAnEmptyMessage(t *testing.T) {
	if _, ok := pb.Body(&pb.Event{}); ok {
		t.Error("an event with no body reports one")
	}
	ev := &pb.Event{}
	if _, err := pb.SetBody(ev, &pb.Halt{}); err != nil {
		t.Fatal(err)
	}
	if _, ok := pb.Body(ev); !ok {
		t.Error("an event whose body is an EMPTY message reports no body — absent and empty are " +
			"different facts and the reader must be able to tell them apart")
	}
}

// ONE LOOKUP EACH WAY, AND EVERY WORD OF THE VOCABULARY GOES THROUGH IT. An event type's word
// names its EventType and its arm of the `body` oneof; every reader that starts from the word
// calls these two, so a word one of them cannot resolve is a word no reader can.
func TestEveryEventWordResolvesToItsTypeAndItsBodyField(t *testing.T) {
	vals := pb.EventType(0).Descriptor().Values()
	var words int
	for i := 0; i < vals.Len(); i++ {
		typ := pb.EventType(vals.Get(i).Number())
		if typ == pb.EventType_EVENT_TYPE_UNSPECIFIED {
			continue
		}
		words++
		word := pb.Word(typ)
		if got, ok := pb.EventTypeOf(word); !ok || got != typ {
			t.Errorf("EventTypeOf(%q) = %v, %v; want %v", word, got, ok, typ)
		}
		fd, ok := pb.BodyField(word)
		if !ok {
			t.Errorf("BodyField(%q) resolves nothing — an event of that type has no body to read", word)
			continue
		}
		if string(fd.Name()) != word || fd.Message() == nil || fd.ContainingOneof() == nil || fd.ContainingOneof().Name() != "body" {
			t.Errorf("BodyField(%q) = %s, which is not that word's message arm of the `body` oneof", word, fd.FullName())
		}
	}

	arms := (&pb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body").Fields()
	if arms.Len() != words {
		t.Errorf("the `body` oneof has %d arms and EventType %d words", arms.Len(), words)
	}
	for i := 0; i < arms.Len(); i++ {
		name := string(arms.Get(i).Name())
		if fd, ok := pb.BodyField(name); !ok || fd != arms.Get(i) {
			t.Errorf("BodyField(%q) does not return the arm of that name", name)
		}
		if _, ok := pb.EventTypeOf(name); !ok {
			t.Errorf("EventTypeOf(%q) resolves nothing for a declared body arm", name)
		}
	}
}

// A NO-MATCH IS A REFUSAL, never the zero type made to look like an answer or a field of the
// envelope. `seat_id` is a field of Event and no body; `unspecified` spells the zero, which names
// no event; the lookup is exact, so a louder or hyphenated spelling is another word.
func TestAWordThatNamesNoEventResolvesNothing(t *testing.T) {
	for _, word := range []string{"", "unspecified", "seat_id", "no_such_event", "MINT", "motion-rule"} {
		if got, ok := pb.EventTypeOf(word); ok || got != pb.EventType_EVENT_TYPE_UNSPECIFIED {
			t.Errorf("EventTypeOf(%q) = %v, %v; want the zero and false", word, got, ok)
		}
		if fd, ok := pb.BodyField(word); ok || fd != nil {
			t.Errorf("BodyField(%q) = %v, %v; want nil and false", word, fd, ok)
		}
	}
}

func dynamicOf(t *testing.T, full string) proto.Message {
	t.Helper()
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(full))
	if err != nil {
		t.Fatalf("cannot resolve %s: %v", full, err)
	}
	return mt.New().Interface()
}
