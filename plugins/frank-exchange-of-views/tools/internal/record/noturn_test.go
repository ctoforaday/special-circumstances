package record

import (
	"slices"
	"testing"
)

// A SITTING THAT PRODUCED NO ADMITTED ENVELOPE IS NO TURN (ruled by gblock 2026-11-11). Blue is
// engaged on G1 and sits once without acting on it. Returning its envelope, that sitting is a null
// turn: the exchange ends and stalls. Ending without one — a failed sitting, how a usage-limit stop
// reaches the record — or stopped at the call limit, it is no turn at all: no exchange, no stall,
// nothing unresolved, and blue stays engaged.
func TestABlueSittingWithNoAdmittedEnvelopeStallsNoGap(t *testing.T) {
	sat := func(t *testing.T, end func(b *stage) *stage) *stage {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").stop("lens-1").
			register("red-chair").dispatch(1, "blue-respond", "G1").registerAs("blue-respond", "blue-1")
		return end(b).register("red-chair")
	}
	for _, c := range []struct {
		name               string
		end                func(b *stage) *stage
		exchanges, stalled int
	}{
		{"returned its envelope: a null turn", func(b *stage) *stage { return b.stop("blue-1") }, 1, 1},
		{"ended without its envelope: a failed sitting", func(b *stage) *stage { return b.fails("blue-1") }, 0, 0},
		// Its envelope admitted all the same, so only the limit sets it apart from the null turn.
		{"stopped at the call limit", func(b *stage) *stage { return b.limit("blue-respond", 1).stop("blue-1") }, 0, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := sat(t, c.end).seed()
			x, err := Exchanges(run, DefaultParams)
			if err != nil {
				t.Fatal(err)
			}
			if g := x["G1"]; g == nil || g.Exchanges != c.exchanges || g.Stalled != c.stalled || g.Unresolved != 0 {
				t.Fatalf("G1 = %+v, want %d exchange(s), %d stalled, nothing unresolved", g, c.exchanges, c.stalled)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
				t.Errorf("blue is not engaged on G1 after its sitting: parties %+v, why %q", plan.Parties, plan.Why)
			}
		})
	}
}

// THE LENS'S SITTING IS NO TURN EITHER. Blue answered G1 and the lens is readied to audit it; a lens
// sitting that returns its envelope in silence ends the exchange and takes the turn
// (TestASilentSittingTakesTheLensTurn). One that fails, or is stopped at the call limit, audited
// nothing: the exchange stays open, and the lens is readied on the answer again.
func TestALensSittingWithNoAdmittedEnvelopeLeavesTheTurnStanding(t *testing.T) {
	for _, c := range []struct {
		name string
		end  func(b *stage) *stage
	}{
		{"a failed sitting", func(b *stage) *stage { return b.fails("lens-fails") }},
		{"stopped at the call limit", func(b *stage) *stage { return b.limit(evLens, 2).stop("lens-fails") }},
	} {
		t.Run(c.name, func(t *testing.T) {
			st := mintedOn(t).blueSitsOn("G1", true)
			st.register("red-chair").dispatch(2, evLens, "G1").registerAs(evLens, "lens-fails").served(evLens, st.head())
			run := c.end(st).register("red-chair").seed()
			x, err := Exchanges(run, DefaultParams)
			if err != nil {
				t.Fatal(err)
			}
			if g := x["G1"]; g == nil || g.Exchanges != 0 || g.Unresolved != 0 {
				t.Fatalf("G1 = %+v, want no exchange ended and nothing unresolved", g)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(gapsOf(plan, evLens), "G1") {
				t.Errorf("the lens's sitting audited nothing, and it is not readied on blue's answer again: parties %+v, why %q", plan.Parties, plan.Why)
			}
		})
	}
}
