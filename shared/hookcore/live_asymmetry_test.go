package hookcore_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

type asymmetryEvent struct {
	Event                 string            `json:"event"`
	HookEventName         string            `json:"hook_event_name"`
	Engine                string            `json:"engine"`
	SessionID             string            `json:"session_id"`
	AgentID               string            `json:"agent_id"`
	AgentType             string            `json:"agent_type"`
	ToolName              string            `json:"tool_name"`
	RawToolName           string            `json:"raw_tool_name"`
	Command               string            `json:"command"`
	Source                string            `json:"source"`
	Error                 string            `json:"error"`
	HasResultField        bool              `json:"has_result_field"`
	InvocationNum         *int              `json:"invocation_num,omitempty"`
	EnvNames              []string          `json:"env_names"`
	EnvValues             map[string]string `json:"env_values,omitempty"`
	ClaudeSessionFiles    []string          `json:"claude_session_files"`
	ClaudeSessionPayloads map[string]string `json:"claude_session_payloads"`
	CWD                   string            `json:"cwd"`
	Raw                   string            `json:"raw"`
}

func liveAsymmetryEnabled() bool {
	return os.Getenv("HOOKCORE_LIVE") == "1"
}

func requireLiveAsymmetry(t *testing.T) {
	t.Helper()
	if !liveAsymmetryEnabled() {
		t.Skip("skipping live asymmetry test without HOOKCORE_LIVE=1")
	}
}

func requireClaudeBin(t *testing.T) string {
	t.Helper()
	bin := findClaudeBin()
	if bin == "" {
		t.Fatal("claude CLI not found on PATH; HOOKCORE_LIVE paired asymmetry tests require both claude and agy")
	}
	return bin
}

func requireAgyBin(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("agy")
	if err != nil {
		t.Fatalf("agy binary not found on PATH; HOOKCORE_LIVE paired asymmetry tests require both claude and agy: %v", err)
	}
	return bin
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

func readAsymmetryEventsIfExists(t *testing.T, path string) []asymmetryEvent {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return readAsymmetryEvents(t, path)
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

// claudeWorld is an isolated Claude Code HOME authenticated ONLY via a lent CLAUDE_CODE_OAUTH_TOKEN
// access token (never copying .credentials.json or its refresh token, #1890).
type claudeWorld struct {
	bin, home, work, pluginData, oauthEnv string
}

func newClaudeWorld(t *testing.T) *claudeWorld {
	t.Helper()
	bin := requireClaudeBin(t)
	oauthEnv, err := testhome.ClaudeOAuthEnv()
	if err != nil {
		t.Fatalf("HOOKCORE_LIVE paired asymmetry test requires Claude credentials: %v", err)
	}
	return &claudeWorld{
		bin:        bin,
		home:       testhome.NewLiveHome(t),
		work:       t.TempDir(),
		pluginData: t.TempDir(),
		oauthEnv:   oauthEnv,
	}
}

func (w *claudeWorld) run(t *testing.T, pluginDir, logPath, resumeSession, prompt string, extraEnv []string, extraArgs ...string) string {
	t.Helper()
	args := []string{"-p", prompt, "--model", "haiku", "--dangerously-skip-permissions"}
	if pluginDir != "" {
		args = append(args, "--plugin-dir", pluginDir)
	}
	if resumeSession != "" {
		args = append(args, "--resume", resumeSession)
	}
	args = append(args, extraArgs...)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, w.bin, args...)
	cmd.Dir = w.work
	// Append oauthEnv AFTER liveTestEnv strips CLAUDE_* so only the short-lived access token reaches claude -p.
	env := liveTestEnv(t, w.home, "CLAUDE_PLUGIN_DATA="+w.pluginData, w.oauthEnv)
	if logPath != "" {
		env = append(env, "ASYMMETRY_LOG="+logPath)
	}
	env = append(env, extraEnv...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("claude -p failed: %v\nOutput:\n%s", err, out)
	}
	return string(out)
}

func findClaudeTranscript(t *testing.T, fakeHome string) string {
	t.Helper()
	projectsDir := filepath.Join(fakeHome, ".claude", "projects")
	var found string
	_ = filepath.Walk(projectsDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(p, ".jsonl") {
			found = p
		}
		return nil
	})
	if found == "" {
		t.Fatalf("no Claude session transcript found under %s", projectsDir)
	}
	b, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("read Claude transcript %s: %v", found, err)
	}
	return string(b)
}

func findClaudeTranscriptByID(t *testing.T, fakeHome, sessionID string) string {
	t.Helper()
	projectsDir := filepath.Join(fakeHome, ".claude", "projects")
	wantName := sessionID + ".jsonl"
	var found string
	_ = filepath.Walk(projectsDir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && info.Name() == wantName {
			found = p
		}
		return nil
	})
	if found == "" {
		t.Fatalf("Claude session transcript %s not found under %s", wantName, projectsDir)
	}
	b, err := os.ReadFile(found)
	if err != nil {
		t.Fatalf("read Claude transcript %s: %v", found, err)
	}
	return string(b)
}

// assertClaudeDebugHookJSONAccepted verifies in Claude's --debug-file that the named hook event's
// JSON stdout containing marker was parsed and validated cleanly (proving acceptance rather than
// merely echoing raw stdout or avoiding a crash). The ±3-line adjacency window relies on each test
// plugin registering only a single JSON-emitting hook event.
func assertClaudeDebugHookJSONAccepted(t *testing.T, debugPath, eventName, marker string) {
	t.Helper()
	raw, err := os.ReadFile(debugPath)
	if err != nil {
		t.Fatalf("read Claude debug file %s: %v", debugPath, err)
	}
	lines := strings.Split(string(raw), "\n")
	for i, line := range lines {
		if !strings.Contains(line, "Successfully parsed and validated hook JSON output") {
			continue
		}
		// Claude emits the parse-success line immediately adjacent to the hook-success line
		// carrying the validated JSON payload for that event.
		for j := max(0, i-3); j < min(len(lines), i+4); j++ {
			if strings.Contains(lines[j], "Hook "+eventName) && strings.Contains(lines[j], marker) {
				return
			}
		}
	}
	t.Fatalf("Claude debug log %s did not record clean JSON parse + validation for %s carrying %q:\n%s",
		debugPath, eventName, marker, string(raw))
}

