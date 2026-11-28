package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

const (
	engLens    = "red-lens-logic"
	engOther   = "red-lens-evidence"
	engCarried = "G-0a000001" // dispatched to blue and its lens this epoch
	engLate    = "G-0a000002" // minted by the lens after the dispatch, before blue sits
	engOnGap   = "G-0a000003" // a gap about engCarried, minted after the dispatch: its author is the lens
)

// more continues a seeded stage: the next events seed onto the same record, after any live write.
func (b *stage) more() *stage { return &stage{t: b.t, runDir: b.runDir, n: b.n + 100} }

// storedPlanAt is the plan event a chair's `dispatch next` stores for epoch under an epoch limit.
func (b *stage) storedPlanAt(epoch, limit int) *stage {
	p := emptyPlan()
	p.Head, p.Epoch, p.MaxEpochs = 1, epoch, limit
	return b.add("red-chair", &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())})
}

// lateGapEpoch is one epoch up to blue's first sitting: the chair stores its plan for epoch under
// limit and dispatches the lens and blue on engCarried; the lens mints engLate and engOnGap (about
// engCarried), closing engCarried where closes says; blue registers.
func lateGapEpoch(t *testing.T, epoch, limit int, closes bool) (*stage, Run) {
	t.Helper()
	b := newStage(t).cast(engLens, engOther, "red-chair", blueSeatID, "judge").ingest().
		bracket(engLens, "lens-0").mint(engLens, engCarried, "high").stop("lens-0").
		bracket("red-chair", "chair-a").storedPlanAt(epoch, limit).dispatch(2, engLens, engCarried).dispatch(2, blueSeatID, engCarried).stop("chair-a").
		bracket(engLens, "lens-a").mint(engLens, engLate, "high").mintOnGap(engOther, engOnGap, engCarried)
	if closes {
		b = b.closeGap(engLens, engCarried)
	}
	b = b.stop("lens-a").bracket(blueSeatID, "blue-a")
	return b, b.seed()
}

func engage(t *testing.T, run Run, gaps ...string) error {
	t.Helper()
	_, err := Append(Identity{Run: run, SeatID: blueSeatID}, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(blueSeatID), GapIds: gaps})
	return err
}

