package posttooluse

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

// fireFully is fire() with stdout kept: the message for the human travels there, where fire's
// callers only ever needed the model's channel.
func fireFully(t *testing.T, units []hookunit.Unit) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(nil, strings.NewReader(`{"tool_name":"Edit","tool_input":{"file_path":"a.go"}}`),
		&o, &e, t.TempDir(), noon, units)
	return o.String(), e.String(), code
}

func saidIn(t *testing.T, stdout string) string {
	t.Helper()
	if strings.TrimSpace(stdout) == "" {
		return ""
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v (%q)", err, stdout)
	}
	if len(doc) != 1 {
		t.Fatalf("this event answers with a message and nothing else: %q", stdout)
	}
	msg, _ := doc["systemMessage"].(string)
	return msg
}

// A GATE THAT IS NOT RUNNING is broken machine state a human can fix, and the line saying so used
// to reach the debug log alone. PostToolUse displays a systemMessage.
func TestASkippedGateIsSaid(t *testing.T) {
	stdout, stderr, code := fireFully(t, []hookunit.Unit{
		unit("sc-quality-gate", hookunit.Result{
			Stderr: "qlty not found — quality gate skipped (run /prosthetic-conscience:doctor --fix). file=a.go",
			Say:    "qlty not found — quality gate skipped (run /prosthetic-conscience:doctor --fix). file=a.go",
		})})
	if code != 0 {
		t.Fatalf("a skipped gate must not change the exit code: %d", code)
	}
	if !strings.Contains(saidIn(t, stdout), "qlty not found") {
		t.Errorf("stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "qlty not found") {
		t.Errorf("the debug log lost its copy: %q", stderr)
	}
}

// EXIT 2 AND A MESSAGE COEXIST, and that was measured rather than assumed (2026-09-16, real
// client): the blocking feedback reached the model AND the systemMessage was displayed. So a
// finding does not suppress an advisory that happens to fire on the same call.
func TestAFindingAndAMessageBothSurvive(t *testing.T) {
	stdout, stderr, code := fireFully(t, []hookunit.Unit{
		unit("finder", hookunit.Result{Stderr: "3 issues in a.go", Exit: 2}),
		unit("sayer", hookunit.Result{Stderr: "qlty not found", Say: "qlty not found"}),
	})
	if code != 2 {
		t.Fatalf("exit = %d, want 2 — the feedback channel is unchanged", code)
	}
	if !strings.Contains(stderr, "3 issues in a.go") {
		t.Errorf("the model's feedback was lost: %q", stderr)
	}
	if !strings.Contains(saidIn(t, stdout), "qlty not found") {
		t.Errorf("the human's message was lost: %q", stdout)
	}
}

// A quiet event stays quiet: no object on stdout for the client to parse.
func TestNothingToSaySaysNothing(t *testing.T) {
	stdout, _, code := fireFully(t, []hookunit.Unit{unit("quiet", hookunit.Result{})})
	if stdout != "" || code != 0 {
		t.Errorf("stdout = %q, exit %d", stdout, code)
	}
}

// The same duty on this event: a failure recorded where nothing could be said is announced here.
func TestARecordedFailureIsAnnouncedOnThisEvent(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	silent := hookfailures.New("prosthetic-conscience", "sc-filechanged-rearm", "FileChanged", noon, io.Discard)
	silent.Fail("rearm-write", "cannot update the re-arm state: read-only file system")
	if msg := silent.Settle(); msg != "" {
		t.Fatalf("FileChanged displays nothing and returned %q", msg)
	}

	stdout, _, _ := fireFully(t, []hookunit.Unit{unit("quiet", hookunit.Result{})})
	if !strings.Contains(saidIn(t, stdout), "rearm-write") {
		t.Fatalf("the next displaying event did not carry it: %q", stdout)
	}
}
