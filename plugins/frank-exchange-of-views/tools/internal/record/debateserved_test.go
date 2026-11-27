package record

import (
	"strings"
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

// A DELIVERY'S LINE SAYS WHAT ITS SEAT WAS HANDED: a lens's is the whole report or the changes since
// its pin, the chair's is its stale areas' changes, the bench's is its convened motions' passages.
// Which of them a row is rides on report_delivery's "whole" being NULL for a seat that is no lens, so
// a reader that takes NULL for a lens's changes-only delivery words the chair's and the bench's
// lines as a lens's.
func TestADeliveryLineNamesWhatItsSeatWasHanded(t *testing.T) {
	b := retBoard(t).sit(2, evLens).served("red-chair", 2).
		benchRegister(recordpb.Occasion_OCCASION_DOCKET).add(benchSeat, &recordpb.ReportRead{Head: proto.Int64(2)})
	merged, err := MergedEvents(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		evLens:      "the whole report",
		"red-chair": "the changes behind its stale areas' pins",
		benchSeat:   "the passages of the gaps its convened motions are about",
	}
	lines := 0
	for _, e := range merged.Events {
		if e.GetReportRead() == nil {
			continue
		}
		lines++
		seat := e.GetSeatId()
		if got, handed := merged.At.DeliveryLine(e), want[seat]+", delivered to "+seat+" at head 2"; !strings.Contains(got, handed) {
			t.Errorf("the delivery to %s reads %q, want it to say %q", seat, got, handed)
		}
	}
	if lines != len(want) {
		t.Fatalf("the record holds %d deliveries, want one to each of %d seats", lines, len(want))
	}
}
