package record

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// mintAbout is lens's mint of gap about anchor, under check kind kind, graded material.
func (b *stage) mintAbout(lens, gap string, kind recordpb.CheckKind, anchor string) *stage {
	return b.add(lens, &recordpb.Mint{GapId: proto.String(gap), AboutKind: recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), AboutRef: proto.String(anchor),
		Class: recordpb.GapClass_GAP_CLASS_VERIFICATION_SCOPE_BLINDSPOT.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: kind.Enum(), Severity: recordpb.Grade_GRADE_HIGH.Enum(), Likelihood: recordpb.Grade_GRADE_HIGH.Enum(), Impact: recordpb.Grade_GRADE_HIGH.Enum()})
}

// retire is blue's retire of anchors, naming no gap.
func (b *stage) retire(anchors ...string) *stage {
	return b.add("blue-respond", &recordpb.Retire{Claim: proto.String("the claim " + anchors[0] + " backs"), Reason: proto.String("it does not hold"), Anchors: anchors})
}

// blueRetiresIn is the chair engaging blue on gap, and blue sitting once and retiring anchors there.
func (b *stage) blueRetiresIn(gap string, anchors ...string) *stage {
	agent := fmt.Sprintf("blue-retires-%d", b.n)
	return b.dispatch(2, "blue-respond", gap).registerAs("blue-respond", agent).retire(anchors...).stop(agent)
}

// A REACHING RETIRE IS BLUE'S ACT ON THE GAP IT REACHES (V1a′.3 (a); ruling A reaching retire is an
// act on the gap). Four gaps, each reached by a retire that names none: a `computation` gap about a
// computation, a `source` gap about a citation, a `source` reading gap carried by a refuting reading
// of a citation, and a `document` gap about a computation. Before the retire the plan readies blue
// alone on it; after, the plan engages its lens, the lens's list holds the met-demand item and not
// "NO edit … answers it", it owes the retire (lensOwesOf), blue's list holds nothing for it, and the
// exchange the lens's audit ends is one the retire moved.
func TestAReachingRetireIsBluesActOnTheGap(t *testing.T) {
	for _, c := range []struct {
		name, anchor, how string
		demand            bool
		place             func(b *stage) *stage
	}{
		{"computation gap about a computation", "K-x", "which G1 is about", true, func(b *stage) *stage {
			return b.computation("K-x", "s1").mintAbout(evLens, "G1", recordpb.CheckKind_CHECK_KIND_COMPUTATION, "K-x")
		}},
		{"source gap about a citation", "C-y", "which G1 is about", true, func(b *stage) *stage {
			return b.cite("C-y").mintAbout(evLens, "G1", recordpb.CheckKind_CHECK_KIND_SOURCE, "C-y")
		}},
		{"source gap about a citation the minter's reading refutes", "C-z", "which G1 is about", true, func(b *stage) *stage {
			b.cite("C-z").foundAgainst(evLens, "C-z", "")
			g := readingGap("G1", "C-z")
			g.CheckKind = recordpb.CheckKind_CHECK_KIND_SOURCE.Enum()
			return b.add(evLens, g)
		}},
		{"document gap about a computation", "K-w", "which G1 is about", false, func(b *stage) *stage {
			return b.computation("K-w", "s1").mintAbout(evLens, "G1", recordpb.CheckKind_CHECK_KIND_DOCUMENT, "K-w")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().
				register("red-chair").dispatch(2, evLens).registerAs(evLens, "lens-mints").served(evLens, 2)
			st = c.place(st).stop("lens-mints").register("red-chair")
			before, err := PlanDispatch(st.seed())
			if err != nil {
				t.Fatal(err)
			}
			if slices.Contains(gapsOf(before, evLens), "G1") || !slices.Contains(gapsOf(before, "blue-respond"), "G1") {
				t.Fatalf("before the retire, blue alone is ready on G1: parties %+v, why %q", before.Parties, before.Why)
			}
			more := st.more()
			run := more.blueRetiresIn("G1", c.anchor).register("red-chair").seed()
			// the retire, before blue's admission, its stop and the chair's register; the first seed
			// stored the settled question after the stage's own acts
			retiredAt := int64(len(st.evs) + len(recordtest.SettledQuestion(t)) + len(more.evs) - 3)

			after, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(gapsOf(after, evLens), "G1") {
				t.Errorf("after the retire, the plan does not engage %s on G1: parties %+v, why %q", evLens, after.Parties, after.Why)
			}
			result := "so its demand is met"
			if !c.demand {
				result = "so the anchor is out of the report"
			}
			lens := sittingOfRunT(t, run, "lens", evLens)
			if want := fmt.Sprintf("gap G1: blue retired %s, %s, at event %d, %s", c.anchor, c.how, retiredAt, result); !listsItem(lens, want) || listsItem(lens, "NO edit") {
				t.Errorf("the lens's list must hold %q and not \"NO edit … answers it\": %+v", want, lens.Open)
			}
			fam, err := MergedEvents(run)
			if err != nil {
				t.Fatal(err)
			}
			ids := fam.At.IDs(fam.Events)
			owes := lensOwesOf(fam.Events, ids, fam.At, []blockerGap{{id: "G1", mintedBy: evLens}}, nil)
			if !slices.Contains(owes[evLens], "blue's retire of "+c.anchor+", "+c.how) {
				t.Errorf("the lens does not owe the retire: %q", owes[evLens])
			}
			w, err := WorkOfSeat(run, "blue", "blue-respond")
			if err != nil {
				t.Fatal(err)
			}
			if listsItem(w.Sitting, "G1") {
				t.Errorf("blue's list holds an item for G1, whose demand the retire met: %+v", w.Sitting.Open)
			}
			for _, g := range w.Open {
				if g.ID == "G1" && (g.AwaitingEvidence || g.YoursToClose) {
					t.Errorf("blue's row for G1 reads awaiting_evidence %v, yours_to_close %v", g.AwaitingEvidence, g.YoursToClose)
				}
			}
			// THE RETIRE MOVES THE GAP, as a named one does: the lens's audit of it ends an exchange
			// that did not stall.
			if x, err := Exchanges(lensAuditsG1(t, more), DefaultParams); err != nil || x["G1"].Exchanges != 1 || x["G1"].Stalled != 0 {
				t.Errorf("the lens's audit after the retire ends one exchange the retire moved: %+v (%v)", x["G1"], err)
			}
		})
	}
}

