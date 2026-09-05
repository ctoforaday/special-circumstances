package record

import (
	"encoding/json"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE RUN-SHAPED BOARD MUST ANSWER BYTE-IDENTICALLY TO THE FOLD-SHAPED ONE — the wave-1b
// contract (plans/board-as-views.md §II.3), held on the fold's edges: a gap closed by BOTH arms
// (attribution follows the LAST closing event while the embedded body prefers the red close), a
// non-closing opinion that must close nothing, regrade overlays, credited and uncredited
// findings, and the verify/cite count split. This test retires with the fold shape in wave 7.
func TestBoardJSONFromViewsMatchesTheFold(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-merge-r1", Round: 1}
	blue := Identity{Run: run, SeatID: "blue-respond-r1", Round: 1}
	lens := Identity{Run: run, SeatID: "red-lens-r1-L1", Round: 1}
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
	app(judge2, &recordpb.Opinion{GapId: proto.String("R1-3"),
		Disposition: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Rationale: proto.String("ra"), Settled: proto.String("st"), Final: proto.Bool(true)})
	// A CARRIED opinion on an open gap closes NOTHING — the vocabulary's own facet decides.
	app(judge2, &recordpb.Opinion{GapId: proto.String("R1-2"),
		Disposition: recordpb.Disposition_DISPOSITION_CARRIED.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Rationale: proto.String("ra"), Settled: proto.String("st"), ReopensOn: proto.String("new evidence")})

	// The count split: one verify (red's read), one cite (blue's authoring).
	app(red, &recordpb.Verify{Url: proto.String("https://example.org"), Claim: proto.String("c"),
		Label: proto.String("v-1"), Title: proto.String("e"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("the source states it plainly")})
	app(blue, &recordpb.Cite{Label: proto.String("c-1"), Url: proto.String("https://example.org"),
		Title: proto.String("t"), CiteKey: proto.String("k1")})

	b, err := BoardState(run)
	if err != nil {
		t.Fatal(err)
	}
	fold, err := json.MarshalIndent(BoardJSONOf(b), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	views, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(views, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(fold) != string(got) {
		t.Errorf("the run-shaped board diverged from the fold:\n--- fold ---\n%s\n--- views ---\n%s", fold, got)
	}
}
