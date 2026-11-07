package doctor_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	builtDoctorOnce sync.Once
	doctorExePath   string
	doctorBuildErr  error
)

// buildDoctorBinary compiles sc-doctor once for real CLI execution tests.
func buildDoctorBinary(t *testing.T) string {
	t.Helper()
	builtDoctorOnce.Do(func() {
		tmp, err := os.MkdirTemp("", "sc-doctor-test-bin-")
		if err != nil {
			doctorBuildErr = err
			t.Fatalf("failed to create temp dir for doctor binary: %v", err)
		}
		ext := ""
		if runtime.GOOS == "windows" {
			ext = ".exe"
		}
		doctorExePath = filepath.Join(tmp, "sc-doctor"+ext)
		// Compile sc-doctor from tools/cmd/sc-doctor
		cmd := exec.Command("go", "build", "-o", doctorExePath, "./cmd/sc-doctor")
		cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
		out, err := cmd.CombinedOutput()
		if err != nil {
			doctorBuildErr = err
			t.Fatalf("failed to build sc-doctor for CLI test: %v\n%s", err, string(out))
		}
	})
	if doctorBuildErr != nil {
		t.Fatalf("sc-doctor compilation failed: %v", doctorBuildErr)
	}
	return doctorExePath
}

// cleanEnv prepares an isolated environment for real CLI subprocess testing,
// preserving PATH for tools like git and go, but stripping harness-specific variables.
func cleanEnv(extra ...string) []string {
	var env []string
	keepPrefixes := []string{"PATH=", "HOME=", "USERPROFILE=", "TMP=", "TEMP=", "TMPDIR=", "SYSTEMROOT="}
	for _, kv := range os.Environ() {
		for _, prefix := range keepPrefixes {
			if strings.HasPrefix(kv, prefix) {
				env = append(env, kv)
				break
			}
		}
	}
	return append(env, extra...)
}

