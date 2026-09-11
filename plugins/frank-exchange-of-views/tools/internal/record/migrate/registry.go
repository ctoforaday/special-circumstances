package migrate

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Entry is one authored translation. Exactly one of Translate / Loss is set.
type Entry struct {
	// Translate builds the new bodies for one old event, in append order — usually one, two
	// for a concept that split (opinion → docket motion + ruling). dst is the DESTINATION
	// run, holding only already-replayed events; it exists for id minting
	// (record.MintMotionID), never for reading the old record.
	Translate func(old OldEvent, dst record.Run) ([]proto.Message, error)
	// Loss marks the word untranslatable, and says why. A loss fails the migration unless
	// the operator accepts it BY WORD (--accept-loss), and an accepted loss is recorded in
	// the manifest: acknowledged loss is visible loss.
	Loss string
}

// Registry maps old words to their translations. A word with no entry translates by
// IDENTITY — same word, fields matched by name — which refuses rather than guesses:
// a word the current vocabulary does not declare, or an old column the current body does
// not carry, is an error naming itself, never a skip. A shape delta inside a kept word is
// the same hazard as a renamed word.
type Registry map[string]Entry

// Translate resolves one old event: the word's entry if one exists, the identity default
// otherwise.
func (r Registry) Translate(old OldEvent, dst record.Run) ([]proto.Message, error) {
	if e, ok := r[old.Word]; ok {
		if e.Loss != "" {
			return nil, LossError{Word: old.Word, Reason: e.Loss}
		}
		return e.Translate(old, dst)
	}
	body, err := identityBody(old)
	if err != nil {
		return nil, err
	}
	return []proto.Message{body}, nil
}

// LossError is a documented untranslatable, distinct from a defect so the replayer can honour
// an operator's --accept-loss for exactly this word and nothing else.
type LossError struct {
	Word, Reason string
}

func (e LossError) Error() string {
	return fmt.Sprintf("migrate: %q is untranslatable: %s", e.Word, e.Reason)
}

// bodyFieldFor resolves an old word against the CURRENT Event body oneof — the oneof field
// name IS the word, which is the correspondence SetBody's own eventTypeOf leans on.
func bodyFieldFor(word string) protoreflect.FieldDescriptor {
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		return nil
	}
	return od.Fields().ByName(protoreflect.Name(word))
}

// newBody constructs a concrete instance of the body message a word names.
func newBody(fd protoreflect.FieldDescriptor) proto.Message {
	return (&recordpb.Event{}).ProtoReflect().NewField(fd).Message().Interface()
}

// identityBody is the default translation: the same word, fields matched by NAME against the
// current body message, converted by the current field's own kind.
func identityBody(old OldEvent) (proto.Message, error) {
	fd := bodyFieldFor(old.Word)
	if fd == nil {
		return nil, fmt.Errorf("migrate: the current vocabulary does not declare %q (%s) and the registry has no entry for it — "+
			"author a translation or rule it a loss; skipping it silently is the one answer this tool refuses to give",
			old.Word, "old event")
	}
	body := newBody(fd)
	m := body.ProtoReflect()
	if err := fillFields(m, old.Fields, old.Word); err != nil {
		return nil, err
	}
	for name, vals := range old.Lists {
		lf := m.Descriptor().Fields().ByName(protoreflect.Name(name))
		if lf == nil || !lf.IsList() {
			return nil, fmt.Errorf("migrate: old table %s_%s holds a list the current %q body does not carry — map it or drop it, with a reason, in a registry entry", old.Word, name, old.Word)
		}
		list := m.Mutable(lf).List()
		for _, v := range vals {
			pv, err := scalarValue(lf, v)
			if err != nil {
				return nil, fmt.Errorf("migrate: %s_%s: %w", old.Word, name, err)
			}
			list.Append(pv)
		}
	}
	for arm, cols := range old.Arms {
		af := m.Descriptor().Fields().ByName(protoreflect.Name(arm))
		if af == nil || af.Message() == nil {
			return nil, fmt.Errorf("migrate: old table %s_%s holds a oneof arm the current %q body does not carry — map it or drop it, with a reason, in a registry entry", old.Word, arm, old.Word)
		}
		sub := m.NewField(af).Message()
		if err := fillFields(sub, cols, old.Word+"_"+arm); err != nil {
			return nil, err
		}
		m.Set(af, protoreflect.ValueOfMessage(sub))
	}
	return body, nil
}

