package fetchcache

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

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

// A span the tool cannot place once in the source's text, typed so the caller's refusal can teach
// the one fix that applies. Only the tool can say which: an OCR reading joins its pages with a
// blank line, which reads the same as a paragraph break.
var (
	ErrSpanStraddles = errors.New("span straddles a page boundary")
	ErrSpanRemoved   = errors.New("span found only in a line the reading removed")
	ErrSpanRepeated  = errors.New("span found more than once")
	ErrSpanAbsent    = errors.New("span not in the source's text")
	ErrNoText        = errors.New("the source has no text")
)

// SpanLocation is where a span sits in a source's text, and the text around it.
type SpanLocation struct {
	// Pages is the PDF page of an OCR reading the span sits on; empty for embedded text, which
	// the extraction holds as one text with no page breaks.
	Pages []int
	// Engine and ReadingSha name the OCR reading the page was found in; empty for embedded text.
	Engine     string
	ReadingSha string
	// Start and End are byte offsets of the span in the text it was found in — the source's text,
	// or the page's reading — and TextSha is that text's sha256.
	Start, End int
	TextSha    string
	// Before, Matched and After are that text's bytes: up to SpanWindow bytes either side of the
	// span, and the span exactly as the source holds it.
	Before, Matched, After string
}

// SpanWindow is how many bytes of the source's text either side of a span a citation keeps. It
// is a sentence or two of an ordinary page — enough to read a definition or a qualifier the span
// cut off, small enough that every citation in a work list can carry it.
const SpanWindow = 300

// spanText is one text a span can be found in: the source's whole text, or one page's reading.
type spanText struct {
	page int // 0 for embedded text
	text string
	sha  string
}

// LocateSourceSpan finds span in the cached source's text, once.
//
// THE COMPARISON IS EXACT AFTER ONE DECLARED NORMALISATION, applied to both sides: a run of
// whitespace is one space, and a word broken by a hyphen at a line end is read either joined
// ("photo-\nsynthesis" as "photosynthesis") or with its hyphen kept ("well-\nknown" as
// "well-known") — a line break is layout, and which reading a hyphen takes is the author's, so
// the span may be typed either way. Case, ligatures, quotation marks and punctuation must match:
// a misread is exactly what an audit has to see, and folding them would accept text the source
// does not hold. There is no minimum length; a span the text holds twice is refused instead, so
// the text kept around it is the place blue quoted.
func LocateSourceSpan(run record.Run, e Entry, span string) (SpanLocation, error) {
	if strings.TrimSpace(span) == "" {
		return SpanLocation{}, fmt.Errorf("%w: the span is empty", ErrSpanAbsent)
	}
	origin, rec, err := TextOrigin(run, e)
	if err != nil {
		return SpanLocation{}, err
	}
	var texts []spanText
	loc := SpanLocation{}
	switch origin {
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
		if err := readingComplete(run, rec); err != nil {
			return SpanLocation{}, err
		}
		for _, p := range rec.Pages {
			b, err := os.ReadFile(PageTextPath(run, rec.Sha, p.Page))
			if err != nil {
				return SpanLocation{}, err
			}
			texts = append(texts, spanText{page: p.Page, text: string(b), sha: p.TextSha})
		}
		loc.Engine, loc.ReadingSha = rec.Engine, rec.TextSha
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED:
		t, err := embeddedText(run, e)
		if err != nil {
			return SpanLocation{}, err
		}
		texts = append(texts, t)
	default:
		return SpanLocation{}, fmt.Errorf("%w: %s", ErrNoText, recordpb.Word(origin))
	}

	for _, mode := range []hyphenMode{hyphenAsIs, hyphenJoined, hyphenKept} {
		want := fold(span, mode).s
		var hits []spanHit
		for _, t := range texts {
			f := fold(t.text, mode)
			for from := 0; ; {
				i := strings.Index(f.s[from:], want)
				if i < 0 {
					break
				}
				hits = append(hits, spanHit{t: t, start: f.at[from+i], end: f.at[from+i+len(want)-1] + 1})
				from += i + 1
			}
		}
		switch {
		case len(hits) > 1:
			return SpanLocation{}, fmt.Errorf("%w: the source's text holds it %d times%s; quote enough of the sentence that it is found once",
				ErrSpanRepeated, len(hits), pagesOf(hits))
		case len(hits) == 1:
			h := hits[0]
			loc.Start, loc.End, loc.TextSha = h.start, h.end, h.t.sha
			if h.t.page > 0 {
				loc.Pages = []int{h.t.page}
			}
			loc.Before, loc.Matched, loc.After = window(h.t.text, h.start, h.end)
			return loc, nil
		}
	}

	want := fold(span, hyphenAsIs).s
	for i := 0; i+1 < len(texts); i++ {
		if joined := fold(texts[i].text+" "+texts[i+1].text, hyphenAsIs).s; strings.Contains(joined, want) {
			return SpanLocation{}, fmt.Errorf("%w: the span runs from page %d onto page %d; quote the part on one page",
				ErrSpanStraddles, texts[i].page, texts[i+1].page)
		}
	}
	// A LINE THE ENGINE READ AND REMOVED IS NOT A MISQUOTE (#1030). Its text was produced by the
	// recognizer from a rule, a leader or an unread script, and the page record keeps it; a quote
	// of it is refused with that said, never with the sentence a misquote earns.
	for _, p := range rec.Pages {
		for _, ln := range p.RemovedLines {
			if strings.Contains(fold(ln, hyphenAsIs).s, want) {
				return SpanLocation{}, fmt.Errorf("%w: page %d's engine read %q and removed it — its letters sit in boxes far wider than print fills, which is a rule, a leader or a script the engine does not read, read as characters; it is not text the page carries, so quote what the reading holds",
					ErrSpanRemoved, p.Page, ln)
			}
		}
	}
	return SpanLocation{}, fmt.Errorf("%w: not in the text at %s (compared with runs of whitespace as one space and a line-end hyphen read either way; case, quotation marks and punctuation must match)",
		ErrSpanAbsent, SourceTextPath(run, e, origin))
}

