package fetchcache

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

func TestTheDeclaredSpanIsReadOnlyFromPageNumbers(t *testing.T) {
	for _, tc := range []struct {
		first, last string
		want        int
	}{
		{"1149", "1160", 12},
		{"291", "455", 165},
		{"7", "7", 1},
		{"e3000410", "", 0}, // an e-locator names an article, not a span
		{"e47", "e47", 0},
		{"10", "9", 0},
		{"10", "5", 0}, // backwards: a span cannot run from a later page to an earlier one
		{"", "", 0},
	} {
		if got := declaredPages(tc.first, tc.last); got != tc.want {
			t.Errorf("declaredPages(%q, %q) = %d, want %d", tc.first, tc.last, got, tc.want)
		}
	}
}

// LONGER IS NORMAL, SHORTER IS PART. The cases are the measured ones: an accepted manuscript at
// 44 pages against a 19-page span, and 10 pages of a 165-page report.
func TestAPDFIsJudgedByWhetherItIsMuchShorterThanTheWork(t *testing.T) {
	for _, tc := range []struct {
		pages, declared int
		want            string
	}{
		{44, 19, CompletenessFull},
		{9, 10, CompletenessFull},
		{10, 165, CompletenessUnverified},
		{12, 0, ""},
		{0, 12, ""},
	} {
		got, reason := pdfCompleteness(tc.pages, tc.declared, "article")
		if got != tc.want || (got != "" && reason == "") {
			t.Errorf("pdfCompleteness(%d, %d) = %q (%q), want %q", tc.pages, tc.declared, got, reason, tc.want)
		}
	}
}

// pagesByBody reports a page count per body, so one test can hold a short pdf and a whole one.
type pagesByBody map[string]int

func (p pagesByBody) Extract(_, mediaType string, body []byte) Extraction {
	if mediaType != "application/pdf" {
		return Extraction{}
	}
	return Extraction{Attempted: true, Text: "text", Pages: p[string(body)], ExtractorID: "stub"}
}

// THE CALL SITE: a live pdf that is a tenth of its work is recorded as part of it, and the rungs
// are asked for the rest.
func TestAShortLivePDFIsRecordedAsPartAndTheWholeIsSought(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	prev := DefaultExtractor
	DefaultExtractor = pagesByBody{"%PDF-1.7 short": 10, "%PDF-1.7 whole": 165}
	t.Cleanup(func() { DefaultExtractor = prev })

	f := fake(func(u string) (*Response, error) {
		switch {
		case u == keepGoingDOI:
			return &Response{Body: []byte("%PDF-1.7 short"), ContentType: "application/pdf"}, nil
		case isIndexLookup(u):
			r, _ := emptyIndexAnswer(u)
			return r, nil
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","biblio":{"first_page":"291","last_page":"455"},` +
				`"locations":[{"pdf_url":"https://repo.example/whole.pdf","is_oa":true}]}`)}, nil
		case u == "https://repo.example/whole.pdf":
			return &Response{Body: []byte("%PDF-1.7 whole"), ContentType: "application/pdf"}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	entry, body, _, err := Resolve(run, keepGoingDOI, f)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if string(body) != "%PDF-1.7 whole" || entry.Completeness != CompletenessFull {
		t.Errorf("the whole work was not taken over its tenth: body %q completeness %q (%s)", body, entry.Completeness, entry.CompletenessReason)
	}
	if entry.Work == nil || entry.Work.DeclaredPages != 165 {
		t.Errorf("the declared span did not reach the record: %+v", entry.Work)
	}
}

// PART OF THE WORK BEATS NONE OF IT, and only that: a short pdf replaces a known abstract, and is
// not traded for a page that may already be the whole text.
func TestAShortPDFIsOnlyAFallback(t *testing.T) {
	short := map[string]*Response{"https://repo.example/short.pdf": {Body: []byte("%PDF-1.7 short"), ContentType: "application/pdf"}}
	for _, tc := range []struct {
		page     string
		replaced bool
	}{
		{"nature-paywalled.html.gz", true},                       // abstract
		{"sciencedirect-oa-abstract-via-wayback.html.gz", false}, // unverified
	} {
		t.Run(tc.page, func(t *testing.T) {
			run := runtest.New(t, t.TempDir())
			prev := DefaultExtractor
			DefaultExtractor = pagesByBody{"%PDF-1.7 short": 10}
			t.Cleanup(func() { DefaultExtractor = prev })
			site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, tc.page), bodies: short}
			site.locs = `{"pdf_url":"https://repo.example/short.pdf","is_oa":true}`
			site.biblio = `{"first_page":"291","last_page":"455"}`
			entry, _, _, err := Resolve(run, keepGoingDOI, site)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got := entry.Backend == ViaOA; got != tc.replaced {
				t.Errorf("replaced by a short pdf = %v, want %v (completeness %q)", got, tc.replaced, entry.Completeness)
			}
			if tc.replaced && entry.Completeness != CompletenessUnverified {
				t.Errorf("the short pdf is not recorded as part of the work: %q", entry.Completeness)
			}
		})
	}
}

// A BOOK'S PDF IS NEVER CLAIMED TO BE THE BOOK: the measured cases are Springer's 23-page front
// matter for a 22-chapter book and a monograph's one-page pdf. A chapter's pdf is its own work.
func TestABooksPDFIsUnverified(t *testing.T) {
	for _, tc := range []struct {
		workType string
		want     string
	}{
		{"book", CompletenessUnverified},
		{"monograph", CompletenessUnverified},
		{"edited-book", CompletenessUnverified},
		{"book-chapter", ""},
	} {
		if got, reason := pdfCompleteness(23, 0, tc.workType); got != tc.want || (got != "" && reason == "") {
			t.Errorf("a %s's 23-page pdf got %q (%q), want %q", tc.workType, got, reason, tc.want)
		}
	}
}

// THE CALL SITE: the book's landing page is replaced by the pdf the index lists — which is the
// front matter — and the record says nobody can tell whether that is the book.
func TestABooksListedPDFIsRecordedAsUnverified(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	prev := DefaultExtractor
	DefaultExtractor = pagesByBody{"%PDF-1.7 front matter": 23}
	t.Cleanup(func() { DefaultExtractor = prev })
	const bfm = "https://link.springer.com/content/pdf/bfm:978-0-387-46312-4/1"
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "springer-book-landing.html.gz"), kind: "book",
		locs:   `{"pdf_url":"` + bfm + `","is_oa":true}`,
		bodies: map[string]*Response{bfm: {Body: []byte("%PDF-1.7 front matter"), ContentType: "application/pdf"}}}
	entry, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Completeness != CompletenessUnverified || !strings.Contains(entry.CompletenessReason, "book") {
		t.Errorf("a book's listed pdf got %q (%s); nothing can say it is the book", entry.Completeness, entry.CompletenessReason)
	}
}
