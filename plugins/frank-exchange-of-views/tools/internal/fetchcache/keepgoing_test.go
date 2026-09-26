package fetchcache

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// keepGoingSite serves one page at the asked doi url and an OpenAlex record listing the given
// locations; every other url is answered by bodies, or refused. It records what was asked.
type keepGoingSite struct {
	mu     sync.Mutex
	asked  string
	page   []byte
	locs   string
	biblio string // OpenAlex's biblio object, where a test declares the work's page span
	kind   string // OpenAlex's type for the work, where a test declares one
	bodies map[string]*Response
	seen   []string
}

func (k *keepGoingSite) Fetch(u string) (*Response, error) {
	k.mu.Lock()
	k.seen = append(k.seen, u)
	k.mu.Unlock()
	switch {
	case u == k.asked:
		return &Response{Body: k.page, ContentType: "text/html; charset=utf-8"}, nil
	case isIndexLookup(u):
		r, _ := emptyIndexAnswer(u)
		return r, nil
	case strings.Contains(u, "openalex"):
		biblio := ""
		if k.biblio != "" {
			biblio = `"biblio":` + k.biblio + `,`
		}
		if k.kind != "" {
			biblio += `"type":"` + k.kind + `",`
		}
		return &Response{Body: []byte(`{"id":"https://openalex.org/W1","open_access":{"oa_url":null},` + biblio + `"locations":[` + k.locs + `]}`)}, nil
	}
	if r, ok := k.bodies[u]; ok {
		return r, nil
	}
	return nil, &Refusal{URL: u, Status: 403}
}

func (k *keepGoingSite) candidatesFetched() []string {
	var out []string
	for _, u := range k.seen {
		if u != k.asked && !isIndexLookup(u) && !strings.Contains(u, "openalex") && !strings.Contains(u, "archive.org") {
			out = append(out, u)
		}
	}
	return out
}

func withText(t *testing.T, text string) {
	t.Helper()
	prev := DefaultExtractor
	DefaultExtractor = fixedExtractor{Extraction{Attempted: true, Text: text, Pages: 12, ExtractorID: "stub"}}
	t.Cleanup(func() { DefaultExtractor = prev })
}

const keepGoingDOI = "https://doi.org/10.1234/x"

// THE ABSTRACT IS A 200, AND A 200 USED TO END THE SEARCH. A publisher's paywalled page is
// replaced by the open copy the indexes list, and the record says why it went looking.
func TestAnAbstractPageIsReplacedByTheOpenCopy(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "the paper's own text")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-paywalled.html.gz"),
		locs:   `{"pdf_url":"https://repo.example/open.pdf","is_oa":true,"version":"acceptedVersion"}`,
		bodies: map[string]*Response{"https://repo.example/open.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}}}
	entry, body, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if string(body) != "%PDF-1.7 the paper" || entry.Backend != ViaOA {
		t.Fatalf("the abstract was kept over an open copy of the paper: backend=%q completeness=%q", entry.Backend, entry.Completeness)
	}
	if !strings.Contains(entry.RetrievedVia, "abstract") || !strings.Contains(entry.RetrievedVia, "repo.example/open.pdf") {
		t.Errorf("the record does not say the asked page was the abstract, or where the copy came from: %s", entry.RetrievedVia)
	}
	if !entry.TextRetrieved {
		t.Error("the open copy's text is not recorded as retrieved")
	}
}

