package tessocr

import (
	"fmt"
	"strings"
	"testing"
)

// THE LATTICE IS PURE GEOMETRY, so every rule of it is pinned without the C stack: the shim's
// dump goes in, cells and a reading come out. The measured pages live in the tagged tests.

// ruleDump writes what the shim writes: "h|v x y w h" per rule.
func ruleDump(rs ...string) string { return strings.Join(rs, "\n") + "\n" }

// A three-column table, four horizontal rules (three bands), whose middle band is a section
// heading: its verticals stop at the band above.
func headingTable() string {
	return ruleDump(
		"h 100 100 900 3", // top
		"h 100 200 900 3", // under the header row
		"h 100 300 900 3", // under the heading band
		"h 100 400 900 3", // bottom
		"v 100 100 3 300", // the table's left edge, full height
		"v 400 100 3 100", // interior, header band only
		"v 700 100 3 100",
		"v 400 300 3 100", // interior, last band only
		"v 700 300 3 100",
		"v 1000 100 3 300", // right edge
	)
}

func TestLatticeReadsRulesAndBoundsCells(t *testing.T) {
	lat := ParseLattice(headingTable())
	if len(lat.Horizontal) != 4 || len(lat.Vertical) != 6 {
		t.Fatalf("parsed %d horizontal and %d vertical rules, want 4 and 6", len(lat.Horizontal), len(lat.Vertical))
	}
	tab := lat.Cells()
	if tab.Cols != 3 {
		t.Errorf("columns = %d, want 3", tab.Cols)
	}
	// THE HEADING BAND IS ONE CELL: no vertical crosses it, so it spans the table rather than
	// arriving as three cells of which two are empty.
	byRow := map[int]int{}
	for _, c := range tab.Cells {
		byRow[c.Row]++
	}
	if got := []int{byRow[0], byRow[1], byRow[2]}; got[0] != 3 || got[1] != 1 || got[2] != 3 {
		t.Errorf("cells per band = %v, want [3 1 3]", got)
	}
	if tab.X0 != 100 || tab.Y0 != 100 || tab.X1 != 1003 || tab.Y1 != 403 {
		t.Errorf("table bounds = (%d,%d)-(%d,%d), want the rules' own extent", tab.X0, tab.Y0, tab.X1, tab.Y1)
	}
}

func TestLatticeJoinsOneRuleDrawnTwice(t *testing.T) {
	// A printed rule arrives as two components a few pixels apart, and a scan's two halves of
	// one line are one line. Joined, this is a 2x2 table; unjoined it would be 3 bands of which
	// one is a hairline.
	lat := ParseLattice(ruleDump(
		"h 100 100 400 3", "h 505 102 400 3", // one rule, two segments
		"h 100 300 805 3", "h 100 303 805 2", // a doubled rule
		"v 100 100 3 205", "v 900 100 3 205", "v 500 100 3 205",
	))
	tab := lat.Cells()
	rows := map[int]bool{}
	for _, c := range tab.Cells {
		rows[c.Row] = true
	}
	if len(rows) != 1 || tab.Cols != 2 || len(tab.Cells) != 2 {
		t.Fatalf("bands=%d columns=%d cells=%d, want one band of two columns", len(rows), tab.Cols, len(tab.Cells))
	}
	// The two segments are ONE boundary at their shared y, and the doubled rule is ONE line: the
	// band runs from the first rule to the second, not to the hairline between the second's halves.
	c := tab.Cells[0]
	if c.Y0 != 101 || c.Y1 != 301 {
		t.Errorf("the band is y %d-%d, want 101-301 — the segments joined and the doubled rule read once", c.Y0, c.Y1)
	}
}

// A HAIRLINE IS NOT A ROW. A rule printed twice, or a rule the scan thickened, leaves two
// coordinates a few pixels apart; the sliver between them is not a band a cell could sit in.
func TestLatticeSkipsABandTooShortToHoldText(t *testing.T) {
	lat := ParseLattice(ruleDump(
		"h 100 100 900 3", "h 100 118 900 3", // 18 px apart: a hairline, under minBandHeight
		"h 100 400 900 3",
		"v 100 100 3 303", "v 500 100 3 303", "v 1000 100 3 303",
	))
	tab := lat.Cells()
	rows := map[int]bool{}
	for _, c := range tab.Cells {
		rows[c.Row] = true
	}
	if len(rows) != 1 {
		t.Errorf("bands = %d, want 1 — the 18 px sliver between two rules is not a row", len(rows))
	}
}

