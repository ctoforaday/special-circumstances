package trajectory

import (
	"strings"
	"testing"
)

const claudeShape = `{"type":"user","uuid":"u-1","timestamp":"2026-07-28T13:20:00Z","message":{"role":"user","content":"do the thing"}}
{"type":"assistant","uuid":"a-1","parentUuid":"u-1","timestamp":"2026-07-28T13:20:05Z","message":{"content":[{"type":"thinking","thinking":"","signature":"xx"},{"type":"tool_use","id":"toolu_1","name":"Bash","input":{"command":"REC=./feov-record ; \"$REC\" finding --seat-id red-1","description":"aliased"}}]}}
{"type":"user","uuid":"u-2","parentUuid":"a-1","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_1","content":"ok"}]}}
{"type":"assistant","uuid":"a-2","parentUuid":"u-2","timestamp":"2026-07-28T13:20:11Z","attributionAgent":"red-auditor","message":{"content":[{"type":"tool_use","id":"toolu_2","name":"Bash","input":{"command":"./feov-record merge --seat-id red-1"}}]}}
{"type":"assistant","uuid":"a-3","parentUuid":"a-2","timestamp":"2026-07-28T13:20:20Z","message":{"content":[{"type":"tool_use","id":"toolu_3","name":"Read","input":{"file_path":"/w/notes.txt"}}]}}
{"type":"system","uuid":"s-1","subtype":"marker"}
`

const agyShape = `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-04T07:26:20Z","content":"inspect the codebase"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T07:26:21Z","input_tokens":100,"output_tokens":50,"thinking":"Let's check notes.txt first.","tool_calls":[{"name":"read_file","args":{"Path":"/w/notes.txt","StartLine":"1","EndLine":"100"}}]}
{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","created_at":"2026-11-04T07:26:22Z","content":"file content here"}
{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-04T07:26:23Z","tool_calls":[{"name":"run_command","args":{"CommandLine":"go test ./...","WaitMsBeforeAsync":5000}}]}
{"step_index":4,"source":"MODEL","type":"GENERIC","status":"ERROR","error":"exit status 1","created_at":"2026-11-04T07:26:24Z","content":"test failure"}
`

func TestParseClaude(t *testing.T) {
	res, err := Parse(strings.NewReader(claudeShape), "/w/claude.jsonl")
	if err != nil {
		t.Fatalf("Parse(claude): %v", err)
	}
	if res.Lines != 6 {
		t.Errorf("Lines = %d, want 6", res.Lines)
	}
	uses := res.ToolUses()
	if len(uses) != 3 {
		t.Fatalf("ToolUses = %d, want 3", len(uses))
	}
	if uses[0].ToolName != "Bash" || uses[0].ToolUseID != "toolu_1" {
		t.Errorf("use[0] = %v, want Bash toolu_1", uses[0])
	}
	if uses[1].AgentType != "red-auditor" {
		t.Errorf("use[1].AgentType = %q, want red-auditor", uses[1].AgentType)
	}
	if uses[2].ToolName != "Read" {
		t.Errorf("use[2].ToolName = %q, want Read", uses[2].ToolName)
	}

	for _, e := range res.Events {
		if !e.Cited() {
			t.Errorf("event not cited: %v", e)
		}
	}
}

func TestParseAntigravity(t *testing.T) {
	res, err := Parse(strings.NewReader(agyShape), "/w/agy.jsonl")
	if err != nil {
		t.Fatalf("Parse(agy): %v", err)
	}
	if res.Lines != 5 {
		t.Errorf("Lines = %d, want 5", res.Lines)
	}

	uses := res.ToolUses()
	if len(uses) != 2 {
		t.Fatalf("ToolUses = %d, want 2", len(uses))
	}

	// First call: read_file normalized to Read
	if uses[0].ToolName != "Read" || uses[0].RawToolName != "read_file" {
		t.Errorf("uses[0] tool name = %q (raw %q), want Read (raw read_file)", uses[0].ToolName, uses[0].RawToolName)
	}
	if p := WriteTarget(uses[0]); p != "" {
		t.Errorf("WriteTarget on Read returned %q, want empty", p)
	}

	// Second call: run_command normalized to Bash
	if uses[1].ToolName != "Bash" || uses[1].RawToolName != "run_command" {
		t.Errorf("uses[1] tool name = %q (raw %q), want Bash (raw run_command)", uses[1].ToolName, uses[1].RawToolName)
	}
	if cmd := BashCommand(uses[1]); cmd != "go test ./..." {
		t.Errorf("BashCommand(uses[1]) = %q, want 'go test ./...'", cmd)
	}

	// Check tool results
	var results []Event
	for _, e := range res.Events {
		if e.Kind == "tool_result" {
			results = append(results, e)
		}
	}
	if len(results) != 2 {
		t.Fatalf("tool_result count = %d, want 2", len(results))
	}
	if results[0].IsError {
		t.Errorf("results[0] IsError = true, want false")
	}
	if !results[1].IsError {
		t.Errorf("results[1] IsError = false, want true")
	}
}

func TestParseAllAntigravity(t *testing.T) {
	events, err := ParseAll(strings.NewReader(agyShape), "/w/agy.jsonl")
	if err != nil {
		t.Fatalf("ParseAll: %v", err)
	}
	var kinds []string
	for _, e := range events {
		kinds = append(kinds, e.Kind)
	}
	// Expected: user_text, assistant_text (thinking), tool_use, tool_result, tool_use, tool_result
	wantKinds := []string{"user_text", "assistant_text", "tool_use", "tool_result", "tool_use", "tool_result"}
	if len(kinds) != len(wantKinds) {
		t.Fatalf("got kinds %v, want %v", kinds, wantKinds)
	}
	for i := range kinds {
		if kinds[i] != wantKinds[i] {
			t.Errorf("[%d] got kind %q, want %q", i, kinds[i], wantKinds[i])
		}
	}
}

func TestUnparseableLines(t *testing.T) {
	body := "not a json line\n" + claudeShape + "{\"broken json\n"
	res, err := Parse(strings.NewReader(body), "/w/broken.jsonl")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if res.BadLines != 2 {
		t.Errorf("BadLines = %d, want 2", res.BadLines)
	}
	if len(res.BadSample) != 2 || res.BadSample[0] != 1 || res.BadSample[1] != 8 {
		t.Errorf("BadSample = %v, want [1, 8]", res.BadSample)
	}
	if len(res.ToolUses()) != 3 {
		t.Errorf("ToolUses = %d, want 3", len(res.ToolUses()))
	}
}
