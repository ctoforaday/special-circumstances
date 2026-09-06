package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

const twoTurns = `{"agentId":"AGENT","timestamp":"2026-09-03T03:00:00.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"thinking"}],"usage":{"input_tokens":100,"output_tokens":3,"cache_read_input_tokens":900,"cache_creation_input_tokens":0}}}
{"agentId":"AGENT","timestamp":"2026-09-03T03:00:30.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"tool_use"}],"usage":{"input_tokens":110,"output_tokens":50,"cache_read_input_tokens":950,"cache_creation_input_tokens":0}}}
`

func ingestFixture(t *testing.T, files map[string]string) (record.Run, string, []string) {
	t.Helper()
	runDir := recordtest.TmpRun(t)
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	var names []string
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	return run, dir, names
}

func seatTurnRows(t *testing.T, run record.Run) int {
	t.Helper()
	n, err := record.CountSeatTurns(run)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// THE TURNS LAND, AND CAPTURE SAYS HOW MANY. A silent ingest would leave an empty table that
// reads exactly like a run whose seats took no turns.
func TestIngestSeatTurnsWritesRowsAndReportsThem(t *testing.T) {
	run, dir, files := ingestFixture(t, map[string]string{
		"agent-one.jsonl": twoTurns,
		"agent-two.jsonl": strings.ReplaceAll(twoTurns, "AGENT", "OTHER"),
	})

	line := ingestSeatTurns(run, dir, files)
	if got := seatTurnRows(t, run); got != 4 {
		t.Errorf("two transcripts of two turns gave %d rows, want 4", got)
	}
	if !strings.Contains(line, "4 row(s)") || !strings.Contains(line, "2 transcript(s)") {
		t.Errorf("the reported line does not state what was ingested: %q", line)
	}

	// Re-running capture on unchanged transcripts must not double the run's measured cost.
	if line := ingestSeatTurns(run, dir, files); !strings.Contains(line, "0 row(s)") {
		t.Errorf("a re-run reported %q, want 0 new rows", line)
	}
	if got := seatTurnRows(t, run); got != 4 {
		t.Errorf("a re-run left %d rows, want 4", got)
	}
}

// A TRANSCRIPT THAT COULD NOT BE INGESTED IS NAMED IN THE COUNT, because the alternative is a
// smaller number that looks like a smaller run.
func TestIngestSeatTurnsNamesWhatItCouldNotRead(t *testing.T) {
	run, dir, _ := ingestFixture(t, map[string]string{"agent-one.jsonl": twoTurns})

	line := ingestSeatTurns(run, dir, []string{"agent-one.jsonl", "agent-missing.jsonl"})
	if !strings.Contains(line, "NOT INGESTED") {
		t.Errorf("an unreadable transcript was swallowed: %q", line)
	}
	if !strings.Contains(line, "2 row(s)") {
		t.Errorf("the readable transcript's turns are missing from %q", line)
	}
}

// A TRANSCRIPT WITH NO USAGE TURNS IS NOT A FAILURE. A seat dispatched and cancelled before its
// first API round has nothing to measure, and counting it as an ingest failure would send someone
// looking for a broken file.
func TestIngestSeatTurnsTreatsATurnlessTranscriptAsEmptyNotBroken(t *testing.T) {
	run, dir, files := ingestFixture(t, map[string]string{
		"agent-quiet.jsonl": `{"agentId":"Q","type":"user","message":{"content":[{"type":"text"}]}}` + "\n",
	})
	line := ingestSeatTurns(run, dir, files)
	if strings.Contains(line, "NOT INGESTED") {
		t.Errorf("a transcript with no API rounds was reported as a failure: %q", line)
	}
	if got := seatTurnRows(t, run); got != 0 {
		t.Errorf("got %d rows from a transcript with no usage", got)
	}
}
