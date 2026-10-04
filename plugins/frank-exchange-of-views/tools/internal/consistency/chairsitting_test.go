package consistency

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// chairBracket is the hook's SubagentStart for a chair dispatched under agent: it opens the chair's
// sitting whether or not the chair ever registers.
func chairBracket(t *testing.T, agent string) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, record.HarnessSeat, "harness:sitting_open:"+agent, &recordpb.SittingOpen{
		AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:red-chair"), SeatId: proto.String("red-chair")})
}

// chairRegisters is the chair's register under agent: it joins agent's bracket when the hook
// bracketed it, and opens a sitting of its own when not.
func chairRegisters(t *testing.T, agent string) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, "red-chair", "red-chair:register:"+agent, &recordpb.Register{AgentId: proto.String(agent)})
}

// THE ORACLE COUNTS THE CHAIR'S SITTINGS THE WAY THE WRITE PATH STORES THEM, from the events. A
// chair sitting the hook opened with no register is an epoch; a bracket and the register that joins
// it are ONE epoch, and a register that repairs a sitting opens none. The board reads the stored
// epoch, so an oracle that counted chair registers reported a divergence on every live run — the
// second opinion disagreeing for its own reason. The oracle reads each of those rules off the raw
// fields, so a write-path predicate that broke one would store an epoch this check disagrees with.
func TestTheOraclesEpochCountsTheChairsStoredSittings(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		chairBracket(t, "C1"), chairRegisters(t, "C1"), // epoch 1: a bracket and its own register
		mint(t, "red-chair", "G1"),
		mint(t, "red-chair", "G2"),
		chairBracket(t, "C2"), // epoch 2: the hook opened it and the chair never registered
		redClose(t, "red-chair", "G1", recordpb.Disposition_DISPOSITION_REPAIRED),
		chairRegisters(t, "C3"), // epoch 3: a register the hook never bracketed (a resume)
		// still epoch 3: a repair of the chair's sitting joins it
		recordtest.At(t, "red-chair", "red-chair:register:C4", &recordpb.Register{AgentId: proto.String("C4"),
			RepairsSitting: proto.String("red-chair:register:C3")}),
		redClose(t, "red-chair", "G2", recordpb.Disposition_DISPOSITION_REPAIRED),
	)
	check(t, dir)
}
