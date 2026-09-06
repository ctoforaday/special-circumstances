package fetchcache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klippa-app/go-pdfium/requests"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// THE AUTOMATIC PATH, AND WHY IT IS ONE CALL RATHER THAN TWO VERBS.
//
// `ocr pages` and `ocr read` are primitives: one rasterises, one reads. Composing them was
// left to the operator, which meant a seat that fetched a scanned standard got
// `text_extracted: false` and a reason, and then had to KNOW two more commands existed. An
// instruction telling it so was the remaining half of #644 — and an instruction is the
// weakest possible carrier for a step the tool can simply take.
//
// So fetch takes it — ONE PAGE AT A TIME, AND NOTHING IS KEPT BUT THE TEXT. The first
// composition rendered every page to disk and then read the directory back, which held every
// page's raster in memory at once (#671) and persisted images the automatic path has no
// further use for: the product of a fetch is the reading, and the reading record carries the
// hash of every image it was read from. So each page is rasterised, read, and released
// before the next is touched; no page PNG and no render record are written. An operator who
// wants the pixels themselves — to check a reading against a page — has `ocr pages`, whose
// whole product is the images.

// ScanReader turns a cached document that yielded no text layer into a reading.
//
// It is an interface for exactly the reason Extractor is: the command layer's tests must
// exercise fetch's PLUMBING — what the summary says, what is reused, what is refused —
// without compiling a 5 MB WebAssembly module per test and without needing the engine's C
// stack. The real implementation ships as the default (see DefaultExtractor for why a no-op
// default would be the dangerous arrangement).
type ScanReader interface {
	// ReadScanned renders and reads e's document, at the engine's own operative resolution
	// (tessocr.RenderDPI — the constants are per-DPI facts, so the resolution is the
	// engine's to name, not a parameter to vary). It returns an error rather than an empty
	// record when it could not: a reading nobody made and a reading that found nothing are
	// different facts, and the caller states which in the reason it prints.
	ReadScanned(ctx context.Context, run record.Run, e Entry) (ReadingRecord, error)
}

// DefaultScanReader is the process-wide ScanReader, swapped by tests and nothing else.
var DefaultScanReader ScanReader = RenderAndRead{}

// RenderAndRead is the real composition: PDFium rasterises, the local engine reads.
type RenderAndRead struct{}

// ReadScanned reads e's document one page at a time — rasterise, read, release — and returns
// the reading. It persists the reading and its record, never the page images.
//
// THE BUDGET IS CHECKED BEFORE ANYTHING IS RENDERED. The page count is a field the extractor
// measured, so a document over the render budget is refused before the first raster. Where
// the count is 0 the extractor could not measure it (an unreadable page tree), and the same
// budget fires again from the document's own page count, after opening it but before any
// page is rendered.
//
// A READING AT THIS RESOLUTION AND ENGINE IS NEVER DERIVED TWICE ON A FETCH. Not because it
// is expensive any more — the engine is local and deterministic — but because rendering is
// deterministic for a fixed renderer, so a stored reading at this DPI by this engine already
// describes exactly the bytes this call would produce, and the record's RenderShas still say
// which. A malformed record is an error, never an absence — treating it as absence would
// silently replace a record somebody may have cited from.
func (RenderAndRead) ReadScanned(ctx context.Context, run record.Run, e Entry) (ReadingRecord, error) {
	if e.ContentType != "application/pdf" {
		return ReadingRecord{}, fmt.Errorf("only application/pdf renders to pages, and sha %s is %s",
			e.Sha, ContentTypeOrUnknown(e.ContentType))
	}
	if err := renderWithinDiskBudget(e.Pages, tessocr.RenderDPI); err != nil {
		return ReadingRecord{}, err
	}

	prev, had, err := ReadReadingRecord(run, e.Sha)
	if err != nil {
		return ReadingRecord{}, err
	}
	if had && prev.DPI == tessocr.RenderDPI && prev.Engine == DefaultPageEngine.Identity() && len(prev.Pages) > 0 {
		return prev, nil
	}

	body, rerr := Read(run, e.Sha)
	if rerr != nil {
		return ReadingRecord{}, fmt.Errorf("the index names sha %s but its content file is unreadable: %w", e.Sha, rerr)
	}
	return renderAndReadPages(run, e.Sha, body)
}

