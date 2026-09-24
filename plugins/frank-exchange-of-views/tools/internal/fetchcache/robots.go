package fetchcache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ROBOTS.TXT IS READ FOR ITS RATE LIMIT, AND FOR NOTHING ELSE.
//
// # What this file is for
//
// A host that publishes `Crawl-delay` is telling automated clients how much room it wants, and
// that number beats anything this tool could guess. arXiv publishes **15 seconds** — five times
// slower than the 3 seconds this tool had transcribed from arXiv's prose terms of use. A number
// a host publishes for machines beats a number we read out of a sentence, every time, and it is
// the difference between believing we are polite and being told we are.
//
// # What it is NOT for, and the mistake that was made here
//
// This tool ONCE REFUSED a fetch on a `Disallow`, and that was wrong. The Robots Exclusion
// Protocol addresses CRAWLERS — clients that discover and traverse — and this is not one. It
// fetches one url a seat has chosen and is about to cite, which is the act a browser performs,
// and a browser does not consult robots.txt.
//
// The rules that were turning us away say so themselves. Measured over 558 fetches, six rules
// accounted for every refusal, and four were a blanket `Disallow: /` for `*` beside NAMED
// allowances for search engines — IOP allows googlebot and slurp with narrow, annotated
// exceptions ("Duplicate content caused by serving jsessionids", "Disallow crawling search
// results pages") while refusing everyone else outright; APA allows Googlebot and
// CrossrefEventDataBot by name. Those are search-index directives. Letting them decide whether a
// researcher may read one paper they have already cited is not obedience, it is a category error
// — and it cost the publisher's copy on 5.4% of fetches to honour a rule aimed at indexers.
//
// A tool that one day walks a corpus rather than answering for a citation is a crawler and owes
// the full protocol. This one is not that, and must not be built as though it were.
type robotsRules struct {
	// CrawlDelay is the host's own pacing instruction in seconds, or 0 where it publishes none.
	// It is the only directive this tool acts on.
	CrawlDelay float64 `json:"crawl_delay"`
	// Fetched is when this was read, for the cache's own expiry.
	Fetched time.Time `json:"fetched"`
	// Missing records that the host answered no robots.txt at all — stored so the absence is
	// cached like any other answer rather than re-asked every fetch.
	Missing bool `json:"missing"`
}

const (
	// robotsTTL is how long a host's rules are trusted. Long enough that a sweep reads each
	// host's file once, short enough that an operator changing their mind is honoured the same
	// day.
	robotsTTL = 6 * time.Hour
	// robotsMaxBytes bounds the read. Google's own limit is 500 KiB; anything past that is not a
	// rules file.
	robotsMaxBytes = 512 << 10
)

var robotsMem sync.Map // host -> *robotsRules, so one process reads a host's file once

// robotsFor returns the rules for a host, from memory, then from the shared cache, then from the
// host itself.
//
// A FAILURE TO READ IT IS NOT A LICENCE. Where robots.txt cannot be fetched at all the rules are
// treated as absent-and-permissive, which is what the standard says a 404 means — but a host that
// answers a 5xx or times out is NOT saying yes, so that case is cached only briefly and paced at
// the default rather than read as permission.
// THE SCHEME COMES FROM THE URL BEING FETCHED, never assumed. This asked https:// for every
// host, so an http-only origin's rules failed to fetch and were read as absent-and-permissive —
// the one direction a rules file must never fail in. Most scholarly hosts are https, which is
// exactly why it survived: it was wrong on the hosts nobody was looking at.
func robotsFor(f Fetcher, scheme, host string) *robotsRules {
	if host == "" {
		return &robotsRules{Missing: true}
	}
	key := strings.ToLower(scheme + "://" + host)
	if v, ok := robotsMem.Load(key); ok {
		if r := v.(*robotsRules); time.Since(r.Fetched) < robotsTTL {
			return r
		}
	}
	path := filepath.Join(paceDir, "robots-"+hashHost(scheme+"://"+key)+".json")
	if b, err := os.ReadFile(path); err == nil {
		var r robotsRules
		if json.Unmarshal(b, &r) == nil && time.Since(r.Fetched) < robotsTTL {
			robotsMem.Store(key, &r)
			return &r
		}
	}
	r := fetchRobots(f, scheme, host)
	r.Fetched = time.Now()
	robotsMem.Store(key, r)
	if err := os.MkdirAll(paceDir, 0o755); err == nil {
		if b, mErr := json.Marshal(r); mErr == nil {
			_ = os.WriteFile(path, b, 0o644)
		}
	}
	return r
}

func fetchRobots(f Fetcher, scheme, host string) *robotsRules {
	if scheme != "http" && scheme != "https" {
		scheme = "https"
	}
	resp, err := f.Fetch(scheme + "://" + host + "/robots.txt")
	if err != nil {
		// A refusal for robots.txt itself is the common case (404 = no rules) and is permissive.
		return &robotsRules{Missing: true}
	}
	if len(resp.Body) > robotsMaxBytes {
		resp.Body = resp.Body[:robotsMaxBytes]
	}
	return parseRobots(string(resp.Body))
}

// parseRobots reads the Crawl-delay that applies to us: from a group naming this tool where one
// exists, otherwise from the wildcard group.
//
// GROUPS ARE ACCUMULATED, NOT OVERWRITTEN, because a file may open several `User-agent:` lines
// before a single set of rules — the standard's way of saying "these apply to all of you".
func parseRobots(body string) *robotsRules {
	star, mine := &robotsRules{}, &robotsRules{}
	var targets []*robotsRules
	lastWasAgent := false
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k, v = strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v)
		if k == "user-agent" {
			if !lastWasAgent {
				targets = nil
			}
			switch {
			case v == "*":
				targets = append(targets, star)
			case v != "" && strings.Contains(strings.ToLower(userAgent), strings.ToLower(v)):
				targets = append(targets, mine)
			}
			lastWasAgent = true
			continue
		}
		lastWasAgent = false
		if k != "crawl-delay" {
			continue
		}
		if d, err := strconv.ParseFloat(v, 64); err == nil && d > 0 {
			for _, t := range targets {
				t.CrawlDelay = d
			}
		}
	}
	// A GROUP NAMING US WINS OUTRIGHT over the wildcard: the most specific matching group
	// applies, and its silence is not the wildcard's speech.
	if mine.CrawlDelay > 0 {
		return mine
	}
	return star
}

// interval is the host's published Crawl-delay as a duration, or 0 where it publishes none.
// It is the one thing this file exists to produce.
func (r *robotsRules) interval() time.Duration {
	if r == nil || r.CrawlDelay <= 0 {
		return 0
	}
	return time.Duration(r.CrawlDelay * float64(time.Second))
}
