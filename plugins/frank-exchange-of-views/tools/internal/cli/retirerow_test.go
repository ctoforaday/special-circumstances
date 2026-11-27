package cli

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A RETIRE TAKES THE ANCHOR, AND A LINE OF SYMBOLS IS NOT A HUSK, through the real verbs. A gap
// quoted on a table row stands on that row; its lens closes it, blue retires its marker, naming the
// anchor (an open gap's marker is refused, F14), and the anchor leaves while the row stays word for
// word — a row of marks or the delimiter row included, though neither holds a letter or a digit.
// Blue then cuts the cell's claim down to a dash, corrects the row of marks and realigns the
// delimiter row (no anchor rides on any of them: none may stand on no words), and the table is what
// the edits wrote, in the report's text and in the report that ships.
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
		if _, err := mintQuote(t, runDir, "G"+key, step.quote); err != nil {
			t.Fatalf("mint on %q: %v", step.quote, err)
		}
		m := regexp.MustCompile(regexp.QuoteMeta(step.quote) + `(<!--gap:(G-[0-9a-f]{8})-->)`).FindStringSubmatch(readReport(t, runDir))
		if m == nil {
			t.Fatalf("no gap anchor after %q:\n%s", step.quote, readReport(t, runDir))
		}
		line := readReport(t, runDir)
		line = line[strings.LastIndex(line[:strings.Index(line, m[1])], "\n")+1:]
		line = strings.Replace(line[:strings.Index(line, "\n")], m[1], "", 1)
		if err := lensClose(t, runDir, lensSeat, m[2], "not_a_defect"); err != nil {
			t.Fatalf("closing %s: %v", m[2], err)
		}
		out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--reason", "refuted", "--anchor", m[2])
		if err != nil || !strings.Contains(out, m[2]) {
			t.Fatalf("retire naming %s on %q: %v\n%s", m[2], step.quote, err, out)
		}
		rep := readReport(t, runDir)
		if strings.Contains(rep, m[1]) {
			t.Errorf("the retired anchor %s still stands:\n%s", m[1], rep)
		}
		if !strings.Contains(rep, "\n"+line+"\n") {
			t.Errorf("the retire of %s took the row %q with its anchor:\n%s", m[2], line, rep)
		}
		if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E"+key,
			"--quote", step.quote, "--new", step.cut, "--reason", "the row as it should read"); err != nil {
			t.Fatalf("edit %q down to symbols: %v\n%s", step.quote, err, out)
		}
		if rep := readReport(t, runDir); !strings.Contains(rep, "\n"+step.row+"\n") {
			t.Errorf("the edit of %q did not write the row %q:\n%s", step.quote, step.row, rep)
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
