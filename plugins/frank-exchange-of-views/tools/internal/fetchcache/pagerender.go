package fetchcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/enums"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// Page rendering turns a document with no text layer into images a SEAT can read (#644).
//
// #636 made the empty case honest: a scan records `text_extracted: false` with a reason
// naming optical recognition as the missing remedy. This is that remedy's deterministic
// half — the pages exist as images, hashed and recorded — and it deliberately stops there.
// Nothing here reads anything. What the images SAY is a later act by a seat, and keeping the
// two apart is what lets this half be verified offline with no model in the loop.

// DPI bounds. The floor is legibility and the ceiling is arithmetic: cost scales with the
// square of the resolution, and an unbounded --dpi is a way to fill a disk by typo —
// renderWithinDiskBudget states the budget that enforces it.
//
// THE DEFAULT IS THE ENGINE'S OPERATIVE RESOLUTION, derived rather than copied so the two
// cannot drift. It was 200 while a model read the pages — 200 DPI sat exactly at that
// model's vision ceiling, and rendering higher bought pixels the API downscaled away —
// and that justification left with the model. 300 is where the local engine's table
// geometry works: full column coverage in word boxes and a perfect rotated-header
// recovery, where 200 loses half the grid (plans/local-ocr.md; the measured cost is
// prose WER 0.40%→0.80%, ruled acceptable there).
const (
	MinRenderDPI = 72
	// MaxRenderDPI is also the cap on following a scan's own resolution (RenderDPIFor): past it a
	// long document renders to gigabytes for nothing, and rendering ABOVE a scan's native
	// resolution was measured to gain no text and to cost a false table on a boundary page (#1031).
	MaxRenderDPI     = 600
	DefaultRenderDPI = tessocr.RenderDPI
)

// RenderDPIFor is the resolution to render a scan at, given the resolution its own images carry.
//
// MEASURED (#1031), three regimes, and the reader was on the wrong side of two of them:
//
//   - BELOW 300 the page is read small and content is lost: a 150-DPI page recovered 2 of 4 printed
//     strings, and at 300 it recovered 4 of 4 and its table verdict with them. So a low-resolution
//     scan is rendered UP to the floor.
//   - AT the scan's own resolution nothing is resampled and nothing is thrown away. Rendering the
//     corpus at native instead of 300 took its expect-failures from 13 to 10, recovering a contents
//     page's page numbers and a cover's publisher line — content a fixed 300 was destroying, since
//     these scans are natively 350 and 501.
//   - ABOVE native there is nothing to recover: 400 and 600 renders read flat to worse, and at 600
//     the tuning document's closest clean rejection flipped to a false table.
//
// So: the scan's own resolution, floored at the engine's comfort zone and capped for memory.
func RenderDPIFor(nativeDPI int) int {
	switch {
	case nativeDPI < DefaultRenderDPI:
		return DefaultRenderDPI
	case nativeDPI > MaxRenderDPI:
		return MaxRenderDPI
	default:
		return nativeDPI
	}
}

