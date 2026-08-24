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
		{&pb.MotionAppeal{}, pb.EventType_EVENT_TYPE_MOTION_APPEAL},
		{&pb.Register{}, pb.EventType_EVENT_TYPE_REGISTER},
		{&pb.Mint{}, pb.EventType_EVENT_TYPE_MINT},
		{&pb.ClassNew{}, pb.EventType_EVENT_TYPE_CLASS_NEW},
		{&pb.RoundVerdict{}, pb.EventType_EVENT_TYPE_VERDICT},
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

func dynamicOf(t *testing.T, full string) proto.Message {
	t.Helper()
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(full))
	if err != nil {
		t.Fatalf("cannot resolve %s: %v", full, err)
	}
	return mt.New().Interface()
}
