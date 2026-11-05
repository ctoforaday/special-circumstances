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

// UNIVERSE-M38, ITS SHAPE REPLAYED (ruled by gblock 2026-11-04). Blue sat in epoch 11 of 12, the last
// that dispatches a lens: it moved the head, reopened a computation, unseated the checks a lens had
// made minutes before, and left six lenses ready at a final sitting that dispatches nobody. Wherever
// blue acts last, a PASS is unreachable. So blue is not readied in the epoch before the last, and
// every act blue makes is followed by a lens's sitting.

// m38Board is the shape under a limit of 4: the lens mints G1 and G2 at epoch 1, blue answers G1 at
// epoch 2, and the chair opens epoch 3 — the last that dispatches a lens — with the lens ready to
// re-audit blue's answer and blue still ready on both gaps.
func m38Board(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "high").mint(evLens, "G2", "high"). // epoch 1
		register("red-chair").dispatch(2, "blue-respond", "G1", "G2").register("blue-respond").     // epoch 2
		edit("G1", "was", "is").
		register("red-chair") // epoch 3
}

func TestM38BlueIsHeldOutOfTheLastEpochThatDispatchesALens(t *testing.T) {
	b := m38Board(t)
	run := b.seed()
	writeRunConfig(t, b.runDir, `{"maxEpochs":4}`)
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := partyOf(plan, blueRespondSeat); ok {
		t.Fatalf("epoch 3 of 4 readies blue: parties %+v", plan.Parties)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G1") {
		t.Errorf("epoch 3 of 4 must still ready the lens to re-audit blue's answer to G1: parties %+v", plan.Parties)
	}
	if plan.FinalSitting || plan.EpochLimitReached || plan.Ceiling || plan.PassPermitted {
		t.Errorf("epoch 3 of 4 dispatches the lens: final %v, limit %v, ceiling %v, pass %v", plan.FinalSitting, plan.EpochLimitReached, plan.Ceiling, plan.PassPermitted)
	}
	if !plan.BlueHeldOut() {
		t.Errorf("the plan at epoch %d of %d does not say it holds blue out", plan.Epoch, plan.MaxEpochs)
	}
	if !whyMentions(plan, "blue-respond: not dispatched — epoch 3 is the last that dispatches a lens under the run's epoch limit of 4") {
		t.Errorf("the plan does not say why blue is not dispatched: %q", plan.Why)
	}
	// NO LINE SAYS BLUE IS READY: G2's names the hold, not readiness.
	if line := planLine(t, run, "G2"); !strings.Contains(line, blueHeldOutClause) || strings.Contains(line, "blue is ready") {
		t.Errorf("G2's line is %q, want the hold and no readiness", line)
	}

	// ONE EPOCH SHORT OF IT, blue is dispatched as ever.
	writeRunConfig(t, b.runDir, `{"maxEpochs":5}`)
	plan, err = PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if gaps := gapsOf(plan, blueRespondSeat); !slices.Equal(gaps, []string{"G1", "G2"}) || plan.BlueHeldOut() || whyMentions(plan, blueHeldOutClause) {
		t.Errorf("epoch 3 of 5 = blue on %v, held out %v, why %q; want blue on G1 and G2", gaps, plan.BlueHeldOut(), plan.Why)
	}
}

// THE FINAL SITTING FINDS NO LENS READY ON BLUE'S ACCOUNT. The chair dispatches exactly what each plan
// readies: the lens re-audits and keeps its grade, and blue answers every gap with an edit whenever it
// sits — m38's blue at epoch 11. Under a limit of 4, blue's last sitting is epoch 2 and the lens audits
// it at epoch 3, so the final sitting reads the head the lens audited and holds no lens ready, where
// m38's final plan held six.
func TestM38TheFinalSittingFindsNoLensReadyOnBluesAccount(t *testing.T) {
	const limit = 4
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "high").mint(evLens, "G2", "high")
	blueSat := 0
	for epoch := 2; ; epoch++ {
		b.register("red-chair")
		run := newStage(t)
		writeRunConfig(t, run.runDir, fmt.Sprintf(`{"maxEpochs":%d}`, limit))
		run.evs, run.n = append(run.evs, b.evs...), b.n
		plan, err := PlanDispatch(run.seed())
		if err != nil {
			t.Fatal(err)
		}
		if plan.FinalSitting {
			if epoch != limit {
				t.Fatalf("the final sitting is epoch %d, want %d", epoch, limit)
			}
			// THE FINAL SITTING HOLDS NOBODY OUT: it dispatches nobody, and says so in its own words.
			if plan.BlueHeldOut() || whyMentions(plan, "blue-respond: not dispatched") {
				t.Errorf("the final sitting says it holds blue out: %q", plan.Why)
			}
			for _, bl := range plan.Blockers {
				if bl.Kind == BlockerLensReady {
					t.Errorf("the final sitting holds %s ready after blue's last act: blockers %+v", bl.Subject, plan.Blockers)
				}
			}
			break
		}
		if len(plan.Parties) == 0 {
			t.Fatalf("epoch %d dispatches nobody: %q", epoch, plan.Why)
		}
		head := plan.Head
		for _, p := range plan.Parties {
			b.dispatch(head, p.SeatID, p.GapIDs...)
		}
		for _, p := range plan.Parties {
			switch p.SeatID {
			case evLens:
				b.register(evLens).served(evLens, head)
				for _, g := range p.GapIDs {
					b.keepGrade(evLens, g, "the answer misses the check")
				}
			case blueRespondSeat:
				blueSat = epoch
				b.register(blueRespondSeat)
				for _, g := range p.GapIDs {
					b.edit(g, fmt.Sprintf("was %d", b.n), fmt.Sprintf("is %d", b.n))
				}
			}
		}
	}
	if blueSat != limit-2 {
		t.Errorf("blue last sat at epoch %d, want %d: the epoch before the last holds it out", blueSat, limit-2)
	}
}

