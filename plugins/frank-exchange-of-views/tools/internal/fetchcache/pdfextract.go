package fetchcache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"unicode"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/tetratelabs/wazero"
)

// PDFExtractor reads PDFs with PDFium compiled to WebAssembly, run under wazero.
//
// WHY THIS ENGINE, AND NOT THE FOUR THAT LOST. Measured 2026-08-29 against the four documents
// this repository's own reports cite, then against a 159-PDF corpus (PDFium's adversarial
// test suite plus real arXiv/ACL/government documents in Latin, Cyrillic and CJK):
//
//   - go2markdown v1.0.0, the library first written into #629, letter-splits its output
//     ("T h e m e t h o d o f t y p e s") and — worse — returns 3,248 characters of
//     "![Image N from page M]" AND NO ERROR on an 80-page scan with no text layer. It would
//     have recorded an unreadable document as successfully extracted.
//   - ledongthuc/pdf and its dslipak fork strip every word boundary:
//     "ActiveLearningLiteratureSurveyBurrSettles". Nothing there is citable or greppable.
//   - pdfcpu has no plain-text extraction API at all.
//   - MuPDF (go-fitz) extracts beautifully and is AGPL-3.0, against this repository's MIT and
//     all four "license": "MIT" plugin manifests. Its nocgo path is not an escape — it dlopens
//     libmupdf and PANICS when absent.
//   - poppler's pdftotext is excellent and is a MACHINE PREREQUISITE, plus a separate pdffonts
//     interrogation call, plus an unsandboxed subprocess parsing bytes off the open web.
//
// PDFium under wazero is MIT + Apache-2.0, needs no cgo and no system library, builds for all
// six pairs this repo ships with CGO_ENABLED=0 and the release job unchanged, and reports
// zero characters natively on a document with no text layer — so the emptiness test is a
// property of the document rather than a hopeful reading of an error.
//
// AND IT SANDBOXES THE PARSER. A malformed or hostile PDF cannot crash this process or reach
// disk, network or memory outside the runtime. For bytes fetched off the open web that is a
// property [[agent-guardrails]] should want, and neither cgo nor a subprocess offers it.
type PDFExtractor struct{}

// extractorIdentity is library plus semver, per #629 D3 — the key an audit re-runs an
// extraction against. It moves with the DEPENDENCY, not with the tool's release, because
// semver is what actually tracks a change in extraction behaviour. A binary hash would be
// finer and would invalidate every cached extraction on every build, including builds that
// never touched this file.
//
// IT IS READ OUT OF THE BUILD, NOT COPIED BESIDE IT. A hand-written "go-pdfium@v1.19.8"
// constant is two copies of one fact with nothing between them: bump go.mod, forget the
// constant, and every cached extraction is thereafter keyed to a version that did not produce
// it — silently, because a stale string is still a well-formed string. [[facts-are-fields]]
// says prefer GENERATING the derived carrier over guarding two hand-written ones, and the Go
// toolchain already stamps the real module graph into every binary it links.
//
// extractorFallbackVersion is the one case generation cannot cover: `go test` builds a test
// binary whose BuildInfo carries an EMPTY Deps list (measured — main is populated, deps=0), so
// under test there is nothing to read. The fallback is checked against go.mod on disk by
// TestExtractorFallbackMatchesGoMod, which is a real gate rather than a second copy, because
// it reads the module file itself.
const (
	extractorModule          = "github.com/klippa-app/go-pdfium"
	extractorFallbackVersion = "v1.19.8"
)

func extractorIdentity() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == extractorModule {
				return extractorModule + "@" + dep.Version
			}
		}
	}
	return extractorModule + "@" + extractorFallbackVersion
}

// wazeroCacheDir is where the compiled WebAssembly module is kept between invocations.
//
// THIS IS THE WHOLE PERFORMANCE STORY. Compiling PDFium's 5.2 MB module costs 3,968 ms; with
// the cache warm the same call takes 220 ms for a 5-page paper and 1,048 ms for a 67-page one.
// `fetch` is a short-lived process invoked once per source, so without the cache every single
// fetch would pay the four seconds. It lives under the run's cache directory because that is
// already the run's scratch space and is already excluded from the record.
const wazeroCacheDir = ".wazero"

