package fetchcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// Reading a rendered page is the second half of #644 — and since plans/local-ocr.md it is
// LOCAL: tesseract + leptonica statically linked into this binary, no model, no
// credentials, no network, no filter. What used to be the expensive, uncorroborated half
// of the pipeline is now the cheap, re-derivable one.
//
// THE READING RE-DERIVES, WHICH IS THE PROPERTY THE MODEL READER HAD TO DISCLAIM. A model
// re-reading the same page returned different bytes, so its record could offer only an
// attestation of provenance — which model, when, against which image hashes — and had to
// say plainly that nothing corroborated it. A deterministic engine restores #636's
// extractor model: the record keys the reading to the engine identity (tessocr.Identity():
// library pins, language data and a hash of the engine's own source), and the same engine
// over the same pixels returns the same bytes, so an audit can re-run the read and compare
// hashes instead of taking anyone's word. That is strictly stronger than the
// one-pass/corroboration argument this comment used to carry, and it is why the token
// fields, the call-count economics and the content-filter machinery are gone from this
// file rather than merely disused.

// PageEngine turns one page image into a reading — text for prose, a reconstructed table
// with stats where a ruled grid fires. It is an interface for the reason the extractor
// is a variable: the command layer's tests exercise the read loops' plumbing with a
// pure-Go fake, and only internal/tessocr's own suite needs the C stack.
//
// It returns an error when it could not read: an empty Text is a legitimate reading (a
// blank page) and must never be how a failure arrives. On a binary built without
// `-tags tessocr` every call fails with the engine-absent error, loudly.
type PageEngine interface {
	// ReadPage reads one page rendered at dpi. The resolution is a parameter because a scan is
	// read at ITS OWN resolution (#1031): the grid thresholds are derived per page through
	// tessocr.GridFor rather than assumed to be the 300-DPI tune, and the geometry below the
	// detector is fed measurements converted into the space its constants were fitted in (#1074).
	ReadPage(png []byte, dpi int) (tessocr.PageResult, error)
	// Identity keys the reading record — the #636 extractor model. Same identity, same
	// pixels, same bytes; a reading whose recorded identity matches this engine's is
	// re-derivable by it.
	Identity() string
}

// DefaultPageEngine is the process-wide engine. A test replaces it; production never does.
var DefaultPageEngine PageEngine = &TessocrPageEngine{}

// TessocrPageEngine is the production engine: internal/tessocr's statically linked
// tesseract + leptonica, initialized once and kept for the life of the process (the CLI
// is one verb per process, and a TessBaseAPI init per page would re-load the embedded
// traineddata eighty times per document). The mutex serializes ReadPage because the
// underlying TessBaseAPI carries per-recognition state.
type TessocrPageEngine struct {
	mu sync.Mutex
	en *tessocr.Engine
}

func (t *TessocrPageEngine) Identity() string { return tessocr.Identity() }

func (t *TessocrPageEngine) ReadPage(png []byte, dpi int) (tessocr.PageResult, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.en == nil {
		en, err := tessocr.New()
		if err != nil {
			return tessocr.PageResult{}, engineAbsentLoudly(err)
		}
		t.en = en
	}
	res, err := t.en.ReadPage(png, tessocr.GridFor(dpi))
	if err != nil {
		return tessocr.PageResult{}, engineAbsentLoudly(err)
	}
	return res, nil
}

// engineAbsentLoudly turns the stub build's refusal into the sentence a seat reads in
// ocr_reason. The wrap matters because that reason is the ONLY thing standing between
// "this binary cannot read scans" and a seat concluding the source itself is unreadable
// — and an empty reading in place of this error would be the plausible zero
// [[facts-are-fields]] names.
func engineAbsentLoudly(err error) error {
	if errors.Is(err, tessocr.ErrNotCompiledIn) {
		return fmt.Errorf("the OCR engine is not compiled into this binary; release binaries "+
			"carry it — build with `-tags tessocr` and the C stack from third_party/pins: %w", err)
	}
	return err
}

