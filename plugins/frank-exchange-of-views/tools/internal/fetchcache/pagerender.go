package fetchcache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/klippa-app/go-pdfium/requests"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Page rendering turns a document with no text layer into images a SEAT can read (#644).
//
// #636 made the empty case honest: a scan records `text_extracted: false` with a reason
// naming optical recognition as the missing remedy. This is that remedy's deterministic
// half — the pages exist as images, hashed and recorded — and it deliberately stops there.
// Nothing here reads anything. What the images SAY is a later act by a seat, and keeping the
// two apart is what lets this half be verified offline with no model in the loop.

// DPI bounds. The floor is legibility and the ceiling is arithmetic: the 534-page Unicode
// CJK chart in #636's corpus renders to roughly 1.5 MB a page at 200 DPI, and cost scales
// with the square of the resolution. An unbounded --dpi is a way to fill a disk by typo.
const (
	MinRenderDPI     = 72
	MaxRenderDPI     = 400
	DefaultRenderDPI = 200
)

// RenderRecord is what a render leaves behind: one JSON document per source document,
// written beside the images it describes.
//
// IT IS NOT A FIELD ON Entry, AND THE REASON IS THE INDEX'S OWN SEMANTICS. The cache index
// is append-only and Lookup takes the FIRST match — "download-once: the first fetch's hash
// is canonical". That is right for a fetch and makes an UPDATE inexpressible: appending a
// second line for the same sha would either be ignored by a first-match read or quietly
// change what first-match means for every other reader. A render is also a distinct later
// act with its own facts (a resolution, a set of image hashes, a time), not a property of
// the fetch that produced the document — so it gets its own record rather than being
// forced into one whose merge rule cannot carry it.
//
// One document, one render record, replaced wholesale when the resolution changes. Every
// fact a reader needs is a field here; nothing is recovered from an image's filename.
type RenderRecord struct {
	// Sha is the SOURCE document's hash — the thing rendered, not the rendering.
	Sha string `json:"sha"`
	// DPI is the resolution these images were rendered at. RECORDED BECAUSE IT DECIDES
	// LEGIBILITY: without it, a reader who gets a poor reading back cannot tell a bad model
	// from a page rendered too small to read.
	DPI int `json:"dpi"`
	// PageShas is one sha256 per page image, in page order. Its LENGTH is the rendered page
	// count — a separate count field would be a second copy of the same fact, free to
	// disagree with the slice beside it.
	PageShas []string `json:"page_shas"`
	// RenderedAt is when, so a render predating a change to this code is identifiable.
	RenderedAt time.Time `json:"rendered_at"`
	// Renderer is library@semver, the same identity key #636 chose for extraction and for
	// the same reason: semver is what tracks a change in output.
	Renderer string `json:"renderer"`
}

// Pages is the rendered page count. It is derived from the slice rather than stored, so the
// two cannot drift.
func (r RenderRecord) Pages() int { return len(r.PageShas) }

// PagesDir holds one document's page images: <run>/cache/<sha>.pages/.
//
// A LOCATION, NOT A CARRIER. Every fact about the render lives in render.json inside it; the
// directory name exists so a human listing the cache can see which blob is which, exactly as
// TextPath's suffix does.
func PagesDir(run record.Run, sha string) string { return Path(run, sha) + ".pages" }

// PagePath is one page's image. n is 1-BASED, matching how a document is cited and how the
// seat reading it will refer to it — PDFium's own page index is 0-based and is converted at
// the boundary rather than leaking a second numbering into the record.
func PagePath(run record.Run, sha string, n int) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.png", n))
}

// renderRecordPath is the render's own record, beside the images.
func renderRecordPath(run record.Run, sha string) string {
	return filepath.Join(PagesDir(run, sha), "render.json")
}

// ReadRenderRecord returns the render for a document, and whether one exists.
//
// A MALFORMED RECORD IS AN ERROR, NEVER AN ABSENCE. Returning (zero, false, nil) for
// unparseable JSON would report "nothing rendered" in the same breath it reports "the
// record is corrupt", and a caller would re-render over a directory it could not read.
func ReadRenderRecord(run record.Run, sha string) (RenderRecord, bool, error) {
	b, err := os.ReadFile(renderRecordPath(run, sha))
	if os.IsNotExist(err) {
		return RenderRecord{}, false, nil
	}
	if err != nil {
		return RenderRecord{}, false, err
	}
	var r RenderRecord
	if err := json.Unmarshal(b, &r); err != nil {
		return RenderRecord{}, false, fmt.Errorf("render record for %s is unreadable: %w", sha, err)
	}
	if r.Sha != sha {
		return RenderRecord{}, false, fmt.Errorf("render record under %s names document %s", sha, r.Sha)
	}
	return r, true, nil
}

