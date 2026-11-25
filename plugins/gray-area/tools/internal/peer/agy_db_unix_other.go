//go:build unix && !linux

package peer

// otherDBHolder fails closed on non-Linux Unix platforms where /proc/<pid>/fd is unavailable.
func otherDBHolder(_, conversation string, _ int) (bool, string) {
	return true, "cannot verify whether another process holds conversations/" + conversation + ".db open on this platform"
}

func flockHeldOnlyBy(string, int) bool { return false }
