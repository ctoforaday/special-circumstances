package migrate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

// dirSyncs: Windows has no directory fsync, and the rename is not write-through. A power loss can
// lose the publish's directory entries; that is a stated residue, not a mechanism.
const dirSyncs = false

// fsyncPath opens a regular file read-write — FlushFileBuffers needs write access — and syncs it.
func fsyncPath(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// publishRename removes an existing empty --to and renames the build onto it, since Windows cannot
// rename over a directory. RemoveDirectory removes only an empty directory, so a --to that became a
// file or gained an entry since it was checked is refused, not deleted. The raw call does not get
// os's long-path handling, so a --to of 248 characters or more fails here after the build, the
// build kept and --to intact. Between the two calls --to is absent; a rename that fails after the
// removal leaves it absent, and the error says so.
func publishRename(from, to string) error {
	removed := false
	if _, err := os.Lstat(to); err == nil {
		p, err := syscall.UTF16PtrFromString(to)
		if err != nil {
			return err
		}
		if err := syscall.RemoveDirectory(p); err != nil {
			return &os.PathError{Op: "remove", Path: to, Err: err}
		}
		removed = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Rename(from, to); err != nil {
		if removed {
			return fmt.Errorf("%w — the empty %s was removed to make way and is now absent", err, to)
		}
		return err
	}
	return nil
}

// sameDevice is always true here: a junction or mount point at --to is refused by its Lstat type.
func sameDevice(a, b string) (bool, error) { return true, nil }
