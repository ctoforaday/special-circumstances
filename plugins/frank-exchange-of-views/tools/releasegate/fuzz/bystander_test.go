package fuzz

import (
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A PROPOSAL BY A SEAT THAT IS NOT THE GAP'S PARTY (§III.B9, amendment 5): another lens, or the chair,
// proposes a close or a regrade of a lens's open gap, and the dispatch dockets it for the bench, whose
// answer for the author applies it and whose answer for the raiser — the gap's lens — applies red's
// text. The sweep gates each (role, act) cell (motion:<role>:<act>); left to blue's drives, no other
// filer ever proposed.

// bystanderTurn rotates each filer role's act across the sweep's runs in one process — close, regrade
// — so every cell is driven whichever runs reach the drive.
var bystanderTurn = map[string]*atomic.Int64{"lens": {}, "chair": {}}

// bystanderActs is the acts a bystander's proposal rotates through.
var bystanderActs = []string{"close", "regrade"}

// proposeAsBystander is seatID — a lens, or the chair — proposing an act on another lens's open gap,
// never in a run forced to a verdict: on a coin where a gap qualifies, and on bystanderSeed's run at
// every sitting one does. The gap has had its counted exchange, holds no proposal or docketing, no
// ruling's tombstone, is not seatID's own and not a dispute scenario's (whose one proposal is blue's
// regrade). The coin is drawn only once a gap qualifies, so a run with none draws as it did. A proposal
// the record admits once its read point is fresh is admitFresh's.
//
// THE FORCED RUN OPENS ON A REGRADE, and the bench sides with the raiser against it (benchRules): the
// raiser's side of a regrade states the grades as they stand, so it is always stated, and
// `--for-raiser` lands. A close reaches that side only where the lens's closing states text.
func (r *runner) proposeAsBystander(seatID string) {
	role := seatclass.RoleOfSeat(seatID)
	turn := bystanderTurn[role]
	if turn == nil || r.forcedVerdict() {
		return
	}
	gaps := r.bystanderGaps(seatID)
	if len(gaps) == 0 || (!r.forceBystander && !r.coin(20)) {
		return
	}
	act := bystanderActs[int((turn.Add(1)-1)%int64(len(bystanderActs)))]
	if r.forceBystander {
		r.mu.Lock()
		if r.bystanderFiled == nil {
			r.bystanderFiled = map[string]int{}
		}
		n := r.bystanderFiled[role]
		r.bystanderFiled[role]++
		r.mu.Unlock()
		act = forcedBystanderActs[n%len(forcedBystanderActs)]
	}
	r.bystanderProposal(seatID, act, gaps[r.rng.Intn(len(gaps))])
}

// forcedBystanderActs is bystanderSeed's rotation, by filer role: a regrade first.
var forcedBystanderActs = []string{"regrade", "close"}

// bystanderGaps is the gaps seatID may propose on as a bystander (proposeAsBystander).
func (r *runner) bystanderGaps(seatID string) []string {
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
		if minter == "" || minter == seatID || seatclass.RoleOfSeat(minter) != "lens" || ruled[id] ||
			d == dirDisputeWon || d == dirDisputeLost || r.beforeTheBench(id) || !r.exchanged(id) {
			continue
		}
		gaps = append(gaps, id)
	}
	return gaps
}

// bystanderProposal is seatID's proposal of act on gap id, as a bystander; it reports the gap, ""
// where the act has nothing to propose.
func (r *runner) bystanderProposal(seatID, act, id string) string {
	reason := "fuzz: " + seatID + ", no party to " + id + ", proposes its act"
	switch act {
	case "close":
		r.bystanderFiles(id, proposeCloseArgs(seatID, id, "not_a_defect", reason))
	case "regrade":
		held := r.benchGrades(id)
		free := slices.DeleteFunc(slices.Clone(disputeDims), func(d string) bool { _, locked := held[d]; return locked })
		if len(free) == 0 {
			return ""
		}
		dim := pick(r.rng, free)
		r.bystanderFiles(id, []string{"motion", "docket", "file", "regrade", "--seat-id", seatID, "--id", id,
			"--" + dim, r.gradeOtherThan(r.currentGrade(id, dim)), "--reason", reason})
	}
	return id
}

