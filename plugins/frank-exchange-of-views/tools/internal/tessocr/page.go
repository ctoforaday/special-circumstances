package tessocr

// The per-page pipeline of plans/local-ocr.md §II, amended by §VI: grid-detect, then
// either plain text or the full grid branch — orientation recovery, two-PSM TSV, the
// rotated-header band, deterministic reconstruction, and an honest fallback. This file is
// deliberately untagged: it orchestrates Engine methods that are themselves tag-split, so
// under the default (stub) build every path here surfaces ErrNotCompiledIn rather than a
// zero that reads like a blank page.
//
// Every pixel constant in this file is a 300-DPI empirical fit to the IEEE 1012 corpus
// and says so where it is declared. Reusing one at another DPI is the mistake the Wave 0
// spot-check caught; the read path refuses other resolutions rather than misapplying
// these numbers quietly.

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"sort"
	"unicode"
)

// PageResult is what the engine produced for one page: the reading, and the facts a
// record needs to say how it was derived. The zero value is meaningless — a result
// arrives only alongside a nil error.
type PageResult struct {
	// Text is the page's reading: plain OCR text for a prose page, the reconstructed
	// |-separated table for a grid page, or plain text again when the reconstruction
	// fell back — in which case Fallback says so.
	Text string
	// Table reports that the grid detector fired. It is the detector's fact, kept even
	// when reconstruction later fell back: "a grid was seen and could not be rebuilt" and
	// "no grid" are different findings.
	Table bool
	// RotatedPage reports that the page read better rotated 90° clockwise — a landscape
	// table on a portrait scan (plan §VI amendment a) — and that Text, the TSV passes and
	// the reconstruction all came from the rotated pixels.
	RotatedPage bool
	// Grid is the detector's measurement. Intersections is the OCR-independent
	// denominator reconstruction stats are judged against (§VI amendment b).
	Grid GridStats
	// Reconstruction is the grid branch's stats — nil when no grid fired, and also nil
	// when the TSV held no marks at all (ErrNoMarks; Fallback states it). It is kept when
	// the placement threshold failed: the failed stats are the evidence for the fallback.
	Reconstruction *Stats
	// Fallback states why a fired grid's reconstruction was not used as the page text.
	// Empty when the reconstruction held (or no grid fired). A page whose grid could not
	// be rebuilt must say so — never emit a plausible half-table, never a silent zero.
	Fallback string
	// Diagnostics is what tesseract printed while reading this page — its resolution
	// estimate among it, which exists nowhere else. Captured, never printed (#644), and
	// kept for debugging: none of it is part of the reading or changes a byte it hashes.
	Diagnostics string
	// Evidence is what the grid branch read on the way to its answer; zero on a prose page.
	Evidence PageEvidence
}

// PageEvidence is the grid branch's intermediate readings, kept so a reconstruction — or
// its fallback — can be debugged after the run from what the engine actually saw. Like
// Diagnostics it is evidence, not the reading.
type PageEvidence struct {
	// PortraitTSV is the level-5 word TSV under PSMAuto on the page as scanned.
	PortraitTSV string
	// RotatedTSV is the same on the page turned 90° clockwise, when the orientation probe
	// ran; empty when it did not.
	RotatedTSV string
	// SparseTSV is the word TSV under PSMSparseText, on the orientation adopted.
	SparseTSV string
	// HeaderBand is the rotated-header band's re-OCR text, when a band was found.
	HeaderBand string
}

// MinMarkPlacement is the first reconstruction fallback threshold: below this placed/total
// ratio the emitted table is discarded for plain text, with the failure stated. It catches
// a SCATTERED placement and nothing else — OCR glyph dropout shrinks MarksTotal along with
// the marks, so the ratio stays green on a page missing most of its grid (p0052 at 33/36
// lost five of seventeen rows). MaxIntersectionRatio is the gate for that.
const MinMarkPlacement = 0.8

