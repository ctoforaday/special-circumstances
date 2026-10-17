package recordpb

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// CheckDeclared refuses a message that holds an enum number its schema does not declare, anywhere
// in it: a field, a list element, a map value, a nested message, a oneof arm.
//
// A GO ENUM IS AN INT32, so `LogType(999)` compiles and a body can carry it. Nothing a seat types
// produces one — the CLI resolves a word to a declared value or refuses it — but every reader past
// the write assumes a declared value: a facet or ruler lookup on the value's descriptor dereferences
// the nil `ByNumber` returns (#1393), and the store writes the value's WORD, which for an undeclared
// number is "". Refused here, once, by a walk over the schema rather than per field, so an enum
// field added to any body is covered without an edit.
//
// SetBody calls it, which puts it ahead of every write's guards and validation: no check
// downstream ever holds an undeclared value.
func CheckDeclared(m proto.Message) error {
	pm := m.ProtoReflect()
	return checkDeclaredIn(pm, string(pm.Descriptor().Name()))
}

func checkDeclaredIn(m protoreflect.Message, path string) error {
	var err error
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		at := path + "." + string(fd.Name())
		switch {
		case fd.IsList():
			l := v.List()
			for i := 0; i < l.Len() && err == nil; i++ {
				err = checkDeclaredValue(fd, l.Get(i), fmt.Sprintf("%s[%d]", at, i))
			}
		case fd.IsMap():
			vd := fd.MapValue()
			v.Map().Range(func(k protoreflect.MapKey, mv protoreflect.Value) bool {
				err = checkDeclaredValue(vd, mv, fmt.Sprintf("%s[%v]", at, k.Interface()))
				return err == nil
			})
		default:
			err = checkDeclaredValue(fd, v, at)
		}
		return err == nil
	})
	return err
}

// checkDeclaredValue checks one value of fd's kind: an enum against its vocabulary, a message by
// walking it. Every other kind holds nothing a schema can fail to declare.
func checkDeclaredValue(fd protoreflect.FieldDescriptor, v protoreflect.Value, at string) error {
	switch fd.Kind() {
	case protoreflect.EnumKind:
		if fd.Enum().Values().ByNumber(v.Enum()) == nil {
			return undeclared(at, fd.Enum(), v.Enum())
		}
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return checkDeclaredIn(v.Message(), at)
	}
	return nil
}

// declaredValue is e's value descriptor, or CheckDeclared's refusal when the schema declares no
// value with e's number. It is the lookup for a reader handed a typed enum value: `ByNumber` alone
// answers nil for the undeclared case, and every descriptor read after it dereferences that nil.
func declaredValue(e protoreflect.Enum) (protoreflect.EnumValueDescriptor, error) {
	vd := e.Descriptor().Values().ByNumber(e.Number())
	if vd == nil {
		return nil, undeclared(string(e.Descriptor().Name()), e.Descriptor(), e.Number())
	}
	return vd, nil
}

// undeclared is the one refusal for an enum number the schema does not declare, naming where the
// value sits, the number, and the words the vocabulary does declare.
func undeclared(at string, ed protoreflect.EnumDescriptor, n protoreflect.EnumNumber) error {
	return feov.Errorf(feov.Validation,
		"%s holds %d, a number %s does not declare — it takes one of %s. A value no word spells "+
			"cannot be stored or read back, so it is refused rather than written",
		at, n, ed.Name(), strings.Join(Words(ed), "|"))
}

// noValue is the refusal a descriptor lookup gives when handed no value descriptor at all — what
// `ByNumber` answers for an undeclared number — instead of dereferencing it.
func noValue(lookup string) error {
	return feov.Errorf(feov.Validation,
		"%s was asked of no enum value — the number has no declaration in this schema", lookup)
}
