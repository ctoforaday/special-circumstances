package fetchcache

import (
	"context"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE AUTOMATIC PATH, AND WHY IT IS ONE CALL RATHER THAN TWO VERBS.
//
// `ocr pages` and `ocr read` are primitives: one rasterises, one reads. Composing them was
// left to the operator, which meant a seat that fetched a scanned standard got
// `text_extracted: false` and a reason, and then had to KNOW two more commands existed. An
// instruction telling it so was the remaining half of #644 — and an instruction is the
// weakest possible carrier for a step the tool can simply take.
//
// So fetch takes it. This is that composition: render if nothing usable is rendered, read if
// this exact set of images has not been read, and hand back the reading. Both primitives keep
// their own verbs, because an operator re-reading a bad scan at 300 DPI still needs them
// separately.

// ScanReader turns a cached document that yielded no text layer into a reading.
//
// It is an interface for exactly the reason Extractor is: the command layer's tests must
// exercise fetch's PLUMBING — what the summary says, what is reused, what is refused —
// without compiling a 5 MB WebAssembly module per test and without spending a model. The
// real implementation ships as the default (see DefaultExtractor for why a no-op default
// would be the dangerous arrangement).
type ScanReader interface {
	// ReadScanned renders and reads e's document. It returns an error rather than an empty
	// record when it could not: a reading nobody made and a reading that found nothing are
	// different facts, and the caller states which in the reason it prints.
	ReadScanned(ctx context.Context, run record.Run, e Entry, model string, dpi int) (ReadingRecord, error)
}

// DefaultScanReader is the process-wide ScanReader, swapped by tests and nothing else.
var DefaultScanReader ScanReader = RenderAndRead{}

// RenderAndRead is the real composition: PDFium rasterises, a model reads.
type RenderAndRead struct{}

// ReadScanned renders e's pages if they are not already rendered at this resolution, then
// reads them if this exact set of images has not been read, and returns the reading.
//
// THE CAP IS CHECKED BEFORE ANYTHING IS RENDERED, which is the difference between this and
// calling the two verbs in sequence. ReadRenderedPages refuses a document over MaxReadPages —
// but by then a 500-page scan has already been rasterised to disk, and disk is a fixed
// per-session allowance this repository has already exhausted once. The page count is a field
// the extractor measured, so the refusal can happen before the first PNG.
//
// Where the count is 0 the extractor could not measure it (an unreadable page tree), and the
// cap in ReadRenderedPages remains the backstop. That is a render this refusal cannot avoid,
// and it is bounded by the same cap one step later.
func (RenderAndRead) ReadScanned(ctx context.Context, run record.Run, e Entry, model string, dpi int) (ReadingRecord, error) {
	if e.ContentType != "application/pdf" {
		return ReadingRecord{}, fmt.Errorf("only application/pdf renders to pages, and sha %s is %s",
			e.Sha, ContentTypeOrUnknown(e.ContentType))
	}
	if e.Pages > MaxReadPages {
		return ReadingRecord{}, fmt.Errorf("the document has %d pages, over the %d-page cap on one "+
			"automatic read (one model call a page); nothing was rendered. Read it deliberately "+
			"with `ocr pages --sha %s` then `ocr read --sha %s`",
			e.Pages, MaxReadPages, e.Sha, e.Sha)
	}

	rd, have, err := ReadRenderRecord(run, e.Sha)
	if err != nil {
		return ReadingRecord{}, err
	}
	if !have || rd.DPI != dpi || rd.Pages() == 0 {
		body, rerr := Read(run, e.Sha)
		if rerr != nil {
			return ReadingRecord{}, fmt.Errorf("the index names sha %s but its content file is unreadable: %w", e.Sha, rerr)
		}
		if rd, err = RenderPages(run, e.Sha, body, dpi); err != nil {
			return ReadingRecord{}, err
		}
	}

	// A READING OF THESE EXACT IMAGES IS NEVER PAID FOR TWICE. It is also not idempotent — a
	// second reading returns different bytes — so redoing it would replace a record a seat may
	// already have cited from.
	//
	// THE COMPARISON'S FALSE BRANCH IS CURRENTLY UNREACHABLE, and that is stated rather than
	// hidden: a mutation pass replacing SameRenders with `true` survived. RenderPages clears the
	// whole reading when it re-renders, so a stored reading whose image hashes DIFFER from the
	// current render cannot exist — `had` is already false in every case this would catch. It is
	// kept as the invariant's guard, not deleted, because the thing making it unreachable is a
	// deletion in another file: the day a re-render preserves a cited reading instead of wiping
	// it, this is the check that stops a 72 DPI transcription being served for 200 DPI pixels.
	// TestARerenderClearsTheReadingOfTheOldPixels is what pins the assumption.
	if prev, had, rerr := ReadReadingRecord(run, e.Sha); rerr == nil && had && SameRenders(prev.RenderShas, rd.PageShas) {
		return prev, nil
	}
	return ReadRenderedPages(ctx, run, e.Sha, model, rd)
}

// SameRenders reports whether two lists name the same page images in the same order.
//
// It is a function rather than two inline loops because both callers — the automatic path
// here and `ocr read`'s reuse guard — are asking one question, and a reuse rule that differed
// between them by a subscript would spend a model on one path and not the other for reasons
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
