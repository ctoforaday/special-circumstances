package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE RUN-SHAPED BOARD MUST ANSWER BYTE-IDENTICALLY TO THE FOLD-SHAPED ONE — the wave-1b
// contract (plans/board-as-views.md §II.3), held on the fold's edges: a gap closed by BOTH arms
// (attribution follows the LAST closing event while the embedded body prefers the red close), a
// non-closing docket ruling that must close nothing, regrade overlays, credited and uncredited
// findings, and the verify/cite count split. This test retires with the fold shape in wave 7.
func TestBoardJSONHoldsTheFoldsEdges(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-chair-r1", Round: 1}
	blue := Identity{Run: run, SeatID: "blue-respond-r1", Round: 1}
	lens := Identity{Run: run, SeatID: "red-lens-r1-evidence", Round: 1}
	judge2 := Identity{Run: run, SeatID: "judge-r2", Round: 2}
	app := func(id Identity, body proto.Message) {
		t.Helper()
		if _, err := Append(id, body); err != nil {
			t.Fatal(err)
		}
	}

	app(lens, &recordpb.Finding{FindingId: proto.String("f-0a0a0a0a"), Label: proto.String("L1-F1"),
		Location: proto.String("¶3"), Text: proto.String("overclaims"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH)})
	app(lens, &recordpb.Finding{FindingId: proto.String("f-0b0b0b0b"), Label: proto.String("L1-F2"),
		Text: proto.String("uncredited")})

	mint := func(id string, extra func(m *recordpb.Mint)) {
		m := &recordpb.Mint{GapId: proto.String(id), Class: proto.String("self-attestation"),
			Problem: proto.String("p " + id), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
			CheckKind:  recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
		if extra != nil {
			extra(m)
		}
		app(red, m)
	}
	mint("R1-1", func(m *recordpb.Mint) {
		m.FoundBy = []string{"L1-F1"}
		m.MintReason = proto.String("the argument")
		m.Location = proto.String("the claimed span")
		m.FixBasis = proto.String("proposed")
		m.FixNew = proto.String("the fix")
		m.CheckKind = recordpb.CheckKind_CHECK_KIND_COMPUTATION.Enum()
	})
	mint("R1-2", func(m *recordpb.Mint) { m.Supersedes = []string{"R1-1"} })
	mint("R1-3", nil)

	// A regrade the projection must both OVERLAY (grades) and EMBED (the regrades list).
	app(red, &recordpb.Regrade{GapId: proto.String("R1-3"),
		Impact: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("moved")})
	// R1-3: closed by red in round 1, then ruled by the bench in round 2 — BOTH arms.
	// Attribution must follow the bench (last), the embedded body must stay red's (closureBody).
	app(red, &recordpb.Close{GapId: proto.String("R1-3"),
		ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
		AnchorSeat:   proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"),
		Prose: proto.String("verified at the leaf")})
	// THE BENCH'S DISPOSITION IS A DOCKET MOTION'S RULING (#681 Scope 2), so each of these is a
	// PAIR: red files the gap to the bench, the bench rules. The gap rides the FILING, which is
	// why both halves have to be on the record for either projection to attribute the closure.
	docket := func(motionID, gapID string, rule *recordpb.DocketRuling) {
		t.Helper()
		app(red, &recordpb.Motion{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:   proto.String("red cannot settle " + gapID),
			Filing:  &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}}})
		app(judge2, &recordpb.MotionRule{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String("ra"),
			Ruling:  &recordpb.MotionRule_Docket{Docket: rule}})
	}
	docket("M1", "R1-3", &recordpb.DocketRuling{
		Disposition: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Settled: proto.String("st"), Final: proto.Bool(true)})
	// A CARRIED ruling on an open gap closes NOTHING — the vocabulary's own facet decides.
	docket("M2", "R1-2", &recordpb.DocketRuling{
		Disposition: recordpb.Disposition_DISPOSITION_CARRIED.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Settled: proto.String("st"), ReopensOn: proto.String("new evidence")})

	// The count split: one verify (red's read), one cite (blue's authoring).
	app(red, &recordpb.Verify{Url: proto.String("https://example.org"), Claim: proto.String("c"),
		Label: proto.String("v-1"), Title: proto.String("e"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("the source states it plainly")})
	app(blue, &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("https://example.org"),
		Title: proto.String("t"), CiteKey: proto.String("k1")})

	bj, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]GapJSON{}
	for _, g := range append(append([]GapJSON{}, bj.Open...), bj.Closed...) {
		byID[g.ID] = g
	}
	// R1-3: closed by red (r1), then ruled by the bench (r2). Attribution follows the bench;
	// the embedded body stays red's close (its prose proves which body rendered).
	g3 := byID["R1-3"]
	if g3.Open || !g3.ClosedByBench || g3.ClosedRound != 2 {
		t.Errorf("R1-3 attribution = open=%v bench=%v round=%d, want closed/bench/2", g3.Open, g3.ClosedByBench, g3.ClosedRound)
	}
	if g3.Closure == nil || g3.Closure["prose"] != "verified at the leaf" {
		t.Errorf("R1-3 embedded closure = %v, want red's close body (closureBody's precedence)", g3.Closure)
	}
	if len(g3.Regrades) != 1 {
		t.Errorf("R1-3 regrades = %v, want the one recorded regrade embedded", g3.Regrades)
	}
	if g3.Impact != "high" {
		t.Errorf("R1-3 impact = %v, want the regrade overlay", g3.Impact)
	}
	// R1-2: a carried docket ruling closes nothing.
	if g2 := byID["R1-2"]; !g2.Open {
		t.Error("a carried ruling closed R1-2 — the vocabulary's own facet says it must not")
	}
	// R1-1 (computation, unproved) and R1-2 (carried) stay open; R1-3 closed by both arms.
	if bj.Counts.Open != 2 || bj.Counts.Closed != 1 || bj.Counts.ClosedByBench != 1 {
		t.Errorf("counts = %+v", bj.Counts)
	}
	if bj.Counts.Citations != 1 || bj.Counts.CitationsAuthored != 1 {
		t.Errorf("citation split = %d/%d, want 1/1", bj.Counts.Citations, bj.Counts.CitationsAuthored)
	}
	credited, uncredited := 0, 0
	for _, o := range bj.Observations {
		if o.Credited {
			credited++
		} else {
			uncredited++
		}
	}
	if credited != 1 || uncredited != 1 || bj.Counts.UncreditedFindings != 1 {
		t.Errorf("credit split = %d/%d (counter %d), want 1/1/1", credited, uncredited, bj.Counts.UncreditedFindings)
	}
}
