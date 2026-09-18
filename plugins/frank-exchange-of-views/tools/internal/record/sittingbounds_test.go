package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE EARLIER OF TWO CLOSES A SITTING, AND A LATE STOP IS NOT IT (#1002, finding 5). A sitting ends
// at whichever comes first after it began — the seat's next register, or the stop of the agent that
// sat it. The hook writes the stop when the agent RETURNS, and an agent that returns after the seat
// has been dispatched again leaves its stop behind the register that already ended the sitting.
// Taking the stop whenever there is one hands the first sitting every act of the second: its edits,
// its receipts and the gaps it was engaged on all land in a sitting that was over.
func TestAStopThatLandsAfterTheNextRegisterClosesNothing(t *testing.T) {
	evs := []*Event{
		recordtest.At(t, "red-chair", "red-chair:dispatch:d1", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, "red-chair", "red-chair:dispatch:d2", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{"G2"}}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e2", &recordpb.BlueEdit{Answers: proto.String("G2")}),
		// blue-a's agent returns only now, one whole sitting late.
		recordtest.At(t, HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		recordtest.At(t, HarnessSeat, "harness:sitting_close:s2", &recordpb.SittingClose{AgentId: proto.String("blue-b")}),
	}
	ss := BlueSittings(evs, WhileRunning)
	if len(ss) != 2 {
		t.Fatalf("blue sittings = %d, want 2", len(ss))
	}
	if got := strings.Join(actTypes(ss[0]), ","); got != "EVENT_TYPE_REGISTER,EVENT_TYPE_BLUE_EDIT" {
		t.Errorf("the first sitting's acts = %s, want its own register and edit — the late stop closed a sitting that had already ended", got)
	}
	if got := strings.Join(actTypes(ss[1]), ","); got != "EVENT_TYPE_REGISTER,EVENT_TYPE_BLUE_EDIT" {
		t.Errorf("the second sitting's acts = %s, want its own register and edit", got)
	}
	if o := ManifestOwed(evs, WhileRunning); strings.Join(o.Gaps, ",") != "G1,G2" || o.Unresolved != 0 {
		t.Errorf("owed = %+v, want one gap per sitting and both sittings closed", o)
	}
}

// A REPAIR JOINS ITS OWN SEAT'S SITTING AND NOBODY ELSE'S (#1002, finding 5). THE WRITER ALREADY
// REFUSES THIS — checkRepair admits only a register that opened a sitting of the SAME seat — so
// this is defence in depth at the READ: the closer joins a repair to the sitting it names by key,
// and a key is not a seat. Without the seat test, one blue seat naming another's register would
// extend that other seat's sitting over its own acts, and the manifest and record-parity would
// answer for a sitting that never held them.
func TestARepairNamingAnotherSeatsSittingJoinsNothing(t *testing.T) {
	opened := "blue-respond:register:#1"
	evs := []*Event{
		recordtest.At(t, "red-chair", "red-chair:dispatch:d1", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.At(t, "blue-respond", opened, &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		// Another blue seat claims blue-respond's sitting as the one it repairs.
		recordtest.At(t, "blue-synthesize", "blue-synthesize:register:#1", &recordpb.Register{AgentId: proto.String("syn-a"), RepairsSitting: proto.String(opened)}),
		// An act by blue-respond after its sitting closed, inside the span the foreign repair would open.
		recordtest.At(t, "blue-respond", "blue-respond:manifest_row:m1", &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("recomputed")}),
	}
	ss := BlueSittings(evs, WhileRunning)
	if len(ss) != 1 {
		t.Fatalf("blue sittings = %d, want 1", len(ss))
	}
	if got := strings.Join(actTypes(ss[0]), ","); got != "EVENT_TYPE_REGISTER,EVENT_TYPE_BLUE_EDIT" {
		t.Errorf("blue-respond's sitting = %s, want it closed by its own stop — another seat's register does not reopen it", got)
	}

	// The writer's half, which is why the reader's is depth rather than the only guard.
	err := checkRepair(evs, "blue-synthesize", opened)
	if err == nil || !strings.Contains(err.Error(), "is not a register that opened a sitting of blue-synthesize") {
		t.Fatalf("the writer's refusal = %v, want it naming the seat the register does not belong to", err)
	}
}
