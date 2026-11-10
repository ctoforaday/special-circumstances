package doctor

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupMockMarketplaceRepo(t *testing.T) (repoDir, tagCommit string, wantHooks []byte) {
	t.Helper()
	repoDir = t.TempDir()
	runGitCmd := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repoDir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, string(out))
		}
		return strings.TrimSpace(string(out))
	}
	runGitCmd("init")
	runGitCmd("config", "user.email", "test@example.com")
	runGitCmd("config", "user.name", "Test")

	pcDir := filepath.Join(repoDir, "plugins", "prosthetic-conscience")
	if err := os.MkdirAll(filepath.Join(pcDir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pcDir, ".claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"prosthetic-conscience","version":"0.54.0","description":"test"}` + "\n"
	if err := os.WriteFile(filepath.Join(pcDir, ".antigravity-plugin", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pcDir, ".claude-plugin", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pcDir, "requirements.json"), []byte(`{"tools":[]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantHooks = []byte("{\n  \"hooks\": {\n    \"SessionStart\": [\n      {\"command\": \"sh hooks/fetch-bin.sh hook SessionStart\"}\n    ]\n  }\n}\n")
	if err := os.WriteFile(filepath.Join(pcDir, "hooks.json"), wantHooks, 0o644); err != nil {
		t.Fatal(err)
	}

	mktDir := filepath.Join(repoDir, ".claude-plugin")
	if err := os.MkdirAll(mktDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mktJSON := `{
  "name": "special-circumstances",
  "plugins": [
    {
      "name": "prosthetic-conscience",
      "source": {
        "source": "git-subdir",
        "path": "plugins/prosthetic-conscience",
        "ref": "prosthetic-conscience--v0.54.0"
      }
    }
  ]
}` + "\n"
	if err := os.WriteFile(filepath.Join(mktDir, "marketplace.json"), []byte(mktJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	runGitCmd("add", "-A")
	runGitCmd("commit", "-m", "release 0.54.0")
	runGitCmd("tag", "prosthetic-conscience--v0.54.0")
	tagCommit = runGitCmd("rev-parse", "prosthetic-conscience--v0.54.0^{commit}")
	return repoDir, tagCommit, wantHooks
}

// TestAgyContentWarningsDetectsStaleAndHalfUpdatedInstall verifies that sc-doctor
// detects:
//  1. Missing .sc-installed-ref
//  2. Mismatched .sc-installed-ref commit
//  3. Matching .sc-installed-ref commit with a stale/Claude-format hooks.json (the 11-07 failure)
//  4. Clean state when both .sc-installed-ref and hooks.json match the pinned tag.
func TestAgyContentWarningsDetectsStaleAndHalfUpdatedInstall(t *testing.T) {
	repoDir, tagCommit, wantHooks := setupMockMarketplaceRepo(t)
	t.Setenv("SC_MARKETPLACE_REPO", repoDir)
	t.Setenv("ANTIGRAVITY_AGENT", "1")

	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome)

	plugRoot := filepath.Join(fakeHome, ".gemini", "config", "plugins", "prosthetic-conscience")
	if err := os.MkdirAll(filepath.Join(plugRoot, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugRoot, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"prosthetic-conscience","version":"0.54.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugRoot, "requirements.json"), []byte(`{"tools":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Case 1: No .sc-installed-ref marker
	w1 := agyContentWarnings(plugRoot)
	if len(w1) != 1 || !strings.Contains(w1[0], "PLUGIN CONTENTS STALE") || !strings.Contains(w1[0], "no .sc-installed-ref") {
		t.Fatalf("expected missing marker warning, got: %v", w1)
	}
	out1, _ := doctor(t, plugRoot)
	if !strings.Contains(out1, "PLUGIN CONTENTS STALE") || !strings.Contains(out1, "VERDICT: DEGRADED") {
		t.Fatalf("missing .sc-installed-ref must downgrade READY to DEGRADED:\n%s", out1)
	}

	// Case 2: Wrong commit in .sc-installed-ref
	if err := os.WriteFile(filepath.Join(plugRoot, agyInstalledRefFile), []byte("0000000000000000000000000000000000000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w2 := agyContentWarnings(plugRoot)
	if len(w2) != 1 || !strings.Contains(w2[0], "PLUGIN CONTENTS STALE") || !strings.Contains(w2[0], "0000000") {
		t.Fatalf("expected commit mismatch warning, got: %v", w2)
	}

	// Case 3: Matching .sc-installed-ref, but hooks.json is old Claude-format (half-updated install)
	if err := os.WriteFile(filepath.Join(plugRoot, agyInstalledRefFile), []byte(tagCommit+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldClaudeHooks := `{"description":"Claude Code quality and safety hooks","hooks":{}}`
	if err := os.WriteFile(filepath.Join(plugRoot, "hooks.json"), []byte(oldClaudeHooks), 0o644); err != nil {
		t.Fatal(err)
	}
	w3 := agyContentWarnings(plugRoot)
	if len(w3) != 1 || !strings.Contains(w3[0], "hooks.json differs") {
		t.Fatalf("expected hooks.json differs warning, got: %v", w3)
	}
	out3, _ := doctor(t, plugRoot)
	if !strings.Contains(out3, "hooks.json differs") || !strings.Contains(out3, "VERDICT: DEGRADED") {
		t.Fatalf("mismatched hooks.json must downgrade READY to DEGRADED:\n%s", out3)
	}

	// Case 4: Matching .sc-installed-ref AND matching hooks.json -> 0 warnings, READY
	if err := os.WriteFile(filepath.Join(plugRoot, "hooks.json"), wantHooks, 0o644); err != nil {
		t.Fatal(err)
	}
	w4 := agyContentWarnings(plugRoot)
	if len(w4) != 0 {
		t.Fatalf("expected 0 warnings when ref and hooks.json match tag, got: %v", w4)
	}
	out4, _ := doctor(t, plugRoot)
	if !strings.Contains(out4, "VERDICT: READY") {
		t.Fatalf("expected VERDICT: READY once ref and hooks.json match tag:\n%s", out4)
	}
}

// TestReinstallAgyPluginsReplacesWholePluginDir verifies that `sc-doctor -fix`
// extracts the pinned tag via `git archive` and runs `agy plugin install <dir>`,
// replacing the whole plugin directory (including removing leftover stale files)
// and stamping .sc-installed-ref.
func TestReinstallAgyPluginsReplacesWholePluginDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell mock for agy uses sh")
	}
	repoDir, tagCommit, wantHooks := setupMockMarketplaceRepo(t)
	t.Setenv("SC_MARKETPLACE_REPO", repoDir)
	t.Setenv("ANTIGRAVITY_AGENT", "1")

	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	t.Setenv("USERPROFILE", fakeHome)

	plugRoot := filepath.Join(fakeHome, ".gemini", "config", "plugins", "prosthetic-conscience")
	if err := os.MkdirAll(filepath.Join(plugRoot, "hooks"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(plugRoot, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugRoot, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"prosthetic-conscience","version":"0.54.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plugRoot, "requirements.json"), []byte(`{"tools":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write stale Claude-format hooks.json and a leftover hooks/hooks.json file
	if err := os.WriteFile(filepath.Join(plugRoot, "hooks.json"), []byte(`{"description":"stale claude format"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	staleNested := filepath.Join(plugRoot, "hooks", "hooks.json")
	if err := os.WriteFile(staleNested, []byte(`{"description":"leftover nested hooks.json"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a mock `agy` binary that implements `agy plugin install <srcDir>` the same way
	// the real `agy` CLI does: rm -rf $HOME/.gemini/config/plugins/<name> && cp -R <srcDir> ...
	mockBinDir := t.TempDir()
	mockAgy := filepath.Join(mockBinDir, "agy")
	script := "#!/bin/sh\n" +
		"set -e\n" +
		"if [ \"$1\" = \"plugin\" ] && [ \"$2\" = \"install\" ]; then\n" +
		"  SRC=\"$3\"\n" +
		"  DEST=\"$HOME/.gemini/config/plugins/prosthetic-conscience\"\n" +
		"  rm -rf \"$DEST\"\n" +
		"  mkdir -p \"$(dirname \"$DEST\")\"\n" +
		"  cp -R \"$SRC\" \"$DEST\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 1\n"
	if err := os.WriteFile(mockAgy, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGY", mockAgy)

	var out bytes.Buffer
	var fixed []binStatus
	if code := run([]string{"-root", plugRoot, "-fix"}, &out, "", noExecutable, noFix(&fixed)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	stdout := out.String()
	if !strings.Contains(stdout, "reinstalled plugin contents from prosthetic-conscience--v0.54.0") {
		t.Fatalf("expected reinstall report in stdout, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "VERDICT: READY") {
		t.Fatalf("expected VERDICT: READY after -fix reinstall, got:\n%s", stdout)
	}

	// Verify hooks.json matches the tag byte-for-byte
	gotHooks, err := os.ReadFile(filepath.Join(plugRoot, "hooks.json"))
	if err != nil {
		t.Fatalf("read reinstalled hooks.json: %v", err)
	}
	if !bytes.Equal(gotHooks, wantHooks) {
		t.Errorf("hooks.json after reinstall:\n%s\nwant:\n%s", string(gotHooks), string(wantHooks))
	}

	// Verify leftover nested hooks/hooks.json was removed by the whole-directory replacement
	if _, err := os.Stat(staleNested); !os.IsNotExist(err) {
		t.Errorf("expected leftover %s to be removed by whole-plugin reinstall, err=%v", staleNested, err)
	}

	// Verify .sc-installed-ref was written with tagCommit
	gotRef, err := os.ReadFile(filepath.Join(plugRoot, agyInstalledRefFile))
	if err != nil {
		t.Fatalf("read .sc-installed-ref: %v", err)
	}
	if strings.TrimSpace(string(gotRef)) != tagCommit {
		t.Errorf(".sc-installed-ref = %q, want %q", strings.TrimSpace(string(gotRef)), tagCommit)
	}

	// Verify ReferenceCommit reads .sc-installed-ref
	if ref := ReferenceCommit(binStatus{Name: "sc-doctor", Root: plugRoot, Built: true}); ref != tagCommit {
		t.Errorf("ReferenceCommit = %q, want %q", ref, tagCommit)
	}
}
