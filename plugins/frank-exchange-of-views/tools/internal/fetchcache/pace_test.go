package fetchcache

import (
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

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
		waits = append(waits, reserveSlot("example.org", 0))
	}
	for i, w := range waits {
		want := time.Duration(i) * defaultHostInterval
		if w < want-200*time.Millisecond || w > want+200*time.Millisecond {
			t.Errorf("reservation %d waits %v, want about %v (all: %v) — callers must queue, not depart together", i, w, want, waits)
		}
	}
	if w := reserveSlot("unrelated.org", 0); w != 0 {
		t.Errorf("an unrelated host waited %v for another host's queue", w)
	}
	// arXiv's own published limit, not the default.
	tempPaceDir(t)
	reserveSlot("arxiv.org", 0)
	if w := reserveSlot("arxiv.org", 0); w < 2500*time.Millisecond {
		t.Errorf("arxiv.org paced at %v; its terms of use say one request every three seconds", w)
	}
}

// GOROUTINES IN ONE PROCESS MUST QUEUE TOO. This is the weaker half of the contract and the one
// an in-process map already satisfied; it is kept so a refactor cannot lose it.
func TestConcurrentGoroutinesQueue(t *testing.T) {
	tempPaceDir(t)
	const n = 6
	var mu sync.Mutex
	var got []time.Duration
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := reserveSlot("example.org", 0)
			mu.Lock()
			got = append(got, w)
			mu.Unlock()
		}()
	}
	wg.Wait()
	seen := map[int64]bool{}
	for _, w := range got {
		// Rounded, not truncated: a wait of 1.9999s is slot 1, and integer division calls it 0.
		slot := int64((w + defaultHostInterval/2) / defaultHostInterval)
		if seen[slot] {
			t.Errorf("two goroutines were given the same slot (%v); waits: %v", slot, got)
		}
		seen[slot] = true
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
		os.Stdout.WriteString(strconv.FormatInt(int64(reserveSlot("concurrent.example", 0)), 10))
		os.Exit(0)
	}
	dir := t.TempDir()
	const n = 8
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

	seen := map[int64]bool{}
	var waits []time.Duration
	for i, r := range out {
		if r.err != nil {
			t.Fatalf("child %d: %v", i, r.err)
		}
		waits = append(waits, r.wait)
		slot := int64((r.wait + 200*time.Millisecond) / defaultHostInterval)
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
		t.Errorf("the longest wait among %d processes was %v, want about %v — they are not queueing (all: %v)",
			n, max, want, waits)
	}
}

// A 429 SLOWS THE WHOLE HOST, in every process, not just the request that was refused.
func TestABackoffPushesTheWholeHostQueue(t *testing.T) {
	tempPaceDir(t)
	backoffHost("example.org", 5*time.Second)
	if w := reserveSlot("example.org", 0); w < 4*time.Second {
		t.Errorf("after a 5s backoff the next request waits %v", w)
	}
	if w := reserveSlot("other.org", 0); w != 0 {
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
	const host = "touchy.example"
	base := reserveSlot(host, 0)
	if base != 0 {
		t.Fatalf("first call waited %v", base)
	}
	// Before learning, the second call waits the default.
	if w := reserveSlot(host, 0); w > defaultHostInterval+200*time.Millisecond {
		t.Fatalf("pre-refusal wait %v already exceeds the default", w)
	}

	backoffHost(host, time.Second) // the host says 429

	// AFTER the refusal the floor itself is higher, not merely the next slot pushed out. The
	// pace directory is deliberately NOT reset here: the point is that the lesson persists in
	// the shared state a later process would read.
	w1 := reserveSlot(host, 0)
	w2 := reserveSlot(host, 0)
	if gap := w2 - w1; gap < 2*defaultHostInterval-300*time.Millisecond {
		t.Errorf("after a refusal consecutive slots are %v apart, want at least double the %v default — "+
			"the lesson was not kept", gap, defaultHostInterval)
	}
}

// AND IT DOES NOT RATCHET TO NEVER. A host refusing every request is declining us, not pacing us,
// and creeping towards an hour between attempts would turn a refusal a seat should see into a
// hang it cannot.
func TestTheLearnedFloorIsCapped(t *testing.T) {
	tempPaceDir(t)
	const host = "hostile.example"
	for i := 0; i < 20; i++ {
		backoffHost(host, time.Millisecond)
	}
	reserveSlot(host, 0)
	w := reserveSlot(host, 0)
	if w > learnedCeiling+time.Second {
		t.Errorf("after 20 refusals the floor is %v, above the %v ceiling", w, learnedCeiling)
	}
}
