package sessionstart

import (
	"bytes"
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hooktest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

var noon = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

func fire(t *testing.T, dir, payload string, units []hookunit.Unit) (out hookOutput, spoke bool, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(nil, strings.NewReader(payload), &o, &e, dir, noon, units)
	if strings.TrimSpace(o.String()) == "" {
		return out, false, code
	}
	if err := json.Unmarshal(o.Bytes(), &out); err != nil {
		t.Fatalf("emitted non-JSON %q: %v", o.String(), err)
	}
	return out, true, code
}

func stub(name, text string, watch ...string) hookunit.Unit {
	return hookunit.Unit{Name: name, Run: func(*hookunit.Ctx) hookunit.Result {
		return hookunit.Result{Name: name, Stdout: text, Watch: watch}
	}}
}

// COMPOSE, not pick-first: every unit's text survives and watchPaths is the union. This is
// the opposite of PreToolUse, where only one document may be emitted.
func TestMergeComposesTextAndUnionsPaths(t *testing.T) {
	text, watch, _ := merge([]hookunit.Result{
		{Name: "a", Stdout: "nudge line", Watch: []string{"tools"}},
		{Name: "b", Stdout: "recovered state", Watch: []string{"tools", "docs"}},
		{Name: "c"},
	})
	for _, want := range []string{"nudge line", "recovered state"} {
		if !strings.Contains(text, want) {
			t.Errorf("text lost %q: %q", want, text)
		}
	}
	if strings.Index(text, "nudge line") > strings.Index(text, "recovered state") {
		t.Error("text must follow unit order")
	}
	if len(watch) != 2 || watch[0] != "tools" || watch[1] != "docs" {
		t.Errorf("watch = %v; want the deduplicated union in first-seen order", watch)
	}
}

// A note can be worth WATCHING with nothing to say about it, so either half is reason to
// speak — and neither half means silence, because an empty response would spend tokens
// telling the session there is nothing to tell it.
func TestEitherHalfSpeaksAndNeitherIsSilent(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct {
		name  string
		units []hookunit.Unit
		spoke bool
	}{
		{"nothing at all", []hookunit.Unit{stub("a", "")}, false},
		{"whitespace only", []hookunit.Unit{stub("a", "   ")}, false},
		{"text alone", []hookunit.Unit{stub("a", "something")}, true},
		{"watch alone", []hookunit.Unit{stub("a", "", "tools")}, true},
		{"both", []hookunit.Unit{stub("a", "something", "tools")}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			out, spoke, code := fire(t, dir, `{"source":"startup"}`, c.units)
			if code != 0 {
				t.Fatalf("a SessionStart hook that fails is a session that fails: exit %d", code)
			}
			if spoke != c.spoke {
				t.Fatalf("spoke = %v, want %v", spoke, c.spoke)
			}
			if spoke && out.HookSpecificOutput.HookEventName != "SessionStart" {
				t.Errorf("hookEventName = %q", out.HookSpecificOutput.HookEventName)
			}
		})
	}
}

// THE DEFECT THE MERGE EXPOSED: two units used to write on different channels, and one
// process cannot emit a bare line and a JSON document on the same stdout. Everything must
// now travel as additionalContext — the channel §5 of the spike actually verified.
func TestEverythingTravelsAsAdditionalContext(t *testing.T) {
	out, spoke, _ := fire(t, t.TempDir(), `{"source":"startup"}`,
		[]hookunit.Unit{stub("nudge", "missing tool(s): qlty"), stub("restore", "recovered state", "tools")})
	if !spoke {
		t.Fatal("expected a response")
	}
	for _, want := range []string{"missing tool(s): qlty", "recovered state"} {
		if !strings.Contains(out.HookSpecificOutput.AdditionalContext, want) {
			t.Errorf("additionalContext lost %q: %q", want, out.HookSpecificOutput.AdditionalContext)
		}
	}
	if len(out.HookSpecificOutput.WatchPaths) != 1 {
		t.Errorf("watchPaths = %v", out.HookSpecificOutput.WatchPaths)
	}
}

// A panicking unit must not cost the other's contribution, nor the session.
func TestPanicIsIsolated(t *testing.T) {
	out, spoke, code := fire(t, t.TempDir(), `{"source":"startup"}`, []hookunit.Unit{
		{Name: "boom", Run: func(*hookunit.Ctx) hookunit.Result { panic("restore exploded") }},
		stub("nudge", "still here"),
	})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !spoke || !strings.Contains(out.HookSpecificOutput.AdditionalContext, "still here") {
		t.Errorf("the surviving unit's text was lost: %+v", out.HookSpecificOutput)
	}
}

