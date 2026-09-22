package fetchcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

// PACING IS SHARED BETWEEN PROCESSES, BECAUSE THE FETCHES ARE.
//
// The first version of this file held its state in a map and paced one process. That is the wrong
// shape for how this tool is actually used: a run dispatches several seats, each seat reaches the
// web by shelling out `feov-record fetch`, and those are separate processes running at the same
// time in different shells. An in-process floor would have let twenty concurrent seats hit one
// origin simultaneously while every one of them believed it was pacing itself — the appearance of
// politeness, which is worse than none, because nobody looks again.
//
// So the floor lives in a file, taken under an advisory lock that every process contends for.
// Twenty parallel fetches of the same host queue behind one another; twenty of different hosts do
// not wait at all.
//
// # Why the state is machine-scoped rather than run-scoped
//
// arXiv states the reason in its own terms of use: the limit applies "to all of the machines under
// your control as a whole". A limit that reset per run would be no limit at all to an origin
// watching one address, and two concurrent runs on this box are, to that origin, one client.
//
// # Why the slot is claimed under the lock and slept OUTSIDE it
//
// Holding the lock for the duration of the wait would serialise every process on the lock rather
// than on the host, and a peer's bounded acquire would time out and proceed unpaced — the floor
// failing open under exactly the load it exists for. Claiming a slot is a short read-modify-write;
// waiting for it is the caller's own business.
const (
	// defaultHostInterval paces a host that publishes no limit of its own.
	//
	// FIFTEEN SECONDS, WHICH IS SLOWER THAN ANY SUCH HOST WOULD DEMAND, AND DELIBERATELY SO. The
	// asymmetry is not close: being too slow costs this run some latency, while being too fast
	// costs it the source, can cost every later run the same source, and spends someone else's
	// server to do it. A host that publishes a number gets that number; a host that publishes
	// nothing has not consented to anything, and the floor should read as caution rather than as
	// the fastest rate we think we can get away with.
	//
	// It is well above the Crawl-delay range robots.txt files commonly carry (1–10s) on purpose:
	// where a host does state a delay, that value wins in either direction, so this only ever
	// governs hosts that have said nothing at all — and a host that has said nothing has agreed
	// to nothing. Fifteen seconds is the pace of a person reading, which is the traffic this tool
	// is supposed to look like.
	defaultHostInterval = 15 * time.Second
	// maxPaceWait bounds a single request's wait. Beyond it the request is REFUSED rather than
	// made to wait, and that distinction is the whole of it.
	//
	// TRUNCATING THE WAIT RECREATES THE BURST THE FLOOR EXISTS TO PREVENT. This used to clamp:
	// a caller whose slot fell past the cap was told to wait the cap instead. Raising the default
	// floor to 15s made it visible — eight callers queueing on one host produced waits of 0, 15,
	// 30, 45, 60, 60, 60, 60, and the last four would have departed together, at the one moment
	// the host is already saturated. A cap that collapses a queue is worse than no cap.
	//
	// So a slot beyond this is not a slot. The fetch fails, the seat is told the host is
	// saturated, and nothing is claimed — because claiming and then refusing would push the queue
	// out for a request that never went.
	maxPaceWait = 120 * time.Second
	// paceLockWait bounds contention on the lock itself, which is held only for a read-modify-
	// write and so should never be contended for long.
	paceLockWait = 10 * time.Second
	// retryAfterCap bounds how long a 429's own Retry-After is honoured inline. Beyond it the
	// refusal is returned, because a run must not silently sleep for minutes inside one fetch.
	retryAfterCap = 20 * time.Second
	// learnedCeiling bounds how slow a host's learned floor may become. A host that refuses us
	// every minute is not rate-limiting us, it is declining us, and waiting longer will not change
	// its mind — better to let the refusal reach the seat than to creep towards never asking.
	learnedCeiling = 60 * time.Second
)

// slot is a host's shared state: when it may next be asked, and what this machine has LEARNED
// about how fast it tolerates being asked.
//
// THE LEARNED FLOOR IS THE ANSWER TO "TUNE IT SO WE DO NOT HIT THE LIMIT". A static table can
// only carry the limits somebody published, and most hosts publish none — so the first time a
// host says 429 is the only time it will ever tell us its number. Throwing that away and
// returning to the default guarantees hitting it again on the next run. Doubling the floor and
// persisting it means each refusal is paid for exactly once per machine.
type slot struct {
	NextNanos int64 `json:"next"`
	// LearnedNanos is 0 until a host refuses us for going too fast.
	LearnedNanos int64 `json:"learned"`
}