// ONLY WHEN THE PAGE IS NOT KNOWN TO BE THE BODY. A full page asks no index for another copy:
// the search costs requests, and a paper already in hand does not pay for it.
func TestAFullPageLooksNoFurther(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-full.html.gz"),
		locs:   `{"pdf_url":"https://repo.example/open.pdf","is_oa":true}`,
		bodies: map[string]*Response{"https://repo.example/open.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}}}
	entry, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Completeness != CompletenessFull || entry.Backend != "" {
		t.Errorf("a full page was not kept: completeness=%q backend=%q", entry.Completeness, entry.Backend)
	}
	for _, u := range site.seen {
		if isIndexLookup(u) || strings.Contains(u, "repo.example") {
			t.Errorf("a full page went looking for another copy: asked %s", u)
		}
	}
}

// THE PAGE IS KEPT WHEN NOTHING BETTER EXISTS: an abstract is still what the authors wrote about
// the work, and it is recorded as the abstract.
func TestAnAbstractIsKeptWhenNoCopyCarriesMore(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-paywalled.html.gz")}
	entry, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Completeness != CompletenessAbstract || entry.Backend != "" {
		t.Errorf("completeness=%q backend=%q, want the abstract kept as the live page", entry.Completeness, entry.Backend)
	}
}

// A SCAN DOES NOT REPLACE A PAGE THAT MAY ALREADY BE THE TEXT. An unverified page is kept over a
// pdf with no text layer; a known abstract is not.
func TestAScanReplacesOnlyAPageKnownNotToBeTheBody(t *testing.T) {
	scan := map[string]*Response{"https://repo.example/scan.pdf": {Body: []byte("%PDF-1.4 images"), ContentType: "application/pdf"}}
	loc := `{"pdf_url":"https://repo.example/scan.pdf","is_oa":true}`
	for _, tc := range []struct {
		page     string
		replaced bool
	}{
		{"sciencedirect-oa-abstract-via-wayback.html.gz", false}, // unverified
		{"springer-link-paywalled.html.gz", true},                // abstract
	} {
		t.Run(tc.page, func(t *testing.T) {
			run := runtest.New(t, t.TempDir())
			withText(t, "")
			site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, tc.page), locs: loc, bodies: scan}
			entry, _, _, err := Resolve(run, keepGoingDOI, site)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got := entry.Backend == ViaOA; got != tc.replaced {
				t.Errorf("replaced by a scan = %v, want %v (completeness %q)", got, tc.replaced, entry.Completeness)
			}
		})
	}
}

// THE OPEN-ACCESS WALK GOES PAST AN ABSTRACT. An index's "open" location is often the publisher's
// landing page; the pdf further down the list is the work.
func TestTheOpenAccessWalkGoesPastAnAbstractToTheBody(t *testing.T) {
	withText(t, "the paper")
	abstract := loadBody(t, "springer-link-paywalled.html.gz")
	site := &keepGoingSite{asked: "unused",
		locs: `{"landing_page_url":"https://publisher.example/abs","is_oa":true},{"pdf_url":"https://repo.example/open.pdf","is_oa":true},{"pdf_url":"https://later.example/c.pdf","is_oa":true}`,
		bodies: map[string]*Response{
			"https://publisher.example/abs": {Body: abstract, ContentType: "text/html"},
			"https://repo.example/open.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"},
		}}
	att := Recover(site, keepGoingDOI, ViaOA, "")
	if att == nil || MediaType(att.ContentType) != "application/pdf" {
		t.Fatalf("the walk stopped at the abstract: %+v", att)
	}
	// It stops at the body: a candidate after the pdf is a request nothing needed.
	for _, u := range site.candidatesFetched() {
		if strings.Contains(u, "later.example") {
			t.Errorf("the walk went on past the body to %s", u)
		}
	}
	// And with nothing after it, the abstract is still the answer — ranked last, not refused.
	site.locs = `{"landing_page_url":"https://publisher.example/abs","is_oa":true}`
	att = Recover(site, keepGoingDOI, ViaOA, "")
	if att == nil || !strings.Contains(MediaType(att.ContentType), "html") {
		t.Fatalf("an abstract with nothing better was dropped: %+v", att)
	}
}

// ONLY A BODY REPLACES THE PAGE. An open repository's record page is unverified — it may be the
// text, and usually is a record like the one in hand — so a known abstract is not traded for it.
func TestAnUnverifiedCopyDoesNotReplaceTheAbstract(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-paywalled.html.gz"),
		locs:   `{"landing_page_url":"https://repo.example/record","is_oa":true}`,
		bodies: map[string]*Response{"https://repo.example/record": {Body: loadBody(t, "bepress-repository-record.html.gz"), ContentType: "text/html"}}}
	entry, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.Backend != "" || entry.Completeness != CompletenessAbstract {
		t.Errorf("the abstract was traded for an unverified page: backend=%q completeness=%q", entry.Backend, entry.Completeness)
	}
}

