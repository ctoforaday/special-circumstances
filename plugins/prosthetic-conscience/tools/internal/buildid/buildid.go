// Package buildid answers "what IS this binary" from the binary itself.
//
// # Why the constants it replaces were worse than useless (#405 item 3)
//
// Every hook binary carried `const version = "0.1.0"`, hand-written at birth and never moved
// again. Thirteen of them, eleven still reading 0.1.0 after a year of change. `-version` therefore
// answered with a number that had never been true of anything shipped — and it answered
// CONFIDENTLY, which is worse than refusing, because the reader has no way to tell a frozen
// constant from a maintained one.
//
// `versionguard` counted them on every run and printed the total as "tolerated", precisely so that
// tolerating them could not decay into denying they existed.
//
// # What replaces them, and why nothing had to be added to any build
//
// The Go toolchain already stamps `vcs.revision` into every binary built inside a git tree, with
// no flags — established and verified in #450, including under the release job's
// `-trimpath -ldflags "-s -w"` cross-compile. So the truthful identity was already inside every
// binary; thirteen hand-written constants sat next to it saying something else.
//
// This reads the stamp of the RUNNING binary (`debug.ReadBuildInfo`), where #450's sibling reads
// the stamp of a binary ON DISK (`debug/buildinfo.ReadFile`). Same fact, two vantage points: a
// binary reporting on itself, and the doctor reporting on the ones it found.
//
// # An unstamped build says so
//
// A binary built outside a git tree carries no revision, and `Line` says `unknown` rather than
// inventing a number. That is the [[facts-are-fields]] clause-3 shape: the honest answer to "which
// build is this" must be distinguishable from an answer, and a fabricated semver is exactly the
// confident-wrong output this change removes.
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