// pdfiumInstance starts a PDFium instance sharing the run's compiled-module cache, and
// returns it with the closer that releases both it and its pool.
//
// SHARED BY EXTRACTION AND RENDERING, which is the reason it exists. Both open the same
// document through the same runtime, and a second hand-rolled copy of this setup would be
// free to drift on the one line that matters — RuntimeConfig, without which every call pays
// the 3,968 ms module compile rather than the 220 ms warm start. Two copies of a
// performance-critical wiring is exactly the shape that ends up half-applied.
//
// The error text is caller-ready: Extract folds it straight into a Reason a seat reads, so
// it names the stage rather than the symptom.
// moduleCacheDir redirects the compiled-module cache away from the run. Production leaves it
// empty and the cache stays under the run, which is the behaviour the constant above describes;
// UseSharedModuleCache is the only thing that sets it.
var moduleCacheDir string

// moduleCacheName is the shared cache's directory, and it is FIXED rather than unique because
// that is the cleanup design. See UseSharedModuleCache.
const moduleCacheName = "feov-pdfium-module-cache"

// UseSharedModuleCache points the compiled-module cache at ONE directory for the lifetime of a
// test binary, and returns the release that removes it. IT IS FOR TESTS, and it is exported
// because two packages need it — this one and internal/cli, whose `ocr` tests drive the real
// rasteriser through the CLI. The alternative was the same fixed-name-and-cleanup reasoning
// hand-written in two TestMains, free to drift on the half that is silent when it breaks.
//
// WHY A TEST BINARY NEEDS A DOOR PRODUCTION DOES NOT. The 3,968 ms compile is amortised per CACHE
// DIRECTORY, and that directory is derived from the run. Right for the product — a `fetch` process
// handles one run — and exactly wrong for a test binary, where every test builds its own run under
// its own t.TempDir() and so its own empty cache. Measured: 21 tests in this package and 6 in
// internal/cli each paid the full compile; 68s and ~15s respectively, and past the 10-minute
// default timeout under -race, where the instrumented compile runs about 10x. fetchcache is in
// cmd/feov-record's import graph, so that is CI's race leg.
//
// THE FIXED NAME IS THE CLEANUP. The obvious shape — MkdirTemp plus a removal before os.Exit —
// cleans up on every path except the one that matters: `go test` PANICS the process on a timeout,
// which is exactly how this package was failing, and a panic runs no deferred removal. Measured:
// that shape leaked a 5 MB module per crash, forever. A fixed name cannot accumulate — a crash
// leaves ONE directory, the next run uses it warm and removes it on the way out.
//
// SAFE TO REUSE AND SAFE TO DELETE, holding nothing but wazero's compilation of a module that
// ships in the binary — keyed by content and wazero version, reproducible at any time for the
// 3,968 ms it costs. Two concurrent test binaries sharing it can at worst make each other
// recompile; there is no state here to corrupt.
//
// The release returns an error so it drops straight into testbuild.Main's post-suite hooks.
func UseSharedModuleCache() (func() error, error) {
	dir := filepath.Join(os.TempDir(), moduleCacheName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("shared PDFium module cache unavailable: %w", err)
	}
	moduleCacheDir = dir
	return func() error { return os.RemoveAll(dir) }, nil
}

func pdfiumInstance(cacheDir string) (pdfium.Pdfium, func(), error) {
	if moduleCacheDir != "" {
		cacheDir = moduleCacheDir
	}
	cache, err := wazero.NewCompilationCacheWithDir(filepath.Join(cacheDir, wazeroCacheDir))
	if err != nil {
		return nil, nil, fmt.Errorf("wazero compilation cache unavailable: %w", err)
	}
	pool, err := webassembly.Init(webassembly.Config{
		MinIdle: 1, MaxIdle: 1, MaxTotal: 1,
		RuntimeConfig: wazero.NewRuntimeConfig().WithCompilationCache(cache),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("pdf runtime failed to start: %w", err)
	}
	inst, err := pool.GetInstance(0)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("pdf runtime unavailable: %w", err)
	}
	return inst, func() { inst.Close(); pool.Close() }, nil
}