// THE RANKING, candidate by candidate: an unverified page beats an abstract, and a page its
// platform marks full is the body — taken at once, with nothing further fetched.
func TestTheOpenAccessWalkRanksWhatItHolds(t *testing.T) {
	withText(t, "the paper")
	abstract := loadBody(t, "springer-link-paywalled.html.gz")
	record := loadBody(t, "bepress-repository-record.html.gz")
	full := loadBody(t, "pmc-full.html.gz")

	site := &keepGoingSite{asked: "unused",
		locs: `{"landing_page_url":"https://publisher.example/abs","is_oa":true},{"landing_page_url":"https://repo.example/record","is_oa":true}`,
		bodies: map[string]*Response{
			"https://publisher.example/abs": {Body: abstract, ContentType: "text/html"},
			"https://repo.example/record":   {Body: record, ContentType: "text/html"},
		}}
	if att := Recover(site, keepGoingDOI, ViaOA, ""); att == nil || !strings.Contains(att.Via, "repo.example/record") {
		t.Errorf("an abstract was held over an unverified page: %+v", att)
	}

	site = &keepGoingSite{asked: "unused",
		// Both landing pages: the list puts pdf locations first, so a later candidate is another page.
		locs: `{"landing_page_url":"https://pmc.example/full","is_oa":true},{"landing_page_url":"https://later.example/c","is_oa":true}`,
		bodies: map[string]*Response{
			"https://pmc.example/full": {Body: full, ContentType: "text/html"},
			"https://later.example/c":  {Body: record, ContentType: "text/html"},
		}}
	att := Recover(site, keepGoingDOI, ViaOA, "")
	if att == nil || !strings.Contains(att.Via, "pmc.example/full") {
		t.Errorf("a page marked full was not taken as the body: %+v", att)
	}
	for _, u := range site.candidatesFetched() {
		if strings.Contains(u, "later.example") {
			t.Errorf("the walk went on past a full page to %s", u)
		}
	}
}

// A LANDING PAGE'S OWN PDF IS THE BODY, and the walk stops there rather than fetching the next
// candidate for nothing.
func TestALandingPagesPDFEndsTheWalk(t *testing.T) {
	withText(t, "the paper")
	landing := []byte(`<html><head><meta name="citation_title" content="x">` +
		`<meta name="citation_pdf_url" content="https://publisher.example/x.pdf"></head><body>` +
		strings.Repeat("abstract words ", 200) + `</body></html>`)
	site := &keepGoingSite{asked: "unused",
		locs: `{"landing_page_url":"https://publisher.example/abs","is_oa":true},{"landing_page_url":"https://later.example/c","is_oa":true}`,
		bodies: map[string]*Response{
			"https://publisher.example/abs":   {Body: landing, ContentType: "text/html"},
			"https://publisher.example/x.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"},
			"https://later.example/c":         {Body: landing, ContentType: "text/html"},
		}}
	att := Recover(site, keepGoingDOI, ViaOA, "")
	if att == nil || string(att.Body) != "%PDF-1.7 the paper" {
		t.Fatalf("the landing page's pdf was not taken: %+v", att)
	}
	for _, u := range site.candidatesFetched() {
		if strings.Contains(u, "later.example") {
			t.Errorf("the walk went on past the landing page's pdf to %s", u)
		}
	}
}

