package migrate

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// CarryPlanVerdict puts on a stored plan the verdict its chair sitting had recorded when the plan
// was stored — the plan's `verdict`, which a plan stored before EventSchema 39 does not print. It is
// read off the migrated record as it stands at the plan's place (record.ChairVerdictAt, the
// predicate `dispatch next` prints from), so it is the record's own fact joined in, never a value
// supplied: a plan whose sitting had recorded none is left as stored. The verdict is not in the
// plan's key, so the key stands.
//
// A PLAN STORED AFTER A FAIL PRINTS IT (EventSchema 39 to 75), and the FAIL translates to no event
// (verdictEntry): the record at the plan's place holds no verdict, so the plan prints none. Only the
// retired word is stripped; any other verdict the record does not hold at the plan's place is
// refused below.
//
// What no translation can supply: before EventSchema 39 a verdict printed no plan, and the chair
// relayed the plan it asked for before recording its verdict. Such a sitting stores no plan carrying
// the verdict, and a migrated record holds none either.
func CarryPlanVerdict(pl *recordpb.Plan, q recordsql.Querier) error {
	p, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		return fmt.Errorf("this stored plan does not read as a plan: %v", err)
	}
	v, err := record.ChairVerdictAt(q)
	if err != nil {
		return err
	}
	switch {
	case v == "" && recordpb.SameWord(string(p.Verdict), retiredVerdict):
		p.Verdict = ""
		pl.Printed = proto.String(p.WithKey().Printed())
		return nil
	case v == "" || p.Verdict == v:
		return nil
	case p.Verdict != "":
		return fmt.Errorf("this stored plan prints verdict %q and its sitting had recorded %q", p.Verdict, v)
	}
	p.Verdict = v
	pl.Printed = proto.String(p.WithKey().Printed())
	return nil
}
