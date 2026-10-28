package report

// MARKDOWN → HTML, FROM THE TREE A MARKDOWN READER BUILDS.
//
// The documents are markdown, and what each line of one is — a heading, a list item, a fence, a
// table, a footnote, prose — is markdown's grammar, read by a markdown parser (pageReader:
// CommonMark with GitHub's extensions and footnotes, and no HTML grammar). This file renders that
// tree; it reads no line of the text itself. The rendering is the reading tier's own — the ids, the
// tab links, the numbered citations — and the tree is the one a markdown reader of the shipped .md
// file sees but for a `<` autolink or angle-bracket destination, which the .md ships as text (ship
// writes every `<` outside code `&lt;`), so the two tiers do not disagree about what a line is.
//
// WHAT IT DOES NOT DO IS AS IMPORTANT: it does not round-trip. The markdown files are the
// durable tier and are written by the composers directly; this reads them ONCE, forward, into
// a rendering. Nothing here can change what ships in a .md.
//
// HTML IN THE TEXT IS TEXT. The renderer passes no HTML through: an inline tag, a comment, an HTML
// block are each shown as typed, whoever wrote them. The one element the tool places in a document
// that markdown cannot say — a proof's anchor — is a node only the tool can produce (proofanchor.go),
// so a seat's copy of it is HTML in the text like any other.
//
// EVERY VALUE IS WRITTEN FOR ITS CONTEXT. The renderer builds its HTML by hand, so each value it
// places is escaped where it is placed: text and attributes by escape, a link's URL decoded, checked
// and encoded (linkURL), a footnote label in an id or a fragment as noteID. RenderSite then holds
// every fragment of the page that holds the record's text — the title, the badges, each body — to an
// allowlist of exactly what this file writes (pageHTML, bodyPolicy).

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
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

