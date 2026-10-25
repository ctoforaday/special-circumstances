package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// planKindRenames is each blocker kind word an archived plan prints that the live table spells
// otherwise (EventSchema 46): a citation or proof nobody has checked as it stands is one kind,
// whether an edit reopened it or nobody ever checked it, and a proof judged unsound is one failed
// check beside a citation nobody could read. These describe ARCHIVED bytes, so they are migrate's
// own table, and the live table never holds the old words.
var planKindRenames = map[record.BlockerKind]record.BlockerKind{
	"reopened":      record.BlockerUnchecked,
	"unsound_proof": record.BlockerFailedCheck,
}

// TranslatePlanKinds respells a stored plan's archived blocker kinds (planKindRenames), and re-keys
// and re-prints the plan when any moved; a plan whose kinds are all current is left untouched, its
// bytes the bytes stored. The kind is the plan's field, so it is translated. The plan's `why` lines
// are the prose it printed, and carry the words they were printed with.
func TranslatePlanKinds(pl *recordpb.Plan) error {
	p, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		return fmt.Errorf("this stored plan does not read as a plan: %v", err)
	}
	changed := false
	for i, b := range p.Blockers {
		if to, ok := planKindRenames[b.Kind]; ok {
			p.Blockers[i].Kind, changed = to, true
		}
	}
	if !changed {
		return nil
	}
	pl.Printed = proto.String(p.WithKey().Printed())
	return nil
}
