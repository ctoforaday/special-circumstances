package cli

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
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
}

// summarize projects a cache entry into what the seat is shown. Paths are absolute-as-given —
// whatever runDir the seat is working in — because the seat's next act is to Read one of them.
func summarize(runDir string, e fetchcache.Entry, bodyLen int, hit bool) fetchSummary {
	s := fetchSummary{
		URL:           e.URL,
		Sha256:        e.Sha,
		Path:          fetchcache.Path(runDir, e.Sha),
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
		s.TextPath = fetchcache.TextPath(runDir, e.Sha)
	}
	return s
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
	return b.String()
}
