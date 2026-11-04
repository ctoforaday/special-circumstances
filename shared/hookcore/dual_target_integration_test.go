package hookcore_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	builtBinariesMu sync.Mutex
	builtBinaries   = map[string]string{}
)

// getHookBinary compiles the hook executable for the given plugin into a temp directory
// and verifies that its -version stamp matches the tree under test.
// It NEVER reads from or populates plugins/*/bin/.
func getHookBinary(t *testing.T, repoRoot, pluginName, binName, buildDir string) string {
	t.Helper()
	builtBinariesMu.Lock()
	defer builtBinariesMu.Unlock()

	key := pluginName + "/" + binName
	if path, ok := builtBinaries[key]; ok {
		return path
	}

	targetBin := binName
	if runtime.GOOS == "windows" && !strings.HasSuffix(targetBin, ".exe") {
		targetBin += ".exe"
	}

	binDir := filepath.Join(buildDir, pluginName)
	_ = os.MkdirAll(binDir, 0755)
	tmpBin := filepath.Join(binDir, targetBin)
	cmd := exec.Command("go", "build", "-C", filepath.Join(repoRoot, "plugins", pluginName, "tools"),
		"-o", tmpBin, "./cmd/"+binName)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build %s/%s: %v\nOutput: %s", pluginName, binName, err, string(out))
	}

	// Assert binary identity: -version must match git HEAD
	vcmd := exec.Command(tmpBin, "-version")
	vout, err := vcmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s -version failed: %v\nOutput: %s", binName, err, string(vout))
	}
	versionStr := strings.TrimSpace(string(vout))

	gitCmd := exec.Command("git", "rev-parse", "--short=7", "HEAD")
	gitCmd.Dir = repoRoot
	gout, err := gitCmd.Output()
	if err == nil {
		expectedRev := strings.TrimSpace(string(gout))
		if !strings.Contains(versionStr, expectedRev) {
			t.Fatalf("%s -version output %q does not match current HEAD %s", binName, versionStr, expectedRev)
		}
		statusCmd := exec.Command("git", "status", "--porcelain")
		statusCmd.Dir = repoRoot
		sout, _ := statusCmd.Output()
		if len(strings.TrimSpace(string(sout))) == 0 {
			if strings.Contains(versionStr, "+dirty") {
				t.Fatalf("%s was stamped +dirty despite clean git tree: %s", binName, versionStr)
			}
		}
	}

	builtBinaries[key] = tmpBin
	return tmpBin
}

func copyDirOrFile(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, fi.Mode())
	}
	if err := os.MkdirAll(dst, fi.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if err := copyDirOrFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

func stageMockPlugin(t *testing.T, repoRoot, pluginName, binName, binPath string) string {
	t.Helper()
	mockRoot := t.TempDir()
	srcPlugin := filepath.Join(repoRoot, "plugins", pluginName)
	for _, item := range []string{".claude-plugin", ".antigravity-plugin", "hooks", "hooks.json", "requirements.json"} {
		src := filepath.Join(srcPlugin, item)
		if _, err := os.Stat(src); err == nil {
			_ = copyDirOrFile(src, filepath.Join(mockRoot, item))
		}
	}
	destBinDir := filepath.Join(mockRoot, "bin")
	if err := os.MkdirAll(destBinDir, 0755); err != nil {
		t.Fatalf("mkdir mockRoot/bin: %v", err)
	}
	destBin := filepath.Join(destBinDir, filepath.Base(binPath))
	binData, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("read binPath %s: %v", binPath, err)
	}
	if err := os.WriteFile(destBin, binData, 0755); err != nil {
		t.Fatalf("write destBin %s: %v", destBin, err)
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(binName, ".exe") {
		_ = os.WriteFile(filepath.Join(destBinDir, binName), binData, 0755)
	}
	return mockRoot
}

func assertCleanPluginsBin(t *testing.T, repoRoot, when string) {
	t.Helper()
	cmd := exec.Command("git", "status", "--ignored", "--short", "plugins/*/bin")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git status plugins/*/bin: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if !strings.HasSuffix(l, ".gitkeep") {
			t.Fatalf("%s test run: plugins/*/bin contains stray files (must be only .gitkeep):\n%s", when, string(out))
		}
	}
}

