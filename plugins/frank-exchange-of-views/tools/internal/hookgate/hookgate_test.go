package hookgate

import (
	"encoding/json"
	"strings"
	"testing"
)

func mkInput(t *testing.T, agentType, tool string, ti map[string]string) Input {
	t.Helper()
	raw, err := json.Marshal(ti)
	if err != nil {
		t.Fatal(err)
	}
	return Input{AgentType: agentType, ToolName: tool, ToolInput: raw}
}

const responder = "frank-exchange-of-views:blue-researcher"

// §V.3 — the PreToolUse allowlist: only the author may raw-write blue/report.md; every
// other seat is denied; reads and non-report targets get no opinion.
func TestPreDecisionAllowlist(t *testing.T) {
	cases := []struct {
		name, agent, tool string
		ti                map[string]string
		wantDeny          bool
	}{
		{"author Write report → allow", AuthorAgentType, "Write", map[string]string{"file_path": "/r/x/blue/report.md"}, false},
		{"author cp report → allow", AuthorAgentType, "Bash", map[string]string{"command": "cp draft.md /r/x/blue/report.md"}, false},
		{"responder Write report → DENY", responder, "Write", map[string]string{"file_path": "/r/x/blue/report.md"}, true},
		{"responder Edit report → DENY", responder, "Edit", map[string]string{"file_path": "/r/x/blue/report.md"}, true},
		{"responder MultiEdit report → DENY", responder, "MultiEdit", map[string]string{"file_path": "/r/x/blue/report.md"}, true},
		{"responder cp → report → DENY", responder, "Bash", map[string]string{"command": "cp /tmp/d.md /r/x/blue/report.md"}, true},
		{"responder redirect → report → DENY", responder, "Bash", map[string]string{"command": "echo hi >> /r/x/blue/report.md"}, true},
		{"responder tee report → DENY", responder, "Bash", map[string]string{"command": "printf x | tee /r/x/blue/report.md"}, true},
		{"responder sed -i report → DENY", responder, "Bash", map[string]string{"command": "sed -i 's/a/b/' /r/x/blue/report.md"}, true},
		{"responder mv report → DENY", responder, "Bash", map[string]string{"command": "mv /tmp/n.md /r/x/blue/report.md"}, true},
		{"red Edit report → DENY", "frank-exchange-of-views:red-auditor", "Edit", map[string]string{"file_path": "/r/x/blue/report.md"}, true},
		// Reads and non-targets → NO opinion (allow).
		{"responder grep report (read) → allow", responder, "Bash", map[string]string{"command": "grep '<!--fx:' /r/x/blue/report.md"}, false},
		{"responder read-redirect report → allow", responder, "Bash", map[string]string{"command": "cat < /r/x/blue/report.md"}, false},
		{"responder count-claims (no path) → allow", responder, "Bash", map[string]string{"command": "feov-record count-claims --run /r/x"}, false},
		{"responder Write TOP-LEVEL report → allow", responder, "Write", map[string]string{"file_path": "/r/x/report.md"}, false},
		{"responder Edit other file → allow", responder, "Edit", map[string]string{"file_path": "/r/x/blue/CHANGELOG.md"}, false},
		{"responder cp elsewhere → allow", responder, "Bash", map[string]string{"command": "cp a.md b.md"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			deny, reason := PreDecision(mkInput(t, c.agent, c.tool, c.ti))
			if deny != c.wantDeny {
				t.Errorf("deny=%v, want %v (reason %q)", deny, c.wantDeny, reason)
			}
			if deny && reason == "" {
				t.Error("a deny must carry a reason")
			}
		})
	}
}

// §V.4 — the PostToolUse backstop: a non-author call that dropped a marker returns the
// missing ids; a clean call or the author returns nil.
func TestPostDropped(t *testing.T) {
	anchors := func(runDir string) ([]string, error) { return []string{"f-a", "f-b"}, nil }
	// Report readers standing in for two states.
	both := func(path string) (string, error) { return "x <!--fx:f-a--> y <!--fx:f-b--> z", nil }
	dropped := func(path string) (string, error) { return "x <!--fx:f-a--> y z", nil } // f-b gone

	edit := mkInput(t, responder, "Edit", map[string]string{"file_path": "/r/x/blue/report.md"})

	if got, _ := PostDropped(edit, anchors, both); len(got) != 0 {
		t.Errorf("no drop → %v, want empty", got)
	}
	if got, _ := PostDropped(edit, anchors, dropped); len(got) != 1 || got[0] != "f-b" {
		t.Errorf("dropped f-b → %v, want [f-b]", got)
	}
	// The author is never checked.
	auth := mkInput(t, AuthorAgentType, "Edit", map[string]string{"file_path": "/r/x/blue/report.md"})
	if got, _ := PostDropped(auth, anchors, dropped); got != nil {
		t.Errorf("author is exempt, got %v", got)
	}
	// A call that never touched report.md → nothing to check.
	other := mkInput(t, responder, "Edit", map[string]string{"file_path": "/r/x/blue/CHANGELOG.md"})
	if got, _ := PostDropped(other, anchors, dropped); got != nil {
		t.Errorf("non-report call → %v, want nil", got)
	}
}

