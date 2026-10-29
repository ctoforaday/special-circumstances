package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// limit is the sitting_limit the hook-spawned writer records when it stops the seat's n-th sitting
// at the per-sitting tool-call limit.
func (b *stage) limit(seat string, n int) *stage {
	return b.add(HarnessSeat, &recordpb.SittingLimit{SeatId: proto.String(seat), Sitting: proto.Int32(int32(n)), Limit: proto.Int32(150)})
}

// sitHooked is one sitting of a lens as the shipped workflow runs it: a dispatch at pin, the hook's
// bracket for agent, and — when stopped — the agent's stop.
func (b *stage) sitHooked(pin int64, lens, agent string, stopped bool) *stage {
	b.dispatch(pin, lens).bracket(lens, agent).served(lens, pin)
	if stopped {
		b.stop(agent)
	}
	return b
}

// A SITTING THE LENS DID NOT SIT THROUGH IS NEITHER SAT NOR BARREN (satThrough). Each row is a lens
// with two barren sittings it sat through at the head, then a third sitting at the head that ends
// some way, then a head move: sat through, the third retires it; stopped at the call limit or never
// closed, it is not counted — the lens is not retired, the moved head readies it at the pin of its
// second sitting, and the reason says which sitting did not count and why.
func TestASittingTheLensDidNotSitThroughCountsNeitherAsSatNorBarren(t *testing.T) {
	for _, c := range []struct {
		name    string
		third   func(b *stage)
		counted bool
		why     string
	}{
		{"sat through: stopped by its agent", func(b *stage) { b.sitHooked(b.at(), evLens, "a3", true) }, true, ""},
		{"stopped at the call limit", func(b *stage) {
			b.dispatch(b.at(), evLens).bracket(evLens, "a3").limit(evLens, 3).stop("a3")
		}, false, "was stopped at the per-sitting tool-call limit"},
		// Never closed: no stop, and the lens has opened a later sitting — here one at a later head
		// that does not end either, so the latest is still open.
		{"never closed", func(b *stage) {
			b.sitHooked(b.at(), evLens, "a3", false).ingest()
			b.sitHooked(b.at(), evLens, "a4", false)
		}, false, "has not closed yet"},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := retBoard(t).sitHooked(2, evLens, "a1", true).ingest()
			pin := b.at() // the head the second sitting is dispatched at
			b.sitHooked(pin, evLens, "a2", true)
			c.third(b)
			b.ingest()
			f := lensStateOf(t, b.register("red-chair").seed())
			if c.counted {
				if !f.retired() || f.ready || f.barren != retiresAfter {
					t.Fatalf("a third barren sitting it sat through retires the lens: %+v", f)
				}
				return
			}
			if f.retired() || !f.ready || f.barren != 2 || f.sittings != 2 || f.pin != pin {
				t.Fatalf("a sitting it did not sit through is neither sat nor barren — want 2 sittings, barren 2, pin %d, ready: %+v", pin, f)
			}
			if !strings.Contains(f.why, c.why) || !strings.Contains(f.why, "counts neither as sat nor toward retirement") {
				t.Errorf("the reason does not say which sitting did not count and why (%q): %s", c.why, f.why)
			}
		})
	}
}

// AT THE FINAL HEAD, a lens cut off at the limit on its only sitting has not audited the report: it
// is ready, the PASS gate refuses on it, and the chair's work list holds the same blocker — the
// plan, the gate and the list read one fold (passBlockersOf).
func TestALensStoppedAtTheLimitIsStillReadyForThePlanTheGateAndTheList(t *testing.T) {
	run := retBoard(t).dispatch(2, evLens).bracket(evLens, "a1").limit(evLens, 1).stop("a1").
		register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !lensReady(plan) || plan.PassPermitted || !strings.Contains(whyOf(plan), "has sat through no sitting") {
		t.Fatalf("the plan does not ready a lens whose only sitting was stopped at the limit: %+v", plan)
	}
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(blockers, func(b Blocker) bool { return b.Kind == BlockerLensReady && b.Subject == evLens }) {
		t.Errorf("the PASS gate does not refuse on the lens the plan readies: %+v", blockers)
	}
	s := sittingOfRunT(t, run, "chair", chairSeat)
	if !slices.ContainsFunc(s.Open, func(i Item) bool { return i.Blocks && strings.Contains(i.What, "lens "+evLens+" is ready") }) {
		t.Errorf("the chair's work list does not hold the blocker the gate refuses on: %+v", s.Open)
	}
}

