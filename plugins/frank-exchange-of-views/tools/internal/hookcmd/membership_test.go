package hookcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// membershipProject is a project with one live run led by testLead, and a worktree nested under it
// the way .claude/worktrees/<name>/ is.
func membershipProject(t *testing.T) (proj, runDir, nested string) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	proj = t.TempDir()
	runDir = filepath.Join(proj, "research", "r")
	nested = filepath.Join(proj, ".claude", "worktrees", "dev", "sub")
	for _, d := range []string{runDir, nested} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runlive.WriteRunLiveMarker(proj, "research/r", nil, time.Now(), "", "", testLead); err != nil {
		t.Fatal(err)
	}
	return proj, runDir, nested
}

// bash fires PreToolUse for a Bash call through the hook binary's entry point and returns the command
// the call will run — the original when the hook rewrote nothing.
func bash(t *testing.T, session, agentID, agentType, cwd string) string {
	t.Helper()
	p := map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": "feov-record ops log"}, "cwd": cwd}
	for k, v := range map[string]string{"session_id": session, "agent_id": agentID, "agent_type": agentType} {
		if v != "" {
			p[k] = v
		}
	}
	b, _ := json.Marshal(p)
	var out bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", Pre, bytes.NewReader(b), &out)
	var got struct {
		HookSpecificOutput struct {
			UpdatedInput struct {
				Command string `json:"command"`
			} `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	_ = json.Unmarshal(out.Bytes(), &got)
	if got.HookSpecificOutput.UpdatedInput.Command == "" {
		return "feov-record ops log"
	}
	return got.HookSpecificOutput.UpdatedInput.Command
}

func injectsRun(command string) bool { return strings.Contains(command, seatenv.Var+"=") }

// (a) THE AUDIT'S DEMONSTRATION, ON PRETOOLUSE. A dev session's Bash — its main thread and its Explore
// agent — in a worktree nested under the run's checkout gets no FEOV_RUN, so a verb there without
// --run writes into no run; and an ordinary unrelated session is said nowhere.
func TestAnotherSessionsBashGetsNoRunInjected(t *testing.T) {
	proj, _, nested := membershipProject(t)
	for _, c := range []struct{ agentID, agentType string }{{"", ""}, {"aexp10re000000001", "Explore"}} {
		if cmd := bash(t, "a-dev-session", c.agentID, c.agentType, nested); injectsRun(cmd) {
			t.Errorf("agent %q of another session was injected the run: %s", c.agentType, cmd)
		}
	}
	if onRecord(t, string(runlive.StageMembership), proj) {
		t.Error("an unrelated session's ordinary calls were recorded as a membership fault")
	}
}

// (b) THE RUN'S OWN SESSION — its lead's Bash and its seats' — is injected as before.
func TestTheLeadSessionsBashIsInjected(t *testing.T) {
	_, runDir, nested := membershipProject(t)
	for _, c := range []struct{ agentID, agentType, cwd string }{
		{"", "", nested},
		{"ase4t000000000001", "frank-exchange-of-views:red-chair", nested},
	} {
		if cmd := bash(t, testLead, c.agentID, c.agentType, c.cwd); !strings.Contains(cmd, seatenv.Var+"='"+runDir+"'") {
			t.Errorf("the lead session's call (agent %q) was not injected the run: %s", c.agentID, cmd)
		}
	}
}

// (c) A RESUME IN A NEW SESSION is injected only after it takes the run over, and is said before.
func TestANewSessionIsInjectedOnlyAfterTheTakeOver(t *testing.T) {
	proj, runDir, _ := membershipProject(t)
	const resumer = "a-resuming-session"
	if cmd := bash(t, resumer, "ase4t000000000001", "frank-exchange-of-views:red-chair", proj); injectsRun(cmd) {
		t.Fatalf("a seat of a session leading no run was injected: %s", cmd)
	}
	if !onRecord(t, string(runlive.StageMembership), proj) {
		t.Fatal("a seat of a session leading no run was not recorded")
	}
	// An unrelated session's ordinary call is no evidence the run's members are found again.
	bash(t, "a-dev-session", "", "", proj)
	if !onRecord(t, string(runlive.StageMembership), proj) {
		t.Fatal("an unrelated session's call cleared the membership entry")
	}
	if _, err := runlive.TakeOver(proj, "research/r", testLead, resumer); err != nil {
		t.Fatal(err)
	}
	if cmd := bash(t, resumer, "ase4t000000000001", "frank-exchange-of-views:red-chair", proj); !strings.Contains(cmd, seatenv.Var+"='"+runDir+"'") {
		t.Fatalf("the resuming session's seat after the take-over was not injected: %s", cmd)
	}
	if onRecord(t, string(runlive.StageMembership), proj) {
		t.Error("an injected call did not clear the membership entry")
	}
	if cmd := bash(t, testLead, "", "", proj); injectsRun(cmd) {
		t.Errorf("the old lead's Bash was injected after the take-over: %s", cmd)
	}
}

// (d) AFTER THE OUTCOME, a bench agent is bound for its assembly sitting only in the lead session.
func TestThePostOutcomeBenchInjectionRequiresTheSession(t *testing.T) {
	proj, runDir, _ := membershipProject(t)
	if err := runlive.WriteOutcomeMark(runDir, "aorig000000000001"); err != nil {
		t.Fatal(err)
	}
	if cmd := bash(t, "a-dev-session", "afresh00000000001", runlive.BenchAgentType, proj); injectsRun(cmd) {
		t.Errorf("another session's bench was injected the decided run: %s", cmd)
	}
	if cmd := bash(t, "a-dev-session", "aorig000000000001", runlive.BenchAgentType, proj); injectsRun(cmd) {
		t.Errorf("another session's agent with the recording agent's id was injected the decided run: %s", cmd)
	}
	if cmd := bash(t, testLead, "afresh00000000001", runlive.BenchAgentType, proj); !strings.Contains(cmd, seatenv.Var+"='"+runDir+"'") {
		t.Errorf("the lead session's bench after the outcome was not injected: %s", cmd)
	}
}

// (e) A PAYLOAD WITH NO session_id is injected nothing, and the hook says so.
func TestAPayloadWithNoSessionIsInjectedNothingAndSaid(t *testing.T) {
	proj, _, _ := membershipProject(t)
	if cmd := bash(t, "", "ase4t000000000001", "frank-exchange-of-views:red-chair", proj); injectsRun(cmd) {
		t.Fatalf("a payload with no session_id was injected the run: %s", cmd)
	}
	if !onRecord(t, string(runlive.StageMembership), proj) {
		t.Fatal("a payload with no session_id was not recorded")
	}
}

// POSTTOOLUSEFAILURE logs a failed call into a run only for the run's own session.
func TestTheFailureHookLogsOnlyTheLeadSessionsSeats(t *testing.T) {
	_, runDir, nested := membershipProject(t)
	var logged []string
	prev := recordFailure
	recordFailure = func(dir, agent, text string) error {
		if dir != runDir {
			t.Errorf("logged into %q, want %q", dir, runDir)
		}
		logged = append(logged, agent)
		return nil
	}
	t.Cleanup(func() { recordFailure = prev })
	fire := func(session, agent string) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"session_id": session, "agent_id": agent, "agent_type": "Explore",
			"tool_name": "Read", "tool_input": map[string]string{"file_path": "/x"}, "error": "File does not exist.", "cwd": nested})
		Run("feov-posttoolusefailure", "PostToolUseFailure", PostFailure, bytes.NewReader(b), &bytes.Buffer{})
	}
	fire("a-dev-session", "aexp10re000000001")
	fire(testLead, "ase4t000000000001")
	if len(logged) != 1 || logged[0] != "ase4t000000000001" {
		t.Fatalf("logged %q, want only the lead session's seat", logged)
	}
}
