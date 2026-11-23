package hookcore

import (
	"fmt"
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
	if gotKey := DrainScopeKey(req); gotKey != "conv-1" {
		t.Fatalf("DrainScopeKey(req) = %q, want conv-1", gotKey)
	}
	got, err := DrainFailedToolSteps(req)
	if err != nil {
		t.Fatalf("first drain error: %v", err)
	}
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
	if again, err := DrainFailedToolSteps(req); err != nil || len(again) != 0 {
		t.Fatalf("second drain = %+v err=%v, want empty, nil", again, err)
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

	got2, err := DrainFailedToolSteps(req)
	if err != nil {
		t.Fatalf("third drain error: %v", err)
	}
	if len(got2) != 1 || got2[0].StepIndex != 6 || got2[0].ExitCode != 2 || got2[0].Command != "exit 2" {
		t.Fatalf("third drain = %+v, want step 6 exit 2", got2)
	}
}

func TestDrainFailedToolStepsDoesNotSkipIncompleteTrailingLine(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-partial",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}

	part1 := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"false"}}]}` + "\n" +
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code `
	if err := os.WriteFile(tp, []byte(part1), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := DrainFailedToolSteps(req)
	if err != nil || len(got) != 0 {
		t.Fatalf("partial drain = %+v err=%v, want empty, nil", got, err)
	}

	// Finish writing the partial line with its trailing newline.
	f, err := os.OpenFile(tp, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`1.\n"}` + "\n")
	_ = f.Close()

	got, err = DrainFailedToolSteps(req)
	if err != nil {
		t.Fatalf("completed drain error: %v", err)
	}
	if len(got) != 1 || got[0].StepIndex != 2 || got[0].ExitCode != 1 || got[0].Command != "false" {
		t.Fatalf("completed drain = %+v, want step 2 exit 1 cmd=false", got)
	}
}

func TestDrainFailedToolStepsStraddlingPlannerResponseAndGeneric(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-straddle",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}

	line1 := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"python3 fail.py"}}]}` + "\n"
	if err := os.WriteFile(tp, []byte(line1), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := DrainFailedToolSteps(req)
	if err != nil || len(got) != 0 {
		t.Fatalf("first drain = %+v err=%v, want empty, nil", got, err)
	}

	line2 := `{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 9.\n\nCommand output:\nboom\n"}` + "\n"
	f, err := os.OpenFile(tp, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(line2)
	_ = f.Close()

	got, err = DrainFailedToolSteps(req)
	if err != nil {
		t.Fatalf("second drain error: %v", err)
	}
	if len(got) != 1 || got[0].StepIndex != 2 || got[0].ToolName != "Bash" || got[0].ExitCode != 9 || got[0].Command != "python3 fail.py" {
		t.Fatalf("straddle drain = %+v, want step 2 Bash exit 9 cmd='python3 fail.py'", got)
	}
}

func TestDrainFailedToolStepsErrorsOnUnrecognizedBashGeneric(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	transcript := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"false"}}]}` + "\n" +
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"exited with status 1\n"}` + "\n"
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-drift",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}
	got, err := DrainFailedToolSteps(req)
	if err == nil {
		t.Fatalf("expected error on unrecognized Bash GENERIC prose, got steps=%+v err=nil", got)
	}
}

func TestDrainFailedToolStepsHandlesLineOver10MB(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	big := strings.Repeat("a", 10*1024*1024)
	transcript := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"bigcmd"}}]}` + "\n" +
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 3.\n\nCommand output:\n` + big + `\n"}` + "\n"
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-big",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}
	got, err := DrainFailedToolSteps(req)
	if err != nil {
		t.Fatalf("10MB line drain error: %v", err)
	}
	if len(got) != 1 || got[0].ExitCode != 3 || got[0].Command != "bigcmd" {
		t.Fatalf("10MB line drain = %+v, want exit 3 cmd=bigcmd", got)
	}
}

func TestDrainFailedToolStepsRecognizesOctoberFormatAndStepExecErrors(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	transcript := strings.Join([]string{
		`{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"missing-cmd"}}]}`,
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"The command exited with code 127.\n\nCommand output:\nnot found\n"}`,
		`{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"hook-err-cmd"}}]}`,
		`{"step_index":4,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Encountered error in step execution: failed to unmarshal result from hook jsonhook__sc_PreToolUse\n"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-oct",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}
	got, err := DrainFailedToolSteps(req)
	if err != nil {
		t.Fatalf("unexpected drain error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d failed steps (%+v), want 2 (October exit 127 + hook unmarshal error)", len(got), got)
	}
	if got[0].StepIndex != 2 || got[0].ExitCode != 127 || got[0].Command != "missing-cmd" {
		t.Fatalf("step 0 = %+v, want step 2 exit 127", got[0])
	}
	if got[1].StepIndex != 4 || got[1].Command != "hook-err-cmd" || !strings.Contains(got[1].Output, "failed to unmarshal result from hook") {
		t.Fatalf("step 1 = %+v, want step 4 hook-unmarshal failure", got[1])
	}

	// Unknown "Encountered error in step execution:" form must NOT be suppressed silently.
	f, err := os.OpenFile(tp, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"step_index":5,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"other"}}]}` + "\n" +
		`{"step_index":6,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Encountered error in step execution: unknown mystery\n"}` + "\n")
	_ = f.Close()
	if _, err := DrainFailedToolSteps(req); err == nil {
		t.Fatalf("unknown step-execution error was silently suppressed, want non-nil error")
	}
}

