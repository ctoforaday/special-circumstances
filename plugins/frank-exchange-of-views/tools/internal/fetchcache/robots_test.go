package fetchcache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// arXiv's REAL robots.txt, trimmed to the groups that decide. The Crawl-delay is the point: this
// tool paced arXiv at 3 seconds, transcribed from the prose terms of use, while arXiv publishes
// 15 for machines. A number addressed to us beats a number we read in a sentence.
const arxivRobots = `User-agent: *
Disallow: /find
Disallow: /pdf/*v[0-9][0-9]
Allow: /abs/
Crawl-delay: 15

User-agent: Twitterbot
Disallow:
`

func TestTheHostsOwnCrawlDelayIsRead(t *testing.T) {
	r := parseRobots(arxivRobots)
	if got := r.interval(); got != 15*time.Second {
		t.Errorf("crawl delay = %v, want 15s — the host's published number", got)
	}
}

// A Disallow IS READ PAST, NOT OBEYED, and that is the decision this test exists to hold.
//
// The Robots Exclusion Protocol addresses CRAWLERS — clients that discover and traverse. This
// tool fetches one url a seat has chosen and is about to cite, which is what a browser does, and
// a browser does not consult robots.txt. Measured over 558 fetches, obeying cost the publisher's
// copy on 5.4% of them, to rules aimed at search indexers: four of the six were a blanket
// `Disallow: /` for `*` sitting beside named allowances for googlebot.
//
// The file is still read. The rate limit in it is a host telling us how much room it wants, and
// that is worth having.
func TestADisallowDoesNotStopAFetch(t *testing.T) {
	blanket := parseRobots("User-agent: *\nDisallow: /\nCrawl-delay: 20\n")
	if got := blanket.interval(); got != 20*time.Second {
		t.Errorf("the rate was lost with the rules: %v", got)
	}
	// The parsed rules carry no path vocabulary at all, which is what makes obeying one
	// impossible to reintroduce by accident rather than merely discouraged.
	if fmt.Sprintf("%+v", *blanket) != fmt.Sprintf("%+v", robotsRules{CrawlDelay: 20}) {
		t.Errorf("robotsRules carries more than the rate: %+v", *blanket)
	}
}

// SEVERAL AGENTS MAY SHARE ONE SET OF RULES, declared as consecutive User-agent lines. Resetting
// the target on each line instead of accumulating would silently apply the rules to only the last.
func TestConsecutiveAgentLinesShareTheRules(t *testing.T) {
	r := parseRobots("User-agent: SomeBot\nUser-agent: *\nCrawl-delay: 9\n")
	if got := r.interval(); got != 9*time.Second {
		t.Errorf("a rate following a shared agent block did not reach the wildcard group: %v", got)
	}
}

// THE AGENT STRING IS A GATE INPUT, NOT DECORATION, and the tokens it must NOT contain are as
// load-bearing as the ones it must.
//
// Anubis scores the user agent against a policy shipped in its source: `Mozilla` or `Opera` adds
// +10, which earns a proof-of-work challenge, and a catch-all that operators do enable denies
// `bot` or `crawler` outright at "difficulty: 16 # impossible". This agent passes at weight zero.
// Cloudflare's Verified Bots policy agrees from the other side: it auto-rejects generic agents and
// asks for honest self-identification with a contact.
//
// The failure this prevents is a well-meant one — someone making the client "look more like a
// browser" to get past a wall, which is both the impersonation this tool will not do and the
// change that would guarantee the wall.
func TestTheAgentDoesNotImpersonateOrTripAGate(t *testing.T) {
	ua := strings.ToLower(userAgent)
	for _, banned := range []string{"mozilla", "opera", "applewebkit", "chrome/", "safari/", "gecko"} {
		if strings.Contains(ua, banned) {
			t.Errorf("the user agent contains %q — that is browser impersonation, and Anubis adds weight for it, "+
				"so it fails on both counts: %s", banned, userAgent)
		}
	}
	for _, banned := range []string{"bot", "crawler", "spider", "scraper"} {
		if strings.Contains(ua, banned) {
			t.Errorf("the user agent contains %q — Anubis's catch-all denies it at impossible difficulty: %s", banned, userAgent)
		}
	}
	// AND IT MUST STILL SAY WHO WE ARE AND HOW TO COMPLAIN, which is the other half of the bar.
	for _, required := range []string{"feov-record/", "mailto:", "github.com/ctoforaday"} {
		if !strings.Contains(ua, strings.ToLower(required)) {
			t.Errorf("the user agent no longer carries %q, which is what makes it honest self-identification: %s", required, userAgent)
		}
	}
}

