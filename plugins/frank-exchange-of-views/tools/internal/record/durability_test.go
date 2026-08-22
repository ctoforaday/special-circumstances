package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// FIVE TESTS ARE GONE FROM THIS FILE, and each named a mechanism the record no longer has.
//
//   - TestAppendEntersACriticalSection — the append took a critical section around a FILE write.
//   - TestAppendHealsATornFinalLine — a crash mid-append left an unterminated fragment, and
//     appending onto it would destroy the next event too. A transaction commits or it does not.
//   - TestAppendAssignsGapFreePerShardSequence — `seq` was the seat's position in its own shard;
//     nothing read it, and numbering it cost every write a query.
//   - TestRegisterRotatesTheNonceAndRepointsTheSeat — the nonce named a shard file and the pointer
//     said which was live. Both are gone; the sitting is the count of a seat's registers.
//   - TestAppendImplicitlyRegistersWhenThePointerIsAbsent — there is no pointer to be absent.
//
// What they protected is not unprotected: atomicity is the transaction (recordsql.InsertTx),
// concurrent writers are covered by TestConcurrentSeatsDoNotLoseEvents in recordsql — which spawns
// real processes, because goroutines did not reproduce the hazard — and the write path's
// registration and seat-id rules are still tested below.

// enterCritical must BLOCK once shutdown has begun: starting a write we may not
// be allowed to finish is the truncation the guard exists to prevent.
func TestEnterCriticalBlocksOnceShuttingDown(t *testing.T) {
	guardMu.Lock()
	shuttingDown = true
	guardMu.Unlock()
	t.Cleanup(func() {
		guardMu.Lock()
		shuttingDown = false
		guardCond.Broadcast()
		guardMu.Unlock()
	})

	entered := make(chan struct{})
	go func() {
		enterCritical()
		exitCritical()
		close(entered)
	}()

	select {
	case <-entered:
		t.Fatal("enterCritical proceeded during shutdown — a write could start that the exit will not wait for")
	default:
	}

	// Releasing the shutdown flag lets the blocked writer through.
	guardMu.Lock()
	shuttingDown = false
	guardCond.Broadcast()
	guardMu.Unlock()
	<-entered
}

// A blocked caller must not increment the depth, or it would delay the very exit
// it is waiting on — a deadlock at shutdown.
func TestBlockedCallerDoesNotHoldTheDepth(t *testing.T) {
	guardMu.Lock()
	shuttingDown = true
	guardMu.Unlock()
	t.Cleanup(func() {
		guardMu.Lock()
		shuttingDown = false
		guardCond.Broadcast()
		guardMu.Unlock()
	})

	started := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		close(started)
		enterCritical()
		exitCritical()
	}()
	<-started

	guardMu.Lock()
	d := guardDepth
	guardMu.Unlock()
	if d != 0 {
		t.Errorf("guardDepth = %d while a caller is BLOCKED on shutdown, want 0", d)
	}

	// Release and JOIN before returning. The depth is process-global, so a
	// goroutine still blocked in enterCritical when this test ends would enter it
	// during a later test and corrupt that test's reading.
	guardMu.Lock()
	shuttingDown = false
	guardCond.Broadcast()
	guardMu.Unlock()
	<-finished
}

// Nested critical sections must both be counted, so an outer section is not
// released by an inner one finishing.
func TestCriticalSectionsNest(t *testing.T) {
	// The depth is process-global, so this asserts on the DELTA rather than an
	// absolute value: what nesting has to guarantee is that each enter adds
	// exactly one and each exit removes exactly one, whatever the starting point.
	depth := func() int {
		guardMu.Lock()
		defer guardMu.Unlock()
		return guardDepth
	}
	base := depth()

	enterCritical()
	enterCritical()
	if got := depth() - base; got != 2 {
		t.Errorf("depth rose by %d after two enters, want 2", got)
	}
	exitCritical()
	if got := depth() - base; got != 1 {
		t.Errorf("depth is +%d after one exit of two, want 1 — the outer section was released early", got)
	}
	exitCritical()
	if got := depth() - base; got != 0 {
		t.Errorf("depth is +%d after balanced exits, want 0", got)
	}
}
func TestRegisterSeatRejectsMalformedSeatIDs(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"leading digit", "1lens"},
		{"leading hyphen", "-lens"},
		{"path traversal", "../escape"},
		{"path separator", "red/lens"},
		{"windows separator", `red\lens`},
		{"underscore is not in the alphabet", "red_lens"},
		{"space", "red lens"},
		{"dot", "red.lens"},
		{"nul byte", "red\x00lens"},
		{"newline", "red\nlens"},
		{"unicode", "red-日本語"},
		{"glob", "red-*"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			_, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: tc.id, Round: RoundOf(tc.id)})
			if err == nil {
				t.Fatalf("RegisterSeat accepted %q — the id becomes a FILENAME", tc.id)
			}
			if !strings.Contains(err.Error(), "invalid --seat-id") {
				t.Errorf("refusal does not name the flag: %v", err)
			}
			// Nothing was written under a rejected id.
			entries, rerr := os.ReadDir(recordsDirT(runDir))
			if rerr == nil {
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), "events-") || strings.HasPrefix(e.Name(), ".active-") {
						t.Errorf("a rejected seat id still produced %s", e.Name())
					}
				}
			}
		})
	}
}