// MaxIntersectionRatio is the dropout gate: a reconstruction whose grid the detector
// measures at more than this many times the lattice the reconstruction accounts for
// (GridStats.Intersections / Stats.ExpectedIntersections) is discarded for plain text.
//
// The boundary is measured on one document (#644, 2026-09-11, every page checked against
// its 300-DPI pixels): the one reconstruction that held, p0054, at 1040/432 = 2.4; the
// four that did not — rows lost, marks miscounted, cells shifted a column — at p0051
// 3347/576 = 5.8, p0052 3295/195 = 16.9, p0050 2688/150 = 17.9 and p0053 2001/36 = 55.6.
// Wave 0 had called p0051 and p0052 healthy on placement alone.
//
// THE RATIO IS NOT UNITLESS, AND THAT LIMITS IT. The numerator counts crossing PIXELS (the
// AND of two morphological openings), the denominator lattice POINTS, so a healthy page
// sits near the rule thickness squared rather than 1; and the expected lattice is built
// from the columns that hold marks, so a sparse grid read correctly also scores high. Both
// errors fall the safe way — to plain text with the reason stated — and a second document
// (#934) is what tests where the boundary really sits.
const MaxIntersectionRatio = 4.0

// Rotation probe. A whole-page rotated table reads as near-silence under PSMAuto: the
// portrait TSV of p0051/p0052 held 9 confident words each where ordinary pages of this
// corpus hold over a hundred. A grid page under this floor is worth one extra TSV pass on
// the rotated pixels; the orientation with more confident words wins, so a misfire costs
// a pass and never the page.
const (
	rotProbeMaxConfidentWords = 25
	confidentWordMinConf      = 80
	confidentWordMinRunes     = 4
)

// Header-band recovery constants. Rotated (bottom-up) column headers appear in the
// portrait TSV as tall-narrow boxes tesseract reads in place with high confidence but
// low accuracy (Wave 0: 7 of p0054's 11 verbatim in place, 11 of 11 after band re-OCR).
// The band is the padded bounding box of those tall boxes; the pads re-create, from
// geometry, the hand-tuned rectangle the Wave 0 spike validated.
const (
	bandMinAspect = 3.0 // a header box is much taller than wide
	bandMinHeight = 60  // px at 300 DPI; excludes stray tall glyphs
	bandMinConf   = 50.0
	bandMinBoxes  = 3 // fewer tall boxes is not a header band
	bandPadX      = 45
	bandPadY      = 50
)

// ReadPage runs the whole per-page pipeline over one PNG. On a stub build the first
// engine call returns ErrNotCompiledIn and the page fails loudly — the caller must
// surface that, never record it as an empty reading.
//
// Tesseract's diagnostics for the page are captured into the result rather than printed.
func (en *Engine) ReadPage(pagePNG []byte, thr GridThresholds) (PageResult, error) {
	var res PageResult
	diag, err := captureDiagnostics(func() error {
		var rerr error
		res, rerr = en.readPage(pagePNG, thr)
		return rerr
	})
	if err != nil {
		return PageResult{}, err
	}
	res.Diagnostics = diag
	return res, nil
}

func (en *Engine) readPage(pagePNG []byte, thr GridThresholds) (PageResult, error) {
	grid, err := DetectGrid(pagePNG, thr)
	if err != nil {
		return PageResult{}, err
	}
	if !thr.Table(grid) {
		text, terr := en.PageText(pagePNG)
		if terr != nil {
			return PageResult{}, terr
		}
		return PageResult{Text: text, Grid: grid}, nil
	}
	return en.readGridPage(pagePNG, grid)
}