// THE PUBLISHED DELAY MUST REACH THE FLOOR, end to end, through a real fetch.
//
// A unit test asserted this and passed while it was broken for every real host, because it
// stored its fixture under a key robotsFor never writes. The cache is keyed `scheme://host`; the
// lookup used the bare host; every entry missed; the override was dead code. arXiv publishes 15
// seconds and was paced at the internal table's 3, with nothing failing and nothing logged.
//
// So this drives the fetcher, lets it read a real rules file, and then asks the pacer what it
// learned — the only version of this test that could have caught it.
func TestAPublishedCrawlDelayReachesThePacer(t *testing.T) {
	tempPaceDir(t)
	pacedLoopback(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nCrawl-delay: 30\n"))
			return
		}
		_, _ = w.Write([]byte("<html><body>" + strings.Repeat("the paper. ", 200) + "</body></html>"))
	}))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	// Before anything is read, the floor is whatever we chose.
	if iv := intervalFor(host); iv != defaultHostInterval {
		t.Fatalf("pre-read interval = %v, want the default %v", iv, defaultHostInterval)
	}
	if _, err := NewHTTPFetcher().Fetch(srv.URL + "/article"); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	// After: the host's own number, which is larger than ours and therefore wins.
	if iv := intervalFor(host); iv != 30*time.Second {
		t.Errorf("after reading a published Crawl-delay of 30s the floor is %v — the host's own "+
			"number never reached the pacer", iv)
	}
}

// THE SHORTCUT IS TAKEN ON THE REAL FETCH PATH, not merely available to it. A test of
// RegisteredTarget is a test of the function; deleting the call in followRedirects left the whole
// suite green, which is the shape that lets a resolved path quietly stop being resolved.
//
// What this asserts is behavioural and not incidental: the resolver is NEVER ASKED. That is the
// whole point — doi.org publishes no rate guidance, so it sits at this tool's fifteen-second kind
// default, and the sweep that measured this spent 41 projected hours there.
func TestTheFetchPathTakesTheRegisteredTargetAndNeverAsksTheResolver(t *testing.T) {
	tempPaceDir(t)
	prevPace := paceLoopback
	paceLoopback = false // the fixture hosts are loopback; pacing them would only slow the test
	t.Cleanup(func() { paceLoopback = prevPace })

	var resolverHits int
	publisher := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("<html><body>" + strings.Repeat("the paper. ", 200) + "</body></html>"))
	}))
	defer publisher.Close()
	crossref := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"message":{"DOI":"10.1234/x","resource":{"primary":{"URL":"` +
			publisher.URL + `/article/1"}}}}`))
	}))
	defer crossref.Close()
	resolver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		resolverHits++
		http.Redirect(w, r, publisher.URL+"/article/1", http.StatusFound)
	}))
	defer resolver.Close()

	ru, err := url.Parse(resolver.URL)
	if err != nil {
		t.Fatal(err)
	}
	doiResolverHosts[ru.Hostname()+":"+ru.Port()] = true
	prevCR := crossrefWorks
	crossrefWorks = crossref.URL + "/works/"
	t.Cleanup(func() {
		crossrefWorks = prevCR
		delete(doiResolverHosts, ru.Hostname()+":"+ru.Port())
	})

	resp, err := NewHTTPFetcher().Fetch(resolver.URL + "/10.1234/x")
	if err != nil {
		t.Fatalf("the doi fetch failed: %v", err)
	}
	if !strings.Contains(string(resp.Body), "the paper.") {
		t.Errorf("the publisher page did not come back: %.120s", resp.Body)
	}
	if resolverHits != 0 {
		t.Errorf("the resolver was asked %d time(s) — the registered target should have replaced it", resolverHits)
	}
}
