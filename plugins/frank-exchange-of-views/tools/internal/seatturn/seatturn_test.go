package seatturn

import (
	"strings"
	"testing"
)

// A transcript in the shape the harness actually writes: assistant rounds carrying usage,
// interleaved with the records that do not.
const fixture = `{"agentId":"a1","type":"user","timestamp":"2026-09-03T03:00:00.000Z","message":{"content":[{"type":"text"}]}}
{"agentId":"a1","timestamp":"2026-09-03T03:00:10.500Z","message":{"model":"claude-opus-4-1","content":[{"type":"thinking"}],"usage":{"input_tokens":100,"output_tokens":3,"cache_read_input_tokens":9000,"cache_creation_input_tokens":50}}}
not json at all
{"agentId":"a1","timestamp":"2026-09-03T03:00:40.000Z","message":{"model":"claude-opus-4-1","content":[{"type":"thinking"},{"type":"tool_use"}],"usage":{"input_tokens":120,"output_tokens":400,"cache_read_input_tokens":9500,"cache_creation_input_tokens":0}}}
{"agentId":"a1","type":"user","timestamp":"2026-09-03T03:00:41.000Z","toolUseResult":{"ok":true}}
{"agentId":"a1","message":{"model":"claude-opus-4-1","content":[{"type":"text"}],"usage":{"input_tokens":10,"output_tokens":20,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
`

func TestParseReadsTheUsageBearingTurns(t *testing.T) {
	agent, turns := Parse(fixture)
	if agent != "a1" {
		t.Errorf("agent = %q, want a1", agent)
	}
	if len(turns) != 3 {
		t.Fatalf("got %d turns, want 3 (only the rounds reporting usage)", len(turns))
	}

	// Index is position among USAGE turns, not among lines — the user records and the
	// unparseable line between them must not advance it, or turn numbers drift from what the
	// seat actually did.
	for i, tn := range turns {
		if tn.Index != i {
			t.Errorf("turn %d has Index %d", i, tn.Index)
		}
	}

	first := turns[0]
	if first.Input != 100 || first.Output != 3 || first.CacheRead != 9000 || first.CacheCreation != 50 {
		t.Errorf("first turn's usage read wrong: %+v", first)
	}
	if first.Model != "claude-opus-4-1" {
		t.Errorf("model = %q", first.Model)
	}
	if !first.Thinking || first.Tool {
		t.Errorf("a turn carrying only a thinking block: Thinking=%v Tool=%v", first.Thinking, first.Tool)
	}
	if !turns[1].Thinking || !turns[1].Tool {
		t.Errorf("a turn carrying BOTH blocks must report both: %+v", turns[1])
	}
	if turns[2].Thinking || turns[2].Tool {
		t.Errorf("a text-only turn is neither thinking nor tool: %+v", turns[2])
	}
}

// THE SHAPE IS READ, NOT INFERRED. The F11 decomposition inferred a thinking turn from a low
// output_tokens with a long span. This turn has three output tokens and no thinking block: under
// the inference it is thinking, and it is not.
func TestParseDoesNotInferThinkingFromTokenCounts(t *testing.T) {
	const line = `{"agentId":"a1","timestamp":"2026-09-03T03:00:00Z","message":{"model":"m","content":[{"type":"tool_use"}],"usage":{"input_tokens":1,"output_tokens":3,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
	_, turns := Parse(line)
	if len(turns) != 1 {
		t.Fatalf("got %d turns", len(turns))
	}
	if turns[0].Thinking {
		t.Error("a 3-output-token tool turn was called thinking — that is the correlation, not the fact")
	}
}

// A MISSING TIMESTAMP IS ZERO AND MUST STAY DISTINGUISHABLE. Parsing it as the epoch would put the
// seat's start in 1970 and make every span computed over it the age of the universe.
func TestParseLeavesAnAbsentTimestampAtZero(t *testing.T) {
	for _, line := range []string{
		`{"agentId":"a1","message":{"model":"m","usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`,
		`{"agentId":"a1","timestamp":"not a time","message":{"model":"m","usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`,
	} {
		_, turns := Parse(line)
		if len(turns) != 1 {
			t.Fatalf("got %d turns", len(turns))
		}
		if turns[0].TSMillis != 0 {
			t.Errorf("absent/unparseable timestamp became %d, want 0", turns[0].TSMillis)
		}
	}
}

func TestParseReadsTheTimestamp(t *testing.T) {
	_, turns := Parse(fixture)
	// 2026-09-03T03:00:10.500Z
	if got := turns[0].TSMillis; got != 1788404410500 {
		t.Errorf("TSMillis = %d, want 1788404410500", got)
	}
	if turns[1].TSMillis-turns[0].TSMillis != 29500 {
		t.Errorf("the gap between the first two turns is %dms, want 29500", turns[1].TSMillis-turns[0].TSMillis)
	}
}

// THE AGENT COMES OFF THE LINE. A transcript whose file was renamed or copied still files its
// turns against the agent that produced them.
func TestParseTakesTheAgentFromTheRecordNotTheCaller(t *testing.T) {
	agent, _ := Parse(strings.ReplaceAll(fixture, `"agentId":"a1"`, `"agentId":"renamed-file-agent"`))
	if agent != "renamed-file-agent" {
		t.Errorf("agent = %q", agent)
	}
}

func TestParseOfNothingIsNothing(t *testing.T) {
	agent, turns := Parse("")
	if agent != "" || len(turns) != 0 {
		t.Errorf("empty transcript gave agent=%q turns=%d", agent, len(turns))
	}
}
