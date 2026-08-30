package fetchcache

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE FIXTURE IS THE ONE THAT MOTIVATED THE FEATURE: a page whose only content is an image,
// which is structurally what IEEE 1012 is and what PDFium reports zero characters for. It is
// built rather than pasted for the reason pdfextract_test.go gives — the absence of a text
// layer is the subject, so it should be legible in the source.

func TestRenderPagesWritesOneImagePerPageAndRecordsTheirHashes(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := pdfWithNoTextLayer()
	sha := Sha(body)

	rec, err := RenderPages(run, sha, body, 72)
	if err != nil {
		t.Fatalf("RenderPages: %v", err)
	}
	if rec.Pages() != 1 {
		t.Fatalf("Pages() = %d, want 1", rec.Pages())
	}
	if rec.DPI != 72 {
		t.Errorf("DPI = %d, want the resolution actually rendered at — a reader who gets a poor "+
			"reading cannot otherwise tell a bad model from a page rendered too small", rec.DPI)
	}
	if rec.Sha != sha {
		t.Errorf("Sha = %q, want the SOURCE document's hash %q", rec.Sha, sha)
	}
	if rec.Renderer == "" || !strings.Contains(rec.Renderer, "@") {
		t.Errorf("Renderer = %q, want library@semver", rec.Renderer)
	}
	if rec.RenderedAt.IsZero() {
		t.Error("RenderedAt is zero — a render predating a change to this code would be unidentifiable")
	}

	// The image is on disk, is a real PNG, and its hash is the one recorded.
	p := PagePath(run, sha, 1)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("page 1 image: %v", err)
	}
	if _, err := png.Decode(strings.NewReader(string(b))); err != nil {
		t.Errorf("page 1 is not a decodable PNG: %v", err)
	}
	if got := Sha(b); got != rec.PageShas[0] {
		t.Errorf("page 1 on disk hashes to %s, record says %s — the record names an image that is "+
			"not the one written", got, rec.PageShas[0])
	}
}

// PAGES ARE 1-BASED IN THE RECORD. PDFium indexes from 0 and a document is cited from 1; the
// conversion happens at the boundary so a second numbering never reaches a seat.
func TestPagePathIsOneBasedAndZeroPadded(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	got := PagePath(run, "abc", 1)
	if filepath.Base(got) != "p0001.png" {
		t.Errorf("PagePath(...,1) = %q, want it to end p0001.png", got)
	}
	if dir := filepath.Dir(got); dir != PagesDir(run, "abc") {
		t.Errorf("page image is not under PagesDir: %q vs %q", dir, PagesDir(run, "abc"))
	}
}

