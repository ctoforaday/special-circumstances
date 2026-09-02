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
	b.WriteString("<meta name=\"color-scheme\" content=\"light dark\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", escape(short))
	b.WriteString("<style>\n" + siteCSS + "</style>\n</head>\n<body>\n")

	fmt.Fprintf(&b, "<header class=\"masthead\">\n<h1>%s</h1>\n", inline(short))
	if verdict, cls := verdictBadge(board); verdict != "" {
		fmt.Fprintf(&b, "<p class=\"badges\"><span class=\"badge %s\">%s</span>%s</p>\n", cls, escape(verdict), countBadges(board))
	}
	b.WriteString("</header>\n")

	// The chrome bar is the STICKY tier and the masthead is not: mid-document the reader keeps
	// the tab strip and the filter, while the title — already read — scrolls away and gives a
	// phone its screen back.
	b.WriteString("<div class=\"chrome\">\n")
	if len(docs) > 1 {
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
	}
	b.WriteString("<input id=\"filter\" type=\"search\" placeholder=\"filter sections…\" aria-label=\"filter this document — type to hide sections that do not match\" autocomplete=\"off\">\n")
	b.WriteString("</div>\n")

	b.WriteString("<div class=\"layout\">\n<nav id=\"toc\" aria-label=\"Contents\" hidden></nav>\n<main>\n")

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
	b.WriteString("</footer>\n<script>\n" + strings.Replace(siteJS, "MERMAID_CDN_URL", mermaidCDN, 1) + "</script>\n</body>\n</html>\n")
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

