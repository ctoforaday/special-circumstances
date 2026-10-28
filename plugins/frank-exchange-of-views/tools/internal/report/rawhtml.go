package report

// A SHIPPED .md HOLDS NO `<` OUTSIDE CODE BUT A PROOF'S ANCHOR.
//
// A shipped document is composed from units — the report's text, and each section a seat's words
// fill — and a markdown viewer parses it WHOLE: a fence one unit leaves open flips the next unit's
// code and prose. Seats read untrusted web content, and a viewer that renders raw HTML from the .md
// is a page the human opens. So the decision is taken on each whole document, once (shipSet), and
// Write writes it.
//
// EVERY `<` OUTSIDE CODE IS WRITTEN `&lt;` (ltEntity). Every HTML node, autolink and angle-bracket
// link destination a reader can build opens at a `<`, so a text with none outside code holds none of
// them for any reader, whatever its HTML and autolink rules: nothing is predicted about which `<` a
// reader would take for HTML. A character reference opens nothing and prints as the `<` typed; a
// backslash that escaped the `<` goes, as the entity is the escape now and the backslash kept would
// print. What that costs, stated plainly: the raw .md shows `&lt;`, and a seat's `<url>` autolink or
// `[a](<x y>)` destination ships as text.
//
// CODE SHIPS BYTE FOR BYTE, SO WHERE CODE IS IS THE ONE READING SHIP TAKES. A `<` in a code span, a
// code block or a fence's info string is a character to every reader, and an entity there would
// print as `&lt;`. ship reads the text with every `<` escaped (codeReader) — the reading the shipped
// text has outside code — and keeps each `<` that reading holds as code but one at the head of a
// line: there a `<` opens an HTML block before the span running across that line is read
// (lineHeads), and anywhere else code's content moves no boundary. That costs a `<` at a line's head
// inside such a span, which ships as `&lt;` in code.
//
// codeReader is report.html's grammar less the HTML grammar it no longer needs, with goldmark's
// readings of code made cmark-gfm's where they differ: a code span opens where cmark-gfm's record of
// backtick runs opens one (cmarkBackticks); a URL GitHub's extended autolink takes holds no code
// span, a backtick in it included (gfmautolink.go); a footnote's text starts after the spaces past
// its colon, never four of them as code, and a footnote inside a footnote is read (viewerFootnote);
// and a list stays open where only a list nested in it has an empty item before a blank line, so the
// line after it is no code block (viewerList).
//
// THE TOOL'S HTML IS THE TOOL'S BY CONSTRUCTION. The one element the tool writes into a .md is a
// proof's anchor in evidence.md (proofanchor.go): evidenceDoc writes a mark no record text holds,
// and resolveMarkSpans makes it the anchor and returns its span. A `<` inside such a span is the
// tool's and is never escaped; the same tag typed anywhere is a seat's. Assembly strips every anchor
// token as text enters composition, but those the weave turns into notes (stripTokens), and the weave
// takes the rest, so a token in a shipped document is a seat's text.
//
// WHAT A READER IS GUARANTEED. A reader that agrees with codeReader on where code is — pinned against
// cmark-gfm 0.29.0.gfm.13, GitHub's parser — finds no HTML in the file but the proofs' anchors, and
// prints each `&lt;` ship wrote as the `<` typed. A viewer that does not parse the file as CommonMark
// finds no `<` in its prose either; what it takes for code is its own.
//
// The seat-facing views — `show report --assembled` (Layout), `show inquiries` — print the
// record's text as written, because a seat quotes and edits it. report.html never reads the shipped
// text: it renders through pageReader (md.go) and an allowlist after it (site.go).

