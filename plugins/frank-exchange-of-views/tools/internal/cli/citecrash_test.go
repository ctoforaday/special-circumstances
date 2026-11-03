package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
)

// A CRASH RETRY IS IDEMPOTENT, AND UNDER REPORT-AS-RECORD (#709) THE TORN-SPLICE WINDOW IS GONE.
//
// The report is the record projection: a marker exists only as its event, so no marker sits on the
// page that no event backs. A placing act and its Anchor are one write (record.AppendPlaced); a record
// migrate brought forward from an archive can still hold the act without its Anchor, where that run
// crashed between the two. The --key retry finishes it, and that is what this file pins.

// `lens finding` on a record holding the finding without its anchor event: the idempotent retry
// places the anchor, rather than answering and leaving the pair half on the record.
func TestFindingRetryFinishesTheHalfAppendedPair(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green<!--fx:F-0badf00d-->.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, runDir, recordtest.At(t, "red-lens-evidence", "red-lens-evidence:finding:F-0badf00d", &recordpb.Finding{
		Id: proto.String("F-0badf00d"), FindingKey: proto.String("K1"),
		Location: proto.String("The sky is blue and the grass is green."), Text: proto.String("an unfounded leap"),
		Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}))
	out, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--quote", "The sky is blue and the grass is green.",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium",
		"--key", "K1", "--reason", "an unfounded leap")
	if err != nil {
		t.Fatalf("the retry itself failed: %v", err)
	}
	if !strings.Contains(out, "F-0badf00d") || !strings.Contains(out, repeatedWords) {
		t.Errorf("the retry did not answer idempotently with the prior label: %q", out)
	}
	owed, err := record.UnplacedLocation(runtest.Open(t, runDir), "F-0badf00d")
	if err != nil {
		t.Fatal(err)
	}
	if owed != "" {
		t.Error("the retry left the pair half-appended: the finding exists and its anchor event still does not")
	}
	assertNoAnchorViolations(t, runDir)
}

func assertNoAnchorViolations(t *testing.T, runDir string) {
	t.Helper()
	violations, err := consistency.Check(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range violations {
		if strings.Contains(v, "anchor-record") {
			t.Errorf("%s", v)
		}
	}
}