func TestDrainFailedToolStepsRefusesSilentReplayOnCorruptOrShrunkenOrUnwritableWatermark(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	pluginData := filepath.Join(dir, "plugin_data")
	transcript := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"exit 1"}}]}` + "\n" +
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 1.\n"}` + "\n"
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-replay",
		TranscriptPath: tp,
		PluginDataDir:  pluginData,
	}
	got, err := DrainFailedToolSteps(req)
	if err != nil || len(got) != 1 {
		t.Fatalf("initial drain = %+v err=%v, want 1, nil", got, err)
	}

	// 1. Corrupt watermark errors once, rebases to EOF (no replay), and subsequent drain succeeds cleanly.
	wmPath := filepath.Join(pluginData, "last_failed_step_conv-replay.json")
	if err := os.WriteFile(wmPath, []byte("{not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	replayed, err := DrainFailedToolSteps(req)
	if err == nil || len(replayed) != 0 {
		t.Fatalf("corrupt watermark: got steps=%+v err=%v, want empty + non-nil error", replayed, err)
	}
	if after, err := DrainFailedToolSteps(req); err != nil || len(after) != 0 {
		t.Fatalf("drain after corrupt rebase: got steps=%+v err=%v, want empty, nil", after, err)
	}

	// 2. Shrunken transcript errors once, rebases to EOF (no replay), and subsequent drain succeeds cleanly.
	if err := os.WriteFile(tp, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	replayed, err = DrainFailedToolSteps(req)
	if err == nil || len(replayed) != 0 {
		t.Fatalf("shrunken transcript: got steps=%+v err=%v, want empty + non-nil error", replayed, err)
	}
	if after, err := DrainFailedToolSteps(req); err != nil || len(after) != 0 {
		t.Fatalf("drain after shrink rebase: got steps=%+v err=%v, want empty, nil", after, err)
	}

	// 3. Misaligned rewrite (file rewritten to >= old size so byte at offset-1 != '\n') errors once and rebases.
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	reqAlign := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-align",
		TranscriptPath: tp,
		PluginDataDir:  pluginData,
	}
	if _, err := DrainFailedToolSteps(reqAlign); err != nil {
		t.Fatal(err)
	}
	// Overwrite with longer content where the byte at old offset-1 is 'x', not '\n'.
	rewritten := strings.Repeat("x", len(transcript)+20) + "\n"
	if err := os.WriteFile(tp, []byte(rewritten), 0o600); err != nil {
		t.Fatal(err)
	}
	if gotAlign, err := DrainFailedToolSteps(reqAlign); err == nil || len(gotAlign) != 0 {
		t.Fatalf("misaligned rewrite: got steps=%+v err=%v, want empty + non-nil error", gotAlign, err)
	}
	if after, err := DrainFailedToolSteps(reqAlign); err != nil || len(after) != 0 {
		t.Fatalf("drain after misaligned rebase: got steps=%+v err=%v, want empty, nil", after, err)
	}

	// 4. Unwritable watermark tmp path returns 0 steps + non-nil error on BOTH drains (no replay).
	req2 := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-unwritable",
		TranscriptPath: tp,
		PluginDataDir:  pluginData,
	}
	if err := os.WriteFile(tp, []byte(transcript), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pluginData, "last_failed_step_conv-unwritable.json.tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		steps, err := DrainFailedToolSteps(req2)
		if err == nil || len(steps) != 0 {
			t.Fatalf("unwritable watermark drain %d: got steps=%+v err=%v, want 0 steps + error", i, steps, err)
		}
	}
}

func TestDrainFailedToolStepsCapsUnmatchedWarnings(t *testing.T) {
	dir := t.TempDir()
	tp := filepath.Join(dir, "transcript_full.jsonl")
	var lines []string
	for i := 0; i < 5; i++ {
		lines = append(lines,
			fmt.Sprintf(`{"step_index":%d,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"false"}}]}`, 1+2*i),
			fmt.Sprintf(`{"step_index":%d,"source":"MODEL","type":"GENERIC","status":"DONE","content":"unrecognized status\n"}`, 2+2*i),
		)
	}
	if err := os.WriteFile(tp, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	req := Request{
		Engine:         EngineAntigravity,
		Event:          "PreInvocation",
		SessionID:      "conv-cap",
		TranscriptPath: tp,
		PluginDataDir:  filepath.Join(dir, "plugin_data"),
	}
	_, err := DrainFailedToolSteps(req)
	if err == nil || !strings.Contains(err.Error(), "+2 more, 5 unmatched total") {
		t.Fatalf("capped warning err = %v, want '+2 more, 5 unmatched total'", err)
	}
}
