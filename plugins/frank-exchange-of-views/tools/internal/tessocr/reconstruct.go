package tessocr

// Deterministic table reconstruction from tesseract TSV word geometry — the grid branch
// of the per-page pipeline (plan §II), ported from the Wave 0 prototype
// (~/scratch-reconstruct, measured p0054 380/385 cells against the oracle and 381/385
// against pixels). Pure Go on purpose: it runs under the default (stub) build too, so the
// reconstruction path and its fixtures are testable with no C stack present.
//
// The pipeline: classify tokens into marks and labels by a confusion-set grammar, cluster
// mark centres into columns and rows, attach label lines, expand merged mark tokens
// against the grid, place marks, name columns from the rotated-band header recovery, emit
// |-separated rows plus Stats.
//
// Every constant is an empirical fit to the 300-DPI corpus and says so. The mark grammar
// only works because X-matrices have no other cell content; a numeric table needs a
// different cell-content model, which the plan scopes out (non-ruled and content-bearing
// tables are named future work).

import (
	"bufio"
	"errors"
	"sort"
	"strconv"
	"strings"
)

// Reconstruction clustering gaps, px at 300 DPI (RenderDPI): x-centres closer than
// colGap300 are one column, y-centres closer than rowGap300 are one row. Measured against
// a column pitch of ~108 px and legitimate row pitches down to ~34 px.
const (
	colGap300 = 30.0
	rowGap300 = 25.0
)

// ErrNoMarks means the TSV carried no recognizable mark tokens, so there is no grid to
// reconstruct. Callers fall back to plain text WITH this stated on the record — a page
// whose reconstruction found nothing must never read as a page with an empty table.
var ErrNoMarks = errors.New("tessocr: no mark tokens in TSV; nothing to reconstruct")

// Stats is what a reconstruction leaves on the page record (plan §II: reconstruction
// confidence is a FIELD, never inferred from the emitted table's shape).
type Stats struct {
	// ColumnsFound counts emitted columns (supercolumns where anchors grouped them);
	// SubColumnsFound counts the raw mark columns before grouping — the grid's actual
	// vertical rule count is SubColumnsFound+1.
	ColumnsFound     int `json:"columns_found"`
	SubColumnsFound  int `json:"subcolumns_found"`
	RowsFound        int `json:"rows_found"`
	HeaderNamesFound int `json:"header_names_found"`
	// MarksPlaced/MarksTotal is the plan's original fallback trigger — and Wave 0 proved
	// it is NOT sufficient alone: when the OCR drops glyphs, MarksTotal drops with them
	// and the ratio stays healthy (p0052: 33/36 placed on a page missing 55% of its
	// marks). PSMDisagreement and ExpectedIntersections are the two independent
	// denominators that catch what this ratio cannot.
	MarksTotal    int `json:"marks_total"`
	MarksPlaced   int `json:"marks_placed"`
	MarksUnplaced int `json:"marks_unplaced"`
	LabellessRows int `json:"labelless_rows"`
	EmptyRows     int `json:"empty_rows"`
	// PSMDisagreement is set by the caller from mark-token counts under PSMAuto vs
	// PSMSparseText (see PSMDisagreement); reconstruction of a single TSV cannot know it.
	PSMDisagreement float64 `json:"psm_disagreement"`
}

// ExpectedIntersections is the rule-lattice size this reconstruction implies:
// (rows+1) x (subcolumns+1) crossings. Compared against the grid detector's measured
// intersection count it is the OCR-independent dropout signal — a lattice far larger than
// the reconstruction accounts for means tesseract dropped grid content that leptonica can
// still see.
func (s Stats) ExpectedIntersections() int {
	return (s.RowsFound + 1) * (s.SubColumnsFound + 1)
}

