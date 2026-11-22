package hookcore_test

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type asymmetryEvent struct {
	Event          string `json:"event"`
	HookEventName  string `json:"hook_event_name"`
	Engine         string `json:"engine"`
	SessionID      string `json:"session_id"`
	AgentID        string `json:"agent_id"`
	AgentType      string `json:"agent_type"`
	ToolName       string `json:"tool_name"`
	RawToolName    string `json:"raw_tool_name"`
	Command        string `json:"command"`
	Source         string `json:"source"`
	Error          string `json:"error"`
	HasResultField bool   `json:"has_result_field"`
	Raw            string `json:"raw"`
}

func buildAsymmetryProbe(t *testing.T, outPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(outPath), err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outPath, "./testdata/asymmetryprobe")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build asymmetryprobe: %v\n%s", err, out)
	}
}

func readAsymmetryEvents(t *testing.T, path string) []asymmetryEvent {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open asymmetry log %s: %v", path, err)
	}
	defer f.Close()
	var out []asymmetryEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 4<<20)
	for sc.Scan() {
		var ev asymmetryEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			t.Fatalf("decode asymmetry event %q: %v", sc.Text(), err)
		}
		out = append(out, ev)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan asymmetry log: %v", err)
	}
	return out
}

func claudeAsymmetryPlugin(t *testing.T, hooksJSON string, agents map[string]string) string {
	t.Helper()
	pluginDir := filepath.Join(t.TempDir(), "asymmetryprobe")
	claudeMeta := filepath.Join(pluginDir, ".claude-plugin")
	if err := os.MkdirAll(claudeMeta, 0o755); err != nil {
		t.Fatal(err)
	}
	buildAsymmetryProbe(t, filepath.Join(pluginDir, "bin", "asymmetryprobe"))
	manifest := `{"name":"asymmetryprobe","version":"0.0.1","description":"Harness asymmetry probe","hooks":"./.claude-plugin/hooks.json"}`
	if err := os.WriteFile(filepath.Join(claudeMeta, "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeMeta, "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if len(agents) > 0 {
		agentsDir := filepath.Join(pluginDir, "agents")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, body := range agents {
			if err := os.WriteFile(filepath.Join(agentsDir, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return pluginDir
}

func runClaudeAsymmetry(t *testing.T, workDir, pluginDir, logPath, resumeSession, prompt string, extraArgs ...string) string {
	t.Helper()
	claudeBin := findClaudeBin()
	if claudeBin == "" {
		t.Skip("claude CLI not found; skipping live Claude asymmetry test")
	}
	args := []string{"-p", prompt, "--model", "haiku", "--dangerously-skip-permissions", "--plugin-dir", pluginDir}
	if resumeSession != "" {
		args = append(args, "--resume", resumeSession)
	}
	args = append(args, extraArgs...)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, claudeBin, args...)
	cmd.Dir = workDir
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "CLAUDE_") || strings.HasPrefix(e, "FEOV_") {
			continue
		}
		env = append(env, e)
	}
	env = append(env, "ASYMMETRY_LOG="+logPath, "CLAUDE_PLUGIN_DATA="+t.TempDir())
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("claude -p failed: %v\nOutput:\n%s", err, out)
	}
	return string(out)
}

// TestLiveAsymmetrySessionStartOnResume pins the resume-hook asymmetry (#1926 item 1):
//   - Claim: Claude Code fires SessionStart(source=resume) on `--resume`; Antigravity does NOT
//     fire SessionStart on `--conversation`.
//   - Relying code:
//     plugins/prosthetic-conscience/tools/internal/sessionstart/main.go (PreInvocation resume restore),
//     plugins/gray-area/tools/cmd/gray-area-capture/link.go (PreInvocation stale-link detection),
//     docs/local-development.md.
func TestLiveAsymmetrySessionStartOnResume(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live asymmetry test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live asymmetry test under -short")
	}

	t.Run("Claude", func(t *testing.T) {
		hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
		pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
		workDir := t.TempDir()
		logPath := filepath.Join(t.TempDir(), "claude-resume.jsonl")

		runClaudeAsymmetry(t, workDir, pluginDir, logPath, "", "Reply with the single word OK.")
		turn1 := readAsymmetryEvents(t, logPath)
		var sessionID string
		for _, ev := range turn1 {
			if ev.Event == "SessionStart" && ev.SessionID != "" {
				sessionID = ev.SessionID
				break
			}
		}
		if sessionID == "" {
			t.Fatalf("Claude did not fire SessionStart on initial turn; events=%+v", turn1)
		}

		if err := os.Truncate(logPath, 0); err != nil {
			t.Fatal(err)
		}
		runClaudeAsymmetry(t, workDir, pluginDir, logPath, sessionID, "Reply with the single word OK2.")
		turn2 := readAsymmetryEvents(t, logPath)
		var sawResumeStart bool
		for _, ev := range turn2 {
			if ev.Event == "SessionStart" && ev.SessionID == sessionID && ev.Source == "resume" {
				sawResumeStart = true
				break
			}
		}
		if !sawResumeStart {
			t.Fatalf("Claude did NOT fire SessionStart(source=resume) on --resume %s; events=%+v", sessionID, turn2)
		}
	})

	t.Run("Antigravity", func(t *testing.T) {
		w := newAgyWorld(t)
		probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
		buildAsymmetryProbe(t, probeBin)
		w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
  "PreInvocation":[{"command":"./bin/asymmetryprobe -event PreInvocation"}]}}`)
		logPath := filepath.Join(t.TempDir(), "agy-resume.jsonl")

		w.run(t, "", "Reply with the single word OK.", "ASYMMETRY_LOG="+logPath)
		convs := w.conversations(t)
		if len(convs) != 1 {
			t.Fatalf("want 1 conversation, got %v", convs)
		}
		conv := convs[0]

		if err := os.Truncate(logPath, 0); err != nil {
			t.Fatal(err)
		}
		w.run(t, conv, "Reply with the single word OK2.", "ASYMMETRY_LOG="+logPath)
		turn2 := readAsymmetryEvents(t, logPath)
		var sawPreInv bool
		for _, ev := range turn2 {
			if ev.Event == "SessionStart" {
				t.Fatalf("ASSUMPTION CHANGED: Antigravity now fires SessionStart on --conversation resume (%+v); PreInvocation resume workarounds can be retired", ev)
			}
			if ev.Event == "PreInvocation" {
				sawPreInv = true
			}
		}
		if !sawPreInv {
			t.Fatalf("Antigravity resume did not fire PreInvocation either; events=%+v", turn2)
		}
	})
}

// TestLiveAsymmetryFailedShellCommandHook pins the failed-command reporting asymmetry (#1926 item 5):
//   - Claim: Claude Code fires PostToolUseFailure with a non-empty error on a non-zero Bash exit;
//     Antigravity fires no PostToolUseFailure and its PostToolUse payload carries error:"" and no
//     result field.
//   - Relying code:
//     shared/hookcore/failedsteps.go (DrainFailedToolSteps),
//     plugins/frank-exchange-of-views/tools/internal/hookcmd/agysitting.go (logAgyFailedToolSteps),
//     plugins/prosthetic-conscience/tools/internal/sessionstart/main.go (PreInvocation strike drain).
func TestLiveAsymmetryFailedShellCommandHook(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live asymmetry test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live asymmetry test under -short")
	}

	t.Run("Claude", func(t *testing.T) {
		hooksJSON := `{
  "hooks": {
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}],
    "PostToolUseFailure": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUseFailure"}]}]
  }
}`
		pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
		workDir := t.TempDir()
		logPath := filepath.Join(t.TempDir(), "claude-fail.jsonl")

		runClaudeAsymmetry(t, workDir, pluginDir, logPath, "",
			"Run the exact Bash command `echo ASYM_FAIL_PROBE && exit 2`. After it finishes, reply with DONE.")
		events := readAsymmetryEvents(t, logPath)
		var sawFailureEvent bool
		for _, ev := range events {
			if !strings.Contains(ev.Command, "ASYM_FAIL_PROBE") {
				continue
			}
			if ev.Event == "PostToolUseFailure" && ev.Error != "" {
				sawFailureEvent = true
			}
		}
		if !sawFailureEvent {
			t.Fatalf("Claude did not fire PostToolUseFailure with non-empty error for failing Bash command; events=%+v", events)
		}
	})

	t.Run("Antigravity", func(t *testing.T) {
		w := newAgyWorld(t)
		probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
		buildAsymmetryProbe(t, probeBin)
		w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PostToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUse"}]}],
  "PostToolUseFailure":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUseFailure"}]}]}}`)
		logPath := filepath.Join(t.TempDir(), "agy-fail.jsonl")

		w.run(t, "", "Run the exact shell command `echo ASYM_FAIL_PROBE && exit 2` via run_command. After it finishes, reply with DONE.", "ASYMMETRY_LOG="+logPath)
		events := readAsymmetryEvents(t, logPath)
		var sawPost bool
		for _, ev := range events {
			if ev.Event == "PostToolUseFailure" {
				t.Fatalf("ASSUMPTION CHANGED: Antigravity now fires PostToolUseFailure (%+v); hookcore.DrainFailedToolSteps can be retired", ev)
			}
			if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "ASYM_FAIL_PROBE") {
				sawPost = true
				if ev.Error != "" || ev.HasResultField {
					t.Fatalf("ASSUMPTION CHANGED: Antigravity PostToolUse on failing run_command now carries error=%q has_result=%v (raw=%s); hookcore.DrainFailedToolSteps can be retired",
						ev.Error, ev.HasResultField, ev.Raw)
				}
			}
		}
		if !sawPost {
			t.Fatalf("never observed Antigravity PostToolUse for ASYM_FAIL_PROBE; events=%+v", events)
		}
	})
}