const siteCSS = `
:root {
  --bg: #fbfbfa; --fg: #1b1b1a; --muted: #5c5c58; --rule: #dedcd6; --rule-strong: #b9b7b0;
  --panel: #ffffff; --accent: #2f5d8a; --code: #f3f2ee;
  --good: #2c6e49; --warn: #8a5a00; --bad: #8a2f2f; --neutral: #4a4a46;
  --sans: ui-sans-serif, system-ui, "Segoe UI", Roboto, sans-serif;
  --serif: ui-serif, Georgia, "Times New Roman", serif;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #14151a; --fg: #e6e5e1; --muted: #9c9b95; --rule: #2c2e36; --rule-strong: #4a4d59;
    --panel: #191b21; --accent: #7fb0e0; --code: #1f222a;
    --good: #7fc0a0; --warn: #d9ac5c; --bad: #e08a8a; --neutral: #a8a8a2;
  }
}
* { box-sizing: border-box; }
html { color-scheme: light dark; }
body { margin: 0; background: var(--bg); color: var(--fg); font: 16px/1.6 var(--sans); }
.shellpad { padding-inline: clamp(1rem, 4vw, 2rem); }

header.masthead { border-bottom: 1px solid var(--rule); }
header.masthead h1 {
  max-width: 72rem; margin: 0 auto; padding: 1.1rem clamp(1rem, 4vw, 2rem) .4rem;
  font: 650 clamp(1.25rem, 1.05rem + 1vw, 1.6rem)/1.3 var(--sans); letter-spacing: -.01em;
}
.badges {
  max-width: 72rem; margin: 0 auto; padding: 0 clamp(1rem, 4vw, 2rem) .8rem;
  display: flex; gap: .4rem; flex-wrap: wrap;
}
.badge {
  font: 600 11px/1 var(--sans); letter-spacing: .06em; text-transform: uppercase;
  padding: .35rem .5rem; border-radius: 3px; border: 1px solid currentColor;
}
.badge.good { color: var(--good); } .badge.warn { color: var(--warn); }
.badge.bad { color: var(--bad); } .badge.neutral, .badge.unknown { color: var(--neutral); }

.chrome {
  position: sticky; top: 0; z-index: 10; background: var(--bg);
  border-bottom: 1px solid var(--rule);
}
.chrome::after { content: ""; display: block; clear: both; }
.chrome { display: flex; align-items: center; gap: 1rem; flex-wrap: wrap;
  max-width: 72rem; margin: 0 auto; padding: 0 clamp(1rem, 4vw, 2rem); }
nav.tabs {
  display: flex; flex: 1 1 auto; min-width: 0; overflow-x: auto; scrollbar-width: none;
}
nav.tabs::-webkit-scrollbar { display: none; }
nav.tabs button {
  font: 600 13px/1 var(--sans); white-space: nowrap; background: none; cursor: pointer;
  color: var(--muted); padding: .8rem .15rem calc(.8rem - 2px); margin-right: 1.15rem;
  border: none; border-bottom: 2px solid transparent;
}
nav.tabs button:hover { color: var(--fg); }
nav.tabs button[aria-selected="true"] { color: var(--fg); border-bottom-color: var(--accent); }
#filter {
  flex: 0 1 13rem; min-width: 9rem; margin: .45rem 0; padding: .4rem .6rem;
  font: 13px var(--sans); background: var(--panel); color: var(--fg);
  border: 1px solid var(--rule); border-radius: 4px;
}
@media (max-width: 480px) { #filter { flex: 1 1 100%; margin-top: 0; } }

.layout { max-width: 72rem; margin: 0 auto; }
#toc { display: none; }
main { max-width: 46rem; margin: 0 auto; padding: 1.5rem clamp(1rem, 4vw, 2rem) 3rem; }
@media (min-width: 1100px) {
  .layout {
    display: grid; grid-template-columns: minmax(11rem, 14rem) minmax(0, 46rem);
    justify-content: center; gap: clamp(2rem, 5vw, 5rem); padding-inline: 2rem;
  }
  #toc:not([hidden]) {
    display: block; position: sticky; top: 4.4rem; align-self: start;
    max-height: calc(100vh - 6rem); overflow-y: auto;
    font: .82rem/1.5 var(--sans); padding: 1.6rem 0 1rem;
  }
  main { margin: 0; padding-inline: 0; }
}
#toc .toc-title {
  font: 650 .72rem/1 var(--sans); letter-spacing: .09em; text-transform: uppercase;
  color: var(--muted); margin: 0 0 .7rem;
}
#toc ol { list-style: none; margin: 0; padding: 0; border-left: 1px solid var(--rule); }
#toc li { margin: 0; }
#toc a {
  display: block; color: var(--muted); text-decoration: none;
  padding: .28rem 0 .28rem .85rem; margin-left: -1px; border-left: 2px solid transparent;
}
#toc a:hover { color: var(--fg); }
#toc a.active { color: var(--accent); border-left-color: var(--accent); }

article { font-family: var(--serif); font-size: clamp(1rem, .96rem + .25vw, 1.0625rem); line-height: 1.7; }
article > .blurb { font: .9rem/1.55 var(--sans); color: var(--muted); margin: .2rem 0 1.6rem; }
article h1 {
  font: 650 clamp(1.4rem, 1.15rem + 1.3vw, 1.8rem)/1.3 var(--sans);
  letter-spacing: -.015em; margin: 1.4rem 0 1rem;
}
h2 {
  font: 650 1.25rem/1.35 var(--sans); letter-spacing: -.01em;
  margin: 2.6rem 0 .8rem; padding-bottom: .3rem; border-bottom: 1px solid var(--rule);
}
h3 { font: 600 1.05rem/1.4 var(--sans); margin: 1.8rem 0 .5rem; }
h4, h5, h6 {
  font: 650 .8rem/1.4 var(--sans); letter-spacing: .07em; text-transform: uppercase;
  color: var(--muted); margin: 1.4rem 0 .4rem;
}
p { margin: 0 0 1rem; }
p, li { overflow-wrap: break-word; }
a { color: var(--accent); text-underline-offset: 2px; }
a.idref {
  font: 600 .85em var(--sans); text-decoration: none;
  border-bottom: 1px dotted var(--accent);
}
code { background: var(--code); padding: .1em .3em; border-radius: 3px; font-size: .86em; }
pre {
  background: var(--code); border: 1px solid var(--rule); border-radius: 6px;
  padding: .85rem 1rem; overflow-x: auto;
}
pre code { background: none; padding: 0; font-size: .8rem; line-height: 1.55; }
blockquote {
  margin: 1.2rem 0; padding: .1rem 0 .1rem 1.1rem;
  border-left: 3px solid var(--rule-strong); color: var(--muted);
}
.tblwrap { overflow-x: auto; margin: 1.25rem 0; }
table { border-collapse: collapse; width: 100%; font: .875rem/1.5 var(--sans); }
th, td { padding: .5rem .6rem; text-align: left; vertical-align: top; }
th {
  font: 650 .78rem/1.4 var(--sans); letter-spacing: .04em; text-transform: uppercase;
  color: var(--muted); border-bottom: 2px solid var(--rule-strong);
}
td { border-bottom: 1px solid var(--rule); }
hr { border: none; border-top: 1px solid var(--rule); margin: 2.4rem 0; }
.cont { margin-left: 1.6rem; color: var(--muted); font-size: .93rem; }

sup.fnref { font-size: .66em; line-height: 0; }
sup.fnref a {
  font: 600 1em var(--sans); letter-spacing: .01em; text-decoration: none;
  padding: .12em .3em; border-radius: 3px;
  background: color-mix(in srgb, var(--accent) 9%, transparent);
}
.fnpop {
  font: .85rem/1.55 var(--sans); background: var(--panel);
  border: 1px solid var(--rule); border-left: 3px solid var(--accent); border-radius: 6px;
  padding: .7rem .9rem; margin: .6rem 0 1rem; overflow-wrap: anywhere;
  box-shadow: 0 4px 16px rgb(0 0 0 / .1);
}
.fnpop .fnback { display: none; }
.footnotes { margin-top: 3.5rem; font-size: .9rem; }
.footnotes ol.fn { list-style: none; padding-left: 0; }
.footnotes li { margin: .6rem 0; overflow-wrap: anywhere; }
.fnlabel { font: 650 .78rem var(--sans); color: var(--accent); margin-right: .3rem; }
.fnback { font-family: var(--sans); text-decoration: none; margin-left: .3rem; }

figure.mermaid { margin: 1.5rem 0; text-align: center; }
figure.mermaid svg { max-width: 100%; height: auto; }
figure.mermaid details { text-align: left; margin-top: .5rem; }
figure.mermaid summary, .mermaid-note { font: .78rem var(--sans); color: var(--muted); cursor: pointer; }
.mermaid-note { margin: 1.2rem 0 .3rem; cursor: default; }

footer {
  border-top: 1px solid var(--rule); margin-top: 2rem;
  color: var(--muted); font: .82rem/1.6 var(--sans);
}
footer p { max-width: 72rem; margin: 0 auto; padding: 1.2rem clamp(1rem, 4vw, 2rem) 2.5rem; }
[id] { scroll-margin-top: 4.6rem; }
:target { background: color-mix(in srgb, var(--accent) 12%, transparent); }
.hidden-by-filter { display: none; }
@media print {
  .chrome, #toc, footer, .fnpop { display: none !important; }
  body { background: #fff; color: #000; font-size: 11pt; }
  header.masthead h1, .badges, main, footer p { max-width: 100%; padding-inline: 0; }
  a { color: inherit; }
  sup.fnref a { background: none; }
  pre, .tblwrap { overflow: visible; }
  pre { white-space: pre-wrap; }
}
`