var (
	// pageReader is the dialect report.html reads a document in: CommonMark with GitHub's extensions
	// and footnotes, and no HTML grammar, so no node of the tree is HTML and a tag, a comment or an
	// unclosed `<!--` is text in the paragraph it stands in — it cannot take the lines after it into
	// an HTML block — and with the proof's anchor, the one node only the tool's mark produces. The
	// References list keeps a definition nothing references (citations).
	pageReader = goldmark.New(goldmark.WithParser(parser.NewParser(
		parser.WithBlockParsers(append(blockParsers(),
			util.Prioritized(extension.NewFootnoteBlockParser(), 999), util.Prioritized(proofAnchorParser{}, 100))...),
		parser.WithInlineParsers(append(inlineParsers(),
			util.Prioritized(extension.NewFootnoteParser(), 101))...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)), goldmark.WithExtensions(extension.GFM))
	// lineReader reads one line of inline markdown — a title — as a paragraph, whatever it opens
	// with: a title that starts "1." is a title, not a list. It has pageReader's inline grammar.
	lineReader = parser.NewParser(
		parser.WithBlockParsers(util.Prioritized(parser.NewParagraphParser(), 1000)),
		parser.WithInlineParsers(append(inlineParsers(),
			util.Prioritized(extension.NewLinkifyParser(), 999),
			util.Prioritized(extension.NewStrikethroughParser(), 500))...),
	)
	// idToken matches the identifiers the record mints and every document quotes: gap, motion
	// and finding ids, and a proof's footnote number (P1). They are the joins the single-file report
	// made a reader scroll for. An id that matches nothing renders exactly like an id nobody else
	// mentioned, so the id shapes are read from the kinds table, never restated.
	idToken = regexp.MustCompile(`\b(` + anchor.IDPattern("gap", "motion", "finding") + `|P\d+)\b`)
)

// CommonMark's grammar is goldmark's default parsers, and its HTML grammar is two of them: the HTML
// block parser and the inline HTML parser. The readers here are built from the rest, named one by one
// from goldmark's constructors at goldmark's priorities, so no parser reaches a reader by being in a
// default list: an upgrade that adds or changes a default parser changes none of them, and
// TestTheGrammarIsGoldmarksDefaultsLessHTML fails until these lists say what to do with it.
const (
	htmlBlockPriority = 900 // parser.NewHTMLBlockParser's place among the block parsers
	rawHTMLPriority   = 400 // parser.NewRawHTMLParser's among the inline parsers
)

// blockParsers is CommonMark's block grammar without its HTML block parser.
func blockParsers() []util.PrioritizedValue {
	return []util.PrioritizedValue{
		util.Prioritized(parser.NewSetextHeadingParser(), 100),
		util.Prioritized(parser.NewThematicBreakParser(), 200),
		util.Prioritized(parser.NewListParser(), 300),
		util.Prioritized(parser.NewListItemParser(), 400),
		util.Prioritized(parser.NewCodeBlockParser(), 500),
		util.Prioritized(parser.NewATXHeadingParser(), 600),
		util.Prioritized(parser.NewFencedCodeBlockParser(), 700),
		util.Prioritized(parser.NewBlockquoteParser(), 800),
		util.Prioritized(parser.NewParagraphParser(), 1000),
	}
}

// inlineParsers is CommonMark's inline grammar without its inline HTML parser.
func inlineParsers() []util.PrioritizedValue {
	return []util.PrioritizedValue{
		util.Prioritized(parser.NewCodeSpanParser(), 100),
		util.Prioritized(parser.NewLinkParser(), 200),
		util.Prioritized(parser.NewAutoLinkParser(), 300),
		util.Prioritized(parser.NewEmphasisParser(), 500),
	}
}

// parsePage is report.html's one parse of a document (pageReader).
func parsePage(src []byte) ast.Node { return pageReader.Parser().Parse(text.NewReader(src)) }

// anchors is the map from a record id to the document and element that DEFINES it, filled as
// each document renders and applied across the set afterwards.
type anchors map[string]string

// mention is one id the text names, where linkIDs will link it: the id, and the id of the element
// it stands in, so an id is never linked at the site that defines it.
type mention struct{ id, here string }

// rendered is one document rendered, its id mentions held for the pass that links them.
type rendered struct {
	html     string
	mentions []mention
}

// renderDoc renders one document from the tree shipSet parsed it to. It records every anchor it
// defines into anchor, keyed by the record id, valued as "<file>#<elementID>".
//
// seen counts the ids given out, ACROSS THE SET: the documents are tabs of one page, a fragment
// resolves to the first element of that id on the page, and a hidden tab's element is no place
// to scroll to. The report's own "## Proofs" and evidence.md's are "proofs" and "proofs-2", and
// the note each of two documents numbers 1 is "fn-1" in the first and "fn-1-2" in the second
// (citations).
//
// files is the document set: a link to one of them is a tab switch, written as one here, because
// in a single page there is no sibling file to open.
func renderDoc(p ParsedText, file string, anchor anchors, seen map[string]int, files map[string]bool) rendered {
	r := &mdRender{src: p.src, file: file, anchor: anchor, seen: seen, files: files, link: true, anchored: map[string]bool{}}
	// THE CITATION LAYER IS HELD, NOT RE-READ. Each reference is recorded where the tree holds it
	// and left in the text as a marker that holds only its index (refMark); each definition is kept
	// as a note. citations numbers and lists them from these once the document is rendered.
	r.ref = func(label string) string {
		r.refs = append(r.refs, label)
		return refMark(len(r.refs) - 1)
	}
	var b strings.Builder
	r.blocks(&b, p.tree)
	return rendered{html: citations(b.String(), r.refs, r.notes, seen), mentions: r.mentions}
}

// unmarked is md with the marker bytes made the replacement character. The markers are control
// characters with no rendering, so no text can be read as a marker; NUL is too, as CommonMark has it.
func unmarked(md string) string {
	if !strings.ContainsAny(md, string([]byte{refMarkByte, mentionMarkByte, 0})) {
		return md
	}
	return strings.NewReplacer(string(refMarkByte), "\uFFFD", string(mentionMarkByte), "\uFFFD", "\x00", "\uFFFD").Replace(md)
}

// mdRender is one document's rendering in progress.
type mdRender struct {
	src    []byte
	file   string
	anchor anchors
	seen   map[string]int
	files  map[string]bool
	ref    func(label string) string // what a footnote reference renders as
	link   bool                      // whether the ids the text names are held for linkIDs
	// looseRefs reads a footnote reference in the text (noteRefsIn), where no definition is in
	// reach for the parser to resolve it to: one line of a title.
	looseRefs bool
	// anchored is the proofs whose anchor this document has placed: one element per label.
	anchored map[string]bool

	refs     []string
	notes    []footnote
	mentions []mention
	here     string // the id of the element being rendered
}

// blocks renders n's block children.
func (r *mdRender) blocks(b *strings.Builder, n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		r.block(b, c)
	}
}

