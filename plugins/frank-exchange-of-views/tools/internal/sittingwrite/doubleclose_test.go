package sittingwrite

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

const lens = "frank-exchange-of-views:red-lens-logic"

// envelopeTurn is the StructuredOutput turn a nudged seat takes after its first stop.
const envelopeTurn = `{"agentId":"AG","timestamp":"2026-09-03T03:00:40.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"tool_use"}],"usage":{"input_tokens":120,"output_tokens":40,"cache_read_input_tokens":990,"cache_creation_input_tokens":0}}}
`

// A SEAT NUDGED TO RETURN ITS ENVELOPE STOPS TWICE, and its sitting is one span with one close
// (m34: 80 closes for 75 opens). The second stop still ingests the envelope turn, which the
// transcript holds only by then.
func TestTwoStopsOfOneSittingWriteOneClose(t *testing.T) {
	run := newRun(t)
	path := writeTranscript(t, oneTurn)
	for _, s := range []Sitting{
		{RunDir: run, Phase: Open, AgentID: "AG", AgentType: lens},
		{RunDir: run, Phase: Close, AgentID: "AG", AgentType: lens, TranscriptPath: path},
	} {
		if err := Write(s, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, []byte(oneTurn+envelopeTurn), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(Sitting{RunDir: run, Phase: Close, AgentID: "AG", AgentType: lens, TranscriptPath: path}, nil); err != nil {
		t.Fatalf("the second stop of one sitting = %v, want it to write nothing and succeed", err)
	}
	if o, c := countByType(t, run); o != 1 || c != 1 {
		t.Errorf("the record holds %d opens and %d closes, want 1 and 1", o, c)
	}
	if n, err := record.CountSeatTurns(mustRun(t, run)); err != nil || n != 3 {
		t.Errorf("seat_turn holds %d rows (err %v), want 3 — the envelope turn the second stop read", n, err)
	}
}

// TWO STOPS THAT DISAGREE ARE NOT ONE FACT: the second is refused naming both paths, and the tool
// does not pick one.
func TestAStopThatDisagreesWithTheCloseIsRefused(t *testing.T) {
	run := newRun(t)
	first, second := writeTranscript(t, oneTurn), writeTranscript(t, oneTurn)
	for _, s := range []Sitting{
		{RunDir: run, Phase: Open, AgentID: "AG", AgentType: lens},
		{RunDir: run, Phase: Close, AgentID: "AG", AgentType: lens, TranscriptPath: first},
	} {
		if err := Write(s, nil); err != nil {
			t.Fatal(err)
		}
	}
	err := Write(Sitting{RunDir: run, Phase: Close, AgentID: "AG", AgentType: lens, TranscriptPath: second}, nil)
	// Quoted as the refusal quotes them: a Windows path's backslashes are escaped there.
	if err == nil || !strings.Contains(err.Error(), strconv.Quote(first)) || !strings.Contains(err.Error(), strconv.Quote(second)) {
		t.Fatalf("a disagreeing second stop = %v, want a refusal naming %s and %s", err, first, second)
	}
	if _, c := countByType(t, run); c != 1 {
		t.Errorf("the record holds %d closes, want 1", c)
	}
}

// THE KEY IS THE SITTING, NOT THE AGENT: an agent opened again after its close has a new span, and
// that span's stop is its own close.
func TestASecondSittingOfOneAgentHasItsOwnClose(t *testing.T) {
	run := newRun(t)
	path := writeTranscript(t, oneTurn)
	for _, p := range []Phase{Open, Close, Open, Close} {
		if err := Write(Sitting{RunDir: run, Phase: p, AgentID: "AG", AgentType: lens, TranscriptPath: path}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if o, c := countByType(t, run); o != 2 || c != 2 {
		t.Errorf("the record holds %d opens and %d closes, want 2 and 2", o, c)
	}
}
