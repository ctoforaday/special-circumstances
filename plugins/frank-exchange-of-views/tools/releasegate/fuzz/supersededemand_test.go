package fuzz

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE DEMAND MOVES TO THE SUCCESSOR (ruled by gblock 2026-11-05), against the built binary: a successor
// at another check kind over an open `source` gap nothing answers is refused naming the demand, and
// the gap stays open; one at `source` is admitted, closes it, and awaits the live answer itself; a
// `source` gap blue's citation answers takes a `document` successor. The drive's draw agrees:
// supersedable offers the unanswered gap only to a `source` successor.
func TestASuccessorCarriesItsAncestorsLiveAnswerDemand(t *testing.T) {
	r := stagedLifecycleRunner(t)
	const blue, lens = "blue-respond", citationOwner
	c := lifecycleClaims
	if _, ok := r.do("edit", blue).set("--quote", sourceDemandOpen).set("--new", sourceDemandOpen+". "+strings.Join(c[:], ". ")).
		set("--reason", "fuzz: the claims the supersede arms rest on").admit(); !ok {
		t.Fatalf("blue's edit adding the claims: %s", r.actErr)
	}
	if !r.openSitting(lens) {
		t.Fatalf("the register of %s: %s", lens, r.actErr)
	}
	g, why := r.lifecycleMint(lens, "source", c[0], "", "low")
	if g == "" {
		t.Fatalf("the source gap's mint: %s", why)
	}
	if open, owing, _ := r.supersedable(lens, "document"); slices.Contains(open, g) || !slices.Contains(owing, g) {
		t.Fatalf("for a document successor, supersedable offers %v and holds back %v; want %s held back", open, owing, g)
	}
	r.refuseDroppedDemand(lens, g, "document")
	if r.actErr != "" || r.supersedeDemandRefused != 1 || !r.openOnRecord(g) {
		t.Fatalf("a document successor of unanswered %s: refused %d, %s open %v (%s), want refused naming the demand", g, r.supersedeDemandRefused, g, r.openOnRecord(g), r.actErr)
	}

	open, _, carries := r.supersedable(lens, "source")
	if !slices.Contains(open, g) || !carries[g] {
		t.Fatalf("for a source successor, supersedable offers %v (carries %v); want %s, carrying its demand", open, carries, g)
	}
	succ, why := r.lifecycleMint(lens, "source", c[0], "", "low", "--supersedes", g)
	if succ == "" {
		t.Fatalf("the source successor of %s: %s", g, why)
	}
	r.supersessionClosed(g, succ, true)
	if r.actErr != "" || r.supersedeDemandCarried != 1 {
		t.Fatalf("the source successor %s of %s: carried %d (%s)", succ, g, r.supersedeDemandCarried, r.actErr)
	}
	bj, err := record.BoardJSONOfRun(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if i := slices.IndexFunc(bj.Open, func(x record.GapJSON) bool { return x.ID == succ }); i < 0 || !bj.Open[i].AwaitingEvidence {
		t.Fatalf("successor %s is not open awaiting its live answer: the demand did not move to it", succ)
	}

	// A MET DEMAND IS UNAFFECTED: blue's citation answers the gap, and a document successor is admitted.
	met, why := r.lifecycleMint(lens, "source", c[1], "", "low")
	if met == "" {
		t.Fatalf("the second source gap's mint: %s", why)
	}
	if r.lifecycleCite(blue, c[1], "/lifecycle-1", met) == "" {
		t.Fatalf("blue's cite --answers %s: %s", met, r.actErr)
	}
	doc, why := r.lifecycleMint(lens, "document", c[1], "", "low", "--supersedes", met)
	if doc == "" {
		t.Fatalf("a document successor of answered %s: %s", met, why)
	}
	r.supersessionClosed(met, doc, false)
	if r.actErr != "" {
		t.Fatalf("the document successor %s of answered %s: %s", doc, met, r.actErr)
	}
}
