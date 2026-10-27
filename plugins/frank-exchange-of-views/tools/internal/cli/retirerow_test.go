package cli

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A RETIRE TAKES THE ANCHOR AND ITS HUSK, AND A LINE OF SYMBOLS IS NOT A HUSK, through the real
// verbs. A finding quoted on a table row stands on that row; blue cuts the cell's claim down to a
// dash, corrects a row of marks, realigns the delimiter row — each edit carrying the anchor on a
// row that holds no letter and no digit — and retires what it cut, naming the anchor. The anchor
// leaves; the row is what the edit wrote and stays, in the report's text and in the report that
// ships.
func TestARetireLeavesARowOfSymbolsWhereItStands(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Checks\n\n| Check | Result |\n| --- | --- |\n| Seven divides it | ✓ |\n| ✓ | ✗ |\n\nAfter.\n")
	registerBlue(t, runDir)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)

	for i, step := range []struct{ quote, cut, row string }{
		{"Seven divides it", "—", "| — | ✓ |"},
		{"| ✓ | ✗ |", "| ✓ | ✓ |", "| ✓ | ✓ |"},
		{"| --- | --- |", "|:---|---:|", "|:---|---:|"},
	} {
		key := string(rune('1' + i))
		if out, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F"+key, "--quote", step.quote,
			"--reason", "the row is questioned", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
			t.Fatalf("finding on %q: %v\n%s", step.quote, err, out)
		}
		m := regexp.MustCompile(regexp.QuoteMeta(step.quote) + `(<!--fx:(F-[0-9a-f]{8})-->)`).FindStringSubmatch(readReport(t, runDir))
		if m == nil {
			t.Fatalf("no finding anchor after %q:\n%s", step.quote, readReport(t, runDir))
		}
		if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E"+key,
			"--quote", step.quote+m[1], "--new", step.cut+m[1], "--reason", "the row as it should read"); err != nil {
			t.Fatalf("edit %q down to symbols: %v\n%s", step.quote, err, out)
		}
		out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", step.quote, "--reason", "refuted", "--anchor", m[2])
		if err != nil || !strings.Contains(out, m[2]) {
			t.Fatalf("retire %q naming %s: %v\n%s", step.quote, m[2], err, out)
		}
		rep := readReport(t, runDir)
		if strings.Contains(rep, m[1]) {
			t.Errorf("the retired anchor %s still stands:\n%s", m[1], rep)
		}
		if !strings.Contains(rep, "\n"+step.row+"\n") {
			t.Errorf("the retire of %q took the row %q with its anchor:\n%s", step.quote, step.row, rep)
		}
	}
	const table = "| Check | Result |\n|:---|---:|\n| — | ✓ |\n| ✓ | ✓ |\n\nAfter.\n"
	if rep := readReport(t, runDir); !strings.HasSuffix(rep, table) {
		t.Errorf("the report's table reads\n%s\nwant it to end\n%s", rep, table)
	}
	docs, err := shippedDocs(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if d.File == report.FileReport && !strings.Contains(d.Body, strings.TrimSuffix(table, "\n\nAfter.\n")) {
			t.Errorf("report.md does not ship the table whole:\n%s", d.Body)
		}
	}
}
