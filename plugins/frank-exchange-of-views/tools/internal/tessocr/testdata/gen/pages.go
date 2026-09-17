package main

import (
	"fmt"
	"strings"
)

// THE LAYOUT CASES, AS PAGES. Each is drawn here rather than cropped from a document: the bytes
// are reproducible from this file, nothing licensed is checked in, and the case a page exists for
// is stated beside it. A page is a PDF content stream at 612x792 points, rendered at the engine's
// own resolution.
//
// A page's expected reading is a golden beside its image (testdata/<name>.golden), written by
// `go test -tags tessocr -run TestTextCellGoldens -update`. A golden is the whole reading, byte
// for byte, so a regression is a diff rather than a judgement.

type page struct {
	name string
	why  string
	draw func(*canvas)
}

var pages = []page{
	{"textcells", "a ruled table of text: header, a heading that spans, two rows of wrapped prose, page text above and below", drawTextCells},
	{"spacedrows", "rows the print separates by SPACE inside one ruled band — IEEE 1012's Table 1 shape, the case that leaves 'the inputs of task 2' unanswered", drawSpacedRows},
	{"brokenrules", "a three-column table whose verticals the scan breaks at every row line", drawBrokenRules},
	{"markgrid", "a grid of tick marks: the mark reconstruction owns it, and the text path must refuse it", drawMarkGrid},
	{"boxedparagraph", "a paragraph in a box — one column, so not a table", drawBoxedParagraph},
	{"prosepage", "prose with a small table in the middle of it: the page is not all table", drawProsePage},
	{"twotables", "two tables on one page, separated by prose", drawTwoTables},
	{"faintscan", "the ruled table again at low contrast, as a tired photocopy prints it", drawFaint},
	{"levelgrid", "a mark table whose activities each print a Levels caption over 4 3 2 1 subcolumns — IEEE 1012's Table 2 shape, where the level is the content (#933)", drawLevelGrid},
}

// boldText writes in Times-Bold, the face a printed standard sets its table headers and marks in.
func (c *canvas) boldText(x, y, size float64, s string) {
	fmt.Fprintf(&c.b, "%.2f g BT /F2 %.0f Tf %.1f %.1f Td (%s) Tj ET\n", c.gray, size, x, y, escape(s))
}

// canvas draws a PDF content stream in points, with the origin at the bottom left.
type canvas struct {
	b    strings.Builder
	gray float64 // ink level: 0 is black, 0.5 is a faint photocopy
}

func (c *canvas) text(x, y, size float64, s string) {
	fmt.Fprintf(&c.b, "%.2f g BT /F1 %.0f Tf %.1f %.1f Td (%s) Tj ET\n", c.gray, size, x, y, escape(s))
}

func (c *canvas) rule(x, y, w, h float64) {
	fmt.Fprintf(&c.b, "%.2f g %.1f %.1f %.1f %.1f re f\n", c.gray, x, y, w, h)
}

// wrapped writes a cell's text inside a width, at 16-point leading.
func (c *canvas) wrapped(x, y, width, size float64, s string) {
	for i, line := range wrap(s, width) {
		c.text(x, y-float64(i)*16, size, line)
	}
}

const (
	pageW, pageH = 612.0, 792.0
	left, right  = 60.0, 552.0
	fontSize     = 11.0
)

// grid draws a table: horizontal rules at every band boundary, verticals down the columns, and
// (where interior is false for a band) no interior verticals, which makes that band one cell.
type grid struct {
	cols   []float64 // x boundaries, left to right
	tops   []float64 // y of each band's top rule, descending, plus the bottom rule last
	merged map[int]bool
}

func (c *canvas) grid(g grid) {
	for _, y := range g.tops {
		c.rule(g.cols[0], y, g.cols[len(g.cols)-1]-g.cols[0], 1.2)
	}
	for i, x := range g.cols {
		edge := i == 0 || i == len(g.cols)-1
		for b := 0; b+1 < len(g.tops); b++ {
			if !edge && g.merged[b] {
				continue
			}
			top, bot := g.tops[b], g.tops[b+1]
			c.rule(x, bot, 1.2, top-bot)
		}
	}
}

func drawTextCells(c *canvas) {
	g := grid{cols: []float64{left, 170, 460, right}, tops: []float64{700, 640, 580, 460, 340}, merged: map[int]bool{1: true}}
	c.grid(g)
	c.text(left, 730, fontSize, "IEEE STANDARD FOR SOFTWARE")
	c.text(left, 300, fontSize, "26")
	for i, h := range []string{"Criticality", "Description", "Outputs"} {
		c.text(g.cols[i]+8, 660, fontSize, h)
	}
	c.text(left+8, 600, fontSize, "5.2.1 Acquisition Support V&V Activity")
	rows := [][]string{
		{"High", "Selected function affects critical performance of the system", "Task Report"},
		{"Major", "Selected function affects important system performance", "Anomaly Report"},
	}
	for r, row := range rows {
		y := 560 - float64(r)*120
		for i, cell := range row {
			c.wrapped(g.cols[i]+8, y, g.cols[i+1]-g.cols[i]-16, fontSize, cell)
		}
	}
}