// renderAndReadPages is the fused loop: one page rasterised, read, and released at a time.
//
// MEMORY IS BOUNDED BY ONE PAGE, BY CONSTRUCTION. The render-everything-then-read shape held
// every raster live at once and was OOM-killed 27 pages into its first real document (#671);
// here the raster is released (renderPagePNG) and the PNG bytes go out of scope before the
// next page is touched.
//
// STALENESS IS HANDLED PER PAGE BY THE RECEIPTS (#679's design, kept): a receipt is reused
// only when its render sha equals the sha of the page as rendered NOW and its engine matches
// this binary's, so a DPI change or an engine bump mismatches every receipt and re-derives
// every page. The document is content-addressed, so the page count cannot drift under a sha.
// What IS still removed up front: the assembled text (a crash mid-loop must leave an honest
// absence, not a stale document that reads as current) and any deliberate render's images
// and record (the reading supersedes them; a stale render.json describing images this path
// is about to orphan would be the record naming pixels nobody kept).
//
// The per-page texts and receipts are written as the loop runs and the record LAST, so a
// crash leaves rows with no record — which reads as "not read", and the re-read validates
// page by page instead of deriving again.
func renderAndReadPages(run record.Run, sha string, body []byte) (ReadingRecord, error) {
	// The resolution is the engine's own constant, not a parameter: the grid and
	// reconstruction thresholds are per-DPI facts, and a caller-supplied DPI here would be
	// an invitation to apply them to pixels they were not tuned for.
	const dpi = tessocr.RenderDPI

	dir := PagesDir(run, sha)
	if err := os.Remove(OCRTextPath(run, sha)); err != nil && !os.IsNotExist(err) {
		return ReadingRecord{}, fmt.Errorf("clearing the previous reading of %s: %w", sha, err)
	}
	if err := os.Remove(renderRecordPath(run, sha)); err != nil && !os.IsNotExist(err) {
		return ReadingRecord{}, fmt.Errorf("clearing the previous render record of %s: %w", sha, err)
	}
	if pngs, err := filepath.Glob(filepath.Join(dir, "p????.png")); err == nil {
		for _, p := range pngs {
			if err := os.Remove(p); err != nil {
				return ReadingRecord{}, fmt.Errorf("clearing a superseded render of %s: %w", sha, err)
			}
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ReadingRecord{}, err
	}

	inst, closer, err := pdfiumInstance(Dir(run))
	if err != nil {
		return ReadingRecord{}, err
	}
	defer closer()

	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &body})
	if err != nil {
		return ReadingRecord{}, fmt.Errorf("pdf could not be opened: %w", err)
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return ReadingRecord{}, fmt.Errorf("pdf page count unavailable: %w", err)
	}
	if pc.PageCount == 0 {
		return ReadingRecord{}, fmt.Errorf("the document reports zero pages, so there is nothing to read")
	}
	if err := renderWithinDiskBudget(pc.PageCount, dpi); err != nil {
		return ReadingRecord{}, err
	}

	out := ReadingRecord{Sha: sha, Engine: DefaultPageEngine.Identity(), ReadAt: time.Now().UTC(), DPI: dpi}
	var assembled strings.Builder
	for i := 0; i < pc.PageCount; i++ {
		png, rerr := renderPagePNG(inst, doc.Document, i, dpi)
		if rerr != nil {
			// NAMED AND FATAL, exactly as in RenderPages: a page silently skipped would leave the
			// assembled text a document with a hole nothing downstream could see.
			return ReadingRecord{}, fmt.Errorf("page %d of %d: %w", i+1, pc.PageCount, rerr)
		}
		out.RenderShas = append(out.RenderShas, Sha(png))

		// The shared per-page step: reuse a matching receipt, keep every error fatal.
		r, norm, rerr := readPageStep(run, sha, i+1, png, dpi)
		if rerr != nil {
			return ReadingRecord{}, fmt.Errorf("page %d: %w", i+1, rerr)
		}
		out.Pages = append(out.Pages, r.pageReading())

		assembled.WriteString(norm)
		assembled.WriteString("\n\n")
	}

	text := strings.TrimRight(assembled.String(), "\n")
	if err := writeReplacing(OCRTextPath(run, sha), []byte(text)); err != nil {
		return ReadingRecord{}, err
	}
	out.TextSha = Sha([]byte(text))

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return ReadingRecord{}, err
	}
	if err := writeReplacing(readingRecordPath(run, sha), append(b, '\n')); err != nil {
		return ReadingRecord{}, err
	}
	return out, nil
}

// SameRenders reports whether two lists name the same page images in the same order.
//
// It is a function rather than two inline loops because both callers — the automatic path
// here and `ocr read`'s reuse guard — are asking one question, and a reuse rule that differed
// between them by a subscript would re-derive on one path and not the other for reasons
// nobody could see.
func SameRenders(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ContentTypeOrUnknown keeps an absent Content-Type from rendering as an empty string in the
// middle of a sentence, where it reads as a missing word rather than a missing measurement.
//
// Exported because the `ocr` verbs refuse a non-PDF in the same words. Two copies of a
// sentence a seat reads when it is already confused is how two spellings of one refusal
// start.
func ContentTypeOrUnknown(ct string) string {
	if ct == "" {
		return "of no recorded content type"
	}
	return ct
}
