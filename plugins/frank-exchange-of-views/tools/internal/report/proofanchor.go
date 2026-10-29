package report

// A PROOF'S ANCHOR IS THE TOOL'S BY CONSTRUCTION, NOT BY SHAPE.
//
// evidence.md carries an element per proof that the footnotes in the other documents link to by
// LABEL. The same tag typed by a seat would give the page a second element of that id, and the
// report's tab comes first, so every footnote linking to the proof would land on the seat's copy.
// So the anchor is never recognised from its text: evidenceDoc writes a mark no record text can
// hold (anchorMark), and each tier turns only that mark into an anchor — report.html through the
// parser's own node for it (proofAnchorNode), the .md by resolveMarks. Anything else shaped like the
// anchor is a seat's text, and both tiers show it as typed.

import (
	"bytes"
	"crypto/rand"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// proofAnchor is the anchor evidence.md ships above each proof: an explicit id, because a heading's
// generated id is a function of its whole text and the footnotes in the other documents link here by
// LABEL, which is stable.
func proofAnchor(label string) string { return `<a id="` + label + `"></a>` }

// anchorMark opens the line evidenceDoc writes where a proof's anchor goes. It is drawn for the
// process, after every seat wrote, so no record text holds it: a mark is the tool's by construction.
var anchorMark = "\x1eproof-anchor:" + rand.Text() + ":"

// proofAnchorMark is what evidenceDoc writes, on a line of its own, for the proof labelled label.
func proofAnchorMark(label string) string { return anchorMark + label }

// resolveMarks returns md with each mark made the anchor it stands for (resolveMarkSpans).
func resolveMarks(md string) string {
	md, _ = resolveMarkSpans(md)
	return md
}

// resolveMarkSpans returns md with each mark made the anchor it stands for, and the span of each
// anchor: all of them the tool's. A label's anchor is written once — a second proof under one label
// adds no second element of that id — and a mark whose label is no proof id writes nothing.
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
		if proofLabel(label) && !seen[label] {
			seen[label] = true
			spans = append(spans, [2]int{b.Len(), b.Len() + len(proofAnchor(label))})
			b.WriteString(proofAnchor(label))
		}
	}
}

// kindProofAnchor is the node kind of a proof's anchor in the tree report.html renders.
var kindProofAnchor = ast.NewNodeKind("ProofAnchor")

// proofAnchorNode is a proof's anchor: a mark line the tool wrote, read by proofAnchorParser.
type proofAnchorNode struct {
	ast.BaseBlock
	label string
}

func (n *proofAnchorNode) Kind() ast.NodeKind { return kindProofAnchor }

func (n *proofAnchorNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"label": n.label}, nil)
}

// proofAnchorParser reads a line that opens with anchorMark as the proof's anchor it marks. A line
// that merely looks like the anchor is HTML in the text, and the renderer shows that as typed.
type proofAnchorParser struct{}

func (proofAnchorParser) Trigger() []byte { return []byte{anchorMark[0]} }

func (proofAnchorParser) Open(_ ast.Node, reader text.Reader, _ parser.Context) (ast.Node, parser.State) {
	line, seg := reader.PeekLine()
	label, ok := bytes.CutPrefix(bytes.TrimLeft(line, " \t"), []byte(anchorMark))
	if !ok {
		return nil, parser.NoChildren
	}
	reader.Advance(seg.Len() - 1)
	return &proofAnchorNode{label: string(bytes.TrimRight(label, "\r\n"))}, parser.NoChildren
}

func (proofAnchorParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

func (proofAnchorParser) Close(ast.Node, text.Reader, parser.Context) {}
func (proofAnchorParser) CanInterruptParagraph() bool                 { return true }
func (proofAnchorParser) CanAcceptIndentedLine() bool                 { return false }
