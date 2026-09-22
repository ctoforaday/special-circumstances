package fetchcache

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// mustReserve claims a slot and fails the test if the host is reported saturated, so a test that
// means "take the next slot" says so and does not silently read a refusal as a zero wait.
func mustReserve(t *testing.T, host string) time.Duration {
	t.Helper()
	w, ok := reserveSlot(host, 0)
	if !ok {
		t.Fatalf("%s reported saturated; this test expects a slot", host)
	}
	return w
}

func tempPaceDir(t *testing.T) {
	t.Helper()
	prev := paceDir
	paceDir = t.TempDir()
	t.Cleanup(func() { paceDir = prev })
}

// THE FLOOR IS PER HOST. A run's breadth must not be charged as pressure: two origins have
// nothing to say to each other about how fast either may be asked.
func TestTheFloorIsPerHostAndCumulative(t *testing.T) {
	tempPaceDir(t)
	var waits []time.Duration
	for i := 0; i < 4; i++ {
		waits = append(waits, mustReserve(t, "example.org"))
	}
	// EACH GAP IS JITTERED OVER [X, 2X), so the assertion is bounds and monotonicity rather than
	// exact multiples: the i-th wait lies in [i*X, i*2X], and no two callers share a slot.
	var prev time.Duration
	for i, w := range waits {
		lo, hi := time.Duration(i)*defaultHostInterval, time.Duration(i)*2*defaultHostInterval
		if w < lo-200*time.Millisecond || w > hi+200*time.Millisecond {
			t.Errorf("reservation %d waits %v, want within [%v, %v] (all: %v)", i, w, lo, hi, waits)
		}
		if i > 0 && w <= prev {
			t.Errorf("reservation %d waits %v, not after the previous %v — callers must queue, not depart together (all: %v)", i, w, prev, waits)
		}
		prev = w
	}
	if w := mustReserve(t, "unrelated.org"); w != 0 {
		t.Errorf("an unrelated host waited %v for another host's queue", w)
	}
	// arXiv's own published limit, not the default.
	tempPaceDir(t)
	mustReserve(t, "arxiv.org")
	if w := mustReserve(t, "arxiv.org"); w < 2500*time.Millisecond {
		t.Errorf("arxiv.org paced at %v; its terms of use say one request every three seconds", w)
	}
}

// GOROUTINES IN ONE PROCESS MUST QUEUE TOO. This is the weaker half of the contract and the one
// an in-process map already satisfied; it is kept so a refactor cannot lose it.
func TestConcurrentGoroutinesQueue(t *testing.T) {
	tempPaceDir(t)
	const n = 4
	var mu sync.Mutex
	var got []time.Duration
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := mustReserve(t, "example.org")
			mu.Lock()
			got = append(got, w)
			mu.Unlock()
		}()
	}
	wg.Wait()
	// DISTINCT WAITS, which is the invariant jitter must not break: two callers departing at the
	// same instant is the burst, whatever the spacing between slots happens to be.
	seen := map[time.Duration]bool{}
	for _, w := range got {
		if seen[w] {
			t.Errorf("two goroutines were given the same wait (%v); all: %v", w, got)
		}
		seen[w] = true
	}
}

