package record

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY ENUM A BODY PROJECTS IS SPELLED AS THE RECORD SPELLS IT. The board embeds closure and
// regrade bodies through bodyMap; protojson alone writes `DISPOSITION_REPAIRED` where the record,
// the CLI and every other view say `repaired`. Each event body is filled with every enum it can
// carry — nested and repeated ones included — and the projection must hold the word at each.
func TestABodysProjectionSpellsEveryEnumAsItsWord(t *testing.T) {
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	enums := 0
	for i := 0; i < od.Fields().Len(); i++ {
		md := od.Fields().Get(i).Message()
		m := dynamicpb.NewMessage(md)
		fillEnums(m, 0)
		out, err := bodyMap(m)
		if err != nil {
			t.Fatalf("%s: %v", md.Name(), err)
		}
		enums += checkEnumWords(t, string(md.Name()), m, out)
	}
	if enums == 0 {
		t.Fatal("no body carried an enum — the walk reached nothing, so this test measures nothing")
	}
}

// fillEnums sets every enum field to its first non-zero value, and recurses into message fields.
func fillEnums(m protoreflect.Message, depth int) {
	if depth > 4 {
		return
	}
	fs := m.Descriptor().Fields()
	for i := 0; i < fs.Len(); i++ {
		fd := fs.Get(i)
		switch {
		case fd.IsMap():
		case fd.IsList() && fd.Kind() == protoreflect.EnumKind:
			m.Mutable(fd).List().Append(protoreflect.ValueOfEnum(fd.Enum().Values().Get(1).Number()))
		case fd.IsList() && fd.Message() != nil:
			l := m.Mutable(fd).List()
			el := l.NewElement()
			fillEnums(el.Message(), depth+1)
			l.Append(el)
		case fd.Kind() == protoreflect.EnumKind:
			m.Set(fd, protoreflect.ValueOfEnum(fd.Enum().Values().Get(1).Number()))
		case fd.Message() != nil:
			fillEnums(m.Mutable(fd).Message(), depth+1)
		}
	}
}

func checkEnumWords(t *testing.T, path string, m protoreflect.Message, out map[string]any) int {
	t.Helper()
	n := 0
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		key, p := string(fd.Name()), path+"."+string(fd.Name())
		switch {
		case fd.IsMap():
		case fd.IsList():
			items, _ := out[key].([]any)
			for j := 0; j < v.List().Len(); j++ {
				if j >= len(items) {
					t.Errorf("%s: %d item(s) projected, the body holds %d", p, len(items), v.List().Len())
					break
				}
				if fd.Kind() == protoreflect.EnumKind {
					n++
					if want := recordpb.Spelling(fd.Enum().Values().ByNumber(v.List().Get(j).Enum())); items[j] != want {
						t.Errorf("%s[%d] = %v, want the word %q", p, j, items[j], want)
					}
				} else if fd.Message() != nil {
					sub, _ := items[j].(map[string]any)
					n += checkEnumWords(t, p, v.List().Get(j).Message(), sub)
				}
			}
		case fd.Kind() == protoreflect.EnumKind:
			n++
			if want := recordpb.Spelling(fd.Enum().Values().ByNumber(v.Enum())); out[key] != want {
				t.Errorf("%s = %v, want the word %q", p, out[key], want)
			}
		case fd.Message() != nil:
			sub, _ := out[key].(map[string]any)
			n += checkEnumWords(t, p, v.Message(), sub)
		}
		return true
	})
	return n
}