// A CLOSED GAP WHOSE ANSWERS ARE ALL RETIRED (V1a′.3 (b); ruling Closures are history): the closure
// stands, and its lens's list holds one item that blocks nothing, until the lens sits through a
// sitting opened after the retire, or mints a gap superseding it. With one of two answers still live
// there is no item.
func TestAClosedGapWhoseAnswersAreRetiredIsItsLenssNews(t *testing.T) {
	const item = "the answers to gap G1, which you closed, have all been retired"
	closed := func(t *testing.T, answers ...string) *stage {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).registerAs(evLens, "lens-mints").served(evLens, 2).mint(evLens, "G1", "high").stop("lens-mints")
		for _, k := range answers {
			st.add("blue-respond", &recordpb.Computation{Argv: []string{"python3", "s.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), ComputationId: proto.String(k),
				BundleSha: proto.String("sha-" + k), Answers: proto.String("G1")})
		}
		return st.register("red-chair").dispatch(2, evLens).registerAs(evLens, "lens-closes").closeGap(evLens, "G1").stop("lens-closes")
	}
	open := func(run Run) bool {
		for _, g := range workGapStatesT(t, run) {
			if g.ID == "G1" {
				return g.Open
			}
		}
		t.Fatal("no G1")
		return false
	}

	t.Run("one of two answers retired: no item", func(t *testing.T) {
		run := closed(t, "K-1", "K-2").retire("K-1").seed()
		if listsItem(sittingOfRunT(t, run, "lens", evLens), item) {
			t.Error("a closed gap with a live answer is listed")
		}
	})
	t.Run("all retired: listed, blocks nothing, and the closure stands", func(t *testing.T) {
		st := closed(t, "K-1").retire("K-1")
		run := st.seed()
		s := sittingOfRunT(t, run, "lens", evLens)
		i := slices.IndexFunc(s.Open, func(it Item) bool { return strings.Contains(it.What, item) })
		if i < 0 || s.Open[i].Blocks || !strings.Contains(s.Open[i].What, "(K-1)") || open(run) {
			t.Fatalf("want a non-blocking item naming K-1 and G1 closed: %+v", s.Open)
		}
		blockers, err := PassBlockers(run)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range blockers {
			if strings.Contains(b.WorkItem(), "G1") {
				t.Errorf("the item holds PASS: %+v", b)
			}
		}
		for _, seat := range []string{"blue-respond", "red-chair"} {
			role := map[string]string{"blue-respond": "blue", "red-chair": "chair"}[seat]
			if listsItem(sittingOfRunT(t, run, role, seat), item) {
				t.Errorf("%s's list holds the lens's item", seat)
			}
		}

		inside := st.more().dispatch(2, evLens).registerAs(evLens, "lens-reads")
		if !listsItem(sittingOfRunT(t, inside.seed(), "lens", evLens), item) {
			t.Error("the item is gone inside the sitting that reads it")
		}
		if listsItem(sittingOfRunT(t, inside.more().stop("lens-reads").seed(), "lens", evLens), item) {
			t.Error("the item is still listed after the lens sat through a sitting opened after the retire")
		}
	})
	t.Run("a mint superseding it removes it", func(t *testing.T) {
		st := closed(t, "K-1").retire("K-1")
		if !listsItem(sittingOfRunT(t, st.seed(), "lens", evLens), item) {
			t.Fatal("the item is not listed before the mint")
		}
		run := st.more().mintSuperseding(evLens, "G2", "high", "G1").seed()
		if listsItem(sittingOfRunT(t, run, "lens", evLens), item) {
			t.Error("a mint naming G1 in supersedes left the item listed")
		}
	})
}

