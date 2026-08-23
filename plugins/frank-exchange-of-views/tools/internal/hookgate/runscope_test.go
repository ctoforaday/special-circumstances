package hookgate

import (
	"path/filepath"
	"testing"
)

// THE LOCKDOWN IS SCOPED TO A LIVE RUN, and these are the cases that were wrong for the whole
// life of the gate rather than broken by a change.
//
// Every doc comment on this file said "a run's blue/report.md". The implementation matched a path
// SUFFIX and consulted no run: PreOutcome has carried a resolved runDir since the first commit in
// this file's history, and handed it only to the injection arm — the lockdown decision sat one
// line above and never took it. So the gate that was meant to be a no-op outside a run denied a
// write to any file named blue/report.md anywhere on disk, in any session.
//
// These are TESTS, not a gate: they pin the behaviour a caller depends on, per case, and they
// fail on the specific thing that regressed rather than reporting a count.

const scopeRun = "/runs/2026-08-22_smoke"

func write(agent, path string) Input {
	return Input{AgentType: agent, ToolName: "Write", ToolInput: []byte(`{"file_path":"` + path + `"}`)}
}

func bashCmd(agent, cmd string) Input {
	return Input{AgentType: agent, ToolName: "Bash", ToolInput: []byte(`{"command":"` + cmd + `"}`)}
}

const responderAgent = "frank-exchange-of-views:blue-researcher"

func TestNoLiveRunIsANoOp(t *testing.T) {
	// The case the whole gate was documented to skip and never did.
	for _, c := range []struct {
		name string
		in   Input
	}{
		{"a report-shaped path with no run at all", write(responderAgent, "/home/user/notes/blue/report.md")},
		{"a relative report path", write(responderAgent, "blue/report.md")},
		{"a Bash write into a report-shaped path", bashCmd(responderAgent, "cp d.md /home/user/notes/blue/report.md")},
		{"even the author's own write", write(AuthorAgentType, "/home/user/notes/blue/report.md")},
	} {
		t.Run(c.name, func(t *testing.T) {
			if deny, reason := PreDecision(c.in, ""); deny {
				t.Errorf("denied outside a live run (%q).\n\n"+
					"With no run resolvable there is no report to protect and every path is somebody "+
					"else's file. This gate is documented as a no-op here, and denying meant a plugin "+
					"installed for one repo policed filenames in every session.", reason)
			}
		})
	}
}

func TestOnlyPathsInsideTheRunAreTheGatesBusiness(t *testing.T) {
	inside := scopeRun + "/blue/report.md"
	for _, c := range []struct {
		name     string
		in       Input
		wantDeny bool
	}{
		{"non-author writes THIS run's report", write(responderAgent, inside), true},
		{"non-author Bash-writes THIS run's report", bashCmd(responderAgent, "cp d.md "+inside), true},
		{"the author writes THIS run's report", write(AuthorAgentType, inside), false},

		// A different run's report is a different run's business. The gate resolves ONE run from
		// the payload's cwd; a path outside it is not the report it is holding.
		{"a DIFFERENT run's report", write(responderAgent, "/runs/2026-01-01_other/blue/report.md"), false},
		{"a finished run under research/", write(responderAgent, "/repo/research/2026-08-02_old/blue/report.md"), false},

		// Inside the run, but not the report.
		{"blue's draft inside the run", write(responderAgent, scopeRun+"/blue/draft.md"), false},
		{"red's report inside the run", write(responderAgent, scopeRun+"/red/report.md"), false},
		{"the run's top-level report.md", write(responderAgent, scopeRun+"/report.md"), false},

		// Reads are never writes, inside a run or out.
		{"reading the report inside the run", bashCmd(responderAgent, "grep x "+inside), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			deny, reason := PreDecision(c.in, scopeRun)
			if deny != c.wantDeny {
				t.Errorf("deny=%v, want %v (reason %q)", deny, c.wantDeny, reason)
			}
		})
	}
}

// A PATH THAT CLIMBS OUT OF THE RUN IS OUTSIDE IT, however it is spelled. filepath.Rel answers
// this, and the ".." prefix is the whole test — a containment check written with strings.HasPrefix
// would accept `/runs/2026-08-22_smoke/../elsewhere/blue/report.md` and, worse, would accept
// `/runs/2026-08-22_smoke-other/blue/report.md` because the run's own name is a string prefix of it.
func TestContainmentIsNotAStringPrefix(t *testing.T) {
	for _, p := range []string{
		scopeRun + "/../elsewhere/blue/report.md",
		scopeRun + "-other/blue/report.md",
		scopeRun + "2/blue/report.md",
	} {
		t.Run(p, func(t *testing.T) {
			if insideRun(p, scopeRun) {
				t.Errorf("%s reported as inside %s — a sibling directory sharing a name prefix is not "+
					"inside the run, and a path climbing out of it has left", p, scopeRun)
			}
		})
	}
	// ...and the genuine cases still resolve.
	for _, p := range []string{
		scopeRun + "/blue/report.md",
		scopeRun + "/./blue/report.md",
		filepath.Join(scopeRun, "blue", "report.md"),
	} {
		if !insideRun(p, scopeRun) {
			t.Errorf("%s reported as OUTSIDE %s, so the lockdown would not hold on its own report", p, scopeRun)
		}
	}
}

// THE FAIL-CLOSED ARM IS SCOPED TOO, and that is the second half of the same defect.
//
// Unreadable tool_input denies, deliberately and before the allowlist, because a payload that did
// not parse cannot prove who is calling. That trade is sound INSIDE a run. Outside one it meant an
// unparseable payload denied a write to any file at all, in any session — the gate refusing calls
// it had already established were none of its business.
func TestUnreadableInputDeniesOnlyInsideARun(t *testing.T) {
	bad := Input{ToolName: "Write", ToolInput: []byte(`{"file_path": `)} // truncated mid-value
	if !InputUnreadable(bad) {
		t.Fatal("malformed tool_input reported as readable — the bypass would stay invisible")
	}
	if deny, _ := PreDecision(bad, scopeRun); !deny {
		t.Error("unreadable tool_input was ALLOWED inside a live run — a gate bypassed by input it " +
			"cannot read is not a gate")
	}
	if deny, reason := PreDecision(bad, ""); deny {
		t.Errorf("unreadable tool_input was DENIED with no live run (%q).\n\n"+
			"There is no report to protect here, so there is nothing to fail closed about. This is "+
			"how the lockdown came to refuse writes to arbitrary files in sessions that had nothing "+
			"to do with a debate.", reason)
	}
}
