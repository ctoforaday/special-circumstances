package report

// A SHIPPED .md HOLDS NO SEAT'S HTML FOR A COMMONMARK READER.
//
// A shipped document is composed from units — the report's text, and each section a seat's words
// fill — and a markdown viewer parses it WHOLE: a fence one unit leaves open flips the next unit's
// code and prose, and an unclosed `<!--` hides every section after it. Seats read untrusted web
// content, and a viewer that renders raw HTML from the .md is a page the human opens. So the
// decision is taken on each whole document, once (Ship), and Write writes it.
//
// WHAT IS HTML IS THE PARSER'S READING, in the grammar a CommonMark viewer reads the file in
// (mdReader: CommonMark, GitHub's extensions, footnotes). ship parses the document and puts a
// backslash before every `<` inside each HTML node — an HTML block's lines, an inline tag or
// comment — but the tool's, and parses again, because text that stops being HTML can let the lines
// around it read differently: a comment line escaped into a paragraph takes the indented line after
// it as a continuation, where it was code. It stops at the first parse that
// finds no seat HTML. Code, autolinks, link destinations and info strings are never edited: the
// parser reads none of them as HTML.
//
// THE TOOL'S HTML IS THE TOOL'S BY CONSTRUCTION. The one element the tool writes into a .md is a
// proof's anchor in evidence.md (proofanchor.go): evidenceDoc writes a mark no record text holds,
// and resolveMarkSpans makes it the anchor and returns its span. A `<` inside such a span is the
// tool's and is never escaped; the same tag typed anywhere is a seat's. Assembly weaves or strips
// every anchor token before a document ships (AssembleAll), so a token in one is a seat's text.
//
// THE PARSES ARE BOUNDED. Each parse costs the document's length, and maxEscapePasses caps them;
// a text that still holds seat HTML after them has every `<` but the tool's escaped, code or not —
// the backslash then prints inside code, which only text built to need that many passes meets.
//
// The seat-facing views — `show report --assembled` (Layout), `show inquiries` — print the
// record's text as written, because a seat quotes and edits it. report.html never reads this text:
// it renders through a reader with no HTML grammar (md.go) and an allowlist after it (site.go).

import (
	"bytes"
	"slices"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// parseMD is a parse of a shipped .md, in the dialect a markdown viewer reads it in (mdReader).
var parseMD = func(src []byte) ast.Node { return mdReader.Parser().Parse(text.NewReader(src)) }

// maxEscapePasses is how many parses ship spends finding seat HTML before it escapes every `<`.
var maxEscapePasses = 4

// ship is md as its .md ships: its proofs' anchors in place, and no HTML in it a CommonMark reader
// finds but those anchors.
//
// WHAT A READER IS GUARANTEED. A reader that parses the file as CommonMark with GitHub's extensions
// finds no HTML in it but the proofs' anchors: every other `<` that would open a tag, a comment or
// an HTML block carries a backslash, and such a reader prints it as the character typed. A viewer
// that does not parse the file as CommonMark is not covered: one that reads `<svg:x onfocus=…>` —
// an autolink to CommonMark — or an angle-bracket link destination as a tag finds a tag there.
func ship(md string) string {
	md, tool := resolveMarkSpans(md)
	src, spans := []byte(md), tool
	for pass := 1; ; pass++ {
		at := seatHTML(parseMD(src), src, spans)
		if len(at) == 0 {
			return string(src)
		}
		if pass == maxEscapePasses {
			break
		}
		src, spans = escapeAt(src, at, spans)
	}
	src, spans = escapeAt([]byte(md), angles([]byte(md), tool), tool)
	if len(seatHTML(parseMD(src), src, spans)) > 0 {
		panic("report: text with every `<` escaped but the proofs' anchors still parses to HTML the tool did not write")
	}
	return string(src)
}

// seatHTML is the offset of every `<` in src that stands inside an HTML node of tree, outside the
// tool's anchors (tool, their spans), and is not already escaped, in order.
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
				if src[i] == '<' && !inSpan(i, tool) && !escaped(src, i) {
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

// escaped reports whether the byte at i is escaped: an odd run of backslashes stands before it.
func escaped(src []byte, i int) bool {
	n := 0
	for k := i - 1; k >= 0 && src[k] == '\\'; k-- {
		n++
	}
	return n%2 == 1
}

// angles is every `<` of src outside spans that is not already escaped: the fallback's edits.
func angles(src []byte, spans [][2]int) []int {
	var at []int
	for i, c := range src {
		if c == '<' && !inSpan(i, spans) && !escaped(src, i) {
			at = append(at, i)
		}
	}
	return at
}

// escapeAt is src with a backslash before each offset of at, which are in order, and spans moved
// to where their bytes now stand. No offset of at lies inside a span.
func escapeAt(src []byte, at []int, spans [][2]int) ([]byte, [][2]int) {
	out := make([]byte, 0, len(src)+len(at))
	last := 0
	for _, i := range at {
		out = append(append(out, src[last:i]...), '\\')
		last = i
	}
	out = append(out, src[last:]...)
	moved := make([][2]int, len(spans))
	for k, s := range spans {
		n, _ := slices.BinarySearch(at, s[0])
		moved[k] = [2]int{s[0] + n, s[1] + n}
	}
	return out, moved
}

// proofLabel reports whether label is a proof id: the anchor package reads the id shape, from the
// token it would spell.
func proofLabel(label string) bool {
	tok := anchor.Token(label)
	return anchor.Kind(label) == "proof" && anchor.SkipRun(tok, 0) == len(tok)
}

// rawText is an inline HTML node's source.
func rawText(x *ast.RawHTML, src []byte) string {
	var b bytes.Buffer
	for i := 0; i < x.Segments.Len(); i++ {
		seg := x.Segments.At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}
