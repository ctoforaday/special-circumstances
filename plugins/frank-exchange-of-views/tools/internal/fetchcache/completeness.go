package fetchcache

import (
	"bytes"
	"regexp"
	"strings"
)

// WHETHER AN HTML PAGE IS THE PAPER OR ITS ABSTRACT, decided only on POSITIVE evidence.
//
// MEASURED 2026-09-25 against a hand-read ground truth of 26 distinct pages this tool had recorded
// as documents: 4 were the paper, 19 were its abstract, 7 were something else (a sign-in form, a
// book's sales page). No metadata flag separates them. `isAccessibleForFree` describes the
// licence, not the response — a free article whose body lives only in its pdf says `true` over an
// abstract. A wall's presence means not-full; its ABSENCE means nothing (7 of the 19 abstracts
// carried no wall at all). And the abstract page is often the LARGER one: Nature's paywalled page
// is 358 KB of navigation around a 177-word abstract.
//
// So the question is asked the one way the ground truth could answer it: does the PLATFORM that
// rendered this page mark a body section? Each platform names its body sections in its own
// markup, and an abstract page has none of them before the reference list. Where no platform
// this tool knows rendered the page, the answer is `unverified` — which costs a further look for
// a better copy, and never claims a paper was read that was not.
//
// THE LIST IS INCOMPLETE BY DESIGN, and incomplete fails safe. A platform missing here yields
// `unverified`, the page is kept as the fallback, and the record says nobody could tell. What
// would NOT fail safe is a marker that says `full` over an abstract — which is why a marker is
// added only once a page from that platform has been read by hand, and the corpus in
// testdata/bodies scores every rule against those readings.
const (
	// CompletenessFull: the platform marks body sections before the reference list.
	CompletenessFull = "full"
	// CompletenessAbstract: the page is the work's own landing page and the body is not on it —
	// the platform's paywall is showing, or the platform renders the work as images.
	CompletenessAbstract = "abstract"
	// CompletenessUnverified: a scholarly page no known marker speaks for.
	CompletenessUnverified = "unverified"
	// CompletenessNotTheWork: the page is not the work in any part — a sign-in form, a book's
	// sales page. Citing it even as an abstract would cite text the author did not write.
	CompletenessNotTheWork = "not_the_work"
)

// Completeness classifies an html page reached for a scholarly work. It answers "" — not asked —
// for anything that is not html, and for a page that neither was reached through a doi nor
// declares itself a scholarly work: a blog post or a standards page has no abstract to be
// confused with, and asking the question there would send every ordinary page looking for a
// "better copy" that does not exist.
func Completeness(contentType string, body []byte, reachedForAWork bool) (verdict, reason string) {
	if !strings.Contains(contentType, "html") || len(body) == 0 {
		return "", ""
	}
	declared := citationTitle.Match(body)
	if !reachedForAWork && !declared {
		return "", ""
	}
	// NOT THE WORK first: a book's sales page carries the platform's paywall text too, and calling
	// it the book's abstract would put the publisher's blurb in the author's mouth.
	if ogBook.Match(body) {
		return CompletenessNotTheWork, "the page declares itself a book's landing page (og:type book): a description and a table of contents, not the work's text"
	}
	if !declared && passwordInput.Match(body) {
		return CompletenessNotTheWork, "a sign-in form carrying no scholarly metadata: the source answered with its login page, not the work"
	}
	for _, w := range abstractWalls {
		if w.re.Match(body) {
			return CompletenessAbstract, w.reason
		}
	}
	for _, p := range bodyMarkers {
		if n := p.sections(body); n > 0 {
			return CompletenessFull, p.reason
		}
	}
	return CompletenessUnverified, "no platform marker this tool knows says whether the body is on this page; it is kept, and a fuller copy is looked for"
}

var (
	citationTitle = regexp.MustCompile(`(?i)<meta[^>]+name\s*=\s*["']?citation_title["'\s>]`)
	ogBook        = regexp.MustCompile(`(?i)<meta[^>]+property\s*=\s*["']og:type["'][^>]*content\s*=\s*["']book["']`)
	passwordInput = regexp.MustCompile(`(?i)<input[^>]+type\s*=\s*["']?password\b`)
)

// abstractWalls are what a platform prints IN PLACE OF the body. Each was read on a ground-truth
// page labelled abstract and is absent from every page labelled full.
var abstractWalls = []struct {
	re     *regexp.Regexp
	reason string
}{
	// Springer Nature's platform (nature.com, link.springer.com) — on all three paywalled pages in
	// the corpus, the article's abstract followed by this notice in place of the body.
	{regexp.MustCompile(`(?i)preview of subscription content`),
		"Springer Nature's platform printed 'preview of subscription content' in place of the body"},
	// HighWire (BMJ and its hosted journals) names the abstract rendering by its view: the full
	// text is `fulltext-view`, the abstract `extract-view`.
	{regexp.MustCompile(`class\s*=\s*["'][^"']*\bextract-view\b`),
		"HighWire rendered its extract view (the abstract), not its full-text view"},
	// PubMed Central serves pre-digital articles as page IMAGES: the text is in the scan pdf, and
	// the page carries none of it.
	{regexp.MustCompile(`class\s*=\s*["']scanned-pages["']`),
		"PubMed Central renders this article as scanned page images; the page carries none of its text"},
}

// bodyMarkers count the body section headings a platform renders BEFORE its reference list.
// Counting before the references is what makes a stray section not count: nature.com numbers its
// supplementary-information section in the body series, and places it after the references.
var bodyMarkers = []struct {
	sections func([]byte) int
	reason   string
}{
	{func(b []byte) int {
		if !bytes.Contains(b, []byte("c-article-body")) {
			return 0
		}
		return len(springerNatureSection.FindAll(beforeRefs(b, springerNatureRefs), -1))
	}, "Springer Nature's platform rendered numbered body sections before the reference list"},
	{func(b []byte) int {
		i := bytes.Index(b, []byte("main-article-body"))
		if i < 0 {
			return 0
		}
		return len(pmcSection.FindAll(beforeRefs(b[i:], pmcRefs), -1))
	}, "PubMed Central rendered body section headings before the reference list"},
}

var (
	springerNatureSection = regexp.MustCompile(`<h2[^>]*\bid\s*=\s*["']Sec\d+["']`)
	springerNatureRefs    = regexp.MustCompile(`id\s*=\s*["']Bib1["']`)
	pmcSection            = regexp.MustCompile(`<h[2-4][^>]*class\s*=\s*["'][^"']*\bpmc_sec_title\b`)
	pmcRefs               = regexp.MustCompile(`id\s*=\s*["']ref-list`)
)

// beforeRefs is the page up to its reference list, or all of it where there is none.
func beforeRefs(b []byte, refs *regexp.Regexp) []byte {
	if loc := refs.FindIndex(b); loc != nil {
		return b[:loc[0]]
	}
	return b
}
