package trajectory

import (
	"strings"
	"testing"
)

// Shape taken from a real Claude Code transcript produced by driving a live
// session; only the paths and ids are shortened. The two Bash commands are
// verbatim from that run, including the aliased one.
const realShape = `{"type":"user","uuid":"u-1","timestamp":"2026-07-28T13:20:00Z","message":{"role":"user","content":"do the thing"}}
{"type":"assistant","uuid":"a-1","parentUuid":"u-1","timestamp":"2026-07-28T13:20:05Z","message":{"content":[{"type":"thinking","thinking":"","signature":"xx"},{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"REC=./feov-record ; \"$REC\" finding --seat-id red-1","description":"aliased"}}]}}
{"type":"user","uuid":"u-2","parentUuid":"a-1","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"ok"}]}}
{"type":"assistant","uuid":"a-2","parentUuid":"u-2","timestamp":"2026-07-28T13:20:11Z","attributionAgent":"red-auditor","message":{"content":[{"type":"tool_use","id":"toolu_2","name":"Bash","input":{"command":"./feov-record merge --seat-id red-1"}}]}}
{"type":"assistant","uuid":"a-3","parentUuid":"a-2","timestamp":"2026-07-28T13:20:20Z","message":{"content":[{"type":"tool_use","id":"toolu_3","name":"Read","input":{"file_path":"/w/notes.txt"}}]}}
{"type":"system","uuid":"s-1","subtype":"marker"}
`

func parse(t *testing.T, body string) Result {
	t.Helper()
	res, err := Parse(strings.NewReader(body), "/w/session.jsonl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return res
}

// Provenance is the whole point: an event that cannot be cited is an opinion.
func TestEveryEventCarriesProvenance(t *testing.T) {
	res := parse(t, realShape)
	if len(res.Events) == 0 {
		t.Fatal("parsed nothing from a real-shaped transcript")
	}
	for i, e := range res.Events {
		if !e.Cited() {
			t.Errorf("event %d cannot be cited: file=%q line=%d uuid=%q", i, e.File, e.Line, e.UUID)
		}
	}
}

func TestParsesToolUsesWithLineNumbers(t *testing.T) {
	res := parse(t, realShape)
	uses := res.ToolUses()
	if len(uses) != 3 {
		t.Fatalf("tool_use count = %d, want 3", len(uses))
	}
	want := []struct {
		line int
		name string
		id   string
	}{
		{2, "Bash", "toolu_1"},
		{4, "Bash", "toolu_2"},
		{5, "Read", "toolu_3"},
	}
	for i, w := range want {
		if uses[i].Line != w.line || uses[i].ToolName != w.name || uses[i].ToolUseID != w.id {
			t.Errorf("[%d] got line=%d name=%s id=%s, want line=%d name=%s id=%s",
				i, uses[i].Line, uses[i].ToolName, uses[i].ToolUseID, w.line, w.name, w.id)
		}
	}
	// attributionAgent identifies the seat when present, and is empty in a main session.
	if uses[1].AgentType != "red-auditor" {
		t.Errorf("lost the seat attribution: %q", uses[1].AgentType)
	}
	if uses[0].AgentType != "" {
		t.Errorf("invented an attribution where the transcript had none: %q", uses[0].AgentType)
	}
}

// A result is linked to its call, which is what makes act-vs-claim possible.
func TestToolResultsLinkToTheirCall(t *testing.T) {
	res := parse(t, realShape)
	var linked int
	for _, e := range res.Events {
		if e.Kind == "tool_result" && e.ToolUseID == "toolu_1" {
			linked++
		}
	}
	if linked != 1 {
		t.Fatalf("tool_result linked to toolu_1 = %d, want 1", linked)
	}
}

// A reader that silently drops input answers confidently from a subset.
func TestUnparseableLinesAreCountedNotSwallowed(t *testing.T) {
	body := realShape + "{not json\n" + `{"type":"assistant","uuid":"a-9","message":{"content":[{"type":"tool_use","id":"t9","name":"Glob","input":{}}]}}` + "\n"
	res := parse(t, body)
	if res.BadLines != 1 {
		t.Fatalf("BadLines = %d, want 1", res.BadLines)
	}
	if len(res.BadSample) != 1 || res.BadSample[0] != 7 {
		t.Errorf("BadSample = %v, want [7]", res.BadSample)
	}
	// Parsing continues past the bad line rather than stopping at it.
	if got := len(res.ToolUses()); got != 4 {
		t.Errorf("tool_use count after a bad line = %d, want 4", got)
	}
}

// Thinking blocks are structurally present and empty by default; they must not
// become events, and must never be mistaken for acts.
func TestThinkingBlocksAreNotEvents(t *testing.T) {
	res := parse(t, realShape)
	for _, e := range res.Events {
		if e.Kind != "tool_use" && e.Kind != "tool_result" {
			t.Errorf("unexpected event kind %q — only acts are events", e.Kind)
		}
	}
}

// A transcript line whose content is a plain string (normal for user lines) is
// not a parse failure.
func TestStringContentIsNotABadLine(t *testing.T) {
	res := parse(t, realShape)
	if res.BadLines != 0 {
		t.Fatalf("BadLines = %d on clean input, want 0 (sample %v)", res.BadLines, res.BadSample)
	}
	if res.Lines != 6 {
		t.Errorf("Lines = %d, want 6", res.Lines)
	}
}
