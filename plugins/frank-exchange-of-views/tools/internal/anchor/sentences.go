package anchor

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// BlockKind is the markdown block a stretch of the report is.
type BlockKind int

const (
	Paragraph BlockKind = iota
	ListItem
	Heading
	TableRow
	Fence
	FootnoteDef
)

// Block is one markdown block: [Start, End) from its first line's start to its last line's end,
// newline excluded; Body where its content starts, past its ">" markers and its list marker; and
// its sentences.
type Block struct {
	Kind             BlockKind
	Start, End, Body int
	Sentences        [][2]int
	depth            int // block-quote depth
}

var (
	fenceRe    = regexp.MustCompile("^\\s*(```|~~~)")
	headingRe  = regexp.MustCompile(`^\s*#{1,6}(?:[ \t]|$)`)
	footnoteRe = regexp.MustCompile(`^\s*\[\^[^\]]+\]:`)
	listRe     = regexp.MustCompile(`^\s*(?:[-*+]|(\d+)([.)]))(?:[ \t]+|$)`)
	delimRe    = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(?:\|\s*:?-+:?\s*)*\|?\s*$`)
	quoteRe    = regexp.MustCompile("^(?: {0,3}>[ \t]?)*") // a line's block-quote markers
)

// Blocks splits text into markdown blocks by markdown's own rules (CommonMark, and GFM for
// tables), because a newline ends a sentence only where it ends a block. A block ends at a blank
// line (a ">"-only line inside a quote) or where the next line opens one: a fence, a heading (one
// to six "#" then a space or the line's end), a footnote definition, a list item (only "1." or
// "1)" interrupts a paragraph; any number opens after a blank line or as the next item of the
// ordered list already open), a table row (a "|" line over a delimiter row, then every line until a
// blank or another opener), or a deeper block quote. Openers are read after a line's ">" markers;
// a shallower line continues a paragraph lazily. A heading and a table row are one line; a fence
// runs opener to closer and is one sentence.
func Blocks(text string) []Block {
	type line struct {
		lo, hi, depth int
		c             string // the line after its ">" markers
	}
	var lines []line
	masked := []byte(text) // ">" markers read as whitespace, offsets kept
	lo := 0
	for _, ln := range strings.Split(text, "\n") {
		q := quoteRe.FindString(ln)
		for i := lo; i < lo+len(q); i++ {
			masked[i] = ' '
		}
		lines = append(lines, line{lo, lo + len(ln), strings.Count(q, ">"), ln[len(q):]})
		lo += len(ln) + 1
	}
	var out []Block
	open, inFence, inTable := false, false, false
	olist := "" // the delimiter of the ordered list open
	end := func(e int) {
		if open {
			out[len(out)-1].End = e
			open = false
		}
	}
	for i, l := range lines {
		c := l.c
		if inFence {
			if fenceRe.MatchString(c) {
				inFence = false
				end(l.hi)
			}
			continue
		}
		if strings.TrimSpace(c) == "" {
			end(l.lo - 1)
			inTable = false
			continue
		}
		k, opens := Paragraph, true
		m := listRe.FindStringSubmatch(c)
		switch {
		case fenceRe.MatchString(c):
			k = Fence
		case headingRe.MatchString(c):
			k = Heading
		case footnoteRe.MatchString(c):
			k = FootnoteDef
		case m != nil:
			k = ListItem
			if m[1] != "" && m[1] != "1" && open && olist != m[2] {
				k, opens = Paragraph, false
			}
		case inTable && !open:
			k = TableRow
		case strings.HasPrefix(strings.TrimSpace(c), "|") && i+1 < len(lines) && delimRe.MatchString(lines[i+1].c):
			k = TableRow
		default:
			opens = false
		}
		inTable = inTable && k == TableRow
		if !opens && open {
			if b := out[len(out)-1]; l.depth == b.depth || (l.depth < b.depth && (b.Kind == Paragraph || b.Kind == ListItem)) {
				continue // a soft wrap, or a lazy continuation
			}
		}
		end(l.lo - 1)
		body := l.hi - len(c)
		if k == ListItem {
			body += len(m[0])
		}
		switch {
		case k == ListItem && m[1] != "":
			olist = m[2]
		case c[0] != ' ' && c[0] != '\t':
			olist = ""
		}
		out = append(out, Block{Kind: k, Start: l.lo, End: l.hi, Body: body, depth: l.depth})
		open = true
		switch k {
		case Fence:
			inFence = true
		case TableRow:
			inTable = true
			end(l.hi)
		case Heading:
			end(l.hi)
		}
	}
	end(len(text))
	ms := string(masked)
	for i := range out {
		b := &out[i]
		if b.Kind == Fence {
			b.Sentences = trimmed(text, b.Body, b.End)
		} else {
			b.Sentences = split(ms, b.Body, b.End)
		}
	}
	return out
}

// Sentences splits text into its sentences, in order: [start, end) spans, each beginning at its
// first non-space byte and ending after its terminator run and closers. Outside every span lie only
// whitespace, ">" markers and list markers.
//
// A terminator run (. ! ?) with the closers right after it ends a sentence when followed by the
// end of its block; by whitespace — a soft wrap is whitespace — and then an uppercase letter, an
// opener or an HTML comment; or, with no closer, by a comment flush, which then opens the next
// sentence (the shape a cut leaves). A comment flush after a closer belongs to the sentence the
// closer ends, and the same test runs again after the comment run. An HTML comment — every anchor
// token — is opaque. There is no abbreviation list: a hand-kept table stops covering its case, so
// "3.5", "e.g. foo" and "x. 2" join, and "Dr. Smith" splits.
func Sentences(text string) [][2]int {
	var out [][2]int
	for _, b := range Blocks(text) {
		out = append(out, b.Sentences...)
	}
	return out
}

const (
	closers = "\"'”’»)]}*_`"
	openers = "\"'“‘«([{*_`"
)

