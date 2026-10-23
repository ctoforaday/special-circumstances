package fetchcache

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrSpanMidClause is a span the source holds once that does not start where a sentence or a
// clause of the source starts.
var ErrSpanMidClause = errors.New("span starts mid-clause")

// SpanBoundary is the boundary rule in a seat's words: the one statement of it that cite's --help
// and the mid-clause refusal both render, beside SpanNormalisation.
const SpanBoundary = "a span starts at the start of the source's text (or of an OCR page), where the source starts a sentence — after . ! or ?, a blank line or a markup tag, at a character that is not a lowercase letter — " +
	"or just after a clause mark: a comma, semicolon or colon, a dash, or an opening bracket or quote; a span may also open with that bracket or quote"

// THE SOURCE'S SENTENCES ARE READ BY A SCANNER OVER PLAIN TEXT, not by anchor.Sentences. That
// splitter reads markdown blocks first, and a source's text is not markdown: an extraction's
// indented line would read as a code block and a page's "1." line as a list item. What the two
// share is the judgement a terminator makes — . ! ? and their closers, then whitespace.
//
// A SENTENCE STARTS AT A CHARACTER THAT IS NOT A LOWERCASE LETTER. A blank line in an extraction
// is a block boundary, and a display formula breaks a sentence into blocks: "Any a such that",
// the formula, "when n is composite is known as a Fermat liar." A fragment that starts after the
// formula drops its subject and reads the other way, and its lowercase first letter is what says
// the sentence began earlier.
//
// THE START OF A TEXT IS A START WHATEVER ITS CASE. An OCR page that opens mid-sentence continues
// the page before, and a span may not run across pages (ErrSpanStraddles): refusing the page's
// opening words would leave the part of that sentence on it unquotable.

// spanStartsAtBoundary reports whether a span whose first byte is text[start] starts a sentence
// or a clause of text, under SpanBoundary.
func spanStartsAtBoundary(text string, start int) bool {
	if opensUnit(text, start) {
		return true
	}
	j, blank := prevNonSpace(text, start)
	if j < 0 {
		return true
	}
	if clauseMarkEnds(text, j, start) || opensUnit(text, j) {
		return true
	}
	return (blank || terminatorEnds(text, j, start) || closesTag(text, j)) && notLower(text, start)
}

// sentenceAround is the source's sentence a span at [start, end) sits in, its whitespace folded
// to single spaces: from the nearest sentence start at or before start to the first sentence end
// at or after end, each looked for no further than spanReach bytes away, an ellipsis marking
// where the search stopped.
func sentenceAround(text string, start, end int) string {
	a, cutA := start, false
	for a > 0 {
		if start-a > spanReach {
			cutA = true
			break
		}
		if !unicode.IsSpace(runeAt(text, a)) && sentenceStarts(text, a) {
			break
		}
		_, n := utf8.DecodeLastRuneInString(text[:a])
		a -= n
	}
	z, cutZ := end, false
	for z < len(text) {
		if z-end > spanReach {
			cutZ = true
			break
		}
		if sentenceEndsAt(text, z) {
			break
		}
		_, n := utf8.DecodeRuneInString(text[z:])
		z += n
	}
	s := strings.Join(strings.Fields(text[a:z]), " ")
	if cutA {
		s = "…" + s
	}
	if cutZ {
		s += "…"
	}
	return s
}

// spanReach is how far sentenceAround looks for a sentence's start or end: a text with no
// punctuation for that long is a table or a list, and the refusal shows a window of it.
const spanReach = 2 * SpanWindow

// sentenceStarts reports whether a sentence starts at the non-space character text[i].
func sentenceStarts(text string, i int) bool {
	j, blank := prevNonSpace(text, i)
	return (j < 0 || blank || terminatorEnds(text, j, i) || closesTag(text, j)) && notLower(text, i)
}

// sentenceEndsAt reports whether a sentence ends just before text[z]: a terminator run and its
// closers end there and whitespace or the text's end follows, or a blank line does.
func sentenceEndsAt(text string, z int) bool {
	if z >= len(text) {
		return true
	}
	if !unicode.IsSpace(runeAt(text, z)) {
		return false
	}
	if blankAhead(text, z) {
		return true
	}
	return z > 0 && !unicode.IsSpace(runeBefore(text, z)) && closesTerminatorRun(text, z-utf8.RuneLen(runeBefore(text, z)))
}

