package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

const twoTurns = `{"agentId":"AGENT","timestamp":"2026-09-03T03:00:00.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"thinking"}],"usage":{"input_tokens":100,"output_tokens":3,"cache_read_input_tokens":900,"cache_creation_input_tokens":0}}}
{"agentId":"AGENT","timestamp":"2026-09-03T03:00:30.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"tool_use"}],"usage":{"input_tokens":110,"output_tokens":50,"cache_read_input_tokens":950,"cache_creation_input_tokens":0}}}
`

func ingestFixture(t *testing.T, files map[string]string) (record.Run, string, []record.Transcript) {
	t.Helper()
	runDir := recordtest.TmpRun(t)
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	var named []record.Transcript
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		named = append(named, record.Transcript{AgentID: name, Path: p})
	}
	return run, dir, named
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
	run, _, files := ingestFixture(t, map[string]string{
		"agent-one.jsonl": twoTurns,
		"agent-two.jsonl": strings.ReplaceAll(twoTurns, "AGENT", "OTHER"),
	})

	line := ingestSeatTurns(run, files)
	if got := seatTurnRows(t, run); got != 4 {
		t.Errorf("two transcripts of two turns gave %d rows, want 4", got)
	}
	if want := "seat turns: 4 row(s) on the record, 4 of them added by this capture from 2 transcript(s)"; line != want {
		t.Errorf("the reported line = %q, want %q", line, want)
	}

	// Re-running capture on unchanged transcripts must not double the run's measured cost, and
	// says the record still holds every row.
	if line, want := ingestSeatTurns(run, files), "seat turns: 4 row(s) on the record, 0 of them added by this capture from 2 transcript(s)"; line != want {
		t.Errorf("a re-run reported %q, want %q", line, want)
	}
	if got := seatTurnRows(t, run); got != 4 {
		t.Errorf("a re-run left %d rows, want 4", got)
	}
}

// A TRANSCRIPT THAT COULD NOT BE INGESTED IS NAMED IN THE COUNT, because the alternative is a
// smaller number that looks like a smaller run.
func TestIngestSeatTurnsNamesWhatItCouldNotRead(t *testing.T) {
	run, dir, one := ingestFixture(t, map[string]string{"agent-one.jsonl": twoTurns})

	line := ingestSeatTurns(run, append(one, record.Transcript{AgentID: "missing", Path: filepath.Join(dir, "agent-missing.jsonl")}))
	if !strings.Contains(line, "NOT INGESTED") {
		t.Errorf("an unreadable transcript was swallowed: %q", line)
	}
	if !strings.Contains(line, "2 row(s) on the record, 2 of them added") {
		t.Errorf("the readable transcript's turns are missing from %q", line)
	}
}

// A TRANSCRIPT WITH NO USAGE TURNS IS NOT A FAILURE. A seat dispatched and cancelled before its
// first API round has nothing to measure, and counting it as an ingest failure would send someone
// looking for a broken file.
func TestIngestSeatTurnsTreatsATurnlessTranscriptAsEmptyNotBroken(t *testing.T) {
	run, _, files := ingestFixture(t, map[string]string{
		"agent-quiet.jsonl": `{"agentId":"Q","type":"user","message":{"content":[{"type":"text"}]}}` + "\n",
	})
	line := ingestSeatTurns(run, files)
	if strings.Contains(line, "NOT INGESTED") {
		t.Errorf("a transcript with no API rounds was reported as a failure: %q", line)
	}
	if got := seatTurnRows(t, run); got != 0 {
		t.Errorf("got %d rows from a transcript with no usage", got)
	}
}

// THE TURNS A SITTING'S CLOSE FILED ARE COUNTED. The close ingests a seat's transcript, so by capture
// the record holds its rows and capture adds none; the line printed what capture added, and read
// "0 row(s)" on a record holding 596 (universe-m33, m34) — a zero indistinguishable from a run
// whose seats took no turns.
func TestCaptureCountsTheTurnsTheSittingCloseFiled(t *testing.T) {
	run, _, files := ingestFixture(t, map[string]string{"agent-one.jsonl": twoTurns})
	agentID, turns := seatturn.Parse(twoTurns)
	if n, err := record.AppendSeatTurns(run, agentID, turns); err != nil || n != 2 {
		t.Fatalf("filing the close's turns: %d, %v", n, err)
	}
	if line, want := ingestSeatTurns(run, files), "seat turns: 2 row(s) on the record, 0 of them added by this capture from 1 transcript(s)"; line != want {
		t.Errorf("capture after the close filed its turns reported %q, want %q", line, want)
	}
}
