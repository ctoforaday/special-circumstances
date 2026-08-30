package cli

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// fetchSummary is what `fetch` prints instead of the document (#629 D1, D2).
//
// THE BODY USED TO GO STRAIGHT TO STDOUT, AND FOR A PDF THAT WAS BINARY. `--json` was worse:
// it put invalid UTF-8 in a JSON string field. But the repair is not "print the PDF's text
// instead" — the Settles survey is 67 pages and IEEE 1012 is 80, and pasting either into a
// seat's context is the same waste whether it is legible or not. So NOTHING is returned inline,
// for ANY content type, and what comes back is a set of paths plus the facts a seat needs to
// decide what to open.
//
// EVERY FACT IS A FIELD, WHICH IS THE POINT. The alternative considered — hand back a bare path
// — would have left the seat to infer the content type from the bytes and to assume text
// existed. Here the absence of text is a field with a reason beside it, so the honest zero and
// the unlooked-at case cannot be confused (see [[facts-are-fields]]).
type fetchSummary struct {
	URL         string `json:"url"`
	Sha256      string `json:"sha256"`
	Path        string `json:"path"`
	ContentType string `json:"content_type,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Bytes       int    `json:"bytes"`
	CacheHit    bool   `json:"cache_hit"`
	Pages       int    `json:"pages,omitempty"`

	// TextExtracted is a pointer for the same three-state reason the Entry field is: nil means
	// nothing was attempted (this content type is not one anything extracts), false means it was
	// attempted and there was none. Rendering both as `false` would tell a seat that an HTML
	// page has no text.
	TextExtracted *bool  `json:"text_extracted,omitempty"`
	TextPath      string `json:"text_path,omitempty"`
	TextSha256    string `json:"text_sha256,omitempty"`
	TextReason    string `json:"text_reason,omitempty"`
	Extractor     string `json:"extractor,omitempty"`
	OCRDerived    bool   `json:"ocr_derived,omitempty"`

	// The automatic read's own facts (#644). They appear ONLY where the document was a PDF
	// with no text layer — the one case fetch spends a model on — and they are separate
	// fields from the extraction's because they answer a different question. TextReason says
	// why the DOCUMENT had no text layer; OCRReason says why there is no reading of its
	// pixels either, and a seat that saw only the first would conclude the source is
	// unreadable when in fact automatic reading was simply switched off.
	OCRReason string `json:"ocr_reason,omitempty"`
	Model     string `json:"model,omitempty"`
	DPI       int    `json:"dpi,omitempty"`
	// Divergences names the pages whose two readings disagreed. A COUNT ALONE WOULD BE
	// USELESS: the point of the two-pass check is to send a human to a specific page.
	Divergences []int `json:"divergent_pages,omitempty"`
	InTokens    int64 `json:"input_tokens,omitempty"`
	OutTokens   int64 `json:"output_tokens,omitempty"`
}

// applyReading folds a model's reading of the page images into the summary.
//
// IT OVERWRITES TextExtracted, and that is the point of the whole change: the extraction
// said false with a reason, the read then succeeded, and what the seat must be told is that
// the text IS there — with `ocr_derived: true` beside it so nobody mistakes a transcription
// for an author's text layer. The path is the reading's own file, never TextPath: two
// producers writing one name is how a re-render at another resolution silently replaces the
// text a citation was taken from.
func (s *fetchSummary) applyReading(run record.Run, r fetchcache.ReadingRecord) {
	yes := true
	s.TextExtracted = &yes
	s.TextReason = ""
	s.TextPath = fetchcache.OCRTextPath(run, r.Sha)
	s.TextSha256 = r.TextSha
	s.OCRDerived = true
	// THE EXTRACTOR DID NOT PRODUCE THIS TEXT, so its id must not be left beside it. PDFium
	// opened the document, counted 80 pages and found no glyphs; the transcription came from a
	// model. Leaving `extractor: pdfium@v1.19.8` on a reading would name a producer an audit
	// could re-run to get different bytes — and `model` is the word `ocr read` already uses for
	// this fact, so it is the word used here rather than a second one meaning the same thing.
	s.Extractor = ""
	s.Model = r.Model
	s.DPI = r.DPI
	s.Divergences = r.Divergences()
	s.InTokens, s.OutTokens = r.InTokens, r.OutTok
}

// summarize projects a cache entry into what the seat is shown. Paths are absolute — a Run
// resolves its directory once, at construction — because the seat's next act is to Read one.
func summarize(run record.Run, e fetchcache.Entry, bodyLen int, hit bool) fetchSummary {
	s := fetchSummary{
		URL:           e.URL,
		Sha256:        e.Sha,
		Path:          fetchcache.Path(run, e.Sha),
		ContentType:   e.ContentType,
		Filename:      e.Filename,
		Bytes:         bodyLen,
		CacheHit:      hit,
		Pages:         e.Pages,
		TextExtracted: e.TextExtracted,
		TextSha256:    e.TextSha,
		TextReason:    e.TextReason,
		Extractor:     e.Extractor,
		OCRDerived:    e.OCRDerived,
	}
	// THE PATH IS NAMED ONLY WHEN THE FILE IS THERE. A text_path pointing at a file that was
	// never written is worse than no field at all: a seat would Read it, get a not-found, and
	// have to guess whether the tool or the document was at fault.
	if e.TextSha != "" {
		s.TextPath = fetchcache.TextPath(run, e.Sha)
	}
	return s
}

// applicableToOCR reports whether this is the one case fetch spends a model on: a PDF the
// extractor looked at and found no text layer in.
//
// THE THREE-STATE POINTER IS LOAD-BEARING HERE. nil means nothing looked — an HTML page, an
// index line written by an older binary — and rendering either to pixels would be absurd. A
// plain bool would have collapsed that into the same false the scanned standard produces,
// and fetch would rasterise every web page it read.
func applicableToOCR(e fetchcache.Entry) bool {
	return e.ContentType == "application/pdf" && e.TextExtracted != nil && !*e.TextExtracted
}

// render is the human-facing form: the same fields, one per line, in the order a seat needs
// them — what it is, then where to read it, then whether the text is there.
//
// It is deliberately NOT a paraphrase of the JSON. Every line is `key: value`, so the two
// renderings carry identical facts and neither has to be parsed back out of prose.
func (s fetchSummary) render() string {
	var b strings.Builder
	line := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	line("url", s.URL)
	line("sha256", s.Sha256)
	line("path", s.Path)
	line("content_type", s.ContentType)
	line("filename", s.Filename)
	line("bytes", fmt.Sprint(s.Bytes))
	line("cache_hit", fmt.Sprint(s.CacheHit))
	if s.Pages > 0 {
		line("pages", fmt.Sprint(s.Pages))
	}
	switch {
	case s.TextExtracted == nil:
		// Nothing attempted. Say so rather than printing a text_extracted line a seat would read
		// as a measured "no".
	case *s.TextExtracted:
		line("text_extracted", "true")
		line("text_path", s.TextPath)
		line("text_sha256", s.TextSha256)
		line("extractor", s.Extractor)
		if s.OCRDerived {
			line("ocr_derived", "true")
		}
	default:
		line("text_extracted", "false")
		line("text_reason", s.TextReason)
		line("extractor", s.Extractor)
	}
	// THE REASON IS PRINTED WHETHER OR NOT THERE IS TEXT. Where a reading succeeded it is
	// empty and this line does not appear; where it did not — switched off, over the page cap,
	// credentials refused — it is the only thing standing between the seat and the conclusion
	// that the source itself is unreadable.
	line("ocr_reason", s.OCRReason)
	if s.OCRDerived {
		line("model", s.Model)
		line("dpi", fmt.Sprint(s.DPI))
		line("input_tokens", fmt.Sprint(s.InTokens))
		line("output_tokens", fmt.Sprint(s.OutTokens))
	}
	if len(s.Divergences) > 0 {
		line("divergent_pages", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(s.Divergences)), ","), "[]"))
		line("divergence_note", "the two readings disagree on these pages; the text marks each in place "+
			"and both passes are kept beside the page image. Nothing picked a winner.")
	}
	return b.String()
}
