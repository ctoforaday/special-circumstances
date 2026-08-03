package fetchcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The scheme allowlist is checked BEFORE any network call, so these cases need no server.
func TestHTTPFetcherRefusesNonWebSchemes(t *testing.T) {
	f := NewHTTPFetcher()
	for _, u := range []string{"file:///etc/passwd", "ftp://host/x", "gopher://h", "data:text/plain,hi"} {
		if _, err := f.Fetch(u); err == nil || !strings.Contains(err.Error(), "scheme") {
			t.Errorf("Fetch(%q) = %v, want a scheme-refusal error", u, err)
		}
	}
}

// A body larger than the cap is DETECTED (errors), never silently truncated into a
// citation. httptest is in-process loopback — CI-safe, no external network.
func TestHTTPFetcherEnforcesSizeCap(t *testing.T) {
	big := strings.Repeat("x", maxFetchBytes+100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, big)
	}))
	defer srv.Close()

	if _, err := NewHTTPFetcher().Fetch(srv.URL); err == nil || !strings.Contains(err.Error(), "cap") {
		t.Errorf("Fetch of an over-size body = %v, want a size-cap error", err)
	}
}

// A body at or under the cap comes back whole.
func TestHTTPFetcherReturnsBodyUnderCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "small source")
	}))
	defer srv.Close()

	b, err := NewHTTPFetcher().Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch under cap: %v", err)
	}
	if string(b) != "small source" {
		t.Errorf("body = %q, want %q", b, "small source")
	}
}

// A redirect chain beyond the cap is refused, not followed forever.
func TestHTTPFetcherCapsRedirects(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL+r.URL.Path+"x", http.StatusFound) // always bounces onward
	}))
	defer srv.Close()

	if _, err := NewHTTPFetcher().Fetch(srv.URL); err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Errorf("Fetch of an endless redirect = %v, want a redirect-cap error", err)
	}
}

// A non-200 is an error, not a cached body.
func TestHTTPFetcherRejectsNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := NewHTTPFetcher().Fetch(srv.URL); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("Fetch of a 404 = %v, want an HTTP-status error", err)
	}
}
