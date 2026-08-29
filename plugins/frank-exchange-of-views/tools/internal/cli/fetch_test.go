package cli

import (
	"encoding/json"
	"errors"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
)

// errFake is the canned failure the offline Fetcher returns when a test drives the
// unreachable-source path.
var errFake = errors.New("fake fetch failure")

// fakeFetcher is the offline Fetcher the fetch-command tests swap in for the prod HTTP one.
type fakeFetcher struct {
	resp        map[string][]byte
	contentType string
	disposition string
	err         error
	calls       int
}

func (f *fakeFetcher) Fetch(u string) (*fetchcache.Response, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if b, ok := f.resp[u]; ok {
		return &fetchcache.Response{Body: b, ContentType: f.contentType, Disposition: f.disposition}, nil
	}
	return nil, errors.New("no fake response for " + u)
}

func withFetcher(t *testing.T, f fetchcache.Fetcher) {
	t.Helper()
	prev := fetchcache.Default
	fetchcache.Default = f
	t.Cleanup(func() { fetchcache.Default = prev })
}

// withExtractor swaps the real PDFium engine out. EVERY cli test needs this, because the real
// one compiles a 5 MB WebAssembly module on first use — a four-second tax on a test suite that
// is not testing extraction. The stub also lets a test STATE what a document contains instead
// of shipping a fixture PDF to make the point.
func withExtractor(t *testing.T, e fetchcache.Extractor) {
	t.Helper()
	prev := fetchcache.DefaultExtractor
	fetchcache.DefaultExtractor = e
	t.Cleanup(func() { fetchcache.DefaultExtractor = prev })
}

// stubExtractor returns whatever a test hands it, for any media type it is given.
type stubExtractor struct{ out fetchcache.Extraction }

func (s stubExtractor) Extract(_, _ string, _ []byte) fetchcache.Extraction { return s.out }

// THE BODY MUST NOT APPEAR, FOR ANY CONTENT TYPE (#629 D1). The assertion is written as an
// ABSENCE on purpose: a test that only checks the summary is present would still pass with the
// body concatenated after it, which is exactly the regression this guards. The sentinel is a
// string that cannot occur in a summary, so a match is proof the document leaked.
func TestFetchNeverPrintsTheBodyAndCachesForReuse(t *testing.T) {
	dir := recordtest.TmpRun(t)
	const body = "SENTINEL-source-body-that-must-not-be-printed"
	f := &fakeFetcher{resp: map[string][]byte{"https://ex/a": []byte(body)}, contentType: "text/html"}
	withFetcher(t, f)
	withExtractor(t, stubExtractor{})

	out, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/a")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if strings.Contains(out, body) {
		t.Errorf("fetch printed the document body; D1 says it never may. got:\n%s", out)
	}
	// And it must still say where the body IS, or the seat has lost the source entirely.
	sha := fetchcache.Sha([]byte(body))
	if !strings.Contains(out, fetchcache.Path(dir, sha)) {
		t.Errorf("summary does not name the cache path; a seat cannot read what it cannot locate. got:\n%s", out)
	}
	if !strings.Contains(out, "content_type: text/html") {
		t.Errorf("summary does not carry the content type measured at fetch. got:\n%s", out)
	}
	// NOTHING WAS EXTRACTED, AND THE SUMMARY MUST NOT CLAIM IT WAS LOOKED AT. An HTML page is
	// entirely text; printing `text_extracted: false` for one would be a lie about the document
	// rather than a statement that no extractor ran.
	if strings.Contains(out, "text_extracted") {
		t.Errorf("summary reported text_extracted for a type nothing extracts — nil must print nothing. got:\n%s", out)
	}
	if f.calls != 1 {
		t.Errorf("fetched %d times, want 1", f.calls)
	}

	// Second fetch of the same URL is served from cache — Fetch not re-entered.
	out2, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/a")
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if f.calls != 1 {
		t.Errorf("second fetch re-entered Fetch: calls=%d, want 1", f.calls)
	}
	if !strings.Contains(out2, "cache_hit: true") {
		t.Errorf("second fetch did not report a cache hit. got:\n%s", out2)
	}
	if strings.Contains(out2, body) {
		t.Errorf("the CACHE-HIT path printed the body; D1 binds both paths. got:\n%s", out2)
	}
}

