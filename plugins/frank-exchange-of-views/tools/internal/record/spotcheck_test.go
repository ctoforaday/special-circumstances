package record

import (
	"sort"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// board builds a Board with the gaps and events a spot-check audit reads.
func spotBoard(t *testing.T, gaps map[string]*Gap, evs ...*Event) Family {
	t.Helper()
	// Sorted ids so the fixture family is deterministic, as the fold's GapOrder was.
	var order []string
	for id := range gaps {
		order = append(order, id)
	}
	sort.Strings(order)
	var ordered []*Gap
	for _, id := range order {
		g := gaps[id]
		if g != nil && g.ID == "" {
			g.ID = id
		}
		ordered = append(ordered, g)
	}
	return loadedFamilyT(t, ordered, evs)
}

func closedIn(epoch int) *Gap { return &Gap{HasClosed: true, ClosedEpoch: epoch, Open: false} }

// THE FLOOR IS OWED ONLY WHERE THERE WAS SOMETHING TO SAMPLE. W1.8 keyed it on the archive's
// state at round START precisely because a round-NUMBER rule degenerated: run 5's round-2 merge
// entered with an empty archive, so "from round 2" became the seat attesting blocks it was about
// to write itself.
func TestSpotCheckDebtOnlyWhereTheArchiveWasNonEmpty(t *testing.T) {
	// A gap archived in epoch 1; the chair sits twice and acts in epoch 2 without sampling it.
	b := spotBoard(t, map[string]*Gap{"G1": closedIn(1)},
		chairSits(t), chairSits(t),
		chairActs(t),
		chairActs(t),
	)
	_, debt, _ := SpotCheckAudit(b)
	if len(debt) != 1 || debt[0] != 2 {
		t.Fatalf("epoch 2 entered with an archived closure and sampled none — expected debt [2], got %v", debt)
	}

	// The SAME shape with the closure landing IN epoch 2 owes nothing: the archive was empty
	// when the epoch opened, which is the whole point of keying on its start.
	b = spotBoard(t, map[string]*Gap{"G2": closedIn(2)},
		chairSits(t), chairSits(t),
		chairActs(t),
	)
	if _, debt, _ := SpotCheckAudit(b); len(debt) != 0 {
		t.Errorf("a closure made DURING the epoch was not in the archive at its start; no debt is owed: %v", debt)
	}

	// An epoch the chair never ACTED in owes nothing — the chair's second register opens it, but
	// only blue does anything there. Demanding a sample from an absent seat is the round-number
	// keying W1.8 replaced, in a new spelling.
	b = spotBoard(t, map[string]*Gap{"G1": closedIn(1)},
		chairSits(t),
		chairActs(t),
		chairSits(t),
		recordtest.Event(t, "blue-respond", &recordpb.Position{Text: proto.String("blue acts")}),
	)
	if _, debt, _ := SpotCheckAudit(b); len(debt) != 0 {
		t.Errorf("the chair did not act in epoch 2; no duty was skipped: %v", debt)
	}
}

// A discharge clears the debt.
func TestSpotCheckDischargeClearsTheDebt(t *testing.T) {
	b := spotBoard(t, map[string]*Gap{"G1": closedIn(1)},
		chairSits(t), chairSits(t),
		chairActs(t),
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{Ids: []string{"G1"}, Reason: proto.String("the anchor still resolves")}),
	)
	checks, debt, falseEmpty := SpotCheckAudit(b)
	if len(debt) != 0 || len(falseEmpty) != 0 {
		t.Fatalf("a real sample discharges the duty: debt=%v falseEmpty=%v", debt, falseEmpty)
	}
	if len(checks) != 1 || checks[0].Archived != 1 {
		t.Fatalf("the discharge must carry the board's count of what was available: %+v", checks)
	}
	if got := checks[0].Describe(); got == "" {
		t.Error("a discharge must render for the reader")
	}
}

// THE DIRECT HEIR OF THE RUN-5 DEGENERACY. "There was nothing to sample" is a claim, and the
// board can now refuse it. Every repair before this one asked the seat for the number it was
// being checked against.
func TestAFalseEmptyClaimIsCaught(t *testing.T) {
	b := spotBoard(t, map[string]*Gap{"G1": closedIn(1)},
		chairSits(t), chairSits(t),
		chairActs(t),
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{None: proto.Bool(true), Reason: proto.String("nothing archived")}),
	)
	_, debt, falseEmpty := SpotCheckAudit(b)
	if len(falseEmpty) != 1 {
		t.Fatalf("a --none claim against a non-empty archive must be caught: %v", falseEmpty)
	}
	if falseEmpty[0].Archived != 1 {
		t.Errorf("the contradiction must carry the board's count: %+v", falseEmpty[0])
	}
	// It is a FALSE CLAIM, not a missing one — reporting it as debt too would double-count the
	// same round under two different accusations.
	if len(debt) != 0 {
		t.Errorf("an epoch that recorded something is not one that recorded nothing: %v", debt)
	}

	// An HONEST --none, against an archive the board agrees was empty, is a discharge.
	b = spotBoard(t, map[string]*Gap{},
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{None: proto.Bool(true), Reason: proto.String("nothing archived")}),
	)
	if _, _, fe := SpotCheckAudit(b); len(fe) != 0 {
		t.Errorf("an honestly-empty epoch is a discharge, not a violation: %v", fe)
	}
}

func TestSpotCheckAuditHandlesANilBoard(t *testing.T) {
	if c, d, f := SpotCheckAudit(NewFamily(nil, Merged{})); c != nil || d != nil || f != nil {
		t.Errorf("a nil board must not panic or invent violations: %v %v %v", c, d, f)
	}
}

// chairActs is an act of the chair's in the sitting it is in — what makes an epoch one the chair
// worked in.
func chairActs(t *testing.T) *Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Position{Text: proto.String("the chair acts")})
}
