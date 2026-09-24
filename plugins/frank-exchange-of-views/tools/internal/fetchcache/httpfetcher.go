package fetchcache

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// The prod Fetcher's guardrails. These are caps, not a full SSRF sandbox: the stated risk
// posture (spec R2) is that a seat could already `curl` the same URL by hand, so the tool
// adds CACHING, not new reach. It refuses non-web schemes, bounds time, size, and redirect
// depth, and sends no credentials — enough that a hostile URL cannot hang the run, exhaust
// memory, or bounce the client through an unbounded redirect chain.
const (
	// fetchTimeout bounds how long ONE HOP may take to start answering — not how long the whole
	// fetch may take.
	//
	// OUR OWN PACING MUST NOT COUNT AGAINST THE NETWORK'S PATIENCE. http.Client.Timeout covers
	// everything including redirects, so once each hop began waiting its turn a paced chain blew
	// the budget before it blew the redirect cap: an endless-redirect fixture failed with a
	// deadline instead of the cap it was written for. A hop that sleeps 15 seconds by our choice
	// has not been slow, and treating it as slow would make politeness look like a broken host.
	fetchTimeout  = 15 * time.Second
	maxFetchBytes = 5 << 20 // 5 MiB — a source document, not a download
	// maxRedirects is the real bound on a paced fetch. Five is generous for the web and
	// pathological for a citation: a doi.org link reaches its publisher in one or two hops, and
	// the deadline in Fetch is sized from this number rather than guessed at.
	maxRedirects = 5
	// maxMetaRefresh bounds the OTHER kind of redirect — the one a page expresses in its markup
	// instead of its status line. It is deliberately smaller than maxRedirects: a header redirect
	// chain is routine, whereas a page that bounces twice through HTML is already unusual.
	maxMetaRefresh = 3

	// userAgent identifies the tool to the sources it reads (see Fetch): who we are, what we are
	// doing, where to read about it, and how to reach a human.
	//
	// THE mailto IS NOT DECORATION — IT IS THE POLITE POOL. Measured against api.crossref.org on
	// 2026-09-22: this agent string without a mailto is answered `x-api-pool: public-single` with
	// `x-rate-limit-limit: 5`, and with one it is `polite-single` at `10`. The public pool is the
	// one that gets shed first under load, and it is the one where an operator who dislikes our
	// traffic has no way to tell us so and must simply block us. A contactable client is both
	// better mannered and better served.
	//
	// Crossref reads it from the User-Agent OR a mailto= query parameter; the header covers every
	// host at once, so it is the one that belongs here. OpenAlex, measured the same day, returns
	// identical headers with and without — its polite pool is gone, replaced by metering — so the
	// mailto it is sent buys nothing measurable and is sent anyway, because identifying yourself
	// is not a trade.
	// WHAT IS IN THIS STRING IS LOAD-BEARING, AND SO IS WHAT IS ABSENT FROM IT.
	//
	// Anubis — the proof-of-work gate in front of PubMed Central, SciPost and DOAB — SCORES the
	// user agent against its shipped policy and challenges anything over zero. Read from its
	// source at v1.28.0-pre2: `Mozilla` or `Opera` in the string adds +10, which is a challenge;
	// a commented-out catch-all that operators do enable hard-denies `bot` or `crawler` at
	// "difficulty: 16 # impossible". This agent matches none of them and passes at weight zero.
	//
	// So making ourselves LOOK like a browser is the one change guaranteed to get us gated, and
	// the honest string is also the one that works. Cloudflare's Verified Bots policy says the
	// same from the other end: it auto-rejects generic agents (`Go-http-client`, `python-requests`)
	// and asks for "honest self-identification" plus a contact.
	//
	// YOU MUST NOT add `Mozilla`, `bot`, `crawler`, or a browser version triple to this string.
	// TestTheAgentDoesNotImpersonateOrTripAGate holds it.
	userAgent = "feov-record/" + fetchUAVersion + " (Special Circumstances research debate; " +
		"+https://github.com/ctoforaday/special-circumstances; mailto:" + ContactEmail + ")"
	// ContactEmail is the one address this tool gives out, to every source and every API that
	// asks. It was three copies of a literal before, one per backend, which is how a contact
	// address silently stops matching the one a site operator would actually reach.
	ContactEmail = "feov@ctoforaday.com"
	// fetchUAVersion is kept separate from cli.Version to avoid an import cycle (cli imports
	// fetchcache). It moves when the FETCH BEHAVIOUR changes, not with every tool release: 1.1
	// follows redirects expressed in markup and paces itself per host, both of which a site
	// operator reading their logs would notice.
	fetchUAVersion = "1.1"
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
			// NO OVERALL DEADLINE HERE. Liveness is enforced per hop by the transport below, and
			// the whole-fetch bound is a context in Fetch sized to include the pacing this client
			// deliberately does.
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				ResponseHeaderTimeout: fetchTimeout,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: time.Second,
			},
			// WE FOLLOW REDIRECTS OURSELVES. Returning ErrUseLastResponse hands each 3xx back
			// as an ordinary response, which is the only way to treat a hop as what it is: a
			// fetch of a different url, on a different host, that must clear the same gates as
			// the first one.
			//
			// Doing it inside CheckRedirect could pace a hop but nothing else. robots.txt was
			// consulted for the url a seat typed and for no hop after it, so a redirect into a
			// path the host disallows went through — the operator's instruction evaded by a
			// 302. Sleeping in the callback also fought the client's own timeout, which counts
			// our waiting as the network being slow.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		maxBytes: maxFetchBytes,
	}
}

