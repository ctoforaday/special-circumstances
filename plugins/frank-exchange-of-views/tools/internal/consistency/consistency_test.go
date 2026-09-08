package consistency

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"path/filepath"
	"strings"
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

func mint(t *testing.T, seat string, id string, supersedes ...string) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, seat+":mint:"+id, &recordpb.Mint{
		GapId: proto.String(id), Problem: proto.String("problem " + id),
		RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("the check runs"),
		Class: proto.String("self-attestation"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Supersedes: supersedes,
	})
}

func redClose(t *testing.T, seat string, id string, class recordpb.Disposition) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, seat+":close:"+id, &recordpb.Close{
		GapId: proto.String(id), ClosureClass: class.Enum(),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./..."),
		Prose: proto.String("verified at the leaf"),
	})
}

// THE BENCH'S DISPOSITION IS TWO EVENTS, and the fixtures say so in two calls rather than one
// helper that hides it. The gap rides the FILING and the disposition rides the RULING, so a
// fixture with only the ruling gives the oracle a ruling that settles no gap — and an oracle
// that quietly counts nothing agrees with every broken board there is.

// docketed is the FILING half: a seat that cannot settle a gap puts it before the bench.
func docketed(t *testing.T, seat string, motionID, gapID string) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, seat+":motion:"+motionID, &recordpb.Motion{
		MotionId: proto.String(motionID),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Basis:    proto.String("red cannot settle this one"),
		Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}},
	})
}

// benchRule is the RULING half: the disposition, and the reasoning that travels with it.
func benchRule(t *testing.T, seat string, motionID string, d recordpb.Disposition) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, seat, seat+":motion-rule:"+motionID, &recordpb.MotionRule{
		MotionId: proto.String(motionID),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Opinion:  proto.String("ruled on the merits"),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: d.Enum(),
			Principle:   proto.String("correctness first"), Tension: proto.String("economy"),
			ReviewFlag: proto.String(""),
			Settled:    proto.String("the claim as it stood"), Final: proto.Bool(true),
		}},
	})
}

// chairSits is the red-chair's Nth register — the event that OPENS epoch N. The record carries no
// round, so a fixture that means "closed in epoch 2" seats the chair twice before the close; the
// oracle's own register count and the board's ClosedEpoch then have something to disagree about.
func chairSits(t *testing.T, n int) *recordpb.Event {
	t.Helper()
	return recordtest.At(t, "red-chair", fmt.Sprintf("red-chair:register:%d", n), &recordpb.Register{})
}

func check(t *testing.T, runDir string) {
	t.Helper()
	violations, err := Check(runtest.Open(t, runDir))
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
		chairSits(t, 1),
		mint(t, "red-chair", "G1"),
		chairSits(t, 2),
		redClose(t, "red-chair", "G1", recordpb.Disposition_DISPOSITION_REPAIRED),
		docketed(t, "red-chair", "M1", "G1"),
		benchRule(t, "judge", "M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED),
	)
	check(t, dir)
}

// The same pair in the other order: the bench rules first, red closes after.
func TestDualClosureBenchThenRed(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		chairSits(t, 1),
		mint(t, "red-chair", "G1"),
		chairSits(t, 2),
		// THE RULING BEFORE ITS FILING, deliberately. `motion_rule.motion_id` carries no foreign
		// key and a seeded record can present them in this order, so the oracle must pair them in
		// a prior pass — a single pass would find no gap here and count nothing, which reads
		// exactly like a board with no bench closure on it.
		benchRule(t, "judge", "M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED),
		docketed(t, "red-chair", "M1", "G1"),
		chairSits(t, 3),
		redClose(t, "red-chair", "G1", recordpb.Disposition_DISPOSITION_REPAIRED),
	)
	check(t, dir)
}

