package report

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE REPORT TEXT DELIVERED TO A LENS IS ON THE DEBATE'S JSON: which lens had which delivery completed
// at which head — delivered, which is all the tool can see.
func TestTheDebateSaysWhichLensWasDeliveredWhatAtWhichHead(t *testing.T) {
	const lens = "red-lens-evidence"
	evs := []*record.Event{
		recordtest.Event(t, "harness", &recordpb.ReportPart{Text: proto.String("# report")}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, lens, &recordpb.Register{}),
		recordtest.Event(t, lens, &recordpb.ReportRead{Head: proto.Int64(1)}),
	}
	f := famOf(t, evs)
	j := record.DebateJSONOfFamily(f)
	var served []record.DebateServedJSON
	for _, e := range j.Epochs {
		served = append(served, e.Served...)
	}
	if len(served) != 1 || served[0] != (record.DebateServedJSON{SeatID: lens, Head: 1, Whole: true}) {
		t.Errorf("the debate's JSON serves %+v, want %s at head 1", served, lens)
	}
}

// A LENS'S THIRD DELIVERY IS THE CHANGES SINCE ITS PIN: its first whole delivery and its full re-read
// behind it, the debate's JSON marks the third not whole.
func TestTheDebateMarksALensThirdDeliveryNotWhole(t *testing.T) {
	const lens = "red-lens-evidence"
	evs := []*record.Event{
		recordtest.Event(t, "harness", &recordpb.ReportPart{Text: proto.String("# report")}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
	}
	for range 3 {
		evs = append(evs, recordtest.Event(t, lens, &recordpb.Register{}),
			recordtest.Event(t, lens, &recordpb.ReportRead{Head: proto.Int64(1)}))
	}
	f := famOf(t, evs)
	var whole []bool
	for _, e := range record.DebateJSONOfFamily(f).Epochs {
		for _, s := range e.Served {
			whole = append(whole, s.Whole)
		}
	}
	if !slices.Equal(whole, []bool{true, true, false}) {
		t.Errorf("the debate's JSON marks the deliveries whole = %v, want [true true false]", whole)
	}
}
