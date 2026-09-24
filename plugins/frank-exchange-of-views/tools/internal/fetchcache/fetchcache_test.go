package fetchcache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubFetcher is the offline Fetcher every cache test drives — it counts calls so a test
// can prove a second read was served from cache (Fetch NOT re-entered).
type stubFetcher struct {
	calls       int
	resp        map[string][]byte
	contentType string
	disposition string
	err         error
}

func (s *stubFetcher) Fetch(url string) (*Response, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if b, ok := s.resp[url]; ok {
		return &Response{Body: b, ContentType: s.contentType, Disposition: s.disposition}, nil
	}
	return nil, fmt.Errorf("no stub response for %s", url)
}

func TestResolveMissThenHit(t *testing.T) {
	run := t.TempDir()
	body := []byte("the source bytes both sides read")
	f := &stubFetcher{resp: map[string][]byte{"https://ex/1": body}}

	e1, got, hit, err := Resolve(runtest.Open(t, run), "https://ex/1", f)
	if err != nil {
		t.Fatalf("first Resolve: %v", err)
	}
	if hit {
		t.Errorf("first Resolve reported a cache hit — it was a miss")
	}
	if f.calls != 1 {
		t.Errorf("first Resolve fetched %d times, want 1", f.calls)
	}
	if string(got) != string(body) {
		t.Errorf("first Resolve bytes = %q, want %q", got, body)
	}
	if e1.Sha != Sha(body) {
		t.Errorf("sha = %q, want %q", e1.Sha, Sha(body))
	}
	if _, err := os.Stat(Path(runtest.Open(t, run), e1.Sha)); err != nil {
		t.Errorf("cache file not written: %v", err)
	}

	// Second read: same URL → served from cache, Fetch NOT re-entered.
	e2, got2, hit2, err := Resolve(runtest.Open(t, run), "https://ex/1", f)
	if err != nil {
		t.Fatalf("second Resolve: %v", err)
	}
	if !hit2 {
		t.Errorf("second Resolve reported a miss — the URL was already cached")
	}
	if f.calls != 1 {
		t.Errorf("second Resolve fetched again (calls=%d) — download-once violated", f.calls)
	}
	if e2.Sha != e1.Sha || string(got2) != string(body) {
		t.Errorf("second Resolve = (%s,%q), want (%s,%q)", e2.Sha, got2, e1.Sha, body)
	}
}

func TestResolveFailureCachesNothing(t *testing.T) {
	run := t.TempDir()
	f := &stubFetcher{err: errors.New("host unreachable")}

	_, _, _, err := Resolve(runtest.Open(t, run), "https://gone/x", f)
	if err == nil {
		t.Fatal("Resolve of an unreachable URL returned nil error")
	}
	// No cache dir/index written on a pure failure — nothing to serve, nothing to leak.
	if entries, _ := os.ReadDir(Dir(runtest.Open(t, run))); len(entries) != 0 {
		t.Errorf("a failed fetch left %d cache entries, want 0", len(entries))
	}
	// A retry after the source comes back succeeds and caches (failure was not sticky).
	f.err = nil
	f.resp = map[string][]byte{"https://gone/x": []byte("now up")}
	if _, _, hit, err := Resolve(runtest.Open(t, run), "https://gone/x", f); err != nil || hit {
		t.Errorf("retry after recovery: hit=%v err=%v, want (false,nil)", hit, err)
	}
}

