package fuzz

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// THE SHIPPED HOOK'S OWN COMMAND LINE, RUN AGAINST THE BUILT BINARY. Nothing else in the suite
// does this, and the gap is the whole reason the plugin shipped able to disable the session it
// was installed into.
//
// MEASURED 2026-08-22. hooks.json invokes `feov-record hook pretooluse` with no --seat-id. When
// the surface became scoped by seat id, no-identity stopped resolving ANY command, so the root
// refused it — and RefuseAndTeach exits 2. Exit 2 from a PreToolUse hook DENIES the tool call.
// The matcher is Write|Edit|MultiEdit|NotebookEdit|Bash, so that is every mutating tool call in
// the session, for anyone whose plugin bin/ holds the binary — the state `doctor --fix` produces.
// Both hooks blocked; the suite was green throughout.
//
// WHY NOTHING SAW IT. hookgate is tested in-process and cli tests construct a root and call it in
// Go, so both ends were exercised and neither was the ARGV. The change that broke it updated
// eight test files to pass --seat-id, which is exactly what kept the suite green: the tests could
// be taught the new model and hooks.json could not, because a hook has no identity to give.
//
// THE INVARIANT IS THE HOOK'S OWN, and it was stated in a doc comment two levels below where it
// could be broken — internal/hookcmd: "EVERY ENTRY POINT RETURNS 0 — the decision travels in the stdout JSON,
// never the exit status is not a channel." A change to the ROOT's argument parsing broke a contract documented on
// the leaf. This asserts it at the process boundary, which is the altitude it lives at.
func TestTheShippedHooksResolveAgainstTheBuiltBinary(t *testing.T) {
	releaseGate(t)
	if runtime.GOOS == "windows" {
		t.Skip("hooks.json's command is /bin/sh; the binary's argv contract is covered by the same gate on Linux")
	}
	sweepPath(t)
	// The binary each command names is DERIVED from the command, never listed here. A hand-kept
	// roster of hook binaries beside the manifest that already names them is the same defect one
	// level up: it goes stale silently, and a gate reading a stale roster builds the wrong thing
	// and reports the right one sound.
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	hooksPath, err := repotree.Plugin(".claude-plugin", "hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("hooks.json does not parse: %v", err)
	}

	sandbox := t.TempDir()
	driven := 0
	for event, matchers := range manifest.Hooks {
		for _, m := range matchers {
			// The matcher decides which tool names reach the hook; drive one the matcher names,
			// so the payload is one this hook actually receives.
			tool := "Write"
			switch {
			case strings.Contains(m.Matcher, "Bash"):
				tool = "Bash"
			case m.Matcher != "" && !strings.ContainsAny(m.Matcher, "|*.^$()[]"):
				// A matcher naming one tool is driven with that tool,
				// so the hook takes the path it takes in a session rather than its early return.
				tool = m.Matcher
			}
			payload := map[string]any{
				"session_id":      "gate",
				"transcript_path": filepath.Join(sandbox, "t.jsonl"),
				"cwd":             sandbox,
				"hook_event_name": event,
				"tool_name":       tool,
				"tool_input":      map[string]any{"command": "echo hi", "file_path": filepath.Join(sandbox, "x.txt"), "content": "hi"},
				"tool_response":   map[string]any{"stdout": "hi"},
			}
			pb, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			for _, h := range m.Hooks {
				driven++
				if event != "SessionStart" || h.Command != ensureCommand {
					installHookBinary(t, pluginRoot, h.Command)
				}
				cmd := exec.Command("sh", "-c", h.Command)
				cmd.Dir = sandbox
				cmd.Env = append(os.Environ(),
					"CLAUDE_PLUGIN_ROOT="+pluginRoot,
					"CLAUDE_PROJECT_DIR="+sandbox,
					// Every hook here records the failures it swallows (internal/hookfailures), and
					// the record resolves from XDG_STATE_HOME before the home directory: without this
					// a driven hook writes into the developer's own ~/.local/state.
					"XDG_STATE_HOME="+filepath.Join(sandbox, ".state"),
				)
				cmd.Stdin = strings.NewReader(string(pb))
				out, err := cmd.CombinedOutput()
				if cmd.ProcessState == nil {
					t.Errorf("%s hook %q did not run: %v", event, h.Command, err)
					continue
				}
				code := cmd.ProcessState.ExitCode()
				if code == 0 {
					continue
				}
				t.Errorf("%s hook exits %d on an ordinary %s call, so Claude Code BLOCKS it.\n\n"+
					"  command: %s\n  output: %s\n\n"+
					"Every hook in this manifest must exit 0 whatever it decides — the decision travels in "+
					"the stdout JSON, never the exit status (internal/hookcmd says so, and hookcmd.Run enforces it). A non-zero exit here is "+
					"not a failed check, it is every matched tool call in the session denied: this matcher "+
					"is %q.",
					event, code, tool, h.Command, strings.TrimSpace(string(out)), m.Matcher)
			}
		}
	}
	if driven == 0 {
		t.Fatal("drove no hooks at all — hooks.json parsed to an empty set, and a gate that runs " +
			"nothing reports every hook sound")
	}

	// Drive root hooks.json (Antigravity) and enforce strict .exa.hooks_pb stdout schemas.
	agyHooksPath, err := repotree.Plugin("hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	agyRaw, err := os.ReadFile(agyHooksPath)
	if err != nil {
		t.Fatal(err)
	}
	type agyEntry struct {
		Command string `json:"command"`
		Hooks   []struct {
			Command string `json:"command"`
		} `json:"hooks"`
	}
	var agyTop map[string]map[string][]agyEntry
	if err := json.Unmarshal(agyRaw, &agyTop); err != nil {
		t.Fatalf("root hooks.json does not parse: %v", err)
	}
	const agyEnsureCommand = `R="${PWD}"; F="$R/hooks/fetch-bin.sh"; if [ -f "$F" ]; then exec sh "$F" ensure SessionStart; fi`
	pluginData := t.TempDir()
	for _, events := range agyTop {
		for event, entries := range events {
			var commands []string
			for _, e := range entries {
				if e.Command != "" {
					commands = append(commands, e.Command)
				}
				for _, h := range e.Hooks {
					if h.Command != "" {
						commands = append(commands, h.Command)
					}
				}
			}
			for _, command := range commands {
				driven++
				if command != agyEnsureCommand {
					installAgyHookBinary(t, pluginRoot, command)
				}
				payload := map[string]any{
					"hookEventName":       event,
					"conversationId":      "conv-fuzz-1873",
					"cwd":                 sandbox,
					"workspacePaths":      []string{sandbox},
					"invocationNum":       0,
					"continuation_active": false,
					"toolCall": map[string]any{
						"id":   "call_fuzz_1",
						"name": "run_command",
						"args": map[string]any{"CommandLine": "echo hi"},
					},
				}
				pb, err := json.Marshal(payload)
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command("sh", "-c", command)
				cmd.Dir = pluginRoot
				cmd.Env = append(os.Environ(),
					"ANTIGRAVITY_CONVERSATION_ID=conv-fuzz-1873",
					"PLUGIN_ROOT="+pluginRoot,
					"PLUGIN_DATA="+pluginData,
					"XDG_STATE_HOME="+filepath.Join(sandbox, ".state"),
				)
				cmd.Stdin = strings.NewReader(string(pb))
				var out, errb bytes.Buffer
				cmd.Stdout, cmd.Stderr = &out, &errb
				if err := cmd.Run(); err != nil {
					t.Errorf("Antigravity %s hook %q exited non-zero: %v\nstderr: %s", event, command, err, errb.String())
					continue
				}
				s := strings.TrimSpace(out.String())
				if s == "" {
					continue
				}
				dec := json.NewDecoder(bytes.NewReader([]byte(s)))
				dec.DisallowUnknownFields()
				var allowed map[string]any
				switch event {
				case "PreInvocation", "SessionStart":
					var res struct {
						InjectSteps         []map[string]any `json:"injectSteps"`
						TerminationBehavior string           `json:"terminationBehavior"`
					}
					if err := dec.Decode(&res); err != nil {
						t.Errorf("Antigravity %s stdout failed strict %sHookResult: %v\n%s", event, event, err, s)
					}
				case "PreToolUse":
					var res struct {
						Decision   string         `json:"decision"`
						Reason     string         `json:"reason"`
						DenyReason string         `json:"deny_reason"`
						Overwrite  map[string]any `json:"overwrite"`
					}
					if err := dec.Decode(&res); err != nil {
						t.Errorf("Antigravity %s stdout failed strict PreToolHookResult: %v\n%s", event, err, s)
					}
				case "PostToolUse":
					var res struct{}
					if err := dec.Decode(&res); err != nil {
						t.Errorf("Antigravity %s stdout failed strict PostToolHookResult: %v\n%s", event, err, s)
					}
				case "Stop":
					var res struct {
						Decision string `json:"decision"`
						Reason   string `json:"reason"`
					}
					if err := dec.Decode(&res); err != nil {
						t.Errorf("Antigravity %s stdout failed strict StopHookResult: %v\n%s", event, err, s)
					}
				default:
					_ = allowed
				}
			}
		}
	}
	t.Logf("%d shipped hook invocation(s) driven against the built binary", driven)
}

// ensureCommand is the ONE hooks.json command that runs no binary: SessionStart's version check,
// which fetches this plugin's release when the installed binaries are not its version. There is
// nothing to build for it; it is still driven below, and must exit 0 like every other hook — in
// this sandbox its script is absent, so its own existence guard is what exits. scripts/fetchbingen
// holds every plugin to exactly this string and fails when this copy drifts from it.
const ensureCommand = `F="${CLAUDE_PLUGIN_ROOT}/hooks/fetch-bin.sh"; if [ -f "$F" ]; then exec sh "$F" ensure SessionStart; fi`

// binFromCommand pulls the ${CLAUDE_PLUGIN_ROOT}/bin/<name> a hook command resolves.
var binFromCommand = regexp.MustCompile(`\$\{CLAUDE_PLUGIN_ROOT\}/bin/([A-Za-z0-9._-]+)`)

var agyBinFromCommand = regexp.MustCompile(`\$R/bin/([A-Za-z0-9._-]+)`)

func installAgyHookBinary(t *testing.T, pluginRoot, command string) {
	t.Helper()
	m := agyBinFromCommand.FindStringSubmatch(command)
	if m == nil {
		t.Fatalf("Antigravity hook command names no $R/bin/<binary>:\n  %s", command)
	}
	name := m[1]
	dest := filepath.Join(pluginRoot, "bin", name)
	if _, err := os.Stat(dest); err == nil {
		return
	}
	src, err := repotree.Plugin("tools", "cmd", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("hooks.json invokes %q, and there is no cmd/%s to build it from: %v", name, name, err)
	}
	built, err := os.ReadFile(testbuild.Binary(t, name))
	if err != nil {
		t.Fatalf("reading built cmd/%s: %v", name, err)
	}
	if err := os.WriteFile(dest, built, 0o755); err != nil {
		t.Fatalf("laying out cmd/%s at %s: %v", name, dest, err)
	}
}

// installHookBinary builds the command's named binary from ./cmd/<name> through testbuild, which
// builds each command once per test binary (or takes it from the CI job's prebuilt set), and lays a
// copy out the way an install does, because ${CLAUDE_PLUGIN_ROOT}/bin/<name> is part of what is
// under test.
func installHookBinary(t *testing.T, pluginRoot, command string) {
	t.Helper()
	m := binFromCommand.FindStringSubmatch(command)
	if m == nil {
		t.Fatalf("this hook command names no ${CLAUDE_PLUGIN_ROOT}/bin/<binary>, so the gate cannot "+
			"build what it invokes:\n  %s", command)
	}
	name := m[1]
	dest := filepath.Join(pluginRoot, "bin", name)
	if _, err := os.Stat(dest); err == nil {
		return // already built for an earlier event
	}
	src, err := repotree.Plugin("tools", "cmd", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("hooks.json invokes %q, and there is no cmd/%s to build it from (%v) — the manifest "+
			"and the tree disagree about which binaries this plugin ships", name, name, err)
	}
	built, err := os.ReadFile(testbuild.Binary(t, name))
	if err != nil {
		t.Fatalf("reading the built cmd/%s, which hooks.json invokes: %v", name, err)
	}
	if err := os.WriteFile(dest, built, 0o755); err != nil {
		t.Fatalf("laying out cmd/%s at %s: %v", name, dest, err)
	}
}
