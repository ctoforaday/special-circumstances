package report

import (
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// WHICH FOOTNOTES A DOCUMENT REFERENCES, AND WHICH IT DEFINES, AS A MARKDOWN READER READS THEM.
//
// The report set is markdown, and what is a footnote in it — not code, not an escaped bracket, a
// definition only where a definition can stand — is markdown's grammar, read here by a markdown
// parser (CommonMark with GitHub's footnotes). A reference the parser resolved is a link to its
// definition; a reference with no definition is left as text, which is what every reader shows:
// the label's own source characters. So a dangling reference is a `[^label]` in the text the
// parser left as prose — never inside code, never escaped — and is read from that text alone.

// footnoteReader is CommonMark with GitHub's tables and footnotes, the dialect the set is written in.
var footnoteReader = goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote))

// FootnotesIn reads md's footnotes: how many distinct labels it references, how many of those
// it defines, and the labels it references with no definition in md, sorted.
func FootnotesIn(md string) (refs, defs int, dangling []string) {
	src := []byte(md)
	doc := footnoteReader.Parser().Parse(text.NewReader(src))
	labelOf := map[int]string{} // a definition's index -> its label
	var links []int
	unresolved := map[string]bool{}
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *east.Footnote:
			labelOf[x.Index] = string(x.Ref)
		case *east.FootnoteLink:
			links = append(links, x.Index)
		case *ast.CodeSpan, *ast.FencedCodeBlock, *ast.CodeBlock, *ast.HTMLBlock, *ast.RawHTML:
			return ast.WalkSkipChildren, nil
		}
		for _, run := range textRuns(n, src) {
			for _, label := range unresolvedRefs(run) {
				unresolved[label] = true
			}
		}
		return ast.WalkContinue, nil
	})
	referenced := map[string]bool{}
	for _, i := range links {
		referenced[strings.ToLower(labelOf[i])] = true
	}
	for label := range unresolved {
		referenced[strings.ToLower(label)] = true
		dangling = append(dangling, label)
	}
	sort.Strings(dangling)
	return len(referenced), len(referenced) - len(unresolved), dangling
}

// textRuns is n's direct text children as the source wrote them, one string per run of adjacent
// text — a link, an emphasis or a code span between two runs ends the first — with a line break
// kept where the source broke the line.
func textRuns(n ast.Node, src []byte) []string {
	var runs []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			runs = append(runs, b.String())
			b.Reset()
		}
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		t, ok := c.(*ast.Text)
		if !ok {
			flush()
			continue
		}
		b.Write(t.Segment.Value(src))
		if t.SoftLineBreak() || t.HardLineBreak() {
			b.WriteByte('\n')
		}
	}
	flush()
	return runs
}

// unresolvedRefs is each `[^label]` a run of prose holds that is not escaped. The parser has
// already turned every reference it could resolve into a link, so what remains in prose resolves
// to nothing.
func unresolvedRefs(run string) (labels []string) {
	for i := 0; ; {
		j := strings.Index(run[i:], noteOpen)
		if j < 0 {
			return labels
		}
		at := i + j
		i = at + len(noteOpen)
		if NoteEscapedAt(run, at) {
			continue
		}
		end := strings.IndexAny(run[i:], "]\n[")
		if end <= 0 || run[i+end] != ']' {
			continue
		}
		labels = append(labels, run[i:i+end])
	}
}