// PSMDisagreement compares candidate-mark counts from the same page's TSV under two
// segmentation modes: 0 when they agree, 1 when one mode saw marks the other missed
// entirely (p0052 measured 0 under PSMAuto vs 26 under PSMSparseText — total silent
// dropout, invisible to every other stat). Two zeros agree: a prose page has no marks
// under either mode, and that is the honest zero, not a miss.
func PSMDisagreement(autoMarks, sparseMarks int) float64 {
	a, b := float64(autoMarks), float64(sparseMarks)
	hi := a
	if b > hi {
		hi = b
	}
	if hi == 0 {
		return 0
	}
	return abs(a-b) / hi
}

// MarkTokenCount counts strong-mark tokens in a TSV — the input to PSMDisagreement.
func MarkTokenCount(tsv string) int {
	n := 0
	for _, w := range parseTSVWords(tsv) {
		if isStrongMark(w.text) {
			n++
		}
	}
	return n
}

// Reconstruct rebuilds a ruled table from level-5 TSV word geometry. headers, when
// non-nil, are the rotated-band recovery's column names in column order (see
// ParseRotatedBandHeaders); without them column names fall back to supercolumn anchors.
// The emitted table is |-separated rows in the corpus's own transcript format.
func Reconstruct(tsv string, headers []string) (string, Stats, error) {
	words := parseTSVWords(tsv)
	return reconstruct(words, headers)
}

