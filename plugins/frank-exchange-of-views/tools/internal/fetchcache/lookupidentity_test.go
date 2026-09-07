package fetchcache

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A CACHE THAT ANSWERS EVERY QUESTION IS NOT A CACHE. Lookup walks the append-only index
// and matches on URL; drop that match and the first readable entry is returned for
// whatever was asked, so a seat citing one source is handed another — with `cache_hit:
// true` beside it and nothing anywhere saying the wrong document was served. The miss
// case is the one that pins the identity, and it had no test.
func TestLookupServesOnlyTheURLItWasAsked(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	for _, s := range []struct{ url, body string }{
		{"https://ex/one", "the first document"},
		{"https://ex/two", "the second document"},
	} {
		if _, err := Store(run, Entry{URL: s.url, ContentType: "text/plain"}, []byte(s.body)); err != nil {
			t.Fatal(err)
		}
	}

	e, b, ok, err := Lookup(run, "https://ex/never-fetched")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("an unfetched url was served from cache as %q (%q) — every fetch in the run "+
			"would be answered with the first document the index holds", e.URL, b)
	}

	// And the hit path still resolves the entry it was asked for, not merely the first one.
	got, body, ok, err := Lookup(run, "https://ex/two")
	if err != nil || !ok {
		t.Fatalf("the stored second entry did not resolve: ok=%v err=%v", ok, err)
	}
	if got.URL != "https://ex/two" || string(body) != "the second document" {
		t.Errorf("lookup returned %q / %q, want the entry named in the request", got.URL, body)
	}
}
