package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// The universe-m31 ids, so a reader can hold each test against that run's record.
const (
	m31Logic      = "red-lens-logic"
	m31Closed     = "G-77cc033b" // epoch 10's dispatched gap, closed by logic (ev636)
	m31Minted     = "G-1ac85c33" // minted by logic in the same sitting (ev635)
	m31Closed8    = "G-c791e4e4" // epoch 8's dispatched gap, closed by logic (ev584), no successor named
	m31NotLineage = "G-77cc033b" // minted in epoch 8 (ev587) — "Lineage: none"
)

// more continues a seeded stage: the next events seed onto the same record, after any live write.
func (b *stage) more() *stage { return &stage{t: b.t, runDir: b.runDir, n: b.n + 100} }

// m31Epoch10 is universe-m31's epoch 10 up to blue's first sitting: the chair dispatches logic and
// blue on G-77cc033b, logic mints G-1ac85c33 and closes G-77cc033b, and blue sits.
func m31Epoch10(t *testing.T) (*stage, Run) {
	b := newStage(t).cast(m31Logic, "red-chair", blueSeatID, "judge").ingest().
		bracket(m31Logic, "logic-0").mint(m31Logic, m31Closed, "high").stop("logic-0").
		bracket("red-chair", "chair-a").dispatch(2, m31Logic, m31Closed).dispatch(2, blueSeatID, m31Closed).dispatch(2, "judge", m31Closed).stop("chair-a").
		bracket(m31Logic, "logic-a").mint(m31Logic, m31Minted, "high").closeGap(m31Logic, m31Closed).stop("logic-a").
		bracket(blueSeatID, "blue-a")
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

// UNIVERSE-M31 EPOCH 8, REPLAYED: logic closed G-c791e4e4 naming no successor and minted G-77cc033b
// with "Lineage: none". The record names no successor, so nothing is owed and an engage is refused
// — blue does not guess one from a gap minted in the same sitting.
func TestAGapClosedWithNoSuccessorOwesNoSecondSitting(t *testing.T) {
	b := newStage(t).cast(m31Logic, "red-chair", blueSeatID, "judge").ingest().
		bracket(m31Logic, "logic-0").mint(m31Logic, m31Closed8, "high").stop("logic-0").
		bracket("red-chair", "chair-a").dispatch(2, m31Logic, m31Closed8).dispatch(2, blueSeatID, m31Closed8).stop("chair-a").
		bracket(m31Logic, "logic-a").closeGap(m31Logic, m31Closed8).mint(m31Logic, m31NotLineage, "high").stop("logic-a").
		bracket(blueSeatID, "blue-a")
	run := b.seed()
	w := workOfBlue(t, run)
	if w.Engaged == nil || len(w.Engaged.FoundClosed) != 1 || w.Engaged.FoundClosed[0].Successors == nil || len(w.Engaged.FoundClosed[0].Successors) != 0 {
		t.Fatalf("engaged = %+v, want %s found closed with an empty successors list", w.Engaged, m31Closed8)
	}
	if blocking(w, "record `engage`") {
		t.Errorf("an engage is owed where the record names no successor: %+v", w.Sitting.Open)
	}
	err := engage(t, run, m31NotLineage)
	if err == nil || !strings.Contains(err.Error(), "owes none") {
		t.Errorf("engage on a gap the record names as no successor = %v, want refused", err)
	}
}

// ONLY blue-respond ENGAGES: a lane writing a dispatch row is refused at the write path.
func TestOnlyTheRespondingSeatEngages(t *testing.T) {
	_, run := m31Epoch10(t)
	_, err := Append(Identity{Run: run, SeatID: seatclass.LaneSeatPrefix + "1"}, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(blueSeatID), GapIds: []string{m31Minted}})
	if err == nil || !strings.Contains(err.Error(), "only "+blueSeatID+" engages") {
		t.Errorf("a lane's dispatch row = %v, want refused", err)
	}
}

// AN ENGAGE FOLLOWS FROM THE SITTING BLUE IS IN, and from no other: a sitting that answers no
// dispatch, and a dispatch blue has not yet sat for, owe none — the found-closed gaps of the sitting
// before are not this one's.
func TestEngageFollowsOnlyFromTheSittingBlueIsIn(t *testing.T) {
	b, run := m31Epoch10(t)
	stray := b.more().stop("blue-a").bracket(blueSeatID, "blue-x")
	recordtest.Seed(t, run.Dir(), stray.evs...)
	if blocking(workOfBlue(t, run), "record `engage`") {
		t.Error("a sitting that answers no dispatch is told it owes an engage")
	}
	if err := engage(t, run, m31Minted); err == nil || !strings.Contains(err.Error(), "not sitting a sitting for a dispatch in this epoch") {
		t.Errorf("engage from a sitting no dispatch answers = %v, want refused", err)
	}

	for _, c := range []struct {
		name string
		gaps []string // the next epoch's dispatch naming blue; nil: the chair dispatches logic alone
	}{{"the chair has sat since and dispatched blue", []string{m31Minted}}, {"the chair has sat since and dispatched others", nil}} {
		t.Run(c.name, func(t *testing.T) {
			b2, run2 := m31Epoch10(t)
			next := b2.more().stop("blue-a").bracket("red-chair", "chair-b").dispatch(2, m31Logic, m31Minted)
			if c.gaps != nil {
				next = next.dispatch(2, blueSeatID, c.gaps...)
			}
			recordtest.Seed(t, run2.Dir(), next.stop("chair-b").evs...)
			if err := engage(t, run2, m31Minted); err == nil || !strings.Contains(err.Error(), "not sitting a sitting for a dispatch in this epoch") {
				t.Errorf("engage after the chair sat again = %v, want refused", err)
			}
		})
	}
}