// requireAgyWorld builds an isolated Antigravity home authenticated ONLY via a refresh-stripped
// access token (#1890), failing closed (never skipping) when agy or credentials are missing.
// Never copies host ~/.config/gh credentials into the isolated home.
func requireAgyWorld(t *testing.T) *agyWorld {
	t.Helper()
	bin := requireAgyBin(t)
	w := &agyWorld{bin: bin, home: testhome.NewLiveHome(t), work: t.TempDir()}
	if err := testhome.StageAgyAccessToken(w.home); err != nil {
		t.Fatalf("HOOKCORE_LIVE paired asymmetry test requires Antigravity credentials: %v", err)
	}
	settings := filepath.Join(w.home, ".gemini", "antigravity-cli", "settings.json")
	if err := os.WriteFile(settings, []byte(`{"toolPermission":"always-proceed"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	w.env = liveTestEnv(t, w.home)
	return w
}

func readAgyCLILogs(fakeHome string) string {
	var b strings.Builder
	logDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli")
	_ = filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if (strings.HasPrefix(info.Name(), "cli") && strings.HasSuffix(info.Name(), ".log")) || info.Name() == "cli.log" {
			if data, err := os.ReadFile(path); err == nil {
				b.Write(data)
				b.WriteByte('\n')
			}
		}
		return nil
	})
	return b.String()
}

type claudeSessionRecord struct {
	PID       int    `json:"pid"`
	SessionID string `json:"sessionId"`
	CWD       string `json:"cwd"`
}

func parseMatchingClaudeSessionRecord(events []asymmetryEvent) (claudeSessionRecord, bool) {
	for _, ev := range events {
		if ev.Event != "SessionStart" || ev.SessionID == "" {
			continue
		}
		for name, raw := range ev.ClaudeSessionPayloads {
			var rec claudeSessionRecord
			if json.Unmarshal([]byte(raw), &rec) == nil && rec.PID > 0 && rec.SessionID == ev.SessionID && rec.CWD != "" {
				if name == fmt.Sprintf("%d.json", rec.PID) {
					return rec, true
				}
			}
		}
	}
	return claudeSessionRecord{}, false
}

// --- Row 1: session-start-on-resume ---

func asymSessionStartOnResumeClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-resume.jsonl")

	cw.run(t, pluginDir, logPath, "", "Reply with the single word OK.", nil)
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
	cw.run(t, pluginDir, logPath, sessionID, "Reply with the single word OK2.", nil)
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
}

func asymSessionStartOnResumeAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
  "PreInvocation":[{"command":"./bin/asymmetryprobe -event PreInvocation"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-resume.jsonl")

	addr1File := filepath.Join(w.work, "addr1.txt")
	addr2File := filepath.Join(w.work, "addr2.txt")
	w.run(t, "", "Run `printf '%s' \"$ANTIGRAVITY_LS_ADDRESS\" > "+addr1File+"` via run_command, then reply OK.", "ASYMMETRY_LOG="+logPath)
	turn1 := readAsymmetryEvents(t, logPath)
	var sawTurn1Start bool
	for _, ev := range turn1 {
		if ev.Event == "SessionStart" {
			sawTurn1Start = true
			break
		}
	}
	if !sawTurn1Start {
		t.Fatalf("positive control failed: Antigravity did not fire SessionStart on turn 1; events=%+v", turn1)
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	conv := convs[0]

	if err := os.Truncate(logPath, 0); err != nil {
		t.Fatal(err)
	}
	w.run(t, conv, "Run `printf '%s' \"$ANTIGRAVITY_LS_ADDRESS\" > "+addr2File+"` via run_command, then reply OK2.", "ASYMMETRY_LOG="+logPath)
	turn2 := readAsymmetryEvents(t, logPath)
	var sawPreInv, sawPreInvZero bool
	for _, ev := range turn2 {
		if ev.Event == "SessionStart" {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity now fires SessionStart on --conversation resume (%+v); PreInvocation resume workarounds can be retired", ev)
		}
		if ev.Event == "PreInvocation" {
			sawPreInv = true
			if ev.InvocationNum != nil && *ev.InvocationNum == 0 {
				sawPreInvZero = true
			}
		}
	}
	if !sawPreInv || !sawPreInvZero {
		t.Fatalf("Antigravity resume did not fire PreInvocation with invocationNum==0 (sawPreInv=%v sawZero=%v); events=%+v", sawPreInv, sawPreInvZero, turn2)
	}
	// Confirm turn 2 ran under a fresh language server (covering LS-restart resume as well as --conversation).
	a1, err1 := os.ReadFile(addr1File)
	a2, err2 := os.ReadFile(addr2File)
	if err1 != nil || err2 != nil || strings.TrimSpace(string(a1)) == "" || strings.TrimSpace(string(a2)) == "" {
		t.Fatalf("positive control failed: could not read turn1/turn2 LS addresses: err1=%v err2=%v a1=%q a2=%q", err1, err2, string(a1), string(a2))
	}
	if strings.TrimSpace(string(a1)) == strings.TrimSpace(string(a2)) {
		t.Fatalf("positive control failed: turn 1 and turn 2 reused the same LS address %q; did not exercise fresh language server", string(a1))
	}
}

// --- Row 2: hook-plugin-root-env ---

func asymHookPluginRootEnvClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-rootenv.jsonl")

	cw.run(t, pluginDir, logPath, "", "Reply with OK.", nil)
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("Claude SessionStart hook did not fire")
	}
	ev := events[0]
	if !slices.Contains(ev.EnvNames, "CLAUDE_PLUGIN_ROOT") {
		t.Fatalf("Claude SessionStart hook missing CLAUDE_PLUGIN_ROOT; env=%v", ev.EnvNames)
	}
	if slices.Contains(ev.EnvNames, "PLUGIN_ROOT") {
		t.Fatalf("ASSUMPTION CHANGED: Claude SessionStart hook now sets PLUGIN_ROOT; env=%v", ev.EnvNames)
	}
}

func asymHookPluginRootEnvAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
  "PreInvocation":[{"command":"./bin/asymmetryprobe -event PreInvocation"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-rootenv.jsonl")

	w.run(t, "", "Reply with OK.", "ASYMMETRY_LOG="+logPath)
	events := readAsymmetryEvents(t, logPath)
	seen := map[string]bool{}
	for _, ev := range events {
		if ev.Event != "SessionStart" && ev.Event != "PreInvocation" {
			continue
		}
		seen[ev.Event] = true
		for _, want := range []string{"PLUGIN_ROOT", "PLUGIN_DATA", "ANTIGRAVITY_CONVERSATION_ID"} {
			if !slices.Contains(ev.EnvNames, want) {
				t.Fatalf("Antigravity %s hook missing %s; env=%v", ev.Event, want, ev.EnvNames)
			}
		}
		if slices.Contains(ev.EnvNames, "CLAUDE_PLUGIN_ROOT") {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity %s hook now sets CLAUDE_PLUGIN_ROOT; env=%v", ev.Event, ev.EnvNames)
		}
	}
	if !seen["SessionStart"] || !seen["PreInvocation"] {
		t.Fatalf("Antigravity did not fire both SessionStart and PreInvocation (seen=%v); events=%+v", seen, events)
	}
}

// --- Row 3: hook-wake-credentials ---

func asymHookWakeCredentialsClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-wake.jsonl")

	cw.run(t, pluginDir, logPath, "", "Reply with the single word OK.", nil)
	events := readAsymmetryEvents(t, logPath)
	if _, ok := parseMatchingClaudeSessionRecord(events); !ok {
		t.Fatalf("Claude SessionStart hook did not observe valid ~/.claude/sessions/<pid>.json matching session_id; events=%+v", events)
	}
}

func asymHookWakeCredentialsAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
  "PreInvocation":[{"command":"./bin/asymmetryprobe -event PreInvocation"}],
  "PostToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUse"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-wake.jsonl")

	out := w.run(t, "", "Run the exact shell command `env | grep -o '^ANTIGRAVITY_[A-Z_]*' | sort` via run_command and report its output verbatim.", "ASYMMETRY_LOG="+logPath)
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	tr := w.transcript(t, convs[0])
	// Verify all three ANTIGRAVITY_* env vars appeared in a non-USER_INPUT transcript step
	// (run_command output / model response), never matching the echoed prompt.
	var nonUserBlob strings.Builder
	for _, line := range strings.Split(tr, "\n") {
		var step struct {
			Type    string `json:"type"`
			Content string `json:"content"`
			Output  string `json:"output"`
		}
		if json.Unmarshal([]byte(line), &step) == nil && step.Type != "USER_INPUT" {
			nonUserBlob.WriteString(step.Content)
			nonUserBlob.WriteByte('\n')
			nonUserBlob.WriteString(step.Output)
			nonUserBlob.WriteByte('\n')
		}
	}
	stepOut := nonUserBlob.String()
	for _, wantShellVar := range []string{"ANTIGRAVITY_LS_ADDRESS", "ANTIGRAVITY_CSRF_TOKEN", "ANTIGRAVITY_CONVERSATION_ID"} {
		if !strings.Contains(stepOut, wantShellVar) {
			t.Fatalf("positive control failed: agent run_command transcript step did not contain %s; out=%s", wantShellVar, out)
		}
	}

	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("expected at least one hook event")
	}
	var sawConvIDInHook bool
	for _, ev := range events {
		if slices.Contains(ev.EnvNames, "ANTIGRAVITY_CONVERSATION_ID") {
			sawConvIDInHook = true
		}
		for _, forbidden := range []string{"ANTIGRAVITY_LS_ADDRESS", "ANTIGRAVITY_CSRF_TOKEN"} {
			if slices.Contains(ev.EnvNames, forbidden) {
				t.Fatalf("ASSUMPTION CHANGED: Antigravity %s hook now receives %s (env=%v); telepathy link can move into the hook", ev.Event, forbidden, ev.EnvNames)
			}
		}
	}
	if !sawConvIDInHook {
		t.Fatalf("positive control failed: no Antigravity hook received ANTIGRAVITY_CONVERSATION_ID; events=%+v", events)
	}
}

// --- Row 7: failed-shell-command-hook ---

func asymFailedShellCommandHookClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}],
    "PostToolUseFailure": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUseFailure"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-fail.jsonl")

	cw.run(t, pluginDir, logPath, "",
		"Run the exact Bash command `echo ASYM_FAIL_PROBE && exit 2`. After it finishes, reply with DONE.", nil)
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
}

func asymFailedShellCommandHookAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PostToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUse"}]}],
  "PostToolUseFailure":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUseFailure"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-fail.jsonl")

	w.run(t, "", "Run the exact shell command `echo ASYM_FAIL_PROBE && exit 2` via run_command. After it finishes, reply with DONE.", "ASYMMETRY_LOG="+logPath)
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	tr := w.transcript(t, convs[0])
	var sawProbeOutput, sawExit2 bool
	for _, line := range strings.Split(tr, "\n") {
		var step struct {
			Type     string `json:"type"`
			ExitCode *int   `json:"exit_code"`
			Content  string `json:"content"`
			Output   string `json:"output"`
			Error    string `json:"error"`
		}
		if json.Unmarshal([]byte(line), &step) != nil || step.Type == "USER_INPUT" {
			continue
		}
		blob := step.Content + "\n" + step.Output + "\n" + step.Error
		if strings.Contains(blob, "ASYM_FAIL_PROBE") {
			sawProbeOutput = true
		}
		if (step.ExitCode != nil && *step.ExitCode == 2) ||
			(step.Type == "GENERIC" && strings.Contains(step.Content, "Command exited with code 2.")) ||
			strings.Contains(blob, "exit code 2") ||
			strings.Contains(blob, "Exit code: 2") ||
			strings.Contains(blob, "exit status 2") {
			sawExit2 = true
		}
	}
	if !sawProbeOutput || !sawExit2 {
		t.Fatalf("positive control failed: transcript steps do not structurally confirm ASYM_FAIL_PROBE exited 2 (probe=%v exit2=%v):\n%s",
			sawProbeOutput, sawExit2, tr)
	}
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
}

// --- Row 11: workflow-child-attribution ---

func asymWorkflowChildAttributionClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SubagentStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SubagentStart"}]}],
    "SubagentStop": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SubagentStop"}]}],
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}]
  }
}`
	agentDef := "---\nname: asym-child\ndescription: Asymmetry child worker.\nmodel: haiku\ntools: Bash\n---\nRun the exact Bash command requested and return its output.\n"
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, map[string]string{"asym-child.md": agentDef})
	logPath := filepath.Join(t.TempDir(), "claude-child.jsonl")

	cw.run(t, pluginDir, logPath, "",
		"Spawn the subagent asymmetryprobe:asym-child via the Agent tool to run `echo ASYM_WF_CHILD` and return its output. Do NOT run Bash yourself.", nil)
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
}

func asymWorkflowChildAttributionAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
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
	var parentConvID string
	for _, ev := range events {
		if ev.Event == "SessionStart" && ev.SessionID != "" {
			parentConvID = ev.SessionID
			break
		}
	}
	if parentConvID == "" {
		t.Fatalf("positive control failed: missing parent SessionStart conversationId; events=%+v", events)
	}
	var sawChildCmd bool
	for _, ev := range events {
		if ev.Event == "SubagentStart" || ev.Event == "SubagentStop" {
			t.Errorf("ASSUMPTION CHANGED: Antigravity now fires %s for run_workflow child (%+v); .agy-sittings workaround can be retired", ev.Event, ev)
		}
		if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "ASYM_WF_CHILD") {
			if ev.SessionID == "" || ev.SessionID == parentConvID {
				t.Fatalf("positive control failed: PostToolUse for ASYM_WF_CHILD ran in parent conversation %q rather than child conversation (got %q)", parentConvID, ev.SessionID)
			}
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
}

// --- Row 12: finish-tool-overwrite ---

func asymFinishToolOverwriteClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "PreToolUse": [{"matcher":"StructuredOutput","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PreToolUse"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-finish.jsonl")
	schema := `{"type":"object","properties":{"answer":{"type":"string"},"stamped_by_hook":{"type":"string"}},"required":["answer"]}`

	out := cw.run(t, pluginDir, logPath, "",
		"Return structured output with answer='ORIGINAL_ANSWER'.",
		[]string{"ASYMMETRY_OVERWRITE_FINISH=1"},
		"--output-format", "json", "--json-schema", schema)
	if !strings.Contains(out, "OVERWRITTEN_BY_PRETOOLUSE") {
		t.Fatalf("Claude did not honor PreToolUse updatedInput on StructuredOutput; out=%s", out)
	}
}

func asymFinishToolOverwriteAgy(t *testing.T) {
	w := requireAgyWorld(t)
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
	if !strings.Contains(out, "ORIGINAL_CHILD_ANSWER") {
		t.Fatalf("positive control failed: workflow did not return ORIGINAL_CHILD_ANSWER (out=%s)", out)
	}
	events := readAsymmetryEvents(t, logPath)
	var sawFinishPre bool
	for _, ev := range events {
		if ev.Event == "PreToolUse" && ev.RawToolName == "finish" {
			sawFinishPre = true
			break
		}
	}
	if !sawFinishPre {
		t.Fatalf("positive control failed: PreToolUse hook never observed RawToolName=finish; events=%+v", events)
	}
	if strings.Contains(out, "OVERWRITTEN_BY_PRETOOLUSE") {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity now honors PreToolUse overwrite on finish (out=%s); debate.py can switch to hook-based envelope stamping", out)
	}
}

// --- Row 10: system-message-metadata-panic ---

func asymSystemMessageMetadataPanicClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-meta.jsonl")
	debugPath := filepath.Join(t.TempDir(), "claude-meta-debug.log")

	out := cw.run(t, pluginDir, logPath, "", "Reply with the single word OK.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=SessionStart",
			`ASYMMETRY_EMIT_RAW={"systemMessage":"META_PROBE_CLAUDE_OK","metadata":{"user_markdown":"META_PROBE_CLAUDE_OK"}}`,
		},
		"--output-format", "stream-json", "--verbose", "--debug-file", debugPath)
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 || events[0].Event != "SessionStart" {
		t.Fatalf("Claude SessionStart hook did not fire; events=%+v out=%s", events, out)
	}
	assertClaudeDebugHookJSONAccepted(t, debugPath, "SessionStart", "META_PROBE_CLAUDE_OK")
}

func asymSystemMessageMetadataPanicAgy(t *testing.T) {
	// 1. Delete-the-cause positive control: systemMessage WITHOUT metadata is delivered intact.
	wCtrl := requireAgyWorld(t)
	buildAsymmetryProbe(t, filepath.Join(wCtrl.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe"))
	wCtrl.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}]}}`)
	ctrlLog := filepath.Join(t.TempDir(), "agy-meta-ctrl.jsonl")
	wCtrl.run(t, "", "Reply with the single word OK.",
		"ASYMMETRY_LOG="+ctrlLog,
		"ASYMMETRY_EMIT_EVENT=SessionStart",
		`ASYMMETRY_EMIT_RAW={"injectSteps":[{"systemMessage":{"systemMessage":"META_CONTROL_DELIVERED"}}]}`)
	ctrlConvs := wCtrl.conversations(t)
	if len(ctrlConvs) != 1 {
		t.Fatalf("control: want 1 conversation, got %v", ctrlConvs)
	}
	if !strings.Contains(wCtrl.transcript(t, ctrlConvs[0]), "META_CONTROL_DELIVERED") {
		t.Fatalf("positive control failed: systemMessage without metadata was not delivered into transcript.jsonl")
	}

	// 2. Bug assertion: adding metadata.user_markdown panics hookcaller and drops the message.
	w := requireAgyWorld(t)
	buildAsymmetryProbe(t, filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe"))
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-meta.jsonl")

	w.run(t, "", "Reply with the single word OK.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=SessionStart",
		`ASYMMETRY_EMIT_RAW={"injectSteps":[{"systemMessage":{"systemMessage":"META_PROBE_AGY_DROPPED","metadata":{"user_markdown":"META_PROBE_AGY_DROPPED"}}}]}`)

	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 || events[0].Event != "SessionStart" {
		t.Fatalf("positive control failed: SessionStart hook did not fire; events=%+v", events)
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	tr := w.transcript(t, convs[0])
	logs := readAgyCLILogs(w.home)
	if strings.Contains(tr, "META_PROBE_AGY_DROPPED") || !strings.Contains(logs, "type mismatch: cannot convert string to message") {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity no longer panics on systemMessage with metadata (#1774 fixed); in_transcript=%v logs_snippet=%q",
			strings.Contains(tr, "META_PROBE_AGY_DROPPED"), logs)
	}
}

// --- Row 13: post-tool-use-feedback ---

func asymPostToolUseFeedbackClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-postfb.jsonl")

	out := cw.run(t, pluginDir, logPath, "",
		"Run the exact Bash command `echo REAL_TOOL_OUT_1` and then report BOTH its stdout and any additional context marker returned by the tool hook verbatim.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=PostToolUse",
			`ASYMMETRY_EMIT_RAW={"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"POST_FEEDBACK_TOKEN_77"}}`,
		})
	if !strings.Contains(out, "POST_FEEDBACK_TOKEN_77") {
		t.Fatalf("Claude did not deliver PostToolUse additionalContext to agent; out=%s", out)
	}
}

func asymPostToolUseFeedbackAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PostToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUse"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-postfb.jsonl")

	out := w.run(t, "",
		"Run the exact shell command `echo REAL_TOOL_OUT_1` via run_command and repeat everything you observed from the tool result verbatim.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=PostToolUse",
		`ASYMMETRY_EMIT_RAW={"overwriteResult":"POST_FEEDBACK_TOKEN_77"}`)

	events := readAsymmetryEvents(t, logPath)
	var sawPost bool
	for _, ev := range events {
		if ev.Event == "PostToolUse" && strings.Contains(ev.Command, "REAL_TOOL_OUT_1") {
			sawPost = true
		}
	}
	if !sawPost {
		t.Fatalf("positive control failed: PostToolUse never fired for REAL_TOOL_OUT_1; events=%+v", events)
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	tr := w.transcript(t, convs[0])
	var sawRealOutInStep bool
	for _, line := range strings.Split(tr, "\n") {
		var step struct {
			Type    string `json:"type"`
			Content string `json:"content"`
			Output  string `json:"output"`
		}
		if json.Unmarshal([]byte(line), &step) == nil && step.Type != "USER_INPUT" {
			if strings.Contains(step.Content+"\n"+step.Output, "REAL_TOOL_OUT_1") {
				sawRealOutInStep = true
				break
			}
		}
	}
	if !sawRealOutInStep {
		t.Fatalf("positive control failed: REAL_TOOL_OUT_1 missing from non-USER_INPUT transcript steps")
	}
	if strings.Contains(out, "POST_FEEDBACK_TOKEN_77") || strings.Contains(tr, "POST_FEEDBACK_TOKEN_77") {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity PostToolUse overwriteResult now reaches the agent; pending_*.jsonl PreInvocation bridge can be retired (out=%s)", out)
	}
}

// --- Row 14: unknown-pretool-hook-stdout ---

func asymUnknownPreToolHookStdoutClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "PreToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PreToolUse"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)

	// Case 1: non-schema JSON stdout from PreToolUse — Claude ignores it and runs Bash.
	// Case 2: plain non-JSON stdout from PreToolUse (#1929 / pretooluse/main.go:58) — Claude also ignores it and runs Bash.
	for _, tc := range []struct {
		name string
		emit string
	}{
		{"non_schema_json", `{"not_claude_schema":"IGNORED_BY_CLAUDE"}`},
		{"plain_non_json", `NOT_JSON_PRETOOL_STDOUT_LINE`},
	} {
		logPath := filepath.Join(t.TempDir(), "claude-prestdout-"+tc.name+".jsonl")
		markerFile := filepath.Join(cw.work, "ran_"+tc.name+".txt")
		cw.run(t, pluginDir, logPath, "",
			"Run the exact Bash command `echo ok > "+markerFile+"` and then reply DONE.",
			[]string{
				"ASYMMETRY_EMIT_EVENT=PreToolUse",
				"ASYMMETRY_EMIT_RAW=" + tc.emit,
			})
		events := readAsymmetryEvents(t, logPath)
		var sawPre bool
		for _, ev := range events {
			if ev.Event == "PreToolUse" && strings.Contains(ev.Command, markerFile) {
				sawPre = true
				break
			}
		}
		if !sawPre {
			t.Fatalf("positive control failed (%s): Claude PreToolUse hook never fired for %s; events=%+v", tc.name, markerFile, events)
		}
		if _, err := os.Stat(markerFile); err != nil {
			t.Fatalf("Claude blocked or failed Bash command after PreToolUse emitted %s (marker %s missing): %v", tc.name, markerFile, err)
		}
	}
}

func asymUnknownPreToolHookStdoutAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PreToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PreToolUse"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-prestdout.jsonl")
	markerFile := filepath.Join(w.work, "ran_ok.txt")

	w.run(t, "",
		"Run the exact shell command `echo PRETOOL_RAN_OK_33 > "+markerFile+"` via run_command. If it errors, do not retry; reply DONE.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=PreToolUse",
		`ASYMMETRY_EMIT_RAW={"systemMessage":"NON_SCHEMA_PRETOOL_MSG"}`)

	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("positive control failed: PreToolUse hook never fired")
	}
	if _, err := os.Stat(markerFile); err == nil {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity executed run_command despite PreToolUse emitting unknown systemMessage field (#1927)")
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	fullTr, err := os.ReadFile(filepath.Join(w.home, ".gemini", "antigravity-cli", "brain", convs[0], ".system_generated", "logs", "transcript_full.jsonl"))
	if err != nil {
		t.Fatalf("read transcript_full.jsonl: %v", err)
	}
	if !strings.Contains(string(fullTr), "failed to unmarshal result from hook") {
		t.Fatalf("ASSUMPTION CHANGED: transcript_full.jsonl lacks 'failed to unmarshal result from hook'; full=%s", string(fullTr))
	}
}

// --- Row 15: strict-protojson-posttool-hook ---

func asymStrictProtojsonPostToolHookClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-poststdout.jsonl")
	debugPath := filepath.Join(t.TempDir(), "claude-poststdout-debug.log")
	markerFile := filepath.Join(cw.work, "posttool_ran_ok.txt")

	cw.run(t, pluginDir, logPath, "",
		"Run the exact Bash command `echo ok > "+markerFile+"` and reply DONE.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=PostToolUse",
			`ASYMMETRY_EMIT_RAW={"systemMessage":"CLAUDE_POSTTOOL_OK","not_claude_schema":"IGNORED_BY_CLAUDE"}`,
		}, "--debug-file", debugPath)
	events := readAsymmetryEvents(t, logPath)
	var sawPost bool
	for _, ev := range events {
		if ev.Event == "PostToolUse" && strings.Contains(ev.Command, markerFile) {
			sawPost = true
			break
		}
	}
	if !sawPost {
		t.Fatalf("positive control failed: Claude PostToolUse hook never fired for %s; events=%+v", markerFile, events)
	}
	if _, err := os.Stat(markerFile); err != nil {
		t.Fatalf("Claude Bash command did not execute marker file %s: %v", markerFile, err)
	}
	assertClaudeDebugHookJSONAccepted(t, debugPath, "PostToolUse", "CLAUDE_POSTTOOL_OK")
}

func asymStrictProtojsonPostToolHookAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PostToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PostToolUse"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-poststdout.jsonl")

	w.run(t, "",
		"Run the exact shell command `echo POSTTOOL_RAN_OK_44` via run_command. After it finishes, reply DONE.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=PostToolUse",
		`ASYMMETRY_EMIT_RAW={"systemMessage":"NON_SCHEMA_POSTTOOL_MSG"}`)

	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("positive control failed: PostToolUse hook never fired")
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	fullTr, _ := os.ReadFile(filepath.Join(w.home, ".gemini", "antigravity-cli", "brain", convs[0], ".system_generated", "logs", "transcript_full.jsonl"))
	logs := readAgyCLILogs(w.home)
	if !strings.Contains(string(fullTr), "failed to unmarshal result from hook") && !strings.Contains(logs, "failed to unmarshal result from hook") {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity PostToolUse no longer rejects unknown systemMessage field via strict protojson;\nfull=%s\nlogs=%s", string(fullTr), logs)
	}
}

// --- Row 16: session-end-hook-event ---

func asymSessionEndHookEventClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}],
    "SessionEnd": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionEnd"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-end.jsonl")

	cw.run(t, pluginDir, logPath, "", "Reply with the single word OK.", nil)
	events := readAsymmetryEvents(t, logPath)
	var sawEnd bool
	var endReason string
	for _, ev := range events {
		if ev.Event == "SessionEnd" {
			sawEnd = true
			var raw struct {
				Reason string `json:"reason"`
			}
			_ = json.Unmarshal([]byte(ev.Raw), &raw)
			endReason = raw.Reason
		}
	}
	if !sawEnd {
		t.Fatalf("Claude did not fire SessionEnd at session exit; events=%+v", events)
	}
	if endReason != "other" {
		t.Fatalf("ASSUMPTION CHANGED: headless claude -p SessionEnd reason = %q, want %q (checkpointseal/main.go:21)", endReason, "other")
	}
}

func asymSessionEndHookEventAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
  "SessionEnd":[{"command":"./bin/asymmetryprobe -event SessionEnd"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-end.jsonl")

	w.run(t, "", "Reply with the single word OK.", "ASYMMETRY_LOG="+logPath)
	events := readAsymmetryEvents(t, logPath)
	var sawStart bool
	for _, ev := range events {
		if ev.Event == "SessionStart" {
			sawStart = true
		}
		if ev.Event == "SessionEnd" {
			t.Fatalf("ASSUMPTION CHANGED: Antigravity now fires SessionEnd (%+v); plugins/*/hooks.json and validatejson can enable SessionEnd", ev)
		}
	}
	if !sawStart {
		t.Fatalf("positive control failed: SessionStart did not fire when SessionEnd was declared alongside it; events=%+v", events)
	}
}

// --- Row 25: post-tool-delivery-cap ---

func findClaudeAdditionalContextAttachments(t *testing.T, tr string) []string {
	t.Helper()
	var out []string
	for _, line := range strings.Split(tr, "\n") {
		var rec struct {
			Attachment struct {
				Type    string   `json:"type"`
				Content []string `json:"content"`
			} `json:"attachment"`
		}
		if json.Unmarshal([]byte(line), &rec) == nil && rec.Attachment.Type == "hook_additional_context" {
			out = append(out, rec.Attachment.Content...)
		}
	}
	return out
}

func findClaudeSpillFiles(fakeHome string) []string {
	var out []string
	_ = filepath.Walk(filepath.Join(fakeHome, ".claude", "projects"), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.Contains(filepath.ToSlash(p), "/tool-results/") && strings.HasSuffix(p, "-additionalContext.txt") {
			out = append(out, p)
		}
		return nil
	})
	return out
}

