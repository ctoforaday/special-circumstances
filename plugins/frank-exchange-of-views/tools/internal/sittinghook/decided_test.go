package sittinghook

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// A RESUMED ASSEMBLY SITTING'S SPAN IS ON THE RECORD. The assembly sitting stamps the outcome first;
// one that stops after the stamp is dispatched again on a resume, as a fresh agent the outcome mark
// does not name. Driven through Start and Stop as the hook binaries run them, with the real writer
// behind the handoff: the fresh bench agent's sitting opens and closes on the record, and its
// assemble-occasion register lands in it; a lens dispatched after the outcome, and any agent once
// the run is captured, hand nothing to the writer.
func TestAResumedAssemblySittingIsRecordedAfterTheOutcome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the identity comes from the PreToolUse rewrite run in a POSIX shell; the binding it tests is platform-independent")
	}
	got := capture(t) // places the writer beside the test binary, which the handoff checks for
	cwd, runDir := liveRun(t)
	t.Cleanup(func() { _ = recordsql.CloseUnder(runDir) })
	spawn = func(_, run, phase, agentID, agentType, transcript string) ([]byte, error) {
		*got = append(*got, handoffArgs{"", run, phase, agentID, agentType, transcript})
		var forSeat bytes.Buffer
		err := sittingwrite.Write(sittingwrite.Sitting{
			RunDir: run, Phase: sittingwrite.Phase(phase),
			AgentID: agentID, AgentType: agentType, TranscriptPath: transcript,
		}, &forSeat)
		if err != nil {
			t.Errorf("the writer refused the %s the hook handed it for %s: %v", phase, agentID, err)
		}
		return forSeat.Bytes(), err
	}
	recordtest.Seed(t, runDir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{SeatIds: []string{"red-chair", "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest", &recordpb.ReportPart{Text: proto.String("# report\n\nOne paragraph.\n")}))
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	const (
		orig  = "a0r1g1na1000000001"
		fresh = "af7e5h0000000000a"
		lens  = "a1e75000000000000"
	)
	start := func(agent, agentType string) {
		t.Helper()
		if err := Start(payload(t, agent, agentType, cwd), &bytes.Buffer{}, testRecorder()); err != nil {
			t.Fatal(err)
		}
	}
	stop := func(agent, agentType string) {
		t.Helper()
		if err := Stop(payload(t, agent, agentType, cwd), &bytes.Buffer{}, testRecorder()); err != nil {
			t.Fatal(err)
		}
	}
	register := func(agent string) {
		t.Helper()
		identityFromTheRewrite(t, runDir, agent, runlive.BenchAgentType)
		if _, _, err := record.RegisterSeat(record.Identity{Run: run, SeatID: "judge"}, string(seatenv.RunFromEnv), "assemble"); err != nil {
			t.Fatalf("agent %s registering the bench for the assembly: %v", agent, err)
		}
	}

	// The assembly sitting stamps the outcome, as `bench outcome` does — the event, then the mark —
	// and stops before it assembles: no SubagentStop reaches the hooks.
	start(orig, runlive.BenchAgentType)
	register(orig)
	if _, err := record.Append(record.Identity{Run: run, SeatID: "judge"}, &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED), Prose: proto.String("nobody was ready"),
	}); err != nil {
		t.Fatal(err)
	}
	if err := runlive.WriteOutcomeMark(runDir, orig); err != nil {
		t.Fatal(err)
	}

	// THE RESUME: the same occasion, a fresh agent.
	start(fresh, runlive.BenchAgentType)
	register(fresh)
	stop(fresh, runlive.BenchAgentType)
	opened, closed := spans(t, run, fresh)
	if opened != 1 || closed != 1 {
		t.Fatalf("the resumed assembly sitting's span: %d open(s), %d close(s) on the record, want one of each", opened, closed)
	}
	if seat, found, err := record.RegisteredSeatOfAgent(run, fresh); err != nil || !found || seat != "judge" {
		t.Errorf("the fresh agent's register binds it to %q (found %v, %v), want judge", seat, found, err)
	}

	// A LENS AFTER THE OUTCOME is in no run: the decided run's row is live for the bench alone.
	before := len(*got)
	start(lens, "frank-exchange-of-views:red-lens-evidence")
	stop(lens, "frank-exchange-of-views:red-lens-evidence")
	if len(*got) != before {
		t.Errorf("a lens dispatched after the outcome was handed to the writer: %+v", (*got)[before:])
	}

	// CAPTURED: nothing binds — a bench agent, and the agent that recorded the outcome.
	if err := os.WriteFile(filepath.Join(runDir, runlive.CapturedMark), []byte("2026-10-24T00:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before = len(*got)
	start("a5econd0000000000b", runlive.BenchAgentType)
	stop(orig, runlive.BenchAgentType)
	if len(*got) != before {
		t.Errorf("an agent was handed to the writer after the run was captured: %+v", (*got)[before:])
	}
}

// spans counts the hook-written sitting opens and closes on the record naming agent.
func spans(t *testing.T, run record.Run, agent string) (opened, closed int) {
	t.Helper()
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if e.GetSittingOpen().GetAgentId() == agent {
			opened++
		}
		if e.GetSittingClose().GetAgentId() == agent {
			closed++
		}
	}
	return opened, closed
}
