package lens

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A grade axis has one definition: the `(means)` on its GradeDimension value. Every verb that
// takes the axis as a flag prints that sentence, so a seat grading a gap and a seat contesting the
// grade read the same words for the same axis.
//
// The axes are walked off the descriptor, and each verb states how many it takes: a verb that
// loses an axis flag, or an axis added to the schema that no verb takes, fails here.
func TestEveryGradeAxisFlagPrintsItsAxisMeaning(t *testing.T) {
	axes := recordpb.GradeDimension(0).Descriptor().Values()
	for _, tc := range []struct {
		verb *cobra.Command
		axes int
	}{
		{newMint(), 4},
		{newRegrade(), 4},
		{newFinding(), 3}, // a finding grades the defect; what the fix costs is the gap's question
	} {
		carried := 0
		for i := 0; i < axes.Len(); i++ {
			vd := axes.Get(i)
			if vd.Number() == 0 {
				continue
			}
			f := tc.verb.Flags().Lookup(recordpb.Spelling(vd))
			if f == nil {
				continue
			}
			carried++
			doc, err := recordpb.EnumValueDoc(vd)
			if err != nil {
				t.Fatalf("%s: %v", vd.FullName(), err)
			}
			if !strings.Contains(f.Usage, doc) {
				t.Errorf("%s --%s reads %q, which does not carry the axis's meaning %q", tc.verb.Name(), f.Name, f.Usage, doc)
			}
		}
		if carried != tc.axes {
			t.Errorf("%s takes %d grade axes as flags, want %d", tc.verb.Name(), carried, tc.axes)
		}
	}
}
