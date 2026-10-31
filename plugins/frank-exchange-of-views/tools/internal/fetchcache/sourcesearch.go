package fetchcache

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// SEARCHING A SOURCE IS A READ OF THE TEXT A CITATION IS FOUND IN, so it reads that text the way
// LocateSourceSpan does: the same texts (each page of an OCR reading, or the embedded text), in the
// same normalisation (whitespace runs as one space, typographic characters as their keyboard
// forms). A seat that found a passage with its shell's grep searched the raw bytes, line by line,
// and then had to work out what `cite --source-span` would accept; m36's seats wrote keyword-in-
// context patterns (`.{0,80}word.{0,80}`) that the seats' grep refuses outright, nine calls in seven
// seats. A hit here carries the source's sentence it sits in and whether cite takes that sentence,
// answered by the locator cite runs, so the read and the act cannot disagree.
//
// IT GATES NOTHING. It writes no event and refuses no act; a hit is information, and the seat still
// chooses the words it cites.

// EVERY HIT IS PRINTED, as every selector prints every match (ruled by gblock 2026-10-31): a list cut
// at N reads as a narrow pattern that found N. A broad pattern says so by its count, which is the
// seat's signal to narrow it.

// ErrPatternMatchesEmpty is a pattern that matches the empty string: it would match at every
// position of the text, which is no question about it.
var ErrPatternMatchesEmpty = errors.New("the pattern matches the empty string")

// SourceSearch is a search of one cached source's text.
type SourceSearch struct {
	// Sha is the cached document's sha256, and URL the address it was fetched from.
	Sha string `json:"sha256"`
	URL string `json:"url"`
	// Origin is where the text searched came from — embedded, or OCR-derived — and TextPath the file
	// that holds it.
	Origin   string `json:"source_text_origin"`
	TextPath string `json:"text_path"`
	// Hits are every match, in text order.
	Hits []SourceHit `json:"hits"`
	// Selection says what was asked and how many matches the text holds.
	Selection SourceSelection `json:"selection"`
}

// SourceSelection is what a search asked and how many matches the text holds.
type SourceSelection struct {
	Criterion string `json:"criterion"`
	Matched   int    `json:"matched"`
}

// SourceHit is one place a search matched.
type SourceHit struct {
	// Page is the PDF page of an OCR reading the match sits on; absent for embedded text.
	Page int `json:"page,omitempty"`
	// Start and End are the match's byte offsets in the text it was found in — the source's text, or
	// the page's reading — whose sha256 is TextSha: the coordinates a citation records its span in.
	Start   int    `json:"start"`
	End     int    `json:"end"`
	TextSha string `json:"text_sha"`
	// Matched is the source's bytes the pattern matched.
	Matched string `json:"matched"`
	// Sentence is the source's sentence the match sits in, its whitespace folded, between
	// SentenceStart and SentenceEnd in the same text; an ellipsis marks an edge the search stopped at
	// without finding the sentence's.
	Sentence      string `json:"sentence"`
	SentenceStart int    `json:"sentence_start"`
	SentenceEnd   int    `json:"sentence_end"`
	// SourceSpan is Sentence when `cite --source-span` takes it as typed — the source holds it once and
	// it starts and ends at a boundary — and empty when it does not, with NotCiteable saying why.
	SourceSpan  string `json:"source_span"`
	NotCiteable string `json:"not_citeable,omitempty"`
}

// SearchSource finds every match of a pattern in a cached source's text, case-insensitively. match
// is an RE2 regex; quote, when given instead, is literal text and is read in the source's
// normalisation, so a ’ typed from the page matches the ’ there. Exactly one of the two is given;
// an empty one, or a pattern that matches the empty string, is refused.
func SearchSource(run record.Run, e Entry, match, quote string) (SourceSearch, error) {
	re, criterion, err := searchPattern(match, quote)
	if err != nil {
		return SourceSearch{}, err
	}
	texts, origin, _, err := sourceTexts(run, e)
	if err != nil {
		return SourceSearch{}, err
	}
	out := SourceSearch{Sha: e.Sha, URL: e.URL, Origin: recordpb.Word(origin), TextPath: SourceTextPath(run, e, origin), Hits: []SourceHit{}}
	for _, t := range texts {
		f := fold(t.text, hyphenAsIs)
		for _, m := range re.FindAllStringIndex(f.s, -1) {
			out.Hits = append(out.Hits, sourceHit(texts, t, f.at[m[0]], f.end[m[1]-1]))
		}
	}
	out.Selection = SourceSelection{Criterion: criterion, Matched: len(out.Hits)}
	return out, nil
}

// sourceHit is the hit at [start, end) of t, with the sentence it sits in and whether cite takes it.
func sourceHit(texts []spanText, t spanText, start, end int) SourceHit {
	a, z, cutA, cutZ := sentenceBounds(t.text, start, end)
	// The bounds stop at the text's own edges, which can be whitespace; the sentence is its words.
	for a < z && unicode.IsSpace(runeAt(t.text, a)) {
		a += utf8.RuneLen(runeAt(t.text, a))
	}
	for z > a && unicode.IsSpace(runeBefore(t.text, z)) {
		z -= utf8.RuneLen(runeBefore(t.text, z))
	}
	h := SourceHit{Page: t.page, Start: start, End: end, TextSha: t.sha, Matched: t.text[start:end],
		Sentence: sentenceAround(t.text, start, end), SentenceStart: a, SentenceEnd: z}
	if cutA || cutZ {
		h.NotCiteable = fmt.Sprintf("the source marks no sentence edge within %d bytes of the match", spanReach)
		return h
	}
	sentence := strings.Join(strings.Fields(t.text[a:z]), " ")
	// The sentence is this text's own words at [a, z), so the locator finds it here: once, and then
	// it is cited at these offsets, or more than once. Found once, it can still end where cite reads
	// no sentence end — "e.g." before a lowercase word.
	_, _, err := locateOnce(texts, sentence)
	switch {
	case errors.Is(err, ErrSpanRepeated):
		h.NotCiteable = "the source holds this sentence more than once"
	case err != nil:
		h.NotCiteable = "the sentence does not start and end at a boundary cite takes"
	default:
		h.SourceSpan = sentence
	}
	return h
}

// searchPattern compiles a search's pattern, case-insensitive, and names it.
func searchPattern(match, quote string) (*regexp.Regexp, string, error) {
	var re *regexp.Regexp
	var criterion string
	switch {
	case match != "" && quote != "":
		return nil, "", errors.New("a search takes a regex or a literal, not both")
	case quote != "":
		want := fold(quote, hyphenAsIs).s
		if want == "" {
			return nil, "", fmt.Errorf("%w: the quote is blank", ErrPatternMatchesEmpty)
		}
		re, criterion = regexp.MustCompile("(?i)"+regexp.QuoteMeta(want)), `the quote "`+quote+`"`
	case match != "":
		var err error
		if re, err = regexp.Compile("(?i)" + match); err != nil {
			return nil, "", err
		}
		criterion = "/" + match + "/"
	default:
		return nil, "", fmt.Errorf("%w: no pattern was given", ErrPatternMatchesEmpty)
	}
	if re.MatchString("") {
		return nil, "", ErrPatternMatchesEmpty
	}
	return re, criterion, nil
}