// ParseRotatedBandHeaders reads the rotated-band re-OCR output: names separated by blank
// lines, a multi-line group being one wrapped name ("Installation and\nCheckout").
func ParseRotatedBandHeaders(band string) []string {
	var out []string
	// CRLF is normalized FIRST: a Windows checkout hands this fixture (and a Windows
	// tesseract run hands this band) \r\n endings, under which the blank-line split below
	// matches nothing and eleven headers collapse into one — caught by CI's Windows leg.
	band = strings.ReplaceAll(band, "\r\n", "\n")
	for _, grp := range strings.Split(strings.TrimSpace(band), "\n\n") {
		name := strings.Join(strings.Fields(grp), " ")
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// ---- TSV parsing ----

type tsvWord struct {
	x, y, w, h int
	conf       float64
	text       string
}

func (wd tsvWord) cx() float64 { return float64(wd.x) + float64(wd.w)/2 }
func (wd tsvWord) cy() float64 { return float64(wd.y) + float64(wd.h)/2 }

// parseTSVWords keeps level-5 (word) rows with non-empty text and drops everything else;
// page/block/line geometry is re-derived from the words rather than trusted.
func parseTSVWords(tsv string) []tsvWord {
	var out []tsvWord
	sc := bufio.NewScanner(strings.NewReader(tsv))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		fld := strings.Split(sc.Text(), "\t")
		if len(fld) < 12 || fld[0] != "5" {
			continue
		}
		txt := strings.TrimSpace(fld[11])
		if txt == "" {
			continue
		}
		atoi := func(s string) int { n, _ := strconv.Atoi(s); return n }
		conf, _ := strconv.ParseFloat(fld[10], 64)
		out = append(out, tsvWord{atoi(fld[6]), atoi(fld[7]), atoi(fld[8]), atoi(fld[9]), conf, txt})
	}
	return out
}

// ---- mark classification ----

const seps = "|/\\(){}[]!;:.,'\"`~—–-_=+<>*^?%$&"

// stripSeps removes separator glyphs (including I/l/i/j/1, common misreads of rule lines
// and the | glyph) from a token.
func stripSeps(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(seps, r) || r == 'I' || r == 'l' || r == 'i' || r == 'j' || r == '1' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// isStrongMark accepts a token whose stripped remainder is 1-24 chars from the X
// confusion set with at least one x/X. The set covers what tesseract actually calls an X
// at 300 DPI (`X`, `xX`, `D4`) and column-run merges of a whole run of X's into one token
// (`X|XPX/X|X|X{X|X/X|KEX|`).
func isStrongMark(s string) bool {
	st := stripSeps(s)
	if st == "" || len(st) > 24 {
		return false
	}
	hasX := false
	for _, r := range st {
		switch r {
		case 'x', 'X':
			hasX = true
		case '4', 'K', 'M', 'D', 'Y', 'P', 'k', 'm', 'y', 'N', 'H', 'E', 'R':
			// confusion glyphs, allowed alongside X
		default:
			return false
		}
	}
	return hasX
}

// isWeakMark accepts a short confusion-set token with no x/X required (`4`, `K|`). Weak
// marks place only by tight snap — they are too ambiguous to seed geometry.
func isWeakMark(s string) bool {
	st := stripSeps(s)
	if st == "" || len(st) > 2 {
		return false
	}
	for _, r := range st {
		switch r {
		case 'x', 'X', '4', 'K', 'M', 'D', 'Y', 'k', 'm', 'y':
		default:
			return false
		}
	}
	return true
}

// ---- 1-D clustering ----

type cluster struct {
	centre float64
	n      int
}

func cluster1D(vals []float64, gap float64) []cluster {
	if len(vals) == 0 {
		return nil
	}
	v := append([]float64(nil), vals...)
	sort.Float64s(v)
	var out []cluster
	start, sum := 0, v[0]
	for i := 1; i <= len(v); i++ {
		if i == len(v) || v[i]-v[i-1] > gap {
			n := i - start
			out = append(out, cluster{sum / float64(n), n})
			if i < len(v) {
				start, sum = i, v[i]
			}
		} else {
			sum += v[i]
		}
	}
	return out
}

// mergeSubSplits merges adjacent clusters closer than frac*medianPitch — repair for a
// column that clustered into two because its marks wobble around the rule.
func mergeSubSplits(cs []cluster, frac float64) []cluster {
	for {
		if len(cs) < 3 {
			return cs
		}
		var pitches []float64
		for i := 1; i < len(cs); i++ {
			pitches = append(pitches, cs[i].centre-cs[i-1].centre)
		}
		med := median(pitches)
		best, bestGap := -1, med*frac
		for i := 1; i < len(cs); i++ {
			if g := cs[i].centre - cs[i-1].centre; g < bestGap {
				best, bestGap = i, g
			}
		}
		if best < 0 {
			return cs
		}
		cs = mergeAt(cs, best)
	}
}

func mergeAt(cs []cluster, i int) []cluster {
	a, b := cs[i-1], cs[i]
	merged := cluster{(a.centre*float64(a.n) + b.centre*float64(b.n)) / float64(a.n+b.n), a.n + b.n}
	return append(cs[:i-1], append([]cluster{merged}, cs[i+1:]...)...)
}

func mergeClosest(cs []cluster) []cluster {
	best, bestGap := 1, 1e18
	for i := 1; i < len(cs); i++ {
		if g := cs[i].centre - cs[i-1].centre; g < bestGap {
			best, bestGap = i, g
		}
	}
	return mergeAt(cs, best)
}

func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s[len(s)/2]
}

func nearest(centres []float64, v float64) (idx int, dist float64) {
	idx, dist = -1, 1e18
	for i, c := range centres {
		if d := abs(v - c); d < dist {
			idx, dist = i, d
		}
	}
	return
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// ---- reconstruction ----

type tableRow struct {
	cy     float64
	labels []tsvWord
	cells  map[int]int // column index -> mark count
}

func reconstruct(words []tsvWord, headers []string) (string, Stats, error) {
	var st Stats
	colGap, rowGap := colGap300, rowGap300

	// 1. classify
	var strong, weak, other []tsvWord
	for _, w := range words {
		switch {
		case isStrongMark(w.text) && w.h >= 12 && w.w >= 8:
			strong = append(strong, w)
		case isWeakMark(w.text) && w.h >= 12 && w.w >= 6:
			weak = append(weak, w)
		default:
			// keep everything: a pure-separator token (`~`, `*`, `>.<`) can still be a
			// misread X, judged by geometry later
			other = append(other, w)
		}
	}
	if len(strong) == 0 {
		return "", st, ErrNoMarks
	}

	// median single-mark box size from strong marks
	var ws, hs []float64
	for _, w := range strong {
		ws = append(ws, float64(w.w))
		hs = append(hs, float64(w.h))
	}
	medW, medH := median(ws), median(hs)

	// 2. columns from single-ish strong marks
	var colXs, rowYs []float64
	for _, w := range strong {
		if float64(w.w) <= 1.7*medW {
			colXs = append(colXs, w.cx())
		}
		if float64(w.h) <= 1.7*medH {
			rowYs = append(rowYs, w.cy())
		}
	}
	cols := mergeSubSplits(cluster1D(colXs, colGap), 0.55)
	// NO sub-split merging for rows: row pitch is legitimately non-uniform
	// (compound-row blocks run at ~half the global pitch) and a median-based merge
	// collapses real adjacent rows.
	markRows := cluster1D(rowYs, rowGap)
	if len(headers) > 0 {
		// the band recovery counted the real columns; extra clusters are sub-splits
		for len(cols) > len(headers) {
			cols = mergeClosest(cols)
		}
	}
	colC := centres(cols)
	rowC := centres(markRows)
	colPitch := medianPitch(colC, 100)
	rowPitch := medianPitch(rowC, 50)

	// 3. label lines: non-mark words left of the grid
	leftBoundary := colC[0] - 0.6*colPitch
	var labelWords []tsvWord
	for _, w := range other {
		if w.cx() < leftBoundary && stripSeps(w.text) != "" {
			labelWords = append(labelWords, w)
		}
	}
	sort.Slice(labelWords, func(i, j int) bool { return labelWords[i].cy() < labelWords[j].cy() })
	var labelLines [][]tsvWord
	for _, w := range labelWords {
		if n := len(labelLines); n > 0 {
			last := labelLines[n-1]
			if abs(w.cy()-lineCy(last)) <= 15 {
				labelLines[n-1] = append(last, w)
				continue
			}
		}
		labelLines = append(labelLines, []tsvWord{w})
	}
	// hyphen-join: a label line ending in "-" is a wrapped label; pull the continuation
	// line into it so the row centre lands between the halves.
	for i := 0; i < len(labelLines)-1; i++ {
		ln := labelLines[i]
		rightmost := ln[0]
		for _, w := range ln {
			if w.x+w.w > rightmost.x+rightmost.w {
				rightmost = w
			}
		}
		if strings.HasSuffix(rightmost.text, "-") &&
			lineCy(labelLines[i+1])-lineCy(ln) < 0.9*rowPitch {
			labelLines[i] = append(ln, labelLines[i+1]...)
			labelLines = append(labelLines[:i+1], labelLines[i+2:]...)
			i--
		}
	}

	// 4. table span and row assembly
	top := rowC[0] - 1.0*rowPitch
	bot := rowC[len(rowC)-1] + 1.0*rowPitch
	rows := make([]*tableRow, len(rowC))
	for i, c := range rowC {
		rows[i] = &tableRow{cy: c, cells: map[int]int{}}
	}
	// each mark row takes at most ONE label line (the nearest); a losing line — e.g. a
	// section header sitting above an indented sub-row block — becomes its own
	// (typically empty) row. Wrapped labels were already fused by the hyphen-join above.
	type claim struct {
		line []tsvWord
		cy   float64
		idx  int
		d    float64
	}
	var claims []claim
	for _, ln := range labelLines {
		c := lineCy(ln)
		if c < top || c > bot {
			continue
		}
		idx, d := nearest(rowC, c)
		if d <= 0.55*rowPitch {
			claims = append(claims, claim{ln, c, idx, d})
		} else {
			rows = append(rows, &tableRow{cy: c, labels: ln, cells: map[int]int{}})
		}
	}
	sort.Slice(claims, func(i, j int) bool { return claims[i].d < claims[j].d })
	taken := map[int]bool{}
	for _, cl := range claims {
		if taken[cl.idx] {
			rows = append(rows, &tableRow{cy: cl.cy, labels: cl.line, cells: map[int]int{}})
			continue
		}
		taken[cl.idx] = true
		rows[cl.idx].labels = append(rows[cl.idx].labels, cl.line...)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].cy < rows[j].cy })
	allRowC := make([]float64, len(rows))
	for i, r := range rows {
		allRowC[i] = r.cy
	}

	// 5. mark placement, expanding merged boxes against the grid
	place := func(w tsvWord) {
		colsHit := spanned(colC, float64(w.x), float64(w.x+w.w), float64(w.w) > 1.7*medW, w.cx(), 0.5*colPitch)
		rowsHit := spanned(allRowC, float64(w.y), float64(w.y+w.h), float64(w.h) > 1.7*medH, w.cy(), 0.55*rowPitch)
		n := len(colsHit) * len(rowsHit)
		if len(colsHit) > len(rowsHit) {
			st.MarksTotal += len(colsHit)
		} else if len(rowsHit) > 0 {
			st.MarksTotal += len(rowsHit)
		} else {
			st.MarksTotal++
		}
		if n == 0 {
			st.MarksUnplaced++
			return
		}
		for _, ci := range colsHit {
			for _, ri := range rowsHit {
				rows[ri].cells[ci]++
				st.MarksPlaced++
			}
		}
	}
	for _, w := range strong {
		place(w)
	}
	for _, w := range weak {
		// weak marks must snap tightly to an existing column and row
		ci, dc := nearest(colC, w.cx())
		ri, dr := nearest(allRowC, w.cy())
		st.MarksTotal++
		if ci >= 0 && dc <= 0.4*colPitch && ri >= 0 && dr <= 0.4*rowPitch {
			rows[ri].cells[ci]++
			st.MarksPlaced++
		} else {
			st.MarksUnplaced++
		}
	}
	// geometric marks: an X glyph can be misread as arbitrary junk (`ras`, `bad`, `»`,
	// `*`). Text tells us nothing, but a single-mark-sized box sitting exactly on a
	// row x column intersection inside the grid is a mark regardless of what the
	// recognizer called it.
	gtolC := minf(16, 0.3*colPitch)
	gtolR := minf(16, 0.3*rowPitch)
	for _, w := range other {
		if w.cx() <= leftBoundary {
			continue
		}
		if float64(w.w) < 10 || float64(w.w) > 1.9*medW || float64(w.h) < 12 || float64(w.h) > 1.9*medH {
			continue
		}
		ci, dc := nearest(colC, w.cx())
		ri, dr := nearest(allRowC, w.cy())
		if ci >= 0 && dc <= gtolC && ri >= 0 && dr <= gtolR {
			st.MarksTotal++
			rows[ri].cells[ci]++
			st.MarksPlaced++
		}
	}

	// 5b. wrapped-label continuation: an empty label-only row whose label starts with a
	// lowercase letter ("verification", "supporting processes") is the second line of
	// the row above it, not a row. A section header ("Inspection", "Management of V&V")
	// starts uppercase and stays a row. TSV alone cannot fully distinguish the two — an
	// uppercase continuation line defeats the heuristic (p0051's one split label); the
	// ruled-line detector owns the principled answer and Wave 2 should pass rule
	// y-positions in rather than improve this.
	for i := 1; i < len(rows); i++ {
		r := rows[i]
		if len(r.cells) != 0 || len(r.labels) == 0 {
			continue
		}
		first := renderLabel(r.labels)
		fr := rune(first[0])
		if fr >= 'a' && fr <= 'z' && r.cy-rows[i-1].cy < 0.9*rowPitch {
			rows[i-1].labels = append(rows[i-1].labels, r.labels...)
			rows = append(rows[:i], rows[i+1:]...)
			i--
		}
	}

	// 6. margin-junk suppression: an empty label-only row whose label sits entirely left
	// of the marked rows' label column is page-margin noise (e.g. a rotated copyright
	// line), not a table row.
	minLabelX := 1e18
	for _, r := range rows {
		if len(r.cells) > 0 && len(r.labels) > 0 {
			for _, w := range r.labels {
				if float64(w.x) < minLabelX {
					minLabelX = float64(w.x)
				}
			}
		}
	}
	if minLabelX < 1e18 {
		var kept []*tableRow
		for _, r := range rows {
			maxX := 0.0
			for _, w := range r.labels {
				if e := float64(w.x + w.w); e > maxX {
					maxX = e
				}
			}
			if len(r.cells) == 0 && len(r.labels) > 0 && maxX < minLabelX-20 {
				continue
			}
			kept = append(kept, r)
		}
		rows = kept
	}

	// 7. supercolumn anchors: a repeated identical caption word above the grid (e.g.
	// "Levels" over every "Levels 4 3 2 1" band) marks the true column centres of a
	// table whose data columns are level sub-columns.
	anchors, anchorCy := findAnchors(other, top, colPitch)
	var superOf []int
	var outColC []float64
	if len(anchors) >= 3 {
		superOf = make([]int, len(colC))
		for i, c := range colC {
			superOf[i], _ = nearest(anchors, c)
		}
		outColC = anchors
		if len(headers) == 0 {
			headers = bandHeaders(other, anchors, anchorCy, leftBoundary)
		}
	} else {
		superOf = make([]int, len(colC))
		for i := range colC {
			superOf[i] = i
		}
		outColC = colC
	}

	// 8. emit
	var b strings.Builder
	corner := cornerLabel(labelLines, top)
	hdr := make([]string, len(outColC))
	for i := range hdr {
		if i < len(headers) {
			hdr[i] = headers[i]
		}
	}
	b.WriteString("| " + corner + " | " + strings.Join(hdr, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat("---|", len(outColC)+1) + "\n")
	for _, r := range rows {
		label := renderLabel(r.labels)
		counts := make([]int, len(outColC))
		hasMark := false
		for ci, n := range r.cells {
			counts[superOf[ci]] += n
			if n > 0 {
				hasMark = true
			}
		}
		cells := make([]string, len(outColC))
		for i, n := range counts {
			cells[i] = strings.TrimSpace(strings.Repeat("X ", n))
		}
		if label == "" {
			st.LabellessRows++
		}
		if !hasMark {
			st.EmptyRows++
		}
		b.WriteString("| " + label + " | " + strings.Join(cells, " | ") + " |\n")
	}
	st.ColumnsFound = len(outColC)
	st.SubColumnsFound = len(colC)
	st.RowsFound = len(rows)
	st.HeaderNamesFound = len(headers)
	return b.String(), st, nil
}

// findAnchors looks above the grid for a token repeated >=4 times at a near-uniform pitch
// wider than the sub-column pitch; returns its centres.
func findAnchors(other []tsvWord, top, colPitch float64) ([]float64, float64) {
	groups := map[string][]tsvWord{}
	for _, w := range other {
		if w.cy() >= top {
			continue
		}
		k := strings.ToLower(strings.Trim(w.text, seps+" "))
		if len(k) < 3 {
			continue
		}
		groups[k] = append(groups[k], w)
	}
	// keep only tokens on the caption's own line: a same-word occurrence in the corner
	// header (e.g. "Software Integrity Levels") sits on another baseline and must not
	// become an anchor
	for k, ws := range groups {
		var cys []float64
		for _, w := range ws {
			cys = append(cys, w.cy())
		}
		med := median(cys)
		var kept []tsvWord
		for _, w := range ws {
			if abs(w.cy()-med) <= 15 {
				kept = append(kept, w)
			}
		}
		groups[k] = kept
	}
	// among repeated captions at a supercolumn-like pitch, take the one closest to the
	// grid (max mean cy): the band's last caption line
	var best []tsvWord
	bestCy := -1.0
	for _, ws := range groups {
		if len(ws) < 4 {
			continue
		}
		var xs []float64
		var cy float64
		for _, w := range ws {
			xs = append(xs, w.cx())
			cy += w.cy()
		}
		sort.Float64s(xs)
		if medianPitch(xs, 0) < 1.5*colPitch {
			continue
		}
		if m := cy / float64(len(ws)); m > bestCy {
			bestCy, best = m, ws
		}
	}
	if best == nil {
		return nil, 0
	}
	var xs []float64
	for _, w := range best {
		xs = append(xs, w.cx())
	}
	sort.Float64s(xs)
	return xs, bestCy
}

// bandHeaders recovers per-supercolumn header text: grid-region words above the anchor
// row, below the widest line gap (which separates the table title from the header band),
// assigned to the nearest anchor.
func bandHeaders(other []tsvWord, anchors []float64, anchorCy, leftBoundary float64) []string {
	var band []tsvWord
	for _, w := range other {
		if w.cy() < anchorCy-5 && w.cx() > leftBoundary && stripSeps(w.text) != "" {
			band = append(band, w)
		}
	}
	if len(band) == 0 {
		return nil
	}
	sort.Slice(band, func(i, j int) bool { return band[i].cy() < band[j].cy() })
	// find the widest vertical gap between consecutive band lines; the title sits above
	// it
	cut := band[0].cy() - 1
	bestGap := 0.0
	prev := band[0].cy()
	for _, w := range band[1:] {
		if g := w.cy() - prev; g > 40 && g > bestGap {
			bestGap, cut = g, prev+g/2
		}
		if w.cy() > prev {
			prev = w.cy()
		}
	}
	perCol := make([][]tsvWord, len(anchors))
	for _, w := range band {
		if w.cy() < cut {
			continue
		}
		i, _ := nearest(anchors, w.cx())
		perCol[i] = append(perCol[i], w)
	}
	out := make([]string, len(anchors))
	for i, ws := range perCol {
		out[i] = renderLabel(ws)
	}
	return out
}

// spanned resolves a mark box to grid indices: a single box snaps to its nearest centre
// within tol; a merged box (wider/taller than the single-mark envelope) marks every
// centre its extent covers — that is how a whole column-run of X's read as one token
// still lands one mark per row.
func spanned(centres []float64, lo, hi float64, merged bool, cv, tol float64) []int {
	if !merged {
		if i, d := nearest(centres, cv); i >= 0 && d <= tol {
			return []int{i}
		}
		return nil
	}
	var out []int
	for i, c := range centres {
		if c >= lo-5 && c <= hi+5 {
			out = append(out, i)
		}
	}
	if len(out) == 0 {
		if i, d := nearest(centres, cv); i >= 0 && d <= tol {
			return []int{i}
		}
	}
	return out
}

func centres(cs []cluster) []float64 {
	out := make([]float64, len(cs))
	for i, c := range cs {
		out[i] = c.centre
	}
	return out
}

func medianPitch(c []float64, fallback float64) float64 {
	if len(c) < 2 {
		return fallback
	}
	var p []float64
	for i := 1; i < len(c); i++ {
		p = append(p, c[i]-c[i-1])
	}
	return median(p)
}

func lineCy(ln []tsvWord) float64 {
	var s float64
	for _, w := range ln {
		s += w.cy()
	}
	return s / float64(len(ln))
}

func cornerLabel(lines [][]tsvWord, top float64) string {
	bestCy, best := -1.0, ""
	for _, ln := range lines {
		c := lineCy(ln)
		if c < top && c > bestCy {
			bestCy, best = c, renderLabel(ln)
		}
	}
	return best
}

func renderLabel(ws []tsvWord) string {
	s := append([]tsvWord(nil), ws...)
	sort.Slice(s, func(i, j int) bool {
		if abs(s[i].cy()-s[j].cy()) > 12 {
			return s[i].cy() < s[j].cy()
		}
		return s[i].x < s[j].x
	})
	// tesseract sometimes emits a label line twice at a 1-2 px offset; drop a word whose
	// box overlaps >60% with an already-kept word on the same line.
	var kept []tsvWord
	for _, w := range s {
		dup := false
		for _, k := range kept {
			if abs(w.cy()-k.cy()) > 14 {
				continue
			}
			lo, hi := maxf(float64(w.x), float64(k.x)), minf(float64(w.x+w.w), float64(k.x+k.w))
			if hi > lo && (hi-lo) > 0.6*minf(float64(w.w), float64(k.w)) {
				dup = true
				break
			}
		}
		if !dup {
			kept = append(kept, w)
		}
	}
	var parts []string
	for _, w := range kept {
		parts = append(parts, w.text)
	}
	return strings.Join(parts, " ")
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
