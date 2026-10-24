package runlive

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// MarkerLock is the marker's lock file, beside it in .claude. Writers hold it exclusive from the
// read to the rename; readers hold it shared across the open and the read.
//
// A RENAME IS NOT ATOMIC TO A CONCURRENT OPEN ON WINDOWS. replaceFile renames a complete file over
// the marker, so a reader never sees torn bytes — but an open that lands while NTFS replaces the
// name finds NO file: ERROR_FILE_NOT_FOUND on 13 to 879 opens per run of a few thousand
// replacements on the Windows runner, with os.Root.Rename's POSIX rename and a full-path
// FileRenameInfoEx rename alike (hooks.yml runs 40088161661, 40109682375, 40110811684,
// 40112622361); os.Rename does not replace a held marker at all (#1555). Go's own toolchain met the same thing (golang.org/issue/32188: "os.Rename is not
// actually atomic on Windows under load") and guards go.mod with a lock rather than a rename
// (cmd/go/internal/lockedfile). A reader that misses the marker reports no live run: the dashboard
// watcher tears down a live run's server and the hooks attribute nothing. Under the lock no open
// overlaps a rename, on any platform.
//
// The exclusive hold also makes each writer's read-modify-write one step: setup and capture
// writing at once can no longer drop each other's row.
//
// The lock file is created by the first writer and never removed; a reader never creates it.
const MarkerLock = "run-live.json.lock"

func markerLockPath(projectDir string) string {
	return filepath.Join(projectDir, ".claude", MarkerLock)
}

// lockMarker takes the marker's lock exclusive for a writer. A project with no .claude directory
// has no marker to change, so there is nothing to lock and the release is a no-op.
func lockMarker(projectDir string) (release func() error, err error) {
	fl := flock.New(markerLockPath(projectDir), flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o644))
	if err := fl.Lock(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return func() error { return nil }, nil
		}
		return nil, err
	}
	return fl.Close, nil
}

// rlockMarker takes the marker's lock shared for a reader. No lock file means no writer has
// written the marker, so there is no rename to wait out.
func rlockMarker(projectDir string) (release func() error, err error) {
	fl := flock.New(markerLockPath(projectDir), flock.SetFlag(os.O_RDONLY))
	if err := fl.RLock(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return func() error { return nil }, nil
		}
		return nil, err
	}
	return fl.Close, nil
}
