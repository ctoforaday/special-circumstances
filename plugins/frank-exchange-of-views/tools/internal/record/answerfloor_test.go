package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN AMEND MOVES THE ANSWER FLOOR (§III.B3, T-B54's floor row): an answer the author gave before the
// raiser amended the gap answers a wording the gap no longer states, so it is read as no answer, as
// one before a remand is; an answer after the amend is read. A regrade moves no floor.
func TestAnAmendMovesTheAnswerFloor(t *testing.T) {
	staged := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").sit(2, evLens).sit(2, evLens).mint(evLens, "G1", "low").register("red-chair").
			blueSitsOn("G1", true)
	}
	answered := func(b *stage) bool {
		evs, win := mustEvents(t, b.seed())
		a, ok := blueAnswersOf(evs, win, map[string]bool{"G1": true}, motionsAt(evs, win.IDs(evs), win))["G1"]
		return ok && a.answered()
	}
	if !answered(staged()) {
		t.Fatal("blue's sitting engaged on G1 is not read as its answer: the fixture answers nothing")
	}
	amended := staged().register(evLens).add(evLens, &recordpb.Amend{GapId: proto.String("G1"), Problem: proto.String("restated"), Basis: proto.String("narrower")})
	if answered(amended) {
		t.Error("blue's answer to G1 from before its minter amended it is read as an answer to the amended demand")
	}
	regraded := staged().register(evLens).add(evLens, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordpb.Grade_GRADE_HIGH.Enum(), Basis: proto.String("heavier")})
	if !answered(regraded) {
		t.Error("a regrade moved G1's answer floor: it changes no fix or check an answer must meet")
	}
}
