package consistency

import (
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// Every seed below is a TOOL-REACHABLE state, seeded directly for speed: the store's own
// constraints still apply (foreign keys, checks, uniqueness), and each scenario is one the
// engine's own sequencing can produce — the dual-closure pair is literally the measured
// 2026-08-22 shape, where red and the bench both acted on one gap in one sitting.
//
// The assertion is ZERO violations. A failure here is not a broken test; it is two readers of one
// record disagreeing, which is the crop this oracle exists to harvest.

func mint(t *testing.T, seat string, round int, id string, supersedes ...string) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, round, seat+":mint:"+id, &recordpb.Mint{
		GapId: proto.String(id), Problem: proto.String("problem " + id),
		RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("the check runs"),
		Class: proto.String("self-attestation"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Supersedes: supersedes,
	})
}

func redClose(t *testing.T, seat string, round int, id string, class recordpb.Disposition) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, round, seat+":close:"+id, &recordpb.Close{
		GapId: proto.String(id), ClosureClass: class.Enum(),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./..."),
		Prose: proto.String("verified at the leaf"),
	})
}

func opinion(t *testing.T, seat string, round int, id string, d recordpb.Disposition) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, round, seat+":opinion:"+id, &recordpb.Opinion{
		GapId: proto.String(id), Disposition: d.Enum(),
		Principle: proto.String("correctness first"), Tension: proto.String("economy"),
		ReviewFlag: proto.String(""), Rationale: proto.String("ruled on the merits"),
		Settled: proto.String("the claim as it stood"), Final: proto.Bool(true),
	})
}

func check(t *testing.T, runDir string) {
	t.Helper()
	violations, err := Check(runDir)
	if err != nil {
		t.Fatalf("oracle: %v", err)
	}
	for _, v := range violations {
		t.Errorf("%s", v)
	}
}

// THE MEASURED SHAPE: red closes, the bench rules the same gap in the same sitting.
func TestDualClosureRedThenBench(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		mint(t, "red-merge-r1", 1, "R1-1"),
		redClose(t, "red-merge-r2", 2, "R1-1", recordpb.Disposition_DISPOSITION_REPAIRED),
		opinion(t, "judge-r2", 2, "R1-1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED),
	)
	check(t, dir)
}

// The same pair in the other order: the bench rules first, red closes after.
func TestDualClosureBenchThenRed(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		mint(t, "red-merge-r1", 1, "R1-1"),
		opinion(t, "judge-r2", 2, "R1-1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED),
		redClose(t, "red-merge-r3", 3, "R1-1", recordpb.Disposition_DISPOSITION_REPAIRED),
	)
	check(t, dir)
}

// A carried ruling defers; a later close ends it. The carry must not count as a closure.
func TestCarriedThenClosed(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		mint(t, "red-merge-r1", 1, "R1-1"),
		opinion(t, "judge-r1", 1, "R1-1", recordpb.Disposition_DISPOSITION_CARRIED),
		redClose(t, "red-merge-r2", 2, "R1-1", recordpb.Disposition_DISPOSITION_REPAIRED),
	)
	check(t, dir)
}

// A carried-only gap stays open everywhere.
func TestCarriedOnlyStaysOpen(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		mint(t, "red-merge-r1", 1, "R1-1"),
		opinion(t, "judge-r1", 1, "R1-1", recordpb.Disposition_DISPOSITION_CARRIED),
	)
	check(t, dir)
}

// Regrades overlay only the fields they carry, including one landing after the closure.
func TestRegradeOverlayAndPostCloseRegrade(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		mint(t, "red-merge-r1", 1, "R1-1"),
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:regrade:R1-1", &recordpb.Regrade{
			GapId: proto.String("R1-1"), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Basis: proto.String("impact only; the rest must survive"),
		}),
		redClose(t, "red-merge-r2", 2, "R1-1", recordpb.Disposition_DISPOSITION_REPAIRED),
		recordtest.At(t, "red-merge-r2", 2, "red-merge-r2:regrade:R1-1", &recordpb.Regrade{
			GapId: proto.String("R1-1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW),
			Basis: proto.String("regrade after close"),
		}),
	)
	check(t, dir)
}

