package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The invisible finding-marker "<!--fx:f-<id>-->" is stripped; blue's own claim
// footnotes ("[^L1]") and prose are untouched.
func TestStripFindingMarkers(t *testing.T) {
	in := "A claim.<!--fx:f-abc--> Another one[^L1], and a third.<!--fx:f-9f2a1c-->\n"
	got := StripFindingMarkers(in)
	want := "A claim. Another one[^L1], and a third.\n"
	if got != want {
		t.Errorf("StripFindingMarkers = %q, want %q", got, want)
	}
	if strings.Contains(got, "<!--fx:") {
		t.Error("a raw finding-marker survived the strip")
	}
}

// THE LEAK REPRO: a marker token that reaches the report via a RECORD-DERIVED section
// (here a mint's problem text, rendered into the findings/risk section) must still be
// stripped — the earlier blue-only strip missed exactly this path.
func TestAssembleStripsMarkersFromRecordDerivedSections(t *testing.T) {
	runDir := newRun(t)
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Blue's report is marker-free; the marker enters ONLY through the record.
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte("# Title\n\nClean prose.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"red-merge-r1", "blue-respond-r1", "judge-terminal"} {
		if _, _, err := record.RegisterSeat(record.Identity{RunDir: runDir, SeatID: s, Round: record.RoundIn(runDir)(s)}, ""); err != nil {
			t.Fatal(err)
		}
	}
	// A gap whose problem text carries a marker token (as a real finding's quoted
	// location/reason would), plus a terminal outcome so assembly composes fully.
	p := record.NewPayload().Set("gap_id", "R1-1").
		Set("problem", "the sentence flagged here <!--fx:f-leak12--> is wrong").
		Set("location", "§1").Set("required_fix", "fix it").Set("acceptance_check", "recheck").Set("check_kind", "document").
		Set("class", "correctness").Set("severity", "high").Set("likelihood", "high").Set("impact", "high")
	if _, err := record.Append(record.Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: record.RoundIn(runDir)("red-merge-r1")}, "mint", p); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{RunDir: runDir, SeatID: "judge-terminal", Round: record.RoundIn(runDir)("judge-terminal")}, "outcome", record.NewPayload().Set("verdict", "CEILING").Set("reason", "the round ceiling arrived before red could pass the final revision")); err != nil {
		t.Fatal(err)
	}

	if _, err := Assemble(runDir); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	shipped, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(shipped), "<!--fx:") {
		t.Errorf("a finding-marker leaked into the shipped report via a record-derived section:\n%s", shipped)
	}
}
