package fetchcache

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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
	if !r.allows("/abs/1706.03762") {
		t.Error("an explicitly allowed path was refused")
	}
	if r.allows("/find") {
		t.Error("a disallowed path was permitted")
	}
}

// THE WILDCARD RULES ARE MATCHED BY LONGEST MATCH, WITH Allow WINNING A TIE — the behaviour every
// major implementation shares, and the one that decides whether a broad Disallow with a narrow
// Allow carved out of it lets us through.
func TestLongestMatchWinsAndAllowBreaksTies(t *testing.T) {
	r := parseRobots("User-agent: *\nDisallow: /articles/\nAllow: /articles/open/\n")
	for path, want := range map[string]bool{
		"/articles/paywalled/1": false,
		"/articles/open/1":      true,
		"/elsewhere":            true,
	} {
		if got := r.allows(path); got != want {
			t.Errorf("allows(%q) = %v, want %v", path, got, want)
		}
	}
}

// A GROUP NAMING US WINS OUTRIGHT, and its silence is not the wildcard's speech — a host that
// gives this tool its own permissive group must not inherit the wildcard's refusals.
func TestAGroupNamingUsOverridesTheWildcard(t *testing.T) {
	r := parseRobots("User-agent: *\nDisallow: /\n\nUser-agent: feov-record\nDisallow: /private\nCrawl-delay: 2\n")
	if !r.allows("/articles/1") {
		t.Error("our own group was permitted everything but /private, and the wildcard's blanket refusal was applied anyway")
	}
	if r.allows("/private/x") {
		t.Error("our own group's refusal was not applied")
	}
	if r.interval() != 2*time.Second {
		t.Errorf("crawl delay = %v, want our group's 2s", r.interval())
	}
}

// SEVERAL AGENTS MAY SHARE ONE SET OF RULES, declared as consecutive User-agent lines. Resetting
// the target on each line instead of accumulating would silently apply the rules to only the last.
func TestConsecutiveAgentLinesShareTheRules(t *testing.T) {
	r := parseRobots("User-agent: SomeBot\nUser-agent: *\nDisallow: /secret\n")
	if r.allows("/secret/x") {
		t.Error("rules following a shared agent block were not applied to the wildcard")
	}
}

// AN ABSENT robots.txt PERMITS EVERYTHING — the standard's reading of a 404, and the common case.
func TestNoRulesPermits(t *testing.T) {
	var missing *robotsRules
	if !missing.allows("/anything") {
		t.Error("a nil rule set refused a path")
	}
	if !(&robotsRules{Missing: true}).allows("/anything") {
		t.Error("an absent robots.txt refused a path")
	}
}

// THE REFUSAL SAYS WHAT IT IS AND WHAT IS STILL POSSIBLE. A seat told only "disallowed" would
// record the source as unreachable, when the indexes that publish for machines may still answer.
func TestTheRobotsRefusalPointsSomewhere(t *testing.T) {
	e := &RobotsRefusal{URL: "https://ex.org/a", Rule: "/a"}
	for _, want := range []string{"robots.txt", "site operator's own instruction", "metadata", "oa"} {
		if !strings.Contains(e.Error(), want) {
			t.Errorf("refusal missing %q: %s", want, e.Error())
		}
	}
}