func workOfBlue(t *testing.T, run Run) WorkJSON {
	t.Helper()
	w, err := WorkOfSeat(run, "blue", blueSeatID)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func blocking(w WorkJSON, sub string) bool {
	for _, it := range w.Sitting.Open {
		if it.Blocks && strings.Contains(it.What, sub) {
			return true
		}
	}
	return false
}

// T-B55: BLUE SITS AGAIN ONLY WHERE NO LATER SITTING ANSWERS (plans/findings-fold.md X22a, §III.B4).
// Under a limit of 6, epoch 4 is the last blue answers in: epoch 5 holds it out. There a gap blue
// authors minted after its dispatch is owed on its work list, with or without a close of the gap it
// was engaged on; `engage` naming exactly that set lands, once; a gap about a gap (its author is a
// lens) and a gap already engaged are not owed. Where blueSitsAfter holds (epoch 3, no limit) nothing
// is owed, and in the held-out and final epochs nothing is owed either — each refusal states why.
func TestBlueSitsAgainOnlyWhereNoLaterSittingAnswers(t *testing.T) {
	for _, closes := range []bool{false, true} {
		t.Run(map[bool]string{false: "nothing found closed", true: "its engaged gap found closed"}[closes], func(t *testing.T) {
			_, run := lateGapEpoch(t, 4, 6, closes)
			w := workOfBlue(t, run)
			if w.Engaged == nil || strings.Join(w.Engaged.Owed, ",") != engLate {
				t.Fatalf("engaged = %+v, want owed exactly %s (not the gap about a gap, not the dispatched gap)", w.Engaged, engLate)
			}
			if !blocking(w, "record `engage`") {
				t.Errorf("the owed engagement does not block the work list: %+v", w.Sitting.Open)
			}
			if err := engage(t, run, engLate, engOnGap); err == nil || !strings.Contains(err.Error(), engOnGap+" is not owed") {
				t.Errorf("engage naming a gap about a gap = %v, want refused naming it", err)
			}
			if err := engage(t, run, engCarried); err == nil || !strings.Contains(err.Error(), engCarried+" is not owed") {
				t.Errorf("engage naming the dispatched gap = %v, want refused naming it", err)
			}
			if err := engage(t, run, engLate); err != nil {
				t.Fatalf("engage naming exactly the owed set: %v", err)
			}
			w = workOfBlue(t, run)
			if len(w.Engaged.Owed) != 0 || blocking(w, "record `engage`") {
				t.Errorf("after the engage, owed = %v, want none", w.Engaged.Owed)
			}
			if err := engage(t, run, engLate); err == nil || !strings.Contains(err.Error(), "already engaged") {
				t.Errorf("a second engage this epoch = %v, want refused: blue sits at most twice", err)
			}
		})
	}
	for _, c := range []struct {
		name         string
		epoch, limit int
		why          string
	}{
		{"an earlier epoch", 3, 6, "a later sitting of yours answers"},
		{"a run held to no limit", 4, 0, "a later sitting of yours answers"},
		{"the held-out epoch", 5, 6, "holds your sittings to their closings"},
		{"the final epoch", 6, 6, "holds your sittings to their closings"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, run := lateGapEpoch(t, c.epoch, c.limit, true)
			w := workOfBlue(t, run)
			if w.Engaged == nil || len(w.Engaged.Owed) != 0 || blocking(w, "record `engage`") {
				t.Errorf("engaged = %+v, want nothing owed", w.Engaged)
			}
			if err := engage(t, run, engLate); err == nil || !strings.Contains(err.Error(), "owes none") || !strings.Contains(err.Error(), c.why) {
				t.Errorf("engage = %v, want refused: owes none, %s", err, c.why)
			}
		})
	}
}

// ONLY blue-respond ENGAGES: a lane writing a dispatch row is refused at the write path.
func TestOnlyTheRespondingSeatEngages(t *testing.T) {
	_, run := lateGapEpoch(t, 4, 6, true)
	_, err := Append(Identity{Run: run, SeatID: seatclass.LaneSeatPrefix + "1"}, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(blueSeatID), GapIds: []string{engLate}})
	if err == nil || !strings.Contains(err.Error(), "only "+blueSeatID+" engages") {
		t.Errorf("a lane's dispatch row = %v, want refused", err)
	}
}

// AN ENGAGE FOLLOWS FROM THE SITTING BLUE IS IN, and from no other: a sitting that answers no
// dispatch, and a dispatch blue has not yet sat for, owe none.
func TestEngageFollowsOnlyFromTheSittingBlueIsIn(t *testing.T) {
	b, run := lateGapEpoch(t, 4, 6, true)
	stray := b.more().stop("blue-a").bracket(blueSeatID, "blue-x")
	recordtest.Seed(t, run.Dir(), stray.evs...)
	if blocking(workOfBlue(t, run), "record `engage`") {
		t.Error("a sitting that answers no dispatch is told it owes an engage")
	}
	if err := engage(t, run, engLate); err == nil || !strings.Contains(err.Error(), "not sitting a sitting for a dispatch in this epoch") {
		t.Errorf("engage from a sitting no dispatch answers = %v, want refused", err)
	}
	b2, run2 := lateGapEpoch(t, 4, 6, true)
	next := b2.more().stop("blue-a").bracket("red-chair", "chair-b").dispatch(2, engLens, engLate).dispatch(2, blueSeatID, engLate)
	recordtest.Seed(t, run2.Dir(), next.stop("chair-b").evs...)
	if err := engage(t, run2, engLate); err == nil || !strings.Contains(err.Error(), "not sitting a sitting for a dispatch in this epoch") {
		t.Errorf("engage after the chair sat again = %v, want refused", err)
	}
}
