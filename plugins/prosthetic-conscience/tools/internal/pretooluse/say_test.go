package pretooluse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
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

// A DECISION THIS HOOK CANNOT READ IS A DENY, ON BOTH HARNESSES, AND A RECORDED FAILURE. Written
// to stdout as-is it would be a decision for neither: Claude Code runs a call whose stdout is not
// its JSON, and Antigravity fails the step with "failed to unmarshal result from hook" (#1927). The
// one unit that decides is the secrets gate, which fails closed, so this does too — and the defect
// is announced, never absorbed.
func TestAnUnreadableDecisionIsADenyOnBothHarnesses(t *testing.T) {
	unreadable := map[string]string{
		"not JSON":           `{"permissionDecision": NOT JSON`,
		"no decision field":  `{"hookSpecificOutput":{"hookEventName":"PreToolUse"}}`,
		"top-level decision": `{"permissionDecision":"deny"}`,
		"unknown decision":   `{"hookSpecificOutput":{"permissionDecision":"block"}}`,
	}
	harnesses := map[string]string{
		"claude":      `{"tool_name":"Bash","tool_input":{"command":"git push"}}`,
		"antigravity": `{"conversationId":"c1","toolCall":{"name":"run_command","args":{"CommandLine":"git push"}}}`,
	}
	for name, doc := range unreadable {
		for harness, payload := range harnesses {
			t.Run(name+"/"+harness, func(t *testing.T) {
				t.Setenv("XDG_STATE_HOME", t.TempDir())
				stdout, _, code := fire(t, payload, []hookunit.Unit{stub("gate", doc, "")})
				if code != 0 {
					t.Fatalf("the block travels in the JSON, never the status: exit %d", code)
				}
				out := responseOf(t, stdout)
				switch harness {
				case "claude":
					hso, _ := out["hookSpecificOutput"].(map[string]any)
					if hso["permissionDecision"] != "deny" || !strings.Contains(fmt.Sprint(hso["permissionDecisionReason"]), "cannot read") {
						t.Errorf("not a deny naming the unreadable decision: %q", stdout)
					}
					if !strings.Contains(fmt.Sprint(out["systemMessage"]), "could not be read") {
						t.Errorf("the failure was not announced: %q", stdout)
					}
				case "antigravity":
					if out["decision"] != "deny" || !strings.Contains(fmt.Sprint(out["deny_reason"]), "cannot read") {
						t.Errorf("not an Antigravity deny naming the unreadable decision: %q", stdout)
					}
				}
			})
		}
	}
}

// A call with no unreadable decision clears the failure an unreadable one recorded, so the record
// says what is true now and stops announcing a defect that is gone.
func TestACallWithoutAnUnreadableDecisionClearsTheFailure(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	payload := `{"tool_name":"Bash","tool_input":{"command":"git push"}}`
	recorded := func() bool {
		t.Helper()
		path, err := hookfailures.Path("prosthetic-conscience")
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			return false // the record removes itself when its last failure clears
		}
		if err != nil {
			t.Fatalf("read the failure record: %v", err)
		}
		var r hookfailures.Record
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatalf("the failure record is not JSON: %v", err)
		}
		for _, f := range r.Failures {
			if f.Stage == StageDecision {
				return true
			}
		}
		return false
	}
	fire(t, payload, []hookunit.Unit{stub("gate", `not json`, "")})
	if !recorded() {
		t.Fatal("an unreadable decision was not recorded as a failure")
	}
	fire(t, payload, []hookunit.Unit{stub("quiet", "", "")})
	if recorded() {
		t.Error("a call with no unreadable decision left the failure recorded")
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

// NO PROJECT ROOT ON THE EVENT THAT FIRES ON EVERY TOOL CALL. PreToolUse never called Explain, so an
// unresolved root was recorded by every hook except this one and PostToolUse. It is recorded now —
// and the secrets gate still runs, because recording decides nothing.
func TestAnUnresolvedRootIsRecordedWithoutStoppingTheGate(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	ran := false
	gate := hookunit.Unit{Name: "gate", Run: func(*hookunit.Ctx) hookunit.Result {
		ran = true
		return hookunit.Result{Name: "gate"}
	}}
	var o, e bytes.Buffer
	run(nil, strings.NewReader(`{"tool_name":"Bash","tool_input":{"command":"ls"}}`), &o, &e, "", noon, []hookunit.Unit{gate})
	if !ran {
		t.Fatal("the gate did not run with no project root — recording the root must decide nothing")
	}
	if msg, _ := responseOf(t, o.String())["systemMessage"].(string); !strings.Contains(msg, "- project-root:") {
		t.Fatalf("an unresolved root was not said: %q", o.String())
	}
}
