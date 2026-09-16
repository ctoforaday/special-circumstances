package tessocr

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// THE RULES A SCAN ACTUALLY CARRIES, and the cells they bound.
//
// DetectGrid counts rule pixels; that is enough to say "this page is a table" and nothing at
// all about where its cells are. A ruled table whose cells hold TEXT reconstructs from the
// lattice: the rules say which rectangle each word sits in, and the row binding survives
// (#932). Mark-grid reconstruction infers its lattice from the marks instead, because on
// those pages the marks are the only ink inside the cells.

// A Rule is one straight run of ink the opening left standing, as a bounding box in page
// pixels. x,y is its top-left corner.
type Rule struct{ X, Y, W, H int }

// Lattice is a page's rule geometry, horizontal and vertical.
type Lattice struct {
	Horizontal []Rule
	Vertical   []Rule
}

// ParseLattice reads the shim's dump: one line per rule, "h|v x y w h".
func ParseLattice(dump string) Lattice {
	var lat Lattice
	for _, line := range strings.Split(dump, "\n") {
		f := strings.Fields(line)
		if len(f) != 5 {
			continue
		}
		n := make([]int, 4)
		bad := false
		for i, s := range f[1:] {
			v, err := strconv.Atoi(s)
			if err != nil {
				bad = true
				break
			}
			n[i] = v
		}
		if bad {
			continue
		}
		r := Rule{X: n[0], Y: n[1], W: n[2], H: n[3]}
		switch f[0] {
		case "h":
			lat.Horizontal = append(lat.Horizontal, r)
		case "v":
			lat.Vertical = append(lat.Vertical, r)
		}
	}
	return lat
}

// Empty reports whether the page carried no rules at all.
func (l Lattice) Empty() bool { return len(l.Horizontal) == 0 && len(l.Vertical) == 0 }

// Cell is one rectangle of the lattice, with the row and column it sits at. Row 0 is the
// topmost band; Col 0 the leftmost. A cell that spans several columns keeps the column it
// starts at and states its Span.
type Cell struct {
	Row, Col, Span int
	X0, Y0, X1, Y1 int
	words          []tsvWord
}

// Text is the cell's words in reading order, joined with a space: a line break inside a cell is
// layout, not content (gblock, 2026-09-16), so a row is one line of the reading.
func (c Cell) Text() string {
	ws := append([]tsvWord(nil), c.words...)
	sort.Slice(ws, func(i, j int) bool {
		if abs(ws[i].cy()-ws[j].cy()) > float64(ws[i].h) {
			return ws[i].cy() < ws[j].cy()
		}
		return ws[i].x < ws[j].x
	})
	parts := make([]string, 0, len(ws))
	for _, w := range ws {
		parts = append(parts, w.text)
	}
	return strings.Join(parts, " ")
}

// Table is the lattice's cells, in reading order.
type Table struct {
	Cells []Cell
	// Cols is the number of column boundaries the widest band carried.
	Cols int
	// X0, Y0, X1, Y1 is the lattice's own bounding box: outside it, a word belongs to the
	// page rather than to the table.
	X0, Y0, X1, Y1 int
}

// Lattice geometry constants, in pixels at RenderDPI. A rule is a run of ink, not a
// mathematical line: two segments of one table rule arrive as separate components with a
// pixel or two of drift, and a cell border shared by two bands is one rule, not two.
const (
	// ruleJoin is how far apart two rules may sit and still be the same table line.
	ruleJoin = 12
	// minBandHeight excludes the hairline between a doubled rule; no cell of this corpus
	// is shorter than a line of text.
	minBandHeight = 24
	// coverage is how much of a band a vertical rule must span to bound its cells. A rule
	// that stops short is a nested box, not a column boundary.
	coverage = 0.8
	// minRowGap is the blank height that separates one ROW from the next inside a band that has
	// no rule between them. Measured on IEEE 1012 p34 at 300 DPI: the gaps between that band's
	// three tasks are 45, 53 and 88 px, and the gaps between lines WITHIN a task's prose are
	// 10-15 px. A band is cut only where no word crosses at all, so this number separates a row
	// boundary from ordinary leading, and nothing else depends on it.
	minRowGap = 32
)

