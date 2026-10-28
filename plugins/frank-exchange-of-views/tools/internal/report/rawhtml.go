package report

// A SHIPPED .md HOLDS NO SEAT'S HTML FOR A COMMONMARK READER.
//
// A shipped document is composed from units — the report's text, and each section a seat's words
// fill — and a markdown viewer parses it WHOLE: a fence one unit leaves open flips the next unit's
// code and prose, and an unclosed `<!--` hides every section after it. Seats read untrusted web
// content, and a viewer that renders raw HTML from the .md is a page the human opens. So the
// decision is taken on each whole document, once (shipSet), and Write writes it.
//
// THE .md READS AS report.html READS. report.html reads a document with no HTML grammar
// (pageReader): a tag, a comment or an HTML block's opener is text in the paragraph it stands in,
// and the lines after it are what they would be without it. The shipped .md is that reading, made
// to hold for a reader that has the grammar (mdReader): ship writes each `<` at which such a reader
// would open an HTML node in that reading as `&lt;` (ltEntity), and no other. Code, autolinks,
// link destinations and info strings are not edited — the reading has no HTML node in them, so no
// `<` there is one — and that includes code that is code only because an enclosing HTML block is
// read as text: `<div>`, a fence, `</div>` ships its fence byte for byte.
//
// WHERE A READER OPENS AN HTML NODE IS THE PARSER'S ANSWER, taken in one parse (htmlAt). probeReader
// is mdReader's grammar with its HTML block parser and inline HTML parser each wrapped: the wrapped
// parser runs, and where it would open a node the wrapper records the `<` and declines. A parser
// that declines is a parser that is absent, so the tree is the viewer's with no HTML grammar, and each
// `<` recorded is one mdReader opens a node at when the text before it reads that way. Escaping every
// one makes that so: a `&lt;` opens nothing, and is the same character to every other parser. Each
// `<` is decided in the reading the shipped text has, so escaping one never exposes another — no
// pass per row of a pasted table, no pass per tag a backtick frees.
//
// WHERE GOLDMARK AND GITHUB'S PARSER READ A `<` APART, THE PROBE READS IT AS GITHUB DOES. The viewer is
// cmark-gfm; the readers are goldmark, and differential runs against cmark-gfm find goldmark reading
// as no HTML a `<` cmark-gfm opens a tag at. Each such reading is corrected where it is made: a
// footnote's text starts after the spaces past its colon, and a footnote inside one is read
// (viewerFootnote); an opener a tab indents less than four columns is an opener (htmlBlockProbe); a
// `<` inside an angle-bracket destination, where cmark-gfm takes no link, is recorded (bracketWatch,
// linkRefProbe), and as `&lt;` it is the same destination to both; a `<?` or `<!` before a byte
// that is not UTF-8, where cmark-gfm's scanner closes a node it found no end of, is recorded
// (rawHTMLProbe); a code span opens where cmark-gfm's record of backtick runs opens one
// (cmarkBackticks); and a list stays open where only a list nested in it has an empty item before a
// blank line (viewerList).
//
// A `<` IS NEUTRALISED BY WHAT THE READER DOES, NOT BY THE BYTES BEFORE IT. The viewer's grammar is
// GitHub's, extended autolink and all (gfmautolink.go): a URL takes every byte up to whitespace or a
// `<`, so a seat's "http://x.co/a\<img …>" is a link ending in a backslash and a live tag, while to
// goldmark the backslash escapes the `<`. The probe reads it as the viewer does and records that `<`;
// its edit is the entity, since a backslash written there would be taken the same way, and the
// backslash that stood before it goes, as the entity is now the escape. The viewer's grammar is
// report.html's but for that autolink, so the readings differ only where a URL in prose takes a byte
// goldmark's linkify would not.
//
// ship then parses the result as a reader does (parseMD) and checks it holds no seat HTML, whatever
// stands before its `<`. A text that defeats the probe — a reader's reading the probe did not predict,
// as when an edit lets a URL run on and take a backtick — ships with every `<` but the tool's written
// `&lt;`, code included. That text holds no seat HTML by construction — no `<` a seat wrote is left,
// and an HTML node opens only at a `<` — so it is not checked again, and no seat's text can stop a
// document shipping.
//
// THE TOOL'S HTML IS THE TOOL'S BY CONSTRUCTION. The one element the tool writes into a .md is a
// proof's anchor in evidence.md (proofanchor.go): evidenceDoc writes a mark no record text holds,
// and resolveMarkSpans makes it the anchor and returns its span. A `<` inside such a span is the
// tool's and is never escaped; the same tag typed anywhere is a seat's. Assembly strips every anchor
// token as text enters composition, but those the weave turns into notes (stripTokens), and the weave
// takes the rest, so a token in a shipped document is a seat's text.
//
// WHAT A READER IS GUARANTEED. A reader that parses the file as CommonMark with GitHub's extensions
// finds no HTML in it but the proofs' anchors, and prints each `&lt;` ship wrote as the `<` typed, but
// inside code a fallback reached and in a URL GitHub's autolink takes it into.
// A viewer that does not parse the file as CommonMark is not covered: one that reads
// `<svg:x onfocus=…>` — an autolink to CommonMark — or an angle-bracket link destination as a tag
// finds a tag there.
//
// The seat-facing views — `show report --assembled` (Layout), `show inquiries` — print the
// record's text as written, because a seat quotes and edits it. report.html never reads the shipped
// text: it renders through pageReader (md.go) and an allowlist after it (site.go).

