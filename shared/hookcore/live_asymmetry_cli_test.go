package hookcore_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

// --- Row 4: standalone-server-daemon-wake-route ---

func asymStandaloneWakeRouteClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-standalone.jsonl")

	cw.run(t, pluginDir, logPath, "", "Reply with OK.", nil)
	events := readAsymmetryEvents(t, logPath)
	rec, ok := parseMatchingClaudeSessionRecord(events)
	if !ok {
		t.Fatalf("standalone Claude session did not advertise ~/.claude/sessions/<pid>.json with pid/sessionId/cwd; events=%+v", events)
	}
	if rec.CWD != cw.work {
		t.Fatalf("Claude session file cwd = %q, want %q", rec.CWD, cw.work)
	}
	// Distinct from hook-wake-credentials: verify the session advertisement file reflects process
	// lifecycle — once standalone claude -p exits, the file is removed so external callers never
	// mistake an exited standalone session for a live route.
	sessFile := filepath.Join(cw.home, ".claude", "sessions", fmt.Sprintf("%d.json", rec.PID))
	if _, err := os.Stat(sessFile); err == nil {
		t.Fatalf("ASSUMPTION CHANGED: standalone Claude left %s behind after process exit", sessFile)
	}
}

func asymStandaloneWakeRouteAgy(t *testing.T) {
	w := requireAgyWorld(t)
	reportFile := filepath.Join(w.work, "live_probe.txt")
	// Plant a canary file inside $HOME/.gemini/antigravity-cli so grep must prove the tree was
	// genuinely scanned (not silently skipped via error suppression).
	canaryDir := filepath.Join(w.home, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(canaryDir, 0o700); err != nil {
		t.Fatal(err)
	}
	canaryFile := filepath.Join(canaryDir, ".scan_canary")
	if err := os.WriteFile(canaryFile, []byte("WAKE_SCAN_CANARY_88\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Inspect LIVE standalone agy state from inside run_command while agy is actively running
	// (using headless `agy -p` as a proxy for non-daemon standalone agy):
	// 1) verify ANTIGRAVITY_LS_ADDRESS and ANTIGRAVITY_CSRF_TOKEN are present in the live shell,
	// 2) verify the grep scan over $HOME/.gemini actually finds the planted canary file,
	// 3) verify NO file anywhere under $HOME/.gemini contains either wake credential on disk.
	script := `set -e
canary_hit="$(grep -rlF "WAKE_SCAN_CANARY_88" "$HOME/.gemini" 2>/dev/null | head -n 1 || true)"
addr_hit="$(grep -rlF "$ANTIGRAVITY_LS_ADDRESS" "$HOME/.gemini" 2>/dev/null | head -n 1 || true)"
tok_hit="$(grep -rlF "$ANTIGRAVITY_CSRF_TOKEN" "$HOME/.gemini" 2>/dev/null | head -n 1 || true)"
printf 'HAS_ADDR=%s\nHAS_TOK=%s\nCANARY_FILE=%s\nADDR_FILE=%s\nTOK_FILE=%s\n' \
  "$([ -n "$ANTIGRAVITY_LS_ADDRESS" ] && echo yes || echo no)" \
  "$([ -n "$ANTIGRAVITY_CSRF_TOKEN" ] && echo yes || echo no)" \
  "$canary_hit" \
  "$addr_hit" \
  "$tok_hit" > "` + reportFile + `"`
	scriptPath := filepath.Join(w.work, "probe.sh")
	if err := os.WriteFile(scriptPath, []byte(script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	w.run(t, "", "Run `sh "+scriptPath+"` via run_command, then reply DONE.")

	raw, err := os.ReadFile(reportFile)
	if err != nil {
		t.Fatalf("positive control failed: standalone agy did not write %s: %v", reportFile, err)
	}
	kv := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			kv[k] = v
		}
	}
	if kv["HAS_ADDR"] != "yes" || kv["HAS_TOK"] != "yes" {
		t.Fatalf("positive control failed: live standalone agy shell lacked wake env vars:\n%s", string(raw))
	}
	if kv["CANARY_FILE"] != canaryFile {
		t.Fatalf("positive control failed: $HOME/.gemini scan did not find planted canary %q (got %q):\n%s", canaryFile, kv["CANARY_FILE"], string(raw))
	}
	if kv["ADDR_FILE"] != "" || kv["TOK_FILE"] != "" {
		t.Fatalf("ASSUMPTION CHANGED: standalone agy now advertises wake credentials on disk under ~/.gemini (addr_file=%q tok_file=%q)", kv["ADDR_FILE"], kv["TOK_FILE"])
	}
}

// --- Row 9: remote-control-serve-help ---

func asymRemoteControlServeHelpClaude(t *testing.T) {
	bin := requireClaudeBin(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "remote-control", "--help")
	cmd.Env = liveTestEnv(t, testhome.NewLiveHome(t))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("claude remote-control --help failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "USAGE") && !strings.Contains(string(out), "Usage") {
		t.Fatalf("claude remote-control --help printed no usage header:\n%s", out)
	}
}

func asymRemoteControlServeHelpAgy(t *testing.T) {
	bin := requireAgyBin(t)
	emptyHome := testhome.NewLiveHome(t)
	env := liveTestEnv(t, emptyHome)

	// Positive control: `agy remote-control --help` prints usage and exits 0.
	ctrlCtx, ctrlCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer ctrlCancel()
	ctrlCmd := exec.CommandContext(ctrlCtx, bin, "remote-control", "--help")
	ctrlCmd.Dir = emptyHome
	ctrlCmd.Env = env
	ctrlOut, ctrlErr := ctrlCmd.CombinedOutput()
	if ctrlErr != nil || (!strings.Contains(string(ctrlOut), "Usage:") && !strings.Contains(string(ctrlOut), "USAGE")) {
		t.Fatalf("positive control failed: `agy remote-control --help` did not print usage and exit 0 (err=%v):\n%s", ctrlErr, ctrlOut)
	}

	// Negative assertion: `agy remote-control serve --help` ignores --help and enters serve's
	// sign-in/startup path (exiting non-zero with "not signed in" in an empty HOME rather than
	// printing usage and exiting 0; upstream #1742).
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "remote-control", "serve", "--help")
	cmd.Dir = emptyHome
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err == nil || strings.Contains(string(out), "USAGE") || strings.Contains(string(out), "Usage:") || !strings.Contains(string(out), "not signed in") {
		t.Fatalf("ASSUMPTION CHANGED: `agy remote-control serve --help` printed usage or no longer entered serve sign-in check (upstream #1742 may be fixed; err=%v):\n%s", err, out)
	}
}

// --- Row 21: logged-out-session-start ---

func asymLoggedOutSessionStartClaude(t *testing.T) {
	bin := requireClaudeBin(t)
	emptyHome := testhome.NewLiveHome(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-loggedout.jsonl")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-p", "Reply OK.", "--model", "haiku", "--dangerously-skip-permissions", "--plugin-dir", pluginDir)
	cmd.Dir = t.TempDir()
	cmd.Env = liveTestEnv(t, emptyHome, "ASYMMETRY_LOG="+logPath, "CLAUDE_PLUGIN_DATA="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("ASSUMPTION CHANGED: logged-out claude -p hung until timeout (%v) instead of exiting cleanly:\n%s", ctx.Err(), out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected logged-out claude -p to exit non-zero with ExitError, got err=%v:\n%s", err, out)
	}
	evs := readAsymmetryEventsIfExists(t, logPath)
	if len(evs) == 0 || evs[0].Event != "SessionStart" {
		t.Fatalf("ASSUMPTION CHANGED: logged-out claude -p no longer fires SessionStart before exiting (out=%s)", out)
	}
}

func asymLoggedOutSessionStartAgy(t *testing.T) {
	bin := requireAgyBin(t)
	emptyHome := testhome.NewLiveHome(t)
	pDir := filepath.Join(emptyHome, ".gemini", "config", "plugins", "asymmetryprobe")
	if err := os.MkdirAll(filepath.Join(pDir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	buildAsymmetryProbe(t, filepath.Join(pDir, "bin", "asymmetryprobe"))
	_ = os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"asymmetryprobe","version":"1.0.0","description":"t"}`), 0o644)
	_ = os.WriteFile(filepath.Join(pDir, "hooks.json"), []byte(`{"asymmetryprobe-hooks":{"SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}]}}`), 0o644)
	logPath := filepath.Join(t.TempDir(), "agy-loggedout.jsonl")

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "-p", "Reply OK.", "--model", "gemini-3.8-flash-low")
	cmd.Dir = t.TempDir()
	cmd.Stdin = nil
	cmd.Env = liveTestEnv(t, emptyHome, "ASYMMETRY_LOG="+logPath)
	out, _ := cmd.CombinedOutput()
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("ASSUMPTION CHANGED: logged-out agy -p did not block waiting for interactive login (ctx.Err=%v):\n%s", ctx.Err(), out)
	}
	if !strings.Contains(string(out), "Authentication required") {
		t.Fatalf("positive control failed: logged-out agy -p did not prompt for authentication:\n%s", out)
	}
	if evs := readAsymmetryEventsIfExists(t, logPath); len(evs) != 0 {
		t.Fatalf("ASSUMPTION CHANGED: logged-out agy -p now fires SessionStart without OAuth (%+v); scripts/installshape can verify agy hooks directly", evs)
	}
}

// --- Row 22: transcript-model-field ---

func asymTranscriptModelFieldClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	out := cw.run(t, "", "", "", "Reply with the single word OK.", nil, "--output-format", "json")
	var doc struct {
		ModelUsage map[string]any `json:"modelUsage"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("parse claude --output-format json: %v\n%s", err, out)
	}
	if len(doc.ModelUsage) == 0 {
		t.Fatalf("Claude --output-format json reported empty modelUsage:\n%s", out)
	}

	tr := findClaudeTranscript(t, cw.home)
	var sawModel string
	for _, line := range strings.Split(tr, "\n") {
		var rec struct {
			Message struct {
				Model string `json:"model"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) == nil && rec.Message.Model != "" {
			sawModel = rec.Message.Model
			break
		}
	}
	if sawModel == "" {
		t.Fatalf("Claude transcript.jsonl carried no message.model field:\n%s", tr)
	}
}

func asymTranscriptModelFieldAgy(t *testing.T) {
	w := requireAgyWorld(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, w.bin, "-p", "Reply with the single word OK.", "--model", "gemini-3.8-flash-low", "--output-format", "json", "--print-timeout", "60s")
	cmd.Dir = w.work
	cmd.Env = w.env
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("agy -p --output-format json failed: %v\n%s", err, outBytes)
	}
	var doc map[string]any
	if err := json.Unmarshal(outBytes, &doc); err != nil {
		t.Fatalf("parse agy --output-format json: %v\n%s", err, outBytes)
	}
	convID, ok := doc["conversation_id"].(string)
	if !ok || convID == "" {
		t.Fatalf("positive control failed: agy JSON output missing string conversation_id:\n%s", outBytes)
	}
	if hitPath, hitVal, ok := findNonEmptyJSONKey(doc, "", "model", "modelUsage", "model_usage"); ok {
		t.Fatalf("ASSUMPTION CHANGED: agy -p --output-format json now reports %s=%v", hitPath, hitVal)
	}

	// Verify the production surface servedmodel.Observe actually reads: PLANNER_RESPONSE.model in transcript.jsonl.
	tr := w.transcript(t, convID)
	var sawPlannerModel string
	for _, line := range strings.Split(tr, "\n") {
		var rec struct {
			Source string `json:"source"`
			Type   string `json:"type"`
			Model  string `json:"model"`
		}
		if json.Unmarshal([]byte(line), &rec) == nil && rec.Source == "MODEL" && rec.Type == "PLANNER_RESPONSE" && rec.Model != "" {
			sawPlannerModel = rec.Model
			break
		}
	}
	if sawPlannerModel == "" {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity transcript.jsonl PLANNER_RESPONSE no longer records top-level model field (servedmodel.Observe will fall back to NotMeasured):\n%s", tr)
	}
}

func findNonEmptyJSONKey(v any, prefix string, targets ...string) (string, any, bool) {
	switch cur := v.(type) {
	case map[string]any:
		for k, val := range cur {
			path := k
			if prefix != "" {
				path = prefix + "." + k
			}
			for _, target := range targets {
				if k == target && val != nil && val != "" {
					return path, val, true
				}
			}
			if p, fv, ok := findNonEmptyJSONKey(val, path, targets...); ok {
				return p, fv, true
			}
		}
	case []any:
		for i, elem := range cur {
			path := fmt.Sprintf("%s[%d]", prefix, i)
			if p, fv, ok := findNonEmptyJSONKey(elem, path, targets...); ok {
				return p, fv, true
			}
		}
	}
	return "", nil, false
}

// --- Row 23: bash-oauth-token-stripping ---

func asymBashOAuthTokenStrippingClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	resultFile := filepath.Join(cw.work, "oauth_check.txt")
	// cw.run already passes CLAUDE_CODE_OAUTH_TOKEN in the parent claude -p process env.
	// Write the env check to a file on disk so model prose cannot spoof the verdict.
	cw.run(t, "", "", "",
		"Run the exact Bash command `if [ -n \"$CLAUDE_CODE_OAUTH_TOKEN\" ]; then echo OAUTH_PRESENT; else echo OAUTH_STRIPPED; fi > "+resultFile+"` and reply DONE.",
		nil)
	raw, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf("positive control failed: Bash did not write %s: %v", resultFile, err)
	}
	got := strings.TrimSpace(string(raw))
	if got != "OAUTH_STRIPPED" {
		t.Fatalf("ASSUMPTION CHANGED: Claude Code did not strip CLAUDE_CODE_OAUTH_TOKEN from Bash tool env (got %q, want OAUTH_STRIPPED; universe.sh relies on scrubbing)", got)
	}
}