// RenderPages rasterises every page of a PDF and records what it produced.
//
// GRAYSCALE, ALWAYS. PDFium hands back an RGBA raster and a scan carries no colour worth
// keeping: measured 2026-08-30 on a letter page of scan-like noise at 200 DPI, RGBA is
// 3.37 MB of PNG and grayscale is 2.31 MB — 31% less disk, 31% less to base64 and upload,
// and nothing the reader could have used either way. (The other lever does not exist:
// png.BestCompression produced byte-identical output to the default at every resolution,
// because scan noise is incompressible. The colour model was the whole saving.)
//
// It is unconditional rather than a flag. A flag would be a fact about the pipeline that the
// record would then have to carry, and inventing a field for a constant is the cost
// [[facts-are-fields]] warns against paying; the day a caller genuinely needs colour, the
// option and the field arrive together.
//
// The images are written first and the record last, so a crash leaves images with no record
// — which reads as "not rendered" and re-renders cleanly. The opposite order would leave a
// record naming images that do not exist, which is the failure that looks like success.
func RenderPages(run record.Run, sha string, body []byte, dpi int) (RenderRecord, error) {
	if dpi < MinRenderDPI || dpi > MaxRenderDPI {
		return RenderRecord{}, fmt.Errorf("dpi %d is outside %d–%d: below the floor a page is "+
			"illegible, above the ceiling a long document renders to gigabytes", dpi, MinRenderDPI, MaxRenderDPI)
	}
	// A RENDER REPLACES A RENDER, WHOLESALE, AND THE DIRECTORY IS CLEARED TO MAKE THAT TRUE.
	//
	// writeAtomic treats an existing destination as already done — correct for the
	// content-addressed cache, where the path IS the hash, and wrong here, where the path is
	// (document, page number) and the CONTENT depends on the resolution. Without this, a
	// re-render at a new DPI wrote a record saying 200 over images still rendered at 72: the
	// field asserted a resolution the pixels did not have, and every reader downstream would
	// have believed the field. Measured, before the clear: 72 and 200 both produced 612x792.
	//
	// Clearing first also makes a failed re-render safe. The record is written last, so a
	// crash leaves images with no record, which reads as "not rendered" and re-renders — the
	// same ordering the whole function relies on. What it costs is the previous render, and
	// that is a cache of something the document can always reproduce.
	// AND THE READING OF THOSE PIXELS GOES WITH THEM. The reading record and its per-pass
	// exhibits already died here — they live inside this directory — but the assembled
	// transcription does not, and it was being left behind: a <sha>.ocr.txt on disk marking
	// disagreements against pass files that no longer existed, attesting to a model, a time and
	// a set of image hashes whose record had just been deleted. That is a transcription
	// outliving its own provenance, which is the state this whole design exists to prevent, and
	// it read to a seat exactly like a reading that was still current.
	dir := PagesDir(run, sha)
	if err := os.RemoveAll(dir); err != nil {
		return RenderRecord{}, fmt.Errorf("clearing the previous render of %s: %w", sha, err)
	}
	if err := os.Remove(OCRTextPath(run, sha)); err != nil && !os.IsNotExist(err) {
		return RenderRecord{}, fmt.Errorf("clearing the previous reading of %s: %w", sha, err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return RenderRecord{}, err
	}

	inst, closer, err := pdfiumInstance(Dir(run))
	if err != nil {
		return RenderRecord{}, err
	}
	defer closer()

	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &body})
	if err != nil {
		return RenderRecord{}, fmt.Errorf("pdf could not be opened: %w", err)
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return RenderRecord{}, fmt.Errorf("pdf page count unavailable: %w", err)
	}
	if pc.PageCount == 0 {
		return RenderRecord{}, fmt.Errorf("the document reports zero pages, so there is nothing to render")
	}

	shas := make([]string, 0, pc.PageCount)
	for i := 0; i < pc.PageCount; i++ {
		rendered, rerr := inst.RenderPageInDPI(&requests.RenderPageInDPI{
			Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: i}},
			DPI:  dpi,
		})
		if rerr != nil {
			// NAMED AND FATAL. A page that silently did not render would leave the record one
			// image short of the document, and nothing downstream could tell that from a
			// document with fewer pages.
			return RenderRecord{}, fmt.Errorf("page %d of %d did not render: %w", i+1, pc.PageCount, rerr)
		}
		if rendered.Result.RenderedImage == nil {
			return RenderRecord{}, fmt.Errorf("page %d of %d rendered to no image", i+1, pc.PageCount)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, toGray(rendered.Result.RenderedImage)); err != nil {
			return RenderRecord{}, fmt.Errorf("encoding page %d: %w", i+1, err)
		}
		if err := writeAtomic(PagePath(run, sha, i+1), buf.Bytes()); err != nil {
			return RenderRecord{}, err
		}
		shas = append(shas, Sha(buf.Bytes()))
	}

	rec := RenderRecord{
		Sha:        sha,
		DPI:        dpi,
		PageShas:   shas,
		RenderedAt: time.Now().UTC(),
		Renderer:   extractorIdentity(),
	}
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return RenderRecord{}, err
	}
	if err := writeAtomic(renderRecordPath(run, sha), append(b, '\n')); err != nil {
		return RenderRecord{}, err
	}
	return rec, nil
}

// toGray converts a rendered page to a single-channel image.
//
// draw.Draw into a Gray destination does the conversion through color.GrayModel — the
// luminance weighting — rather than dropping channels, so coloured text keeps the contrast
// against paper that makes it legible. Dropping to the red channel instead would render red
// ink as white.
func toGray(src image.Image) *image.Gray {
	if g, ok := src.(*image.Gray); ok {
		return g
	}
	g := image.NewGray(src.Bounds())
	draw.Draw(g, g.Bounds(), src, src.Bounds().Min, draw.Src)
	return g
}
