package fetchcache

import (
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