func asymBashOAuthTokenStrippingAgy(t *testing.T) {
	w := requireAgyWorld(t)
	resultFile := filepath.Join(w.work, "oauth_check.txt")
	const sentinel = "sentinel-not-a-secret-9921"
	// Use a helper script writing only MATCH/MISMATCH to disk so the model neither refuses to
	// write an OAuth-named env var nor can spoof the verdict in prose.
	scriptPath := filepath.Join(w.work, "check_env.sh")
	script := "if [ \"$CLAUDE_CODE_OAUTH_TOKEN\" = \"" + sentinel + "\" ]; then echo MATCH > \"" + resultFile + "\"; else echo MISMATCH > \"" + resultFile + "\"; fi\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	w.run(t, "",
		"Run `sh "+scriptPath+"` via run_command, then reply DONE.",
		"CLAUDE_CODE_OAUTH_TOKEN="+sentinel)
	raw, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf("positive control failed: run_command did not write %s: %v", resultFile, err)
	}
	if strings.TrimSpace(string(raw)) != "MATCH" {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity now strips or alters CLAUDE_CODE_OAUTH_TOKEN in run_command env (got %q, want MATCH)", strings.TrimSpace(string(raw)))
	}
}

// --- Row 24: thinking-summaries-persistence ---

func asymThinkingSummariesPersistenceClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	// Use opus so Claude actually reasons and emits a thinking content block; haiku skips
	// thinking on short prompts, which would make an empty-summary assertion vacuous.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cw.bin, "-p",
		"Find the smallest integer n > 100 such that n, n+2, and n+6 are all prime. Check candidates carefully, then output only n.",
		"--model", "opus", "--dangerously-skip-permissions")
	cmd.Dir = cw.work
	cmd.Env = liveTestEnv(t, cw.home, "CLAUDE_PLUGIN_DATA="+cw.pluginData, cw.oauthEnv)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("claude -p --model opus failed: %v\n%s", err, out)
	}
	tr := findClaudeTranscript(t, cw.home)
	var sawThinkingBlock, sawSignature bool
	for _, line := range strings.Split(tr, "\n") {
		var rec struct {
			Message struct {
				Content []struct {
					Type      string `json:"type"`
					Thinking  string `json:"thinking"`
					Signature string `json:"signature"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		for _, c := range rec.Message.Content {
			if c.Type == "thinking" {
				sawThinkingBlock = true
				if c.Signature != "" {
					sawSignature = true
				}
				if strings.TrimSpace(c.Thinking) != "" {
					t.Fatalf("ASSUMPTION CHANGED: Claude Code now persists non-empty thinking summaries by default without showThinkingSummaries=true (doctor/capture.go warning may be obsolete): %q", c.Thinking)
				}
			}
		}
	}
	if !sawThinkingBlock || !sawSignature {
		t.Fatalf("positive control failed: Claude opus transcript carried no thinking block with signature (sawBlock=%v sawSig=%v)", sawThinkingBlock, sawSignature)
	}
}

func asymThinkingSummariesPersistenceAgy(t *testing.T) {
	w := requireAgyWorld(t)
	// Use a thinking-enabled tier (gemini-3.8-flash-medium) with default Antigravity settings
	// (no extra thinking-persistence flags): gemini-3.8-flash-low can skip reasoning on short
	// prompts and omit the thinking field entirely.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, w.bin, "-p",
		"Think carefully step by step: find the sum of all prime numbers strictly between 40 and 70, checking each odd number for divisibility, then reply with only the final integer.",
		"--model", "gemini-3.8-flash-medium", "--print-timeout", "90s")
	cmd.Dir = w.work
	cmd.Env = w.env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("agy -p failed: %v\n%s", err, out)
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	tr := w.transcript(t, convs[0])
	var sawThinking bool
	for _, line := range strings.Split(tr, "\n") {
		var rec struct {
			Source   string `json:"source"`
			Type     string `json:"type"`
			Thinking string `json:"thinking"`
		}
		if json.Unmarshal([]byte(line), &rec) == nil && rec.Source == "MODEL" && rec.Type == "PLANNER_RESPONSE" && strings.TrimSpace(rec.Thinking) != "" {
			sawThinking = true
			break
		}
	}
	if !sawThinking {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity no longer persists non-empty thinking field natively in transcript.jsonl PLANNER_RESPONSE (doctor/capture.go assumes native persistence):\n%s", tr)
	}
}

// --- Row 31: workflow-resume-from-run-id ---

func asymWorkflowResumeFromRunIDClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	countFile := filepath.Join(cw.work, "wf_count.txt")
	wfPath := filepath.Join(cw.work, "wf_resume_probe.js")
	// Native top-level Workflow script syntax (no `export default`), matching debate.js.
	// agent(prompt, opts) takes the prompt string as its first positional argument.
	wfSrc := `export const meta = {
  name: 'resume-probe',
  description: 'Pin Workflow resumeFromRunId replay with a cached agent() step.',
  phases: [{ title: 'Tick', detail: 'append one tick' }],
}

const res = await agent(
  "Run the exact Bash command: printf 'tick\\n' >> ` + countFile + ` && wc -l < ` + countFile + `. Reply ONLY with CLAUDE_WF_TICK_OK and the resulting line count.",
  { name: 'ticker', model: 'haiku' },
)

return { marker: 'CLAUDE_WF_OK', agentResult: res }
`
	if err := os.WriteFile(wfPath, []byte(wfSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	out := cw.run(t, "", "", "",
		"Do NOT call Bash, Read, Edit, or Write yourself — call ONLY the Workflow tool.\n"+
			"1) Call Workflow with scriptPath='"+wfPath+"' and args={}. Wait for it to finish and note its workflow run ID (the wf_... directory name in its Transcript dir).\n"+
			"2) Call Workflow a second time with the exact same scriptPath and args={}, passing resumeFromRunId set to that wf_... run ID. Wait for it to finish.\n"+
			"3) Call Workflow a third time with the exact same scriptPath and args={} and NO resumeFromRunId (control run). Wait for it to finish, then reply DONE.",
		nil, "--output-format", "stream-json", "--verbose")

	type wfCall struct {
		id              string
		resumeFromRunID string
	}
	type wfCompletion struct {
		status     string
		outputFile string
		summary    string
		body       string
	}
	var calls []wfCall
	var outerMutatingTools []string
	resultsByID := map[string]struct {
		isError bool
		content string
	}{}
	completionsByID := map[string]wfCompletion{}
	maxTokensByID := map[string]int{}
	maxToolsByID := map[string]int{}

	for _, line := range strings.Split(out, "\n") {
		var rec struct {
			Type        string `json:"type"`
			Subtype     string `json:"subtype"`
			ToolUseID   string `json:"tool_use_id"`
			Status      string `json:"status"`
			OutputFile  string `json:"output_file"`
			Summary     string `json:"summary"`
			ParentUseID string `json:"parent_tool_use_id"`
			Usage       struct {
				TotalTokens int `json:"total_tokens"`
				ToolUses    int `json:"tool_uses"`
			} `json:"usage"`
			Message struct {
				Content []struct {
					Type      string          `json:"type"`
					ID        string          `json:"id"`
					Name      string          `json:"name"`
					Input     json.RawMessage `json:"input"`
					ToolUseID string          `json:"tool_use_id"`
					IsError   bool            `json:"is_error"`
					Content   json.RawMessage `json:"content"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if rec.Type == "system" && rec.Subtype == "task_progress" && rec.ToolUseID != "" {
			if rec.Usage.TotalTokens > maxTokensByID[rec.ToolUseID] {
				maxTokensByID[rec.ToolUseID] = rec.Usage.TotalTokens
			}
			if rec.Usage.ToolUses > maxToolsByID[rec.ToolUseID] {
				maxToolsByID[rec.ToolUseID] = rec.Usage.ToolUses
			}
		}
		if rec.Type == "system" && rec.Subtype == "task_notification" && rec.ToolUseID != "" {
			comp := wfCompletion{
				status:     rec.Status,
				outputFile: rec.OutputFile,
				summary:    rec.Summary,
			}
			if rec.OutputFile != "" {
				if b, err := os.ReadFile(rec.OutputFile); err == nil {
					comp.body = string(b)
				}
			}
			completionsByID[rec.ToolUseID] = comp
		}
		if rec.ParentUseID != "" {
			continue
		}
		for _, c := range rec.Message.Content {
			if c.Type == "tool_use" {
				if c.Name == "Workflow" {
					var in struct {
						ResumeFromRunID string `json:"resumeFromRunId"`
					}
					_ = json.Unmarshal(c.Input, &in)
					calls = append(calls, wfCall{id: c.ID, resumeFromRunID: in.ResumeFromRunID})
				} else if c.Name == "Bash" || c.Name == "Edit" || c.Name == "Write" || c.Name == "Agent" || c.Name == "Task" || c.Name == "NotebookEdit" {
					outerMutatingTools = append(outerMutatingTools, c.Name)
				}
			}
			if c.Type == "tool_result" && c.ToolUseID != "" {
				resultsByID[c.ToolUseID] = struct {
					isError bool
					content string
				}{isError: c.IsError, content: string(c.Content)}
			}
		}
	}

	if len(outerMutatingTools) != 0 {
		t.Fatalf("outer Claude session called mutating tools %v instead of only Workflow; cannot trust tick counter:\n%s", outerMutatingTools, out)
	}
	type wfOutputDoc struct {
		Result struct {
			Marker      string `json:"marker"`
			AgentResult string `json:"agentResult"`
		} `json:"result"`
		WorkflowProgress []struct {
			Type   string `json:"type"`
			Cached bool   `json:"cached"`
		} `json:"workflowProgress"`
		TotalTokens    int `json:"totalTokens"`
		TotalToolCalls int `json:"totalToolCalls"`
	}
	parsedByID := map[string]wfOutputDoc{}
	var firstCalls, resumeCalls []wfCall
	for _, c := range calls {
		res, ok := resultsByID[c.id]
		if !ok || res.isError || strings.Contains(res.content, "tool_use_error") {
			t.Fatalf("Workflow call %+v failed launch (ok=%v err=%v content=%s)", c, ok, res.isError, res.content)
		}
		comp, cok := completionsByID[c.id]
		if !cok || comp.status != "completed" {
			t.Fatalf("Workflow call %+v did not complete cleanly (cok=%v comp=%+v)", c, cok, comp)
		}
		t.Logf("workflow call id=%s resumeFromRunId=%q tokens=%d tool_uses=%d summary=%q body=%q",
			c.id, c.resumeFromRunID, maxTokensByID[c.id], maxToolsByID[c.id], comp.summary, comp.body)
		var doc wfOutputDoc
		if err := json.Unmarshal([]byte(comp.body), &doc); err != nil {
			t.Fatalf("parse Workflow output_file JSON for %+v: err=%v body=%q", c, err, comp.body)
		}
		if doc.Result.Marker != "CLAUDE_WF_OK" || !strings.Contains(doc.Result.AgentResult, "CLAUDE_WF_TICK_OK") {
			t.Fatalf("Workflow call %+v returned unexpected result (marker=%q agentResult=%q)", c, doc.Result.Marker, doc.Result.AgentResult)
		}
		parsedByID[c.id] = doc
		if c.resumeFromRunID == "" {
			firstCalls = append(firstCalls, c)
		} else {
			resumeCalls = append(resumeCalls, c)
		}
	}
	if len(firstCalls) < 2 || len(resumeCalls) < 1 {
		t.Fatalf("expected 2 non-resume Workflow calls (initial + control) and 1 resumeFromRunId call; got non-resume=%d resume=%d calls=%+v",
			len(firstCalls), len(resumeCalls), calls)
	}
	// Verify the resumed call actually targeted call 1's workflow run ID (wf_...).
	reWFRunID := regexp.MustCompile(`wf_[A-Za-z0-9-]+`)
	call1RunID := reWFRunID.FindString(resultsByID[firstCalls[0].id].content)
	if call1RunID == "" {
		t.Fatalf("could not extract call 1 wf_... run ID from tool_result %q", resultsByID[firstCalls[0].id].content)
	}
	if resumeCalls[0].resumeFromRunID != call1RunID {
		t.Fatalf("resumed Workflow call passed resumeFromRunId=%q, want call 1 run ID %q", resumeCalls[0].resumeFromRunID, call1RunID)
	}
	call1Doc := parsedByID[firstCalls[0].id]
	resumeDoc := parsedByID[resumeCalls[0].id]
	call3Doc := parsedByID[firstCalls[1].id]
	// Replay proof 1: initial & control calls ran the agent (totalToolCalls >= 1), whereas the
	// resumed call replayed from cache (totalTokens == 0, totalToolCalls == 0, cached == true)
	// and returned the exact same agentResult as call 1.
	if call1Doc.TotalToolCalls == 0 || call3Doc.TotalToolCalls == 0 {
		t.Fatalf("positive control failed: uncached Workflow calls did not execute ticker agent tool calls (call1=%d call3=%d)",
			call1Doc.TotalToolCalls, call3Doc.TotalToolCalls)
	}
	var sawCachedAgent bool
	for _, wp := range resumeDoc.WorkflowProgress {
		if wp.Type == "workflow_agent" && wp.Cached {
			sawCachedAgent = true
			break
		}
	}
	if !sawCachedAgent || resumeDoc.TotalToolCalls != 0 || resumeDoc.TotalTokens != 0 {
		t.Fatalf("ASSUMPTION CHANGED: Workflow resumeFromRunId did not replay ticker agent from cache (cached=%v tokens=%d toolCalls=%d)",
			sawCachedAgent, resumeDoc.TotalTokens, resumeDoc.TotalToolCalls)
	}
	if resumeDoc.Result.AgentResult != call1Doc.Result.AgentResult {
		t.Fatalf("ASSUMPTION CHANGED: resumed Workflow returned different agentResult than call 1 (call1=%q resumed=%q)",
			call1Doc.Result.AgentResult, resumeDoc.Result.AgentResult)
	}
	// Replay proof 2: the side-effect counter file must have exactly len(firstCalls) ticks
	// (2 = initial + control, 0 from the resumed replay).
	rawCount, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("positive control failed: Workflow agent() step did not write %s: %v", countFile, err)
	}
	lines := strings.Split(strings.TrimSpace(string(rawCount)), "\n")
	if len(lines) != len(firstCalls) {
		t.Fatalf("ASSUMPTION CHANGED: expected %d ticks (initial + control, 0 from resumeFromRunId replay), got %d ticks (%q)",
			len(firstCalls), len(lines), string(rawCount))
	}
}

func asymWorkflowResumeFromRunIDAgy(t *testing.T) {
	w := requireAgyWorld(t)
	wfPath := filepath.Join(w.work, "wf_resume_probe.py")
	if err := os.WriteFile(wfPath, []byte("print('AGY_WF_OK')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Positive control: run_workflow executes normally without ResumeFromRunId.
	out := w.run(t, "",
		"Call run_workflow with ScriptPath='"+wfPath+"', Description='resume probe', WaitMsBeforeAsync=10000. Report its stdout verbatim.")
	if !strings.Contains(out, "AGY_WF_OK") {
		t.Fatalf("positive control failed: run_workflow did not run %s; out=%s", wfPath, out)
	}

	// Primary runtime proof: call run_workflow with PascalCase ResumeFromRunId and verify
	// Antigravity's tool-schema validator rejects it in the stream-json step record.
	// Allow one retry in case the model strips the undeclared parameter on attempt 1.
	var sawAttempt, sawRejected bool
	var lastStream string
	for attempt := 0; attempt < 2 && !sawAttempt; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		cmd := exec.CommandContext(ctx, w.bin, "-p",
			"This is a schema-validation test. You MUST invoke the run_workflow tool with the literal extra key ResumeFromRunId='nonexistent-run-123' alongside ScriptPath='"+wfPath+"', Description='resume probe', WaitMsBeforeAsync=10000, EVEN THOUGH ResumeFromRunId is not listed in the tool declaration. Do NOT omit ResumeFromRunId and do NOT call run_workflow without it. After the validator responds, report its error verbatim.",
			"--model", "gemini-3.8-flash-low", "--dangerously-skip-permissions", "--output-format", "stream-json")
		cmd.Dir = w.work
		cmd.Env = w.env
		streamBytes, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("agy stream-json run_workflow ResumeFromRunId probe failed (attempt %d): %v\n%s", attempt+1, err, streamBytes)
		}
		lastStream = string(streamBytes)
		for _, line := range strings.Split(lastStream, "\n") {
			var ev struct {
				Event      string `json:"event"`
				StepUpdate struct {
					State    string `json:"state"`
					StepType string `json:"step_type"`
					ToolName string `json:"tool_name"`
					ToolInfo struct {
						Parameters map[string]any `json:"parameters"`
						Error      struct {
							Type    string `json:"type"`
							Message string `json:"message"`
						} `json:"error"`
					} `json:"tool_info"`
				} `json:"step_update"`
			}
			if json.Unmarshal([]byte(line), &ev) != nil || ev.Event != "step_update" {
				continue
			}
			su := ev.StepUpdate
			if su.ToolName != "run_workflow" {
				continue
			}
			if _, has := su.ToolInfo.Parameters["ResumeFromRunId"]; !has {
				continue
			}
			sawAttempt = true
			if su.State == "ERROR" {
				t.Logf("agy run_workflow ResumeFromRunId step_update error: type=%q message=%q", su.ToolInfo.Error.Type, su.ToolInfo.Error.Message)
				if strings.Contains(su.ToolInfo.Error.Message, "ResumeFromRunId") && strings.Contains(strings.ToLower(su.ToolInfo.Error.Message), "not allowed") {
					sawRejected = true
				}
			}
		}
	}
	if !sawAttempt {
		t.Fatalf("positive control failed: agy did not attempt run_workflow with ResumeFromRunId in stream-json:\n%s", lastStream)
	}
	if !sawRejected {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity run_workflow no longer rejects ResumeFromRunId as a disallowed additional property; scripts/universe.sh:892 can revisit agy resume:\n%s", lastStream)
	}

	// Secondary hermetic doc check (case-insensitive) from the isolated home's unpacked builtin skill.
	skillPath := filepath.Join(w.home, ".gemini", "antigravity-cli", "builtin", "skills", "run-workflow", "SKILL.md")
	skillBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("positive control failed: isolated home missing builtin run-workflow skill %s: %v", skillPath, err)
	}
	skillText := string(skillBytes)
	if !strings.Contains(skillText, "ScriptPath") || !strings.Contains(skillText, "Args") {
		t.Fatalf("positive control failed: builtin run-workflow skill missing expected ScriptPath/Args docs:\n%s", skillText)
	}
	lowerSkill := strings.ToLower(skillText)
	if strings.Contains(lowerSkill, "resumefromrunid") || strings.Contains(lowerSkill, "resume_from_run_id") {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity builtin run-workflow SKILL.md now documents resumeFromRunId; scripts/universe.sh:892 can enable agy resume")
	}
}