func drawSpacedRows(c *canvas) {
	// ONE ruled band holding three rows, separated by space the way IEEE 1012's Table 1 does.
	g := grid{cols: []float64{left, 300, right}, tops: []float64{700, 660, 360}}
	c.grid(g)
	for i, h := range []string{"V&V task", "Required inputs"} {
		c.text(g.cols[i]+8, 675, fontSize, h)
	}
	tasks := [][]string{
		{"(1) Scoping the V&V Effort. Define the project criticality.", "Preliminary System Description"},
		{"(2) Planning the Interface. Plan the V&V schedule for each task.", "SVVP RFP or tender"},
		{"(3) System Requirements Review. Review the system requirements.", "User Needs Contract"},
	}
	for r, row := range tasks {
		y := 630 - float64(r)*95 // 95 pt apart: a row boundary, not leading
		c.wrapped(g.cols[0]+8, y, 232, fontSize, row[0])
		c.wrapped(g.cols[1]+8, y, 232, fontSize, row[1])
	}
}

func drawBrokenRules(c *canvas) {
	cols := []float64{left, 230, 400, right}
	tops := []float64{700, 620, 540, 460, 380, 300, 220, 140}
	for _, y := range tops {
		c.rule(cols[0], y, cols[len(cols)-1]-cols[0], 1.2)
	}
	// Every vertical arrives in pieces, cut where it meets a row line — what a scan does to a
	// printed rule. The pieces stay longer than the detector's opening length (151 px at 300 DPI,
	// about 36 points): a shorter segment is not a faint rule, it is an invisible one, because the
	// opening that finds rules removes every run below its own length.
	for _, x := range cols {
		for b := 0; b+1 < len(tops); b++ {
			top, bot := tops[b], tops[b+1]
			mid := (top + bot) / 2
			c.rule(x, bot, 1.2, mid-bot-3)   // touches the rule below
			c.rule(x, mid+3, 1.2, top-mid-3) // touches the rule above
		}
	}
	for i, h := range []string{"Form", "Technical", "Management"} {
		c.text(cols[i]+8, 660, fontSize, h)
	}
	rows := [][]string{
		{"Classical review of the form", "rigorous independent check", "full management oversight"},
		{"Modified review of the form", "limited independent check", "shared management oversight"},
		{"Internal review of the form", "partial internal check", "local management oversight"},
		{"Embedded review of the form", "minimal internal check", "team level oversight"},
		{"Deferred review of the form", "no check performed yet", "oversight still pending"},
		{"Waived review of the form", "the check was waived", "no oversight required"},
	}
	for r, row := range rows {
		for i, cell := range row {
			c.wrapped(cols[i]+8, 600-float64(r)*80, cols[i+1]-cols[i]-16, fontSize, cell)
		}
	}
}

func drawMarkGrid(c *canvas) {
	cols := []float64{left, 200, 260, 320, 380, 440, 500, right}
	tops := []float64{700, 660, 620, 580, 540, 500}
	c.grid(grid{cols: cols, tops: tops})
	c.text(cols[0]+8, 675, fontSize, "Task")
	for i := 1; i+1 < len(cols); i++ {
		c.text(cols[i]+20, 675, fontSize, fmt.Sprint(i))
	}
	for r, name := range []string{"Algorithm Analysis", "Audit Performance", "Control Flow", "Database Analysis"} {
		y := 635 - float64(r)*40
		c.text(cols[0]+8, y, fontSize, name)
		for i := 1; i+1 < len(cols); i++ {
			if (r+i)%2 == 0 {
				c.text(cols[i]+22, y, fontSize, "X")
			}
		}
	}
}

func drawBoxedParagraph(c *canvas) {
	c.rule(left, 700, right-left, 1.2)
	c.rule(left, 560, right-left, 1.2)
	c.rule(left, 560, 1.2, 141)
	c.rule(right, 560, 1.2, 141)
	c.wrapped(left+10, 680, right-left-20, fontSize,
		"NOTE: This standard does not require a particular life cycle model. The user of this standard maps the V&V activities to the life cycle model the project has chosen.")
}

func drawProsePage(c *canvas) {
	c.wrapped(left, 730, right-left, fontSize,
		"Software integrity levels denote a range of software criticality values necessary to maintain risks within acceptable limits. The assignment of a level is a project decision, and the table below states the scheme this standard uses.")
	g := grid{cols: []float64{left, 200, right}, tops: []float64{620, 580, 540, 500}}
	c.grid(g)
	for i, h := range []string{"Criticality", "Description"} {
		c.text(g.cols[i]+8, 595, fontSize, h)
	}
	for r, row := range [][]string{{"High", "Selected function affects critical performance"}, {"Low", "Selected function has noticeable effect"}} {
		y := 555 - float64(r)*40
		c.text(g.cols[0]+8, y, fontSize, row[0])
		c.wrapped(g.cols[1]+8, y, g.cols[2]-g.cols[1]-16, fontSize, row[1])
	}
	c.wrapped(left, 460, right-left, fontSize,
		"The V&V effort assigns a level to every component. Where the assignment changes during development, the plan is updated and the affected tasks are performed again.")
}

