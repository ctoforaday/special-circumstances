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

// undocumentedEnums are the enums whose values no seat ever types, so no --help renders them.
// Listed explicitly: an exemption somebody decided beats an exemption that emerges from a
// missing map entry.
var undocumentedEnums = map[protoreflect.FullName]bool{
	"feov.record.v1.EventType":     true, // stamped by the tool, never chosen
	"feov.record.v1.SchemaVersion": true, // the format discriminator
}

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

// Usage renders a flag's help from the enum itself, so the contract a seat reads is the contract
// the write path enforces. One declaration, two readers — which is the rule the old EnumFields
// table existed to keep and the reason it could not simply be deleted.
func Usage(e protoreflect.EnumDescriptor) (string, error) {
	var b strings.Builder
	for i := 0; i < e.Values().Len(); i++ {
		v := e.Values().Get(i)
		if isZeroValue(v) {
			continue
		}
		doc, err := EnumValueDoc(v)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "\n  %s — %s", Spelling(v), doc)
	}
	return b.String(), nil
}

// Spelling is the word a seat types: the enum value's name with its type prefix removed and
// lowercased, so CLOSURE_CLASS_RISK_ACCEPTED reads as `risk_accepted`.
func Spelling(v protoreflect.EnumValueDescriptor) string {
	prefix := enumPrefix(v.Parent().(protoreflect.EnumDescriptor))
	return strings.ToLower(strings.TrimPrefix(string(v.Name()), prefix))
}

// Names is the bare vocabulary, in declaration order, for the readers that only need the words.
func Names(e protoreflect.EnumDescriptor) []string {
	var out []string
	for i := 0; i < e.Values().Len(); i++ {
		if v := e.Values().Get(i); !isZeroValue(v) {
			out = append(out, Spelling(v))
		}
	}
	return out
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
// `closed` and `closed_with_regression` are not. Carried over from enums.go, where it exists
// because closure_class gates an invariant and a near-miss silently took the other branch.
func SameWord(a, b string) bool {
	strip := func(s string) string {
		return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(s))
	}
	return strip(a) == strip(b)
}

// NearMiss finds a declared value that differs from the seat's word only by the typo class, so a
// refusal can name what would have worked instead of only listing the set.
func NearMiss(e protoreflect.EnumDescriptor, word string) (string, bool) {
	for _, n := range Names(e) {
		if n != word && SameWord(n, word) {
			return n, true
		}
	}
	return "", false
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