func TestStoreDedupsOnContent(t *testing.T) {
	run := t.TempDir()
	body := []byte("identical bytes")
	se1, err := Store(runtest.Open(t, run), Entry{URL: "https://a"}, body)
	if err != nil {
		t.Fatal(err)
	}
	se2, err := Store(runtest.Open(t, run), Entry{URL: "https://b"}, body) // different URL, same bytes
	if err != nil {
		t.Fatal(err)
	}
	if se1.Sha != se2.Sha {
		t.Errorf("same bytes hashed differently: %s vs %s", se1.Sha, se2.Sha)
	}
	// One content file (content-addressed dedup), two index lines (two URLs).
	files, _ := filepath.Glob(filepath.Join(Dir(runtest.Open(t, run)), "*"))
	content := 0
	for _, f := range files {
		if filepath.Base(f) == se1.Sha {
			content++
		}
	}
	if content != 1 {
		t.Errorf("content files for one hash = %d, want 1", content)
	}
	if g1, _, ok1, _ := Lookup(runtest.Open(t, run), "https://a"); !ok1 || g1.Sha != se1.Sha {
		t.Errorf("Lookup a missed or resolved to the wrong hash (ok=%v), want %s", ok1, se1.Sha)
	}
	if g2, _, ok2, _ := Lookup(runtest.Open(t, run), "https://b"); !ok2 || g2.Sha != se2.Sha {
		t.Errorf("Lookup b missed or resolved to the wrong hash (ok=%v), want %s", ok2, se2.Sha)
	}
}

