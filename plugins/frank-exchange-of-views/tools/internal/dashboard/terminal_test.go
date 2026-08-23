package dashboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
	e := record.Event{SeatID: "red-merge-r1", Nonce: "0000000a", Round: 1, Type: "mint",
		Key: "red-merge-r1:mint:R1-1",
		Payload: record.NewPayload().Set("gap_id", "R1-1").Set("problem", "p").
			Set("severity", "medium").Set("likelihood", "medium").Set("impact", "medium")}
	line, err := record.MarshalEvent(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(recs, "events-red-merge-r1-0000000a.jsonl"), append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAStubbedReportIsNotATerminalRun(t *testing.T) {
	m := BuildModel(runWithStubbedReportButNoOutcome(t), t.TempDir(), Config{}, 0)
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
	m := BuildModel(runWithStubbedReportButNoOutcome(t), t.TempDir(), Config{}, 0)
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
	e := record.Event{SeatID: "bench-r1", Nonce: "0000000c", Round: 1, Type: "outcome",
		Key: "bench-r1:outcome:1", Payload: record.NewPayload().Set("verdict", "CEILING")}
	line, err := record.MarshalEvent(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "records", "events-bench-r1-0000000c.jsonl"), append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	m := BuildModel(dir, t.TempDir(), Config{}, 0)
	if !m.Terminal {
		t.Error("the bench recorded an outcome; the run is terminal")
	}
	if m.TerminalVerdict != "CEILING" {
		t.Errorf("the verdict comes off the outcome event; got %q", m.TerminalVerdict)
	}
}
