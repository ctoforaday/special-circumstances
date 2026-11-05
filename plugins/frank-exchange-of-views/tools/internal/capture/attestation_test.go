package capture

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE PRODUCTION FAILURE, PINNED. Measured 2026-08-22 on research/2026-08-22_is-7-prime: two
// closures anchored `report text at line 103` and `report text at line 62` were reported as
// claimed acts with no matching tool call — under an audit whose own message says "This rules on
// whether the RECORD IS HONEST". Neither target has a fragment longer than six characters
// (`report` is exactly six), so the prose needle was "" and the closures were filed as findings.
// The third sampled closure reconciled only because its target happened to contain the ten-letter
// word `projection`. Honesty was decided by word length; the audit now joins on anchor ids alone.
//
// AttestationAudit had no test of any kind, which is how it shipped.

// attestGap is the one gap attestRun's board holds; attestOther is another gap a closure of it can
// name as an anchor; attestInquiry is an inquiry it can rest on.
const (
	attestGap     = "G-0dd40334"
	attestOther   = "G-5e10a3c2"
	attestInquiry = "Q-1a2b3c4d"
	// closerAgent is the agent whose sitting records the closure.
	closerAgent = "a0c105e5"
)

// The stamps the fixtures place: the closer's sitting opens at 00:00:00 and records its close at
// 00:00:10, inside the window of the call that made it.
const (
	sittingOpens = "2026-01-01T00:00:00.000000000Z"
	closeLands   = "2026-01-01T00:00:10.000000000Z"
)

// attestRun writes a run whose board holds two gaps, the first closed by closerAgent's sitting with
// the given verification and anchors.
//
// SEEDED THROUGH THE STORE, not by hand-writing shard files. A fixture that builds the record by a
// route production cannot take is testing a shape rather than the system.
func attestRun(t *testing.T, anchorTool, anchorTarget string, anchors ...string) string {
	t.Helper()
	return attestRunAt(t, closeLands, anchorTool, anchorTarget, anchors...)
}

// attestRunAt is attestRun with the close stamped closeTS.
func attestRunAt(t *testing.T, closeTS, anchorTool, anchorTarget string, anchors ...string) string {
	t.Helper()
	dir := t.TempDir()
	mint := func(id string) *recordpb.Mint {
		return &recordpb.Mint{
			GapId:           proto.String(id),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			Class:           recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}
	}
	recordtest.Seed(t, dir,
		recordtest.Stamped(recordtest.At(t, "red-lens-logic", "red-lens-logic:register", &recordpb.Register{AgentId: proto.String(closerAgent)}), sittingOpens),
		recordtest.Stamped(recordtest.At(t, "red-lens-logic", "red-lens-logic:mint:G1", mint(attestGap)), "2026-01-01T00:00:01.000000000Z"),
		recordtest.Stamped(recordtest.At(t, "red-lens-logic", "red-lens-logic:mint:G2", mint(attestOther)), "2026-01-01T00:00:02.000000000Z"),
		recordtest.Stamped(recordtest.At(t, "red-lens-logic", "red-lens-logic:close:G1", &recordpb.Close{
			GapId:        proto.String(attestGap),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   proto.String("red-lens-logic"),
			AnchorTool:   proto.String(anchorTool),
			AnchorTarget: proto.String(anchorTarget),
			Anchors:      anchors,
			Prose:        proto.String("verified at the leaf"),
		}), closeTS),
	)
	return dir
}

// call is one tool call a fixture transcript carries: its command, what came back, and the window
// from the call's line to its result's.
type call struct{ command, result, from, to string }

// read is a call the closer's sitting made before its close.
func read(command, result string) call {
	return call{command, result, "2026-01-01T00:00:03.000Z", "2026-01-01T00:00:04.000Z"}
}

// closing is the call that recorded the close: its window holds closeLands.
func closing(command string) call {
	return call{command, "closed " + attestGap + " (repaired)", "2026-01-01T00:00:09.500Z", "2026-01-01T00:00:10.500Z"}
}