// runHook runs a hook executable with the given engine environment and stdin payload.
func runHook(t *testing.T, binPath string, args []string, stdinPayload, envEngine string, mockRoot string, extraEnv map[string]string) (stdout, stderr string, rc int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tempDir := t.TempDir()
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = tempDir
	cmd.Stdin = strings.NewReader(stdinPayload)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Clean environment
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "CLAUDE_") || strings.HasPrefix(e, "FEOV_") || strings.HasPrefix(e, "ANTIGRAVITY_") ||
			strings.HasPrefix(e, "HOME=") || strings.HasPrefix(e, "USERPROFILE=") ||
			strings.HasPrefix(e, "XDG_") {
			continue
		}
		env = append(env, e)
	}

	env = append(env,
		"HOME="+tempDir,
		"USERPROFILE="+tempDir,
		"XDG_CONFIG_HOME="+tempDir,
		"XDG_DATA_HOME="+tempDir,
		"XDG_STATE_HOME="+tempDir,
		"CLAUDE_PLUGIN_DATA="+tempDir,
		"CLAUDE_PLUGIN_ROOT="+mockRoot,
	)

	if envEngine == "claude" {
		env = append(env, "CLAUDECODE=1")
	} else if envEngine == "antigravity" {
		env = append(env, "ANTIGRAVITY_CONVERSATION_ID=test-conv-123")
	}

	for k, v := range extraEnv {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	err := cmd.Run()
	rc = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			rc = exitErr.ExitCode()
		} else {
			t.Fatalf("exec %s failed: %v", binPath, err)
		}
	}
	return outBuf.String(), errBuf.String(), rc
}

