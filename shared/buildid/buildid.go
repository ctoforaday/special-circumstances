// Package buildid answers "what IS this binary" from the binary itself.
//
// It provides build-provenance helpers shared across all plugin tools and dev tooling.
package buildid

import (
	"fmt"
	"runtime/debug"
)

// Unknown is what an unstamped binary reports. A word, not an empty string: `sc-doctor ` with a
// trailing space reads like a bug, while `sc-doctor unknown` reads like an answer.
const Unknown = "unknown"

// Revision is the short commit this binary was built from, or Unknown.
//
// Suffixed `+dirty` when the tree had uncommitted changes at build time, because "built from
// abc1234" is a false claim about a binary that also contains edits which never reached abc1234.
func Revision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return Unknown
	}
	rev, modified := "", false
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			rev = kv.Value
		case "vcs.modified":
			modified = kv.Value == "true"
		}
	}
	if rev == "" {
		return Unknown
	}
	if len(rev) > 7 {
		rev = rev[:7]
	}
	if modified {
		return rev + "+dirty"
	}
	return rev
}

// Line is the whole `-version` output for a tool: its name and the build it came from.
func Line(name string) string { return fmt.Sprintf("%s %s", name, Revision()) }
