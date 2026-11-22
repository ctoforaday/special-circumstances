package hookcore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDrainFailedToolStepsExtractsNonZeroRunCommandOnce(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	transcript := strings.Join([]string{
		`{"step_index":0,"source":"USER_EXPLICIT","type":"USER_INPUT","status":"DONE","content":"hi"}`,
		`{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"\"echo ok\""}}]}`,
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 0.\n\nCommand output:\nok\n"}`,
		`{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"\"/no/such/bin --flag\""}}]}`,
		`{"step_index":4,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 127.\n\nCommand output:\nbash: /no/such/bin: No such file or directory\n"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}

	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-1",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}
	got := DrainFailedToolSteps(req)
	if len(got) != 1 {
		t.Fatalf("first drain = %+v, want 1 failed step", got)
	}
	if got[0].StepIndex != 4 || got[0].ToolName != "Bash" || got[0].ExitCode != 127 || got[0].Command != "/no/such/bin --flag" {
		t.Fatalf("unexpected failed step: %+v", got[0])
	}
	if !strings.Contains(got[0].Output, "No such file or directory") {
		t.Fatalf("unexpected output: %q", got[0].Output)
	}

	// Second drain with no new steps returns nothing (watermark prevents duplicates).
	if again := DrainFailedToolSteps(req); len(again) != 0 {
		t.Fatalf("second drain = %+v, want empty", again)
	}

	// Append another failed step with a PLAIN (unquoted) CommandLine string; third drain returns only the new one.
	extra := `{"step_index":5,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"exit 2"}}]}` + "\n" +
		`{"step_index":6,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 2.\n\nCommand output:\nbad arg\n"}` + "\n"
	f, err := os.OpenFile(tp, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(extra)
	_ = f.Close()

	got2 := DrainFailedToolSteps(req)
	if len(got2) != 1 || got2[0].StepIndex != 6 || got2[0].ExitCode != 2 || got2[0].Command != "exit 2" {
		t.Fatalf("third drain = %+v, want step 6 exit 2", got2)
	}
}
