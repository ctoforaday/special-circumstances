package fuzz

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// stagedLifecycleArms is a staged run the lifecycle arms stand on (lifecycleSetUp).
func stagedLifecycleArms(t *testing.T) *runner {
	t.Helper()
	r := stagedLifecycleRunner(t)
	if why := r.lifecycleSetUp("blue-respond"); why != "" {
		t.Fatalf("the lifecycle arms' setup: %s (%s)", why, r.actErr)
	}
	return r
}

// armDrives runs one lifecycle arm on a staged run and holds it to its count: the arm returns "", no act
// failed, and the count reads want.
func armDrives(t *testing.T, arm func(r *runner) string, count func(c lifecycleCounts) int, want int) *runner {
	t.Helper()
	r := stagedLifecycleArms(t)
	if why := arm(r); why != "" || r.actErr != "" {
		t.Fatalf("the arm failed: %s (%s)", why, r.actErr)
	}
	r.mu.Lock()
	got := count(r.lifecycle)
	r.mu.Unlock()
	if got != want {
		t.Errorf("the arm counted %d, want %d", got, want)
	}
	return r
}

// BLUE'S DRAWN CUT OF A CITED SENTENCE GOES THROUGH cutCitingFirst (citedCut): applied by a computation
// on another sentence, the cut is refused naming it (counted in citedCutsRefused, the lifecycle seed's
// cell), the computation is retired, and the cut lands taking the citation; not applied, the cut lands
// at once and nothing is refused.
func TestTheDrawnCitedCutGoesThroughCutCitingFirst(t *testing.T) {
	for _, applied := range []bool{true, false} {
		r := stagedLifecycleRunner(t)
		r.citedCut("blue-respond", applied)
		if r.actErr != "" {
			t.Fatalf("applied %v: the drawn cut failed: %s", applied, r.actErr)
		}
		want := 0
		if applied {
			want = 1
		}
		if r.lifecycle.citedCutsRefused != want || r.lifecycle.computationRetires != want {
			t.Errorf("applied %v: %d cut(s) refused and %d computation(s) retired, want %d each", applied, r.lifecycle.citedCutsRefused, r.lifecycle.computationRetires, want)
		}
		md, err := record.RenderRun(r.run())
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(md, "The drawn cut claim 1 holds") || !strings.Contains(md, "The drawn method claim 1 holds") {
			t.Errorf("applied %v: the cut claim stands, or its neighbour left with it:\n%s", applied, md)
		}
	}
}

// A CUT THAT MOVES A TOKEN ONTO THE NEIGHBOURING SENTENCE IS A REWRITE (lifecycleCutMovesToken): the
// token is carried and nothing leaves.
func TestACutMovingATokenIsARewrite(t *testing.T) {
	armDrives(t, func(r *runner) string {
		return r.lifecycleCutMovesToken("blue-respond", lifecycleClaims[19], lifecycleClaims[20])
	}, func(c lifecycleCounts) int { return c.cutMovedTokens }, 1)
}

// A PURE CUT OF A SENTENCE WHOSE CITATION STANDS ON A FAILED CHECK CLEARS THE CHECK, and the reader's gap
// then closes (lifecycleCutClearsCheck).
func TestAPureCutClearsTheFailedCheck(t *testing.T) {
	armDrives(t, func(r *runner) string {
		return r.lifecycleCutClearsCheck("blue-respond", citationOwner, lifecycleClaims[21])
	}, func(c lifecycleCounts) int { return c.cutClearedChecks }, 1)
}

// A PURE CUT ABUTTING A TOKEN REPLAYS AS WRITTEN, whether its quote takes the token or stops before it,
// and each token it took has one remove (lifecycleAbuttingCuts).
func TestAPureCutAbuttingATokenReplaysAsWritten(t *testing.T) {
	armDrives(t, func(r *runner) string {
		return r.lifecycleAbuttingCuts("blue-respond", lifecycleClaims[22], lifecycleClaims[23])
	}, func(c lifecycleCounts) int { return c.abuttingQuotedCuts + 10*c.abuttingUnquotedCuts }, 11)
}

// A CLOSED GAP'S OWN MARKER IS RETIRED (F14, lifecycleClosedMarkerRetire).
func TestAClosedGapsMarkerRetireLands(t *testing.T) {
	armDrives(t, func(r *runner) string {
		return r.lifecycleClosedMarkerRetire("blue-respond", citationOwner, lifecycleClaims[24])
	}, func(c lifecycleCounts) int { return c.closedMarkerRetires }, 1)
}

// A RETIRE OF AN ANCHOR OF EACH KIND KEEPS ITS SENTENCE AND THE HEAD (lifecycleKeptRetires): one per
// kind the anchor table holds.
func TestARetireOfEachKindKeepsItsSentenceAndTheHead(t *testing.T) {
	armDrives(t, func(r *runner) string {
		return r.lifecycleKeptRetires("blue-respond", citationOwner, lifecycleClaims[25:29])
	}, func(c lifecycleCounts) int { return c.keptRetiresHeadStill }, len(anchor.Kinds()))
}

// EVERY LIFECYCLE CELL IS A GATE THAT FAILS AT ZERO: a tally nothing drove reads zero on every gate, the
// sweep's loop over them (judgeSweep) failing each; a tally with one cell driven reads that cell's gate
// alone; and every field of the tally is one gate, so a cell added without its gate fails here.
func TestEveryLifecycleCellFailsAtZero(t *testing.T) {
	var zero lifecycleTally
	gates := zero.gates()
	fields := reflect.TypeOf(zero).NumField()
	if len(gates) != fields {
		t.Fatalf("%d lifecycle gates for %d tally cells: each cell is one gate", len(gates), fields)
	}
	for _, g := range gates {
		if g.n != 0 {
			t.Errorf("a tally nothing drove reads %d on %q", g.n, g.what)
		}
	}
	for i := range fields {
		var one lifecycleTally
		reflect.ValueOf(&one).Elem().Field(i).SetInt(1)
		var driven []string
		for _, g := range one.gates() {
			if g.n != 0 {
				driven = append(driven, g.what)
			}
		}
		if len(driven) != 1 {
			t.Errorf("cell %s driven alone moves gates %v, want its own alone", reflect.TypeOf(zero).Field(i).Name, driven)
		}
	}
	// THE NEW CUT AND RETIRE CELLS ARE AMONG THEM, by the count each arm drives.
	var c lifecycleCounts
	c.cutMovedTokens, c.cutClearedChecks, c.abuttingQuotedCuts, c.abuttingUnquotedCuts, c.closedMarkerRetires, c.keptRetiresHeadStill = 1, 1, 1, 1, 1, 1
	var moved int
	for _, g := range c.tally().gates() {
		if g.n != 0 {
			moved++
		}
	}
	if moved != 6 {
		t.Errorf("six cut and retire counts moved %d gate(s), want 6", moved)
	}
	if !slices.ContainsFunc(gates, func(g struct {
		n    int
		what string
	}) bool {
		return strings.Contains(g.what, "one of each anchor kind")
	}) {
		t.Error("no gate counts the anchor-only retire of each kind")
	}
}