// A crash between the content write and the index append leaves an index line pointing at
// a missing file. Lookup must treat that as a MISS so the next fetch self-heals, not as a
// hit that reads a phantom.
func TestLookupTreatsMissingContentAsMiss(t *testing.T) {
	run := t.TempDir()
	if err := os.MkdirAll(Dir(runtest.Open(t, run)), 0o755); err != nil {
		t.Fatal(err)
	}
	// Hand-write an index entry for a sha whose content file does not exist.
	if err := os.WriteFile(indexPath(runtest.Open(t, run)), []byte(`{"sha":"deadbeef","url":"https://orphan"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := Lookup(runtest.Open(t, run), "https://orphan"); err != nil || ok {
		t.Errorf("Lookup of an orphaned index line = (ok=%v,err=%v), want (false,nil)", ok, err)
	}
}

func TestLookupUncachedURLIsMiss(t *testing.T) {
	run := t.TempDir()
	if _, _, ok, err := Lookup(runtest.Open(t, run), "https://never"); err != nil || ok {
		t.Errorf("Lookup before any fetch = (ok=%v,err=%v), want (false,nil)", ok, err)
	}
}

// AN INDEX WRITTEN BY AN OLDER BINARY MUST STILL RESOLVE. Every field #629 adds is optional on
// read, because a run started before this change carries two-field lines and a run resumed after
// it must not read them as corrupt — nor, worse, as measured answers of "" and false.
func TestAnIndexLineWithoutTheNewFieldsStillResolves(t *testing.T) {
	run := t.TempDir()
	body := []byte("bytes cached by a binary that knew nothing of content types")
	sha := Sha(body)
	if err := os.MkdirAll(Dir(runtest.Open(t, run)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(runtest.Open(t, run), sha), body, 0o644); err != nil {
		t.Fatal(err)
	}
	// The old two-field shape, verbatim.
	line := `{"sha":"` + sha + `","url":"https://old/entry"}` + "\n"
	if err := os.WriteFile(indexPath(runtest.Open(t, run)), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	got, gotBody, ok, err := Lookup(runtest.Open(t, run), "https://old/entry")
	if err != nil {
		t.Fatalf("Lookup on a pre-#629 index: %v", err)
	}
	if !ok {
		t.Fatal("a pre-#629 index line did not resolve — an in-flight run would lose its cache")
	}
	if got.Sha != sha || string(gotBody) != string(body) {
		t.Errorf("round-trip mismatch: sha %s/%s, %d/%d bytes", got.Sha, sha, len(gotBody), len(body))
	}
	// AND THE ABSENT FIELDS READ AS ABSENT, NOT AS ANSWERS. TextExtracted nil is "never asked";
	// a plain false would tell a reader this document was examined and found to have no text.
	if got.TextExtracted != nil {
		t.Errorf("TextExtracted = %v on a line that never carried it, want nil (not measured)", *got.TextExtracted)
	}
	if got.ContentType != "" || got.Filename != "" || got.Extractor != "" {
		t.Errorf("an old line produced content_type=%q filename=%q extractor=%q, want all empty",
			got.ContentType, got.Filename, got.Extractor)
	}
}

// THE FACTS MEASURED AT FETCH REACH THE INDEX, which is the whole of #629 defect 2: Content-Type
// arrived in the response header and was thrown away at the only moment it existed.
func TestResolveRecordsWhatTheResponseSaidAndWhatWasExtracted(t *testing.T) {
	run := t.TempDir()
	const body = "%PDF-1.7 pretend"
	f := &stubFetcher{
		resp:        map[string][]byte{"https://ex/paper": []byte(body)},
		contentType: "application/pdf; charset=binary",
		disposition: `attachment; filename="served-name.pdf"`,
	}
	prev := DefaultExtractor
	DefaultExtractor = fixedExtractor{Extraction{
		Attempted: true, Text: "extracted words", Title: "The Document Title",
		Pages: 12, ExtractorID: "stub@v1",
	}}
	t.Cleanup(func() { DefaultExtractor = prev })

	got, _, _, err := Resolve(runtest.Open(t, run), "https://ex/paper", f)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.ContentType != "application/pdf" {
		t.Errorf("ContentType = %q, want the bare media type from the header", got.ContentType)
	}
	if got.Filename != "The Document Title" {
		t.Errorf("Filename = %q — the document Title outranks the served name (D4)", got.Filename)
	}
	if got.Pages != 12 {
		t.Errorf("Pages = %d, want 12", got.Pages)
	}
	if got.TextExtracted == nil || !*got.TextExtracted {
		t.Fatalf("TextExtracted = %v, want true", got.TextExtracted)
	}
	if got.TextSha != Sha([]byte("extracted words")) {
		t.Errorf("TextSha = %q, want the hash of the EXTRACTION — the body's own hash proves the "+
			"document is unchanged and says nothing about the text a claim came from", got.TextSha)
	}

	// And a second Resolve returns the SAME entry from the index, not a re-derived one: the
	// facts must survive the round-trip through JSON, or a cache hit silently loses them.
	again, _, hit, err := Resolve(runtest.Open(t, run), "https://ex/paper", f)
	if err != nil || !hit {
		t.Fatalf("second Resolve: hit=%v err=%v", hit, err)
	}
	// Compared field-wise on purpose: TextExtracted is a *bool, so `==` on the struct would
	// compare ADDRESSES and pass or fail for reasons that have nothing to do with the record.
	if !sameEntry(again, got) {
		t.Errorf("cache hit returned a different entry:\n got %s\nwant %s", showEntry(again), showEntry(got))
	}
}

// fixedExtractor returns one prepared answer for anything, so a cache test can state what a
// document contains without shipping a fixture PDF to make the point.
type fixedExtractor struct{ out Extraction }

func (f fixedExtractor) Extract(_, _ string, _ []byte) Extraction { return f.out }

// sameEntry compares two entries by value, dereferencing the three-state TextExtracted.
// sameEntry compares two entries BY THEIR SERIALISED FORM, which is the round-trip the caller is
// actually testing.
//
// IT USED TO NIL OUT TextExtracted AND COMPARE THE STRUCTS, and that quietly compared every other
// pointer field BY ADDRESS. NotRenderable never caught it because both sides were nil in the one
// test that used this; adding a third pointer field made two entries with identical contents
// compare unequal, and the failure read as a cache defect. An enumerating helper that has to be
// widened for each new field is the same shape as the bug it was guarding against.
func sameEntry(a, b Entry) bool {
	ja, erra := json.Marshal(a)
	jb, errb := json.Marshal(b)
	return erra == nil && errb == nil && bytes.Equal(ja, jb)
}

func showEntry(e Entry) string {
	extracted := "nil"
	if e.TextExtracted != nil {
		extracted = fmt.Sprintf("%v", *e.TextExtracted)
	}
	e.TextExtracted = nil
	return fmt.Sprintf("%+v text_extracted=%s", e, extracted)
}

// A RECOVERED WALL IS NOT A RECOVERED DOCUMENT. The wall detector ran on the live path only, so
// every archive snapshot, open-access copy and bibliographic record was stored unexamined — and
// the archive is the rung MOST likely to hand back a landing page rather than a paper, which the
// tool's own summary says in as many words.
//
// Measured on a 50-url scan: an archived Journal of Chemical Physics page was recorded as a
// retrieved document at 1,525 bytes carrying ten visible characters. A seat would have read that
// as the paper.
func TestARecoveredWallIsClassifiedLikeALiveOne(t *testing.T) {
	starved := []byte(`<html><head><title>x</title></head><body><div id="root"></div></body></html>`)
	live := Entry{ContentType: "text/html"}
	Classify(&live, starved)
	if live.NotRenderable == nil || !*live.NotRenderable {
		t.Fatal("the live path stopped refusing a starved page")
	}

	// The same bytes arriving by recovery must reach the same verdict.
	rec := Entry{ContentType: "text/html", RetrievedVia: "archive.org capture of 2019-05-20",
		Backend: ViaArchive, TextRetrieved: true}
	Classify(&rec, starved)
	if rec.NotRenderable == nil || !*rec.NotRenderable {
		t.Error("a recovered page was stored without being classified — the route decided the " +
			"verdict instead of the bytes")
	}
	if rec.NotRenderableReason == "" {
		t.Error("a refusal with no reason is a verdict a reader cannot check")
	}

	// AND A REAL DOCUMENT IS STILL A DOCUMENT, whichever route it came by.
	good := Entry{ContentType: "text/html", Backend: ViaArchive}
	Classify(&good, []byte("<html><body><article><p>"+strings.Repeat("real prose here. ", 80)+"</p></article></body></html>"))
	if good.NotRenderable == nil || *good.NotRenderable {
		t.Error("an archived article was refused")
	}
}

// THE CALL SITE, NOT THE FUNCTION. A test of Classify passes whether or not the recovery path
// calls it — which is exactly how this shipped: the live path classified, the recovery path did
// not, and the archive is the rung most likely to return a landing page.
//
// This drives Resolve: the live url is refused, the archive answers with a starved page, and the
// question is what the stored entry says about it.
func TestResolveClassifiesWhatTheRecoveryPathStored(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	wall := []byte(`<html><head><title>x</title></head><body><div id="root"></div></body></html>`)
	cdx := `[["timestamp","original","digest"],["20190520000000","https://ex/a","D1"]]`
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "cdx/search"):
			return &Response{Body: []byte(cdx), ContentType: "application/json"}, nil
		case strings.Contains(u, "web.archive.org/web/"):
			return &Response{Body: wall, ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	prev := DefaultExtractor
	DefaultExtractor = fixedExtractor{Extraction{}}
	t.Cleanup(func() { DefaultExtractor = prev })

	entry, _, _, err := Resolve(run, "https://ex/a", f)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if entry.NotRenderable == nil || !*entry.NotRenderable {
		t.Errorf("the archive handed back a starved page and it was stored as a document: "+
			"not_renderable=%v reason=%q", entry.NotRenderable, entry.NotRenderableReason)
	}
	// AND THE CLAIM THAT TEXT WAS RETRIEVED IS WITHDRAWN WITH IT. A seat reading `--json` must
	// not be told a wall is the source's text.
	if entry.TextRetrieved {
		t.Error("a recovered wall still claims text_retrieved — that is what a seat branches on")
	}
}

// THE PAPER WE COULD READ IS THE ONE THAT MOST NEEDS THE RETRACTION CHECK, and it was the one
// not getting it.
//
// The index lookup lived in Recover's stamp, which runs only when the live fetch FAILED. Measured
// over 47 works on 2026-09-23: all 28 rows carrying work facts were `metadata` or `oa` answers,
// and all 8 rows that returned a readable document carried none. The fact that can void a
// citation was present on exactly the documents nobody could quote.
func TestASuccessfulLiveFetchStillAsksTheIndex(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	yes := true
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","is_retracted":true,` +
				`"type":"article","open_access":{"oa_status":"bronze"}}`)}, nil
		case strings.Contains(u, "publisher.example"):
			return &Response{Body: []byte("<html><body>" + strings.Repeat("the paper. ", 300) + "</body></html>"),
				ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	e, _, _, err := Resolve(run, "https://publisher.example/doi/10.1234/withdrawn", f)
	if err != nil {
		t.Fatalf("the live fetch failed: %v", err)
	}
	if !e.TextRetrieved && e.Sha == "" {
		t.Fatal("no document was stored, so this tests nothing")
	}
	if e.Work == nil || e.Work.Retracted == nil || *e.Work.Retracted != yes {
		t.Fatalf("a successfully READ paper carries no retraction check: %+v", e.Work)
	}
	if e.Work.OAStatus != "bronze" {
		t.Errorf("the rest of the work record did not travel: %+v", e.Work)
	}

	// AND A URL WITH NO DOI ASKS NOBODY. There is nothing to ask about, and a request per fetch
	// on every non-scholarly url would be a cost paid for no possible answer.
	var asked int
	g := fake(func(u string) (*Response, error) {
		if strings.Contains(u, "openalex") {
			asked++
		}
		return &Response{Body: []byte("<html><body>" + strings.Repeat("plain. ", 300) + "</body></html>"),
			ContentType: "text/html"}, nil
	})
	if _, _, _, err := Resolve(run, "https://example.org/a-page", g); err != nil {
		t.Fatalf("the second fetch failed: %v", err)
	}
	if asked != 0 {
		t.Errorf("an index was asked about a url carrying no doi (%d times)", asked)
	}
}

// AN ABSTRACT PAGE IS NOT THE PAPER, AND THE PUBLISHER SAYS WHERE THE PAPER IS.
//
// MEASURED over 120 works: of 25 html bodies this tool recorded as documents, eleven were
// abstract or landing pages — 401 to 1,885 words of navigation, abstract and references. That is
// the same false claim as calling a zip the source's text and worse, because an abstract reads
// like a paper: a seat quoting one would be quoting the summary while saying it read the study.
//
// Of the sixteen short pages, TWELVE carried the publisher's own pointer to the readable copy.
// The fix is that stated fact, not a word-count threshold.
func TestALandingPageIsFollowedToTheFullTextItNames(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	const landing = "https://publisher.example/article/1"
	const full = "https://publisher.example/article/1/fulltext"
	abstract := `<html><head>` +
		`<meta name="citation_pdf_url" content="https://publisher.example/article/1.pdf">` +
		`<meta name="citation_fulltext_html_url" content="` + full + `">` +
		`</head><body>` + strings.Repeat("the abstract. ", 40) + `</body></html>`
	body := "<html><body>" + strings.Repeat("methods results discussion. ", 900) + "</body></html>"
	f := fake(func(u string) (*Response, error) {
		switch u {
		case landing:
			return &Response{Body: []byte(abstract), ContentType: "text/html"}, nil
		case full:
			return &Response{Body: []byte(body), ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	e, got, _, err := Resolve(run, landing, f)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(got) != len(body) {
		t.Fatalf("the abstract page was stored as the document (%d bytes, full text is %d)", len(got), len(body))
	}
	if !strings.Contains(e.RetrievedVia, "citation metadata") {
		t.Errorf("the hop is not recorded as provenance: %q", e.RetrievedVia)
	}

	// A SHORTER ANSWER IS REFUSED. The pointer is the publisher's, so this is not deciding which
	// is the paper — it is refusing to trade a page for a smaller one, which is what a paywall
	// stub or an error page would be.
	g := fake(func(u string) (*Response, error) {
		switch u {
		case landing:
			return &Response{Body: []byte(abstract), ContentType: "text/html"}, nil
		case full:
			return &Response{Body: []byte("<html>register to continue</html>"), ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 404}
	})
	run2 := runtest.New(t, t.TempDir())
	if _, kept, _, err := Resolve(run2, landing, g); err != nil {
		t.Fatalf("resolve: %v", err)
	} else if len(kept) != len(abstract) {
		t.Errorf("a smaller answer replaced the page we had: %d bytes", len(kept))
	}
}
