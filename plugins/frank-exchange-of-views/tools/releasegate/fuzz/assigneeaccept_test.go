package fuzz

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE ASSIGNEE'S ACCEPT IS DRIVEN, AND COUNTED BY ITS ACT (§III.B9 assignee_accepts_by_act): blue's
// close proposal on the lens's gap, filed after one counted exchange, is accepted by the lens at its
// next sitting through `motion docket accept` — the close and its ruling in one transaction — and the
// drive counts it; the gap is then closed with no docketing written for it.
func TestTheAssigneeAcceptsAProposalAsFiled(t *testing.T) {
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
	r := newRunner(bin, runDir, newLockedRand(7))
	r.forceVerified = true
	const sentence = "A § fuzz sentence to anchor findings."
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	r.chairPlans(t)
	lens := fuzzLensSeats[0]
	r.dispatchSits("lens", lens)
	r.mustExec(t, "--json", "--seat-id", lens, "show", "report")
	r.mustExec(t, "--json", "--seat-id", lens, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	r.dispatchReturns(lens, false)
	gapID := r.openGaps()[0]
	sitBlue := func() {
		r.chairPlans(t)
		r.dispatchOpens("blue-respond")
		r.openUnderAgent = true
		r.sit("blue", "blue-respond")
		r.openUnderAgent = false
	}
	sitBlue()
	r.do("edit", "blue-respond").set("--quote", "fuzz sentence").set("--new", "fuzz statement").
		set("--answers", gapID).set("--reason", "fuzz: blue's answer").admit()
	r.splitQuestion("blue-respond")
	r.coverQuestionParts("blue-respond")
	r.dispatchReturns("blue-respond", r.envelopeEnds("blue-respond"))
	r.dispatchSits("chair", "red-chair")
	r.mustExec(t, "--json", "dispatch", "next", "--seat-id", "red-chair")
	r.dispatchReturns("red-chair", false)
	r.dispatchOpens(lens)
	r.openUnderAgent = true
	r.sit("lens", lens)
	r.openUnderAgent = false
	r.dispatchReturns(lens, r.envelopeEnds(lens))
	sitBlue()
	r.mustExec(t, withJSON(proposeCloseArgs("blue-respond", gapID, "not_a_defect", "fuzz: blue's dispute of "+gapID))...)
	// Blue's closing on its own proposal is owed in the assignee's window; the accept ends the window.
	r.dispatchReturns("blue-respond", false)
	ms := r.unruledDocketMotions(gapID)
	if len(ms) != 1 {
		t.Fatalf("proposals standing on %s after blue's filing = %v, want one", gapID, ms)
	}
	r.dispatchOpens(lens)
	r.openUnderAgent = true
	r.sit("lens", lens)
	r.openUnderAgent = false
	r.acceptAsAssignee(lens, gapID, ms[0], "close")
	if r.actErr != "" {
		t.Fatalf("the assignee's accept: %s", r.actErr)
	}
	if got := r.docketTally()["assignee_accepts:close"]; got != 1 {
		t.Errorf("assignee_accepts:close = %d, want 1", got)
	}
	if r.openOnRecord(gapID) {
		t.Errorf("%s is open after its assignee accepted blue's close", gapID)
	}
	if ds := r.docketings(); len(ds) != 0 {
		t.Errorf("an accepted proposal was docketed: %d docketing(s)", len(ds))
	}
}
