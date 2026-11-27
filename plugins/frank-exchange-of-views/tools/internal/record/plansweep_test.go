package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
	_, _, _ = AnchorRemovedAt(run, "A1")
	_, _ = EvidenceDemandOf(run, "G1")
	if db, err := openRunForRead(run); err == nil && db != nil {
		_ = requireCitationUncited(db, []string{"C-1"}, "retire")
	}
	_, _ = ReportAtPart(run, "harness")
	_, _ = ReportHasParts(run)
	_, _, _ = PartStanding(run, "harness")
	_, _ = ExistingProposalByText(run, "blue-respond", "q", "h")
	// A proposed close and a tombstone, so the decode reads their tables.
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
	// Red's closing on the open gap a proposed close stands on states its text: here none, with
	// --no-text. Its refusal or landing is not this test's; it is issued so its statements are planned.
	_, _ = Append(Identity{Run: run, SeatID: evLens}, &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("the gap stands"),
		NoText: proto.Bool(true)})
	if db, err := openRunForRead(run); err == nil && db != nil {
		_, _, _ = closingTakesText(db, "G1")
	}
	// The tool dockets the proposal the assignee did not accept, each party's closing states what it
	// states, and the bench sides with the proposed close in the docket sitting convened for G1. The
	// closings' and the ruling's refusals are not this test's: each is issued so its statements are
	// planned.
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Docketing{DocketingId: proto.String("K1"), GapId: proto.String("G1"),
		Cause: recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL.Enum(), MotionId: proto.String("M2"), Basis: proto.String("the assignee did not accept it")})
	_, _ = Append(blue, &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("the proposed close states the defect moot")})
	_, _ = Append(Identity{Run: run, SeatID: evLens}, &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("the gap stands"),
		NoText: proto.Bool(true)})
	if _, err := Append(convene(t, run, "G1"), &recordpb.MotionRule{MotionId: proto.String("M2"), Opinion: proto.String("ruled"),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{DocketingId: proto.String("K1"),
			SidedWith: recordpb.SidedWith_SIDED_WITH_AUTHOR.Enum(), Disposition: recordpb.Disposition_DISPOSITION_MOOT.Enum(), Final: proto.Bool(true),
			Principle: proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"), Settled: proto.String("st")}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: run, SeatID: "judge"}, &recordpb.Tombstone{GapId: proto.String("G1"),
		Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL.Enum()}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: run, SeatID: "judge"}, &recordpb.Tombstone{GapId: proto.String("G1"),
		Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING.Enum(), DocketingId: proto.String("K1")}); err != nil {
		t.Fatal(err)
	}
	for _, seat := range []string{evLens, "red-chair", "judge"} {
		_, _ = CutDelivery(run, seat, []string{evLens}, func(string, string) bool { return true })
	}
	if db, err := openRunForRead(run); err == nil && db != nil {
		_, _, _ = DeliveryThisSitting(db, evLens)
		_, _ = NextDeliveryWhole(db, evLens)
		_, _ = PinAt(db, evLens)
		_ = requireDeliveredText(db, evLens, "a close", []string{"G1"})
		_ = requireReportRead(db, evLens, &recordpb.ReportRead{Head: proto.Int64(1)})
	}
	// The release fuzz's gap-state oracle (releasegate/fuzz/anchororacle_test.go); its rows are
	// asserted by recordsql's gap-state tests.
	if db, err := openRun(run); err == nil && db != nil {
		_, _ = recordsql.GapStateRefreshDiff(db)
	}
	_, _ = MergedEvents(run) // decodes every arm some row names, the position motion's included
	// The readers of the standing rows: the proposal arm, the class arm's held set and the work list's
	// protected gaps.
	if db, err := openRunForRead(run); err == nil && db != nil {
		_ = requireNotEstoppedOn(db, "G1")
		_, _ = withRuledGapsOfClassAt(db, nil, recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION)
		_, _ = rulingsProtectingAll(db)
	}

	// Whether red's closing takes text where a proposal stands on an open gap about no inquiry or
	// section reads the proposal's act.
	sentence := retBoard(t).sit(2, evLens).add(evLens, &recordpb.Mint{GapId: proto.String("G9"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Problem: proto.String("p"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}).
		add("blue-respond", &recordpb.Motion{MotionId: proto.String("M9"), Basis: proto.String("b"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G9")}},
			Act: &recordpb.Motion_Close{Close: &recordpb.Close{GapId: proto.String("G9"), ClosureClass: recordpb.Disposition_DISPOSITION_MOOT.Enum(),
				Prose: proto.String("moot"), AnchorSeat: proto.String("blue-respond"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("s")}}}).seed()
	if db, err := openRunForRead(sentence); err == nil && db != nil {
		if takes, _, err := closingTakesText(db, "G9"); err != nil || !takes {
			t.Errorf("red's closing on G9, a proposed close standing on it, takes text: %v (%v)", takes, err)
		}
	}

	_, _ = AppendSeatTurns(run, "agent-1", []seatturn.Turn{{Index: 0, Model: "m"}})
	_, _ = SeatTimingOf(run)
	_, _ = SeatMetrics(run)

	admitted := &recordpb.EnvelopeAdmitted{AgentId: proto.String("agent-1")}
	_, _ = Append(Identity{Run: run, SeatID: "harness"}, admitted, OneAdmissionASitting(admitted))
	_, _ = Append(Identity{Run: run, SeatID: "harness"}, &recordpb.Log{Text: proto.String("t")}, Check(func(l *Landing) error {
		_, _ = l.AboutAnchor("G1")
		return ErrWritesNothing
	}))
}
