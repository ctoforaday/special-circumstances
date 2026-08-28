package fetchcache

import (
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"testing"
)

// stubFetcher is the offline Fetcher every cache test drives — it counts calls so a test
// can prove a second read was served from cache (Fetch NOT re-entered).
type stubFetcher struct {
	calls int
	resp  map[string][]byte
	err   error
}

func (s *stubFetcher) Fetch(url string) ([]byte, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if b, ok := s.resp[url]; ok {
		return b, nil
	}
	return nil, fmt.Errorf("no stub response for %s", url)
}

func TestResolveMissThenHit(t *testing.T) {
	run := t.TempDir()
	body := []byte("the source bytes both sides read")
	f := &stubFetcher{resp: map[string][]byte{"https://ex/1": body}}

	sha, got, hit, err := Resolve(runtest.Open(t, run), "https://ex/1", f)
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
	if sha != Sha(body) {
		t.Errorf("sha = %q, want %q", sha, Sha(body))
	}
	if _, err := os.Stat(Path(runtest.Open(t, run), sha)); err != nil {
		t.Errorf("cache file not written: %v", err)
	}

	// Second read: same URL → served from cache, Fetch NOT re-entered.
	sha2, got2, hit2, err := Resolve(runtest.Open(t, run), "https://ex/1", f)
	if err != nil {
		t.Fatalf("second Resolve: %v", err)
	}
	if !hit2 {
		t.Errorf("second Resolve reported a miss — the URL was already cached")
	}
	if f.calls != 1 {
		t.Errorf("second Resolve fetched again (calls=%d) — download-once violated", f.calls)
	}
	if sha2 != sha || string(got2) != string(body) {
		t.Errorf("second Resolve = (%s,%q), want (%s,%q)", sha2, got2, sha, body)
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
	sha1, err := Store(runtest.Open(t, run), "https://a", body)
	if err != nil {
		t.Fatal(err)
	}
	sha2, err := Store(runtest.Open(t, run), "https://b", body) // different URL, same bytes
	if err != nil {
		t.Fatal(err)
	}
	if sha1 != sha2 {
		t.Errorf("same bytes hashed differently: %s vs %s", sha1, sha2)
	}
	// One content file (content-addressed dedup), two index lines (two URLs).
	files, _ := filepath.Glob(filepath.Join(Dir(runtest.Open(t, run)), "*"))
	content := 0
	for _, f := range files {
		if filepath.Base(f) == sha1 {
			content++
		}
	}
	if content != 1 {
		t.Errorf("content files for one hash = %d, want 1", content)
	}
	if s1, _, ok1, _ := Lookup(runtest.Open(t, run), "https://a"); !ok1 || s1 != sha1 {
		t.Errorf("Lookup a = (%s,%v), want (%s,true)", s1, ok1, sha1)
	}
	if s2, _, ok2, _ := Lookup(runtest.Open(t, run), "https://b"); !ok2 || s2 != sha2 {
		t.Errorf("Lookup b = (%s,%v), want (%s,true)", s2, ok2, sha2)
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