// THE ONE THAT MATTERS: SEPARATE PROCESSES. Several seats fetch at once from different shells,
// so a floor that lives in one process's memory is decoration. This spawns real processes that
// each claim a slot for the same host and prints what they were told to wait; correct pacing
// gives each a distinct, increasing slot, and an in-process-only floor gives every one of them
// zero.
//
// It re-executes THIS test binary with an environment switch rather than building a helper,
// which is the standard way to get a real second process out of `go test`.
func TestSeparateProcessesShareTheFloor(t *testing.T) {
	if os.Getenv("PACE_CHILD") != "" {
		// Child: claim one slot and print the wait it was given, in nanoseconds.
		paceDir = os.Getenv("PACE_DIR")
		w, ok := reserveSlot("concurrent.example", 0)
		if !ok {
			os.Stdout.WriteString("saturated")
			os.Exit(0)
		}
		os.Stdout.WriteString(strconv.FormatInt(int64(w), 10))
		os.Exit(0)
	}
	dir := t.TempDir()
	const n = 4
	type res struct {
		wait time.Duration
		err  error
	}
	out := make([]res, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run", "TestSeparateProcessesShareTheFloor")
			cmd.Env = append(os.Environ(), "PACE_CHILD=1", "PACE_DIR="+dir)
			b, err := cmd.Output()
			if err != nil {
				out[i] = res{err: err}
				return
			}
			ns, perr := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
			out[i] = res{wait: time.Duration(ns), err: perr}
		}(i)
	}
	wg.Wait()

	seen := map[time.Duration]bool{}
	var waits []time.Duration
	for i, r := range out {
		if r.err != nil {
			t.Fatalf("child %d: %v", i, r.err)
		}
		waits = append(waits, r.wait)
		slot := r.wait
		if seen[slot] {
			t.Errorf("TWO PROCESSES WERE GIVEN THE SAME SLOT. %d processes claimed slots for one host "+
				"and at least two were told to wait the same %v — they would hit the origin together. "+
				"All waits: %v", n, r.wait, waits)
		}
		seen[slot] = true
	}
	// And the queue is a queue: with n processes the last one waits about (n-1) intervals.
	var max time.Duration
	for _, w := range waits {
		if w > max {
			max = w
		}
	}
	if want := time.Duration(n-1) * defaultHostInterval; max < want-500*time.Millisecond {
		t.Errorf("the longest wait among %d processes was %v, want at least %v — they are not queueing (all: %v)",
			n, max, want, waits)
	}
}

// A 429 SLOWS THE WHOLE HOST, in every process, not just the request that was refused.
func TestABackoffPushesTheWholeHostQueue(t *testing.T) {
	tempPaceDir(t)
	backoffHost("example.org", 5*time.Second)
	if w := mustReserve(t, "example.org"); w < 4*time.Second {
		t.Errorf("after a 5s backoff the next request waits %v", w)
	}
	if w := mustReserve(t, "other.org"); w != 0 {
		t.Errorf("one host's backoff delayed another by %v", w)
	}
}

// RETRY-AFTER IS THE HOST'S OWN ANSWER and arrives in two shapes; an absent or unreadable one
// must read as "no answer", never as "retry now".
func TestRetryAfterIsReadInBothForms(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		wantAtLeast time.Duration
		wantZero    bool
	}{
		{"delta seconds", "5", 5 * time.Second, false},
		{"http date", time.Now().Add(4 * time.Second).UTC().Format(http.TimeFormat), 2 * time.Second, false},
		{"absent", "", 0, true},
		{"unparseable", "soon please", 0, true},
		{"negative", "-3", 0, true},
		{"a date in the past", time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			if tc.value != "" {
				h.Set("Retry-After", tc.value)
			}
			got := retryAfter(h)
			if tc.wantZero && got != 0 {
				t.Errorf("retryAfter(%q) = %v, want 0 — an unreadable header is not permission to retry now", tc.value, got)
			}
			if !tc.wantZero && got < tc.wantAtLeast {
				t.Errorf("retryAfter(%q) = %v, want at least %v", tc.value, got, tc.wantAtLeast)
			}
		})
	}
}

