package fuzz

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
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
	if runtime.GOOS == "windows" {
		t.Skip("hooks.json's command is /bin/sh; the binary's argv contract is covered by the same gate on Linux")
	}
	// The binary each command names is DERIVED from the command, never listed here. A hand-kept
	// roster of hook binaries beside the manifest that already names them is the same defect one
	// level up: it goes stale silently, and a gate reading a stale roster builds the wrong thing
	// and reports the right one sound.
	pluginRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(pluginRoot, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	hooksPath, err := repotree.Plugin("hooks", "hooks.json")
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
			if strings.Contains(m.Matcher, "Bash") {
				tool = "Bash"
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
				installHookBinary(t, pluginRoot, h.Command)
				cmd := exec.Command("sh", "-c", h.Command)
				cmd.Dir = sandbox
				cmd.Env = append(os.Environ(),
					"CLAUDE_PLUGIN_ROOT="+pluginRoot,
					"CLAUDE_PROJECT_DIR="+sandbox,
				)
				cmd.Stdin = strings.NewReader(string(pb))
				out, err := cmd.CombinedOutput()
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
				_ = err
			}
		}
	}
	if driven == 0 {
		t.Fatal("drove no hooks at all — hooks.json parsed to an empty set, and a gate that runs " +
			"nothing reports every hook sound")
	}
	t.Logf("%d shipped hook invocation(s) driven against the built binary", driven)
}

// binFromCommand pulls the ${CLAUDE_PLUGIN_ROOT}/bin/<name> a hook command resolves.
var binFromCommand = regexp.MustCompile(`\$\{CLAUDE_PLUGIN_ROOT\}/bin/([A-Za-z0-9._-]+)`)

// installHookBinary builds the command's named binary from ./cmd/<name> and lays it out the way
// an install does, because ${CLAUDE_PLUGIN_ROOT}/bin/<name> is part of what is under test.
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
	toolsDir, err := repotree.Plugin("tools")
	if err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", dest, "./cmd/"+name)
	build.Dir = toolsDir // ./cmd is the MODULE's, not this test package's
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("building cmd/%s, which hooks.json invokes: %v\n%s", name, err, out)
	}
}
