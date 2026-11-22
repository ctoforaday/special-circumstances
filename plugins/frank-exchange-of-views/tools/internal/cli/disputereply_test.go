package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A REPLY MUST NAME WHAT IT ANSWERS — and after #344 it names an ID, not a pair.
//
// THE HISTORY MATTERS BECAUSE IT IS THE ARGUMENT FOR THE ID. A dispute was filed against ONE
// grade, and blue could contest more than one on the same gap. The answer recorded only the gap,
// so two disputes and one "rejected" left a record nobody could resolve. The fix at the time was
// to make `--dimension` required on the answer, so the join became (gap_id, dimension) — the
// orchestrator had always matched on that pair off the ENVELOPE, which evaporates with the
// session, while the durable record kept the lossier half.
//
// That pair was still a composed key, and #312 is where it ran out: two petitions from one seat
// in one round could not be told apart at all, because their pair had no third component to
// reach for. A motion has an id, and the docketing that puts it before the bench has its own:
// `motion docket rule --id D` answers D, and through it D's motion, and nothing else, so both
// failures below stop being checks against a validator and become unrepresentable.
//
// These tests survive the collapse asserting the SAME property against the new mechanism: an
// answer joins to exactly one ask, and an answer to an ask nobody filed is refused.

// TWO MOTIONS ON THE SAME GAP ARE STILL DISTINCT.
//
// This is the case a pair key could not represent at all, and it is why the id exists. Blue puts
// G1 before the bench twice — one at a time: a second proposal while the first stands is refused
// naming it, and blue files again once the bench has answered the first. The two filings are
// separate motions with separate ids, and the ruling on the first leaves the second open.
func TestTwoMotionsOnOneGapAreTellableApart(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nA claim.\n")
	g := mintGap(t, runDir, "G1", "spec_underspecification")
	exchangeOn(t, runDir, g)
	first := proposeClose(t, runDir, "blue-respond", g, "not_a_defect", "the harm needs two failures")
	second := "and the second failure is gated upstream"
	if out, err := runAt(t, closeProposalArgs(runDir, "blue-respond", g, "not_a_defect", second)...); err == nil ||
		!strings.Contains(err.Error()+out, first+" stands unanswered") {
		t.Fatalf("a second proposal while %s stands = %v %q, want refused naming it", first, err, out)
	}

	// Rule only the first. Keyed on the gap, this ruling would match BOTH filings, and the record
	// would show one answer to two questions.
	k := docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, first)
	conveneBench(t, runDir, g)
	args := benchRuleArgs(k, "remand", "the gap needs one more exchange")
	args = append([]string{args[0], args[1], args[2], "--run", runDir, "--seat-id", "judge"}, args[3:]...)
	if _, err := runAt(t, args...); err != nil {
		t.Fatalf("a properly addressed ruling was refused: %v", err)
	}
	later := proposeClose(t, runDir, "blue-respond", g, "not_a_defect", second)
	if later == first {
		t.Fatalf("the second filing reused the first's id %s", first)
	}

	ev := lastBody(t, runDir, &recordpb.MotionRule{})
	if got := ev.GetMotionId(); got != first {
		t.Errorf("motion_id = %q, want %s — the record must carry the join, not leave it to be inferred", got, first)
	}
	if got := ev.GetDocket().GetDisposition(); got != recordpb.Disposition_DISPOSITION_REMANDED {
		t.Errorf("ruling = %q, want remanded", recordpb.Word(got))
	}
	b, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range record.MotionsOf(b.Events, b.At) {
		if m.ID == later && m.Ruled() {
			t.Errorf("the ruling on %s answered %s too", first, later)
		}
	}
}

// AN ANSWER TO A QUESTION NOBODY ASKED IS REFUSED — the dangling-reference rule this codebase
// applies to gap ids, applied to the other side of the conversation.
func TestARulingOnAnUnfiledMotionIsRefused(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nA claim.\n")
	g := mintGap(t, runDir, "G1", "spec_underspecification")
	docketFile(t, runDir, "blue-respond", g, "s")

	args := benchRuleArgs("D-00000007", "remand", "sure")
	_, err := runAt(t, append([]string{args[0], args[1], args[2], "--run", runDir, "--seat-id", "judge"}, args[3:]...)...)
	if err == nil {
		t.Fatal("the bench ruled a docketing the tool never made — a reply with no question, joining to nothing")
	}
	if !strings.Contains(err.Error(), "no docketing D-00000007 is on the record") {
		t.Errorf("the refusal must say the reference resolves to nothing: %v", err)
	}
}