// BLUE ALONE, HELD OUT, IS THE LIMIT REACHED ONE SITTING EARLY. Nobody else is ready, so nobody is
// dispatched and the run ends CEILING: the limit is what leaves blue undispatched, and the record
// derives the word after the exit.
func TestBlueAloneHeldOutEndsTheRunAtItsCeiling(t *testing.T) {
	b := epochTwo(t)
	run := b.seed()
	writeRunConfig(t, b.runDir, `{"maxEpochs":3}`)
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.EpochLimitReached || !plan.Ceiling || plan.PassPermitted || plan.FinalSitting || len(plan.Parties) != 0 {
		t.Fatalf("plan at epoch 2 of 3 with blue alone ready = %+v, want nobody dispatched and the ceiling", plan)
	}
	if !whyMentions(plan, "epoch limit 3 reached — blue is the only party the board readies, and it is held out of epoch 2") {
		t.Errorf("the plan does not say why the limit is reached: %q", plan.Why)
	}
	got, why, ok, err := DeriveVerdict(run)
	if err != nil || !ok || got != "CEILING" {
		t.Errorf("derived (%q, %q, %v, %v), want CEILING", got, why, ok, err)
	}
}

// A DOCKET IN THE EPOCH THAT HOLDS BLUE OUT readies the lens for its closing and says blue files none.
func TestADocketInTheHeldOutEpochAwaitsTheLensAlone(t *testing.T) {
	run := atImpasse(t) // the chair opens epoch 5 with G1 at impasse
	writeRunConfig(t, run.dir, `{"maxEpochs":6}`)
	plan, _, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plan.Docket, []string{"G1"}) {
		t.Fatalf("plan = %+v, want G1 docketed", plan)
	}
	if _, ok := partyOf(plan, evLens); !ok {
		t.Errorf("the lens owes its closing and is not readied: parties %+v", plan.Parties)
	}
	if _, ok := partyOf(plan, blueRespondSeat); ok {
		t.Errorf("blue is readied for its closing in the epoch that holds it out: parties %+v", plan.Parties)
	}
	if line := planLine(t, run, "G1"); !strings.Contains(line, "red-lens-evidence is readied to file its own; "+blueHeldOutClause) || strings.Contains(line, "each is readied") {
		t.Errorf("G1's line is %q, want the lens readied and blue held out", line)
	}
	// WITH THE LENS'S CLOSING FILED, nothing a dispatched party can file is awaited: the bench sits now,
	// in the last epoch that dispatches it, rather than wait on a closing blue cannot file.
	recordtest.Seed(t, run.Dir(), recordtest.At(t, evLens, "m38:closing:G1", &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("the case on G1")}))
	plan, err = PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, "judge"), "G1") {
		t.Errorf("with the lens's closing in and blue held out, the bench is not ready on G1: parties %+v, why %q", plan.Parties, plan.Why)
	}
	if line := planLine(t, run, "G1"); !strings.Contains(line, blueHeldOutClause+", so it files no closing on the gap; no other party's closing is awaited, so the bench is ready") {
		t.Errorf("G1's line is %q, want blue held out and the bench ready", line)
	}
}

