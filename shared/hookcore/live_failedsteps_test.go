package hookcore_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLiveAntigravityFailedRunCommandDrainedOnPreInvocation exercises the real agy CLI on a
// failing run_command (`exit 2`), where Antigravity fires PostToolUse with error:"" and no result
// field, and verifies that hookcore.DrainFailedToolSteps recovers the exit code on the following
// PreInvocation and injects it into the conversation.
func TestLiveAntigravityFailedRunCommandDrainedOnPreInvocation(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live Antigravity CLI test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live Antigravity CLI test under -short")
	}
	agyBin, err := exec.LookPath("agy")
	if err != nil {
		t.Skipf("agy binary not found on PATH: %v", err)
	}

	fakeHome := t.TempDir()
	copyAgyCredentials(t, fakeHome)

	pluginDir := filepath.Join(fakeHome, ".gemini", "config", "plugins", "hookprobe")
	probeBin := filepath.Join(pluginDir, "bin", "hookprobe")
	buildProbeBinary(t, probeBin)

	if err := os.MkdirAll(filepath.Join(pluginDir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"hookprobe","version":"1.0.0","description":"probe"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	hooksJSON := `{
  "probe-hooks": {
    "PreInvocation": [{"command": "./bin/hookprobe -event PreInvocation"}],
    "PostToolUse": [{"matcher": "run_command", "hooks": [{"command": "./bin/hookprobe -event PostToolUse"}]}]
  }
}`
	if err := os.WriteFile(filepath.Join(pluginDir, "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	workDir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, agyBin,
		"-p", "Run the shell command `echo hookcore-fail-probe && exit 2` via run_command. After it finishes, reply with DONE.",
		"--model", "gemini-3.8-flash-low",
		"--dangerously-skip-permissions",
		"--output-format", "text",
	)
	cmd.Dir = workDir
	cmd.Env = liveTestEnv(fakeHome)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("agy -p failed: %v\n%s", err, out.String())
	}

	brainBase := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "brain")
	entries, err := os.ReadDir(brainBase)
	if err != nil || len(entries) == 0 {
		t.Fatalf("no conversation dir created under %s: %v", brainBase, err)
	}
	found := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		trPath := filepath.Join(brainBase, e.Name(), ".system_generated", "logs", "transcript.jsonl")
		data, err := os.ReadFile(trPath)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), "HOOKCORE_DRAIN_ERR:") {
			t.Fatalf("DrainFailedToolSteps returned an error in live agy transcript:\n%s", string(data))
		}
		if strings.Contains(string(data), "HOOKCORE_FAILED_STEP_CODE_2:") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected HOOKCORE_FAILED_STEP_CODE_2 in transcript after failing run_command; agy output:\n%s", out.String())
	}
}
