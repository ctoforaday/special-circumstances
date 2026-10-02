package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// A READ THAT FAILS IS AN ERROR, NOT A DIFFERENT NUMBER. MintAvenueID counts the proposals that
// STAND through live_event; when that read fails — locked, busy, a malformed view — the caller
// sees the error. It does not get an id minted from a second count that includes every struck
// proposal, which is what a "for an older record" second query did, on a record the open had
// already held to this binary's schema (recordsql.requireDeclaredSchema), so the only way to
// reach it was a real error.
//
// The failure is made by dropping the view through the run's cached handle: every later read
// goes through the same handle and fails the same way a busy database would, and the fixture is
// arranged so the corrected count (Q2) and the uncorrected count (Q3) differ — a fold to the
// wrong query is visible as the wrong id, not as a coincidence.
func TestAFailedLiveReadIsAnErrorNotAnUncorrectedCount(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	judge := sit(t, run, "judge")
	propose := func(line string) *recordpb.Avenue {
		return &recordpb.Avenue{AvenueId: proto.String("Q1"), Line: proto.String(line), Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum()}
	}
	k := mustAppend(t, blue, propose("try the  method")).GetKey()
	mustAppend(t, blue, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordpb.AvenueStatus_AVENUE_STATUS_PURSUED.Enum(), SupersedesStatus: proto.String("proposed")})
	if _, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_AVENUE, k, "the shell ate it"), propose("try the recorded method")); err != nil {
		t.Fatal(err)
	}
	mustAppend(t, judge, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED), Prose: proto.String("ended")})

	// The healthy answers first, so the assertions below are about the failure and not the fixture.
	if id, err := MintAvenueID(run); err != nil || id != "Q2" {
		t.Fatalf("MintAvenueID over a healthy record = (%q, %v), want Q2: one proposal stands, its struck original does not count", id, err)
	}
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

	if id, err := MintAvenueID(run); err == nil {
		t.Errorf("MintAvenueID over a failing live read = (%q, nil), want the error — an id here is minted from the uncorrected count", id)
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
