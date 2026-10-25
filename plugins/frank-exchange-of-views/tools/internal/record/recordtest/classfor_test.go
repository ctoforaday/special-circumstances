package recordtest

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A FIXTURE THAT SETS ITS CLASS MATERIAL BY HAND TAKES ITS CLASS FROM ClassFor, so the class it
// names and the stamp it carries agree.
func TestClassForHasTheMaterialityItNames(t *testing.T) {
	for _, cm := range []recordpb.ClassMaterial{
		recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS,
		recordpb.ClassMaterial_CLASS_MATERIAL_NEVER,
		recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE,
	} {
		if got := recordpb.MaterialDefault(ClassFor(cm)); got != cm {
			t.Errorf("ClassFor(%s) = %s, whose material default is %s", recordpb.Word(cm), recordpb.Word(ClassFor(cm)), recordpb.Word(got))
		}
	}
}
