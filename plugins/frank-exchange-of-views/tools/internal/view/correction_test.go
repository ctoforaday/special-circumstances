package view

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// `show debate` SHOWS A CORRECTED POSITION STRUCK, then the position that replaced it.
func TestShowDebateShowsACorrectedPositionStruck(t *testing.T) {
	k := "blue-respond:position:#1"
	evs := []*record.Event{
		recordtest.At(t, "blue-respond", k, &recordpb.Position{Text: proto.String("the report is  now")}),
		recordtest.At(t, "blue-respond", k+"~1", &recordpb.Position{Text: proto.String("the report is sound now")}),
		recordtest.At(t, "blue-respond", "blue-respond:correction:"+k, &recordpb.Correction{
			Corrects: proto.String(k), Replacement: proto.String(k + "~1"), Why: proto.String("a word was lost")}),
	}
	out := string(debateMD(inputT(t, evs...)))
	struck := strings.Index(out, "~~the report is  now~~ (struck by blue-respond: a word was lost)")
	stands := strings.Index(out, "the report is sound now")
	if struck < 0 || stands < 0 || stands < struck {
		t.Errorf("show debate must list the struck position, marked, and then the one that stands:\n%s", out)
	}
}

// THE LINES-OF-AVENUE LISTING SHOWS A CORRECTED PROPOSAL'S WORDING STRUCK under the line.
func TestAvenuesShowTheStruckWording(t *testing.T) {
	prop := func(line string) *recordpb.Avenue {
		return &recordpb.Avenue{AvenueId: proto.String("Q1"), Line: proto.String(line), Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum()}
	}
	k := "blue-respond:avenue:#1"
	evs := []*record.Event{
		recordtest.At(t, "blue-respond", k, prop("try the  method")),
		recordtest.At(t, "blue-respond", k+"~1", prop("try the recorded method")),
		recordtest.At(t, "blue-respond", "blue-respond:correction:"+k, &recordpb.Correction{
			Corrects: proto.String(k), Replacement: proto.String(k + "~1"), Why: proto.String("a word was lost")}),
	}
	in := inputT(t, evs...)
	out := AvenueBody(in.Events, in.At)
	if !strings.Contains(out, "Q1 try the recorded method") || !strings.Contains(out, "~~try the  method~~ (struck by blue-respond: a word was lost)") {
		t.Errorf("the listing must carry the line that stands and the struck wording under it:\n%s", out)
	}
}

// inputT is the render input over evs as the write path stores them: every epoch and `seat #N` a
// view prints is the stored window, so a view's fixture is seeded and loaded back.
func inputT(t *testing.T, evs ...*record.Event) Input {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	in, err := InputOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	return in
}
