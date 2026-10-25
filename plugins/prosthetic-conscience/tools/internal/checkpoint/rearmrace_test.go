package checkpoint

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// A RELEASE THAT CANNOT READ ITS LOCK SAYS SO (#1553). It used to be dropped: the lock stayed,
// every waiting hook timed out, and nothing named why.
func TestAReleaseThatFailsIsReported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")
	lock, err := lockRearm(path)
	if err != nil {
		t.Fatal(err)
	}
	// The lock's path becomes something no read can open as a file, on any platform.
	if err := os.Remove(lock.path); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(lock.path, "occupant"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := lock.release(); err == nil {
		t.Fatal("a release that could not read its lock reported success")
	}
}

// A COMMITTED UPDATE WHOSE RELEASE FAILED SAYS BOTH HALVES: the event is recorded, and the lock is
// left for the next hook to wait on. The hook reads the type to keep reporting the commit.
func TestALockLeftErrorSaysTheEventIsRecorded(t *testing.T) {
	var err error = &LockLeftError{Path: "rearmed.json", Err: os.ErrPermission}
	var left *LockLeftError
	if !errors.As(err, &left) || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("LockLeftError must be matchable and unwrap its cause: %v", err)
	}
	for _, want := range []string{"IS recorded", "could not be released", "waits up to"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must carry %q: %v", want, err)
		}
	}
}

// THE TIMEOUT NAMES THE HOLDER, THE LOCK'S AGE AND THE LAST OPEN ERROR (#1553): the next lock that
// does not free has to say whose it was and whether it was being worked or left behind.
func TestTheLockTimeoutNamesItsHolder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rearmed.json")
	if err := os.WriteFile(path+".lock", []byte("4242-holder"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Stamped a second ahead, so the wait's deadline arrives before the lock ages into a
	// stale-break and the timeout is what this test sees.
	ahead := time.Now().Add(time.Second)
	if err := os.Chtimes(path+".lock", ahead, ahead); err != nil {
		t.Fatal(err)
	}
	_, err := lockRearm(path)
	if err == nil {
		t.Fatal("a fresh lock held by another hook was taken")
	}
	for _, want := range []string{`held by "4242-holder"`, "lock age", "last open error"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the timeout must carry %q: %v", want, err)
		}
	}
}