func (r *mdRender) block(b *strings.Builder, n ast.Node) {
	switch x := n.(type) {
	case *ast.Heading:
		text := r.plain(x)
		id := r.slug(text)
		r.define(text, id)
		fmt.Fprintf(b, "<h%d id=\"%s\">%s</h%d>\n", x.Level, id, r.inlineIn(x, id), x.Level)
	case *ast.Paragraph:
		b.WriteString("<p>" + r.inlineIn(x, "") + "</p>\n")
	case *ast.TextBlock:
		b.WriteString(r.inlineIn(x, "") + "\n")
	case *ast.ThematicBreak:
		b.WriteString("<hr>\n")
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		// Taken verbatim, escaped, never read: a "## " or a "|" inside a script is not a heading
		// and not a table.
		cls := ""
		if f, ok := x.(*ast.FencedCodeBlock); ok {
			if lang := f.Language(r.src); lang != nil {
				cls = ` class="language-` + escape(string(lang)) + `"`
			}
		}
		b.WriteString("<pre><code" + cls + ">" + escape(strings.TrimSuffix(r.lines(x), "\n")) + "</code></pre>\n")
	case *proofAnchorNode:
		// A proof's anchor, which only evidenceDoc's mark produces: once per label, and nothing for
		// a label that is no proof id.
		if proofLabel(x.label) && !r.anchored[x.label] {
			r.anchored[x.label] = true
			b.WriteString(proofAnchor(x.label) + "\n")
		}
	case *ast.Blockquote:
		if p, ok := x.FirstChild().(*ast.Paragraph); ok && x.ChildCount() == 1 {
			b.WriteString("<blockquote>" + r.inlineIn(p, "") + "</blockquote>\n")
			return
		}
		b.WriteString("<blockquote>\n")
		r.blocks(b, x)
		b.WriteString("</blockquote>\n")
	case *ast.List:
		tag := "ul"
		if x.IsOrdered() {
			tag = "ol"
		}
		b.WriteString("<" + tag)
		if x.IsOrdered() && x.Start != 1 {
			fmt.Fprintf(b, ` start="%d"`, x.Start)
		}
		b.WriteString(">")
		for c := x.FirstChild(); c != nil; c = c.NextSibling() {
			r.item(b, c, !x.IsOrdered())
		}
		b.WriteString("</" + tag + ">\n")
	case *east.Table:
		// The wrapper is what scrolls: a wide table inside it pans on a phone instead of
		// stretching the whole page and dragging every paragraph with it.
		b.WriteString("<div class=\"tblwrap\"><table>")
		for row := x.FirstChild(); row != nil; row = row.NextSibling() {
			cell := "td"
			if _, head := row.(*east.TableHeader); head {
				cell = "th"
				b.WriteString("<thead>")
			}
			b.WriteString("<tr>")
			for c := row.FirstChild(); c != nil; c = c.NextSibling() {
				b.WriteString("<" + cell + ">" + r.inlineIn(c, "") + "</" + cell + ">")
			}
			if cell == "th" {
				b.WriteString("</tr></thead><tbody>\n")
			} else {
				b.WriteString("</tr>\n")
			}
		}
		b.WriteString("</tbody></table></div>\n")
	case *east.FootnoteList:
		// The definitions, which citations lists at the document's end, each under its label.
		for c := x.FirstChild(); c != nil; c = c.NextSibling() {
			fn := c.(*east.Footnote)
			note := footnote{label: string(fn.Ref), html: r.noteBody(fn)}
			title := *r
			title.ref, title.link = supRef, false
			note.scent = scentOf(title.noteBody(fn))
			r.notes = append(r.notes, note)
		}
	default:
		// A block this renderer has no element for is shown as its text, never dropped.
		b.WriteString("<p>" + escape(r.lines(n)) + "</p>\n")
	}
}

