package migrate

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// TranslatePlanParties takes `occasions` off a stored plan's parties (EventSchema 53), and re-keys and
// re-prints the plan when any carried one; a plan with none is left untouched, its bytes the bytes
// stored. A bench party with no gap was convened for a hearing alone, and leaves the plan: a hearing is
// on no plan now. The plan's `why` lines are the prose it printed, and carry the words they were
// printed with.
func TranslatePlanParties(pl *recordpb.Plan) error {
	var plan map[string]json.RawMessage
	if err := json.Unmarshal([]byte(pl.GetPrinted()), &plan); err != nil {
		return fmt.Errorf("this stored plan does not read as an object: %v", err)
	}
	var parties []map[string]json.RawMessage
	if err := json.Unmarshal(plan["parties"], &parties); err != nil {
		return fmt.Errorf("this stored plan's parties do not read as a list: %v", err)
	}
	changed := false
	kept := make([]map[string]json.RawMessage, 0, len(parties))
	for _, p := range parties {
		if _, carried := p["occasions"]; !carried {
			kept = append(kept, p)
			continue
		}
		changed = true
		delete(p, "occasions")
		var seat string
		var gaps []string
		if err := json.Unmarshal(p["seat_id"], &seat); err != nil {
			return fmt.Errorf("this stored plan's party does not name a seat: %v", err)
		}
		if err := json.Unmarshal(p["gap_ids"], &gaps); err != nil {
			return fmt.Errorf("this stored plan's party %s does not list its gaps: %v", seat, err)
		}
		if !record.SeatOwesOccasion(seat) || len(gaps) > 0 {
			kept = append(kept, p)
		}
	}
	if !changed {
		return nil
	}
	b, err := json.Marshal(kept)
	if err != nil {
		return err
	}
	plan["parties"] = b
	raw, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	p, _, err := record.ReadPlan(raw)
	if err != nil {
		return fmt.Errorf("this stored plan does not read as a plan with its occasions gone: %v", err)
	}
	pl.Printed = proto.String(p.WithKey().Printed())
	return nil
}