import (
	"slices"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// codeReader is the reading ship takes of where a shipped .md's code is: pageReader's grammar with no
// HTML grammar and no proof's anchor, its code read as cmark-gfm reads it.
var codeReader = goldmark.New(goldmark.WithParser(parser.NewParser(
	parser.WithBlockParsers(append(codeBlockParsers(),
		util.Prioritized(viewerFootnote{extension.NewFootnoteBlockParser()}, 999))...),
	parser.WithInlineParsers(append(codeInlineParsers(),
		util.Prioritized(extension.NewFootnoteParser(), 101))...),
	parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
)), goldmark.WithExtensions(extension.Table, extension.TaskList))

// parseCode is ship's one parse of a text (codeReader).
func parseCode(src []byte) ast.Node { return codeReader.Parser().Parse(text.NewReader(src)) }

// codeBlockParsers is pageReader's block grammar with the list parser reading as cmark-gfm does
// (viewerList).
func codeBlockParsers() []util.PrioritizedValue {
	ps := blockParsers()
	for i, p := range ps {
		if p.Value == parser.NewListParser() {
			ps[i].Value = viewerList{p.Value.(parser.BlockParser)}
		}
	}
	return ps
}

// viewerFootnote is goldmark's footnote definition parser, reading a definition as cmark-gfm does: the
// spaces after its colon are the marker's, so text four spaces after it is no code block, and a
// definition inside a definition stays where it stands — goldmark's own close moves it to the list
// before the outer one closes, and the outer one, moved into a list it now holds, leaves the document
// unread.
type viewerFootnote struct{ parser.BlockParser }

func (p viewerFootnote) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	n, state := p.BlockParser.Open(parent, reader, pc)
	if n != nil && state&parser.HasChildren != 0 {
		line, _ := reader.PeekLine()
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i < len(line) && line[i] != '\n' && line[i] != '\r' {
			reader.Advance(i)
		}
	}
	return n, state
}

func (p viewerFootnote) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	for a := node.Parent(); a != nil; a = a.Parent() {
		if _, ok := a.(*east.Footnote); ok {
			return
		}
	}
	p.BlockParser.Close(node, reader, pc)
}

// viewerList is goldmark's list parser, keeping a list open past a blank line that follows an empty
// item of a list nested in it. goldmark marks that blank line in one flag per parse, and every list
// open then closes at the next line; cmark-gfm closes only the list whose item is empty, so a line
// indented to the outer item's content is the outer item's.
type viewerList struct{ parser.BlockParser }

func (p viewerList) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	state := p.BlockParser.Continue(node, reader, pc)
	line, _ := reader.PeekLine()
	last, ok := node.LastChild().(*ast.ListItem)
	if state != parser.Close || util.IsBlank(line) || !ok || last.ChildCount() == 0 {
		return state
	}
	if indent, _ := util.IndentWidth(line, reader.LineOffset()); indent < last.Offset {
		return state
	}
	return parser.Continue | parser.HasChildren
}

// backticksKey holds, in one parse's context, the backtick runs cmark-gfm has scanned in the text
// being parsed (cmarkBackticks).
var backticksKey = parser.NewContextKey()

// maxBackticks is the longest run cmark-gfm opens a code span with.
const maxBackticks = 1000

// tickScan is cmark-gfm's record of one text's backtick runs: whether a scan has reached its end with
// no closer, and where the scans last passed a run of each length.
type tickScan struct {
	ended bool
	last  map[int]int
}

// cmarkBackticks is the code span parser, opening a span only where cmark-gfm does. cmark-gfm scans
// for a closer once per opener, notes each run it passes, and once a scan has reached the text's end
// reads an opener whose length it last passed before that opener as having no closer — though a
// later scan, finding a closer, has moved that note back. Its backticks are then text, and so is what
// stands between them.
type cmarkBackticks struct{ parser.InlineParser }

func (p cmarkBackticks) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	n := 0
	for n < len(line) && line[n] == '`' {
		n++
	}
	scan, _ := pc.Get(backticksKey).(*tickScan)
	if scan == nil {
		scan = &tickScan{last: map[int]int{}}
		pc.Set(backticksKey, scan)
	}
	after := seg.Start + n
	if n > maxBackticks || scan.ended && scan.last[n] <= after {
		block.Advance(n)
		return ast.NewTextSegment(text.NewSegment(seg.Start, after))
	}
	scan.toCloser(parent.Lines(), block.Source(), n, after)
	return p.InlineParser.Parse(parent, block, pc)
}

// toCloser is cmark-gfm's scan from offset from for a run of n backticks, through lines.
func (s *tickScan) toCloser(lines *text.Segments, src []byte, n, from int) {
	for i := 0; i < lines.Len(); i++ {
		l := lines.At(i)
		for j := max(l.Start, from); j < l.Stop; {
			if src[j] != '`' {
				j++
				continue
			}
			k := j
			for k < l.Stop && src[k] == '`' {
				k++
			}
			if k-j <= maxBackticks {
				s.last[k-j] = j
			}
			if k-j == n {
				return
			}
			j = k
		}
	}
	s.ended = true
}