// THE FLOOR LEARNS. A static table only carries limits somebody published, and most hosts publish
// none — so a 429 is the only time such a host will ever tell us its number. Discarding it
// guarantees hitting the same wall on the next run; this is what "tune it so we don't hit them"
// has to mean for a host that never said what the limit was.
func TestAHostThatRefusesUsRaisesItsOwnFloorPermanently(t *testing.T) {
	tempPaceDir(t)
	// A HOST WITH A SHORT PUBLISHED FLOOR, so this measures the LEARNING rather than the
	// saturation cap. At the 15s default a doubled, jittered floor runs past maxPaceWait and the
	// second reservation is legitimately refused — which is correct behaviour and the wrong thing
	// for this test to be asserting.
	const host = "touchy.example"
	hostIntervals[host] = time.Second
	t.Cleanup(func() { delete(hostIntervals, host) })

	if w := mustReserve(t, host); w != 0 {
		t.Fatalf("first call waited %v", w)
	}
	if w := mustReserve(t, host); w > 2*time.Second+200*time.Millisecond {
		t.Fatalf("pre-refusal wait %v exceeds one jittered second", w)
	}

	backoffHost(host, time.Millisecond) // the host says 429

	// AFTER the refusal the floor itself is higher, not merely the next slot pushed out. The
	// pace directory is deliberately NOT reset: the point is that the lesson persists in the
	// shared state a later process would read.
	if learned := time.Duration(readSlot(slotFile(host)).LearnedNanos); learned < 2*time.Second {
		t.Errorf("after a refusal the learned floor is %v, want at least double the %v published "+
			"interval — the lesson was not kept", learned, time.Second)
	}
}

// AND IT DOES NOT RATCHET TO NEVER. A host refusing every request is declining us, not pacing
// us, and creeping towards an hour between attempts would turn a refusal a seat should see into
// a hang it cannot.
func TestTheLearnedFloorIsCapped(t *testing.T) {
	tempPaceDir(t)
	const host = "hostile.example"
	for i := 0; i < 20; i++ {
		backoffHost(host, time.Millisecond)
	}
	// Read the learned floor from the shared state rather than through a reservation: past the
	// ceiling the next reservation is legitimately REFUSED, so measuring it by waiting would be
	// measuring the refusal instead.
	got := time.Duration(readSlot(slotFile(host)).LearnedNanos)
	if got > learnedCeiling {
		t.Errorf("after 20 refusals the learned floor is %v, above the %v ceiling", got, learnedCeiling)
	}
	if got < defaultHostInterval {
		t.Errorf("the learned floor is %v, below the default — the refusals taught nothing", got)
	}
}

// A QUEUE LONGER THAN WE WILL WAIT IS REFUSED, NOT TRUNCATED. Clamping the wait to the cap gave
// every caller past it the SAME slot, so they departed together — the burst the floor exists to
// prevent, produced by the safety valve, at the one moment the host is already saturated.
func TestASaturatedHostIsRefusedRatherThanCollapsedIntoABurst(t *testing.T) {
	tempPaceDir(t)
	const host = "busy.example"
	var waits []time.Duration
	refusals := 0
	for i := 0; i < 20; i++ {
		w, ok := reserveSlot(host, 0)
		if !ok {
			refusals++
			continue
		}
		waits = append(waits, w)
	}
	if refusals == 0 {
		t.Fatalf("20 reservations on one host produced no refusal; waits: %v", waits)
	}
	// EVERY GRANTED SLOT IS DISTINCT. That is the invariant the clamp broke.
	seen := map[time.Duration]bool{}
	for _, w := range waits {
		if seen[w] {
			t.Errorf("two granted slots collide at %v; all waits: %v", w, waits)
		}
		seen[w] = true
		if w > maxPaceWait {
			t.Errorf("a granted slot waits %v, beyond the %v cap", w, maxPaceWait)
		}
	}
}

// A 5xx MEANS "NOT NOW", AND CONTINUING AT THE PACE THAT MET IT IS THE WRONG ANSWER. Europe PMC
// falls over with a bare nginx 503 under load and sends no Retry-After; treated as an ordinary
// refusal, the next request leaves at exactly the rate that just failed.
func TestAnOverloadedServiceIsBackedOffFromLikeARateLimit(t *testing.T) {
	for code, want := range map[int]bool{
		429: true, 502: true, 503: true, 504: true,
		403: false, 404: false, 200: false, 500: false,
	} {
		if got := isOverloadStatus(code); got != want {
			t.Errorf("isOverloadStatus(%d) = %v, want %v", code, got, want)
		}
	}
}

