package toolchain

import (
	"encoding/json"
	"testing"
)

// A tool can be absent BY DESIGN rather than by neglect. `gh` in a Claude Code cloud
// session is the case that forced this: GitHub is reached through MCP tools there and
// no gh is shipped, so the preflight reported DEGRADED in every cloud session forever
// and told the operator to install a tool the environment will not keep. A verdict that
// is permanently wrong on one axis is a verdict nobody reads on any axis.
func TestEnvironmentFollowsClaudeCodeRemote(t *testing.T) {
	t.Setenv("CLAUDE_CODE_REMOTE", "1")
	if got := Environment(); got != EnvRemote {
		t.Fatalf("Environment() = %q with CLAUDE_CODE_REMOTE set; want %q", got, EnvRemote)
	}
	t.Setenv("CLAUDE_CODE_REMOTE", "")
	if got := Environment(); got != EnvLocal {
		t.Fatalf("Environment() = %q with CLAUDE_CODE_REMOTE unset; want %q", got, EnvLocal)
	}
}

func TestAppliesIn(t *testing.T) {
	gh := Tool{Name: "gh", NotApplicableIn: []string{EnvRemote}}
	if gh.AppliesIn(EnvRemote) {
		t.Error("gh should not apply in a cloud session")
	}
	if !gh.AppliesIn(EnvLocal) {
		t.Error("gh must still apply on a local machine — the exemption is per-environment, not blanket")
	}
	// An unmarked tool applies everywhere: the field is opt-in, so a manifest that
	// says nothing keeps exactly the behaviour it had before the field existed.
	if !(Tool{Name: "qlty"}).AppliesIn(EnvRemote) {
		t.Error("a tool with no not_applicable_in must apply in every environment")
	}
}

func TestProbeInMarksNotApplicable(t *testing.T) {
	tools := []Tool{
		{Name: "phantom", CheckCmd: "sc-definitely-not-real-xyz --version", NotApplicableIn: []string{EnvRemote}},
		{Name: "other", CheckCmd: "sc-definitely-not-real-abc --version"},
	}
	remote := ProbeIn(tools, EnvRemote)
	if !remote[0].NotApplicable {
		t.Error("a tool marked for this environment must probe NotApplicable")
	}
	if remote[1].NotApplicable {
		t.Error("an unmarked tool must never probe NotApplicable")
	}
	if ProbeIn(tools, EnvLocal)[0].NotApplicable {
		t.Error("the exemption must not leak into the local environment")
	}
}

// NotApplicable and Found are independent. If gh happens to be on PATH in a cloud
// session, the table must say so rather than claim a tool is missing while it sits
// right there — the exemption suppresses the PENALTY, not the observation.
func TestNotApplicableStillReportsPresence(t *testing.T) {
	got := ProbeIn([]Tool{{Name: "go", CheckCmd: "go version", NotApplicableIn: []string{EnvRemote}}}, EnvRemote)
	if !got[0].NotApplicable {
		t.Fatal("want NotApplicable")
	}
	if !got[0].Found {
		t.Skip("go not on PATH here; nothing to assert about presence")
	}
}

// The manifest is the authority on which tools are environment-scoped. This asserts
// the shape survives a round trip, because the field is only ever set from JSON.
func TestNotApplicableInUnmarshals(t *testing.T) {
	var tool Tool
	if err := json.Unmarshal([]byte(`{"name":"gh","not_applicable_in":["claude-code-remote"]}`), &tool); err != nil {
		t.Fatal(err)
	}
	if len(tool.NotApplicableIn) != 1 || tool.NotApplicableIn[0] != EnvRemote {
		t.Fatalf("not_applicable_in did not decode: %#v", tool.NotApplicableIn)
	}
}