// readGridPage is the grid branch: orientation, TSV under both segmentation modes, the
// rotated-header band, reconstruction, and the stated fallback.
func (en *Engine) readGridPage(pagePNG []byte, grid GridStats) (PageResult, error) {
	out := PageResult{Table: true, Grid: grid}

	tsvAuto, err := en.PageTSV(pagePNG, PSMAuto)
	if err != nil {
		return PageResult{}, err
	}
	out.Evidence.PortraitTSV = tsvAuto

	// Orientation (plan §VI amendment a). Only probed when the portrait pass read almost
	// nothing; adopted only when the rotated pass measurably reads better. The rotation is
	// clockwise because that is the corpus's observed layout (both Wave 0 pages); a table
	// rotated the other way loses the probe and degrades through the stated fallback
	// below rather than emitting an upside-down reading.
	if confidentWordCount(tsvAuto) < rotProbeMaxConfidentWords {
		rotPNG, rerr := rotate90CW(pagePNG)
		if rerr != nil {
			return PageResult{}, rerr
		}
		rotTSV, rerr := en.PageTSV(rotPNG, PSMAuto)
		if rerr != nil {
			return PageResult{}, rerr
		}
		out.Evidence.RotatedTSV = rotTSV
		if confidentWordCount(rotTSV) > confidentWordCount(tsvAuto) {
			out.RotatedPage = true
			pagePNG, tsvAuto = rotPNG, rotTSV
		}
	}

	tsvSparse, err := en.PageTSV(pagePNG, PSMSparseText)
	if err != nil {
		return PageResult{}, err
	}
	out.Evidence.SparseTSV = tsvSparse

	// PSMAuto's layout analysis silently drops isolated marks it cannot attach to a text
	// block (Wave 0, p0052: zero mark tokens on a page with 74 marks); the sparse pass is
	// the second witness. It replaces the auto TSV only in the total-dropout case — per
	// page it otherwise helps or harms, and the geometric union of the two is named
	// future work in the Wave 0 record, not silently attempted here.
	marksAuto, marksSparse := MarkTokenCount(tsvAuto), MarkTokenCount(tsvSparse)
	tsv := tsvAuto
	if marksAuto == 0 && marksSparse > 0 {
		tsv = tsvSparse
	}

	var headers []string
	if x, y, w, h, ok := headerBand(tsv, pagePNG); ok {
		band, berr := en.RotatedBand(pagePNG, x, y, w, h)
		if berr != nil {
			return PageResult{}, berr
		}
		out.Evidence.HeaderBand = band
		headers = ParseRotatedBandHeaders(band)
	}

	table, st, rerr := Reconstruct(tsv, headers)
	if rerr != nil && rerr != ErrNoMarks {
		return PageResult{}, rerr
	}
	if rerr == nil {
		st.PSMDisagreement = PSMDisagreement(marksAuto, marksSparse)
		out.Reconstruction = &st
	}

	out.Fallback = fallbackReason(rerr, grid, st)
	if out.Fallback == "" {
		out.Text = table
		return out, nil
	}

	text, terr := en.PageText(pagePNG)
	if terr != nil {
		return PageResult{}, terr
	}
	out.Text = text
	return out, nil
}

// fallbackReason is the whole acceptance decision for a reconstruction, a pure function of
// what the page measured: "" when the reconstructed table may stand as the page text,
// otherwise the sentence the record carries for why it did not. Pure so the measured
// boundary pages can be pinned without the C stack.
func fallbackReason(rerr error, grid GridStats, st Stats) string {
	switch {
	case rerr == ErrNoMarks:
		// Every ruled table whose cells hold TEXT rather than marks lands here, and so does
		// the detector's standing false positive (p0025, boxed text). Over-detection degrades
		// to plain text WITH the failure stated — never a fabricated grid, never a silent
		// zero — and the sentence says what the reader loses, not only why: tesseract's plain
		// reading serializes such a table column by column, so a cell's words survive and its
		// row does not (#932).
		return "grid detected but the TSV held no mark tokens; page kept as plain text — " +
			"its words are read, but which row and column each belongs to is not preserved, " +
			"so check a cell against the page image before relying on it"
	case st.MarksTotal > 0 && float64(st.MarksPlaced)/float64(st.MarksTotal) < MinMarkPlacement:
		return fmt.Sprintf("reconstruction placed %d of %d marks, below the %.2f "+
			"placement threshold; page kept as plain text", st.MarksPlaced, st.MarksTotal, MinMarkPlacement)
	case float64(grid.Intersections) > MaxIntersectionRatio*float64(st.ExpectedIntersections()):
		return fmt.Sprintf("the detector measured %d grid intersections, %.1f times the %d the "+
			"reconstruction accounts for (limit %.1f): the OCR dropped rows or columns the grid "+
			"still shows; page kept as plain text", grid.Intersections,
			float64(grid.Intersections)/float64(st.ExpectedIntersections()),
			st.ExpectedIntersections(), MaxIntersectionRatio)
	}
	return ""
}

