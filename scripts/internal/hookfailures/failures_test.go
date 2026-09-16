package hookfailures

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The authored package is tested HERE, once, in the module that owns it. Each plugin's own suite
// then tests its WIRING — that its hooks reach these calls — rather than re-testing the record.
//
// MUTATION PASS, 2026-09-16: 22 mutations of the call sites and branches below, one at a time
// (`~/.claude/scratch/hookvoice/mutate.py`, throwaway, not committed — the repo has no mutation
// tool and this is CLAUDE.md's "delete the row, invert the branch" done mechanically). 21 killed.
// The one survivor is EQUIVALENT, not a gap: forcing `merge` to save when nothing changed makes a
// healthy invocation call `save` with an empty record, which removes a file that does not exist and
// leaves exactly the state it found. Four earlier survivors were real gaps and are closed by the
// tests below — the displaying set is now stated by hand rather than read back from `Displays`, a
// scope's own success is asserted to clear it, `Last` is asserted to move while `Since` holds, and
// the record is asserted to be per plugin.

var noon = time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)

const (
	stageA Stage = "stage-a"
	stageB Stage = "stage-b"
)

// isolate points the record at a temporary state directory. Every test needs it: without it the
// package would read and WRITE the developer's own ~/.local/state.
func isolate(t *testing.T) (plugin, path string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	plugin = "test-plugin"
	p, err := Path(plugin)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(p, dir) {
		t.Fatalf("the record resolved outside the test's state dir: %s", p)
	}
	return plugin, p
}

// EACH PLUGIN GETS ITS OWN RECORD. Two plugins sharing one file would have each clearing stages it
// knows nothing about, and the stage names are only unique within a plugin.
func TestTheRecordIsPerPlugin(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	one, err := Path("plugin-one")
	if err != nil {
		t.Fatal(err)
	}
	two, err := Path("plugin-two")
	if err != nil {
		t.Fatal(err)
	}
	if one == two {
		t.Fatalf("two plugins share one record: %s", one)
	}
	if !strings.Contains(one, "plugin-one") || !strings.Contains(two, "plugin-two") {
		t.Errorf("a record does not name its plugin: %s / %s", one, two)
	}
}

func recorded(t *testing.T, path string) []Failure {
	t.Helper()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("the record does not parse: %v\n%s", err, b)
	}
	if rec.Schema != Schema {
		t.Fatalf("schema = %d, want %d", rec.Schema, Schema)
	}
	return rec.Failures
}

func entry(fs []Failure, s Stage, scope string) (Failure, bool) {
	for _, f := range fs {
		if f.Stage == s && f.Scope == scope {
			return f, true
		}
	}
	return Failure{}, false
}

// displayingEvents is written out HERE, by hand, and never read back from Displays: a test that asks
// the code under test what to expect agrees with it by construction. Measured against the real
// client for SessionStart, Stop, PreToolUse and PostToolUse (see the plan's §I); PostToolUseFailure
// is documented with them.
var displayingEvents = map[string]bool{
	"SessionStart": true, "PreToolUse": true, "PostToolUse": true, "PostToolUseFailure": true, "Stop": true,
	"PreCompact": false, "PostCompact": false, "SessionEnd": false,
	"SubagentStart": false, "SubagentStop": false, "FileChanged": false,
}

// A failure is recorded whatever the event, and SAID only on an event the client displays.
func TestAFailureIsRecordedAlwaysAndSaidOnlyWhereItIsDisplayed(t *testing.T) {
	for ev, shows := range displayingEvents {
		t.Run(ev, func(t *testing.T) {
			if Displays(ev) != shows {
				t.Fatalf("Displays(%s) = %v, want %v", ev, Displays(ev), shows)
			}
			plugin, path := isolate(t)
			var stderr strings.Builder
			r := New(plugin, "test-hook", ev, noon, &stderr)
			r.Fail(stageA, "the store would not open")
			msg := r.Settle()

			f, ok := entry(recorded(t, path), stageA, "")
			if !ok {
				t.Fatalf("%s: nothing recorded", ev)
			}
			if f.Error != "the store would not open" || f.Event != ev || !f.Since.Equal(noon) {
				t.Errorf("entry = %+v", f)
			}
			if !strings.Contains(stderr.String(), "test-hook: the store would not open") {
				t.Errorf("the debug log lost its line: %q", stderr.String())
			}
			if shows {
				if !strings.Contains(msg, "- stage-a: the store would not open") {
					t.Errorf("%s displays and said %q", ev, msg)
				}
				if f.Notified.IsZero() {
					t.Error("said, but the entry does not record that it was")
				}
			} else {
				if msg != "" {
					t.Errorf("%s displays nothing and said %q", ev, msg)
				}
				if !f.Notified.IsZero() {
					t.Error("not said, but the entry claims it was")
				}
			}
		})
	}
}