// §V.5 — the backstop sweeps BOTH anchor classes: a dropped CITATION anchor is caught the
// same way a dropped finding marker is (the EXPECTED set is the union of both).
func TestPostDroppedCatchesCitationAndBothClasses(t *testing.T) {
	// EXPECTED union: a finding and a citation.
	anchors := func(runDir string) ([]string, error) { return []string{"f-a", "c-1"}, nil }
	intact := func(path string) (string, error) { return "x <!--fx:f-a--> y <!--cite:c-1--> z", nil }
	citeGone := func(path string) (string, error) { return "x <!--fx:f-a--> y z", nil }   // c-1 dropped
	findGone := func(path string) (string, error) { return "x y <!--cite:c-1--> z", nil } // f-a dropped

	edit := mkInput(t, responder, "Edit", map[string]string{"file_path": "/r/x/blue/report.md"})

	if got, _ := PostDropped(edit, anchors, intact); len(got) != 0 {
		t.Errorf("both anchors present → %v, want empty", got)
	}
	if got, _ := PostDropped(edit, anchors, citeGone); len(got) != 1 || got[0] != "c-1" {
		t.Errorf("dropped citation → %v, want [c-1]", got)
	}
	if got, _ := PostDropped(edit, anchors, findGone); len(got) != 1 || got[0] != "f-a" {
		t.Errorf("dropped finding → %v, want [f-a]", got)
	}
}

// A GATE THAT CANNOT READ ITS INPUT MUST LEAVE A TRACE.
//
// `_ = json.Unmarshal(in.ToolInput, &ti)` made malformed input bypass the blue-report lockdown
// in silence: FilePath empty -> writesBlueReport false -> Decide "no opinion" -> the write is
// allowed. The fail direction is deliberately unchanged; what is asserted here is that the
// condition is now DETECTABLE, so internal/cli/hook.go can record it as a friction event of kind
// FRICTION_KIND_TOOL_ERROR instead of nothing happening.
func TestUnreadableToolInputFailsClosed(t *testing.T) {
	bad := Input{ToolName: "Write", ToolInput: []byte(`{"file_path": `)} // truncated mid-value

	if !InputUnreadable(bad) {
		t.Fatal("malformed tool_input reported as readable — the bypass would stay invisible")
	}

	// THE GATE DENIES. Before this it returned false here (FilePath empty reads as "not the
	// report"), so unreadable input walked straight through the blue-report lockdown.
	deny, reason := PreDecision(bad)
	if !deny {
		t.Error("unreadable tool_input was ALLOWED — a gate bypassed by input it cannot read " +
			"is not a gate")
	}

	// The refusal reaches the model through permissionDecisionReason, so it must say the seat did
	// nothing wrong and what to do — a bare denial on input the seat never controlled is unactionable.
	for _, want := range []string{"could not parse tool_input", "Nothing about your call is known to be wrong", "Retry"} {
		if !strings.Contains(reason, want) {
			t.Errorf("deny reason missing %q — a denial the seat cannot act on is a mystery:\n%s", want, reason)
		}
	}

	// AND THE AUTHOR IS NOT EXEMPT, which is the cost this decision accepts. The allowlist is
	// checked AFTER readability, because an unreadable payload cannot prove who is calling.
	author := Input{AgentType: AuthorAgentType, ToolName: "Write", ToolInput: []byte(`{"file_path": `)}
	if deny, _ := PreDecision(author); !deny {
		t.Error("the author was exempted from the unreadable-input denial — the exemption is " +
			"read from a payload that did not parse")
	}
}

func TestReadableInputIsUnaffected(t *testing.T) {
	// The ordinary paths must not move: this change is about input the gate cannot read, and
	// nothing else.
	report := Input{ToolName: "Write", ToolInput: []byte(`{"file_path":"/x/blue/report.md"}`)}
	if InputUnreadable(report) {
		t.Fatal("well-formed input reported unreadable — this would deny every call")
	}
	if deny, _ := PreDecision(report); !deny {
		t.Error("a non-author write to blue/report.md must still be denied")
	}
	authored := Input{AgentType: AuthorAgentType, ToolName: "Write", ToolInput: []byte(`{"file_path":"/x/blue/report.md"}`)}
	if deny, _ := PreDecision(authored); deny {
		t.Error("the allowlisted author must still be permitted")
	}
	other := Input{ToolName: "Write", ToolInput: []byte(`{"file_path":"/x/notes.md"}`)}
	if deny, _ := PreDecision(other); deny {
		t.Error("a write to another file must still draw no opinion")
	}
}
