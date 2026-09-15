package fetchcache

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// TextOrigin is where a cached source's text came from, derived from the index entry and the
// reading on disk, first match wins. EMBEDDED needs positive evidence that the bytes carry text:
// a nil TextExtracted means nobody asked, and a PDF index line from before that field may be a
// scan. It returns the reading when the origin is OCR, because that reading is what a quote is
// located in.
//
// The index's own ocr_derived flag is not consulted: the automatic read updates the fetch
// summary, never the index line, so that flag was false on every OCR source.
func TextOrigin(run record.Run, e Entry) (recordpb.SourceTextOrigin, ReadingRecord, error) {
	if e.TextExtracted != nil && *e.TextExtracted {
		return recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED, ReadingRecord{}, nil
	}
	// A reading on disk makes the text OCR-derived even when a page file has since gone missing:
	// completeness is LocateSpan's question, asked only when a quote has to be found in it.
	rec, ok, err := ReadReadingRecord(run, e.Sha)
	if err != nil {
		return 0, ReadingRecord{}, err
	}
	if ok {
		return recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR, rec, nil
	}
	if e.NotRenderable != nil && *e.NotRenderable {
		return recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE, ReadingRecord{}, nil
	}
	if e.TextExtracted == nil && CarriesText(MediaType(e.ContentType)) {
		return recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED, ReadingRecord{}, nil
	}
	return recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE, ReadingRecord{}, nil
}

// ErrReadingIncomplete is a reading record whose page texts are not all on disk as recorded.
var ErrReadingIncomplete = errors.New("reading incomplete")

// readingComplete holds a reading to its record: one text per rendered page, each hashing to
// what the record says. A reading with a page missing would locate a quote in part of a
// document and report the rest as "not found".
func readingComplete(run record.Run, rec ReadingRecord) error {
	if len(rec.Pages) == 0 || len(rec.Pages) != len(rec.RenderShas) {
		return fmt.Errorf("%w: the reading of %s records %d pages against %d rendered",
			ErrReadingIncomplete, rec.Sha, len(rec.Pages), len(rec.RenderShas))
	}
	for _, p := range rec.Pages {
		b, err := os.ReadFile(PageTextPath(run, rec.Sha, p.Page))
		if err != nil || Sha(b) != p.TextSha {
			return fmt.Errorf("%w: the reading of %s is incomplete on disk (page %d's text is missing or altered)",
				ErrReadingIncomplete, rec.Sha, p.Page)
		}
	}
	return nil
}

// CarriesText reports whether a bare media type is text a reader reads as-is.
func CarriesText(mt string) bool {
	switch {
	case strings.HasPrefix(mt, "text/"),
		mt == "application/json", mt == "application/xml", mt == "application/xhtml+xml",
		strings.HasSuffix(mt, "+json"), strings.HasSuffix(mt, "+xml"):
		return true
	}
	return false
}

// A span that is not on any single page, typed so the caller's refusal can teach the one fix
// that applies. Only the tool can say which: the assembled reading joins pages with a blank
// line, which reads the same as a paragraph break.
var (
	ErrSpanStraddles = errors.New("span straddles a page boundary")
	ErrSpanHyphen    = errors.New("span found only across a line-end hyphen")
	ErrSpanAbsent    = errors.New("span not in the reading")
)

// SpanLocation is where a span sits in a reading.
type SpanLocation struct {
	Pages   []int  // every page holding the whole span, ascending
	Engine  string // the reading's engine identity
	TextSha string // the reading's assembled text sha
}

// LocateSpan finds every page of a complete reading whose text holds span.
//
// WHITESPACE-NORMALISED, GLYPH-EXACT. Line breaks are layout, so runs of whitespace compare as
// one space on both sides. Case, hyphens, ligatures and punctuation must match: a misread is
// exactly what an audit of OCR text has to see, and folding them would accept text the reading
// does not hold. There is no minimum length; whether a span is adequate is red's to judge from
// the recorded quote.
func LocateSpan(run record.Run, rec ReadingRecord, span string) (SpanLocation, error) {
	want := normalizeSpace(span)
	if want == "" {
		return SpanLocation{}, fmt.Errorf("%w: the span is empty", ErrSpanAbsent)
	}
	if err := readingComplete(run, rec); err != nil {
		return SpanLocation{}, err
	}
	texts := make([]string, len(rec.Pages))
	loc := SpanLocation{Engine: rec.Engine, TextSha: rec.TextSha}
	for i, p := range rec.Pages {
		b, err := os.ReadFile(PageTextPath(run, rec.Sha, p.Page))
		if err != nil {
			return SpanLocation{}, err
		}
		texts[i] = string(b)
		if strings.Contains(normalizeSpace(texts[i]), want) {
			loc.Pages = append(loc.Pages, p.Page)
		}
	}
	if len(loc.Pages) > 0 {
		return loc, nil
	}
	for i := 0; i+1 < len(texts); i++ {
		if joined := normalizeSpace(texts[i] + " " + texts[i+1]); strings.Contains(joined, want) {
			return SpanLocation{}, fmt.Errorf("%w: the span runs from page %d onto page %d; quote the part on one page",
				ErrSpanStraddles, rec.Pages[i].Page, rec.Pages[i+1].Page)
		}
	}
	for i, t := range texts {
		if frag, ok := hyphenOnlyMatch(t, want); ok {
			return SpanLocation{}, fmt.Errorf("%w: found on page %d only as %q across a line break; quote it as the reading has it",
				ErrSpanHyphen, rec.Pages[i].Page, frag)
		}
	}
	return SpanLocation{}, fmt.Errorf("%w: not in the reading at %s (compared with runs of whitespace as one space; case, hyphens and punctuation must match)",
		ErrSpanAbsent, OCRTextPath(run, rec.Sha))
}

func normalizeSpace(s string) string { return strings.Join(strings.Fields(s), " ") }

var lineEndHyphen = regexp.MustCompile(`(\S+)-[ \t]*\r?\n[ \t]*(\S+)`)

// hyphenOnlyMatch reports whether want is on the page once each line-end hyphen is joined, and
// the first hyphenated word as the reading holds it.
func hyphenOnlyMatch(page, want string) (string, bool) {
	joined := lineEndHyphen.ReplaceAllString(page, "$1$2")
	if joined == page || !strings.Contains(normalizeSpace(joined), want) {
		return "", false
	}
	for _, m := range lineEndHyphen.FindAllStringSubmatch(page, -1) {
		if strings.Contains(want, m[1]+m[2]) {
			return m[1] + "-", true
		}
	}
	return lineEndHyphen.FindStringSubmatch(page)[1] + "-", true
}