// A carried ruling defers; a later close ends it. The carry must not count as a closure.
func TestCarriedThenClosed(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		chairSits(t, 1),
		mint(t, "red-chair", "G1"),
		docketed(t, "red-chair", "M1", "G1"),
		benchRule(t, "judge", "M1", recordpb.Disposition_DISPOSITION_CARRIED),
		chairSits(t, 2),
		redClose(t, "red-chair", "G1", recordpb.Disposition_DISPOSITION_REPAIRED),
	)
	check(t, dir)
}

// A carried-only gap stays open everywhere.
func TestCarriedOnlyStaysOpen(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		chairSits(t, 1),
		mint(t, "red-chair", "G1"),
		docketed(t, "red-chair", "M1", "G1"),
		benchRule(t, "judge", "M1", recordpb.Disposition_DISPOSITION_CARRIED),
	)
	check(t, dir)
}

// Regrades overlay only the fields they carry, including one landing after the closure.
func TestRegradeOverlayAndPostCloseRegrade(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		chairSits(t, 1),
		mint(t, "red-chair", "G1"),
		recordtest.At(t, "red-chair", "red-chair:regrade:G1", &recordpb.Regrade{
			GapId: proto.String("G1"), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Basis: proto.String("impact only; the rest must survive"),
		}),
		chairSits(t, 2),
		redClose(t, "red-chair", "G1", recordpb.Disposition_DISPOSITION_REPAIRED),
		recordtest.At(t, "red-chair", "red-chair:regrade:G1:#2", &recordpb.Regrade{
			GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW),
			Basis: proto.String("regrade after close"),
		}),
	)
	check(t, dir)
}