// Disk-denominated render budget. The old cap was denominated in MODEL CALLS ("a page is
// a model call") and left with the model; what remains bounded is the disk a render can
// fill by typo. The arithmetic: a letter-size grayscale scan page measures ~350 KB of PNG
// at 300 DPI, the floor a page renders at, so the 1 GB budget is ~3000 pages there — the 534-page Unicode CJK
// chart that motivated the old cap estimates to ~187 MB and now renders, which is the
// point: it was always a disk question wearing a billing cap. Cost scales with the square
// of the resolution, so the estimate does too.
const (
	approxPageBytesAt300DPI = 350 << 10
	MaxRenderBudgetBytes    = 1 << 30
)

// renderWithinDiskBudget refuses a document whose full render would exceed the budget,
// by name and before any page is rasterised.
func renderWithinDiskBudget(pages, dpi int) error {
	est := int64(pages) * approxPageBytesAt300DPI * int64(dpi) * int64(dpi) / (300 * 300)
	if est <= MaxRenderBudgetBytes {
		return nil
	}
	return fmt.Errorf("%d pages at %d DPI estimate to %d MB of page images (~%d KB a page at "+
		"300 DPI, scaled by resolution squared), over the %d MB render budget; nothing was "+
		"rendered — render a subset, or raise the budget deliberately: this refuses rather "+
		"than filling the disk and calling it the document",
		pages, dpi, est>>20, approxPageBytesAt300DPI>>10, MaxRenderBudgetBytes>>20)
}

// PageReading is what the engine produced for one page.
type PageReading struct {
	// Page is 1-based, matching PagePath and how a document is cited.
	Page int `json:"page"`
	// TextSha is the sha256 of this page's reading. The text itself is on disk beside
	// the image; the hash keeps the record small and makes a changed page visible.
	TextSha string `json:"text_sha"`
	// Length is the reading's length in runes — a cheap shape check on a page that came
	// back far shorter than its neighbours, which is what a truncation looks like.
	Length int `json:"length"`
	// DPI is the resolution THIS page was rendered and read at, and part of the re-derivation
	// key: the grid constants are derived from it (tessocr.GridFor). It sits on the page rather
	// than on the document because a scan is read at its own resolution (#1031), and a document
	// that mixes them has no single number that is true of every page.
	DPI int `json:"dpi"`
	// Table reports the grid detector fired on this page. Kept even when reconstruction
	// fell back: a grid that was seen and could not be rebuilt is a different fact from
	// no grid.
	Table bool `json:"table,omitempty"`
	// RotatedPage reports the page read better rotated 90° clockwise — a landscape table
	// on a portrait scan — and that the reading came from the rotated pixels.
	RotatedPage bool `json:"rotated_page,omitempty"`
	// GridIntersections is the detector's measured rule-crossing count AT THE PAGE'S OWN
	// RESOLUTION, the OCR-independent denominator: compared against
	// Reconstruction.ExpectedIntersections(), a lattice far larger than the
	// reconstruction accounts for means the OCR dropped grid content leptonica can still
	// see — and past tessocr.MaxIntersectionRatio the reconstruction is discarded, which
	// ReconstructionFallback then states.
	//
	// THE COMPARISON IS NOT MADE IN THESE UNITS, and the field says so rather than leaving a
	// reader to divide two numbers that do not divide. Crossings are PIXELS and scale with the
	// square of the resolution; ExpectedIntersections counts lattice POINTS and does not scale at
	// all, so the ratio is taken after converting the crossings to 300 DPI (#1074). This field is
	// what the detector measured; ReconstructionFallback prints what the gate compared.
	GridIntersections int `json:"grid_intersections,omitempty"`
	// Reconstruction is the grid branch's confidence, a FIELD rather than something
	// inferred from the emitted table's shape (plan §II). Nil on prose pages and on grid
	// pages whose TSV held no marks at all.
	Reconstruction *tessocr.Stats `json:"reconstruction,omitempty"`
	// ReconstructionFallback states why a fired grid's reconstruction was not used as the
	// page text — a bad reconstruction must never read as a good one, and this field is
	// where it says so.
	ReconstructionFallback string `json:"reconstruction_fallback,omitempty"`

	// TextCells is what the rules' own cells recovered on a page whose marks did not answer
	// (#932): the lattice's shape and how much of the page it explained. Absent when the mark
	// reconstruction stood, and absent when the cells were refused — TextCellFallback says why,
	// and one of the two is always present on a table page that is not a mark grid.
	TextCells *tessocr.CellStats `json:"text_cells,omitempty"`
	// TextCellFallback states why the rules' cells were not used as the page text either.
	TextCellFallback string `json:"text_cell_fallback,omitempty"`
}