// item renders one list item. An item of a bullet list that opens with a record id is where that
// id is defined, and carries the element id.
func (r *mdRender) item(b *strings.Builder, n ast.Node, bullet bool) {
	id := ""
	if first := n.FirstChild(); bullet && first != nil {
		if tok := leadingID(r.plain(first)); tok != "" {
			id = r.slug(tok)
			r.define(tok, id)
		}
	}
	if id != "" {
		b.WriteString(`<li id="` + id + `">`)
	} else {
		b.WriteString("<li>")
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *ast.TextBlock:
			b.WriteString(r.inlineIn(c, id))
		case *ast.Paragraph:
			if c == n.FirstChild() {
				b.WriteString(r.inlineIn(c, id))
			} else {
				b.WriteString("<p>" + r.inlineIn(c, id) + "</p>")
			}
		default:
			r.block(b, c)
		}
	}
	b.WriteString("</li>\n")
}

// noteBody renders a footnote definition's content: its one paragraph as a line, or its blocks.
func (r *mdRender) noteBody(fn *east.Footnote) string {
	if p, ok := fn.FirstChild().(*ast.Paragraph); ok && fn.ChildCount() == 1 {
		return r.inlineIn(p, "")
	}
	var b strings.Builder
	r.blocks(&b, fn)
	return strings.TrimSuffix(b.String(), "\n")
}

// lines is a block's source lines, as written.
func (r *mdRender) lines(n ast.Node) string {
	var b strings.Builder
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		b.Write(seg.Value(r.src))
	}
	return b.String()
}

// define records "this element is where id lives", first definition winning: a heading that
// introduces G3 is its home, and the twenty later mentions are references to it.
func (r *mdRender) define(text, elemID string) {
	if tok := leadingID(text); tok != "" {
		if _, ok := r.anchor[tok]; !ok {
			r.anchor[tok] = r.file + "#" + elemID
		}
	}
}

// leadingID is the record id text opens with, or "".
func leadingID(text string) string {
	text = strings.TrimSpace(text)
	if loc := idToken.FindStringIndex(text); loc != nil && loc[0] == 0 {
		return text[:loc[1]]
	}
	return ""
}

// slug is an element id for text: its letters and digits, lower-cased, each other run one hyphen.
func (r *mdRender) slug(text string) string {
	var b strings.Builder
	hyphen := false
	for _, c := range strings.ToLower(text) {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			if hyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(c)
			hyphen = false
		} else {
			hyphen = true
		}
	}
	s := b.String()
	if s == "" {
		s = "section"
	}
	return uniqueID(r.seen, s)
}

