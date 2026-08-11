package record

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

// Locking, on real OS advisory locks.
//
// The shards are single-writer by design, but the seat POINTER and the shared
// PROJECTIONS are not: six lens processes mutating concurrently means six
// renders racing onto one ledger.md, and on Windows a rename-over-existing under
// contention fails rather than last-writer-wins.
//
// WHY THIS REPLACED THE MKDIR LOCK. The ported algorithm acquired by mkdir and
// STOLE any lock whose directory was older than ten seconds, on the theory that
// the holder had crashed. That theory is unfalsifiable from outside the process:
// a merge seat rendering a large board, or any seat on a loaded machine, is
// indistinguishable from a dead one — so the heuristic could revoke a lock from
// a LIVE holder and admit two writers to the critical section, which is the
// exact failure the lock exists to prevent. It was also untestable, since no
// test can hold a lock "slowly but legitimately" on demand.
//
// flock has no such hazard: the kernel releases the lock when the holder's
// process dies, so a crashed holder needs no timeout and a slow holder is never
// robbed. LockFileEx on Windows, flock(2) on Unix, one dependency, no heuristic
// to tune. The wait stays BOUNDED — a seat that cannot acquire in time proceeds
// rather than deadlocking, because a lost render self-heals on the next mutation
// while a hung seat costs the round.
const lockWait = 5 * time.Second

// heldFlocks lets the signal guard release locks on an interrupted seat. The
// kernel would release them at exit anyway; doing it explicitly means release
// happens BEFORE teardown, so a peer waiting on the lock proceeds immediately
// rather than serving out its own bounded wait.
var heldFlocks sync.Map // *flock.Flock -> struct{}

// withLock takes the RESOLVED record directory. The lock files live beside the shards they
// guard, so a separated run must lock in the separated root — locking under the run directory
// would leave concurrent seats guarding a path neither of them writes.
func withLock(recDir, name string, fn func()) {
	os.MkdirAll(recDir, 0o755)
	fl := flock.New(filepath.Join(recDir, ".lock-"+name))

	ctx, cancel := context.WithTimeout(context.Background(), lockWait)
	defer cancel()
	locked, err := fl.TryLockContext(ctx, 50*time.Millisecond)
	if err == nil && locked {
		heldFlocks.Store(fl, struct{}{})
		defer func() {
			heldFlocks.Delete(fl)
			fl.Unlock()
		}()
	}
	// Failing to acquire is survivable and deliberate: proceed unlocked rather
	// than abandon the seat's write. Projections are full-state, so the worst
	// case is a render that loses a race and is rewritten by the next mutation.
	fn()
}

func releaseHeldLocks() {
	heldFlocks.Range(func(k, _ any) bool {
		if fl, ok := k.(*flock.Flock); ok {
			fl.Unlock()
		}
		return true
	})
}

// writeAtomic is temp + fsync + rename + dir-fsync (plan-audit sitting-4 note 1,
// hardened): atomic for a concurrent reader AND durable across a crash.
func writeAtomic(path string, content []byte) error {
	b := make([]byte, 3)
	if _, err := crand.Read(b); err != nil {
		return err
	}
	tmp := path + ".tmp-" + hex.EncodeToString(b)
	if err := durableWrite(path, content, tmp); err != nil {
		// A projection is FULL STATE, so a dropped render is a delay and never a
		// loss: the next mutation rewrites it from the event log, which is the
		// surface that had to be durable. Failing the seat's command here would
		// turn a cosmetic stall into a lost record.
		return nil
	}
	return nil
}

// renameWithRetry absorbs the Windows failure mode: rename-over-existing can
// return EPERM under a concurrent reader or racer even inside the lock
// (antivirus, search indexer). Brief backoff, then let the caller give up.
func renameWithRetry(tmp, path string) error {
	var err error
	for i := 0; i < 5; i++ {
		if err = os.Rename(tmp, path); err == nil {
			return nil
		}
		if !isRetryableRename(err) {
			return err
		}
		time.Sleep(time.Duration(20*(i+1)) * time.Millisecond)
	}
	return err
}

func isRetryableRename(err error) bool {
	s := err.Error()
	for _, code := range []string{"EPERM", "EBUSY", "EACCES", "EEXIST",
		"permission denied", "being used by another process", "Access is denied", "file exists"} {
		if strings.Contains(s, code) {
			return true
		}
	}
	return os.IsPermission(err) || os.IsExist(err)
}