// SubagentStart and SubagentStop are the two events that must stay mute even if the set of
// displaying events is ever widened by hand: an emission there re-invokes the seat.
func TestTheSubagentEventsNeverDisplay(t *testing.T) {
	for _, ev := range []string{"SubagentStart", "SubagentStop"} {
		if Displays(ev) {
			t.Errorf("%s must never display: an emission there re-invokes the seat", ev)
		}
	}
}

// A stage's own success clears it, and nothing else does.
func TestOnlyTheStagesOwnSuccessClearsIt(t *testing.T) {
	plugin, path := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard)
	r.Fail(stageA, "a")
	r.Fail(stageB, "b")
	r.Settle()

	r = New(plugin, "test-hook", "Stop", noon.Add(time.Minute), io.Discard)
	r.OK(stageA)
	r.Settle()

	if _, ok := entry(recorded(t, path), stageA, ""); ok {
		t.Error("stage-a worked and its entry stayed")
	}
	if _, ok := entry(recorded(t, path), stageB, ""); !ok {
		t.Error("stage-a's success cleared stage-b")
	}
}

// Scope is part of the identity: a stage that works in one project says nothing about another.
func TestScopeIsPartOfTheIdentity(t *testing.T) {
	plugin, path := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard)
	r.FailIn(stageA, "/p/one", "broken here")
	r.Settle()

	r = New(plugin, "test-hook", "Stop", noon.Add(time.Minute), io.Discard)
	r.OKIn(stageA, "/p/two")
	r.Settle()

	if _, ok := entry(recorded(t, path), stageA, "/p/one"); !ok {
		t.Errorf("a success in another scope cleared it: %+v", recorded(t, path))
	}

	// ... and a success in the SAME scope does clear it, which is what makes the scope a key
	// rather than a label nothing reads.
	r = New(plugin, "test-hook", "Stop", noon.Add(2*time.Minute), io.Discard)
	r.OKIn(stageA, "/p/one")
	r.Settle()
	if _, ok := entry(recorded(t, path), stageA, "/p/one"); ok {
		t.Errorf("the scope worked and its entry stayed: %+v", recorded(t, path))
	}
}

// A failure and a success for the same stage in ONE invocation settle as a CONTINUING failure: its
// first time and its last announcement are kept, so it is not re-announced as new.
func TestAFailureOutranksASuccessInOneInvocation(t *testing.T) {
	plugin, path := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard)
	r.Fail(stageA, "boom")
	r.Settle()

	r = New(plugin, "test-hook", "Stop", noon.Add(time.Minute), io.Discard)
	r.Fail(stageA, "boom")
	r.OK(stageA)
	if msg := r.Settle(); msg != "" {
		t.Errorf("re-announced inside the quiet period: %q", msg)
	}
	f, ok := entry(recorded(t, path), stageA, "")
	if !ok || !f.Since.Equal(noon) || !f.Notified.Equal(noon) {
		t.Errorf("the failure restarted as new: %+v", f)
	}
}

// A stage that keeps failing is said once per NotifyEvery, and a NEW stage is not held back by an
// older one's quiet period.
func TestAFailingStageIsSaidOnceEveryNotifyPeriod(t *testing.T) {
	plugin, _ := isolate(t)
	say := func(at time.Time, stages ...Stage) string {
		r := New(plugin, "test-hook", "Stop", at, io.Discard)
		for _, s := range stages {
			r.Fail(s, string(s)+" is broken")
		}
		return r.Settle()
	}
	if msg := say(noon, stageA); !strings.Contains(msg, "stage-a") {
		t.Fatalf("first: %q", msg)
	}
	if msg := say(noon.Add(2*time.Minute), stageA); msg != "" {
		t.Fatalf("said again after 2 minutes: %q", msg)
	}
	msg := say(noon.Add(5*time.Minute), stageA, stageB)
	if !strings.Contains(msg, "stage-b") || strings.Contains(msg, "stage-a") {
		t.Fatalf("the new stage should be said alone: %q", msg)
	}
	if msg := say(noon.Add(NotifyEvery-time.Second), stageA); msg != "" {
		t.Fatalf("said a second early: %q", msg)
	}
	if msg := say(noon.Add(NotifyEvery), stageA); !strings.Contains(msg, "stage-a") {
		t.Fatalf("not said at the period: %q", msg)
	}
	// A failure that persists keeps the time it STARTED and moves the time it was last seen: the
	// first answers "how long has this been broken", which is the question a stale alarm hides.
	path, err := Path(plugin)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := entry(recorded(t, path), stageA, "")
	if !f.Since.Equal(noon) || !f.Last.Equal(noon.Add(NotifyEvery)) {
		t.Errorf("since/last = %s/%s, want %s/%s", f.Since, f.Last, noon, noon.Add(NotifyEvery))
	}
}

