//go:build windows

package peer

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

// ConversationBusy fails closed on Windows where Unix flock and /proc fd inspection are unavailable.
func ConversationBusy(string, string, int) (bool, string) {
	return true, "cannot verify presence lock or open conversation handles on this platform"
}

// lockRegistry serialises concurrent RegisterWith read-modify-write updates in dir using
// Windows LockFileEx on .register.lock.
func lockRegistry(dir string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(dir, ".register.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	var ol windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &ol); err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		var uol windows.Overlapped
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &uol)
		_ = f.Close()
	}, nil
}

// renameRecord replaces dst with src, retrying briefly when a concurrent reader on Windows holds
// dst open and causes MoveFileEx to return ERROR_ACCESS_DENIED or ERROR_SHARING_VIOLATION.
func renameRecord(src, dst string) error {
	var err error
	for attempt := 0; attempt < 20; attempt++ {
		err = os.Rename(src, dst)
		if err == nil {
			return nil
		}
		if !errors.Is(err, windows.ERROR_ACCESS_DENIED) && !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}
