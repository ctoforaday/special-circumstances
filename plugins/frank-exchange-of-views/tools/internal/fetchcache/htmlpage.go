package fetchcache

import (
	"bytes"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// AN HTML PAGE IS PARSED ONCE, AND EVERY QUESTION THIS PACKAGE ASKS OF IT READS THE RESULT.
//
// The wall detector, the completeness verdict, the rights tags, the landing page's pointers to its
// own full text, the markup redirect and the page's readable text all ask about the same bytes.
// They are answered from one parse by golang.org/x/net/html — the HTML standard's tree builder —
// rather than by a pattern each over the raw bytes. A pattern over markup reads attributes in
// one order, ends an element at the first `>` inside a quoted value, closes a `<script>` at a
// `</style>`, and finds a `<title>` inside an inline image; each of those was a silent miss or a
// wrong fact on this path. The parser reads an attribute by its name wherever it sits, decodes
// its value, keeps script and style text out of the page's text, and implies the `</head>`,
// `</p>` and `</li>` a page leaves out, as a browser does.
//
// The page is parsed as a browser with scripting ON parses it, so a `<noscript>` block is one
// string of raw text that is not the page's text — parsed with scripting off, a `<noscript>` in
// the head carrying prose ends the head there and its cookie notice lands in the body.

// htmlPage is what this package reads from one HTML page.
type htmlPage struct {
	// metas and links are every <meta> and <link> element's attributes, in document order. Keys
	// are lower case (the parser's rule); values are entity-decoded.
	metas, links []map[string]string
	// title is the document's own <title> — never an inline SVG's, which is a different element.
	title    string
	password bool            // an <input type=password> is on the page
	dataTest map[string]bool // every data-test attribute value
	classes  map[string]bool // every class token on any element

	// Springer Nature: the article-body element, whether its reference list (id="Bib1") was
	// reached, and the numbered body sections (h2 id="Sec<n>") inside the body before it.
	snBody     bool
	snRefs     bool
	snSections int
	// PubMed Central: the same three facts — the main-article-body element, its reference list
	// (an id beginning "ref-list"), and the pmc_sec_title headings (h2–h4) before it.
	pmcBody     bool
	pmcRefs     bool
	pmcSections int

	// mount reports a body whose only content is empty <div>s, <noscript> and <script> — the
	// mount point a client-side app renders into.
	mount bool
	// text is the page's readable text with a line break where a block breaks (htmlText's
	// output); flat is the same text as one whitespace-collapsed line.
	text, flat string
}

// meta returns the content of the first <meta> whose attribute key equals value (compared
// without case), and whether there was one.
func (p *htmlPage) meta(key, value string) (string, bool) {
	for _, m := range p.metas {
		if v, ok := m[key]; ok && strings.EqualFold(strings.TrimSpace(v), value) {
			return m["content"], true
		}
	}
	return "", false
}

// readHTML returns the parse of an HTML page, parsing it at most once while it is in use.
//
// ONE PARSE PER PAGE. The fetch path asks about the same bytes from several places — the fetcher
// reads the rights tags and the markup redirect, the store reads the pointers, the text, the wall
// and the completeness — and each would otherwise parse the page again. The cache is keyed on the
// slice itself (its first byte's address and its length) and holds a reference to it, so the
// address cannot be reused by another body while the entry lives. Bodies are never written after
// they are read off the wire; a caller that changed one in place would read a stale parse.
func readHTML(b []byte) *htmlPage {
	if len(b) == 0 {
		return parseHTML(b)
	}
	parsed.Lock()
	for _, e := range parsed.ring {
		if len(e.body) == len(b) && &e.body[0] == &b[0] {
			parsed.Unlock()
			return e.page
		}
	}
	parsed.Unlock()
	p := parseHTML(b)
	parsed.Lock()
	parsed.ring[parsed.next] = parsedEntry{body: b, page: p}
	parsed.next = (parsed.next + 1) % len(parsed.ring)
	parsed.Unlock()
	return p
}

type parsedEntry struct {
	body []byte
	page *htmlPage
}

// parsed holds the last few pages parsed: enough for the fetches a run makes concurrently.
var parsed struct {
	sync.Mutex
	ring [8]parsedEntry
	next int
}

func parseHTML(b []byte) *htmlPage {
	p := &htmlPage{dataTest: map[string]bool{}, classes: map[string]bool{}}
	// Parse fails only when its reader does, and a bytes.Reader does not.
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return p
	}
	var raw strings.Builder
	p.walk(doc, walkState{}, &raw)
	p.text = readerLines(raw.String())
	p.flat = strings.Join(strings.Fields(p.text), " ")
	p.mount = mountOnly(doc)
	return p
}

