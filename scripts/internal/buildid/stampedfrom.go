package buildid

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// StampedFrom reports the working tree whose revision the Go toolchain stamps into a
// binary built from dir, and whether that could be determined.
//
// IT IS NOT ALWAYS dir. `go build` resolves VCS state through `.git`, and in a WORKTREE
// that is a file pointing at `<main>/.git/worktrees/<name>` — so the revision and the dirty
// flag both come from the MAIN checkout, not from the tree the source was read from. On a
// worktree whose HEAD has diverged (which is every worktree, by definition) the stamp
// disagrees with the local HEAD for a reason that has nothing to do with staleness.
//
// That cost this repository a triage cycle per session: CLAUDE.md puts work in worktrees and
// the harness creates one per session, so the default place this code is developed was the
// one place two provenance tests could not pass. A red meaning "not applicable here" is how
// a suite teaches people to skim past reds (#532).
//
// This resolves the question rather than dodging it, so the assertion states what the
// toolchain actually promises: the binary reports the revision Go read, and Go read it from
// HERE.
func StampedFrom(dir string) (string, bool) {
	common, err := exec.Command("git", "-C", dir, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", false
	}
	p := strings.TrimSpace(string(common))
	if p == "" {
		return "", false
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(dir, p)
	}
	// --git-common-dir names the .git DIRECTORY; the checkout is its parent. In a plain
	// clone this resolves back to dir itself, so callers need no special case.
	return filepath.Dir(p), true
}
