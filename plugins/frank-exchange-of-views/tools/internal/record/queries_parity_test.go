package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// EACH QUERY IS HELD AGAINST THE FOLD IT REPLACED, on one record written through the real write
// path — the test plans/record-sqlite.md step 4 demands per converted projection. The fixture
// deliberately exercises the folds' edges: a regrade overlaying a mint grade, a superseded
// ancestor left open, a computation gap with and without a proof, and a record that does not
// exist at all (whose every answer must be the honest zero, never an error).
func TestQueriesAgreeWithTheFoldsTheyReplaced(t *testing.T) {
	runDir := newRun(t)
	red := Identity{Run: mustRun(t, runDir), SeatID: "red-merge-r1", Round: 1}
	blue := Identity{Run: mustRun(t, runDir), SeatID: "blue-respond-r1", Round: 1}
	judge := Identity{Run: mustRun(t, runDir), SeatID: "judge-r1", Round: 1}

	mint := func(id string, kind recordpb.CheckKind, extra func(*recordpb.Mint)) {
		t.Helper()
		m := &recordpb.Mint{
			GapId:           proto.String(id),
			Class:           proto.String("self-attestation"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       kind.Enum(),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}
		if extra != nil {
			extra(m)
		}
		if _, err := Append(red, m); err != nil {
			t.Fatal(err)
		}
	}
	mint("R1-1", recordpb.CheckKind_CHECK_KIND_COMPUTATION, func(m *recordpb.Mint) {
		m.FixBasis = proto.String("verified")
		m.Location = proto.String("the sky is teal at noon")
		m.FixNew = proto.String("the sky is blue at noon")
	})
	mint("R1-2", recordpb.CheckKind_CHECK_KIND_DOCUMENT, func(m *recordpb.Mint) {
		m.Supersedes = []string{"R1-1"}
	})
	mint("R1-3", recordpb.CheckKind_CHECK_KIND_COMPUTATION, nil)

	// The regrade overlay: impact moves MEDIUM -> HIGH on R1-1; every other axis keeps the mint's.
	if _, err := Append(red, &recordpb.Regrade{GapId: proto.String("R1-1"),
		Impact: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("new evidence")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(blue, &recordpb.Proof{Answers: proto.String("R1-3"),
		Text: proto.String("7 is prime"), Script: proto.String("factor 7"), Exit: proto.Int32(0)}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(red, &recordpb.Close{GapId: proto.String("R1-3"),
		ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
		AnchorSeat:   proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"),
		Prose: proto.String("verified at the leaf")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(blue, &recordpb.Avenue{AvenueId: proto.String("Q1"),
		Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("a direction")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(blue, &recordpb.BlueEdit{Answers: proto.String("R1-2"),
		Old: proto.String("we claim the moon is cheese"), New: proto.String("retracted"), Text: proto.String("r")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(blue, &recordpb.Revision{Text: proto.String("round 1 edits")}); err != nil {
		t.Fatal(err)
	}

	run := mustRun(t, runDir)
	b, err := BoardState(run)
	if err != nil {
		t.Fatal(err)
	}

	// BoardCounts vs the fold's own tally.
	wantOpen, wantClosed := 0, 0
	for _, id := range b.GapOrder {
		if b.Gaps[id].Open {
			wantOpen++
		} else {
			wantClosed++
		}
	}
	open, closed, err := BoardCounts(run)
	if err != nil || open != wantOpen || closed != wantClosed {
		t.Errorf("BoardCounts = (%d, %d, %v), fold says (%d, %d)", open, closed, err, wantOpen, wantClosed)
	}

	// gapState against the board.
	for id, g := range b.Gaps {
		got, err := gapState(run, id)
		if err != nil || got != !g.Open {
			t.Errorf("gapState(%s) = (%v, %v), board says closed=%v", id, got, err, !g.Open)
		}
	}
	if got, err := gapState(run, "R9-99"); got || err != nil {
		t.Errorf("gapState on an unknown gap = (%v, %v), want (false, nil) — existence is requireGap's question", got, err)
	}

	// The grade overlay: proposing the CURRENT post-regrade impact must be refused as no change;
	// proposing a different one must pass. A pre-regrade reading would get both wrong.
	noChange := &recordpb.GradeMotion{GapId: proto.String("R1-1"),
		Dimension: recordpb.GradeDimension_GRADE_DIMENSION_IMPACT.Enum(),
		Proposed:  recordpb.Grade_GRADE_HIGH.Enum()}
	if err := requireGradeMotionAsksForAChange(run, noChange); err == nil {
		t.Error("a motion proposing the regraded current grade was not refused — the query missed the regrade overlay")
	}
	change := &recordpb.GradeMotion{GapId: proto.String("R1-1"),
		Dimension: recordpb.GradeDimension_GRADE_DIMENSION_IMPACT.Enum(),
		Proposed:  recordpb.Grade_GRADE_LOW.Enum()}
	if err := requireGradeMotionAsksForAChange(run, change); err != nil {
		t.Errorf("a motion proposing a different grade was refused: %v", err)
	}
	untouched := &recordpb.GradeMotion{GapId: proto.String("R1-1"),
		Dimension: recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD.Enum(),
		Proposed:  recordpb.Grade_GRADE_MEDIUM.Enum()}
	if err := requireGradeMotionAsksForAChange(run, untouched); err == nil {
		t.Error("a motion proposing the mint's grade on an axis no regrade touched was not refused")
	}

	// The stranded-ancestor join: R1-1 is superseded by R1-2 and still open.
	if err := requireSupersededAreClosed(run); err == nil || !strings.Contains(err.Error(), "R1-1 (superseded by R1-2)") {
		t.Errorf("requireSupersededAreClosed = %v, want the stranded pair named", err)
	}

	// requireInquiry: the proposed line resolves, an unknown one is refused.
	if err := requireInquiry(run, "Q1", "move", "--id"); err != nil {
		t.Errorf("requireInquiry(Q1) = %v", err)
	}
	if err := requireInquiry(run, "Q9", "move", "--id"); err == nil || !strings.Contains(err.Error(), "names no line of inquiry") {
		t.Errorf("requireInquiry on an unknown id = %v, want the refusal", err)
	}

	// Estoppel: the byte-exact pair, the proof join, the edit's old span.
	if ok, err := ProposalAppliedVerbatim(run, "R1-1", "the sky is teal at noon", "the sky is blue at noon"); !ok || err != nil {
		t.Errorf("ProposalAppliedVerbatim on the exact pair = (%v, %v)", ok, err)
	}
	if ok, _ := ProposalAppliedVerbatim(run, "R1-1", "the sky is teal at noon ", "the sky is blue at noon"); ok {
		t.Error("ProposalAppliedVerbatim matched a near-application — exactness is the whole point")
	}
	if !ProofAnswers(run, "R1-3") || ProofAnswers(run, "R1-1") || ProofAnswers(run, "") {
		t.Error("ProofAnswers disagrees with the recorded proof join")
	}
	if !ClaimAppearsInAnEdit(run, "moon is cheese") || ClaimAppearsInAnEdit(run, "never written") {
		t.Error("ClaimAppearsInAnEdit disagrees with the recorded old spans")
	}

	// GapsAwaitingProof: R1-1 is the open computation gap with no proof; R1-3's proof discharges
	// it and R1-2 is a document check.
	if got := GapsAwaitingProof(run); len(got) != 1 || got[0] != "R1-1" {
		t.Errorf("GapsAwaitingProof = %v, want [R1-1]", got)
	}

	if kind, err := MintCheckKind(run, "R1-2"); err != nil || kind != recordpb.CheckKind_CHECK_KIND_DOCUMENT {
		t.Errorf("MintCheckKind(R1-2) = (%v, %v)", kind, err)
	}
	if kind, err := MintCheckKind(run, "R9-99"); err != nil || kind != recordpb.CheckKind_CHECK_KIND_UNSPECIFIED {
		t.Errorf("MintCheckKind on an unknown gap = (%v, %v), want the unspecified zero", kind, err)
	}

	if n := RoundsWithRevision(run); n != 1 {
		t.Errorf("RoundsWithRevision = %d, want 1", n)
	}

	// The bench records the outcome; RecordedOutcome is that act and TerminalVerdict serves it.
	if v := RecordedOutcome(run); v != "" {
		t.Errorf("RecordedOutcome before any outcome = %q, want the honest empty", v)
	}
	if _, err := Append(judge, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED),
		Prose: proto.String("ended on safety grounds")}); err != nil {
		t.Fatal(err)
	}
	want := recordpb.Word(recordpb.RunOutcome_RUN_OUTCOME_HALTED)
	if v := RecordedOutcome(run); v != want {
		t.Errorf("RecordedOutcome = %q, want %q", v, want)
	}
	if v := TerminalVerdict(run); v != want {
		t.Errorf("TerminalVerdict = %q, want the bench's own act %q", v, want)
	}
}

// A RUN THAT HAS RECORDED NOTHING answers every question with its honest zero — the same answers
// the folds gave over an empty event stream, and never an error. openRunForRead's third state
// (a record this binary cannot read) stays an error and is recordroot's test to hold.
func TestQueriesAnswerTheHonestZeroOverNoRecord(t *testing.T) {
	run := mustRun(t, newRun(t))

	if open, closed, err := BoardCounts(run); open != 0 || closed != 0 || err != nil {
		t.Errorf("BoardCounts = (%d, %d, %v)", open, closed, err)
	}
	if v := RecordedOutcome(run); v != "" {
		t.Errorf("RecordedOutcome = %q", v)
	}
	if closed, err := gapState(run, "R1-1"); closed || err != nil {
		t.Errorf("gapState = (%v, %v)", closed, err)
	}
	if err := requireSupersededAreClosed(run); err != nil {
		t.Errorf("requireSupersededAreClosed = %v", err)
	}
	if got := GapsAwaitingProof(run); got != nil {
		t.Errorf("GapsAwaitingProof = %v", got)
	}
	if kind, err := MintCheckKind(run, "R1-1"); err != nil || kind != recordpb.CheckKind_CHECK_KIND_UNSPECIFIED {
		t.Errorf("MintCheckKind = (%v, %v)", kind, err)
	}
	if n := RoundsWithRevision(run); n != 0 {
		t.Errorf("RoundsWithRevision = %d", n)
	}
}
