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
// same record, for a different question. No network request, no asset directory, no server:
// a reader opens it out of the run's tarball months later and everything works.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

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
	for i, d := range docs {
		bodies[i] = siteLinks(linkIDs(bodies[i], anchor, d.File), files)
	}

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", escape(short))
	b.WriteString("<style>\n" + siteCSS + "</style>\n</head>\n<body>\n")

	fmt.Fprintf(&b, "<header>\n<h1>%s</h1>\n", inline(short))
	if verdict, cls := verdictBadge(board); verdict != "" {
		fmt.Fprintf(&b, "<p class=\"badges\"><span class=\"badge %s\">%s</span>%s</p>\n", cls, escape(verdict), countBadges(board))
	}
	b.WriteString("<nav class=\"tabs\" role=\"tablist\">\n")
	for i, d := range docs {
		sel := ""
		if i == 0 {
			sel = " aria-selected=\"true\""
		}
		fmt.Fprintf(&b, "<button role=\"tab\" data-doc=\"%s\"%s title=\"%s\">%s</button>\n",
			escape(d.File), sel, escape(d.Blurb), escape(d.Nav))
	}
	b.WriteString("</nav>\n")
	b.WriteString("<input id=\"filter\" type=\"search\" placeholder=\"filter this document — type to hide sections that do not match\" autocomplete=\"off\">\n")
	b.WriteString("</header>\n<main>\n")

	for i, d := range docs {
		hidden := " hidden"
		if i == 0 {
			hidden = ""
		}
		fmt.Fprintf(&b, "<article id=\"doc-%s\" data-doc=\"%s\"%s>\n", escape(slugFile(d.File)), escape(d.File), hidden)
		fmt.Fprintf(&b, "<p class=\"blurb\">%s</p>\n", escape(d.Blurb))
		b.WriteString(bodies[i])
		b.WriteString("</article>\n")
	}

	b.WriteString("</main>\n<footer>\n")
	b.WriteString("<p>Rendered from this run's own record by <code>feov-record bench assemble</code>. The markdown set beside this file is the same content, one document per file; <code>records/</code> and <code>proofs/</code> are what both were derived from.</p>\n")
	b.WriteString("</footer>\n<script>\n" + siteJS + "</script>\n</body>\n</html>\n")
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

const siteCSS = `
:root {
  --bg: #fbfbfa; --fg: #1b1b1a; --muted: #5c5c58; --rule: #dedcd6;
  --panel: #ffffff; --accent: #2f5d8a; --code: #f3f2ee;
  --good: #2c6e49; --warn: #8a5a00; --bad: #8a2f2f; --neutral: #4a4a46;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #14151a; --fg: #e6e5e1; --muted: #9c9b95; --rule: #2c2e36;
    --panel: #191b21; --accent: #7fb0e0; --code: #1f222a;
    --good: #7fc0a0; --warn: #d9ac5c; --bad: #e08a8a; --neutral: #a8a8a2;
  }
}
* { box-sizing: border-box; }
body {
  margin: 0; background: var(--bg); color: var(--fg);
  font: 16px/1.65 ui-serif, Georgia, "Times New Roman", serif;
}
header {
  position: sticky; top: 0; z-index: 5; background: var(--bg);
  border-bottom: 1px solid var(--rule); padding: 1rem 1.5rem 0;
}
header h1 { font-size: 1.35rem; margin: 0 0 .4rem; line-height: 1.3; }
.badges { margin: 0 0 .6rem; display: flex; gap: .4rem; flex-wrap: wrap; }
.badge {
  font: 600 11px/1 ui-sans-serif, system-ui, sans-serif; letter-spacing: .06em;
  text-transform: uppercase; padding: .35rem .5rem; border-radius: 3px;
  border: 1px solid currentColor;
}
.badge.good { color: var(--good); } .badge.warn { color: var(--warn); }
.badge.bad { color: var(--bad); } .badge.neutral, .badge.unknown { color: var(--neutral); }
nav.tabs { display: flex; gap: .25rem; flex-wrap: wrap; }
nav.tabs button {
  font: 500 13px/1 ui-sans-serif, system-ui, sans-serif;
  background: none; border: 1px solid transparent; border-bottom: none;
  color: var(--muted); padding: .55rem .8rem; cursor: pointer;
  border-radius: 4px 4px 0 0;
}
nav.tabs button:hover { color: var(--fg); }
nav.tabs button[aria-selected="true"] {
  color: var(--fg); background: var(--panel);
  border-color: var(--rule); margin-bottom: -1px;
}
#filter {
  width: 100%; margin: .6rem 0 .8rem; padding: .45rem .6rem;
  font: 13px ui-sans-serif, system-ui, sans-serif;
  background: var(--panel); color: var(--fg);
  border: 1px solid var(--rule); border-radius: 4px;
}
main { max-width: 46rem; margin: 0 auto; padding: 1.5rem; }
article > .blurb { color: var(--muted); font-style: italic; margin-top: 0; }
h2 { font-size: 1.15rem; margin: 2.2rem 0 .6rem; padding-bottom: .25rem; border-bottom: 1px solid var(--rule); }
h3 { font-size: 1rem; margin: 1.6rem 0 .4rem; }
h4, h5, h6 { font-size: .95rem; margin: 1.2rem 0 .3rem; }
p, li { overflow-wrap: break-word; }
a { color: var(--accent); }
a.idref {
  font: 600 .85em ui-sans-serif, system-ui, sans-serif;
  text-decoration: none; border-bottom: 1px dotted var(--accent);
}
code { background: var(--code); padding: .1em .3em; border-radius: 3px; font-size: .87em; }
pre { background: var(--code); padding: .8rem; border-radius: 4px; overflow-x: auto; }
pre code { background: none; padding: 0; font-size: .8rem; line-height: 1.5; }
blockquote {
  margin: 1rem 0; padding: .4rem 0 .4rem 1rem;
  border-left: 3px solid var(--rule); color: var(--muted);
}
table { border-collapse: collapse; width: 100%; margin: 1rem 0; font-size: .9rem; }
th, td { border: 1px solid var(--rule); padding: .4rem .55rem; text-align: left; vertical-align: top; }
th { background: var(--panel); font: 600 .85rem ui-sans-serif, system-ui, sans-serif; }
hr { border: none; border-top: 1px solid var(--rule); margin: 2rem 0; }
.cont { margin-left: 1.6rem; color: var(--muted); font-size: .93rem; }
sup.fnref a { text-decoration: none; }
.footnotes { margin-top: 3rem; border-top: 1px solid var(--rule); font-size: .9rem; }
.footnotes ol.fn { list-style: none; padding-left: 0; }
.footnotes li { margin: .5rem 0; }
.fnlabel { font: 600 .8rem ui-sans-serif, system-ui, sans-serif; color: var(--muted); }
footer { max-width: 46rem; margin: 0 auto; padding: 0 1.5rem 3rem; color: var(--muted); font-size: .85rem; }
:target { background: color-mix(in srgb, var(--accent) 12%, transparent); }
.hidden-by-filter { display: none; }
`