// A healthy invocation writes nothing at all — not even the directory.
func TestAHealthyInvocationWritesNothing(t *testing.T) {
	plugin, path := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard)
	r.OK(stageA)
	if msg := r.Settle(); msg != "" {
		t.Errorf("a healthy invocation spoke: %q", msg)
	}
	if _, err := os.Stat(filepath.Dir(path)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a healthy invocation created the record's directory (%v)", err)
	}
}

// With no record location, the failure is still said — unthrottled, because nothing remembers it —
// and the message says why. A record that cannot be kept must not become a second silent failure.
func TestWithNoRecordTheFailureIsStillSaid(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	if _, err := Path("test-plugin"); err == nil {
		t.Skip("this platform resolves a home directory without HOME or USERPROFILE")
	}
	for i := range 2 {
		var stderr strings.Builder
		r := New("test-plugin", "test-hook", "Stop", noon.Add(time.Duration(i)*time.Second), &stderr)
		r.Fail(stageA, "boom")
		msg := r.Settle()
		if !strings.Contains(msg, "- stage-a") || !strings.Contains(msg, "the failure record cannot be kept") {
			t.Fatalf("invocation %d: %q", i, msg)
		}
		if !strings.Contains(stderr.String(), "failure record not kept") {
			t.Errorf("the debug log does not say the record was not kept: %q", stderr.String())
		}
	}
}

// An unreadable record is REPLACED: kept, it would disable the throttle for good; read as empty, it
// would be the silent zero this record exists to prevent.
func TestAnUnreadableRecordIsReplaced(t *testing.T) {
	for name, body := range map[string]string{
		"not json":     "{torn",
		"other schema": `{"schema": 99, "failures": []}`,
	} {
		t.Run(name, func(t *testing.T) {
			plugin, path := isolate(t)
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			var stderr strings.Builder
			r := New(plugin, "test-hook", "Stop", noon, &stderr)
			r.Fail(stageA, "boom")
			msg := r.Settle()
			if !strings.Contains(stderr.String(), "replacing the failure record") {
				t.Errorf("replaced silently: %q", stderr.String())
			}
			if _, ok := entry(recorded(t, path), stageA, ""); !ok || !strings.Contains(msg, "stage-a") {
				t.Fatalf("the replacement lost this invocation's failure: %+v / %q", recorded(t, path), msg)
			}

			// And a healthy invocation replaces an unreadable record with no record at all.
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			r = New(plugin, "test-hook", "Stop", noon, io.Discard)
			r.OK(stageA)
			r.Settle()
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("an unreadable record survived a healthy invocation (%v)", err)
			}
		})
	}
}

// Emit writes the one object a hook with no document of its own uses to speak — and writes NOTHING
// for an empty message, because silence is a valid response and an empty object is not.
func TestEmitWritesOneObjectOrNothing(t *testing.T) {
	var out strings.Builder
	Emit(&out, "")
	if out.String() != "" {
		t.Errorf("an empty message wrote %q", out.String())
	}
	Emit(&out, "something failed")
	var obj map[string]any
	if err := json.Unmarshal([]byte(out.String()), &obj); err != nil {
		t.Fatalf("not one JSON object: %v (%q)", err, out.String())
	}
	if len(obj) != 1 || obj["systemMessage"] != "something failed" {
		t.Errorf("Emit wrote %q", out.String())
	}
	if strings.Count(out.String(), "\n") != 1 {
		t.Errorf("Emit wrote more than one line: %q", out.String())
	}
}

// Failed answers the caller that must compose its response BEFORE it settles.
func TestFailedReportsThisInvocation(t *testing.T) {
	plugin, _ := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard)
	if r.Failed() {
		t.Error("nothing failed and Failed() is true")
	}
	r.Fail(stageA, "boom")
	if !r.Failed() {
		t.Error("a failure was recorded and Failed() is false")
	}
}
