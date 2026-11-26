//go:build unix && !linux

package peer

import "errors"

// otherDBHolder fails closed on non-Linux Unix platforms where /proc/<pid>/fd is unavailable.
func otherDBHolder(_, conversation string, _ int) Holder {
	return Holder{Kind: HolderUninspectable, Detail: "whether another process has conversations/" + conversation + ".db open cannot be inspected on this platform"}
}

// flockHolder cannot read a lock's holder where /proc/locks is unavailable.
func flockHolder(string, int) (int, error) {
	return 0, errors.New("/proc/locks is not available on this platform")
}
