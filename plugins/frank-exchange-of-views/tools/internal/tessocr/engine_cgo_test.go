//go:build tessocr && cgo

package tessocr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The embedded traineddata must be the pinned bytes — the pin in PINS.txt is the record,
// and an embed that drifted from it would ship an engine whose Identity() attests to
// language data it does not carry.
func TestEmbeddedTraineddataPin(t *testing.T) {
	const pinned = "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2"
	sum := sha256.Sum256(engTraineddata)
	if got := hex.EncodeToString(sum[:]); got != pinned {
		t.Fatalf("embedded eng.traineddata sha256 = %s, want the PINS.txt pin %s", got, pinned)
	}
	pins, err := os.ReadFile(filepath.Join("..", "..", "third_party", "pins", "PINS.txt"))
	if err != nil {
		t.Fatalf("reading PINS.txt: %v", err)
	}
	if !strings.Contains(string(pins), pinned) {
		t.Errorf("PINS.txt no longer carries the traineddata pin %s — update BOTH carriers together", pinned)
	}
}

// The one real image in the tree, end to end through the cgo pixel path: a 2180x800 crop
// of p0054's ruled grid, row labels included (105 KB — the boundary set proper is tested
// over recorded counts, which is why eleven full-page scans are not checked in). The
// counts are pinned exactly: the morphology is deterministic, so any drift is a change in
// the shim or the libraries, not noise.
func TestDetectGridOnRealCrop(t *testing.T) {
	png, err := os.ReadFile(filepath.Join("testdata", "gridcrop.png"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := DetectGrid(png, Grid300)
	if err != nil {
		t.Fatalf("DetectGrid: %v", err)
	}
	want := GridStats{HPix: 23375, VPix: 21440, Intersections: 367}
	if got != want {
		t.Errorf("DetectGrid = %+v, want %+v (measured at SEL=151 when the fixture was cut)", got, want)
	}
	if !Grid300.Table(got) {
		t.Errorf("the grid crop must clear the 300-DPI thresholds")
	}
}

func TestDetectGridRefusesBrokenImage(t *testing.T) {
	// A failed decode is an ERROR, never a page measuring zero — zero is what a clean
	// prose page legitimately measures.
	if _, err := DetectGrid([]byte("not a png"), Grid300); err == nil {
		t.Fatal("DetectGrid on garbage bytes returned no error")
	}
	if _, err := DetectGrid(nil, Grid300); err == nil {
		t.Fatal("DetectGrid on an empty image returned no error")
	}
}

// OCR of the real crop through the embedded-traineddata engine. Asserted loosely on
// CONTENT (row labels that are unmistakably on the crop) rather than pinned bytes: OCR
// output is deterministic for one build of the C stack, but the exact text is the
// stack's fact, not this package's, and pinning it would fail every legitimate library
// bump with a wall of diff.
func TestEngineOCRRealCrop(t *testing.T) {
	en, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer en.Close()

	png, err := os.ReadFile(filepath.Join("testdata", "gridcrop.png"))
	if err != nil {
		t.Fatal(err)
	}
	text, err := en.PageText(png)
	if err != nil {
		t.Fatalf("PageText: %v", err)
	}
	if len(text) < 100 {
		t.Fatalf("PageText returned %d bytes — the crop carries a header band and rows", len(text))
	}
	for _, wantWord := range []string{"Database", "Analysis"} {
		if !strings.Contains(text, wantWord) {
			t.Errorf("PageText output lacks %q; got:\n%s", wantWord, text)
		}
	}

	tsv, err := en.PageTSV(png, PSMAuto)
	if err != nil {
		t.Fatalf("PageTSV: %v", err)
	}
	if MarkTokenCount(tsv) == 0 {
		t.Errorf("PageTSV under PSMAuto found no mark tokens on the grid crop")
	}
}

func TestEngineRefusalsAfterClose(t *testing.T) {
	en, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	en.Close()
	en.Close() // double-close is safe by contract
	if _, err := en.PageText([]byte{1}); err == nil {
		t.Fatal("PageText on a closed engine returned no error")
	}
}

// The full per-page pipeline against the real corpus — env-gated because the corpus
// (~80 pages of 300-DPI scans) lives outside the repo. Point TESSOCR_CORPUS_DIR at a
// .pages directory to run it; CI and clean checkouts skip, and the skip is visible in
// verbose output rather than folded into a pass.
func TestEndToEndCorpus(t *testing.T) {
	dir := os.Getenv("TESSOCR_CORPUS_DIR")
	if dir == "" {
		t.Skip("TESSOCR_CORPUS_DIR not set; end-to-end corpus check skipped")
	}
	en, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer en.Close()

	// p0010: the prose benchmark page. No grid, plain text.
	prose, err := os.ReadFile(filepath.Join(dir, "p0010.png"))
	if err != nil {
		t.Fatal(err)
	}
	gs, err := DetectGrid(prose, Grid300)
	if err != nil {
		t.Fatal(err)
	}
	if Grid300.Table(gs) {
		t.Errorf("p0010 measured %+v — the prose page must not fire", gs)
	}
	text, err := en.PageText(prose)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("p0010: grid=%+v table=false text=%d bytes", gs, len(text))
	if len(text) < 2000 {
		t.Errorf("p0010 text is %d bytes; the page carries ~500 words", len(text))
	}

	// p0054: the table benchmark page, through the WHOLE pipeline — ReadPage runs detect,
	// the two TSV passes, the DERIVED rotated-band recovery and reconstruction — asserted
	// at the Wave 0 measurements. Driving ReadPage rather than the components is the
	// point: it proves the band derivation lands where the hand-tuned rectangle did.
	tablePng, err := os.ReadFile(filepath.Join(dir, "p0054.png"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := en.ReadPage(tablePng, Grid300)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Table || res.RotatedPage {
		t.Fatalf("p0054 = table:%v rotated:%v, want a portrait table page", res.Table, res.RotatedPage)
	}
	if res.Fallback != "" {
		t.Fatalf("p0054 fell back (%q); Wave 0 reconstructed it at 380/385", res.Fallback)
	}
	st := *res.Reconstruction
	t.Logf("p0054: grid=%+v stats=%+v expected_intersections=%d text_bytes=%d",
		res.Grid, st, st.ExpectedIntersections(), len(res.Text))
	if st.HeaderNamesFound != 11 {
		t.Errorf("derived band recovered %d headers, want 11 (Wave 0: all 11 verbatim)", st.HeaderNamesFound)
	}
	if st.ColumnsFound != 11 || st.RowsFound != 35 {
		t.Errorf("reconstruction found %dx%d, want 11 columns x 35 rows", st.ColumnsFound, st.RowsFound)
	}
	if st.MarksPlaced < 185 {
		t.Errorf("marks placed = %d, Wave 0 measured 189/191", st.MarksPlaced)
	}

	// p0051: the whole-page ROTATED table (plan §VI amendment a) — the pipeline must
	// probe, rotate, and reconstruct 10 columns via the Levels anchors.
	rotPng, err := os.ReadFile(filepath.Join(dir, "p0051.png"))
	if err != nil {
		t.Fatal(err)
	}
	rres, err := en.ReadPage(rotPng, Grid300)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("p0051: table=%v rotated=%v fallback=%q recon=%+v", rres.Table, rres.RotatedPage, rres.Fallback, rres.Reconstruction)
	if !rres.Table || !rres.RotatedPage {
		t.Errorf("p0051 = table:%v rotated:%v, want the rotation recovered", rres.Table, rres.RotatedPage)
	}
	if rres.Reconstruction != nil && rres.Fallback == "" && rres.Reconstruction.ColumnsFound != 10 {
		t.Errorf("p0051 reconstruction found %d columns, Wave 0 measured 10/10", rres.Reconstruction.ColumnsFound)
	}
}
