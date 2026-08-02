package scorecard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

func seedReport(t *testing.T, runDir, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func anchorEv(id string) record.Event {
	return record.Event{Type: "anchor", Payload: record.NewPayload().Set("id", id).Set("location", "§x")}
}

// The immortal-marker tampering detector: an anchored finding id absent from the report
// is a hard violation (blue dropped red's marker); a present one is fine; an orphan
// marker in the report with no anchor event is NOT expected, so it never trips.
func TestDroppedFindingMarkersDetector(t *testing.T) {
	// f-a present, f-b MISSING, plus an orphan f-orphan in the report with no anchor.
	runDir := t.TempDir()
	seedReport(t, runDir, "A claim.<!--fx:f-a--> Stray.<!--fx:f-orphan-->")
	board := &record.Board{Events: []record.Event{anchorEv("f-a"), anchorEv("f-b")}}

	r := rowByMetric(blueRows(runDir, nil, nil, board), "dropped_finding_markers")
	if r == nil || r.Value == nil {
		t.Fatalf("row not computed: %+v", r)
	}
	if v, _ := r.Value.(int); v != 1 {
		t.Errorf("dropped = %v, want 1 (f-b anchored but missing; f-a present; f-orphan not expected)", r.Value)
	}
}

// All anchored markers present → no hit.
func TestDroppedFindingMarkersAllPresent(t *testing.T) {
	runDir := t.TempDir()
	seedReport(t, runDir, "One.<!--fx:f-a--> Two.<!--fx:f-b-->")
	board := &record.Board{Events: []record.Event{anchorEv("f-a"), anchorEv("f-b")}}
	r := rowByMetric(blueRows(runDir, nil, nil, board), "dropped_finding_markers")
	if v, _ := r.Value.(int); v != 0 {
		t.Errorf("dropped = %v, want 0 (both markers present)", r.Value)
	}
}
