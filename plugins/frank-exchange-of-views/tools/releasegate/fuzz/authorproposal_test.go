package fuzz

import (
	"encoding/json"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A GAP HAS TWO PARTIES (§III.B9): its raiser, the lens that minted it, and its subject's author,
// blue-respond for every report subject. Only the author proposes an act on the gap; the raiser answers
// the proposal by taking it (`regrade --accept M`, `close --accept M`) or leaves it, and a proposal left
// untaken is docketed for the bench, whose answer for the author applies it and whose answer for the
// raiser applies the raiser's text. Under a docketing either party concedes: the raiser by closing its
// gap or by taking the proposal the docketing holds, the author by accepting the raiser's statement
// (`edit --accept G`, `retire --accept G`). The sweep gates each act (motion:blue:<act>,
// accept:proposal:<act>, concession:<party>:<act>) but the retire, whose statement only a closing on an
// anchor gap makes: `retire --accept` is held by the flag gate.

// authorProposalTurn rotates the author's act across the sweep's runs in one process — close, regrade
// — so both are driven whichever runs reach the drive.
var authorProposalTurn atomic.Int64

// authorProposalActs is the acts the author's proposal rotates through.
var authorProposalActs = []string{"close", "regrade"}

// forcedAuthorProposalActs is authorProposalSeed's rotation: a regrade first.
var forcedAuthorProposalActs = []string{"regrade", "close"}

// proposeAsAuthor is blue, the gaps' author, proposing an act on a lens's open gap in its own sitting,
// never in a run forced to a verdict: on a coin where a gap qualifies, and on authorProposalSeed's run
// at every sitting one does. The gap has had its counted exchange, holds no proposal or docketing, no
// ruling's tombstone, no closing of blue's this sitting, and is not a dispute scenario's (whose one
// proposal is blue's regrade). The coin is drawn only once a gap qualifies, so a run with none draws as
// it did. A proposal the record admits once its read point is fresh is admitFresh's.
//
// THE FORCED RUN OPENS ON A REGRADE, and the bench sides with the raiser against it (benchRules): the
// raiser's side of a regrade states the grades as they stand, so it is always stated, and
// `--for-raiser` lands. A close reaches that side only where the lens's closing states text.
func (r *runner) proposeAsAuthor(seatID string) {
	if r.forcedVerdict() {
		return
	}
	gaps := r.authorProposalGaps(seatID)
	if len(gaps) == 0 || (!r.forceAuthorProposal && !r.coin(20)) {
		return
	}
	act := authorProposalActs[int((authorProposalTurn.Add(1)-1)%int64(len(authorProposalActs)))]
	if r.forceAuthorProposal {
		r.mu.Lock()
		n := r.authorProposalsFiled
		r.authorProposalsFiled++
		r.mu.Unlock()
		act = forcedAuthorProposalActs[n%len(forcedAuthorProposalActs)]
	}
	r.authorProposal(seatID, act, gaps[r.rng.Intn(len(gaps))])
}

// authorProposalGaps is the gaps blue may propose on (proposeAsAuthor).
func (r *runner) authorProposalGaps(seatID string) []string {
	r.mu.Lock()
	final := finalEpoch(r.planEpoch, r.planLimit)
	r.mu.Unlock()
	// A PROPOSAL FILED AT THE FINAL EPOCH IS NOT HEARD IN TIME, and the record refuses it.
	if final {
		return nil
	}
	ruled, ok := r.tombstoned()
	if !ok {
		return nil
	}
	var gaps []string
	for _, id := range r.openGaps() {
		r.mu.Lock()
		minter := r.minter[id]
		r.mu.Unlock()
		d := r.scenarioOf(id)
		if minter == "" || seatclass.RoleOfSeat(minter) != "lens" || ruled[id] || d == dirDisputeWon || d == dirDisputeLost ||
			r.closedThisSitting(seatID, id) || r.beforeTheBench(id) || !r.exchanged(id) {
			continue
		}
		gaps = append(gaps, id)
	}
	return gaps
}

// authorProposal is blue's proposal of act on gap id; it reports the gap, "" where the act has nothing
// to propose.
func (r *runner) authorProposal(seatID, act, id string) string {
	reason := "fuzz: " + seatID + ", the author of " + id + "'s subject, proposes its act"
	switch act {
	case "close":
		r.admitFresh(id, proposeCloseArgs(seatID, id, "not_a_defect", reason)...)
	case "regrade":
		held := r.benchGrades(id)
		free := slices.DeleteFunc(slices.Clone(disputeDims), func(d string) bool { _, locked := held[d]; return locked })
		if len(free) == 0 {
			return ""
		}
		dim := pick(r.rng, free)
		r.admitFresh(id, "motion", "docket", "file", "regrade", "--seat-id", seatID, "--id", id,
			"--"+dim, r.gradeOtherThan(r.currentGrade(id, dim)), "--reason", reason)
	}
	return id
}

// concessionTurn rotates a party's concession across the sweep's runs in one process: one docketing in
// three that a drive may concede it concedes, and the rest stay the bench's, so the bench's answers
// keep their cells.
var concessionTurn atomic.Int64

// docketsAreTheBenchs reports whether this run leaves every proposal and docketing to the bench, so no
// party takes or concedes one: a run forced to a verdict, whose bound a concession would move, and
// authorProposalSeed's and benchOrdersSeed's runs, whose arms are the bench's answers.
func (r *runner) docketsAreTheBenchs() bool {
	return r.forcedVerdict() || r.forceAuthorProposal || r.forceBenchOrders
}

// concedes reports whether a party's drive concedes a docketing on gapID now: never where the dockets
// are the bench's (docketsAreTheBenchs); never on a dispute scenario's gap, whose proposal is the
// bench's to rule, or a computation scenario's, which closes only behind a reproduction (the COMPUTE
// oracle); and on the rotation's turn.
func (r *runner) concedes(gapID string) bool {
	if r.docketsAreTheBenchs() {
		return false
	}
	switch r.scenarioOf(gapID) {
	case dirDisputeWon, dirDisputeLost, dirCompute, dirComputeDrifts:
		return false
	}
	return (concessionTurn.Add(1)-1)%3 == 0
}

// raiserTakesTurn rotates the raiser's answer to the proposals it meets across the sweep's runs in one
// process: it takes two in five, and leaves the rest to be docketed and ruled.
var raiserTakesTurn atomic.Int64

// takeProposals is the raiser seatID taking, by its own act, each proposal standing on a gap of its own
// on the rotation's turn (raiserTakesTurn): `regrade --accept M` or `close --accept M`, whichever the
// proposal proposes, the tool supplying its grades or its word and verification. Under a docketing the
// take is the raiser's concession, ending the docketing and the proposal with no ruling. Never where the
// dockets are the bench's (docketsAreTheBenchs); never on a
// dispute scenario's gap, whose proposal is the bench's to rule, or a computation scenario's, which
// closes only behind a reproduction; and a close only where nothing the close guard reads refuses it —
// a live-answer demand nothing meets, a failed check standing on the gap.
func (r *runner) takeProposals(seatID string) {
	if r.docketsAreTheBenchs() {
		return
	}
	b, ok := r.board()
	if !ok {
		r.noteActErr("the record could not be read for the proposals standing on " + seatID + "'s gaps")
		return
	}
	for _, m := range record.MotionsOf(b.Events, b.At) {
		if m == nil || m.Subject != "docket" || !m.Standing() || (m.Act != "close" && m.Act != "regrade") ||
			r.minterOf(m.GapID) != seatID || !r.openOnRecord(m.GapID) {
			continue
		}
		switch r.scenarioOf(m.GapID) {
		case dirDisputeWon, dirDisputeLost, dirCompute, dirComputeDrifts:
			continue
		}
		if m.Act == "close" {
			if d, ok := r.demandOf(m.GapID); !ok || d.Unmet() {
				continue
			}
			if _, standing := r.standingFailedCheckOf(m.GapID); standing {
				continue
			}
		}
		if (raiserTakesTurn.Add(1)-1)%5 >= 2 {
			continue
		}
		docketed := r.standingDocketings(m.GapID)
		if _, ok := r.admitFresh(m.GapID, m.Act, "--seat-id", seatID, "--accept", m.ID,
			"--reason", "fuzz: "+seatID+" takes "+m.ID+", its author's proposal on "+m.GapID); !ok {
			continue
		}
		r.countGate("accept:proposal:" + m.Act)
		if len(docketed) > 0 {
			r.countGate("accept:proposal:concession")
			r.endedUnruled(docketed, seatID+"'s take of "+m.ID)
		}
	}
}

// endedUnruled holds a concession to ending each docketing it was filed under with no ruling: none
// stands after it, and none reads ruled.
func (r *runner) endedUnruled(ds []*record.Docketing, act string) {
	ended := map[string]bool{}
	for _, d := range ds {
		ended[d.ID] = true
	}
	for _, d := range r.docketings() {
		if !ended[d.ID] {
			continue
		}
		switch {
		case d.Standing():
			r.noteActErr(act + ", a concession, left docketing " + d.ID + " standing")
		case d.Ruled():
			r.noteActErr(act + ", a concession, reads docketing " + d.ID + " ruled")
		}
	}
}

// concedeByAccept is the author's concession in its sitting: blue takes the raiser's statement on a gap
// docketed at impasse or holding blue's proposal, ending the docketing and the proposal with no ruling,
// on the concession rotation (concedes). A pair is taken by `edit --accept G` where it is the mint's
// verified prescription with its span still in the report — the pair the apply drive sends — and no
// closing of the raiser's withdrew or replaced it; a retire the raiser's closing states is taken by
// `retire --accept G` while its anchor is still in the report. Never on an IGNORE scenario's gap, whose
// oracle holds blue to no act answering it. One per sitting.
func (r *runner) concedeByAccept(seatID string) {
	if r.docketsAreTheBenchs() {
		return
	}
	for _, d := range r.standingDocketings("") {
		switch d.Cause {
		case recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL:
		default:
			continue
		}
		if !r.openOnRecord(d.GapID) || r.scenarioOf(d.GapID) == dirIgnore {
			continue
		}
		cur, err := record.RenderRun(r.run())
		if err != nil {
			r.noteActErr("the report could not be rendered for blue's accept of " + d.GapID + ": " + err.Error())
			return
		}
		st, found, err := record.StatementOf(r.run(), d.GapID)
		if err != nil {
			r.noteActErr("the raiser's statement on " + d.GapID + " could not be read: " + err.Error())
			return
		}
		if !found || st.Withdrawn {
			continue
		}
		verb, key, fn := "edit", "concession:author:edit", ""
		if st.Retire {
			if st.AboutRef == "" || !strings.Contains(cur, st.AboutRef) {
				continue
			}
			verb, key = "retire", "concession:author:retire"
		} else {
			gid, fo, pairNew := r.proposalFor(d.GapID, cur)
			// THE STATEMENT IS STILL THE MINT'S PAIR, compared as text: the board's pair carries the
			// gap's marker, the statement does not.
			if gid == "" || !strings.Contains(cur, fo) || plainText(st.FixNew) != plainText(pairNew) {
				continue
			}
			fn = pairNew
		}
		if !r.concedes(d.GapID) {
			return
		}
		if _, ok := r.do(verb, seatID).set("--accept", d.GapID).
			set("--reason", "fuzz: blue takes the raiser's statement on "+d.GapID+", conceding its docketing").admit(); !ok {
			return
		}
		r.countGate(key)
		if fn != "" {
			r.noteVerbatimGap(d.GapID, fn)
		}
		r.endedUnruled([]*record.Docketing{d}, "blue's accept of "+d.GapID)
		return
	}
}

// AUTHORPROPOSALSEED'S RUN REACHES THE BENCH ON THE AUTHOR'S REGRADE, pinned deterministically: filed on
// a lens's gap after its counted exchange, it is docketed once the raiser's sitting leaves it untaken,
// and the bench sides with the raiser — the gap's lens.
func TestTheAuthorsRegradeIsLeftToTheBenchWhichSidesWithTheRaiser(t *testing.T) {
	r, lens, gapID := proposalRun(t, dirIgnore)
	r.forceAuthorProposal = true

	r.dispatchSits("blue", "blue-respond")
	if got := r.authorProposalGaps("blue-respond"); !slices.Contains(got, gapID) {
		t.Fatalf("blue may propose on %v, want %s among them", got, gapID)
	}
	r.proposeAsAuthor("blue-respond")
	if r.actErr != "" {
		t.Fatalf("blue's regrade on %s: %s", gapID, r.actErr)
	}
	r.dispatchReturns("blue-respond", false)
	if ms := r.standingProposals(gapID); len(ms) != 1 {
		t.Fatalf("proposals standing on %s after blue's filing = %v, want one", gapID, ms)
	}
	// THE FORCED RUN'S RAISER TAKES NOTHING: the proposal is the bench's.
	raiserTakesTurn.Store(0)
	r.dispatchSits("lens", lens)
	r.takeProposals(lens)
	r.dispatchReturns(lens, false)
	if got := r.docketTally()["accept:proposal:regrade"]; got != 0 {
		t.Fatalf("the raiser took the forced run's proposal: accept:proposal:regrade = %d", got)
	}

	r.partiesClose(t, lens)
	r.dispatchSits("bench", "judge", "docket")
	r.benchDisposes("judge", gapID, "repaired", "--reason", "fuzz: the grades as they stand hold")
	if r.actErr != "" {
		t.Fatalf("the bench's ruling: %s", r.actErr)
	}
	tally := r.docketTally()
	if tally["answer:proposal:raiser"] != 1 || tally["motion:blue:regrade"] != 1 {
		t.Errorf("answer:proposal:raiser = %d, motion:blue:regrade = %d, want the author's one regrade and the bench's one answer for the raiser; tally %v",
			tally["answer:proposal:raiser"], tally["motion:blue:regrade"], tally)
	}
	r.dispatchReturns("judge", r.envelopeEnds("judge"))
}

// THE RAISER TAKES THE AUTHOR'S PROPOSAL BY ITS OWN ACT, in its sitting: the take answers the proposal
// accepted and lands the act it proposes; under the proposal's docketing it is the raiser's concession,
// ending the docketing with no ruling. The rotation (raiserTakesTurn) is set to a taking turn first.
func TestTheRaiserTakesTheAuthorsProposal(t *testing.T) {
	for _, c := range []struct {
		name     string
		act      string
		docketed bool
	}{
		{"a regrade, undocketed", "regrade", false},
		{"a close, under its docketing", "close", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, lens, gapID := proposalRun(t, dirIgnore)
			r.dispatchSits("blue", "blue-respond")
			if r.authorProposal("blue-respond", c.act, gapID) != gapID || r.actErr != "" {
				t.Fatalf("blue's %s on %s: %s", c.act, gapID, r.actErr)
			}
			r.dispatchReturns("blue-respond", false)
			motions := r.standingProposals(gapID)
			if len(motions) != 1 {
				t.Fatalf("proposals standing on %s = %v, want one", gapID, motions)
			}
			if c.docketed {
				// LEFT UNTAKEN through the raiser's sitting, the proposal is docketed at the next plan.
				r.dispatchSits("lens", lens)
				r.dispatchReturns(lens, false)
				r.chairPlans(t)
				if len(r.standingDocketings(gapID)) != 1 {
					t.Fatalf("no docketing stands on %s after its raiser's sitting left %s untaken", gapID, motions[0])
				}
				r.partiesFileClosings(t, lens)
			}
			raiserTakesTurn.Store(0)
			r.dispatchSits("lens", lens)
			r.docketedOffTheList("lens", lens)
			r.takeProposals(lens)
			r.sittingOwesNothing(t, lens)
			if r.actErr != "" {
				t.Fatalf("the raiser's take: %s", r.actErr)
			}
			tally := r.docketTally()
			if tally["accept:proposal:"+c.act] != 1 {
				t.Errorf("accept:proposal:%s = %d, want 1; tally %v", c.act, tally["accept:proposal:"+c.act], tally)
			}
			if want := map[bool]int{true: 1}[c.docketed]; tally["accept:proposal:concession"] != want {
				t.Errorf("accept:proposal:concession = %d, want %d", tally["accept:proposal:concession"], want)
			}
			if left := r.standingOn(gapID); len(left) > 0 {
				t.Errorf("%v still stands on %s after the raiser took %s", left, gapID, motions[0])
			}
			b, ok := r.board()
			if !ok {
				t.Fatal("the board could not be read")
			}
			for _, m := range record.MotionsOf(b.Events, b.At) {
				if m.ID == motions[0] && m.Ruling != record.AcceptedWord {
					t.Errorf("%s reads answered %q, want %q", m.ID, m.Ruling, record.AcceptedWord)
				}
			}
			// THE MOTIONS PAGE A SEAT READS says the same.
			out := r.mustExec(t, seat.GroupOf("motions"), "motions", "--seat-id", "judge")
			var page struct {
				Motions []struct {
					ID     string `json:"id"`
					Ruled  bool   `json:"ruled"`
					Ruling string `json:"ruling"`
				} `json:"motions"`
			}
			if err := json.Unmarshal([]byte(out), &page); err != nil {
				t.Fatalf("the motions page: %v", err)
			}
			for _, m := range page.Motions {
				if m.ID == motions[0] && (!m.Ruled || m.Ruling != record.AcceptedWord) {
					t.Errorf("the motions page reads %s ruled %v %q, want ruled %q", m.ID, m.Ruled, m.Ruling, record.AcceptedWord)
				}
			}
			if open := r.openOnRecord(gapID); open != (c.act == "regrade") {
				t.Errorf("%s open = %v after the raiser took a %s", gapID, open, c.act)
			}
		})
	}
}

