package migrate

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// TranslatePlanVerdict carries a stored plan's `verdict` to `passed` (EventSchema 78): a plan stored
// from EventSchema 39 prints the verdict its sitting recorded, and the plan now prints whether that
// sitting recorded the PASS. A PASS becomes `passed: true`; a FAIL becomes nothing, as the FAIL
// itself does (verdictEntry); any other word is refused. The verdict is not in the plan's key, so the
// key stands. A plan printing no verdict is left untouched, its bytes the bytes stored.
//
// IT RUNS FIRST, AND LEAVES THE BYTES UNPRINTED: every later translation reads the plan as a plan,
// which a `verdict` key is not, and CarryPlanPass, the last, prints the plan as the verb does.
func TranslatePlanVerdict(pl *recordpb.Plan) error {
	var plan map[string]json.RawMessage
	if err := json.Unmarshal([]byte(pl.GetPrinted()), &plan); err != nil {
		return fmt.Errorf("this stored plan does not read as an object: %v", err)
	}
	raw, printed := plan["verdict"]
	if !printed {
		return nil
	}
	var w string
	if err := json.Unmarshal(raw, &w); err != nil {
		return fmt.Errorf("this stored plan's verdict does not read as a word: %v", err)
	}
	delete(plan, "verdict")
	switch {
	case recordpb.SameWord(w, archivedPass):
		plan["passed"] = json.RawMessage("true")
	case recordpb.SameWord(w, retiredVerdict):
	default:
		return fmt.Errorf("this stored plan prints verdict %q, which is neither the PASS nor the archived FAIL", w)
	}
	b, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	pl.Printed = proto.String(string(b))
	return nil
}

// CarryPlanPass puts on a stored plan the PASS its chair sitting had recorded when the plan was
// stored — the plan's `passed`, which a plan stored before EventSchema 39 does not print. It is read
// off the migrated record as it stands at the plan's place (record.ChairPassedAt, the predicate
// `dispatch next` prints from), so it is the record's own fact joined in, never a value supplied: a
// plan printing a PASS the record does not hold at its place is refused. A plan whose bytes are not
// as the verb prints them — TranslatePlanVerdict's, or one no translation touched — is printed.
//
// What no translation can supply: before EventSchema 39 a verdict printed no plan, and the chair
// relayed the plan it asked for before recording its verdict. Such a sitting stores no plan carrying
// the PASS, and a migrated record holds none either.
func CarryPlanPass(pl *recordpb.Plan, q recordsql.Querier) error {
	p, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		return fmt.Errorf("this stored plan does not read as a plan: %v", err)
	}
	passed, err := record.ChairPassedAt(q)
	if err != nil {
		return err
	}
	if p.Passed && !passed {
		return fmt.Errorf("this stored plan prints the PASS and its sitting had recorded none")
	}
	p.Passed = passed
	if out := p.Printed(); out != pl.GetPrinted() {
		pl.Printed = proto.String(out)
	}
	return nil
}
