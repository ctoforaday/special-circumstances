package hookcore_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func runValidator(t *testing.T, home, bin string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = home
	cmd.Env = liveTestEnv(t, home)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// --- Row 5: hooks-manifest-location ---

func asymHooksManifestLocationClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	valOut, valErr := runValidator(t, cw.home, cw.bin, "plugin", "validate", pluginDir)
	if valErr != nil {
		t.Fatalf("claude plugin validate rejected .claude-plugin/hooks.json: %v\n%s", valErr, valOut)
	}
	logPath := filepath.Join(t.TempDir(), "claude-hooks-loc.jsonl")
	cw.run(t, pluginDir, logPath, "", "Reply with OK.", nil)
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 || events[0].Event != "SessionStart" {
		t.Fatalf("Claude did not load .claude-plugin/hooks.json at runtime; events=%+v", events)
	}
}

func asymHooksManifestLocationAgy(t *testing.T) {
	w := requireAgyWorld(t)
	pDir := filepath.Join(w.home, ".gemini", "config", "plugins", "innerhooks")
	if err := os.MkdirAll(filepath.Join(pDir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	buildAsymmetryProbe(t, filepath.Join(pDir, "bin", "asymmetryprobe"))
	manifest := `{"name":"innerhooks","version":"1.0.0","description":"t","hooks":"./.antigravity-plugin/hooks.json"}`
	if err := os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	hooksBody := `{"innerhooks-hooks":{"SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}]}}`
	if err := os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "hooks.json"), []byte(hooksBody), 0o644); err != nil {
		t.Fatal(err)
	}

	valOut, valErr := runValidator(t, w.home, w.bin, "plugin", "validate", "--strict", pDir)
	if valErr == nil || !strings.Contains(valOut, "never loaded") || !strings.Contains(valOut, `unknown field "hooks" is ignored`) {
		t.Fatalf("ASSUMPTION CHANGED: agy plugin validate --strict no longer rejects .antigravity-plugin/hooks.json and plugin.json hooks field (err=%v):\n%s", valErr, valOut)
	}

	// Runtime check: inner hooks never fire, while a root hooks.json plugin in the same home DOES fire (positive control).
	innerLog := filepath.Join(t.TempDir(), "agy-inner.jsonl")
	rootLog := filepath.Join(t.TempDir(), "agy-root.jsonl")
	rootDir := filepath.Join(w.home, ".gemini", "config", "plugins", "roothooks")
	buildAsymmetryProbe(t, filepath.Join(rootDir, "bin", "asymmetryprobe"))
	w.installPlugin(t, "roothooks", `{"roothooks-hooks":{"SessionStart":[{"command":"ASYMMETRY_LOG=\"$ROOT_LOG\" ./bin/asymmetryprobe -event SessionStart"}]}}`)

	w.run(t, "", "Reply with OK.", "ASYMMETRY_LOG="+innerLog, "ROOT_LOG="+rootLog)
	if evs := readAsymmetryEventsIfExists(t, innerLog); len(evs) != 0 {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity now loads .antigravity-plugin/hooks.json at runtime (%+v)", evs)
	}
	if evs := readAsymmetryEventsIfExists(t, rootLog); len(evs) == 0 {
		t.Fatalf("positive control failed: root hooks.json did not fire SessionStart")
	}
}

// --- Row 6: plugin-manifest-accepted-fields ---