// proposalRun is a record holding one document-check gap a lens minted with scenario fix, and the
// mint's flags in prescribe, past its counted exchange: the runner, the lens, and the gap.
func proposalRun(t *testing.T, fix string, prescribe ...string) (*runner, string, string) {
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
	r := newRunner(bin, runDir, newLockedRand(7))
	const sentence = "A § fuzz sentence to anchor findings."
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	r.chairPlans(t)
	lens := fuzzLensSeats[0]
	r.dispatchSits("lens", lens)
	r.mustExec(t, "--json", "--seat-id", lens, "show", "report")
	r.mustExec(t, append([]string{"--json", "--seat-id", lens, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", fix, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason"}, prescribe...)...)
	r.dispatchReturns(lens, false)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	gapID := open[0]
	r.noteMint(lens, gapID)
	r.exchangeOn(t)
	return r, lens, gapID
}

// closedByParty reports whether gapID closed by a party's answer rather than a ruling or a repair: a
// close on evs that takes a proposal on it (Close.accepts), or its raiser's close where a concession
// ended a docketing on it unruled.
func (r *runner) closedByParty(evs []*record.Event, gapID string) bool {
	byRaiser := false
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok && c.GetGapId() == gapID {
			if c.GetAccepts() != "" {
				return true
			}
			byRaiser = byRaiser || e.GetSeatId() == r.minterOf(gapID)
		}
	}
	if !byRaiser {
		return false
	}
	for _, d := range r.docketings() {
		if d.GapID == gapID && !d.Standing() && !d.Ruled() {
			return true
		}
	}
	return false
}

