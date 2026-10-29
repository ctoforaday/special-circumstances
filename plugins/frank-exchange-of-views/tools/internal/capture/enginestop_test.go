package capture

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

const engineError = "agent({schema}): subagent completed without calling StructuredOutput (after in-conversation nudge)"

// stoppedRun is a run whose chair sat once and returned with no admitted envelope, followed by the
// engine's stop as the lead records it; with resumed, the chair opens a sitting after the stop.
func stoppedRun(t *testing.T, last time.Time, resumed bool) record.Run {
	t.Helper()
	dir := t.TempDir()
	stamp := "2006-01-02T15:04:05.000000000Z07:00"
	n := 0
	at := func(seat string, body proto.Message) *recordpb.Event {
		n++
		return recordtest.Stamped(recordtest.At(t, seat, "", body), last.Add(time.Duration(n-20)*time.Second).UTC().Format(stamp))
	}
	chair := func(agent string) *recordpb.SittingOpen {
		return &recordpb.SittingOpen{AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:red-chair"), SeatId: proto.String("red-chair")}
	}
	evs := []*recordpb.Event{
		at(record.HarnessSeat, &recordpb.Cast{SeatIds: []string{"red-chair"}}),
		at(record.HarnessSeat, chair("C1")),
		at(record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("C1")}),
		at(record.HarnessSeat, &recordpb.Log{Text: proto.String(engineError), Type: recordpb.LogType_LOG_TYPE_FAILURE.Enum(), Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()}),
	}
	if resumed {
		evs = append(evs, at(record.HarnessSeat, chair("C2")))
	}
	recordtest.Seed(t, dir, evs...)
	return runtest.Open(t, dir)
}

// A RUN THE ENGINE STOPPED IS NOT A RUN IN FLIGHT, however recently the stop was written: with no
// outcome and the engine's failure entry last, liveness FAILS quoting the error and naming the
// sitting that failed — where it read "captured mid-run" off the stop's own recency (m35).
func TestLivenessReadsTheEnginesStopAsStopped(t *testing.T) {
	now := time.Date(2026, 10, 27, 17, 0, 0, 0, time.UTC)
	a := LivenessAudit(stoppedRun(t, now, false), now.Add(10*time.Second))
	if a.Verdict != "FAIL" {
		t.Fatalf("a run the engine stopped audited %s — want FAIL:\n%s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "ENGINE STOPPED") || !strings.Contains(a.Detail, engineError) {
		t.Errorf("the FAIL does not say the engine stopped the run, quoting its error:\n%s", a.Detail)
	}
	if !strings.Contains(a.Detail, "red-chair sitting 1") {
		t.Errorf("the FAIL does not name the sitting that closed with no admitted envelope:\n%s", a.Detail)
	}
}

// A STOP A RESUME SAT PAST IS NOT THE RUN'S END: a sitting opened after it, and liveness reads the
// record's movement as it did before.
func TestLivenessReadsAResumedRunByItsMovement(t *testing.T) {
	now := time.Date(2026, 10, 27, 17, 0, 0, 0, time.UTC)
	a := LivenessAudit(stoppedRun(t, now, true), now.Add(10*time.Second))
	if strings.Contains(a.Detail, "ENGINE STOPPED") {
		t.Errorf("a run resumed past the engine's stop audited as stopped:\n%s", a.Detail)
	}
}

// A SITTING THAT CLOSED WITH NO ADMITTED ENVELOPE WARNS, named; a record with none passes.
func TestFailedSittingsAuditNamesEachFailedSitting(t *testing.T) {
	now := time.Date(2026, 10, 27, 17, 0, 0, 0, time.UTC)
	a := FailedSittingsAudit(stoppedRun(t, now, false))
	if a.Verdict != "WARN" || !strings.Contains(a.Detail, "red-chair sitting 1") || !strings.Contains(a.Detail, "agent C1") {
		t.Errorf("a chair sitting closed with no admission audited %s:\n%s", a.Verdict, a.Detail)
	}
	if a := FailedSittingsAudit(writeRunForLiveness(t, 4, time.Second, now, true)); a.Verdict != "PASS" {
		t.Errorf("a record with no failed sitting audited %s:\n%s", a.Verdict, a.Detail)
	}
}