// THE RECORD IS WRITTEN LAST, so a crash leaves images with no record — which reads as "not
// rendered" and re-renders cleanly. The opposite order leaves a record naming images that do
// not exist, which is the failure that looks like success.
func TestReadRenderRecordReportsAbsenceAndRefusesCorruption(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if _, ok, err := ReadRenderRecord(run, "never-rendered"); err != nil || ok {
		t.Errorf("ReadRenderRecord on an unrendered document = (%v, %v), want (false, nil)", ok, err)
	}

	// A CORRUPT RECORD IS AN ERROR, NEVER AN ABSENCE. Reporting "nothing rendered" for
	// unparseable JSON would have a caller re-render over a directory it could not read.
	sha := "corrupt"
	if err := os.MkdirAll(PagesDir(run, sha), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(PagesDir(run, sha), "render.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, ok, err := ReadRenderRecord(run, sha)
	if err == nil {
		t.Error("a corrupt render record returned no error — the miss and the honest zero are the same bytes")
	}
	if ok {
		t.Error("a corrupt render record reported ok=true")
	}

	// And a record naming a DIFFERENT document is refused: a copied directory would otherwise
	// hand a reader another document's page hashes.
	other, _ := json.Marshal(RenderRecord{Sha: "someone-else", DPI: 200})
	if err := os.WriteFile(filepath.Join(PagesDir(run, sha), "render.json"), other, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadRenderRecord(run, sha); err == nil || !strings.Contains(err.Error(), "names document") {
		t.Errorf("a record naming another document was accepted: err = %v", err)
	}
}

// DPI IS BOUNDED IN BOTH DIRECTIONS. Below the floor a page is illegible and the reading is
// garbage for a reason nobody will attribute to resolution; above the ceiling the 534-page
// CJK chart in #636's corpus renders to gigabytes, and cost scales with the square.
func TestRenderPagesRefusesADPIOutsideTheBounds(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := pdfWithNoTextLayer()
	for _, dpi := range []int{0, MinRenderDPI - 1, MaxRenderDPI + 1, 5000} {
		if _, err := RenderPages(run, Sha(body), body, dpi); err == nil {
			t.Errorf("dpi %d was accepted; want a refusal naming the bounds", dpi)
		}
	}
	// The bounds themselves are usable — a guard that refused its own limits would be a
	// different bug wearing this one's clothes.
	for _, dpi := range []int{MinRenderDPI, DefaultRenderDPI} {
		if _, err := RenderPages(run, Sha(body), body, dpi); err != nil {
			t.Errorf("dpi %d was refused: %v", dpi, err)
		}
	}
}

// GARBAGE IN, A NAMED ERROR OUT — the same contract extraction has. What must never happen
// is a render that quietly produces nothing, because an empty pages directory is
// indistinguishable from a document nobody rendered.
func TestRenderPagesNamesWhatItCouldNotOpen(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := []byte("this is not a PDF at all")
	_, err := RenderPages(run, Sha(body), body, DefaultRenderDPI)
	if err == nil {
		t.Fatal("unparseable bytes rendered without error")
	}
	if !strings.Contains(err.Error(), "could not be opened") {
		t.Errorf("err = %v, want it to say the document could not be opened", err)
	}
}

// A DOCUMENT WITH A TEXT LAYER STILL RENDERS. The refusal that keeps a seat from burning
// tokens on a document already read belongs at the VERB, where --force can lift it; the
// library must not decide that for every caller.
func TestRenderPagesDoesNotItselfRefuseADocumentThatHasText(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := pdfWithTextLayer()
	rec, err := RenderPages(run, Sha(body), body, MinRenderDPI)
	if err != nil {
		t.Fatalf("a text-layer PDF was refused by the renderer: %v", err)
	}
	if rec.Pages() != 1 {
		t.Errorf("Pages() = %d, want 1", rec.Pages())
	}
}

// LookupSha answers "what do we know about these bytes", which is the question a --sha verb
// asks and which Lookup (keyed on URL) could not.
func TestLookupShaFindsAnEntryAndReportsAnHonestMiss(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	yes := true
	stored, err := Store(run, Entry{URL: "https://ex/a.pdf", ContentType: "application/pdf", TextExtracted: &yes}, []byte("body"))
	if err != nil {
		t.Fatal(err)
	}
	got, ok, err := LookupSha(run, stored.Sha)
	if err != nil || !ok {
		t.Fatalf("LookupSha(stored) = (%v, %v), want a hit", ok, err)
	}
	if got.URL != "https://ex/a.pdf" || got.ContentType != "application/pdf" {
		t.Errorf("LookupSha returned %+v, want the stored entry", got)
	}
	if _, ok, err := LookupSha(run, "0000000000000000000000000000000000000000000000000000000000000000"); err != nil || ok {
		t.Errorf("LookupSha(absent) = (%v, %v), want an honest miss", ok, err)
	}
	// An empty index is a miss, not an error — a run that has fetched nothing is a normal state.
	if _, ok, err := LookupSha(runtest.New(t, t.TempDir()), "anything"); err != nil || ok {
		t.Errorf("LookupSha on a run with no index = (%v, %v), want (false, nil)", ok, err)
	}
}

// THE TEST THAT CAUGHT A RECORD LYING ABOUT ITS OWN IMAGES, and the reason it asserts PIXELS
// rather than the record.
//
// writeAtomic treats an existing destination as already done — right for the content-addressed
// cache where the path is the hash, wrong here where the path is (document, page) and the
// content depends on the resolution. A re-render at 200 DPI therefore wrote a record saying
// 200 over images still rendered at 72. Every test that read RenderRecord.DPI passed: the
// field was correct, and it described an artifact that did not exist. Measured before the
// fix: dpi 72 and dpi 200 both produced 612x792 px at 2,969 bytes, byte-identical.
//
// So this measures the IMAGE. A US Letter page is 612x792 POINTS, and 72 points is an inch,
// so the rendered width at N DPI must be 612/72*N. That arithmetic is the only thing that can
// tell a resolution that reached the rasteriser from one that only reached the record.
func TestRenderedPixelsHonourTheRecordedDPI(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := pdfWithNoTextLayer()
	sha := Sha(body)

	var prev string
	for _, dpi := range []int{72, 150} {
		rec, err := RenderPages(run, sha, body, dpi)
		if err != nil {
			t.Fatalf("dpi %d: %v", dpi, err)
		}
		f, err := os.Open(PagePath(run, sha, 1))
		if err != nil {
			t.Fatal(err)
		}
		cfg, _, derr := image.DecodeConfig(f)
		f.Close()
		if derr != nil {
			t.Fatalf("dpi %d: decoding the page image: %v", dpi, derr)
		}
		if want := 612 * dpi / 72; cfg.Width != want {
			t.Errorf("dpi %d rendered %dx%d, want width %d — the resolution reached the record but "+
				"not the rasteriser", dpi, cfg.Width, cfg.Height, want)
		}
		// And the RECORD's hash is the hash of the image actually on disk, at every resolution.
		b, rerr := os.ReadFile(PagePath(run, sha, 1))
		if rerr != nil {
			t.Fatal(rerr)
		}
		if got := Sha(b); got != rec.PageShas[0] {
			t.Errorf("dpi %d: image hashes to %s, record says %s", dpi, got, rec.PageShas[0])
		}
		if prev != "" && rec.PageShas[0] == prev {
			t.Errorf("dpi %d produced a byte-identical image to the previous resolution — the "+
				"re-render did not replace the old pixels", dpi)
		}
		prev = rec.PageShas[0]
	}
}