// Tables splits a page's rules into the TABLES they bound: a run of bands with no blank stretch
// between them. A page carrying two tables and prose between them is two lattices, and treating
// it as one merges their columns and pulls the prose into cells — measured on the generated
// twotables fixture, where one lattice read 7 columns across both tables and put the sentence
// between them in a cell.
//
// The cut is a gap between two horizontal rules that NO VERTICAL CROSSES. Height alone does not
// say it: this corpus has single rows 676 px tall (a task with its inputs) and two tables 416 px
// apart, so any threshold on the gap either splits a tall row or merges two tables. A column runs
// through a table's rows and stops at its bottom rule, which is exactly the fact needed.
func (l Lattice) Tables() []Table {
	rows := groupRules(l.Horizontal, true)
	if len(rows) < 2 {
		return nil
	}
	var out []Table
	start := 0
	for i := 1; i <= len(rows); i++ {
		if i < len(rows) && crossedByAVertical(l.Vertical, rows[i-1], rows[i]) {
			continue
		}
		if i-start >= 2 {
			sub := Lattice{Horizontal: rulesBetween(l.Horizontal, rows[start], rows[i-1]), Vertical: l.Vertical}
			if t := sub.Cells(); len(t.Cells) > 0 {
				// ITS OWN BOUNDS, NOT THE PAGE'S. The verticals belong to the whole page, so the
				// lattice's extent spans every table on it; a table's own extent is its cells,
				// and a word between two tables belongs to neither.
				t.X0, t.Y0, t.X1, t.Y1 = t.Cells[0].X0, t.Cells[0].Y0, t.Cells[0].X1, t.Cells[0].Y1
				for _, c := range t.Cells {
					t.X0, t.Y0 = mini(t.X0, c.X0), mini(t.Y0, c.Y0)
					t.X1, t.Y1 = maxi(t.X1, c.X1), maxi(t.Y1, c.Y1)
				}
				out = append(out, t)
			}
		}
		start = i
	}
	return out
}

// columnCoverage is how much of [top, bot] each COLUMN covers: the segments at one x are one
// column, because a scan breaks a printed rule where it meets the row lines. Shared by the two
// readers of that fact, so a column that bounds a cell and a column that joins two bands into one
// table are the same column.
func columnCoverage(vs []Rule, top, bot int) map[int]int {
	out := map[int]int{}
	for _, x := range groupRules(vs, false) {
		covered := 0
		for _, v := range vs {
			if abs(float64(v.X+v.W/2-x)) > ruleJoin {
				continue
			}
			if lo, hi := maxi(v.Y, top), mini(v.Y+v.H, bot); hi > lo {
				covered += hi - lo
			}
		}
		out[x] = covered
	}
	return out
}

// crossedByAVertical reports whether any column runs through the gap between two rules — the test
// for "one table with a tall row" against "two tables with something between them".
func crossedByAVertical(vs []Rule, top, bot int) bool {
	if bot <= top {
		return true
	}
	for _, covered := range columnCoverage(vs, top, bot) {
		if float64(covered) >= coverage*float64(bot-top) {
			return true
		}
	}
	return false
}

// rulesBetween is the horizontal rules whose line lies within [top, bot], inclusive of the join
// tolerance — the rules of one table.
func rulesBetween(rs []Rule, top, bot int) []Rule {
	var out []Rule
	for _, r := range rs {
		if y := r.Y + r.H/2; y >= top-ruleJoin && y <= bot+ruleJoin {
			out = append(out, r)
		}
	}
	return out
}

// Cells recovers the lattice's cell rectangles.
//
// ROW BANDS COME FROM THE HORIZONTAL RULES, columns from the verticals that actually cross
// each band — which is what makes a spanning cell a spanning cell: a section heading row
// carries no interior vertical, so it is one cell wide rather than three empty ones. The
// boundary case this exists to refuse is a band with no verticals at all on a page whose
// other bands have them: that is a real merged row, and it reads as one cell.
func (l Lattice) Cells() Table {
	rows := groupRules(l.Horizontal, true)
	if len(rows) < 2 {
		return Table{}
	}
	var t Table
	t.X0, t.Y0, t.X1, t.Y1 = boundsOf(l.Horizontal, l.Vertical)
	for i := 0; i+1 < len(rows); i++ {
		top, bot := rows[i], rows[i+1]
		if bot-top < minBandHeight {
			continue
		}
		xs := columnsIn(l.Vertical, top, bot, t.X0, t.X1)
		for j := 0; j+1 < len(xs); j++ {
			t.Cells = append(t.Cells, Cell{
				Row: len(t.Cells), Col: j, Span: 1,
				X0: xs[j], Y0: top, X1: xs[j+1], Y1: bot,
			})
		}
		if len(xs)-1 > t.Cols {
			t.Cols = len(xs) - 1
		}
	}
	// Row indexes are per band, not per cell.
	row, prevTop := -1, -1
	for i := range t.Cells {
		if t.Cells[i].Y0 != prevTop {
			row++
			prevTop = t.Cells[i].Y0
		}
		t.Cells[i].Row = row
	}
	return t
}

