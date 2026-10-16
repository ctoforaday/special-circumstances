//go:build unix

package migrate

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// dirSyncs: a directory can be fsynced here. On macOS a directory's Sync is F_FULLFSYNC, whose
// semantics for a directory are undocumented and which no CI leg runs; a failure is a warning.
const dirSyncs = true

// fsyncPath opens path read-only and fsyncs it; for a directory that makes its entries durable.
func fsyncPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// publishRename is rename(2) itself, retried on EINTR as os.Rename does internally. Not os.Rename:
// it Lstats the destination and refuses any existing directory before the syscall (os/file_unix.go,
// go1.27.1), while rename(2) replaces an EMPTY directory atomically — so --to is never absent and
// nothing removes it first — and refuses a non-empty one (ENOTEMPTY) or a file (ENOTDIR) without
// deleting anything.
func publishRename(from, to string) error {
	for {
		err := syscall.Rename(from, to)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return &os.LinkError{Op: "rename", Old: from, New: to, Err: err}
		}
		return nil
	}
}

// sameDevice reports whether a and b are on one device — false for a mount point and its parent.
func sameDevice(a, b string) (bool, error) {
	var sa, sb syscall.Stat_t
	if err := syscall.Lstat(a, &sa); err != nil {
		return false, fmt.Errorf("reading the device of %s: %w", a, err)
	}
	if err := syscall.Stat(b, &sb); err != nil {
		return false, fmt.Errorf("reading the device of %s: %w", b, err)
	}
	return sa.Dev == sb.Dev, nil
}
