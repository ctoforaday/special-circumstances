package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE CHAIR IS TOLD WHAT CHANGED BEHIND EACH PIN, NOT ASKED (gblock, 2026-11-10). The verb prints the
// stamp the write put on the spot-check — each area's pin, and the count and event ids of the changes
// to the report after it — in both forms. The run's lens retired at its pin (staleAreaRun) and the
// report moved behind it: the part the fixture appended, and blue's edit here — which `show changes`
// names by the same event id the stamp names it by. Every value is read off the record, never typed.
func TestASpotCheckEchoesTheChangesBehindEachPin(t *testing.T) {
	runDir := staleAreaRun(t)
	registerSeatOnce(t, runDir, "blue-respond")
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond", "--quote", "the parser accepts an empty body in this line.",
		"--new", "the parser accepts an empty body on this line.", "--reason", "r"); err != nil {
		t.Fatalf("blue's edit: %v", err)
	}
	db := openRunDB(t, runDir)
	pin, err := record.PinAt(db, "red-lens-evidence")
	if err != nil || pin == 0 {
		t.Fatalf("the retired lens's pin = %d, %v; want the head of its latest delivery", pin, err)
	}
	var part, edit int64
	if err := db.QueryRow(`SELECT MAX("event_id") FROM "report_part"`).Scan(&part); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT MAX("event_id") FROM "blue_edit"`).Scan(&edit); err != nil {
		t.Fatal(err)
	}
	if part <= pin || edit <= part {
		t.Fatalf("the fixture's changes (part %d, edit %d) are not both behind the pin %d", part, edit, pin)
	}
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
	if !ok || !slices.Equal(sc.GetChanges(), []int64{part, edit}) {
		t.Fatalf("the stored spot-check stamps %v; want the part and the edit behind the pin, %v", sc.GetChanges(), []int64{part, edit})
	}
	a := env.Result.Areas
	if len(a) != 1 || a[0].SeatID != "red-lens-evidence" || a[0].Pin != pin || a[0].Count != 2 || !slices.Equal(a[0].Changes, sc.GetChanges()) {
		t.Fatalf("the echo %+v is not the stamp %v at pin %d", a, sc.GetChanges(), pin)
	}
	log, err := run(t, "show", "changes", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("the change log: %v\n%s", err, log)
	}
	var changes record.ChangesJSON
	if err := json.Unmarshal([]byte(log), &changes); err != nil {
		t.Fatalf("%v\n%s", err, log)
	}
	if len(changes.Edits) != 1 || changes.Edits[0].EventID != edit {
		t.Errorf("the change log names the edit %+v; the stamp names it as event %d", changes.Edits, edit)
	}
	// THE HUMAN FORM says the same, on a run staged the same way without the edit.
	humanDir := staleAreaRun(t)
	humanPin, err := record.PinAt(openRunDB(t, humanDir), "red-lens-evidence")
	if err != nil {
		t.Fatal(err)
	}
	human, err := run(t, "spot-check", "--run", humanDir, "--seat-id", "red-chair", "--areas", "red-lens-evidence",
		"--reason", "the changes keep every cited sentence on its source")
	if err != nil {
		t.Fatalf("spot-check: %v\n%s", err, human)
	}
	if want := fmt.Sprintf("pin %d: 1 change(s) to the report since", humanPin); !strings.Contains(human, want) {
		t.Errorf("the human echo does not state the record's count %q: %q", want, human)
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
