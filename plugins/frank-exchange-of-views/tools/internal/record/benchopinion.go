package record

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// benchOpinion is a docket ruling's opinion fields: the rule applied, what pulled against it, what a
// human should look at again, and what the losing party may no longer assert.
var benchOpinion = []protoreflect.Name{"principle", "tension", "review_flag", "settled"}

// requireBenchOpinion is the opinion a docket ruling owes by its writer, which no column can see: the
// bench states all four fields (`settled` may be empty — a ruling can bar no proposition), and the
// gap's assignee, accepting a proposed act at its own sitting, states none: its accept is its own act
// through its own verb, and nothing on the docket reached the bench.
func requireBenchOpinion(seatID string, d *recordpb.DocketRuling) error {
	if seatclass.RoleOfSeat(seatID) == "bench" {
		return recordpb.CheckSaid(d, benchOpinion, "settled")
	}
	m := d.ProtoReflect()
	for _, n := range benchOpinion {
		if fd := m.Descriptor().Fields().ByName(n); m.Has(fd) {
			return feov.Errorf(feov.Validation, "--%s is the bench's opinion on a ruling it writes; an accept by the gap's assignee carries none. Nothing was recorded", recordpb.FlagFor(fd))
		}
	}
	return nil
}
