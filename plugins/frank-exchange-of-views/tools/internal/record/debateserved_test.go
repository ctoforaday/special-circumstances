package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE DEBATE VIEW LISTS EVERY SEAT'S DELIVERY, not a lens's alone: the chair and the bench are handed
// report text on the same chain (gblock 2026-11-13), and debate.md renders each of their deliveries,
// so the view the bench reads names them too. It reads them off the report_delivery rows a narrowed
// read must load with its report reads.
func TestTheDebateViewListsEverySeatsDelivery(t *testing.T) {
	b := retBoard(t).sit(2, evLens).served("red-chair", 2).
		benchRegister(recordpb.Occasion_OCCASION_DOCKET).add(benchSeat, &recordpb.ReportRead{Head: proto.Int64(2)})
	debate, err := debateView.of(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, ep := range debate.Epochs {
		for _, s := range ep.Served {
			got[s.SeatID] = true
		}
	}
	for _, seat := range []string{evLens, "red-chair", benchSeat} {
		if !got[seat] {
			t.Errorf("the debate view lists no delivery to %s; it lists %v", seat, got)
		}
	}
}
