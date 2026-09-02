package report

// MARKDOWN → HTML, IN THE SUBSET THIS ASSEMBLER EMITS.
//
// go.mod carries no markdown library and does not get one for a rendering convenience: the
// dependency would be load-bearing for the deliverable of every run, and the surface it would
// buy is a superset of what any composer here produces. What the assembler emits is a closed
// set — headings, fenced code, GFM tables, lists, blockquotes, rules, footnotes and a handful
// of inline forms — and the one open surface (blue's lifted prose) is written by a model
// following the same template.
//
// WHAT IT DOES NOT DO IS AS IMPORTANT: it does not round-trip. The markdown files are the
// durable tier and are written by the composers directly; this reads them ONCE, forward, into
// a rendering. Nothing here can change what ships in a .md, which is why an unsupported
// construct degrades to visible text rather than to a silent drop.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
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
	// idToken matches the identifiers the record mints and every document quotes: gaps
	// (R4-3), motions (M2), proofs (P1) and lens findings (L2-F1). They are the joins the
	// single-file report made a reader scroll for.
	idToken = regexp.MustCompile(`\b(R\d+-\d+|M\d+|P\d+|L\d+-F\d+)\b`)
)

// anchors is the map from a record id to the document and element that DEFINES it, filled as
// each document renders and applied across the set afterwards.
type anchors map[string]string

// mdToHTML renders one document. It records every anchor it defines into anchor, keyed by the
// record id, valued as "<file>#<elementID>".
func mdToHTML(md, file string, anchor anchors) string {
	lines := strings.Split(md, "\n")
	var b strings.Builder
	var foot []string
	seen := map[string]int{}

	slug := func(text string) string {
		s := strings.ToLower(strings.TrimSpace(stripInline(text)))
		s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
		s = strings.Trim(s, "-")
		if s == "" {
			s = "section"
		}
		seen[s]++
		if n := seen[s]; n > 1 {
			s = fmt.Sprintf("%s-%d", s, n)
		}
		return s
	}
	// define records "this element is where id lives", first definition winning: a heading
	// that introduces R4-3 is its home, and the twenty later mentions are references to it.
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
		b.WriteString("<p>" + inline(strings.Join(para, "\n")) + "</p>\n")
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
		// definition, so it registers, and it renders as itself.
		if m := reAnchorTag.FindStringSubmatch(t); m != nil {
			flushPara()
			b.WriteString(`<a id="` + escape(m[1]) + `"></a>` + "\n")
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
			foot = append(foot, `<li id="fn-`+escape(m[1])+`"><span class="fnlabel" data-nolink>`+escape(m[1])+`</span> `+inline(m[2])+
				` <a class="fnback" href="#fnref-`+escape(m[1])+`" aria-label="back to the citing text">&#8617;</a></li>`)
			continue
		}

		if m := reHeading.FindStringSubmatch(t); m != nil {
			flushPara()
			closeList(&listStack)
			level := len(m[1])
			id := slug(m[2])
			define(m[2], id)
			fmt.Fprintf(&b, "<h%d id=\"%s\">%s</h%d>\n", level, id, inline(m[2]), level)
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
			b.WriteString("<blockquote>" + inline(strings.Join(quote, " ")) + "</blockquote>\n")
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
				b.WriteString("<th>" + inline(c) + "</th>")
			}
			b.WriteString("</tr></thead><tbody>\n")
			for i += 2; i < len(lines); i++ {
				row := strings.TrimSpace(lines[i])
				if row == "" || !strings.Contains(row, "|") {
					break
				}
				b.WriteString("<tr>")
				for _, c := range splitRow(row) {
					b.WriteString("<td>" + inline(c) + "</td>")
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
			writeItem(&b, id, inline(m[2]))
			continue
		}
		if m := reOLItem.FindStringSubmatch(ln); m != nil {
			flushPara()
			b.WriteString(openList(&listStack, "ol", len(m[1])))
			writeItem(&b, "", inline(m[2]))
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
			b.WriteString("<div class=\"cont\">" + inline(t) + "</div>\n")
			continue
		}
		closeList(&listStack)
		// LEFT-trimmed only: a trailing double space is markdown's hard break and TrimSpace
		// eats it, which silently un-does the composer's line breaks one layer down.
		para = append(para, strings.TrimLeft(ln, " \t"))
	}
	flushPara()
	closeList(&listStack)
	if len(foot) > 0 {
		// "References", because that is what the entries are — the citation layer the
		// research protocol requires — and a reader looking for a bibliography must be able
		// to find it by its own name.
		b.WriteString("<section class=\"footnotes\"><h2>References</h2><ol class=\"fn\">\n" + strings.Join(foot, "\n") + "\n</ol></section>\n")
	}
	return withFirstRefAnchors(b.String())
}

// withFirstRefAnchors gives the FIRST in-text citation of each id an anchor the References
// entry can link back to. Every later citation of the same id stays a plain marker: one
// return address per entry, and it is the earliest, which is where the claim was made.
func withFirstRefAnchors(html string) string {
	seen := map[string]bool{}
	return reFnrefSup.ReplaceAllStringFunc(html, func(m string) string {
		id := reFnrefSup.FindStringSubmatch(m)[1]
		if seen[id] {
			return m
		}
		seen[id] = true
		return `<sup class="fnref" id="fnref-` + id + `">` + m[len(`<sup class="fnref">`):]
	})
}

var reFnrefSup = regexp.MustCompile(`<sup class="fnref"><a href="#fn-([^"]+)">`)

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

// inline renders one line's inline markdown. Code spans are lifted out FIRST and restored
// last, so nothing inside a backtick is treated as emphasis, a link, or an id.
func inline(s string) string {
	var spans []string
	s = reCodeSpan.ReplaceAllStringFunc(s, func(m string) string {
		spans = append(spans, "<code>"+escape(strings.Trim(m, "`"))+"</code>")
		return fmt.Sprintf("\x00%d\x00", len(spans)-1)
	})
	s = escape(s)
	s = reFootRef.ReplaceAllString(s, `<sup class="fnref"><a href="#fn-$1">$1</a></sup>`)
	s = reLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
	s = reAutolink.ReplaceAllString(s, `$1<a href="$2">$2</a>`)
	s = reStrong.ReplaceAllString(s, "<strong>$1</strong>")
	s = reEmphasis.ReplaceAllString(s, "$1<em>$2</em>$3")
	s = reEmphasisU.ReplaceAllString(s, "$1<em>$2</em>$3")
	s = reHardBreak.ReplaceAllString(s, "<br>\n")
	s = strings.ReplaceAll(s, "\n", " ")
	for i, span := range spans {
		s = strings.ReplaceAll(s, fmt.Sprintf("\x00%d\x00", i), span)
	}
	return s
}

// reHardBreak is markdown's two-trailing-spaces line break, applied after escaping so it
// cannot be confused with anything inside a code span.
var reHardBreak = regexp.MustCompile(`  +\n`)

// linkIDs turns every mention of a record id into a link to the element that defines it —
// the join a seven-document set can express and a markdown file cannot.
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
