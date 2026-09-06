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
	if err := checkRequiredIn(verb, body.ProtoReflect()); err != nil {
		return err
	}
	// AND THE SUBJECT'S OWN ARM, WHOSE ANNOTATIONS ARE AS BINDING AS THE BODY'S.
	//
	// The bench's five reasoning fields used to sit on a top-level body (`Opinion`) and are now a
	// oneof ARM — `MotionRule.ruling.docket`. A walk over the body's own fields sees `docket` as
	// one message field carrying no `required` marking of its own, so every annotation inside it
	// became unreachable the moment it moved: `--principle ""` recorded a ruling with no stated
	// rule, which is the exact decoration this verb exists to refuse.
	//
	// It did not fail loudly, because the DDL derives NOT NULL from the SAME annotations and the
	// database still refused the empty-STRING case — as raw driver text naming a column, not a
	// flag. Half-mediated: the record was safe and the seat was untaught.
	//
	// SYNTHETIC ONEOFS ARE SKIPPED. Every proto3 `optional` field is one, so a walk that does not
	// filter them would recurse into every scalar on every body.
	m := body.ProtoReflect()
	md := m.Descriptor()
	for i := 0; i < md.Oneofs().Len(); i++ {
		od := md.Oneofs().Get(i)
		if od.IsSynthetic() {
			continue
		}
		fd := m.WhichOneof(od)
		if fd == nil || fd.Kind() != protoreflect.MessageKind {
			continue
		}
		if err := checkRequiredIn(verb, m.Get(fd).Message()); err != nil {
			return err
		}
	}
	return nil
}

// checkRequiredIn is the walk over ONE message's own fields. Split out so the body and its filing
// or ruling arm are checked by the same code rather than by two copies that can disagree.
func checkRequiredIn(verb string, m protoreflect.Message) error {
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
