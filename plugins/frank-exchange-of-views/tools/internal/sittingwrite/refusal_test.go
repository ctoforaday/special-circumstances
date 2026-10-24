package sittingwrite

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"

	"google.golang.org/protobuf/proto"
)

// A CALL THE HOOK DENIED, OR ONE THAT FAILED, IS LOGGED AS THE SEAT'S — under the seat the record binds the
// agent to, as the tool's entry — and an agent the record binds to no seat is no seat, so nothing is written.
func TestAHookRefusalIsLoggedUnderTheRefusedSeat(t *testing.T) {
	runDir := newRun(t)
	registerAs(t, runDir, "agent-1", "red-lens-adversary")
	run := mustRun(t, runDir)
	if err := WriteToolEntry(runDir, Refusal, "agent-1", "refused a `Read` the run does not serve: the report is not a file a seat reads"); err != nil {
		t.Fatal(err)
	}
	if err := WriteToolEntry(runDir, Failure, "agent-1", "`Read` failed on `/x`: File does not exist <!--fx:F-1234abcd-->"); err != nil {
		t.Fatal(err)
	}
	if err := WriteToolEntry(runDir, Refusal, "agent-nobody", "refused a `Read`"); err != nil {
		t.Fatalf("an unregistered agent is no seat and is skipped, not an error: %v", err)
	}
	evs, _, err := record.EventsOf(run, recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("want the registered seat's refusal and failure, got %d", len(evs))
	}
	for i, want := range []recordpb.LogType{recordpb.LogType_LOG_TYPE_REFUSAL, recordpb.LogType_LOG_TYPE_FAILURE} {
		l, _ := recordpb.BodyAs[*recordpb.Log](evs[i])
		if evs[i].GetSeatId() != "red-lens-adversary" || l.GetType() != want || l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
			t.Errorf("entry %d is not the tool's %v under the seat: %s %v", i, want, evs[i].GetSeatId(), l)
		}
		if strings.Contains(l.GetText(), "<!--") {
			t.Errorf("a live anchor reached the log: %q", l.GetText())
		}
	}
	// The anchor the failure named is still named, by its id.
	if l, _ := recordpb.BodyAs[*recordpb.Log](evs[1]); !strings.HasSuffix(l.GetText(), "File does not exist F-1234abcd") {
		t.Errorf("the failure's entry lost the anchor it named: %q", l.GetText())
	}
}

// otherSeat is another seat's own connection on runDir's record, as a seat in another process holds.
func otherSeat(t *testing.T, runDir string) *sql.DB {
	t.Helper()
	db, err := recordsql.OpenSeparate(filepath.Join(runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// AN ENTRY IS WRITTEN AFTER THE OUTCOME TOO. A seat's call the hook saw denied or failed after the
// run ended is friction the operator reads (record.requireOpenToSeats lets every log entry land),
// including one whose outcome commits at the moment the hook logs it; and the harness's own facts
// about a sitting are no seat's act, so they land as well.
func TestAHookEntryIsWrittenAfterTheOutcome(t *testing.T) {
	runDir := newRun(t)
	registerAs(t, runDir, "agent-1", "red-lens-adversary")
	outcome := recordtest.At(t, "judge", "race:outcome", &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(),
		VerdictBasis: proto.String(record.VerdictAsserted), Prose: proto.String("the run ended")})
	if err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{outcome}, func() error {
		return WriteToolEntry(runDir, Failure, "agent-1", "`Read` failed on `/x`: File does not exist")
	}); err != nil {
		t.Fatalf("an entry on a run whose outcome landed at that moment was refused: %v", err)
	}
	if err := WriteToolEntry(runDir, Refusal, "agent-1", "refused a `Read` the run does not serve"); err != nil {
		t.Fatalf("an entry on a run holding an outcome was refused: %v", err)
	}
	evs, _, err := record.EventsOf(mustRun(t, runDir), recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("want both entries under the seat on a record holding an outcome, got %d", len(evs))
	}
	for _, e := range evs {
		if e.GetSeatId() != "red-lens-adversary" {
			t.Errorf("an entry is attributed to %q, want the seat the agent is bound to", e.GetSeatId())
		}
	}
	// The harness's own fact about the sitting lands too: it is no seat's act.
	if err := WriteLimit(runDir, "agent-1", 1, 150); err != nil {
		t.Fatalf("the harness's sitting limit after the outcome was refused: %v", err)
	}
	if limits, _, err := record.EventsOf(mustRun(t, runDir), recordpb.EventType_EVENT_TYPE_SITTING_LIMIT); err != nil || len(limits) != 1 {
		t.Errorf("the harness's sitting limit after the outcome: %d event(s), err %v; want 1", len(limits), err)
	}
}

// AN ENDED RUN TAKES NO HOOK ENTRY, AND SAYS SO. Once capture has marked the run's end
// (record.MarkEnded, which capture calls at its close), the hook's entry is refused loudly — the
// error names the ended run — and nothing is written; so is the harness's sitting limit.
func TestAHookEntryIsRefusedOnceTheRunHasEnded(t *testing.T) {
	runDir := newRun(t)
	registerAs(t, runDir, "agent-1", "red-lens-adversary")
	if err := record.MarkEnded(mustRun(t, runDir), "2026-10-24T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if err := WriteToolEntry(runDir, Failure, "agent-1", "`Read` failed on `/x`: File does not exist"); !errors.Is(err, record.ErrRunEnded) {
		t.Errorf("a hook entry on an ended run = %v, want the ended run's refusal", err)
	}
	if err := WriteLimit(runDir, "agent-1", 1, 150); !errors.Is(err, record.ErrRunEnded) {
		t.Errorf("the harness's sitting limit on an ended run = %v, want the ended run's refusal", err)
	}
	if evs, _, err := record.EventsOf(mustRun(t, runDir), recordpb.EventType_EVENT_TYPE_LOG); err != nil || len(evs) != 0 {
		t.Errorf("the ended run took %d log entr(ies) (err %v)", len(evs), err)
	}
}
