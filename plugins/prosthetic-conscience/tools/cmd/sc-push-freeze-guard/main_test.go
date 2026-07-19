package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
)

func TestDecide(t *testing.T) {
	live := &runlive.Marker{RunDir: "research/x", PinnedPaths: []string{"ideas/backlog.md", "research/old"}, Started: "2026-07-16T00:00:00Z"}
	cases := []struct {
		name    string
		m       *runlive.Marker
		cmd     string
		wantHit bool
		wantSub string // required substring of the warning when wantHit
	}{
		{"no marker is silent even on push", nil, "git push origin main", false, ""},
		{"live + push warns with the frozen pins", live, "cd repo && git push", true, "FROZEN"},
		{"live + non-push is silent", live, "git status", false, ""},
		{"live + commit without push is silent", live, "git commit -m x", false, ""},
		// W1.13 — the 2026-07-17 incident classes.
		{"live + add -A warns", live, "git add -A && git commit -m x", true, "sweeping"},
		{"live + add . warns", live, "cd repo && git add .", true, "sweeping"},
		{"live + add --all warns", live, "git add --all", true, "sweeping"},
		{"live + explicit-path add is silent", live, "git add ideas/backlog.md plans/x.md", false, ""},
		{"live + checkout branch warns", live, "git checkout main", true, "worktree"},
		{"live + checkout -b is silent (no working-tree deletion)", live, "git checkout -b feat/x", false, ""},
		{"live + switch warns", live, "git switch main", true, "worktree"},
		{"live + switch -c is silent", live, "git switch -c feat/x", false, ""},
		{"live + stash -u warns", live, "git stash -u", true, "untracked"},
		{"live + stash --include-untracked warns", live, "git stash push --include-untracked -m x", true, "untracked"},
		{"live + stash pop warns on clobber risk", live, "git stash pop", true, "never pop"},
		{"live + plain stash push is silent (tracked-only)", live, "git stash push -m wip", false, ""},
		{"no marker is silent on add -A", nil, "git add -A", false, ""},
		{"flag in a LATER command never leaks into the git verb", live, "git add file.md && rm -A", false, ""},

		// DEFECT 1 — global git options hid the verb from the guard. `git -C repo
		// checkout main` deletes the untracked blackboard exactly like the bare form,
		// but the old positional check only ever looked at the token right after `git`.
		{"live + -C checkout warns", live, "git -C repo checkout main", true, "worktree"},
		{"live + -C add -A warns", live, "git -C repo add -A", true, "sweeping"},
		{"live + -C push warns", live, "git -C repo push origin main", true, "FROZEN"},
		{"live + -c option checkout warns", live, "git -c core.pager=cat checkout main", true, "worktree"},
		{"live + --no-pager switch warns", live, "git --no-pager switch main", true, "worktree"},
		{"live + --git-dir stash -u warns", live, "git --git-dir=/x stash -u", true, "untracked"},
		{"-C value is never mistaken for the verb", live, "git -C checkout status", false, ""},
		{"-b still exempt behind a global option", live, "git -C repo checkout -b feat/x", false, ""},

		// DEFECT 2 — shell separators had to be space-padded to be seen.
		{"unpadded semicolon still warns", live, "git add -A;git commit -m x", true, "sweeping"},
		{"unpadded && still warns", live, "cd repo&&git checkout main", true, "worktree"},
		{"unpadded separator no longer leaks a later flag in", live, "git add file.md&&rm -A", false, ""},
		{"pipe ends the git verb's arguments", live, "git add file.md|grep -A 2 x", false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := decide(c.m, c.cmd)
			if (got != "") != c.wantHit {
				t.Fatalf("decide(%v, %q) = %q; wantHit=%v", c.m != nil, c.cmd, got, c.wantHit)
			}
			if c.wantHit {
				if !strings.Contains(got, "research/x") || !strings.Contains(got, c.wantSub) {
					t.Fatalf("warning missing run dir or %q: %q", c.wantSub, got)
				}
			}
		})
	}
}
