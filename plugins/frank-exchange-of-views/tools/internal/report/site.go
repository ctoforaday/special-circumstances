package report

// ONE FILE, REAL TABS, NO BUILD STEP.
//
// The markdown set is the durable tier: greppable, diffable, reviewable in a pull request,
// readable in any editor. This is the READING tier, and it exists because three things the
// research set needs cannot be said in markdown at all — there are no tabs in CommonMark or
// GitHub-flavored markdown, no cross-document link that lands the reader in the right pane,
// and no filter over a 100 KB transcript.
//
// It is not a site generator and it does not become one. dashboard.html is already a
// self-contained Go-rendered HTML+SVG artifact in this tool; this is the same shape, over the
// same record, for a different question. No network request at load, no asset directory, no
// server: a reader opens it out of the run's tarball months later and everything works. The
// style, script and skeleton live in site.css / site.js / site.html.tmpl, COMPILED INTO the
// binary by go:embed — files so they are edited as what they are, embedded so the artifact
// still needs nothing beside itself.

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

//go:embed site.css
var siteCSS string

//go:embed site.js
var siteJS string

//go:embed site.html.tmpl
var siteHTML string

var siteTmpl = template.Must(template.New("site").Parse(siteHTML))

// sitePage is the template's whole world. The template escapes every string; what arrives as
// template.HTML is this package's own rendering, and every such fragment that holds a word of the
// record's text — the title, the badges, each body — is made only by pageHTML, so the allowlist
// stands over all of them. The board chart is the one fragment outside it: the tool draws it from
// the board's counts and it holds no text but numbers and the tool's own labels.
type sitePage struct {
	TitleText string
	TitleHTML template.HTML
	Badges    template.HTML
	MultiDoc  bool
	Docs      []siteDoc
	CSS       template.CSS
	JS        template.JS
}

type siteDoc struct {
	File, Element, Nav, Blurb string // Element: the id of the article that holds the document (docElement)
	Selected                  bool
	Chart                     template.HTML // the board chart, the tool's drawing of counts (boardChart)
	Body                      template.HTML
}

// pageHTML is a fragment of the renderer's output as the page carries it: held to bodyPolicy, the
// allowlist of exactly what the renderer writes. On honest output it changes nothing a browser reads;
// it is the second wall, should the renderer ever let a seat's markup through.
func pageHTML(s string) template.HTML { return template.HTML(bodyPolicy.Sanitize(s)) }

// RenderSite builds the whole set into one HTML file. It renders each document's body as assembly
// composed it — the record's text, unescaped, in the tree Ship parsed it to with no HTML grammar —
// and not the shipped .md: no markdown link bar and no repeated H1, because the tab strip and the
// header say those things better here, and none of the .md's backslashes.
func RenderSite(set Shipped, fam record.Family) string { return renderSite(set, fam, stdRenderer()) }

// renderer is what renderSite writes each document's body and the title with.
type renderer struct {
	body  func(p ParsedText, file string, anchor anchors, seen map[string]int, files map[string]bool) rendered
	title func(string) string
}

// stdRenderer is the renderer the page is written with (md.go).
func stdRenderer() renderer { return renderer{body: renderDoc, title: inline} }

