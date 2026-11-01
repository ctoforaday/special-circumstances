package sittinghook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// membershipRun is a live run led by testLead with the REAL writer behind the handoff, so what a
// payload does is read off the record rather than off the spawn seam.
func membershipRun(t *testing.T) (cwd string, run record.Run) {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	got := capture(t)
	cwd, runDir := liveRun(t)
	t.Cleanup(func() { _ = recordsql.CloseUnder(runDir) })
	spawn = func(_, dir, phase, agentID, agentType, transcript string) ([]byte, error) {
		*got = append(*got, handoffArgs{"", dir, phase, agentID, agentType, transcript})
		var forSeat bytes.Buffer
		err := sittingwrite.Write(sittingwrite.Sitting{
			RunDir: dir, Phase: sittingwrite.Phase(phase),
			AgentID: agentID, AgentType: agentType, TranscriptPath: transcript,
		}, &forSeat)
		return forSeat.Bytes(), err
	}
	recordtest.Seed(t, runDir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{SeatIds: []string{"red-chair", "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest", &recordpb.ReportPart{Text: proto.String("# report\n\nOne paragraph.\n")}))
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	return cwd, run
}

// sessionPayload is a SubagentStart/SubagentStop payload as the client sends it, session included
// (omitted when session is "").
func sessionPayload(t *testing.T, session, agentID, agentType, cwd string) *strings.Reader {
	t.Helper()
	p := map[string]string{"agent_id": agentID, "agent_type": agentType, "cwd": cwd}
	if session != "" {
		p["session_id"] = session
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReader(string(b))
}

// sitThrough fires SubagentStart then SubagentStop for one agent, through the hook entry points, settling
// each event's failures as the binaries do.
func sitThrough(t *testing.T, session, agentID, agentType, cwd string) {
	t.Helper()
	settle(t, "SubagentStart", func(rec *hookfailures.Recorder) {
		if err := Start(sessionPayload(t, session, agentID, agentType, cwd), &bytes.Buffer{}, rec); err != nil {
			t.Fatal(err)
		}
	})
	settle(t, "SubagentStop", func(rec *hookfailures.Recorder) {
		if err := Stop(sessionPayload(t, session, agentID, agentType, cwd), &bytes.Buffer{}, rec); err != nil {
			t.Fatal(err)
		}
	})
}

// noSittingOf asserts the record holds no sitting span naming agent.
func noSittingOf(t *testing.T, run record.Run, agent string) {
	t.Helper()
	if opened, closed := spans(t, run, agent); opened != 0 || closed != 0 {
		t.Errorf("agent %s: %d sitting open(s) and %d close(s) on the run's record, want none", agent, opened, closed)
	}
}

// (a) THE AUDIT'S DEMONSTRATION. Another conversation's Explore agent, in a worktree nested under the
// run's checkout, finds the run's marker by walking up — and is no member of the run: nothing of it
// reaches the record, and an ordinary agent of an unrelated session is said nowhere.
func TestAnotherSessionsAgentInANestedWorktreeIsNotAttributed(t *testing.T) {
	cwd, run := membershipRun(t)
	nested := filepath.Join(cwd, ".claude", "worktrees", "dev", "sub")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	sitThrough(t, "a-dev-session", "aexp10re000000001", "Explore", nested)
	noSittingOf(t, run, "aexp10re000000001")
	if st := sittingStages(t); st[string(runlive.StageMembership)+"@"+cwd] {
		t.Errorf("an unrelated session's ordinary agent was recorded as a membership fault: %v", st)
	}
}

// (b) A SEAT OF THE RUN'S OWN SESSION is attributed as before, from the nested directory too.
func TestASeatOfTheLeadSessionIsAttributed(t *testing.T) {
	cwd, run := membershipRun(t)
	sitThrough(t, testLead, "ase4t000000000001", "frank-exchange-of-views:red-chair", cwd)
	if opened, closed := spans(t, run, "ase4t000000000001"); opened != 1 || closed != 1 {
		t.Fatalf("the lead session's seat: %d open(s), %d close(s), want one of each", opened, closed)
	}
}

// (c) A RESUME IN A NEW SESSION. Before the take-over, the new session's seat is attributed to
// nothing and SAID — it is one of this plugin's agents, the shape of a resume nobody took over.
// After the take-over it is attributed, and the old lead's agents no longer are.
func TestAResumeInANewSessionIsAttributedOnlyAfterTheTakeOver(t *testing.T) {
	cwd, run := membershipRun(t)
	const resumer = "a-resuming-session"
	sitThrough(t, resumer, "abefore0000000001", "frank-exchange-of-views:red-chair", cwd)
	noSittingOf(t, run, "abefore0000000001")
	if !sittingStages(t)[string(runlive.StageMembership)+"@"+cwd] {
		t.Fatalf("a seat of a session leading no run was not recorded: %v", sittingStages(t))
	}
	if _, err := runlive.TakeOver(cwd, run.Dir(), "not-the-lead", resumer); err == nil {
		t.Fatal("a take-over naming the wrong lead was accepted")
	}
	if _, err := runlive.TakeOver(cwd, run.Dir(), testLead, resumer); err != nil {
		t.Fatal(err)
	}
	sitThrough(t, resumer, "aafter00000000001", "frank-exchange-of-views:red-chair", cwd)
	if opened, closed := spans(t, run, "aafter00000000001"); opened != 1 || closed != 1 {
		t.Fatalf("the resumed seat after the take-over: %d open(s), %d close(s), want one of each", opened, closed)
	}
	if sittingStages(t)[string(runlive.StageMembership)+"@"+cwd] {
		t.Error("an attributed sitting did not clear the membership entry")
	}
	sitThrough(t, testLead, "aold1ead000000001", "frank-exchange-of-views:red-chair", cwd)
	noSittingOf(t, run, "aold1ead000000001")
}

// (d) THE BENCH'S POST-OUTCOME EXCEPTION IS THE LEAD SESSION'S. Between the outcome and the capture a
// fresh lead-judge agent is bound to the run for its assembly sitting — only when its session leads
// the run.
func TestThePostOutcomeBenchExceptionRequiresTheSession(t *testing.T) {
	cwd, run := membershipRun(t)
	if err := runlive.WriteOutcomeMark(run.Dir(), "aorig000000000001"); err != nil {
		t.Fatal(err)
	}
	sitThrough(t, "a-dev-session", "aother00000000001", runlive.BenchAgentType, cwd)
	noSittingOf(t, run, "aother00000000001")
	sitThrough(t, testLead, "afresh00000000001", runlive.BenchAgentType, cwd)
	if opened, closed := spans(t, run, "afresh00000000001"); opened != 1 || closed != 1 {
		t.Fatalf("the lead session's fresh bench after the outcome: %d open(s), %d close(s), want one of each", opened, closed)
	}
}

// (e) A PAYLOAD WITH NO session_id is attributed to nothing, and recorded loudly: a run's members
// cannot be told without it, and silence would read as "not a member".
func TestAPayloadWithNoSessionIsNotAttributedAndIsRecorded(t *testing.T) {
	cwd, run := membershipRun(t)
	sitThrough(t, "", "anosess1on0000001", "frank-exchange-of-views:red-chair", cwd)
	noSittingOf(t, run, "anosess1on0000001")
	if !sittingStages(t)[string(runlive.StageMembership)+"@"+cwd] {
		t.Fatalf("a payload with no session_id was not recorded: %v", sittingStages(t))
	}
}