type walkState struct {
	hidden      bool // inside an element whose text a reader never sees
	inSN, inPMC bool // inside Springer Nature's / PubMed Central's article body
}

func (p *htmlPage) walk(n *html.Node, st walkState, raw *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		if !st.hidden {
			raw.WriteString(collapseSource(n.Data))
		}
		return
	case html.ElementNode:
		st = p.element(n, st)
		if n.Namespace == "" {
			if n.DataAtom == atom.Br || n.DataAtom == atom.Hr {
				raw.WriteByte('\n') // a void element is one break, not an opening and a closing
			} else if blockElements[n.DataAtom] {
				raw.WriteByte('\n')
				defer raw.WriteByte('\n')
			} else if n.DataAtom == atom.Td || n.DataAtom == atom.Th {
				raw.WriteByte(' ')
				defer raw.WriteByte(' ')
			}
		}
	case html.DocumentNode:
	default: // comments and the doctype carry no text
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.walk(c, st, raw)
	}
}

// element records one element's facts and returns the state its children are read in.
func (p *htmlPage) element(n *html.Node, st walkState) walkState {
	attrs := make(map[string]string, len(n.Attr))
	for _, a := range n.Attr {
		if a.Namespace == "" {
			attrs[a.Key] = a.Val
		}
	}
	if v, ok := attrs["data-test"]; ok {
		p.dataTest[v] = true
	}
	classes := strings.Fields(attrs["class"])
	for _, c := range classes {
		p.classes[c] = true
	}
	if n.Namespace != "" {
		// An inline SVG's <title> and text are a picture's, not the page's. MathML's text is the
		// formula a reader sees, and is kept.
		if n.Namespace == "svg" {
			st.hidden = true
		}
		return st
	}
	id := attrs["id"]
	switch n.DataAtom {
	case atom.Meta:
		p.metas = append(p.metas, attrs)
	case atom.Link:
		p.links = append(p.links, attrs)
	case atom.Title:
		if p.title == "" && n.FirstChild != nil {
			p.title = strings.Join(strings.Fields(n.FirstChild.Data), " ")
		}
	case atom.Input:
		if strings.EqualFold(strings.TrimSpace(attrs["type"]), "password") {
			p.password = true
		}
	}
	// THE REFERENCE LIST ENDS THE COUNT for its platform: a section after it is not body.
	if id == "Bib1" {
		p.snRefs = true
	}
	if p.pmcBody && strings.HasPrefix(id, "ref-list") {
		p.pmcRefs = true
	}
	if has(classes, "c-article-body") {
		p.snBody, st.inSN = true, true
	}
	if has(classes, "main-article-body") {
		p.pmcBody, st.inPMC = true, true
	}
	if st.inSN && !p.snRefs && n.DataAtom == atom.H2 && numbered(id, "Sec") {
		p.snSections++
	}
	if st.inPMC && !p.pmcRefs && (n.DataAtom == atom.H2 || n.DataAtom == atom.H3 || n.DataAtom == atom.H4) && has(classes, "pmc_sec_title") {
		p.pmcSections++
	}
	if readerHidden[n.DataAtom] {
		st.hidden = true
	}
	return st
}

