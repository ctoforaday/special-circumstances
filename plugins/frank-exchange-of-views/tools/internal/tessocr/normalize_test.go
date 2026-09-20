package tessocr

import (
	"fmt"
	"testing"
)

// THE PROPERTY #1074 BUYS, PINNED WHERE IT CAN BE PINNED EXACTLY.
//
// The same page, printed at the same physical size, scanned at a different resolution, must produce
// the same TABLE. Not similar — the same: the rules and the words are in the same places on paper,
// and the only thing that changed is how many pixels describe them.
//
// This is a PURE-GO test, and that is deliberate. The obvious version reads one real scan at two
// resolutions through the engine and demands equal structure; it would be wrong, because tesseract
// reads different pixels at 300 and 600 and returns different WORDS, so the structure legitimately
// differs and the test could only be met by weakening it. Here the words are held fixed and only
// the coordinate system moves, which isolates exactly the arithmetic this change fixes. It also
// runs on every CI leg, including Windows, and costs no engine time.
//
// Before #1074 the geometry read page pixels directly, so the 600-DPI arm of this test produced a
// different table from the 300-DPI arm and this test did not pass.
func TestGeometryIsResolutionInvariant(t *testing.T) {
	// One ruled table of text cells, as measured at 300 DPI: three horizontal rules, four
	// verticals, a heading row and a wrapped body row, with a running head above and a folio below
	// so the page around the table is in the comparison too.
	words300 := []word{
		{"Running", 100, 40, 90, 20, 96}, {"head", 200, 40, 60, 20, 96},
		{"Criticality", 110, 150, 120, 20, 95}, {"Description", 410, 150, 130, 20, 95}, {"Notes", 710, 150, 70, 20, 95},
		{"High", 110, 250, 60, 20, 95}, {"Affects", 410, 250, 80, 20, 95}, {"critical", 500, 250, 80, 20, 95},
		{"performance", 410, 275, 140, 20, 95}, {"of", 560, 275, 30, 20, 95}, {"the", 600, 275, 40, 20, 95},
		{"system", 650, 275, 70, 20, 95}, {"no", 710, 250, 40, 20, 95}, {"workaround", 710, 275, 120, 20, 95},
		{"26", 120, 500, 30, 20, 96},
	}
	rules300 := []string{
		"h 100 100 900 3", "h 100 200 900 3", "h 100 300 900 3",
		"v 100 100 3 203", "v 400 100 3 203", "v 700 100 3 203", "v 1000 100 3 203",
	}

	baseText, baseStats, why := TextCells(
		normalizeLattice(ParseLattice(ruleDump(rules300...)), normScale(300)),
		normalizeWords(parseTSVWords(fakeTSV(words300)), normScale(300)), nil)
	if why != "" {
		t.Fatalf("the 300-DPI baseline refused, so there is nothing to compare: %s", why)
	}

	// 450 is the half-pixel case: at 1.5x an odd coordinate lands between two pixels on the way out
	// and has to come back to where it started. 600 is the clean doubling. Both are inside the band
	// RenderDPIFor renders in.
	for _, dpi := range []int{450, 600} {
		t.Run(fmt.Sprintf("%ddpi", dpi), func(t *testing.T) {
			up := float64(dpi) / float64(tuneDPI)
			scale := normScale(dpi)

			lat := normalizeLattice(scaleLatticeForTest(ParseLattice(ruleDump(rules300...)), up), scale)
			words := normalizeWords(scaleWordsForTest(parseTSVWords(fakeTSV(words300)), up), scale)

			text, st, why := TextCells(lat, words, nil)
			if why != "" {
				t.Fatalf("refused at %d DPI while the same page held at 300: %s", dpi, why)
			}
			if text != baseText {
				t.Errorf("the same page read a different table at %d DPI.\n--- at 300\n%s\n--- at %d\n%s",
					dpi, baseText, dpi, text)
			}
			if st.Rows != baseStats.Rows || st.Columns != baseStats.Columns ||
				st.Cells != baseStats.Cells || st.Bands != baseStats.Bands {
				t.Errorf("structure at %d DPI = %d bands %d rows %d columns %d cells, want 300's %d/%d/%d/%d",
					dpi, st.Bands, st.Rows, st.Columns, st.Cells,
					baseStats.Bands, baseStats.Rows, baseStats.Columns, baseStats.Cells)
			}
		})
	}
}