// THE SEAT SITTING NOW IS NOT FAILED. A lens reading its own work list inside an open sitting owes
// nothing for the sitting it is in (sittingFor: its opening satisfied the dispatch), is told nothing
// was cut short, and its sitting reads as not closed yet — never as one that failed.
func TestTheLensSittingNowIsNotReadAsFailed(t *testing.T) {
	b := retBoard(t).sitHooked(2, evLens, "a1", false)
	run := b.seed()
	s := sittingOfRunT(t, run, "lens", evLens)
	if owed := owedItems(s); len(owed) != 0 {
		t.Errorf("the lens in an open sitting owes %+v", owed)
	}
	for _, i := range s.Open {
		if strings.Contains(i.What, "never closed") || strings.Contains(i.What, "call limit") {
			t.Errorf("the lens sitting now is told its sitting failed: %q", i.What)
		}
	}
	m := loadedT(t, b.evs...)
	ids := m.At.IDs(m.Events)
	_, openings := dispatchLedger(m.Events, ids, m.At)
	ends := satThroughOf(m.Events, ids, m.At, openings)
	o := openings[evLens][0]
	if ends.sat(evLens, o) || ends.short[o] != "has not closed yet" {
		t.Errorf("the open sitting reads %q, sat through %v — want not closed yet, and not sat through", ends.short[o], ends.sat(evLens, o))
	}
}

// AN OPENING WITH NO AGENT IS SAT THROUGH AT ITS OPENING: no stop can join it (sittingCloser), and a
// headless seat's register is that case.
func TestASittingNoStopCanJoinIsSatThroughAtItsOpening(t *testing.T) {
	f := lensStateOf(t, retBoard(t).sit(2, evLens).register("red-chair").seed())
	if f.sittings != 1 || f.pin != 2 || f.ready {
		t.Fatalf("a register with no agent is a sitting sat through: %+v", f)
	}
}

// WHAT IS NEW ON ITS GAP IS OWED UNTIL IT SITS THROUGH A SITTING (lensOwesOf). Blue answers the
// lens's gap; a sitting stopped at the limit after it leaves the answer owed, and the lens is not
// retired by its barren sittings; a sitting it sits through ends the debt.
func TestADebtOutlivesASittingStoppedAtTheLimit(t *testing.T) {
	board := func(t *testing.T, through bool) lensFold {
		b := retBoard(t)
		b.add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
		for range retiresAfter {
			b.sit(2, evLens)
		}
		blueAnswersGap(b, "G1")
		b.dispatch(2, evLens, "G1").bracket(evLens, "a9")
		if !through {
			b.limit(evLens, retiresAfter+1)
		}
		b.stop("a9").ingest()
		return lensStateOf(t, b.register("red-chair").seed())
	}
	if f := board(t, false); !slices.Contains(f.owes, "blue's answer to G1") || f.retired() || !f.ready {
		t.Errorf("a sitting stopped at the limit ended the debt it was dispatched for: %+v", f)
	}
	if f := board(t, true); slices.Contains(f.owes, "blue's answer to G1") || !f.retired() {
		t.Errorf("a sitting it sat through did not end the debt: %+v", f)
	}
}

