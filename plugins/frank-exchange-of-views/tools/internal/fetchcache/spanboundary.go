package fetchcache

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrSpanMidClause is a span the source holds once that does not start where a sentence or a
// clause of the source starts, or does not end where one ends.
var ErrSpanMidClause = errors.New("span cuts into a clause")

// SpanBoundary is the boundary rule in a seat's words: the one statement of it that cite's --help
// and the mid-clause refusal both render, beside SpanNormalisation.
const SpanBoundary = "a span starts and ends at a boundary; a clause mark is a comma, semicolon or colon followed by whitespace, or a dash (a hyphen counts when spaced or doubled). " +
	"A span STARTS at the start of the text (or of an OCR page); with an opening bracket or quote; just after a clause mark or an opening bracket or quote; " +
	"or just after one of these when its first character is not a lowercase letter: a sentence end (. ! ? or …, closers and reference marks such as [12] allowed, then whitespace), a blank line, or a markup tag. " +
	"It ENDS at the end of the text (or of an OCR page), or before a blank line or a markup tag; with a clause mark or a closing bracket or quote; just before a clause mark, a closing bracket or quote, or a spaced opening bracket ( or [; " +
	"or on or just before a sentence end: . ! ? or … (closers and reference marks allowed), then whitespace and a character that is not a lowercase letter. " +
	"In «" + spanExample + "» the spans «" + spanExampleWhole + "» and «" + spanExampleClause + "» are accepted; «" + spanExampleCutStart + "» is refused (it starts mid-clause), as is «" + spanExampleCutEnd + "» (it ends mid-clause). " +
	"The tool never widens a span: quote more"

// The worked example SpanBoundary shows, each span admitted or refused as it says
// (TestSpanBoundarysExampleIsWhatTheToolDoes).
const (
	spanExample         = "Fermat's test is fast, but it is not proof."
	spanExampleWhole    = "Fermat's test is fast"
	spanExampleClause   = "but it is not proof."
	spanExampleCutStart = "test is fast"
	spanExampleCutEnd   = "but it is"
)

// THE SOURCE'S SENTENCES ARE READ BY A SCANNER OVER PLAIN TEXT, not by anchor.Sentences. That
// splitter reads markdown blocks first, and a source's text is not markdown: an extraction's
// indented line would read as a code block and a page's "1." line as a list item. A terminator
// here is . ! ? or …, then its closers and any reference marks ("[127]"), then whitespace.
//
// A SENTENCE STARTS AT A CHARACTER THAT IS NOT A LOWERCASE LETTER. A blank line in an extraction
// is a block boundary, and a display formula breaks a sentence into blocks: "Any a such that",
// the formula, "when n is composite is known as a Fermat liar." A fragment that starts after the
// formula drops its subject and reads the other way, and its lowercase first letter is what says
// the sentence began earlier.
//
// A SPAN ENDS WHERE THE SOURCE ENDS A SENTENCE OR A CLAUSE, the start rule mirrored. A span
// cut before the clause ends can drop what the sentence goes on to say about its subject — m31's
// "by each integer from 2 up to the square root of" stops before its object, and a span stopping
// before a trailing formula ("…of a given integer" before "n") is refused on the same terms,
// whatever the formula adds. A spaced opening bracket ends a span: what it opens restates or
// qualifies, as "p divides b (that is, p ∣ ab ⟹ …)" does. An unspaced one is inside an expression,
// as in "f(x)".
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

// spanAtBoundaries refuses a span at [start, end) that does not start or does not end at a
// boundary, showing the source's sentence it sits in and SpanBoundary.
func spanAtBoundaries(text string, start, end int) error {
	from, to := spanStartsAtBoundary(text, start), spanEndsAtBoundary(text, end)
	var cut, quote string
	switch {
	case from && to:
		return nil
	case !from && !to:
		cut, quote = "starts and ends inside one", "that sentence, or a clause of it, whole"
	case !from:
		cut, quote = "starts inside one", "from that sentence's start or from a clause boundary inside it"
	default:
		cut, quote = "ends inside one", "to that sentence's end or to a clause boundary inside it"
	}
	return fmt.Errorf("%w — it %s: the source's sentence it sits in reads «%s»; quote %s (%s)",
		ErrSpanMidClause, cut, sentenceAround(text, start, end), quote, SpanBoundary)
}

