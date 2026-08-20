package record

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE WRITE PATH'S HALF OF EVERY ENUM JOIN.
//
// recordpb.Word turns a value into the schema's spelling; these turn a seat's word back into the
// value. The pairing is the rule, not a convenience: a conversion that exists in only one
// direction is how two vocabularies drift apart — the writer invents its own mapping, the reader
// keeps another, and nothing can see them disagree. GradeOf states the same thing beside GradeStr.
//
// EVERY ONE RETURNS ok, AND EVERY CALLER MUST REFUSE ON false. The zero of each of these enums is
// UNSPECIFIED, which the record reserves for "the seat never said". Recording it for a word the
// schema does not know would turn a typo into a silently unset field, and an unset field reads
// downstream exactly like a question nobody was asked — which is the defect class this migration
// exists to remove.
//
// Resolution is EXACT and case-sensitive (recordpb.BySpelling is, by design, and its own test pins
// that `PASS` does not resolve to `pass`). Where a seat-facing vocabulary differs from the schema's
// spelling by a mechanical rule, the fold belongs in that enum's own function — see RunOutcomeOf,
// which lowercases because `bench outcome` still takes the capital word.

func enumOf[E ~int32](d protoreflect.EnumDescriptor, word string) (E, bool) {
	vd, ok := recordpb.BySpelling(d, word)
	if !ok {
		return E(0), false
	}
	return E(vd.Number()), true
}

// SoundnessOf resolves `sound` / `unsound` — the lens's judgement of what a re-run script actually
// computes, which is a different question from whether it reproduced.
func SoundnessOf(word string) (recordpb.Soundness, bool) {
	return enumOf[recordpb.Soundness](recordpb.Soundness(0).Descriptor(), word)
}

// SourceOutcomeOf resolves a citation verdict. Its negative half is load-bearing: `refutes` means
// the source contradicts the claim and `absent` means it is simply not there, and collapsing either
// into a weaker word is what the axis was widened to prevent.
func SourceOutcomeOf(word string) (recordpb.SourceOutcome, bool) {
	return enumOf[recordpb.SourceOutcome](recordpb.SourceOutcome(0).Descriptor(), word)
}

// ConfidenceOf resolves how sure the seat is of the determination it just made — a separate
// question from what the determination WAS.
func ConfidenceOf(word string) (recordpb.Confidence, bool) {
	return enumOf[recordpb.Confidence](recordpb.Confidence(0).Descriptor(), word)
}

// GradeDimensionOf resolves the axis a grade motion contests. The seat's word for one of these is
// `cx` while the schema spells `complexity_cost` — that fold is the flag's business, not this
// function's, and it is why the surface has its own alias rather than a second spelling here.
func GradeDimensionOf(word string) (recordpb.GradeDimension, bool) {
	return enumOf[recordpb.GradeDimension](recordpb.GradeDimension(0).Descriptor(), word)
}

// PetitionClassOf resolves the ground a petition stands on. Four values, and the distinction
// between them is what the bench rules against — a petition filed under the wrong one asks a
// different question than the seat meant.
func PetitionClassOf(word string) (recordpb.PetitionClass, bool) {
	return enumOf[recordpb.PetitionClass](recordpb.PetitionClass(0).Descriptor(), word)
}

// AvenueStatusOf resolves a line of inquiry's fate. `proposed` is a real state and not a failure
// to choose one — it is the state the pre-schema shape could not express, which forced blue to
// declare a fate before it had one.
func AvenueStatusOf(word string) (recordpb.AvenueStatus, bool) {
	return enumOf[recordpb.AvenueStatus](recordpb.AvenueStatus(0).Descriptor(), word)
}