// transcriptOf writes agent's transcript carrying the calls, as the harness writes one: a tool_use
// line and its tool_result line, each stamped, joined by the call's id.
func transcriptOf(t *testing.T, agent string, calls ...call) record.Transcript {
	t.Helper()
	var b strings.Builder
	for i, c := range calls {
		id := fmt.Sprintf("toolu_%s_%02d", agent, i)
		b.WriteString(`{"type":"assistant","timestamp":"` + c.from + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"` + id + `","name":"Bash","input":{"command":` +
			jsStringify(c.command) + `}}]}}` + "\n")
		b.WriteString(`{"type":"user","timestamp":"` + c.to + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"` + id + `","content":` +
			jsStringify(c.result) + `}]}}` + "\n")
	}
	p := filepath.Join(t.TempDir(), "trajectory-of-"+agent+".jsonl")
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return record.Transcript{AgentID: agent, Path: p}
}

// closeCommand is the close as the seat types it, every anchor named in it.
func closeCommand(anchors ...string) string {
	return "feov-record close --id " + attestGap + " --anchor " + strings.Join(anchors, ",") + " --verified-against \"report text at line 103\""
}

// A PRECISE PROSE TARGET IS NOT A VAGUE ONE. The seat did the work, anchored it to an exact line,
// and named the anchor it read. That must reconcile.
func TestAPreciseAnchorWithShortWordsReconcilesByItsID(t *testing.T) {
	run := attestRun(t, "show report --anchor", "report text at line 103", attestOther)
	trs := []record.Transcript{transcriptOf(t, closerAgent,
		read(`feov-record show report --anchor `+attestOther, "the sentence"), closing(closeCommand(attestOther)))}
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict != "PASS" {
		t.Errorf("a closure citing an id the sitting read must reconcile.\ngot %s: %s", got.Verdict, got.Detail)
	}
	// RECONCILED, NOT UNMEASURED: an unmeasured closure reports the same PASS with 0/1 behind it.
	if !strings.HasPrefix(got.Detail, "1/1 anchored closure(s) sampled and reconciled") || strings.Contains(got.Detail, "NOT BE MEASURED") {
		t.Errorf("the closure must reconcile BY ITS ID, not go unmeasured:\n%s", got.Detail)
	}
}

// A READ IS WHAT CAME BACK, not only what was typed: a `show report` names no anchor in its command
// and returns every anchor of the sentences it prints. m30's voice lens read its two computations that way.
func TestAnAnchorTheSittingReadInAResultReconciles(t *testing.T) {
	run := attestRun(t, "show report", "report text at line 103", "K-24a6af66")
	trs := []record.Transcript{transcriptOf(t, closerAgent,
		read(`feov-record show report --json`, `91 = 7 x 13 <!--computation:K-24a6af66-->`), closing(closeCommand("K-24a6af66")))}
	if got := AttestationAudit(runtest.Open(t, run), trs, 3); got.Verdict != "PASS" {
		t.Errorf("an anchor the sitting's own call returned must reconcile; got %s: %s", got.Verdict, got.Detail)
	}
}

// THE CLOSE IS NOT ITS OWN EVIDENCE. Every anchored closure names its anchors in the close that
// records it, and the audit searched every call — so it could not fail. m30's G-968365ad named two
// computations this way and reconciled on its own `close --anchor`.
func TestAnAnchorOnlyTheCloseNamesIsAFinding(t *testing.T) {
	run := attestRun(t, "show report", "report text at line 103", "K-24a6af66")
	trs := []record.Transcript{transcriptOf(t, closerAgent,
		read(`feov-record show board`, "G-0dd40334 open"), closing(closeCommand("K-24a6af66")))}
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("an anchor no call but the close carries must be a finding; got %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "K-24a6af66") || !strings.Contains(got.Detail, "other than its closes") {
		t.Errorf("the finding must name the anchor and say the close was not counted:\n%s", got.Detail)
	}
}

