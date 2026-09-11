package seatprobe

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A SEAT THAT CORRECTED ITS LOG ENTRY SAID ONE THING, and the probe reads it once — the wording it
// meant — and counts no `correction` verb, because no surface offers one: the seat re-ran `log`.
func TestACorrectedLogIsReadOnceAndIsNotAVerb(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	run := runtest.Open(t, runDir)
	id := record.Identity{Run: run, SeatID: "blue-respond"}
	if _, _, err := record.RegisterSeat(id, ""); err != nil {
		t.Fatal(err)
	}
	entry := func(text string) *recordpb.Log {
		return &recordpb.Log{Text: proto.String(text), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
	}
	first, err := record.Append(id, entry("the tool  refused"))
	if err != nil {
		t.Fatal(err)
	}
	fix := id
	fix.Correct = &record.Correct{Type: recordpb.EventType_EVENT_TYPE_LOG, Key: first.GetKey(), Why: "a word was lost"}
	if _, err := record.Append(fix, entry("the tool refused the cite")); err != nil {
		t.Fatal(err)
	}

	c, err := Read(NewSurface(nil), run, "blue-respond")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Friction) != 1 || c.Friction[0] != "the tool refused the cite" {
		t.Errorf("friction = %q, want only the corrected wording", c.Friction)
	}
	if n := c.Used["correction"]; n != 0 {
		t.Errorf("the probe counted %d use(s) of a `correction` verb no surface offers", n)
	}
	if n := c.Used["log"]; n != 2 {
		t.Errorf("log counted %d time(s), want 2 — the seat ran it twice", n)
	}
}
