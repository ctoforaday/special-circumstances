package recordpb

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// THE GAP CLASS'S OPTIONS, READ OFF THE VALUE. Each GapClass value declares its materiality
// (`material_default`, a facet the vocabulary table carries), the value nearest to it and the
// question that tells the two apart (`neighbor`, `distinguisher`, read by help and its test).

func gapClassValue(c GapClass) protoreflect.EnumValueDescriptor {
	return c.Descriptor().Values().ByNumber(c.Number())
}

// MaterialDefault is where a gap of class c starts its materiality. UNSPECIFIED for a value that
// declares none, which the schema refuses to build, so a live value always answers.
func MaterialDefault(c GapClass) ClassMaterial {
	n, ok, err := EnumFacet(gapClassValue(c), "material_default")
	if err != nil || !ok {
		return ClassMaterial_CLASS_MATERIAL_UNSPECIFIED
	}
	return ClassMaterial(n)
}

// ValueNeighbor is the word a value's `neighbor` option names; "" where it declares none.
func ValueNeighbor(vd protoreflect.EnumValueDescriptor) string {
	if vd == nil {
		return ""
	}
	a, err := annotationsOf(vd)
	if err != nil {
		return ""
	}
	return a.Neighbor
}

// ValueDistinguisher is a value's `distinguisher` option; "" where it declares none.
func ValueDistinguisher(vd protoreflect.EnumValueDescriptor) string {
	if vd == nil {
		return ""
	}
	a, err := annotationsOf(vd)
	if err != nil {
		return ""
	}
	return a.Distinguisher
}

// Neighbor is the value nearest to c, which its distinguisher tells it apart from. False for the
// catch-all, which has none, and for a word that resolves to no value.
func Neighbor(c GapClass) (GapClass, bool) {
	w := ValueNeighbor(gapClassValue(c))
	if w == "" {
		return 0, false
	}
	nv, ok := BySpelling(c.Descriptor(), w)
	if !ok {
		return 0, false
	}
	return GapClass(nv.Number()), true
}

// GapClassesOf is the gap classes a reading of a citation with outcome o may mint its gap in — the
// outcome's `gap_classes` option, in declaration order — and nil for an outcome that declares none:
// an outcome that does not find against a sentence admits no gap.
func GapClassesOf(o SourceOutcome) []GapClass {
	return outcomeClasses(o, func(a *ValueAnnotations) []GapClass { return a.GapClasses })
}

// FoundSourceGapClassesOf is the gap classes a reading of a source the lens found (no anchor) with
// outcome o may mint its gap in — the outcome's `found_source_gap_classes` option — and nil for an
// outcome that declares none (`absent`, `unreachable`).
func FoundSourceGapClassesOf(o SourceOutcome) []GapClass {
	return outcomeClasses(o, func(a *ValueAnnotations) []GapClass { return a.FoundSourceGapClasses })
}

// outcomeClasses reads one of o's class lists off its row of the generated annotation table.
func outcomeClasses(o SourceOutcome, pick func(*ValueAnnotations) []GapClass) []GapClass {
	vd := o.Descriptor().Values().ByNumber(o.Number())
	if vd == nil {
		return nil
	}
	a, err := annotationsOf(vd)
	if err != nil {
		return nil
	}
	return pick(a)
}

// Distinguisher is the question that tells c from its neighbour; empty for the catch-all.
func Distinguisher(c GapClass) string { return ValueDistinguisher(gapClassValue(c)) }
