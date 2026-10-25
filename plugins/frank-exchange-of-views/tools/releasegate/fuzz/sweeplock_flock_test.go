//go:build linux || darwin

package fuzz

import (
	"errors"
	"os"
	"syscall"
)

const sweepLockSupported = true

// tryLock takes an exclusive flock on f without blocking, and reports false while another
// process holds it.
func tryLock(f *os.File) (bool, error) {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, syscall.EWOULDBLOCK), errors.Is(err, syscall.EINTR):
		return false, nil
	}
	return false, err
}
