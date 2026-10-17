package migrate

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A STORED PLAN IS CARRIED ONLY WHERE THE EVENTS IT NAMES KEEP THEIR NUMBERS: its head and each
// stale area's pin, inside the printed bytes.
func TestAStoredPlanWhoseEventsWereRenumberedIsRefused(t *testing.T) {
	p := record.Plan{Head: 4, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{}, Exchanges: []record.ExchangeCount{},
		StaleAreas: []record.StaleArea{{SeatID: "red-lens-voice", Pin: 3}}, Why: []string{}, Blockers: []record.PlanBlocker{}, PassPermitted: true}
	pl := &recordpb.Plan{Printed: proto.String(p.Printed())}
	if got := planRenumbered(pl, map[int64]int64{3: 3, 4: 4}); got != "" {
		t.Errorf("a plan whose events keep their numbers is refused: %s", got)
	}
	for name, newID := range map[string]map[int64]int64{
		"the head renumbered":   {3: 3, 4: 5},
		"a pin renumbered":      {3: 2, 4: 4},
		"the head never landed": {3: 3},
	} {
		if got := planRenumbered(pl, newID); !strings.Contains(got, "carried as printed") {
			t.Errorf("%s: %q, want a refusal", name, got)
		}
	}
}