// spanEndsAtBoundary reports whether a span whose last byte is text[end-1] ends a sentence or a
// clause of text, under SpanBoundary.
func spanEndsAtBoundary(text string, end int) bool {
	k, blank := nextNonSpace(text, end)
	if k < 0 || blank || opensTag(text, k) {
		return true
	}
	lastAt := end - utf8.RuneLen(runeBefore(text, end))
	if closesTerminatorRun(text, lastAt) && endsSentence(text, end) {
		return true
	}
	if closesUnit(text, lastAt, end) || clauseMarkEnds(text, lastAt, k) {
		return true
	}
	if z, ok := terminatorRunFrom(text, k); ok && endsSentence(text, z) {
		return true
	}
	return clauseMarkStarts(text, k) || closesUnit(text, k, k+utf8.RuneLen(runeAt(text, k))) ||
		(k > end && (text[k] == '(' || text[k] == '['))
}

// endsSentence reports whether the terminator run that stops at z ends a sentence: after any
// reference marks, whitespace and a character that is not a lowercase letter follow, or a blank
// line, a markup tag or the text's end does. "e.g. two" is an abbreviation, not a sentence end.
func endsSentence(text string, z int) bool {
	z = pastReferences(text, z)
	j, blank := nextNonSpace(text, z)
	return j < 0 || blank || opensTag(text, j) || (j > z && notLower(text, j))
}

// refMarkMax is the longest reference mark read as one, brackets included: [127], [a], [n 3] is not.
const refMarkMax = 10

// pastReferences is where the reference marks starting at z stop — "[127]", a bracketed run of
// letters and digits, as many as follow one another. An encyclopedia sets them after the
// terminator, and a sentence quoted to its end is not cut by its footnote.
func pastReferences(text string, z int) int {
	for z < len(text) && text[z] == '[' {
		k := z + 1
		for k < len(text) && k-z < refMarkMax && isAlnum(text[k]) {
			k++
		}
		if k == z+1 || k >= len(text) || text[k] != ']' {
			break
		}
		z = k + 1
	}
	return z
}

// beforeReferences is the last character before the reference marks that end at j, or j when the
// character at j closes none.
func beforeReferences(text string, j int) int {
	for j > 0 && text[j] == ']' {
		k := j - 1
		for k >= 0 && j-k < refMarkMax && isAlnum(text[k]) {
			k--
		}
		if k == j-1 || k <= 0 || text[k] != '[' {
			break
		}
		j = k - 1
	}
	return j
}

