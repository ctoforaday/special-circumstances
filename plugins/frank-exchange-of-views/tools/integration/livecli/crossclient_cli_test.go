package livecli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/servedmodel"
)

func requireLiveCLI(t *testing.T, needClaude, needAgy bool) (claudeBin, agyBin string) {
	t.Helper()
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live CLI integration test without HOOKCORE_LIVE=1")
	}
	if needClaude {
		var err error
		claudeBin, err = exec.LookPath("claude")
		if err != nil {
			t.Skip("claude CLI not found on PATH; skipping live Claude test")
		}
	}
	if needAgy {
		var err error
		agyBin, err = exec.LookPath("agy")
		if err != nil {
			t.Skip("agy CLI not found on PATH; skipping live Antigravity test")
		}
	}
	return claudeBin, agyBin
}

func feovPluginRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	cur := wd
	for {
		if _, err := os.Stat(filepath.Join(cur, "requirements.json")); err == nil {
			if _, err := os.Stat(filepath.Join(cur, ".claude-plugin", "plugin.json")); err == nil {
				return cur
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			t.Fatalf("could not locate frank-exchange-of-views plugin root above %s", wd)
		}
		cur = parent
	}
}

func buildFEOVBin(t *testing.T, toolsDir, pkg, outPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(outPath), err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "-C", toolsDir, "build", "-o", outPath, pkg)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build %s: %v\n%s", pkg, err, out)
	}
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	if err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		info, err := d.Info()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode()&0o777)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	}); err != nil {
		t.Fatalf("copyTree %s -> %s: %v", src, dst, err)
	}
}

func copyFileIfPresent(src, dst string, mode os.FileMode) {
	b, err := os.ReadFile(src)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(dst), 0o755)
	_ = os.WriteFile(dst, b, mode)
}

