package capture

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"slices"
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
// (`report` is exactly six), so mostDistinctive returned "" and the closures were filed as
// findings. The third sampled closure reconciled only because its target happened to contain the
// ten-letter word `projection`. Honesty was decided by word length.
//
// AttestationAudit had no test of any kind, which is how it shipped.

// attestGap is the one gap attestRun's board holds, and the anchor a closure of it can name.
const attestGap = "G-0dd40334"

// attestRun writes a run whose board holds one gap, minted and then closed with the given
// verification and the report anchors the closure names.
//
// SEEDED THROUGH THE STORE, not by hand-writing shard files. This wrote `events-<seat>-<nonce>.jsonl`
// with record.MarshalEvent; there are no shards and no nonce, and a fixture that builds the record
// by a route production cannot take is testing a shape rather than the system.
func attestRun(t *testing.T, anchorTool, anchorTarget string, anchors ...string) string {
	t.Helper()
	dir := t.TempDir()
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-chair", "red-chair:mint:G1", &recordpb.Mint{
			GapId:           proto.String(attestGap),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			Class:           proto.String("self-attestation"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
		recordtest.At(t, "red-chair", "red-chair:close:G1", &recordpb.Close{
			GapId:        proto.String(attestGap),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   proto.String("red-chair"),
			AnchorTool:   proto.String(anchorTool),
			AnchorTarget: proto.String(anchorTarget),
			Anchors:      anchors,
			Prose:        proto.String("verified at the leaf"),
		}),
	)
	return dir
}

// transcriptWith writes one agent transcript whose tool_use inputs carry the given commands.
func transcriptWith(t *testing.T, commands ...string) []record.Transcript {
	t.Helper()
	tr := t.TempDir()
	var b strings.Builder
	for _, c := range commands {
		b.WriteString(`{"message":{"role":"assistant","content":[{"type":"tool_use","input":{"command":` +
			jsStringify(c) + `}}]}}` + "\n")
	}
	p := filepath.Join(tr, "trajectory-of-x.jsonl")
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return []record.Transcript{{AgentID: "x", Path: p}}
}

// A PRECISE PROSE TARGET IS NOT A VAGUE ONE. The seat did the work, anchored it to an exact line,
// and named the anchor it read. That must reconcile.
func TestAPreciseAnchorWithShortWordsReconcilesByItsID(t *testing.T) {
	run := attestRun(t, "show report --anchor", "report text at line 103", attestGap)
	trs := transcriptWith(t, `feov-record red-chair show report --anchor `+attestGap)
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict != "PASS" {
		t.Errorf("a closure citing an id the transcript carries must reconcile.\ngot %s: %s", got.Verdict, got.Detail)
	}
	// RECONCILED, NOT UNMEASURED: an id the reader does not recognise falls through to the prose path,
	// finds no fragment, and reports the same PASS with 0/1 behind it.
	if !strings.HasPrefix(got.Detail, "1/1 anchored closure(s) sampled and reconciled") || strings.Contains(got.Detail, "NOT BE MEASURED") {
		t.Errorf("the closure must reconcile BY ITS ID, not go unmeasured:\n%s", got.Detail)
	}
}

// AND THE DISHONEST CASE STILL FAILS, which is the half that makes the fix worth having.
func TestAClosureCitingAnIDNobodyRanIsAFinding(t *testing.T) {
	run := attestRun(t, "show report --anchor", "report text at line 103", attestGap)
	trs := transcriptWith(t, `feov-record red-chair show board`)
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("an id in no tool call must be a finding; got %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, attestGap) {
		t.Errorf("the refusal must name the id it could not find:\n%s", got.Detail)
	}
}

