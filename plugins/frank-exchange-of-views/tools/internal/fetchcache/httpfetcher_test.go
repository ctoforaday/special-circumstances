package fetchcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The scheme allowlist is checked BEFORE any network call, so these cases need no server.
//
// IT ASSERTS OUR REFUSAL, NOT ANY REFUSAL, and the difference is the whole test. The
// first version matched an error merely CONTAINING "scheme" — and net/http's own
// transport says `unsupported protocol scheme "file"`, so the standard library's
// accident satisfied the assertion and this barrier could be deleted with the suite
// green (measured: disabling the allowlist left this test passing). Two things are
// pinned now: the message is the one this package writes, and no RoundTrip happens at
// all, which is what "checked before any network call" means.
func TestHTTPFetcherRefusesNonWebSchemes(t *testing.T) {
	var attempted []string
	f := &httpFetcher{
		client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			attempted = append(attempted, r.URL.String())
			return nil, fmt.Errorf("no request should have been made")
		})},
		maxBytes: maxFetchBytes,
	}
	for _, u := range []string{"file:///etc/passwd", "ftp://host/x", "gopher://h", "data:text/plain,hi"} {
		_, err := f.Fetch(u)
		if err == nil {
			t.Errorf("Fetch(%q) was allowed", u)
			continue
		}
		for _, want := range []string{"refused scheme", "only http and https"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("Fetch(%q) = %v, want THIS package's refusal naming %q", u, err, want)
			}
		}
	}
	if len(attempted) > 0 {
		t.Errorf("the allowlist let %v reach the transport; it must refuse before any network call", attempted)
	}
}

// roundTripFunc adapts a function to http.RoundTripper so a test can prove a request was
// never made.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

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

	resp, err := NewHTTPFetcher().Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch under cap: %v", err)
	}
	if string(resp.Body) != "small source" {
		t.Errorf("body = %q, want %q", resp.Body, "small source")
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

// THE FETCHER MUST IDENTIFY ITSELF. Go's default "Go-http-client/1.1" is blocked outright by
// major sources: the 2026-08-04 smoke lost four citations to Wikipedia 403s, including the most
// obvious source for the question being researched. This asserts the header actually reaches the
// server — the previous code compiled, passed every test, and sent nothing.
func TestHTTPFetcherSendsDescriptiveUserAgent(t *testing.T) {
	got := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.UserAgent()
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	if _, err := NewHTTPFetcher().Fetch(srv.URL); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	ua := <-got
	if ua == "" || strings.HasPrefix(ua, "Go-http-client") {
		t.Fatalf("User-Agent = %q — the default (or empty) agent is what sources 403", ua)
	}
	if !strings.Contains(ua, "feov-record") {
		t.Errorf("User-Agent = %q, want it to name the tool so an operator can attribute the traffic", ua)
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