func TestRegisterSeatAcceptsTheEngineAssignedShapes(t *testing.T) {
	for _, id := range []string{"red-lens-r1-L1", "red-merge-r12", "blue-lane-3", "blue-synthesize", "frontier", "judge-r1", "assemble", "a"} {
		runDir := t.TempDir()
		if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: id, Round: RoundOf(id)}); err != nil {
			t.Errorf("RegisterSeat(%q) = %v, want accepted", id, err)
		}
	}
}

// writeAtomic must never leave a temp file behind, and must never publish a
// half-written projection.
func TestWriteAtomicLeavesNoTempAndPublishesWholeContent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ledger.md")
	content := []byte(strings.Repeat("projection line\n", 5000))
	if err := writeAtomic(target, content); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(content) {
		t.Errorf("published %d bytes, want %d", len(got), len(content))
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leaked temp artifact: %s", e.Name())
		}
	}
	// Overwriting an existing projection is the common case; it must replace whole.
	shorter := []byte("replaced\n")
	if err := writeAtomic(target, shorter); err != nil {
		t.Fatal(err)
	}
	got, _ = os.ReadFile(target)
	if string(got) != string(shorter) {
		t.Errorf("overwrite left %q, want %q — the old tail survived", got, shorter)
	}
}

// Concurrent writers to the SAME projection: the lock plus atomic rename means a
// reader never sees a partial file, whichever writer wins.
func TestConcurrentWriteAtomicNeverPublishesAPartialFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ledger.md")
	a := []byte(strings.Repeat("A", 200000))
	b := []byte(strings.Repeat("B", 200000))

	var writers, reader sync.WaitGroup
	stop := make(chan struct{})
	// A reader hammering the file while writers swap it.
	var bad []string
	var mu sync.Mutex
	reader.Add(1)
	go func() {
		defer reader.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			got, err := os.ReadFile(target)
			if err != nil || len(got) == 0 {
				continue
			}
			if s := string(got); s != string(a) && s != string(b) {
				mu.Lock()
				bad = append(bad, fmt.Sprintf("torn read of %d bytes", len(got)))
				mu.Unlock()
				return
			}
		}
	}()
	for i := 0; i < 20; i++ {
		writers.Add(1)
		go func(i int) {
			defer writers.Done()
			c := a
			if i%2 == 1 {
				c = b
			}
			if err := writeAtomic(target, c); err != nil {
				mu.Lock()
				bad = append(bad, err.Error())
				mu.Unlock()
			}
		}(i)
	}
	writers.Wait()
	close(stop)
	reader.Wait()

	mu.Lock()
	defer mu.Unlock()
	for _, s := range bad {
		t.Error(s)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leaked temp artifact under contention: %s", e.Name())
		}
	}
}

// A lock the process already holds must not be lost by releaseHeldLocks running
// on a seat that holds nothing — the signal path calls it unconditionally.
func TestReleaseHeldLocksIsSafeWhenNothingIsHeld(t *testing.T) {
	releaseHeldLocks()
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	releaseHeldLocks()
	if _, err := BoardState(runDir); err != nil {
		t.Errorf("board replay after a spurious release: %v", err)
	}
}

func TestIsRetryableRename(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"windows sharing violation", fmt.Errorf("rename: The process cannot access the file because it is being used by another process"), true},
		{"windows access denied", fmt.Errorf("rename: Access is denied"), true},
		{"posix permission", fmt.Errorf("rename: permission denied"), true},
		{"posix exists", fmt.Errorf("rename: file exists"), true},
		{"errno spelling", fmt.Errorf("EBUSY"), true},
		{"a real, non-transient failure is NOT retried", fmt.Errorf("no such file or directory"), false},
		{"disk full is not transient", fmt.Errorf("no space left on device"), false},
		{"os.ErrPermission is recognized by identity", os.ErrPermission, true},
		{"os.ErrExist is recognized by identity", os.ErrExist, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryableRename(tc.err); got != tc.want {
				t.Errorf("isRetryableRename(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
