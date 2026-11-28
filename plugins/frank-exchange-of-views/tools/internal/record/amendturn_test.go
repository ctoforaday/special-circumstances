package record

import (
	"fmt"
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE AMEND IS THE AUTHOR'S TURN (T-B45's amend rows, §III.B3): the raiser's amend of G is its half of
// the exchange, counted as its moving regrade is (one exchange when its sitting ends, and movement), and
// the next round's turn is the author's; the bench's amend is no party's act — no exchange, no
// movement — and opens a round on G with the author's turn, as a remand ruling does; A's amend of G1
// naming G2, a gap about G1, is A's answer on G2: G2's raiser's turn comes, and it moves G2.
func TestAnAmendIsTheAuthorsTurn(t *testing.T) {
	amend := func(gap, answers string) *recordpb.Amend {
		a := &recordpb.Amend{GapId: proto.String(gap), Problem: proto.String(fmt.Sprintf("restated %s", gap)), Basis: proto.String("narrower")}
		if answers != "" {
			a.Answers = proto.String(answers)
		}
		return a
	}
	exchanges := func(t *testing.T, b *stage) map[string]*GapExchanges {
		t.Helper()
		x, err := Exchanges(b.register("red-chair").seed(), DefaultParams)
		if err != nil {
			t.Fatal(err)
		}
		return x
	}

	// blueActsStill is G1 minted and blue's sitting on it acting without moving it (an edit that
	// changes nothing): the lens's turn has come, and the round is open.
	blueActsStill := func(t *testing.T) *stage {
		st := mintedOn(t).dispatch(2, "blue-respond", "G1")
		agent := fmt.Sprintf("blue-acts-%d", st.n)
		return st.registerAs("blue-respond", agent).edit("G1", "same", "same").stop(agent)
	}
	// lensSits is the chair engaging the lens on G1 and the lens sitting, with the act given.
	lensSits := func(b *stage, act proto.Message) *stage {
		b.register("red-chair").dispatch(2, evLens, "G1")
		agent := fmt.Sprintf("%s-sits-%d", evLens, b.n)
		b.registerAs(evLens, agent).served(evLens, b.head())
		if act != nil {
			b.add(evLens, act)
		}
		return b.stop(agent)
	}

	t.Run("the raiser's amend is movement, and one exchange when its sitting ends", func(t *testing.T) {
		still := exchanges(t, lensSits(blueActsStill(t), nil))["G1"]
		if still == nil || still.Exchanges != 1 || still.Stalled != 1 {
			t.Fatalf("the fixture: the lens's null audit after blue's still act ends one stalled round: %+v", still)
		}
		amended := func() *stage { return lensSits(blueActsStill(t), amend("G1", "")) }
		g := exchanges(t, amended())["G1"]
		if g == nil || g.Exchanges != 1 || g.Stalled != 0 || len(g.rounds) != 1 || !g.rounds[0].moved || !g.rounds[0].audited {
			t.Errorf("the raiser's amend in its auditing sitting is not counted as one exchange that moved G1: %+v", g)
		}
		// THE NEXT ROUND'S TURN IS THE AUTHOR'S: blue is readied on G1, the lens is not.
		plan := planOfT(t, amended().register("red-chair").seed())
		if !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
			t.Errorf("after the raiser's amend of G1 blue is not readied on it: %+v", plan.Parties)
		}
		if slices.Contains(gapsOf(plan, evLens), "G1") {
			t.Errorf("after its own amend of G1 the raiser is readied on it before blue answers: %+v", plan.Parties)
		}
	})

	// THE RAISER'S AMEND IS ITS ACT ON G (gapActsOf), so the lens has read G where it amended, even in a
	// sitting no stop has closed yet: blue's act before it is no longer news to the lens.
	t.Run("the raiser's amend is its act on G", func(t *testing.T) {
		st := blueActsStill(t)
		st.register("red-chair").dispatch(2, evLens, "G1").registerAs(evLens, "lens-mid").served(evLens, st.head())
		st.add(evLens, amend("G1", ""))
		at := st.at()
		g := exchanges(t, st)["G1"]
		if g == nil || g.lensRead != at || g.lensTurn {
			t.Errorf("the raiser's amend of G1 is not read as its act there: lensRead %d (want %d), lensTurn %v", g.lensRead, at, g.lensTurn)
		}
	})

	t.Run("an amend restating the current wording moves nothing", func(t *testing.T) {
		same := &recordpb.Amend{GapId: proto.String("G1"), Problem: proto.String("p"), Basis: proto.String("again")}
		g := exchanges(t, lensSits(blueActsStill(t), same))["G1"]
		if g == nil || g.Stalled != 1 {
			t.Errorf("an amend restating G1's own problem is read as movement: %+v", g)
		}
	})

	t.Run("the bench's amend is no exchange and no movement, and opens the author's round", func(t *testing.T) {
		// Blue acts again on G1 — the round that act opens has not been audited when the bench amends.
		base := func() *stage {
			b := lensSits(blueActsStill(t), nil)
			agent := fmt.Sprintf("blue-again-%d", b.n)
			return b.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", agent).edit("G1", "same", "same").stop(agent)
		}
		benched := func() *stage { return base().register("judge").add("judge", amend("G1", "")) }
		before := exchanges(t, base())["G1"]
		if before == nil || !before.pending.acted {
			t.Fatalf("the fixture: blue's second act is in the open round: %+v", before)
		}
		b := benched()
		at := b.at()
		g := exchanges(t, b)["G1"]
		if g == nil || g.Exchanges != before.Exchanges || g.Stalled != before.Stalled {
			t.Errorf("the bench's amend counted an exchange or moved the stall: before %+v, after %+v", before, g)
		}
		if g.pending.moved {
			t.Errorf("the bench's amend is read as movement: %+v", g.pending)
		}
		if g.pending.start != at || g.pending.acted {
			t.Errorf("the bench's amend does not open a round on G1 with the author's turn: the open round starts at %d (want %d), acted %v", g.pending.start, at, g.pending.acted)
		}
		if plan := planOfT(t, benched().register("red-chair").seed()); !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
			t.Errorf("after the bench's amend of G1 blue is not readied on it: %+v", plan.Parties)
		}
	})

	t.Run("A's amend naming G2 is G2's raiser's turn, and moves G2", func(t *testing.T) {
		const author, raiser = evLens, outsideLens
		amended := func() *stage {
			st := authorStage(t)
			agent := fmt.Sprintf("author-amends-%d", st.n)
			return st.register("red-chair").dispatch(2, author, "G2").registerAs(author, agent).add(author, amend("G1", "G2")).stop(agent)
		}
		g := exchanges(t, amended())["G2"]
		if g == nil || !g.lensTurn || !g.pending.acted || !g.pending.moved {
			t.Fatalf("after the author's amend of G1 naming G2, G2's raiser's turn has not come, or the round reads no moving act: %+v", g)
		}
		if plan := planOfT(t, amended().register("red-chair").seed()); !slices.Contains(gapsOf(plan, raiser), "G2") {
			t.Errorf("after the author's amend naming G2 its raiser is not readied on it: %+v", plan.Parties)
		}
	})
}