// TestLiveAsymmetryWorkflowChildAttribution pins the subagent/workflow child attribution asymmetry:
//   - Claim: Claude Code's Agent tool fires SubagentStart/SubagentStop and carries agent_id on child
//     tool hooks; Antigravity's run_workflow child agents fire no SubagentStart/SubagentStop and
//     carry neither parentConversationId nor agentName on child tool hooks.
//   - Relying code:
//     plugins/frank-exchange-of-views/tools/internal/hookcmd/agysitting.go (.agy-sittings / --seat-id bind),
//     plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.py (seat registration).
func TestLiveAsymmetryWorkflowChildAttribution(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live asymmetry test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live asymmetry test under -short")
	}

	t.Run("Claude", func(t *testing.T) {
		hooksJSON := `{
  "hooks": {
    "SubagentStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SubagentStart"}]}],
    "SubagentStop": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SubagentStop"}]}],
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}]
  }
}`
		agentDef := "---\nname: asym-child\ndescription: Asymmetry child worker.\nmodel: haiku\ntools: Bash\n---\nRun the exact Bash command requested and return its output.\n"
		pluginDir := claudeAsymmetryPlugin(t, hooksJSON, map[string]string{"asym-child.md": agentDef})
		workDir := t.TempDir()
		logPath := filepath.Join(t.TempDir(), "claude-child.jsonl")

		runClaudeAsymmetry(t, workDir, pluginDir, logPath, "",
			"Spawn the subagent asymmetryprobe:asym-child via the Agent tool to run `echo ASYM_WF_CHILD` and return its output. Do NOT run Bash yourself.")
		events := readAsymmetryEvents(t, logPath)
		var sawStart, sawStop, sawChildPost bool
		for _, ev := range events {
			if ev.Event == "SubagentStart" && ev.AgentID != "" {
				sawStart = true
			}
			if ev.Event == "SubagentStop" && ev.AgentID != "" {
				sawStop = true
			}
			if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "ASYM_WF_CHILD") && ev.AgentID != "" {
				sawChildPost = true
			}
		}
		if !sawStart || !sawStop || !sawChildPost {
			t.Fatalf("Claude Agent child attribution incomplete: start=%v stop=%v childPost=%v; events=%+v", sawStart, sawStop, sawChildPost, events)
		}
	})

	t.Run("Antigravity", func(t *testing.T) {
		w := newAgyWorld(t)
		probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
		buildAsymmetryProbe(t, probeBin)
		w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PostToolUse":[{"matcher":".*","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUse"}]}],
  "SubagentStart":[{"command":"./bin/asymmetryprobe -event SubagentStart"}],
  "SubagentStop":[{"command":"./bin/asymmetryprobe -event SubagentStop"}]}}`)

		wfPath := filepath.Join(w.work, "wf_attr.py")
		wfSrc := `await agent(
    prompt="Run the shell command 'echo ASYM_WF_CHILD' via run_command, then finish with done=True.",
    schema={"type": "object", "properties": {"done": {"type": "boolean"}}, "required": ["done"]},
    role="Asymmetry Child",
    type_name="self",
    model="flash_lite",
)
print("WF_ATTR_DONE")
`
		if err := os.WriteFile(wfPath, []byte(wfSrc), 0o644); err != nil {
			t.Fatal(err)
		}
		logPath := filepath.Join(t.TempDir(), "agy-wf-attr.jsonl")
		out := w.run(t, "",
			"Call run_workflow with ScriptPath='"+wfPath+"', Description='attribution test', WaitMsBeforeAsync=10000.",
			"ASYMMETRY_LOG="+logPath)

		events := readAsymmetryEvents(t, logPath)
		var sawChildCmd bool
		for _, ev := range events {
			if ev.Event == "SubagentStart" || ev.Event == "SubagentStop" {
				t.Errorf("ASSUMPTION CHANGED: Antigravity now fires %s for run_workflow child (%+v); .agy-sittings workaround can be retired", ev.Event, ev)
			}
			if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "ASYM_WF_CHILD") {
				sawChildCmd = true
				if ev.AgentID != "" || ev.AgentType != "" || strings.Contains(ev.Raw, "parentConversationId") || strings.Contains(ev.Raw, "agentName") {
					t.Errorf("ASSUMPTION CHANGED: run_workflow child hook now carries parent/agent attribution (agent_id=%q agent_type=%q raw=%s); .agy-sittings / --seat-id binding can be retired",
						ev.AgentID, ev.AgentType, ev.Raw)
				}
			}
		}
		if !sawChildCmd {
			t.Fatalf("never observed run_workflow child running ASYM_WF_CHILD; events=%+v\nout:\n%s", events, out)
		}
	})
}

// TestLiveAsymmetryFinishToolOverwrite pins the structured-output PreToolUse overwrite asymmetry:
//   - Claim: Claude Code honors PreToolUse updatedInput on StructuredOutput (used by debate.js);
//     Antigravity ignores PreToolUse overwrite on the run_workflow `finish` tool.
//   - Relying code:
//     plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.py (feov-sitting-write
//     -phase envelope workaround instead of hook-based envelope stamping).
func TestLiveAsymmetryFinishToolOverwrite(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live asymmetry test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live asymmetry test under -short")
	}

	t.Run("Claude", func(t *testing.T) {
		hooksJSON := `{
  "hooks": {
    "PreToolUse": [{"matcher":"StructuredOutput","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PreToolUse"}]}]
  }
}`
		pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
		workDir := t.TempDir()
		logPath := filepath.Join(t.TempDir(), "claude-finish.jsonl")
		schema := `{"type":"object","properties":{"answer":{"type":"string"},"stamped_by_hook":{"type":"string"}},"required":["answer"]}`

		// Pass ASYMMETRY_OVERWRITE_FINISH=1 via env? Wait: runClaudeAsymmetry doesn't take extraEnv;
		// set it via t.Setenv so the child inherits it cleanly.
		t.Setenv("ASYMMETRY_OVERWRITE_FINISH", "1")
		out := runClaudeAsymmetry(t, workDir, pluginDir, logPath, "",
			"Return structured output with answer='ORIGINAL_ANSWER'.",
			"--output-format", "json", "--json-schema", schema)
		if !strings.Contains(out, "OVERWRITTEN_BY_PRETOOLUSE") {
			t.Fatalf("Claude did not honor PreToolUse updatedInput on StructuredOutput; out=%s", out)
		}
	})

	t.Run("Antigravity", func(t *testing.T) {
		w := newAgyWorld(t)
		probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
		buildAsymmetryProbe(t, probeBin)
		w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PreToolUse":[{"matcher":".*","hooks":[{"command":"./bin/asymmetryprobe -event PreToolUse"}]}]}}`)

		wfPath := filepath.Join(w.work, "wf_finish.py")
		wfSrc := `import json
schema = {
    "type": "object",
    "properties": {
        "answer": {"type": "string"},
        "stamped_by_hook": {"type": "string"},
    },
    "required": ["answer"],
}
res = await agent(
    prompt="Finish immediately with answer='ORIGINAL_CHILD_ANSWER'.",
    schema=schema,
    role="Finish Overwrite Child",
    type_name="self",
    model="flash_lite",
)
print("WF_RESULT:" + json.dumps(res))
`
		if err := os.WriteFile(wfPath, []byte(wfSrc), 0o644); err != nil {
			t.Fatal(err)
		}
		logPath := filepath.Join(t.TempDir(), "agy-wf-finish.jsonl")
		out := w.run(t, "",
			"Call run_workflow with ScriptPath='"+wfPath+"', Description='finish overwrite test', WaitMsBeforeAsync=10000. Report the WF_RESULT line verbatim.",
			"ASYMMETRY_LOG="+logPath, "ASYMMETRY_OVERWRITE_FINISH=1")
		if strings.Contains(out, "OVERWRITTEN_BY_PRETOOLUSE") {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity now honors PreToolUse overwrite on finish (out=%s); debate.py can switch to hook-based envelope stamping", out)
		}
	})
}