// pageReceipt is one page's provenance row: what was rendered, what read it, and what
// came back. Receipts are written as the loop runs — the reading record is a PROJECTION
// of them, assembled last (#757) — so a crash mid-document leaves rows a re-run validates
// page by page. Their original resume purpose (protecting model spend) retired with the
// model; what keeps them is the render-sha validation, which is what guarantees a reading
// is never served for pixels it did not come from.
type pageReceipt struct {
	Page      int       `json:"page"`
	RenderSha string    `json:"render_sha"`
	DPI       int       `json:"dpi"`
	Engine    string    `json:"engine"`
	ReadAt    time.Time `json:"read_at"`
	TextSha   string    `json:"text_sha"`
	Length    int       `json:"length"`

	Table                  bool           `json:"table,omitempty"`
	RotatedPage            bool           `json:"rotated_page,omitempty"`
	GridIntersections      int            `json:"grid_intersections,omitempty"`
	Reconstruction         *tessocr.Stats `json:"reconstruction,omitempty"`
	ReconstructionFallback string         `json:"reconstruction_fallback,omitempty"`

	TextCells        *tessocr.CellStats `json:"text_cells,omitempty"`
	TextCellFallback string             `json:"text_cell_fallback,omitempty"`
}

// pageReading projects the receipt's page-level facts into the record's row — one
// projection, used by both read loops, so the two cannot select different fields.
func (r pageReceipt) pageReading() PageReading {
	return PageReading{
		Page: r.Page, TextSha: r.TextSha, Length: r.Length, DPI: r.DPI,
		Table: r.Table, RotatedPage: r.RotatedPage, GridIntersections: r.GridIntersections,
		Reconstruction: r.Reconstruction, ReconstructionFallback: r.ReconstructionFallback,
		TextCells: r.TextCells, TextCellFallback: r.TextCellFallback,
	}
}

// receiptPath is one page's receipt, beside its text.
func receiptPath(run record.Run, sha string, n int) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.receipt.json", n))
}

// readReceipt returns a page's receipt, and whether one exists. A malformed receipt is an
// ERROR, never an absence: treated as absence it would silently redo work over a row
// somebody may already have cited from — the same rule every record reader here follows.
func readReceipt(run record.Run, sha string, n int) (pageReceipt, bool, error) {
	b, err := os.ReadFile(receiptPath(run, sha, n))
	if os.IsNotExist(err) {
		return pageReceipt{}, false, nil
	}
	if err != nil {
		return pageReceipt{}, false, err
	}
	var r pageReceipt
	if err := json.Unmarshal(b, &r); err != nil {
		return pageReceipt{}, false, fmt.Errorf("receipt for page %d of %s is unreadable: %w — "+
			"`ocr read --force` discards receipts and reads again", n, sha, err)
	}
	return r, true, nil
}

// ClearReceipts removes every receipt, which is what `ocr read --force` means: discard
// the rows and derive everything again. Force is cheap now — seconds of local compute —
// but it keeps its meaning, because the receipts are also what a citation's provenance
// hangs on and discarding them must stay a deliberate act.
func ClearReceipts(run record.Run, sha string) error {
	matches, err := filepath.Glob(filepath.Join(PagesDir(run, sha), "p????.receipt.json"))
	if err != nil {
		return err
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			return err
		}
	}
	return nil
}