// AND THE FLOOR FOR AN UNSPOKEN HOST IS CAUTION, NOT THE FASTEST RATE WE THINK WE CAN GET AWAY
// WITH. A host that publishes nothing has consented to nothing.
func TestTheDefaultFloorIsSlowerThanAnyPublishedCrawlDelay(t *testing.T) {
	// robots.txt Crawl-delay values in the wild run 1–10s; the default sits above the common
	// range because it governs only hosts that have said nothing at all.
	if defaultHostInterval < 15*time.Second {
		t.Errorf("defaultHostInterval = %v; the asymmetry is not close — too slow costs latency, "+
			"too fast can cost every later run the source", defaultHostInterval)
	}
	// A host that DOES publish a delay still wins, in either direction of our table.
	tempPaceDir(t)
	robotsMem.Store("slow.example", &robotsRules{CrawlDelay: 20, Fetched: time.Now()})
	if iv := intervalFor("slow.example"); iv != 20*time.Second {
		t.Errorf("a published Crawl-delay of 20s was paced at %v", iv)
	}
	robotsMem.Delete("slow.example")
}

// THE CALL SITE, NOT THE PREDICATE. A test of isOverloadStatus passes whether or not anything
// consults it: reverting the fetcher to back off on 429 alone left the predicate's own test green.
// This drives a real 503 through the real fetcher and asks whether the host's floor moved.
func TestA503ThroughTheFetcherActuallyRaisesTheFloor(t *testing.T) {
	tempPaceDir(t)
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		hits++
		w.WriteHeader(http.StatusServiceUnavailable) // no Retry-After, as EBI sends it
		_, _ = w.Write([]byte("<center>nginx</center>"))
	}))
	defer srv.Close()

	h := NewHTTPFetcher()
	host := strings.TrimPrefix(srv.URL, "http://")
	if _, err := h.Fetch(srv.URL + "/paper"); err == nil {
		t.Fatal("a 503 was reported as success")
	}
	// IT RETRIED ONCE, because a 5xx means "not now" rather than "not ever".
	if hits < 2 {
		t.Errorf("the fetcher hit the overloaded service %d time(s); it should honour one retry", hits)
	}
	// AND THE FLOOR FOR THAT HOST IS NOW ABOVE THE DEFAULT, so the next request — in this process
	// or any other — does not leave at the rate that just failed.
	mustReserve(t, host)
	if w := mustReserve(t, host); w <= defaultHostInterval {
		t.Errorf("after a 503 the next slot waits %v, no more than the %v default — the overload "+
			"was reported and not learned from", w, defaultHostInterval)
	}
}

// JITTER SPREADS A QUEUE INSTEAD OF MARCHING IT. Every process computes its spacing from the same
// shared file, so a fixed interval keeps callers in lockstep and the origin sees a pulse rather
// than a trickle — the shape that trips a rate limiter even when the average rate is well inside
// it.
func TestTheFloorIsJitteredAndNeverShorterThanItself(t *testing.T) {
	const iv = 4 * time.Second
	var sawLonger bool
	for i := 0; i < 60; i++ {
		got := jittered(iv)
		if got < iv {
			t.Fatalf("jittered(%v) = %v, SHORTER than the floor — jitter may only ever add", iv, got)
		}
		if got >= 2*iv {
			t.Fatalf("jittered(%v) = %v, beyond [X,2X)", iv, got)
		}
		if got > iv {
			sawLonger = true
		}
	}
	if !sawLonger {
		t.Error("60 draws never exceeded the floor; the interval is not actually being spread")
	}
	// A zero or negative floor is returned untouched rather than panicking in rand.
	if got := jittered(0); got != 0 {
		t.Errorf("jittered(0) = %v", got)
	}
}
