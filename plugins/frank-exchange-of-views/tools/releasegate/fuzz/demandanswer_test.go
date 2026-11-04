package fuzz

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// demandBoard is a runner over one lane's part, red's seats registered, and a gap of each check kind
// that demands a live answer: kinds["citation"] a `source` gap, kinds["computation"] a `computation` gap.
func demandBoard(t *testing.T) (*runner, map[string]string) {
	t.Helper()
	r := newRunner(buildBinary(t), recordtest.TmpRun(t), newLockedRand(7))
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: r.run(), SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatal(err)
	}
	if err := r.stagePart("A § fuzz sentence to anchor findings.\n\nThe cost is rising over time across the whole sampled corpus.\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	r.register("chair", "red-chair")
	// `mint` draws the check kind, so take gaps until one of each demanding kind stands. Bounded, so a
	// change that stops drawing them fails here rather than passing on nothing.
	kinds := map[string]string{}
	for i := 0; i < 60 && len(kinds) < 2; i++ {
		lens := fuzzLensSeats[i%len(fuzzLensSeats)]
		r.register("lens", lens)
		id := r.mint(lens)
		if kind, _ := r.demandKindOf(id); id != "" && kind != "" && kinds[kind] == "" {
			kinds[kind] = id
		}
	}
	if len(kinds) < 2 || r.actErr != "" {
		t.Fatalf("after 60 mints the demanding gaps are %v (%s): this test cannot exercise both answers", kinds, r.actErr)
	}
	return r, kinds
}

// THE DRIVE ANSWERS EACH DEMAND BEFORE IT CLOSES. closeGap answers a `source` gap with a citation and a
// `computation` gap with a computation placed as its answer, so the lens's close is admitted; with
// either arm gone, that kind's close is refused by the live-answer demand and the gap stays open.
func TestTheDriveAnswersEachDemandBeforeItCloses(t *testing.T) {
	r, kinds := demandBoard(t)
	for kind, g := range kinds {
		r.closeGap("red-chair", g, false)
		if r.actErr != "" {
			t.Fatalf("closeGap of %s gap %s: %s", kind, g, r.actErr)
		}
		answered, err := record.EvidenceAnswers(r.run(), g)
		if err != nil || !answered {
			t.Errorf("closeGap left %s gap %s's demand unmet (%v)", kind, g, err)
		}
		for _, open := range r.openGaps() {
			if open == g {
				t.Errorf("closeGap left %s gap %s open", kind, g)
			}
		}
	}
}

// BLUE DISCHARGES THE DEMAND ITS SITTING IS ENGAGED ON. Blue's work list holds an item for each
// engaged gap awaiting evidence, and its envelope is refused while one stands: dischargeOwed places a
// citation as a `source` gap's answer and a computation as a `computation` gap's.
func TestBlueDischargesTheDemandOfEachEngagedGap(t *testing.T) {
	r, kinds := demandBoard(t)
	gaps := []string{kinds["citation"], kinds["computation"]}
	_, head, err := record.ReportAtHead(r.run())
	if err != nil || head == 0 {
		t.Fatalf("the report head: %d, %v", head, err)
	}
	if _, err := record.Append(record.Identity{Run: r.run(), SeatID: "red-chair"},
		&recordpb.Dispatch{Pin: proto.Int64(head), SeatId: proto.String("blue-respond"), GapIds: gaps}); err != nil {
		t.Fatalf("dispatch blue on %v: %v", gaps, err)
	}
	r.register("blue", "blue-respond")
	r.dischargeOwed("blue-respond")
	for kind, g := range kinds {
		if answered, err := record.EvidenceAnswers(r.run(), g); err != nil || !answered {
			t.Errorf("blue's sitting left %s gap %s's demand unmet (%v; %s)", kind, g, err, r.actErr)
		}
	}
}