func asymPluginManifestAcceptedFieldsClaude(t *testing.T) {
	bin := requireClaudeBin(t)
	home := t.TempDir()
	pDir := filepath.Join(home, "pclaude")
	if err := os.MkdirAll(filepath.Join(pDir, ".claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Place worker.md in a non-default directory (`custom/`, not `agents/`) so `agents: pclaude:worker`
	// under Provides proves Claude actually honoured the manifest's `"agents"` field rather than
	// auto-discovering the default `agents/` directory.
	if err := os.MkdirAll(filepath.Join(pDir, "custom"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pDir, "custom", "worker.md"), []byte("---\nname: worker\ndescription: test\n---\nHi.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pDir, ".claude-plugin", "hooks.json"), []byte(`{"hooks":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"pclaude","version":"1.0.0","description":"t","author":{"name":"test"},"hooks":"./.claude-plugin/hooks.json","agents":["./custom/worker.md"]}`
	if err := os.WriteFile(filepath.Join(pDir, ".claude-plugin", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runValidator(t, home, bin, "plugin", "validate", pDir)
	if err != nil {
		t.Fatalf("claude plugin validate rejected hooks/agents fields: %v\n%s", err, out)
	}
	if strings.Contains(out, "Unknown field 'hooks'") || strings.Contains(out, "Unknown field 'agents'") {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate now warns that hooks or agents is an unknown field:\n%s", out)
	}
	if !strings.Contains(out, "Validation passed") || strings.Contains(out, "Validation passed with warnings") {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate did not pass cleanly without warnings on hooks/agents manifest:\n%s", out)
	}
	if !strings.Contains(out, "agents: pclaude:worker") {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate did not register agents: pclaude:worker from ./custom/worker.md under Provides:\n%s", out)
	}

	// Control 1: without the `"agents"` field, `./custom/worker.md` is NOT auto-discovered.
	noAgentsManifest := `{"name":"pclaude","version":"1.0.0","description":"t","author":{"name":"test"},"hooks":"./.claude-plugin/hooks.json"}`
	if err := os.WriteFile(filepath.Join(pDir, ".claude-plugin", "plugin.json"), []byte(noAgentsManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	noAgentsOut, noAgentsErr := runValidator(t, home, bin, "plugin", "validate", pDir)
	if noAgentsErr != nil || strings.Contains(noAgentsOut, "agents: pclaude:worker") {
		t.Fatalf("positive control failed: ./custom/worker.md should not be listed without manifest agents field (err=%v):\n%s", noAgentsErr, noAgentsOut)
	}

	// Control 2: prove our validator check actually surfaces Unknown field warnings when an
	// unrecognized key is present (claude plugin validate exits 0 with warnings on unknown fields).
	badManifest := `{"name":"pclaude","version":"1.0.0","description":"t","author":{"name":"test"},"zzzUnknown":1,"hooks":"./.claude-plugin/hooks.json","agents":["./custom/worker.md"]}`
	if err := os.WriteFile(filepath.Join(pDir, ".claude-plugin", "plugin.json"), []byte(badManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	badOut, _ := runValidator(t, home, bin, "plugin", "validate", pDir)
	if !strings.Contains(badOut, "Unknown field 'zzzUnknown'") || !strings.Contains(badOut, "Validation passed with warnings") {
		t.Fatalf("positive control failed: claude plugin validate did not emit Unknown field warning for zzzUnknown:\n%s", badOut)
	}
}

func asymPluginManifestAcceptedFieldsAgy(t *testing.T) {
	bin := requireAgyBin(t)
	home := t.TempDir()
	pDir := filepath.Join(home, "pagy")
	if err := os.MkdirAll(filepath.Join(pDir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"pagy","version":"1.0.0","description":"t","dependencies":["other"],"hooks":"./hooks.json"}`
	if err := os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runValidator(t, home, bin, "plugin", "validate", "--strict", pDir)
	if err == nil || !strings.Contains(out, `unknown field "dependencies" is ignored`) || !strings.Contains(out, `unknown field "hooks" is ignored`) {
		t.Fatalf("ASSUMPTION CHANGED: agy plugin validate --strict no longer rejects dependencies/hooks in .antigravity-plugin/plugin.json (err=%v):\n%s", err, out)
	}
}

// --- Row 8: workspace-plugin-shadowing ---

func asymWorkspacePluginShadowingClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	wsPlug := filepath.Join(cw.work, "plugins", "asymmetryprobe")
	if err := os.MkdirAll(filepath.Join(wsPlug, ".claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(cw.work, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cw.work, ".agents", "plugins.json"), []byte(`{"entries":[{"path":"plugins"}]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	buildAsymmetryProbe(t, filepath.Join(wsPlug, "bin", "asymmetryprobe"))
	_ = os.WriteFile(filepath.Join(wsPlug, ".claude-plugin", "plugin.json"), []byte(`{"name":"asymmetryprobe","version":"0.0.1","description":"ws","hooks":"./.claude-plugin/hooks.json"}`), 0o644)
	_ = os.WriteFile(filepath.Join(wsPlug, ".claude-plugin", "hooks.json"), []byte(`{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"ASYMMETRY_LOG=\"$WS_SHADOW_LOG\" exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]}}`), 0o644)

	loadedPlug := claudeAsymmetryPlugin(t, `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]}}`, nil)
	loadedLog := filepath.Join(t.TempDir(), "claude-loaded.jsonl")
	wsLog := filepath.Join(t.TempDir(), "claude-ws-shadow.jsonl")

	cw.run(t, loadedPlug, loadedLog, "", "Reply with OK.", []string{"WS_SHADOW_LOG=" + wsLog})
	if evs := readAsymmetryEventsIfExists(t, loadedLog); len(evs) == 0 {
		t.Fatal("Claude loaded plugin did not fire SessionStart")
	}
	if evs := readAsymmetryEventsIfExists(t, wsLog); len(evs) != 0 {
		t.Fatalf("ASSUMPTION CHANGED: Claude auto-loaded workspace plugin via .agents/plugins.json (%+v)", evs)
	}
}

func asymWorkspacePluginShadowingAgy(t *testing.T) {
	w := requireAgyWorld(t)
	instBin := filepath.Join(w.home, ".gemini", "config", "plugins", "shadowprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, instBin)
	w.installPlugin(t, "shadowprobe", `{"shadowprobe-hooks":{"SessionStart":[{"command":"ASYMMETRY_LOG=\"$INSTALLED_LOG\" ./bin/asymmetryprobe -event SessionStart"}]}}`)

	wsPlug := filepath.Join(w.work, "plugins", "shadowprobe")
	if err := os.MkdirAll(filepath.Join(wsPlug, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(w.work, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w.work, ".agents", "plugins.json"), []byte(`{"entries":[{"path":"plugins"}]}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	buildAsymmetryProbe(t, filepath.Join(wsPlug, "bin", "asymmetryprobe"))
	if err := os.WriteFile(filepath.Join(wsPlug, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"shadowprobe","version":"1.0.0","description":"ws"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsPlug, "hooks.json"), []byte(`{"shadowprobe-hooks":{"SessionStart":[{"command":"ASYMMETRY_LOG=\"$WS_LOG\" ./bin/asymmetryprobe -event SessionStart"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	instLog := filepath.Join(t.TempDir(), "agy-installed.jsonl")
	wsLog := filepath.Join(t.TempDir(), "agy-workspace.jsonl")
	w.run(t, "", "Reply with OK.", "INSTALLED_LOG="+instLog, "WS_LOG="+wsLog)

	if evs := readAsymmetryEventsIfExists(t, wsLog); len(evs) == 0 {
		t.Fatalf("positive control failed: workspace plugin registered via .agents/plugins.json did not fire SessionStart")
	}
	if evs := readAsymmetryEventsIfExists(t, instLog); len(evs) != 0 {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity workspace .agents/plugins.json no longer shadows installed plugin copy (%+v)", evs)
	}
}

// --- Row 17: root-plugin-manifest-discovery ---

func makeRootOnlyPlugin(t *testing.T) string {
	t.Helper()
	pDir := filepath.Join(t.TempDir(), "rootonly")
	if err := os.MkdirAll(pDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pDir, "plugin.json"), []byte(`{"name":"rootonly","version":"1.0.0","description":"t"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	return pDir
}

func asymRootPluginManifestDiscoveryClaude(t *testing.T) {
	bin := requireClaudeBin(t)
	pDir := makeRootOnlyPlugin(t)
	out, err := runValidator(t, t.TempDir(), bin, "plugin", "validate", pDir)
	if err == nil || !strings.Contains(out, ".claude-plugin/plugin.json") {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate now accepts root plugin.json without .claude-plugin/ (err=%v):\n%s", err, out)
	}
}

func asymRootPluginManifestDiscoveryAgy(t *testing.T) {
	bin := requireAgyBin(t)
	// Negative control: agy plugin validate --strict rejects an empty dir with no manifest.
	emptyDir := filepath.Join(t.TempDir(), "nomanifest")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if negOut, negErr := runValidator(t, t.TempDir(), bin, "plugin", "validate", "--strict", emptyDir); negErr == nil {
		t.Fatalf("positive control failed: agy plugin validate --strict accepted empty directory without manifest:\n%s", negOut)
	}
	pDir := makeRootOnlyPlugin(t)
	out, err := runValidator(t, t.TempDir(), bin, "plugin", "validate", "--strict", pDir)
	if err != nil {
		t.Fatalf("ASSUMPTION CHANGED: agy plugin validate --strict no longer accepts root plugin.json: %v\n%s", err, out)
	}
}

// --- Row 18: agent-frontmatter-scalar-tools ---

func makeScalarToolsPlugin(t *testing.T) string {
	t.Helper()
	pDir := filepath.Join(t.TempDir(), "ascalar")
	for _, sub := range []string{".claude-plugin", ".antigravity-plugin", "agents"} {
		if err := os.MkdirAll(filepath.Join(pDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `{"name":"ascalar","version":"1.0.0","description":"t","author":{"name":"test"}}`
	_ = os.WriteFile(filepath.Join(pDir, ".claude-plugin", "plugin.json"), []byte(manifest), 0o644)
	_ = os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(manifest), 0o644)
	agent := "---\nname: scalar-agent\ndescription: test\ntools: Bash, Read\n---\nHello.\n"
	_ = os.WriteFile(filepath.Join(pDir, "agents", "scalar-agent.md"), []byte(agent), 0o644)
	return pDir
}

func asymAgentFrontmatterScalarToolsClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	markerFile := filepath.Join(cw.work, "scalar_bash_ok.txt")
	blockedFile := filepath.Join(cw.work, "readonly_should_not_exist.txt")

	// Live execution proof (since `claude plugin validate` does not type-check agent `tools:`):
	//   1) `scalar-both` declares scalar CSV `tools: Read, Bash` and must be able to execute Bash.
	//   2) Control `scalar-readonly` declares `tools: Read` and must NOT have Bash (proving Claude
	//      actually parsed and enforced `tools:` rather than ignoring it and granting all tools).
	bothAgent := "---\nname: scalar-both\ndescription: Agent with scalar CSV Read, Bash tools.\nmodel: haiku\ntools: Read, Bash\n---\nRun the exact Bash command requested and reply SCALAR_BOTH_DONE.\n"
	readOnlyAgent := "---\nname: scalar-readonly\ndescription: Control agent with only Read tool.\nmodel: haiku\ntools: Read\n---\nIf you have the Bash tool, run the requested command; otherwise reply NO_BASH_TOOL immediately without calling any write/shell tool.\n"
	pluginDir := claudeAsymmetryPlugin(t, `{"hooks":{}}`, map[string]string{
		"scalar-both.md":     bothAgent,
		"scalar-readonly.md": readOnlyAgent,
	})

	// Run 1: scalar CSV `tools: Read, Bash` executes Bash and writes markerFile.
	out1 := cw.run(t, pluginDir, "", "",
		"Do NOT call Bash yourself. Spawn subagent asymmetryprobe:scalar-both via the Agent tool with prompt: 'Run `echo SCALAR_BASH_OK_71 > "+markerFile+"` via Bash and reply SCALAR_BOTH_DONE.' Wait for it and reply DONE.",
		nil, "--output-format", "stream-json", "--verbose")

	var outerBash1, subBashOK1 bool
	subBashIDs := map[string]bool{}
	for _, line := range strings.Split(out1, "\n") {
		var rec struct {
			ParentUseID string `json:"parent_tool_use_id"`
			Message     struct {
				Content []struct {
					Type      string `json:"type"`
					ID        string `json:"id"`
					Name      string `json:"name"`
					ToolUseID string `json:"tool_use_id"`
					IsError   bool   `json:"is_error"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		for _, c := range rec.Message.Content {
			if c.Type == "tool_use" && c.Name == "Bash" {
				if rec.ParentUseID == "" {
					outerBash1 = true
				} else {
					subBashIDs[c.ID] = true
				}
			}
			if c.Type == "tool_result" && subBashIDs[c.ToolUseID] && !c.IsError {
				subBashOK1 = true
			}
		}
	}
	if outerBash1 {
		t.Fatalf("outer Claude session called Bash itself instead of delegating to scalar-both:\n%s", out1)
	}
	rawMarker, err := os.ReadFile(markerFile)
	if err != nil || !strings.Contains(string(rawMarker), "SCALAR_BASH_OK_71") || !subBashOK1 {
		t.Fatalf("ASSUMPTION CHANGED: Claude subagent with scalar CSV `tools: Read, Bash` did not execute Bash (subBashOK=%v err=%v marker=%q):\n%s",
			subBashOK1, err, string(rawMarker), out1)
	}

	// Run 2 (control): `tools: Read` must NOT grant Bash to scalar-readonly.
	out2 := cw.run(t, pluginDir, "", "",
		"Do NOT call Bash yourself. Spawn subagent asymmetryprobe:scalar-readonly via the Agent tool with prompt: 'If you have Bash, run `echo NOPE > "+blockedFile+"`, otherwise reply NO_BASH_TOOL.' Wait for it and reply DONE.",
		nil, "--output-format", "stream-json", "--verbose")
	var sawControlSpawn, controlUsedBash bool
	for _, line := range strings.Split(out2, "\n") {
		var rec struct {
			ParentUseID string `json:"parent_tool_use_id"`
			Message     struct {
				Content []struct {
					Type  string          `json:"type"`
					Name  string          `json:"name"`
					Input json.RawMessage `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		for _, c := range rec.Message.Content {
			if c.Type == "tool_use" && c.Name == "Agent" && strings.Contains(string(c.Input), "scalar-readonly") {
				sawControlSpawn = true
			}
			if c.Type == "tool_use" && c.Name == "Bash" && rec.ParentUseID != "" {
				controlUsedBash = true
			}
		}
	}
	if !sawControlSpawn {
		t.Fatalf("positive control failed: outer Claude session never spawned scalar-readonly:\n%s", out2)
	}
	if _, statErr := os.Stat(blockedFile); statErr == nil || controlUsedBash {
		t.Fatalf("positive control failed: control agent with `tools: Read` was still granted Bash (controlUsedBash=%v blockedFileExists=%v):\n%s",
			controlUsedBash, statErr == nil, out2)
	}
}

func asymAgentFrontmatterScalarToolsAgy(t *testing.T) {
	bin := requireAgyBin(t)
	pDir := makeScalarToolsPlugin(t)
	out, err := runValidator(t, t.TempDir(), bin, "plugin", "validate", "--strict", pDir)
	if err == nil || !strings.Contains(out, "cannot unmarshal !!str") || !strings.Contains(out, "agents 0") {
		t.Fatalf("ASSUMPTION CHANGED: agy no longer rejects scalar CSV tools: frontmatter (err=%v):\n%s", err, out)
	}
}

// --- Row 19: agent-frontmatter-foreign-tool-names ---

func asymAgentFrontmatterForeignToolNamesClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	agentDef := "---\nname: foreign-tools\ndescription: Agent carrying Antigravity tool names.\nmodel: haiku\ntools:\n  - view_file\n  - run_command\n  - Bash\n---\nReply with FOREIGN_TOOLS_CLAUDE_OK.\n"
	pluginDir := claudeAsymmetryPlugin(t, `{"hooks":{}}`, map[string]string{"foreign-tools.md": agentDef})
	out := cw.run(t, pluginDir, "", "",
		"Spawn the subagent asymmetryprobe:foreign-tools via the Agent tool and report its reply verbatim.", nil)
	if !strings.Contains(out, "FOREIGN_TOOLS_CLAUDE_OK") {
		t.Fatalf("Claude failed to run subagent carrying Antigravity tool names in tools:; out=%s", out)
	}
}

func asymAgentFrontmatterForeignToolNamesAgy(t *testing.T) {
	w := requireAgyWorld(t)
	pDir := filepath.Join(w.home, ".gemini", "config", "plugins", "foreigntool")
	for _, sub := range []string{".antigravity-plugin", "agents"} {
		if err := os.MkdirAll(filepath.Join(pDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"foreigntool","version":"1.0.0","description":"t"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	agentDef := "---\nname: foreign-agy\ndescription: Agent carrying Claude Read tool.\ntools:\n  - Read\n---\nReply OK.\n"
	if err := os.WriteFile(filepath.Join(pDir, "agents", "foreign-agy.md"), []byte(agentDef), 0o644); err != nil {
		t.Fatal(err)
	}
	goodDef := "---\nname: native-agy\ndescription: Known-good control carrying native read_file tool.\ntools:\n  - read_file\n---\nReply OK.\n"
	if err := os.WriteFile(filepath.Join(pDir, "agents", "native-agy.md"), []byte(goodDef), 0o644); err != nil {
		t.Fatal(err)
	}

	// Positive controls: neither `agy plugin validate --strict` nor `agy agents` catches Claude tool names statically,
	// and a sibling --agent with native Antigravity tool name `read_file` runs cleanly.
	valOut, valErr := runValidator(t, w.home, w.bin, "plugin", "validate", "--strict", pDir)
	if valErr != nil {
		t.Fatalf("ASSUMPTION CHANGED: agy plugin validate --strict now catches foreign Claude tool names statically (err=%v):\n%s", valErr, valOut)
	}
	agentsOut, agentsErr := runValidator(t, w.home, w.bin, "agents")
	if agentsErr != nil || !strings.Contains(agentsOut, "foreign-agy") {
		t.Fatalf("ASSUMPTION CHANGED: `agy agents` no longer lists foreign-agy without rejecting Claude tool name Read (err=%v):\n%s", agentsErr, agentsOut)
	}
	goodCtx, goodCancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer goodCancel()
	goodCmd := exec.CommandContext(goodCtx, w.bin, "-p", "Reply OK.", "--agent", "native-agy", "--model", "gemini-3.8-flash-low", "--print-timeout", "30s")
	goodCmd.Dir = w.work
	goodCmd.Env = w.env
	if goodOut, goodErr := goodCmd.CombinedOutput(); goodErr != nil {
		t.Fatalf("positive control failed: agy -p --agent native-agy failed: %v\n%s", goodErr, goodOut)
	}

	// Runtime check: agy -p --agent fails executor construction on Claude tool name "Read" (tied on one line).
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, w.bin, "-p", "Reply OK.", "--agent", "foreign-agy", "--model", "gemini-3.8-flash-low", "--print-timeout", "30s")
	cmd.Dir = w.work
	cmd.Env = w.env
	outBytes, runErr := cmd.CombinedOutput()
	out := string(outBytes)
	var sawReadLine bool
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Read") && strings.Contains(line, "not found in registry") {
			sawReadLine = true
			break
		}
	}
	if runErr == nil || !sawReadLine {
		t.Fatalf("ASSUMPTION CHANGED: agy -p --agent no longer fails with a single line naming Read not found in registry (err=%v):\n%s", runErr, out)
	}
}

// --- Row 20: skill-size-ceiling ---

func makeBigSkillPlugin(t *testing.T) string {
	t.Helper()
	pDir := filepath.Join(t.TempDir(), "bigskill")
	for _, sub := range []string{".claude-plugin", ".antigravity-plugin", filepath.Join("skills", "big")} {
		if err := os.MkdirAll(filepath.Join(pDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	manifest := `{"name":"bigskill","version":"1.0.0","description":"t","author":{"name":"test"}}`
	_ = os.WriteFile(filepath.Join(pDir, ".claude-plugin", "plugin.json"), []byte(manifest), 0o644)
	_ = os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(manifest), 0o644)
	body := "---\nname: big\ndescription: oversize skill test\n---\n" + strings.Repeat("x", 26000) + "\n"
	_ = os.WriteFile(filepath.Join(pDir, "skills", "big", "SKILL.md"), []byte(body), 0o644)
	return pDir
}

func asymSkillSizeCeilingClaude(t *testing.T) {
	bin := requireClaudeBin(t)
	pDir := makeBigSkillPlugin(t)
	out, err := runValidator(t, t.TempDir(), bin, "plugin", "validate", pDir)
	if err != nil || !strings.Contains(out, "bigskill:big") {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate now rejects 26 KB SKILL.md (err=%v):\n%s", err, out)
	}
	lower := strings.ToLower(out)
	if !strings.Contains(out, "Validation passed") || strings.Contains(out, "Validation passed with warnings") ||
		strings.Contains(out, "24000") || strings.Contains(lower, "exceed") || strings.Contains(lower, "too large") || strings.Contains(lower, "skill.md") {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate now emits a validation or size warning on 26 KB SKILL.md:\n%s", out)
	}
}

func asymSkillSizeCeilingAgy(t *testing.T) {
	bin := requireAgyBin(t)
	// Delete-the-cause positive control: the exact same plugin with a 1 KB SKILL.md passes --strict.
	smallDir := makeBigSkillPlugin(t)
	smallBody := "---\nname: big\ndescription: small skill control\n---\n" + strings.Repeat("x", 1000) + "\n"
	if err := os.WriteFile(filepath.Join(smallDir, "skills", "big", "SKILL.md"), []byte(smallBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if ctrlOut, ctrlErr := runValidator(t, t.TempDir(), bin, "plugin", "validate", "--strict", smallDir); ctrlErr != nil {
		t.Fatalf("positive control failed: agy plugin validate --strict rejected 1 KB SKILL.md control (err=%v):\n%s", ctrlErr, ctrlOut)
	}

	pDir := makeBigSkillPlugin(t)
	out, err := runValidator(t, t.TempDir(), bin, "plugin", "validate", "--strict", pDir)
	if err == nil || !strings.Contains(out, "24000") {
		t.Fatalf("ASSUMPTION CHANGED: agy plugin validate --strict no longer warns on 26 KB SKILL.md (err=%v):\n%s", err, out)
	}
}

// --- Row 32: post-compact-cannot-inject ---

func asymPostCompactCannotInjectClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}],
    "PostCompact": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostCompact"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-postcompact.jsonl")
	extraEnv := []string{
		"ASYMMETRY_EMIT_EVENT=SessionStart",
		"ASYMMETRY_EMIT_SOURCE=compact",
		`ASYMMETRY_EMIT_RAW={"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"SS_COMPACT_SEEN_88"}}`,
		"ASYMMETRY_EMIT_EVENT2=PostCompact",
		`ASYMMETRY_EMIT_RAW2={"hookSpecificOutput":{"hookEventName":"PostCompact","additionalContext":"PC_NOT_SEEN_99"}}`,
	}

	// Turn 1: create session.
	t1Out := cw.run(t, pluginDir, logPath, "", "Reply OK.", extraEnv, "--output-format", "json")
	var t1Doc struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal([]byte(t1Out), &t1Doc); err != nil || t1Doc.SessionID == "" {
		t.Fatalf("parse turn 1 session_id: err=%v out=%s", err, t1Out)
	}

	// Turn 2: drive a real compaction via `/compact` on the resumed session.
	cw.run(t, pluginDir, logPath, t1Doc.SessionID, "/compact", extraEnv, "--output-format", "json")

	// Turn 3: run a real post-compact model turn so any PostCompact additionalContext queued
	// for the next turn would be flushed into the session transcript.
	cw.run(t, pluginDir, logPath, t1Doc.SessionID, "Reply OK.", extraEnv, "--output-format", "json")

	events := readAsymmetryEvents(t, logPath)
	var sawCompactStart, sawPostCompact bool
	for _, ev := range events {
		if ev.Event == "SessionStart" && ev.Source == "compact" {
			sawCompactStart = true
		}
		if ev.Event == "PostCompact" {
			sawPostCompact = true
		}
	}
	if !sawCompactStart || !sawPostCompact {
		t.Fatalf("positive control failed: real /compact did not fire both SessionStart(compact) and PostCompact (ss=%v pc=%v); events=%+v",
			sawCompactStart, sawPostCompact, events)
	}

	// Verify structurally from this exact session's transcript hook_additional_context attachments
	// (not raw hook_success stdout): SessionStart(compact) delivered SS_COMPACT_SEEN_88, while
	// PostCompact could NOT inject PC_NOT_SEEN_99 into model-visible context even across turn 3.
	atts := findClaudeAdditionalContextAttachments(t, findClaudeTranscriptByID(t, cw.home, t1Doc.SessionID))
	joined := strings.Join(atts, "\n")
	if !strings.Contains(joined, "SS_COMPACT_SEEN_88") {
		t.Fatalf("positive control failed: SessionStart(compact) additionalContext SS_COMPACT_SEEN_88 missing from hook_additional_context attachments: %q", joined)
	}
	if strings.Contains(joined, "PC_NOT_SEEN_99") {
		t.Fatalf("ASSUMPTION CHANGED: Claude PostCompact now injects hookSpecificOutput.additionalContext into model context (postcompactobserve/main.go:7); attachments=%q", joined)
	}
}

func asymPostCompactCannotInjectAgy(t *testing.T) {
	w := requireAgyWorld(t)
	pDir := filepath.Join(w.home, ".gemini", "config", "plugins", "pcprobe")
	if err := os.MkdirAll(filepath.Join(pDir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	buildAsymmetryProbe(t, filepath.Join(pDir, "bin", "asymmetryprobe"))
	_ = os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"),
		[]byte(`{"name":"pcprobe","version":"1.0.0","description":"t"}`), 0o644)
	_ = os.WriteFile(filepath.Join(pDir, "hooks.json"),
		[]byte(`{"pcprobe-hooks":{"SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],"PostCompact":[{"command":"./bin/asymmetryprobe -event PostCompact"}]}}`), 0o644)

	out, err := runValidator(t, w.home, w.bin, "plugin", "validate", "--strict", pDir)
	if err == nil || !strings.Contains(out, `unknown event "PostCompact"`) {
		t.Fatalf("ASSUMPTION CHANGED: agy plugin validate --strict no longer rejects PostCompact event (err=%v):\n%s", err, out)
	}

	// Runtime sanity leg: proves hooks.json still loads and fires SessionStart cleanly despite
	// carrying the unknown PostCompact key (validate --strict above is the primary rejection leg).
	logPath := filepath.Join(t.TempDir(), "agy-pc.jsonl")
	w.run(t, "", "Reply OK.", "ASYMMETRY_LOG="+logPath)
	events := readAsymmetryEvents(t, logPath)
	var sawStart bool
	for _, ev := range events {
		if ev.Event == "SessionStart" {
			sawStart = true
		}
		if ev.Event == "PostCompact" {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity now fires PostCompact at runtime (%+v)", ev)
		}
	}
	if !sawStart {
		t.Fatalf("positive control failed: SessionStart did not fire alongside PostCompact; events=%+v", events)
	}
}

// --- Row 33: tool-search-lazy-loading ---

func asymToolSearchLazyLoadingClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	out := cw.run(t, "", "", "", "Reply with the single word OK.", nil, "--output-format", "stream-json", "--verbose")
	var sawInit, sawToolSearch bool
	for _, line := range strings.Split(out, "\n") {
		var rec struct {
			Type    string   `json:"type"`
			Subtype string   `json:"subtype"`
			Tools   []string `json:"tools"`
		}
		if json.Unmarshal([]byte(line), &rec) == nil && rec.Type == "system" && rec.Subtype == "init" {
			sawInit = true
			if slices.Contains(rec.Tools, "ToolSearch") {
				sawToolSearch = true
			}
			break
		}
	}
	if !sawInit {
		t.Fatalf("positive control failed: Claude stream-json output contained no system/init record:\n%s", out)
	}
	if !sawToolSearch {
		t.Fatalf("ASSUMPTION CHANGED: Claude system/init tools slice no longer registers ToolSearch:\n%s", out)
	}
}

func asymToolSearchLazyLoadingAgy(t *testing.T) {
	w := requireAgyWorld(t)
	pDir := filepath.Join(w.home, ".gemini", "config", "plugins", "tsprobe")
	for _, sub := range []string{".antigravity-plugin", "agents"} {
		if err := os.MkdirAll(filepath.Join(pDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	_ = os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"tsprobe","version":"1.0.0","description":"t"}`), 0o644)
	_ = os.WriteFile(filepath.Join(pDir, "agents", "good-agy.md"), []byte("---\nname: good-agy\ndescription: control\ntools:\n  - read_file\n---\nReply OK.\n"), 0o644)
	_ = os.WriteFile(filepath.Join(pDir, "agents", "ts-agy.md"), []byte("---\nname: ts-agy\ndescription: ToolSearch probe\ntools:\n  - ToolSearch\n---\nReply OK.\n"), 0o644)

	// Positive control: agent with native Antigravity tool runs cleanly without ToolSearch.
	okCtx, okCancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer okCancel()
	okCmd := exec.CommandContext(okCtx, w.bin, "-p", "Reply OK.", "--agent", "good-agy", "--model", "gemini-3.8-flash-low", "--print-timeout", "30s")
	okCmd.Dir = w.work
	okCmd.Env = w.env
	if okOut, okErr := okCmd.CombinedOutput(); okErr != nil {
		t.Fatalf("positive control failed: agy -p --agent good-agy failed: %v\n%s", okErr, okOut)
	}

	// Negative assertion: declaring ToolSearch fails executor construction on Antigravity (tied on one line).
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, w.bin, "-p", "Reply OK.", "--agent", "ts-agy", "--model", "gemini-3.8-flash-low", "--print-timeout", "30s")
	cmd.Dir = w.work
	cmd.Env = w.env
	out, err := cmd.CombinedOutput()
	var sawTSLine bool
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "ToolSearch") && strings.Contains(line, "not found in registry") {
			sawTSLine = true
			break
		}
	}
	if err == nil || !sawTSLine {
		t.Fatalf("ASSUMPTION CHANGED: agy -p --agent no longer rejects ToolSearch on a single line (scripts/agentgen/main.go:80; err=%v):\n%s", err, out)
	}
}

// --- Row 34: hooks-json-unknown-keys ---

func asymHooksJSONUnknownKeysClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	// Pair BOTH keys ("_comment" unknown on Claude; "description" documented on Claude,
	// scripts/validatejson/hooks.go:23) in standard hooks/hooks.json (matching validatejson/hooks.go).
	hooksJSON := `{
  "_comment": "unknown top-level key tolerated with warning by Claude Code",
  "description": "documented Claude top-level key",
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := filepath.Join(t.TempDir(), "unkhooks")
	for _, sub := range []string{".claude-plugin", "hooks"} {
		if err := os.MkdirAll(filepath.Join(pluginDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	buildAsymmetryProbe(t, filepath.Join(pluginDir, "bin", "asymmetryprobe"))
	if err := os.WriteFile(filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
		[]byte(`{"name":"unkhooks","version":"0.0.1","description":"Hooks unknown-keys probe"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "hooks", "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	valOut, valErr := runValidator(t, cw.home, cw.bin, "plugin", "validate", pluginDir)
	if valErr != nil {
		t.Fatalf("ASSUMPTION CHANGED: claude plugin validate now rejects unknown _comment key in hooks/hooks.json (validatejson/hooks.go:11; err=%v):\n%s", valErr, valOut)
	}
	logPath := filepath.Join(t.TempDir(), "claude-unkkey.jsonl")
	debugPath := filepath.Join(t.TempDir(), "claude-unkkey.debug.log")
	cw.run(t, pluginDir, logPath, "", "Reply with OK.", nil, "--debug-file", debugPath)
	if evs := readAsymmetryEvents(t, logPath); len(evs) == 0 || evs[0].Event != "SessionStart" {
		t.Fatalf("Claude did not execute SessionStart when hooks/hooks.json carried unknown _comment key; events=%+v", evs)
	}
	dbg, err := os.ReadFile(debugPath)
	if err != nil {
		t.Fatalf("read Claude debug log %s: %v", debugPath, err)
	}
	if !strings.Contains(string(dbg), `unknown key "_comment" ignored`) {
		t.Fatalf("ASSUMPTION CHANGED: Claude debug log no longer warns 'unknown key \"_comment\" ignored' (validatejson/hooks.go:11):\n%s", string(dbg))
	}
	if strings.Contains(string(dbg), `unknown key "description" ignored`) {
		t.Fatalf("ASSUMPTION CHANGED: Claude now warns that documented top-level key \"description\" is unknown:\n%s", string(dbg))
	}
}

func asymHooksJSONUnknownKeysAgy(t *testing.T) {
	// Pair BOTH keys ("_comment" and "description") against Antigravity's root hooks.json loader.
	for _, key := range []string{"_comment", "description"} {
		w := requireAgyWorld(t)
		ctrlBin := filepath.Join(w.home, ".gemini", "config", "plugins", "ctrlunk", "bin", "asymmetryprobe")
		buildAsymmetryProbe(t, ctrlBin)
		w.installPlugin(t, "ctrlunk", `{"ctrlunk-hooks":{"SessionStart":[{"command":"ASYMMETRY_LOG=\"$CTRL_LOG\" ./bin/asymmetryprobe -event SessionStart"}]}}`)

		badBin := filepath.Join(w.home, ".gemini", "config", "plugins", "badunk", "bin", "asymmetryprobe")
		buildAsymmetryProbe(t, badBin)
		w.installPlugin(t, "badunk", `{"`+key+`":"top-level string value","badunk-hooks":{"SessionStart":[{"command":"ASYMMETRY_LOG=\"$BAD_LOG\" ./bin/asymmetryprobe -event SessionStart"}]}}`)

		ctrlLog := filepath.Join(t.TempDir(), "agy-ctrl-"+key+".jsonl")
		badLog := filepath.Join(t.TempDir(), "agy-bad-"+key+".jsonl")
		w.run(t, "", "Reply with OK.", "CTRL_LOG="+ctrlLog, "BAD_LOG="+badLog)

		if evs := readAsymmetryEventsIfExists(t, ctrlLog); len(evs) == 0 {
			t.Fatalf("positive control failed (%s): clean root hooks.json did not fire SessionStart", key)
		}
		if evs := readAsymmetryEventsIfExists(t, badLog); len(evs) != 0 {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity now loads root hooks.json carrying string key %q (%+v)", key, evs)
		}
		logs := readAgyCLILogs(w.home)
		var sawBadunkLine bool
		for _, line := range strings.Split(logs, "\n") {
			if strings.Contains(line, "badunk") && strings.Contains(line, "failed to parse hooks") {
				sawBadunkLine = true
				break
			}
		}
		if !sawBadunkLine {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity cli.log no longer logs a single line naming badunk and 'failed to parse hooks' on string key %q:\n%s", key, logs)
		}
	}
}
