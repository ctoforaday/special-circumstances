package hookcore_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildProbeBinary compiles the test probe helper binary linking hookcore into the given bin directory.
func buildProbeBinary(t *testing.T, binPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
		t.Fatalf("mkdir probe bin dir: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./testdata/probe")
	cmd.Dir = "."
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build test probe binary failed: %v\nstderr: %s", err, stderr.String())
	}
	if err := os.Chmod(binPath, 0755); err != nil {
		t.Fatalf("chmod probe bin: %v", err)
	}
}

// archiveDir compresses srcDir into a .tar.gz archive at destTarGz.
func archiveDir(t *testing.T, srcDir, destTarGz string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(destTarGz), 0755); err != nil {
		t.Fatalf("mkdir archive dir: %v", err)
	}
	outFile, err := os.Create(destTarGz)
	if err != nil {
		t.Fatalf("create archive file %s: %v", destTarGz, err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			header.Name += "/"
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("archive directory walk failed: %v", err)
	}
}

// archiveSingleFile compresses a single file into a .tar.gz archive.
func archiveSingleFile(t *testing.T, srcFile, destTarGz string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(destTarGz), 0755); err != nil {
		t.Fatalf("mkdir archive dir: %v", err)
	}
	info, err := os.Stat(srcFile)
	if err != nil {
		t.Fatalf("stat srcFile %s: %v", srcFile, err)
	}
	outFile, err := os.Create(destTarGz)
	if err != nil {
		t.Fatalf("create archive file: %v", err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		t.Fatalf("header: %v", err)
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	inF, err := os.Open(srcFile)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer inF.Close()
	if _, err := io.Copy(tw, inF); err != nil {
		t.Fatalf("copy: %v", err)
	}
}

func getArchiveBaseDir(t *testing.T) string {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("get user home: %v", err)
	}
	dir := filepath.Join(home, ".claude", "scratch", "hookcore-live-archive")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func TestLiveClaudeContract(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live Claude CLI test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live Claude CLI test under -short")
	}

	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		t.Skipf("claude binary not found on PATH: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("user home dir: %v", err)
	}
	credPath := filepath.Join(home, ".claude", ".credentials.json")
	if _, err := os.Stat(credPath); err != nil {
		t.Skipf("claude credentials not found at %s: skipping", credPath)
	}

	// 1. Isolate test in a dedicated fake home directory
	fakeHome := t.TempDir()
	fakeClaudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(fakeClaudeDir, 0755); err != nil {
		t.Fatalf("mkdir fakeHome .claude: %v", err)
	}
	credData, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("read credentials: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fakeClaudeDir, ".credentials.json"), credData, 0600); err != nil {
		t.Fatalf("write fake credentials: %v", err)
	}

	// Set up temporary plugin directory for Claude Code
	pluginDir := t.TempDir()
	probeBin := filepath.Join(pluginDir, "bin", "hookprobe")
	buildProbeBinary(t, probeBin)

	claudePluginDir := filepath.Join(pluginDir, ".claude-plugin")
	if err := os.MkdirAll(claudePluginDir, 0755); err != nil {
		t.Fatalf("mkdir .claude-plugin: %v", err)
	}

	pluginJSON := `{
  "name": "hookprobe",
  "version": "1.0.0",
  "description": "Hookcore live test probe",
  "author": { "name": "Antigravity Research Swarm" },
  "hooks": "./.claude-plugin/hooks.json"
}
`
	if err := os.WriteFile(filepath.Join(claudePluginDir, "plugin.json"), []byte(pluginJSON), 0644); err != nil {
		t.Fatalf("write plugin.json: %v", err)
	}

	hooksJSON := `{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "exec \"${CLAUDE_PLUGIN_ROOT}/bin/hookprobe\" -event SessionStart"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "exec \"${CLAUDE_PLUGIN_ROOT}/bin/hookprobe\" -event PreToolUse"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "exec \"${CLAUDE_PLUGIN_ROOT}/bin/hookprobe\" -event PostToolUse"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "exec \"${CLAUDE_PLUGIN_ROOT}/bin/hookprobe\" -event Stop"
          }
        ]
      }
    ]
  }
}
`
	if err := os.WriteFile(filepath.Join(claudePluginDir, "hooks.json"), []byte(hooksJSON), 0644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "hooks.json"), []byte(hooksJSON), 0644); err != nil {
		t.Fatalf("write root hooks.json: %v", err)
	}

	// 2. Execute single session with unified multi-turn prompt in dedicated workspace
	workDir := t.TempDir()
	prompt := "First, run the shell command echo 'TRIGGER_BLOCK'. When that is blocked, run the shell command echo 'TRIGGER_REWRITE'. Next, run the shell command echo 'TRIGGER_FEEDBACK'. Report the exact output of each."
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, claudeBin,
		"-p", prompt,
		"--model", "haiku",
		"--plugin-dir", pluginDir,
		"--dangerously-skip-permissions",
	)
	cmd.Dir = workDir
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "HOME=") || strings.HasPrefix(e, "CLAUDE_") || strings.HasPrefix(e, "FEOV_") {
			continue
		}
		env = append(env, e)
	}
	env = append(env,
		"HOME="+fakeHome,
		"PATH="+os.Getenv("PATH"),
	)
	cmd.Env = env

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Logf("Claude output:\n%s", out.String())
		t.Fatalf("Claude execution failed: %v", err)
	}
	output := out.String()

	// 3. Find and archive the new session file from fakeHome, then clean it up
	projectsDir := filepath.Join(fakeHome, ".claude", "projects")
	var newSessionFiles []string
	_ = filepath.Walk(projectsDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".jsonl") {
			newSessionFiles = append(newSessionFiles, p)
		}
		return nil
	})

	archiveBaseDir := getArchiveBaseDir(t)
	timestamp := time.Now().UTC().Format("2006-01-02_150405")
	for i, sessionFile := range newSessionFiles {
		archivePath := filepath.Join(archiveBaseDir, fmt.Sprintf("claude-%s-%d.tar.gz", timestamp, i))
		archiveSingleFile(t, sessionFile, archivePath)
		t.Logf("Archived Claude session to %s", archivePath)
		_ = os.Remove(sessionFile)
	}

	// 4. Verify behavioral assertions in the single continuous session
	if !strings.Contains(output, "BLOCKED_BY_HOOKCORE_PROBE") {
		t.Errorf("Claude output missing PreToolUse block assertion: %s", output)
	}
	if !strings.Contains(output, "HOOKCORE_REWRITE_SUCCESS") {
		t.Errorf("Claude output missing PreToolUse rewrite assertion: %s", output)
	}
}

