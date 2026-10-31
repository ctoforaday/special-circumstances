package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// CarryPlanHalted puts on a stored plan whether a halt was on the record when the plan was stored —
// the plan's `halted`, which a plan stored before EventSchema 54 does not print. It is read off the
// migrated record as it stands at the plan's place (record.HaltedAt, the predicate `dispatch next`
// prints from), so it is the record's own fact joined in, never a value supplied. The key leaves
// `halted` out, but it is taken over the plan as the verb prints it, a field the archive never printed
// included, so the plan is re-keyed and re-printed.
//
// What no translation can supply: before EventSchema 54 a halted record still readied parties, so an
// archived plan stored after a halt prints `halted` true beside the parties it dispatched.
func CarryPlanHalted(pl *recordpb.Plan, q recordsql.Querier) error {
	p, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		return fmt.Errorf("this stored plan does not read as a plan: %v", err)
	}
	if p.Halted, err = record.HaltedAt(q); err != nil {
		return err
	}
	pl.Printed = proto.String(p.WithKey().Printed())
	return nil
}