// NativeDPIs is the resolution each page carries: for every page, the highest resolution of any
// image on it, in page order.
//
// PER PAGE, NOT PER DOCUMENT. A first cut took the median across pages on the grounds that real
// scans are uniform — and they mostly are, SP 602 running 350-351 over 60 pages — but a median is a
// statistic about a document, and the resolution is a fact about a PAGE. A document that mixes (a
// fold-out plate scanned finer, an inserted exhibit scanned coarser) is exactly the case where the
// summary is wrong, and it is wrong silently: every page but one renders at a resolution none of
// them has.
//
// Zero for a page means it carries no image to measure — a born-digital page, or one drawn rather
// than scanned. The caller renders that page at the floor.
func NativeDPIs(run record.Run, body []byte) ([]int, error) {
	inst, closer, err := pdfiumInstance(Dir(run))
	if err != nil {
		return nil, err
	}
	defer closer()
	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &body})
	if err != nil {
		return nil, fmt.Errorf("pdf could not be opened: %w", err)
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return nil, fmt.Errorf("pdf page count unavailable: %w", err)
	}
	perPage := make([]int, 0, pc.PageCount)
	for i := 0; i < pc.PageCount; i++ {
		page := requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: i}}
		n, cerr := inst.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{Page: page})
		if cerr != nil {
			return nil, fmt.Errorf("page %d: %w", i+1, cerr)
		}
		best := 0
		for j := 0; j < n.Count; j++ {
			o, gerr := inst.FPDFPage_GetObject(&requests.FPDFPage_GetObject{Page: page, Index: j})
			if gerr != nil {
				return nil, fmt.Errorf("page %d object %d: %w", i+1, j, gerr)
			}
			ty, terr := inst.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: o.PageObject})
			if terr != nil || ty.Type != enums.FPDF_PAGEOBJ_IMAGE {
				continue
			}
			md, merr := inst.FPDFImageObj_GetImageMetadata(&requests.FPDFImageObj_GetImageMetadata{
				ImageObject: o.PageObject, Page: page,
			})
			if merr != nil {
				continue
			}
			if d := int(md.ImageMetadata.HorizontalDPI + 0.5); d > best {
				best = d
			}
		}
		perPage = append(perPage, best)
	}
	return perPage, nil
}

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
	// Renders is one row per page image, in page order, each carrying the resolution THAT PAGE
	// was rendered at. Its LENGTH is the rendered page count — a separate count field would be a
	// second copy of the same fact, free to disagree with the slice beside it.
	//
	// PER PAGE BECAUSE THE RESOLUTION IS A FACT ABOUT A PAGE (#1031): a scan is rendered at its
	// own resolution, and a document that mixes them has no single number to record. A document
	// field here would be that number, wrong for every page but one, and silent about it.
	Renders []PageRender `json:"renders"`
	// RenderedAt is when, so a render predating a change to this code is identifiable.
	RenderedAt time.Time `json:"rendered_at"`
	// Renderer is library@semver, the same identity key #636 chose for extraction and for
	// the same reason: semver is what tracks a change in output.
	Renderer string `json:"renderer"`
}

// PageRender is one rendered page: what it hashes to, and the resolution it was drawn at.
//
// RECORDED BECAUSE IT DECIDES LEGIBILITY: without it, a reader who gets a poor reading back cannot
// tell a bad engine from a page rendered too small to read.
type PageRender struct {
	Sha string `json:"sha"`
	DPI int    `json:"dpi"`
}

// Pages is the rendered page count. It is derived from the slice rather than stored, so the
// two cannot drift.
func (r RenderRecord) Pages() int { return len(r.Renders) }

// DPIRange is the lowest and highest resolution any page here was rendered at.
func (r RenderRecord) DPIRange() (lo, hi int) {
	for i, p := range r.Renders {
		if i == 0 || p.DPI < lo {
			lo = p.DPI
		}
		if p.DPI > hi {
			hi = p.DPI
		}
	}
	return lo, hi
}

