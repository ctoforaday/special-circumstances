package fuzz

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// disputeMovedGap is the author's dispute of a gap its raiser regraded while the author sat: the close
// it proposes is filed on the gap as the author read it when its sitting opened, so the record refuses
// it as stale, naming the regrade, and admits it again once read (admitFresh). Once per run, and never
// in a forced-verdict run, whose bound a hand-filed docket motion would move. The gap's raiser has
// returned from its dispatch, so the filing races nothing but the regrade it is refused for. The
// dispatch sits blue after the lenses, so a sweep run meets the race only where a raiser's sitting
// overlaps blue's; TestADisputeOfAGapItsMinterKeepsRegradingLandsOnceItStops drives it.
func (r *runner) disputeMovedGap(seatID string) {
	if r.forcedVerdict() {
		return
	}
	r.mu.Lock()
	done := r.movedDisputed
	r.mu.Unlock()
	if done {
		return
	}
	b, ok := r.board()
	if !ok {
		return
	}
	sitting := b.At.LatestSittingOf(seatID)
	if sitting == 0 {
		return
	}
	open := map[string]bool{}
	for _, id := range r.openGaps() {
		open[id] = true
	}
	// A RULING'S TOMBSTONE REFUSES EVERY PROPOSAL ON ITS GAP: the dispute is filed on a gap none stands on.
	ruled, ok := r.tombstoned()
	if !ok {
		return
	}
	gap := ""
	for _, e := range b.Events {
		rg := e.GetRegrade()
		if rg == nil || e.GetSeatId() == seatID || b.At.Of(e).ID <= sitting {
			continue
		}
		id := rg.GetGapId()
		r.mu.Lock()
		minter := r.minter[id]
		sitting := r.unreturned[minter] > 0
		r.mu.Unlock()
		d := r.scenarioOf(id)
		if open[id] && minter == e.GetSeatId() && !sitting && !r.closedThisSitting(seatID, id) && !r.beforeTheBench(id) && r.exchanged(id) && !ruled[id] &&
			d != dirDisputeWon && d != dirDisputeLost {
			gap = id
		}
	}
	if gap == "" {
		return
	}
	r.mu.Lock()
	if r.movedDisputed {
		r.mu.Unlock()
		return
	}
	r.movedDisputed = true
	r.mu.Unlock()
	before := r.readPointOf(seatID, gap)
	if _, ok := r.admitFresh(gap, proposeCloseArgs(seatID, gap, "not_a_defect",
		"fuzz: "+gap+" was regraded while blue sat, and its weight is the bench's to settle")...); !ok {
		return
	}
	if r.readPointOf(seatID, gap) <= before {
		r.noteActErr("a docket motion on " + gap + ", regraded by its raiser after " + seatID + "'s sitting opened, landed with no stale refusal")
		return
	}
	r.mu.Lock()
	r.movedDisputes++
	r.mu.Unlock()
}

// THE AUTHOR'S DISPUTE OF A GAP ITS RAISER REGRADED WHILE THE AUTHOR SAT IS REFUSED AS STALE, AND
// LANDS ONCE READ: blue sits, the raiser sits beside it, regrades the gap and returns, and blue's docket
// motion meets the stale refusal naming the regrade, then lands on the retry. Between blue's attempts
// the raiser sits again and regrades (a sitting holds one regrade per gap) until it has regraded three
// times, so each attempt is refused for a change the previous refusal could not show, and the motion
// lands at the first attempt no regrade preceded. While the raiser still sits, the drive files nothing;
// once the motion lands, blue's sitting owes nothing dischargeOwed leaves.
func TestADisputeOfAGapItsMinterKeepsRegradingLandsOnceItStops(t *testing.T) {
	const author = "blue-respond"
	r, gap := staleRun(t)
	r.exchangeOn(t) // a proposal waits for the gap's first exchange
	r.dispatchSits("blue", author)
	grades := []string{"low", "medium", "high"}
	regrade := func(i int) {
		r.dispatchSits("lens", staleMinter)
		r.mustExec(t, "--json", "--seat-id", staleMinter, "regrade", "--id", gap, "--severity", grades[i], "--reason", "fuzz: regrade "+grades[i])
	}
	regrade(0)
	r.noteMint(staleMinter, gap)
	r.disputeMovedGap(author)
	if r.movedDisputed {
		t.Fatalf("the drive filed on %s while its raiser still sat", gap)
	}
	r.dispatchReturns(staleMinter, false)
	next := 1
	r.betweenStaleRetries = func(seatID, gapID string) {
		if seatID != author || gapID != gap {
			t.Errorf("a retry of %s on %s, want %s's on %s", seatID, gapID, author, gap)
		}
		if next < 3 {
			regrade(next)
			r.dispatchReturns(staleMinter, false)
			next++
		}
	}
	r.disputeMovedGap(author)
	r.dischargeOwed(author)
	if refusal, err := record.EnvelopeRefusal(r.run(), author, true); err != nil || refusal != "" {
		t.Errorf("blue's sitting owes after its dispute landed: %q (%v)", refusal, err)
	}
	r.dispatchReturns(author, false)
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if r.movedDisputes != 1 {
		t.Fatalf("the drive landed %d dispute(s) of the moved gap, want 1", r.movedDisputes)
	}
	if got := r.staleRefusals(author, gap); got != 3 {
		t.Errorf("blue met %d stale refusal(s) on %s, want one per regrade, 3", got, gap)
	}
	// ONCE PER RUN: a second sitting's call files nothing.
	r.dispatchSits("blue", author)
	r.disputeMovedGap(author)
	r.dispatchReturns(author, false)
	if r.movedDisputes != 1 {
		t.Errorf("a second sitting's drive landed another dispute: %d", r.movedDisputes)
	}
}

// THE ONE-GAP REFUSAL STATES THE REFUSED SEAT'S ROAD BY WHOSE GAP HOLDS THE SUBJECT, and neither
// names a finding: the minter's second mint on the sentence meets its own gap's road, a sibling's
// meets another seat's.
func TestTheOneGapRefusalStatesEachRoad(t *testing.T) {
	r, gap := staleRun(t)
	for _, c := range []struct{ seat, road string }{{staleMinter, "own"}, {staleFiler, "another"}} {
		r.dispatchSits("lens", c.seat)
		args := []string{"--json", "mint", "--seat-id", c.seat, "--quote", "A § fuzz sentence to anchor findings.",
			"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz: again",
			"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason"}
		out, err := r.exec(args...)
		if err == nil || !strings.Contains(out, gap) {
			t.Fatalf("%s's mint over %s: %s", c.seat, gap, out)
		}
		r.oneGapAnswered(c.seat, out, args)
		r.dispatchReturns(c.seat, false)
		if r.oneGapRedirects[c.road] != 1 {
			t.Errorf("%s's refusal was counted on roads %v, want one on %q", c.seat, r.oneGapRedirects, c.road)
		}
	}
	if r.actErr != "" || r.oneGapNamingFinding != 0 {
		t.Errorf("the drive: %q; %d road(s) naming a finding", r.actErr, r.oneGapNamingFinding)
	}
}
