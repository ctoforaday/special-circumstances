package report

import (
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
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

// escapeNotes returns md with every `[^` outside code escaped, so that nothing in it opens a
// footnote, reading md with rd. One already escaped is left as it is. A `[^` is left as it is only
// where both tiers read it as code (noteCode): report.html's reading (rd.page) and the shipped .md's,
// read as ship reads it (shippedCode, rd.code). Where the two differ — a `[^1]` between two pairs
// of backticks after "www.a.com", the first pair GitHub's extended autolink takes into the URL, is
// prose there and code to report.html — it is escaped, which is text to both, at the cost of a
// backslash printed in the code of the tier that holds it as code.
func (rd readers) escapeNotes(md string) string {
	if !strings.Contains(md, noteOpen) {
		return md
	}
	src := []byte(md)
	return escapeNotesOutside(md, noteCode(codeIn(rd.page(src), src), shippedCode(src, rd.code)))
}

// escapeNotesInLine is escapeNotes for a fragment assembly writes on a line after the line's own
// label — a citation's title, after `[^N]: ` or `- `, and a proof's reason — each value in it made
// one line. In place the fragment is inline text of that line, so the only code in it is a code span
// it holds whole, and each tier reads it as one line of a paragraph (lineReader, codeLineReader), never
// as a document of its own: read alone, "```[^1] spec" opens a fence whose info string holds the
// `[^1]`, which in place is prose. Nothing else on its line holds a backtick — a citation's URLs are
// links that hold none (plainURL) — so a span the fragment holds is a span in place.
func escapeNotesInLine(s string) string {
	if !strings.Contains(s, noteOpen) {
		return s
	}
	src := []byte(s)
	return escapeNotesOutside(s, noteCode(codeIn(lineReader.Parse(text.NewReader(src)), src), shippedCode(src, parseCodeLine)))
}

// codeLineReader is lineReader as the shipped .md reads a line: one line of inline markdown as a
// paragraph, in codeReader's inline grammar.
var codeLineReader = parser.NewParser(
	parser.WithBlockParsers(util.Prioritized(parser.NewParagraphParser(), 1000)),
	parser.WithInlineParsers(append(codeInlineParsers(),
		util.Prioritized(extension.NewFootnoteParser(), 101))...),
)

// parseCodeLine is the shipped .md's parse of a line fragment (codeLineReader).
func parseCodeLine(src []byte) ast.Node { return codeLineReader.Parse(text.NewReader(src)) }

// shippedCode is every stretch of src that is code as the shipped .md reads it, parsed by parse: ship
// reads the text with each line ending `\n` and every `<` written ltEntity (ship), and each stretch
// that reading holds as code is mapped back to the bytes of src it was read from.
func shippedCode(src []byte, parse func([]byte) ast.Node) [][2]int {
	read := make([]byte, 0, len(src))
	back := make([]int, 0, len(src))
	for i := 0; i < len(src); i++ {
		switch c := src[i]; {
		case c == '\r' && i+1 < len(src) && src[i+1] == '\n':
		case c == '\r':
			read, back = append(read, '\n'), append(back, i)
		case c == '<':
			if escapedAt(src, i) {
				read, back = read[:len(read)-1], back[:len(back)-1]
			}
			for k := range len(ltEntity) {
				read, back = append(read, ltEntity[k]), append(back, i)
			}
		default:
			read, back = append(read, c), append(back, i)
		}
	}
	var out [][2]int
	for _, c := range codeIn(parse(read), read) {
		if c[0] < c[1] {
			out = append(out, [2]int{back[c[0]], back[c[1]-1] + 1})
		}
	}
	return out
}

// noteCode is the stretches that are code in both a and b, each in order and not overlapping: what
// one tier reads as code and the other as prose is prose to the note escape and the refusal.
func noteCode(a, b [][2]int) [][2]int {
	var out [][2]int
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if lo, hi := max(a[i][0], b[j][0]), min(a[i][1], b[j][1]); lo < hi {
			out = append(out, [2]int{lo, hi})
		}
		if a[i][1] < b[j][1] {
			i++
		} else {
			j++
		}
	}
	return out
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
