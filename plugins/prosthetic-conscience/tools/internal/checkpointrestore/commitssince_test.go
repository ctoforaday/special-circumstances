package checkpointrestore

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

// git runs a command in dir and fails the test on error, because a fixture that
// half-built produces a count that looks like a finding.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(cmd.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// mergeRepo builds the one shape that distinguishes the two counts:
//
//	A ── C ── M   (first parent: this line)
//	 \       /
//	  ─── B ─
//
// From A: three commits are REACHABLE (B, C, M) and two are on this branch's own
// line (C, M). A count that does not say which it means is a count of somebody
// else's work as often as your own.
func mergeRepo(t *testing.T) (dir, base string) {
	t.Helper()
	dir = t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "A")
	base = git(t, dir, "rev-parse", "--short=7", "HEAD")

	git(t, dir, "checkout", "-q", "-b", "side")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "B")

	git(t, dir, "checkout", "-q", "main")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "C")
	git(t, dir, "merge", "-q", "--no-ff", "-m", "M", "side")
	return dir, base
}

// The defect this fixes: a note written before a routine merge was reported as
// stale by every commit that arrived in it, most of which the session never did.
func TestCommitsSinceCountsThisBranchsLineNotEverythingReachable(t *testing.T) {
	dir, base := mergeRepo(t)
	t.Chdir(dir)

	reachable := git(t, dir, "rev-list", "--count", base+"..HEAD")
	firstParent := git(t, dir, "rev-list", "--count", "--first-parent", base+"..HEAD")
	if reachable == firstParent {
		t.Fatalf("fixture does not distinguish the counts: both %s — the merge did not take", reachable)
	}

	n, ok := commitsSince(base)
	if !ok {
		t.Fatalf("commitsSince(%s) unreachable in a repo that contains it", base)
	}
	if got, want := n, 2; got != want {
		t.Errorf("commitsSince = %d, want %d (first-parent). Reachable-but-not-ours: %s vs %s — "+
			"counting everything reachable reports another branch's work as this session's staleness",
			got, want, reachable, firstParent)
	}
}

// Unreachable must stay a REPORTED state rather than a zero: "written 0 commits
// ago" and "that commit is not in this repo" are different claims, and the
// restore hook must never fail over provenance either way.
func TestCommitsSinceReportsUnreachableRatherThanZero(t *testing.T) {
	dir, _ := mergeRepo(t)
	t.Chdir(dir)

	if n, ok := commitsSince("0000000"); ok {
		t.Errorf("commitsSince(unknown ref) = (%d, true); want reachable=false", n)
	}
}

// Schema 3 splits `updated:` — "something happened" — into two facts that mean
// different things. The digest must render the one that answers "how old is this
// note", and must say when a note was only RE-AFFIRMED, because that is the case
// where the file moved and the content did not.
func TestDigestRendersWrittenAtAndReaffirmedAt(t *testing.T) {
	const n3 = `---
schema: 3
written_at: 2026-08-23T04:00:00Z
reaffirmed_at: 2026-08-23T06:30:00Z
session_id: sess-abc
objective: "prove the provenance line speaks schema 3"
---
## Next intended steps
1. anything
`
	got := digest(n3, "CHECKPOINT.md", "startup", checkpoint.RearmState{})
	if !strings.Contains(got, "written: 2026-08-23T04:00:00Z") {
		t.Errorf("digest does not carry written_at:\n%s", got)
	}
	if !strings.Contains(got, "reaffirmed: 2026-08-23T06:30:00Z") {
		t.Errorf("digest drops reaffirmed_at, so a note that was touched but not changed "+
			"reads exactly like one that was rewritten:\n%s", got)
	}
}

// A note with no re-affirmation must not grow an empty field: absent is a state, and
// rendering "reaffirmed: " for it spends the budget on nothing.
func TestDigestOmitsReaffirmedWhenAbsent(t *testing.T) {
	const n3 = `---
schema: 3
written_at: 2026-08-23T04:00:00Z
objective: "no re-affirmation here"
---
## Next intended steps
1. anything
`
	if got := digest(n3, "CHECKPOINT.md", "startup", checkpoint.RearmState{}); strings.Contains(got, "reaffirmed") {
		t.Errorf("digest rendered a re-affirmation that never happened:\n%s", got)
	}
}
