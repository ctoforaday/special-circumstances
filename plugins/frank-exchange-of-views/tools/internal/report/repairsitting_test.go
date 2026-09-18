package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// blueSatAndRepaired is blue's sitting on G1 — dispatched, registered, edited — its agent's stop,
// and then the sitting-record repair: a second prompt, registering as the repair of that sitting,
// which files the receipt and the retirement the first sitting owed.
func blueSatAndRepaired(t *testing.T) []*record.Event {
	t.Helper()
	return []*record.Event{
		recordtest.At(t, "red-chair", "red-chair:dispatch:d1", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		recordtest.At(t, record.HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "blue-respond", "blue-respond:register:#2", &recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String("blue-respond:register:#1")}),
		recordtest.At(t, "blue-respond", "blue-respond:manifest_row:m1", &recordpb.ManifestRow{GapId: proto.String("G1"), Row: proto.String("recomputed the figure")}),
		recordtest.At(t, "blue-respond", "blue-respond:retire:r1", &recordpb.Retire{Claim: proto.String("the old figure"), Reason: proto.String("superseded"), RemovalBasis: proto.String(record.RemovalAsserted)}),
	}
}

// AN ACT FILED IN A REPAIR IS RENDERED UNDER THE SITTING IT COMPLETES (#1002, gblock 2026-09-18).
// The repair is a sitting — the seat was handed a prompt — but its acts are the repaired sitting's,
// and these two sections attribute acts. Rendered under the turn count they said `blue-respond #2`,
// a sitting the closer bounds nothing for: the receipt for a gap sitting 1 owed, filed by sitting 1,
// shown against a sitting that answers for nothing.
func TestARepairsActsRenderUnderTheSittingTheyComplete(t *testing.T) {
	evs := blueSatAndRepaired(t)
	g := &record.Gap{ID: "G1"}
	fam := (&boardT{GapOrder: []string{"G1"}, Gaps: map[string]*record.Gap{"G1": g}, Events: evs}).fam()

	manifest := correctnessManifest(fam)
	if !strings.Contains(manifest, "**G1** (blue-respond #1): recomputed the figure") {
		t.Errorf("the receipt is not filed under the sitting it completes:\n%s", manifest)
	}
	if strings.Contains(manifest, "#2") {
		t.Errorf("the receipt is rendered under the repair's own turn count:\n%s", manifest)
	}

	withdrawn := withdrawnClaims(evs)
	if !strings.Contains(withdrawn, "(blue-respond #1)") || strings.Contains(withdrawn, "#2") {
		t.Errorf("the retirement is not filed under the sitting it completes:\n%s", withdrawn)
	}
}
