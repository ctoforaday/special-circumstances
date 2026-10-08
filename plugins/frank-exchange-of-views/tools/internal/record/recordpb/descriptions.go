package recordpb

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// WHAT EACH ENUM VALUE MEANS, DECLARED ONCE — and the SET is generated, only the prose is written.
//
// This replaces record.EnumValue, whose whole point was that a set rendered as six bare words
// leaves a seat guessing which situation warrants which. That prose is not decoration: it is
// generated into the CLI's --help AND into the refusal a seat reads when it gets one wrong, and
// enums.go was explicit that the help must not be allowed to drift from the check because it is
// BUILT from it. Proto enums carry no descriptions, so the prose lives here.
//
// WHY THIS IS NOT THE HAND-KEPT ALLOWLIST facts-are-fields WARNS ABOUT. That rule's objection is
// to a guard whose own list is maintained by hand — it reproduces the defect one level up. Here
// the authoritative set is the generated descriptor; this map supplies only the sentence. An
// exhaustiveness test walks the descriptor and fails when any value lacks prose, and fails again
// when this map names a value that no longer exists. Neither direction can rot silently, which
// is the property a bare list cannot offer.
//
// Rejected: custom proto options. They would put the prose in the .proto, which reads well, at
// the price of descriptor plumbing and a protoc-gen-go extension to read it back — a build
// dependency for a string table.

// THE EXEMPTION IS GONE, AND ITS PREMISE IS WHY.
//
// `undocumentedEnums` held EventType (and, until it was removed, SchemaVersion), on the reason "whose values no seat ever
// types, so no --help renders them". That was true, and it was scoped to the only consumer `means`
// had when it was written. There are three now: --help, the refusal a seat reads, and the record's
// own vocabulary TABLES — and the third has an audience the first two do not, a human reading the
// database directly. `events.type` is the column every join keys on. Leaving it undocumented meant
// the one word that says WHAT AN ACT WAS could not be joined to what it means, in the artifact that
// exists to be read after the run.
//
// The exemption also cost the schema a wall: with no vocabulary table there was nothing for
// `events.type` to reference, so it was bare TEXT — the only such column left once the arms were
// repaired.

// EnumValueDoc returns the prose for one enum value.
//
// The miss is LOUD. A silent "" would render an empty --help line, which reads as a value with
// no meaning rather than a value whose meaning nobody wrote — and that is the failure this whole
// table exists to prevent.
func EnumValueDoc(v protoreflect.EnumValueDescriptor) (string, error) {
	if m, _ := proto.GetExtension(v.Options(), E_Means).(string); m != "" {
		return m, nil
	}
	return "", fmt.Errorf("recordpb: no meaning on enum value %s — put one on the value itself, "+
		"`%s = N [(means) = \"…\"]`; a set rendered as bare words leaves a seat guessing which "+
		"situation warrants which", v.FullName(), v.Name())
}

// Spelling is the word a seat types: the enum value's name with its type prefix removed and
// lowercased, so CLOSURE_CLASS_DEFECT_ACCEPTED reads as `defect_accepted`.
func Spelling(v protoreflect.EnumValueDescriptor) string {
	prefix := enumPrefix(v.Parent().(protoreflect.EnumDescriptor))
	return strings.ToLower(strings.TrimPrefix(string(v.Name()), prefix))
}

// BySpelling resolves a seat's word back to its value, exactly and case-sensitively: the gates
// downstream compare literally, so anything looser here re-opens the hole one layer down.
func BySpelling(e protoreflect.EnumDescriptor, word string) (protoreflect.EnumValueDescriptor, bool) {
	for i := 0; i < e.Values().Len(); i++ {
		v := e.Values().Get(i)
		if !isZeroValue(v) && Spelling(v) == word {
			return v, true
		}
	}
	return nil, false
}

// SameWord reports whether two spellings differ only in case or separators — the typo class, and
// nothing wider. `closed-with-regression` and `Closed_With_Regression` are the same word;
// `closed` and `repaired_with_regression` are not. Two readers: the enum refusal names a near miss
// with it (record/enumvalue.go), and migrate resolves an era's respelling with it
// (migrate/registry.go). A near miss nobody names is a refusal that hides that the word was right.
func SameWord(a, b string) bool {
	strip := func(s string) string {
		return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(s))
	}
	return strip(a) == strip(b)
}

// enumPrefix derives the SCREAMING_SNAKE prefix protoc-gen-go gives an enum's values.
func enumPrefix(e protoreflect.EnumDescriptor) string {
	var b strings.Builder
	name := string(e.Name())
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String()) + "_"
}

// isZeroValue reports whether this is the enum's UNSPECIFIED zero, which names no seat-facing
// choice. Grade's zero is the exception that proves the rule — it is the `undefined` sentinel and
// is still never something a seat TYPES.
func isZeroValue(v protoreflect.EnumValueDescriptor) bool { return v.Number() == 0 }

// RulerOf returns the seat role that holds the gavel for a motion subject.
//
// ONE DECLARATION, TWO READERS. `internal/cli/motion` adds `rule` only to this role's command
// tree. The PASS gate, in `internal/record`, cannot import the CLI, so its refusal names the role
// from here: a seat blocked behind an unruled petition is told the bench rules it, not told to
// rule a motion whose gavel it does not hold and cannot obtain.
//
// The miss is LOUD for the same reason EnumValueDoc's is: a silent "" would put a role-shaped
// hole in a refusal message, which reads as a motion nobody has to rule.
func RulerOf(v protoreflect.EnumValueDescriptor) (string, error) {
	if r, _ := proto.GetExtension(v.Options(), E_RuledBy).(string); r != "" {
		return r, nil
	}
	return "", fmt.Errorf("recordpb: no ruler on motion subject %s — put one on the value itself, "+
		"`%s = N [(ruled_by) = \"…\"]`; a subject nobody holds the gavel for is a motion that "+
		"blocks a PASS and can never be answered", v.FullName(), v.Name())
}

// SubjectRuler is RulerOf keyed by the enum value a record carries.
func SubjectRuler(s MotionSubject) (string, error) {
	return RulerOf(s.Descriptor().Values().ByNumber(s.Number()))
}

// SeatMayFile reports whether a SEAT may put this log type on the record, and whether the value
// answered the question at all.
//
// The undeclared case is returned rather than defaulted, for the reason `Facet` returns it: a
// default here is an answer given on behalf of whoever added the value without one, and this
// facet exists because exactly that happened — `estoppel` means "the tool refused a mint" and the
// seat verb accepted it anyway, stamping source=SEAT on a refusal that never occurred (#782).
//
// The schema build already refuses a facet declared on some of an enum's values and not others,
// so in practice `declared` is false only for a vocabulary that declares it nowhere.
func SeatMayFile(t LogType) (may bool, declared bool, err error) {
	return Facet(t.Descriptor().Values().ByNumber(t.Number()), "seat_may_file")
}