// NOT MEASURED IS NOT A FINDING, and it must say so where a reader will see it.
func TestAnUnmeasurableAnchorIsReportedAsNotMeasuredRatherThanAsDishonesty(t *testing.T) {
	run := attestRun(t, "show report", "report text at line 103")
	trs := transcriptWith(t, `feov-record red-chair show report`)
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict == "FAIL" {
		t.Errorf("an anchor with nothing to join on is not evidence of a dishonest record.\ngot %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "NOT BE MEASURED") {
		t.Errorf("the unmeasured closure must be counted out loud, or a 3-sample where 2 were "+
			"unmeasurable reads as a clean board:\n%s", got.Detail)
	}
}

// The prose path still works where prose is actually distinctive — this is the case that carried
// the whole audit before, and it must not regress.
func TestADistinctiveProseTargetStillReconciles(t *testing.T) {
	run := attestRun(t, "show evidence", "evidence projection and report text")
	trs := transcriptWith(t, `feov-record red-chair show evidence --projection full`)
	if got := AttestationAudit(runtest.Open(t, run), trs, 3); got.Verdict != "PASS" {
		t.Errorf("distinctive prose must still reconcile; got %s: %s", got.Verdict, got.Detail)
	}
}

// The six-character floor is deliberate and this pins WHY: admitting `report` would make the
// needle match every `show report` call in a run, turning a false FAIL into a false PASS.
func TestTheProseFloorStillRefusesAWordThatWouldMatchEverything(t *testing.T) {
	if got := mostDistinctive("report text at line 103"); got != "" {
		t.Errorf("mostDistinctive(%q) = %q — a needle this common reconciles anything", "report text at line 103", got)
	}
	if got := mostDistinctive("evidence projection and report text"); got != "projection" {
		t.Errorf("longest fragment over six chars: want projection, got %q", got)
	}
}

func TestNeedlesForPrefersTheRecordsOwnIdentifiers(t *testing.T) {
	ns, byID := needlesFor(Claim{Tool: "show evidence", Target: "evidence projection", Anchors: []string{"C-5058e9f8"}})
	if !byID || len(ns) != 1 || ns[0] != "C-5058e9f8" {
		t.Errorf("a named anchor must win over prose in the target; got %v byID=%v", ns, byID)
	}
	// EVERY NAMED ANCHOR IS A NEEDLE.
	want := []string{"F-0dd40334", "G-5e10a3c2"}
	if ns, byID := needlesFor(Claim{Tool: "show report", Target: "evidence projection", Anchors: want}); !byID || !slices.Equal(ns, want) {
		t.Errorf("needlesFor = %v byID=%v, want %v by id", ns, byID, want)
	}
	// THE TOOL'S WORDS ARE NOT READ FOR AN ID. The anchors are a field the write checked; an id typed
	// into --verified-with is prose, and the target's prose is what is left to join on.
	ns, byID = needlesFor(Claim{Tool: "show report --anchor F-0dd40334", Target: "evidence projection"})
	if byID || !slices.Equal(ns, []string{"projection"}) {
		t.Errorf("with no anchor named, fall back to prose; got %v byID=%v", ns, byID)
	}
	if ns, _ := needlesFor(Claim{Tool: "show report", Target: "report text at line 103"}); len(ns) != 0 {
		t.Errorf("nothing to join on must be empty, not a guess; got %v", ns)
	}
}

// THE AUDIT READS THE CLOSURE'S ANCHORS FROM THE RECORD. An id in --verified-with's words alone is no
// anchor named, so a closure whose target has no distinctive fragment is not measured — rather than
// reconciled, or accused, on an id read out of prose.
func TestAnIDInTheToolsWordsAloneIsNoAnchor(t *testing.T) {
	run := attestRun(t, "show report --anchor "+attestGap, "report text at line 103")
	trs := transcriptWith(t, `feov-record red-chair show report --anchor `+attestGap)
	got := AttestationAudit(runtest.Open(t, run), trs, 3)
	if got.Verdict == "FAIL" || !strings.Contains(got.Detail, "NOT BE MEASURED") {
		t.Errorf("a closure naming no anchor and no distinctive target must be counted not measured; got %s: %s", got.Verdict, got.Detail)
	}
}