// hostIntervals are the published limits, as each service states them. A host absent from this
// map gets defaultHostInterval.
//
// EVERY ENTRY CITES THE SERVICE'S OWN WORDS, because a limit invented here would be a number
// nobody could check and a promise to the origin that nobody made.
var hostIntervals = map[string]time.Duration{
	// arXiv's Terms of Use: "make no more than one request every three seconds, and limit
	// requests to a single connection at a time."
	"arxiv.org":        3 * time.Second,
	"export.arxiv.org": 3 * time.Second,
	"oaipmh.arxiv.org": 3 * time.Second,
	// Crossref answers `x-rate-limit-limit: 10` per `x-rate-limit-interval: 1s` in the polite
	// pool this tool's mailto puts it in (measured 2026-09-22; without the mailto it is 5). Half
	// the ceiling, because that pool is shared with every other polite client.
	"api.crossref.org": 200 * time.Millisecond,
	// NCBI asks for no more than 3 requests per second without an API key.
	"eutils.ncbi.nlm.nih.gov": 334 * time.Millisecond,
	"pmc.ncbi.nlm.nih.gov":    334 * time.Millisecond,
	// INSPIRE: "every IP address is allowed 15 requests in a 5s window", and a blocked request
	// still counts against the quota.
	"inspirehep.net": 334 * time.Millisecond,
	// The Internet Archive rate-limits, and says so at 200 with a page headed "Rate limit
	// reached" — measured 2026-09-22, while this very sweep was running. Both the CDX endpoint
	// and the snapshot host are the same service.
	"web.archive.org": time.Second,
	// EBI publishes no rate for the Europe PMC REST service and is visibly flaky under load,
	// answering a bare nginx 503 with no Retry-After. A service that falls over is telling you
	// something whether or not it means to, so this one is paced SLOWER than the default rather
	// than at it.
	"www.ebi.ac.uk": 15 * time.Second,
	// The PMC open-access bucket is plain S3 and NCBI names it a sanctioned automated route. Its
	// stated ceiling without a key is 3 requests a second; taken at a third of that, because a
	// sanctioned route is a courtesy to keep rather than a budget to spend.
	"pmc-oa-opendata.s3.amazonaws.com": time.Second,
	// api.openalex.org is the open-access index this tool asks. Its singleton lookups cost no
	// credits, but the budget is per IP and shared with anything else on this box.
	"api.openalex.org": 500 * time.Millisecond,
}

// paceDir is where the shared slots live. A package var so a test can point it somewhere
// disposable instead of contending with whatever else is running on this machine.
var paceDir = defaultPaceDir()

func defaultPaceDir() string {
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "feov-record", "pace")
}

// inProcess serialises this process's own goroutines before they contend for the file lock. It is
// an optimisation, not the mechanism: correctness comes from the lock, and removing this map
// changes only how much lock traffic one process generates.
var inProcess sync.Map // host -> *sync.Mutex

func hostMutex(host string) *sync.Mutex {
	m, _ := inProcess.LoadOrStore(host, &sync.Mutex{})
	return m.(*sync.Mutex)
}

// intervalFor is the floor for a host: our table, the default, and — winning whenever it asks
// for MORE room — the host's own published Crawl-delay.
//
// THE HOST'S NUMBER BEATS OURS, and measurement is why. arXiv's terms of use say "one request
// every three seconds" in prose and this table transcribed that; arXiv's robots.txt says
// `Crawl-delay: 15`. A number published for machines is the one addressed to us.
//
// Only ever slower. A host asking to be hit FASTER than our own floor is not a reason to take it
// up on the offer — the floor is also what protects a run from a host that would rather be
// hammered than lose a crawl.
func intervalFor(host string) time.Duration {
	iv, ok := hostIntervals[strings.ToLower(host)]
	if !ok {
		iv = defaultHostInterval
	}
	if r, loaded := robotsMem.Load(strings.ToLower(host)); loaded {
		if d := r.(*robotsRules).interval(); d > iv {
			return d
		}
	}
	return iv
}

// hashHost names a host's state files. A hash, so that a host with a colon, a slash or a case
// difference cannot produce a path that collides with another host's or escapes the directory.
func hashHost(host string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(host)))
	return hex.EncodeToString(sum[:8])
}

func slotFile(host string) string { return filepath.Join(paceDir, hashHost(host)) }

