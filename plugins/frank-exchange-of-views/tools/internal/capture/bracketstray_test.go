package capture

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// A SEAT NO ROW NAMED THAT SAT BY ITS BRACKET ALONE IS A STRAY (#1262): the hook's bracket opens its
// sitting, the SubagentStart writer's own write, and the parity audit reads sittings, not registers.
func TestASeatThatSatByItsBracketAloneIsAStray(t *testing.T) {
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	reg := func(seat string) *record.Event { return at(seat, &recordpb.Register{}) }
	dispatch := func(seat string) *record.Event {
		return at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(seat)})
	}
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-voice", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		reg("red-chair"), dispatch("red-lens-evidence"),
		reg("red-lens-evidence"))
	if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: dir, Phase: sittingwrite.Open,
		AgentID: "agent_voice", AgentType: "frank-exchange-of-views:red-lens-voice"}, io.Discard); err != nil {
		t.Fatalf("the hook's bracket: %v", err)
	}
	if a := DispatchParityAudit(runtest.Open(t, dir), nil, false); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "red-lens-voice sat after dispatch 1 and was not a party to it") {
		t.Fatalf("a seat nobody dispatched sat by its bracket unnoticed: %s: %s", a.Verdict, a.Detail)
	}
}