const siteJS = `
(function () {
  var slice = function (x) { return Array.prototype.slice.call(x); };
  var tabs = slice(document.querySelectorAll('nav.tabs button'));
  var docs = slice(document.querySelectorAll('main article'));
  var filter = document.getElementById('filter');
  var toc = document.getElementById('toc');

  function active() { return docs.filter(function (d) { return !d.hidden; })[0]; }

  // ---- table of contents (wide screens): built from the ACTIVE document's h2 spine ----
  var tocLinks = [];
  function buildToc(d) {
    if (!toc || !d) { return; }
    var hs = slice(d.querySelectorAll('h2'));
    toc.innerHTML = ''; tocLinks = [];
    if (hs.length < 2) { toc.hidden = true; return; }
    var title = document.createElement('p');
    title.className = 'toc-title'; title.textContent = 'Contents';
    toc.appendChild(title);
    var ol = document.createElement('ol');
    hs.forEach(function (h, i) {
      if (!h.id) { h.id = 's-' + (d.dataset.doc || 'doc').replace(/[^a-z0-9]/gi, '') + '-' + i; }
      var li = document.createElement('li'); var a = document.createElement('a');
      a.href = '#' + h.id; a.textContent = h.textContent;
      li.appendChild(a); ol.appendChild(li); tocLinks.push({ a: a, h: h });
    });
    toc.appendChild(ol); toc.hidden = false;
    spy();
  }
  function spy() {
    if (!tocLinks.length) { return; }
    var current = tocLinks[0];
    tocLinks.forEach(function (t) { if (t.h.getBoundingClientRect().top <= 90) { current = t; } });
    tocLinks.forEach(function (t) { t.a.className = t === current ? 'active' : ''; });
  }
  var ticking = false;
  window.addEventListener('scroll', function () {
    if (ticking) { return; }
    ticking = true;
    requestAnimationFrame(function () { spy(); ticking = false; });
  }, { passive: true });

  function show(file, frag) {
    tabs.forEach(function (t) { t.setAttribute('aria-selected', String(t.dataset.doc === file)); });
    docs.forEach(function (d) { d.hidden = d.dataset.doc !== file; });
    closePop();
    if (frag) {
      var el = document.getElementById(frag);
      if (el) { el.scrollIntoView({ block: 'start' }); }
    } else {
      window.scrollTo(0, 0);
    }
    if (history.replaceState) { history.replaceState(null, '', '#' + file + (frag ? '/' + frag : '')); }
    apply();
    buildToc(active());
    mermaidize(active());
  }

  tabs.forEach(function (t) { t.addEventListener('click', function () { show(t.dataset.doc, ''); }); });
  tabs.forEach(function (t, i) {
    t.addEventListener('keydown', function (e) {
      var j = e.key === 'ArrowRight' ? i + 1 : e.key === 'ArrowLeft' ? i - 1 : -1;
      if (j < 0 || j >= tabs.length) { return; }
      tabs[j].focus(); show(tabs[j].dataset.doc, '');
    });
  });

  // A cross-document id link switches the tab, THEN jumps. This is the whole reason the site
  // exists: in the markdown set the same link is a file the reader has to open and search.
  document.addEventListener('click', function (e) {
    var a = e.target.closest ? e.target.closest('a.idref') : null;
    if (!a) { return; }
    var doc = a.dataset.doc;
    var frag = a.getAttribute('href').slice(1);
    if (doc) { e.preventDefault(); show(doc, frag); }
  });

  // ---- citations: a tap on the in-text marker opens the References entry IN PLACE ----
  var pop = null;
  function closePop() { if (pop && pop.parentNode) { pop.parentNode.removeChild(pop); } pop = null; }
  document.addEventListener('click', function (e) {
    var a = e.target.closest ? e.target.closest('sup.fnref a') : null;
    if (!a) { return; }
    var art = active();
    if (!art) { return; }
    var id = a.getAttribute('href').slice(1);
    var li = art.querySelector('[id="' + id + '"]');
    if (!li) { return; }
    e.preventDefault();
    var again = pop && pop.dataset.fn === id;
    closePop();
    if (again) { return; }
    pop = document.createElement('div');
    pop.className = 'fnpop'; pop.dataset.fn = id;
    pop.innerHTML = li.innerHTML;
    var host = e.target.closest('p, li, td, blockquote') || a.parentNode.parentNode;
    host.insertAdjacentElement('afterend', pop);
  });

  // Filtering hides SECTIONS, not lines: a transcript filtered to matching sentences is a
  // different document. A section is a heading and everything under it up to the next one of
  // the same or higher rank.
  function apply() {
    var q = (filter.value || '').trim().toLowerCase();
    var act = active();
    if (!act) { return; }
    var kids = slice(act.children);
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

  // ---- mermaid: OPTIONAL enhancement over a page that is already complete without it ----
  // The import is dynamic and pinned; offline or blocked, the fenced source stays exactly
  // where it was with a note saying what it is. new Function keeps the syntax out of old
  // parsers so the rest of this script survives them.
  var dynImport = null;
  try { dynImport = new Function('u', 'return import(u)'); } catch (err) { dynImport = null; }
  var merLib = null;
  function mermaidize(d) {
    if (!d) { return; }
    var codes = slice(d.querySelectorAll('pre > code.language-mermaid')).filter(function (c) {
      return !c.parentNode.dataset.mmDone;
    });
    if (!codes.length) { return; }
    if (!dynImport) { codes.forEach(function (c) { note(c.parentNode); }); return; }
    merLib = merLib || dynImport('MERMAID_CDN_URL').then(function (m) {
      var mm = m.default;
      mm.initialize({
        startOnLoad: false, securityLevel: 'strict',
        theme: window.matchMedia && matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'neutral'
      });
      return mm;
    });
    merLib.then(function (mm) {
      codes.forEach(function (c, i) {
        var pre = c.parentNode;
        if (pre.dataset.mmDone) { return; }
        pre.dataset.mmDone = '1';
        mm.render('mm-' + Date.now() + '-' + i, c.textContent).then(function (r) {
          var fig = document.createElement('figure');
          fig.className = 'mermaid'; fig.innerHTML = r.svg;
          var det = document.createElement('details');
          det.innerHTML = '<summary>diagram source</summary>';
          pre.parentNode.insertBefore(fig, pre);
          det.appendChild(pre); fig.appendChild(det);
        }).catch(function () { note(pre); });
      });
    }).catch(function () {
      merLib = null;
      codes.forEach(function (c) { note(c.parentNode); });
    });
  }
  function note(pre) {
    if (pre.dataset.mmNote) { return; }
    pre.dataset.mmNote = '1';
    var p = document.createElement('p');
    p.className = 'mermaid-note';
    p.textContent = 'diagram source — rendering it needs network access this copy did not have';
    pre.parentNode.insertBefore(p, pre);
  }

  var hash = (location.hash || '').slice(1);
  var parts = hash.split('/');
  if (hash && docs.some(function (d) { return d.dataset.doc === parts[0]; })) {
    show(parts[0], parts[1] || '');
  } else {
    buildToc(active());
    mermaidize(active());
  }
})();
`
