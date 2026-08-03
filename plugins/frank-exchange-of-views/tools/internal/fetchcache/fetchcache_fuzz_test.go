package fetchcache

import (
	"testing"
	"unicode/utf8"
)

// FuzzStoreRoundTrips asserts the content-addressed cache holds its invariants for
// arbitrary bytes and URLs: Store's hash is stable (same bytes → same sha), storing dedups
// on content, and a Lookup after a Store returns the exact bytes under that URL. It must
// never panic on any input.
func FuzzStoreRoundTrips(f *testing.F) {
	f.Add([]byte("hello"), "https://a")
	f.Add([]byte(""), "https://empty")
	f.Add([]byte("bytes with \x00 nulls and \n newlines"), "https://weird/path?q=1&r=2")

	f.Fuzz(func(t *testing.T, body []byte, url string) {
		// The cache's input domain is real URLs — a value that already passed the Fetcher's
		// url-parse + GET before Store ever sees it. A URL is always non-empty and valid
		// UTF-8 (JSON, and HTTP itself, cannot losslessly carry invalid UTF-8); anything else
		// is outside the contract, so it is not a round-trip case.
		if url == "" || !utf8.ValidString(url) {
			return
		}
		run := t.TempDir()

		sha, err := Store(run, url, body)
		if err != nil {
			t.Fatalf("Store: %v", err)
		}
		if sha != Sha(body) {
			t.Fatalf("Store sha %s != Sha(body) %s — hash not content-stable", sha, Sha(body))
		}

		// Dedup: storing the same bytes again yields the same hash and no error.
		if sha2, err := Store(run, url, body); err != nil || sha2 != sha {
			t.Fatalf("re-Store: sha=%s err=%v, want stable %s", sha2, err, sha)
		}

		// Round-trip: Lookup returns the exact bytes under this URL.
		gotSha, gotBody, ok, err := Lookup(run, url)
		if err != nil {
			t.Fatalf("Lookup: %v", err)
		}
		if !ok {
			t.Fatalf("Lookup miss for a URL just stored (%q)", url)
		}
		if gotSha != sha || string(gotBody) != string(body) {
			t.Fatalf("Lookup round-trip mismatch: sha %s/%s, body %d/%d bytes", gotSha, sha, len(gotBody), len(body))
		}
	})
}
