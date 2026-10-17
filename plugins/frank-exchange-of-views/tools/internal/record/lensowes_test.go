package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// A LENS THAT OWES IS NOT RETIRED (retirement.go). Each row is one thing a lens's work list names as
// its own to clear, staged on a lens that then sits three times at the head minting nothing: with the
// debt standing it is not retired and the moved head readies it, the reason naming what it owes; the
// same board without the debt retires it. universe-m23's voice lens retired holding an open gap
// that is not material, and a computation lens retired with a proof it alone is offered unjudged.
func TestALensThatOwesIsNotRetired(t *testing.T) {
	for _, c := range []struct {
		name string
		lens string
		cast []string
		debt func(b *stage)
		owes string
	}{
		{"an open gap it minted that is not material", evLens, []string{evLens},
			func(b *stage) { b.add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high")) }, "open gap G1"},
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

// A DEBT BELOW THE BARREN LIMIT IS NOT WHAT HOLDS THE LENS. A lens that minted a gap still open and
// has sat once since is ready because the head moved; its reason says it retires at the limit only
// owing nothing, and names the debt, rather than claiming the debt is what keeps it from retiring.
func TestALensBelowItsLimitSaysWhenItRetires(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
	b.add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
	b.sit(2, evLens)
	b.ingest()
	f := foldOf(t, b.register("red-chair").seed(), evLens)
	if !f.ready || f.barren >= retiresAfter {
		t.Fatalf("fixture: want a ready lens below its barren limit: %+v", f)
	}
	if want := "it retires at 3 owing nothing, and it owes open gap G1"; !strings.Contains(f.why, want) || strings.Contains(f.why, "does not retire while") {
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
