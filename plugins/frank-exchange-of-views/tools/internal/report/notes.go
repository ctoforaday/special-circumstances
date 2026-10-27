package report

import (
	"regexp"
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// ONLY THE WEAVE OPENS A FOOTNOTE IN WHAT ASSEMBLY COMPOSES.
//
// A footnote reference is two characters and a label — `[^P5]` — and a seat types those in prose
// for its own reasons: a finding that names the report's footnote ("the one footnote [^P5] cites"),
// a log entry that quotes a regex (`grep -o '[^<>]{0,120}'`). Assembly printed that text as the
// seat wrote it, into documents whose footnotes are the tool's, and markdown has one reading of
// `[^x]`: in the board document the finding's words became a reference to a note the document does
// not define, and report.html made them a citation link to nowhere. The capture audit that reads
// each document for references with no definition reported both, correctly.
//
// So text assembly composes is escaped where it is composed (Section.block, and the three places
// outside the table that print a seat's words: the inquiries document, the evidence document, and
// the notes the weave writes) — `\[^P5]`, which every markdown reader prints as the five
// characters the seat typed — and the weave, which runs after, is then the one writer of an
// unescaped `[^` in a tool-composed block. The report's own text is not escaped: a footnote under
// a label that is not the weave's is the report's to write (WovenNoteIn refuses the others at the
// write), and it ships as written.
//
// CODE IS LEFT AS IT IS. Inside a fence or a code span nothing is a reference and a backslash is a
// character, so an escape there would print.

// reCodeSpan is a code span as the refusal of a woven label in the report's text reads one
// (WovenNoteIn). Escaping and rendering read code from the markdown parser's tree (codeIn).
var reCodeSpan = regexp.MustCompile("`[^`]+`")

// noteOpen is what opens a footnote reference or definition, and noteEscape is that with the
// backslash that makes it text.
const (
	noteOpen   = "[^"
	noteEscape = `\` + noteOpen
)

// EscapeNotes returns md with every `[^` outside code escaped, so that nothing in it opens a
// footnote. One already escaped is left as it is. What is code is the parser's reading (codeIn): a
// fence of four backticks holds a line of three, which a scan of fence lines takes for its close.
func EscapeNotes(md string) string {
	if !strings.Contains(md, noteOpen) {
		return md
	}
	var b strings.Builder
	last := 0
	for _, c := range codeIn(parseMD([]byte(md)), []byte(md)) {
		b.WriteString(escapeNoteOpens(md[last:c[0]]))
		b.WriteString(md[c[0]:c[1]])
		last = c[1]
	}
	b.WriteString(escapeNoteOpens(md[last:]))
	return b.String()
}

// escapeNoteOpens escapes each `[^` of a stretch of prose that holds no code.
func escapeNoteOpens(s string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, noteOpen)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		if !NoteEscapedAt(s, i) {
			b.WriteByte('\\')
		}
		b.WriteString(noteOpen)
		s = s[i+len(noteOpen):]
	}
}

// NoteEscapedAt reports whether the `[^` at byte i of md is escaped, and so is text: the one
// reading of the escape, for the reader that renders a document and the audit that counts its
// references.
func NoteEscapedAt(md string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && md[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

// codeIn is every stretch of src the tree reads as code, in order: a fence's info string, a code
// block's lines and a code span's content.
func codeIn(tree ast.Node, src []byte) [][2]int {
	var code [][2]int
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			if f, ok := x.(*ast.FencedCodeBlock); ok && f.Info != nil {
				code = append(code, [2]int{f.Info.Segment.Start, f.Info.Segment.Stop})
			}
			for i := 0; i < x.Lines().Len(); i++ {
				seg := x.Lines().At(i)
				code = append(code, [2]int{seg.Start, seg.Stop})
			}
			return ast.WalkSkipChildren, nil
		case *ast.CodeSpan:
			first, last := x.FirstChild(), x.LastChild()
			if first != nil {
				code = append(code, [2]int{first.(*ast.Text).Segment.Start, last.(*ast.Text).Segment.Stop})
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	slices.SortFunc(code, func(a, b [2]int) int { return a[0] - b[0] })
	return code
}
