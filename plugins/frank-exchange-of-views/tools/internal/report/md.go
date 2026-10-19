package report

// MARKDOWN → HTML, IN THE SUBSET THIS ASSEMBLER EMITS.
//
// The rendering is this file, not the markdown parser go.mod carries (which reads the set's
// footnotes, footnotes.go): a library renderer would be load-bearing for the deliverable of every
// run, and the surface it would buy is a superset of what any composer here produces. What the assembler emits is a closed
// set — headings, fenced code, GFM tables, lists, blockquotes, rules, footnotes and a handful
// of inline forms — and the one open surface (the report's own prose) is written by a model
// following the same template.
//
// WHAT IT DOES NOT DO IS AS IMPORTANT: it does not round-trip. The markdown files are the
// durable tier and are written by the composers directly; this reads them ONCE, forward, into
// a rendering. Nothing here can change what ships in a .md, which is why an unsupported
// construct degrades to visible text rather than to a silent drop.

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

var (
	reHeading  = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	reFence    = regexp.MustCompile("^(```|~~~)\\s*([A-Za-z0-9_+-]*)\\s*$")
	reULItem   = regexp.MustCompile(`^(\s*)[-*+]\s+(.*)$`)
	reOLItem   = regexp.MustCompile(`^(\s*)\d+[.)]\s+(.*)$`)
	reTableSep = regexp.MustCompile(`^\s*\|?\s*:?-{1,}:?\s*(\|\s*:?-{1,}:?\s*)*\|?\s*$`)
	reFootnote = regexp.MustCompile(`^\[\^([A-Za-z0-9_-]+)\]:\s*(.*)$`)
	// RE2 has no backreferences, so the three rule characters are spelled out rather than
	// matched against the first one.
	reHR       = regexp.MustCompile(`^\s*(?:-\s*-\s*-[-\s]*|\*\s*\*\s*\*[*\s]*|_\s*_\s*_[_\s]*)$`)
	reCodeSpan = regexp.MustCompile("`[^`]+`")
	reLink     = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)\)`)
	reFootRef  = regexp.MustCompile(`\[\^([A-Za-z0-9_-]+)\]`)
	reStrong   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reEmphasis = regexp.MustCompile(`(^|[\s(])\*([^*\n]+)\*($|[\s).,;:!?])`)
	// Underscore emphasis, which every composer here uses for its parenthetical notes. The
	// boundary conditions are what keep `required_fix` and `run_id` out of it: an opening
	// underscore must follow a space or a line start, and the closing one must precede a space,
	// a line end, or punctuation.
	reEmphasisU = regexp.MustCompile(`(^|[\s(])_([^_\n]+)_($|[\s).,;:!?])`)
	reAutolink  = regexp.MustCompile(`(^|[\s(])(https?://[^\s<>()\[\]]+)`)
	reAnchorTag = regexp.MustCompile(`(?i)^<a\s+id="([^"]+)"`)
	// idToken matches the identifiers the record mints and every document quotes: gap, motion
	// and finding ids, and a proof's footnote number (P1). They are the joins the single-file report
	// made a reader scroll for. An id that matches nothing renders exactly like an id nobody else
	// mentioned, so the id shapes are read from the kinds table, never restated.
	idToken = regexp.MustCompile(`\b(` + anchor.IDPattern("gap", "motion", "finding") + `|P\d+)\b`)
)

// anchors is the map from a record id to the document and element that DEFINES it, filled as
// each document renders and applied across the set afterwards.
type anchors map[string]string

// mdToHTML renders one document. It records every anchor it defines into anchor, keyed by the
// record id, valued as "<file>#<elementID>".
//
// seen counts the ids given out, ACROSS THE SET: the documents are tabs of one page, a fragment
// resolves to the first element of that id on the page, and a hidden tab's element is no place
// to scroll to. The report's own "## Proofs" and evidence.md's are "proofs" and "proofs-2", and
// the note each of two documents numbers 1 is "fn-1" in the first and "fn-1-2" in the second
// (noteIDs).
//
// files is the document set: a link to one of them is a tab switch, written as one here, because
// in a single page there is no sibling file to open.
func mdToHTML(md, file string, anchor anchors, seen map[string]int, files map[string]bool) string {
	// The marker byte is a control character with no rendering; one in the text is shown as the
	// replacement character, so no text can be read as a reference.
	md = strings.ReplaceAll(md, string(refMarkByte), "\uFFFD")
	lines := strings.Split(md, "\n")
	var b strings.Builder
	// THE CITATION LAYER IS HELD, NOT RE-READ. Each reference is recorded where inline meets it and
	// left in the text as a marker that holds only its index (refMark); each definition is kept as
	// a note. citations numbers and lists them from these once the document is rendered.
	var refs []string
	var notes []footnote
	in := func(s string) string {
		return renderInline(s, func(label string) string {
			refs = append(refs, label)
			return refMark(len(refs) - 1)
		}, files)
	}

	slug := func(text string) string {
		s := strings.ToLower(strings.TrimSpace(stripInline(text)))
		s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
		s = strings.Trim(s, "-")
		if s == "" {
			s = "section"
		}
		return uniqueID(seen, s)
	}
	// define records "this element is where id lives", first definition winning: a heading
	// that introduces G3 is its home, and the twenty later mentions are references to it.
	define := func(text, elemID string) {
		if m := idToken.FindStringSubmatch(strings.TrimSpace(stripInline(text))); m != nil && strings.HasPrefix(strings.TrimSpace(stripInline(text)), m[1]) {
			if _, ok := anchor[m[1]]; !ok {
				anchor[m[1]] = file + "#" + elemID
			}
		}
	}

	closeList := func(stack *[]string) {
		for len(*stack) > 0 {
			b.WriteString("</" + (*stack)[len(*stack)-1] + ">\n")
			*stack = (*stack)[:len(*stack)-1]
		}
	}

	var listStack []string
	var para []string
	flushPara := func() {
		if len(para) == 0 {
			return
		}
		// A line ending in two spaces is a markdown hard break, and the composers use it to
		// keep a gap's location, grades, demanded fix and acceptance check on four lines
		// instead of fusing them into one sentence.
		b.WriteString("<p>" + in(strings.Join(para, "\n")) + "</p>\n")
		para = nil
	}

	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		t := strings.TrimSpace(ln)

		// Fenced code: taken verbatim to its closing fence, escaped, never parsed. A "## " or
		// a "|" inside a script is not a heading and not a table.
		if m := reFence.FindStringSubmatch(t); m != nil {
			flushPara()
			closeList(&listStack)
			lang := m[2]
			var code []string
			for i++; i < len(lines); i++ {
				if reFence.MatchString(strings.TrimSpace(lines[i])) {
					break
				}
				code = append(code, lines[i])
			}
			cls := ""
			if lang != "" {
				cls = ` class="language-` + escape(lang) + `"`
			}
			b.WriteString("<pre><code" + cls + ">" + escape(strings.Join(code, "\n")) + "</code></pre>\n")
			continue
		}

		if t == "" {
			flushPara()
			closeList(&listStack)
			continue
		}

		// An explicit anchor the composer placed (evidence.md does this per proof). It is a
		// definition, so it registers, and it renders as itself. A line that opens with one and
		// goes on is the report's own HTML, and the whole line passes through.
		if m := reAnchorTag.FindStringSubmatch(t); m != nil {
			flushPara()
			if t == `<a id="`+m[1]+`"></a>` {
				b.WriteString(`<a id="` + escape(m[1]) + `"></a>` + "\n")
			} else {
				b.WriteString(ln + "\n")
			}
			if _, ok := anchor[m[1]]; !ok {
				anchor[m[1]] = file + "#" + m[1]
			}
			continue
		}

		if m := reFootnote.FindStringSubmatch(t); m != nil {
			flushPara()
			closeList(&listStack)
			// data-nolink: the label IS the definition of that footnote, so the id-linker must
			// leave it alone. A label that links away reads as a reference to somewhere else.
			// The number leads (the <ol> renders it); the slug trails as a muted tag, because
			// it is the markdown tier's name for this entry and grep must find it here too.
			notes = append(notes, footnote{label: m[1], html: in(m[2]), scent: scentOf(inline(m[2]))})
			continue
		}

		if m := reHeading.FindStringSubmatch(t); m != nil {
			flushPara()
			closeList(&listStack)
			level := len(m[1])
			id := slug(m[2])
			define(m[2], id)
			fmt.Fprintf(&b, "<h%d id=\"%s\">%s</h%d>\n", level, id, in(m[2]), level)
			continue
		}

		if reHR.MatchString(t) {
			flushPara()
			closeList(&listStack)
			b.WriteString("<hr>\n")
			continue
		}

		if strings.HasPrefix(t, ">") {
			flushPara()
			closeList(&listStack)
			var quote []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				quote = append(quote, strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), ">")), " "))
				i++
			}
			i--
			b.WriteString("<blockquote>" + in(strings.Join(quote, " ")) + "</blockquote>\n")
			continue
		}

		// GFM table: a row followed by a delimiter row. Without the delimiter it is a
		// paragraph that happens to contain pipes, which is what prose about a shell command
		// looks like.
		if strings.Contains(t, "|") && i+1 < len(lines) && reTableSep.MatchString(lines[i+1]) {
			flushPara()
			closeList(&listStack)
			// The wrapper is what scrolls: a wide table inside it pans on a phone instead of
			// stretching the whole page and dragging every paragraph with it.
			b.WriteString("<div class=\"tblwrap\"><table><thead><tr>")
			for _, c := range splitRow(t) {
				b.WriteString("<th>" + in(c) + "</th>")
			}
			b.WriteString("</tr></thead><tbody>\n")
			for i += 2; i < len(lines); i++ {
				row := strings.TrimSpace(lines[i])
				if row == "" || !strings.Contains(row, "|") {
					break
				}
				b.WriteString("<tr>")
				for _, c := range splitRow(row) {
					b.WriteString("<td>" + in(c) + "</td>")
				}
				b.WriteString("</tr>\n")
			}
			i--
			b.WriteString("</tbody></table></div>\n")
			continue
		}

		if m := reULItem.FindStringSubmatch(ln); m != nil {
			flushPara()
			b.WriteString(openList(&listStack, "ul", len(m[1])))
			id := ""
			if tok := idToken.FindStringSubmatch(stripInline(m[2])); tok != nil && strings.HasPrefix(stripInline(m[2]), tok[1]) {
				id = slug(tok[1])
				define(m[2], id)
			}
			writeItem(&b, id, in(m[2]))
			continue
		}
		if m := reOLItem.FindStringSubmatch(ln); m != nil {
			flushPara()
			b.WriteString(openList(&listStack, "ol", len(m[1])))
			writeItem(&b, "", in(m[2]))
			continue
		}

		// Raw HTML the composer wrote on purpose — passed through, never escaped.
		if strings.HasPrefix(t, "<") {
			flushPara()
			b.WriteString(ln + "\n")
			continue
		}

		if len(listStack) > 0 && strings.HasPrefix(ln, "  ") {
			// A continuation line under a list item: appended to the item rather than
			// promoted to a paragraph that breaks the list open.
			b.WriteString("<div class=\"cont\">" + in(t) + "</div>\n")
			continue
		}
		closeList(&listStack)
		// LEFT-trimmed only: a trailing double space is markdown's hard break and TrimSpace
		// eats it, which silently un-does the composer's line breaks one layer down.
		para = append(para, strings.TrimLeft(ln, " \t"))
	}
	flushPara()
	closeList(&listStack)
	return citations(b.String(), refs, notes, seen)
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