// THIS SITTING, NOT THE RUN: another seat's read is not the closer's.
func TestAnAnchorAnotherSittingReadIsAFinding(t *testing.T) {
	run := attestRun(t, "show report --anchor", "report text at line 103", attestOther)
	trs := []record.Transcript{
		transcriptOf(t, "someone-else", read(`feov-record show report --anchor `+attestOther, "the sentence")),
		transcriptOf(t, closerAgent, closing(closeCommand(attestOther))),
	}
	if got := AttestationAudit(runtest.Open(t, run), trs, 3); got.Verdict != "FAIL" {
		t.Errorf("an anchor only another agent read must be a finding against the closer; got %s: %s", got.Verdict, got.Detail)
	}
}

// AN INQUIRY JOINS LIKE ANY OTHER ANCHOR (gblock, 2026-10-19).
func TestAnInquiryAnchorJoinsOnTheSittingsReads(t *testing.T) {
	run := attestRun(t, "show inquiries", "report text at line 103", attestInquiry)
	read := transcriptOf(t, closerAgent,
		read(`feov-record show inquiries`, attestInquiry+" abandoned: the method could not settle it"), closing(closeCommand(attestInquiry)))
	if got := AttestationAudit(runtest.Open(t, run), []record.Transcript{read}, 3); got.Verdict != "PASS" {
		t.Errorf("an inquiry the sitting read must reconcile; got %s: %s", got.Verdict, got.Detail)
	}
	unread := transcriptOf(t, closerAgent, closing(closeCommand(attestInquiry)))
	if got := AttestationAudit(runtest.Open(t, run), []record.Transcript{unread}, 3); got.Verdict != "FAIL" {
		t.Errorf("an inquiry only the close names must be a finding; got %s: %s", got.Verdict, got.Detail)
	}
}

// WHERE THE CLOSE CANNOT BE FOUND, NOTHING IS MEASURED. A close no call's window holds cannot be told
// from the sitting's reads, and counting it as a read is the tautology this audit had; nor is a
// sitting whose transcript is not here a finding against its seat.
func TestACloseTheTranscriptDoesNotHoldIsNotMeasured(t *testing.T) {
	run := attestRun(t, "show report --anchor", "report text at line 103", attestOther)
	stray := call{closeCommand(attestOther), "closed", "2026-01-01T00:00:20.000Z", "2026-01-01T00:00:21.000Z"}
	for name, trs := range map[string][]record.Transcript{
		"the close in no call's window": {transcriptOf(t, closerAgent, read(`feov-record show report --anchor `+attestOther, "x"), stray)},
		"no transcript of the closer":   {transcriptOf(t, "someone-else", read(`feov-record show report --anchor `+attestOther, "x"))},
	} {
		t.Run(name, func(t *testing.T) {
			got := AttestationAudit(runtest.Open(t, run), trs, 3)
			if got.Verdict == "FAIL" || !strings.Contains(got.Detail, "NOT BE MEASURED") {
				t.Errorf("want the closure counted not measured; got %s: %s", got.Verdict, got.Detail)
			}
		})
	} // A STAMP THAT DOES NOT READ locates no call either: dropped, the close would count as a read.
	unstamped := attestRunAt(t, "2026-01-01T00:00:10Z", "show report --anchor", "report text at line 103", attestOther)
	trs := []record.Transcript{transcriptOf(t, closerAgent, closing(closeCommand(attestOther)))}
	if got := AttestationAudit(runtest.Open(t, unstamped), trs, 3); got.Verdict == "FAIL" || !strings.Contains(got.Detail, "no readable stamp") {
		t.Errorf("a close whose stamp does not read: want it counted not measured; got %s: %s", got.Verdict, got.Detail)
	}
}

// AND THE DISHONEST CASE STILL FAILS, which is the half that makes the fix worth having.
func TestAClosureCitingAnIDNobodyRanIsAFinding(t *testing.T) {
	run := attestRun(t, "show report --anchor", "report text at line 103", attestOther)
	trs := []record.Transcript{transcriptOf(t, closerAgent, read(`feov-record red-chair show board`, "nothing"), closing(closeCommand(attestOther)))}
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("an id in no tool call must be a finding; got %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, attestOther) {
		t.Errorf("the refusal must name the id it could not find:\n%s", got.Detail)
	}
}

