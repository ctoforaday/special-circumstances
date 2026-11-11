package cli

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE CHAIR IS TOLD WHAT CHANGED BEHIND EACH PIN, NOT ASKED (gblock, 2026-11-10). The verb prints the
// stamp the write put on the spot-check — each area's pin, and the count and event ids of the changes
// to the report after it — in both forms. The staged lens was never served the report, so its pin is
// 0 and every change on the record is behind it: the ingested part (event 2), the part added here and
// blue's edit — which `show changes` names by the same event id the stamp names it by.
func TestASpotCheckEchoesTheChangesBehindEachPin(t *testing.T) {
	runDir := newRun(t)
	stageEvents(t, runDir, recordtest.At(t, "harness", "harness:stage:part", &recordpb.ReportPart{Text: proto.String("## more")}),
		recordtest.At(t, "blue-respond", "blue-respond:stage:edit", &recordpb.BlueEdit{Old: proto.String("more"), New: proto.String("most"), Text: proto.String("r")}))
	out, err := run(t, "spot-check", "--run", runDir, "--seat-id", "red-chair", "--areas", "red-lens-evidence",
		"--reason", "the changes keep every cited sentence on its source", "--json")
	if err != nil {
		t.Fatalf("spot-check: %v\n%s", err, out)
	}
	var env struct {
		Result struct {
			Areas []struct {
				SeatID  string  `json:"seat_id"`
				Pin     int64   `json:"pin"`
				Count   int     `json:"count"`
				Changes []int64 `json:"changes"`
			} `json:"areas"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	sc, ok := recordpb.BodyAs[*recordpb.SpotCheck](lastEventOf(t, runDir, recordpb.EventType_EVENT_TYPE_SPOT_CHECK))
	if !ok || len(sc.GetChanges()) != 3 {
		t.Fatalf("the stored spot-check stamps %v; want the two parts and the edit", sc.GetChanges())
	}
	a := env.Result.Areas
	if len(a) != 1 || a[0].SeatID != "red-lens-evidence" || a[0].Pin != 0 || a[0].Count != 3 || !slices.Equal(a[0].Changes, sc.GetChanges()) {
		t.Fatalf("the echo %+v is not the stamp %v at pin 0", a, sc.GetChanges())
	}
	log, err := run(t, "show", "changes", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("the change log: %v\n%s", err, log)
	}
	var changes record.ChangesJSON
	if err := json.Unmarshal([]byte(log), &changes); err != nil {
		t.Fatalf("%v\n%s", err, log)
	}
	if edit := sc.GetChanges()[2]; len(changes.Edits) != 1 || changes.Edits[0].EventID != edit {
		t.Errorf("the change log names the edit %+v; the stamp names it as event %d", changes.Edits, edit)
	}
	// THE HUMAN FORM says the same, on a run staged the same way.
	humanDir := newRun(t)
	stageEvents(t, humanDir, recordtest.At(t, "harness", "harness:stage:part", &recordpb.ReportPart{Text: proto.String("## more")}))
	human, err := run(t, "spot-check", "--run", humanDir, "--seat-id", "red-chair", "--areas", "red-lens-evidence",
		"--reason", "the changes keep every cited sentence on its source")
	if err != nil {
		t.Fatalf("spot-check: %v\n%s", err, human)
	}
	if !strings.Contains(human, "pin 0: 2 change(s) to the report since") {
		t.Errorf("the human echo does not state the record's count: %q", human)
	}
}

// lastEventOf is the latest event of typ on the run's record.
func lastEventOf(t *testing.T, runDir string, typ recordpb.EventType) *record.Event {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), typ)
	if err != nil || len(evs) == 0 {
		t.Fatalf("no %s on the record: %v", typ, err)
	}
	return evs[len(evs)-1]
}
