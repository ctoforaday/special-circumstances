// Package buildid answers "what IS this binary" from the binary itself.
//
// Replaces `const version = "0.x.0"`, hand-written at birth and never moved again, which answered
// `-version` with a number that had never been true of anything shipped — and answered it
// confidently, which is worse than refusing (#405 item 3).
//
// The Go toolchain already stamps `vcs.revision` into every binary built inside a git tree, with
// no flags; established and verified under the release job's own cross-compile flags in #450. The
// truthful identity was already inside every binary while a hand-written constant sat beside it
// saying something else.
//
// # Why this is a second copy rather than a shared package
//
// gray-area is its own Go module, by design: each plugin ships and releases independently, so it
// cannot import prosthetic-conscience's internal packages. That makes this duplicated CODE and not
// a duplicated FACT — each copy reads the stamp of the binary it is compiled into, so there is no
// shared value that could drift between them. The [[facts-are-fields]] objection is to two
// carriers of ONE fact; this is two readers of two different ones.
package buildid

import (
	"fmt"
	"runtime/debug"
)

// Unknown is what an unstamped binary reports. A word, not an empty string: `gray-area ` with a
// trailing space reads like a bug, while `gray-area unknown` reads like an answer.
const Unknown = "unknown"

// Revision is the short commit this binary was built from, or Unknown. Suffixed `+dirty` when the
// tree carried uncommitted changes, because "built from abc1234" is a false claim about a binary
// that also contains edits which never reached abc1234.
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
