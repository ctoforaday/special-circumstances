//go:build windows

package runlive

import (
	"io/fs"
	"os"
	"syscall"
)

// openSharingDelete opens path for reading with FILE_SHARE_DELETE, which os.Open on Windows does
// not grant. The run-live files are polled by readers (the hook guards, the served dashboard's
// teardown watcher) while their writers remove or replace them: a reader holding a handle that
// refuses delete makes the writer's os.Remove fail with a sharing violation, and the writers drop
// that error — the run's row then outlives its capture.
func openSharingDelete(path string) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: err}
	}
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(h), path), nil
}