// split applies the terminator rule to m[lo:hi], one block.
func split(m string, lo, hi int) [][2]int {
	var out [][2]int
	start := -1
	for i := lo; i < hi; {
		if n := commentLen(m, i, hi); n > 0 {
			if start < 0 {
				start = i
			}
			i += n
			continue
		}
		c := m[i]
		if isSpace(c) {
			i++
			continue
		}
		if start < 0 {
			start = i
		}
		i++
		if c != '.' && c != '!' && c != '?' {
			continue
		}
		for i < hi && strings.IndexByte(".!?", m[i]) >= 0 {
			i++
		}
		t := i
		for i < hi {
			r, n := utf8.DecodeRuneInString(m[i:hi])
			if !strings.ContainsRune(closers, r) {
				break
			}
			i += n
		}
		if e := sentenceEnd(m, i, hi, i > t); e >= 0 {
			out = append(out, [2]int{start, e})
			start, i = -1, e
		}
	}
	if start >= 0 {
		out = append(out, trimmed(m, start, hi)...)
	}
	return out
}

// sentenceEnd is where the sentence whose terminator run and closers end at e ends, or -1 when it
// runs on.
func sentenceEnd(m string, e, hi int, closed bool) int {
	r := e
	if closed {
		for n := commentLen(m, r, hi); n > 0; n = commentLen(m, r, hi) {
			r += n
		}
	} else if commentLen(m, e, hi) > 0 {
		return e
	}
	j := r
	for j < hi && isSpace(m[j]) {
		j++
	}
	if r < hi && j == r {
		return -1
	}
	if j == hi || commentLen(m, j, hi) > 0 {
		return r
	}
	if c, _ := utf8.DecodeRuneInString(m[j:hi]); unicode.IsUpper(c) || strings.ContainsRune(openers, c) {
		return r
	}
	return -1
}

// trimmed is m[lo:hi] without its leading and trailing whitespace, as one span, or none.
func trimmed(m string, lo, hi int) [][2]int {
	for lo < hi && isSpace(m[lo]) {
		lo++
	}
	for hi > lo && isSpace(m[hi-1]) {
		hi--
	}
	if lo == hi {
		return nil
	}
	return [][2]int{{lo, hi}}
}

// commentLen is the length of the HTML comment starting at i and closing before hi, or 0.
func commentLen(m string, i, hi int) int {
	if !strings.HasPrefix(m[i:hi], tokenOpen) {
		return 0
	}
	if j := strings.Index(m[i:hi], tokenClose); j >= 0 {
		return j + len(tokenClose)
	}
	return 0
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }
