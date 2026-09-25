package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
)

// PRETOOLUSE IS FEOV'S ONLY DISPLAYING EVENT, so Run is where every failure this plugin records —
// here or on the sitting events, which display nothing to a human — reaches one. These tests drive Run.

func runEntry(t *testing.T, event string, f Entry) (stdout string) {
	t.Helper()
	var out bytes.Buffer
	if code := Run("feov-test", event, f, strings.NewReader(`{}`), &out); code != 0 {
		t.Fatalf("Run exited %d — a hook must never block", code)
	}
	return out.String()
}

func messageOf(t *testing.T, stdout string) (doc map[string]any, msg string) {
	t.Helper()
	if strings.TrimSpace(stdout) == "" {
		return nil, ""
	}
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("stdout is not one JSON object: %v (%q)", err, stdout)
	}
	msg, _ = doc["systemMessage"].(string)
	return doc, msg
}

// A PANIC IS SWALLOWED — a hook that dies denies every tool call — AND SAID, which it never was.
func TestAPanicIsSaidOnPreToolUse(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	out := runEntry(t, "PreToolUse", func(io.Reader, io.Writer, *hookfailures.Recorder) error { panic("nil map") })
	if _, msg := messageOf(t, out); !strings.Contains(msg, "- hook-panic:") || !strings.Contains(msg, "nil map") {
		t.Fatalf("a panic was not said: %q", out)
	}
}

func TestAnErrorIsSaidOnPreToolUse(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	out := runEntry(t, "PreToolUse", func(io.Reader, io.Writer, *hookfailures.Recorder) error {
		return errors.New("the gate could not read its input")
	})
	if _, msg := messageOf(t, out); !strings.Contains(msg, "- hook-error:") {
		t.Fatalf("an error was not said: %q", out)
	}
}

// THE DECISION AND THE MESSAGE TRAVEL IN ONE DOCUMENT (measured: the client honours both).
func TestADecisionCarriesTheMessage(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	out := runEntry(t, "PreToolUse", func(_ io.Reader, w io.Writer, rec *hookfailures.Recorder) error {
		rec.Fail("hook-error", "something broke")
		emitPreDeny(w, "past the sitting limit")
		return nil
	})
	doc, msg := messageOf(t, out)
	hso, _ := doc["hookSpecificOutput"].(map[string]any)
	if hso["permissionDecision"] != "deny" || hso["permissionDecisionReason"] != "past the sitting limit" {
		t.Fatalf("the decision did not survive: %q", out)
	}
	if !strings.Contains(msg, "something broke") {
		t.Errorf("the message is not in the response: %q", out)
	}
}

// A DECISION THIS BINARY CANNOT DECODE GOES OUT EXACTLY AS WRITTEN. It is what protects the session.
func TestAnUndecodableDecisionGoesOutIntact(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	out := runEntry(t, "PreToolUse", func(_ io.Reader, w io.Writer, rec *hookfailures.Recorder) error {
		rec.Fail("hook-error", "something broke")
		_, _ = w.Write([]byte(`{"permissionDecision": NOT JSON`))
		return nil
	})
	if out != `{"permissionDecision": NOT JSON` {
		t.Fatalf("the decision was altered: %q", out)
	}
}

// A SITTING EVENT SAYS NOTHING ABOUT ITS OWN FAULT, whatever failed — the failure goes on the record
// and the next displaying event reads it out. Asserted on BOTH ends, because a panicking entry has
// written no document and there is nothing legitimate left to emit; the systemMessage a displaying
// event would carry is what must not appear here.
//
// THIS IS NOT "the sitting events never speak", which is what it used to say and covered both events
// with a measurement of one. SubagentStop may never emit at all — an emission re-invokes the seat,
// nine firings measured — and that is asserted at the source, in sittinghook. SubagentStart's
// document is how a seat's work list reaches it; what it may not do is turn its own fault into output.
func TestASittingEventNeverSpeaksItsOwnFault(t *testing.T) {
	for _, ev := range []string{"SubagentStart", "SubagentStop"} {
		t.Run(ev, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", t.TempDir())
			out := runEntry(t, ev, func(io.Reader, io.Writer, *hookfailures.Recorder) error { panic("boom") })
			if out != "" {
				t.Fatalf("%s wrote to stdout: %q", ev, out)
			}
		})
	}
}

