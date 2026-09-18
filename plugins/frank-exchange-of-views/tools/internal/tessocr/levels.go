package tessocr

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"sort"
	"strings"
	"unicode"
)

// THE LEVEL HEADER OF A MARK TABLE, READ ONE CELL AT A TIME (#933).
//
// IEEE 1012's Table 2 prints, under each activity, a "Levels" caption and a row of level
// subcolumns (4 3 2 1). The level is the table's content — which integrity levels require a task —
// and the page-level passes cannot read it: the digits are bold and boxed tightly by rules, and on
// p51 the page TSV held one of forty and a re-read of the band held none. The lattice already knows
// every one of those boxes, so each is cropped inside its rules and read alone.

// levelCaptionMin is how many times a caption must repeat on one baseline to mark supercolumns —
// the same count findAnchors requires, so the level band is looked for exactly where the
// reconstruction will group supercolumns.
const levelCaptionMin = 4

// LevelBandCells is the row of header cells that holds a mark table's levels: the first lattice
// band that starts below a caption repeated across the table ("Levels" over every activity). It
// returns nothing on a page with no repeated caption, which is every table without levels.
//
// Only cells UNDER the captions are level cells: the band also crosses the row-label column,
// whose header ("Software Integrity Levels") is not a level. A cell whose centre lies outside the
// captions' span — half a caption pitch beyond the first and the last — is dropped. A cell the
// lattice merged across a missed rule is then split by the band's own cell width, so every digit
// is read alone: read across the missing rule, a three-level cell of IEEE 1012 p51 returned
// "3|2|4" for 3 2 1.
func LevelBandCells(lat Lattice, tsv string) []Cell {
	captions := repeatedCaption(parseTSVWords(tsv))
	if len(captions) == 0 {
		return nil
	}
	captionY := captions[0].cy()
	lo, hi := captionSpan(captions)
	for _, t := range lat.Tables() {
		var band []Cell
		bandTop := -1
		for _, c := range t.Cells {
			if float64(c.Y0) <= captionY {
				continue
			}
			if bandTop < 0 {
				bandTop = c.Y0
			}
			if centre := float64(c.X0+c.X1) / 2; c.Y0 == bandTop && centre >= lo && centre <= hi {
				band = append(band, c)
			}
		}
		if len(band) >= levelCaptionMin {
			sort.Slice(band, func(i, j int) bool { return band[i].X0 < band[j].X0 })
			return splitWideCells(band)
		}
	}
	return nil
}

// repeatedCaption is the words of a caption of three letters or more that repeats at least
// levelCaptionMin times on one baseline — the supercolumn caption — in x order, or nothing.
func repeatedCaption(words []tsvWord) []tsvWord {
	byText := map[string][]tsvWord{}
	for _, w := range words {
		k := strings.ToLower(strings.Trim(w.text, seps+" "))
		if len(k) >= 3 {
			byText[k] = append(byText[k], w)
		}
	}
	var best []tsvWord
	for _, ws := range byText {
		var cys []float64
		for _, w := range ws {
			cys = append(cys, w.cy())
		}
		med := median(cys)
		var line []tsvWord
		for _, w := range ws {
			if abs(w.cy()-med) <= 15 {
				line = append(line, w)
			}
		}
		if len(line) >= levelCaptionMin && len(line) > len(best) {
			best = line
		}
	}
	sort.Slice(best, func(i, j int) bool { return best[i].cx() < best[j].cx() })
	return best
}

// captionSpan is the x range the captions govern: from half a caption pitch before the first to
// half a pitch after the last.
func captionSpan(captions []tsvWord) (float64, float64) {
	var xs []float64
	for _, w := range captions {
		xs = append(xs, w.cx())
	}
	half := medianPitch(xs, 0) / 2
	return xs[0] - half, xs[len(xs)-1] + half
}

// splitWideCells divides a cell the lattice merged across a missed rule into the cells it spans,
// judged by the band's median cell width.
func splitWideCells(band []Cell) []Cell {
	var ws []float64
	for _, c := range band {
		ws = append(ws, float64(c.X1-c.X0))
	}
	med := median(ws)
	if med <= 0 {
		return band
	}
	var out []Cell
	for _, c := range band {
		n := int(float64(c.X1-c.X0)/med + 0.5)
		if n < 2 {
			out = append(out, c)
			continue
		}
		w := float64(c.X1-c.X0) / float64(n)
		for k := 0; k < n; k++ {
			part := c
			part.X0 = c.X0 + int(float64(k)*w)
			part.X1 = c.X0 + int(float64(k+1)*w)
			if k == n-1 {
				part.X1 = c.X1
			}
			out = append(out, part)
		}
	}
	return out
}

