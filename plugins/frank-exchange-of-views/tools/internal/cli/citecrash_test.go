package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
)

// THE SPLICE AND THE EVENT ARE TWO ACTS, AND A CRASH BETWEEN THEM IS ORDINARY.
//
// `blue cite` mutates blue/report.md first and records the cite event second — deliberately, so a
// mis-quote is refused before anything is written. The cost of that order is a window: a process
// killed (or a db write refused) after the splice commits leaves an anchor token on the page that
// no event backs. The record's own design says retries are ordinary — mint, cite and prove all
// carry crash-retry keys — so the question is not whether the torn state occurs but what the
// RETRY does when it meets it.
//
// What it did: ExistingCiteByKey looks for a PRIOR EVENT, and the crash recorded none, so the
// retry saw a fresh cite — minted a NEW label and spliced a SECOND anchor beside the orphan. One
// sentence, two markers: the orphan is immortal (anchors are), backs nothing, and every later
// edit of that sentence must carry both tokens forever.
func TestCiteRetryAfterTornSpliceAdoptsTheOrphan(t *testing.T) {
	runDir := newRun(t)
	// The state a crash leaves: the splice committed, the event did not.
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green<!--cite:c-0badf00d-->.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": []byte("<html>a source</html>")}})

	if _, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", "The sky is blue and the grass is green.",
		"--url", "https://sky/1", "--title", "Sky Facts", "--key", "K1"); err != nil {
		t.Fatalf("the retry itself failed: %v", err)
	}

	report := readReport(t, runDir)
	if n := len(citeAnchorRe.FindAllString(report, -1)); n != 1 {
		t.Errorf("the sentence carries %d citation anchors after the retry, want 1:\n%s", n, report)
	}
	ev := firstCiteEvent(t, runDir)
	if ev == nil {
		t.Fatal("no cite event recorded")
	}
	if ev.GetLabel() != "c-0badf00d" {
		t.Errorf("the retry minted %s instead of adopting the orphan c-0badf00d — the orphan now backs nothing forever", ev.GetLabel())
	}
	violations, err := consistency.Check(runDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		if strings.Contains(v, "anchor-record") {
			t.Errorf("%s", v)
		}
	}
}

// The same window on `lens finding` (window A: splice committed, neither event appended).
func TestFindingRetryAfterTornSpliceAdoptsTheOrphan(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green<!--fx:f-0badf00d-->.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--quote", "The sky is blue and the grass is green.",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium",
		"--key", "K1", "--reason", "an unfounded leap"); err != nil {
		t.Fatalf("the retry itself failed: %v", err)
	}
	report := readReport(t, runDir)
	if n := strings.Count(report, "<!--fx:"); n != 1 {
		t.Errorf("the sentence carries %d finding markers after the retry, want 1:\n%s", n, report)
	}
	if !strings.Contains(report, "<!--fx:f-0badf00d-->") {
		t.Errorf("the orphan marker was replaced rather than adopted:\n%s", report)
	}
	assertNoAnchorViolations(t, runDir)
}

// Window B on `lens finding`: the finding appended, the anchor event did not, and the idempotent
// retry used to seal that state forever — it saw the finding, answered, and never looked.
func TestFindingRetryFinishesTheHalfAppendedPair(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green<!--fx:f-0badf00d-->.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, runDir, recordtest.At(t, "red-lens-r1-L1", 1, "red-lens-r1-L1:finding:L1-F1", &recordpb.Finding{
		Label: proto.String("L1-F1"), FindingId: proto.String("f-0badf00d"), FindingKey: proto.String("K1"),
		Location: proto.String("The sky is blue and the grass is green."), Text: proto.String("an unfounded leap"),
	}))
	out, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--quote", "The sky is blue and the grass is green.",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium",
		"--key", "K1", "--reason", "an unfounded leap")
	if err != nil {
		t.Fatalf("the retry itself failed: %v", err)
	}
	if !strings.Contains(out, "L1-F1") {
		t.Errorf("the retry did not answer idempotently with the prior label: %q", out)
	}
	anchored, err := record.AnchorEventExists(runDir, "f-0badf00d")
	if err != nil {
		t.Fatal(err)
	}
	if !anchored {
		t.Error("the retry left the pair half-appended: the finding exists and its anchor event still does not")
	}
	assertNoAnchorViolations(t, runDir)
}

// The same window on `blue prove` (window A).
func TestProveRetryAfterTornSpliceAdoptsTheOrphan(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sum of the first ten primes is 129<!--proof:p-0badf00d-->.\n")
	registerBlue(t, runDir)
	script := filepath.Join(runDir, "sum.js")
	if err := os.WriteFile(script, []byte("console.log('129')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "prove", "--run", runDir, "--seat-id", citeSeat,
		"--quote", "The sum of the first ten primes is 129.",
		"--script", script, "--key", "K1", "--reason", "computed, not asserted"); err != nil {
		t.Fatalf("the retry itself failed: %v", err)
	}
	report := readReport(t, runDir)
	if n := strings.Count(report, "<!--proof:"); n != 1 {
		t.Errorf("the sentence carries %d proof markers after the retry, want 1:\n%s", n, report)
	}
	if !strings.Contains(report, "<!--proof:p-0badf00d-->") {
		t.Errorf("the orphan marker was replaced rather than adopted:\n%s", report)
	}
	assertNoAnchorViolations(t, runDir)
}

func assertNoAnchorViolations(t *testing.T, runDir string) {
	t.Helper()
	violations, err := consistency.Check(runDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		if strings.Contains(v, "anchor-record") {
			t.Errorf("%s", v)
		}
	}
}
