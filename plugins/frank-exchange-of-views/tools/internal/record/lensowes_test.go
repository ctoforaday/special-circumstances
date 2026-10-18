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

// foldOf is the gate's fold for seat, asked of the record.
func foldOf(t *testing.T, run Run, seat string) lensFold {
	t.Helper()
	g, ok, err := passLensGateOfRun(run)
	if err != nil || !ok {
		t.Fatalf("gate: %+v %v %v", g, ok, err)
	}
	for _, f := range g.lenses {
		if f.seat == seat {
			return f
		}
	}
	t.Fatalf("no fold for %s in %+v", seat, g.lenses)
	return lensFold{}
}

// A LENS THAT OWES IS NOT RETIRED (retirement.go). Each row is a debt that stands until the lens acts
// on it, staged on a lens that then sits three times at the head minting nothing: with the debt
// standing it is not retired and the moved head readies it, the reason naming what it owes; the same
// board without the debt retires it. A computation lens retired with a proof it alone is offered
// unjudged. What is new on a gap of its own is owed until it sits (TestNewsOnItsGapIsOwedUntilItSits).
func TestALensThatOwesIsNotRetired(t *testing.T) {
	for _, c := range []struct {
		name string
		lens string
		cast []string
		debt func(b *stage)
		owes string
	}{
		{"a proof it owns that nobody judged", compLens, []string{compLens, evLens},
			func(b *stage) { b.proof("P-1", "s1") }, "unjudged proof P-1"},
		{"a citation it owns reopened since its check", evLens, []string{evLens},
			func(b *stage) { b.cite("C-1").supports(evLens, "C-1").reopen("C-1") }, "reopened C-1"},
		{"a proof it owns reopened since its judgement", compLens, []string{compLens, evLens},
			func(b *stage) {
				b.proof("P-1", "s1").reproduce(compLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND).reopen("P-1")
			}, "reopened P-1"},
		{"a contradiction it read that nothing answers", evLens, []string{evLens},
			func(b *stage) { b.add(evLens, contradictingVerify("the sky is green")) }, "contradiction "},
	} {
		for _, owed := range []bool{true, false} {
			name := c.name
			if !owed {
				name = "without " + name
			}
			// board is the debt (when owed), three barren sittings of every cast lens at the head, a head
			// move, and — satAgain — a sitting of the lens at the moved head.
			board := func(t *testing.T, satAgain bool) Run {
				b := newStage(t).cast(append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")...).ingest().register("red-chair")
				if owed {
					c.debt(b)
				}
				for _, l := range c.cast {
					b.sit(2, l).sit(2, l).sit(2, l)
				}
				b.ingest()
				if satAgain {
					b.sit(b.at(), c.lens)
				}
				return b.register("red-chair").seed()
			}
			t.Run(name, func(t *testing.T) {
				run := board(t, false)
				f := foldOf(t, run, c.lens)
				plan, err := PlanDispatch(run)
				if err != nil {
					t.Fatal(err)
				}
				_, ready := partyOf(plan, c.lens)
				// The PASS gate reads the same fold: the lens the plan readies holds the PASS.
				gated := false
				for _, bl := range mustBlockers(t, run) {
					gated = gated || bl.Kind == BlockerLensReady && bl.Subject == c.lens
				}
				if gated != ready {
					t.Errorf("%s holds the PASS as a ready lens = %v, and the plan readies it = %v", c.lens, gated, ready)
				}
				if f.barren != retiresAfter {
					t.Fatalf("fixture: %s sat %d barren sittings, want %d: %+v", c.lens, f.barren, retiresAfter, f)
				}
				if !owed {
					if f.retired() == false || ready {
						t.Fatalf("a lens that owes nothing retires at its third barren sitting and the moved head does not ready it: %+v, parties %+v", f, plan.Parties)
					}
					return
				}
				if f.retired() || !f.ready || !ready {
					t.Fatalf("a lens that owes %s is not retired, and the moved head readies it: %+v, parties %+v", c.owes, f, plan.Parties)
				}
				if !slices.ContainsFunc(f.owes, func(o string) bool { return strings.HasPrefix(o, c.owes) }) || !strings.Contains(f.why, "it does not retire while it owes "+c.owes) {
					t.Errorf("the reason does not name the debt %q: owes %q, why %q", c.owes, f.owes, f.why)
				}
				// THE BOUND: once it has sat at the moved head it is not ready again until the head moves,
				// and the reason still says what keeps it from retiring.
				if f := foldOf(t, board(t, true), c.lens); f.ready || f.retired() || !strings.Contains(f.why, "it does not retire while it owes "+c.owes) {
					t.Errorf("a lens that owes and has sat at the head is not ready until the head moves, still not retired, and says what it owes: %+v", f)
				}
			})
		}
	}
}

// THE DEBT IS THE OWNER'S ALONE. A proof nobody judged is the computation lens's where the cast seats
// it, so the evidence lens beside it owes nothing for it and retires; with no computation lens, the
// evidence lens owns it and does not.
func TestAnUnjudgedProofKeepsOnlyItsOwnerFromRetiring(t *testing.T) {
	for _, c := range []struct {
		name  string
		cast  []string
		owner string
	}{
		{"computation cast", []string{compLens, evLens}, compLens},
		{"no computation lens", []string{evLens}, evLens},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")...).ingest().register("red-chair").proof("P-1", "s1")
			for _, l := range c.cast {
				b.sit(2, l).sit(2, l).sit(2, l)
			}
			run := b.ingest().register("red-chair").seed()
			for _, l := range c.cast {
				if f := foldOf(t, run, l); f.retired() == (l == c.owner) {
					t.Errorf("%s retired = %v with P-1 unjudged; want retired exactly when it does not own the proof (owner %s): %+v", l, f.retired(), c.owner, f)
				}
			}
		})
	}
}