// readPageStep is the ONE per-page act both loops share — the deliberate `ocr read` over
// rendered images and fetch's fused rasterise-and-read (#679 required them identical, and
// a copy would let them drift on semantics that took a plan to settle).
//
// Given the page's PNG bytes as rendered NOW, it: reuses a receipt whose render sha and
// engine identity match — the reading re-derives, so a matching receipt IS the reading —
// after verifying the text on disk still matches the receipt's hash; otherwise runs the
// engine. Text first, receipt second, so a crash between the two leaves a page that
// re-reads rather than a receipt naming text that was never written.
func readPageStep(run record.Run, sha string, page int, png []byte, dpi int) (pageReceipt, string, error) {
	renderSha := Sha(png)
	if r, ok, err := readReceipt(run, sha, page); err != nil {
		return pageReceipt{}, "", err
	} else if ok && r.RenderSha == renderSha && r.Engine == DefaultPageEngine.Identity() {
		b, rerr := os.ReadFile(PageTextPath(run, sha, page))
		if rerr != nil || Sha(b) != r.TextSha {
			return pageReceipt{}, "", fmt.Errorf("page %d of %s has a receipt but its text is missing or "+
				"altered (%v) — the pair is corrupt; `ocr read --force` discards receipts and reads again",
				page, sha, rerr)
		}
		return r, string(b), nil
	}

	res, rerr := DefaultPageEngine.ReadPage(png, dpi)
	if rerr != nil {
		// AN ENGINE ERROR IS AN ERROR, NOT AN EMPTY PAGE — including the stub build's
		// engine-absent refusal, which must reach the operator as a sentence, never as a
		// blank reading.
		return pageReceipt{}, "", rerr
	}

	if err := writePageEvidence(run, sha, page, res); err != nil {
		return pageReceipt{}, "", err
	}
	norm := tessocr.NormalizeReading(res.Text)
	if err := writeReplacing(PageTextPath(run, sha, page), []byte(norm)); err != nil {
		return pageReceipt{}, "", err
	}
	r := pageReceipt{
		Page: page, RenderSha: renderSha, DPI: dpi, Engine: DefaultPageEngine.Identity(),
		ReadAt: time.Now().UTC(), TextSha: Sha([]byte(norm)), Length: len([]rune(norm)),
		Table: res.Table, RotatedPage: res.RotatedPage,
		Reconstruction: res.Reconstruction, ReconstructionFallback: res.Fallback,
		TextCells: res.TextCells, TextCellFallback: res.TextCellFallback,
	}
	if res.Table {
		r.GridIntersections = res.Grid.Intersections
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return pageReceipt{}, "", err
	}
	if err := writeReplacing(receiptPath(run, sha, page), append(b, '\n')); err != nil {
		return pageReceipt{}, "", err
	}
	return r, norm, nil
}

// DPIRange is the lowest and highest resolution any page of this reading was read at. Equal values
// are the ordinary case — a scan is usually uniform — and a spread is the document that mixes,
// which is the one a summary of a single number used to misreport.
func (r ReadingRecord) DPIRange() (lo, hi int) {
	for i, p := range r.Pages {
		if i == 0 || p.DPI < lo {
			lo = p.DPI
		}
		if p.DPI > hi {
			hi = p.DPI
		}
	}
	return lo, hi
}

// TablePages is how many pages of this reading carried a detected grid. Derived from the
// pages slice rather than stored — a count field beside it would be a second copy of the
// same fact, free to disagree.
func (r ReadingRecord) TablePages() int {
	n := 0
	for _, p := range r.Pages {
		if p.Table {
			n++
		}
	}
	return n
}

// TextCellPages is how many of those grid pages were rebuilt from the RULES' own cells — a
// ruled table of text, where the marks path has nothing to infer a grid from (#932). Derived
// for the same reason TablePages is.
func (r ReadingRecord) TextCellPages() int {
	n := 0
	for _, p := range r.Pages {
		if p.TextCells != nil {
			n++
		}
	}
	return n
}

