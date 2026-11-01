package report

// A COMPUTATION'S ANCHOR IS THE TOOL'S BY CONSTRUCTION, NOT BY SHAPE.
//
// evidence.md carries an element per computation that the footnotes in the other documents link to by
// LABEL. The same tag typed by a seat would give the page a second element of that id, and the
// report's tab comes first, so every footnote linking to the computation would land on the seat's copy.
// So the anchor is never recognised from its text: evidenceDoc writes a mark no record text can
// hold (anchorMark), and each tier turns only that mark into an anchor — report.html through the
// parser's own node for it (computationAnchorNode), the .md by resolveMarks. Anything else shaped like the
// anchor is a seat's text, and both tiers show it as typed.

import (
	"bytes"
	"crypto/rand"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// computationAnchor is the anchor evidence.md ships above each computation: an explicit id, because a heading's
// generated id is a function of its whole text and the footnotes in the other documents link here by
// LABEL, which is stable.
func computationAnchor(label string) string { return `<a id="` + label + `"></a>` }

// anchorMark opens the line evidenceDoc writes where a computation's anchor goes. It is drawn for the
// process, after every seat wrote, so no record text holds it: a mark is the tool's by construction.
var anchorMark = "\x1ecomputation-anchor:" + rand.Text() + ":"

// computationAnchorMark is what evidenceDoc writes, on a line of its own, for the computation labelled label.
func computationAnchorMark(label string) string { return anchorMark + label }

// resolveMarks returns md with each mark made the anchor it stands for (resolveMarkSpans).
func resolveMarks(md string) string {
	md, _ = resolveMarkSpans(md)
	return md
}

// resolveMarkSpans returns md with each mark made the anchor it stands for, and the span of each
// anchor: all of them the tool's. A label's anchor is written once — a second computation under one label
// adds no second element of that id — and a mark whose label is no computation id writes nothing.
func resolveMarkSpans(md string) (string, [][2]int) {
	if !strings.Contains(md, anchorMark) {
		return md, nil
	}
	var b strings.Builder
	var spans [][2]int
	seen := map[string]bool{}
	for {
		i := strings.Index(md, anchorMark)
		if i < 0 {
			b.WriteString(md)
			return b.String(), spans
		}
		b.WriteString(md[:i])
		md = md[i+len(anchorMark):]
		end := strings.IndexByte(md, '\n')
		if end < 0 {
			end = len(md)
		}
		label := md[:end]
		md = md[end:]
		if computationLabel(label) && !seen[label] {
			seen[label] = true
			spans = append(spans, [2]int{b.Len(), b.Len() + len(computationAnchor(label))})
			b.WriteString(computationAnchor(label))
		}
	}
}

// kindComputationAnchor is the node kind of a computation's anchor in the tree report.html renders.
var kindComputationAnchor = ast.NewNodeKind("ComputationAnchor")

// computationAnchorNode is a computation's anchor: a mark line the tool wrote, read by computationAnchorParser.
type computationAnchorNode struct {
	ast.BaseBlock
	label string
}

func (n *computationAnchorNode) Kind() ast.NodeKind { return kindComputationAnchor }

func (n *computationAnchorNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"label": n.label}, nil)
}

// computationAnchorParser reads a line that opens with anchorMark as the computation's anchor it marks. A line
// that merely looks like the anchor is HTML in the text, and the renderer shows that as typed.
type computationAnchorParser struct{}

func (computationAnchorParser) Trigger() []byte { return []byte{anchorMark[0]} }

func (computationAnchorParser) Open(_ ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	line, seg := reader.PeekLine()
	label, ok := bytes.CutPrefix(bytes.TrimLeft(line, " \t"), []byte(anchorMark))
	if !ok {
		return nil, parser.NoChildren
	}
	reader.Advance(seg.Len() - 1)
	return &computationAnchorNode{label: string(bytes.TrimRight(label, "\r\n"))}, parser.NoChildren
}

func (computationAnchorParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

func (computationAnchorParser) Close(ast.Node, text.Reader, parser.Context) {}
func (computationAnchorParser) CanInterruptParagraph() bool                 { return true }
func (computationAnchorParser) CanAcceptIndentedLine() bool                 { return false }