import (
	"slices"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// parseMD is a parse of a shipped .md, in the dialect a markdown viewer reads it in (mdReader).
func parseMD(src []byte) ast.Node { return mdReader.Parser().Parse(text.NewReader(src)) }

// probeReader is mdReader's grammar with its two HTML parsers each wrapped to record where it would open
// a node, and decline: the viewer's grammar with no HTML in it, and the HTML grammar as a probe.
var probeReader = goldmark.New(goldmark.WithParser(parser.NewParser(
	parser.WithBlockParsers(append(viewerBlockParsers(),
		util.Prioritized(htmlBlockProbe{parser.NewHTMLBlockParser()}, htmlBlockPriority),
		util.Prioritized(viewerFootnote{extension.NewFootnoteBlockParser()}, 999))...),
	parser.WithInlineParsers(append(viewerInlineParsers(),
		util.Prioritized(rawHTMLProbe{parser.NewRawHTMLParser()}, rawHTMLPriority),
		util.Prioritized(extension.NewFootnoteParser(), 101))...),
	parser.WithParagraphTransformers(util.Prioritized(linkRefProbe{parser.LinkReferenceParagraphTransformer}, 100)),
)), goldmark.WithExtensions(extension.Table, extension.TaskList))

// htmlAtKey holds, in one parse's context, the offsets the probes recorded.
var htmlAtKey = parser.NewContextKey()

// htmlAt is the offset of each `<` of src at which a CommonMark reader opens an HTML node, src read as
// pageReader reads it (probeReader), in order.
func htmlAt(src []byte) []int {
	pc := parser.NewContext()
	probeReader.Parser().Parse(text.NewReader(src), parser.WithContext(pc))
	at, _ := pc.Get(htmlAtKey).([]int)
	slices.Sort(at)
	return slices.Compact(at)
}

// recordHTML adds offset i of the source to the parse's record.
func recordHTML(pc parser.Context, i int) {
	at, _ := pc.Get(htmlAtKey).([]int)
	pc.Set(htmlAtKey, append(at, i))
}

// firstAngle is the offset of the first `<` of src in seg: the one an HTML parser opened on.
func firstAngle(src []byte, seg text.Segment) int {
	for i := seg.Start; i < seg.Stop; i++ {
		if src[i] == '<' {
			return i
		}
	}
	return seg.Start
}

// htmlBlockProbe is the HTML block parser, recording the `<` of a line it would open a block on.
type htmlBlockProbe struct{ parser.BlockParser }

func (p htmlBlockProbe) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, pos := reader.Position()
	l, seg := reader.PeekLine()
	n, _ := p.BlockParser.Open(parent, reader, pc)
	if n == nil {
		n = p.tabbed(parent, l, reader.LineOffset(), pc)
	}
	if n != nil {
		recordHTML(pc, firstAngle(reader.Source(), seg))
		reader.SetPosition(line, pos)
	}
	return nil, parser.NoChildren
}