// docketProposal is blue's proposal of act on gapID, left untaken through its raiser's sitting and
// docketed at the next plan; it reports the docketing.
func (r *runner) docketProposal(t *testing.T, lens, act, gapID string) *record.Docketing {
	t.Helper()
	r.dispatchSits("blue", "blue-respond")
	if r.authorProposal("blue-respond", act, gapID) != gapID || r.actErr != "" {
		t.Fatalf("blue's %s on %s: %s", act, gapID, r.actErr)
	}
	r.dispatchReturns("blue-respond", false)
	r.dispatchSits("lens", lens)
	r.dispatchReturns(lens, false)
	r.chairPlans(t)
	ds := r.standingDocketings(gapID)
	if len(ds) != 1 {
		t.Fatalf("docketings standing on %s after its raiser left blue's %s untaken = %d, want one", gapID, act, len(ds))
	}
	return ds[0]
}

// THE RAISER'S CLOSE OF ITS GAP UNDER THE AUTHOR'S PROPOSAL: refused naming the proposal while it stands
// undocketed, and once it is docketed admitted as the raiser's concession, ending the docketing and the
// proposal with no ruling.
func TestTheRaisersCloseIsRefusedUnderAProposalAndConcedesItsDocketing(t *testing.T) {
	r, lens, gapID := proposalRun(t, dirIgnore)
	r.dispatchSits("blue", "blue-respond")
	if r.authorProposal("blue-respond", "regrade", gapID) != gapID || r.actErr != "" {
		t.Fatalf("blue's regrade on %s: %s", gapID, r.actErr)
	}
	r.dispatchReturns("blue-respond", false)
	r.dispatchSits("lens", lens)
	r.docketedOffTheList("lens", lens)
	r.closeBeforeTheBench(lens, gapID)
	r.dispatchReturns(lens, false)
	if r.actErr != "" || r.docketedClosesRefused != 1 {
		t.Fatalf("the raiser's close under the undocketed proposal: %d refused (want 1); %s", r.docketedClosesRefused, r.actErr)
	}
	r.chairPlans(t)
	if len(r.standingDocketings(gapID)) != 1 {
		t.Fatalf("no docketing stands on %s once its raiser left the proposal untaken", gapID)
	}
	r.partiesFileClosings(t, lens)
	concessionTurn.Store(0)
	r.dispatchSits("lens", lens)
	r.docketedOffTheList("lens", lens)
	r.closeBeforeTheBench(lens, gapID)
	r.sittingOwesNothing(t, lens)
	if r.actErr != "" {
		t.Fatalf("the raiser's concession: %s", r.actErr)
	}
	if got := r.docketTally()["concession:raiser:close"]; got != 1 {
		t.Errorf("concession:raiser:close = %d, want 1", got)
	}
	if left := r.standingOn(gapID); len(left) > 0 || r.openOnRecord(gapID) {
		t.Errorf("after the raiser's concession %s is open %v with %v standing on it", gapID, r.openOnRecord(gapID), left)
	}
}