// NOT MEASURED IS NOT A FINDING, and it must say so where a reader will see it.
func TestAnUnmeasurableAnchorIsReportedAsNotMeasuredRatherThanAsDishonesty(t *testing.T) {
	run := attestRun(t, "show report", "report text at line 103")
	trs := []record.Transcript{transcriptOf(t, closerAgent, read(`feov-record red-chair show report`, "x"), closing(closeCommand()))}
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict == "FAIL" {
		t.Errorf("an anchor with nothing to join on is not evidence of a dishonest record.\ngot %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "NOT BE MEASURED") {
		t.Errorf("the unmeasured closure must be counted out loud, or a 3-sample where 2 were "+
			"unmeasurable reads as a clean board:\n%s", got.Detail)
	}
}

// A CLOSURE THAT NAMES NO ANCHOR IS NOT MEASURED, whatever its prose holds. m39's G-c41c68ef named
// no anchor id; the audit cut "(scratch" from its target, found it in no call, and FAILed a voice
// sitting that served and read the report. A long word that a call does carry is no better: it
// would reconcile on word length, not on an act. Neither is read; both are counted not measured.
func TestAClosureNamingNoAnchorIsNotMeasuredWhateverItsProse(t *testing.T) {
	target := "the report as served (scratch copy of show report, 76 lines)"
	run := attestRun(t, "show report", target)
	closeIt := closing(`feov-record close --id ` + attestGap + ` --verified-against "` + target + `"`)
	for name, reads := range map[string]call{
		"no call carries the prose":  read(`feov-record red-chair show report`, "the report"),
		"a call carries a long word": read(`feov-record red-chair show report --match scratch`, "(scratch copy of show report, 76 lines)"),
	} {
		t.Run(name, func(t *testing.T) {
			got := AttestationAudit(runtest.Open(t, run), []record.Transcript{transcriptOf(t, closerAgent, reads, closeIt)}, 3)
			if got.Verdict != "PASS" || !strings.HasPrefix(got.Detail, "0/1 anchored closure(s)") ||
				!strings.Contains(got.Detail, "NOT BE MEASURED") || !strings.Contains(got.Detail, "names no anchor id") {
				t.Errorf("a closure with no anchor id must be counted not measured, neither reconciled nor accused; got %s: %s", got.Verdict, got.Detail)
			}
		})
	}
}

// THE AUDIT READS THE CLOSURE'S ANCHORS FROM THE RECORD. An id in --verified-with's words alone is no
// anchor named, so the closure is not measured — rather than reconciled, or accused, on an id read
// out of prose.
func TestAnIDInTheToolsWordsAloneIsNoAnchor(t *testing.T) {
	run := attestRun(t, "show report --anchor "+attestOther, "report text at line 103")
	trs := []record.Transcript{transcriptOf(t, closerAgent, read(`feov-record red-chair show report --anchor `+attestOther, "x"), closing(closeCommand()))}
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict == "FAIL" || !strings.Contains(got.Detail, "NOT BE MEASURED") {
		t.Errorf("a closure naming no anchor must be counted not measured; got %s: %s", got.Verdict, got.Detail)
	}
}

// A REPLAYED CALL IS THE SAME CALL. The harness writes each request's payload into the transcript
// again — an `api-request-blob` line with no timestamp — so the close recurs, outside any window; read
// as a call of its own, it would be the close reconciling itself once more.
func TestAReplayedCloseIsStillTheClose(t *testing.T) {
	run := attestRun(t, "show report", "report text at line 103", "K-24a6af66")
	tr := transcriptOf(t, closerAgent, closing(closeCommand("K-24a6af66")))
	body, err := os.ReadFile(tr.Path)
	if err != nil {
		t.Fatal(err)
	}
	blob := `{"type":"api-request-blob","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_` + closerAgent + `_00","name":"Bash","input":{"command":` +
		jsStringify(closeCommand("K-24a6af66")) + `}}]}}` + "\n"
	if err := os.WriteFile(tr.Path, append(body, blob...), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := AttestationAudit(runtest.Open(t, run), []record.Transcript{tr}, 3); got.Verdict != "FAIL" {
		t.Errorf("an anchor only the close names, replayed in a request blob, must still be a finding; got %s: %s", got.Verdict, got.Detail)
	}
}