// groupRules collapses rules that describe one table line into one coordinate: the median of
// the group along the axis it bounds.
func groupRules(rs []Rule, horizontal bool) []int {
	if len(rs) == 0 {
		return nil
	}
	pos := make([]int, len(rs))
	for i, r := range rs {
		if horizontal {
			pos[i] = r.Y + r.H/2
		} else {
			pos[i] = r.X + r.W/2
		}
	}
	sort.Ints(pos)
	var out []int
	start := 0
	for i := 1; i <= len(pos); i++ {
		if i == len(pos) || pos[i]-pos[i-1] > ruleJoin {
			out = append(out, pos[(start+i-1)/2])
			start = i
		}
	}
	return out
}

// columnsIn is the x boundaries a band actually has: the table's own edges plus every column
// whose rule crosses most of the band.
//
// A COLUMN IS ITS SEGMENTS, NOT ONE RULE. A printed vertical arrives from a scan broken at the
// row lines it meets, so asking whether any SINGLE component spans the band answers no on a page
// whose columns are perfectly real — measured on p45 of IEEE 1012, where six verticals bound a
// three-column table and not one of them crossed a band whole. The segments at one x are one
// column, and their covered length is what crosses.
func columnsIn(vs []Rule, top, bot, x0, x1 int) []int {
	var xs []int
	for x, covered := range columnCoverage(vs, top, bot) {
		if float64(covered) >= coverage*float64(bot-top) {
			xs = append(xs, x)
		}
	}
	sort.Ints(xs)
	if len(xs) == 0 {
		return []int{x0, x1}
	}
	if xs[0]-x0 > ruleJoin {
		xs = append([]int{x0}, xs...)
	}
	if x1-xs[len(xs)-1] > ruleJoin {
		xs = append(xs, x1)
	}
	return xs
}

func boundsOf(h, v []Rule) (x0, y0, x1, y1 int) {
	x0, y0 = 1<<30, 1<<30
	for _, r := range append(append([]Rule{}, h...), v...) {
		x0, y0 = mini(x0, r.X), mini(y0, r.Y)
		x1, y1 = maxi(x1, r.X+r.W), maxi(y1, r.Y+r.H)
	}
	if x0 == 1<<30 {
		return 0, 0, 0, 0
	}
	return x0, y0, x1, y1
}

func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CellStats is what a text-cell reconstruction measured, for the record. A reader deciding
// whether to trust a row needs the shape it came from and how much of the page it explains.
type CellStats struct {
	// Tables is how many tables the page's rules bound: a page carrying two tables and prose
	// between them is two lattices, and reading it as one merges their columns.
	Tables int `json:"tables"`
	// Bands and Columns are the lattice's shape: row bands the rules bound, and the widest band's
	// column count. Rows is what the reading emits: a band whose print separates its rows by
	// space rather than by a rule splits into several (IEEE 1012's Table 1 rules its section
	// headings and not its tasks), so Rows is at least Bands.
	Bands   int `json:"bands"`
	Rows    int `json:"rows"`
	Columns int `json:"columns"`
	// Cells is how many rectangles the rules bound.
	Cells int `json:"cells"`
	// WordsPlaced landed in a cell; WordsLoose sat inside the table's own bounding box and in no
	// cell (a word straddling a rule); WordsAround sat outside the table and are kept as the
	// page's prose, before or after the rows.
	WordsPlaced int `json:"words_placed"`
	WordsLoose  int `json:"words_loose,omitempty"`
	WordsAround int `json:"words_around,omitempty"`
	// WordsFromSparse were read by the sparse pass into cells the layout pass left EMPTY. Page
	// segmentation drops a short word standing alone in a narrow column the same way it drops an
	// isolated mark, and an empty cell beside a full one is where that shows.
	WordsFromSparse int `json:"words_from_sparse,omitempty"`
}

