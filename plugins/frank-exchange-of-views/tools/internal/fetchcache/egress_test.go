package fetchcache

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A BLOCKED HOST IS NOT AN ANSWER, and the seat that sees the failure is the one that has to
// tell the difference.
//
// MEASURED 2026-08-23 (#592): openai.com returns 403 through the session proxy, and run B's
// deep-research-persistence question shipped "open rather than resolved" on that basis — a fact
// about this container's allowlist recorded as a fact about the world. "returned HTTP 403" was
// the whole of what the seat had to go on.
func TestAForbiddenResponseUnderAProxyNamesTheOtherReading(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://proxy.invalid:8080")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := NewHTTPFetcher().Fetch(srv.URL)
	if err == nil {
		t.Fatal("a 403 was accepted as a body")
	}
	for _, want := range []string{"403", "EGRESS PROXY", "UNREACHABLE FROM HERE"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the failure does not carry %q — the seat cannot tell a blocked host from a "+
				"refusing origin, and grades the question instead of the reach:\n%v", want, err)
		}
	}
	// IT STATES A POSSIBILITY, NOT A VERDICT. Nothing in the response separates a proxy refusal
	// from an origin refusal, so asserting the first would swap one unfounded certainty for
	// another — which is the defect, not the fix.
	if !strings.Contains(err.Error(), "MAY BE") {
		t.Errorf("the note asserts a proxy block it cannot know:\n%v", err)
	}
}

// WITH NO PROXY CONFIGURED THE QUESTION DOES NOT ARISE, and the sentence would be noise on
// every 403 a local run ever sees.
func TestNoProxyMeansNoEgressNote(t *testing.T) {
	for _, k := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		t.Setenv(k, "")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := NewHTTPFetcher().Fetch(srv.URL)
	if err == nil {
		t.Fatal("a 403 was accepted as a body")
	}
	if strings.Contains(err.Error(), "EGRESS PROXY") {
		t.Errorf("a direct-egress environment raised the proxy reading:\n%v", err)
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("the status itself went missing:\n%v", err)
	}
}

// AND ONLY THE STATUSES A PROXY ACTUALLY USES. A 404 is the origin's answer under any topology;
// attaching the note there would teach seats to discount every real miss as possible censorship,
// which is the same failure pointed the other way.
func TestAnOrdinaryMissCarriesNoEgressNote(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://proxy.invalid:8080")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := NewHTTPFetcher().Fetch(srv.URL)
	if err == nil {
		t.Fatal("a 404 was accepted as a body")
	}
	if strings.Contains(err.Error(), "EGRESS PROXY") {
		t.Errorf("a 404 raised the proxy reading — every honest miss would read as a possible block:\n%v", err)
	}
}
