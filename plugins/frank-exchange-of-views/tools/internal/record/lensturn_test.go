package record

import (
	"fmt"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// m36Archive is universe-m36: 540 events, VERIFIED at epoch 8, 51 sittings of which 33 were lenses'.
const m36Archive = "2026-10-29_is-91-prime-m36.tar.gz"

// mintedOn is a stage at its second chair sitting: the lens sat in epoch 1 and minted G1 there, and
// nobody has acted on G1 since.
func mintedOn(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high").
		register("red-chair")
}

// blueSitsOn is the chair engaging blue alone on gap and blue sitting once; answer, when set, is its
// edit answering the gap, and otherwise it records nothing on it — a null turn.
func (b *stage) blueSitsOn(gap string, answer bool) *stage {
	b.dispatch(2, "blue-respond", gap)
	if !answer {
		return b.sitClosed("blue-respond")
	}
	agent := fmt.Sprintf("blue-answers-%s-%d", gap, b.n)
	return b.registerAs("blue-respond", agent).edit(gap, "old", "new").stop(agent)
}

// THE LENS'S TURN ON A GAP COMES WITH BLUE'S ACT (ruled by gblock 2026-10-30) — universe-m36's
// sequence. The lens mints G1 in epoch 1. Epoch 2's plan readies blue on G1 and not the lens, whose
// sitting first would audit an answer that does not exist; the reason says why. Blue answers, and
// that answer completes G1's first exchange — the mint was the lens's opening turn. Epoch 3's plan
// readies the lens on G1 to audit the answer, and blue with it.
func TestALensIsNotReadyOnAGapBlueHasNotActedOn(t *testing.T) {
	st := mintedOn(t)
	plan, err := PlanDispatch(st.seed())
	if err != nil {
		t.Fatal(err)
	}
	if _, lens := partyOf(plan, evLens); lens || !slices.Equal(gapsOf(plan, "blue-respond"), []string{"G1"}) {
		t.Fatalf("epoch 2: blue alone is ready on the fresh G1: parties %+v, why %q", plan.Parties, plan.Why)
	}
	if !whyMentions(plan, "G1: open, material, 0 exchange(s) (0 stalled) — below its limits; blue is ready on it, and "+evLens+" is not until blue has acted on it since the lens's last act there") {
		t.Errorf("epoch 2: the reason does not say why the lens waits: %q", plan.Why)
	}

	more := st.more()
	run := more.blueSitsOn("G1", true).register("red-chair").seed()
	after, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(gapsOf(after, evLens), []string{"G1"}) || !slices.Equal(gapsOf(after, "blue-respond"), []string{"G1"}) {
		t.Fatalf("epoch 3: blue answered G1, so the lens re-audits it, and blue answers again: parties %+v, why %q", after.Parties, after.Why)
	}
	x, err := Exchanges(run, DefaultParams)
	if err != nil || x["G1"].Exchanges != 1 || x["G1"].Stalled != 0 {
		t.Errorf("blue's answer completes G1's first exchange, opened by the mint, and moves it: %+v (%v)", x["G1"], err)
	}
}

// A GAP BLUE NEVER ACTS ON STILL REACHES THE BENCH. The lens's turn never comes, so the lens is never
// readied on G1; each of blue's null turns completes an exchange whose lens turn passed, and at k
// (2) stalled exchanges G1 is at impasse and docketed, its two parties readied for their closings.
func TestAGapBlueNeverAnswersReachesTheBenchWithoutTheLens(t *testing.T) {
	st := mintedOn(t)
	st.seed()
	more := st.more()
	more.blueSitsOn("G1", false).register("red-chair")
	once := more.seed()
	plan, err := PlanDispatch(once)
	if err != nil {
		t.Fatal(err)
	}
	if _, lens := partyOf(plan, evLens); lens || !slices.Equal(gapsOf(plan, "blue-respond"), []string{"G1"}) || len(plan.ToFile) != 0 {
		t.Fatalf("one null turn: blue alone is ready again, nothing docketed: parties %+v, to file %v, why %q", plan.Parties, plan.ToFile, plan.Why)
	}
	if x, _ := Exchanges(once, DefaultParams); x["G1"].Exchanges != 1 || x["G1"].Stalled != 1 {
		t.Errorf("one null turn is one stalled exchange: %+v", x["G1"])
	}
	again := more.more()
	twice := again.blueSitsOn("G1", false).register("red-chair").seed()
	docketed, err := PlanDispatch(twice)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(docketed.ToFile, []string{"G1"}) || slices.Contains(gapsOf(docketed, evLens), "G1") || slices.Contains(gapsOf(docketed, "blue-respond"), "G1") {
		t.Errorf("two null turns stall G1 to impasse, and it is docketed rather than debated: to file %v, parties %+v, why %q", docketed.ToFile, docketed.Parties, docketed.Why)
	}
}

// lensReadiedBeforeBlue is each lens engaged on a gap blue has not acted on since that lens's last act
// there and since the gap's latest remand ruling, in a plan folded from evs: "seat:gap" rows.
func lensReadiedBeforeBlue(evs []*Event, win WindowIndex, remands map[string][]remandRow, plan Plan) []string {
	ids := win.IDs(evs)
	acts, mintedAt := gapActsOf(evs, ids), mintPlaces(evs, ids)
	var out []string
	for _, p := range plan.Parties {
		if seatclass.RoleOfSeat(p.SeatID) != "lens" {
			continue
		}
		for _, g := range p.GapIDs {
			if !lensTurnOn(acts[g], p.SeatID, mintedAt[g], latestBefore(remandPlaces(remands[g]), ids[len(ids)-1]), ids[len(ids)-1]) {
				out = append(out, p.SeatID+":"+g)
			}
		}
	}
	return out
}

// UNIVERSE-M36, REPLAYED. Every plan its chair stored is folded again on the record as it stood
// before that plan's dispatch rows. Three plans — events 191, 317 and 430 — engaged seven lenses on
// gaps minted the epoch before, which blue had not answered: each lens sat first, recorded a
// keep-the-grade regrade ("Blue has not edited the sentence since this gap was minted") and nothing
// else readied it. Folded now, each of the three readies blue alone; no plan engages a lens on a gap
// blue has not acted on since the lens's last act; and every exchange count the run's plans relayed
// is the count the fold gives now — the mint opens the exchange the lens's no-op sitting used to.
func TestM36LensesWaitForBlue(t *testing.T) {
	whole := archivedEvents(t, m36Archive)
	stored, err := StoredPlans(whole)
	if err != nil {
		t.Fatal(err)
	}
	blueAlone := map[int]bool{191: true, 317: true, 430: true}
	lensSittings := map[string]int{} // stored, replayed
	for _, sp := range stored {
		start := sp.At
		for start > 0 {
			if _, ok := recordpb.BodyAs[*recordpb.Dispatch](whole[start-1]); !ok {
				break
			}
			start--
		}
		run := archiveAt(t, m36Archive, whole, start)
		plan, err := PlanDispatch(run)
		if err != nil {
			t.Fatal(err)
		}
		var evs []*Event
		var win WindowIndex
		var remands map[string][]remandRow
		if err := readSnapshot(run, func(q recordsql.Querier) error {
			var err error
			if evs, win, err = eventsAt(q); err != nil {
				return err
			}
			remands, err = remandRulingsOf(q)
			return err
		}); err != nil {
			t.Fatal(err)
		}
		at := sp.At + 1
		if early := lensReadiedBeforeBlue(evs, win, remands, plan); len(early) > 0 {
			t.Errorf("plan at event %d engages %v on a gap blue has not acted on since the lens's last act", at, early)
		}
		lenses := func(p Plan) (n int) {
			for _, q := range p.Parties {
				if seatclass.RoleOfSeat(q.SeatID) == "lens" {
					n++
				}
			}
			return n
		}
		lensSittings["stored"] += lenses(sp.Plan)
		lensSittings["replayed"] += lenses(plan)
		if blueAlone[at] {
			if len(plan.Parties) != 1 || !slices.Equal(gapsOf(plan, "blue-respond"), gapsOf(sp.Plan, "blue-respond")) {
				t.Errorf("plan at event %d readies %+v, want blue alone on the gaps it was engaged on", at, plan.Parties)
			}
		}
		want := map[string]ExchangeCount{}
		for _, x := range sp.Plan.Exchanges {
			want[x.GapID] = x
		}
		for _, x := range plan.Exchanges {
			if w, ok := want[x.GapID]; ok && w != x {
				t.Errorf("plan at event %d counts %+v on %s, the run relayed %+v", at, x, x.GapID, w)
			}
		}
	}
	if got := lensSittings["stored"] - lensSittings["replayed"]; got != 7 {
		t.Errorf("the replayed plans dispatch %d lens sitting(s) fewer than m36's (%d stored, %d replayed), want the 7 no-op sittings", got, lensSittings["stored"], lensSittings["replayed"])
	}

	// CAPTURE NAMES THEM: the run's own dispatch rows hold the nine early engagements, in the seven
	// sittings above.
	early, err := EarlyLensDispatches(archiveAt(t, m36Archive, whole, len(whole)))
	if err != nil {
		t.Fatal(err)
	}
	sittings := map[string]bool{}
	for _, e := range early {
		sittings[fmt.Sprintf("%s@%d", e.SeatID, e.At)] = true
	}
	if len(early) != 9 || len(sittings) != 7 {
		t.Errorf("m36's record holds %d early lens engagement(s) in %d dispatch(es), want 9 in 7: %+v", len(early), len(sittings), early)
	}
}

// A RUN DISPATCHED BY THE RULE HOLDS NO EARLY ENGAGEMENT, and one that engaged the lens on its fresh
// gap the epoch after its mint holds one: the row capture's lens-turn audit fails on.
func TestEarlyLensDispatchesAreTheEngagementsBeforeBlue(t *testing.T) {
	st := mintedOn(t)
	st.seed()
	more := st.more()
	ruled := more.blueSitsOn("G1", true).register("red-chair").dispatch(2, evLens, "G1").seed()
	if early, err := EarlyLensDispatches(ruled); err != nil || len(early) != 0 {
		t.Errorf("the lens engaged on G1 after blue's answer is no early engagement: %+v (%v)", early, err)
	}
	old := mintedOn(t).dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").seed()
	early, err := EarlyLensDispatches(old)
	if err != nil || len(early) != 1 || early[0].SeatID != evLens || early[0].GapID != "G1" {
		t.Errorf("the lens engaged on G1 the epoch after its mint, before blue: %+v (%v), want that one row", early, err)
	}
}

// THE BENCH'S REMAND IS THE FLOOR OF THE LENS'S TURN. G1 reached impasse with blue's proof the last
// act on it, after the lens's kept grade; the bench heard that answer and remanded G1. Blue's answer
// before the ruling does not give the lens its turn: the plan readies blue alone on the gap.
func TestARemandRulingIsTheFloorOfTheLensTurn(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.actingExchangeOn("G1").exchangeOn("G1")
	run := st.docketAndRemand("M1", "G1").register("red-chair").seed()
	if x, err := Exchanges(run, DefaultParams); err != nil || x["G1"].Remand != remandOwed {
		t.Fatalf("setup: G1 must be remanded at impasse with its exchange owed: %+v (%v)", x["G1"], err)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Errorf("blue's answer before the ruling gives the lens no turn on the remanded G1: parties %+v, why %q", plan.Parties, plan.Why)
	}
}
