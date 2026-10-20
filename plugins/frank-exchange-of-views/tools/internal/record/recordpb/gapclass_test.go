package recordpb_test

import (
	"testing"

	pb "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// gapClasses is every GapClass value but the zero.
func gapClasses() []pb.GapClass {
	ed := pb.GapClass(0).Descriptor()
	var out []pb.GapClass
	for i := 0; i < ed.Values().Len(); i++ {
		if n := ed.Values().Get(i).Number(); n != 0 {
			out = append(out, pb.GapClass(n))
		}
	}
	return out
}

// EVERY CLASS SAYS WHERE ITS GAPS' MATERIALITY STARTS. The schema refuses a partly-annotated facet;
// this names the value that dropped it, and holds the read the write stamps from.
func TestEveryGapClassDeclaresItsMaterialDefault(t *testing.T) {
	for _, c := range gapClasses() {
		if pb.MaterialDefault(c) == pb.ClassMaterial_CLASS_MATERIAL_UNSPECIFIED {
			t.Errorf("%s declares no material_default — a gap minted under it would have no materiality to start from", pb.Word(c))
		}
	}
}

// EVERY CLASS BUT THE CATCH-ALL NAMES ITS NEAREST VALUE AND THE QUESTION THAT TELLS THEM APART —
// the membership rule: a value stays only if both can be written. The neighbour is another value's
// word, so a renamed value cannot leave a neighbour pointing at nothing.
func TestEveryGapClassNamesItsNeighbour(t *testing.T) {
	for _, c := range gapClasses() {
		if c == pb.GapClass_GAP_CLASS_UNCLASSIFIED {
			continue
		}
		vd := c.Descriptor().Values().ByNumber(c.Number())
		n, ok := pb.Neighbor(c)
		switch {
		case pb.ValueNeighbor(vd) == "":
			t.Errorf("%s declares no neighbor", pb.Word(c))
		case !ok:
			t.Errorf("%s names neighbor %q, which is not a GapClass word", pb.Word(c), pb.ValueNeighbor(vd))
		case n == c:
			t.Errorf("%s names itself as its neighbour", pb.Word(c))
		}
		if pb.Distinguisher(c) == "" {
			t.Errorf("%s declares no distinguisher from its neighbour", pb.Word(c))
		}
	}
}

// THE CATCH-ALL GOES BY GRADE AND HAS NO NEIGHBOUR: "no class fits" says nothing about how much a
// gap matters, and there is no value it is nearest to.
func TestTheCatchAllIsByGradeAndHasNoNeighbour(t *testing.T) {
	u := pb.GapClass_GAP_CLASS_UNCLASSIFIED
	if got := pb.MaterialDefault(u); got != pb.ClassMaterial_CLASS_MATERIAL_BY_GRADE {
		t.Errorf("unclassified's material default is %s, want by_grade", pb.Word(got))
	}
	if _, ok := pb.Neighbor(u); ok {
		t.Error("unclassified names a neighbour")
	}
	if pb.Distinguisher(u) != "" {
		t.Error("unclassified carries a distinguisher")
	}
	if pb.Word(u) != "unclassified" {
		t.Errorf("the catch-all is spelled %q, want unclassified", pb.Word(u))
	}
}
