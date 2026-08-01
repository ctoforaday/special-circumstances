package pretooluse

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

var noon = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

func fire(t *testing.T, payload string, units []hookunit.Unit) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(nil, strings.NewReader(payload), &o, &e, t.TempDir(), noon, units)
	return o.String(), e.String(), code
}

func stub(name, stdout, stderr string) hookunit.Unit {
	return hookunit.Unit{Name: name, Run: func(*hookunit.Ctx) hookunit.Result {
		return hookunit.Result{Name: name, Stdout: stdout, Stderr: stderr}
	}}
}

// Only ONE permission document may be emitted, so stdout is pick-first rather than
// composed — the opposite of PostToolUse, where every unit's stderr must survive.
func TestMergeTakesTheFirstDecisionAndKeepsEveryWarning(t *testing.T) {
	d, w := merge([]hookunit.Result{
		{Name: "a", Stderr: "warned first"},
		{Name: "b", Stdout: `{"deny":1}`, Stderr: "and denied"},
		{Name: "c", Stdout: `{"deny":2}`},
	})
	if d != `{"deny":1}` {
		t.Errorf("decision = %q; the first in unit order wins", d)
	}
	for _, want := range []string{"warned first", "and denied"} {
		if !strings.Contains(w, want) {
			t.Errorf("warnings lost %q: %q", want, w)
		}
	}
	if d, w := merge([]hookunit.Result{{Name: "a"}, {Name: "b"}}); d != "" || w != "" {
		t.Errorf("silence must stay silent: %q %q", d, w)
	}
}

// A warning is independent of a decision: the operator needs both facts, and a unit that
// only warns must still warn when another unit denies.
func TestWarningSurvivesADenial(t *testing.T) {
	stdout, stderr, code := fire(t, `{"tool_name":"Bash","tool_input":{"command":"x"}}`,
		[]hookunit.Unit{stub("gate", `{"permissionDecision":"deny"}`, ""), stub("freeze", "", "a run is LIVE")})
	if code != 0 {
		t.Fatalf("the block travels in the JSON, never the status: exit %d", code)
	}
	if !strings.Contains(stdout, "deny") {
		t.Errorf("stdout = %q", stdout)
	}
	if !strings.Contains(stderr, "a run is LIVE") {
		t.Errorf("the advisory unit's warning was suppressed by the denial: %q", stderr)
	}
}

// THE REGRESSION THIS EVENT IS MOST EXPOSED TO. #211 closed a bypass where an undecodable
// payload was allowed. A merge that handed units only the PARSED fragment would reintroduce
// it while every existing test still passed, because the fragment is empty either way.
func TestUndecodablePayloadStillDenies(t *testing.T) {
	truncated := `{"tool_name":"Bash","tool_input":{"command":"echo AKIAIOSFODNN7EXAMPLE"`

	stdout, _, code := fire(t, truncated, Units())
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, `"permissionDecision":"deny"`) {
		t.Fatalf("a truncated payload carrying a key got through the merged binary:\n%s", stdout)
	}
	if !strings.Contains(stdout, "could not be parsed") {
		t.Errorf("the reason must say the payload did not parse: %s", stdout)
	}
}

// The real units, end to end: each must act on its own matcher and no other.
func TestRealUnitsKeepTheirOwnMatchers(t *testing.T) {
	// A WebFetch carrying a key: the secrets gate watches it, the freeze guard does not.
	stdout, stderr, _ := fire(t, `{"tool_name":"WebFetch","tool_input":{"url":"https://x/?k=AKIAIOSFODNN7EXAMPLE"}}`, Units())
	if !strings.Contains(stdout, `"deny"`) {
		t.Errorf("the secrets gate must watch WebFetch: %s", stdout)
	}
	if strings.Contains(stderr, "push-freeze") {
		t.Errorf("the freeze guard claimed a WebFetch — merging must not widen a matcher: %s", stderr)
	}

	// A clean Bash call with no live run: both apply, neither has anything to say.
	stdout, stderr, code := fire(t, `{"tool_name":"Bash","tool_input":{"command":"git status"}}`, Units())
	if stdout != "" || stderr != "" || code != 0 {
		t.Errorf("a clean call must be silent: stdout=%q stderr=%q exit=%d", stdout, stderr, code)
	}

	// A tool neither unit watches.
	if stdout, stderr, _ := fire(t, `{"tool_name":"Read","tool_input":{"file_path":"a.go"}}`, Units()); stdout != "" || stderr != "" {
		t.Errorf("an unwatched tool must be silent: %q %q", stdout, stderr)
	}
}

// The emitted document must be exactly what the protocol expects — one JSON object.
func TestDecisionIsOneValidDocument(t *testing.T) {
	stdout, _, _ := fire(t, `{"tool_name":"Bash","tool_input":{"command":"echo AKIAIOSFODNN7EXAMPLE"}}`, Units())
	var d struct {
		HookSpecificOutput struct {
			HookEventName      string `json:"hookEventName"`
			PermissionDecision string `json:"permissionDecision"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(stdout), &d); err != nil {
		t.Fatalf("stdout is not one JSON document (%v): %q", err, stdout)
	}
	if d.HookSpecificOutput.HookEventName != "PreToolUse" || d.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("decision = %+v", d.HookSpecificOutput)
	}
}

// A panicking unit must not take the event — and must not silently become an ALLOW.
func TestPanicIsIsolatedAndReported(t *testing.T) {
	_, stderr, code := fire(t, `{"tool_name":"Bash","tool_input":{"command":"x"}}`, []hookunit.Unit{
		{Name: "boom", Run: func(*hookunit.Ctx) hookunit.Result { panic("gate exploded") }},
		stub("freeze", "", "still warned"),
	})
	if code != 0 {
		t.Errorf("exit %d", code)
	}
	for _, want := range []string{"panicked and was isolated", "still warned"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q: %q", want, stderr)
		}
	}
}
