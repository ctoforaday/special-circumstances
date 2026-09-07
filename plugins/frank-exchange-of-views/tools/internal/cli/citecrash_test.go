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
// The report is the record projection: a marker exists only as its event, so there is no longer a
// state where a marker sits on the page that no event backs — the "torn splice" the file world had,
// and its orphan-adoption retry, cannot arise. What remains is the two-EVENT window: a verb that
// appends more than one event (lens finding appends the Finding and then its Anchor) can crash
// between them, leaving a half-appended pair. The --key retry finishes it, and that is what this
// file now pins.

// Window B on `lens finding`: the finding appended, the anchor event did not, and the idempotent
// retry used to seal that state forever — it saw the finding, answered, and never looked.
func TestFindingRetryFinishesTheHalfAppendedPair(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green<!--fx:f-0badf00d-->.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-evidence"); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, runDir, recordtest.At(t, "red-lens-r1-evidence", 1, "red-lens-r1-evidence:finding:L1-F1", &recordpb.Finding{
		Label: proto.String("L1-F1"), FindingId: proto.String("f-0badf00d"), FindingKey: proto.String("K1"),
		Location: proto.String("The sky is blue and the grass is green."), Text: proto.String("an unfounded leap"),
	}))
	out, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-evidence",
		"--quote", "The sky is blue and the grass is green.",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium",
		"--key", "K1", "--reason", "an unfounded leap")
	if err != nil {
		t.Fatalf("the retry itself failed: %v", err)
	}
	if !strings.Contains(out, "L1-F1") {
		t.Errorf("the retry did not answer idempotently with the prior label: %q", out)
	}
	anchored, err := record.AnchorEventExists(runtest.Open(t, runDir), "f-0badf00d")
	if err != nil {
		t.Fatal(err)
	}
	if !anchored {
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