// plain is the text n shows: what a heading id and an id at an element's start are read from.
func (r *mdRender) plain(n ast.Node) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := c.(type) {
		case *ast.Text:
			b.Write(textValue(x.Segment.Value(r.src)))
			if x.SoftLineBreak() || x.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(x.Value)
		case *ast.AutoLink:
			b.Write(x.Label(r.src))
		case *east.FootnoteLink:
			// The reference reads as written, which is what a heading's id has always been made of.
			b.WriteString(noteOpen + r.footnoteLabel(x) + "]")
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// inlineIn renders n's inline content, as the element whose id is here.
func (r *mdRender) inlineIn(n ast.Node, here string) string {
	was := r.here
	r.here = here
	defer func() { r.here = was }()
	var b strings.Builder
	r.inlines(&b, n, r.link)
	return b.String()
}

// inlines renders n's inline children. Runs of adjacent text are read as one, so an id the text
// names is found whole wherever the parser cut the run; link is whether those ids are held for
// linkIDs, which inside a link or code they are not.
func (r *mdRender) inlines(b *strings.Builder, n ast.Node, link bool) {
	var run strings.Builder // the run's source, escapes and all
	flush := func() {
		s := run.String()
		run.Reset()
		if !r.looseRefs {
			r.text(b, string(textValue([]byte(s))), link)
			return
		}
		last := 0
		for _, sp := range noteRefsIn(s) {
			r.text(b, string(textValue([]byte(s[last:sp[0]]))), link)
			b.WriteString(r.ref(s[sp[0]+len(noteOpen) : sp[1]-1]))
			last = sp[1]
		}
		r.text(b, string(textValue([]byte(s[last:]))), link)
	}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			run.Write(t.Segment.Value(r.src))
			switch {
			case t.HardLineBreak():
				flush()
				b.WriteString("<br> ")
			case t.SoftLineBreak():
				run.WriteByte(' ')
			}
			continue
		}
		flush()
		switch x := c.(type) {
		case *ast.String:
			r.text(b, string(x.Value), link)
		case *ast.CodeSpan:
			var code strings.Builder
			for t := x.FirstChild(); t != nil; t = t.NextSibling() {
				v := t.(*ast.Text).Segment.Value(r.src)
				code.Write(v)
				if t.(*ast.Text).SoftLineBreak() {
					code.WriteByte(' ')
				}
			}
			b.WriteString("<code>" + escape(strings.ReplaceAll(code.String(), "\n", " ")) + "</code>")
		case *ast.Emphasis:
			tag := "em"
			if x.Level >= 2 {
				tag = "strong"
			}
			b.WriteString("<" + tag + ">")
			r.inlines(b, x, link)
			b.WriteString("</" + tag + ">")
		case *east.Strikethrough:
			b.WriteString("<del>")
			r.inlines(b, x, link)
			b.WriteString("</del>")
		case *ast.Link:
			r.linkTo(b, linkURL(x.Destination), func() { r.inlines(b, x, false) })
		case *ast.AutoLink:
			url := x.URL(r.src)
			if x.AutoLinkType == ast.AutoLinkEmail {
				url = append([]byte("mailto:"), url...)
			}
			r.linkTo(b, linkURL(url), func() { b.WriteString(escape(string(x.Label(r.src)))) })
		case *ast.Image:
			// An image is shown as a link to it: the page loads nothing the text names.
			r.linkTo(b, linkURL(x.Destination), func() { r.inlines(b, x, false) })
		case *east.FootnoteLink:
			b.WriteString(r.ref(r.footnoteLabel(x)))
		case *east.TaskCheckBox:
			if x.IsChecked {
				b.WriteString("[x] ")
			} else {
				b.WriteString("[ ] ")
			}
		default:
			r.inlines(b, c, link)
		}
	}
	flush()
}

// linkTo writes a link to url around what text writes: a tab switch to a document of the set
// (siteLink), a link to anything else, and, where url is empty — a destination linkURL refuses, or
// none — the text alone, as the allowlist would leave it.
func (r *mdRender) linkTo(b *strings.Builder, url string, text func()) {
	if url == "" {
		text()
		return
	}
	b.WriteString(siteLink(url, r.files))
	text()
	b.WriteString("</a>")
}

// footnoteLabel is the label of the definition a reference resolved to.
func (r *mdRender) footnoteLabel(l *east.FootnoteLink) string {
	for p := ast.Node(l); p != nil; p = p.Parent() {
		if _, ok := p.(*ast.Document); ok {
			for c := p.FirstChild(); c != nil; c = c.NextSibling() {
				if list, ok := c.(*east.FootnoteList); ok {
					for fn := list.FirstChild(); fn != nil; fn = fn.NextSibling() {
						if f := fn.(*east.Footnote); f.Index == l.Index {
							return string(f.Ref)
						}
					}
				}
			}
		}
	}
	panic("report: a footnote reference the parser resolved to no definition")
}