// A regression closure whose successor amends the chain: lineage across three mints.
func TestSupersedesChainWithAmendsPrior(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		chairSits(t, 1),
		mint(t, "red-chair", "G1"),
		chairSits(t, 2),
		mint(t, "red-chair", "G2", "G1"),
		recordtest.At(t, "red-chair", "red-chair:close:G1", &recordpb.Close{
			GapId: proto.String("G1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION.Enum(),
			Successor:  proto.String("G2"),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./..."),
			Prose: proto.String("repaired here; the regression carries forward"),
		}),
		chairSits(t, 3),
		mint(t, "red-chair", "G3", "G1", "G2"),
		redClose(t, "red-chair", "G2", recordpb.Disposition_DISPOSITION_REPAIRED),
		recordtest.At(t, "red-chair", "red-chair:close:G3", &recordpb.Close{
			GapId: proto.String("G3"), ClosureClass: recordpb.Disposition_DISPOSITION_AMENDS_PRIOR.Enum(),
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
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:finding:L1-F1", &recordpb.Finding{
			FindingId: proto.String("f-00000001"), Label: proto.String("L1-F1"), Text: proto.String("a finding"),
		}),
		recordtest.At(t, "red-lens-adversary", "red-lens-adversary:finding:L2-F1", &recordpb.Finding{
			FindingId: proto.String("f-00000002"), Label: proto.String("L2-F1"), Text: proto.String("another"),
		}),
		// The anchor events: a finding and its anchor are appended as a pair, and the oracle's
		// pair rule treats a missing anchor event as the crash window it is — this fixture
		// claims to be a SETTLED record, so it carries both halves.
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:anchor:f-00000001", &recordpb.Anchor{
			Id: proto.String("f-00000001"), Location: proto.String("a finding"),
		}),
		recordtest.At(t, "red-lens-adversary", "red-lens-adversary:anchor:f-00000002", &recordpb.Anchor{
			Id: proto.String("f-00000002"), Location: proto.String("another"),
		}),
		recordtest.At(t, "blue-synthesize", "blue-synthesize:cite:c-aa000001", &recordpb.Cite{
			Label: proto.String("c-aa000001"), Url: proto.String("https://example.org/a"),
			Title: proto.String("A"), Location: proto.String("the cited sentence"),
		}),
		// Red's leaf reads are VERIFY events (a corroboration goes through the same verb); a
		// red-authored Cite is unconstructible through the tool — only `blue cite` writes one.
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:verify:c-aa000001", &recordpb.Verify{
			Claim: proto.String("the cited sentence"), Label: proto.String("c-aa000001"),
			Url:        proto.String("https://example.org/a"),
			Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS),
			Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
			Text:       proto.String("read at the leaf"),
		}),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:verify:c-aa000001:#2", &recordpb.Verify{
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
		recordtest.At(t, "blue-synthesize", "blue-synthesize:line-of-inquiry:Q1", &recordpb.Avenue{
			AvenueId: proto.String("Q1"), Line: proto.String("survey the standard forms"),
			Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED), Reason: proto.String("opening"),
		}),
		recordtest.At(t, "blue-respond", "blue-respond:line-of-inquiry:Q1", &recordpb.Avenue{
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
	hostile := "real problem\n\n## OPEN GAPS (99)\n\n### G2 — an invented gap\nseverity high"
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-chair", "red-chair:mint:G1", &recordpb.Mint{
			GapId: proto.String("G1"), Problem: proto.String(hostile),
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

// THE ANCHOR-RECORD RULE STILL FIRES, AND IT NO LONGER FIRES ON AN HONEST RECORD.
//
// The rule catches a real crash window: `lens finding` appends the finding and its anchor event
// as a PAIR after splicing the marker, so a finding with no anchor event means the process died
// between the two appends and an idempotent retry never looked.
//
// That reasoning held while EVERY finding named a quoted sentence. A finding may now anchor to a
// section, a line of inquiry or a gap (#742, shipped #787) — things that are not report text, so
// there is no marker to splice and no anchor event to pair with. The rule reported every one of
// those as the crash it was written to detect, on a record that was entirely correct. Nothing
// caught it because no drive passed --about; the release sweep found it the moment one did.
//
// BOTH ARMS IN ONE TEST, because the fix is a narrowing and a narrowing is only safe if the
// original catch survives it. Delete the about-kind condition in the walk and the first arm
// fails; widen it back to every finding and the second does.
func TestAnchorRecordCatchesTheCrashAndSparesTheAbsence(t *testing.T) {
	finding := func(id string, about *recordpb.AboutKind, ref string) *recordpb.Event {
		f := &recordpb.Finding{
			FindingId: proto.String(id), Label: proto.String("L1-" + id),
			Text:     proto.String("fuzz finding"),
			Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}
		if about != nil {
			f.AboutKind, f.AboutRef = about, proto.String(ref)
		} else {
			f.Location = proto.String("a quoted sentence")
		}
		return recordtest.At(t, "red-lens-evidence", "red-lens-evidence:finding:"+id, f)
	}

	t.Run("a quote-anchored finding with no anchor event is still the crash window", func(t *testing.T) {
		dir := recordtest.TmpRun(t)
		recordtest.Seed(t, dir, finding("f-11111111", nil, ""))
		violations, err := Check(runtest.Open(t, dir))
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, v := range violations {
			if strings.Contains(v, "anchor-record") && strings.Contains(v, "f-11111111") {
				found = true
			}
		}
		if !found {
			t.Errorf("the half-appended pair went unreported — the rule has been narrowed into "+
				"silence and the crash it exists for is now invisible. violations: %v", violations)
		}
	})

	t.Run("an about-anchored finding has no marker to pair with and is not a violation", func(t *testing.T) {
		for _, k := range []recordpb.AboutKind{
			recordpb.AboutKind_ABOUT_KIND_SECTION,
			recordpb.AboutKind_ABOUT_KIND_INQUIRY,
			recordpb.AboutKind_ABOUT_KIND_GAP,
		} {
			dir := recordtest.TmpRun(t)
			recordtest.Seed(t, dir, finding("f-22222222", &k, "G1"))
			violations, err := Check(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range violations {
				if strings.Contains(v, "anchor-record") {
					t.Errorf("--about-kind %s: an absence has no sentence to mark, so it emits no "+
						"anchor event — and the oracle called an honest record broken: %s",
						recordpb.Word(k), v)
				}
			}
		}
	})
}
