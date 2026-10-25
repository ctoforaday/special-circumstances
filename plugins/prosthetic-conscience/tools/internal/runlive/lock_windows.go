//go:build windows

package runlive

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockFile takes a LockFileEx lock on f's first byte — the range frank-exchange-of-views's writer
// locks on the same file (gofrs/flock), so the two interlock.
func lockFile(f *os.File, exclusive bool) error {
	var flags uint32
	if exclusive {
		flags = windows.LOCKFILE_EXCLUSIVE_LOCK
	}
	return windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, 1, 0, &windows.Overlapped{})
}