// tabbed is the block the HTML block parser opens on line once the whitespace it opens with is read
// as columns, or nil. goldmark's parser matches its openers from the line's first byte, so a tab that
// indents less than four columns — what a blockquote's `> ` leaves of one — hides an opener cmark-gfm
// reads.
func (p htmlBlockProbe) tabbed(parent ast.Node, line []byte, column int, pc parser.Context) ast.Node {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if !slices.Contains(line[:i], '\t') {
		return nil
	}
	if w, _ := util.IndentWidth(line, column); w >= 4 {
		return nil
	}
	n, _ := p.BlockParser.Open(parent, text.NewReader(line[i:]), pc)
	return n
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

// linkRefProbe is the link reference definition transformer, recording each `<` inside an
// angle-bracket destination of a definition it takes: goldmark takes one there, and cmark-gfm takes
// no such definition, so the line is text and the `<` may open a tag. Written `&lt;`, it is the same
// destination to both. The record is every `]:` of the lines taken followed by such a destination,
// a title's included: an entity there is the same character too.
type linkRefProbe struct{ parser.ParagraphTransformer }

func (p linkRefProbe) Transform(node *ast.Paragraph, reader text.Reader, pc parser.Context) {
	parent, lines := node.Parent(), slices.Clone(node.Lines().Sliced(0, node.Lines().Len()))
	p.ParagraphTransformer.Transform(node, reader, pc)
	kept := 0
	if node.Parent() == parent {
		kept = node.Lines().Len()
	}
	src := reader.Source()
	taken := lines[:len(lines)-kept]
	for k, seg := range taken {
		line := src[seg.Start:seg.Stop]
		for j := 0; j+1 < len(line); j++ {
			if line[j] == '\\' {
				j++
				continue
			}
			if line[j] != ']' || line[j+1] != ':' {
				continue
			}
			rest, at := line[j+2:], seg.Start+j+2
			if t := skipBlank(rest); t == len(rest) || rest[t] == '\n' || rest[t] == '\r' {
				if k+1 == len(taken) {
					continue
				}
				rest, at = src[taken[k+1].Start:taken[k+1].Stop], taken[k+1].Start
			}
			t := skipBlank(rest)
			for _, i := range angleDestAngles(rest[t:]) {
				recordHTML(pc, at+t+i)
			}
		}
	}
}

// viewerBlockParsers is pageReader's block grammar with the list parser reading as cmark-gfm does
// (viewerList).
func viewerBlockParsers() []util.PrioritizedValue {
	ps := blockParsers()
	for i, p := range ps {
		if p.Value == parser.NewListParser() {
			ps[i].Value = viewerList{p.Value.(parser.BlockParser)}
		}
	}
	return ps
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

// skipBlank is the length of the spaces and tabs b opens with.
func skipBlank(b []byte) int {
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\t') {
		i++
	}
	return i
}

// angleDestAngles is the offset in b of each unescaped `<` inside the angle-bracket destination b opens
// with, as goldmark reads one — to the first unescaped `>` on the line — or nil where b opens with
// none. CommonMark admits no `<` there, and cmark-gfm takes no link whose destination holds one.
func angleDestAngles(b []byte) []int {
	if len(b) == 0 || b[0] != '<' {
		return nil
	}
	var at []int
	for i := 1; i < len(b); i++ {
		switch {
		case b[i] == '\\' && i+1 < len(b) && util.IsPunct(b[i+1]):
			i++
		case b[i] == '>':
			return at
		case b[i] == '<':
			at = append(at, i)
		case b[i] == '\n':
			return nil
		}
	}
	return nil
}

// rawHTMLProbe is the inline HTML parser, recording the `<` it would open a tag or comment on. The
// inline parse restores the reader's position after a parser that declines.
type rawHTMLProbe struct{ parser.InlineParser }

func (p rawHTMLProbe) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if p.InlineParser.Parse(parent, block, pc) != nil || unterminated(line, parent, block.Source(), seg) {
		recordHTML(pc, firstAngle(block.Source(), seg))
	}
	return nil
}

// unterminated reports whether cmark-gfm opens a node at the `<?` or `<!` line opens with although no
// terminator closes it: its scanners for a processing instruction, a declaration and a CDATA section
// stop at a byte that is not UTF-8 and count the terminator's length past it as matched, so the node
// ends wherever that count lands. Any such byte from here to the end of the text parent holds is taken
// as that one.
func unterminated(line []byte, parent ast.Node, src []byte, seg text.Segment) bool {
	if len(line) < 2 || line[1] != '?' && line[1] != '!' {
		return false
	}
	lines := parent.Lines()
	if lines.Len() == 0 {
		return false
	}
	return !utf8.Valid(src[seg.Start:max(seg.Start, lines.At(lines.Len()-1).Stop)])
}

// readers is the parses shipping and assembly take: report.html's reading of a text (page), a markdown
// viewer's (md), and where that viewer opens HTML in report.html's reading (probe). They are passed, so
// a test counts what a shipping or an assembly parses by handing it readers that count.
type readers struct {
	page  func([]byte) ast.Node
	md    func([]byte) ast.Node
	probe func([]byte) []int
}

