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
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

//go:embed site.css
var siteCSS string

//go:embed site.js
var siteJS string

//go:embed site.html.tmpl
var siteHTML string

var siteTmpl = template.Must(template.New("site").Parse(siteHTML))

// sitePage is the template's whole world. Bodies and badges arrive as template.HTML because
// they are this package's OWN rendered output — the template escapes everything else.
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
	File, Slug, Nav, Blurb string
	Selected               bool
	Body                   template.HTML
}

// RenderSite builds the whole set into one HTML file. It takes the RAW document bodies (no
// markdown link bar, no repeated H1) because the tab strip and the header say those things
// better here.
func RenderSite(title string, docs []Doc, board *record.Board) string {
	short := strings.TrimSpace(strings.TrimPrefix(title, "# "))

	// TWO PASSES, and the reason is the join: pass one renders each document and records
	// where every record id is DEFINED; pass two links every mention across the whole set. A
	// one-pass render can only link backwards, which is the half of the joins nobody needs.
	anchor := anchors{}
	bodies := make([]string, len(docs))
	for i, d := range docs {
		bodies[i] = mdToHTML(d.Body, d.File, anchor)
	}
	files := map[string]bool{}
	for _, d := range docs {
		files[d.File] = true
	}

	page := sitePage{
		TitleText: short,
		TitleHTML: template.HTML(inline(short)),
		MultiDoc:  len(docs) > 1,
		CSS:       template.CSS(siteCSS),
		JS:        template.JS(strings.Replace(siteJS, "MERMAID_CDN_URL", mermaidCDN, 1)),
	}
	if verdict, cls := verdictBadge(board); verdict != "" {
		page.Badges = template.HTML(fmt.Sprintf("<span class=\"badge %s\">%s</span>%s", cls, escape(verdict), countBadges(board)))
	}
	for i, d := range docs {
		body := siteLinks(linkIDs(bodies[i], anchor, d.File), files)
		if d.File == FileRun {
			// The trajectory opens the run document: "how did the board converge" is the
			// first question behind the verdict, and it is a picture. Markdown cannot carry
			// it (GitHub strips inline SVG), so it is drawn HERE, in the reading tier, while
			// the durable tier keeps the same numbers as text.
			if c := boardChart(board); c != "" {
				body = "<h2>The board, by round</h2>\n" + c + body
			}
		}
		page.Docs = append(page.Docs, siteDoc{
			File: d.File, Slug: slugFile(d.File), Nav: d.Nav, Blurb: d.Blurb,
			Selected: i == 0,
			Body:     template.HTML(body),
		})
	}

	var b strings.Builder
	if err := siteTmpl.Execute(&b, page); err != nil {
		// The template and its data are both compiled in; an execute error is a programming
		// error, and a panic in tests is how it gets found.
		panic(fmt.Sprintf("report: site template: %v", err))
	}
	return b.String()
}

// siteLinks rewrites the markdown set's own cross-file links — `[the docket](docket.md)`,
// `[evidence.md](evidence.md#p-deadbeef)` — into tab switches.
//
// WITHOUT THIS THEY ARE DEAD. In the markdown tier they are exactly right: a relative link to
// a sibling file. In one HTML file there is no sibling file, so the browser would try to open
// a document that is already open, in a pane it cannot reach. Same links, two tiers, and only
// one of them can follow a path.
func siteLinks(html string, files map[string]bool) string {
	return reHref.ReplaceAllStringFunc(html, func(m string) string {
		target := reHref.FindStringSubmatch(m)[1]
		file, frag, _ := strings.Cut(target, "#")
		if !files[file] {
			return m
		}
		return fmt.Sprintf(`<a class="idref" data-doc="%s" href="#%s"`, escape(file), escape(frag))
	})
}

var reHref = regexp.MustCompile(`<a href="([^"]+)"`)

// slugFile turns "CHANGELOG.md" into an element id fragment.
func slugFile(f string) string {
	f = strings.TrimSuffix(f, ".md")
	return strings.ToLower(strings.ReplaceAll(f, ".", "-"))
}

// verdictBadge is the one fact the header exists for, and it comes off the RECORD. The colour
// is a rendering of the verdict, never a judgement added to it: a ceiling termination is not a
// failure, and must not be painted as one.
func verdictBadge(board *record.Board) (string, string) {
	if board == nil {
		return "", ""
	}
	o := outcomeOf(board.Events)
	if o == nil {
		return "no terminal outcome recorded", "unknown"
	}
	word := verdictWord(o)
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
func countBadges(board *record.Board) string {
	open, closed := 0, 0
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
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
