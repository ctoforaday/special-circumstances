package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// `register --repair-sitting` IS THE REPAIR, THROUGH THE VERB (#1002). Blue sat on G1, edited, and
// its agent returned with no position or revision. The re-prompted seat registers as the repair:
// the verb writes the register naming that sitting and says so; without the flag the same call
// writes a register that opens a sitting of its own.
func TestRegisterWithRepairSittingNamesTheSittingItRepairs(t *testing.T) {
	blueSat := func(t *testing.T) string {
		runDir := newRun(t)
		stageEvents(t, runDir,
			recordtest.At(t, "red-chair", "red-chair:stage:d", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
			recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
			recordtest.At(t, "blue-respond", "blue-respond:stage:e", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("b")}),
			recordtest.At(t, record.HarnessSeat, "harness:stage:s", &recordpb.SittingClose{AgentId: proto.String("blue-a"), AgentType: proto.String("frank-exchange-of-views:blue-researcher")}),
		)
		return runDir
	}

	runDir := blueSat(t)
	out, err := run(t, "register", "--run", runDir, "--seat-id", "blue-respond", "--repair-sitting")
	if err != nil {
		t.Fatalf("register --repair-sitting: %v\n%s", err, out)
	}
	if !strings.Contains(out, `as the repair of the sitting blue-respond:register:#1 opened`) {
		t.Errorf("the verb does not say which sitting it repairs:\n%s", out)
	}
	if r := lastBody(t, runDir, &recordpb.Register{}); r.GetRepairsSitting() != "blue-respond:register:#1" {
		t.Errorf("the written register = %v, want repairs_sitting naming the sitting it repairs", r)
	}

	plain := blueSat(t)
	if _, err := run(t, "register", "--run", plain, "--seat-id", "blue-respond"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if r := lastBody(t, plain, &recordpb.Register{}); r.RepairsSitting != nil {
		t.Errorf("a register without the flag claimed a repair: %v", r)
	}
}