// Shas is the page hashes in page order, for readers that only need identity.
func (r RenderRecord) Shas() []string {
	out := make([]string, len(r.Renders))
	for i, p := range r.Renders {
		out[i] = p.Sha
	}
	return out
}

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
// 3.37 MB of PNG and grayscale is 2.31 MB — 31% less disk, less to decode, and nothing
// the engine could have used either way. (The other lever does not exist:
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
	// The disk budget fires here, after the count is known and before the first raster:
	// this is the verb that PERSISTS pixels, so it is the one the budget exists for.
	if err := renderWithinDiskBudget(pc.PageCount, dpi); err != nil {
		return RenderRecord{}, err
	}

	renders := make([]PageRender, 0, pc.PageCount)
	for i := 0; i < pc.PageCount; i++ {
		b, rerr := renderPagePNG(inst, doc.Document, i, dpi)
		if rerr != nil {
			// NAMED AND FATAL. A page that silently did not render would leave the record one
			// image short of the document, and nothing downstream could tell that from a
			// document with fewer pages.
			return RenderRecord{}, fmt.Errorf("page %d of %d: %w", i+1, pc.PageCount, rerr)
		}
		if err := writeAtomic(PagePath(run, sha, i+1), b); err != nil {
			return RenderRecord{}, err
		}
		renders = append(renders, PageRender{Sha: Sha(b), DPI: dpi})
	}

	rec := RenderRecord{
		Sha:        sha,
		Renders:    renders,
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

// PageImagePath is one page's image drawn for a verification, OUTSIDE PagesDir. A render and a
// reading both clear PagesDir, and an image kept there would vanish under either and pose as a
// partial `ocr pages` render.
func PageImagePath(run record.Run, sha string, page int) string {
	return filepath.Join(Path(run, sha)+".page-images", fmt.Sprintf("p%04d.png", page))
}

// ErrPageOutOfRange is a page number the document does not have.
var ErrPageOutOfRange = errors.New("page out of range")

// RenderOnePage draws ONE page of a PDF at dpi, the same grayscale raster the reader reads, and
// writes it to PageImagePath, replacing an earlier image. page is 1-based. It returns the path,
// the image's sha256 and the document's page count.
func RenderOnePage(run record.Run, sha string, body []byte, page, dpi int) (string, string, int, error) {
	inst, closer, err := pdfiumInstance(Dir(run))
	if err != nil {
		return "", "", 0, err
	}
	defer closer()
	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &body})
	if err != nil {
		return "", "", 0, fmt.Errorf("pdf could not be opened: %w", err)
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return "", "", 0, fmt.Errorf("pdf page count unavailable: %w", err)
	}
	if page < 1 || page > pc.PageCount {
		return "", "", pc.PageCount, fmt.Errorf("%w: page %d of a %d-page document", ErrPageOutOfRange, page, pc.PageCount)
	}
	b, err := renderPagePNG(inst, doc.Document, page-1, dpi)
	if err != nil {
		return "", "", pc.PageCount, fmt.Errorf("page %d of %d: %w", page, pc.PageCount, err)
	}
	p := PageImagePath(run, sha, page)
	if err := writeReplacing(p, b); err != nil {
		return "", "", pc.PageCount, err
	}
	return p, Sha(b), pc.PageCount, nil
}

// renderPagePNG rasterises ONE page to grayscale PNG bytes and releases the bitmap before
// returning.
//
// THE RELEASE IS THE LOAD-BEARING LINE. go-pdfium under WebAssembly documents the contract on
// the response itself: "you MUST call Cleanup() when you are done with the image object to
// release resources." The first real-document run of this path never called it, so every
// page's raster (~15 MB at 200 DPI letter, plus the wasm-side allocation) stayed live until
// process exit — ~60 MB a page, and the 80-page IEEE 1012 scan was OOM-killed at 1.6 GB
// before its first model call (#671). The encode happens BEFORE the release because the
// pixel buffer is documented invalid after Cleanup, and toGray may return the pdfium-backed
// image unwrapped rather than a copy.
//
// index is 0-based, pdfium's own numbering; callers wrap errors with the 1-based page number
// a citation would use.
func renderPagePNG(inst pdfium.Pdfium, doc references.FPDF_DOCUMENT, index, dpi int) ([]byte, error) {
	rendered, err := inst.RenderPageInDPI(&requests.RenderPageInDPI{
		Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc, Index: index}},
		DPI:  dpi,
	})
	if err != nil {
		return nil, fmt.Errorf("did not render: %w", err)
	}
	defer rendered.Cleanup()
	if rendered.Result.RenderedImage == nil {
		return nil, fmt.Errorf("rendered to no image")
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, toGray(rendered.Result.RenderedImage)); err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}
	return buf.Bytes(), nil
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
