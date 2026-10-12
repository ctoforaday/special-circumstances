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

// A spot-check carries the board's count of what was available to sample when its epoch opened: a
// closure made in an earlier epoch counts, one made in the spot-check's own epoch does not.
func TestASpotCheckCarriesTheBoardsCountOfTheArchive(t *testing.T) {
	b := spotBoard(t, map[string]*Gap{"G1": closedIn(1), "G2": closedIn(2)},
		chairSits(t), chairSits(t),
		chairActs(t),
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{Ids: []string{"G1"}, Reason: proto.String("the anchor still resolves")}),
	)
	checks, falseEmpty := SpotCheckAudit(b)
	if len(falseEmpty) != 0 {
		t.Fatalf("a real sample claims no empty archive: falseEmpty=%v", falseEmpty)
	}
	if len(checks) != 1 || checks[0].Archived != 1 || checks[0].Epoch != 2 {
		t.Fatalf("the spot-check carries the board's count of what was available at its epoch's start, want 1 in epoch 2: %+v", checks)
	}
	if got := checks[0].Describe(); got == "" {
		t.Error("a spot-check must render for the reader")
	}
}

// "There was nothing to sample" is a claim, and the board contradicts it where it held archived
// closures.
func TestAFalseEmptyClaimIsCaught(t *testing.T) {
	b := spotBoard(t, map[string]*Gap{"G1": closedIn(1)},
		chairSits(t), chairSits(t),
		chairActs(t),
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{None: proto.Bool(true), Reason: proto.String("nothing archived")}),
	)
	_, falseEmpty := SpotCheckAudit(b)
	if len(falseEmpty) != 1 {
		t.Fatalf("a --none claim against a non-empty archive must be caught: %v", falseEmpty)
	}
	if falseEmpty[0].Archived != 1 {
		t.Errorf("the contradiction must carry the board's count: %+v", falseEmpty[0])
	}

	// An HONEST --none, against an archive the board agrees was empty, contradicts nothing.
	b = spotBoard(t, map[string]*Gap{},
		recordtest.Event(t, "red-chair", &recordpb.SpotCheck{None: proto.Bool(true), Reason: proto.String("nothing archived")}),
	)
	if _, fe := SpotCheckAudit(b); len(fe) != 0 {
		t.Errorf("an honestly-empty claim is not a contradiction: %v", fe)
	}
}

func TestSpotCheckAuditHandlesANilBoard(t *testing.T) {
	if c, f := SpotCheckAudit(NewFamily(nil, Merged{})); c != nil || f != nil {
		t.Errorf("a nil board must not panic or invent violations: %v %v", c, f)
	}
}

// chairActs is an act of the chair's in the sitting it is in — what makes an epoch one the chair
// worked in.
func chairActs(t *testing.T) *Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Position{Text: proto.String("the chair acts")})
}