// End to end with the real units: a live checkpoint must still be restored, and its
// trigger surfaces still registered.
func TestRealUnitsRestoreANote(t *testing.T) {
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	note := "---\nschema: 3\nobjective: finish the merge\n---\n" +
		"## Validation loop\n1. `go test ./...`  → ok  · re-armed by: tools/\n   last run: pass\n"
	if err := os.WriteFile(filepath.Join(cp, "CHECKPOINT.md"), []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}

	out, spoke, code := fire(t, dir, `{"source":"startup"}`, Units(""))
	if code != 0 || !spoke {
		t.Fatalf("expected a restore: spoke=%v exit=%d", spoke, code)
	}
	if !strings.Contains(out.HookSpecificOutput.AdditionalContext, "finish the merge") {
		t.Errorf("the objective did not reach the session: %q", out.HookSpecificOutput.AdditionalContext)
	}
	if len(out.HookSpecificOutput.WatchPaths) == 0 {
		t.Error("the note's trigger surface was not registered")
	}
	if strings.Contains(out.HookSpecificOutput.AdditionalContext, "After this compaction") {
		t.Errorf("a startup carries the compaction resume line: %q", out.HookSpecificOutput.AdditionalContext)
	}

	// The restore unit runs last, so after a compaction the merged response ends with its line.
	out, _, _ = fire(t, dir, `{"source":"compact"}`, Units(""))
	if !strings.HasSuffix(strings.TrimRight(out.HookSpecificOutput.AdditionalContext, "\n"), "first next action, or ask if that action is the human's to decide.") {
		t.Errorf("the compaction response does not end with the resume line: %q", out.HookSpecificOutput.AdditionalContext)
	}
}

// No test in this package may write the developer's own state.
func TestMain(m *testing.M) { os.Exit(hooktest.Isolated(m)) }

// THE VERSION CHECK SPEAKS IN THIS HOOK'S ONE RESPONSE. It was a second SessionStart hook beside this
// one; folded in, its message rides the systemMessage this binary already sends, and a check that
// cannot run is recorded rather than lost.
func TestTheVersionCheckSpeaksInTheOneResponse(t *testing.T) {
	prev := ensure
	t.Cleanup(func() { ensure = prev })

	ensure = func(event string) (string, error) {
		if event != "SessionStart" {
			t.Errorf("the check was asked about %q", event)
		}
		return "installing prosthetic-conscience v2", nil
	}
	out, spoke, _ := fire(t, t.TempDir(), `{"source":"startup"}`, []hookunit.Unit{stub("a", "")})
	if !spoke || !strings.Contains(out.SystemMessage, "installing prosthetic-conscience v2") {
		t.Fatalf("the version check's message is not in the response: spoke=%v %+v", spoke, out)
	}

	ensure = func(string) (string, error) { return "", errString("fetch-bin.sh is missing") }
	out, spoke, _ = fire(t, t.TempDir(), `{"source":"startup"}`, []hookunit.Unit{stub("a", "")})
	if !spoke || !strings.Contains(out.SystemMessage, "fetch-bin.sh is missing") {
		t.Fatalf("a version check that could not run went unsaid: spoke=%v %+v", spoke, out)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestPreInvocationDrainsFailedRunCommandsIntoStrikesAndFlushesPending(t *testing.T) {
	t.Cleanup(func() { ensure = hookfailures.Ensure })
	ensure = func(string) (string, error) { return "", nil }

	proj := t.TempDir()
	pluginData := t.TempDir()
	t.Setenv("PLUGIN_DATA", pluginData)

	// Queue a pending feedback message (as sc-posttooluse does on Antigravity).
	hookcore.QueueSay(pluginData, "conv-preinv-1", "QUEUED_SAY_MSG")

	// Write a transcript with 3 failing run_command calls of the same command so the 3-strike
	// rule trips on PreInvocation.
	tp := filepath.Join(t.TempDir(), "transcript_full.jsonl")
	tr := strings.Join([]string{
		`{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"go test ./..."}}]}`,
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 1.\n\nCommand output:\nFAIL\n"}`,
		`{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"go test ./..."}}]}`,
		`{"step_index":4,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 1.\n\nCommand output:\nFAIL\n"}`,
		`{"step_index":5,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"go test ./..."}}]}`,
		`{"step_index":6,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 1.\n\nCommand output:\nFAIL\n"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(tp, []byte(tr), 0o600); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreInvocation",
		"conversationId": "conv-preinv-1",
		"cwd":            proj,
		"transcriptPath": tp,
		"invocationNum":  1,
	})
	var stdout, stderr bytes.Buffer
	unitRan := false
	units := []hookunit.Unit{{
		Name: "should-not-run",
		Run: func(*hookunit.Ctx) hookunit.Result {
			unitRan = true
			return hookunit.Result{Stdout: "SESSIONSTART_ONLY"}
		},
	}}
	code := run([]string{"-event", "PreInvocation"}, bytes.NewReader(payload), &stdout, &stderr, proj, time.Now(), units)
	if code != 0 {
		t.Fatalf("PreInvocation exit %d, stderr=%s", code, stderr.String())
	}
	if unitRan {
		t.Error("PreInvocation must not run SessionStart units")
	}
	outStr := stdout.String()
	if !strings.Contains(outStr, "QUEUED_SAY_MSG") {
		t.Errorf("PreInvocation did not flush queued Say: %s", outStr)
	}
	if !strings.Contains(outStr, "anti-spinning:") || !strings.Contains(outStr, "failed 3 times") {
		t.Errorf("PreInvocation did not trip 3-strike warning from drained run_command failures: %s", outStr)
	}
}
