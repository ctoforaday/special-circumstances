package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

// ONE SWEEP PER MACHINE. A local sweep spawns the real binary from NumCPU*3 workers, and two
// sessions sweeping at once on one machine do not each take twice as long — they push the load
// average past the point where every other session's `go test` times out too. So a local run
// with FEOV_RELEASE_GATE set takes an exclusive lock under the user cache before the first test
// starts, and a second one REFUSES, naming the holder.
//
// IT REFUSES RATHER THAN WAITS: `go test` kills the test process at -timeout plus one minute of
// wall clock from the process's start (cmd/go's testKillTimeout), so time spent waiting in
// TestMain is time the sweep no longer has.
//
// CI TAKES NO LOCK: a runner is one job's machine, and GitHub Actions sets GITHUB_ACTIONS. The
// lock is released when the process exits, however it exits — the kernel closes the file.

// realCacheDir is the developer's user cache, read at package initialisation: TestMain hands HOME
// to testbuild's sandbox, under which os.UserCacheDir names a temp directory no other session sees.
var realCacheDir, realCacheDirErr = os.UserCacheDir()

// heldSweepLock keeps the lock reachable for the life of the process. An *os.File nothing
// references is closed by its finalizer, and closing it releases the lock mid-sweep.
var heldSweepLock *flock.Flock

// holdSweepLockIfLocal takes the machine-wide sweep lock when this process is a LOCAL sweep: the
// gate is set and GITHUB_ACTIONS is not. It returns nil, nil when no lock applies.
func holdSweepLockIfLocal() (*flock.Flock, error) {
	if os.Getenv("FEOV_RELEASE_GATE") == "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return nil, nil
	}
	if realCacheDirErr != nil {
		return nil, fmt.Errorf("no user cache directory for the sweep lock: %w", realCacheDirErr)
	}
	return takeSweepLock(filepath.Join(realCacheDir, "feov", "release-fuzz.lock"))
}

// takeSweepLock takes the exclusive lock on path without waiting, then records this process as
// its holder. While another process holds it, it refuses, naming that holder.
func takeSweepLock(path string) (*flock.Flock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("sweep lock directory: %w", err)
	}
	fl := flock.New(path, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o644))
	ok, err := fl.TryLock()
	if err != nil {
		return nil, fmt.Errorf("sweep lock %s: %w", path, err)
	}
	if !ok {
		return nil, fmt.Errorf("another sweep holds %s: %s. One release fuzz invocation per machine at a time; run again once it has exited",
			path, sweepLockHolder(path))
	}
	if err := stampSweepLock(path); err != nil {
		_ = fl.Close()
		return nil, fmt.Errorf("sweep lock %s: %w", path, err)
	}
	return fl, nil
}

// holderPath is where the holder records itself: a file beside the lock rather than the lock file,
// because on Windows the lock is a byte-range lock no other handle may read or write.
func holderPath(lockPath string) string { return lockPath + ".holder" }

// stampSweepLock records this process as the holder: pid, start time, worktree.
func stampSweepLock(lockPath string) error {
	stamp := fmt.Sprintf("pid=%d started=%s worktree=%s\n", os.Getpid(), time.Now().Format(time.RFC3339), sweepWorktree())
	return os.WriteFile(holderPath(lockPath), []byte(stamp), 0o644)
}

// sweepLockHolder is the holder line beside the lock. A holder that has taken the lock and not yet
// written itself leaves that file absent or naming its predecessor; the absent case says so.
func sweepLockHolder(lockPath string) string {
	b, err := os.ReadFile(holderPath(lockPath))
	if err != nil && !os.IsNotExist(err) {
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
