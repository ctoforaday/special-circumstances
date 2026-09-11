package tessocr

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The pipeline's pure halves, tested against REAL recorded TSVs (no C stack needed):
// p0054.300.tsv is the portrait table page with rotated headers, p0051.300.tsv is the
// portrait TSV of a whole-page ROTATED table — the two ends of the orientation and
// band-recovery decisions.

func readFixtureTSV(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// THE ORIENTATION SIGNAL, PINNED ON BOTH SIDES OF THE FLOOR: the rotated page's portrait
// TSV reads as near-silence (its 9 confident words are the horizontal margin line), the
// portrait table page reads over a hundred. Both directions matter — a floor above the
// portrait page would rotate-probe every table, a floor below the rotated page would
// never recover one.
func TestConfidentWordCountSeparatesRotatedFromPortrait(t *testing.T) {
	rotated := confidentWordCount(readFixtureTSV(t, "p0051.300.tsv"))
	portrait := confidentWordCount(readFixtureTSV(t, "p0054.300.tsv"))
	if rotated >= rotProbeMaxConfidentWords {
		t.Errorf("p0051 (whole-page rotated) measured %d confident words, at or over the %d probe "+
			"floor — the rotation would never be probed", rotated, rotProbeMaxConfidentWords)
	}
	if portrait < rotProbeMaxConfidentWords {
		t.Errorf("p0054 (portrait) measured %d confident words, under the %d probe floor — every "+
			"ordinary table page would pay a rotation probe", portrait, rotProbeMaxConfidentWords)
	}
}

// THE BAND IS DERIVED, NOT HAND-TUNED — and this pins that the derivation lands on the
// rectangle the Wave 0 spike validated by hand (930,525,1260,330 recovered all 11 headers
// verbatim). The tall-narrow high-confidence boxes in p0054's portrait TSV
// ("Implementation" at 29x208 px and its ten siblings) are the locator.
func TestHeaderBandIsDerivedFromTallBoxes(t *testing.T) {
	pagePNG := grayPagePNG(t, 2550, 3300)
	x, y, w, h, ok := headerBand(readFixtureTSV(t, "p0054.300.tsv"), pagePNG)
	if !ok {
		t.Fatal("no header band derived from the page whose band recovery Wave 0 validated")
	}
	// The derived rectangle must COVER the tall header boxes (x 976–2084, y 573–780
	// measured from the fixture) and stay in the validated rectangle's neighbourhood: a
	// band that drifted far would crop away a header or swallow the grid.
	if x > 976 || x+w < 2084 || y > 573 || y+h < 780 {
		t.Errorf("band (%d,%d,%d,%d) does not cover the tall header boxes (x 976–2084, y 573–780)", x, y, w, h)
	}
	if x < 850 || y < 450 || y+h > 900 {
		t.Errorf("band (%d,%d,%d,%d) strays far from the validated rectangle (930,525,1260,330)", x, y, w, h)
	}
}

// A page with no tall boxes over its columns derives no band, and says so with ok=false —
// reconstruction then falls back to supercolumn anchors rather than re-OCRing a
// rectangle that holds nothing.
func TestHeaderBandAbsentIsStated(t *testing.T) {
	// A TSV with marks but no tall header boxes: three single marks, one label word.
	tsv := "level\tpage\tblock\tpar\tline\tword\tleft\ttop\twidth\theight\tconf\ttext\n" +
		"5\t1\t1\t1\t1\t1\t100\t500\t20\t22\t95\tX\n" +
		"5\t1\t1\t1\t1\t2\t208\t500\t20\t22\t95\tX\n" +
		"5\t1\t1\t1\t2\t1\t100\t560\t20\t22\t95\tX\n"
	if _, _, _, _, ok := headerBand(tsv, grayPagePNG(t, 600, 800)); ok {
		t.Error("a band was derived from a TSV with no tall header boxes")
	}
}

// The rotation is exactly 90° clockwise: src(x,y) lands at dst(h-1-y, x), dimensions
// swap, and the output is grayscale like everything the render path produces.
func TestRotate90CWGeometry(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 2, 3))
	src.SetGray(0, 0, color.Gray{Y: 255}) // top-left marker
	src.SetGray(1, 2, color.Gray{Y: 128}) // bottom-right marker
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	out, err := rotate90CW(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	got, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if b := got.Bounds(); b.Dx() != 3 || b.Dy() != 2 {
		t.Fatalf("rotated bounds = %v, want 3x2 (dimensions swapped)", b)
	}
	if r, _, _, _ := got.At(2, 0).RGBA(); r>>8 != 255 {
		t.Error("src(0,0) did not land at dst(h-1-0, 0) = (2,0) — not a clockwise rotation")
	}
	if r, _, _, _ := got.At(0, 1).RGBA(); r>>8 != 128 {
		t.Error("src(1,2) did not land at dst(h-1-2, 1) = (0,1)")
	}
	if _, err := rotate90CW([]byte("not a png")); err == nil {
		t.Error("garbage bytes rotated without error")
	}
}