// cellPad is how far inside its rules a header cell is cropped: a rule left in the crop is the ink
// that defeats the read.
const cellPad = 6

// CropCell cuts a cell out of a page, inside its rules, on a white margin — the input a single
// character read wants.
func CropCell(pagePNG []byte, c Cell) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pagePNG))
	if err != nil {
		return nil, fmt.Errorf("tessocr: decoding the page to crop a header cell: %w", err)
	}
	r := image.Rect(c.X0+cellPad, c.Y0+cellPad, c.X1-cellPad, c.Y1-cellPad).Intersect(img.Bounds())
	if r.Empty() {
		return nil, fmt.Errorf("tessocr: header cell (%d,%d)-(%d,%d) is empty inside its rules", c.X0, c.Y0, c.X1, c.Y1)
	}
	const margin = 20
	out := image.NewGray(image.Rect(0, 0, r.Dx()+2*margin, r.Dy()+2*margin))
	draw.Draw(out, out.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(out, image.Rect(margin, margin, margin+r.Dx(), margin+r.Dy()), img, r.Min, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// LevelBoxes turns each header cell's read into level subcolumns. A cell that read several digits
// spans that many subcolumns — the lattice missed a rule there, and the digits say how many — so
// its width is divided equally, one digit per part, left to right. A cell that read no digit is one
// column whose level is unread.
func LevelBoxes(cells []Cell, reads []string) []LevelBox {
	var out []LevelBox
	for i, c := range cells {
		var digits []rune
		if i < len(reads) {
			for _, r := range reads[i] {
				if unicode.IsDigit(r) {
					digits = append(digits, r)
				}
			}
		}
		if len(digits) == 0 {
			out = append(out, LevelBox{X0: c.X0, X1: c.X1})
			continue
		}
		w := float64(c.X1-c.X0) / float64(len(digits))
		for k, d := range digits {
			x0 := c.X0 + int(float64(k)*w)
			x1 := c.X0 + int(float64(k+1)*w)
			if k == len(digits)-1 {
				x1 = c.X1
			}
			out = append(out, LevelBox{X0: x0, X1: x1, Label: string(d)})
		}
	}
	return out
}

// AgreeWithTable keeps a level read only where the table itself agrees with it.
//
// EVERY ACTIVITY PRINTS THE SAME LEVEL HEADER, so the reads at one position under different
// captions are independent witnesses of one printed digit. A read that disagrees with the
// majority at its position is a misread, and it is made UNREAD — never replaced by the majority,
// which would be a level nobody read in that cell. Measured on IEEE 1012: p51 reads 4 3 2 1 under
// all ten activities; p52 reads one "1" as "4"; p53, a sparse page, reads a "5". A position with no
// majority keeps no read at all.
//
// THE MAJORITY NEEDS A QUORUM: at least half the activities must have read that position. On p53
// the lattice gave one activity five boxes, and its fifth — the next activity's "4" — was the only
// read at its position, so it was its own majority.
func AgreeWithTable(boxes []LevelBox, captionCentres []float64) []LevelBox {
	if len(boxes) == 0 || len(captionCentres) == 0 {
		return boxes
	}
	out := append([]LevelBox(nil), boxes...)
	// position of each box within its caption's group, left to right
	pos := make([]int, len(out))
	groups := map[int][]int{}
	for i, b := range out {
		g, _ := nearest(captionCentres, float64(b.X0+b.X1)/2)
		groups[g] = append(groups[g], i)
	}
	for _, idx := range groups {
		sort.Slice(idx, func(a, b int) bool { return out[idx[a]].X0 < out[idx[b]].X0 })
		for k, i := range idx {
			pos[i] = k
		}
	}
	votes := map[int]map[string]int{}
	for i, b := range out {
		if b.Label == "" {
			continue
		}
		if votes[pos[i]] == nil {
			votes[pos[i]] = map[string]int{}
		}
		votes[pos[i]][b.Label]++
	}
	for i := range out {
		if out[i].Label == "" {
			continue
		}
		tally := votes[pos[i]]
		total, best := 0, 0
		for _, n := range tally {
			total += n
			if n > best {
				best = n
			}
		}
		if tally[out[i].Label] != best || 2*best <= total || 2*best < len(captionCentres) {
			out[i].Label = ""
		}
	}
	return out
}
