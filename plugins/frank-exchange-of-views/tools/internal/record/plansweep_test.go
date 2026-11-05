package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

// THE ENTRY POINTS NO OTHER TEST OF THIS PACKAGE CALLS, CALLED ONCE SO THE GUARD PLANS THEIR
// STATEMENTS. Their behaviour is tested by the packages that use them (the CLI, the dashboard,
// the hooks); what this suite owes them is a plan, and planguard.Suite fails the run naming any
// statement in the source that no test reached. Nothing here is asserted: an entry point that
// stops reaching its statement is reported by that census, by file and line.
func TestPlanSweepReachesTheStatementsNoOtherTestIssues(t *testing.T) {
	b := retBoard(t).sit(2, evLens).mint(evLens, "G1", "high").edit("G1", "report", "account").
		add(evLens, &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("b"),
			Filing: &recordpb.Motion_Position{Position: &recordpb.PositionMotion{PositionId: proto.String("P1")}}})
	run := b.seed()

	_, _, _ = RegisteredSeatOfAgent(run, "agent-1")
	_, _ = ExistingBlueEditByKey(run, "blue-respond", "E1")
	_, _ = ChangesOf(run, "G1")
	_, _ = ExistingCiteByKey(run, evLens, "k")
	_, _ = ExistingComputationByKey(run, evLens, "k")
	_, _, _, _, _ = AnchorRetiredAt(run, "A1")
	_, _ = ReportAtPart(run, "harness")
	_, _ = ReportHasParts(run)
	_, _, _ = PartStanding(run, "harness")
	_, _ = ExistingProposalByText(run, "blue-respond", "q", "h")
	_, _, _ = ServeReport(Identity{Run: run, SeatID: evLens})
	_, _ = MergedEvents(run) // decodes every arm some row names, the position motion's included

	_, _ = AppendSeatTurns(run, "agent-1", []seatturn.Turn{{Index: 0, Model: "m"}})
	_, _ = SeatTimingOf(run)
	_, _ = SeatMetrics(run)

	admitted := &recordpb.EnvelopeAdmitted{AgentId: proto.String("agent-1")}
	_, _ = Append(Identity{Run: run, SeatID: "harness"}, admitted, OneAdmissionASitting(admitted))
	_, _ = Append(Identity{Run: run, SeatID: "harness"}, &recordpb.Log{Text: proto.String("t")}, Check(func(l *Landing) error {
		_, _ = l.EditSpans()
		return ErrWritesNothing
	}))
}
