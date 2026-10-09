package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// A READ THAT FAILS IS AN ERROR, NOT A DIFFERENT ANSWER. RecordedOutcome reads the outcome that
// STANDS through live_event; when that read fails — locked, busy, a malformed view — the caller
// sees the error. It does not get a verdict read by a second query over every outcome row, struck
// ones included, on a record the open has already held to this binary's schema
// (recordsql.requireDeclaredSchema), so the only way to reach it is a real error.
//
// The failure is made by dropping the view through the run's cached handle: every later read
// goes through the same handle and fails the same way a busy database would.
func TestAFailedLiveReadIsAnErrorNotAnUncorrectedAnswer(t *testing.T) {
	run := corrRun(t)
	judge := sit(t, run, "judge")
	mustAppend(t, judge, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED), Prose: proto.String("ended")})

	// The healthy answers first, so the assertions below are about the failure and not the fixture.
	if v, err := RecordedOutcome(run); err != nil || v != recordpb.Word(recordpb.RunOutcome_RUN_OUTCOME_HALTED) {
		t.Fatalf("RecordedOutcome over a healthy record = (%q, %v)", v, err)
	}

	db, err := openRunForRead(run)
	if err != nil || db == nil {
		t.Fatalf("open the run's handle: (%v, %v)", db, err)
	}
	if _, err := db.Exec(`DROP VIEW "live_event"`); err != nil {
		t.Fatal(err)
	}

	// RecordedOutcome returns the read it cannot make as the error, with no verdict beside it. What
	// it must NOT do is answer the question from a second query over every outcome row, struck
	// ones included — and what TerminalVerdict must not do is take that error for "not ended".
	if v, err := RecordedOutcome(run); err == nil || v != "" {
		t.Errorf("RecordedOutcome over a failing live read = (%q, %v), want the error and no verdict — a verdict here was read around the failure", v, err)
	}
	if v, err := TerminalVerdict(run); err == nil || v != "" {
		t.Errorf("TerminalVerdict over a failing live read = (%q, %v), want the error — \"\" here reads as a run in flight", v, err)
	}
}
