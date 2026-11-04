package trajectory

import (
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
	// A session with no tool calls, but workspace mentioned in prompt
	input := `{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","created_at":"2026-11-04T00:00:00Z","content":"Working on /home/gblock_ctoforaday_com/projects/special-circumstances/README.md"}
`
	proj := Project(strings.NewReader(input), 0)
	if len(proj.CWDs) == 0 {
		t.Fatalf("expected CWDs to be extracted from prompt path")
	}
	if !strings.Contains(proj.CWDs[0], "special-circumstances") {
		t.Errorf("expected CWD to be special-circumstances repo root, got %s", proj.CWDs[0])
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
