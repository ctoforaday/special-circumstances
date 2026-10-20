package recordpb

import (
	"google.golang.org/protobuf/proto"
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
	if vd == nil || vd.Options() == nil || !proto.HasExtension(vd.Options(), E_Neighbor) {
		return ""
	}
	return proto.GetExtension(vd.Options(), E_Neighbor).(string)
}

// ValueDistinguisher is a value's `distinguisher` option; "" where it declares none.
func ValueDistinguisher(vd protoreflect.EnumValueDescriptor) string {
	if vd == nil || vd.Options() == nil || !proto.HasExtension(vd.Options(), E_Distinguisher) {
		return ""
	}
	return proto.GetExtension(vd.Options(), E_Distinguisher).(string)
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

// Distinguisher is the question that tells c from its neighbour; empty for the catch-all.
func Distinguisher(c GapClass) string { return ValueDistinguisher(gapClassValue(c)) }