// TestAllHooksDualTargetIntegration systematically tests every single hook binary across
// all plugins under both Claude Code and Antigravity wire protocols.
func TestAllHooksDualTargetIntegration(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs repo root: %v", err)
	}

	assertCleanPluginsBin(t, repoRoot, "before")
	defer assertCleanPluginsBin(t, repoRoot, "after")

	tempProj := t.TempDir()
	// Initialize git repo in tempProj so tools recognize it
	initCmd := exec.Command("git", "init", "-q", tempProj)
	if err := initCmd.Run(); err != nil {
		t.Fatalf("git init tempProj: %v", err)
	}

	secretKey := "AKIAIOSFODNN7EXAMPL2"

	tests := []struct {
		name        string
		plugin      string
		binary      string
		args        []string
		claudeInput string
		agyInput    string
		checkClaude func(t *testing.T, stdout, stderr string, rc int)
		checkAgy    func(t *testing.T, stdout, stderr string, rc int)
	}{
		// 1. sc-sessionstart
		{
			name:   "sc-sessionstart",
			plugin: "prosthetic-conscience",
			binary: "sc-sessionstart",
			claudeInput: fmt.Sprintf(`{
				"session_id": "claude-sess-1",
				"cwd": %q,
				"transcript_path": %q,
				"hook_event_name": "SessionStart",
				"source": "startup"
			}`, tempProj, filepath.Join(tempProj, "transcript.jsonl")),
			agyInput: fmt.Sprintf(`{
				"conversationId": "agy-sess-1",
				"cwd": %q,
				"transcriptPath": %q,
				"hook_event_name": "SessionStart"
			}`, tempProj, filepath.Join(tempProj, "transcript.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("expected rc 0, got %d (stderr: %s)", rc, stderr)
				}
				if stdout != "" {
					var out struct {
						HookSpecificOutput *struct {
							HookEventName     string `json:"hookEventName"`
							AdditionalContext string `json:"additionalContext"`
						} `json:"hookSpecificOutput"`
						SystemMessage string `json:"systemMessage"`
					}
					if err := json.Unmarshal([]byte(stdout), &out); err != nil {
						t.Errorf("claude output failed to unmarshal: %v\nOutput: %s", err, stdout)
					}
					if out.HookSpecificOutput != nil && out.HookSpecificOutput.HookEventName != "SessionStart" {
						t.Errorf("hookEventName = %q, want SessionStart", out.HookSpecificOutput.HookEventName)
					}
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("expected rc 0, got %d (stderr: %s)", rc, stderr)
				}
				if stdout != "" {
					var raw map[string]any
					if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
						t.Errorf("agy output failed to unmarshal: %v\nOutput: %s", err, stdout)
					}
				}
			},
		},

		// 2. sc-pretooluse (allow)
		{
			name:   "sc-pretooluse-allow",
			plugin: "prosthetic-conscience",
			binary: "sc-pretooluse",
			claudeInput: fmt.Sprintf(`{
				"session_id": "claude-sess-1",
				"cwd": %q,
				"hook_event_name": "PreToolUse",
				"tool_name": "Bash",
				"tool_input": {"command": "echo test_clean"}
			}`, tempProj),
			agyInput: fmt.Sprintf(`{
				"conversationId": "agy-sess-1",
				"cwd": %q,
				"tool_name": "run_command",
				"tool_input": {"CommandLine": "echo test_clean"}
			}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				// In Claude Code, clean allow can be silent or explicit allow
				if strings.TrimSpace(stdout) != "" {
					var out struct {
						HookSpecificOutput struct {
							PermissionDecision string `json:"permissionDecision"`
						} `json:"hookSpecificOutput"`
					}
					if err := json.Unmarshal([]byte(stdout), &out); err != nil {
						t.Errorf("claude unmarshal: %v\n%s", err, stdout)
					}
					if out.HookSpecificOutput.PermissionDecision == "deny" {
						t.Errorf("unexpected deny on clean tool call")
					}
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				// In Antigravity, clean allow can be silent or explicit allow
				if strings.TrimSpace(stdout) != "" {
					var out struct {
						Decision      string `json:"decision"`
						SystemMessage string `json:"systemMessage"`
					}
					if err := json.Unmarshal([]byte(stdout), &out); err != nil {
						t.Errorf("agy unmarshal: %v\n%s", err, stdout)
					}
					if out.Decision == "deny" {
						t.Errorf("unexpected deny on clean tool call")
					}
					if out.SystemMessage != "" {
						t.Errorf("Antigravity response leaked systemMessage: %q", out.SystemMessage)
					}
				}
			},
		},

		// 3. sc-pretooluse (secret deny)
		{
			name:   "sc-pretooluse-secret-deny",
			plugin: "prosthetic-conscience",
			binary: "sc-pretooluse",
			claudeInput: fmt.Sprintf(`{
				"session_id": "claude-sess-1",
				"cwd": %q,
				"hook_event_name": "PreToolUse",
				"tool_name": "Bash",
				"tool_input": {"command": "echo %s"}
			}`, tempProj, secretKey),
			agyInput: fmt.Sprintf(`{
				"conversationId": "agy-sess-1",
				"cwd": %q,
				"tool_name": "run_command",
				"tool_input": {"CommandLine": "echo %s"}
			}`, tempProj, secretKey),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				var out struct {
					HookSpecificOutput struct {
						PermissionDecision       string `json:"permissionDecision"`
						PermissionDecisionReason string `json:"permissionDecisionReason"`
					} `json:"hookSpecificOutput"`
				}
				if err := json.Unmarshal([]byte(stdout), &out); err != nil {
					t.Errorf("claude unmarshal: %v\n%s", err, stdout)
				}
				if out.HookSpecificOutput.PermissionDecision != "deny" {
					t.Errorf("decision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
				}
				if !strings.Contains(out.HookSpecificOutput.PermissionDecisionReason, "aws-access-token") {
					t.Errorf("reason %q does not mention aws-access-token", out.HookSpecificOutput.PermissionDecisionReason)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				var out struct {
					Decision      string `json:"decision"`
					Reason        string `json:"reason"`
					SystemMessage string `json:"systemMessage"`
				}
				if err := json.Unmarshal([]byte(stdout), &out); err != nil {
					t.Errorf("agy unmarshal: %v\n%s", err, stdout)
				}
				if out.Decision != "deny" {
					t.Errorf("decision = %q, want deny", out.Decision)
				}
				if !strings.Contains(out.Reason, "aws-access-token") {
					t.Errorf("reason %q does not mention aws-access-token", out.Reason)
				}
				if out.SystemMessage != "" {
					t.Errorf("Antigravity response leaked systemMessage (violates protojson): %q", out.SystemMessage)
				}
			},
		},

		// 4. sc-posttooluse
		{
			name:   "sc-posttooluse",
			plugin: "prosthetic-conscience",
			binary: "sc-posttooluse",
			claudeInput: fmt.Sprintf(`{
				"session_id": "claude-sess-1",
				"cwd": %q,
				"hook_event_name": "PostToolUse",
				"tool_name": "Write",
				"tool_input": {"file_path": %q, "content": "package x"},
				"tool_response": {"success": true}
			}`, tempProj, filepath.Join(tempProj, "x.go")),
			agyInput: fmt.Sprintf(`{
				"conversationId": "agy-sess-1",
				"cwd": %q,
				"tool_name": "write_file",
				"tool_input": {"Path": %q, "Content": "package x"},
				"tool_response": {"success": true}
			}`, tempProj, filepath.Join(tempProj, "x.go")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				var raw map[string]any
				if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
					t.Errorf("agy unmarshal: %v\n%s", err, stdout)
				}
				if _, ok := raw["systemMessage"]; ok {
					t.Errorf("Antigravity PostToolUse leaked systemMessage: %s", stdout)
				}
			},
		},

		// 5. sc-strike-counter
		{
			name:   "sc-strike-counter",
			plugin: "prosthetic-conscience",
			binary: "sc-strike-counter",
			claudeInput: fmt.Sprintf(`{
				"session_id": "claude-sess-1",
				"cwd": %q,
				"hook_event_name": "PostToolUseFailure",
				"tool_name": "Bash",
				"tool_input": {"command": "make broken"},
				"error": "exit status 2"
			}`, tempProj),
			agyInput: fmt.Sprintf(`{
				"conversationId": "agy-sess-1",
				"cwd": %q,
				"tool_name": "run_command",
				"tool_input": {"CommandLine": "make broken"},
				"error": "exit status 2"
			}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				var raw map[string]any
				if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
					t.Errorf("agy unmarshal: %v\n%s", err, stdout)
				}
				if _, ok := raw["systemMessage"]; ok {
					t.Errorf("Antigravity PostToolUseFailure leaked systemMessage: %s", stdout)
				}
			},
		},

		// 6. sc-precompact
		{
			name:        "sc-precompact",
			plugin:      "prosthetic-conscience",
			binary:      "sc-precompact",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"PreCompact","trigger":"auto"}`, tempProj),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"hook_event_name":"PreCompact","trigger":"auto"}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 7. sc-postcompact-observe
		{
			name:        "sc-postcompact-observe",
			plugin:      "prosthetic-conscience",
			binary:      "sc-postcompact-observe",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"PostCompact","trigger":"auto"}`, tempProj),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"hook_event_name":"PostCompact","trigger":"auto"}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 8. sc-filechanged-rearm
		{
			name:        "sc-filechanged-rearm",
			plugin:      "prosthetic-conscience",
			binary:      "sc-filechanged-rearm",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"FileChanged","file_path":"x.go"}`, tempProj),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"hook_event_name":"FileChanged","file_path":"x.go"}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 9. sc-stop
		{
			name:        "sc-stop",
			plugin:      "prosthetic-conscience",
			binary:      "sc-stop",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"Stop","stop_hook_active":false}`, tempProj),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"hook_event_name":"Stop","stop_hook_active":false}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 10. sc-subagentstop
		{
			name:        "sc-subagentstop",
			plugin:      "prosthetic-conscience",
			binary:      "sc-subagentstop",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SubagentStop","agent_id":"a1","agent_type":"general-purpose","agent_transcript_path":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"agent_id":"a1","agent_type":"general-purpose","agent_transcript_path":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 11. sc-sessionend
		{
			name:        "sc-sessionend",
			plugin:      "prosthetic-conscience",
			binary:      "sc-sessionend",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SessionEnd","reason":"other"}`, tempProj),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"hook_event_name":"SessionEnd","reason":"other"}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 12. gray-area-capture SessionStart
		{
			name:        "gray-area-capture-SessionStart",
			plugin:      "gray-area",
			binary:      "gray-area-capture",
			args:        []string{"-event", "SessionStart"},
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SessionStart","source":"startup","transcript_path":%q}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"transcriptPath":%q}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0 (stderr: %s)", rc, stderr)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0 (stderr: %s)", rc, stderr)
				}
				if strings.Contains(stderr, "no project root") {
					t.Errorf("gray-area-capture failed to resolve project root under Antigravity: %s", stderr)
				}
			},
		},

		// 13. gray-area-capture Stop
		{
			name:        "gray-area-capture-Stop",
			plugin:      "gray-area",
			binary:      "gray-area-capture",
			args:        []string{"-event", "Stop"},
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"Stop","transcript_path":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"transcriptPath":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 14. gray-area-capture SubagentStop
		{
			name:        "gray-area-capture-SubagentStop",
			plugin:      "gray-area",
			binary:      "gray-area-capture",
			args:        []string{"-event", "SubagentStop"},
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SubagentStop","agent_id":"a1","agent_type":"general-purpose","agent_transcript_path":%q,"transcript_path":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl"), filepath.Join(tempProj, "t.jsonl")),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"agent_id":"a1","agent_type":"general-purpose","agent_transcript_path":%q,"transcriptPath":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl"), filepath.Join(tempProj, "t.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 15. gray-area-capture SessionEnd
		{
			name:        "gray-area-capture-SessionEnd",
			plugin:      "gray-area",
			binary:      "gray-area-capture",
			args:        []string{"-event", "SessionEnd"},
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SessionEnd","reason":"other","transcript_path":%q}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"reason":"other","transcriptPath":%q}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 16. feov-pretooluse
		{
			name:   "feov-pretooluse",
			plugin: "frank-exchange-of-views",
			binary: "feov-pretooluse",
			claudeInput: fmt.Sprintf(`{
				"session_id": "c1",
				"cwd": %q,
				"hook_event_name": "PreToolUse",
				"tool_name": "Bash",
				"tool_input": {"command": "ls"}
			}`, tempProj),
			agyInput: fmt.Sprintf(`{
				"conversationId": "a1",
				"cwd": %q,
				"tool_name": "run_command",
				"tool_input": {"CommandLine": "ls"}
			}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				// Verify no synthetic key leakage into updatedInput if present
				var raw map[string]any
				if err := json.Unmarshal([]byte(stdout), &raw); err == nil {
					if hso, ok := raw["hookSpecificOutput"].(map[string]any); ok {
						if up, ok := hso["updatedInput"].(map[string]any); ok {
							if _, hasCmdLine := up["CommandLine"]; hasCmdLine {
								t.Errorf("Claude updatedInput leaked synthetic CommandLine: %+v", up)
							}
						}
					}
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				if strings.TrimSpace(stdout) != "" {
					var raw map[string]any
					if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
						t.Errorf("agy unmarshal: %v\n%s", err, stdout)
					}
					if _, ok := raw["systemMessage"]; ok {
						t.Errorf("Antigravity feov-pretooluse leaked systemMessage: %s", stdout)
					}
				}
			},
		},

		// 17. feov-posttooluse
		{
			name:   "feov-posttooluse",
			plugin: "frank-exchange-of-views",
			binary: "feov-posttooluse",
			claudeInput: fmt.Sprintf(`{
				"session_id": "c1",
				"cwd": %q,
				"hook_event_name": "PostToolUse",
				"tool_name": "StructuredOutput",
				"tool_input": {},
				"tool_response": {}
			}`, tempProj),
			agyInput: fmt.Sprintf(`{
				"conversationId": "a1",
				"cwd": %q,
				"tool_name": "StructuredOutput",
				"tool_input": {},
				"tool_response": {}
			}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				var raw map[string]any
				if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
					t.Errorf("agy unmarshal: %v\n%s", err, stdout)
				}
				if _, ok := raw["systemMessage"]; ok {
					t.Errorf("Antigravity feov-posttooluse leaked systemMessage: %s", stdout)
				}
			},
		},

		// 18. feov-posttoolusefailure
		{
			name:   "feov-posttoolusefailure",
			plugin: "frank-exchange-of-views",
			binary: "feov-posttoolusefailure",
			claudeInput: fmt.Sprintf(`{
				"session_id": "c1",
				"cwd": %q,
				"hook_event_name": "PostToolUseFailure",
				"tool_name": "Bash",
				"tool_input": {"command": "broken"},
				"error": "err"
			}`, tempProj),
			agyInput: fmt.Sprintf(`{
				"conversationId": "a1",
				"cwd": %q,
				"tool_name": "run_command",
				"tool_input": {"CommandLine": "broken"},
				"error": "err"
			}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
				var raw map[string]any
				if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
					t.Errorf("agy unmarshal: %v\n%s", err, stdout)
				}
				if _, ok := raw["systemMessage"]; ok {
					t.Errorf("Antigravity feov-posttoolusefailure leaked systemMessage: %s", stdout)
				}
			},
		},

		// 19. feov-subagentstart
		{
			name:        "feov-subagentstart",
			plugin:      "frank-exchange-of-views",
			binary:      "feov-subagentstart",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SubagentStart","agent_id":"a1","agent_type":"frank-exchange-of-views:red-chair"}`, tempProj),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"agent_id":"a1","agent_type":"frank-exchange-of-views:red-chair"}`, tempProj),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},

		// 20. feov-subagentstop
		{
			name:        "feov-subagentstop",
			plugin:      "frank-exchange-of-views",
			binary:      "feov-subagentstop",
			claudeInput: fmt.Sprintf(`{"session_id":"c1","cwd":%q,"hook_event_name":"SubagentStop","agent_id":"a1","agent_type":"frank-exchange-of-views:red-chair","agent_transcript_path":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			agyInput:    fmt.Sprintf(`{"conversationId":"a1","cwd":%q,"agent_id":"a1","agent_type":"frank-exchange-of-views:red-chair","agent_transcript_path":%q,"stop_hook_active":false}`, tempProj, filepath.Join(tempProj, "t.jsonl")),
			checkClaude: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
			checkAgy: func(t *testing.T, stdout, stderr string, rc int) {
				if rc != 0 {
					t.Errorf("rc = %d, want 0", rc)
				}
			},
		},
	}

	rootBuildDir := t.TempDir()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			binPath := getHookBinary(t, repoRoot, tc.plugin, tc.binary, rootBuildDir)
			mockRoot := stageMockPlugin(t, repoRoot, tc.plugin, tc.binary, binPath)

			extraEnv := map[string]string{
				"CLAUDE_PROJECT_DIR": tempProj,
			}

			// 1. Run under Claude
			t.Run("Claude", func(t *testing.T) {
				stdout, stderr, rc := runHook(t, binPath, tc.args, tc.claudeInput, "claude", mockRoot, extraEnv)
				tc.checkClaude(t, stdout, stderr, rc)
			})

			// 2. Run under Antigravity
			t.Run("Antigravity", func(t *testing.T) {
				stdout, stderr, rc := runHook(t, binPath, tc.args, tc.agyInput, "antigravity", mockRoot, extraEnv)
				tc.checkAgy(t, stdout, stderr, rc)
			})
		})
	}
}
