package fetchcache

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// ONE RESOURCE, TWO SPELLINGS — AND TWO RESOURCES THAT ONLY LOOK ALIKE. Each pair is equivalent
// under RFC 3986 §6.2.2 or it is not; the second half keeps the normalisation from swallowing a
// different path, query or host.
func TestSameResourceIsRFC3986Equivalence(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		same bool
	}{
		{"https://sites.math.rutgers.edu/%7Esdmiller/a%20b.pdf", "https://sites.math.rutgers.edu/~sdmiller/a%20b.pdf", true},
		{"https://Example.ORG/x", "https://example.org/x", true},
		{"https://example.org:443/x", "https://example.org/x", true},
		{"http://example.org:80/x", "http://example.org/x", true},
		{"https://example.org", "https://example.org/", true},
		{"https://example.org/x#sec", "https://example.org/x", true},
		{"https://example.org/a%2fb", "https://example.org/a%2Fb", true},
		{"https://example.org/q?a=%7E", "https://example.org/q?a=~", true},
		{"https://example.org/a%2Fb", "https://example.org/a/b", false},
		{"https://example.org/x/", "https://example.org/x", false},
		{"https://example.org/X", "https://example.org/x", false},
		{"https://example.org/q?a=1", "https://example.org/q?a=2", false},
		{"https://example.org:8443/x", "https://example.org/x", false},
		{"http://example.org/x", "https://example.org/x", false},
		{"not a url", "not  a url", false},
	} {
		if got := SameResource(tc.a, tc.b); got != tc.same {
			t.Errorf("SameResource(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.same)
		}
	}
}

// A PERCENT-ENCODING TWIN IS NOT A BOT WALL (smoke 07d, log 59 and 199). The lecture was cached
// under `%7E` and under `~`; the shared-body warning named the one beside the other, and two seats
// each spent a read confirming the source was real. A genuinely different url on the same bytes
// is still named.
func TestATwinSpellingIsNotASharedBody(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := []byte("%PDF-1.4 the Miller-Rabin lecture\n")
	tilde := "https://sites.math.rutgers.edu/~sdmiller/Lecture%2014.pdf"
	escaped := "https://sites.math.rutgers.edu/%7Esdmiller/Lecture%2014.pdf"
	for _, u := range []string{tilde, escaped} {
		if _, err := Store(run, Entry{URL: u}, body); err != nil {
			t.Fatal(err)
		}
	}
	if shared, err := URLsSharingBody(run, Sha(body), escaped); err != nil || len(shared) != 0 {
		t.Fatalf("a twin spelling was named as a shared body: %v (err %v)", shared, err)
	}
	other := "https://www.jstor.org/stable/1"
	if _, err := Store(run, Entry{URL: other}, body); err != nil {
		t.Fatal(err)
	}
	shared, err := URLsSharingBody(run, Sha(body), escaped)
	if err != nil || len(shared) != 1 || shared[0] != other {
		t.Errorf("shared = %v (err %v), want only %s", shared, err, other)
	}
}
