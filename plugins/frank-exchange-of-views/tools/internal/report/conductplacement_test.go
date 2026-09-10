package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// WHAT ANSWERED THE SEATS IS ENVELOPE, SO IT IS IN run.md AND NOT IN report.md.
//
// gblock, 2026-09-10: "It's envelope. Move out of report.md." The section sat in report.md "in the same
// breath" as the verdict, and its tool-authored heading was the one voice tell no run could remove
// from the research document (#792, §V item 3). The difftest goldens render report.md only, so after
// the move they showed the section LEAVING and nothing showing it ARRIVE — a section composed
// nowhere would have passed every one of them. This asserts both halves on an assembled run.
//
// ONE SEAT IS REGISTERED ON PURPOSE. conduct() renders nothing for a run with no seats, so a seatless
// fixture would pass this test whichever document the section was composed into.
func TestWhatAnsweredTheSeatsIsInRunMdNotReportMd(t *testing.T) {
	runDir := newRun(t)
	blue := "# Is 91 prime? — research report\n\n## TL;DR\n91 = 7 x 13, so it is composite.\n"
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}
	id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, ""); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(blue)}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(runDir, name))
		if err != nil {
			t.Fatalf("assemble did not write %s: %v", name, err)
		}
		return string(b)
	}
	report, runmd := read(FileReport), read(FileRun)

	const heading = "## How this run was conducted"
	if strings.Contains(report, heading) {
		t.Errorf("report.md still carries the conduct section — it is envelope, not research:\n%s", report)
	}
	if !strings.Contains(runmd, heading) {
		t.Errorf("run.md does not carry the conduct section, so the served-model fact is composed nowhere:\n%s", runmd)
	}
	// The TABLE, not only the heading: a heading with no body would satisfy the line above.
	if !strings.Contains(runmd, "what answered") {
		t.Errorf("run.md's conduct section has no measured-model table:\n%s", runmd)
	}
}
