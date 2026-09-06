package tessocr

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
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

// The fallback threshold's both directions, on the Wave 0 boundary evidence: every
// healthy reconstruction measured at or above 0.92 placed/total, so the 0.80 floor keeps
// them all while a genuinely scattered placement falls to plain text. The constant is
// asserted so a change has to restate the boundary pages, not just edit a number.
func TestMinMarkPlacementHoldsTheMeasuredBoundary(t *testing.T) {
	if MinMarkPlacement != 0.8 {
		t.Errorf("MinMarkPlacement = %v — the Wave 0 pages bound it (healthy: 189/191, 129/135, "+
			"33/36, all ≥ 0.91); change it with the boundary restated", MinMarkPlacement)
	}
	for _, healthy := range []struct{ placed, total int }{{189, 191}, {129, 135}, {33, 36}} {
		if ratio := float64(healthy.placed) / float64(healthy.total); ratio < MinMarkPlacement {
			t.Errorf("Wave 0 healthy page %d/%d would fall back under the threshold", healthy.placed, healthy.total)
		}
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