// A REOPENED PROOF IS ITS OWNER'S, A REOPENED CITATION THE EVIDENCE LENS'S (reopenedOwnerOf): the
// blocker names that owner, the plan readies it and no other lens for the anchor, and its work list —
// opened after the edit — blocks on it.
func TestAReopenedAnchorReadiesItsOwner(t *testing.T) {
	proof := func(b *stage) {
		b.proof("P-1", "s1").reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND).reopen("P-1")
	}
	citation := func(b *stage) { b.cite("C-1").supports(compLens, "C-1").reopen("C-1") }
	for _, c := range []struct {
		name   string
		cast   []string
		stage  func(b *stage)
		anchor string
		owner  string
	}{
		{"a proof, computation cast", []string{compLens, evLens}, proof, "P-1", compLens},
		{"a proof, no computation lens", []string{evLens}, proof, "P-1", evLens},
		{"a citation, computation cast", []string{compLens, evLens}, citation, "C-1", evLens},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")...).ingest().register("red-chair")
			for _, l := range c.cast {
				b.sit(2, l)
			}
			c.stage(b)
			edit := b.at()
			run := b.register("red-chair").seed()
			var owners []string
			for _, bl := range reopenedOf(t, run) {
				owners = append(owners, bl.Subject+" "+bl.Owner)
			}
			if !slices.Equal(owners, []string{c.anchor + " " + c.owner}) {
				t.Fatalf("reopened blockers = %q, want %s owned by %s", owners, c.anchor, c.owner)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			for _, l := range c.cast {
				_, ready := partyOf(plan, l)
				// The edit moved the head, so every lens is ready to audit it; owing the anchor's check
				// is the owner's reason alone.
				readied := slices.ContainsFunc(plan.Why, func(w string) bool {
					return strings.HasPrefix(w, l+": head ") && strings.Contains(w, "it owes reopened "+c.anchor)
				})
				if readied != (l == c.owner) || (l == c.owner && !ready) {
					t.Errorf("%s readied for the reopened anchor = %v (party %v), want %v: %q", l, readied, ready, l == c.owner, plan.Why)
				}
			}
			// The owner's sitting after the edit lists it as blocking.
			after := newStage(t).cast(append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")...).ingest().register("red-chair")
			for _, l := range c.cast {
				after.sit(2, l)
			}
			c.stage(after)
			after.sit(after.at(), c.owner)
			if !slices.ContainsFunc(blockingItems(sittingOfRunT(t, after.seed(), "lens", c.owner)), func(it string) bool {
				return strings.Contains(it, c.anchor+"'s sentence changed at edit ") && strings.Contains(it, c.owner+" owes that check")
			}) {
				t.Errorf("%s's work list, opened after edit %d, does not block on %s", c.owner, edit, c.anchor)
			}
		})
	}
}

