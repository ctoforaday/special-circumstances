//go:build !windows

package runlive

import (
	"os"
	"syscall"
)

// lockFile takes an flock(2) lock on f — the primitive frank-exchange-of-views's writer takes on
// the same file (gofrs/flock), so the two interlock.
func lockFile(f *os.File, exclusive bool) error {
	how := syscall.LOCK_SH
	if exclusive {
		how = syscall.LOCK_EX
	}
	for {
		err := syscall.Flock(int(f.Fd()), how)
		if err != syscall.EINTR {
			return err
		}
	}
}