// M38 DEFECT 5: THE FINAL SITTING'S DOCKET OWES NO CLOSING ANYONE CAN FILE. The docket stands, and the
// terminal bench rules it at the exit; the plan says so, and readies no party for a closing.
func TestTheFinalSittingsDocketSaysNoClosingCanBeFiled(t *testing.T) {
	run := atImpasse(t)
	writeRunConfig(t, run.dir, `{"maxEpochs":5}`)
	plan, _, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.FinalSitting || !plan.EpochLimitReached || len(plan.Parties) != 0 || !slices.Equal(plan.Docket, []string{"G1"}) {
		t.Fatalf("plan = %+v, want the final sitting, nobody dispatched, and G1 docketed", plan)
	}
	want := "G1: docketed and unruled — the closing of red-lens-evidence and blue-respond is owed on it, and no seat is dispatched after this sitting, so none can be filed; the bench rules it at the exit, on the record as it stands"
	if !slices.Contains(plan.Why, want) {
		t.Errorf("the final sitting's line for G1 is not %q: %q", want, plan.Why)
	}
	if whyMentions(plan, "readied to file") {
		t.Errorf("the final sitting's plan readies a party to file a closing: %q", plan.Why)
	}
}

// EVERY SOURCE THAT READIES BLUE IS HELD OUT, AND ITS LINE SAYS SO. Each board readies blue at the epoch
// it has reached; under a limit one past that epoch, the same board readies no blue, and the line that
// named blue ready names the hold instead.
func TestEverySourceThatReadiesBlueIsHeldOut(t *testing.T) {
	for _, c := range []struct {
		name, line, ready string
		board             func(t *testing.T) *stage
	}{
		{"a remand's one more exchange", "G1: at impasse", "blue is ready for the one more exchange", func(t *testing.T) *stage {
			return remandedAtImpasse(t).register("red-chair")
		}},
		{"a failed check's remedy", "blue-respond: computation K-1 stands on a failed check", "ready to remedy it", unsoundBoard},
		{"an open inquiry", "blue-respond: inquiry Q1", "ready, to record a fate", func(t *testing.T) *stage {
			return inquiryBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true))
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := c.board(t)
			run := b.seed()
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := partyOf(plan, blueRespondSeat); !ok || !strings.Contains(lineOf(plan, c.line), c.ready) {
				t.Fatalf("setup: with no limit the board readies blue: parties %+v, why %q", plan.Parties, plan.Why)
			}
			writeRunConfig(t, b.runDir, fmt.Sprintf(`{"maxEpochs":%d}`, plan.Epoch+1))
			if plan, err = PlanDispatch(run); err != nil {
				t.Fatal(err)
			}
			if _, ok := partyOf(plan, blueRespondSeat); ok {
				t.Errorf("epoch %d of %d readies blue: parties %+v", plan.Epoch, plan.MaxEpochs, plan.Parties)
			}
			if line := lineOf(plan, c.line); !strings.Contains(line, blueHeldOutClause) || strings.Contains(line, c.ready) {
				t.Errorf("the line is %q, want the hold and not %q", line, c.ready)
			}
		})
	}
}

// A REMAND WHOSE EXCHANGE WAITS ON BLUE IS NOT OWED WORK IN THE EPOCH THAT HOLDS BLUE OUT. The release
// sweep on #1744 (seeds 9, 13, 17, 18, 20, 21, 24, 31, 36, 39): the plan named the remand on
// remand_owed and engaged nobody on its gap. Blue sits neither there nor at the final sitting, so the
// exchange cannot be had before the run ends: the plan does not list it, its line says so, and the gap
// still holds the PASS.
func TestARemandWaitingOnBlueIsNotOwedInTheHeldOutEpoch(t *testing.T) {
	b := remandedAtImpasse(t).register("red-chair")
	run := b.seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.RemandOwed) != 1 || !slices.Contains(gapsOf(plan, blueRespondSeat), "G1") {
		t.Fatalf("setup: with no limit the remand is owed and readies blue: remand_owed %+v, parties %+v", plan.RemandOwed, plan.Parties)
	}
	writeRunConfig(t, b.runDir, fmt.Sprintf(`{"maxEpochs":%d}`, plan.Epoch+1))
	if plan, err = PlanDispatch(run); err != nil {
		t.Fatal(err)
	}
	if len(plan.RemandOwed) != 0 {
		t.Errorf("epoch %d of %d lists the remand as owed while nobody can have its exchange: %+v, parties %+v", plan.Epoch, plan.MaxEpochs, plan.RemandOwed, plan.Parties)
	}
	if line := lineOf(plan, "G1: at impasse"); !strings.Contains(line, "so that exchange cannot be had before the run ends") {
		t.Errorf("G1's line is %q, want the exchange named as one that cannot be had", line)
	}
	if plan.PassPermitted {
		t.Errorf("the remanded gap no longer holds the PASS: %+v", plan.Blockers)
	}
}
