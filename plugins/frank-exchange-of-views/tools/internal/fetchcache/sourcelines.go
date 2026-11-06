package fetchcache

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// TWO LOCATORS CHOOSE A CITATION'S EVIDENCE, AS PEERS (ruled by gblock 2026-11-06): the words,
// quoted (LocateSourceSpan), or lines, by number (LocateSourceLines). The seat picks; the tool
// infers nothing about what kind of source it is. A row of a data file — m39's OEIS b-file, where
// "66 91" sits on its own line with no sentence punctuation anywhere — is chosen by its line, and
// prose is chosen by its words.
//
// A LOCATOR ONLY CHOOSES. Each resolves to the same SpanLocation: the bytes it names, where they
// sit, and the text around them. From the write on, the citation is that text; nothing reads the
// locator again, so a line number that would mean another line after the source is read again
// cannot move a citation.
//
// LINES ARE COUNTED IN ONE TEXT, NAMED BY ITS HASH. A source's texts are its extraction, or one
// reading per OCR page, and a page's reading can be replaced within a run (`ocr read --force`, or a
// fetch whose page renders differ from the receipts', pageread.go readPageStep). A line number is
// only a pointer into the text it was read from, so the seat names that text by the hash `show
// source` printed beside the number, and a hash no text of the source has any longer is refused:
// the numbers were read from a text that is not there to resolve them. The hash also says which
// page, so a range never runs across pages.

// ErrTextGone is a line locator naming a text the source does not hold, or no longer holds.
var ErrTextGone = errors.New("the source holds no text with that hash")

// ErrLinesOutside is a line range that is malformed, or that runs past the text's last line.
var ErrLinesOutside = errors.New("the lines are not in the text")

// ParseLineRange reads a line locator: N, or N-M with N ≤ M, numbered from 1.
func ParseLineRange(s string) (first, last int, err error) {
	a, z, ranged := strings.Cut(strings.TrimSpace(s), "-")
	first, ferr := strconv.Atoi(strings.TrimSpace(a))
	last = first
	var lerr error
	if ranged {
		last, lerr = strconv.Atoi(strings.TrimSpace(z))
	}
	if ferr != nil || lerr != nil || first < 1 || last < first {
		return 0, 0, fmt.Errorf("%w: %q is not a line or a range of lines — give N or N-M, numbered from 1 as `show source` prints them", ErrLinesOutside, s)
	}
	return first, last, nil
}

// LocateSourceLines resolves lines first through last of the source's text whose sha256 is
// textSha to the words they hold: from the first line's first non-blank character to the last
// line's last, the line breaks between them kept as the source holds them.
func LocateSourceLines(run record.Run, e Entry, textSha string, first, last int) (SpanLocation, error) {
	texts, origin, rec, err := sourceTexts(run, e)
	if err != nil {
		return SpanLocation{}, err
	}
	for _, t := range texts {
		if t.sha != textSha {
			continue
		}
		h, err := linesOf(t, first, last)
		if err != nil {
			return SpanLocation{}, err
		}
		return placedAt(origin, rec, h), nil
	}
	return SpanLocation{}, fmt.Errorf("%w %q, so the line numbers were read from a text that is not there to count them in — search the source again and cite the lines its hits now print, with the text_sha beside them",
		ErrTextGone, textSha)
}

// linesOf is lines first through last of t, trimmed of the blanks at either end.
func linesOf(t spanText, first, last int) (spanHit, error) {
	starts := lineStarts(t.text)
	if last > len(starts) {
		return spanHit{}, fmt.Errorf("%w: the text has %d lines, and the range ends at line %d", ErrLinesOutside, len(starts), last)
	}
	a, z := starts[first-1], len(t.text)
	if last < len(starts) {
		z = starts[last] - 1 // the line break that ends the last line
	}
	for a < z && isBlank(t.text[a]) {
		a++
	}
	for z > a && isBlank(t.text[z-1]) {
		z--
	}
	if a == z {
		return spanHit{}, fmt.Errorf("%w: lines %d-%d of the text are blank", ErrLinesOutside, first, last)
	}
	return spanHit{t: t, start: a, end: z}, nil
}

// lineStarts is the byte offset each line of text starts at. A line break ends a line, so a text
// ending in one has no empty line after it.
func lineStarts(text string) []int {
	if text == "" {
		return nil
	}
	starts := []int{0}
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' && i+1 < len(text) {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// lineAt is the line number of the byte at i, counted from 1.
func lineAt(text string, i int) int { return strings.Count(text[:i], "\n") + 1 }

func isBlank(b byte) bool { return b == ' ' || b == '\t' || b == '\r' || b == '\n' }

// placedAt is the SpanLocation of a resolved placement — the one shape every locator resolves to.
func placedAt(origin recordpb.SourceTextOrigin, rec ReadingRecord, h spanHit) SpanLocation {
	loc := SpanLocation{Start: h.start, End: h.end, TextSha: h.t.sha}
	if origin == recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR {
		loc.Engine, loc.ReadingSha = rec.Engine, rec.TextSha
	}
	if h.t.page > 0 {
		loc.Pages = []int{h.t.page}
	}
	loc.Before, loc.Matched, loc.After = window(h.t.text, h.start, h.end)
	return loc
}
