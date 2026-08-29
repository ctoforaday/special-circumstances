package dashboard

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

// A STUBBED report.md IS NOT A FINISHED RUN, and for the life of every run it was read as one.
//
// setup's skeleton creates report.md — its own comment documents that file as `bench assemble`'s
// output and stubs it anyway — and Terminal was fileExists(runDir/report.md). So the dashboard
// said "run complete — the assembler wrote the report" from the moment setup ran, before a seat
// was dispatched. Measured 2026-08-22: it said so for 55 minutes with blue-lane-1 live in the
// very next section of the same page.

func runWithStubbedReportButNoOutcome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	// Exactly what setup leaves behind: the heading, and nothing else.
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte("# report.md — a topic\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, dir, recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Problem:         proto.String("p"),
		RequiredFix:     proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"),
		Class:           proto.String("self-attestation"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}))
	return dir
}

func TestAStubbedReportIsNotATerminalRun(t *testing.T) {
	m := BuildModel(runtest.Open(t, runWithStubbedReportButNoOutcome(t)), t.TempDir(), Config{}, 0)
	if m.Terminal {
		t.Error("a run whose report.md is setup's stub, with no terminal act on the record, is NOT " +
			"complete. Terminal read a filename; the fact lives on the `outcome` event.")
	}
	if m.TerminalVerdict != "" {
		t.Errorf("no outcome event, so there is no terminal verdict to report; got %q", m.TerminalVerdict)
	}
}

// AND THE PAGE MUST NOT SAY IT EITHER — the field and the sentence it gates are separate failures,
// and only the sentence is what an operator actually reads.
func TestTheLiveRunPageDoesNotAnnounceCompletion(t *testing.T) {
	m := BuildModel(runtest.Open(t, runWithStubbedReportButNoOutcome(t)), t.TempDir(), Config{}, 0)
	html := RenderHTML(m)
	for _, bad := range []string{"run complete", "the assembler wrote the report"} {
		if strings.Contains(html, bad) {
			t.Errorf("a live run's dashboard says %q", bad)
		}
	}
	if !strings.Contains(html, "Seats live now") {
		t.Error("a run with no terminal act should render the live-seats section")
	}
}

// The other direction: a recorded outcome IS terminal, and the verdict travels from the record.
func TestARecordedOutcomeIsTerminal(t *testing.T) {
	dir := runWithStubbedReportButNoOutcome(t)
	recordtest.Seed(t, dir, recordtest.At(t, "judge-terminal", 1, "judge-terminal:outcome:#1", &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
		Prose:   proto.String("the round ceiling arrived before red could pass the final revision"),
	}))
	m := BuildModel(runtest.Open(t, dir), t.TempDir(), Config{}, 0)
	if !m.Terminal {
		t.Error("the bench recorded an outcome; the run is terminal")
	}
	// THE SPELLING IS THE SCHEMA'S, not a literal. The point is that the verdict comes off the
	// OUTCOME EVENT rather than the rendered report; a hardcoded "CEILING" was really pinning how
	// the payload record stored the seat's uppercase word.
	if want := recordpb.Word(recordpb.RunOutcome_RUN_OUTCOME_CEILING); m.TerminalVerdict != want {
		t.Errorf("the verdict comes off the outcome event; got %q, want %q", m.TerminalVerdict, want)
	}
}