// readerHidden are the elements whose content a reader of the page does not see as its text: the
// head, and the title wherever a page put it (it is the extraction's Title), code, styles,
// templates, the no-script fallback, and an embedded frame's fallback content — which the parser
// keeps as raw text, markup and all.
var readerHidden = map[atom.Atom]bool{
	atom.Head: true, atom.Title: true, atom.Script: true, atom.Style: true, atom.Noscript: true,
	atom.Template: true, atom.Iframe: true,
}

// blockElements are where a reader sees a line break.
var blockElements = func() map[atom.Atom]bool {
	m := map[atom.Atom]bool{}
	for _, a := range []atom.Atom{
		atom.P, atom.Div, atom.Br, atom.Li, atom.Ul, atom.Ol, atom.Dl, atom.Dd, atom.Dt,
		atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6,
		atom.Table, atom.Thead, atom.Tbody, atom.Tfoot, atom.Tr, atom.Caption,
		atom.Section, atom.Article, atom.Header, atom.Footer, atom.Nav, atom.Aside, atom.Main,
		atom.Blockquote, atom.Pre, atom.Figure, atom.Figcaption, atom.Hr, atom.Form, atom.Fieldset,
		atom.Address, atom.Details, atom.Summary,
	} {
		m[a] = true
	}
	return m
}()

func has(tokens []string, want string) bool {
	for _, t := range tokens {
		if t == want {
			return true
		}
	}
	return false
}

// numbered reports whether s is prefix followed by one or more ASCII digits and nothing else.
func numbered(s, prefix string) bool {
	rest, ok := strings.CutPrefix(s, prefix)
	return ok && rest != "" && strings.Trim(rest, "0123456789") == ""
}

// mountOnly reports a body whose element children are only empty <div>s, <noscript> and
// <script>, with no text beside them — and at least one such child.
func mountOnly(doc *html.Node) bool {
	body := findElement(doc, atom.Body)
	if body == nil {
		return false
	}
	any := false
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.CommentNode:
		case html.TextNode:
			if strings.TrimSpace(c.Data) != "" {
				return false
			}
		case html.ElementNode:
			switch {
			case c.DataAtom == atom.Noscript || c.DataAtom == atom.Script:
			case c.DataAtom == atom.Div && emptyElement(c):
			default:
				return false
			}
			any = true
		}
	}
	return any
}

func emptyElement(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode || c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			return false
		}
	}
	return true
}

func findElement(n *html.Node, a atom.Atom) *html.Node {
	if n.Type == html.ElementNode && n.DataAtom == a && n.Namespace == "" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if f := findElement(c, a); f != nil {
			return f
		}
	}
	return nil
}

// collapseSource is HTML's own whitespace rule: a run of spaces, tabs and line breaks in the
// source is one space on the page. Line breaks come from the blocks, never from the source's
// layout — a pretty-printed formula otherwise lands one token to a line.
func collapseSource(s string) string {
	return collapseRuns(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v'
	})
}

// collapseRuns replaces each run of runes space accepts with one space.
func collapseRuns(s string, space func(rune) bool) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if space(r) {
			if !in {
				b.WriteByte(' ')
			}
			in = true
			continue
		}
		in = false
		b.WriteRune(r)
	}
	return b.String()
}

// readerLines turns the walk's text — source whitespace collapsed, a newline at every block
// boundary — into lines: each trimmed, a run of in-line space (no-break space included) one
// space, and at most one blank line in a row.
func readerLines(s string) string {
	var b strings.Builder
	blank := true // no leading blank line
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(collapseRuns(line, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\f' || r == '\v' || r == ' '
		}))
		if line == "" {
			if !blank {
				b.WriteString("\n")
				blank = true
			}
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
		blank = false
	}
	return strings.TrimRight(b.String(), "\n")
}