// WordsPerCell is the measure that separates a table of TEXT cells from a table of marks: a mark
// cell holds one glyph or none (measured 0.43-0.96 over this corpus's mark grids), a text cell
// holds a phrase (4.3 and up). Zero cells is zero, never a division by zero.
func (s CellStats) WordsPerCell() float64 {
	if s.Cells == 0 {
		return 0
	}
	return float64(s.WordsPlaced) / float64(s.Cells)
}

// Text-cell acceptance, measured over the 35 grid pages of IEEE 1012 (plans/ocr-text-cell-rows.md
// §III D4). A page below any of these reads as plain text with the reason stated.
const (
	MinTextCellColumns = 2
	// MinTextCellRows counts the rows the READING emits, after a band is cut at the row
	// boundaries its print leaves as space: a one-band table of two rows is a table.
	MinTextCellRows = 2
	MinWordsPerCell = 2.0
)

// fill puts each word in the cell that contains its centre, and reports what happened to the
// rest. A word's CENTRE decides: a word touching a rule belongs to one cell, never to two.
func (t *Table) fill(words []tsvWord) (above, below []tsvWord, st CellStats) {
	st = CellStats{Cells: len(t.Cells), Columns: t.Cols}
	for i := range t.Cells {
		if t.Cells[i].Col == 0 {
			st.Bands++
		}
	}

	for _, w := range words {
		if strings.TrimSpace(w.text) == "" {
			continue
		}
		cx, cy := int(w.cx()), int(w.cy())
		if cx < t.X0 || cx >= t.X1 || cy < t.Y0 || cy >= t.Y1 {
			st.WordsAround++
			if cy < t.Y0 {
				above = append(above, w)
			} else {
				below = append(below, w)
			}
			continue
		}
		hit := -1
		for i := range t.Cells {
			c := &t.Cells[i]
			if cx >= c.X0 && cx < c.X1 && cy >= c.Y0 && cy < c.Y1 {
				hit = i
				break
			}
		}
		if hit < 0 {
			st.WordsLoose++
			continue
		}
		st.WordsPlaced++
		t.Cells[hit].words = append(t.Cells[hit].words, w)
	}
	return above, below, st
}

// fillEmpty puts the sparse pass's words into cells the layout pass left empty, and reports how
// many it recovered. A cell that already holds a word is never touched.
func (t *Table) fillEmpty(words []tsvWord) int {
	n := 0
	for _, w := range words {
		if strings.TrimSpace(w.text) == "" {
			continue
		}
		cx, cy := int(w.cx()), int(w.cy())
		for i := range t.Cells {
			c := &t.Cells[i]
			if len(c.words) > 0 || cx < c.X0 || cx >= c.X1 || cy < c.Y0 || cy >= c.Y1 {
				continue
			}
			c.words = append(c.words, w)
			n++
			break
		}
	}
	return n
}

// splitBands cuts each band into the rows its print separates by SPACE rather than by a rule.
//
// THE ROWS OF A TABLE ARE NOT ALWAYS RULED. IEEE 1012's Table 1 rules its section headings and
// leaves the tasks inside a section separated by white space — so a band holds three tasks, their
// inputs and their outputs, and binding the band alone still leaves "the inputs of task 2"
// unanswered. A cut is a y where NO word of the band crosses, at least minRowGap tall; the
// columns' own gaps line up there, because the print aligns a task with its inputs.
//
// Where a band has no such gap it stays one row, which is the ruled reading.
func (t *Table) splitBands() {
	var out []Cell
	for i := 0; i < len(t.Cells); {
		j := i
		for j < len(t.Cells) && t.Cells[j].Y0 == t.Cells[i].Y0 {
			j++
		}
		band := t.Cells[i:j]
		for _, cut := range rowCuts(band) {
			for k := range band {
				var keep, rest []tsvWord
				for _, w := range band[k].words {
					if int(w.cy()) < cut {
						keep = append(keep, w)
					} else {
						rest = append(rest, w)
					}
				}
				split := band[k]
				split.Y0, split.Y1 = cut, band[k].Y1
				split.words = rest
				band[k].Y1 = cut
				band[k].words = keep
				out = append(out, band[k])
				band[k] = split
			}
		}
		out = append(out, band...)
		i = j
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].Y0 != out[b].Y0 {
			return out[a].Y0 < out[b].Y0
		}
		return out[a].X0 < out[b].X0
	})
	row, prevTop := -1, -1
	for i := range out {
		if out[i].Y0 != prevTop {
			row++
			prevTop = out[i].Y0
		}
		out[i].Row = row
	}
	t.Cells = out
}

