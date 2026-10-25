//go:build !(linux || darwin)

package fuzz

import "os"

// No sweep lock off Linux and macOS: the lock is a flock, and a sweep elsewhere runs unqueued.
const sweepLockSupported = false

func tryLock(*os.File) (bool, error) { return true, nil }