func TestLatticeIgnoresAVerticalThatDoesNotCrossTheBand(t *testing.T) {
	// A nested box inside a cell is not a column boundary: it spans a fraction of the band.
	lat := ParseLattice(ruleDump(
		"h 100 100 900 3", "h 100 400 900 3",
		"v 100 100 3 303", "v 1000 100 3 303",
		"v 550 340 3 40", // a short stroke inside the band
	))
	if tab := lat.Cells(); tab.Cols != 1 {
		t.Errorf("columns = %d, want 1 — a stroke that crosses %d%% of the band is not a column", tab.Cols, 13)
	}
}

// fakeTSV writes the level-5 lines parseTSVWords reads: one word per line, at a given box.
func fakeTSV(words []struct {
	text       string
	x, y, w, h int
	conf       int
}) string {
	var b strings.Builder
	b.WriteString("level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n")
	for _, w := range words {
		fmt.Fprintf(&b, "5\t1\t1\t1\t1\t1\t%d\t%d\t%d\t%d\t%d\t%s\n", w.x, w.y, w.w, w.h, w.conf, w.text)
	}
	return b.String()
}

type word = struct {
	text       string
	x, y, w, h int
	conf       int
}

func TestTextCellsKeepsRowBindingAndThePageAroundIt(t *testing.T) {
	tsv := fakeTSV([]word{
		{"Running", 100, 40, 90, 20, 96}, {"head", 200, 40, 60, 20, 96}, // above the table
		{"Criticality", 110, 150, 120, 20, 95}, {"Description", 410, 150, 130, 20, 95}, {"Notes", 710, 150, 70, 20, 95},
		{"High", 110, 250, 60, 20, 95}, {"Affects", 410, 250, 80, 20, 95}, {"critical", 500, 250, 80, 20, 95},
		{"performance", 410, 275, 140, 20, 95}, {"of", 560, 275, 30, 20, 95}, {"the", 600, 275, 40, 20, 95},
		{"system", 650, 275, 70, 20, 95}, {"no", 710, 250, 40, 20, 95}, {"workaround", 710, 275, 120, 20, 95},
		{"26", 120, 500, 30, 20, 96}, // the folio, below the table
	})
	lat := ParseLattice(ruleDump(
		"h 100 100 900 3", "h 100 200 900 3", "h 100 300 900 3",
		"v 100 100 3 203", "v 400 100 3 203", "v 700 100 3 203", "v 1000 100 3 203",
	))
	text, st, why := TextCells(lat, parseTSVWords(tsv), nil)
	if why != "" {
		t.Fatalf("refused: %s", why)
	}
	want := "Running head\n\n" +
		"| Criticality | Description | Notes |\n|---|---|---|\n" +
		"| High | Affects critical performance of the system | no workaround |\n\n26"
	if text != want {
		t.Errorf("reading =\n%q\nwant\n%q", text, want)
	}
	if st.Bands != 2 || st.Columns != 3 || st.Cells != 6 || st.WordsPlaced != 12 || st.WordsAround != 3 {
		t.Errorf("stats = %+v, want 2 bands, 3 columns, 6 cells, 12 words placed, 3 around", st)
	}
}

func TestTextCellsRefusesWhatIsNotATableOfText(t *testing.T) {
	marks := func(n int) string {
		var ws []word
		for i := 0; i < n; i++ {
			ws = append(ws, word{"X", 110 + 100*(i%8), 150 + 100*(i/8), 20, 20, 90})
		}
		return fakeTSV(ws)
	}
	// A mark grid: nine columns, and one glyph per cell.
	var rules []string
	for y := 100; y <= 400; y += 100 {
		rules = append(rules, fmt.Sprintf("h 100 %d 900 3", y))
	}
	for x := 100; x <= 1000; x += 100 {
		rules = append(rules, fmt.Sprintf("v %d 100 3 303", x))
	}
	cases := []struct {
		name, dump, tsv, want string
	}{
		{"a box, one column", ruleDump("h 100 100 900 3", "h 100 300 900 3", "v 100 100 3 203", "v 1000 100 3 203"),
			fakeTSV([]word{{"A", 200, 150, 30, 20, 95}, {"paragraph", 250, 150, 90, 20, 95}}), "column(s)"},
		{"one band", ruleDump("h 100 100 900 3", "h 100 200 900 3", "v 100 100 3 103", "v 500 100 3 103", "v 1000 100 3 103"),
			fakeTSV([]word{{"only", 200, 150, 40, 20, 95}, {"one", 600, 150, 40, 20, 95}}), "row(s)"},
		{"a grid of marks", ruleDump(rules...), marks(24), "carry marks rather than text"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			text, _, why := TextCells(ParseLattice(tc.dump), parseTSVWords(tc.tsv), nil)
			if why == "" {
				t.Fatalf("accepted, want a refusal naming %q. text:\n%s", tc.want, text)
			}
			if !strings.Contains(why, tc.want) {
				t.Errorf("refusal %q does not say %q", why, tc.want)
			}
			if text != "" {
				t.Errorf("a refused lattice still returned text: %q", text)
			}
		})
	}
}