// bystanderRefusals is each refusal a bystander's proposal meets as the record working as ruled, by the
// words that name it, and its tally key: another seat's proposal landed on the gap since the drive read
// it — the lenses sit in parallel, so two may draw one gap (a gap holds one proposal at a time, F5) —
// or the gap closed since the read.
var bystanderRefusals = []struct {
	code       feov.Code
	words, key string
}{
	{feov.Conflict, "a gap holds one proposal at a time", "one_proposal_refused"},
	{feov.Conflict, "no proposed act moves a closed gap", "closed_gap_refused"},
}

// bystanderFiles is a bystander's filing of args on gapID: admitted, or refused by one of
// bystanderRefusals, which is counted under its key and reported; any other refusal is admitFresh's.
// It reports the refusal's key, "" where the filing landed or admitFresh took it.
func (r *runner) bystanderFiles(gapID string, args []string) string {
	out, err := r.exec(withJSON(args)...)
	if err == nil {
		return ""
	}
	if env, ok := refusalEnvelopeIn(out); ok {
		for _, b := range bystanderRefusals {
			if env.Code == string(b.code) && strings.Contains(env.Error, b.words) {
				r.countGate(b.key)
				return b.key
			}
		}
	}
	r.admitFresh(gapID, args...)
	return ""
}

// bystanderProposedOn reports whether a seat other than blue filed a proposal on gapID.
func (r *runner) bystanderProposedOn(gapID string) bool {
	evs, _, err := record.EventsOf(r.run(), recordpb.EventType_EVENT_TYPE_MOTION)
	if err != nil {
		r.noteActErr("the record's proposals could not be read: " + err.Error())
		return false
	}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok && m.GetDocket().GetGapId() == gapID && seatclass.RoleOfSeat(e.GetSeatId()) != "blue" {
			return true
		}
	}
	return false
}

// BYSTANDERSEED'S RUN REACHES THE BENCH ON ANOTHER LENS'S REGRADE, pinned deterministically: filed on a
// lens's gap after its counted exchange, it is docketed at the next plan, and the bench sides with the
// raiser — the gap's lens.
func TestABystandersRegradeIsLeftToTheBenchWhichSidesWithTheRaiser(t *testing.T) {
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
	r.forceBystander = true
	const sentence = "A § fuzz sentence to anchor findings."
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	r.chairPlans(t)
	lens, bystander := fuzzLensSeats[0], fuzzLensSeats[1]
	r.dispatchSits("lens", lens)
	r.mustExec(t, "--json", "--seat-id", lens, "show", "report")
	r.mustExec(t, "--json", "--seat-id", lens, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	r.dispatchReturns(lens, false)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	gapID := open[0]
	r.noteMint(lens, gapID)
	r.exchangeOn(t)

	r.dispatchSits("lens", bystander)
	if got := r.bystanderGaps(bystander); !slices.Contains(got, gapID) {
		t.Fatalf("%s may propose on %v, want %s among them", bystander, got, gapID)
	}
	if r.bystanderProposal(bystander, "regrade", gapID) != gapID || r.actErr != "" {
		t.Fatalf("%s's regrade on %s: %s", bystander, gapID, r.actErr)
	}
	r.dispatchReturns(bystander, false)
	if ms := r.unruledDocketMotions(gapID); len(ms) != 1 {
		t.Fatalf("proposals standing on %s after the bystander's filing = %v, want one", gapID, ms)
	}

	r.partiesClose(t, lens)
	r.dispatchSits("bench", "judge", "docket")
	r.benchDisposes("judge", gapID, "repaired", "--reason", "fuzz: the grades as they stand hold")
	if r.actErr != "" {
		t.Fatalf("the bench's ruling: %s", r.actErr)
	}
	if got := r.docketTally()["answer:proposal:raiser"]; got != 1 {
		t.Errorf("answer:proposal:raiser = %d, want the bench's one answer for the raiser; tally %v", got, r.docketTally())
	}
	r.dispatchReturns("judge", r.envelopeEnds("judge"))
}