// fillFields writes loose columns onto a message, refusing any column the message has no
// field for. The one structural exception is a `<oneof>_case` discriminator column: the arm
// row itself carries the fact, so the discriminator is derivable and skipped — but ONLY when
// the message actually declares that oneof; otherwise the suffix is a coincidence and the
// column is as unmapped as any other.
func fillFields(m protoreflect.Message, cols map[string]any, table string) error {
	md := m.Descriptor()
	for name, v := range cols {
		fd := md.Fields().ByName(protoreflect.Name(name))
		if fd == nil {
			if base, ok := strings.CutSuffix(name, "_case"); ok && md.Oneofs().ByName(protoreflect.Name(base)) != nil {
				continue
			}
			return fmt.Errorf("migrate: old column %s.%s has no field on the current body — map it or drop it, with a reason, in a registry entry", table, name)
		}
		pv, err := scalarValue(fd, v)
		if err != nil {
			return fmt.Errorf("migrate: %s.%s: %w", table, name, err)
		}
		m.Set(fd, pv)
	}
	return nil
}

// scalarValue converts one SQL value by the CURRENT field's kind. Enum columns hold the
// word, so the reverse of recordpb's spelling is looked up on the enum's own descriptor;
// a word the vocabulary does not hold is an error naming it, exactly like the write path's
// own refusal.
func scalarValue(fd protoreflect.FieldDescriptor, v any) (protoreflect.Value, error) {
	switch fd.Kind() {
	case protoreflect.StringKind:
		s, ok := v.(string)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("field %s wants a string, old record holds %T", fd.Name(), v)
		}
		return protoreflect.ValueOfString(s), nil
	case protoreflect.BoolKind:
		n, ok := v.(int64)
		if !ok || (n != 0 && n != 1) {
			return protoreflect.Value{}, fmt.Errorf("field %s wants a 0/1 bool, old record holds %v (%T)", fd.Name(), v, v)
		}
		return protoreflect.ValueOfBool(n == 1), nil
	case protoreflect.Int32Kind:
		n, ok := v.(int64)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("field %s wants an integer, old record holds %T", fd.Name(), v)
		}
		return protoreflect.ValueOfInt32(int32(n)), nil
	case protoreflect.Int64Kind:
		n, ok := v.(int64)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("field %s wants an integer, old record holds %T", fd.Name(), v)
		}
		return protoreflect.ValueOfInt64(n), nil
	case protoreflect.BytesKind:
		b, ok := v.([]byte)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("field %s wants bytes, old record holds %T", fd.Name(), v)
		}
		return protoreflect.ValueOfBytes(b), nil
	case protoreflect.EnumKind:
		w, ok := v.(string)
		if !ok {
			return protoreflect.Value{}, fmt.Errorf("enum field %s wants a word, old record holds %T", fd.Name(), v)
		}
		num, err := enumValue(fd.Enum(), w)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfEnum(num), nil
	}
	return protoreflect.Value{}, fmt.Errorf("field %s has kind %s, which no old column can carry", fd.Name(), fd.Kind())
}

// enumNumberOf is the reverse of the schema's enum spelling: the value whose word matches —
// exactly first, then by the schema's own same-word class (recordpb.SameWord: case and
// separators only), because the eras changed exactly those. `low-medium` and `FAIL` are the
// same words as `low_medium` and `fail`; `closed` and `repaired` are not, and stay a
// refusal for an AUTHORED map to answer.
func enumNumberOf(ed protoreflect.EnumDescriptor, word string) (protoreflect.EnumNumber, error) {
	for i := 0; i < ed.Values().Len(); i++ {
		vd := ed.Values().Get(i)
		if recordpb.Spelling(vd) == word {
			return vd.Number(), nil
		}
	}
	for i := 0; i < ed.Values().Len(); i++ {
		vd := ed.Values().Get(i)
		if recordpb.SameWord(recordpb.Spelling(vd), word) {
			return vd.Number(), nil
		}
	}
	return 0, fmt.Errorf("the current %s vocabulary does not hold the word %q — author its mapping in a registry entry", ed.Name(), word)
}