func (PDFExtractor) Extract(cacheDir, mediaType string, body []byte) Extraction {
	if mediaType != "application/pdf" {
		// NOT A REFUSAL — a statement that nobody looked. Attempted stays false, so the record
		// says "not measured" rather than "no text found", which for an HTML page would be a
		// lie about a document that is entirely text.
		return Extraction{}
	}
	out := Extraction{Attempted: true, ExtractorID: extractorIdentity()}

	inst, closer, err := pdfiumInstance(cacheDir)
	if err != nil {
		out.Reason = err.Error()
		return out
	}
	defer closer()

	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &body})
	if err != nil {
		// MEASURED ACROSS 159 FILES: this is where encrypted, permission-locked and malformed
		// documents land, and PDFium names which — "invalid password", "incorrect format" —
		// rather than returning an empty string. pdftotext extracted real text from NONE of
		// those 24 files either, so nothing is lost by recording the refusal instead of it.
		out.Reason = fmt.Sprintf("pdf could not be opened: %v", err)
		return out
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	if meta, merr := inst.FPDF_GetMetaText(&requests.FPDF_GetMetaText{Document: doc.Document, Tag: "Title"}); merr == nil {
		out.Title = meta.Value
	}
	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		out.Reason = fmt.Sprintf("pdf page count unavailable: %v", err)
		return out
	}
	out.Pages = pc.PageCount

	var sb strings.Builder
	var firstPageErr error
	for i := 0; i < pc.PageCount; i++ {
		t, terr := inst.GetPageText(&requests.GetPageText{
			Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: i}},
		})
		if terr != nil {
			if firstPageErr == nil {
				firstPageErr = fmt.Errorf("page %d: %w", i+1, terr)
			}
			continue
		}
		sb.WriteString(t.Text)
	}
	out.Text = normalizeExtracted(sb.String())
	if out.Text == "" {
		// THE DOCUMENT ANSWERED, AND THE ANSWER WAS NONE. This is the IEEE 1012 case: a 1998
		// Acrobat PDFWriter scan, 80 pages, not one glyph of text layer — and 11 of the 33 PDF
		// reads in the 2026-08-23 programme. The reason is REQUIRED here, because a bare empty
		// string is exactly the plausible zero [[facts-are-fields]] exists to forbid.
		switch {
		case firstPageErr != nil:
			out.Reason = fmt.Sprintf("no text could be read (%v)", firstPageErr)
		case pc.PageCount == 0:
			out.Reason = "the document reports zero pages"
		default:
			out.Reason = fmt.Sprintf("no text layer: %d pages, all empty — this is a scanned or "+
				"image-only document and needs optical recognition, which this build does not do", pc.PageCount)
		}
	}
	return out
}

// normalizeExtracted strips the control characters PDFium emits alongside real text.
//
// MEASURED, NOT GUESSED AT: PDFium writes U+0002 at every hyphenation point — 337 of them in the
// 67-page Settles survey (0.22 % of its characters), 74 in the bandit paper, and up to 1.59 %
// in the worst real case. Left in, they land in the middle of words a seat then quotes into a
// report, and a `reproduce` comparing that quote against a re-extraction would still pass while
// a human reading it sees mojibake.
//
// Newlines and tabs survive; every other C0 and C1 control character does not. Nothing else is
// touched — no whitespace collapsing, no line rejoining — because the extraction must stay a
// faithful rendering of the document, not a prettied one.
func normalizeExtracted(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' || r == '\t':
			b.WriteRune(r)
		case r == '\r':
			// PDFium emits CRLF; the record keeps LF so a hash of the text is the same on every
			// platform that re-extracts it.
			continue
		case r < 0x20 || (r >= 0x7f && r <= 0x9f):
			continue
		default:
			b.WriteRune(r)
		}
	}
	// A document of nothing but control characters must come back as "", not as whitespace that
	// reads like content to a length check.
	if strings.TrimFunc(b.String(), unicode.IsSpace) == "" {
		return ""
	}
	return b.String()
}