// ITS LAST SITTING IS THE LAST IT SAT THROUGH (lastSittingBefore). A lens whose only earlier sitting
// was stopped at the limit is told to audit in full, as on a first look; one that sat through it is
// told where it last sat.
func TestALensLastSittingSkipsASittingStoppedAtTheLimit(t *testing.T) {
	for _, limited := range []bool{true, false} {
		b := retBoard(t).dispatch(2, evLens).bracket(evLens, "a1").served(evLens, 2)
		if limited {
			b.limit(evLens, 1)
		}
		b.stop("a1").ingest()
		b.dispatch(b.at(), evLens).bracket(evLens, "a2")
		ls := sittingOfRunT(t, b.seed(), "lens", evLens).LastRead
		want := map[bool]string{true: "first", false: "behind"}[limited]
		if ls == nil || ls.Kind != want {
			t.Errorf("limited=%v: last_read = %+v, want kind %s", limited, ls, want)
		}
	}
}

// A CONTRADICTION IS READIED ONCE PER CAUSE, AND A SITTING STOPPED AT THE LIMIT IS NOT THE ONE: the
// lens that read it is readied again.
func TestAContradictionIsReadiedAgainAfterASittingStoppedAtTheLimit(t *testing.T) {
	for _, limited := range []bool{true, false} {
		b := retBoard(t).sit(2, evLens).add(evLens, contradictingVerify("the sky is green")).
			dispatch(2, evLens).bracket(evLens, "a1")
		if limited {
			b.limit(evLens, 2)
		}
		plan := planOf(t, b.stop("a1").register("red-chair"))
		if lensReady(plan) != limited {
			t.Errorf("limited=%v: lens ready = %v, why %q", limited, lensReady(plan), plan.Why)
		}
	}
}

// A FRESH MATERIAL MINT IS ON THE RECORD WHETHER OR NOT ITS SITTING WAS SAT THROUGH, and it resets
// the count: two barren sittings, one stopped at the limit after minting fresh material, and one
// more barren sitting leave the lens one barren sitting in, not retired.
func TestAFreshMintInASittingStoppedAtTheLimitResetsTheCount(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens)
	b.dispatch(2, evLens).bracket(evLens, "a3").add(evLens, freshHigh("G1")).limit(evLens, 3).stop("a3")
	b.sit(2, evLens).ingest()
	if f := lensStateOf(t, b.register("red-chair").seed()); f.retired() || f.barren != 1 || f.sittings != 3 {
		t.Fatalf("the fresh mint resets the count and the stopped sitting is not sat: want barren 1 of 3 sittings: %+v", f)
	}
}

// EACH END HAS ITS OWN ANSWER. On one record: a sitting its agent stopped is sat through; one the
// hook stopped at the call limit is not, though its agent's stop is on the record; one with no stop
// that the lens has sat again since never closed; and the latest, with no stop, has not closed yet.
func TestEachEndOfASittingHasItsOwnAnswer(t *testing.T) {
	b := retBoard(t).sitHooked(2, evLens, "a1", true)
	b.dispatch(2, evLens).bracket(evLens, "a2").limit(evLens, 2).stop("a2")
	b.sitHooked(2, evLens, "a3", false).sitHooked(2, evLens, "a4", false)
	m := loadedT(t, b.evs...)
	ids := m.At.IDs(m.Events)
	_, openings := dispatchLedger(m.Events, ids, m.At)
	ends := satThroughOf(m.Events, ids, m.At, openings)
	os := openings[evLens]
	if len(os) != 4 {
		t.Fatalf("fixture: %d sittings of the lens, want 4", len(os))
	}
	want := []string{"", "was stopped at the per-sitting tool-call limit", "never closed — its agent's stop is not on the record", "has not closed yet"}
	for k, o := range os {
		if got := ends.short[o]; got != want[k] || ends.sat(evLens, o) != (want[k] == "") {
			t.Errorf("sitting %d: short %q, sat through %v — want %q", k+1, got, ends.sat(evLens, o), want[k])
		}
	}
}