func TestLiveAntigravityContract(t *testing.T) {
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

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("user home dir: %v", err)
	}
	tokenPath := filepath.Join(home, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if _, err := os.Stat(tokenPath); err != nil {
		t.Skipf("antigravity oauth token not found at %s: skipping", tokenPath)
	}

	// 1. Isolate test in a dedicated fake home directory
	fakeHome := t.TempDir()
	fakeAgyDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(fakeAgyDir, 0755); err != nil {
		t.Fatalf("mkdir fakeHome .gemini/antigravity-cli: %v", err)
	}
	tokenData, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("read token: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fakeAgyDir, "antigravity-oauth-token"), tokenData, 0600); err != nil {
		t.Fatalf("write fake token: %v", err)
	}
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		src := filepath.Join(home, ".gemini", "antigravity-cli", extra)
		if data, err := os.ReadFile(src); err == nil {
			_ = os.WriteFile(filepath.Join(fakeAgyDir, extra), data, 0600)
		}
	}
	// Tool calls proceed without prompting — the setting a headless run needs, scoped to this throwaway home.
	settingsPath := filepath.Join(fakeAgyDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"toolPermission":"always-proceed"}`), 0600); err != nil {
		t.Fatalf("write fake settings: %v", err)
	}

	// Install probe plugin into fakeHome's global plugins directory
	pluginDir := filepath.Join(fakeHome, ".gemini", "config", "plugins", "hookprobe")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("mkdir plugin dir: %v", err)
	}

	probeBin := filepath.Join(pluginDir, "bin", "hookprobe")
	buildProbeBinary(t, probeBin)

	agyPluginDir := filepath.Join(pluginDir, ".antigravity-plugin")
	if err := os.MkdirAll(agyPluginDir, 0755); err != nil {
		t.Fatalf("mkdir .antigravity-plugin: %v", err)
	}

	pluginJSON := `{
  "name": "hookprobe",
  "version": "1.0.0",
  "description": "Hookcore live test probe",
  "author": "Antigravity Research Swarm"
}
`
	if err := os.WriteFile(filepath.Join(agyPluginDir, "plugin.json"), []byte(pluginJSON), 0644); err != nil {
		t.Fatalf("write plugin.json: %v", err)
	}

	hooksJSON := `{
  "probe-hooks": {
    "SessionStart": [
      {
        "command": "./bin/hookprobe -event SessionStart"
      }
    ],
    "PreInvocation": [
      {
        "command": "./bin/hookprobe -event PreInvocation"
      }
    ],
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "command": "./bin/hookprobe -event PreToolUse"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "command": "./bin/hookprobe -event PostToolUse"
          }
        ]
      }
    ],
    "PostInvocation": [
      {
        "command": "./bin/hookprobe -event PostInvocation"
      }
    ]
  }
}
`
	if err := os.WriteFile(filepath.Join(pluginDir, "hooks.json"), []byte(hooksJSON), 0644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	// 2. Execute single session with unified multi-turn prompt in dedicated workspace
	workDir := t.TempDir()
	prompt := "First, run the shell command echo 'TRIGGER_BLOCK'. When that is denied, run the shell command echo 'TRIGGER_REWRITE'. Next, run the shell command echo 'TRIGGER_FEEDBACK'. Report the exact output of each."
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, agyBin,
		"-p", prompt,
		"--model", "gemini-3.8-flash-low",
	)
	cmd.Dir = workDir
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "HOME=") || strings.HasPrefix(e, "CLAUDE_") || strings.HasPrefix(e, "FEOV_") {
			continue
		}
		env = append(env, e)
	}
	env = append(env,
		"HOME="+fakeHome,
		"PATH="+os.Getenv("PATH"),
	)
	cmd.Env = env

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Logf("Antigravity output:\n%s", out.String())
		t.Fatalf("Antigravity execution failed: %v", err)
	}
	output := out.String()

	// 3. Find and archive the new conversation directory from fakeHome, then clean it up
	brainBase := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "brain")
	var newBrainDirs []string
	if entries, err := os.ReadDir(brainBase); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				newBrainDirs = append(newBrainDirs, filepath.Join(brainBase, e.Name()))
			}
		}
	}

	// 4. Verify behavioral assertions in the single continuous session
	if !strings.Contains(output, "BLOCKED_BY_HOOKCORE_PROBE") {
		t.Errorf("Antigravity output missing PreToolUse block assertion: %s", output)
	}
	if !strings.Contains(output, "HOOKCORE_REWRITE_SUCCESS") {
		t.Errorf("Antigravity output missing PreToolUse rewrite assertion: %s", output)
	}

	// 5. Verify that hookcore Say message arrived in transcript.jsonl
	foundSay := false
	for _, bDir := range newBrainDirs {
		trFile := filepath.Join(bDir, ".system_generated", "logs", "transcript.jsonl")
		if data, err := os.ReadFile(trFile); err == nil {
			if strings.Contains(string(data), "HOOKCORE_SAY_NOTICE") {
				foundSay = true
				break
			}
		}
	}
	if !foundSay {
		t.Errorf("expected hookcore Say message HOOKCORE_SAY_NOTICE in transcript.jsonl across %v", newBrainDirs)
	}

	// 6. Verify no hook panics occurred in cli.log
	assertNoAgyPanics(t, fakeHome)

	archiveBaseDir := getArchiveBaseDir(t)
	timestamp := time.Now().UTC().Format("2006-01-02_150405")
	for i, bDir := range newBrainDirs {
		archivePath := filepath.Join(archiveBaseDir, fmt.Sprintf("antigravity-%s-%d.tar.gz", timestamp, i))
		archiveDir(t, bDir, archivePath)
		t.Logf("Archived Antigravity session to %s", archivePath)
		_ = os.RemoveAll(bDir)
	}
}
