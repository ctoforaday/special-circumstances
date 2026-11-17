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
	_, _, _ = AnchorRetiredAt(run, "A1")
	_, _ = EvidenceDemandOf(run, "G1")
	if db, err := openRunForRead(run); err == nil && db != nil {
		_ = requireCitationUncited(db, []string{"C-1"})
	}
	_, _ = ReportAtPart(run, "harness")
	_, _ = ReportHasParts(run)
	_, _, _ = PartStanding(run, "harness")
	_, _ = ExistingProposalByText(run, "blue-respond", "q", "h")
	_, _, _ = ServeReport(Identity{Run: run, SeatID: evLens})
	// A proposed close and a proposed supersede, and a tombstone, so the decode reads their tables.
	// A seat's proposal waits on the gap's one exchange: blue's edit leaves the lens its re-audit, so
	// the round ends with the lens's sitting.
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(evLens), GapIds: []string{"G1"}})
	sit(t, run, evLens)
	sit(t, run, evLens)
	blue := Identity{Run: run, SeatID: "blue-respond"}
	if _, err := Append(blue, &recordpb.Motion{MotionId: proto.String("M2"), Basis: proto.String("b"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
		Act: &recordpb.Motion_Close{Close: &recordpb.Close{GapId: proto.String("G1"),
			ClosureClass: recordpb.Disposition_DISPOSITION_MOOT.Enum(), Prose: proto.String("moot"),
			AnchorSeat: proto.String("blue-respond"), AnchorTool: proto.String("the report read"), AnchorTarget: proto.String("the report's sentence")}}}); err != nil {
		t.Fatal(err)
	}
	// A gap holds one proposal at a time (F5): the assignee accepts the close, which closes G1, and a
	// closed gap takes the proposed supersede (F8).
	mustAppend(t, Identity{Run: run, SeatID: evLens}, actRule("M2", true, evLens))
	successor := successorMint("G2", "G1")
	if _, err := Append(blue, &recordpb.Motion{MotionId: proto.String("M3"), Basis: proto.String("b"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
		Act:    &recordpb.Motion_Supersede{Supersede: successor}}); err != nil {
		t.Fatal(err)
	}
	// The bench accepts the supersede in the docket sitting convened for G1, and a Mint under the
	// bench — the successor as the replay meets it after that accept — asks for the accept it stands
	// on (requireBenchSuccessor).
	if _, err := Append(convene(t, run, "G1"), actRule("M3", true, "judge")); err != nil {
		t.Fatal(err)
	}
	_ = wouldReplay(t, run, "judge", recordpb.EventType_EVENT_TYPE_MINT, successorMint("G3", "G1"))
	if _, err := Append(Identity{Run: run, SeatID: "judge"}, &recordpb.Tombstone{GapId: proto.String("G1"),
		Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL.Enum()}); err != nil {
		t.Fatal(err)
	}
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
