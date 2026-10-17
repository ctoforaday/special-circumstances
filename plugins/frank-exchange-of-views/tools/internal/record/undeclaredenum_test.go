package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// undeclaredNumber is an enum number no vocabulary in the schema declares.
const undeclaredNumber = 999

// enumSite is one place a body can hold an enum: the body it was built on, the path a refusal
// names, and the body with that place set to undeclaredNumber.
type enumSite struct {
	path string
	body proto.Message
}

// enumSites builds, for every arm of Event's `body` oneof, one body per enum field reachable from
// it — directly, in a list, as a map value, through a nested message or a oneof arm — with that one
// field holding undeclaredNumber. Read off the descriptors, so an enum field added to any body is a
// new row here without an edit.
func enumSites(t *testing.T) []enumSite {
	t.Helper()
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	var out []enumSite
	for i := 0; i < od.Fields().Len(); i++ {
		arm := od.Fields().Get(i)
		var walk func(md protoreflect.MessageDescriptor, steps []protoreflect.FieldDescriptor, seen map[protoreflect.FullName]bool)
		walk = func(md protoreflect.MessageDescriptor, steps []protoreflect.FieldDescriptor, seen map[protoreflect.FullName]bool) {
			if seen[md.FullName()] {
				return
			}
			seen[md.FullName()] = true
			defer delete(seen, md.FullName())
			for j := 0; j < md.Fields().Len(); j++ {
				fd := md.Fields().Get(j)
				path := append(append([]protoreflect.FieldDescriptor{}, steps...), fd)
				leaf := fd
				if fd.IsMap() {
					leaf = fd.MapValue()
				}
				switch leaf.Kind() {
				case protoreflect.EnumKind:
					out = append(out, buildSite(arm, path))
				case protoreflect.MessageKind:
					if !fd.IsMap() {
						walk(leaf.Message(), path, seen)
					}
				}
			}
		}
		walk(arm.Message(), nil, map[protoreflect.FullName]bool{})
	}
	return out
}

// buildSite makes a fresh body for arm and sets the enum at the end of path to undeclaredNumber,
// creating each message on the way — one element of a list of messages.
func buildSite(arm protoreflect.FieldDescriptor, path []protoreflect.FieldDescriptor) enumSite {
	root := (&recordpb.Event{}).ProtoReflect().NewField(arm).Message()
	m := root
	name := string(arm.Message().Name())
	for _, fd := range path[:len(path)-1] {
		name += "." + string(fd.Name())
		if fd.IsList() {
			l := m.Mutable(fd).List()
			el := l.NewElement()
			l.Append(el)
			m = el.Message()
			name += "[0]"
			continue
		}
		m = m.Mutable(fd).Message()
	}
	fd := path[len(path)-1]
	name += "." + string(fd.Name())
	bad := protoreflect.ValueOfEnum(undeclaredNumber)
	switch {
	case fd.IsList():
		m.Mutable(fd).List().Append(bad)
		name += "[0]"
	case fd.IsMap():
		k := fd.MapKey().Default().MapKey()
		m.Mutable(fd).Map().Set(k, bad)
		name += fmt.Sprintf("[%v]", k.Interface())
	default:
		m.Set(fd, bad)
	}
	return enumSite{path: name, body: root.Interface()}
}

// AN UNDECLARED ENUM NUMBER IS REFUSED BY THE WRITE, FOR EVERY ENUM FIELD OF EVERY BODY (#1393).
//
// `LogType(999)` used to reach `recordpb.SeatMayFile` through validation and panic on the nil
// descriptor `ByNumber` answers; other fields reached the store, which wrote the number's word —
// "" for a number nothing spells. Every row drives the real Append and expects the one refusal:
// code validation, naming where the value sits and the number.
func TestAnUndeclaredEnumNumberIsRefusedByTheWrite(t *testing.T) {
	sites := enumSites(t)
	// The reported shape, complete in every other field, so nothing but the enum can refuse it.
	sites = append(sites, enumSite{path: "Log.type", body: &recordpb.Log{
		Text:   proto.String("the reported body"),
		Type:   recordpb.LogType(undeclaredNumber).Enum(),
		Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum(),
	}})
	// Anti-vacuity: the reported field and a nested one are among the rows.
	var paths []string
	for _, s := range sites {
		paths = append(paths, s.path)
	}
	joined := strings.Join(paths, " ")
	for _, want := range []string{"Log.type", "MotionRule.binds"} {
		if !strings.Contains(joined, want+" ") && !strings.HasSuffix(joined, want) {
			t.Fatalf("no row for %s among %d sites — the walk lost a field it must reach", want, len(sites))
		}
	}
	dir := newRun(t)
	for _, s := range sites {
		t.Run(s.path, func(t *testing.T) {
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Append panicked on %s = %d: %v", s.path, undeclaredNumber, r)
					}
				}()
				_, err = Append(Identity{Run: mustRun(t, dir), SeatID: HarnessSeat}, s.body)
			}()
			if err == nil {
				t.Fatalf("Append wrote %s = %d, a number its vocabulary does not declare", s.path, undeclaredNumber)
			}
			if got := feov.CodeOf(err); got != string(feov.Validation) {
				t.Errorf("code %q, want %q: %v", got, feov.Validation, err)
			}
			if want := fmt.Sprintf("%s holds %d", s.path, undeclaredNumber); !strings.Contains(err.Error(), want) {
				t.Errorf("refusal does not say %q: %v", want, err)
			}
		})
	}
}

// THE DESCRIPTOR LOOKUPS HANDED A TYPED VALUE REFUSE AN UNDECLARED ONE, rather than dereference
// the nil `ByNumber` answers for it. These are the reads #1393's panic went through; a caller that
// is not behind a write still gets an error.
func TestTheValueLookupsRefuseAnUndeclaredNumber(t *testing.T) {
	if _, _, err := recordpb.SeatMayFile(recordpb.LogType(undeclaredNumber)); err == nil {
		t.Error("SeatMayFile answered for an undeclared log type")
	}
	if _, err := recordpb.SubjectRuler(recordpb.MotionSubject(undeclaredNumber)); err == nil {
		t.Error("SubjectRuler answered for an undeclared motion subject")
	}
	if _, _, err := recordpb.Facet(nil, "seat_may_file"); err == nil {
		t.Error("Facet answered for no value")
	}
	if _, _, err := recordpb.Number(nil, "mass"); err == nil {
		t.Error("Number answered for no value")
	}
	if _, err := recordpb.RulerOf(nil); err == nil {
		t.Error("RulerOf answered for no value")
	}
	if _, err := recordpb.EnumValueDoc(nil); err == nil {
		t.Error("EnumValueDoc answered for no value")
	}
}