// Fetch GETs a url, following redirects — INCLUDING THE ONES EXPRESSED IN MARKUP.
//
// A `<meta http-equiv="refresh">` is a redirect that Go's client does not follow, because it is
// not a status line. Measured over the top-cited works of 26 fields, one url in five landed on
// one: Elsevier bounces a cookie-less client through `<title>Redirecting</title>` with eleven
// characters of visible text, and the article sits one hop further on. The tool reported those as
// "something stands between this container and the source", which was true and useless — the
// source had said where to go and nothing followed it.
//
// IT IS FOLLOWED ONLY FROM A PAGE WITH NO PROSE OF ITS OWN. A document that happens to carry a
// refresh — a live-updating dashboard, an article with a stale meta tag — must be returned, not
// abandoned for whatever it points at. The prose floor that decides is the same one ShellReason
// uses, so the two rules cannot drift apart: below it the page is a bouncer, above it the page is
// the answer.
func (h *httpFetcher) Fetch(rawURL string) (*Response, error) {
	seen := map[string]bool{rawURL: true}
	cur := rawURL
	var reserved bool
	var policy string
	for hop := 0; ; hop++ {
		out, final, err := h.fetchOnce(cur)
		if err != nil {
			return nil, err
		}
		// A RESERVATION SURVIVES THE HOP THAT DECLARED IT. Keep the first one seen, because the
		// page that states it is usually the one being redirected away from.
		if out.TDMReserved {
			reserved, policy = true, out.TDMPolicy
		}
		out.TDMReserved, out.TDMPolicy = reserved, policy
		next := metaRefreshTarget(out.ContentType, out.Body, final)
		// THE LAST HOP RETURNS WHAT IT HAS rather than failing: a bouncer is still a fact about
		// the source, and ShellReason will say so. Refusing here would turn a describable page
		// into no answer at all.
		if next == "" || hop >= maxMetaRefresh || seen[next] {
			return out, nil
		}
		seen[next] = true
		cur = next
	}
}

