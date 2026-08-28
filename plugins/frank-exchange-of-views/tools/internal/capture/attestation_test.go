package capture

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

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

// attestRun writes a run whose board holds one gap, minted at r1 and closed at r2 with the given
// anchor fields.
//
// SEEDED THROUGH THE STORE, not by hand-writing shard files. This wrote `events-<seat>-<nonce>.jsonl`
// with record.MarshalEvent; there are no shards and no nonce, and a fixture that builds the record
// by a route production cannot take is testing a shape rather than the system.
func attestRun(t *testing.T, anchorTool, anchorTarget string) string {
	t.Helper()
	dir := t.TempDir()
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
			GapId:           proto.String("R1-1"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			Class:           proto.String("self-attestation"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
		recordtest.At(t, "red-merge-r2", 2, "red-merge-r2:close:R1-1", &recordpb.Close{
			GapId:        proto.String("R1-1"),
			ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat:   proto.String("red-merge-r2"),
			AnchorTool:   proto.String(anchorTool),
			AnchorTarget: proto.String(anchorTarget),
			Prose:        proto.String("verified at the leaf"),
		}),
	)
	return dir
}

// transcriptWith writes one agent transcript whose tool_use inputs carry the given commands.
func transcriptWith(t *testing.T, commands ...string) (string, []string) {
	t.Helper()
	tr := t.TempDir()
	var b strings.Builder
	for _, c := range commands {
		b.WriteString(`{"message":{"role":"assistant","content":[{"type":"tool_use","input":{"command":` +
			jsStringify(c) + `}}]}}` + "\n")
	}
	if err := os.WriteFile(filepath.Join(tr, "agent-x.jsonl"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return tr, []string{"agent-x.jsonl"}
}

// A PRECISE PROSE TARGET IS NOT A VAGUE ONE. The seat did the work, anchored it to an exact line,
// and named the id in the tool it ran. That must reconcile.
func TestAPreciseAnchorWithShortWordsReconcilesByItsID(t *testing.T) {
	run := attestRun(t, "show report --anchor f-0dd40334", "report text at line 103")
	tr, files := transcriptWith(t, `feov-record red-merge-r2 show report --anchor f-0dd40334`)
	got := AttestationAudit(runtest.Open(t, run), tr, files, 3)
	if got.Verdict != "PASS" {
		t.Errorf("a closure citing an id the transcript carries must reconcile.\ngot %s: %s", got.Verdict, got.Detail)
	}
}

// AND THE DISHONEST CASE STILL FAILS, which is the half that makes the fix worth having.
func TestAClosureCitingAnIDNobodyRanIsAFinding(t *testing.T) {
	run := attestRun(t, "show report --anchor f-0dd40334", "report text at line 103")
	tr, files := transcriptWith(t, `feov-record red-merge-r2 show board`)
	got := AttestationAudit(runtest.Open(t, run), tr, files, 3)
	if got.Verdict != "FAIL" {
		t.Fatalf("an id in no tool call must be a finding; got %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "f-0dd40334") {
		t.Errorf("the refusal must name the id it could not find:\n%s", got.Detail)
	}
}

// NOT MEASURED IS NOT A FINDING, and it must say so where a reader will see it.
func TestAnUnmeasurableAnchorIsReportedAsNotMeasuredRatherThanAsDishonesty(t *testing.T) {
	run := attestRun(t, "show report", "report text at line 103")
	tr, files := transcriptWith(t, `feov-record red-merge-r2 show report`)
	got := AttestationAudit(runtest.Open(t, run), tr, files, 3)
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
	tr, files := transcriptWith(t, `feov-record red-merge-r2 show evidence --projection full`)
	if got := AttestationAudit(runtest.Open(t, run), tr, files, 3); got.Verdict != "PASS" {
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
	ns, byID := needlesFor(Claim{Tool: "show evidence; show report --anchor c-5058e9f8", Target: "evidence projection"})
	if !byID || len(ns) != 1 || ns[0] != "c-5058e9f8" {
		t.Errorf("an id in the tool must win over prose in the target; got %v byID=%v", ns, byID)
	}
	ns, byID = needlesFor(Claim{Tool: "show report", Target: "evidence projection"})
	if byID || len(ns) != 1 || ns[0] != "projection" {
		t.Errorf("with no id, fall back to prose; got %v byID=%v", ns, byID)
	}
	if ns, _ := needlesFor(Claim{Tool: "show report", Target: "report text at line 103"}); len(ns) != 0 {
		t.Errorf("nothing to join on must be empty, not a guess; got %v", ns)
	}
}
