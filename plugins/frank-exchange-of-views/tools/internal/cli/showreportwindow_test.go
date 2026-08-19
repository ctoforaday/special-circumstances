package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"strings"
	"testing"
)

// `show report --anchor <id> [--window N]` — reading the report AT one anchor instead of pulling
// the whole document to check one sentence.
//
// The unit contract (which lines come back, what a missing or duplicated anchor does) is pinned in
// internal/anchor/window_test.go. What only exists at THIS layer is the wiring: that the flags are
// registered on the right subcommand, that the window arm is reached before the whole-document
// write, and that --window without --anchor is REFUSED rather than silently ignored.

const windowReport = `# Report

## Findings

far above, out of every window.

three paragraphs up.

two paragraphs up.

a paragraph before the anchored one.

the corpus holds 340 records.<!--fx:f-a1b2c3-->

a paragraph after it.

two paragraphs down.

three paragraphs down.

## Method

method text, far below and out of every window.
`

func TestShowReportAtAnAnchorReadsTheLiveTextAndSaysWhereItIs(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, windowReport)

	out, err := run(t, "show", "report", "--seat-id", "blue-respond-r1", "--run", runDir, "--anchor", "f-a1b2c3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "the corpus holds 340 records.") {
		t.Errorf("the anchored text is not in the window:\n%s", out)
	}
	if !strings.Contains(out, "## Findings") {
		t.Errorf("the window does not say which section it is in — a reader's first question:\n%s", out)
	}
	// LINE NUMBERS ARE OUTPUT, so a seat can say what it read. They are never an address; see
	// internal/anchor/window.go.
	if !strings.Contains(out, "showing lines") {
		t.Errorf("the window does not report the lines it covers:\n%s", out)
	}
	// The whole document is NOT what came back — that is the point of the flag.
	if strings.Contains(out, "method text, far below") || strings.Contains(out, "far above, out of every window.") {
		t.Errorf("--anchor returned the whole report; the window arm was not reached:\n%s", out)
	}
}

// --window SIZES THE READ, and the size has to actually take effect. A flag that parses and
// changes nothing is the plausible-zero shape wearing a different hat.
func TestTheWindowSizeChangesWhatComesBack(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, windowReport)

	narrow, err := run(t, "show", "report", "--seat-id", "blue-respond-r1", "--run", runDir, "--anchor", "f-a1b2c3", "--window", "0")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(narrow, "a paragraph before the anchored one.") {
		t.Errorf("--window 0 returned neighbours; the size is not reaching the reader:\n%s", narrow)
	}
	wide, err := run(t, "show", "report", "--seat-id", "blue-respond-r1", "--run", runDir, "--anchor", "f-a1b2c3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(wide, "a paragraph before the anchored one.") {
		t.Errorf("the default window does not reach a neighbouring paragraph:\n%s", wide)
	}
}

// A STALE ANCHOR REFUSES. An empty read would say "the report has nothing here", which is a
// different fact and a false one.
func TestShowReportRefusesAnAnchorThatIsNotThere(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, windowReport)

	out, err := run(t, "show", "report", "--seat-id", "blue-respond-r1", "--run", runDir, "--anchor", "f-deadbeef")
	if err == nil {
		t.Fatalf("a stale anchor produced output rather than a refusal:\n%s", out)
	}
	if !strings.Contains(err.Error(), "stale reference") {
		t.Errorf("the refusal does not say what a missing anchor MEANS:\n%v", err)
	}
}

// --window WITHOUT --anchor IS REFUSED. Silently ignoring it hands back the whole report, which a
// seat cannot distinguish from a report that is simply that long — the same defect as an unknown
// --format rendering the default and exiting 0.
func TestWindowWithoutAnAnchorIsRefusedRatherThanIgnored(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, windowReport)

	out, err := run(t, "show", "report", "--seat-id", "blue-respond-r1", "--run", runDir, "--window", "1")
	if err == nil {
		t.Fatalf("--window alone returned the whole report and exited 0:\n%s", out)
	}
	if !strings.Contains(err.Error(), "--anchor") {
		t.Errorf("the refusal does not name the flag that makes --window meaningful:\n%v", err)
	}
}

// EVERY ROLE READS THE SAME ARTIFACT THE SAME WAY. `show report` is defined once in
// internal/cli/seat, so a flag registered on one role and missing on another would mean the
// projection had grown a per-role surface — which is exactly what defining it once is for.
func TestEveryRoleCanReadAtAnAnchor(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, windowReport)

	for _, role := range []string{"blue", "lens", "merge", "bench"} {
		out, err := run(t, "show", "report", "--run", runDir, "--anchor", "f-a1b2c3", "--seat-id", record.SampleSeatOf(role))
		if err != nil {
			t.Errorf("%s cannot read at an anchor: %v", role, err)
			continue
		}
		if !strings.Contains(out, "the corpus holds 340 records.") {
			t.Errorf("%s got a window without the anchored text:\n%s", role, out)
		}
	}
}