// The dropout gate's limit is a RATIO of two things that scale differently — crossing pixels over
// lattice points — so it is the one measurement that needs the square. A page's ratio must read the
// same whatever it was scanned at; if normalizeCrossings scaled by dpi rather than dpi squared, or
// not at all, this is what would catch it.
func TestTheDropoutRatioIsResolutionInvariant(t *testing.T) {
	const crossings300 = 1040 // p0054, the one Wave 0 reconstruction that held
	st := Stats{RowsFound: 35, SubColumnsFound: 11}
	want := float64(crossings300) / float64(st.ExpectedIntersections())

	for _, dpi := range []int{300, 350, 450, 501, 600} {
		up := float64(dpi) / float64(tuneDPI)
		atDPI := round(float64(crossings300) * up * up) // what the detector would count at that DPI
		got := float64(normalizeCrossings(atDPI, normScale(dpi))) / float64(st.ExpectedIntersections())
		if rel := abs(got-want) / want; rel > 0.01 {
			t.Errorf("at %d DPI the ratio reads %.3f, want %.3f (%.1f%% off): the gate's limit is "+
				"measured at 300 and this is what it would be compared against", dpi, got, want, rel*100)
		}
	}
}

// A page at 300 DPI must be untouched, not approximately untouched: the transform is the identity
// there, which is what lets this change claim it moves no 300-DPI reading.
func TestNormalizingAt300ChangesNothing(t *testing.T) {
	lat := ParseLattice(ruleDump("h 100 100 900 3", "v 100 100 3 203"))
	words := parseTSVWords(fakeTSV([]word{{"High", 110, 250, 60, 20, 95}}))

	if got := normalizeLattice(lat, normScale(300)); !sameLattice(got, lat) {
		t.Errorf("the lattice moved at 300 DPI: %+v, want %+v", got, lat)
	}
	if got := normalizeWords(words, normScale(300)); !sameWords(got, words) {
		t.Errorf("the words moved at 300 DPI: %+v, want %+v", got, words)
	}
	if got := normalizeCrossings(1040, normScale(300)); got != 1040 {
		t.Errorf("crossings = %d at 300 DPI, want 1040", got)
	}
	c := Cell{Row: 2, Col: 3, Span: 1, X0: 10, Y0: 20, X1: 30, Y1: 40, words: words}
	got := pageCell(c, normScale(300))
	if got.X0 != c.X0 || got.Y0 != c.Y0 || got.X1 != c.X1 || got.Y1 != c.Y1 {
		t.Errorf("the cell's rectangle moved at 300 DPI: %+v, want %+v", got, c)
	}
	// The rest of the cell must survive the conversion whatever the scale — a map-back that
	// rebuilds the struct drops the words and the reading goes empty with every count intact.
	if got.Row != c.Row || got.Col != c.Col || got.Span != c.Span || got.Text() != c.Text() {
		t.Errorf("converting a cell dropped part of it: %+v, want %+v", got, c)
	}
	if moved := pageCell(c, normScale(600)); moved.Row != c.Row || moved.Text() != c.Text() {
		t.Errorf("converting a cell at 600 DPI dropped part of it: %+v", moved)
	}
}

// A cell crops the page IMAGE, so it has to come back to the resolution the image is at. Round trip
// it and the cell must land where it started, within the rounding the conversion costs.
func TestACellComesBackToPagePixels(t *testing.T) {
	for _, dpi := range []int{350, 450, 501, 600} {
		up := float64(dpi) / float64(tuneDPI)
		page := Cell{X0: 240, Y0: 700, X1: 980, Y1: 1160} // a cell in page pixels at dpi
		page = Cell{
			X0: round(float64(page.X0) * up), Y0: round(float64(page.Y0) * up),
			X1: round(float64(page.X1) * up), Y1: round(float64(page.Y1) * up),
		}
		scale := normScale(dpi)
		back := pageCell(Cell{
			X0: round(float64(page.X0) * scale), Y0: round(float64(page.Y0) * scale),
			X1: round(float64(page.X1) * scale), Y1: round(float64(page.Y1) * scale),
		}, scale)
		for _, d := range []struct {
			name     string
			got, was int
		}{{"X0", back.X0, page.X0}, {"Y0", back.Y0, page.Y0}, {"X1", back.X1, page.X1}, {"Y1", back.Y1, page.Y1}} {
			if diff := d.got - d.was; diff > 1 || diff < -1 {
				t.Errorf("at %d DPI the cell's %s came back %d px away (%d, was %d)", dpi, d.name, diff, d.got, d.was)
			}
		}
	}
}

func scaleLatticeForTest(lat Lattice, up float64) Lattice {
	out := Lattice{}
	for _, r := range lat.Horizontal {
		out.Horizontal = append(out.Horizontal, scaleRule(r, up))
	}
	for _, r := range lat.Vertical {
		out.Vertical = append(out.Vertical, scaleRule(r, up))
	}
	return out
}

func scaleWordsForTest(ws []tsvWord, up float64) []tsvWord { return normalizeWords(ws, up) }

func sameLattice(a, b Lattice) bool {
	return fmt.Sprint(a.Horizontal) == fmt.Sprint(b.Horizontal) &&
		fmt.Sprint(a.Vertical) == fmt.Sprint(b.Vertical)
}

func sameWords(a, b []tsvWord) bool { return fmt.Sprint(a) == fmt.Sprint(b) }
