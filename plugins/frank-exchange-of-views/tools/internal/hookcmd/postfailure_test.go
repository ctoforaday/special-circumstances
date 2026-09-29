package hookcmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// A SEAT'S FAILED CALL IS LOGGED, AIMED AND WORDED AS IT FAILED — and nothing the record tool already
// logged, nothing the operator did, and no interrupt. Each row is a failure universe-m13 or m14
// actually produced and nobody filed.
func TestAFailedCallBecomesTheSeatsFailureEntry(t *testing.T) {
	for _, tc := range []struct {
		name, tool, input, err string
		want                   []string // "" means no entry
	}{
		{"a jq that misread the shape", "Bash", `{"command":"feov-record inquest telemetry --json | jq '.[-1]'"}`,
			"Exit code 5\njq: error (at <stdin>:1): Cannot index object with number",
			[]string{"`Bash` failed on `feov-record inquest telemetry --json | jq '.[-1]'`", "jq: error (at <stdin>:1): Cannot index object with number"}},
		{"the seat's command, not the engine's exports", "Bash",
			`{"command":"export FEOV_RUN='/r'; export FEOV_AGENT_ID='a1'; jq '.open' board.json"}`,
			"Exit code 5\njq: error: Cannot index array",
			[]string{"`Bash` failed on `jq '.open' board.json`"}},
		{"a Read at a guessed path", "Read", `{"file_path":"/run/references/catechism_template.md"}`,
			"File does not exist. Note: your current working directory is /src.",
			[]string{"`Read` failed on `/run/references/catechism_template.md`", "File does not exist."}},
		{"the record tool's own refusal is its entry, not this", "Bash", `{"command":"feov-record avenue move --key AV1"}`,
			"Exit code 2\nfeov-record: unknown flag: --key", nil},
		{"and its --json refusal too", "Bash", `{"command":"feov-record --json mint"}`,
			"Exit code 2\n{\"ok\":false,\"code\":\"validation\",\"error\":\"…\"}", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := failureEntry(tc.tool, json.RawMessage(tc.input), tc.err)
			if tc.want == nil {
				if ok {
					t.Fatalf("an already-logged refusal was logged again: %q", got)
				}
				return
			}
			if !ok {
				t.Fatal("a seat's failed call produced no entry")
			}
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Errorf("the entry does not carry %q: %s", w, got)
				}
			}
			if strings.Contains(got, "Exit code") {
				t.Errorf("the harness's exit-code line is not the failure: %s", got)
			}
		})
	}
}

// THE HOOK HANDS A SEAT'S FAILURE TO THE WRITER, and only a seat's, only in a run, never an interrupt.
func TestTheFailureHookLogsOnlyASeatsOwnFailures(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	var logged []string
	prev := recordFailure
	recordFailure = func(dir, agent, text string) error {
		if dir != runDir {
			t.Errorf("the failure was attributed to run %q, not %q", dir, runDir)
		}
		logged = append(logged, agent+" | "+text)
		return nil
	}
	t.Cleanup(func() { recordFailure = prev })
	fire := func(p map[string]any) {
		t.Helper()
		p["cwd"] = cwd
		b, _ := json.Marshal(p)
		if err := PostFailure(bytes.NewReader(b), &bytes.Buffer{}, testRecorder()); err != nil {
			t.Fatal(err)
		}
	}
	read := map[string]any{"tool_name": "Read", "tool_input": map[string]string{"file_path": "/x"}, "error": "File does not exist."}
	with := func(extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range read {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	fire(with(map[string]any{"agent_id": "a1"}))
	fire(with(map[string]any{"agent_id": "a1", "is_interrupt": true}))
	fire(with(map[string]any{})) // the main session: no agent id
	if len(logged) != 1 || !strings.HasPrefix(logged[0], "a1 | `Read` failed on `/x`") {
		t.Fatalf("want exactly the seat's one real failure, got %q", logged)
	}
}
