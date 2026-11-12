package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// hookSittings writes each sitting through the writer the sitting hooks spawn, in order: an opening
// for every sitting, and a close for every sitting closes holds true at its index. The last sitting
// of each seat opens and is left in flight when inFlight is set. The closes are the writer's own, so
// a writer that stopped closing reads here as the hook that stopped.
func hookSittings(t *testing.T, closes []bool, inFlight bool) record.Run {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.CreatedRecord(t, dir)
	types := []string{"frank-exchange-of-views:red-chair", "frank-exchange-of-views:red-lens-evidence"}
	write := func(phase sittingwrite.Phase, agent, agentType string) {
		t.Helper()
		if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: dir, Phase: phase, AgentID: agent, AgentType: agentType}, nil); err != nil {
			t.Fatalf("the writer refused the %s end of agent %s: %v", phase, agent, err)
		}
	}
	for i, closed := range closes {
		agent, agentType := fmt.Sprintf("a%02d", i), types[i%len(types)]
		write(sittingwrite.Open, agent, agentType)
		if closed {
			write(sittingwrite.Close, agent, agentType)
		}
	}
	if inFlight {
		for i, agentType := range types {
			write(sittingwrite.Open, fmt.Sprintf("live%d", i), agentType)
		}
	}
	return runtest.Open(t, dir)
}

// A RUN THE HOOK CLOSED EVERY ENDED SITTING OF PASSES, its in-flight sittings named as not closed yet.
func TestSittingClosesPassesARunWhoseEndedSittingsClosed(t *testing.T) {
	a := SittingClosesAudit(hookSittings(t, []bool{true, true, true, true}, true))
	if a.Verdict != "PASS" || !strings.Contains(a.Detail, "4 closed, 2 not closed yet") {
		t.Errorf("a run whose every ended sitting closed audited %s:\n%s", a.Verdict, a.Detail)
	}
}

// A RUN WITH NOT ONE CLOSE FAILS: its sittings ended — each seat sat again — and nothing closed any,
// so FailedSittingsOf reads zero and no lens retires. This is the SubagentStop hook gone silent.
func TestSittingClosesFailsARunWithNoClose(t *testing.T) {
	run := hookSittings(t, []bool{false, false, false, false}, true)
	a := SittingClosesAudit(run)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "6 sitting(s) opened under an agent, 4 of them ended") {
		t.Errorf("a run with no close audited %s:\n%s", a.Verdict, a.Detail)
	}
	if f, err := record.FailedSittings(run); err != nil || len(f) != 0 {
		t.Fatalf("failed sittings = %v (%v), want none: the fixture must be the plausible zero this audit exists for", f, err)
	}
}

// ONE ENDED SITTING WITH NO CLOSE AMONG CLOSES THAT LANDED WARNS, naming it: an agent can die before
// SubagentStop fires, and the record cannot tell that from one close the hook lost.
func TestSittingClosesWarnsNamingAnEndedSittingWithNoClose(t *testing.T) {
	a := SittingClosesAudit(hookSittings(t, []bool{true, false, true, true}, false))
	if a.Verdict != "WARN" || !strings.Contains(a.Detail, "1 of 4") || !strings.Contains(a.Detail, "red-lens-evidence sitting 1") || !strings.Contains(a.Detail, "agent a01") {
		t.Errorf("one ended sitting with no close audited %s:\n%s", a.Verdict, a.Detail)
	}
}

// A CAPTURE WHILE EVERY SITTING IS IN FLIGHT IS NOT A HOOK THAT STOPPED: no sitting has ended, so no
// close is owed yet.
func TestSittingClosesDoesNotConvictSittingsInFlight(t *testing.T) {
	a := SittingClosesAudit(hookSittings(t, nil, true))
	if a.Verdict != "PASS" || !strings.Contains(a.Detail, "0 closed, 2 not closed yet") {
		t.Errorf("a run whose sittings are all in flight audited %s:\n%s", a.Verdict, a.Detail)
	}
}

// A RECORD NO HOOK REACHED IS NOT MEASURED, never a pass: no opening carries an agent to owe a close.
func TestSittingClosesIsNotMeasuredWithoutAnAgentOpening(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, recordtest.At(t, record.HarnessSeat, "", &recordpb.Cast{SeatIds: []string{"red-chair"}}))
	a := SittingClosesAudit(runtest.Open(t, dir))
	if a.Verdict != "SKIP" || !strings.Contains(a.Detail, "NOT MEASURED") {
		t.Errorf("a record with no agent opening audited %s:\n%s", a.Verdict, a.Detail)
	}
}

// CAPTURE RUNS THE AUDIT, AND ITS FAIL FAILS CAPTURE: a test of the function is not a test of the
// call, and a run whose hook went silent exits 2 like any other run an audit failed.
func TestCaptureFailsARunWithNoClose(t *testing.T) {
	run := hookSittings(t, []bool{false, false}, true)
	if err := os.MkdirAll(filepath.Join(run.Dir(), "trajectories"), 0o755); err != nil {
		t.Fatal(err)
	}
	audits, _, exitFail, err := Run(run, t.TempDir(), t.TempDir(), time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	var got *Audit
	for i := range audits {
		if audits[i].Check == "sitting-closes" {
			got = &audits[i]
		}
	}
	if got == nil || got.Verdict != "FAIL" {
		t.Fatalf("capture's sitting-closes audit = %+v, want it run and FAILED on a run with no close", got)
	}
	if !exitFail {
		t.Error("capture reports no failure for a run its sitting-closes audit FAILED")
	}
}