func drawTwoTables(c *canvas) {
	// Rows a real page's height: a 26-point band of running text, not an acre of white with two
	// words in it. Tesseract drops text on a sparse page in BOTH passes — measured while building
	// this fixture — and a fixture that reproduces that measures the engine, not the lattice.
	firstRows := [][]string{
		{"Acquisition process", "Acquisition Support V&V Activity"},
		{"Supply process", "Planning V&V Activity and review"},
		{"Development process", "Requirements V&V Activity"},
		{"Operation process", "Operation V&V Activity"},
	}
	first := grid{cols: []float64{left, 250, right}, tops: []float64{720, 694, 668, 642, 616, 590}}
	c.grid(first)
	for i, h := range []string{"Life cycle process", "V&V activity"} {
		c.text(first.cols[i]+8, 701, fontSize, h)
	}
	for r, row := range firstRows {
		y := 675 - float64(r)*26
		c.text(first.cols[0]+8, y, fontSize, row[0])
		c.text(first.cols[1]+8, y, fontSize, row[1])
	}
	c.wrapped(left, 560, right-left, fontSize,
		"The second table states the outputs each activity produces, and the two are read together by the reviewer.")
	secondRows := [][]string{
		{"Acquisition Support V&V", "Updated SVVP and task report"},
		{"Planning V&V", "Task Report and anomaly report"},
		{"Requirements V&V", "Updated SVVP and review notes"},
		{"Operation V&V", "Anomaly Report and evaluation"},
	}
	second := grid{cols: []float64{left, 250, right}, tops: []float64{500, 474, 448, 422, 396, 370}}
	c.grid(second)
	for i, h := range []string{"V&V activity", "Required outputs"} {
		c.text(second.cols[i]+8, 481, fontSize, h)
	}
	for r, row := range secondRows {
		y := 455 - float64(r)*26
		c.text(second.cols[0]+8, y, fontSize, row[0])
		c.text(second.cols[1]+8, y, fontSize, row[1])
	}
}

func drawFaint(c *canvas) {
	c.gray = 0.45 // a tired photocopy: ink well above the detector's binarization threshold
	drawTextCells(c)
	c.gray = 0
}

func drawLevelGrid(c *canvas) {
	// Landscape content on a portrait page would trip the rotation probe, so the table is drawn
	// upright: a label column, then four activities of four level subcolumns.
	const label, sub = 150.0, 24.0
	activities := []string{"Acquisition", "Supply", "Development", "Operation"}
	x0 := 40.0
	actW := 4 * sub
	right := x0 + label + float64(len(activities))*actW
	tops := []float64{740, 700, 676, 652} // activity names, Levels caption, level digits, grid top
	rows := []string{"Algorithm analysis", "Audit performance", "Cost analysis", "Hazard analysis", "Risk analysis", "Test witnessing"}
	for i := range rows {
		tops = append(tops, 628-float64(i)*24)
	}
	for _, y := range tops {
		c.rule(x0, y, right-x0, 1.2)
	}
	bottom := tops[len(tops)-1]
	c.rule(x0, bottom, 1.2, tops[0]-bottom)
	c.rule(x0+label, bottom, 1.2, tops[0]-bottom)
	for a := range activities {
		ax := x0 + label + float64(a)*actW
		c.rule(ax+actW, bottom, 1.2, tops[0]-bottom)
		for k := 1; k < 4; k++ {
			c.rule(ax+float64(k)*sub, bottom, 1.2, tops[2]-bottom) // level rules start at the digit band
		}
		c.boldText(ax+6, 715, 10, activities[a])
		c.boldText(ax+30, 685, 9, "Levels")
		for k, d := range []string{"4", "3", "2", "1"} {
			c.boldText(ax+float64(k)*sub+8, 659, 11, d)
		}
	}
	c.boldText(x0+6, 685, 9, "Integrity level")
	// marks: which levels each task requires, not contiguous from 4 on every row
	marks := [][]int{{0, 1}, {0, 1, 2, 3}, {0}, {1, 2}, {0, 1, 2}, {3}}
	for r, name := range rows {
		y := 634 - float64(r)*24
		c.text(x0+6, y, 10, name)
		for a := range activities {
			ax := x0 + label + float64(a)*actW
			for _, k := range marks[(r+a)%len(marks)] {
				c.boldText(ax+float64(k)*sub+7, y, 12, "X")
			}
		}
	}
}
