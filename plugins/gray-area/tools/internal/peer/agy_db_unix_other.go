//go:build unix && !linux

package peer

import (
	"errors"
	"os"
)

// otherDBHolder fails closed on non-Linux Unix platforms where /proc/<pid>/fd is unavailable.
func otherDBHolder(_, conversation string, _ int) Holder {
	return Holder{Kind: HolderUninspectable, Detail: "whether another process has conversations/" + conversation + ".db open cannot be inspected on this platform"}
}

// flockHolders cannot read a lock's holders where /proc/locks is unavailable; a lock file that does
// not exist is still reported as such.
func flockHolders(path string) ([]int, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return nil, errors.New("/proc/locks is not available on this platform")
}