// ReadingRecord is one document's reading, written beside its page images.
//
// IT IS ITS OWN RECORD FOR THE REASON RenderRecord IS: the cache index is append-only and
// Lookup takes the first match, so it cannot express an update. A reading is also a
// distinct later act with its own facts — which engine, when, against which rendered
// images.
//
// THE RECORD SUPPORTS RE-DERIVATION, which is #636's extractor model restored. Engine is
// a deterministic identity: the same binary over the pixels RenderShas names produces the
// same bytes, so `reproduce` can re-run the reading and compare hashes — the check the
// model reader's record had to state it could not offer.
type ReadingRecord struct {
	Sha string `json:"sha"`
	// Engine is the identity key of what did the reading — tessocr.Identity(), naming the
	// library pins, the language data and a hash of the engine package's source (or a
	// test fake's own name). Unlike a model name, it does not change underneath itself:
	// any change to what decides the bytes is a new identity.
	Engine string `json:"engine"`
	// ReadAt is when — provenance for a human, not part of the re-derivation key.
	ReadAt time.Time `json:"read_at"`
	// RenderShas binds this reading to the exact images it read. On the deliberate path
	// those images are on disk under `ocr pages`; on fetch's automatic path they were
	// released as they were read and these hashes are the only identity the pixels have.
	// A re-render at a new DPI makes the reading stale, and this is what lets a reader
	// notice rather than assume.
	RenderShas []string      `json:"render_shas"`
	Pages      []PageReading `json:"pages"`
	TextSha    string        `json:"text_sha"`
}

func readingRecordPath(run record.Run, sha string) string {
	return filepath.Join(PagesDir(run, sha), "reading.json")
}

// PageTextPath is one page's reading, beside the image it was read from.
//
// THE PER-PAGE FILES ARE KEPT EVEN THOUGH THE ASSEMBLED TEXT CONTAINS THEM. A citation
// lands on a page, and a reader checking one against the scan wants that page's text next
// to that page's image — not an offset into a document-length file they must count into.
func PageTextPath(run record.Run, sha string, page int) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.txt", page))
}

// evidenceKinds are the per-page evidence files kept beside a reading, by suffix, each
// with the engine output it holds. They are debugging artifacts, not the reading: no
// receipt, record or hash names them, and a page served from its receipt keeps the ones
// its original read wrote.
var evidenceKinds = []struct {
	suffix string
	of     func(tessocr.PageResult) string
}{
	{"engine.log", func(r tessocr.PageResult) string { return r.Diagnostics }},
	{"portrait.tsv", func(r tessocr.PageResult) string { return r.Evidence.PortraitTSV }},
	{"rotated.tsv", func(r tessocr.PageResult) string { return r.Evidence.RotatedTSV }},
	{"sparse.tsv", func(r tessocr.PageResult) string { return r.Evidence.SparseTSV }},
	{"band.txt", func(r tessocr.PageResult) string { return r.Evidence.HeaderBand }},
	{"refused.md", func(r tessocr.PageResult) string { return r.Evidence.RefusedTable }},
}

// EvidencePath is one page's evidence file of the given kind (a suffix in evidenceKinds).
func EvidencePath(run record.Run, sha string, page int, suffix string) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.%s", page, suffix))
}

// writePageEvidence replaces a page's evidence with what this read produced. Every kind is
// removed first, so a page re-read under a new engine never keeps evidence from the old one
// beside the new reading; an empty kind (no diagnostics, or a prose page's TSVs) leaves no
// file, and an absent file means the engine produced nothing of that kind.
func writePageEvidence(run record.Run, sha string, page int, res tessocr.PageResult) error {
	for _, k := range evidenceKinds {
		p := EvidencePath(run, sha, page, k.suffix)
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clearing page %d's old %s: %w", page, k.suffix, err)
		}
		if body := k.of(res); body != "" {
			if err := writeReplacing(p, []byte(body)); err != nil {
				return err
			}
		}
	}
	return nil
}

// OCRTextPath is the assembled reading, beside the extraction TextPath would hold.
// The suffix says which is which to a human listing the directory; no reader recovers a
// fact from it — `ocr_derived` on the record is what states the difference.
func OCRTextPath(run record.Run, sha string) string { return Path(run, sha) + ".ocr.txt" }