// The placement threshold's both directions. Wave 0 listed 189/191, 129/135 and 33/36 as
// healthy; checked against the pixels (#644) only p0054's 189/191 is — the other two lost
// rows the ratio cannot see, and TestFallbackReasonOnTheMeasuredPages pins the gate that
// catches them. The constant is asserted so a change has to restate the boundary pages,
// not just edit a number.
func TestMinMarkPlacementHoldsTheMeasuredBoundary(t *testing.T) {
	if MinMarkPlacement != 0.8 {
		t.Errorf("MinMarkPlacement = %v — p0054 (189/191) bounds it from above; change it "+
			"with the boundary restated", MinMarkPlacement)
	}
	for _, healthy := range []struct{ placed, total int }{{189, 191}} {
		if ratio := float64(healthy.placed) / float64(healthy.total); ratio < MinMarkPlacement {
			t.Errorf("healthy page %d/%d would fall back under the threshold", healthy.placed, healthy.total)
		}
	}
}

// The acceptance decision on the #644 measurement: the five pages of IEEE 1012 whose
// reconstruction ran, each checked against its 300-DPI pixels. Only p0054 held; the other
// four clear the placement threshold and must be caught by the dropout gate. Deleting the
// gate's case from fallbackReason turns four rows here red.
func TestFallbackReasonOnTheMeasuredPages(t *testing.T) {
	for _, p := range []struct {
		page                         string
		gx, rows, sub, placed, total int
		held                         bool
	}{
		{"p0054", 1040, 35, 11, 189, 191, true},
		{"p0051", 3347, 17, 31, 129, 135, false},
		{"p0052", 3295, 12, 14, 33, 36, false},
		{"p0050", 2688, 14, 9, 22, 26, false},
		{"p0053", 2001, 8, 3, 10, 12, false},
	} {
		st := Stats{RowsFound: p.rows, SubColumnsFound: p.sub, MarksPlaced: p.placed, MarksTotal: p.total}
		reason := fallbackReason(nil, GridStats{Intersections: p.gx}, st)
		if held := reason == ""; held != p.held {
			t.Errorf("%s: held = %v (reason %q), want %v", p.page, held, reason, p.held)
			continue
		}
		if !p.held && !strings.Contains(reason, "grid intersections") {
			t.Errorf("%s fell back, but not on the dropout gate: %q", p.page, reason)
		}
	}
}

// Each fallback says why, and the two older reasons still fire on their own evidence.
func TestFallbackReasonStatesEachCause(t *testing.T) {
	noMarks := fallbackReason(ErrNoMarks, GridStats{Intersections: 200}, Stats{})
	for _, want := range []string{"no mark tokens", "row", "page image"} {
		if !strings.Contains(noMarks, want) {
			t.Errorf("no-marks fallback %q does not state %q", noMarks, want)
		}
	}
	scattered := fallbackReason(nil, GridStats{Intersections: 100},
		Stats{RowsFound: 10, SubColumnsFound: 5, MarksPlaced: 5, MarksTotal: 20})
	if !strings.Contains(scattered, "placed 5 of 20") {
		t.Errorf("scattered placement fell back as %q, want the placement reason", scattered)
	}
}

func grayPagePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
