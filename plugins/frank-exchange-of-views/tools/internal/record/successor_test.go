package record

import (
	"slices"
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
	m31Successor  = "G-1ac85c33" // minted by logic in the same sitting, supersedes G-77cc033b (ev635)
	m31Closed8    = "G-c791e4e4" // epoch 8's dispatched gap, closed by logic (ev584), no successor named
	m31NotLineage = "G-77cc033b" // minted in epoch 8 (ev587) — "Lineage: none"
)

func (b *stage) closeSuperseded(lens, gap, successor string) *stage {
	return b.add(lens, &recordpb.Close{GapId: proto.String(gap), Successor: proto.String(successor),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf; the rest is carried forward")})
}

// more continues a seeded stage: the next events seed onto the same record, after any live write.
func (b *stage) more() *stage { return &stage{t: b.t, runDir: b.runDir, n: b.n + 100} }

// m31Epoch10 is universe-m31's epoch 10 up to blue's first sitting: the chair dispatches logic and
// blue on G-77cc033b, logic mints G-1ac85c33 superseding it and closes it naming that successor, and
// blue sits.
func m31Epoch10(t *testing.T) (*stage, Run) {
	b := newStage(t).cast(m31Logic, "red-chair", blueSeatID, "judge").ingest().
		bracket(m31Logic, "logic-0").mint(m31Logic, m31Closed, "high").stop("logic-0").
		bracket("red-chair", "chair-a").dispatch(2, m31Logic, m31Closed).dispatch(2, blueSeatID, m31Closed).dispatch(2, "judge", m31Closed).stop("chair-a").
		bracket(m31Logic, "logic-a").mintSuperseding(m31Logic, m31Successor, "high", m31Closed).closeSuperseded(m31Logic, m31Closed, m31Successor).stop("logic-a").
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

// UNIVERSE-M31 EPOCH 10, REPLAYED: blue finds G-77cc033b closed, its work list names the successor
// the record holds, an `engage` on it is owed and blocks until recorded, the engage row engages
// blue's NEXT sitting (not this one), that sitting engages the successor and owes a position on it,
// and the bound holds — a second engage that epoch is refused.
func TestBlueSitsAgainOnTheSuccessorOfAGapItFoundClosed(t *testing.T) {
	b, run := m31Epoch10(t)
	w := workOfBlue(t, run)
	if w.Engaged == nil || len(w.Engaged.FoundClosed) != 1 {
		t.Fatalf("engaged = %+v, want %s found closed", w.Engaged, m31Closed)
	}
	if got := w.Engaged.FoundClosed[0].Successors; !slices.Equal(got, []string{m31Successor}) {
		t.Errorf("found_closed successors = %v, want [%s] — the close's superseded_by and the mint's supersedes both name it, once", got, m31Successor)
	}
	if !blocking(w, "record `engage`") {
		t.Errorf("blue's work list does not block on the engage it owes: %+v", w.Sitting.Open)
	}
	if err := engage(t, run, m31Successor); err != nil {
		t.Fatalf("engage on the record's successor was refused: %v", err)
	}
	w = workOfBlue(t, run)
	if blocking(w, "record `engage`") {
		t.Errorf("the engage item still blocks after the act: %+v", w.Sitting.Open)
	}
	if blocking(w, "register for this sitting") {
		t.Errorf("the engage row is owed by the NEXT sitting, yet the sitting that wrote it is told to register: %+v", w.Sitting.Open)
	}
	if err := engage(t, run, m31Successor); err == nil || !strings.Contains(err.Error(), "owes no `engage`") {
		t.Errorf("a second engage in the same sitting = %v, want it refused: the successor is engaged this epoch", err)
	}

	// The second sitting: the hook brackets it after the engage row.
	b2 := b.more().bracket(blueSeatID, "blue-b")
	recordtest.Seed(t, run.Dir(), b2.evs...)
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	ss := BlueSittings(m.Events, m.At, WhileRunning)
	if len(ss) != 2 {
		t.Fatalf("blue sittings = %d, want 2 in the epoch", len(ss))
	}
	if !slices.Equal(ss[0].Open, nil) || len(ss[0].FoundClosed) != 1 {
		t.Errorf("first sitting = open %v, found closed %d; want nothing open and %s found closed", ss[0].Open, len(ss[0].FoundClosed), m31Closed)
	}
	if !slices.Equal(ss[1].Engaged, []string{m31Successor}) || !slices.Equal(ss[1].Open, []string{m31Successor}) {
		t.Errorf("second sitting engaged %v open %v, want the successor %s open", ss[1].Engaged, ss[1].Open, m31Successor)
	}
	w = workOfBlue(t, run)
	if w.Engaged == nil || !slices.Equal(w.Engaged.GapIDs, []string{m31Successor}) {
		t.Errorf("second sitting's engaged = %+v, want %s", w.Engaged, m31Successor)
	}
	if !blocking(w, "position is missing") {
		t.Errorf("the second sitting found the successor open and owes a position: %+v", w.Sitting.Open)
	}
	if err := engage(t, run, m31Successor); err == nil {
		t.Error("an engage in the second sitting on the successor already engaged this epoch was recorded — the bound is once per successor per epoch")
	}

	// ONE DISPATCH: the engage row joins the chair's, so the bench that sits after blue is still a
	// party of it, and blue's party row is the engage row its second sitting answers.
	gs := DispatchGroups(m.Events, m.At)
	if len(gs) != 1 {
		t.Fatalf("dispatch groups = %d, want the engage row in the chair's group", len(gs))
	}
	if r := gs[0].PartyRows[blueSeatID]; !slices.Equal(r.GapIDs, []string{m31Successor}) {
		t.Errorf("blue's party row = %+v, want the engage row", r)
	}
	if _, sat := gs[0].Sat[blueSeatID]; !sat {
		t.Error("blue's second sitting does not answer the engage row")
	}
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
	if err == nil || !strings.Contains(err.Error(), "owes no `engage`") {
		t.Errorf("engage on a gap the record names as no successor = %v, want refused", err)
	}
}

// A SUCCESSOR THE RECORD DOES NOT NAME IS REFUSED, AND SO IS A SET THAT LEAVES ONE OUT. The close
// names one successor, a second mint supersedes the same gap, and an unrelated gap is open: each id
// is held to the record's set, whole.
func TestEngageIsHeldToTheSuccessorsTheRecordNames(t *testing.T) {
	const second, unrelated = "G-00000a02", "G-00000a03"
	b := newStage(t).cast(m31Logic, "red-chair", blueSeatID, "judge").ingest().
		bracket(m31Logic, "logic-0").mint(m31Logic, m31Closed, "high").mint(m31Logic, unrelated, "high").stop("logic-0").
		bracket("red-chair", "chair-a").dispatch(2, m31Logic, m31Closed).dispatch(2, blueSeatID, m31Closed).stop("chair-a").
		bracket(m31Logic, "logic-a").mint(m31Logic, m31Successor, "high").closeSuperseded(m31Logic, m31Closed, m31Successor).
		mintSuperseding(m31Logic, second, "high", m31Closed).stop("logic-a").
		bracket(blueSeatID, "blue-a")
	run := b.seed()
	if got := workOfBlue(t, run).Engaged.FoundClosed[0].Successors; !slices.Equal(got, []string{m31Successor, second}) {
		t.Errorf("successors = %v, want the close's %s then the mint's %s", got, m31Successor, second)
	}
	for _, c := range []struct {
		name string
		ids  []string
		says string
	}{
		{"an unrelated open gap", []string{m31Successor, second, unrelated}, "no successor the record names"},
		{"one left out", []string{m31Successor}, "leaves out " + second},
		{"named twice", []string{m31Successor, second, second}, "named twice"},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := engage(t, run, c.ids...)
			if err == nil || !strings.Contains(err.Error(), c.says) {
				t.Errorf("engage %v = %v, want refused saying %q", c.ids, err, c.says)
			}
		})
	}
	if err := engage(t, run, second, m31Successor); err != nil {
		t.Errorf("engage on exactly the owed set, in another order, was refused: %v", err)
	}
}

// A SUCCESSOR THE CHAIR ALREADY ENGAGED BLUE ON THIS EPOCH IS NOT SAT TWICE: blue answers it in
// this sitting, and no engage is owed for it.
func TestASuccessorAlreadyEngagedThisEpochIsNotOwed(t *testing.T) {
	b := newStage(t).cast(m31Logic, "red-chair", blueSeatID, "judge").ingest().
		bracket(m31Logic, "logic-0").mint(m31Logic, m31Closed, "high").mintSuperseding(m31Logic, m31Successor, "high", m31Closed).stop("logic-0").
		bracket("red-chair", "chair-a").dispatch(2, m31Logic, m31Closed).dispatch(2, blueSeatID, m31Closed, m31Successor).stop("chair-a").
		bracket(m31Logic, "logic-a").closeGap(m31Logic, m31Closed).stop("logic-a").
		bracket(blueSeatID, "blue-a")
	run := b.seed()
	if blocking(workOfBlue(t, run), "record `engage`") {
		t.Error("an engage is owed on a successor this epoch's dispatch already engages blue on")
	}
	if err := engage(t, run, m31Successor); err == nil {
		t.Error("engage on a successor the chair's row already engages was recorded")
	}
}

// ONLY blue-respond ENGAGES: a lane writing a dispatch row is refused at the write path.
func TestOnlyTheRespondingSeatEngages(t *testing.T) {
	_, run := m31Epoch10(t)
	_, err := Append(Identity{Run: run, SeatID: seatclass.LaneSeatPrefix + "1"}, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(blueSeatID), GapIds: []string{m31Successor}})
	if err == nil || !strings.Contains(err.Error(), "only "+blueSeatID+" engages") {
		t.Errorf("a lane's dispatch row = %v, want refused", err)
	}
}