func isAlnum(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

// clauseMarkStarts reports whether the character at k, the first after a span, is a clause mark:
// a comma, semicolon or colon followed by whitespace or the text's end, a typographic dash, or a
// keyboard hyphen standing as a dash — spaced, or doubled.
func clauseMarkStarts(text string, k int) bool {
	r, n := utf8.DecodeRuneInString(text[k:])
	after := k+n >= len(text) || unicode.IsSpace(runeAt(text, k+n))
	switch r {
	case ',', ';', ':':
		return after
	case '‒', '–', '—', '―': // ‒ – — ―
		return true
	case '-':
		spacedBefore := k > 0 && unicode.IsSpace(runeBefore(text, k))
		doubled := k+1 < len(text) && text[k+1] == '-'
		return (after && spacedBefore) || doubled
	}
	return false
}

// terminatorRunFrom reports whether a terminator run — . ! ? or …, then any closers — starts at k,
// and where it stops.
func terminatorRunFrom(text string, k int) (int, bool) {
	z := k
	for z < len(text) && isTerminator(runeAt(text, z)) {
		z += utf8.RuneLen(runeAt(text, z))
	}
	if z == k {
		return 0, false
	}
	for z < len(text) && isCloser(runeAt(text, z)) {
		z += utf8.RuneLen(runeAt(text, z))
	}
	return z, true
}

// closesUnit reports whether the character at i, ending at j, is a closing bracket or quote: ) ] }
// ” ’ », or a straight quote with no letter or digit after it.
func closesUnit(text string, i, j int) bool {
	switch r := runeAt(text, i); r {
	case ')', ']', '}', '”', '’', '»':
		return true
	case '"', '\'':
		if j >= len(text) {
			return true
		}
		n := runeAt(text, j)
		return !unicode.IsLetter(n) && !unicode.IsDigit(n)
	}
	return false
}

// nextNonSpace is the index of the first non-space character at or after i, -1 for none, and
// whether the whitespace before it holds a blank line — two line breaks.
func nextNonSpace(text string, i int) (int, bool) {
	blank := blankAhead(text, i)
	for i < len(text) {
		r, n := utf8.DecodeRuneInString(text[i:])
		if !unicode.IsSpace(r) {
			return i, blank
		}
		i += n
	}
	return -1, blank
}

// opensTag reports whether the < at k opens a markup tag — a letter, / ! or ? after it, and a >
// later on its line with no < between. A tag is a block boundary in a source read as-is.
func opensTag(text string, k int) bool {
	if text[k] != '<' || k+1 >= len(text) {
		return false
	}
	if !unicode.IsLetter(runeAt(text, k+1)) && strings.IndexByte("/!?", text[k+1]) < 0 {
		return false
	}
	for j := k + 1; j < len(text) && text[j] != '\n' && text[j] != '<'; j++ {
		if text[j] == '>' {
			return true
		}
	}
	return false
}

// sentenceAround is the source's sentence a span at [start, end) sits in, its whitespace folded
// to single spaces, an ellipsis marking where sentenceBounds stopped looking.
func sentenceAround(text string, start, end int) string {
	a, z, cutA, cutZ := sentenceBounds(text, start, end)
	s := strings.Join(strings.Fields(text[a:z]), " ")
	if cutA {
		s = "…" + s
	}
	if cutZ {
		s += "…"
	}
	return s
}

// sentenceBounds is where the source's sentence a span at [start, end) sits in starts and ends:
// from the nearest sentence start at or before start to the first sentence end at or after end,
// each looked for no further than spanReach bytes away. cutA and cutZ say the search stopped
// there rather than at the sentence's edge.
func sentenceBounds(text string, start, end int) (a, z int, cutA, cutZ bool) {
	a = start
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
	z = end
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
	return a, z, cutA, cutZ
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
	return z > 0 && !unicode.IsSpace(runeBefore(text, z)) && closesTerminatorRun(text, beforeReferences(text, z-utf8.RuneLen(runeBefore(text, z))))
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
// . ! ? or …, then any closers and reference marks — with whitespace before i.
func terminatorEnds(text string, j, i int) bool {
	return j+1 < i && closesTerminatorRun(text, beforeReferences(text, j))
}

// closesTerminatorRun reports whether the character at j is a terminator, or a closer after one.
func closesTerminatorRun(text string, j int) bool {
	for j > 0 && isCloser(runeAt(text, j)) {
		j -= utf8.RuneLen(runeBefore(text, j))
	}
	return isTerminator(runeAt(text, j))
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

// isTerminator is . ! ? or an ellipsis, which SpanNormalisation reads as the ... a seat types.
func isTerminator(r rune) bool { return strings.ContainsRune(".!?…", r) }

// isCloser is a closing bracket or quote, straight quotes included.
func isCloser(r rune) bool { return strings.ContainsRune(")]}”’»\"'", r) }

// notLower reports whether the character at i is not a lowercase letter.
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
