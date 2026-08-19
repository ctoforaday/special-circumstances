package recordpb

import "google.golang.org/protobuf/reflect/protoreflect"

// Word is Spelling for a TYPED enum value — `Word(ClosureClass_CLOSURE_CLASS_RISK_ACCEPTED)` is
// `risk_accepted`, without the caller having to walk to a descriptor first.
//
// It exists because the descriptor walk is four lines of ceremony
// (`v.Descriptor().Values().ByNumber(protoreflect.EnumNumber(v))`, plus a nil check for a number
// the schema does not declare) and EVERY reader that renders an enum needs it. Written out at each
// site it would be the duplication this migration exists to remove, and — worse — the nil arm is
// the one a hurried copy drops, which turns a value from a newer release into a panic instead of a
// blank.
//
// THE ZERO RENDERS AS THE EMPTY STRING, and that is the load-bearing half.
//
// `Spelling` on the UNSPECIFIED zero returns the literal word `unspecified`, which no seat ever
// types and which the pre-migration record never contained: there the fact was ABSENT and
// `p.Str(key)` returned "". A reader that printed `unspecified` into a ledger, or looked it up in
// a mass table, would be reading a word the schema invented — so the zero is mapped back to the
// empty string, which is what every existing consumer already branches on. `isZeroValue` states
// the same rule from the other direction for the help text.
//
// IT IS NOT THE WHOLE VOCABULARY STORY. Two enums are spelled with HYPHENS on the seat-facing
// surface while `Spelling` derives underscores from the generated names — see `record.GradeStr`,
// which is where that join is made and why. Word is the plain case; anything hyphenated must not
// use it directly.
func Word(e protoreflect.Enum) string {
	if e == nil {
		return ""
	}
	n := e.Number()
	if n == 0 {
		return ""
	}
	vd := e.Descriptor().Values().ByNumber(n)
	if vd == nil {
		// A number this binary's schema does not declare. The read path refuses such a line
		// before it reaches any reader (ClassifyLine stage 4/5), so this is reachable only from
		// an in-memory value — and a blank is the honest answer, not a panic.
		return ""
	}
	return Spelling(vd)
}
