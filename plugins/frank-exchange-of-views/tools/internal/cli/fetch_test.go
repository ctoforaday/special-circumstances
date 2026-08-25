package cli

import (
	"errors"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
)

// errFake is the canned failure the offline Fetcher returns when a test drives the
// unreachable-source path.
var errFake = errors.New("fake fetch failure")

// fakeFetcher is the offline Fetcher the fetch-command tests swap in for the prod HTTP one.
type fakeFetcher struct {
	resp  map[string][]byte
	err   error
	calls int
}

func (f *fakeFetcher) Fetch(u string) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	if b, ok := f.resp[u]; ok {
		return b, nil
	}
	return nil, errors.New("no fake response for " + u)
}

func withFetcher(t *testing.T, f fetchcache.Fetcher) {
	t.Helper()
	prev := fetchcache.Default
	fetchcache.Default = f
	t.Cleanup(func() { fetchcache.Default = prev })
}

func TestFetchPrintsBodyAndCachesForReuse(t *testing.T) {
	dir := recordtest.TmpRun(t)
	f := &fakeFetcher{resp: map[string][]byte{"https://ex/a": []byte("source body")}}
	withFetcher(t, f)

	out, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/a")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if out != "source body" {
		t.Errorf("fetch printed %q, want %q", out, "source body")
	}
	if f.calls != 1 {
		t.Errorf("fetched %d times, want 1", f.calls)
	}

	// Second fetch of the same URL is served from cache — Fetch not re-entered.
	out2, err := run(t, "fetch", "--seat-id", "operator", "--run", dir, "--url", "https://ex/a")
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if out2 != "source body" || f.calls != 1 {
		t.Errorf("second fetch: out=%q calls=%d, want (%q,1)", out2, f.calls, "source body")
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
