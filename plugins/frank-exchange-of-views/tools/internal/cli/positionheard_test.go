package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// heardPosition is the chair holding back a PASS its plan permits, on the act fixture: the chair's
// sitting stores a plan permitting the PASS, then files its position through the verb, which the
// bench hears (gblock, 2026-10-30). It returns the position's id, as the verb's envelope names it.
func heardPosition(t *testing.T, runDir string) string {
	t.Helper()
	p := record.Plan{PassPermitted: true, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{},
		Exchanges: []record.ExchangeCount{}, Why: []string{}, StaleAreas: []record.StaleArea{}, Blockers: []record.PlanBlocker{}}
	recordtest.Seed(t, runDir, recordtest.Event(t, "red-chair", &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}))
	out := must(t, runDir, "position", "--seat-id", "red-chair", "--reason", "the PASS is withheld: the core claim rests on one source", "--json")
	var env struct {
		Result struct {
			PositionID string `json:"position_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.PositionID == "" {
		t.Fatalf("no position id on the position's envelope (%v): %s", err, out)
	}
	return env.Result.PositionID
}

func positionFixture(t *testing.T) drive {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := actFixture(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	return drive{t: t, run: runDir}
}

// THE VERB CARRIES THE WORK AN UPHELD RULING NAMES onto the record, each way: --ids as the list of
// gaps in the order named, --lens as the lens; and an overruled ruling as no work at all.
func TestThePositionRulingCarriesTheWorkItNames(t *testing.T) {
	for _, c := range []struct {
		name  string
		args  []string
		check func(h *recordpb.PositionHearing, gap string) string
	}{
		{"upheld on open gaps", []string{"--as", "upheld", "--ids", "G1", "--direction", "source the figure"}, func(h *recordpb.PositionHearing, gap string) string {
			if fmt.Sprint(h.GetGapIds()) != "["+gap+"]" || h.Lens != nil || h.GetDirection() != "source the figure" {
				return fmt.Sprintf("gaps %v, lens %v, direction %q", h.GetGapIds(), h.Lens, h.GetDirection())
			}
			return ""
		}},
		{"upheld on a lens", []string{"--as", "upheld", "--lens", lensSeat, "--direction", "audit the inference"}, func(h *recordpb.PositionHearing, _ string) string {
			if h.GetLens() != lensSeat || len(h.GetGapIds()) != 0 {
				return fmt.Sprintf("lens %q, gaps %v", h.GetLens(), h.GetGapIds())
			}
			return ""
		}},
		{"overruled", []string{"--as", "overruled"}, func(h *recordpb.PositionHearing, _ string) string {
			if h.Lens != nil || len(h.GetGapIds()) != 0 || h.Direction != nil {
				return fmt.Sprintf("an overruled ruling carries work: %v", h)
			}
			return ""
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := positionFixture(t)
			id := heardPosition(t, d.run)
			d.admitted("the bench's ruling", "judge", append([]string{"motion", "position", "rule", "--id", id, "--reason", "the bench's reasoning"}, c.args...)...)
			rules := eventsOfWord(t, d.run, "motion_rule")
			if len(rules) != 1 {
				t.Fatalf("the record holds %d rulings, want the one", len(rules))
			}
			r, _ := recordpb.BodyAs[*recordpb.MotionRule](rules[0])
			if r.GetMotionId() != id || r.GetPosition() == nil {
				t.Fatalf("the ruling is %v, want a position ruling on %s", r, id)
			}
			if m := c.check(r.GetPosition(), onlyID(t, d.run, "G")); m != "" {
				t.Error(m)
			}
		})
	}
}
