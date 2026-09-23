package fetchcache

import (
	"compress/gzip"
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
		// THE FETCHER ASKS FOR THE RULES BEFORE IT ASKS FOR THE DOCUMENT, so a fixture that
		// records every request sees two. This one has a 1-buffered channel and deadlocked on
		// the second until it learned to answer the first.
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
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

// TRANSPORT COMPRESSION ARRIVES DECOMPRESSED, and the way to break that is to try to help.
//
// net/http adds `Accept-Encoding: gzip` and decodes the response itself — but only while the
// caller sets no Accept-Encoding header of its own. Setting one, even to the same value, puts the
// transport in raw mode and hands back compressed bytes with the header still on. Every gzipped
// page would then sniff as a container and be recorded as "not the source's text", with a reason
// that reads entirely plausible.
//
// The invariant is therefore a SILENCE — a header we do NOT set — which nothing would notice
// breaking. This asserts the behaviour rather than the absence, so it fails whether the cause is
// an Accept-Encoding header, DisableCompression, or a hand-rolled transport.
func TestAGzippedResponseArrivesDecompressed(t *testing.T) {
	tempPaceDir(t)
	body := "<html><body>" + strings.Repeat("the paper. ", 200) + "</body></html>"
	var sawAcceptEncoding string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		sawAcceptEncoding = r.Header.Get("Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/html")
		zw := gzip.NewWriter(w)
		_, _ = zw.Write([]byte(body))
		_ = zw.Close()
	}))
	defer srv.Close()

	resp, err := NewHTTPFetcher().Fetch(srv.URL + "/article")
	if err != nil {
		t.Fatalf("a gzipped page failed to fetch: %v", err)
	}
	if !strings.Contains(sawAcceptEncoding, "gzip") {
		t.Errorf("the transport did not offer gzip (Accept-Encoding: %q) — it does so only while we set no header of our own", sawAcceptEncoding)
	}
	if string(resp.Body) != body {
		t.Errorf("the body arrived still compressed: %d bytes, starting %q", len(resp.Body), resp.Body[:min(8, len(resp.Body))])
	}
	// AND THE TYPE IS STILL THE DOCUMENT'S. A body left compressed would sniff as a container
	// and be withdrawn as "not the source's text" — the failure this guards, one step on.
	if !TextBearing(resp.ContentType) {
		t.Errorf("content type %q no longer reads as the source's text", resp.ContentType)
	}
}

// AN ENCODING WE NEVER OFFERED IS REFUSED BY NAME, because of what accepting it does next.
//
// net/http strips Content-Encoding when it decoded the body itself, so a header still present
// names a coding nothing decoded. Measured before this guard: a `Content-Encoding: br` response
// arrived with err nil, the source's own `text/html`, and 38 bytes of undecoded data. Stored,
// the shell detector would have run on it, found no prose, and recorded `not_renderable` with
// the reason "too little prose to be a document" — a confident diagnosis of the wrong thing.
func TestAnEncodingWeNeverOfferedIsRefusedByName(t *testing.T) {
	tempPaceDir(t)
	for _, enc := range []string{"br", "zstd", "deflate"} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Encoding", enc)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("\x1b\x2e\x00\x00opaque"))
		}))
		_, err := NewHTTPFetcher().Fetch(srv.URL + "/a")
		srv.Close()
		if err == nil {
			t.Fatalf("%s: an undecoded body was accepted as the document", enc)
		}
		if !strings.Contains(err.Error(), enc) {
			t.Errorf("%s: the refusal does not name the coding, so nobody can act on it: %v", enc, err)
		}
	}
	// AND gzip STILL PASSES, decoded by the transport with the header stripped. A guard that
	// refused the one coding this client does negotiate would break every compressed page.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "text/html")
		zw := gzip.NewWriter(w)
		_, _ = zw.Write([]byte("<html><body>" + strings.Repeat("the paper. ", 200) + "</body></html>"))
		_ = zw.Close()
	}))
	defer srv.Close()
	if _, err := NewHTTPFetcher().Fetch(srv.URL + "/a"); err != nil {
		t.Errorf("a gzipped page was refused by the encoding guard: %v", err)
	}
}