// rowCuts is every y inside a band that no word crosses, with minRowGap of blank on both sides —
// the band's own row boundaries, in order.
func rowCuts(band []Cell) []int {
	type span struct{ lo, hi int }
	var ws []span
	for _, c := range band {
		for _, w := range c.words {
			ws = append(ws, span{w.y, w.y + w.h})
		}
	}
	if len(ws) == 0 {
		return nil
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].lo < ws[j].lo })
	var cuts []int
	end := ws[0].hi
	for _, s := range ws[1:] {
		if s.lo-end >= minRowGap {
			cuts = append(cuts, (end+s.lo)/2)
		}
		end = maxi(end, s.hi)
	}
	return cuts
}

// render writes the filled table as the same markdown rows the mark reconstruction emits: a
// header row, a separator, then one line per band. A band with one cell across the table — a
// section heading — keeps the empty columns after it, as the oracle's own rendering does.
func (t Table) render() string {
	rows := map[int][]Cell{}
	var order []int
	for _, c := range t.Cells {
		if _, seen := rows[c.Row]; !seen {
			order = append(order, c.Row)
		}
		rows[c.Row] = append(rows[c.Row], c)
	}
	var b strings.Builder
	for n, r := range order {
		cs := rows[r]
		sort.Slice(cs, func(i, j int) bool { return cs[i].X0 < cs[j].X0 })
		var parts []string
		for _, c := range cs {
			parts = append(parts, c.Text())
		}
		b.WriteString("| " + strings.Join(parts, " | ") + " |")
		// A short row spans the columns it does not have, which is what the pipes say.
		for i := len(parts); i < t.Cols; i++ {
			b.WriteString("|")
		}
		b.WriteString("\n")
		if n == 0 {
			b.WriteString("|" + strings.Repeat("---|", t.Cols) + "\n")
		}
	}
	return b.String()
}

// lines renders loose words as the page's own prose, in reading order: words whose baselines sit
// within one line height of each other are one line.
func linesOf(words []tsvWord) string {
	if len(words) == 0 {
		return ""
	}
	ws := append([]tsvWord(nil), words...)
	sort.Slice(ws, func(i, j int) bool {
		if abs(ws[i].cy()-ws[j].cy()) > float64(ws[i].h) {
			return ws[i].cy() < ws[j].cy()
		}
		return ws[i].x < ws[j].x
	})
	var b strings.Builder
	prev := ws[0]
	for i, w := range ws {
		switch {
		case i == 0:
		case abs(w.cy()-prev.cy()) > float64(w.h):
			b.WriteString("\n")
		default:
			b.WriteString(" ")
		}
		b.WriteString(w.text)
		prev = w
	}
	return b.String()
}