func TestParseLatticeIgnoresWhatIsNotARule(t *testing.T) {
	lat := ParseLattice("h 1 2 3 4\nnonsense\nv 5 6 7\nx 1 2 3 4\nv 5 6 7 8\n\n")
	if len(lat.Horizontal) != 1 || len(lat.Vertical) != 1 {
		t.Errorf("parsed %d horizontal and %d vertical rules from a dump with three bad lines, want 1 and 1",
			len(lat.Horizontal), len(lat.Vertical))
	}
	if lat.Empty() {
		t.Error("a lattice with rules reports itself empty")
	}
	if !ParseLattice("").Empty() {
		t.Error("a page with no rules does not report an empty lattice")
	}
}

// A COLUMN BROKEN BY THE SCAN IS STILL A COLUMN. On p45 of IEEE 1012 six verticals bound a
// three-column table and not one component crossed a band whole: the printed rule arrives in
// pieces, cut where it meets each row line. The segments at one x are one column.
func TestLatticeReadsAColumnBrokenIntoSegments(t *testing.T) {
	lat := ParseLattice(ruleDump(
		"h 100 100 900 3", "h 100 300 900 3",
		"v 100 100 3 203", "v 1000 100 3 203",
		// One interior column, arriving as three pieces with gaps where the scan lost it.
		"v 550 105 3 60", "v 550 172 3 60", "v 551 238 3 58",
	))
	if tab := lat.Cells(); tab.Cols != 2 {
		t.Errorf("columns = %d, want 2 — a column broken into segments is one column", tab.Cols)
	}
}

// A BAND HOLDS THE ROWS ITS PRINT SEPARATES BY SPACE. IEEE 1012's Table 1 rules its section
// headings and leaves the tasks inside a section separated by white space, so binding the band
// alone still leaves "the inputs of task 2" unanswered (#932). A band is cut where no word
// crosses; ordinary leading inside a paragraph is not a cut.
func TestTextCellsSplitsABandWhereNoWordCrosses(t *testing.T) {
	// Two "tasks" in one band, 45 px apart, each with its own inputs cell; the first task's own
	// two lines sit 12 px apart, which is leading rather than a row boundary.
	tsv := fakeTSV([]word{
		{"Task", 110, 150, 60, 20, 95}, {"one", 180, 150, 40, 20, 95},
		{"continues", 110, 182, 100, 20, 95}, // 12 px below: same row
		{"Input", 410, 150, 60, 20, 95}, {"A", 480, 150, 20, 20, 95},
		{"Task", 110, 247, 60, 20, 95}, {"two", 180, 247, 40, 20, 95}, // 45 px below: a new row
		{"Input", 410, 247, 60, 20, 95}, {"B", 480, 247, 20, 20, 95},
	})
	// The table is two columns wide: its rules stop at the right-hand edge of the second.
	lat := ParseLattice(ruleDump(
		"h 100 100 603 3", "h 100 320 603 3",
		"v 100 100 3 223", "v 400 100 3 223", "v 700 100 3 223",
	))
	text, st, why := TextCells(lat, parseTSVWords(tsv), nil)
	if why != "" {
		t.Fatalf("refused: %s", why)
	}
	if st.Bands != 1 || st.Rows != 2 {
		t.Errorf("bands=%d rows=%d, want one ruled band holding two rows", st.Bands, st.Rows)
	}
	want := "| Task one continues | Input A |\n|---|---|\n| Task two | Input B |"
	if text != want {
		t.Errorf("reading =\n%q\nwant\n%q", text, want)
	}
}
