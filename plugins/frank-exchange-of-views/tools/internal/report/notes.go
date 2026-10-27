package report

import (
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
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
// the notes and Bibliography entries the weave writes) — `\[^P5]`, which every markdown reader prints as the five
// characters the seat typed — and the weave, which runs after, is then the one writer of an
// unescaped `[^` in a tool-composed block. The report's own text is not escaped: a footnote under
// a label that is not the weave's is the report's to write (WovenNoteIn refuses the others at the
// write), and it ships as written.
//
// CODE IS LEFT AS IT IS. Inside a fence or a code span nothing is a reference and a backslash is a
// character, so an escape there would print.

// noteOpen is what opens a footnote reference or definition, and noteEscape is that with the
// backslash that makes it text.
const (
	noteOpen   = "[^"
	noteEscape = `\` + noteOpen
)

// EscapeNotes returns md with every `[^` outside code escaped, so that nothing in it opens a
// footnote. One already escaped is left as it is. What is code is report.html's reading (pageReader,
// codeIn), which the shipped .md shares (ship): a `[^` is code or prose in the same place in all
// three, so a line an HTML block would hold as code to a reader with the grammar is prose here, as it
// is to both shipped readings.
func EscapeNotes(md string) string { return stdReaders().escapeNotes(md) }

// escapeNotes is EscapeNotes, reading md with rd.
func (rd readers) escapeNotes(md string) string {
	if !strings.Contains(md, noteOpen) {
		return md
	}
	return escapeNotesOutside(md, codeIn(rd.page([]byte(md)), []byte(md)))
}

// escapeNotesInLine is EscapeNotes for a fragment assembly splices into the middle of a line it
// writes — a citation's title, a proof's reason, each made one line — after a note's `[^N]: `. In
// place the fragment is inline text of that line, so the only code in it is a code span it holds
// whole, and it is read as one line of a paragraph (lineReader), never as a document of its own: read
// alone, "```[^1] spec" opens a fence whose info string holds the `[^1]`, which in place is prose.
// What stands before the fragment on its line holds no backtick, so a span the fragment holds is a
// span in place.
func escapeNotesInLine(s string) string {
	if !strings.Contains(s, noteOpen) {
		return s
	}
	src := []byte(s)
	return escapeNotesOutside(s, codeIn(lineReader.Parse(text.NewReader(src)), src))
}

// escapeNotesOutside escapes each `[^` of md outside code, the stretches of md that are code in order.
func escapeNotesOutside(md string, code [][2]int) string {
	var b strings.Builder
	last := 0
	for _, c := range code {
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
		if !escapedAt(s, i) {
			b.WriteByte('\\')
		}
		b.WriteString(noteOpen)
		s = s[i+len(noteOpen):]
	}
}

// escapedAt reports whether the byte at i of s is escaped, and so is text: an odd run of backslashes
// stands before it. It is the one reading of the escape, for a `[^` (escapeNoteOpens, noteRefsIn)
// and for a `<` (ship).
func escapedAt[T ~string | ~[]byte](s T, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
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