// THE EXTRACTION IS AN ARTIFACT ON DISK, NOT A LINE OF OUTPUT. The test reads the file the
// summary points at, because a text_path naming a file that is not there is worse than no
// field: a seat would Read it, get a not-found, and not know whether the tool or the document
// was at fault.
func TestFetchWritesTheExtractionBesideTheDocumentAndNamesBoth(t *testing.T) {
	dir := recordtest.TmpRun(t)
	const pdf = "%PDF-1.7 pretend bytes"
	const text = "The key idea behind active learning"
	f := &fakeFetcher{resp: map[string][]byte{"https://ex/p": []byte(pdf)}, contentType: "application/pdf"}
	withFetcher(t, f)
	withExtractor(t, stubExtractor{out: fetchcache.Extraction{
		Attempted: true, Text: text, Title: "Active Learning Literature Survey",
		Pages: 67, ExtractorID: "stub@v1",
	}})

	out, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/p", "--json")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	var got fetchSummary
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("fetch --json did not emit an object: %v (%s)", err, out)
	}
	if got.TextExtracted == nil || !*got.TextExtracted {
		t.Fatalf("text_extracted = %v, want true", got.TextExtracted)
	}
	onDisk, rerr := os.ReadFile(got.TextPath)
	if rerr != nil {
		t.Fatalf("summary named text_path %q and nothing is there: %v", got.TextPath, rerr)
	}
	if string(onDisk) != text {
		t.Errorf("extraction on disk = %q, want %q", onDisk, text)
	}
	// The extraction is HASHED, which is the whole point: the body's sha proves the document is
	// unchanged and says nothing about the text a claim was drawn from.
	if got.TextSha256 != fetchcache.Sha([]byte(text)) {
		t.Errorf("text_sha256 = %q, want the hash of the extraction itself", got.TextSha256)
	}
	if got.TextSha256 == got.Sha256 {
		t.Error("the extraction and the document hashed identically — one of them is not being hashed")
	}
	if got.Filename != "Active Learning Literature Survey" {
		t.Errorf("filename = %q, want the document Title (D4 rung 1)", got.Filename)
	}
	if got.Pages != 67 {
		t.Errorf("pages = %d, want 67", got.Pages)
	}
	if strings.Contains(out, pdf) {
		t.Error("the PDF body reached the summary")
	}
}

// AN EMPTY EXTRACTION IS A STATED FACT, NEVER AN EMPTY FILE. This is the IEEE 1012 case, and
// the assertion that matters is the ABSENCE of the .txt: a zero-byte extraction on disk reads
// exactly like a successful extraction of an empty document, which is the plausible zero
// [[facts-are-fields]] forbids.
func TestAnEmptyExtractionWritesNoFileAndStatesWhy(t *testing.T) {
	dir := recordtest.TmpRun(t)
	f := &fakeFetcher{resp: map[string][]byte{"https://ex/scan": []byte("%PDF-1.2 a 1998 scan")}, contentType: "application/pdf"}
	withFetcher(t, f)
	withExtractor(t, stubExtractor{out: fetchcache.Extraction{
		Attempted: true, Text: "", Pages: 80, ExtractorID: "stub@v1",
		Reason: "no text layer: 80 pages, all empty",
	}})

	out, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/scan", "--json")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	var got fetchSummary
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("fetch --json: %v (%s)", err, out)
	}
	if got.TextExtracted == nil {
		t.Fatal("text_extracted is absent — an attempted-and-empty extraction must record false, not nothing")
	}
	if *got.TextExtracted {
		t.Error("text_extracted = true for an extraction that produced no text")
	}
	if got.TextReason == "" {
		t.Error("text_extracted is false with no reason — the miss must be loud")
	}
	if got.TextPath != "" {
		t.Errorf("text_path = %q for an empty extraction; it must name nothing", got.TextPath)
	}
	if _, statErr := os.Stat(fetchcache.TextPath(dir, got.Sha256)); !os.IsNotExist(statErr) {
		t.Errorf("an empty .txt was written; it is indistinguishable from a successful empty extraction (stat err: %v)", statErr)
	}
}

func TestFetchFailureIsANonZeroErrorNotAFriction(t *testing.T) {
	dir := recordtest.TmpRun(t)
	withFetcher(t, &fakeFetcher{err: errors.New("host unreachable")})

	_, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://gone")
	if err == nil {
		t.Fatal("fetch of an unreachable URL returned nil error")
	}
	if !strings.Contains(err.Error(), "fetch:") {
		t.Errorf("error = %v, want a fetch-prefixed message", err)
	}
	// fetch takes no seat and writes no event — so nothing (least of all a friction) is
	// on the record. The records dir is never even created by a bare read.
	if _, statErr := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://gone"); statErr == nil {
		t.Error("a repeated failed fetch unexpectedly succeeded")
	}
}

func TestFetchRequiresRunAndURL(t *testing.T) {
	dir := recordtest.TmpRun(t)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{}})
	if _, err := run(t, "fetch", "--seat-id", "operator", "--run", dir); err == nil {
		t.Error("fetch without --url did not error")
	}
	if _, err := run(t, "fetch", "--seat-id", "operator", "--url", "https://x"); err == nil {
		t.Error("fetch without --run did not error")
	}
}

func TestFetchJSONReportsCacheHit(t *testing.T) {
	dir := recordtest.TmpRun(t)
	f := &fakeFetcher{resp: map[string][]byte{"https://ex/j": []byte("jbody")}}
	withFetcher(t, f)

	first, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/j", "--json")
	if err != nil {
		t.Fatalf("json fetch: %v", err)
	}
	if !strings.Contains(first, `"cache_hit":false`) {
		t.Errorf("first --json fetch = %s, want cache_hit false", first)
	}
	second, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/j", "--json")
	if err != nil {
		t.Fatalf("json fetch 2: %v", err)
	}
	if !strings.Contains(second, `"cache_hit":true`) {
		t.Errorf("second --json fetch = %s, want cache_hit true", second)
	}
}
