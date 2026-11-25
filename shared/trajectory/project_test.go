package trajectory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectAntigravity(t *testing.T) {
	proj := Project(strings.NewReader(agyShape), 10)
	if len(proj.Words) != 1 {
		t.Errorf("Words = %d, want 1", len(proj.Words))
	}
	if len(proj.Thoughts) != 1 {
		t.Errorf("Thoughts = %d, want 1", len(proj.Thoughts))
	}
	if len(proj.Acts) != 2 {
		t.Fatalf("Acts = %d, want 2", len(proj.Acts))
	}

	// Act 1: read_file
	a1 := proj.Acts[0]
	if a1.Seq != 10 || a1.Tool != "Read" || a1.Target != "/w/notes.txt" || a1.Outcome != OutcomeOK {
		t.Errorf("Act[0] = %+v, want Seq=10, Tool=Read, Target=/w/notes.txt, Outcome=ok", a1)
	}

	// Act 2: run_command
	a2 := proj.Acts[1]
	if a2.Seq != 11 || a2.Tool != "Bash" || a2.Target != "go test ./..." || a2.Outcome != OutcomeError {
		t.Errorf("Act[1] = %+v, want Seq=11, Tool=Bash, Target='go test ./...', Outcome=error", a2)
	}
}

func TestProjectAntigravityCWDFromPrompt(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tempDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Escape backslashes for JSON on Windows
	escapedPath := strings.ReplaceAll(testFile, `\`, `\\`)
	input := fmt.Sprintf(`{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-04T00:00:00Z","content":"Working on %s"}
`, escapedPath)

	proj := Project(strings.NewReader(input), 0)
	if len(proj.CWDs) == 0 {
		t.Fatalf("expected CWDs to be extracted from prompt path")
	}
	if filepath.Clean(proj.CWDs[0]) != filepath.Clean(tempDir) {
		t.Errorf("expected CWD to be %s, got %s", tempDir, proj.CWDs[0])
	}
}

func TestProjectClaude(t *testing.T) {
	proj := Project(strings.NewReader(claudeShape), 0)
	if len(proj.Acts) != 3 {
		t.Fatalf("Acts = %d, want 3", len(proj.Acts))
	}
	if proj.Acts[0].Tool != "Bash" || proj.Acts[0].Outcome != OutcomeOK {
		t.Errorf("Act[0] = %+v, want Bash ok", proj.Acts[0])
	}
	if proj.Acts[1].Tool != "Bash" || proj.Acts[1].Outcome != OutcomeUnresolved {
		t.Errorf("Act[1] = %+v, want Bash unresolved", proj.Acts[1])
	}
	if proj.Acts[2].Tool != "Read" || proj.Acts[2].Outcome != OutcomeUnresolved {
		t.Errorf("Act[2] = %+v, want Read unresolved", proj.Acts[2])
	}
}

func TestProjectAntigravitySystemMessageAndToolCallID(t *testing.T) {
	input := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-14T18:00:00Z","content":"Run task"}
{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","created_at":"2026-11-14T18:00:01Z","tool_calls":[{"id":"call_agy_154","name":"run_command","args":{"CommandLine":"echo ok"}}]}
{"step_index":2,"source":"TOOL","type":"GENERIC","status":"DONE","created_at":"2026-11-14T18:00:02Z","content":"ok"}
{"step_index":3,"source":"SYSTEM","type":"SYSTEM_MESSAGE","status":"DONE","created_at":"2026-11-14T18:00:03Z","content":"Task finished"}
`
	proj := Project(strings.NewReader(input), 0)
	if len(proj.Words) != 2 {
		t.Fatalf("Words = %d, want 2", len(proj.Words))
	}
	if proj.Words[1].OriginKind != "task-notification" || proj.Words[1].Role != "system" {
		t.Errorf("Words[1] = %+v, want OriginKind=task-notification Role=system", proj.Words[1])
	}
	if len(proj.Acts) != 1 || proj.Acts[0].Outcome != OutcomeOK {
		t.Errorf("Acts = %+v, want 1 Act with OutcomeOK", proj.Acts)
	}

	res, err := Parse(strings.NewReader(input), "transcript.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Events) != 2 {
		t.Fatalf("Parse events = %d, want 2", len(res.Events))
	}
	if res.Events[0].ToolUseID != "call_agy_154" || res.Events[1].ToolUseID != "call_agy_154" {
		t.Errorf("ToolUseID = (%q, %q), want call_agy_154", res.Events[0].ToolUseID, res.Events[1].ToolUseID)
	}
}
