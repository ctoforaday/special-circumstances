package doctor

// A STALE HOOK BINARY IS SILENT, WHERE AN EMPTY ONE IS LOUD (#450).
//
// The empty-bin case already announces itself: no binary means every hook event spawn-fails, the
// crash storm is unmissable, and a bootstrap guard turns it into a pointer at `doctor --fix`.
//
// The STALE case has none of that. A binary built weeks ago runs, exits 0, and does exactly what it
// did then. There is no crash, no warning, and no output that differs from a healthy hook — so the
// absent fix and the applied fix are the same bytes, which is the plausible-zero shape this suite
// keeps finding, here in its own deployment path.
//
// MEASURED 2026-08-17: `sc-filechanged-rearm` was running from a binary built on 2026-08-01 while
// its source had moved 23 commits. The re-arm keying change in #449 had merged and was simply not
// running. It was noticed only because a migration had been predicted to a specific number and the
// number did not move; with no prediction, "nothing changed" would have read as "nothing to change".
//
// # The fix needed no new mechanism, and that is the point
//
// The obvious proposal — and the one this issue was filed recommending — was to stamp the build with
// `-ldflags -X`. That would have been a fact composed at build time by FIVE hand-written command
// lines: the release workflow, `scripts/bootstrap-plugins.sh`, two in `docs/setup-script.md`, and a
// command doc that humans copy-paste. Guarding five hand-written copies of one string is precisely
// what [[facts-are-fields]] says to prefer GENERATION over.
//
// Generation had already happened. The Go toolchain stamps `vcs.revision`, `vcs.time` and
// `vcs.modified` into every binary built inside a git tree, with no flags at all — and it survives
// the release job's `-trimpath -ldflags "-s -w"` cross-compile, verified. The existing binaries on
// disk, built weeks before this file, already carried it. Nothing wrote the fact down because
// nothing needed to; nothing READ it, which is the whole defect.
//
// So this package reads what is already there. There is no build-path change in it, and no carrier
// to keep in step.

import (
	"debug/buildinfo"
	"fmt"
	"os/exec"
	"strings"
)

// HeadCommit is the commit a binary in this tree should have been built from.
//
// Returns "" rather than an error on any failure — outside a checkout, without git on PATH, in a
// shallow copy. Compare treats "" as UNANSWERABLE rather than as clean, so a doctor run in a
// non-git environment reports that it could not judge instead of certifying every binary current.
func HeadCommit(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// BuildStamp is what a binary on disk says about its own provenance.
//
// Absent is a distinct state from stale and is reported as itself: a binary built outside a git
// tree, or by a toolchain with `-buildvcs=false`, carries no revision. Folding that into "stale"
// would be a guess presented as a measurement.
type BuildStamp struct {
	Revision string // full commit sha, "" when the binary carries none
	Modified bool   // the tree had uncommitted changes at build time
	Time     string // RFC3339 commit time, "" when absent
}

// Known reports whether the stamp says anything at all.
func (b BuildStamp) Known() bool { return b.Revision != "" }

// ReadBuildStamp reads one binary's VCS stamp. A binary that cannot be read at all is not an
// error here — the caller already knows whether it exists, and `Built: false` is that report.
func ReadBuildStamp(path string) BuildStamp {
	info, err := buildinfo.ReadFile(path)
	if err != nil || info == nil {
		return BuildStamp{}
	}
	var s BuildStamp
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			s.Revision = kv.Value
		case "vcs.time":
			s.Time = kv.Value
		case "vcs.modified":
			s.Modified = kv.Value == "true"
		}
	}
	return s
}

// Staleness is the verdict for one binary against the tree it should have been built from.
type Staleness string

const (
	// Current: the binary was built from exactly this commit.
	Current Staleness = "current"
	// Stale: built from a DIFFERENT commit. The source has moved and this binary has not.
	Stale Staleness = "stale"
	// Dirty: built from this commit but with uncommitted changes in the tree, so what it
	// contains cannot be recovered from the commit alone.
	Dirty Staleness = "dirty"
	// Unstamped: the binary carries no revision. NOT the same as current, and reported as its
	// own word — an unanswerable question must not fold into the healthy answer.
	Unstamped Staleness = "unstamped"
	// NoReference: the binary IS stamped and nothing here says which commit it should carry.
	//
	// A SEPARATE WORD BECAUSE IT IS A SEPARATE FACT, and the one it used to share a word with
	// pointed at the wrong party. An installed plugin lives in a cache directory, which is not
	// a git checkout, so HeadCommit returned "" for every binary in every real installation —
	// and the verdict line reported that as "carry no build stamp". The binaries were stamped
	// throughout. What was missing was the reference, which is the doctor's end of the
	// comparison, not the binary's.
	NoReference Staleness = "no-reference"
)

// Compare judges a stamp against the current HEAD.
//
// `head` empty means the caller could not determine the commit — outside a git checkout, say. That
// is unanswerable rather than clean, so it returns Unstamped for the same reason a missing stamp
// does: the comparison did not happen, and saying "current" would report a check that never ran.
func Compare(s BuildStamp, head string) Staleness {
	if !s.Known() {
		return Unstamped
	}
	if head == "" {
		return NoReference
	}
	if s.Revision != head {
		return Stale
	}
	if s.Modified {
		return Dirty
	}
	return Current
}

// Describe is the operator-facing line for a verdict, naming the remedy where there is one.
//
// The stale message says WHAT to run, because the whole failure mode is that nothing about a stale
// binary suggests a rebuild is needed.
func Describe(v Staleness, s BuildStamp, head string) string {
	switch v {
	case Stale:
		return fmt.Sprintf("built from %s, tree is at %s — this binary predates the source it ships with; run `doctor --fix` to rebuild",
			short(s.Revision), short(head))
	case Dirty:
		return fmt.Sprintf("built from %s with uncommitted changes — what it contains is not recoverable from that commit", short(s.Revision))
	case Unstamped:
		return "carries no VCS stamp — built outside a git tree or with -buildvcs=false, so staleness cannot be judged"
	default:
		return fmt.Sprintf("built from %s", short(s.Revision))
	}
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	if sha == "" {
		return "(none)"
	}
	return sha
}
