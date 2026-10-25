package fuzz

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// THE SWEEPS QUEUE ON A SHARED BOX. A local sweep spawns the real binary from NumCPU*3 workers,
// and two sessions sweeping at once on one machine do not each take twice as long — they push the
// load average past the point where every other session's `go test` times out too. So a local run
// with FEOV_RELEASE_GATE set takes an exclusive lock under the user cache before the first test
// starts, and a second sweep waits for it, printing who holds it.
//
// CI TAKES NO LOCK: a runner is one job's machine, and CI sets CI. The lock is taken in TestMain,
// before m.Run starts the test binary's -timeout alarm, so the time spent waiting does not count
// against the sweep's own bound. It is released when the process exits, however it exits — flock
// belongs to the open file, and the kernel closes that.

// realCacheDir is the developer's user cache, read at package initialisation: TestMain hands HOME
// to testbuild's sandbox, under which os.UserCacheDir names a temp directory no other session sees.
var realCacheDir, realCacheDirErr = os.UserCacheDir()

// heldSweepLock keeps the lock file reachable for the life of the process. An *os.File nothing
// references is closed by its finalizer, and closing it releases the lock mid-sweep.
var heldSweepLock *os.File

// sweepLockPoll is how often a waiting sweep retries, and re-reads who holds the lock.
const sweepLockPoll = 5 * time.Second

// holdSweepLockIfLocal takes the box-wide sweep lock when this process is a LOCAL sweep: the gate
// is set and CI is not. It returns nil, nil when no lock applies.
func holdSweepLockIfLocal() (*os.File, error) {
	if os.Getenv("FEOV_RELEASE_GATE") == "" || os.Getenv("CI") != "" || !sweepLockSupported {
		return nil, nil
	}
	if realCacheDirErr != nil {
		return nil, fmt.Errorf("no user cache directory for the sweep lock: %w", realCacheDirErr)
	}
	return acquireSweepLock(filepath.Join(realCacheDir, "feov", "release-fuzz.lock"), sweepLockPoll, os.Stderr)
}

// acquireSweepLock blocks until it holds the exclusive lock on path, then records this process as
// its holder in the file. While another process holds it, it writes that holder to w, again
// whenever the holder changes.
func acquireSweepLock(path string, poll time.Duration, w io.Writer) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sweep lock directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("sweep lock: %w", err)
	}
	began := time.Now()
	waited, last := false, ""
	for {
		ok, err := tryLock(f)
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("sweep lock %s: %w", path, err)
		}
		if ok {
			break
		}
		holder := sweepLockHolder(path)
		if !waited || holder != last {
			_, _ = fmt.Fprintf(w, "release fuzz: waiting for %s, held by %s (waiting since %s)\n",
				path, holder, began.Format(time.RFC3339))
		}
		waited, last = true, holder
		time.Sleep(poll)
	}
	if err := stampSweepLock(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("sweep lock %s: %w", path, err)
	}
	if waited {
		_, _ = fmt.Fprintf(w, "release fuzz: took %s after waiting %s\n", path, time.Since(began).Round(time.Second))
	}
	return f, nil
}

// stampSweepLock writes this process into the lock file it holds: pid, start time, worktree.
func stampSweepLock(f *os.File) error {
	stamp := fmt.Sprintf("pid=%d started=%s worktree=%s\n", os.Getpid(), time.Now().Format(time.RFC3339), sweepWorktree())
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.WriteAt([]byte(stamp), 0); err != nil {
		return err
	}
	return f.Sync()
}

// sweepLockHolder is the holder line the lock file carries. A holder that has taken the lock and
// not yet written itself leaves the file empty or naming its predecessor; the empty case says so.
func sweepLockHolder(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("an unreadable holder (%v)", err)
	}
	if s := strings.TrimSpace(string(b)); s != "" {
		return s
	}
	return "a process that has not yet recorded itself"
}

// sweepWorktree is the checkout this sweep runs in: the nearest directory above the working
// directory holding a .git entry (a directory in the main checkout, a file in a worktree).
func sweepWorktree() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	for d := wd; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return d
		}
		if filepath.Dir(d) == d {
			return wd
		}
	}
}
