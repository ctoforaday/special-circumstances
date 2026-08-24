package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE SECTION EXISTS SO A HUMAN SEES THE ANSWER (#415). `verify` is the operator's verb and the
// fuzzer's oracle; neither is read by the person the report is for, while the engine declined to
// fail a lineage gap on the grounds that "the record is authoritative and already checked there".
// This is what makes that sentence true in the register it meant.
func TestRecordVerificationRendersEveryInvariantWithItsStatus(t *testing.T) {
	b := &record.Board{
		Events: []*record.Event{
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.Register{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}),
		},
		GapOrder: []string{"R1-1"},
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Open: false, Closure: &recordpb.Close{ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED)}},
		},
	}
	got := recordVerification(b)
	// THE HEADING IS CHECKED AS A WHOLE LINE, not as a prefix (#447). `HasPrefix` was satisfied
	// by "## Record verification (injected)" — so a rename of the section a human actually reads
	// passed here, and passed the golden suite too, because no golden covered the assembled
	// report at all. The golden is the real fix; this line stops the unit test from reading like
	// a check it was not.
	if first, _, _ := strings.Cut(got, "\n"); first != "## Record verification" {
		t.Fatalf("heading line = %q, want exactly \"## Record verification\":\n%s", first, got)
	}
	for _, want := range []string{"gaps-disposed", "pass-closes-all-gaps", "register-before-append"} {
		if !strings.Contains(got, want) {
			t.Errorf("invariant %q is not named in the section:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "held") || !strings.Contains(got, "violated") {
		t.Errorf("the section must carry a summary count:\n%s", got)
	}
	// It says where the answer came from. A reader must be able to tell this from a seat's
	// self-report, which is the distinction the whole record model rests on.
	if !strings.Contains(got, "replaying the event record") {
		t.Errorf("the section must say it is computed, not reported:\n%s", got)
	}
}

// A VIOLATION MUST READ AS ONE. The section is not a gate, so the only thing standing between a
// contradictory record and a reader who believes it is this text.
func TestRecordVerificationNamesAViolationAndItsOffender(t *testing.T) {
	b := &record.Board{
		Events: []*record.Event{
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.Register{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}),
		},
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
	}
	got := recordVerification(b)
	if !strings.Contains(got, "**FAIL**") {
		t.Errorf("a PASS over an open gap must render as FAIL:\n%s", got)
	}
	if !strings.Contains(got, "R1-1") {
		t.Errorf("the offender must be named, not merely counted:\n%s", got)
	}
	if !strings.Contains(got, "1 violated") {
		t.Errorf("the summary must count the violation:\n%s", got)
	}
	if !strings.Contains(got, "record-integrity finding") {
		t.Errorf("a violation must tell the reader what it means for the counts:\n%s", got)
	}
}

// THE THIRD STATE, AND THE REASON IT IS LOAD-BEARING HERE (#411). An inapplicable invariant
// checked nothing. Rendering it as a pass is exactly how pass-closes-all-gaps stayed invisible
// for its whole life, and a report is read by someone who cannot inspect the code.
func TestRecordVerificationDistinguishesNotApplicableFromHeld(t *testing.T) {
	b := &record.Board{
		Events: []*record.Event{
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.Register{}),
			recordtest.Event(t, "red-merge-r1", 1, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}),
		},
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
	}
	got := recordVerification(b)
	if !strings.Contains(got, "**n/a**") {
		t.Errorf("an inapplicable invariant must be marked n/a, not ok:\n%s", got)
	}
	if !strings.Contains(got, "1 did not apply") {
		t.Errorf("the summary must separate did-not-apply from held:\n%s", got)
	}
	if !strings.Contains(got, "checked NOTHING") {
		t.Errorf("the reader must be told an n/a is not a pass:\n%s", got)
	}
	// And it must NOT be counted among the ones that held.
	if strings.Contains(got, "7 held") {
		t.Errorf("an n/a invariant was counted as held:\n%s", got)
	}
}
