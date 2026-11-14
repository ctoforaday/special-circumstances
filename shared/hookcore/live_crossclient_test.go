package hookcore_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type subagentProbeEvent struct {
	Event               string         `json:"event"`
	Engine              string         `json:"engine"`
	SessionID           string         `json:"session_id"`
	TranscriptPath      string         `json:"transcript_path"`
	AgentID             string         `json:"agent_id"`
	AgentType           string         `json:"agent_type"`
	AgentTranscriptPath string         `json:"agent_transcript_path"`
	ToolName            string         `json:"tool_name"`
	RawToolName         string         `json:"raw_tool_name"`
	Command             string         `json:"command"`
	RawToolInput        map[string]any `json:"raw_tool_input,omitempty"`
}

func readSubagentProbeEvents(t *testing.T, path string) []subagentProbeEvent {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open probe log %s: %v", path, err)
	}
	defer f.Close()

	var out []subagentProbeEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var ev subagentProbeEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			t.Fatalf("decode probe event %q: %v", sc.Text(), err)
		}
		out = append(out, ev)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan probe log: %v", err)
	}
	return out
}

func buildSubagentProbe(t *testing.T, outPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(outPath), err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", outPath, "./testdata/subagentprobe")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build subagentprobe: %v\n%s", err, out)
	}
}

