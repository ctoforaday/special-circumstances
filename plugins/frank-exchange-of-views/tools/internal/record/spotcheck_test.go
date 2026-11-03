package record

import (
	"sort"
	"strings"
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
	checks := SpotChecks(b)
	if len(checks) != 1 || checks[0].Archived != 1 || checks[0].Epoch != 2 {
		t.Fatalf("the spot-check carries the board's count of what was available at its epoch's start, want 1 in epoch 2: %+v", checks)
	}
	if got := checks[0].Describe(); got == "" {
		t.Error("a spot-check must render for the reader")
	}
}

// A SPOT-CHECK'S PROSE IS PRINTED WHATEVER IT NAMES. Over a converged board the chair records no
// verdict and states a defect it found in a stale area in that area's spot-check, which names areas
// and no closure — so a rendering that keeps the prose only beside named closures delivers the
// defect to no reader.
func TestADescribedSpotCheckCarriesItsProseWhateverItNames(t *testing.T) {
	for _, c := range []struct {
		name string
		sc   SpotCheck
		want []string
	}{
		{"stale areas only", SpotCheck{Sitting: 4, SeatID: "red-chair", Areas: []string{"red-lens-logic", "red-lens-voice"}, Prose: "the new paragraph contradicts the table"},
			[]string{"#4 (red-chair)", "read stale area(s) red-lens-logic, red-lens-voice", "the new paragraph contradicts the table"}},
		{"stale areas and closures", SpotCheck{Sitting: 4, SeatID: "red-chair", Areas: []string{"red-lens-logic"}, Sampled: []string{"G1"}, Archived: 3, Prose: "G1's anchor resolves"},
			[]string{"read stale area(s) red-lens-logic", "re-verified G1 of 3 archived closure(s)", "G1's anchor resolves"}},
		{"closures only", SpotCheck{Sitting: 4, SeatID: "red-chair", Sampled: []string{"G1"}, Archived: 1, Prose: "G1's anchor resolves"},
			[]string{"re-verified G1 of 1 archived closure(s)", "G1's anchor resolves"}},
		{"naming nothing", SpotCheck{Sitting: 4, SeatID: "red-chair", Prose: "read the plan"},
			[]string{"naming no stale area and no closure", "read the plan"}},
	} {
		got := c.sc.Describe()
		for _, w := range c.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s: Describe() = %q, missing %q", c.name, got, w)
			}
		}
	}
}

func TestSpotChecksHandlesANilBoard(t *testing.T) {
	if c := SpotChecks(NewFamily(nil, Merged{})); c != nil {
		t.Errorf("a nil board must not panic or invent spot-checks: %v", c)
	}
}

// chairActs is an act of the chair's in the sitting it is in — what makes an epoch one the chair
// worked in.
func chairActs(t *testing.T) *Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("the chair acts")})
}