// isRedirect names the statuses that carry a Location worth following.
func isRedirect(code int) bool {
	switch code {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

func isOverloadStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

func (h *httpFetcher) fetchOnce(rawURL string) (*Response, *url.URL, error) {
	return h.followRedirects(rawURL, false)
}

// robotsClient fetches a rules file WITHOUT consulting rules, for every hop of that fetch.
//
// EXEMPTING THE PATH WAS NOT ENOUGH. The gate skipped `/robots.txt`, but a host that REDIRECTS
// its rules file sends the chain somewhere whose path is not `/robots.txt` — which triggers a
// robots check, which fetches the rules file, which redirects. That recursed until the stack
// overflowed, and it took a fixture that redirects everything to find it. The standard's
// exemption is for the act of reading the rules, not for one url that happens to spell them.
type robotsClient struct{ h *httpFetcher }

func (r robotsClient) Fetch(u string) (*Response, error) {
	out, _, err := r.h.followRedirects(u, true)
	return out, err
}

// followRedirects walks a redirect chain by hand, so every hop is a full fetch: its host paced,
// its path checked against that host's robots.txt, its status inspected.
func (h *httpFetcher) followRedirects(rawURL string, skipRobots bool) (*Response, *url.URL, error) {
	cur := rawURL
	// A DOI URL IS AN IDENTIFIER, AND CROSSREF HOLDS WHERE IT POINTS. Asking it replaces a hop
	// through a resolver that publishes no rate guidance — and therefore sits at this tool's
	// fifteen-second kind default — with a 200ms lookup against an API built to be asked. Where
	// Crossref does not answer, cur is unchanged and the resolver's own redirect is followed, so
	// this is a shortcut and never a dependency.
	if target := RegisteredTarget(robotsClient{h}, cur); target != "" {
		cur = target
	}
	seen := map[string]bool{rawURL: true, cur: true}
	for hop := 0; ; hop++ {
		resp, final, loc, err := h.fetchOnceRetry(cur, false, skipRobots)
		if err != nil || loc == "" {
			return resp, final, err
		}
		if hop >= maxRedirects {
			return nil, nil, fmt.Errorf("fetch: stopped after %d redirects, last at %s", maxRedirects, cur)
		}
		next, rerr := final.Parse(loc)
		if rerr != nil {
			return nil, nil, fmt.Errorf("fetch: %s redirected to an unparseable location %q", cur, loc)
		}
		if next.Scheme != "http" && next.Scheme != "https" {
			return nil, nil, fmt.Errorf("fetch: %s redirected out of the web, to %s", cur, next.Scheme)
		}
		if seen[next.String()] {
			return nil, nil, fmt.Errorf("fetch: %s redirects in a loop, back to %s", rawURL, next)
		}
		seen[next.String()] = true
		cur = next.String()
	}
}

func (h *httpFetcher) fetchOnceRetry(rawURL string, retried, skipRobots bool) (out *Response, final *url.URL, location string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fetch: unparseable url %q: %w", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, nil, "", fmt.Errorf("fetch: refused scheme %q — only http and https are fetched", u.Scheme)
	}
	// THE CHAIN LENGTH IS THE BOUND; THIS DEADLINE IS A BACKSTOP DERIVED FROM IT.
	//
	// What actually limits a paced fetch is maxRedirects — five hops, each waiting at most one
	// jittered floor. The deadline is computed from that cap rather than chosen as a time budget,
	// so the two cannot disagree: raise the cap and this follows. It exists only to stop a
	// connection that never returns from holding a seat open, which is the job Client.Timeout did
	// before our own waiting made that number mean two different things.
	var decodedAs string // the content coding this function decoded, for the read error below
	ctx, cancel := context.WithTimeout(context.Background(),
		fetchTimeout+time.Duration(maxRedirects+1)*2*defaultHostInterval)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fetch: %w", err)
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
	// THE OPERATOR'S OWN INSTRUCTION, READ BEFORE WE ASK FOR ANYTHING ELSE.
	//
	// Reading robots.txt is itself a fetch, so it is exempted by path — otherwise checking the
	// rules would need the rules. The exemption is the standard's own: robots.txt is never
	// governed by robots.txt.
	if !skipRobots {
		rules := robotsFor(robotsClient{h}, u.Scheme, u.Host)
		if rule := robotsBlocks(rules, u); rule != "" {
			return nil, nil, "", &RobotsRefusal{URL: rawURL, Rule: rule}
		}
	}
	// WAIT OUR TURN FOR THIS HOST. The floor is enforced here, at the only place a request
	// leaves, so no caller can forget it and no new backend has to remember.
	//
	// EXCEPT FOR robots.txt, AND THE COST OF NOT EXEMPTING IT WAS MEASURED. A rules file fetch
	// claimed the host's slot like any other request, so first contact with a host went: read
	// robots.txt (no wait, nothing queued yet), arm the floor, then wait a full jittered
	// interval — 15 to 30 seconds for a host that publishes no rate — before the document this
	// was all for. Every new host paid that, and the open-access chain visits many.
	//
	// Measured across three sweeps as candidate hosts multiplied: median 38s per work, then 44s,
	// then 88s, with the 180s timeout rate going 4% to 12% to 19%. Most of that is this.
	//
	// It is also what the standard expects. robots.txt is outside the crawl budget it defines —
	// a rules file is small, fetched once per TTL, and delaying the first real request by half a
	// minute to be polite about the file that grants permission is politeness spent on nothing.
	// Two immediate requests to a cold host is not a flood.
	if !skipRobots {
		w, ok := reserveSlot(u.Host, 0)
		if !ok {
			return nil, nil, "", fmt.Errorf("fetch: %s is saturated — this machine already has more than %v "+
				"of queued requests waiting for that host, so this one is refused rather than added to the "+
				"queue. It is a fact about how much this box is asking of one origin, not about the source",
				u.Host, maxPaceWait)
		}
		if w > 0 {
			time.Sleep(w)
		}
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	// A 429 IS THE HOST TELLING US THE PACE WAS WRONG, and it is the one refusal worth obeying
	// rather than merely reporting. Honour its own Retry-After once, and push the whole host's
	// queue out by it so every later request slows too — backing off only the refused request
	// would keep the pressure that caused it. A second one is returned: at that point the answer
	// is not "wait a little longer", it is that this host does not want this traffic now.
	//
	// A 5xx COUNTS AS THE SAME MESSAGE. A service answering 502, 503 or 504 is overloaded or
	// broken, and either way continuing at the pace that met it is the wrong response — Europe
	// PMC falls over this way under load, with no Retry-After to read. Treating a 5xx as an
	// ordinary refusal means the next request goes out at exactly the rate that just failed.
	if isOverloadStatus(resp.StatusCode) && !retried {
		wait := retryAfter(resp.Header)
		if wait <= 0 {
			// THIS HOST'S FLOOR, not the global default. A host with a published Crawl-delay or a
			// learned floor has already told us what "a while" means for it, and waiting the
			// generic default instead would be slower than it asked in one direction and faster
			// in the other.
			wait = intervalFor(u.Host)
		}
		backoffHost(u.Host, wait)
		if wait <= retryAfterCap {
			time.Sleep(wait)
			return h.fetchOnceRetry(rawURL, true, skipRobots)
		}
	}
	// A REDIRECT IS NOT A REFUSAL. The client no longer follows them, so a 3xx arrives here with
	// its Location and is handed back to followRedirects, which fetches it as a fresh url —
	// paced on its own host, and checked against that host's robots.txt. Following inside the
	// client meant neither happened for any hop after the first.
	if isRedirect(resp.StatusCode) {
		if loc := resp.Header.Get("Location"); loc != "" {
			return nil, u, loc, nil
		}
		return nil, nil, "", &Refusal{URL: rawURL, Status: resp.StatusCode,
			Note: " — the host sent a redirect with no Location to follow"}
	}
	if resp.StatusCode != http.StatusOK {
		// A REFUSAL STILL HAS A BODY, AND IT OFTEN SAYS WHAT IT IS. Challenge detection ran only
		// on a 200 until now, so a wall that refuses with a status — Cloudflare, Anubis — reached
		// a seat as the bare number. Measured 2026-09-22: DOAB's book pages answer 403 with
		// `<title>Making sure you're not a bot!</title>` and Anubis cookies, while DOAB's own API
		// is open and hands out the DOI and an OAPEN handle for the same book. "403" sends a seat
		// away from a book it could have read; "a proof-of-work wall, and this index publishes an
		// API" sends it one rung further.
		note := egressNote(resp.StatusCode)
		if lim := int64(1 << 20); resp.ContentLength <= lim {
			if b, rerr := io.ReadAll(io.LimitReader(resp.Body, lim)); rerr == nil {
				if why := ShellReason(resp.Header.Get("Content-Type"), b); why != "" {
					note = " — " + why + note
				}
			}
		}
		return nil, nil, "", &Refusal{URL: rawURL, Status: resp.StatusCode, Note: note}
	}
	// THE HEADERS ARE READ HERE OR NEVER. Content-Type is the source's own statement of what it
	// just sent, available for exactly the length of this function and previously discarded at
	// the end of it. Everything downstream then had to sniff magic bytes or read an extension
	// off a URL that may not have one.
	// AN ENCODING WE DID NOT ASK FOR, AND THE ANSWER DEPENDS ON WHETHER WE CAN READ IT.
	//
	// net/http strips Content-Encoding when it decoded the body itself, so a header still present
	// here names a coding the transport left alone. Servers are not supposed to send one that was
	// never offered, and they do: a CDN normalises to brotli, a proxy re-encodes, an origin
	// answers `deflate` because it always has.
	//
	// DECODE WHAT WE HAVE A READER FOR. Refusing a body we could perfectly well read would throw
	// away a document over a header — the server being non-compliant is not the document's fault,
	// and gzip and deflate are both in the standard library. Only a coding with no reader here is
	// refused, and the refusal names it, so what is missing is a decoder and everyone can see
	// which one.
	//
	// Bounded like every other read on this path: a decompression bomb is a small response that
	// becomes an unbounded one, so the decoded stream gets the same cap and the same
	// one-byte-past detection as the wire body.
	if enc := strings.TrimSpace(strings.ToLower(resp.Header.Get("Content-Encoding"))); enc != "" && enc != "identity" {
		decoded, derr := decodeCoding(enc, resp.Body, h.maxBytes)
		if derr != nil {
			return nil, nil, "", fmt.Errorf("fetch: %s answered with Content-Encoding %q: %w", rawURL, enc, derr)
		}
		resp.Body = io.NopCloser(decoded)
		decodedAs = enc
	}
	out = &Response{
		ContentType: resp.Header.Get("Content-Type"),
		Disposition: resp.Header.Get("Content-Disposition"),
		LinkHeader:  strings.Join(resp.Header.Values("Link"), ", "),
	}
	// Read one byte past the cap so an over-size body is DETECTED, not silently truncated
	// into a citation.
	b, rerr := io.ReadAll(io.LimitReader(resp.Body, h.maxBytes+1))
	if rerr != nil {
		// A STREAM THAT FAILS HALFWAY NAMES THE CODING TOO. gzip and deflate decode lazily, so a
		// truncated or mislabelled body surfaces here rather than at the constructor — and
		// "corrupt input before offset 3" with no coding named sends the reader looking at the
		// network instead of at the header that lied.
		if decodedAs != "" {
			return nil, nil, "", fmt.Errorf("fetch: reading %s, which declared Content-Encoding %q: %w", rawURL, decodedAs, rerr)
		}
		return nil, nil, "", fmt.Errorf("fetch: reading %s: %w", rawURL, rerr)
	}
	if int64(len(b)) > h.maxBytes {
		return nil, nil, "", fmt.Errorf("fetch: %s exceeds the %d-byte cap — cite a smaller source or a specific page", rawURL, h.maxBytes)
	}
	out.Body = b
	out.TDMReserved, out.TDMPolicy = TDMReservation(out.ContentType, b)
	// resp.Request.URL is the url AFTER any header redirects, which is what a relative refresh
	// target must resolve against — resolving against the ORIGINAL would build a url on the
	// wrong host the moment a doi.org link is involved, which is most of them.
	return out, u, "", nil
}

// AN EGRESS BLOCK IS NOT AN EPISTEMIC RESULT, and telling them apart is the seat's job the
// moment a fetch fails.
//
// MEASURED 2026-08-23 (#592): openai.com returns 403 through the session proxy. Run B's
// deep-research-persistence question came back "open rather than resolved" on the strength of
// that 403 — a fact about this container's allowlist, recorded as a fact about the world. The
// two are indistinguishable in "returned HTTP 403", which is all the seat saw.
//
// THE NOTE STATES A POSSIBILITY, NEVER A VERDICT. A 403 from an origin that dislikes our user
// agent and a 403 from a proxy refusing the host produce the same status line; nothing in the
// response reliably separates them, and a message asserting "blocked at the proxy" would trade
// one unfounded certainty for another. So it names both readings and the check that decides,
// and it fires only where a proxy is actually configured — in a direct-egress environment the
// question does not arise and the extra sentence would be noise.
// refusalClass answers WHO refused, and says `unknown` whenever it cannot tell. It fires on the
// same statuses egressNote explains, and for the same reason: with a proxy configured, a refusal
// is genuinely ambiguous between the container and the origin, and asserting either would trade
// an honest unknown for an unfounded certainty.
func refusalClass(status int) string {
	switch status {
	case http.StatusForbidden, http.StatusMethodNotAllowed, http.StatusProxyAuthRequired:
		if proxyEnv() != "" {
			return "unknown"
		}
		return "origin"
	}
	return "origin"
}

func egressNote(status int) string {
	switch status {
	case http.StatusForbidden, http.StatusMethodNotAllowed, http.StatusProxyAuthRequired:
	default:
		return ""
	}
	if proxyEnv() == "" {
		return ""
	}
	return " — THIS MAY BE THE EGRESS PROXY REFUSING THE HOST rather than the origin refusing us, " +
		"and the two are different findings: a proxy block is a fact about this container's allowlist, " +
		"an origin refusal is a fact about the source. Do not record an unreached source as evidence of " +
		"absence. Check which it is before grading (the proxy's own status endpoint, or the same URL from " +
		"a host outside this environment), and where you cannot, say the source was UNREACHABLE FROM HERE " +
		"rather than that the question is unresolved."
}

// proxyEnv reports the configured egress proxy, if any. Both spellings, because the
// environment sets the upper-case form and Go's own ProxyFromEnvironment reads either.
func proxyEnv() string {
	for _, k := range []string{"HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// metaRefreshRe matches `<meta http-equiv="refresh" content="2; url=...">`.
//
// THE TWO QUOTE STYLES ARE SPELLED OUT because RE2 has no backreference, and a pattern that
// accepts either quote and closes on either one gets this wrong on the page it was written for:
// Elsevier sends content="2; url='/retrieve/…'", so a `["']([^"']*)["']` capture stops at the
// inner apostrophe and yields `2; url=` with the target discarded. It matched, so it looked like
// a parse rather than a miss.
var metaRefreshRe = regexp.MustCompile(`(?is)<meta[^>]+http-equiv\s*=\s*["']?refresh["']?[^>]*?content\s*=\s*(?:"([^"]*)"|'([^']*)')`)

// metaRefreshURL pulls the target out of a refresh directive's content value: `2; url=/next`.
var metaRefreshURL = regexp.MustCompile(`(?is)^\s*[\d.]*\s*;?\s*url\s*=\s*['"]?([^'"\s]+)`)

// metaRefreshTarget returns the absolute url a bouncer page points at, or "" where this response
// is not one.
//
// THREE CONDITIONS, AND EACH REMOVES A WAY TO GO WRONG. It must be HTML, or the pattern is
// matching something that merely looks like markup. It must carry no prose of its own, or it is a
// document that happens to refresh and following it would discard the answer. And the target must
// resolve to http or https, so a `data:` or `javascript:` refresh cannot move the fetch somewhere
// the scheme check at the top of fetchOnce exists to forbid.
func metaRefreshTarget(contentType string, body []byte, base *url.URL) string {
	if !strings.Contains(strings.ToLower(contentType), "html") || base == nil {
		return ""
	}
	if len(visibleText(body)) >= shellProseFloor {
		return ""
	}
	m := metaRefreshRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	content := m[1]
	if len(content) == 0 {
		content = m[2] // the single-quoted alternative
	}
	t := metaRefreshURL.FindSubmatch(content)
	if t == nil {
		return ""
	}
	ref, err := url.Parse(strings.TrimSpace(html.UnescapeString(string(t[1]))))
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	return abs.String()
}

// decodeCoding returns a reader over a content coding the transport did not decode, or an error
// naming the coding when there is no reader for it.
//
// gzip is here even though net/http normally handles it: it does so only for the request it added
// the header to, and "normally" is not a guarantee to build a stored citation on.
//
// `deflate` is both things it is called. RFC 9110 defines it as zlib-wrapped, and a long tail of
// servers send raw DEFLATE instead, so the zlib reader is tried first and a raw reader second —
// the ambiguity is the wire's, and guessing one of the two would drop half the responses.
func decodeCoding(enc string, body io.Reader, cap int64) (io.Reader, error) {
	limited := io.LimitReader(body, cap+1)
	switch enc {
	case "gzip", "x-gzip":
		zr, err := gzip.NewReader(limited)
		if err != nil {
			return nil, fmt.Errorf("the body is not a valid gzip stream: %w", err)
		}
		return zr, nil
	case "deflate":
		raw, err := io.ReadAll(limited)
		if err != nil {
			return nil, err
		}
		if zr, zerr := zlib.NewReader(bytes.NewReader(raw)); zerr == nil {
			return zr, nil
		}
		return flate.NewReader(bytes.NewReader(raw)), nil
	}
	// NO READER HERE, AND THE REFUSAL NAMES THE CODING so the gap is countable rather than a
	// mystery. brotli and zstd both need a dependency; whether that is worth taking is a question
	// about how often this fires, which this message is what makes measurable.
	return nil, fmt.Errorf("nothing here decodes it, and this client never offered it — "+
		"the body is not the document and must not be read as one (gzip and deflate are decoded; %q is not)", enc)
}