// A DOCKETED GAP'S ANSWERS-RETIRED ITEM IS OFF ITS LENS'S LIST UNTIL THE BENCH RULES (ruled by gblock
// 2026-10-31, the docket hold; release fuzz seed 27): a docket motion standing unruled on the closed
// gap holds the item off, as it holds every other item on the gap. A remand leaves the gap closed and
// the item listed again; a closing ruling is a later closing act, which clears it.
func TestADocketedGapsAnswersRetiredItemWaitsOnTheBench(t *testing.T) {
	const item = "the answers to gap G1, which you closed, have all been retired"
	docketed := func(t *testing.T) *stage {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).registerAs(evLens, "lens-mints").served(evLens, 2).mint(evLens, "G1", "high").stop("lens-mints")
		st.add("blue-respond", &recordpb.Computation{Argv: []string{"python3", "s.py"}, Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), ComputationId: proto.String("K-1"),
			BundleSha: proto.String("sha-K-1"), Answers: proto.String("G1")})
		st.register("red-chair").dispatch(2, evLens).registerAs(evLens, "lens-closes").closeGap(evLens, "G1").stop("lens-closes").retire("K-1")
		if !listsItem(sittingOfRunT(t, st.seed(), "lens", evLens), item) {
			t.Fatal("the item is not listed before the docket motion")
		}
		return st.more().register("red-chair").docketMotion("red-chair", "M1", "G1")
	}
	t.Run("unruled: held off", func(t *testing.T) {
		if s := sittingOfRunT(t, docketed(t).seed(), "lens", evLens); listsItem(s, item) {
			t.Errorf("the item is listed with docket motion M1 standing unruled on G1: %+v", s.Open)
		}
	})
	t.Run("remanded: listed again", func(t *testing.T) {
		run := docketed(t).dispatch(2, "judge", "G1").register("judge").add("judge", docketRuling("M1", recordpb.Disposition_DISPOSITION_REMANDED)).seed()
		if !listsItem(sittingOfRunT(t, run, "lens", evLens), item) {
			t.Error("the item is not listed once the bench has ruled M1 and the gap stays closed")
		}
	})
	t.Run("closed by the bench: cleared", func(t *testing.T) {
		run := docketed(t).dispatch(2, "judge", "G1").register("judge").add("judge", docketRuling("M1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)).seed()
		if s := sittingOfRunT(t, run, "lens", evLens); listsItem(s, "have all been retired") {
			t.Errorf("a closing ruling after the retire left the item listed: %+v", s.Open)
		}
	})
}

// A CITATION PLACED AS ITS ANSWER BEGINS A REMAND'S EXCHANGE (V1a′.4, Gap 6's case): a remanded gap,
// and a blue sitting whose only act on it is a citation placed --answers it. Blue has acted, so the
// plan engages the lens alone on the gap, and blue's remand item stays listed without blocking.
func TestACitationPlacedAsItsAnswerBeginsARemandsExchange(t *testing.T) {
	run := remandedAtImpasse(t).register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-cites").
		add("blue-respond", &recordpb.Cite{Label: proto.String("C-0000b0b0"), Url: proto.String("https://x/b"), Sha256: proto.String("abc"), Answers: proto.String("G1")}).
		register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("blue's citation began the remand's exchange, so the lens alone is ready on G1: parties %+v, why %q", plan.Parties, plan.Why)
	}
	if remandItemBlocks(t, run, "blue", "blue-respond") || !workMentions(t, run, "blue", "blue-respond", "you have recorded an act on it since the ruling") {
		t.Errorf("blue cited --answers G1 since the ruling: its item stays on the list and does not block")
	}
}

