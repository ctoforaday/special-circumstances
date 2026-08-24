package cli

import (
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