func asymPostToolDeliveryCapClaude(t *testing.T) {
	hooksJSON := `{
  "hooks": {
    "PostToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PostToolUse"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)

	// Positive control at sittingwrite.DeliveryBudget (9,000 bytes, <10,000 cap): hook fires,
	// transcript hook_additional_context carries tail marker inline without <persisted-output>,
	// and zero spill file is created.
	cwCtrl := newClaudeWorld(t)
	ctrlLog := filepath.Join(t.TempDir(), "claude-cap-ctrl.jsonl")
	cwCtrl.run(t, pluginDir, ctrlLog, "",
		"Run `echo CAP_CTRL` via Bash, then reply DONE.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=PostToolUse",
			"ASYMMETRY_EMIT_BIG_LEN=9000",
			"ASYMMETRY_EMIT_BIG_MARKER=CTRL_MARKER_4422",
		})
	ctrlEvents := readAsymmetryEvents(t, ctrlLog)
	if len(ctrlEvents) == 0 {
		t.Fatal("positive control failed: PostToolUse hook never fired on 9,000-byte run")
	}
	ctrlAtts := findClaudeAdditionalContextAttachments(t, findClaudeTranscript(t, cwCtrl.home))
	ctrlJoined := strings.Join(ctrlAtts, "\n")
	if !strings.Contains(ctrlJoined, "CTRL_MARKER_4422") || strings.Contains(ctrlJoined, "<persisted-output>") {
		t.Fatalf("positive control failed: Claude did not deliver 9,000-byte PostToolUse additionalContext inline (DeliveryBudget=9000); attachments=%q", ctrlJoined)
	}
	if spills := findClaudeSpillFiles(cwCtrl.home); len(spills) != 0 {
		t.Fatalf("positive control failed: 9,000-byte payload unexpectedly spilled to %v", spills)
	}

	// Negative assertion: 12,000-byte payload (>10,000 cap) spills to a tool-results file and
	// replaces the inline attachment with <persisted-output> (delivery.go:16).
	cwBig := newClaudeWorld(t)
	bigLog := filepath.Join(t.TempDir(), "claude-cap-big.jsonl")
	cwBig.run(t, pluginDir, bigLog, "",
		"Run `echo CAP_BIG` via Bash, then reply DONE.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=PostToolUse",
			"ASYMMETRY_EMIT_BIG_LEN=12000",
			"ASYMMETRY_EMIT_BIG_MARKER=BIG_TAIL_MARKER_9911",
		})
	bigEvents := readAsymmetryEvents(t, bigLog)
	if len(bigEvents) == 0 {
		t.Fatal("positive control failed: PostToolUse hook never fired on 12,000-byte run")
	}
	bigAtts := findClaudeAdditionalContextAttachments(t, findClaudeTranscript(t, cwBig.home))
	bigJoined := strings.Join(bigAtts, "\n")
	if strings.Contains(bigJoined, "BIG_TAIL_MARKER_9911") || !strings.Contains(bigJoined, "<persisted-output>") {
		t.Fatalf("ASSUMPTION CHANGED: Claude delivered a 12,000-byte PostToolUse additionalContext inline without <persisted-output> spill; attachments=%q", bigJoined)
	}
	spills := findClaudeSpillFiles(cwBig.home)
	if len(spills) == 0 {
		t.Fatalf("ASSUMPTION CHANGED: Claude spilled 12,000-byte PostToolUse additionalContext without writing a tool-results/*-additionalContext.txt file; attachments=%q", bigJoined)
	}
	spillBody, err := os.ReadFile(spills[0])
	if err != nil || !strings.Contains(string(spillBody), "BIG_TAIL_MARKER_9911") {
		t.Fatalf("positive control failed: spill file %s did not contain BIG_TAIL_MARKER_9911 (err=%v)", spills[0], err)
	}
}

func asymPostToolDeliveryCapAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PreInvocation":[{"command":"./bin/asymmetryprobe -event PreInvocation"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-cap.jsonl")

	w.run(t, "", "Reply with OK.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=PreInvocation",
		"ASYMMETRY_EMIT_BIG_LEN=12000",
		"ASYMMETRY_EMIT_BIG_MARKER=BIG_TAIL_MARKER_9911")

	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	// Read transcript_full.jsonl (the full non-elided transcript, whereas compact transcript.jsonl
	// elides the middle of long steps) and verify head, middle, and tail markers plus >=11,800
	// contiguous padding bytes survive inline on a SINGLE step record (cannot be satisfied by
	// JSON wrapper overhead).
	fullPath := filepath.Join(w.home, ".gemini", "antigravity-cli", "brain", convs[0], ".system_generated", "logs", "transcript_full.jsonl")
	fullBytes, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read %s: %v", fullPath, err)
	}
	wantPadA := strings.Repeat("a", 5900)
	wantPadB := strings.Repeat("b", 5900)
	var sawIntactStep bool
	for _, line := range strings.Split(string(fullBytes), "\n") {
		if strings.Contains(line, "BIG_HEAD_9911") &&
			strings.Contains(line, wantPadA) &&
			strings.Contains(line, "BIG_MID_9911") &&
			strings.Contains(line, wantPadB) &&
			strings.Contains(line, "BIG_TAIL_MARKER_9911") {
			sawIntactStep = true
			break
		}
	}
	if !sawIntactStep {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity truncated/elided a 12,000-byte PreInvocation injectSteps systemMessage (no single transcript_full.jsonl step carried HEAD + 5900*'a' + MID + 5900*'b' + TAIL)")
	}
}

// --- Row 26: strict-protojson-postinvocation-hook ---

func asymStrictProtojsonPostInvocationHookClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SubagentStop": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SubagentStop"}]}]
  }
}`
	agentDef := "---\nname: postinv-child\ndescription: Subagent for SubagentStop stdout tolerance check.\nmodel: haiku\n---\nReply CHILD_OK immediately without calling any tools.\n"
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, map[string]string{"postinv-child.md": agentDef})
	logPath := filepath.Join(t.TempDir(), "claude-substop.jsonl")
	debugPath := filepath.Join(t.TempDir(), "claude-substop-debug.log")

	cw.run(t, pluginDir, logPath, "",
		"Spawn the subagent asymmetryprobe:postinv-child via the Agent tool and reply DONE.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=SubagentStop",
			`ASYMMETRY_EMIT_RAW={"systemMessage":"CLAUDE_SUBSTOP_OK"}`,
		}, "--debug-file", debugPath)
	events := readAsymmetryEvents(t, logPath)
	var stopCount int
	for _, ev := range events {
		if ev.Event == "SubagentStop" {
			stopCount++
		}
	}
	// Note: emitting JSON {"systemMessage":...} is parsed cleanly without re-invoking the subagent
	// (stopCount==1), unlike emitting additionalContext (feov-subagentstop/main.go:4-6) which
	// checkpointseal/main.go:586-587 avoids emitting on SubagentStop.
	if stopCount != 1 {
		t.Fatalf("expected Claude SubagentStop to fire exactly once without re-invoking subagent; got stopCount=%d events=%+v", stopCount, events)
	}
	assertClaudeDebugHookJSONAccepted(t, debugPath, "SubagentStop", "CLAUDE_SUBSTOP_OK")
}

func asymStrictProtojsonPostInvocationHookAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PostInvocation":[{"command":"./bin/asymmetryprobe -event PostInvocation"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-postinv.jsonl")

	w.run(t, "", "Reply with OK.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=PostInvocation",
		`ASYMMETRY_EMIT_RAW={"systemMessage":"NON_SCHEMA_POSTINV_MSG"}`)

	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("positive control failed: PostInvocation hook never fired")
	}
	logs := readAgyCLILogs(w.home)
	var sawLine bool
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, "post-invocation hook") && strings.Contains(line, "failed to unmarshal result from hook") {
			sawLine = true
			break
		}
	}
	if !sawLine {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity PostInvocation no longer logs a single post-invocation hook unmarshal failure on systemMessage stdout (fetch-bin.sh:88-97);\nlogs=%s", logs)
	}
}

// --- Row 27: strict-protojson-stop-hook ---

func asymStrictProtojsonStopHookClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "Stop": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event Stop"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-stop.jsonl")
	debugPath := filepath.Join(t.TempDir(), "claude-stop-debug.log")

	cw.run(t, pluginDir, logPath, "", "Reply with OK.",
		[]string{
			"ASYMMETRY_EMIT_EVENT=Stop",
			`ASYMMETRY_EMIT_RAW={"systemMessage":"CLAUDE_STOP_OK"}`,
		}, "--debug-file", debugPath)
	events := readAsymmetryEvents(t, logPath)
	var saw bool
	for _, ev := range events {
		if ev.Event == "Stop" {
			saw = true
			break
		}
	}
	if !saw {
		t.Fatalf("positive control failed: Claude Stop hook never fired; events=%+v", events)
	}
	assertClaudeDebugHookJSONAccepted(t, debugPath, "Stop", "CLAUDE_STOP_OK")
}

func asymStrictProtojsonStopHookAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "Stop":[{"command":"./bin/asymmetryprobe -event Stop"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-stop.jsonl")

	w.run(t, "", "Reply with OK.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_EMIT_EVENT=Stop",
		`ASYMMETRY_EMIT_RAW={"systemMessage":"NON_SCHEMA_STOP_MSG"}`)

	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("positive control failed: Stop hook never fired")
	}
	logs := readAgyCLILogs(w.home)
	var sawLine bool
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, "stop hook") && strings.Contains(line, "failed to unmarshal result from hook") {
			sawLine = true
			break
		}
	}
	if !sawLine {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity Stop no longer logs a single stop hook unmarshal failure on systemMessage stdout (fetch-bin.sh:88-97);\nlogs=%s", logs)
	}
}

// --- Row 28: headless-workspace-paths ---

func asymHeadlessWorkspacePathsClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-ws.jsonl")

	cw.run(t, pluginDir, logPath, "", "Reply with OK.", nil)
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("Claude SessionStart hook did not fire")
	}
	var raw struct {
		CWD string `json:"cwd"`
	}
	if err := json.Unmarshal([]byte(events[0].Raw), &raw); err != nil {
		t.Fatalf("decode Claude SessionStart raw JSON: %v", err)
	}
	if raw.CWD != cw.work {
		t.Fatalf("Claude SessionStart stdin cwd = %q, want %q", raw.CWD, cw.work)
	}
}

func asymHeadlessWorkspacePathsAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "SessionStart":[{"command":"./bin/asymmetryprobe -event SessionStart"}],
  "PreInvocation":[{"command":"./bin/asymmetryprobe -event PreInvocation"}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-ws.jsonl")

	w.run(t, "", "Reply with OK.", "ASYMMETRY_LOG="+logPath)
	events := readAsymmetryEvents(t, logPath)
	seen := map[string]bool{}
	for _, ev := range events {
		if ev.Event != "SessionStart" && ev.Event != "PreInvocation" {
			continue
		}
		seen[ev.Event] = true
		var raw struct {
			WorkspacePaths []string `json:"workspacePaths"`
		}
		if err := json.Unmarshal([]byte(ev.Raw), &raw); err != nil {
			t.Fatalf("decode Antigravity %s raw JSON: %v", ev.Event, err)
		}
		if len(raw.WorkspacePaths) == 0 || !slices.Contains(raw.WorkspacePaths, w.work) {
			t.Fatalf("ASSUMPTION CHANGED: headless agy -p %s workspacePaths = %v, expected to contain %q", ev.Event, raw.WorkspacePaths, w.work)
		}
	}
	if !seen["SessionStart"] || !seen["PreInvocation"] {
		t.Fatalf("Antigravity did not fire both SessionStart and PreInvocation (seen=%v); events=%+v", seen, events)
	}
}

// --- Row 29: pretool-deny-transcript-step ---

func asymPreToolDenyTranscriptStepClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "PreToolUse": [{"matcher":"Bash","hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event PreToolUse"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-deny.jsonl")
	markerFile := filepath.Join(cw.work, "should_not_exist.txt")

	cw.run(t, pluginDir, logPath, "",
		"Run the exact Bash command `echo nope > "+markerFile+"`. If it is blocked, do not retry; reply DONE.",
		[]string{"ASYMMETRY_DENY_REASON=ASYM_DENY_SENTINEL_77"})
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("positive control failed: Claude PreToolUse hook never fired")
	}
	if _, err := os.Stat(markerFile); err == nil {
		t.Fatalf("PreToolUse deny did not block Bash command: %s was created", markerFile)
	}
	tr := findClaudeTranscript(t, cw.home)
	var sawErrorResult bool
	for _, line := range strings.Split(tr, "\n") {
		var rec struct {
			Message struct {
				Content []struct {
					Type    string `json:"type"`
					IsError bool   `json:"is_error"`
					Content string `json:"content"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		for _, c := range rec.Message.Content {
			if c.Type == "tool_result" && c.IsError && strings.Contains(c.Content, "ASYM_DENY_SENTINEL_77") {
				sawErrorResult = true
				break
			}
		}
	}
	if !sawErrorResult {
		t.Fatalf("Claude transcript did not record PreToolUse deny as tool_result(is_error=true) with ASYM_DENY_SENTINEL_77:\n%s", tr)
	}
}

func asymPreToolDenyTranscriptStepAgy(t *testing.T) {
	w := requireAgyWorld(t)
	probeBin := filepath.Join(w.home, ".gemini", "config", "plugins", "asymmetryprobe", "bin", "asymmetryprobe")
	buildAsymmetryProbe(t, probeBin)
	w.installPlugin(t, "asymmetryprobe", `{"asymmetryprobe-hooks":{
  "PreToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/asymmetryprobe -event PreToolUse"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "agy-deny.jsonl")
	markerFile := filepath.Join(w.work, "should_not_exist.txt")

	w.run(t, "",
		"Run the exact shell command `echo nope > "+markerFile+"` via run_command. If it is blocked, do not retry; reply DONE.",
		"ASYMMETRY_LOG="+logPath,
		"ASYMMETRY_DENY_REASON=ASYM_DENY_SENTINEL_77")
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("positive control failed: Antigravity PreToolUse hook never fired")
	}
	if _, err := os.Stat(markerFile); err == nil {
		t.Fatalf("PreToolUse deny did not block run_command: %s was created", markerFile)
	}
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want 1 conversation, got %v", convs)
	}
	tr := w.transcript(t, convs[0])
	var sawGenericErr bool
	for _, line := range strings.Split(tr, "\n") {
		var step struct {
			Type   string `json:"type"`
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if json.Unmarshal([]byte(line), &step) == nil &&
			step.Type == "GENERIC" && step.Status == "ERROR" &&
			strings.Contains(step.Error, "ASYM_DENY_SENTINEL_77") {
			sawGenericErr = true
			break
		}
	}
	if !sawGenericErr {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity no longer records PreToolUse deny as GENERIC/ERROR step with deny reason (live_system_install_test.go:381-391):\n%s", tr)
	}
}

// --- Row 30: plugin-root-inheritance-precedence ---

func asymPluginRootInheritancePrecedenceClaude(t *testing.T) {
	cw := newClaudeWorld(t)
	hooksJSON := `{
  "hooks": {
    "SessionStart": [{"hooks":[{"type":"command","command":"exec \"${CLAUDE_PLUGIN_ROOT}/bin/asymmetryprobe\" -event SessionStart"}]}]
  }
}`
	pluginDir := claudeAsymmetryPlugin(t, hooksJSON, nil)
	logPath := filepath.Join(t.TempDir(), "claude-pr-inherit.jsonl")
	const foreignRoot = "/foreign/antigravity/plugin"

	cw.run(t, pluginDir, logPath, "", "Reply with OK.", []string{"PLUGIN_ROOT=" + foreignRoot})
	events := readAsymmetryEvents(t, logPath)
	if len(events) == 0 {
		t.Fatal("Claude SessionStart hook did not fire")
	}
	ev := events[0]
	if ev.EnvValues["CLAUDE_PLUGIN_ROOT"] != pluginDir {
		t.Fatalf("positive control failed: CLAUDE_PLUGIN_ROOT = %q, want %q", ev.EnvValues["CLAUDE_PLUGIN_ROOT"], pluginDir)
	}
	if ev.EnvValues["PLUGIN_ROOT"] != foreignRoot {
		t.Fatalf("ASSUMPTION CHANGED: Claude now strips inherited PLUGIN_ROOT from hook env (got %q, want %q); pluginroot.Resolve precedence rationale can be revisited", ev.EnvValues["PLUGIN_ROOT"], foreignRoot)
	}
}

func asymPluginRootInheritancePrecedenceAgy(t *testing.T) {
	// Case 1 (SHOULD 4): install a real plugin with a skill, invoke that skill in agy with
	// PLUGIN_ROOT unset in the parent process, and verify run_command inside that active skill
	// receives an empty PLUGIN_ROOT (pinning research/SKILL.md:36's fallback).
	wUnset := requireAgyWorld(t)
	unsetFile := filepath.Join(wUnset.work, "shell_plugin_root_unset.txt")
	pDir := filepath.Join(wUnset.home, ".gemini", "config", "plugins", "prcheck")
	skillDir := filepath.Join(pDir, "skills", "check-root")
	for _, d := range []string{filepath.Join(pDir, ".antigravity-plugin"), skillDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(pDir, ".antigravity-plugin", "plugin.json"),
		[]byte(`{"name":"prcheck","version":"1.0.0","description":"Plugin root check"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: check-root\ndescription: Check PLUGIN_ROOT inside run_command.\n---\n" +
		"Run this exact shell command via run_command and then reply SKILL_RAN_OK:\n" +
		"`printf 'SKILL_OK:%s' \"${PLUGIN_ROOT:-}\" > " + unsetFile + "`\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	// Positive control: prove agy discovered prcheck as an enabled plugin registering check-root.
	listOut, listErr := runValidator(t, wUnset.home, wUnset.bin, "plugin", "list", "--json")
	if listErr != nil {
		t.Fatalf("positive control failed: agy plugin list --json failed: %v\n%s", listErr, listOut)
	}
	var plugins []struct {
		Name    string   `json:"name"`
		Enabled bool     `json:"enabled"`
		Skills  []string `json:"skills"`
	}
	if err := json.Unmarshal([]byte(listOut), &plugins); err != nil {
		t.Fatalf("parse agy plugin list --json: %v\n%s", err, listOut)
	}
	var sawPrcheck bool
	for _, p := range plugins {
		if p.Name == "prcheck" && p.Enabled && slices.Contains(p.Skills, "check-root") {
			sawPrcheck = true
			break
		}
	}
	if !sawPrcheck {
		t.Fatalf("positive control failed: agy plugin list --json did not show enabled plugin prcheck with skill check-root:\n%s", listOut)
	}
	wUnset.run(t, "",
		"Read and follow the `check-root` skill instructions now, then reply DONE.")
	gotUnset, err := os.ReadFile(unsetFile)
	if err != nil {
		t.Fatalf("positive control failed (unset plugin-skill case): skill did not write %s: %v", unsetFile, err)
	}
	if string(gotUnset) != "SKILL_OK:" {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity now populates PLUGIN_ROOT in run_command inside a plugin skill (got %q, want %q)", string(gotUnset), "SKILL_OK:")
	}

	// Case 2: when PLUGIN_ROOT is pre-set in the parent process, agy's run_command shell inherits
	// it verbatim (which reaches nested Claude sessions, requiring CLAUDE_PLUGIN_ROOT precedence).
	w := requireAgyWorld(t)
	outFile := filepath.Join(w.work, "shell_plugin_root.txt")
	const foreignRoot = "/foreign/antigravity/plugin"

	w.run(t, "",
		"Run the exact shell command `printf '%s' \"$PLUGIN_ROOT\" > "+outFile+"` via run_command, then reply DONE.",
		"PLUGIN_ROOT="+foreignRoot)
	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("positive control failed (inherited case): run_command did not write %s: %v", outFile, err)
	}
	if string(got) != foreignRoot {
		t.Fatalf("ASSUMPTION CHANGED: Antigravity run_command shell no longer inherits PLUGIN_ROOT (got %q, want %q)", string(got), foreignRoot)
	}
}
