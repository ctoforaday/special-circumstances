package buildid

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// StampedFrom reports the working tree whose revision the Go toolchain stamps into a binary
// built from dir, and whether that could be determined.
//
// RESTATED, NOT SHARED. prosthetic-conscience carries the same function, and the two plugins
// are separate Go modules by design — a consumer must be able to take one without the other.
// The duplication is the price of that boundary, and it is the same trade already made for the
// checkpoint-note parser in internal/claims.
//
// It is not always dir: `go build` resolves VCS state through `.git`, which in a WORKTREE is a
// file pointing at `<main>/.git/worktrees/<name>`. The revision and the dirty flag both come
// from the MAIN checkout, so on a worktree whose HEAD has diverged the stamp disagrees with the
// local HEAD for a reason that has nothing to do with staleness (#532).
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
	// --git-common-dir names the .git DIRECTORY; the checkout is its parent. In a plain clone
	// this resolves back to dir itself, so callers need no special case.
	return filepath.Dir(p), true
}
