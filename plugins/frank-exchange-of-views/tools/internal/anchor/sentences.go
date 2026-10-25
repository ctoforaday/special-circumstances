package anchor

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	gtext "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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
}

// blockReader is CommonMark with GitHub's tables and footnote definitions: the block structure of
// the report as any markdown reader of it sees it. A fence's closing line is recorded on the fence
// (fenceCloser), because a fence runs to it and the parser keeps no node for it.
var blockReader = parser.NewParser(
	parser.WithBlockParsers(append(slices.DeleteFunc(parser.DefaultBlockParsers(), func(p util.PrioritizedValue) bool {
		return p.Priority == fencePriority
	}),
		util.Prioritized(fenceCloser{parser.NewFencedCodeBlockParser()}, fencePriority),
		util.Prioritized(extension.NewFootnoteBlockParser(), 999),
	)...),
	parser.WithInlineParsers(parser.DefaultInlineParsers()...),
	parser.WithParagraphTransformers(append(parser.DefaultParagraphTransformers(),
		util.Prioritized(extension.NewTableParagraphTransformer(), 200))...),
	parser.WithASTTransformers(util.Prioritized(extension.NewTableASTTransformer(), 0)),
)

// fencePriority is the fenced-code parser's place among parser.DefaultBlockParsers.
const fencePriority = 700

// fenceCloser is the fenced-code parser, recording where the line that closed a fence ends.
type fenceCloser struct{ parser.BlockParser }

const fenceClose = "fence-close"

func (f fenceCloser) Continue(n ast.Node, r gtext.Reader, pc parser.Context) parser.State {
	_, seg := r.PeekLine()
	st := f.BlockParser.Continue(n, r, pc)
	if st == parser.Close {
		n.SetAttributeString(fenceClose, seg.Start)
	}
	return st
}

// Blocks splits text into its markdown blocks as the parser reads them, because a newline ends a
// sentence only where it ends a block: a paragraph, a heading, a footnote definition, a list item's
// first paragraph (a later one is a paragraph), each line of a table, and a fence or an indented
// code block, which is one sentence. HTML on its own lines reads as a paragraph, and an HTML
// comment — every anchor token — is prose to the block reader, as it is to the sentence rule. A
// thematic break and a link reference definition hold no sentence.
func Blocks(text string) []Block {
	src := []byte(text)
	// The parser reads each comment as a run of letters of its length: a line that opens with an
	// anchor token is the paragraph it reads as, never an HTML block, and every offset stands.
	parsed := make([]byte, len(src))
	copy(parsed, src)
	for i := 0; i < len(parsed); i++ {
		if parsed[i] != '<' {
			continue
		}
		if n := commentLen(text, i, lineEndAt(text, i)); n > 0 {
			for j := i; j < i+n; j++ {
				parsed[j] = 'x'
			}
			i += n - 1
		}
	}
	doc := blockReader.Parse(gtext.NewReader(parsed))
	w := &blockWalk{text: text, masked: []byte(text), opaque: map[int]int{}}
	w.walk(doc)
	slices.SortStableFunc(w.out, func(a, b Block) int { return a.Start - b.Start })
	m := string(w.masked)
	for i := range w.out {
		b := &w.out[i]
		if b.Kind == Fence {
			b.Sentences = trimmed(m, b.Body, b.End)
		} else {
			b.Sentences = splitter{m, w.opaque}.split(b.Body, b.End)
		}
	}
	return w.out
}

// blockWalk collects the blocks of a parsed text. masked is the text with each content line's
// container prefix — its ">" markers and indentation — read as whitespace; opaque is each code
// span, start to end, in which no sentence ends.
type blockWalk struct {
	text   string
	masked []byte
	opaque map[int]int
	out    []Block
}

func (w *blockWalk) walk(n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		w.node(c)
	}
}