// AN ARCHIVED ABSTRACT DOES NOT END THE WALK. A refused candidate's snapshot is often the landing
// page; the next candidate may be the paper.
func TestAnArchivedAbstractDoesNotEndTheWalk(t *testing.T) {
	withText(t, "the paper")
	const walled = "https://publisher.example/abs"
	const snapshot = "https://web.archive.org/web/20250812050548/" + walled
	site := &keepGoingSite{asked: "unused",
		// Both landing pages: pdf locations sort first, and the walled page must be asked before the body.
		locs: `{"landing_page_url":"` + walled + `","is_oa":true},{"landing_page_url":"https://pmc.example/full","is_oa":true}`,
		bodies: map[string]*Response{
			"https://archive.org/wayback/available?url=https%3A%2F%2Fpublisher.example%2Fabs": {Body: []byte(
				`{"archived_snapshots":{"closest":{"available":true,"status":"200","timestamp":"20250812050548","url":"` + snapshot + `"}}}`)},
			"https://web.archive.org/web/20250812050548if_/" + walled: {Body: loadBody(t, "springer-link-paywalled.html.gz"), ContentType: "text/html"},
			"https://pmc.example/full":                                {Body: loadBody(t, "pmc-full.html.gz"), ContentType: "text/html"},
		}}
	att := Recover(site, keepGoingDOI, ViaOA, "")
	if att == nil || !strings.Contains(att.Via, "pmc.example/full") {
		t.Fatalf("an archived abstract ended the walk before the paper: %+v", att)
	}
	if !func() bool {
		for _, u := range site.seen {
			if strings.Contains(u, "if_/"+walled) {
				return true
			}
		}
		return false
	}() {
		t.Fatal("the walled candidate's snapshot was never asked for: the test did not reach the archive")
	}
}

// A DOCUMENT URL ANSWERED WITH THE ABSTRACT HAS BEEN REFUSED, and the archive is asked for it as
// for any refusal. Measured on nature.com: the .pdf an index names 303s back to the paywalled
// page. A landing page answering with its abstract is only being what it is, and costs no lookup.
func TestADocumentURLAnsweredWithItsAbstractIsSoughtInTheArchive(t *testing.T) {
	withText(t, "the paper")
	const pdfURL = "https://publisher.example/x.pdf"
	abstract := loadBody(t, "springer-link-paywalled.html.gz")
	site := &keepGoingSite{asked: "unused",
		locs: `{"pdf_url":"` + pdfURL + `","is_oa":true}`,
		bodies: map[string]*Response{
			pdfURL: {Body: abstract, ContentType: "text/html"},
			"https://archive.org/wayback/available?url=https%3A%2F%2Fpublisher.example%2Fx.pdf": {Body: []byte(
				`{"archived_snapshots":{"closest":{"available":true,"status":"200","timestamp":"20240101000000","url":"https://web.archive.org/web/20240101000000/` + pdfURL + `"}}}`)},
			"https://web.archive.org/web/20240101000000if_/" + pdfURL: {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"},
		}}
	att := Recover(site, keepGoingDOI, ViaOA, "")
	if att == nil || string(att.Body) != "%PDF-1.7 the paper" || !strings.Contains(att.Via, "not its body") {
		t.Fatalf("the archived pdf was not sought after its url answered with the abstract: %+v", att)
	}

	landing := &keepGoingSite{asked: "unused",
		locs:   `{"landing_page_url":"https://publisher.example/abs","is_oa":true}`,
		bodies: map[string]*Response{"https://publisher.example/abs": {Body: abstract, ContentType: "text/html"}}}
	_ = Recover(landing, keepGoingDOI, ViaOA, "")
	for _, u := range landing.seen {
		if strings.Contains(u, "archive.org") {
			t.Errorf("a landing page answering with its abstract sent the walk to the archive: %s", u)
		}
	}
}

// THE TEXT THE RECORD NAMES IS THE TEXT ON DISK. A recovered document's extraction was written
// under an empty sha — `<run>/cache.txt`, shared by every recovery — while the record pointed at
// `<sha>.txt`, which did not exist. A seat following text_path found nothing; one reading the
// shared file read whichever paper was recovered last.
func TestARecoveredDocumentsTextIsWhereTheRecordSays(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	withText(t, "the paper's own text")
	site := &keepGoingSite{asked: keepGoingDOI, page: loadBody(t, "nature-paywalled.html.gz"),
		locs:   `{"pdf_url":"https://repo.example/open.pdf","is_oa":true}`,
		bodies: map[string]*Response{"https://repo.example/open.pdf": {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"}}}
	entry, _, _, err := Resolve(run, keepGoingDOI, site)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.TextSha == "" {
		t.Fatal("no text was recorded for the recovered document")
	}
	got, err := os.ReadFile(TextPath(run, entry.Sha))
	if err != nil || string(got) != "the paper's own text" {
		t.Errorf("the text the record names is not on disk: %q, %v", got, err)
	}
}