// TestLiveCrossClientSubagentLifecycleAndOverwrite drives real `agy` and `claude` CLIs to spawn
// a custom plugin subagent while exercising hookcore's UpdatedInput model overwrite, subagent ID/type
// normalization, and per-conversation pending queue isolation across both harnesses.
func TestLiveCrossClientSubagentLifecycleAndOverwrite(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live test without HOOKCORE_LIVE=1")
	}

	t.Run("Antigravity", func(t *testing.T) {
		w := newAgyWorld(t)
		probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "subagentprobe", "bin", "subagentprobe")
		buildSubagentProbe(t, probeBin)

		hooksJSON := `{
  "subagentprobe-hooks": {
    "SessionStart": [{"command": "./bin/subagentprobe -event SessionStart"}],
    "PreInvocation": [{"command": "./bin/subagentprobe -event PreInvocation"}],
    "PreToolUse": [
      {
        "matcher": ".*",
        "hooks": [{"command": "./bin/subagentprobe -event PreToolUse"}]
      }
    ],
    "PostToolUse": [
      {
        "matcher": ".*",
        "hooks": [{"command": "./bin/subagentprobe -event PostToolUse"}]
      }
    ],
    "SubagentStart": [{"command": "./bin/subagentprobe -event SubagentStart"}],
    "SubagentStop": [{"command": "./bin/subagentprobe -event SubagentStop"}]
  }
}`
		pluginDir := w.installPlugin(t, "subagentprobe", hooksJSON)
		agentDir := filepath.Join(pluginDir, "agents")
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			t.Fatal(err)
		}
		agentDef := `---
name: child-worker
description: Verification subagent for live cross-client hook tests.
model: pro
tools: [read_file, run_command]
---
You are child-worker. Run the exact shell command requested and return its output verbatim.
`
		if err := os.WriteFile(filepath.Join(agentDir, "child-worker.md"), []byte(agentDef), 0o644); err != nil {
			t.Fatal(err)
		}

		logPath := filepath.Join(t.TempDir(), "agy-subagent.jsonl")
		prompt := "Call invoke_subagent with TypeName='child-worker', Role='Worker', Model='pro', Prompt=\"Run this exact shell command: echo 'CHILD_FEEDBACK_CANARY' and report its output verbatim.\" Do NOT run run_command yourself; delegate to child-worker."
		agyOut := w.run(t, "", prompt, "SUBAGENTPROBE_LOG="+logPath)

		events := readSubagentProbeEvents(t, logPath)
		var parentSessionID, childConvID, childTranscript string
		var sawInvokePre, sawInvokeOverwrittenPost, sawChildPost bool
		for _, ev := range events {
			if ev.RawToolName == "invoke_subagent" && ev.Event == "PreToolUse" {
				sawInvokePre = true
				parentSessionID = ev.SessionID
			}
			if ev.RawToolName == "invoke_subagent" && ev.Event == "PostToolUse" {
				if m, _ := ev.RawToolInput["Model"].(string); m == "flash_lite" {
					sawInvokeOverwrittenPost = true
				}
			}
			if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "CHILD_FEEDBACK_CANARY") && parentSessionID != "" && ev.SessionID != parentSessionID {
				sawChildPost = true
				childConvID = ev.SessionID
				childTranscript = ev.TranscriptPath
			}
		}
		if !sawInvokePre {
			t.Fatalf("never observed PreToolUse on invoke_subagent; events=%+v\nout:\n%s", events, agyOut)
		}
		if !sawInvokeOverwrittenPost {
			t.Fatalf("PostToolUse did not reflect UpdatedInput Model=flash_lite overwrite on invoke_subagent; events=%+v", events)
		}
		if !sawChildPost || childConvID == "" {
			var transcripts []string
			for _, c := range w.conversations(t) {
				p := filepath.Join(w.home, ".gemini", "antigravity-cli", "brain", c, ".system_generated", "logs", "transcript.jsonl")
				b, err := os.ReadFile(p)
				transcripts = append(transcripts, fmt.Sprintf("=== %s (err=%v) ===\n%s", c, err, string(b)))
			}
			t.Fatalf("never observed child PostToolUse with CHILD_FEEDBACK_CANARY in distinct subagent conversation (parent=%s); events=%+v\nout:\n%s\ntranscripts:\n%s",
				parentSessionID, events, agyOut, strings.Join(transcripts, "\n"))
		}

		// Verify child and parent transcripts exist and pending queues stayed strictly isolated:
		// parent got HOOK_MODEL_OVERWRITE_AGY, child got CHILD_QUEUE_OK, neither leaked to the other.
		if childTranscript == "" {
			t.Fatalf("expected non-empty TranscriptPath for child %s", childConvID)
		}
		rawChild, err := os.ReadFile(childTranscript)
		if err != nil {
			t.Fatalf("read child transcript %s: %v", childTranscript, err)
		}
		if !strings.Contains(string(rawChild), "CHILD_QUEUE_OK") {
			t.Errorf("child transcript missing flushed CHILD_QUEUE_OK ephemeral feedback:\n%s", string(rawChild))
		}
		if strings.Contains(string(rawChild), "HOOK_MODEL_OVERWRITE_AGY") {
			t.Errorf("parent Say leaked into child transcript %s", childConvID)
		}
		parentRaw := w.transcript(t, parentSessionID)
		for _, line := range strings.Split(parentRaw, "\n") {
			if strings.Contains(line, `"EPHEMERAL_MESSAGE"`) && strings.Contains(line, "CHILD_QUEUE_OK") {
				t.Errorf("child Feedback leaked into parent transcript %s as EPHEMERAL_MESSAGE: %s", parentSessionID, line)
			}
		}
		if !strings.Contains(parentRaw, "HOOK_MODEL_OVERWRITE_AGY") {
			t.Errorf("parent transcript missing HOOK_MODEL_OVERWRITE_AGY system message:\n%s", parentRaw)
		}
		pluginDataRoot := filepath.Join(w.home, ".claude", "plugins", "data", "subagentprobe")
		if left, err := os.ReadFile(filepath.Join(pluginDataRoot, fmt.Sprintf("pending_%s.jsonl", childConvID))); err == nil && len(strings.TrimSpace(string(left))) > 0 {
			t.Errorf("child pending queue still non-empty after child completed: %s", string(left))
		}
	})

	t.Run("Claude", func(t *testing.T) {
		claudeBin := findClaudeBin()
		if claudeBin == "" {
			t.Skip("claude CLI not found; skipping live Claude subagent test")
		}

		stagedDir := t.TempDir()
		pluginDir := filepath.Join(stagedDir, "subagentprobe")
		claudeMeta := filepath.Join(pluginDir, ".claude-plugin")
		agentsDir := filepath.Join(pluginDir, "agents")
		if err := os.MkdirAll(claudeMeta, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatal(err)
		}

		probeBin := filepath.Join(pluginDir, "bin", "subagentprobe")
		buildSubagentProbe(t, probeBin)

		if err := os.WriteFile(filepath.Join(claudeMeta, "plugin.json"), []byte(`{"name":"subagentprobe","version":"0.0.1","description":"Live cross-client subagent test","hooks":"./.claude-plugin/hooks.json"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/subagentprobe\" -event SessionStart"}]}],
    "SubagentStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/subagentprobe\" -event SubagentStart"}]}],
    "PreToolUse": [{"matcher":"Agent","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/subagentprobe\" -event PreToolUse"}]}],
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/subagentprobe\" -event PostToolUse"}]}],
    "SubagentStop": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/subagentprobe\" -event SubagentStop"}]}]
  }
}`
		if err := os.WriteFile(filepath.Join(claudeMeta, "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
			t.Fatal(err)
		}
		agentDef := `---
name: child-worker
description: Verification subagent for live cross-client hook tests.
model: sonnet
tools: Bash
---
You are child-worker. Run the exact Bash command requested and return its output verbatim.
`
		if err := os.WriteFile(filepath.Join(agentsDir, "child-worker.md"), []byte(agentDef), 0o644); err != nil {
			t.Fatal(err)
		}

		logPath := filepath.Join(t.TempDir(), "claude-subagent.jsonl")
		workDir := t.TempDir()
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		prompt := "Spawn the subagent subagentprobe:child-worker via the Agent tool with model=\"sonnet\" to run the Bash command `echo 'CHILD_FEEDBACK_CANARY'` and return its output. Do NOT run Bash yourself."
		cmd := exec.CommandContext(ctx, claudeBin,
			"-p", prompt,
			"--model", "haiku",
			"--dangerously-skip-permissions",
			"--plugin-dir", pluginDir,
		)
		cmd.Dir = workDir
		var env []string
		for _, e := range os.Environ() {
			if strings.HasPrefix(e, "CLAUDE_") || strings.HasPrefix(e, "FEOV_") {
				continue
			}
			env = append(env, e)
		}
		env = append(env, "SUBAGENTPROBE_LOG="+logPath, "CLAUDE_PLUGIN_DATA="+t.TempDir())
		cmd.Env = env

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("claude -p failed: %v\nOutput:\n%s", err, out)
		}

		events := readSubagentProbeEvents(t, logPath)
		var sawAgentPre, sawSubStart, sawChildPost, sawSubStop bool
		var childTranscript string
		for _, ev := range events {
			if ev.Event == "PreToolUse" && ev.RawToolName == "Agent" {
				sawAgentPre = true
			}
			if ev.Event == "SubagentStart" && ev.AgentID != "" && strings.Contains(ev.AgentType, "child-worker") {
				sawSubStart = true
			}
			if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "CHILD_FEEDBACK_CANARY") && ev.AgentID != "" {
				sawChildPost = true
			}
			if ev.Event == "SubagentStop" && ev.AgentID != "" {
				sawSubStop = true
				if ev.AgentTranscriptPath != "" {
					childTranscript = ev.AgentTranscriptPath
				}
			}
		}
		if !sawAgentPre {
			t.Fatalf("never observed PreToolUse on Agent tool; events=%+v\noutput:\n%s", events, out)
		}
		if !sawSubStart {
			t.Fatalf("never observed SubagentStart for child-worker; events=%+v", events)
		}
		if !sawChildPost {
			t.Fatalf("never observed child PostToolUse with CHILD_FEEDBACK_CANARY; events=%+v", events)
		}
		if !sawSubStop {
			t.Fatalf("never observed SubagentStop with non-empty AgentID; events=%+v", events)
		}
		if childTranscript == "" {
			t.Fatalf("expected non-empty AgentTranscriptPath on Claude SubagentStop")
		}
		rawChild, err := os.ReadFile(childTranscript)
		if err != nil {
			t.Fatalf("read Claude child transcript %s: %v", childTranscript, err)
		}
		if !strings.Contains(string(rawChild), "claude-haiku") {
			t.Errorf("expected Agent updatedInput model=haiku to run child as claude-haiku; transcript=%s", string(rawChild))
		}
	})
}
