package record

import (
	"fmt"
	"slices"
	"strings"
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
// that answer opens G1's first exchange — the mint was the lens's opening turn. Epoch 3's plan
// readies the lens on G1 to audit the answer, and blue with it; the lens's audit ends the exchange
// (ruled by gblock 2026-10-31).
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
	if err != nil || x["G1"].Exchanges != 0 {
		t.Errorf("no exchange is complete until the lens has audited blue's answer: %+v (%v)", x["G1"], err)
	}
	audited := lensAuditsG1(t, more)
	if x, err := Exchanges(audited, DefaultParams); err != nil || x["G1"].Exchanges != 1 || x["G1"].Stalled != 0 {
		t.Errorf("the lens's audit ends G1's first exchange, opened by the mint, and blue's answer moved it: %+v (%v)", x["G1"], err)
	}
}

// lensAuditsG1 is the lens sitting on G1, and blue not, on a stage appended to and not yet seeded.
func lensAuditsG1(t *testing.T, st *stage) Run {
	t.Helper()
	more := st.more()
	agent := fmt.Sprintf("lens-audits-%d", more.n)
	return more.dispatch(2, evLens, "G1").registerAs(evLens, agent).served(evLens, 2).stop(agent).register("red-chair").seed()
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
// there and since the gap's latest remand ruling, in a plan folded from evs: "seat:gap" rows. The turns
// are the dispatch's own (lensTurnActsOf): an edit that changed the gap's sentence is the lens's turn.
func lensReadiedBeforeBlue(evs []*Event, win WindowIndex, remands map[string][]remandRow, plan Plan) []string {
	ids := win.IDs(evs)
	minted := mintedBy(evs)
	acts, mintedAt := lensTurnActsOf(evs, ids, gapActsOf(evs, ids, win), gapSittingsOf(evs, ids, win, AfterTheRun, minted), minted), mintPlaces(evs, ids)
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
// else readied it. Folded now, none of the three engages a lens on a gap the stored plan engaged it on,
// blue is ready on every gap it was engaged on, and no plan engages a lens on a gap blue has not acted
// on since the lens's last act.
//
// REPORT_VOICE IS ALWAYS MATERIAL (gblock, 2026-11-05), so m36's low voice gaps hold the board now:
// blue is ready on them too, and at event 430 the voice lens is engaged on G-380a3523, whose sentence
// an edit changed — the lens's turn. That one sitting is why the replay saves six lens sittings, not
// the seven no-op sittings it saved while the class went by grade.
//
// M36'S INQUIRIES WERE SETTLED AND NEVER CHECKED, so the replay also readies the logic lens, on no gap,
// for their check (inquirycheck.go). That readiness engages no lens on a gap, and is left out of both
// counts. So is the voice lens readied, on no gap, for the text the tool composes from a fate or a mint
// it had not sat through since (composedsource.go).
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
				if seatclass.RoleOfSeat(q.SeatID) == "lens" && !checksAnInquiry(p, q) && !auditsComposedText(p, q) {
					n++
				}
			}
			return n
		}
		lensSittings["stored"] += lenses(sp.Plan)
		lensSittings["replayed"] += lenses(plan)
		if blueAlone[at] {
			for _, q := range sp.Plan.Parties {
				if seatclass.RoleOfSeat(q.SeatID) != "lens" {
					continue
				}
				for _, g := range q.GapIDs {
					if slices.Contains(gapsOf(plan, q.SeatID), g) {
						t.Errorf("plan at event %d engages %s on %s again, before blue's answer: %+v", at, q.SeatID, g, plan.Parties)
					}
				}
			}
			for _, g := range gapsOf(sp.Plan, "blue-respond") {
				if !slices.Contains(gapsOf(plan, "blue-respond"), g) {
					t.Errorf("plan at event %d does not ready blue on %s, which the stored plan engaged it on: %+v", at, g, plan.Parties)
				}
			}
		}
	}
	if got := lensSittings["stored"] - lensSittings["replayed"]; got != 6 {
		t.Errorf("the replayed plans dispatch %d lens sitting(s) fewer than m36's (%d stored, %d replayed), want the 7 no-op sittings less voice's turn at event 430", got, lensSittings["stored"], lensSittings["replayed"])
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

// THE BENCH'S REMAND IS THE FLOOR OF THE LENS'S TURN. G1 reached impasse, and blue's computation is the
// last act on it, after the lens's last sitting; the bench heard that answer and remanded G1. Blue's answer
// before the ruling does not give the lens its turn: the plan readies blue alone on the gap.
func TestARemandRulingIsTheFloorOfTheLensTurn(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.exchangeOn("G1").exchangeOn("G1").registerAs("blue-respond", "blue-before-ruling").blueComputes("G1").stop("blue-before-ruling")
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

// checksAnInquiry reports whether party q is the logic lens readied, on no gap, for an inquiry's check
// and for nothing else: its line of the lens gate does not ready it to audit the report.
func checksAnInquiry(p Plan, q Party) bool {
	return q.SeatID == inquiryOwner && len(q.GapIDs) == 0 &&
		slices.ContainsFunc(p.Why, func(w string) bool { return strings.HasPrefix(w, inquiryOwner+" owns the check of inquiry") }) &&
		!slices.ContainsFunc(p.Why, func(w string) bool {
			return strings.HasPrefix(w, inquiryOwner+":") && strings.Contains(w, "audits the report")
		})
}

// auditsComposedText reports whether party q is the voice lens readied, on no gap, for the text the
// tool composes from the record and for nothing else: its line of the lens gate readies it for that
// text alone (lensFold.readiness).
func auditsComposedText(p Plan, q Party) bool {
	return q.SeatID == composedSourceAuditor && len(q.GapIDs) == 0 &&
		slices.ContainsFunc(p.Why, func(w string) bool {
			return strings.HasPrefix(w, composedSourceAuditor+":") && strings.Contains(w, "audits the text the tool composes")
		})
}