// embeddedText is the text an embedded source's span is found in: the extraction stored beside a
// document that was extracted, and the bytes themselves for a type a reader reads as-is. The
// extraction is held to the hash its entry records, so offsets never name bytes nobody fetched.
func embeddedText(run record.Run, e Entry) (spanText, error) {
	if e.TextExtracted == nil {
		b, err := Read(run, e.Sha)
		if err != nil {
			return spanText{}, err
		}
		return spanText{text: string(b), sha: e.Sha}, nil
	}
	b, err := os.ReadFile(TextPath(run, e.Sha))
	if err != nil {
		return spanText{}, fmt.Errorf("%w: the source's text is missing from the cache (%v)", ErrNoText, err)
	}
	if e.TextSha != "" && Sha(b) != e.TextSha {
		return spanText{}, fmt.Errorf("the source's text at %s no longer hashes to what its fetch recorded", TextPath(run, e.Sha))
	}
	return spanText{text: string(b), sha: Sha(b)}, nil
}

type spanHit struct {
	t          spanText
	start, end int
}

func pagesOf(hits []spanHit) string {
	var pages []string
	for _, h := range hits {
		if h.t.page > 0 {
			pages = append(pages, fmt.Sprint(h.t.page))
		}
	}
	if len(pages) == 0 {
		return ""
	}
	return " (on pages " + strings.Join(pages, ", ") + ")"
}

// window is text's bytes from SpanWindow before [start, end) to SpanWindow after it, widened to
// whole characters, split at the span.
func window(text string, start, end int) (string, string, string) {
	a, z := max(0, start-SpanWindow), min(len(text), end+SpanWindow)
	for a > 0 && !utf8.RuneStart(text[a]) {
		a--
	}
	for z < len(text) && !utf8.RuneStart(text[z]) {
		z++
	}
	return text[a:start], text[start:end], text[end:z]
}

// hyphenMode is how fold reads a word broken by a hyphen at a line end.
type hyphenMode int

const (
	hyphenAsIs   hyphenMode = iota // the hyphen stays and the line break is a space
	hyphenJoined                   // the hyphen and the line break go: the word is whole
	hyphenKept                     // the line break goes and the hyphen stays: a compound
)

// folded is a text in fold's normalisation, and the source byte each of its bytes came from.
type folded struct {
	s  string
	at []int
}

// fold normalises text for LocateSourceSpan: each run of whitespace becomes one space, leading
// and trailing whitespace goes, and a line-end hyphen — a hyphen after a non-space, then spaces
// or tabs, a line break, spaces or tabs, and a non-space — is read as mode says. at maps each
// byte of the result to the byte of text it came from, so a match is placed in the source's own
// bytes.
func fold(text string, mode hyphenMode) folded {
	var b strings.Builder
	at := make([]int, 0, len(text))
	emit := func(s string, from int) {
		b.WriteString(s)
		for k := range len(s) {
			at = append(at, from+k)
		}
	}
	space := false // a separator is owed before the next non-space
	for i := 0; i < len(text); {
		r, n := utf8.DecodeRuneInString(text[i:])
		if unicode.IsSpace(r) {
			space = b.Len() > 0
			i += n
			continue
		}
		if space {
			emit(" ", i-1)
			space = false
		}
		if r == '-' && mode != hyphenAsIs && i > 0 {
			if next, ok := lineEndHyphen(text, i); ok {
				if mode == hyphenKept {
					emit("-", i)
				}
				i = next
				continue
			}
		}
		emit(text[i:i+n], i)
		i += n
	}
	return folded{s: b.String(), at: at}
}

// lineEndHyphen reports whether the hyphen at text[i] ends a line inside a word, and where the
// word resumes.
func lineEndHyphen(text string, i int) (int, bool) {
	if prev, _ := utf8.DecodeLastRuneInString(text[:i]); unicode.IsSpace(prev) {
		return 0, false
	}
	j := i + 1
	for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
		j++
	}
	if j < len(text) && text[j] == '\r' {
		j++
	}
	if j >= len(text) || text[j] != '\n' {
		return 0, false
	}
	j++
	for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
		j++
	}
	if j >= len(text) {
		return 0, false
	}
	if r, _ := utf8.DecodeRuneInString(text[j:]); unicode.IsSpace(r) {
		return 0, false
	}
	return j, true
}

// SourceTextPath is the file a seat reads a cached source's text in, for the origin TextOrigin
// gives: the OCR reading, the extraction beside a document, or the bytes of a type read as-is.
// "" where the tool holds no text.
func SourceTextPath(run record.Run, e Entry, origin recordpb.SourceTextOrigin) string {
	switch origin {
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR:
		return OCRTextPath(run, e.Sha)
	case recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED:
		if e.TextExtracted == nil {
			return Path(run, e.Sha)
		}
		return TextPath(run, e.Sha)
	}
	return ""
}