// THE AUTHOR'S ACCEPT OF THE RAISER'S TEXT UNDER A DOCKETING IS ITS CONCESSION: `edit --accept G` lands
// the mint's prescription, answers the gap, and ends the docketing and the proposal it holds unruled.
func TestTheAuthorsAcceptConcedesItsDocketing(t *testing.T) {
	r, lens, gapID := proposalRun(t, dirApply, "--new", "A § fuzz sentence, repaired.")
	d := r.docketProposal(t, lens, "close", gapID)
	r.partiesFileClosings(t, lens)
	concessionTurn.Store(0)
	r.dispatchSits("blue", "blue-respond")
	r.docketedOffTheList("blue", "blue-respond")
	r.concedeByAccept("blue-respond")
	r.sittingOwesNothing(t, "blue-respond")
	if r.actErr != "" {
		t.Fatalf("blue's concession: %s", r.actErr)
	}
	if got := r.docketTally()["concession:author:edit"]; got != 1 {
		t.Fatalf("concession:author:edit = %d, want 1; tally %v", got, r.docketTally())
	}
	for _, e := range r.docketings() {
		if e.ID == d.ID && (e.Standing() || e.Ruled()) {
			t.Errorf("docketing %s standing %v, ruled %v after blue's accept; want ended unruled", d.ID, e.Standing(), e.Ruled())
		}
	}
	if left := r.standingProposals(gapID); len(left) > 0 {
		t.Errorf("proposal(s) %v still stand on %s after blue's concession", left, gapID)
	}
}

// partiesFileClosings is the docketed gap's parties filing their closings on it, as the sweep's
// sittings do before either concedes: the chair plans, and the lens and blue each sit and file theirs.
func (r *runner) partiesFileClosings(t *testing.T, lens string) {
	t.Helper()
	r.chairPlans(t)
	for _, p := range []struct{ role, seat string }{{"lens", lens}, {"blue", "blue-respond"}} {
		r.dispatchSits(p.role, p.seat)
		r.fileClosings(p.seat)
		r.dispatchReturns(p.seat, false)
	}
}

// sittingOwesNothing ends seatID's sitting as the drive does — every item it owes discharged — and holds
// its envelope admitted.
func (r *runner) sittingOwesNothing(t *testing.T, seatID string) {
	t.Helper()
	r.dischargeOwed(seatID)
	if refusal, err := record.EnvelopeRefusal(r.run(), seatID, true); err != nil || refusal != "" {
		t.Errorf("%s's sitting owes after its act: %q (%v)", seatID, refusal, err)
	}
	r.dispatchReturns(seatID, false)
}
