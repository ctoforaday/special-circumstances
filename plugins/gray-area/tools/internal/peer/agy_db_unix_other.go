//go:build unix && !linux

package peer

func otherAgyDBHolder(string, string, int) (int, bool) { return 0, false }
