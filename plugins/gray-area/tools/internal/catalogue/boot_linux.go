//go:build linux

package catalogue

import (
	"os"
	"strconv"
	"strings"
)

// BootTime is when this host last booted, in Unix seconds, from the `btime` line of /proc/stat.
//
// ok is false whenever the line cannot be read — no file, no such line, a value that is not a
// positive integer. Never a zero taken for a boot at the epoch: every session would then read as
// running since after the boot, and a restart would cut off nothing, which is the plausible zero
// `telepathy agents --lost` must not print.
func BootTime() (int64, bool) { return bootFrom("/proc/stat") }

func bootFrom(path string) (int64, bool) {
	body, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(body), "\n") {
		f := strings.Fields(line)
		if len(f) != 2 || f[0] != "btime" {
			continue
		}
		n, err := strconv.ParseInt(f[1], 10, 64)
		if err != nil || n <= 0 {
			return 0, false
		}
		return n, true
	}
	return 0, false
}
