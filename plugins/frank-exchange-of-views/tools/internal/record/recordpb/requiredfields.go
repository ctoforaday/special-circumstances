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
		if !o.GetRequired() {
			continue
		}
		if !m.Has(fd) {
			return fmt.Errorf("record: %s requires --%s%s", verb, flagFor(fd, o), because(o))
		}
		// PRESENT IS NOT ENOUGH FOR PROSE. A required string that is EMPTY is a duty discharged by
		// silence — an acceptance check demanding nothing, a friction entry saying nothing — and
		// the check was presence-only for a while after the Go table's two flavours collapsed into
		// one annotation. `allow_empty` is the narrow exception, declared at the field: a
		// `--review-flag false` is a real ruling, so an empty answer there is an answer.
		if fd.Kind() == protoreflect.StringKind && !o.GetAllowEmpty() && m.Get(fd).String() == "" {
			return fmt.Errorf("record: %s requires --%s to say something%s", verb, flagFor(fd, o), because(o))
		}
	}
	return nil
}

// FlagFor is the exported form: the word a seat types for a field, off the field's own
// annotation. The help and the contract gate both need it, and deriving it by rule instead is
// what put `--motion-id` in a refusal for a flag spelled `--id`.
func FlagFor(fd protoreflect.FieldDescriptor) string { return flagFor(fd, nil) }

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