func (p cmarkBackticks) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	pc.Set(backticksKey, nil)
	if c, ok := p.InlineParser.(parser.CloseBlocker); ok {
		c.CloseBlock(parent, block, pc)
	}
}

// readers is the parses shipping and assembly take: report.html's reading of a text (page), and where
// a shipped .md's code is (code). They are passed, so a test counts what a shipping or an assembly
// parses by handing it readers that count.
type readers struct {
	page func([]byte) ast.Node
	code func([]byte) ast.Node
}

// stdReaders is the readers the set ships with.
func stdReaders() readers { return readers{page: parsePage, code: parseCode} }

// ship is md as its .md ships (readers.ship).
func ship(md string) string { return stdReaders().ship(md) }

// ship is md as its .md ships: its proofs' anchors in place, and every other `<` outside code written
// ltEntity. Which `<` stand in code is read from the text with all of them escaped: escaping a `<`
// outside code changes how a reader reads the text around it — an autolink, an angle-bracket
// destination or a tag that held a backtick is text, and the backtick may open a span — and that
// reading is the shipped one.
func (rd readers) ship(md string) string {
	md, tool := resolveMarkSpans(md)
	src := []byte(md)
	at := angles(src, tool)
	if len(at) == 0 {
		return md
	}
	read, pos := escapeAt(src, at)
	tree := rd.code(read)
	code, heads := codeIn(tree, read), lineHeads(tree, read)
	var prose []int
	for k, i := range at {
		if _, head := slices.BinarySearch(heads, pos[k]); head || !inSpan(pos[k], code) {
			prose = append(prose, i)
		}
	}
	if len(prose) == len(at) {
		return string(read)
	}
	out, _ := escapeAt(src, prose)
	return string(out)
}

// lineHeads is the offset of the first byte of each line of inline text in tree, in order: where a
// code span that runs across lines stands at a line's head, a `<` there opens an HTML block before any
// span is read, and the span with it.
func lineHeads(tree ast.Node, src []byte) []int {
	var heads []int
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			return ast.WalkSkipChildren, nil
		}
		if entering && n.Type() == ast.TypeBlock {
			for i := 0; i < n.Lines().Len(); i++ {
				j := n.Lines().At(i).Start
				for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
					j++
				}
				heads = append(heads, j)
			}
		}
		return ast.WalkContinue, nil
	})
	slices.Sort(heads)
	return heads
}

// inSpan reports whether offset i lies inside one of spans, which are in order and do not overlap.
func inSpan(i int, spans [][2]int) bool {
	k, _ := slices.BinarySearchFunc(spans, i+1, func(s [2]int, i int) int { return s[0] - i })
	return k > 0 && i < spans[k-1][1]
}

// angles is every `<` of src outside spans, escaped or not.
func angles(src []byte, spans [][2]int) []int {
	var at []int
	for i, c := range src {
		if c == '<' && !inSpan(i, spans) {
			at = append(at, i)
		}
	}
	return at
}

// ltEntity is how ship writes a `<` outside code: a character reference opens nothing in any
// CommonMark reader and prints as the `<` typed, and, unlike a backslash, no GitHub autolink can take
// it as a character of its URL and leave the `<` behind it live.
const ltEntity = "&lt;"

// escapeAt is src with each `<` at an offset of at, which are in order, written ltEntity, and a
// backslash that escapes it taken out — the entity is the escape now, and the backslash kept would
// print — and the offset in the result of each entity. A backslash and the `<` it escapes stand on
// the same side of every code boundary: a span's closing backticks, or a line's end, stand between
// code and what follows it.
func escapeAt(src []byte, at []int) ([]byte, []int) {
	out := make([]byte, 0, len(src)+len(at)*(len(ltEntity)-1))
	pos := make([]int, len(at))
	last := 0
	for k, i := range at {
		cut := i
		if escapedAt(src, i) {
			cut--
		}
		out = append(out, src[last:cut]...)
		pos[k] = len(out)
		out = append(out, ltEntity...)
		last = i + 1
	}
	return append(out, src[last:]...), pos
}

// proofLabel reports whether label is a proof id: the anchor package reads the id shape, from the
// token it would spell.
func proofLabel(label string) bool {
	tok := anchor.Token(label)
	return anchor.Kind(label) == "proof" && anchor.SkipRun(tok, 0) == len(tok)
}
