package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A DISPATCH'S RETURN WRITES WHAT THE HOOKS WRITE, through the writer they spawn: the envelope's
// admission, then the sitting's close, both under the agent its opening carried. Pinned here because
// the sweep reaches it by a draw and the deterministic tests in this package draw it on no seed.
func TestADispatchReturnWritesTheHooksAdmissionAndClose(t *testing.T) {
	r, agent := returnedRun(t, "red-chair", true)
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, e := range fam.Events {
		if b, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok && b.GetAgentId() == agent {
			order = append(order, "admitted")
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && b.GetAgentId() == agent {
			order = append(order, "closed")
		}
	}
	if len(order) != 2 || order[0] != "admitted" || order[1] != "closed" {
		t.Fatalf("the record holds %v for %s, want its admission and then its close", order, agent)
	}
	failed, err := record.FailedSittings(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 0 || r.failedSittings != 0 {
		t.Errorf("a sitting closed with its envelope admitted reads as failed: the record reads %+v, the harness counted %d", failed, r.failedSittings)
	}
}

// THE CLOSE NAMES THE AGENT'S TRANSCRIPT, as the SubagentStop hook's spawn does, and the transcript
// the harness writes reads NOT MEASURED for what it is: a conversation with no model turn in it.
func TestTheCloseNamesATranscriptThatReadsNotMeasured(t *testing.T) {
	r, agent := returnedRun(t, "red-chair", true)
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		t.Fatal(err)
	}
	path, err := record.TranscriptOf(fam, agent)
	if err != nil {
		t.Fatalf("the record names no readable transcript for %s: %v", agent, err)
	}
	if filepath.Dir(path) != filepath.Join(r.runDir, "transcripts") {
		t.Errorf("the close names %s, want a transcript under the run's transcripts/", path)
	}
	for _, m := range record.SeatModels(record.Run{}, fam) {
		if m.SeatID != "red-chair" {
			continue
		}
		if m.Measured() || !strings.Contains(m.NotMeasured, "carries no assistant turn naming a model") {
			t.Errorf("red-chair's served model = %+v, want NOT MEASURED because its transcript carries no model turn", m)
		}
		return
	}
	t.Error("record.SeatModels has no row for red-chair")
}

// A SITTING CLOSED WITH NO ADMISSION IS A FAILED ONE, and the harness counts it as one, so the
// sweep's check that the record reads as failed what the harness closed unadmitted has both sides.
func TestASittingClosedUnadmittedIsCountedFailed(t *testing.T) {
	r, agent := returnedRun(t, "red-chair", false)
	failed, err := record.FailedSittings(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 1 || failed[0].AgentID != agent || r.failedSittings != 1 {
		t.Fatalf("the record reads %+v as failed and the harness counted %d, want %s's one sitting on both", failed, r.failedSittings, agent)
	}
}

// A CLOSE THE WRITER REFUSED IS NOT COUNTED CLOSED, so the sweep's census counts the sittings the
// record holds closed, not the attempts.
func TestARefusedCloseIsNotCountedClosed(t *testing.T) {
	r := stagedRunner(t)
	r.forceVerified = true
	agent := r.registerTwice("red-chair", "")
	if agent == "" {
		t.Fatalf("the agent-bearing register was refused: %s", r.actErr)
	}
	r.sittingAgent = map[string]string{"red-chair": agent}
	r.writer = filepath.Join(t.TempDir(), "no-such-writer")
	before := sittingsClosed.Load()
	r.dispatchReturns("red-chair", false)
	if r.actErr == "" {
		t.Fatal("a writer that does not exist reported no failure")
	}
	if got := sittingsClosed.Load() - before; got != 0 || r.failedSittings != 0 {
		t.Errorf("a refused close counted %d closed and %d failed, want neither", got, r.failedSittings)
	}
}

// THE GOJA agent BINDING RETURNS EVERY DISPATCH IT MAKES (dispatchReturns), so a sitting the engine
// dispatches ends as a live one does. Driven through driveDebate, the sweep's own loop, without the
// release gate.
func TestTheAgentBindingReturnsItsDispatch(t *testing.T) {
	r := stagedRunner(t)
	r.register("blue", "blue-synthesize")
	if _, err := r.do("ingest", "blue-synthesize").result(); err != nil {
		t.Fatalf("ingest the report: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, settledErr := driveDebate(r, `globalThis.__result = agent("SEAT_ID: blue-synthesize", {label: "blue-synthesize", model: "haiku"})
			.then(function(){ return {verdict: "VERIFIED", epochs: 0}; });`)
		if settledErr != "" {
			t.Fatalf("dispatch %d: %s", i+1, settledErr)
		}
	}
	requireEverySittingReturned(t, r)
}

// returnedRun is a run where seatID sat under an agent and its dispatch returned, admitted or not.
func returnedRun(t *testing.T, seatID string, admitted bool) (*runner, string) {
	t.Helper()
	r := stagedRunner(t)
	// A forced run closes every sitting, so no draw leaves this one open.
	r.forceVerified = true
	agent := r.registerTwice(seatID, "")
	if agent == "" {
		t.Fatalf("the agent-bearing register was refused: %s", r.actErr)
	}
	r.sittingAgent = map[string]string{seatID: agent}
	r.dispatchReturns(seatID, admitted)
	if r.actErr != "" {
		t.Fatalf("the return's writes: %s", r.actErr)
	}
	return r, agent
}

// stagedRunner is a runner over a run holding a report and a one-lane cast.
func stagedRunner(t *testing.T) *runner {
	t.Helper()
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"),
		[]byte("# § fuzz\n\nA § fuzz sentence to anchor findings.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatal(err)
	}
	return newRunner(bin, runDir, newLockedRand(7))
}
