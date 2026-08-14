package fetchcache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// The prod Fetcher's guardrails. These are caps, not a full SSRF sandbox: the stated risk
// posture (spec R2) is that a seat could already `curl` the same URL by hand, so the tool
// adds CACHING, not new reach. It refuses non-web schemes, bounds time, size, and redirect
// depth, and sends no credentials — enough that a hostile URL cannot hang the run, exhaust
// memory, or bounce the client through an unbounded redirect chain.
const (
	fetchTimeout  = 15 * time.Second
	maxFetchBytes = 5 << 20 // 5 MiB — a source document, not a download
	maxRedirects  = 5

	// userAgent identifies the tool to the sources it reads (see Fetch). Contact-bearing by
	// convention so a site operator can attribute and complain rather than only block.
	userAgent = "feov-record/" + fetchUAVersion + " (Special Circumstances research debate; +https://github.com/ctoforaday/special-circumstances)"
	// fetchUAVersion is kept separate from cli.Version to avoid an import cycle (cli imports
	// fetchcache). It moves when the fetch behaviour changes, not with every tool release.
	fetchUAVersion = "1.0"
)

type httpFetcher struct {
	client   *http.Client
	maxBytes int64
}

// NewHTTPFetcher builds the prod Fetcher: an http/https GET with a timeout, a redirect cap,
// and (in Fetch) a response-size cap.
func NewHTTPFetcher() Fetcher {
	return &httpFetcher{
		client: &http.Client{
			Timeout: fetchTimeout,
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			},
		},
		maxBytes: maxFetchBytes,
	}
}

func (h *httpFetcher) Fetch(rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetch: unparseable url %q: %w", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("fetch: refused scheme %q — only http and https are fetched", u.Scheme)
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	// IDENTIFY OURSELVES OR BE REFUSED. Go sends "Go-http-client/1.1" by default, which major
	// sources block outright. Measured on the 2026-08-04 smoke: blue tried to cite the Fundamental
	// Theorem of Arithmetic for "is 7 a prime number" and Wikipedia returned 403 — four cites lost,
	// the most obvious source for the question among them. Verified at the leaf: the default UA and
	// an empty UA both 403; a descriptive one returns 200. Neither gate could catch this — the
	// fuzzer serves from loopback with no UA policy, and the real-data check used example.com,
	// which does not care. A descriptive agent string is also simply what a well-behaved fetcher
	// owes the sites it reads: who we are and how to complain.
	req.Header.Set("User-Agent", userAgent)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch: %s returned HTTP %d", rawURL, resp.StatusCode)
	}
	// Read one byte past the cap so an over-size body is DETECTED, not silently truncated
	// into a citation.
	b, err := io.ReadAll(io.LimitReader(resp.Body, h.maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("fetch: reading %s: %w", rawURL, err)
	}
	if int64(len(b)) > h.maxBytes {
		return nil, fmt.Errorf("fetch: %s exceeds the %d-byte cap — cite a smaller source or a specific page", rawURL, h.maxBytes)
	}
	return b, nil
}
