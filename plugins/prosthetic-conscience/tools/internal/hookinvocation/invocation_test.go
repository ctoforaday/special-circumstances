// Package hookinvocation tests the SHIPPED COMMAND LINES against the BUILT BINARIES.
//
// Nothing else in this plugin does. Every other test constructs a run() and calls it
// in-process, which cannot see the two failures that actually shipped:
//
//   - `1082275` — the hook could not run the binary it names, so the plugin disabled
//     every session it was installed into. In-process tests were green throughout.
//   - `e40b35d` — the hook was registered as a verb rather than a program.
//
// Both are properties of the COMMAND STRING and the PROCESS, not of any function. This
// file reads hooks.json, builds what it names, runs it the way the client would, and
// asserts the I/O contract that the client actually depends on.
package hookinvocation

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// binRef pulls the binary name out of the shipped command string. The commands are
// `B="${CLAUDE_PLUGIN_ROOT}/bin/<name>"; if [ -x "$B" ] …`, and that shape is itself
// load-bearing: a bare path to a missing binary is what disabled sessions in 1082275,
// so the wrapper degrades instead of failing the event.
var binRef = regexp.MustCompile(`\$\{CLAUDE_PLUGIN_ROOT\}/bin/([a-z0-9-]+)`)

type hooksFile struct {
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func load(t *testing.T) hooksFile {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "hooks", "hooks.json"))
	if err != nil {
		t.Fatalf("cannot read the shipped hooks.json: %v", err)
	}
	var h hooksFile
	if err := json.Unmarshal(b, &h); err != nil {
		t.Fatalf("shipped hooks.json is not valid JSON: %v", err)
	}
	if len(h.Hooks) == 0 {
		t.Fatal("hooks.json declares no hooks; this test would pass vacuously")
	}
	return h
}

// registered maps each event to the binary it invokes.
func registered(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for event, entries := range load(t).Hooks {
		for _, e := range entries {
			for _, h := range e.Hooks {
				m := binRef.FindStringSubmatch(h.Command)
				if m == nil {
					t.Errorf("%s: command names no ${CLAUDE_PLUGIN_ROOT}/bin/<binary>: %q", event, h.Command)
					continue
				}
				out[event] = m[1]
			}
		}
	}
	return out
}

// REGISTRATION PARITY, which NOTHING ELSE GATES.
//
// scripts/pluginparity does not read hooks.json (grep returns 0) and the only hooks.json
// check in CI tests bootstrap-guard degradation. So a binary that is built, declared in
// requirements.json and counted in the docs — but never registered — passes every other
// command in the suite, and sits on disk doing nothing while the gates read green.
//
// This is the check for the direction that matters: every registration must name a
// binary that exists. The reverse (a cmd/ with no registration) is deliberately NOT an
// error — sc-doctor is a command an operator runs, not a hook.
func TestEveryRegistrationNamesABinaryThatExists(t *testing.T) {
	for event, bin := range registered(t) {
		if _, err := os.Stat(filepath.Join("..", "..", "cmd", bin)); err != nil {
			t.Errorf("%s registers %q, which has no cmd/ directory: %v", event, bin, err)
		}
	}
}

// The degradation wrapper is not decoration. A bare path to a missing binary made every
// hook fail its event, which disabled the session (1082275). Every command must test for
// the binary and exit cleanly when it is absent.
func TestEveryCommandDegradesWhenTheBinaryIsMissing(t *testing.T) {
	for event, entries := range load(t).Hooks {
		for _, e := range entries {
			for _, h := range e.Hooks {
				if !strings.Contains(h.Command, "if [ -x") {
					t.Errorf("%s: command does not check the binary exists before running it — "+
						"a missing binary would fail the event rather than degrade:\n%s", event, h.Command)
				}
			}
		}
	}
	// And prove the degradation actually works, rather than trusting the shape: run one
	// command with a plugin root that holds no binaries at all.
	var cmd *exec.Cmd
	for _, entries := range load(t).Hooks {
		cmd = exec.Command("bash", "-c", entries[0].Hooks[0].Command)
		break
	}
	cmd.Env = append(os.Environ(), "CLAUDE_PLUGIN_ROOT="+t.TempDir())
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.Output()
	if err != nil {
		t.Errorf("a hook with no binary present exited non-zero (%v); it must degrade, not fail the event", err)
	}
	if len(out) != 0 {
		t.Errorf("a degraded hook wrote to STDOUT: %q — stdout is the client's channel and must stay clean", out)
	}
}