// A regression closure whose successor amends the chain: lineage across three mints.
func TestSupersedesChainWithAmendsPrior(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		mint(t, "red-merge-r1", 1, "R1-1"),
		mint(t, "red-merge-r2", 2, "R2-1", "R1-1"),
		recordtest.At(t, "red-merge-r2", 2, "red-merge-r2:close:R1-1", &recordpb.Close{
			GapId: proto.String("R1-1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION.Enum(),
			Successor:  proto.String("R2-1"),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./..."),
			Prose: proto.String("repaired here; the regression carries forward"),
		}),
		mint(t, "red-merge-r3", 3, "R3-1", "R1-1", "R2-1"),
		redClose(t, "red-merge-r3", 3, "R2-1", recordpb.Disposition_DISPOSITION_REPAIRED),
		recordtest.At(t, "red-merge-r3", 3, "red-merge-r3:close:R3-1", &recordpb.Close{
			GapId: proto.String("R3-1"), ClosureClass: recordpb.Disposition_DISPOSITION_AMENDS_PRIOR.Enum(),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./..."),
			Prose: proto.String("a defect between two clean repairs"),
		}),
	)
	check(t, dir)
}

// Findings and citations: label bijections and the citation count.
func TestFindingsAndCitations(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-lens-r1-L1", 1, "red-lens-r1-L1:finding:L1-F1", &recordpb.Finding{
			FindingId: proto.String("f-00000001"), Label: proto.String("L1-F1"), Text: proto.String("a finding"),
		}),
		recordtest.At(t, "red-lens-r1-L2", 1, "red-lens-r1-L2:finding:L2-F1", &recordpb.Finding{
			FindingId: proto.String("f-00000002"), Label: proto.String("L2-F1"), Text: proto.String("another"),
		}),
		// The anchor events: a finding and its anchor are appended as a pair, and the oracle's
		// pair rule treats a missing anchor event as the crash window it is — this fixture
		// claims to be a SETTLED record, so it carries both halves.
		recordtest.At(t, "red-lens-r1-L1", 1, "red-lens-r1-L1:anchor:f-00000001", &recordpb.Anchor{
			Id: proto.String("f-00000001"), Location: proto.String("a finding"),
		}),
		recordtest.At(t, "red-lens-r1-L2", 1, "red-lens-r1-L2:anchor:f-00000002", &recordpb.Anchor{
			Id: proto.String("f-00000002"), Location: proto.String("another"),
		}),
		recordtest.At(t, "blue-synthesize", 0, "blue-synthesize:cite:c-aa000001", &recordpb.Cite{
			Label: proto.String("c-aa000001"), Url: proto.String("https://example.org/a"),
			Title: proto.String("A"), Location: proto.String("the cited sentence"),
		}),
		// Red's leaf reads are VERIFY events (a corroboration goes through the same verb); a
		// red-authored Cite is unconstructible through the tool — only `blue cite` writes one.
		recordtest.At(t, "red-lens-r1-L1", 1, "red-lens-r1-L1:verify:c-aa000001", &recordpb.Verify{
			Claim: proto.String("the cited sentence"), Label: proto.String("c-aa000001"),
			Url:        proto.String("https://example.org/a"),
			Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
			Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
			Text:       proto.String("read at the leaf"),
		}),
		recordtest.At(t, "red-lens-r2-L1", 2, "red-lens-r2-L1:verify:c-aa000001", &recordpb.Verify{
			Claim: proto.String("the cited sentence"), Label: proto.String("c-aa000001"),
			Url:        proto.String("https://example.org/a"),
			Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
			Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
			Text:       proto.String("re-read in round 2 — counts a second time, which the doc now admits"),
		}),
	)
	check(t, dir)
}

// The avenue lifecycle: proposed, then moved, and the projection follows the LAST status.
func TestAvenueLifecycle(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, "blue-synthesize", 0, "blue-synthesize:line-of-inquiry:Q1", &recordpb.Avenue{
			AvenueId: proto.String("Q1"), Line: proto.String("survey the standard forms"),
			Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Reason: proto.String("opening"),
		}),
		recordtest.At(t, "blue-respond-r1", 1, "blue-respond-r1:line-of-inquiry:Q1", &recordpb.Avenue{
			AvenueId: proto.String("Q1"), Line: proto.String("survey the standard forms"),
			Status:           recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED),
			SupersedesStatus: proto.String("proposed"), Reason: proto.String("nothing standard exists"),
		}),
	)
	check(t, dir)
}

// Prose is hostile by default: a problem text that carries the ledger's own markup must not
// derail any projection or any parser of one.
func TestMarkdownInjectionInProblemText(t *testing.T) {
	dir := recordtest.TmpRun(t)
	hostile := "real problem\n\n## OPEN GAPS (99)\n\n### R9-9 — an invented gap\nseverity high"
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
			GapId: proto.String("R1-1"), Problem: proto.String(hostile),
			RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("the check runs"),
			Class: proto.String("self-attestation"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
	)
	check(t, dir)
	_ = filepath.Join // keep the import while scenarios grow
}