// prevNonSpace is the index of the last non-space character before i, -1 for none, and whether
// the whitespace between holds a blank line — two line breaks.
func prevNonSpace(text string, i int) (int, bool) {
	breaks := 0
	for i > 0 {
		r, n := utf8.DecodeLastRuneInString(text[:i])
		if !unicode.IsSpace(r) {
			return i - n, breaks >= 2
		}
		if r == '\n' {
			breaks++
		}
		i -= n
	}
	return -1, breaks >= 2
}

// clauseMarkEnds reports whether the character at j, the last before a span at start, is a clause
// mark: a comma, semicolon or colon followed by whitespace, a typographic dash, or a keyboard
// hyphen standing as a dash — spaced, or doubled.
func clauseMarkEnds(text string, j, start int) bool {
	switch r := runeAt(text, j); r {
	case ',', ';', ':':
		return j+1 < start
	case '‒', '–', '—', '―': // ‒ – — ―
		return true
	case '-':
		spacedAfter := j+1 < start
		spacedBefore := j > 0 && unicode.IsSpace(runeBefore(text, j))
		doubled := j > 0 && text[j-1] == '-'
		return spacedAfter && (spacedBefore || doubled)
	}
	return false
}

// terminatorEnds reports whether the character at j, the last before i, ends a terminator run —
// . ! or ?, then any closers — with whitespace before i.
func terminatorEnds(text string, j, i int) bool { return j+1 < i && closesTerminatorRun(text, j) }

// closesTerminatorRun reports whether the character at j is a terminator, or a closer after one.
func closesTerminatorRun(text string, j int) bool {
	for j > 0 && isCloser(runeAt(text, j)) {
		j -= utf8.RuneLen(runeBefore(text, j))
	}
	return strings.ContainsRune(".!?", runeAt(text, j))
}

// closesTag reports whether the > at j closes a markup tag — a < then a letter, / ! or ? earlier on
// its line, with no > between. A source read as-is (XML, a feed) holds its text between tags, and
// a tag is a block boundary there, as a blank line is in an extraction.
func closesTag(text string, j int) bool {
	if text[j] != '>' {
		return false
	}
	for k := j - 1; k >= 0 && text[k] != '\n' && text[k] != '>'; k-- {
		if text[k] == '<' {
			return k+1 < j && (unicode.IsLetter(runeAt(text, k+1)) || strings.IndexByte("/!?", text[k+1]) >= 0)
		}
	}
	return false
}

// opensUnit reports whether text[i] is an opening bracket or quote: ( [ { “ ‘ «, or a straight
// quote with no letter or digit before it.
func opensUnit(text string, i int) bool {
	switch r := runeAt(text, i); r {
	case '(', '[', '{', '“', '‘', '«':
		return true
	case '"', '\'':
		if i == 0 {
			return true
		}
		p := runeBefore(text, i)
		return !unicode.IsLetter(p) && !unicode.IsDigit(p)
	}
	return false
}

// isCloser is a closing bracket or quote, straight quotes included.
func isCloser(r rune) bool { return strings.ContainsRune(")]}”’»\"'", r) }

// notLower reports whether the first letter-bearing character at i is not a lowercase letter.
func notLower(text string, i int) bool { return !unicode.IsLower(runeAt(text, i)) }

func runeAt(text string, i int) rune {
	r, _ := utf8.DecodeRuneInString(text[i:])
	return r
}

func runeBefore(text string, i int) rune {
	r, _ := utf8.DecodeLastRuneInString(text[:i])
	return r
}

// blankAhead reports whether the whitespace run from i holds a blank line — two line breaks.
func blankAhead(text string, i int) bool {
	breaks := 0
	for i < len(text) {
		r, n := utf8.DecodeRuneInString(text[i:])
		if !unicode.IsSpace(r) {
			break
		}
		if r == '\n' {
			breaks++
		}
		i += n
	}
	return breaks >= 2
}