// text writes a run of prose, escaped, each record id it names held as a mention when link is set.
func (r *mdRender) text(b *strings.Builder, s string, link bool) {
	if !link {
		b.WriteString(escape(s))
		return
	}
	last := 0
	for _, loc := range idToken.FindAllStringIndex(s, -1) {
		b.WriteString(escape(s[last:loc[0]]))
		r.mentions = append(r.mentions, mention{id: s[loc[0]:loc[1]], here: r.here})
		b.WriteString(mark(mentionMarkByte, len(r.mentions)-1))
		last = loc[1]
	}
	b.WriteString(escape(s[last:]))
}

// textValue is a text node's characters as a reader sees them: backslash escapes taken out and
// character references resolved.
func textValue(v []byte) []byte {
	v = util.UnescapePunctuations(v)
	v = util.ResolveNumericReferences(v)
	v = util.ResolveEntityNames(v)
	if !utf8.Valid(v) {
		v = []byte(strings.ToValidUTF8(string(v), "\uFFFD"))
	}
	return v
}

// linkURL is a destination as a link's URL, percent-encoded and not yet escaped for an attribute
// (escape does that where it is written), and empty unless it is relative or names a scheme
// bodyPolicy allows. A scheme is what precedes the first ':' when no '/', '?' or '#' comes before it.
//
// THE SCHEME IS READ FROM THE URL A BROWSER FOLLOWS, NOT FROM THE SOURCE. The source spells a
// destination with backslash escapes and character references — `javascript&colon;`,
// `&#106;avascript:` — which a reader resolves, so the destination is decoded first (textValue),
// then checked, then encoded. Checked before decoding, `&colon;` hid the colon from the check and
// the encoder put it back.
func linkURL(dest []byte) string {
	u := textValue(dest)
	if i := bytes.IndexAny(u, ":/?#"); i >= 0 && u[i] == ':' && !slices.Contains(linkSchemes, strings.ToLower(string(u[:i]))) {
		return ""
	}
	return string(util.URLEscape(u, false))
}

// supRef is a footnote reference outside a document's citation layer — a title, a note's scent —
// where it renders as its label.
func supRef(label string) string {
	return `<sup class="fnref"><a href="#fn-` + noteID(label) + `">` + escape(label) + `</a></sup>`
}

// noteID is a footnote label as it stands in an element id and the fragment that links to it: its
// letters, digits and `.`, `:` and `-` as written, and each other byte `_` and two hex digits. A
// label is the seat's to write and may hold a quotation mark or a space, which an attribute must
// escape and a URL must encode, two spellings of one id that the page's script compares; this one
// needs neither, and no two labels share it.
func noteID(label string) string {
	var b strings.Builder
	for i := 0; i < len(label); i++ {
		switch c := label[i]; {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '.', c == ':', c == '-':
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "_%02x", c)
		}
	}
	return b.String()
}

// uniqueID is id the first time the set asks for it, and id with a count after that.
func uniqueID(seen map[string]int, id string) string {
	seen[id]++
	if n := seen[id]; n > 1 {
		return uniqueID(seen, fmt.Sprintf("%s-%d", id, n))
	}
	return id
}

// footnote is one definition as the document holds it: its label, its text rendered (reference
// markers and all), and the scent its markers carry as a title.
type footnote struct {
	label, html, scent string
}

// refMarkByte brackets a reference marker and mentionMarkByte an id mention's (mention). Each is a
// control character no composer writes; one in a document's text is shown as the replacement
// character (unmarked), so no text can be read as a marker.
const (
	refMarkByte     = '\x01'
	mentionMarkByte = '\x02'
)