// $ ANCHORS AND * WILDCARDS are in wide use in real files, including arXiv's own.
func TestWildcardAndAnchorMatching(t *testing.T) {
	for _, tc := range []struct {
		pattern, path string
		want          bool
	}{
		{"/pdf/*v[0-9]", "/pdf/1234v[0-9]", true},
		{"/*.pdf$", "/a/b/c.pdf", true},
		{"/*.pdf$", "/a/b/c.pdf?x=1", false},
		{"/abs/", "/abs/1706.03762", true},
		{"/abs/", "/pdf/1706.03762", false},
	} {
		if got := matchRobotsPath(tc.pattern, tc.path); got != tc.want {
			t.Errorf("matchRobotsPath(%q,%q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
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

// A REDIRECT USED TO WALK STRAIGHT PAST robots.txt. The client followed 3xx internally, so the
// rules were consulted for the url a seat typed and for no hop after it — and since nearly every
// scholarly citation is a doi.org link that redirects to a publisher, the only host whose rules
// were ever checked was a resolver that publishes none. An operator's instruction evaded by a 302.
func TestRobotsIsHonouredOnARedirectTarget(t *testing.T) {
	tempPaceDir(t)
	var served int
	dest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /private/\n"))
			return
		}
		served++
		_, _ = w.Write([]byte("<html><body>" + strings.Repeat("the paper. ", 200) + "</body></html>"))
	}))
	defer dest.Close()
	resolver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r) // the resolver publishes nothing, exactly as doi.org does
			return
		}
		http.Redirect(w, r, dest.URL+"/private/article", http.StatusFound)
	}))
	defer resolver.Close()

	_, err := NewHTTPFetcher().Fetch(resolver.URL + "/10.1234/x")
	var rr *RobotsRefusal
	if !errors.As(err, &rr) {
		t.Fatalf("a redirect into a disallowed path was followed; err = %v", err)
	}
	if served != 0 {
		t.Errorf("the disallowed path was served %d time(s) — the refusal came too late", served)
	}
	// AND THE PERMITTED PATH ON THE SAME HOST STILL WORKS, so this refuses a path and not a host.
	if _, err := NewHTTPFetcher().Fetch(dest.URL + "/public/article"); err != nil {
		t.Errorf("an allowed path on the same host was refused: %v", err)
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

// A ROBOTS REFUSAL MUST STILL TRY THE OTHER ROUTES. The recovery chain keyed on *Refusal alone,
// so a disallowed page ended the fetch — while the refusal's own text tells the seat to ask `oa`
// whether a copy exists elsewhere. The tool named the route and declined to take it.
//
// IOP publishes `Disallow: /`, so every IOP paper was lost. Measured on one: `--via oa` returns a
// 633 KB PDF from arXiv. Being told we may not read the publisher's copy says nothing about the
// copy the author put in a repository.
func TestARobotsRefusalStillReachesTheOpenAccessRung(t *testing.T) {
	tempPaceDir(t)
	run := runtest.New(t, t.TempDir())
	pub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n")) // IOP's shape
			return
		}
		t.Error("the disallowed host was fetched after all")
	}))
	defer pub.Close()

	// The publisher is disallowed; a repository copy exists and is listed by the index.
	repo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("%PDF-1.7 the author's copy"))
	}))
	defer repo.Close()

	prev := Default
	Default = fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "openalex"):
			return &Response{Body: []byte(`{"id":"https://openalex.org/W1","locations":[{"pdf_url":"` +
				repo.URL + `/copy.pdf","is_oa":true}]}`)}, nil
		case strings.Contains(u, "ebi.ac.uk"):
			return &Response{Body: []byte(`{"resultList":{"result":[]}}`)}, nil
		case strings.Contains(u, "semanticscholar"):
			return &Response{Body: []byte(`{"paperId":"a","openAccessPdf":null}`)}, nil
		case strings.Contains(u, "doaj.org"):
			return &Response{Body: []byte(`{"total":0,"results":[]}`)}, nil
		}
		return NewHTTPFetcher().Fetch(u)
	})
	t.Cleanup(func() { Default = prev })

	entry, body, _, err := Resolve(run, pub.URL+"/article/10.1234/x", Default)
	if err != nil {
		t.Fatalf("a robots refusal ended the fetch instead of recovering: %v", err)
	}
	if !strings.HasPrefix(string(body), "%PDF") {
		t.Errorf("recovered %q, want the repository copy", string(body[:min(24, len(body))]))
	}
	// AND THE RECORD SAYS WHY THE PUBLISHER WAS NOT USED — without inventing an HTTP status the
	// origin never returned, because nothing was asked of it.
	if entry.RefusalClass != "robots" {
		t.Errorf("RefusalClass = %q, want robots", entry.RefusalClass)
	}
	if entry.HTTPStatus != 0 {
		t.Errorf("HTTPStatus = %d — nothing was asked of the origin, so it refused nothing", entry.HTTPStatus)
	}
}