// ReadRenderedPages reads every rendered page through the engine and records what came
// back.
func ReadRenderedPages(run record.Run, sha string, rd RenderRecord) (ReadingRecord, error) {
	if rd.Pages() == 0 {
		return ReadingRecord{}, fmt.Errorf("the render record for %s names no pages", sha)
	}
	// THE RESOLUTION IS A BAND, AND IT IS CHECKED PER PAGE (#1031). The grid thresholds are derived
	// from each page's own DPI (tessocr.GridFor), so pixels at 350 are read with a tune for 350; the
	// geometry below the detector reaches its 300-DPI constants already converted (#1074).
	// What stays refused is a page outside the band the derivation was measured over: below the
	// floor a page is read small and content is lost, above the ceiling nothing is recovered and a
	// boundary page was measured flipping to a false table.
	for i, pr := range rd.Renders {
		if pr.DPI < DefaultRenderDPI || pr.DPI > MaxRenderDPI {
			return ReadingRecord{}, fmt.Errorf("page %d was rendered at %d DPI, outside the %d–%d "+
				"the engine's constants are derived over — re-render with `ocr pages --sha %s` and read again",
				i+1, pr.DPI, DefaultRenderDPI, MaxRenderDPI, sha)
		}
	}

	out := ReadingRecord{
		Sha: sha, Engine: DefaultPageEngine.Identity(), ReadAt: time.Now().UTC(),
		RenderShas: rd.Shas(),
	}
	var texts []string

	for i := 1; i <= rd.Pages(); i++ {
		png, err := os.ReadFile(PagePath(run, sha, i))
		if err != nil {
			return ReadingRecord{}, fmt.Errorf("page %d image: %w", i, err)
		}
		// The shared per-page step: reuse a matching receipt, keep every error fatal.
		r, norm, rerr := readPageStep(run, sha, i, png, rd.Renders[i-1].DPI)
		if rerr != nil {
			return ReadingRecord{}, fmt.Errorf("page %d: %w", i, rerr)
		}
		out.Pages = append(out.Pages, r.pageReading())
		texts = append(texts, norm)
	}

	text := tessocr.AssembleReading(texts)
	if err := writeReplacing(OCRTextPath(run, sha), []byte(text)); err != nil {
		return ReadingRecord{}, err
	}
	out.TextSha = Sha([]byte(text))

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return ReadingRecord{}, err
	}
	// The record is written LAST, so a crash leaves pages with no record — which reads as
	// "not read" and re-reads cleanly, the same ordering RenderPages relies on.
	if err := writeReplacing(readingRecordPath(run, sha), append(b, '\n')); err != nil {
		return ReadingRecord{}, err
	}
	return out, nil
}

// ReadReadingRecord returns a document's reading, and whether one exists. A malformed
// record is an error rather than an absence, for the reason ReadRenderRecord gives.
func ReadReadingRecord(run record.Run, sha string) (ReadingRecord, bool, error) {
	b, err := os.ReadFile(readingRecordPath(run, sha))
	if os.IsNotExist(err) {
		return ReadingRecord{}, false, nil
	}
	if err != nil {
		return ReadingRecord{}, false, err
	}
	var r ReadingRecord
	if err := json.Unmarshal(b, &r); err != nil {
		return ReadingRecord{}, false, fmt.Errorf("reading record for %s is unreadable: %w", sha, err)
	}
	if r.Sha != sha {
		return ReadingRecord{}, false, fmt.Errorf("reading record under %s names document %s", sha, r.Sha)
	}
	return r, true, nil
}

// writeReplacing writes b to dst, REPLACING whatever is there.
//
// Not writeAtomic, and the difference is the bug #658 shipped and caught: writeAtomic
// treats an existing destination as already done, which is right for a content-addressed
// path and wrong for a stable name whose content varies. A second reading of the same
// page is new content at the same path, exactly as a re-render at a new DPI was.
func writeReplacing(dst string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, dst); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}