func (w *blockWalk) node(n ast.Node) {
	switch x := n.(type) {
	case *ast.Paragraph, *ast.TextBlock, *ast.HTMLBlock:
		lines := w.mask(n)
		if len(lines) == 0 {
			return
		}
		first, last := lines[0], lines[len(lines)-1]
		b := Block{Kind: Paragraph, Start: lineStartAt(w.text, first.Start), Body: first.Start, End: lineEndAt(w.text, last.Start)}
		if h, ok := n.(*ast.HTMLBlock); ok && h.HasClosure() {
			b.End = lineEndAt(w.text, h.ClosureLine.Start)
		}
		switch p := n.Parent().(type) {
		case *ast.ListItem:
			if p.FirstChild() == n {
				b.Kind, b.Start = ListItem, lineStartAt(w.text, p.Pos())
			}
		case *east.Footnote:
			if p.FirstChild() == n {
				b.Kind, b.Start, b.Body = FootnoteDef, lineStartAt(w.text, p.Pos()), p.Pos()
			}
		}
		copy(w.masked[b.Body:first.Start], w.text[b.Body:first.Start])
		w.out = append(w.out, b)
		w.codeSpans(n)
	case *ast.ListItem:
		if c := x.FirstChild(); c == nil || !isParagraph(c) {
			// An item whose first line holds no paragraph — empty, or a fence — is a block of its own.
			at := lineEndAt(w.text, x.Pos())
			w.out = append(w.out, Block{Kind: ListItem, Start: lineStartAt(w.text, x.Pos()), Body: at, End: at})
		}
		w.walk(x)
	case *east.Footnote:
		if c := x.FirstChild(); c == nil || !isParagraph(c) {
			at := lineEndAt(w.text, x.Pos())
			w.out = append(w.out, Block{Kind: FootnoteDef, Start: lineStartAt(w.text, x.Pos()), Body: x.Pos(), End: at})
		}
		w.walk(x)
	case *ast.Heading:
		w.mask(n)
		if l := x.Lines(); l.Len() > 0 && x.Pos() >= 0 {
			copy(w.masked[x.Pos():l.At(0).Start], w.text[x.Pos():l.At(0).Start])
		}
		b := Block{Kind: Heading, Start: lineStartAt(w.text, x.Pos()), Body: x.Pos(), End: lineEndAt(w.text, x.Pos())}
		if x.Pos() < 0 { // a setext heading: its text lines
			lines := x.Lines()
			b.Body = lines.At(0).Start
			b.Start = lineStartAt(w.text, b.Body)
			b.End = lineEndAt(w.text, lines.At(lines.Len()-1).Start)
		}
		w.out = append(w.out, b)
		w.codeSpans(n)
	case *ast.FencedCodeBlock:
		w.mask(n)
		end := len(w.text) // a fence nothing closes runs to the end of the text
		if v, ok := x.AttributeString(fenceClose); ok {
			end = lineEndAt(w.text, v.(int))
		} else if l := x.Lines(); l.Len() > 0 && x.NextSibling() != nil {
			end = lineEndAt(w.text, l.At(l.Len()-1).Start)
		}
		w.out = append(w.out, Block{Kind: Fence, Start: lineStartAt(w.text, x.Pos()), Body: x.Pos(), End: end})
	case *ast.CodeBlock:
		lines := w.mask(n)
		if len(lines) > 0 {
			w.out = append(w.out, Block{Kind: Fence, Start: lineStartAt(w.text, lines[0].Start), Body: lines[0].Start, End: lineEndAt(w.text, lines[len(lines)-1].Start)})
		}
	case *east.Table:
		// Each line of a table is a block: the header, the delimiter row under it, and each row.
		for row := x.FirstChild(); row != nil; row = row.NextSibling() {
			w.row(row.Pos())
			if _, head := row.(*east.TableHeader); head {
				w.row(lineEndAt(w.text, row.Pos()) + 1)
			}
			w.codeSpans(row)
		}
	default:
		w.walk(n)
	}
}