// renderSite is RenderSite with rr writing each fragment that holds the record's text.
func renderSite(set Shipped, fam record.Family, rr renderer) string {
	docs := set.docs
	short := shortTitle(set)
	title, bodies := renderSet(set, rr)
	page := sitePage{
		TitleText: short,
		TitleHTML: pageHTML(title),
		MultiDoc:  len(docs) > 1,
		CSS:       template.CSS(siteCSS),
		JS:        template.JS(strings.Replace(siteJS, "MERMAID_CDN_URL", mermaidCDN, 1)),
	}
	if outcome, cls := outcomeBadge(fam); outcome != "" {
		page.Badges = pageHTML(fmt.Sprintf("<span class=\"badge %s\">%s</span>%s", cls, escape(outcome), countBadges(fam)))
	}
	for i, d := range docs {
		doc := siteDoc{
			File: d.File, Element: docElement(d.File), Nav: d.Nav, Blurb: d.Blurb,
			Selected: i == 0,
			Body:     pageHTML(bodies[i]),
		}
		if d.File == FileRun {
			// The trajectory opens the run document: "how did the board converge" is the
			// first question behind the verdict, and it is a picture. Markdown cannot carry
			// it (GitHub strips inline SVG), so it is drawn HERE, in the reading tier, while
			// the durable tier keeps the same numbers as text.
			if c := boardChart(fam); c != "" {
				doc.Chart = template.HTML("<h2>The board, by epoch</h2>\n" + c)
			}
		}
		page.Docs = append(page.Docs, doc)
	}

	var b strings.Builder
	if err := siteTmpl.Execute(&b, page); err != nil {
		// The template and its data are both compiled in; an execute error is a programming
		// error, and a panic in tests is how it gets found.
		panic(fmt.Sprintf("site template: %v", err))
	}
	return b.String()
}

// shortTitle is the run's title as the page's header and <title> show it.
func shortTitle(set Shipped) string { return strings.TrimSpace(strings.TrimPrefix(set.title, "# ")) }

// renderSet is each fragment of the page that holds the record's text, as the renderer writes it and
// before the wall (pageHTML): the title, and each document's body with its ids linked across the set.
func renderSet(set Shipped, rr renderer) (title string, bodies []string) {
	// TWO PASSES, and the reason is the join: pass one renders each document and records
	// where every record id is DEFINED; pass two links every mention across the whole set. A
	// one-pass render can only link backwards, which is the half of the joins nobody needs.
	anchor := anchors{}
	files := map[string]bool{}
	for _, d := range set.docs {
		files[d.File] = true
	}
	docs := make([]rendered, len(set.docs))
	slugs := reservedIDs(set.docs)
	for i, d := range set.docs {
		docs[i] = rr.body(set.html[i], d.File, anchor, slugs, files)
	}
	for i, d := range set.docs {
		bodies = append(bodies, linkIDs(docs[i], anchor, d.File))
	}
	return rr.title(shortTitle(set)), bodies
}

// reservedIDs is the ids the page gives out before a document is rendered, each counted once, so a
// heading's id is never one of them: the page's own (pageIDs) and each document's article
// (docElement). A heading "Doc report" is "doc-report-2", and a fragment naming the report's article
// still lands on it. The citation layer takes its ids from the same count (citations).
func reservedIDs(docs []Doc) map[string]int {
	seen := map[string]int{}
	for _, id := range pageIDs {
		seen[id] = 1
	}
	for _, d := range docs {
		seen[docElement(d.File)] = 1
	}
	return seen
}

// pageIDs is every id the page's skeleton (site.html.tmpl) writes as it stands, read from it.
var pageIDs = func() []string {
	var ids []string
	z := html.NewTokenizer(strings.NewReader(siteHTML))
	for tt := z.Next(); tt != html.ErrorToken; tt = z.Next() {
		for _, a := range z.Token().Attr {
			if a.Key == "id" && !strings.Contains(a.Val, "{{") {
				ids = append(ids, a.Val)
			}
		}
	}
	return ids
}()

// siteLink opens a link to target: the set's own cross-file links — `[the board](docket.md)`,
// `[evidence.md](evidence.md#p-deadbeef)` — as tab switches, any other as itself.
//
// WITHOUT THIS THEY ARE DEAD. In the markdown tier they are exactly right: a relative link to
// a sibling file. In one HTML file there is no sibling file, so the browser would try to open
// a document that is already open, in a pane it cannot reach. Same links, two tiers, and only
// one of them can follow a path.
//
// target is a URL (linkURL), escaped here, once, for the attribute. A link to a document with no
// fragment names the document's own element, the article site.js opens at its top: an empty
// fragment is no URL to the allowlist, and a link with no href is no link.
func siteLink(target string, files map[string]bool) string {
	file, frag, _ := strings.Cut(target, "#")
	if !files[file] {
		return `<a href="` + escape(target) + `">`
	}
	if frag == "" {
		frag = docElement(file)
	}
	return `<a class="idref" data-doc="` + escape(file) + `" href="#` + escape(frag) + `">`
}