// refMark is the marker for the k-th reference inline met: where the cite will stand, and which
// reference it is. The label is held in the document's refs, never re-read from the text.
func refMark(k int) string { return mark(refMarkByte, k) }

func mark(c byte, k int) string { return fmt.Sprintf("%c%d%c", c, k, c) }

// eachRefMark calls f with the index of every reference marker in html, in order, and returns
// html with each marker replaced by what f returns.
func eachRefMark(html string, f func(k int) string) string { return eachMark(html, refMarkByte, f) }

// eachMark is eachRefMark for the markers c brackets.
func eachMark(html string, c byte, f func(k int) string) string {
	var b strings.Builder
	for {
		i := strings.IndexByte(html, c)
		if i < 0 {
			return b.String() + html
		}
		j := strings.IndexByte(html[i+1:], c)
		k := -1
		if j >= 0 {
			k, _ = strconv.Atoi(html[i+1 : i+1+j])
		}
		if k < 0 {
			panic("report: a marker the renderer did not write: " + strconv.Quote(html[i:]))
		}
		b.WriteString(html[:i])
		b.WriteString(f(k))
		html = html[i+1+j+1:]
	}
}

// citations renders the citation layer the way a reader of arXiv or PubMed expects: in-text
// markers become bracketed NUMBERED links — [1], [2] — assigned in order of first citation, and the
// References list is ORDERED to match, so entry 3 is the third source the reader met. The slug ids
// stay underneath (in the anchors and as a muted tag on each entry): they are the markdown tier's
// names, and grep must keep finding them in both tiers.
//
// The FIRST citation of each id also carries the anchor its entry links back to: one return address
// per entry, and it is the earliest, which is where the claim was made. A definition nothing cites
// keeps its place at the end, unnumbered in text but present in the list — dropping it would un-say
// a source the author recorded. Two definitions under one label are one entry, the later one.
//
// THE IDS ARE THE SET'S. The documents are tabs of one page, a fragment resolves to the first
// element of that id on the page, and every document with a citation has a note 1, so a label an
// earlier document also numbered takes a count — in the entry's id, the marker's and every link
// between the two — given in the order the page shows them. The label a reader sees, and the title
// on the marker, stay as written.
func citations(body string, refs []string, notes []footnote, seen map[string]int) string {
	byLabel := map[string]footnote{}
	var authored []string
	for _, n := range notes {
		if _, ok := byLabel[n.label]; !ok {
			authored = append(authored, n.label)
		}
		byLabel[n.label] = n
	}
	// Numbered in the order the citations are met: the text first, then the notes as written.
	num := map[string]int{}
	first := map[int]bool{}
	var order []string
	count := func(k int) string {
		if _, ok := num[refs[k]]; !ok {
			num[refs[k]] = len(order) + 1
			order = append(order, refs[k])
			first[k] = true
		}
		return ""
	}
	eachRefMark(body, count)
	for _, l := range authored {
		eachRefMark(byLabel[l].html, count)
	}
	for _, l := range authored {
		if _, cited := num[l]; !cited {
			order = append(order, l)
		}
	}
	listed := slices.DeleteFunc(slices.Clone(order), func(l string) bool { _, ok := byLabel[l]; return !ok })

	// The ids, in the order the page shows each label.
	// An entry's id is "fn-" and the marker's "fnref-" before one name, and the name is the first,
	// of the label and the label with a count, under which neither is given out yet: a heading
	// "Fnref 1" holds "fnref-1", and the note labelled 1 is then "fn-1-2" and "fnref-1-2".
	names := map[string]string{}
	id := func(l string) string {
		if _, ok := names[l]; !ok {
			name := noteID(l)
			for n := 2; seen["fn-"+name] > 0 || seen["fnref-"+name] > 0; n++ {
				name = fmt.Sprintf("%s-%d", noteID(l), n)
			}
			seen["fn-"+name]++
			seen["fnref-"+name]++
			names[l] = name
		}
		return names[l]
	}
	cite := func(k int) string {
		l := refs[k]
		// The scent is text of the note's rendering, escaped already; the label is the seat's.
		title := escape(l)
		if s := byLabel[l].scent; s != "" {
			title += " — " + s
		}
		if first[k] {
			return fmt.Sprintf(`<a class="cite" id="fnref-%s" href="#fn-%s" title="%s">[%d]</a>`, id(l), id(l), title, num[l])
		}
		return fmt.Sprintf(`<a class="cite" href="#fn-%s" title="%s">[%d]</a>`, id(l), title, num[l])
	}
	html := eachRefMark(body, cite)
	if len(listed) == 0 {
		return html
	}
	var items []string
	for _, l := range listed {
		li := `<li id="fn-` + id(l) + `">` + eachRefMark(byLabel[l].html, cite) +
			// data-nolink: the label IS the definition of that footnote, so the id-linker must
			// leave it alone. A label that links away reads as a reference to somewhere else.
			// The number leads (the <ol> renders it); the slug trails as a muted tag, because
			// it is the markdown tier's name for this entry and grep must find it here too.
			` <span class="fnlabel" data-nolink="">` + escape(l) + `</span>` +
			` <a class="fnback" href="#fnref-` + id(l) + `" aria-label="back to the citing text">&#8617;</a></li>`
		items = append(items, li)
	}
	// "References", because that is what the entries are — the citation layer the research
	// protocol requires — and a reader looking for a bibliography must be able to find it by its
	// own name.
	return html + "<section class=\"footnotes\"><h2>References</h2><ol class=\"fn\">\n" + strings.Join(items, "\n") + "\n</ol></section>\n"
}