// row adds the table line whose content starts at body.
func (w *blockWalk) row(body int) {
	start := lineStartAt(w.text, body)
	for i := start; i < body; i++ {
		w.masked[i] = ' '
	}
	w.out = append(w.out, Block{Kind: TableRow, Start: start, Body: body, End: lineEndAt(w.text, body)})
}

// mask reads each of n's lines' container prefixes as whitespace, and returns its lines.
func (w *blockWalk) mask(n ast.Node) []gtext.Segment {
	var segs []gtext.Segment
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		for j := lineStartAt(w.text, seg.Start); j < seg.Start; j++ {
			w.masked[j] = ' '
		}
		segs = append(segs, seg)
	}
	return segs
}

// codeSpans records each code span in n's inline content as opaque.
func (w *blockWalk) codeSpans(n ast.Node) {
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if cs, ok := c.(*ast.CodeSpan); ok && entering {
			lo, hi := -1, -1
			for t := cs.FirstChild(); t != nil; t = t.NextSibling() {
				if tx, ok := t.(*ast.Text); ok {
					if lo < 0 {
						lo = tx.Segment.Start
					}
					hi = tx.Segment.Stop
				}
			}
			if lo >= 0 && hi > lo {
				w.opaque[lo] = hi
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
}

func isParagraph(n ast.Node) bool {
	switch n.(type) {
	case *ast.Paragraph, *ast.TextBlock, *ast.HTMLBlock:
		return true
	}
	return false
}

// lineStartAt and lineEndAt are where the line holding byte i starts, and where it ends before its
// newline.
func lineStartAt(text string, i int) int {
	return strings.LastIndexByte(text[:min(i, len(text))], '\n') + 1
}

func lineEndAt(text string, i int) int {
	if i >= len(text) {
		return len(text)
	}
	if j := strings.IndexByte(text[i:], '\n'); j >= 0 {
		return i + j
	}
	return len(text)
}

// Sentences splits text into its sentences, in order: [start, end) spans, each beginning at its
// first non-space byte and ending after its terminator run and closers. Outside every span lie only
// whitespace, ">" markers and list markers.
//
// THE SENTENCE RULE IS A JUDGEMENT ABOUT PROSE, and it is the one thing here the markdown parser
// does not decide. A terminator run (. ! ?) with the closers right after it ends a sentence when
// followed by the end of its block; by whitespace — a soft wrap is whitespace — and then an
// uppercase letter, an opener, an HTML comment, or a digit after a word of five letters or more
// (wordBefore) when the number is whole (decimalAt); or, with no closer, by a comment flush, which then opens the next sentence (the
// shape a cut leaves). A comment flush after a closer belongs to the sentence the closer ends, and
// the same test runs again after the comment run. An HTML comment — every anchor token — and a code
// span are opaque. There is no abbreviation list: a hand-kept table stops covering its case, so
// "3.5", "e.g. foo", "Fig. 3", "approx. 0.99" and "is 7. 13" join, and "Dr. Smith" and "approx. 10"
// split.
func Sentences(text string) [][2]int {
	var out [][2]int
	for _, b := range Blocks(text) {
		out = append(out, b.Sentences...)
	}
	return out
}

const (
	terminators = ".!?"
	closers     = "\"'”’»)]}*_`"
	openers     = "\"'“‘«([{*_`"
	spaces      = " \t\n\r"
)

// splitter applies the terminator rule to the masked text m; opaque maps the start of each code
// span to its end, and no sentence ends inside one.
type splitter struct {
	m      string
	opaque map[int]int
}

// split applies the terminator rule to m[lo:hi], one block.
func (s splitter) split(lo, hi int) [][2]int {
	m := s.m
	var out [][2]int
	start := -1
	for i := lo; i < hi; {
		n := commentLen(m, i, hi)
		if e, ok := s.opaque[i]; ok && n == 0 {
			n = min(e, hi) - i
		}
		if n == 0 && strings.IndexByte(spaces, m[i]) >= 0 {
			i++
			continue
		}
		if start < 0 {
			start = i
		}
		if n > 0 || strings.IndexByte(terminators, m[i]) < 0 {
			i += max(n, 1)
			continue
		}
		t := skip(m, i, hi, terminators)
		i = skip(m, t, hi, closers)
		if e := sentenceEnd(m, start, i, hi, i > t); e >= 0 {
			out = append(out, [2]int{start, e})
			start, i = -1, e
		}
	}
	if start >= 0 {
		out = append(out, trimmed(m, start, hi)...)
	}
	return out
}

// sentenceEnd is where the sentence that starts at from, and whose terminator run and closers end
// at e, ends, or -1 when it runs on.
func sentenceEnd(m string, from, e, hi int, closed bool) int {
	r := e
	if closed {
		for n := commentLen(m, r, hi); n > 0; n = commentLen(m, r, hi) {
			r += n
		}
	} else if commentLen(m, e, hi) > 0 {
		return e
	}
	j := skip(m, r, hi, spaces)
	c, _ := utf8.DecodeRuneInString(m[j:hi])
	if j == hi || j > r && (commentLen(m, j, hi) > 0 || unicode.IsUpper(c) || strings.ContainsRune(openers, c) || unicode.IsDigit(c) && wordBefore(m, from, e) >= minWordBeforeDigit && !decimalAt(m, j, hi)) {
		return r
	}
	return -1
}

// minWordBeforeDigit is the shortest word that, ending a sentence, lets the next one open with a
// digit. The abbreviations a number follows — "Fig. 3", "No. 5", "p. 12", "Def. 11", "e.g. 3",
// "vol. 2" — are shorter; a word that ends a sentence mostly is not.
const minWordBeforeDigit = 5

// decimalAt reports whether the number at j has a decimal or a thousands separator — "0.999",
// "3,215" — the shape a number after an abbreviation such as "approx." takes.
func decimalAt(m string, j, hi int) bool {
	k := j
	for k < hi && m[k] >= '0' && m[k] <= '9' {
		k++
	}
	return k+1 < hi && (m[k] == '.' || m[k] == ',') && m[k+1] >= '0' && m[k+1] <= '9'
}

// wordBefore is the length in letters of the word the terminator run that ends at e follows, in a
// sentence from from: the terminators, the comments and the closers before them are passed over,
// so "composite<!--cite:C-1-->." and "(mod 91)." read their word; a word holding anything but
// letters — "3.", "e.g." — is no word.
func wordBefore(m string, from, e int) int {
	i := e
	for i > from && strings.IndexByte(terminators+closers, m[i-1]) >= 0 {
		i--
	}
	for i > from && strings.HasSuffix(m[from:i], tokenClose) {
		o := strings.LastIndex(m[from:i], tokenOpen)
		if o < 0 {
			break
		}
		i = from + o
		for i > from && strings.IndexByte(closers, m[i-1]) >= 0 {
			i--
		}
	}
	n := 0
	for i > from {
		c, size := utf8.DecodeLastRuneInString(m[from:i])
		if !unicode.IsLetter(c) {
			if strings.IndexByte(spaces, m[i-1]) < 0 && strings.IndexByte(openers, m[i-1]) < 0 {
				return 0
			}
			break
		}
		n++
		i -= size
	}
	return n
}

// skip is the index past the run of set's characters starting at i, before hi.
func skip(m string, i, hi int, set string) int { return hi - len(strings.TrimLeft(m[i:hi], set)) }

// trimmed is m[lo:hi] without its leading and trailing whitespace, as one span, or none.
func trimmed(m string, lo, hi int) [][2]int {
	lo = skip(m, lo, hi, spaces)
	if hi = lo + len(strings.TrimRight(m[lo:hi], spaces)); lo == hi {
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
