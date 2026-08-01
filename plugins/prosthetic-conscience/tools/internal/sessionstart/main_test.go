package sessionstart

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
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
	text, watch := merge([]hookunit.Result{
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
	note := "---\nschema: 2\nobjective: finish the merge\n---\n" +
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
}