// stdReaders is the readers the set ships with.
func stdReaders() readers { return readers{page: parsePage, md: parseMD, probe: htmlAt} }

// ship is md as its .md ships (readers.ship).
func ship(md string) string { return stdReaders().ship(md) }

// ship is md as its .md ships: its proofs' anchors in place, and no HTML in it a CommonMark reader
// finds but those anchors.
func (rd readers) ship(md string) string {
	md, tool := resolveMarkSpans(md)
	src := []byte(md)
	if len(angles(src, tool)) == 0 {
		return md
	}
	var at []int
	for _, i := range rd.probe(src) {
		if !inSpan(i, tool) {
			at = append(at, i)
		}
	}
	if out, spans := escapeAt(src, at, tool); len(seatHTML(rd.md(out), out, spans)) == 0 {
		return string(out)
	}
	out, _ := escapeAt(src, angles(src, tool), tool)
	return string(out)
}

// seatHTML is the offset of every `<` in src that stands inside an HTML node of tree, outside the
// tool's anchors (tool, their spans), in order. A backslash before one is no defence: the node is
// there because the reader did not take the backslash as an escape.
func seatHTML(tree ast.Node, src []byte, tool [][2]int) []int {
	var at []int
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		var segs []text.Segment
		switch x := n.(type) {
		case *ast.HTMLBlock:
			segs = x.Lines().Sliced(0, x.Lines().Len())
			if x.HasClosure() {
				segs = append(segs, x.ClosureLine)
			}
		case *ast.RawHTML:
			segs = x.Segments.Sliced(0, x.Segments.Len())
		default:
			return ast.WalkContinue, nil
		}
		for _, seg := range segs {
			for i := seg.Start; i < seg.Stop; i++ {
				if src[i] == '<' && !inSpan(i, tool) {
					at = append(at, i)
				}
			}
		}
		return ast.WalkSkipChildren, nil
	})
	slices.Sort(at)
	return slices.Compact(at)
}

// inSpan reports whether offset i lies inside one of spans, which are in order and do not overlap.
func inSpan(i int, spans [][2]int) bool {
	k, _ := slices.BinarySearchFunc(spans, i+1, func(s [2]int, i int) int { return s[0] - i })
	return k > 0 && i < spans[k-1][1]
}

// angles is every `<` of src outside spans, escaped or not: the fallback's edits.
func angles(src []byte, spans [][2]int) []int {
	var at []int
	for i, c := range src {
		if c == '<' && !inSpan(i, spans) {
			at = append(at, i)
		}
	}
	return at
}

// ltEntity is how ship writes a `<` a reader would open HTML at: a character reference opens nothing
// in any CommonMark reader and prints as the `<` typed, and, unlike a backslash, no GitHub autolink
// can take it as a character of its URL and leave the `<` behind it live.
const ltEntity = "&lt;"

// escapeAt is src with each `<` at an offset of at, which are in order, written ltEntity, and a
// backslash that escapes it taken out — the entity is the escape now, and the backslash kept would
// print — with spans moved to where their bytes now stand. No offset of at lies inside a span, and no
// span ends on the backslash before one: a span ends with its tag's `>`.
func escapeAt(src []byte, at []int, spans [][2]int) ([]byte, [][2]int) {
	out := make([]byte, 0, len(src)+len(at)*(len(ltEntity)-1))
	shift := make([]int, len(at)) // what the edits up to and including each one add to the length
	last, grown := 0, 0
	for k, i := range at {
		cut := i
		if escapedAt(src, i) {
			cut--
		}
		out = append(append(out, src[last:cut]...), ltEntity...)
		last = i + 1
		grown += len(ltEntity) - 1 - (i - cut)
		shift[k] = grown
	}
	out = append(out, src[last:]...)
	moved := make([][2]int, len(spans))
	for k, s := range spans {
		d := 0
		if n, _ := slices.BinarySearch(at, s[0]); n > 0 {
			d = shift[n-1]
		}
		moved[k] = [2]int{s[0] + d, s[1] + d}
	}
	return out, moved
}

// proofLabel reports whether label is a proof id: the anchor package reads the id shape, from the
// token it would spell.
func proofLabel(label string) bool {
	tok := anchor.Token(label)
	return anchor.Kind(label) == "proof" && anchor.SkipRun(tok, 0) == len(tok)
}