// A DEBT BELOW THE BARREN LIMIT IS NOT WHAT HOLDS THE LENS. A lens whose gap blue answered after its
// one sitting is ready because the head moved; its reason says it retires at the limit only
// owing nothing, and names the debt, rather than claiming the debt is what keeps it from retiring.
func TestALensBelowItsLimitSaysWhenItRetires(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
	b.add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
	b.sit(2, evLens)
	blueAnswersGap(b, "G1")
	f := foldOf(t, b.register("red-chair").seed(), evLens)
	if !f.ready || f.barren >= retiresAfter {
		t.Fatalf("fixture: want a ready lens below its barren limit: %+v", f)
	}
	if want := "it retires at 3 owing nothing, and it owes blue's answer to G1"; !strings.Contains(f.why, want) || strings.Contains(f.why, "does not retire while") {
		t.Errorf("why = %q, want it to say %q", f.why, want)
	}
}

// A PROOF NAMES ITS SCRIPT. A proof with no sha could never be re-run or judged, so the lens that
// owns proofs would owe a check no act can record: the write refuses it.
func TestAProofWithNoScriptIsRefused(t *testing.T) {
	runDir := newRun(t)
	_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond"}, &recordpb.Proof{ProofId: proto.String("P-00000001"), Script: proto.String("s.py")})
	if err == nil || !strings.Contains(err.Error(), "names no script") {
		t.Fatalf("a proof with no sha was not refused: %v", err)
	}
}

// blueAnswersGap is blue dispatched on gap, sitting, answering it with an edit — which moves the
// report head — and returning.
func blueAnswersGap(b *stage, gap string) *stage {
	agent := fmt.Sprintf("blue-agent-%d", b.n+1)
	return b.dispatch(b.at(), "blue-respond", gap).registerAs("blue-respond", agent).edit(gap, "a", "b").stop(agent)
}

// A GAP A LENS HOLDS WITH NOTHING NEW ON IT OWES NOTHING (gblock, 2026-10-18). universe-m28's
// architecture and voice lenses each held a gap of their own that is not material, which readies
// nobody and which blue never answered, and sat after every edit to the epoch limit: lens sittings
// went from 29–30 to 50. A lens holding only its own unanswered gap retires at its third barren
// sitting and the moved head does not ready it; a material gap still engages it as a party to its
// own dispute, on that gap, as the debate readies any minter (foldPlan).
func TestALensHoldingOnlyItsOwnUnansweredGapRetires(t *testing.T) {
	for _, c := range []struct {
		name     string
		cm       recordpb.ClassMaterial
		material bool
	}{
		{"not material", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, false},
		{"material", recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
			b.add(evLens, cmMint("G1", c.cm, "high"))
			b.sit(2, evLens).sit(2, evLens).sit(2, evLens)
			run := b.ingest().register("red-chair").seed()
			f := foldOf(t, run, evLens)
			if f.barren != retiresAfter || !f.retired() || f.ready || len(f.owes) != 0 {
				t.Fatalf("a lens holding only its own unanswered G1 retires at its third barren sitting, owing nothing, and the moved head does not ready it: %+v", f)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			p, party := partyOf(plan, evLens)
			if want := c.material; party != want || (party && !slices.Equal(p.GapIDs, []string{"G1"})) {
				t.Errorf("a party = %v on %v, want a party only on its material gap (%v): %+v", party, p.GapIDs, want, plan.Parties)
			}
		})
	}
}

