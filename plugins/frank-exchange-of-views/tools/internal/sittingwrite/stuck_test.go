package sittingwrite

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A DELIVERY OF REPORT TEXT THE WRITER CANNOT STORE IS STUCK, AND SAYS SO TWICE: the writer returns
// hookgate.ErrDeliveryStuck, which ends the seat's turn, and the record holds the tool's DEFECT log
// naming why. Here the sitting's delivery file cannot be kept — the delivery directory is a file — so
// neither its lock nor its write can happen; a seat told "could not be read" instead would wait on an
// envelope its report text never reaches.
func TestADeliveryThatCannotBeStoredIsStuckAndLogged(t *testing.T) {
	dir := newRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{
			SeatIds: []string{"red-chair", "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest",
			&recordpb.ReportPart{Text: proto.String("# report\n\nOne paragraph.\n")}),
		recordtest.At(t, "red-chair", "chair:dispatch", &recordpb.Dispatch{
			SeatId: proto.String("red-lens-evidence"), Pin: proto.Int64(1)}))
	const lens = "frank-exchange-of-views:red-lens-evidence"
	if err := Write(Sitting{RunDir: dir, Phase: Open, AgentID: "agent_01", AgentType: lens}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, deliveryDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, deliveryDir), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := WriteNextItem(dir, "agent_01", lens, "claude", &out)
	if !errors.Is(err, hookgate.ErrDeliveryStuck) {
		t.Fatalf("a delivery that cannot be stored returned %v — want hookgate.ErrDeliveryStuck", err)
	}
	if out.Len() != 0 {
		t.Errorf("a stuck delivery handed the seat text: %s", out.String())
	}
	logs := 0
	evs, _, err := record.EventsOf(mustRun(t, dir), recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs {
		if l, _ := recordpb.BodyAs[*recordpb.Log](e); l != nil && l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL && l.GetType() == recordpb.LogType_LOG_TYPE_DEFECT &&
			strings.Contains(l.GetText(), "could not advance") {
			logs++
		}
	}
	if logs != 1 {
		t.Errorf("the record holds %d tool DEFECT log(s) for the stuck delivery, want 1", logs)
	}
}
