package fuzz

import (
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
	r.dispatchOpens("red-chair")
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
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2")); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	for i := 0; i < 3; i++ {
		_, settledErr := driveDebate(r, `globalThis.__result = agent("SEAT_ID: blue-synthesize", {label: "blue-synthesize", model: "haiku"})
			.then(function(){ return {verdict: "VERIFIED", epochs: 0}; });`)
		if settledErr != "" {
			t.Fatalf("dispatch %d: %s", i+1, settledErr)
		}
	}
	requireEveryDispatchReturned(t, r)
}

// A DISPATCH THE HARNESS GAVE UP ON REACHES THE ENGINE AS A THROW, and nothing from its envelope
// reaches the record: the seat ended without returning an envelope the record and the schema would
// accept (endsWithoutEnvelope), so the close lands with no admission and the record reads the sitting
// as failed. Driven through driveDebate, without the release gate.
func TestADispatchEndedWithoutItsEnvelopeThrowsAndReadsFailed(t *testing.T) {
	r := givenUpDispatch(t, `{type: "object"}`, func(r *runner) { r.endWithoutEnvelope = 2 })
	if refused := r.distinctSchemaRefusals(); len(refused) != 0 {
		t.Errorf("the schema refused an envelope it accepts: %v", refused)
	}
}

// A SCHEMA-REFUSED ENVELOPE NEVER REACHES THE ENGINE: the Workflow tool enforces the seat's schema,
// so live the seat retries until it ends without a valid envelope and agent() throws. The binding
// settles the dispatch as it does any give-up, and the record reads the sitting as failed.
func TestASchemaRefusedEnvelopeThrowsAndReadsFailed(t *testing.T) {
	r := givenUpDispatch(t, `{type: "object", required: ["no_such_field"], properties: {no_such_field: {type: "string"}}}`, nil)
	if refused := r.distinctSchemaRefusals(); len(refused) != 1 {
		t.Errorf("the schema refused %v, want the one envelope it was handed", refused)
	}
}

// givenUpDispatch drives one red-chair dispatch carrying schema, its sitting opened under an agent,
// and asserts that debate.js received the give-up form both ways it awaits a dispatch — a throw
// from an awaited agent() and a null from parallel() — and that the record holds each sitting
// closed, unadmitted and failed.
func givenUpDispatch(t *testing.T, schema string, arrange func(*runner)) *runner {
	t.Helper()
	r := stagedRunner(t)
	r.openUnderAgent = true
	if arrange != nil {
		arrange(r)
	}
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2")); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	before := failedSittingsClosed.Load()
	result, settledErr := driveDebate(r, `globalThis.__result = (async function () {
		var opts = {label: "red-chair", model: "haiku", schema: `+schema+`};
		var awaited = "fulfilled";
		try { await agent("SEAT_ID: red-chair", opts); } catch (e) { awaited = "threw: " + (e && e.message ? e.message : e); }
		var fanned = await parallel([function () { return agent("SEAT_ID: red-chair", opts); }]);
		return {awaited: awaited, fanned: fanned[0] === null ? "null" : "fulfilled"};
	})();`)
	if settledErr != "" {
		t.Fatal(settledErr)
	}
	if got, want := result["awaited"], "threw: "+endedWithoutEnvelope("red-chair"); got != want {
		t.Errorf("an awaited dispatch the harness gave up on settled %q, want %q", got, want)
	}
	if got := result["fanned"]; got != "null" {
		t.Errorf("a dispatch the harness gave up on under parallel() settled %v, want null", got)
	}
	requireEveryDispatchReturned(t, r)
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		t.Fatal(err)
	}
	var admitted, closed int
	for _, e := range fam.Events {
		if _, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok {
			admitted++
		}
		if _, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok {
			closed++
		}
	}
	if admitted != 0 || closed != 2 {
		t.Errorf("the record holds %d admission(s) and %d close(s), want no admission and both closes", admitted, closed)
	}
	failed, err := record.FailedSittings(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if len(failed) != 2 || r.failedSittings != 2 || failedSittingsClosed.Load()-before != 2 {
		t.Errorf("the record reads %+v as failed and the harness counted %d, want both sittings on both", failed, r.failedSittings)
	}
	return r
}

// A RESUMED ASSEMBLY SITTING FINDS THE OUTCOME ITS FAILED ONE STAMPED and goes on to assemble, as
// debate.js sits the assembly again in place: the second sitting drives no outcome the record,
// closed to seat acts once the outcome stands, would refuse.
func TestAnAssemblySatAgainAfterAGiveUpAssemblesOnTheStampedOutcome(t *testing.T) {
	r := stagedRunner(t)
	r.openUnderAgent = true
	r.endWithoutEnvelope = 1
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2")); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	result, settledErr := driveDebate(r, `globalThis.__result = (async function () {
		var opts = {label: "judge", model: "haiku", schema: {type: "object"}};
		var prompt = "SEAT_ID: judge OCCASION: assemble Debate outcome: UNVERIFIED";
		var first = "fulfilled";
		try { await agent(prompt, opts); } catch (e) { first = "threw"; }
		var second = await agent(prompt, opts);
		return {first: first, second: second ? "fulfilled" : "null"};
	})();`)
	if settledErr != "" {
		t.Fatal(settledErr)
	}
	if result["first"] != "threw" || result["second"] != "fulfilled" {
		t.Errorf("the assembly sittings settled %v then %v, want a throw and then the envelope", result["first"], result["second"])
	}
	if got, err := record.RecordedOutcome(r.run()); err != nil || got != "unverified" {
		t.Errorf("the recorded outcome = %q (%v), want unverified", got, err)
	}
	requireEveryDispatchReturned(t, r)
}

// A RECEIVED ENVELOPE'S SUCCESSORS ARE WHAT THE HARNESS EXPECTS THE ENGINE TO SIT BLUE ON NEXT, read
// off the envelope debate.js was handed (envelopeReceived). The binding calls it for no envelope it
// gave up on; the sweep holds that call, since a blue-respond sitting engaged on a successor needs a
// whole debate behind it.
func TestAReceivedEnvelopeReadiesItsSuccessorSitting(t *testing.T) {
	r := &runner{}
	r.envelopeReceived(map[string]any{"manifest": []any{}})
	if r.againOn != nil {
		t.Fatalf("an envelope naming no successor readied %v", r.againOn)
	}
	r.envelopeReceived(map[string]any{"successors": []any{"G-00000001"}})
	if len(r.againOn) != 1 || r.againOn[0] != "G-00000001" {
		t.Fatalf("a received envelope's successors readied %v, want G-00000001", r.againOn)
	}
}

// A RETURN WITH NO DISPATCH OPEN IS A HARNESS DEFECT, and it is loud: it fails the run's dispatch
// check rather than folding into a count of zero.
func TestAReturnWithNoDispatchOpenIsAHarnessError(t *testing.T) {
	r := stagedRunner(t)
	r.dispatchReturns("red-chair", false)
	if r.actErr == "" || r.unreturnedDispatches() == "" {
		t.Errorf("a return with no dispatch open was not recorded: actErr %q, dispatches %q", r.actErr, r.unreturnedDispatches())
	}
}

// dispatchSits is a dispatch's opening as the agent binding makes it: the dispatch opens, and the
// seat sits.
func (r *runner) dispatchSits(role, seatID string, occasion ...string) {
	r.dispatchOpens(seatID)
	r.sit(role, seatID, occasion...)
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
	r.dispatchOpens(seatID)
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
