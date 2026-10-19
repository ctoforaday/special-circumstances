package fetchcache

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
	return completeness(htmlOf(contentType, body), reachedForAWork)
}

// completeness is Completeness over a parsed page; nil is not HTML.
func completeness(p *htmlPage, reachedForAWork bool) (verdict, reason string) {
	if p == nil {
		return "", ""
	}
	_, declared := p.meta("name", "citation_title")
	if !reachedForAWork && !declared {
		return "", ""
	}
	// NOT THE WORK first: a book's sales page carries the platform's paywall text too, and calling
	// it the book's abstract would put the publisher's blurb in the author's mouth.
	if t, ok := p.meta("property", "og:type"); ok && strings.EqualFold(strings.TrimSpace(t), "book") {
		return CompletenessNotTheWork, "the page declares itself a book's landing page (og:type book): a description and a table of contents, not the work's text"
	}
	if !declared && p.password {
		return CompletenessNotTheWork, "a sign-in form carrying no scholarly metadata: the source answered with its login page, not the work"
	}
	for _, w := range abstractWalls {
		if w.match(p) {
			return CompletenessAbstract, w.reason
		}
	}
	for _, m := range bodyMarkers {
		n, unknown := m.sections(p)
		if unknown != "" {
			return CompletenessUnverified, unknown
		}
		if n > 0 {
			return CompletenessFull, m.reason
		}
	}
	return CompletenessUnverified, "no platform marker this tool knows says whether the body is on this page; it is kept, and a fuller copy is looked for"
}

// abstractWalls are what a platform prints IN PLACE OF the body. Each was read on a ground-truth
// page labelled abstract and is absent from every page labelled full.
var abstractWalls = []struct {
	match  func(*htmlPage) bool
	reason string
}{
	// Springer Nature's platform names its paywall in its markup on nature.com: the notice that
	// stands in place of the body is the element carrying data-test="institution-access-preview".
	// The element, not its English: a markup attribute does not move when the wording does.
	{func(p *htmlPage) bool { return p.dataTest["institution-access-preview"] },
		"Springer Nature's platform rendered its institution-access preview (data-test=\"institution-access-preview\") in place of the body"},
	// link.springer.com prints the same notice in a generic notes element that names nothing, so
	// there the notice's own words are the signal — read in the page's VISIBLE text, where a copy
	// of the phrase in a script or a translation table cannot fire it.
	{func(p *htmlPage) bool {
		return strings.Contains(strings.ToLower(p.flat), "preview of subscription content")
	},
		"Springer Nature's platform printed 'preview of subscription content' in place of the body"},
	// HighWire (BMJ and its hosted journals) names the abstract rendering by its view: the full
	// text is `fulltext-view`, the abstract `extract-view`.
	{func(p *htmlPage) bool { return p.classes["extract-view"] },
		"HighWire rendered its extract view (the abstract), not its full-text view"},
	// PubMed Central serves pre-digital articles as page IMAGES: the text is in the scan pdf, and
	// the page carries none of it.
	{func(p *htmlPage) bool { return p.classes["scanned-pages"] },
		"PubMed Central renders this article as scanned page images; the page carries none of its text"},
}

// bodyMarkers count the body section headings a platform renders inside its article body BEFORE
// its reference list. Counting before the references is what makes a stray section not count:
// nature.com numbers its supplementary-information section in the body series, and places it
// after the references.
//
// WHERE THE REFERENCE LIST CANNOT BE FOUND, NOTHING IS COUNTED. A platform's page with no marker
// for its references is a page whose body cannot be told from what follows it, and counting the
// whole page there is the unsafe direction: nature.com's paywalled abstract carries a numbered
// supplementary section after its references, so one renamed attribute would turn the abstract
// into a full paper. The answer there is `unverified`, with that reason.
var bodyMarkers = []struct {
	sections func(*htmlPage) (n int, unknown string)
	reason   string
}{
	{func(p *htmlPage) (int, string) {
		if !p.snBody {
			return 0, ""
		}
		if !p.snRefs {
			return 0, "Springer Nature's platform rendered this page, and its reference-list marker (id=\"Bib1\") is absent, so where the body ends cannot be told; it is kept, and a fuller copy is looked for"
		}
		return p.snSections, ""
	}, "Springer Nature's platform rendered numbered body sections before the reference list"},
	{func(p *htmlPage) (int, string) {
		if !p.pmcBody {
			return 0, ""
		}
		if !p.pmcRefs {
			return 0, "PubMed Central rendered this page, and its reference-list marker (id=\"ref-list\") is absent, so where the body ends cannot be told; it is kept, and a fuller copy is looked for"
		}
		return p.pmcSections, ""
	}, "PubMed Central rendered body section headings before the reference list"},
}

// pdfCompleteness judges a pdf by its page count against the span the index declares for the
// work. A pdf is taken as the document — there is no evidence of a pdf of only an abstract — and
// this asks the one question the evidence supports: is it much SHORTER than the work?
//
// MEASURED over 20 retrieved pdfs against their declared spans: longer is normal (an accepted
// manuscript ran 44 pages against 19, another 15 against 7 — double spacing, cover sheets,
// supplements), and one was short: 10 pages of a 165-page report. So `full` is anything that is at
// least half the span, `unverified` is less, and a pdf with no numeric span is not asked.
//
// A BOOK DECLARES NO SPAN, and its open "copies" are often not the book. MEASURED: an index listed
// Springer's front matter (`/content/pdf/bfm:…`, 23 pages: preface and contents) as the open copy
// of a 22-chapter book, and a monograph's listed pdf was one page. Nothing can check a book's pdf,
// so it is `unverified` — kept, and never claimed to be the book.
func pdfCompleteness(pages, declared int, workType string) (verdict, reason string) {
	if pages > 0 && bookTypes[workType] {
		return CompletenessUnverified, fmt.Sprintf("the work is a %s, which declares no page span, so nothing says whether "+
			"these %d pages are the book or part of it (front matter, a chapter)", workType, pages)
	}
	if pages <= 0 || declared <= 0 {
		return "", ""
	}
	if pages*2 >= declared {
		return CompletenessFull, fmt.Sprintf("the pdf's %d pages cover the %d the index declares for the work", pages, declared)
	}
	return CompletenessUnverified, fmt.Sprintf("the pdf has %d pages where the index declares %d for the work: part of it, not all of it", pages, declared)
}

// bookTypes are the index's words for a whole book: OpenAlex's `book`, Crossref's `monograph` and
// `edited-book`. A chapter is not here — a chapter's pdf is that work.
var bookTypes = map[string]bool{"book": true, "monograph": true, "edited-book": true}

// SourceCompleteness is the record's word for this entry's completeness — what a citation of it
// carries. A word this switch does not know maps to UNSPECIFIED, which the record's write guard
// refuses: a new verdict added here without a record word must fail at the write, not reach a
// reader as some other verdict.
func (e Entry) SourceCompleteness() recordpb.SourceCompleteness {
	switch e.Completeness {
	case CompletenessFull:
		return recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL
	case CompletenessAbstract:
		return recordpb.SourceCompleteness_SOURCE_COMPLETENESS_ABSTRACT
	case CompletenessUnverified:
		return recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNVERIFIED
	case CompletenessNotTheWork:
		return recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
	case "":
		return recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED
	}
	return recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNSPECIFIED
}
