package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

// THE BUG THIS FILE EXISTS FOR, and it was measured before it was fixed.
//
// Six FileChanged events fired concurrently against a six-check note recorded
// FOUR records. Two were lost. Every hook process read the state, added its own
// key and wrote the whole map back, so whichever wrote last erased everything
// that landed after its own read. That is #165's "coverage stops advancing":
// records were never stopping, they were being overwritten by their siblings.
func TestConcurrentUpdatesDoNotLoseRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")

	const n = 24
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "check-" + strconv.Itoa(i)
			if err := UpdateRearm(path, func(s *RearmState) {
				s.Rearmed[key] = Rearm{Index: i, Check: key, By: "x.go", Event: "change", At: "t"}
			}); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("UpdateRearm: %v", err)
	}

	s, err := LoadRearm(os.ReadFile, path)
	if err != nil {
		t.Fatalf("state is unreadable after concurrent writes — it tore: %v", err)
	}
	if len(s.Rearmed) != n {
		t.Fatalf("kept %d of %d records; concurrent updates are erasing each other", len(s.Rearmed), n)
	}
}

// The other half of the same failure: a short write landing over a longer file
// leaves the previous document's tail behind. A live rearmed.json in this repo
// was found in exactly that state — a complete 287-byte document followed by the
// tail of a longer one, with five records stamped the same second.
func TestConcurrentUpdatesNeverLeaveATornFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")

	// Seed a large state, then race many small ones against it. Shrinking writes
	// are what tear.
	if err := UpdateRearm(path, func(s *RearmState) {
		for i := 0; i < 40; i++ {
			k := "seed-" + strconv.Itoa(i)
			s.Rearmed[k] = Rearm{Index: i, Check: k, By: "a/very/long/path/to/make/this/document/large.go", Event: "change", At: "t"}
		}
	}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = UpdateRearm(path, func(s *RearmState) {
				s.Rearmed = map[string]Rearm{"small": {Index: 1, Check: "small", At: "t"}}
			})
		}(i)
	}
	wg.Wait()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var probe map[string]any
	if err := json.Unmarshal(b, &probe); err != nil {
		t.Fatalf("file is not valid JSON after racing writes — it tore:\n%s", b)
	}
}

// A lock left behind by a process that died must not wedge the mechanism.
// Losing one re-arm is cheap; losing the file is not, and a permanent block
// would be worse than either.
func TestAnAbandonedLockIsBrokenRatherThanWaitedOnForever(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rearmed.json")
	stale := path + ".lock"
	if err := os.WriteFile(stale, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	// Backdate it well past the timeout.
	old := timeNowMinus(2 * lockRearmTimeout)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}

	if err := UpdateRearm(path, func(s *RearmState) {
		s.Rearmed["k"] = Rearm{Index: 1, Check: "k", At: "t"}
	}); err != nil {
		t.Fatalf("an abandoned lock blocked the update: %v", err)
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("the lock file was left behind")
	}
}

// timeNowMinus keeps the test's intent readable without importing time at the
// top for one call.
func timeNowMinus(d time.Duration) time.Time { return time.Now().Add(-d) }

// #215: unlock must not remove a lock it no longer owns.
//
// The scenario, which the first version got wrong: A holds the lock; B judges it
// abandoned, breaks it and takes its own; A then finishes. A's unlock used to
// remove B's lock, handing a third process exclusion B still believed it had.
func TestUnlockDoesNotRemoveASuccessorsLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rearmed.json")
	lock := path + ".lock"

	unlockA, err := lockRearm(path)
	if err != nil {
		t.Fatal(err)
	}
	held, err := os.ReadFile(lock)
	if err != nil {
		t.Fatalf("the lock file carries no holder: %v", err)
	}

	// B takes over, exactly as a stale break would: different nonce, same path.
	if err := os.WriteFile(lock, []byte("someone-else"), 0o644); err != nil {
		t.Fatal(err)
	}

	unlockA() // A finishes late

	after, err := os.ReadFile(lock)
	if err != nil {
		t.Fatalf("A's unlock removed a lock it did not hold — B is now unprotected: %v", err)
	}
	if string(after) != "someone-else" {
		t.Errorf("lock content = %q, want B's; A's was %q", after, held)
	}
}

// A lock that is live must not be broken just because it looks old: the break
// re-reads the nonce and only removes the exact file it judged abandoned.
func TestAStaleBreakLeavesALockThatWasReplacedUnderIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rearmed.json")
	lock := path + ".lock"

	// An abandoned-looking lock.
	if err := os.WriteFile(lock, []byte("dead-holder"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := timeNowMinus(2 * lockRearmTimeout)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}

	// It is broken and acquired, and the acquisition is ours.
	unlock, err := lockRearm(path)
	if err != nil {
		t.Fatalf("an abandoned lock blocked the update: %v", err)
	}
	got, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) == "dead-holder" {
		t.Error("the stale lock was adopted rather than replaced; unlock would then be a no-op forever")
	}
	unlock()
	if _, err := os.Stat(lock); err == nil {
		t.Error("our own lock was left behind")
	}
}