// A SUCCESSOR ALREADY CLOSED IS OWED NOTHING: the record names it, and there is nothing in it to
// answer. The work list names no successor, and an engage on it is refused.
func TestAClosedSuccessorIsNotOwed(t *testing.T) {
	b := newStage(t).cast(m31Logic, "red-chair", blueSeatID, "judge").ingest().
		bracket(m31Logic, "logic-0").mint(m31Logic, m31Closed, "high").stop("logic-0").
		bracket("red-chair", "chair-a").dispatch(2, m31Logic, m31Closed).dispatch(2, blueSeatID, m31Closed).stop("chair-a").
		bracket(m31Logic, "logic-a").mintSuperseding(m31Logic, m31Successor, "high", m31Closed).closeSuperseded(m31Logic, m31Closed, m31Successor).
		closeGap(m31Logic, m31Successor).stop("logic-a").
		bracket(blueSeatID, "blue-a")
	run := b.seed()
	w := workOfBlue(t, run)
	if got := w.Engaged.FoundClosed[0].Successors; len(got) != 0 {
		t.Errorf("successors = %v, want none: %s is closed", got, m31Successor)
	}
	if blocking(w, "record `engage`") {
		t.Errorf("an engage is owed on a closed successor: %+v", w.Sitting.Open)
	}
	if err := engage(t, run, m31Successor); err == nil {
		t.Error("an engage on a closed successor was recorded")
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
	if err := engage(t, run, m31Successor); err == nil || !strings.Contains(err.Error(), "not sitting a sitting for a dispatch") {
		t.Errorf("engage from a sitting no dispatch answers = %v, want refused", err)
	}

	for _, c := range []struct {
		name string
		gaps []string // the next epoch's dispatch naming blue; nil: the chair dispatches logic alone
	}{{"the chair has sat since and dispatched blue", []string{m31Successor}}, {"the chair has sat since and dispatched others", nil}} {
		t.Run(c.name, func(t *testing.T) {
			b2, run2 := m31Epoch10(t)
			next := b2.more().stop("blue-a").bracket("red-chair", "chair-b").dispatch(2, m31Logic, m31Successor)
			if c.gaps != nil {
				next = next.dispatch(2, blueSeatID, c.gaps...)
			}
			recordtest.Seed(t, run2.Dir(), next.stop("chair-b").evs...)
			if err := engage(t, run2, m31Successor); err == nil || !strings.Contains(err.Error(), "not sitting a sitting for a dispatch") {
				t.Errorf("engage after the chair sat again = %v, want refused", err)
			}
		})
	}
}