const siteJS = `
(function () {
  var tabs = Array.prototype.slice.call(document.querySelectorAll('nav.tabs button'));
  var docs = Array.prototype.slice.call(document.querySelectorAll('main article'));
  var filter = document.getElementById('filter');

  function show(file, frag) {
    tabs.forEach(function (t) { t.setAttribute('aria-selected', String(t.dataset.doc === file)); });
    docs.forEach(function (d) { d.hidden = d.dataset.doc !== file; });
    if (frag) {
      var el = document.getElementById(frag);
      if (el) { el.scrollIntoView({ block: 'start' }); }
    } else {
      window.scrollTo(0, 0);
    }
    if (history.replaceState) { history.replaceState(null, '', '#' + file + (frag ? '/' + frag : '')); }
    apply();
  }

  tabs.forEach(function (t) { t.addEventListener('click', function () { show(t.dataset.doc, ''); }); });

  // A cross-document id link switches the tab, THEN jumps. This is the whole reason the site
  // exists: in the markdown set the same link is a file the reader has to open and search.
  document.addEventListener('click', function (e) {
    var a = e.target.closest ? e.target.closest('a.idref') : null;
    if (!a) { return; }
    var doc = a.dataset.doc;
    var frag = a.getAttribute('href').slice(1);
    if (doc) { e.preventDefault(); show(doc, frag); }
  });

  // Filtering hides SECTIONS, not lines: a transcript filtered to matching sentences is a
  // different document. A section is a heading and everything under it up to the next one of
  // the same or higher rank.
  function apply() {
    var q = (filter.value || '').trim().toLowerCase();
    var active = docs.filter(function (d) { return !d.hidden; })[0];
    if (!active) { return; }
    var kids = Array.prototype.slice.call(active.children);
    var group = [], groupText = '', groups = [];
    kids.forEach(function (el) {
      if (/^H[1-3]$/.test(el.tagName) && group.length) {
        groups.push({ els: group, text: groupText });
        group = []; groupText = '';
      }
      group.push(el);
      groupText += ' ' + (el.textContent || '').toLowerCase();
    });
    if (group.length) { groups.push({ els: group, text: groupText }); }
    groups.forEach(function (g) {
      var hit = !q || g.text.indexOf(q) >= 0;
      g.els.forEach(function (el) { el.classList.toggle('hidden-by-filter', !hit); });
    });
  }
  filter.addEventListener('input', apply);

  var hash = (location.hash || '').slice(1);
  if (hash) {
    var parts = hash.split('/');
    if (docs.some(function (d) { return d.dataset.doc === parts[0]; })) { show(parts[0], parts[1] || ''); }
  }
})();
`