// THE CARRIER, THROUGH Run: a sitting fails where nothing can be said, and the seat's next tool call
// — PreToolUse — says it.
func TestASittingFailureIsSaidOnTheNextToolCall(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	runEntry(t, "SubagentStop", func(_ io.Reader, _ io.Writer, rec *hookfailures.Recorder) error {
		rec.Fail("sitting-write", "feov-sitting-write: exit status 1: disk full")
		return nil
	})
	out := runEntry(t, "PreToolUse", func(io.Reader, io.Writer, *hookfailures.Recorder) error { return nil })
	if _, msg := messageOf(t, out); !strings.Contains(msg, "sitting-write") || !strings.Contains(msg, "disk full") {
		t.Fatalf("the sitting failure was not carried to the next tool call: %q", out)
	}
}

// AN UNUSABLE MARKER ON PRETOOLUSE: the run directory was simply never injected, and the seat then
// hit "no run directory" from a different layer. Recorded in the marker's project; cleared by a
// usable one; and two runs open — legitimate — is not a fault.
func TestPreRecordsOnlyAFaultyMarker(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	proj := t.TempDir()
	marker := filepath.Join(proj, ".claude", "run-live.json")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	pre := func(body string) string {
		if err := os.WriteFile(marker, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		payload := `{"tool_name":"Bash","tool_input":{"command":"ls"},"cwd":` + jsonString(proj) + `}`
		var out bytes.Buffer
		Run("feov-pretooluse", "PreToolUse", Pre, strings.NewReader(payload), &out)
		return out.String()
	}
	stage := string(runlive.StageUnusable)

	if _, msg := messageOf(t, pre("{ not json")); !strings.Contains(msg, "- "+stage+" in "+proj) {
		t.Fatalf("an unreadable marker was not said: %q", msg)
	}
	if err := os.MkdirAll(filepath.Join(proj, "run"), 0o755); err != nil {
		t.Fatal(err)
	}
	pre(`{"runs":[{"runDir":"run"}]}`)
	if onRecord(t, stage, proj) {
		t.Error("a usable marker left the entry")
	}
	for _, d := range []string{"a", "b"} {
		if err := os.MkdirAll(filepath.Join(proj, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pre(`{"runs":[{"runDir":"a"},{"runDir":"b"}]}`)
	if onRecord(t, stage, proj) {
		t.Error("two runs open were recorded as a fault")
	}
}

func onRecord(t *testing.T, stage, scope string) bool {
	t.Helper()
	path, err := hookfailures.Path("frank-exchange-of-views")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	for _, f := range rec.Failures {
		if string(f.Stage) == stage && f.Scope == scope {
			return true
		}
	}
	return false
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// THE TURN-LIMIT FAULT. The seat is still refused — enforcement survives — but a limit the record
// could not be told about used to be a stderr line nobody read. Recorded now, and said.
func TestATurnLimitTheRecordCannotHoldIsSaid(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, runDir, _ := limitRun(t, 1)
	sit(t, runDir, "agent_x", "red-lens-evidence", 1)
	recordLimit = func(string, string, string, int, int) error {
		return errors.New("feov-sitting-write is not beside this hook, so agent_x sitting 1 reaching the limit of 1 is not on the record")
	}

	pre := func() (map[string]any, string) {
		p, _ := json.Marshal(map[string]any{
			"tool_name": "Read", "tool_input": map[string]string{"file_path": "/x"}, "cwd": cwd,
			"agent_id": "agent_x", "agent_type": "frank-exchange-of-views:red-lens-evidence",
		})
		var out bytes.Buffer
		Run("feov-pretooluse", "PreToolUse", Pre, bytes.NewReader(p), &out)
		return messageOf(t, out.String())
	}
	var doc map[string]any
	var msg string
	for i := 0; i < 3 && msg == ""; i++ { // the first call is under the limit; a later one is refused
		doc, msg = pre()
	}
	hso, _ := doc["hookSpecificOutput"].(map[string]any)
	if hso["permissionDecision"] != "deny" {
		t.Fatalf("the seat past its limit was not refused: %v", doc)
	}
	if !strings.Contains(msg, "- turn-limit:") || !strings.Contains(msg, "not on the record") {
		t.Fatalf("the limit the record could not hold was not said: %q", msg)
	}
}

// EVERY FAILURE RUN RECORDS IS CLEARED BY THE NEXT CALL THAT DOES NOT HAVE IT. Seeded, then driven
// clean; asserted against the record FILE, because an entry already shown is held back by the
// 10-minute throttle whether it was cleared or not.
func TestACleanCallClearsWhatRunRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	seed := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "PreToolUse", time.Now(), io.Discard)
	for _, s := range []hookfailures.Stage{StagePanic, StageError, StageTurnLimit} {
		seed.Fail(s, "seeded")
	}
	seed.Settle()

	var out bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", Pre, strings.NewReader(`{"tool_name":"Read","tool_input":{"file_path":"/x"}}`), &out)
	for _, s := range []hookfailures.Stage{StagePanic, StageError, StageTurnLimit} {
		if onRecord(t, string(s), "") {
			t.Errorf("a clean call left %s on the record", s)
		}
	}
}

// A PAYLOAD PRETOOLUSE CANNOT READ: no turn limit counted, no run directory injected, for this call.
func TestAnUnreadablePreToolUsePayloadIsSaidAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	var out bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", Pre, strings.NewReader("{ not json"), &out)
	if _, msg := messageOf(t, out.String()); !strings.Contains(msg, "- "+string(StageInput)+":") {
		t.Fatalf("an unreadable payload was not said: %q", out.String())
	}
	out.Reset()
	Run("feov-pretooluse", "PreToolUse", Pre, strings.NewReader(`{"tool_name":"Read","tool_input":{"file_path":"/x"}}`), &out)
	if onRecord(t, string(StageInput), "") {
		t.Error("a readable payload left the entry")
	}
}

// AN UNPARSABLE BASH TOOL_INPUT IN A LIVE RUN is recorded — and the DECISION DOES NOT CHANGE: no
// document is written for it, exactly as before, because changing decisions is outside this work.
func TestAnUnparsableToolInputIsRecordedWithoutADecision(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, _, _ := limitRun(t, 100)
	payload := `{"tool_name":"Bash","tool_input":"not an object","cwd":` + jsonString(cwd) + `}`
	var out bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", Pre, strings.NewReader(payload), &out)
	doc, msg := messageOf(t, out.String())
	if _, decided := doc["hookSpecificOutput"]; decided {
		t.Errorf("a decision was emitted where there was none before: %q", out.String())
	}
	if !strings.Contains(msg, "- "+string(StageToolInput)+":") {
		t.Fatalf("an unparsable tool_input was not said: %q", out.String())
	}
}

type brokenPipe struct{}

func (brokenPipe) Write([]byte) (int, error) { return 0, errors.New("write |1: broken pipe") }

// A DECISION THAT CANNOT BE DELIVERED is recorded, and the next delivered call — whose response
// carries the message — clears it.
func TestAnUndeliveredResponseIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	deny := func(_ io.Reader, w io.Writer, _ *hookfailures.Recorder) error {
		emitPreDeny(w, "refused")
		return nil
	}
	Run("feov-pretooluse", "PreToolUse", deny, strings.NewReader(`{}`), brokenPipe{})
	if !onRecord(t, string(StageDeliver), "") {
		t.Fatal("an undelivered response was not recorded")
	}
	var out bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", deny, strings.NewReader(`{}`), &out)
	if _, msg := messageOf(t, out.String()); !strings.Contains(msg, "- "+string(StageDeliver)+":") {
		t.Errorf("the delivery failure was not said on the next call: %q", out.String())
	}
	if onRecord(t, string(StageDeliver), "") {
		t.Error("a delivered response left the entry — the OK after the first Settle never reached the record")
	}
}

// THE LIMIT MARKER CANNOT BE CREATED. That folded silently into "a sibling call got there first", so
// no call ever counted as first and the sitting's limit never reached the record. It is recorded now
// — and the seat is STILL REFUSED, because the refusal is enforcement and the marker is bookkeeping.
func TestAnUncreatableLimitMarkerIsRecordedAndTheSeatIsStillRefused(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, runDir, handoffs := limitRun(t, 1)
	sit(t, runDir, "agent_m", "red-lens-evidence", 1)
	if d, _ := call(t, cwd, "agent_m", "Read", ""); d != "" { // under the limit: counted, and .calls now exists
		t.Fatalf("setup: the first call was refused (%q)", d)
	}
	counters := filepath.Join(runDir, sittingcap.Dir)
	if err := os.Chmod(counters, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(counters, 0o755) })
	// A CAPABILITY probe, not a uid check: skip only if the directory is still writable.
	if f, err := os.CreateTemp(counters, "probe-"); err == nil {
		f.Close()
		_ = os.Remove(f.Name())
		t.Skip("chmod 0500 did not stop this process creating files in the counters directory")
	}

	p, _ := json.Marshal(map[string]any{
		"tool_name": "Read", "tool_input": map[string]string{"file_path": "/x"}, "cwd": cwd,
		"agent_id": "agent_m", "agent_type": "frank-exchange-of-views:red-lens-evidence",
	})
	var out bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", Pre, bytes.NewReader(p), &out)
	doc, msg := messageOf(t, out.String())
	hso, _ := doc["hookSpecificOutput"].(map[string]any)
	if hso["permissionDecision"] != "deny" {
		t.Fatalf("the seat past its limit was NOT refused — bookkeeping must never weaken enforcement: %q", out.String())
	}
	if !strings.Contains(msg, "- turn-limit:") || !strings.Contains(msg, "cannot mark") {
		t.Fatalf("the uncreatable marker was not said: %q", msg)
	}
	if len(*handoffs) != 0 {
		t.Errorf("the limit was handed to the record although no call counted as first: %v", *handoffs)
	}
}

// EVERY OUTCOME CLEARS AN UNPARSABLE TOOL_INPUT — a rewrite, a deny, and a call with nothing to do
// all mean the next tool_input parsed. Seeded, driven down each path, asserted against the record.
func TestEachOutcomeClearsAToolInputFailure(t *testing.T) {
	for name, tc := range map[string]struct {
		tool, command string
		want          string // the decision the path produces, "" for none
	}{
		"rewrite": {"Bash", "ls", "allow"},
		// The gate's own pinned fixture (hookgate/substitution_test.go): a backtick inside a
		// double-quoted free-text flag, which the shell would run.
		"deny": {"Bash", "feov-record mint --fix \"replace it with `x`\" --reason 'r'", "deny"},
		// A Bash call PreOutcome has no opinion on. A non-Bash call is NOT this case: it returns
		// before tool_input is looked at, and says nothing about whether the next one parses.
		"none": {"Bash", "", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", t.TempDir())
			cwd, _, _ := limitRun(t, 100)
			seed := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "PreToolUse", time.Now(), io.Discard)
			seed.Fail(StageToolInput, "seeded")
			seed.Persist()

			ti := map[string]string{"file_path": "/x"}
			if tc.tool == "Bash" {
				ti = map[string]string{"command": tc.command}
			}
			p, _ := json.Marshal(map[string]any{"tool_name": tc.tool, "tool_input": ti, "cwd": cwd})
			var out bytes.Buffer
			Run("feov-pretooluse", "PreToolUse", Pre, bytes.NewReader(p), &out)
			doc, _ := messageOf(t, out.String())
			hso, _ := doc["hookSpecificOutput"].(map[string]any)
			if got, _ := hso["permissionDecision"].(string); got != tc.want {
				t.Fatalf("setup: the %s path produced decision %q, want %q (%q)", name, got, tc.want, out.String())
			}
			if onRecord(t, string(StageToolInput), "") {
				t.Errorf("the %s path left the tool-input entry on the record", name)
			}
		})
	}
}
