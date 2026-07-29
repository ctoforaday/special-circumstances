package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/trajectory"
)

const sample = `{"type":"assistant","uuid":"a-1","timestamp":"2026-07-28T13:20:05Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"REC=./feov-record ; \"$REC\" finding --seat-id red-1"}}]}}
{"type":"assistant","uuid":"a-2","attributionAgent":"red-auditor","message":{"content":[{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/w/notes.txt"}}]}}
`

func call(t *testing.T, body string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, &o, &e, func(string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	})
	return o.String(), e.String(), code
}

// Regression for the argument-order bug that only showed up when the tool was
// actually run: `tools -binary X file` must parse the flag.
func TestFlagsAfterSubcommandAreParsed(t *testing.T) {
	stdout, _, code := call(t, sample, "tools", "-binary", "feov-record", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "aliased via $REC") {
		t.Fatalf("the -binary flag was ignored; output:\n%s", stdout)
	}
}

// The companion bug: a flag VALUE must never be mistaken for the subcommand.
// Supporting flags-before-subcommand is ambiguous, so it is an error, not a guess.
func TestFlagValueIsNotMistakenForTheSubcommand(t *testing.T) {
	_, stderr, code := call(t, sample, "-binary", "feov-record", "tools", "/w/s.jsonl")
	if code == 0 {
		t.Fatal("flags before the subcommand should be rejected, not silently misread")
	}
	if strings.Contains(stderr, `unknown command "feov-record"`) {
		t.Fatal("the flag's value was taken as the subcommand — the bug this guards")
	}
}

// Provenance travels with every row; that is what makes a finding checkable.
func TestJSONRowsCarryProvenance(t *testing.T) {
	stdout, _, code := call(t, sample, "tools", "-json", "-binary", "feov-record", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("row count = %d, want 2", len(lines))
	}
	for i, l := range lines {
		var r row
		if err := json.Unmarshal([]byte(l), &r); err != nil {
			t.Fatalf("row %d is not JSON: %v", i, err)
		}
		if r.File == "" || r.Line == 0 || r.UUID == "" {
			t.Errorf("row %d cannot be cited: %+v", i, r)
		}
	}
	var first row
	_ = json.Unmarshal([]byte(lines[0]), &first)
	if len(first.Invocations) != 1 || first.Invocations[0].Verb != "finding" || !first.Invocations[0].Aliased {
		t.Errorf("aliased invocation not resolved in JSON: %+v", first.Invocations)
	}
}

// An answer over a subset must not read like an answer over the whole.
func TestUnparseableLinesWarnOnStderr(t *testing.T) {
	_, stderr, code := call(t, sample+"{not json\n", "tools", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d — a bad line is a warning, not a hard failure", code)
	}
	for _, want := range []string{"WARNING", "SUBSET", "did not parse"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q; got %q", want, stderr)
		}
	}
}

// Refuse to emit what cannot be checked, and say how much was withheld.
func TestUncitableEventsAreSuppressedAndCounted(t *testing.T) {
	// A tool_use on a line with no uuid cannot be cited.
	noUUID := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"t9","name":"Glob","input":{}}]}}` + "\n"
	stdout, stderr, code := call(t, noUUID, "tools", "/w/s.jsonl")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(stdout, "Glob") {
		t.Error("emitted an event with no provenance")
	}
	if !strings.Contains(stderr, "suppressed") {
		t.Errorf("suppression was silent; stderr = %q", stderr)
	}
}

func TestUsageAndVersion(t *testing.T) {
	if _, _, code := call(t, sample); code != 2 {
		t.Errorf("no args should be a usage error, got %d", code)
	}
	if _, _, code := call(t, sample, "mine", "/w/s.jsonl"); code != 2 {
		t.Errorf("unknown command should be an error, got %d", code)
	}
	stdout, _, code := call(t, sample, "-version")
	if code != 0 || !strings.Contains(stdout, "gray-area") {
		t.Errorf("version: code=%d stdout=%q", code, stdout)
	}
}

// The target is a reading convenience; it must never be the only handle on a row.
func TestTargetPrefersCommandThenPath(t *testing.T) {
	e := trajectory.Event{ToolInput: json.RawMessage(`{"command":"ls","file_path":"/x"}`)}
	if got := target(e); got != "ls" {
		t.Errorf("target = %q, want the command", got)
	}
	e = trajectory.Event{ToolInput: json.RawMessage(`{"file_path":"/x"}`)}
	if got := target(e); got != "/x" {
		t.Errorf("target = %q, want the file path", got)
	}
	if got := target(trajectory.Event{}); got != "" {
		t.Errorf("target on empty input = %q, want empty", got)
	}
}