// scentOf is the tooltip a marker carries: the text of its note, cut short — a scent, not the
// reference. html is the note's own rendering, which this package wrote and escaped, so every '<'
// in it opens a tag.
func scentOf(html string) string {
	var b strings.Builder
	for depth, r := 0, []rune(html); len(r) > 0; r = r[1:] {
		switch {
		case r[0] == '<':
			depth++
		case r[0] == '>' && depth > 0:
			depth--
		case depth == 0:
			b.WriteRune(r[0])
		}
	}
	text := strings.TrimSpace(b.String())
	if len(text) > 140 {
		if cut := strings.LastIndexByte(text[:140], ' '); cut > 60 {
			text = text[:cut]
		} else {
			text = text[:140]
		}
		text += "…"
	}
	return text
}

// escape is deliberately attribute-safe as well as text-safe: the same helper builds a <title>
// and a tab's title="…", and a run whose topic contains a quotation mark must not be able to
// terminate an attribute. Model-written prose and repository-quoted findings both reach here.
func escape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// inline renders one line's inline markdown outside a document — a title — where a footnote
// reference has no list to number it against and renders as its label.
func inline(s string) string {
	s = unmarked(s)
	r := &mdRender{src: []byte(s), seen: map[string]int{}, ref: supRef, looseRefs: true}
	var b strings.Builder
	doc := lineReader.Parse(text.NewReader(r.src))
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		r.inlines(&b, c, false)
	}
	return b.String()
}

// linkIDs turns every mention of a record id into a link to the element that defines it —
// the join a multi-document set can express and a markdown file cannot. An id is never linked at
// the element that defines it: a heading linking to itself makes every definition look like a
// reference to somewhere else. `self` is the document being rendered: a link into the same file
// stays a fragment, so the tab does not reload itself.
func linkIDs(doc rendered, anchor anchors, self string) string {
	return eachMark(doc.html, mentionMarkByte, func(k int) string {
		m := doc.mentions[k]
		target, ok := anchor[m.id]
		if !ok {
			return m.id
		}
		file, frag, _ := strings.Cut(target, "#")
		if frag == m.here && file == self {
			return m.id
		}
		if file == self {
			return `<a class="idref" href="#` + escape(frag) + `">` + m.id + `</a>`
		}
		return `<a class="idref" href="#` + escape(frag) + `" data-doc="` + escape(file) + `">` + m.id + `</a>`
	})
}