// exeSuffix is what Windows requires to execute a file at all.
//
// The SHIPPED hook command already knows this — `if [ -x "$B" ] || [ -x "$B.exe" ]` — and
// this test did not, so it built extensionless binaries and then could not start them.
// The plugin's own command string was the documentation for the platform difference.
func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// build compiles every registered binary once, into one directory laid out the way the
// plugin ships: <root>/bin/<name>.
func build(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, name := range registered(t) {
		if seen[name] {
			continue
		}
		seen[name] = true
		cmd := exec.Command("go", "build", "-o", filepath.Join(bin, name+exeSuffix()), "./cmd/"+name)
		cmd.Dir = filepath.Join("..", "..") // the tools module root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("building %s: %v\n%s", name, err, out)
		}
	}
	return root
}

// stdoutContract is what the CLIENT does with a hook's stdout, per event.
//
// It differs by event, which is exactly why this must be asserted rather than assumed:
// PreCompact's stdout becomes the custom compact instructions (spike §5, VERIFIED), so
// plain prose there is correct. Everywhere else the client parses stdout as the
// hookSpecificOutput document, so anything that is neither empty nor JSON is corruption
// — and it fails silently, because a client that cannot parse it simply ignores it.
func stdoutContract(event string) string {
	if event == "PreCompact" {
		return "prose" // becomes the compact instructions
	}
	return "json-or-empty"
}

// marshal builds a payload the way the client does. Concatenating a path into a JSON
// literal is not portable: on Windows the separators become escape sequences, and \t is
// a tab.
func marshal(t *testing.T, kv map[string]any) string {
	t.Helper()
	b, err := json.Marshal(kv)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// THE I/O CONTRACT, against the real processes, with the payloads a client sends and
// several it never would.
func TestTheShippedBinariesHonourTheIOContract(t *testing.T) {
	root := build(t)
	project := t.TempDir()
	if err := os.MkdirAll(filepath.Join(project, ".claude", "checkpoints"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".claude", "checkpoints", "CHECKPOINT.md"),
		[]byte("---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. go test ./...\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	payloads := map[string]string{
		"realistic":   marshal(t, map[string]any{"session_id": "s1", "cwd": project, "hook_event_name": "X", "trigger": "auto", "stop_hook_active": false}),
		"empty":       ``,
		"not json":    `{`,
		"json null":   `null`,
		"wrong types": `{"session_id":42,"cwd":[],"stop_hook_active":"yes"}`,
	}

	for event, name := range registered(t) {
		for label, payload := range payloads {
			t.Run(event+"/"+label, func(t *testing.T) {
				cmd := exec.Command(filepath.Join(root, "bin", name+exeSuffix()))
				cmd.Env = append(os.Environ(),
					"CLAUDE_PLUGIN_ROOT="+root,
					"CLAUDE_PROJECT_DIR="+project,
				)
				cmd.Dir = project
				cmd.Stdin = strings.NewReader(payload)
				var out, errb strings.Builder
				cmd.Stdout, cmd.Stderr = &out, &errb

				done := make(chan error, 1)
				if err := cmd.Start(); err != nil {
					t.Fatalf("cannot start %s: %v", name, err)
				}
				go func() { done <- cmd.Wait() }()
				select {
				case err := <-done:
					// A hook that exits non-zero can deny a tool call or fail an event. None
					// of these are gates, so all of them must exit 0 whatever they are fed.
					if err != nil {
						t.Errorf("%s (%s) exited non-zero: %v\nstderr: %s", name, label, err, errb.String())
					}
				case <-time.After(20 * time.Second):
					_ = cmd.Process.Kill()
					t.Fatalf("%s (%s) did not terminate — a hanging hook stalls the event", name, label)
				}

				s := strings.TrimSpace(out.String())
				if s == "" {
					return // silence is always contract-clean
				}
				switch stdoutContract(event) {
				case "json-or-empty":
					var any map[string]any
					if err := json.Unmarshal([]byte(s), &any); err != nil {
						t.Errorf("%s (%s) wrote non-JSON to stdout, which the client parses as its "+
							"response document and will silently ignore:\n%s", name, label, s)
					}
				case "prose":
					// PreCompact's stdout IS the instruction text; any content is valid. What
					// would not be is JSON pretending to be a response document.
					if strings.HasPrefix(s, "{") {
						t.Errorf("%s (%s) wrote a JSON document to stdout, but this event's stdout "+
							"becomes the compact instructions verbatim:\n%s", name, label, s)
					}
				}
			})
		}
	}
}
