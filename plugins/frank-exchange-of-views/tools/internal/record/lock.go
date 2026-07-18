package record

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Locking, dependency-free.
//
// The shards are single-writer by design, but the seat POINTER and the shared
// PROJECTIONS are not: six lens processes mutating concurrently means six renders
// racing onto one ledger.md, and on Windows a rename-over-existing under
// contention throws rather than last-writer-wins.
//
// mkdir is the atomic primitive every lockfile package wraps, so the oracle's
// algorithm ports directly and stays stdlib-only and portable (the R2g plan
// chose this over flock/LockFileEx deliberately: no x/sys dependency, no
// platform fork, and the semantics are already test-covered). What Go adds is
// `go test -race` over the concurrency the mjs could only spawn-test.
//
// Acquire = mkdir (EEXIST -> wait); a stale lock from a crashed holder is stolen
// after staleMS by mtime; the wait is BOUNDED and then proceeds unlocked rather
// than deadlocking a seat — a lost render self-heals on the next mutation.
const (
	staleMS = 10 * time.Second
	waitMS  = 5 * time.Second
)

func withLock(runDir, name string, fn func()) {
	lockDir := filepath.Join(recordsDir(runDir), ".lock-"+name)
	os.MkdirAll(recordsDir(runDir), 0o755)
	deadline := time.Now().Add(waitMS)
	held := false
	for !held {
		err := os.Mkdir(lockDir, 0o755)
		if err == nil {
			held = true
			heldLocks.Store(lockDir, struct{}{}) // released on signal, not just on return
			break
		}
		if !os.IsExist(err) {
			break // an unexpected error proceeds unlocked, as in the oracle's throw path
		}
		if st, serr := os.Stat(lockDir); serr == nil {
			if time.Since(st.ModTime()) > staleMS {
				os.Remove(lockDir)
				continue
			}
		} else {
			continue
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	}
	defer func() {
		if held {
			heldLocks.Delete(lockDir)
			os.Remove(lockDir)
		}
	}()
	fn()
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

// timeNowMinus is a test seam for ageing a lock without sleeping.
func timeNowMinus(d time.Duration) time.Time { return time.Now().Add(-d) }