// reserveSlot claims this host's next slot across every process on the machine and returns how
// long the caller must wait for it. ok is false when the queue for this host is longer than
// maxPaceWait, in which case NOTHING is claimed and the caller must not proceed.
func reserveSlot(host string, extra time.Duration) (time.Duration, bool) {
	if host == "" {
		return 0, true
	}
	mu := hostMutex(host)
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll(paceDir, 0o755); err != nil {
		return defaultHostInterval, true // cannot coordinate: pace conservatively rather than freely
	}
	path := slotFile(host)
	fl := flock.New(path + ".lock")
	ctx, cancel := context.WithTimeout(context.Background(), paceLockWait)
	defer cancel()
	locked, err := fl.TryLockContext(ctx, 20*time.Millisecond)
	if err == nil && locked {
		defer fl.Unlock()
	} else {
		// FAILING OPEN WOULD BE THE ONE UNACCEPTABLE OUTCOME. Unlike the record's lock, where
		// proceeding unlocked risks a lost write, proceeding unpaced here sends traffic at an
		// origin that has no say in it. Wait the full interval instead.
		return intervalFor(host), true
	}

	now := time.Now()
	cur := readSlot(path)
	at := now
	if t := time.Unix(0, cur.NextNanos); cur.NextNanos > 0 && t.After(at) {
		at = t
	}
	wait := at.Sub(now)
	if wait > maxPaceWait {
		// REFUSE, AND CLAIM NOTHING. The queue stays as it was, so the requests already in it are
		// unaffected and this one simply does not join.
		return 0, false
	}
	iv := intervalFor(host)
	if l := time.Duration(cur.LearnedNanos); l > iv {
		iv = l
	}
	cur.NextNanos = at.Add(jittered(iv) + extra).UnixNano()
	writeSlot(path, cur)
	if wait < 0 {
		return 0, true
	}
	return wait, true
}

// jittered spreads a floor over [iv, 2*iv).
//
// A FIXED INTERVAL SYNCHRONISES CALLERS RATHER THAN SEPARATING THEM. Every process on this
// machine computes the same slot spacing from the same shared file, so a queue that forms on one
// host marches in lockstep — and several hosts whose queues started together stay together, which
// turns a polite per-host floor into a machine-wide pulse. The origin sees a burst every interval
// and quiet in between, which is the shape that trips a rate limiter even when the average rate
// is well inside its limit.
//
// Randomising the GAP rather than the start also means the spread compounds down a queue instead
// of shifting it: the tenth slot is scattered across a much wider window than the second, which
// is exactly where a long queue would otherwise be most regular.
//
// It only ever adds. The floor stays a floor — a jittered interval is never shorter than the
// number the host published or we chose.
func jittered(iv time.Duration) time.Duration {
	if iv <= 0 {
		return iv
	}
	return iv + time.Duration(rand.Int64N(int64(iv)))
}

// backoffHost pushes a host's next slot out by d, so a 429 slows every later request to that host
// — in every process — and not merely the one that was refused.
//
// AND IT RAISES THE FLOOR, which is the part that stops the next run repeating this one's mistake.
// The host has just told us our pace was wrong; the only number it will ever volunteer is the one
// implied by that refusal. Doubling, rather than adopting Retry-After as the interval, is
// deliberate: Retry-After says when to come back, not how often we may ask.
func backoffHost(host string, d time.Duration) {
	if host == "" || d <= 0 {
		return
	}
	if err := os.MkdirAll(paceDir, 0o755); err == nil {
		path := slotFile(host)
		mu := hostMutex(host)
		mu.Lock()
		cur := readSlot(path)
		learned := time.Duration(cur.LearnedNanos)
		if learned <= 0 {
			learned = intervalFor(host)
		}
		if learned *= 2; learned > learnedCeiling {
			learned = learnedCeiling
		}
		cur.LearnedNanos = int64(learned)
		writeSlot(path, cur)
		mu.Unlock()
	}
	reserveSlot(host, d)
}

func readSlot(path string) slot {
	var s slot
	b, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	if json.Unmarshal(b, &s) != nil {
		return slot{}
	}
	return s
}

func writeSlot(path string, s slot) {
	if b, err := json.Marshal(s); err == nil {
		_ = os.WriteFile(path, b, 0o644)
	}
}

// retryAfter reads a 429's own answer to "how long should I wait", in both forms the header
// takes: delta-seconds, and an HTTP-date. It returns 0 when the header is absent or unreadable,
// which the caller treats as "back off by the host's interval" rather than as "retry now".
func retryAfter(h http.Header) time.Duration {
	v := strings.TrimSpace(h.Get("Retry-After"))
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