// --- Row 35: agy-oauth-refresh-threshold ---

func asymAgyOAuthRefreshThresholdClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	// Claude Code accepts a lent CLAUDE_CODE_OAUTH_TOKEN env var with zero credentials file on
	// disk in fakeHome (no file-based expiry threshold check by the child CLI).
	credPath := filepath.Join(cw.home, ".claude", ".credentials.json")
	if _, err := os.Stat(credPath); err == nil {
		t.Fatalf("unexpected .credentials.json in isolated Claude home %s", credPath)
	}
	out := cw.run(t, "", "", "", "Reply with the exact token CLAUDE_TOKEN_ALIVE_91 and nothing else.", nil)
	if !strings.Contains(out, "CLAUDE_TOKEN_ALIVE_91") {
		t.Fatalf("Claude failed to run with lent CLAUDE_CODE_OAUTH_TOKEN and no credentials file; out=%s", out)
	}
}

func asymAgyOAuthRefreshThresholdAgy(t *testing.T) {
	bin := requireAgyBin(t)
	// Stage a valid host refresh-stripped access token first via testhome.StageAgyAccessToken,
	// then rewrite ONLY its expiry field in two isolated throwaway homes to pin agy's 5m0s
	// threshold (shared/testhome.AgyRefreshThreshold):
	//   - 315s (>5m): refresh-stripped token succeeds (rc=0).
	//   - 240s (<5m): refresh-stripped token cannot run and prompts for interactive OAuth login.
	baseHome := testhome.NewLiveHome(t)
	if err := testhome.StageAgyAccessToken(baseHome); err != nil {
		t.Fatalf("HOOKCORE_LIVE paired asymmetry test requires Antigravity credentials: %v", err)
	}
	tokRel := filepath.Join(".gemini", "antigravity-cli", "antigravity-oauth-token")
	baseBytes, err := os.ReadFile(filepath.Join(baseHome, tokRel))
	if err != nil {
		t.Fatalf("read staged agy token: %v", err)
	}

	makeTimedHome := func(left time.Duration) string {
		h := testhome.NewLiveHome(t)
		dstDir := filepath.Join(h, ".gemini", "antigravity-cli")
		if err := os.MkdirAll(dstDir, 0o700); err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(baseBytes, &doc); err != nil {
			t.Fatal(err)
		}
		tokObj := doc["token"].(map[string]any)
		tokObj["expiry"] = time.Now().Add(left).UTC().Format(time.RFC3339Nano)
		b, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dstDir, "antigravity-oauth-token"), b, 0o600); err != nil {
			t.Fatal(err)
		}
		for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
			if data, rerr := os.ReadFile(filepath.Join(baseHome, ".gemini", "antigravity-cli", extra)); rerr == nil {
				_ = os.WriteFile(filepath.Join(dstDir, extra), data, 0o600)
			}
		}
		return h
	}

	// Positive control: 315s (>5m AgyRefreshThreshold) succeeds without a refresh_token.
	okHome := makeTimedHome(315 * time.Second)
	okCtx, okCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer okCancel()
	okCmd := exec.CommandContext(okCtx, bin, "-p", "Reply with ok.", "--model", "gemini-3.8-flash-low", "--output-format", "text", "--print-timeout", "20s")
	okCmd.Dir = t.TempDir()
	okCmd.Env = liveTestEnv(t, okHome)
	if out, err := okCmd.CombinedOutput(); err != nil {
		t.Fatalf("ASSUMPTION CHANGED: refresh-stripped agy token with 315s (>5m) left failed (err=%v):\n%s", err, out)
	}

	// Negative assertion: 240s (<5m AgyRefreshThreshold) triggers OAuth refresh and prompts
	// "Authentication required" because refresh_token is stripped.
	shortHome := makeTimedHome(240 * time.Second)
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shortCancel()
	shortCmd := exec.CommandContext(shortCtx, bin, "-p", "Reply with ok.", "--model", "gemini-3.8-flash-low", "--output-format", "text", "--print-timeout", "8s")
	shortCmd.Dir = t.TempDir()
	shortCmd.Stdin = nil
	shortCmd.Env = liveTestEnv(t, shortHome)
	shortOut, shortErr := shortCmd.CombinedOutput()
	if shortErr == nil || !strings.Contains(string(shortOut), "Authentication required") {
		t.Fatalf("ASSUMPTION CHANGED: refresh-stripped agy token with 240s (<5m) left no longer prompts for OAuth login (shared/testhome.AgyRefreshThreshold may have changed; err=%v):\n%s", shortErr, shortOut)
	}
}
