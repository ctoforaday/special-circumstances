package hookfailures

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

// SubagentStart and SubagentStop carry nothing to a HUMAN even if the set of displaying events is
// ever widened by hand. Widening it would put a systemMessage on a dispatch or a return, where no
// human is reading — and on Stop an emission of any kind re-invokes the seat.
func TestTheSubagentEventsNeverDisplay(t *testing.T) {
	for _, ev := range []string{"SubagentStart", "SubagentStop"} {
		if Displays(ev) {
			t.Errorf("%s must never display: no human reads a dispatch or a return", ev)
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

// BoundedStage associates a TTL with a stage so any event/harness merge prunes expired entries
// once shown, marks un-announced recoveries as Recovered to show once and drop, resets Since when
// a recovered entry fails again, and never loses still-failing or non-TTL stages.
func TestStageTTLSemantics(t *testing.T) {
	plugin, path := isolate(t)
	boundedA := BoundedStage{Stage: stageA, TTL: 30 * time.Minute}

	// 1. G2 + H3: Fail at t=0 on non-displaying PreInvocation, recover (OKIn) at +1m before ever being
	// shown -> displaying event at +2m shows it ONCE with singular recovered header & cleared footer
	// (never "failing since") and drops it.
	g2a := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	g2a.FailBounded(boundedA, "conv-rec", "drain failed")
	g2a.Persist()
	g2b := New(plugin, "test-hook", "PreInvocation", noon.Add(time.Minute), io.Discard)
	g2b.OKIn(stageA, "conv-rec")
	g2b.Persist()
	g2c := New(plugin, "test-hook", "PreToolUse", noon.Add(2*time.Minute), io.Discard)
	recMsg := g2c.Settle()
	if !strings.Contains(recMsg, "1 stage failed and has since worked") ||
		!strings.Contains(recMsg, "has since worked at") ||
		!strings.Contains(recMsg, "Each recovered entry above has been cleared.") ||
		strings.Contains(recMsg, "failing since") {
		t.Fatalf("singular recovered message unexpected: %q", recMsg)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("recovered entry survived its one-time announcement: %+v", recorded(t, path))
	}
	g2d := New(plugin, "test-hook", "PreToolUse", noon.Add(12*time.Minute), io.Discard)
	if again := g2d.Settle(); again != "" {
		t.Fatalf("recovered entry re-announced at +12m: %q", again)
	}

	// 1b. H3 plural recovered-only header: 2 recovered stages use plural "2 stages failed and have since worked".
	p1 := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	p1.FailBounded(boundedA, "conv-p1", "err1")
	p1.FailBounded(boundedA, "conv-p2", "err2")
	p1.Persist()
	p2 := New(plugin, "test-hook", "PreInvocation", noon.Add(time.Minute), io.Discard)
	p2.OKIn(stageA, "conv-p1")
	p2.OKIn(stageA, "conv-p2")
	p2.Persist()
	p3 := New(plugin, "test-hook", "PreToolUse", noon.Add(2*time.Minute), io.Discard)
	if pmsg := p3.Settle(); !strings.Contains(pmsg, "2 stages failed and have since worked") || !strings.Contains(pmsg, "Each recovered entry above has been cleared.") {
		t.Fatalf("plural recovered message unexpected: %q", pmsg)
	}

	// 1c. H2 + H5: fail@0 -> OK@1m (Recovered) -> fail@2m clears Recovered AND resets Since to +2m,
	// so Settle@3m renders "failing since <+2m>" and NOT "has since worked".
	h2a := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	h2a.FailBounded(boundedA, "conv-flaky", "first fail")
	h2a.Persist()
	h2b := New(plugin, "test-hook", "PreInvocation", noon.Add(time.Minute), io.Discard)
	h2b.OKIn(stageA, "conv-flaky")
	h2b.Persist()
	h2c := New(plugin, "test-hook", "PreInvocation", noon.Add(2*time.Minute), io.Discard)
	h2c.FailBounded(boundedA, "conv-flaky", "second fail")
	h2c.Persist()
	h2d := New(plugin, "test-hook", "PreToolUse", noon.Add(3*time.Minute), io.Discard)
	flakyMsg := h2d.Settle()
	wantSince := noon.Add(2 * time.Minute).UTC().Format(time.RFC3339)
	if !strings.Contains(flakyMsg, "failing since "+wantSince) || strings.Contains(flakyMsg, "has since worked") {
		t.Fatalf("fail->OK->fail message = %q, want 'failing since %s' without 'has since worked'", flakyMsg, wantSince)
	}
	// Clean up conv-flaky before continuing.
	h2e := New(plugin, "test-hook", "PreInvocation", noon.Add(4*time.Minute), io.Discard)
	h2e.OKIn(stageA, "conv-flaky")
	h2e.Persist()

	// 1d. G4: FailBounded (TTL 30m) followed by plain FailIn on the same key resets ExpiresAfter to 0.
	g4a := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	g4a.FailBounded(boundedA, "conv-g4", "bounded")
	g4a.Persist()
	g4b := New(plugin, "test-hook", "PreInvocation", noon.Add(time.Minute), io.Discard)
	g4b.FailIn(stageA, "conv-g4", "plain")
	g4b.Persist()
	if got, ok := entry(recorded(t, path), stageA, "conv-g4"); !ok || got.ExpiresAfter != 0 {
		t.Fatalf("plain FailIn after FailBounded kept sticky TTL: ok=%v got=%+v", ok, got)
	}
	g4c := New(plugin, "test-hook", "PreInvocation", noon.Add(2*time.Minute), io.Discard)
	g4c.OKIn(stageA, "conv-g4")
	g4c.Persist()

	// 2. Record a TTL stage on non-displaying PreInvocation and a non-TTL stageB.
	r1 := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	r1.FailBounded(boundedA, "conv-ended", "drain failed")
	r1.Fail(stageB, "permanent failure")
	r1.Persist()

	// 3. Un-announced (Notified is zero): +31m non-displaying merge must NOT prune before shown.
	r2 := New(plugin, "test-hook", "PreInvocation", noon.Add(31*time.Minute), io.Discard)
	r2.Persist()
	if _, ok := entry(recorded(t, path), stageA, "conv-ended"); !ok {
		t.Fatalf("un-announced TTL entry was pruned before ever being shown")
	}

	// 4. Displaying event (Stop, e.g. from Claude) at +31m shows it and stamps Notified.
	r3 := New(plugin, "test-hook", "Stop", noon.Add(31*time.Minute), io.Discard)
	msg := r3.Settle()
	if !strings.Contains(msg, "conv-ended") {
		t.Fatalf("displaying event at +31m did not show un-announced TTL entry: %q", msg)
	}

	// 5. Now that Notified is set and now-Last >= 30m, ANY subsequent merge prunes stageA while
	// leaving non-TTL stageB untouched.
	r4 := New(plugin, "test-hook", "Stop", noon.Add(32*time.Minute), io.Discard)
	r4.Persist()
	rec := recorded(t, path)
	if _, ok := entry(rec, stageA, "conv-ended"); ok {
		t.Fatalf("shown + expired TTL entry was not pruned: %+v", rec)
	}
	if _, ok := entry(rec, stageB, ""); !ok {
		t.Fatalf("non-TTL stageB was wrongly pruned: %+v", rec)
	}

	// 6. Still-failing entry ages from Last, NOT Since: fail at t=0, fail again at t=20m, show at
	// t=20m, settle at t=35m (now-Since=35m >= 30m, now-Last=15m < 30m) -> MUST be kept.
	r5a := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	r5a.FailBounded(boundedA, "conv-active", "still failing")
	r5a.Persist()
	r5b := New(plugin, "test-hook", "Stop", noon.Add(20*time.Minute), io.Discard)
	r5b.FailBounded(boundedA, "conv-active", "still failing")
	r5b.Settle()
	r5c := New(plugin, "test-hook", "Stop", noon.Add(35*time.Minute), io.Discard)
	r5c.Persist()
	if _, ok := entry(recorded(t, path), stageA, "conv-active"); !ok {
		t.Fatalf("still-failing entry was expired from Since instead of Last")
	}

	// 7. G3: Failing-now is never expired even if now-Last >= TTL, AND preserves Since and Notified.
	r6 := New(plugin, "test-hook", "Stop", noon.Add(60*time.Minute), io.Discard)
	r6.FailBounded(boundedA, "conv-active", "failing right now")
	r6.Persist()
	got, ok := entry(recorded(t, path), stageA, "conv-active")
	if !ok {
		t.Fatalf("entry failing in current invocation was wrongly expired")
	}
	if !got.Since.Equal(noon) || !got.Notified.Equal(noon.Add(20*time.Minute)) {
		t.Fatalf("failing-now entry was pruned and re-created: Since=%s Notified=%s, want %s / %s",
			got.Since, got.Notified, noon, noon.Add(20*time.Minute))
	}
}

// G6: A non-TTL stage recorded on a non-displaying event clears immediately on OK before ever being shown.
func TestNonTTLEntryClearsOnOKBeforeShown(t *testing.T) {
	plugin, path := isolate(t)
	r1 := New(plugin, "test-hook", "PreInvocation", noon, io.Discard)
	r1.Fail(stageB, "transient")
	r1.Persist()
	if _, ok := entry(recorded(t, path), stageB, ""); !ok {
		t.Fatalf("stageB was not persisted")
	}
	r2 := New(plugin, "test-hook", "PreInvocation", noon.Add(time.Minute), io.Discard)
	r2.OK(stageB)
	r2.Persist()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("non-TTL entry was not cleared by OK before being shown: %+v", recorded(t, path))
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

// CONCURRENT UNITS SHARE ONE RECORDER. prosthetic-conscience's hookunit fans an event's units out
// over goroutines and every one of them records; unguarded maps crashed the process with "concurrent
// map writes" the first time two units were wired in. A hook that panics is worse than the silence
// this package exists to end, so this runs under -race in CI.
func TestConcurrentUnitsMayShareARecorder(t *testing.T) {
	plugin, path := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard)
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := Stage(fmt.Sprintf("stage-%d", i%4))
			if i%2 == 0 {
				r.FailIn(s, fmt.Sprintf("/p/%d", i%3), "boom")
				return
			}
			r.OKIn(s, fmt.Sprintf("/p/%d", i%3))
		}(i)
	}
	wg.Wait()
	if !r.Failed() {
		t.Error("eight failures were recorded and Failed() is false")
	}
	if msg := r.Settle(); msg == "" {
		t.Error("nothing was said for a record with failures in it")
	}
	if len(recorded(t, path)) == 0 {
		t.Error("nothing reached the record")
	}
}

// PERSIST RECORDS AND MARKS NOTHING AS SAID. A Settle whose message is discarded on a displaying event
// stamps the entry as notified, so the human is never told and the next call that could tell them
// stays quiet for NotifyEvery. Persist is what a caller that cannot speak uses instead.
func TestPersistRecordsWithoutMarkingAnythingSaid(t *testing.T) {
	plugin, path := isolate(t)
	r := New(plugin, "test-hook", "Stop", noon, io.Discard) // Stop DISPLAYS
	r.Fail(stageA, "could not speak on this call")
	r.Persist()

	f, ok := entry(recorded(t, path), stageA, "")
	if !ok {
		t.Fatal("Persist did not record")
	}
	if !f.Notified.IsZero() {
		t.Fatalf("Persist marked the entry as said: %+v", f)
	}
	// The next call that CAN speak says it at once.
	next := New(plugin, "test-hook", "Stop", noon.Add(time.Second), io.Discard)
	if msg := next.Settle(); !strings.Contains(msg, "could not speak on this call") {
		t.Errorf("the next displaying call did not say it: %q", msg)
	}
}

// ENSURE RETURNS WHAT fetch-bin.sh WOULD HAVE SAID, for a binary to merge into its one response —
// and nothing when there is nothing to say or no plugin root to say it from.
func TestEnsureReturnsTheVersionChecksMessage(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fetch-bin.sh is a POSIX shell script; the hooks run it through sh")
	}
	root := t.TempDir()
	script := filepath.Join(root, "hooks", "fetch-bin.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(script, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("CLAUDE_PLUGIN_ROOT", root)

	write(`[ "$1" = ensure ] && [ "$2" = SessionStart ] && printf '{"systemMessage":"installing v2","hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"installing v2"}}\n'`)
	if got, err := Ensure("SessionStart"); err != nil || got != "installing v2" {
		t.Errorf("a stale install's message did not come back: %q %v", got, err)
	}
	// Every installed plugin carries tools/go.mod; it marks no checkout, and the check still runs.
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "go.mod"), []byte("module tools\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := Ensure("SessionStart"); err != nil || got != "installing v2" {
		t.Errorf("an installed root holding tools/go.mod skipped its version check: %q %v", got, err)
	}
	write(`exit 0`)
	if got, err := Ensure("SessionStart"); err != nil || got != "" {
		t.Errorf("an up-to-date install said something: %q %v", got, err)
	}
	write(`echo not json`)
	if _, err := Ensure("SessionStart"); err == nil {
		t.Error("output that is not the message document was taken as a message")
	}
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	if got, err := Ensure("SessionStart"); err != nil || got != "" {
		t.Errorf("outside a plugin hook there is nothing to ensure: %q %v", got, err)
	}
}

func TestJoinMessages(t *testing.T) {
	for _, tc := range []struct {
		a, b, want string
	}{
		{"", "", ""},
		{"a", "", "a"},
		{"", "b", "b"},
		{"a", "b", "a\n\nb"},
	} {
		if got := JoinMessages(tc.a, tc.b); got != tc.want {
			t.Errorf("JoinMessages(%q, %q) = %q, want %q", tc.a, tc.b, got, tc.want)
		}
	}
}
