package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// demandMint is a gap about the Findings section at check kind kind, superseding ancestors.
func demandMint(id string, kind recordpb.CheckKind, ancestors ...string) *recordpb.Mint {
	return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"),
		GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
		RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("the check runs"), CheckKind: kind.Enum(),
		Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Supersedes: ancestors}
}

// supersededBy is the successor the board names as having closed gap id, and whether id reads open.
func supersededBy(t *testing.T, run Run, id string) (string, bool) {
	t.Helper()
	b := mustBoardJSONT(t, run)
	for _, g := range b.Open {
		if g.ID == id {
			return "", true
		}
	}
	for _, g := range b.Closed {
		if g.ID == id {
			return g.SupersededBy, false
		}
	}
	t.Fatalf("gap %s is not on the board", id)
	return "", false
}

// THE DEMAND MOVES TO THE SUCCESSOR (ruled by gblock 2026-11-05). A mint superseding an open `source` or
// `computation` gap that nothing answers is admitted at the same check kind — the ancestor closes and
// the successor awaits the live answer — and refused at any other, naming the demand. An ancestor whose
// demand is met takes a successor of any kind; a closed one is lineage only; the replay writes an
// archived supersede of another kind as it was recorded.
func TestASuccessorCarriesTheLiveAnswerDemandOfTheGapItSupersedes(t *testing.T) {
	run := corrRun(t)
	chair := sit(t, run, "red-chair")
	blue := sit(t, run, "blue-respond")
	mustAppend(t, chair, demandMint("G1", recordpb.CheckKind_CHECK_KIND_SOURCE))
	mustAppend(t, chair, demandMint("G2", recordpb.CheckKind_CHECK_KIND_COMPUTATION))

	for _, c := range []struct {
		id, ancestor, ancestorKind string
		kind                       recordpb.CheckKind
	}{
		{"G3", "G1", "source", recordpb.CheckKind_CHECK_KIND_DOCUMENT},
		{"G3", "G1", "source", recordpb.CheckKind_CHECK_KIND_COMPUTATION},
		{"G3", "G2", "computation", recordpb.CheckKind_CHECK_KIND_SOURCE},
	} {
		_, err := Append(chair, demandMint(c.id, c.kind, c.ancestor))
		if err == nil || feov.CodeOf(err) != string(feov.Conflict) ||
			!strings.Contains(err.Error(), "--supersedes names "+c.ancestor+", an open gap minted --check-kind "+c.ancestorKind) ||
			!strings.Contains(err.Error(), recordpb.DemandClause) || !strings.Contains(err.Error(), "Mint the successor --check-kind "+c.ancestorKind) {
			t.Fatalf("a %s successor of unanswered %s gap %s = %v, want refused naming the demand and the check kind it carries", recordpb.Word(c.kind), c.ancestorKind, c.ancestor, err)
		}
		if _, open := supersededBy(t, run, c.ancestor); !open {
			t.Fatalf("%s closed under a refused mint", c.ancestor)
		}
	}

	// THE SAME KIND IS ADMITTED, and the demand lives on the successor.
	mustAppend(t, chair, demandMint("G3", recordpb.CheckKind_CHECK_KIND_SOURCE, "G1"))
	if by, open := supersededBy(t, run, "G1"); open || by != "G3" {
		t.Fatalf("G1 open=%v superseded by %q, want closed by G3's mint", open, by)
	}
	if owed := awaitingEvidenceT(t, run); len(owed) != 2 || owed[0] != "G2" || owed[1] != "G3" {
		t.Fatalf("awaiting evidence = %v, want [G2 G3]: the demand moved to the successor", owed)
	}

	// A MET DEMAND IS UNAFFECTED: blue's computation answers G2, and a document successor is admitted.
	mustAppend(t, blue, &recordpb.Computation{Argv: []string{"python3", "script.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(),
		ComputationId: proto.String("K-0000c002"), Answers: proto.String("G2"), BundleSha: proto.String("sha-s")})
	mustAppend(t, chair, demandMint("G4", recordpb.CheckKind_CHECK_KIND_DOCUMENT, "G2"))
	if by, open := supersededBy(t, run, "G2"); open || by != "G4" {
		t.Fatalf("G2 open=%v superseded by %q, want closed by G4's mint", open, by)
	}
	// A CLOSED ANCESTOR IS LINEAGE ONLY: G1 was superseded unanswered, and naming it again closes nothing.
	mustAppend(t, chair, demandMint("G5", recordpb.CheckKind_CHECK_KIND_DOCUMENT, "G1"))

	// THE REPLAY ASKS NO ACT JUDGEMENT: an archived document successor of unanswered G3 is written as recorded.
	b := mustBegin(t, run)
	if _, err := b.Write(chair, demandMint("G6", recordpb.CheckKind_CHECK_KIND_DOCUMENT, "G3"), replayAt); err != nil {
		t.Fatalf("the replay refused an archived supersede of another check kind: %v", err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	if by, open := supersededBy(t, run, "G3"); open || by != "G6" {
		t.Fatalf("G3 open=%v superseded by %q after the replay, want closed by G6's archived mint", open, by)
	}
}