func newIsolatedAgyHome(t *testing.T) string {
	t.Helper()
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("real home: %v", err)
	}
	cliSrc := filepath.Join(realHome, ".gemini", "antigravity-cli")
	if _, err := os.Stat(filepath.Join(cliSrc, "antigravity-oauth-token")); err != nil {
		t.Skipf("antigravity-oauth-token not found under %s: skipping live agy test", cliSrc)
	}

	fakeHome := t.TempDir()
	cliDst := filepath.Join(fakeHome, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(cliDst, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"antigravity-oauth-token", "antigravity_state.pbtxt", "installation_id"} {
		copyFileIfPresent(filepath.Join(cliSrc, name), filepath.Join(cliDst, name), 0o600)
	}
	if err := os.WriteFile(filepath.Join(cliDst, "settings.json"), []byte(`{"toolPermission":"always-proceed"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// Preserve Claude subscription credentials if present so setup billing checks run cleanly.
	copyFileIfPresent(filepath.Join(realHome, ".claude", ".credentials.json"), filepath.Join(fakeHome, ".claude", ".credentials.json"), 0o600)
	return fakeHome
}

func cleanEnv(extra ...string) []string {
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "FEOV_") || strings.HasPrefix(e, "CLAUDE_") {
			continue
		}
		env = append(env, e)
	}
	return append(env, extra...)
}

func agyEnv(fakeHome string) []string {
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "FEOV_") || strings.HasPrefix(e, "CLAUDE_") ||
			strings.HasPrefix(e, "ANTIGRAVITY_") || strings.HasPrefix(e, "XDG_") ||
			strings.HasPrefix(e, "HOME=") || strings.HasPrefix(e, "GOCACHE=") ||
			strings.HasPrefix(e, "GOMODCACHE=") || strings.HasPrefix(e, "GOFLAGS=") {
			continue
		}
		env = append(env, e)
	}
	gc, _ := exec.Command("go", "env", "GOCACHE").Output()
	gmc, _ := exec.Command("go", "env", "GOMODCACHE").Output()
	env = append(env, "HOME="+fakeHome, "GOFLAGS=-modcacherw")
	if s := strings.TrimSpace(string(gc)); s != "" {
		env = append(env, "GOCACHE="+s)
	}
	if s := strings.TrimSpace(string(gmc)); s != "" {
		env = append(env, "GOMODCACHE="+s)
	}
	return env
}

func newUUID(t *testing.T) string {
	t.Helper()
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatal(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "livecli@test.local"},
		{"config", "user.name", "livecli"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func stageFullFEOVPlugin(t *testing.T, destPlugin string) string {
	t.Helper()
	srcPlugin := feovPluginRoot(t)
	toolsDir := filepath.Join(srcPlugin, "tools")

	for _, item := range []string{".claude-plugin", ".antigravity-plugin", "hooks", "hooks.json", "requirements.json"} {
		src := filepath.Join(srcPlugin, item)
		if info, err := os.Stat(src); err == nil {
			if info.IsDir() {
				copyTree(t, src, filepath.Join(destPlugin, item))
			} else {
				b, err := os.ReadFile(src)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(destPlugin, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(destPlugin, item), b, 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	bins := []string{
		"feov-record",
		"feov-pretooluse",
		"feov-posttooluse",
		"feov-posttoolusefailure",
		"feov-subagentstart",
		"feov-subagentstop",
	}
	for _, name := range bins {
		buildFEOVBin(t, toolsDir, "./cmd/"+name, filepath.Join(destPlugin, "bin", name))
	}

	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if pData, err := os.ReadFile(filepath.Join(destPlugin, ".claude-plugin", "plugin.json")); err == nil {
		if json.Unmarshal(pData, &manifest) == nil && manifest.Name != "" && manifest.Version != "" {
			_ = os.MkdirAll(filepath.Join(destPlugin, ".fetch"), 0o755)
			_ = os.WriteFile(filepath.Join(destPlugin, ".fetch", "installed"), []byte(manifest.Name+"--v"+manifest.Version+"\n"), 0o644)
		}
	}
	return filepath.Join(destPlugin, "bin", "feov-record")
}

// TestLiveFEOVSetupAndTiersCrossClient exercises the real `feov-record setup`, `show tiers`, and
// `dashboard` commands against live `agy models` discovery — verifying acceptance of a real listed
// Antigravity model alongside Claude models, and fail-fast exit-2 refusal on an unknown model.
func TestLiveFEOVSetupAndTiersCrossClient(t *testing.T) {
	_, agyBin := requireLiveCLI(t, true, true)
	fakeHome := newIsolatedAgyHome(t)

	// Confirm live `agy models` returns usable models on this host before testing setup validation.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	modelsCmd := exec.CommandContext(ctx, agyBin, "models")
	modelsCmd.Env = agyEnv(fakeHome)
	modelsOut, err := modelsCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("agy models failed: %v\n%s", err, modelsOut)
	}
	const targetAgyModel = "gemini-3.8-flash-low"
	if !strings.Contains(string(modelsOut), targetAgyModel) {
		t.Fatalf("expected live `agy models` output to include %q, got:\n%s", targetAgyModel, modelsOut)
	}

	pluginStage := filepath.Join(t.TempDir(), "frank-exchange-of-views")
	srcRoot := feovPluginRoot(t)
	feovRecord := filepath.Join(pluginStage, "bin", "feov-record")
	buildFEOVBin(t, filepath.Join(srcRoot, "tools"), "./cmd/feov-record", feovRecord)
	copyFileIfPresent(filepath.Join(srcRoot, "requirements.json"), filepath.Join(pluginStage, "requirements.json"), 0o644)

	workDir := t.TempDir()
	initGitRepo(t, workDir)
	sessionID := newUUID(t)
	setupEnv := append(agyEnv(fakeHome), seatenv.SessionVar+"="+sessionID)

	// 1. Unknown Antigravity model MUST fail live model validation with exit 2 and create nothing.
	badRunDir := filepath.Join(workDir, "research", "bad-agy-model")
	badCmd := exec.Command(feovRecord, "setup", badRunDir,
		"--seat-id", "operator",
		"--topic", "live model validation refusal",
		"--bulk", "haiku",
		"--judgment", "sonnet",
		"--seat", "red-lens-dark-side=antigravity:nonexistent-agy-model-999",
		"--cite", "README.md",
	)
	badCmd.Dir = workDir
	badCmd.Env = setupEnv
	badOut, badErr := badCmd.CombinedOutput()
	ee, ok := badErr.(*exec.ExitError)
	if !ok || ee.ExitCode() != 2 {
		t.Fatalf("expected exit 2 on invalid agy model, got err=%v out:\n%s", badErr, badOut)
	}
	if _, statErr := os.Stat(filepath.Join(badRunDir, "blue")); statErr == nil {
		t.Fatalf("setup created run directories after model refusal")
	}

	// 2. Valid cross-client configuration succeeds live against real `agy models` and `claude`.
	runDir := filepath.Join(workDir, "research", "2026-11-13_cross-client-live")
	goodCmd := exec.Command(feovRecord, "setup", runDir,
		"--seat-id", "operator",
		"--topic", "live cross-client setup",
		"--bulk", "haiku",
		"--judgment", "sonnet",
		"--seat", "red-lens-dark-side=antigravity:"+targetAgyModel,
		"--cite", "README.md",
	)
	goodCmd.Dir = workDir
	goodCmd.Env = setupEnv
	if out, err := goodCmd.CombinedOutput(); err != nil {
		t.Fatalf("feov-record setup failed: %v\n%s", err, out)
	}

	cfgBytes, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		t.Fatalf("read run-config.json: %v", err)
	}
	var cfg struct {
		Seats map[string]struct {
			Harness string `json:"harness"`
			Model   string `json:"model"`
		} `json:"seats"`
	}
	if err := json.Unmarshal(cfgBytes, &cfg); err != nil {
		t.Fatalf("unmarshal run-config.json: %v\n%s", err, cfgBytes)
	}
	if got := cfg.Seats["red-lens-dark-side"]; got.Harness != "antigravity" || got.Model != targetAgyModel {
		t.Fatalf("red-lens-dark-side = %+v, want antigravity:%s", got, targetAgyModel)
	}
	if got := cfg.Seats["red-chair"]; got.Harness != "claude" || got.Model != "sonnet" {
		t.Fatalf("red-chair = %+v, want claude:sonnet", got)
	}

	// 3. `ops tiers --format json` and `dashboard` run cleanly over the fresh cross-client record.
	tiersCmd := exec.Command(feovRecord, "ops", "tiers", "--run", runDir, "--seat-id", "operator", "--format", "json")
	tiersCmd.Dir = workDir
	tiersCmd.Env = setupEnv
	tiersOut, err := tiersCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("show tiers failed: %v\n%s", err, tiersOut)
	}
	var tierRep struct {
		ConfiguredSeats map[string]struct {
			Harness string `json:"harness"`
			Model   string `json:"model"`
		} `json:"configuredSeats"`
	}
	if err := json.Unmarshal(tiersOut, &tierRep); err != nil {
		t.Fatalf("decode show tiers output: %v\n%s", err, tiersOut)
	}
	if got := tierRep.ConfiguredSeats["red-lens-dark-side"]; got.Harness != "antigravity" || got.Model != targetAgyModel {
		t.Fatalf("tiers configuredSeats[red-lens-dark-side] = %+v, want antigravity:%s", got, targetAgyModel)
	}

	dashCmd := exec.Command(feovRecord, "dashboard", "--seat-id", "operator", runDir, filepath.Join(runDir, "trajectories"))
	dashCmd.Dir = workDir
	dashCmd.Env = setupEnv
	if out, err := dashCmd.CombinedOutput(); err != nil {
		t.Fatalf("dashboard failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(runDir, "dashboard.html")); err != nil {
		t.Fatalf("dashboard.html missing after dashboard command: %v", err)
	}
}

// TestLiveCrossClientTranscriptIngestionCostAndServedModel runs real 1-turn sessions on both
// `claude` and `agy` CLIs and feeds the genuine generated transcripts through `seatturn.Parse`,
// `servedmodel.Observe`, and `cost.ScanTranscript` to lock format compatibility across clients.
func TestLiveCrossClientTranscriptIngestionCostAndServedModel(t *testing.T) {
	claudeBin, agyBin := requireLiveCLI(t, true, true)
	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("real home: %v", err)
	}

	// 1. Live Claude session JSONL transcript.
	claudeSessionID := newUUID(t)
	claudeWork := t.TempDir()
	ctxC, cancelC := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelC()
	claudeCmd := exec.CommandContext(ctxC, claudeBin,
		"-p", "Reply with exactly: CLAUDE_TRANSCRIPT_OK",
		"--model", "haiku",
		"--session-id", claudeSessionID,
		"--dangerously-skip-permissions",
	)
	claudeCmd.Dir = claudeWork
	claudeCmd.Env = cleanEnv()
	if out, err := claudeCmd.CombinedOutput(); err != nil {
		t.Fatalf("claude -p failed: %v\n%s", err, out)
	}

	var claudeTranscript string
	projectsDir := filepath.Join(realHome, ".claude", "projects")
	_ = filepath.WalkDir(projectsDir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == claudeSessionID+".jsonl" {
			claudeTranscript = p
		}
		return nil
	})
	if claudeTranscript == "" {
		t.Fatalf("could not locate Claude session transcript %s.jsonl under %s", claudeSessionID, projectsDir)
	}
	t.Cleanup(func() { _ = os.Remove(claudeTranscript) })

	rawClaude, err := os.ReadFile(claudeTranscript)
	if err != nil {
		t.Fatalf("read Claude transcript %s: %v", claudeTranscript, err)
	}
	_, claudeTurns := seatturn.Parse(string(rawClaude))
	if len(claudeTurns) == 0 {
		t.Fatalf("seatturn.Parse yielded 0 turns on real Claude transcript %s:\n%s", claudeTranscript, string(rawClaude))
	}
	if claudeTurns[0].Prompt()+claudeTurns[0].Output == 0 {
		t.Errorf("expected non-zero Claude token usage on turn 0, got %+v", claudeTurns[0])
	}
	obs, err := servedmodel.Observe(claudeTranscript)
	if err != nil {
		t.Fatalf("servedmodel.Observe(claude) failed: %v", err)
	}
	if !strings.Contains(strings.ToLower(obs.Served), "haiku") {
		t.Errorf("servedmodel.Observe(claude).Served = %q, want haiku model id", obs.Served)
	}
	cRow := cost.ScanTranscript(string(rawClaude))
	if cRow.Cost <= 0 || len(cRow.Unpriced) != 0 {
		t.Errorf("cost.ScanTranscript(claude) = %+v, want Cost > 0 and Unpriced empty", cRow)
	}

	// 2. Live Antigravity session JSONL transcript.
	fakeHome := newIsolatedAgyHome(t)
	agyWork := t.TempDir()
	ctxA, cancelA := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelA()
	agyCmd := exec.CommandContext(ctxA, agyBin,
		"-p", "Reply with exactly: AGY_TRANSCRIPT_OK",
		"--model", "gemini-3.8-flash-low",
	)
	agyCmd.Dir = agyWork
	agyCmd.Env = agyEnv(fakeHome)
	if out, err := agyCmd.CombinedOutput(); err != nil {
		t.Fatalf("agy -p failed: %v\n%s", err, out)
	}

	brainDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "brain")
	entries, err := os.ReadDir(brainDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 Antigravity brain dir in %s, got entries=%v err=%v", brainDir, entries, err)
	}
	agyTranscript := filepath.Join(brainDir, entries[0].Name(), ".system_generated", "logs", "transcript.jsonl")
	rawAgy, err := os.ReadFile(agyTranscript)
	if err != nil {
		t.Fatalf("read Antigravity transcript %s: %v", agyTranscript, err)
	}
	_, agyTurns := seatturn.Parse(string(rawAgy))
	if len(agyTurns) == 0 {
		t.Fatalf("seatturn.Parse yielded 0 turns on real Antigravity transcript %s:\n%s", agyTranscript, string(rawAgy))
	}
	if agyTurns[0].Input+agyTurns[0].Output == 0 {
		t.Errorf("expected non-zero Antigravity token usage on turn 0, got %+v", agyTurns[0])
	}
	if agyTurns[0].Model != "" {
		t.Logf("Antigravity now emits model field (%q) — update servedmodel expectations if persistent", agyTurns[0].Model)
	} else {
		if _, err := servedmodel.Observe(agyTranscript); err == nil || !strings.Contains(err.Error(), "antigravity transcript carrying no model field") {
			t.Errorf("expected Antigravity servedmodel.Observe to fail explicitly as NOT MEASURED, got err=%v", err)
		}
		aRow := cost.ScanTranscript(string(rawAgy))
		if aRow.Turns == 0 || aRow.Inp+aRow.Cr+aRow.Out == 0 || aRow.Cost != 0 {
			t.Errorf("cost.ScanTranscript(agy) = %+v, want non-zero tokens and Cost == 0", aRow)
		}
		if len(aRow.Unpriced) != 1 || aRow.Unpriced[0] != "antigravity (model unreported)" {
			t.Errorf("cost.ScanTranscript(agy).Unpriced = %v, want [\"antigravity (model unreported)\"]", aRow.Unpriced)
		}
	}
}

// TestLiveCrossClientShippedFEOVHooks stages the complete shipped frank-exchange-of-views plugin
// (all compiled hook binaries + hooks.json + .claude-plugin/hooks.json) and drives a real tool
// call through both `claude -p` (attributed lead session) and `agy -p`, catching wiring bugs
// like `hooks.json` pointing PostToolUse at `feov-pretooluse`.
func TestLiveCrossClientShippedFEOVHooks(t *testing.T) {
	claudeBin, agyBin := requireLiveCLI(t, true, true)
	fakeHome := newIsolatedAgyHome(t)

	// Stage Antigravity plugin into isolated fakeHome and Claude plugin for `--plugin-dir`.
	agyPluginDir := filepath.Join(fakeHome, ".gemini", "config", "plugins", "frank-exchange-of-views")
	feovRecord := stageFullFEOVPlugin(t, agyPluginDir)

	claudePluginDir := filepath.Join(t.TempDir(), "frank-exchange-of-views")
	copyTree(t, agyPluginDir, claudePluginDir)

	workDir := t.TempDir()
	initGitRepo(t, workDir)
	leadSession := newUUID(t)
	runDir := filepath.Join(workDir, "research", "2026-11-13_hooks-smoke")

	setupCmd := exec.Command(feovRecord, "setup", runDir,
		"--seat-id", "operator",
		"--topic", "live shipped FEOV hooks smoke",
		"--bulk", "haiku",
		"--judgment", "sonnet",
		"--seat", "red-lens-dark-side=antigravity:gemini-3.8-flash-low",
		"--skip-model-check",
		"--cite", "README.md",
	)
	setupCmd.Dir = workDir
	setupCmd.Env = append(agyEnv(fakeHome), seatenv.SessionVar+"="+leadSession)
	if out, err := setupCmd.CombinedOutput(); err != nil {
		t.Fatalf("feov-record setup failed: %v\n%s", err, out)
	}

	// 1. Live Claude lead session executing a Bash tool call under shipped FEOV hooks.
	ctxC, cancelC := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelC()
	claudeCmd := exec.CommandContext(ctxC, claudeBin,
		"-p", "Run the Bash command: echo 'FEOV_CLAUDE_HOOK_OK'",
		"--model", "haiku",
		"--session-id", leadSession,
		"--dangerously-skip-permissions",
		"--plugin-dir", claudePluginDir,
	)
	claudeCmd.Dir = workDir
	claudeCmd.Env = cleanEnv()
	if out, err := claudeCmd.CombinedOutput(); err != nil {
		t.Fatalf("claude -p under shipped FEOV hooks failed: %v\n%s", err, out)
	}
	t.Cleanup(func() {
		if realHome, err := os.UserHomeDir(); err == nil {
			_ = filepath.WalkDir(filepath.Join(realHome, ".claude", "projects"), func(p string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() && d.Name() == leadSession+".jsonl" {
					_ = os.Remove(p)
				}
				return nil
			})
		}
	})

	// 2. Live Antigravity session executing a run_command tool call under shipped FEOV hooks.
	ctxA, cancelA := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelA()
	agyCmd := exec.CommandContext(ctxA, agyBin,
		"-p", "Run this exact shell command: echo 'FEOV_AGY_HOOK_OK' and output only its stdout.",
		"--model", "gemini-3.8-flash-low",
	)
	agyCmd.Dir = workDir
	agyCmd.Env = agyEnv(fakeHome)
	var agyBuf bytes.Buffer
	agyCmd.Stdout = &agyBuf
	agyCmd.Stderr = &agyBuf
	if err := agyCmd.Run(); err != nil {
		t.Fatalf("agy -p under shipped FEOV hooks failed: %v\n%s", err, agyBuf.String())
	}
	if !strings.Contains(agyBuf.String(), "FEOV_AGY_HOOK_OK") {
		t.Errorf("expected Antigravity output to contain FEOV_AGY_HOOK_OK, got:\n%s", agyBuf.String())
	}

	// Assert no hook errors or panics were recorded during either client's session.
	if b, err := os.ReadFile(filepath.Join(runDir, "records", "failures.json")); err == nil && len(bytes.TrimSpace(b)) > 0 {
		t.Errorf("shipped FEOV hooks logged failures.json during live sessions:\n%s", string(b))
	}
	_ = filepath.WalkDir(filepath.Join(fakeHome, ".gemini", "antigravity-cli"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".log") {
			if b, err := os.ReadFile(p); err == nil {
				txt := string(b)
				if strings.Contains(txt, "panicked") || strings.Contains(txt, "unknown hook") || strings.Contains(txt, "exec: ") {
					t.Errorf("Antigravity log %s reported hook execution error:\n%s", p, txt)
				}
			}
		}
		return nil
	})
}