// WHAT IS NEW ON A GAP OF ITS OWN IS OWED UNTIL IT SITS. Each row stages, after the lens's third
// barren sitting, one thing its work list names about the gap it minted — blue's answer, an accepted
// grade motion awaiting its regrade, the bench's remand, another lens's finding — and then a head
// move. The lens is not retired and the moved head readies it, the reason naming the debt; it is
// kept, and not ready, at an unmoved head (blue's answer moves the head itself); and once it has sat at the moved head it owes nothing and
// is retired again, whether or not that sitting acted on the item, which stays on its list.
func TestNewsOnItsGapIsOwedUntilItSits(t *testing.T) {
	const critic = "red-lens-logic"
	never, always := recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS
	for _, c := range []struct {
		name string
		cm   recordpb.ClassMaterial
		news func(b *stage)
		owes string
		item string // the work list's item it is owed for
	}{
		{"blue answered it", never, func(b *stage) { blueAnswersGap(b, "G1") }, "blue's answer to G1", "gap G1 has been ANSWERED by an edit"},
		{"a grade motion on it was accepted", never, func(b *stage) {
			b.add("blue-respond", gradeMotionFiled(b.t, "M1", "G1").GetMotion())
			b.add("red-chair", gradeMotionRuled(b.t, "M1", recordpb.GradeRuling_GRADE_RULING_ACCEPTED).GetMotionRule())
		}, "the accepted grade motion on G1", regradeOwed},
		// At impasse, as the dispatch dockets a gap: two stalled exchanges under the default terms. Blue's
		// silence in the second is an answer it owes too.
		{"the bench remanded it", always, func(b *stage) {
			b.exchangeOn("G1").exchangeOn("G1").docketAndRemand("M1", "G1")
		}, "the bench's remand of G1", "gap G1, which you minted, was REMANDED by the bench"},
		{"another lens filed a finding about it", never, func(b *stage) {
			b.add(critic, &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("G1 overstates the defect"),
				Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW),
				AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String("G1")})
		}, critic + "'s finding about G1", critic + " filed a finding about your gap G1"},
	} {
		// board is the gap, three barren sittings, the news (when staged), a head move (when moved), and —
		// satAgain — a sitting of the lens at the head.
		board := func(t *testing.T, news, moved, satAgain bool) Run {
			b := newStage(t).cast(evLens, critic, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
			b.add(evLens, cmMint("G1", c.cm, "high"))
			for range retiresAfter {
				b.sit(2, evLens)
			}
			b.sit(2, critic)
			if news {
				c.news(b)
			}
			if moved {
				b.ingest()
			}
			if satAgain {
				b.sit(b.at(), evLens)
			}
			return b.register("red-chair").seed()
		}
		t.Run(c.name, func(t *testing.T) {
			if f := foldOf(t, board(t, false, true, false), evLens); !f.retired() || f.ready {
				t.Fatalf("fixture: without the news the lens is retired and the moved head does not ready it: %+v", f)
			}
			run := board(t, true, true, false)
			f := foldOf(t, run, evLens)
			if f.retired() || !f.ready || !slices.Contains(f.owes, c.owes) || !strings.Contains(f.why, "it does not retire while it owes ") || !strings.Contains(f.why, c.owes) {
				t.Fatalf("a lens owing %q is not retired, the moved head readies it, and the reason names the debt: %+v", c.owes, f)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if _, party := partyOf(plan, evLens); !party {
				t.Errorf("the plan does not ready the lens owing %q: %+v", c.owes, plan.Parties)
			}
			if c.item != "" && !mentions(sittingOfRunT(t, run, "lens", evLens).Open, c.item) {
				t.Errorf("the lens owes %q and its work list does not name it (%q): %v", c.owes, c.item, hows(sittingOfRunT(t, run, "lens", evLens).Open))
			}
			// Blue's answer is an edit, so it moves the head itself.
			if f := foldOf(t, board(t, true, false, false), evLens); c.owes != "blue's answer to G1" && (f.retired() || f.ready || !strings.Contains(f.why, "ready when the head moves; it does not retire while it owes ") || !strings.Contains(f.why, c.owes)) {
				t.Errorf("at an unmoved head the lens owing %q is kept, not ready, and says so: %+v", c.owes, f)
			}
			sat := board(t, true, true, true)
			if f := foldOf(t, sat, evLens); !f.retired() || f.ready || len(f.owes) != 0 {
				t.Errorf("once it has sat since %q it owes nothing and is retired again: %+v", c.owes, f)
			}
			if c.item != "" && !mentions(sittingOfRunT(t, sat, "lens", evLens).Open, c.item) {
				t.Errorf("the item %q left the lens's work list with the debt; only the debt ends at its sitting", c.item)
			}
		})
	}
}
