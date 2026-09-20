package tessocr

// THE GEOMETRY LAYER READS ONE COORDINATE SPACE (#1074).
//
// Every constant below the detector — ruleJoin, minBandHeight, minRowGap, the clustering gaps, the
// mark minima, cellPad, the same-line tolerances — was fitted against pixels at 300 DPI. Until
// #1058 that was safe because every page rendered at 300. It no longer is: a page renders at its
// own scan resolution, floored 300 and capped 600, so a constant fitted at 300 was being compared
// against coordinates up to twice as large. Measured on usfs-birds-markgrid (native 501), one
// physical table read at 300/350/400/501/600 reported five different shapes — 2 bands and 7 rows at
// 300, 3 and 10 at 501, 2 and 18 at 600.
//
// THE FIX IS NOT TWENTY DERIVED CONSTANTS. That was the first design, and its census was missing
// twelve literals when it was audited — medianPitch's fallback pitches, gtolC's 16 px arm, four
// same-line tolerances, a second mark-minimum site. A hand-kept census with nothing that fails when
// the thirteenth is written is the same defect one layer up: it is how #1058 produced this one.
//
// Instead the MEASUREMENTS are converted into the space the constants were fitted in, once, at the
// boundary into the geometry layer, and everything below is left exactly as measured. A quantity is
// a length, an area, or neither; the transform follows from which, not from a list. At 300 DPI the
// scale is 1.0 and every conversion is the identity, so a page at 300 reads byte for byte as it did
// before this existed.
//
// What does NOT cross: confidences and text (not geometry), counts of things (rows, columns,
// words), and ratios (coverage, bandMinAspect, MinMarkPlacement). What comes BACK: a Cell handed to
// CropCell addresses the page image, so it is converted to page pixels at that one site, and
// headerBand is given the native words for the same reason.

// normScale takes a page-pixel length at dpi into the 300-DPI space. A dpi of 0 means the caller
// built a GridThresholds by hand without one; the scale is then 1.0, which is the pre-#1074
// behaviour and the only answer that cannot silently corrupt a hand-built fixture.
func normScale(dpi int) float64 {
	if dpi <= 0 {
		return 1.0
	}
	return float64(tuneDPI) / float64(dpi)
}

// normalizeLattice puts a page's rule geometry into the 300-DPI space.
func normalizeLattice(lat Lattice, scale float64) Lattice {
	if scale == 1.0 {
		return lat
	}
	out := Lattice{
		Horizontal: make([]Rule, len(lat.Horizontal)),
		Vertical:   make([]Rule, len(lat.Vertical)),
	}
	for i, r := range lat.Horizontal {
		out.Horizontal[i] = scaleRule(r, scale)
	}
	for i, r := range lat.Vertical {
		out.Vertical[i] = scaleRule(r, scale)
	}
	return out
}

// scaleRule scales a rule's ORIGIN and its FAR EDGE, then takes the width back off the two, because
// the far edge is what bounds a column (Lattice.Cells reads the outermost rule's extent) and
// scaling X and W independently does not preserve it.
//
// STATED BECAUSE IT WAS CHECKED AND IS NOT LOAD-BEARING TODAY: scaling W directly instead differs by
// at most one pixel, on a rule a few pixels thick, and every tolerance downstream is 12 px or more.
// The delete-the-row pass on this file killed the length transform, the squared crossings and the
// cell map-back; this one survived, and no fixture distinguishes the two forms. It is kept as the
// faithful transform, not as a fix for an observed defect, and a comment claiming otherwise would be
// a behaviour nothing tests.
func scaleRule(r Rule, scale float64) Rule {
	x0, y0 := round(float64(r.X)*scale), round(float64(r.Y)*scale)
	x1, y1 := round(float64(r.X+r.W)*scale), round(float64(r.Y+r.H)*scale)
	return Rule{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
}

// normalizeWords puts TSV word boxes into the 300-DPI space, leaving confidence and text alone.
func normalizeWords(words []tsvWord, scale float64) []tsvWord {
	if scale == 1.0 {
		return words
	}
	out := make([]tsvWord, len(words))
	for i, w := range words {
		x0, y0 := round(float64(w.x)*scale), round(float64(w.y)*scale)
		x1, y1 := round(float64(w.x+w.w)*scale), round(float64(w.y+w.h)*scale)
		out[i] = tsvWord{x: x0, y: y0, w: x1 - x0, h: y1 - y0, conf: w.conf, text: w.text}
	}
	return out
}

// normalizeCrossings puts the detector's crossing-PIXEL count into the 300-DPI space. It counts the
// pixels of a shape whose physical size is fixed, so it scales with the SQUARE of the resolution —
// the distinction GridFor's comment makes for MinHPix and MinVPix, and the reason MaxIntersectionRatio
// could not simply be reused at another DPI: its numerator is this count and its denominator is a
// count of lattice POINTS, which does not scale at all.
func normalizeCrossings(pixels int, scale float64) int {
	if scale == 1.0 {
		return pixels
	}
	return round(float64(pixels) * scale * scale)
}

// pageCell converts a cell back to page-image pixels, for the one consumer that crops the image
// with it.
func pageCell(c Cell, scale float64) Cell {
	if scale == 1.0 {
		return c
	}
	// The cell is COPIED and only its rectangle moved: Row, Col, Span and the words it bound are
	// facts about the table, not about the page image, and rebuilding the struct field by field is
	// how a later field silently stops being carried.
	inv := 1.0 / scale
	out := c
	out.X0, out.Y0 = round(float64(c.X0)*inv), round(float64(c.Y0)*inv)
	out.X1, out.Y1 = round(float64(c.X1)*inv), round(float64(c.Y1)*inv)
	return out
}

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}