// THE BACKING OVER BOTH CHECK VERBS (V1a′.4): a gap about a computation the computation lens judged
// sound lists that judgement in its row's backing — soundness, seat and event — and one about a
// citation its latest reading.
func TestTheBackingListsBothCheckVerbs(t *testing.T) {
	st := newStage(t).cast(evLens, computationSeat, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		computation("K-z", "s1").reproduce(computationSeat, "K-z", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND)
	judged := int64(st.n)
	st.cite("C-w").supports(evLens, "C-w")
	read := int64(st.n)
	st.mintAbout(evLens, "G1", recordpb.CheckKind_CHECK_KIND_COMPUTATION, "K-z").mintAbout(evLens, "G2", recordpb.CheckKind_CHECK_KIND_SOURCE, "C-w")
	w, err := WorkOfSeat(st.seed(), "lens", evLens)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]GapBackingJSON{
		"G1": {Anchor: "K-z", Kind: "computation", Soundness: "sound", VerifiedBy: computationSeat, Event: judged},
		"G2": {Anchor: "C-w", Kind: "citation", Outcome: "supports", Confidence: "high", VerifiedBy: evLens, Event: read},
	}
	for _, g := range w.Open {
		if b, ok := want[g.ID]; ok && (len(g.Backing) != 1 || g.Backing[0] != b) {
			t.Errorf("%s's backing = %+v, want [%+v]", g.ID, g.Backing, b)
		}
	}
}

// THE BACKING READS THE LEDGER'S "STANDS" (m40 defect 5): a check made before an edit changed the
// sentence is of a sentence the report no longer carries, and the backing lists the anchor
// unverified, as the unchecked blocker does — until a check from a sitting opened after the edit.
func TestTheBackingListsNoCheckAnEditOutdated(t *testing.T) {
	backing := func(run Run) GapBackingJSON {
		t.Helper()
		w, err := WorkOfSeat(run, "lens", evLens)
		if err != nil {
			t.Fatal(err)
		}
		if len(w.Open) != 1 || len(w.Open[0].Backing) != 1 {
			t.Fatalf("open = %+v, want G1 with C-1 its one backing", w.Open)
		}
		return w.Open[0].Backing[0]
	}
	board := func() *stage {
		return reopenedBoard(t).mintAbout(evLens, "G1", recordpb.CheckKind_CHECK_KIND_SOURCE, "C-1")
	}
	if got := backing(board().seed()); got.Event == 0 || got.Outcome != "supports" {
		t.Fatalf("before the edit the backing = %+v, want the standing reading", got)
	}
	edited := board().reopen("C-1").seed()
	if got, want := backing(edited), (GapBackingJSON{Anchor: "C-1", Kind: "citation"}); got != want {
		t.Errorf("after the edit the backing = %+v, want %+v — the reading is of the sentence the edit replaced", got, want)
	}
	if board, err := BoardJSONOfRun(edited); err != nil || len(board.Open) != 1 || len(board.Open[0].Backing) != 1 ||
		board.Open[0].Backing[0] != (GapBackingJSON{Anchor: "C-1", Kind: "citation"}) {
		t.Errorf("after the edit the board's open gaps = %+v (%v), want G1 backed by C-1 unverified, as the work list reads it", board.Open, err)
	}
	if rows := reopenedOf(t, edited); len(rows) != 1 || rows[0].Subject != "C-1" {
		t.Fatalf("unchecked blockers = %+v, want C-1: the backing and the plan read one rule", rows)
	}
	b := board().reopen("C-1")
	b.sit(b.at(), evLens).supports(evLens, "C-1")
	if got := backing(b.seed()); got.Event != b.at() || got.Outcome != "supports" {
		t.Errorf("after a later sitting's reading the backing = %+v, want that reading at event %d", got, b.at())
	}
}

// ONE EVENT, ONE ITEM. `retire --anchor K-x --answers G1`, where G1 is about K-x, is blue's answer to
// G1 and a reaching retire of it at once: the lens owes it once, as the retire it is.
func TestARetireAnsweringTheGapItReachesIsOneItem(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().
		register("red-chair").dispatch(2, evLens).registerAs(evLens, "lens-mints").served(evLens, 2).
		computation("K-x", "s1").mintAbout(evLens, "G1", recordpb.CheckKind_CHECK_KIND_COMPUTATION, "K-x").stop("lens-mints").register("red-chair").
		dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-retires").
		add("blue-respond", &recordpb.Retire{Claim: proto.String("the claim K-x backs"), Reason: proto.String("it does not hold"), Anchors: []string{"K-x"},
			ClaimKept: proto.Bool(true), Answers: proto.String("G1")}).
		stop("blue-retires").register("red-chair")
	fam, err := MergedEvents(st.seed())
	if err != nil {
		t.Fatal(err)
	}
	owes := lensOwesOf(fam.Events, fam.At.IDs(fam.Events), fam.At, []blockerGap{{id: "G1", mintedBy: evLens}}, nil)
	if want := []string{"blue's retire of K-x, which G1 is about"}; !slices.Equal(owes[evLens], want) {
		t.Errorf("the lens owes %q for one retire, want %q", owes[evLens], want)
	}
}
