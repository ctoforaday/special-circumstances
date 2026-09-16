package pretooluse

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

// A unit that only ever WARNS has its warning as its whole mechanism, and until now that warning
// went to stderr at exit 0 — the client's debug log, and nobody. These tests hold the response to
// carrying it, beside a decision when there is one.

func saying(name, say string) hookunit.Unit {
	return hookunit.Unit{Name: name, Run: func(*hookunit.Ctx) hookunit.Result {
		return hookunit.Result{Name: name, Say: say, Stderr: say}
	}}
}

func responseOf(t *testing.T, stdout string) map[string]any {
	t.Helper()
	if strings.TrimSpace(stdout) == "" {
		return nil
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("the response is not one JSON object: %v (%q)", err, stdout)
	}
	return doc
}

// THE MESSAGE AND THE DECISION TRAVEL IN ONE OBJECT. Measured against the real client on
// 2026-09-16: both were honoured — the message displayed, the decision applied, the tool run.
func TestADecisionCarriesTheMessageWithoutLosingEither(t *testing.T) {
	stdout, stderr, code := fire(t, `{"tool_name":"Bash","tool_input":{"command":"git add -A"}}`,
		[]hookunit.Unit{
			stub("gate", `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"a secret"}}`, ""),
			saying("freeze", "sc-push-freeze-guard: a run is LIVE — a sweeping `git add`"),
		})
	if code != 0 {
		t.Fatalf("the block travels in the JSON, never the status: exit %d", code)
	}
	doc := responseOf(t, stdout)
	hso, ok := doc["hookSpecificOutput"].(map[string]any)
	if !ok || hso["permissionDecision"] != "deny" || hso["permissionDecisionReason"] != "a secret" {
		t.Fatalf("the decision did not survive the merge: %q", stdout)
	}
	if msg, _ := doc["systemMessage"].(string); !strings.Contains(msg, "a sweeping `git add`") {
		t.Errorf("the warning is not in the response: %q", stdout)
	}
	if !strings.Contains(stderr, "a sweeping `git add`") {
		t.Errorf("the debug log lost its copy: %q", stderr)
	}
}

// With no decision to carry it, the message is the whole response.
func TestAMessageWithNoDecisionIsTheWholeResponse(t *testing.T) {
	stdout, _, _ := fire(t, `{"tool_name":"Bash","tool_input":{"command":"git push"}}`,
		[]hookunit.Unit{saying("freeze", "sc-push-freeze-guard: pinned paths are FROZEN")})
	doc := responseOf(t, stdout)
	if len(doc) != 1 {
		t.Fatalf("the response carries more than the message: %q", stdout)
	}
	if msg, _ := doc["systemMessage"].(string); !strings.Contains(msg, "FROZEN") {
		t.Errorf("stdout = %q", stdout)
	}
}

// A DECISION IS NEVER RISKED FOR A MESSAGE. A unit's decision document that this binary cannot
// decode still goes out exactly as written — it is what protects the session — and the message
// follows rather than replacing it.
func TestAnUndecodableDecisionStillGoesOutIntact(t *testing.T) {
	stdout, _, _ := fire(t, `{"tool_name":"Bash","tool_input":{"command":"git push"}}`,
		[]hookunit.Unit{
			stub("gate", `{"permissionDecision": NOT JSON`, ""),
			saying("freeze", "sc-push-freeze-guard: a run is LIVE"),
		})
	if !strings.Contains(stdout, `{"permissionDecision": NOT JSON`) {
		t.Fatalf("the decision was dropped or rewritten: %q", stdout)
	}
}

// Silence stays silent: no message, no object, nothing for the client to parse.
func TestNothingToSaySaysNothing(t *testing.T) {
	stdout, _, code := fire(t, `{"tool_name":"Bash","tool_input":{"command":"ls"}}`,
		[]hookunit.Unit{stub("quiet", "", "")})
	if stdout != "" || code != 0 {
		t.Errorf("stdout = %q, exit %d", stdout, code)
	}
}

// A FAILURE RECORDED EARLIER, ON AN EVENT THAT COULD NOT SPEAK, IS ANNOUNCED HERE. That is what the
// record buys over saying things as they happen: five of this plugin's events display nothing, and
// their failures would otherwise wait for a session nobody starts.
func TestARecordedFailureIsAnnouncedOnThisEvent(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	// A seal hook on SessionEnd — an event the client displays nothing for — fails.
	silent := hookfailures.New("prosthetic-conscience", "sc-sessionend", "SessionEnd", noon, io.Discard)
	silent.Fail("snapshot-write", "cannot write snapshot: read-only file system")
	if msg := silent.Settle(); msg != "" {
		t.Fatalf("SessionEnd displays nothing and returned %q", msg)
	}

	stdout, _, _ := fire(t, `{"tool_name":"Bash","tool_input":{"command":"ls"}}`,
		[]hookunit.Unit{stub("quiet", "", "")})
	msg, _ := responseOf(t, stdout)["systemMessage"].(string)
	if !strings.Contains(msg, "snapshot-write") || !strings.Contains(msg, "read-only file system") {
		t.Fatalf("the next displaying event did not carry it: %q", stdout)
	}
}
