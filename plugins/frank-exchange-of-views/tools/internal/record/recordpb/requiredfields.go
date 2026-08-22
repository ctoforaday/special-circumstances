package recordpb

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// CheckRequired refuses a body missing a field its own schema marks required.
//
// # Why this replaces four things
//
// "What a verb may not omit" was stated in four places: a map keyed by event type feeding the
// help's REQUIRED marks, a map keyed by field name feeding nothing, per-verb `if` blocks in
// validate doing the actual refusing, and — once the SQLite schema arrived — a NOT NULL derived
// from the annotation. Four artifacts, one fact, and they had already diverged: `Outcome.verdict`
// was added to the field-keyed map and refused nothing, because no writer consulted it.
//
// The annotation sits ON the field, so there is one statement and every enforcer reads it.
//
// # Why presence and not emptiness
//
// Every field is optional in the schema, so `Has` answers the question the record actually cares
// about: did the seat SAY this. An empty string a seat passed deliberately — `--review-flag ""` is
// a legitimate ruling — is present, and a field never passed is absent. Testing for "" instead
// would refuse the first and accept nothing extra.
func CheckRequired(verb string, body proto.Message) error {
	m := body.ProtoReflect()
	md := m.Descriptor()
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		o, _ := proto.GetExtension(fd.Options(), E_Sql).(*Sql)
		if !o.GetRequired() || m.Has(fd) {
			continue
		}
		return fmt.Errorf("record: %s requires --%s%s", verb, flagFor(fd, o), because(o))
	}
	return nil
}

// flagFor is the word a seat types. It is the field's own name unless the field says otherwise —
// which only the prose fields do, because a close stores `prose` and an opinion `rationale` while
// the word for both is `--reason`.
func flagFor(fd protoreflect.FieldDescriptor, o *Sql) string {
	if o == nil {
		o, _ = proto.GetExtension(fd.Options(), E_Sql).(*Sql)
	}
	if f := o.GetFlag(); f != "" {
		return f
	}
	return strings.ReplaceAll(string(fd.Name()), "_", "-")
}

func because(o *Sql) string {
	if w := o.GetWhy(); w != "" {
		return " (" + w + ")"
	}
	return ""
}

// RequiredOf lists the fields a message requires, in declaration order, for the help's REQUIRED
// marks. One list, so what the help promises and what the write refuses are the same set.
func RequiredOf(md protoreflect.MessageDescriptor) []protoreflect.FieldDescriptor {
	var out []protoreflect.FieldDescriptor
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if o, _ := proto.GetExtension(fd.Options(), E_Sql).(*Sql); o.GetRequired() {
			out = append(out, fd)
		}
	}
	return out
}