// refMarkByte brackets a reference marker. It is a control character no composer writes and inline
// leaves alone, and the code-span placeholders use NUL, so the two cannot be read for each other.
const refMarkByte = '\x01'

// refMark is the marker for the k-th reference inline met: where the cite will stand, and which
// reference it is. The label is held in the document's refs, never re-read from the text.
func refMark(k int) string { return fmt.Sprintf("%c%d%c", refMarkByte, k, refMarkByte) }

// eachRefMark calls f with the index of every reference marker in html, in order, and returns
// html with each marker replaced by what f returns.
func eachRefMark(html string, f func(k int) string) string {
	var b strings.Builder
	for {
		i := strings.IndexByte(html, refMarkByte)
		if i < 0 {
			return b.String() + html
		}
		j := strings.IndexByte(html[i+1:], refMarkByte)
		k := -1
		if j >= 0 {
			k, _ = strconv.Atoi(html[i+1 : i+1+j])
		}
		if k < 0 {
			panic("report: a reference marker that inline did not write: " + strconv.Quote(html[i:]))
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
	suffix := map[string]string{}
	id := func(l string) string {
		if _, ok := suffix[l]; !ok {
			suffix[l] = strings.TrimPrefix(uniqueID(seen, "fn-"+l), "fn-"+l)
		}
		return escape(l) + suffix[l]
	}
	cite := func(k int) string {
		l := refs[k]
		title := l
		if s := byLabel[l].scent; s != "" {
			title = l + " — " + s
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
			` <span class="fnlabel" data-nolink>` + escape(l) + `</span>` +
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

// openList emits whatever open/close tags are needed to reach the requested kind and depth.
func openList(stack *[]string, kind string, indent int) string {
	depth := indent/2 + 1
	var out strings.Builder
	for len(*stack) > depth {
		out.WriteString("</" + (*stack)[len(*stack)-1] + ">")
		*stack = (*stack)[:len(*stack)-1]
	}
	for len(*stack) < depth {
		out.WriteString("<" + kind + ">")
		*stack = append(*stack, kind)
	}
	return out.String()
}

func writeItem(b *strings.Builder, id, html string) {
	if id != "" {
		b.WriteString(`<li id="` + id + `">` + html + "</li>\n")
		return
	}
	b.WriteString("<li>" + html + "</li>\n")
}

func splitRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	cells := strings.Split(row, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
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

// stripInline reduces markdown to the text a heading id and an id-token scan should see.
func stripInline(s string) string {
	s = reCodeSpan.ReplaceAllStringFunc(s, func(m string) string { return strings.Trim(m, "`") })
	s = reLink.ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// inline renders one line's inline markdown outside a document — a title — where a footnote
// reference has no list to number it against and renders as its label.
func inline(s string) string {
	return renderInline(s, func(label string) string {
		return `<sup class="fnref"><a href="#fn-` + label + `">` + label + `</a></sup>`
	}, nil)
}

// renderInline renders one line's inline markdown, each footnote reference as ref renders it, and
// a link to a document in files as a switch to that document's tab (siteLink).
// Code spans are lifted out FIRST and restored last, so nothing inside a backtick is treated as
// emphasis, a link, or an id.
func renderInline(s string, ref func(label string) string, files map[string]bool) string {
	var spans []string
	s = reCodeSpan.ReplaceAllStringFunc(s, func(m string) string {
		spans = append(spans, "<code>"+escape(strings.Trim(m, "`"))+"</code>")
		return fmt.Sprintf("\x00%d\x00", len(spans)-1)
	})
	// AN ESCAPED `[^` IS TEXT (notes.go): lifted out as a code span is, so the reference pattern
	// below does not read it, and restored as the two characters without their backslash.
	s = strings.ReplaceAll(s, noteEscape, "\x00note\x00")
	s = escape(s)
	s = reFootRef.ReplaceAllStringFunc(s, func(m string) string { return ref(reFootRef.FindStringSubmatch(m)[1]) })
	s = reLink.ReplaceAllStringFunc(s, func(m string) string {
		l := reLink.FindStringSubmatch(m)
		return siteLink(l[2], files) + l[1] + "</a>"
	})
	s = reAutolink.ReplaceAllString(s, `$1<a href="$2">$2</a>`)
	s = reStrong.ReplaceAllString(s, "<strong>$1</strong>")
	s = reEmphasis.ReplaceAllString(s, "$1<em>$2</em>$3")
	s = reEmphasisU.ReplaceAllString(s, "$1<em>$2</em>$3")
	s = reHardBreak.ReplaceAllString(s, "<br>\n")
	s = strings.ReplaceAll(s, "\n", " ")
	for i, span := range spans {
		s = strings.ReplaceAll(s, fmt.Sprintf("\x00%d\x00", i), span)
	}
	return strings.ReplaceAll(s, "\x00note\x00", noteOpen)
}

// reHardBreak is markdown's two-trailing-spaces line break, applied after escaping so it
// cannot be confused with anything inside a code span.
var reHardBreak = regexp.MustCompile(`  +\n`)

// linkIDs turns every mention of a record id into a link to the element that defines it —
// the join a multi-document set can express and a markdown file cannot.
//
// IT WALKS TAGS, not text. A blind replace over rendered HTML would rewrite ids inside an
// href, inside a heading's own id attribute, and inside the anchor that defines them, which
// produces a document that links every id to itself.
func linkIDs(html string, anchor anchors, self string) string {
	if len(anchor) == 0 {
		return html
	}
	ids := make([]string, 0, len(anchor))
	for id := range anchor {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	depthA, inCode := 0, 0
	// here is the id of the element currently being rendered, so an id is never linked at the
	// site that DEFINES it: a heading linking to itself is the shape a naive replace produces,
	// and it makes every definition look like a reference to somewhere else.
	here := ""
	noLink := false
	for len(html) > 0 {
		lt := strings.IndexByte(html, '<')
		if lt < 0 {
			b.WriteString(linkText(html, anchor, self, here, noLink || depthA > 0 || inCode > 0))
			break
		}
		b.WriteString(linkText(html[:lt], anchor, self, here, noLink || depthA > 0 || inCode > 0))
		gt := strings.IndexByte(html[lt:], '>')
		if gt < 0 {
			b.WriteString(html[lt:])
			break
		}
		tag := html[lt : lt+gt+1]
		b.WriteString(tag)
		switch {
		case strings.HasPrefix(tag, "<a"):
			depthA++
		case strings.HasPrefix(tag, "</a"):
			depthA--
		case strings.HasPrefix(tag, "<code"), strings.HasPrefix(tag, "<pre"):
			inCode++
		case strings.HasPrefix(tag, "</code"), strings.HasPrefix(tag, "</pre"):
			inCode--
		}
		switch {
		case strings.Contains(tag, "data-nolink"):
			noLink = true
		case strings.HasPrefix(tag, "</"):
			noLink = false
		}
		if m := reIDAttr.FindStringSubmatch(tag); m != nil {
			here = m[1]
		} else if strings.HasPrefix(tag, "</") {
			here = ""
		}
		html = html[lt+gt+1:]
	}
	return b.String()
}

var reIDAttr = regexp.MustCompile(`(?i)^<[a-z0-9]+[^>]*\sid="([^"]+)"`)

// linkText rewrites the ids in one run of plain text. `self` is the document being rendered:
// a link into the same file stays a fragment, so the tab does not reload itself.
func linkText(text string, anchor anchors, self, here string, skip bool) string {
	if skip || text == "" {
		return text
	}
	return idToken.ReplaceAllStringFunc(text, func(tok string) string {
		target, ok := anchor[tok]
		if !ok {
			return tok
		}
		file, frag, _ := strings.Cut(target, "#")
		if frag == here && file == self {
			return tok
		}
		if file == self {
			return `<a class="idref" href="#` + frag + `">` + tok + `</a>`
		}
		return `<a class="idref" href="#` + frag + `" data-doc="` + file + `">` + tok + `</a>`
	})
}