// docElement is the id of the element that holds file's document on the page (site.html.tmpl).
func docElement(file string) string { return "doc-" + slugFile(file) }

// slugFile turns "CHANGELOG.md" into an element id fragment.
func slugFile(f string) string {
	f = strings.TrimSuffix(f, ".md")
	return strings.ToLower(strings.ReplaceAll(f, ".", "-"))
}

// outcomeBadge is the one fact the header exists for, and it comes off the RECORD. The colour
// is a rendering of the verdict, never a judgement added to it: a ceiling termination is not a
// failure, and must not be painted as one.
func outcomeBadge(fam record.Family) (string, string) {
	o := outcomeOf(fam.Events)
	if o == nil {
		return "PENDING", "unknown"
	}
	word := outcomeWord(o)
	cls := "neutral"
	switch {
	case strings.HasPrefix(word, "VERIFIED"):
		cls = "good"
	case strings.HasPrefix(word, "HALTED"):
		cls = "bad"
	case strings.HasPrefix(word, "CEILING"), strings.HasPrefix(word, "UNVERIFIED"):
		cls = "warn"
	}
	return word, cls
}

// countBadges puts the board's shape beside the verdict: how many gaps are still open is the
// second question every reader asks and the single report made them scroll for it.
func countBadges(fam record.Family) string {
	open, closed := 0, 0
	for _, g := range fam.Gaps {
		if g == nil {
			continue
		}
		if g.Open {
			open++
		} else {
			closed++
		}
	}
	if open == 0 && closed == 0 {
		return ""
	}
	cls := "good"
	if open > 0 {
		cls = "warn"
	}
	return fmt.Sprintf("<span class=\"badge %s\">%d open</span><span class=\"badge neutral\">%d closed</span>", cls, open, closed)
}

// mermaidCDN is the ONE address this artifact may reach, and only when a document actually
// carries a ```mermaid fence. Everything else — style, script, fonts — is inline: the page
// still opens complete from a tarball with no network, and an unreachable CDN degrades to the
// diagram's source text with a note, never to a broken page. The version is pinned so the
// artifact renders the same way years later or not at all — never differently.
const mermaidCDN = "https://cdn.jsdelivr.net/npm/mermaid@11.12.0/dist/mermaid.esm.min.mjs"

// linkSchemes are the schemes a link in a document may name; any other renders as a link to nothing
// (linkURL), and bodyPolicy drops it all the same.
var linkSchemes = []string{"http", "https", "mailto"}

// bodyPolicy is every element and attribute the document renderer (md.go, citations, linkIDs,
// siteLink) writes, and nothing else. The renderer passes no HTML of the text through, so on honest
// output the policy changes nothing; it is the second wall, should the first ever let a tag through.
// An element the renderer starts writing goes here in the same change, or the page loses it.
var bodyPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "hr", "pre", "blockquote", "ul", "em", "strong", "del",
		"table", "thead", "tbody", "tr", "th", "td")
	p.AllowAttrs("id").OnElements("h1", "h2", "h3", "h4", "h5", "h6", "li")
	p.AllowAttrs("start", "class").OnElements("ol")
	p.AllowAttrs("class").OnElements("code", "div", "sup", "section")
	p.AllowAttrs("class", "data-nolink").OnElements("span")
	p.AllowAttrs("href", "id", "class", "title", "data-doc", "aria-label").OnElements("a")
	p.AllowURLSchemes(linkSchemes...)
	p.AllowRelativeURLs(true)
	return p
}()
