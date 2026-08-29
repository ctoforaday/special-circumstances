package fetchcache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	e1, got, hit, err := Resolve(run, "https://ex/1", f)
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
	if _, err := os.Stat(Path(run, e1.Sha)); err != nil {
		t.Errorf("cache file not written: %v", err)
	}

	// Second read: same URL → served from cache, Fetch NOT re-entered.
	e2, got2, hit2, err := Resolve(run, "https://ex/1", f)
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

	_, _, _, err := Resolve(run, "https://gone/x", f)
	if err == nil {
		t.Fatal("Resolve of an unreachable URL returned nil error")
	}
	// No cache dir/index written on a pure failure — nothing to serve, nothing to leak.
	if entries, _ := os.ReadDir(Dir(run)); len(entries) != 0 {
		t.Errorf("a failed fetch left %d cache entries, want 0", len(entries))
	}
	// A retry after the source comes back succeeds and caches (failure was not sticky).
	f.err = nil
	f.resp = map[string][]byte{"https://gone/x": []byte("now up")}
	if _, _, hit, err := Resolve(run, "https://gone/x", f); err != nil || hit {
		t.Errorf("retry after recovery: hit=%v err=%v, want (false,nil)", hit, err)
	}
}

func TestStoreDedupsOnContent(t *testing.T) {
	run := t.TempDir()
	body := []byte("identical bytes")
	se1, err := Store(run, Entry{URL: "https://a"}, body)
	if err != nil {
		t.Fatal(err)
	}
	se2, err := Store(run, Entry{URL: "https://b"}, body) // different URL, same bytes
	if err != nil {
		t.Fatal(err)
	}
	if se1.Sha != se2.Sha {
		t.Errorf("same bytes hashed differently: %s vs %s", se1.Sha, se2.Sha)
	}
	// One content file (content-addressed dedup), two index lines (two URLs).
	files, _ := filepath.Glob(filepath.Join(Dir(run), "*"))
	content := 0
	for _, f := range files {
		if filepath.Base(f) == se1.Sha {
			content++
		}
	}
	if content != 1 {
		t.Errorf("content files for one hash = %d, want 1", content)
	}
	if g1, _, ok1, _ := Lookup(run, "https://a"); !ok1 || g1.Sha != se1.Sha {
		t.Errorf("Lookup a missed or resolved to the wrong hash (ok=%v), want %s", ok1, se1.Sha)
	}
	if g2, _, ok2, _ := Lookup(run, "https://b"); !ok2 || g2.Sha != se2.Sha {
		t.Errorf("Lookup b missed or resolved to the wrong hash (ok=%v), want %s", ok2, se2.Sha)
	}
}

// A crash between the content write and the index append leaves an index line pointing at
// a missing file. Lookup must treat that as a MISS so the next fetch self-heals, not as a
// hit that reads a phantom.
func TestLookupTreatsMissingContentAsMiss(t *testing.T) {
	run := t.TempDir()
	if err := os.MkdirAll(Dir(run), 0o755); err != nil {
		t.Fatal(err)
	}
	// Hand-write an index entry for a sha whose content file does not exist.
	if err := os.WriteFile(indexPath(run), []byte(`{"sha":"deadbeef","url":"https://orphan"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := Lookup(run, "https://orphan"); err != nil || ok {
		t.Errorf("Lookup of an orphaned index line = (ok=%v,err=%v), want (false,nil)", ok, err)
	}
}

func TestLookupUncachedURLIsMiss(t *testing.T) {
	run := t.TempDir()
	if _, _, ok, err := Lookup(run, "https://never"); err != nil || ok {
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
	if err := os.MkdirAll(Dir(run), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(run, sha), body, 0o644); err != nil {
		t.Fatal(err)
	}
	// The old two-field shape, verbatim.
	line := `{"sha":"` + sha + `","url":"https://old/entry"}` + "\n"
	if err := os.WriteFile(indexPath(run), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	got, gotBody, ok, err := Lookup(run, "https://old/entry")
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

	got, _, _, err := Resolve(run, "https://ex/paper", f)
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
	again, _, hit, err := Resolve(run, "https://ex/paper", f)
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
func sameEntry(a, b Entry) bool {
	if (a.TextExtracted == nil) != (b.TextExtracted == nil) {
		return false
	}
	if a.TextExtracted != nil && *a.TextExtracted != *b.TextExtracted {
		return false
	}
	a.TextExtracted, b.TextExtracted = nil, nil
	return a == b
}

func showEntry(e Entry) string {
	extracted := "nil"
	if e.TextExtracted != nil {
		extracted = fmt.Sprintf("%v", *e.TextExtracted)
	}
	e.TextExtracted = nil
	return fmt.Sprintf("%+v text_extracted=%s", e, extracted)
}
