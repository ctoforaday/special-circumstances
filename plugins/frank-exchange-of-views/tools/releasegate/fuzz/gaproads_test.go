package fuzz

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// aboutGapWords is the clause `mint --about-kind gap`'s refusal opens its reason with
// (cli/lens/mint.go aboutGapRefusal).
const aboutGapWords = "a gap is not the subject of another gap"

// aboutGapRefusedDrive is a lens's one `mint --about-kind gap` of the run, naming an open gap: the
// word is refused before any validate, naming both roads a point about a gap takes — a proposed act
// on it, or a mint superseding it (plans/findings-fold.md §III.A3). No draw types the word, so this
// drive is the gate's (about_gap_refused). An admission, another code, or a refusal naming no road
// fails the run.
func (r *runner) aboutGapRefusedDrive(seatID string) {
	gaps := r.openGaps()
	if len(gaps) == 0 {
		return
	}
	r.mu.Lock()
	if r.aboutGapDriven {
		r.mu.Unlock()
		return
	}
	r.aboutGapDriven = true
	r.mu.Unlock()
	gap := gaps[r.rng.Intn(len(gaps))]
	args := []string{"--json", "mint", "--seat-id", seatID, "--about-kind", "gap", "--about", gap,
		"--class", pick(r.rng, mintPoolClasses), "--problem", "fuzz: a point about " + gap,
		"--check-kind", "document", "--check", "acc", "--fix", dirApply,
		"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(), "--reason", "fuzz: about another gap"}
	out, err := r.exec(args...)
	if err == nil {
		r.noteActErr("mint --about-kind gap was admitted: " + firstLine(out))
		return
	}
	env, ok := refusalEnvelopeIn(out)
	switch {
	case !ok:
		r.noteActErr("mint --about-kind gap was refused with no refusal envelope: " + firstLine(out))
	case env.Code != string(feov.Validation):
		r.noteActErr("mint --about-kind gap was refused with code " + env.Code + ", not validation: " + firstLine(env.Error))
	case !strings.Contains(env.Error, aboutGapWords) || !strings.Contains(env.Error, "motion docket file") || !strings.Contains(env.Error, "--supersedes"):
		r.noteActErr("mint --about-kind gap was refused without naming both roads: " + firstLine(env.Error))
	default:
		r.mu.Lock()
		r.aboutGapRefused++
		r.mu.Unlock()
	}
}

// disputeMovedGap is a lens's dispute of a gap another seat regraded while the lens sat: the close it
// proposes is filed on the gap as the lens read it when its sitting opened, so the record refuses it as
// stale, naming the regrade, and admits it again once read (admitFresh). Lenses sit side by side, so
// this is the race the stale refusal exists for; once per run, and never in a forced-verdict run, whose
// bound a hand-filed docket motion would move. The gap's minter has returned from its dispatch: only it
// closes, supersedes or regrades the gap in a lens wave, so the filing races nothing but the regrade
// it is refused for. A run whose lenses never regrade under a sibling's
// sitting files none, which the sweep's stale-event gate counts across runs.
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
		own, sitting := minter == seatID, r.unreturned[minter] > 0
		r.mu.Unlock()
		if open[id] && !own && !sitting && !r.closedThisSitting(seatID, id) && !r.beforeTheBench(id) && r.exchanged(id) && !ruled[id] {
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
		"fuzz: "+gap+" was regraded while this lens sat, and its weight is the bench's to settle")...); !ok {
		return
	}
	if r.readPointOf(seatID, gap) <= before {
		r.noteActErr("a docket motion on " + gap + ", regraded by another seat after " + seatID + "'s sitting opened, landed with no stale refusal")
		return
	}
	r.mu.Lock()
	r.movedDisputes++
	r.mu.Unlock()
}

// A LENS'S DISPUTE OF A GAP ITS MINTER REGRADED WHILE THE LENS SAT IS REFUSED AS STALE, AND LANDS ONCE
// READ: the filer sits, the minter sits beside it, regrades the gap and returns, and the filer's docket
// motion meets the stale refusal naming the regrade, then lands on the retry. Between the filer's
// attempts the minter sits again and regrades (a sitting holds one regrade per gap) until it has
// regraded three times, so each attempt is refused for a change the previous refusal could not show,
// and the motion lands at the first attempt no regrade preceded. While the minter still sits, the
// drive files nothing; once the motion lands, the filer's sitting owes nothing dischargeOwed leaves.
func TestADisputeOfAGapItsMinterKeepsRegradingLandsOnceItStops(t *testing.T) {
	r, gap := staleRun(t)
	r.exchangeOn(t) // a proposal waits for the gap's first exchange
	r.dispatchSits("lens", staleFiler)
	grades := []string{"low", "medium", "high", "low"}
	regrade := func(i int) {
		r.dispatchSits("lens", staleMinter)
		r.mustExec(t, "--json", "--seat-id", staleMinter, "regrade", "--id", gap, "--severity", grades[i], "--reason", "fuzz: regrade "+grades[i])
	}
	regrade(0)
	r.noteMint(staleMinter, gap)
	r.disputeMovedGap(staleFiler)
	if r.movedDisputed {
		t.Fatalf("the drive filed on %s while its minter still sat", gap)
	}
	r.dispatchReturns(staleMinter, false)
	next := 1
	r.betweenStaleRetries = func(seatID, gapID string) {
		if seatID != staleFiler || gapID != gap {
			t.Errorf("a retry of %s on %s, want %s's on %s", seatID, gapID, staleFiler, gap)
		}
		if next < 3 {
			regrade(next)
			r.dispatchReturns(staleMinter, false)
			next++
		}
	}
	r.disputeMovedGap(staleFiler)
	r.dischargeOwed(staleFiler)
	if refusal, err := record.EnvelopeRefusal(r.run(), staleFiler, true); err != nil || refusal != "" {
		t.Errorf("the filer's sitting owes after its dispute landed: %q (%v)", refusal, err)
	}
	r.dispatchReturns(staleFiler, false)
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if r.movedDisputes != 1 {
		t.Fatalf("the drive landed %d dispute(s) of the moved gap, want 1", r.movedDisputes)
	}
	if got := r.staleRefusals(staleFiler, gap); got != 3 {
		t.Errorf("the filer met %d stale refusal(s) on %s, want one per regrade, 3", got, gap)
	}
	// ONCE PER RUN: a second call files nothing.
	r.dispatchSits("lens", staleFiler)
	regrade(3)
	r.dispatchReturns(staleMinter, false)
	r.disputeMovedGap(staleFiler)
	r.dispatchReturns(staleFiler, false)
	if r.movedDisputes != 1 {
		t.Errorf("a second sitting's drive landed another dispute: %d", r.movedDisputes)
	}
}

// THE ONE `--about-kind gap` OF A RUN IS REFUSED NAMING BOTH ROADS, and a second call drives none.
func TestTheAboutGapDriveIsRefusedNamingBothRoads(t *testing.T) {
	r, _ := staleRun(t)
	r.dispatchSits("lens", staleFiler)
	r.aboutGapRefusedDrive(staleFiler)
	r.aboutGapRefusedDrive(staleFiler)
	r.dispatchReturns(staleFiler, false)
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if r.aboutGapRefused != 1 {
		t.Errorf("the drive counted %d refusal(s), want 1", r.aboutGapRefused)
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
