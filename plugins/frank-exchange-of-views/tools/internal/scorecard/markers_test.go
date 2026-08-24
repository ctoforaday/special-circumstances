package scorecard

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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

func anchorEv(t *testing.T, id string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "", 0, &recordpb.Anchor{Id: proto.String(id), Location: proto.String("§x")})
}

// The immortal-marker tampering detector: an anchored finding id absent from the report
// is a hard violation (blue dropped red's marker); a present one is fine; an orphan
// marker in the report with no anchor event is NOT expected, so it never trips.
func TestDroppedFindingMarkersDetector(t *testing.T) {
	// f-a present, f-b MISSING, plus an orphan f-orphan in the report with no anchor.
	runDir := t.TempDir()
	seedReport(t, runDir, "A claim.<!--fx:f-a--> Stray.<!--fx:f-orphan-->")
	board := &record.Board{Events: []*record.Event{anchorEv(t, "f-a"), anchorEv(t, "f-b")}}

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
	board := &record.Board{Events: []*record.Event{anchorEv(t, "f-a"), anchorEv(t, "f-b")}}
	r := rowByMetric(blueRows(runDir, nil, nil, board), "dropped_finding_markers")
	if v, _ := r.Value.(int); v != 0 {
		t.Errorf("dropped = %v, want 0 (both markers present)", r.Value)
	}
}

func citeEv(t *testing.T, label string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "", 0, &recordpb.Cite{Label: proto.String(label), Url: proto.String("https://x"), Title: proto.String("T")})
}

// §V.6 — the unbacked_citations detector: a cite event whose anchor is gone from the
// report breaks the bijection (a hand-typed footnote or a tampered anchor); a fully
// bijective report scores 0.
func TestUnbackedCitationsDetector(t *testing.T) {
	// c-a present, c-b MISSING (its cite event exists but the anchor is gone).
	runDir := t.TempDir()
	seedReport(t, runDir, "A cited claim<!--cite:c-a-->. Another claim with a hand-typed [^b] footnote.")
	board := &record.Board{Events: []*record.Event{citeEv(t, "c-a"), citeEv(t, "c-b")}}

	r := rowByMetric(blueRows(runDir, nil, nil, board), "unbacked_citations")
	if r == nil || r.Value == nil {
		t.Fatalf("row not computed: %+v", r)
	}
	if v, _ := r.Value.(int); v != 1 {
		t.Errorf("unbacked = %v, want 1 (c-b has a cite event but no anchor; c-a present)", r.Value)
	}
}

// A fully bijective report — every cite event's anchor present — scores 0.
func TestUnbackedCitationsBijectiveIsZero(t *testing.T) {
	runDir := t.TempDir()
	seedReport(t, runDir, "One<!--cite:c-a-->. Two<!--cite:c-b-->.")
	board := &record.Board{Events: []*record.Event{citeEv(t, "c-a"), citeEv(t, "c-b")}}
	r := rowByMetric(blueRows(runDir, nil, nil, board), "unbacked_citations")
	if v, _ := r.Value.(int); v != 0 {
		t.Errorf("unbacked = %v, want 0 (both citation anchors present)", r.Value)
	}
}
