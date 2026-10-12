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
		{"another tool's JSON failure is a failure", "Bash", `{"command":"some-tool --json"}`,
			"Exit code 2\n{\"ok\":false}",
			[]string{"`Bash` failed on `some-tool --json`", `{"ok":false}`}},
		{"and so is an object that is not the envelope the record tool writes", "Bash", `{"command":"some-tool --json"}`,
			"Exit code 2\n{\"ok\":false,\"code\":\"validation\",\"error\":\"x\"}",
			[]string{"`Bash` failed on `some-tool --json`"}},
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

// A CHAIN'S FAILURE IS NOT ITS FIRST COMMAND'S. The call is the bench's at the wave B smoke's event
// 430: `register` succeeded and printed its line, and the `ls` chained after it exited 2. The entry
// read "`Bash` failed on `R=…; "$R/.bin/feov-record" --seat-id judge register --occasion doc…`:
// registered judge" — the command cut before the `ls`, the first command's success standing as the
// error — and the seat logged "register exits 2 after it succeeds" from the same reading.
func TestAChainedCallsFailureIsNotGivenToItsFirstCommand(t *testing.T) {
	const run = "/home/operator/.claude/scratch/universe-wave-b/src/research/2026-10-12_is-91-prime"
	command := `R=` + run + `; "$R/.bin/feov-record" --seat-id judge register --occasion docket --run "$R"; echo "---LAW"; ls -la "$R/inputs/law/" 2>&1`
	input, _ := json.Marshal(map[string]string{"command": command})
	last := "ls: cannot access '" + run + "/inputs/law/': No such file or directory"
	got, ok := failureEntry("Bash", input, "Exit code 2\nregistered judge\n2 STILL TO CLEAR before you may end your turn (of 2 item(s) open to you):\n---LAW\n"+last)
	if !ok {
		t.Fatal("the failed chain produced no entry")
	}
	if !strings.Contains(got, "`"+command+"`") {
		t.Errorf("the entry does not quote the whole command, so the part that failed is not in it:\n%s", got)
	}
	if !strings.Contains(got, "ends: "+last) {
		t.Errorf("the entry does not carry the output's last line, which is where the failing `ls` spoke:\n%s", got)
	}
	if strings.Contains(got, "`: registered judge") {
		t.Errorf("the entry gives the first command's success line as the call's error:\n%s", got)
	}
	if !strings.Contains(got, "does not say which part of it failed") {
		t.Errorf("the entry does not say the status is the whole call's:\n%s", got)
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
