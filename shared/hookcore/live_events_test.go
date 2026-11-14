package hookcore_test

import (
	"bufio"
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

// agyWorld is an isolated Antigravity home with one plugin installed and a workspace to run in.
type agyWorld struct {
	bin, home, work string
	env             []string
}

// newAgyWorld returns an isolated agy home holding the caller's credentials, or skips. Tool calls
// proceed without prompting — the setting a headless run needs, scoped to this throwaway home.
func newAgyWorld(t *testing.T) *agyWorld {
	t.Helper()
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live Antigravity test without HOOKCORE_LIVE=1")
	}
	if testing.Short() {
		t.Skip("skipping live Antigravity test under -short")
	}
	bin, err := exec.LookPath("agy")
	if err != nil {
		t.Skipf("agy binary not found on PATH: %v", err)
	}
	w := &agyWorld{bin: bin, home: t.TempDir(), work: t.TempDir()}
	copyAgyCredentials(t, w.home)
	settings := filepath.Join(w.home, ".gemini", "antigravity-cli", "settings.json")
	if err := os.WriteFile(settings, []byte(`{"toolPermission":"always-proceed"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	w.env = liveTestEnv(w.home)
	return w
}

// run sends one prompt, in a new conversation or continuing conv, and returns the combined output.
func (w *agyWorld) run(t *testing.T, conv, prompt string, extraEnv ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
	defer cancel()
	args := []string{"-p", prompt, "--model", "gemini-3.8-flash-low", "--print-timeout", "220s"}
	if conv != "" {
		args = append(args, "--conversation", conv)
	}
	cmd := exec.CommandContext(ctx, w.bin, args...)
	cmd.Dir = w.work
	cmd.Env = append(append([]string(nil), w.env...), extraEnv...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("agy -p failed: %v\n%s", err, out)
	}
	return string(out)
}

// conversations lists the conversation ids this world's agy has created.
func (w *agyWorld) conversations(t *testing.T) []string {
	t.Helper()
	entries, _ := os.ReadDir(filepath.Join(w.home, ".gemini", "antigravity-cli", "brain"))
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids
}

// transcript returns a conversation's transcript text.
func (w *agyWorld) transcript(t *testing.T, conv string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(w.home, ".gemini", "antigravity-cli", "brain", conv, ".system_generated", "logs", "transcript.jsonl"))
	if err != nil {
		t.Fatalf("transcript of %s: %v", conv, err)
	}
	return string(b)
}

// installPlugin writes a plugin with the given hooks.json into the world's global plugin directory.
func (w *agyWorld) installPlugin(t *testing.T, name, hooksJSON string) string {
	t.Helper()
	dir := filepath.Join(w.home, ".gemini", "config", "plugins", name)
	if err := os.MkdirAll(filepath.Join(dir, ".antigravity-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"` + name + `","version":"1.0.0","description":"live test plugin"}`
	if err := os.WriteFile(filepath.Join(dir, ".antigravity-plugin", "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hooks.json"), []byte(hooksJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

type eventLine struct {
	Event         string   `json:"event"`
	HookEventName string   `json:"hook_event_name"`
	Engine        string   `json:"engine"`
	SessionID     string   `json:"session_id"`
	CWD           string   `json:"cwd"`
	ProjectDir    string   `json:"project_dir"`
	PluginRoot    string   `json:"plugin_root"`
	PluginDataDir string   `json:"plugin_data_dir"`
	ToolCallID    string   `json:"tool_call_id"`
	InvocationNum *int     `json:"invocation_num"`
	EnvNames      []string `json:"env_names"`
}

func readEvents(t *testing.T, path string) []eventLine {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("no hook ran: %v", err)
	}
	defer f.Close()
	var out []eventLine
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var l eventLine
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			t.Fatalf("event log line %q: %v", sc.Text(), err)
		}
		out = append(out, l)
	}
	return out
}

// TestLiveAntigravityEventContract locks the agy behaviour gray-area's telepathy link rests on. Each
// assertion is a fact a design decision depends on; when agy changes one, this fails and names it.
//
//   - SessionStart fires once for a new conversation and NOT when a conversation resumes on a new
//     language server — so a stale link can only be noticed from PreInvocation.
//   - PreInvocation fires before every model step, invocationNum counting from 0 and restarting at 0
//     on each turn — so "the turn's first step" is invocationNum 0.
//   - An injected ephemeralMessage is acted on — so a hook can ASK for a command.
//   - A hook process receives no language-server address or token, while the agent's shell gets
//     both — so only a command the agent runs can record a link.
func TestLiveAntigravityEventContract(t *testing.T) {
	w := newAgyWorld(t)
	probe := filepath.Join(w.home, ".gemini", "config", "plugins", "eventlog", "bin", "eventlog")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "go", "build", "-o", probe, "./testdata/eventlog").CombinedOutput(); err != nil {
		t.Fatalf("build eventlog: %v\n%s", err, out)
	}
	w.installPlugin(t, "eventlog", `{"eventlog-hooks":{
  "SessionStart":[{"command":"./bin/eventlog -event SessionStart"}],
  "PreInvocation":[{"command":"./bin/eventlog -event PreInvocation"}],
  "PreToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/eventlog -event PreToolUse"}]}],
  "PostToolUse":[{"matcher":"run_command","hooks":[{"command":"./bin/eventlog -event PostToolUse"}]}]}}`)
	logPath := filepath.Join(t.TempDir(), "events.jsonl")
	logEnv := "EVENTLOG_FILE=" + logPath

	// Turn 1: a new conversation that runs one command of its own.
	out := w.run(t, "", "Run this shell command once and report its output verbatim: env | grep -o '^ANTIGRAVITY_[A-Z_]*' | sort. Then stop.", logEnv)
	convs := w.conversations(t)
	if len(convs) != 1 {
		t.Fatalf("want one conversation, have %v", convs)
	}
	conv := convs[0]
	first := readEvents(t, logPath)

	var starts, preTools, postTools int
	var invs []int
	for _, l := range first {
		if l.SessionID != conv || l.Engine != "antigravity" {
			t.Errorf("hook call for session %q engine %q, want %s antigravity", l.SessionID, l.Engine, conv)
		}
		if l.HookEventName == "" {
			t.Errorf("%s payload carried no hookEventName (agy 1.5.4+)", l.Event)
		}
		if l.CWD == "" || l.ProjectDir == "" {
			t.Errorf("%s payload carried empty cwd=%q or projectDir=%q (agy 1.5.4+)", l.Event, l.CWD, l.ProjectDir)
		}
		for _, wantEnv := range []string{"PLUGIN_ROOT", "PLUGIN_DATA"} {
			if !slices.Contains(l.EnvNames, wantEnv) {
				t.Errorf("%s hook did not receive %s in env=%v (agy 1.5.3+)", l.Event, wantEnv, l.EnvNames)
			}
		}
		switch l.Event {
		case "SessionStart":
			starts++
		case "PreInvocation":
			if l.InvocationNum == nil {
				t.Errorf("a PreInvocation payload carried no invocationNum")
				continue
			}
			invs = append(invs, *l.InvocationNum)
		case "PreToolUse":
			preTools++
			if l.ToolCallID == "" {
				t.Errorf("PreToolUse payload carried empty toolCall.id (agy 1.5.4+)")
			}
		case "PostToolUse":
			postTools++
			if l.ToolCallID == "" {
				t.Errorf("PostToolUse payload carried empty toolCall.id (agy 1.5.4+)")
			}
		}
		for _, secret := range []string{"ANTIGRAVITY_CSRF_TOKEN", "ANTIGRAVITY_LS_ADDRESS"} {
			if slices.Contains(l.EnvNames, secret) {
				t.Errorf("%s hook received %s: a hook can now record a link itself, and telepathy link's design should change", l.Event, secret)
			}
		}
	}
	if starts != 1 {
		t.Errorf("SessionStart fired %d times for a new conversation, want 1", starts)
	}
	if preTools == 0 || postTools == 0 {
		t.Errorf("PreToolUse=%d PostToolUse=%d on a turn with a tool call, want ≥1 each", preTools, postTools)
	}
	if len(invs) < 2 {
		t.Errorf("PreInvocation fired %d times on a turn with a tool call, want one per model step (≥2)", len(invs))
	}
	for i, n := range invs {
		if n != i {
			t.Errorf("PreInvocation invocationNums = %v, want 0, 1, 2 … in order", invs)
			break
		}
	}
	tr := w.transcript(t, conv)
	if !strings.Contains(tr, "EPHEMERAL_OBEYED") {
		t.Errorf("the model did not act on the injected ephemeral request (no EPHEMERAL_OBEYED in the transcript)")
	}
	for _, name := range []string{"ANTIGRAVITY_CSRF_TOKEN", "ANTIGRAVITY_LS_ADDRESS", "ANTIGRAVITY_CONVERSATION_ID"} {
		if !strings.Contains(tr, name) {
			t.Errorf("the agent's shell did not list %s; telepathy link reads it there\n%s", name, out)
		}
	}

	// Turn 2: the same conversation resumed in a new one-shot, on a new language server.
	if err := os.Truncate(logPath, 0); err != nil {
		t.Fatal(err)
	}
	w.run(t, conv, "Reply with the single word OK, then stop.", logEnv)
	second := readEvents(t, logPath)
	invs = nil
	for _, l := range second {
		if l.Event == "SessionStart" {
			t.Errorf("SessionStart fired on resume; a stale link could then be caught there instead of on PreInvocation")
		}
		if l.Event == "PreInvocation" && l.InvocationNum != nil {
			invs = append(invs, *l.InvocationNum)
		}
	}
	if len(invs) == 0 || invs[0] != 0 {
		t.Errorf("on resume, PreInvocation invocationNums = %v, want a turn starting at 0", invs)
	}

	assertNoAgyPanics(t, w.home)
}
