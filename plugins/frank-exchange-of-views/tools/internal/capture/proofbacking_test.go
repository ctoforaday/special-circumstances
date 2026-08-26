package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// proofBackingRun seeds proof events with the given labels and writes a blue report anchoring
// the labels named. The anchor form is the tool's own — the same token weaveProofs consumes.
func proofBackingRun(t *testing.T, onRecord []string, anchored []string) string {
	t.Helper()
	run := t.TempDir()
	var body strings.Builder
	body.WriteString("# report\n\n")
	for i, label := range anchored {
		body.WriteString("A sentence a computation backs<!--proof:" + label + "-->.\n\n")
		_ = i
	}
	write(t, filepath.Join(run, "blue", "report.md"), body.String())

	var evs []*recordpb.Event
	for i, label := range onRecord {
		evs = append(evs, recordtest.At(t, "blue-lane-1", 1, "blue-lane-1:proof:#"+itoa(i+1),
			&recordpb.Proof{
				ProofId:    proto.String(label),
				ProofSha:   proto.String(strings.Repeat("a", 8) + itoa(i)),
				ProofBasis: proto.String("reproducible"),
				Script:     proto.String("blue/candidates/x.sh"),
			}))
	}
	recordtest.Seed(t, run, evs...)
	return run
}

func TestProofBackingPassesWhenEveryProofIsAnchoredAndEveryAnchorResolves(t *testing.T) {
	run := proofBackingRun(t, []string{"p-aaaa1111", "p-bbbb2222"}, []string{"p-aaaa1111", "p-bbbb2222"})
	got := ProofBackingAudit(run)
	if got.Verdict != "PASS" {
		t.Fatalf("got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "2 proof(s) on the record, 2 anchored") {
		t.Errorf("the counts belong in the detail; got:\n%s", got.Detail)
	}
}

// THE ASK, STATED EXACTLY: a computation this run performed that the report points at nowhere.
// The evidence sits in the record and the cache while the sentence it was run for stands on its
// own authority.
func TestProofBackingCatchesAComputationTheReportNeverPointsAt(t *testing.T) {
	run := proofBackingRun(t, []string{"p-aaaa1111", "p-bbbb2222"}, []string{"p-aaaa1111"})
	got := ProofBackingAudit(run)
	if got.Verdict != "FAIL" {
		t.Fatalf("an unanchored proof must FAIL: got %s — %s", got.Verdict, got.Detail)
	}
	for _, want := range []string{"reach the report nowhere", "p-bbbb2222"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got.Detail)
		}
	}
	if strings.Contains(got.Detail, "p-aaaa1111") {
		t.Errorf("only the unanchored proof may be named; got:\n%s", got.Detail)
	}
}

// THE MIRROR: an anchor resolving to no proof event — a claim pointing at nothing. The assembler
// already renders this as prose ("unresolved proof …"); nothing read that line, which is why it
// is computed from the two sets here instead of recovered from the sentence.
func TestProofBackingCatchesAnAnchorThatResolvesToNothing(t *testing.T) {
	run := proofBackingRun(t, []string{"p-aaaa1111"}, []string{"p-aaaa1111", "p-cccc3333"})
	got := ProofBackingAudit(run)
	if got.Verdict != "FAIL" {
		t.Fatalf("an unresolved anchor must FAIL: got %s — %s", got.Verdict, got.Detail)
	}
	for _, want := range []string{"resolve to NO proof event", "p-cccc3333"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got.Detail)
		}
	}
}

// Both directions at once are two findings, not one — they are different defects and a reader
// fixing one should not be told the other is the same thing.
func TestProofBackingReportsBothDirectionsSeparately(t *testing.T) {
	run := proofBackingRun(t, []string{"p-aaaa1111"}, []string{"p-cccc3333"})
	got := ProofBackingAudit(run)
	if got.Verdict != "FAIL" {
		t.Fatalf("got %s — %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "resolve to NO proof event") || !strings.Contains(got.Detail, "reach the report nowhere") {
		t.Errorf("both directions must be named; got:\n%s", got.Detail)
	}
}

// THE UNCHECKABLE STATES ARE NOT PASSES, and one of them is subtle: assembly WEAVES the anchors
// away, so pointing this at the assembled report would find none and — without this distinction —
// report every proof as unanchored on a perfectly healthy run.
func TestProofBackingWillNotReportAnUncheckedRunAsAClearOne(t *testing.T) {
	// No blue/report.md at all.
	bare := t.TempDir()
	recordtest.Seed(t, bare)
	if got := ProofBackingAudit(bare); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "woven away at assembly") {
		t.Errorf("absent blue report: got %s — %s", got.Verdict, got.Detail)
	}

	// A run with neither proofs nor anchors has nothing to join, and that is not a finding.
	empty := proofBackingRun(t, nil, nil)
	if got := ProofBackingAudit(empty); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "recorded no proofs") {
		t.Errorf("no proofs and no anchors: got %s — %s", got.Verdict, got.Detail)
	}

	// THE FALSE FAIL THIS AUDIT SHIPPED ON ITS FIRST RUN AGAINST REAL ARTIFACTS. Both 2026-08-23
	// run directories in research/ carry the committed report and NOT records/, which is
	// gitignored — and the record layer reads an absent record directory as a legal EMPTY run
	// rather than an error. The audit convicted 6 and 2 perfectly good anchors of "resolving to
	// nothing". An event-less record beside a written report is a report without its record, and
	// every real sitting registers, so it is not a run this audit can speak about.
	reportOnly := t.TempDir()
	write(t, filepath.Join(reportOnly, "blue", "report.md"),
		"a claim<!--proof:p-aaaa1111--> and another<!--proof:p-bbbb2222-->\n")
	recordtest.Seed(t, reportOnly) // a record with no events at all
	got0 := ProofBackingAudit(reportOnly)
	if got0.Verdict != "SKIP" {
		t.Fatalf("a report with no record behind it must not be convicted: got %s — %s", got0.Verdict, got0.Detail)
	}
	for _, want := range []string{"carries no events at all", "rather than 2 claims pointing at nothing"} {
		if !strings.Contains(got0.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got0.Detail)
		}
	}

	// A record that cannot be read: the report is there, the join is not possible, and that must
	// not read as claims that all resolve.
	broken := t.TempDir()
	write(t, filepath.Join(broken, "blue", "report.md"), "x<!--proof:p-aaaa1111-->\n")
	if err := os.MkdirAll(filepath.Join(broken, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "records", "record.db"), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := ProofBackingAudit(broken)
	if got.Verdict != "SKIP" || !strings.Contains(got.Detail, "NOT a run whose claims all resolve") {
		t.Errorf("unreadable record: got %s — %s", got.Verdict, got.Detail)
	}
}