func writeJSONFile(t *testing.T, path string, data any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyExecutable(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestRealCLIClaudeCodeLayout tests sc-doctor running under the Claude Code
// versioned marketplace cache layout:
//
//	<cacheDir>/<marketplace>/<plugin>/<version>/
func TestRealCLIClaudeCodeLayout(t *testing.T) {
	doctorBin := buildDoctorBinary(t)
	cacheDir := t.TempDir()
	market := filepath.Join(cacheDir, "special-circumstances")

	// 1. Self plugin: prosthetic-conscience/0.52.0
	selfRoot := filepath.Join(market, "prosthetic-conscience", "0.52.0")
	writeJSONFile(t, filepath.Join(selfRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "git", "tier": "required", "check_cmd": "git --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(selfRoot, "tools", "cmd", "sc-doctor"), 0o755); err != nil {
		t.Fatal(err)
	}
	selfBin := filepath.Join(selfRoot, "bin", filepath.Base(doctorBin))
	copyExecutable(t, doctorBin, selfBin)

	// 2. Sibling plugin: frank-exchange-of-views/1.73.0
	feovRoot := filepath.Join(market, "frank-exchange-of-views", "1.73.0")
	writeJSONFile(t, filepath.Join(feovRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "node", "tier": "required", "check_cmd": "node --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(feovRoot, "tools", "cmd", "feov-record"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(feovRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	// 3. Sibling plugin: gray-area/0.18.0
	grayRoot := filepath.Join(market, "gray-area", "0.18.0")
	writeJSONFile(t, filepath.Join(grayRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "git", "tier": "required", "check_cmd": "git --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(grayRoot, "tools", "cmd", "telepathy"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(grayRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Execute sc-doctor with CLAUDE_PLUGIN_ROOT pointing to selfRoot
	cmd := exec.Command(selfBin)
	cmd.Env = cleanEnv("CLAUDE_PLUGIN_ROOT=" + selfRoot)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sc-doctor failed with %v\nOutput:\n%s", err, string(out))
	}
	stdout := string(out)

	// Assertions for Claude Code execution
	if !strings.Contains(stdout, "git") {
		t.Errorf("expected stdout to check git, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "-- frank-exchange-of-views --") {
		t.Errorf("expected sibling heading for feov, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "node") {
		t.Errorf("expected feov requirement node, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "-- gray-area --") {
		t.Errorf("expected sibling heading for gray-area, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "feov-record") {
		t.Errorf("expected feov-record binary reported, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "telepathy") {
		t.Errorf("expected telepathy binary reported, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "VERDICT: DEGRADED") {
		t.Errorf("expected VERDICT: DEGRADED due to unbuilt sibling binaries, got:\n%s", stdout)
	}
}

// TestRealCLIAntigravityLayout tests sc-doctor running under the Google Antigravity
// flat plugin layout:
//
//	<configRoot>/plugins/<plugin>/
func TestRealCLIAntigravityLayout(t *testing.T) {
	doctorBin := buildDoctorBinary(t)
	pluginsDir := t.TempDir()

	// 1. Self plugin: prosthetic-conscience (flat directory)
	selfRoot := filepath.Join(pluginsDir, "prosthetic-conscience")
	writeJSONFile(t, filepath.Join(selfRoot, "plugin.json"), map[string]any{
		"name": "prosthetic-conscience", "version": "0.52.0",
	})
	writeJSONFile(t, filepath.Join(selfRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "git", "tier": "required", "check_cmd": "git --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(selfRoot, "tools", "cmd", "sc-doctor"), 0o755); err != nil {
		t.Fatal(err)
	}
	selfBin := filepath.Join(selfRoot, "bin", filepath.Base(doctorBin))
	copyExecutable(t, doctorBin, selfBin)

	// 2. Sibling plugin: frank-exchange-of-views (flat directory)
	feovRoot := filepath.Join(pluginsDir, "frank-exchange-of-views")
	writeJSONFile(t, filepath.Join(feovRoot, "plugin.json"), map[string]any{
		"name": "frank-exchange-of-views", "version": "1.73.0",
	})
	writeJSONFile(t, filepath.Join(feovRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "node", "tier": "required", "check_cmd": "node --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(feovRoot, "tools", "cmd", "feov-record"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(feovRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	// 3. Sibling plugin: gray-area (flat directory)
	grayRoot := filepath.Join(pluginsDir, "gray-area")
	writeJSONFile(t, filepath.Join(grayRoot, ".antigravity-plugin", "plugin.json"), map[string]any{
		"name": "gray-area", "version": "0.18.0",
	})
	writeJSONFile(t, filepath.Join(grayRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "git", "tier": "required", "check_cmd": "git --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(grayRoot, "tools", "cmd", "telepathy"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(grayRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	// 4. Non-plugin directory beside plugins (must be ignored)
	if err := os.MkdirAll(filepath.Join(pluginsDir, "scratch-area"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Execute sc-doctor with ANTIGRAVITY_AGENT=1 and PLUGIN_ROOT set, CLAUDE_PLUGIN_ROOT unset
	cmd := exec.Command(selfBin)
	cmd.Env = cleanEnv(
		"ANTIGRAVITY_AGENT=1",
		"ANTIGRAVITY_APP_DATA_DIR="+filepath.Join(pluginsDir, "antigravity-cli"),
		"PLUGIN_ROOT="+selfRoot,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sc-doctor under Antigravity failed with %v\nOutput:\n%s", err, string(out))
	}
	stdout := string(out)

	// Assertions for Antigravity execution
	if !strings.Contains(stdout, "git") {
		t.Errorf("expected stdout to check git, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "-- frank-exchange-of-views --") {
		t.Errorf("expected sibling heading for feov in flat layout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "node") {
		t.Errorf("expected feov requirement node, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "-- gray-area --") {
		t.Errorf("expected sibling heading for gray-area in flat layout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "feov-record") {
		t.Errorf("expected feov-record binary reported in flat layout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "telepathy") {
		t.Errorf("expected telepathy binary reported in flat layout, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "scratch-area") {
		t.Errorf("non-plugin directory scratch-area must not appear in output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "VERDICT: DEGRADED") {
		t.Errorf("expected VERDICT: DEGRADED, got:\n%s", stdout)
	}
}

// TestRealCLIAntigravityGrandparentDiscovery tests that sc-doctor resolves its root
// via executable grandparent when executed from an outside working directory
// without -root or environment variables.
func TestRealCLIAntigravityGrandparentDiscovery(t *testing.T) {
	doctorBin := buildDoctorBinary(t)
	pluginsDir := t.TempDir()

	selfRoot := filepath.Join(pluginsDir, "prosthetic-conscience")
	writeJSONFile(t, filepath.Join(selfRoot, "plugin.json"), map[string]any{
		"name": "prosthetic-conscience", "version": "0.52.0",
	})
	writeJSONFile(t, filepath.Join(selfRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "git", "tier": "required", "check_cmd": "git --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(selfRoot, "tools", "cmd", "sc-doctor"), 0o755); err != nil {
		t.Fatal(err)
	}
	selfBin := filepath.Join(selfRoot, "bin", filepath.Base(doctorBin))
	copyExecutable(t, doctorBin, selfBin)

	// Sibling plugin
	feovRoot := filepath.Join(pluginsDir, "frank-exchange-of-views")
	writeJSONFile(t, filepath.Join(feovRoot, "plugin.json"), map[string]any{
		"name": "frank-exchange-of-views", "version": "1.73.0",
	})
	writeJSONFile(t, filepath.Join(feovRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "node", "tier": "required", "check_cmd": "node --version"},
		},
	})
	if err := os.MkdirAll(filepath.Join(feovRoot, "tools", "cmd", "feov-record"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(feovRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	outsideWorkDir := t.TempDir()
	cmd := exec.Command(selfBin)
	cmd.Dir = outsideWorkDir
	cmd.Env = cleanEnv("ANTIGRAVITY_AGENT=1") // CLAUDE_PLUGIN_ROOT and PLUGIN_ROOT unset
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sc-doctor grandparent fallback failed: %v\nOutput:\n%s", err, string(out))
	}
	stdout := string(out)

	if !strings.Contains(stdout, "-- frank-exchange-of-views --") {
		t.Errorf("grandparent fallback failed to locate sibling feov:\n%s", stdout)
	}
	if !strings.Contains(stdout, "feov-record") {
		t.Errorf("grandparent fallback failed to report feov-record:\n%s", stdout)
	}
}

// TestRealCLIFixProvisionsSibling verifies that sc-doctor -fix provisions missing
// binaries in sibling plugins under flat layout.
func TestRealCLIFixProvisionsSibling(t *testing.T) {
	doctorBin := buildDoctorBinary(t)
	pluginsDir := t.TempDir()

	selfRoot := filepath.Join(pluginsDir, "prosthetic-conscience")
	writeJSONFile(t, filepath.Join(selfRoot, "plugin.json"), map[string]any{
		"name": "prosthetic-conscience", "version": "0.52.0",
	})
	writeJSONFile(t, filepath.Join(selfRoot, "requirements.json"), map[string]any{
		"tools": []any{
			map[string]any{"name": "git", "tier": "required", "check_cmd": "git --version"},
		},
	})
	selfBin := filepath.Join(selfRoot, "bin", filepath.Base(doctorBin))
	copyExecutable(t, doctorBin, selfBin)

	// Sibling plugin with source command
	feovRoot := filepath.Join(pluginsDir, "frank-exchange-of-views")
	writeJSONFile(t, filepath.Join(feovRoot, "plugin.json"), map[string]any{
		"name": "frank-exchange-of-views", "version": "1.73.0",
	})
	writeJSONFile(t, filepath.Join(feovRoot, "requirements.json"), map[string]any{
		"tools": []any{},
	})
	cmdDir := filepath.Join(feovRoot, "tools", "cmd", "mock-feov-tool")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write go.mod in tools/
	if err := os.WriteFile(filepath.Join(feovRoot, "tools", "go.mod"), []byte("module mockfeov\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(selfBin, "-fix")
	cmd.Env = cleanEnv(
		"ANTIGRAVITY_AGENT=1",
		"PLUGIN_ROOT="+selfRoot,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sc-doctor -fix failed: %v\nOutput:\n%s", err, string(out))
	}
	stdout := string(out)

	if !strings.Contains(stdout, "mock-feov-tool") {
		t.Errorf("expected fix output to mention mock-feov-tool, got:\n%s", stdout)
	}

	// Verify the binary was actually built into feovRoot/bin
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	builtTarget := filepath.Join(feovRoot, "bin", "mock-feov-tool"+ext)
	if _, err := os.Stat(builtTarget); err != nil {
		t.Errorf("expected built binary at %s, not found: %v", builtTarget, err)
	}

	// Verify .fetch/installed was stamped
	installedPath := filepath.Join(feovRoot, ".fetch", "installed")
	data, err := os.ReadFile(installedPath)
	if err != nil {
		t.Errorf("expected .fetch/installed at %s: %v", installedPath, err)
	} else if !strings.Contains(string(data), "frank-exchange-of-views--v1.73.0") {
		t.Errorf("unexpected content in .fetch/installed: %s", string(data))
	}
}
