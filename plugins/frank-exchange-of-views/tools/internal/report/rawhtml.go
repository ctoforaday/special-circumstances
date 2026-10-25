package report

// WHAT BECOMES HTML IS DECIDED ONCE, ON THE WHOLE SHIPPED DOCUMENT.
//
// A shipped document is composed from units — the report's text, and each section a seat's words
// fill — and every reader of it parses it WHOLE: report.html's renderer, a markdown viewer, an
// artifact pane. A unit checked alone is parsed alone, so its fence and HTML state is not the
// state the reader meets: a fence one unit leaves open flips the next unit's code and prose, an
// indented line one unit wrote as code is HTML inside the list the unit before it ended with, and
// an unclosed `<!--` hides every section after it. Seats read untrusted web content, so a seat's
// HTML reaching a page the human opens is a prompt-injection path into a browser.
//
// So the decision is taken on the parse of the document as it ships (escapeRawHTML), and it is
// one decision for both tiers: render writes each .md through it, and mdToHTML renders through it.
// HTML stays HTML only in the shapes the tool itself writes (toolHTML):
//   - evidence.md's anchor per proof (proofAnchor), which the documents link to by label;
//   - a comment that is exactly a run of anchor tokens, which renders as nothing.
//
// Every other `<` that opens HTML — a block, an inline tag, a paragraph of nothing but tags — is
// escaped with a backslash, which every markdown reader prints as the character typed. Code is
// left as written: inside a fence or a code span nothing is HTML and a backslash would print. An
// autolink is markdown, not HTML, and stays a link. The board chart is not in any document: the
// site inserts it after rendering (RenderSite).

import (
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// proofAnchor is the anchor evidence.md writes above each proof: an explicit id, because a
// heading's generated id is a function of its whole text and the footnotes in the other documents
// link here by LABEL, which is stable.
func proofAnchor(label string) string { return `<a id="` + label + `"></a>` }

// escapeRawHTML returns md with every `<` that opens HTML the tool did not write escaped.
//
// Escaping one opener can expose another — the lines of a `<div>` block become a paragraph, and a
// `<script>` line in it opens a block of its own — so the document is parsed again until a parse
// finds nothing to escape. It ends: each pass escapes at least one `<` that opened HTML, and an
// escaped `<` opens nothing again.
func escapeRawHTML(md string) string {
	for {
		src := []byte(md)
		doc := mdReader.Parser().Parse(text.NewReader(src))
		var at []int
		_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			switch x := n.(type) {
			case *ast.HTMLBlock:
				if !toolHTML(x, src) {
					at = append(at, opener(src, x.Lines().At(0).Start))
				}
				return ast.WalkSkipChildren, nil
			case *ast.RawHTML:
				if !toolHTML(x, src) {
					at = append(at, opener(src, x.Segments.At(0).Start))
				}
			}
			return ast.WalkContinue, nil
		})
		if len(at) == 0 {
			return md
		}
		slices.Sort(at)
		var b strings.Builder
		last := 0
		for _, i := range slices.Compact(at) {
			b.Write(src[last:i])
			b.WriteByte('\\')
			last = i
		}
		b.Write(src[last:])
		md = b.String()
	}
}

// opener is the offset of the `<` that opens the HTML whose first segment starts at i: a block's
// line may begin with the indent its container leaves.
func opener(src []byte, i int) int {
	j := strings.IndexByte(string(src[i:]), '<')
	if j < 0 {
		panic("report: an HTML node with no '<' in its source")
	}
	return i + j
}

// toolHTML reports whether n — an HTML block or an inline HTML node — is HTML in a shape the tool
// writes, and so stays HTML in every tier.
func toolHTML(n ast.Node, src []byte) bool {
	switch x := n.(type) {
	case *ast.HTMLBlock:
		// A block of tokens, one run to a line: a comment block the parser closes on its own line.
		lines := strings.Split(strings.TrimSpace(rawLines(x, src)), "\n")
		for _, ln := range lines {
			if !tokenRun(strings.TrimSpace(ln)) {
				return false
			}
		}
		return true
	case *ast.RawHTML:
		if tokenRun(rawText(x, src)) {
			return true
		}
		// evidence.md's anchor is a paragraph of two inline nodes, the opening tag and the closing.
		p, ok := x.Parent().(*ast.Paragraph)
		if !ok || p.ChildCount() != 2 {
			return false
		}
		open, okOpen := p.FirstChild().(*ast.RawHTML)
		shut, okShut := p.LastChild().(*ast.RawHTML)
		if !okOpen || !okShut {
			return false
		}
		whole := rawText(open, src) + rawText(shut, src)
		label := strings.TrimSuffix(strings.TrimPrefix(whole, `<a id="`), `"></a>`)
		return proofLabel(label) && whole == proofAnchor(label)
	}
	return false
}

// tokenRun reports whether s is one or more anchor tokens and nothing else.
func tokenRun(s string) bool { return s != "" && anchor.SkipRun(s, 0) == len(s) }

// proofLabel reports whether label is a proof id: the anchor package reads the id shape, from the
// token it would spell.
func proofLabel(label string) bool {
	return anchor.Kind(label) == "proof" && tokenRun(anchor.Token(label))
}

// rawText is an inline HTML node's source.
func rawText(x *ast.RawHTML, src []byte) string {
	var b strings.Builder
	for i := 0; i < x.Segments.Len(); i++ {
		seg := x.Segments.At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}

// rawLines is an HTML block's source, its closing line with it.
func rawLines(x *ast.HTMLBlock, src []byte) string {
	var b strings.Builder
	for i := 0; i < x.Lines().Len(); i++ {
		seg := x.Lines().At(i)
		b.Write(seg.Value(src))
	}
	if x.HasClosure() {
		b.Write(x.ClosureLine.Value(src))
	}
	return b.String()
}