// confidentWordCount counts TSV words tesseract is sure of — the orientation signal. A
// rotated page's real text runs vertically and yields almost none; its horizontal margin
// line survives, which is why the floor is a count and not zero.
func confidentWordCount(tsv string) int {
	n := 0
	for _, w := range parseTSVWords(tsv) {
		if w.conf >= confidentWordMinConf && len([]rune(w.text)) >= confidentWordMinRunes {
			n++
		}
	}
	return n
}

// headerBand derives the rotated-header band rectangle from TSV geometry: the padded
// bounding box of tall-narrow high-confidence boxes sitting over the mark columns and
// above the first mark row. ok is false when the page has no such band — a table with
// horizontal headers, or a rotated page already normalized — and the caller then lets
// reconstruction fall back to supercolumn anchors.
func headerBand(tsv string, pagePNG []byte) (x, y, w, h int, ok bool) {
	words := parseTSVWords(tsv)
	var marks []tsvWord
	for _, wd := range words {
		if isStrongMark(wd.text) {
			marks = append(marks, wd)
		}
	}
	if len(marks) == 0 {
		return 0, 0, 0, 0, false
	}
	var colXs, rowYs []float64
	for _, m := range marks {
		colXs = append(colXs, m.cx())
		rowYs = append(rowYs, m.cy())
	}
	cols := centres(mergeSubSplits(cluster1D(colXs, colGap300), 0.55))
	sort.Float64s(rowYs)
	firstRow := rowYs[0]
	pitch := medianPitch(cols, 100)
	left, right := cols[0]-0.7*pitch, cols[len(cols)-1]+0.7*pitch

	minX, minY := 1<<30, 1<<30
	maxX, maxY, n := 0, 0, 0
	for _, wd := range words {
		if wd.conf < bandMinConf || float64(wd.h) < float64(wd.w)*bandMinAspect || wd.h < bandMinHeight {
			continue
		}
		if letters(wd.text) < 3 || wd.cx() < left || wd.cx() > right || float64(wd.y+wd.h) > firstRow {
			continue
		}
		n++
		minX, minY = min(minX, wd.x), min(minY, wd.y)
		maxX, maxY = max(maxX, wd.x+wd.w), max(maxY, wd.y+wd.h)
	}
	if n < bandMinBoxes {
		return 0, 0, 0, 0, false
	}
	// THE BAND'S BOTTOM IS THE FIRST MARK ROW'S CENTRE, not a pad below the tall boxes.
	// Short rotated headers ("Test", "Design") descend past the tall boxes' extent, all
	// the way to the table's top rule: measured on p0054, a bottom 20 px too high loses
	// exactly those three, while overlapping the first mark row costs nothing — the marks
	// read as junk the blank-line split discards.
	bottom := int(firstRow)
	if bottom <= maxY {
		bottom = maxY + bandPadY
	}
	// Clamp the pads to the image so leptonica is never asked to clip past the edge.
	cfg, err := png.DecodeConfig(bytes.NewReader(pagePNG))
	if err != nil {
		return 0, 0, 0, 0, false
	}
	x = max(minX-bandPadX, 0)
	y = max(minY-bandPadY, 0)
	return x, y, min(maxX+bandPadX, cfg.Width) - x, min(bottom, cfg.Height) - y, true
}

func letters(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			n++
		}
	}
	return n
}

// rotate90CW rotates a PNG 90° clockwise in pure Go — the whole-page orientation
// recovery. Grayscale out, matching what the render path produces; the OCR never sees
// colour either way.
func rotate90CW(pagePNG []byte) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pagePNG))
	if err != nil {
		return nil, fmt.Errorf("tessocr: rotate: %w", err)
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewGray(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(h-1-y, x, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("tessocr: rotate: %w", err)
	}
	return buf.Bytes(), nil
}