// TextCells reconstructs a ruled table whose cells hold TEXT: the lattice says where the cells
// are, the TSV says which words sit in each, and the page's own prose is kept around the rows in
// reading order (a page is not always all table — one of this corpus's is prose with a five-row
// table in it). It returns the reading, what it measured, and the reason it is not acceptable,
// which is "" when it is.
//
// THE SPARSE PASS IS THE SECOND WITNESS, AND ONLY FOR EMPTY CELLS. Layout analysis drops a short
// word alone in a narrow column — measured on the generated fixture, where "High" and "Major" are
// absent from the auto TSV and present in the sparse one — exactly as it drops isolated marks.
// Filling only cells the first pass left empty takes those words back without letting the two
// passes each contribute their own reading of the same cell.
func TextCells(lat Lattice, tsv, sparse string) (string, CellStats, string) {
	tables := lat.Tables()
	if len(tables) == 0 {
		return "", CellStats{}, "the page's rules bound no cells, so there is no row to recover"
	}
	words, sparseWords := parseTSVWords(tsv), parseTSVWords(sparse)
	// THE BETTER WITNESS FILLS FIRST. Layout analysis reads a dense page best and drops words on a
	// sparse one — measured on the generated two-table page, where it read 22 words and the sparse
	// pass read the rest. Whichever pass read more words fills the cells; the other fills the cells
	// it left empty, so neither pass's dropout decides what the page says.
	if len(sparseWords) > len(words) {
		words, sparseWords = sparseWords, words
	}
	var st CellStats
	// Each table takes the words inside it; what no table claims is the page's own prose, and it
	// is written where it sits — before, between or after the tables, in reading order.
	claimed := map[int]bool{}
	type block struct {
		y    int
		text string
	}
	var blocks []block
	for i := range tables {
		t := &tables[i]
		for j, w := range words {
			if int(w.cx()) >= t.X0 && int(w.cx()) < t.X1 && int(w.cy()) >= t.Y0 && int(w.cy()) < t.Y1 {
				claimed[j] = true
			}
		}
		_, _, one := t.fill(words)
		one.WordsFromSparse = t.fillEmpty(sparseWords)
		one.WordsPlaced += one.WordsFromSparse
		t.splitBands()
		one.Rows = t.Cells[len(t.Cells)-1].Row + 1
		st = st.add(one)
		blocks = append(blocks, block{t.Y0, strings.TrimRight(t.render(), "\n")})
	}
	if why := st.refuse(); why != "" {
		return "", st, why
	}
	var loose []tsvWord
	for j, w := range words {
		if !claimed[j] && strings.TrimSpace(w.text) != "" {
			loose = append(loose, w)
		}
	}
	st.WordsAround = len(loose)
	for _, run := range proseRuns(loose, tables) {
		blocks = append(blocks, block{run.y, linesOf(run.words)})
	}
	sort.SliceStable(blocks, func(a, b int) bool { return blocks[a].y < blocks[b].y })
	var parts []string
	for _, b := range blocks {
		if b.text != "" {
			parts = append(parts, b.text)
		}
	}
	return strings.Join(parts, "\n\n"), st, ""
}

// proseRuns groups the page's own words into the runs BETWEEN tables, so a sentence that sits
// under one table and above the next is written there rather than collected at the end.
func proseRuns(words []tsvWord, tables []Table) []struct {
	y     int
	words []tsvWord
} {
	type run = struct {
		y     int
		words []tsvWord
	}
	byBand := map[int][]tsvWord{}
	for _, w := range words {
		band := 0
		for _, t := range tables {
			if int(w.cy()) >= t.Y1 {
				band++
			}
		}
		byBand[band] = append(byBand[band], w)
	}
	var out []run
	for band, ws := range byBand {
		y := 0
		if band > 0 {
			y = tables[band-1].Y1
		}
		out = append(out, run{y, ws})
	}
	sort.Slice(out, func(a, b int) bool { return out[a].y < out[b].y })
	return out
}

// add folds one table's measurements into the page's: a page carrying two tables reports both,
// and its shape is the widest and the tallest of them.
func (s CellStats) add(o CellStats) CellStats {
	s.Tables++
	s.Cells += o.Cells
	s.Bands += o.Bands
	s.Rows += o.Rows
	s.WordsPlaced += o.WordsPlaced
	s.WordsLoose += o.WordsLoose
	s.WordsFromSparse += o.WordsFromSparse
	s.Columns = maxi(s.Columns, o.Columns)
	return s
}

// refuse states why this lattice may not stand as the page's text, or "" when it may. Every
// condition names what it measured: a reader must be able to tell a table of marks from a
// figure from a page that simply has no grid.
func (s CellStats) refuse() string {
	switch {
	case s.Columns < MinTextCellColumns:
		return fmt.Sprintf("the page's rules bound %d column(s), and a table needs at least %d — this is a box or a rule, not a grid", s.Columns, MinTextCellColumns)
	case s.Rows < MinTextCellRows:
		return fmt.Sprintf("the page's rules bound %d row(s), and a table needs at least %d", s.Rows, MinTextCellRows)
	case s.WordsPerCell() < MinWordsPerCell:
		return fmt.Sprintf("the rules bound %d cells holding %d words (%.2f per cell, under %.2f): these cells carry marks rather than text, and reading them as text would drop what they mean", s.Cells, s.WordsPlaced, s.WordsPerCell(), MinWordsPerCell)
	}
	return ""
}
