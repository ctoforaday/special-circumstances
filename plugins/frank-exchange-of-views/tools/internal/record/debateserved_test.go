package record

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE DEBATE VIEW LISTS EVERY SEAT'S DELIVERY, not a lens's alone: the chair and the bench are handed
// report text on the same chain (gblock 2026-11-13), so the view the bench reads names them too. It
// reads them off the report_delivery rows a narrowed read must load with its report reads.
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

// A SERVED ROW SAYS WHOSE DELIVERY IT IS AND WHETHER IT WAS THE WHOLE REPORT: a lens's first two are
// whole — its first whole delivery and its full re-read — and its third, the changes since its pin, is
// not; the chair's stale areas' changes and the bench's convened motions' passages never are. Which of
// them a row is rides on report_delivery's "whole" being NULL for a seat that is no lens, so a reader
// that counts the chair's or the bench's first delivery as a lens's first marks it whole.
func TestAServedRowSaysWhoseDeliveryItIsAndWhetherItWasWhole(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).sit(2, evLens).served("red-chair", 2).
		benchRegister(recordpb.Occasion_OCCASION_DOCKET).add(benchSeat, &recordpb.ReportRead{Head: proto.Int64(2)})
	debate, err := debateView.of(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	var got []DebateServedJSON
	for _, ep := range debate.Epochs {
		got = append(got, ep.Served...)
	}
	want := []DebateServedJSON{
		{SeatID: evLens, Head: 2, Whole: true},
		{SeatID: evLens, Head: 2, Whole: true},
		{SeatID: evLens, Head: 2, Whole: false},
		{SeatID: "red-chair", Head: 2, Whole: false},
		{SeatID: benchSeat, Head: 2, Whole: false},
	}
	if !slices.Equal(got, want) {
		t.Errorf("the debate view serves %+v, want %+v", got, want)
	}
}
