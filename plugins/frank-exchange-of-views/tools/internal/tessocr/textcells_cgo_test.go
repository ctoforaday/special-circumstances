//go:build tessocr && cgo

package tessocr

import (
	"os"
	"strings"
	"testing"
)

// A RULED TABLE OF TEXT KEEPS ITS ROWS (#932), through the real engine on real pixels. The
// fixture is generated (testdata/gen), so what this pins is the pipeline's behaviour rather than
// one scan's luck: the detector fires, the marks path finds nothing to rebuild, and the rules'
// own cells carry each row's text together.
func TestTextCellTableKeepsItsRowsThroughTheEngine(t *testing.T) {
	png, err := os.ReadFile("testdata/textcells.png")
	if err != nil {
		t.Fatal(err)
	}
	en, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer en.Close()
	res, err := en.ReadPage(png, Grid300)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Table {
		t.Fatalf("the grid detector did not fire on a ruled table: %+v", res.Grid)
	}
	if res.TextCells == nil {
		t.Fatalf("no text-cell reconstruction: fallback %q / %q\n%s", res.Fallback, res.TextCellFallback, res.Text)
	}
	if st := *res.TextCells; st.Columns != 3 || st.Bands != 4 {
		t.Errorf("shape = %d columns x %d bands, want 3 x 4 (header, heading, two rows)", st.Columns, st.Bands)
	}
	// THE ROW BINDING IS THE POINT: the first row's own description and output sit on its line.
	// The plain reading of this page emits every description, then every output.
	line := ""
	for _, l := range strings.Split(res.Text, "\n") {
		if strings.HasPrefix(l, "| High ") {
			line = l
		}
	}
	if line == "" {
		t.Fatalf("no row for the first criticality:\n%s", res.Text)
	}
	for _, want := range []string{"critical performance", "Task Report"} {
		if !strings.Contains(line, want) {
			t.Errorf("the row for High does not carry %q:\n%s", want, line)
		}
	}
	// A cell that wrapped reads as one run of text — the break inside it is layout (gblock,
	// 2026-09-16) — so the row is one line of three cells.
	if strings.Count(line, "|") != 4 {
		t.Errorf("the row is not one line of three cells: %q", line)
	}
	// The spanning heading is one cell, and the page's own prose is kept around the table.
	if !strings.Contains(res.Text, "| 5.2.1 Acquisition Support V&V Activity |") {
		t.Errorf("the spanning heading did not read as one cell:\n%s", res.Text)
	}
	if !strings.Contains(res.Text, "IEEE STANDARD FOR SOFTWARE") || !strings.Contains(res.Text, "26") {
		t.Errorf("the page's prose above and below the table was dropped:\n%s", res.Text)
	}
}
