package fetchcache

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
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

// AN ENCODING WE DID NOT ASK FOR IS DECODED WHEN WE CAN READ IT, AND NAMED WHEN WE CANNOT.
//
// Servers are not supposed to send a coding that was never offered, and they do — a CDN
// normalises to brotli, a proxy re-encodes, an origin answers `deflate` because it always has.
// Refusing a body we could perfectly well read throws away a document over a header.
//
// The first version of this guard refused every coding alike. That was wrong for exactly the case
// worth handling: one we support and did not request.
func TestAnUnsolicitedCodingIsDecodedWhereWeHaveAReader(t *testing.T) {
	tempPaceDir(t)
	page := "<html><body>" + strings.Repeat("the paper. ", 200) + "</body></html>"

	serve := func(enc string, encode func(w io.Writer)) (*Response, error) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Encoding", enc)
			w.Header().Set("Content-Type", "text/html")
			encode(w)
		}))
		defer srv.Close()
		return NewHTTPFetcher().Fetch(srv.URL + "/a")
	}

	// zlib-wrapped deflate, which is what RFC 9110 defines `deflate` to be.
	resp, err := serve("deflate", func(w io.Writer) {
		zw := zlib.NewWriter(w)
		_, _ = zw.Write([]byte(page))
		_ = zw.Close()
	})
	if err != nil {
		t.Fatalf("zlib-wrapped deflate was refused: %v", err)
	}
	if string(resp.Body) != page {
		t.Errorf("zlib deflate did not round-trip: %d bytes", len(resp.Body))
	}

	// AND RAW DEFLATE, which a long tail of servers send under the same name. Guessing one of the
	// two would drop half the responses, and the ambiguity is the wire's rather than ours.
	resp, err = serve("deflate", func(w io.Writer) {
		fw, _ := flate.NewWriter(w, flate.DefaultCompression)
		_, _ = fw.Write([]byte(page))
		_ = fw.Close()
	})
	if err != nil {
		t.Fatalf("raw deflate was refused: %v", err)
	}
	if string(resp.Body) != page {
		t.Errorf("raw deflate did not round-trip: %d bytes", len(resp.Body))
	}

	// A CODING WITH NO READER IS REFUSED BY NAME, so what is missing is a decoder and anyone
	// reading the record can see which. Accepting it would keep the source's `text/html`, run the
	// shell detector over undecoded data, and record "too little prose to be a document" — a
	// confident diagnosis of the wrong thing.
	for _, enc := range []string{"br", "zstd"} {
		_, err := serve(enc, func(w io.Writer) { _, _ = w.Write([]byte("\x1b\x2e\x00\x00opaque")) })
		if err == nil {
			t.Fatalf("%s: an undecodable body was accepted as the document", enc)
		}
		if !strings.Contains(err.Error(), enc) {
			t.Errorf("%s: the refusal does not name the coding: %v", enc, err)
		}
	}
}

// THE Link HEADER IS CAPTURED OFF THE WIRE, which is the half a hand-built Response cannot test.
//
// Signposting had a parser here and no caller: both passed an empty string, because nothing read
// the header off the response. Deleting the capture left the whole suite green — the tests
// constructed their own Response and so exercised the parser while the plumbing was missing.
func TestTheLinkHeaderIsReadOffTheResponse(t *testing.T) {
	tempPaceDir(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		w.Header().Add("Link", `<https://repo.example/12345/paper.pdf>; rel="item"; type="application/pdf"`)
		w.Header().Add("Link", `<https://repo.example/12345/>; rel="cite-as"`)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>" + strings.Repeat("record page. ", 60) + "</body></html>"))
	}))
	defer srv.Close()

	resp, err := NewHTTPFetcher().Fetch(srv.URL + "/12345/")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !strings.Contains(resp.LinkHeader, "rel=\"item\"") {
		t.Fatalf("the Link header did not reach the Response: %q", resp.LinkHeader)
	}
	// BOTH VALUES, JOINED. A host may send several Link headers rather than one comma-joined
	// value, and taking only the first would read whichever the server happened to put first.
	if !strings.Contains(resp.LinkHeader, "cite-as") {
		t.Errorf("a second Link header was dropped: %q", resp.LinkHeader)
	}
	base, _ := neturl.Parse(srv.URL + "/12345/")
	if got := LandingPageFullText(resp.ContentType, resp.Body, resp.LinkHeader, base); got != "https://repo.example/12345/paper.pdf" {
		t.Errorf("the signpost was not read end to end: %q", got)
	}
}

// A 2xx THAT IS NOT 200 IS NOT A REFUSAL.
//
// MEASURED on doi 10.1109/5.18626 during the 2026-09-24 rerun: IEEE answered 202 Accepted and the
// record said `refusal_class: origin` — the publisher turning us away, when what the publisher
// said was "received, being processed". A seat reading that concludes the source is closed to it.
// "403" and "202" license opposite next moves and the bare number reads the same either way.
func TestANonOKSuccessIsNotRecordedAsARefusal(t *testing.T) {
	tempPaceDir(t)
	for _, status := range []int{http.StatusAccepted, http.StatusNoContent, http.StatusPartialContent} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(status)
		}))
		_, err := NewHTTPFetcher().Fetch(srv.URL + "/a")
		srv.Close()
		var ref *Refusal
		if !errors.As(err, &ref) {
			t.Fatalf("%d: no typed refusal at all: %v", status, err)
		}
		if got := refusalClass(ref.Status); got != "incomplete" {
			t.Errorf("%d recorded as %q, want incomplete — the host accepted the request", status, got)
		}
		if !strings.Contains(ref.Note, "NOT a refusal") {
			t.Errorf("%d: the note does not say what happened: %q", status, ref.Note)
		}
	}
	// AND A REAL REFUSAL IS STILL ONE. Widening this must not soften a 403 into "not yet".
	if got := refusalClass(http.StatusForbidden); got != "origin" && got != "unknown" {
		t.Errorf("a 403 is classed %q", got)
	}
}
