package pushfreezeguard

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
)

// tokenize had no direct test: every assertion about it went through decide(), which
// cannot see the difference between `&&` as one token and `&&` as two `&` tokens —
// both are separators, so the verdict is identical either way. A mutation sweep
// confirmed it, flipping the separator-doubling condition with the suite green.
//
// The doc comment promises the separators come out as their OWN tokens, and that
// promise is what the argument-list logic in gitArgs is built on. It is asserted here,
// at the unit, rather than inferred from a downstream verdict that happens to agree.
func TestTokenize(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"git push", []string{"git", "push"}},
		{"git add -A && git commit", []string{"git", "add", "-A", "&&", "git", "commit"}},
		{"git add -A&&git commit", []string{"git", "add", "-A", "&&", "git", "commit"}},
		{"a||b", []string{"a", "||", "b"}},
		{"a|b", []string{"a", "|", "b"}},
		{"a&b", []string{"a", "&", "b"}},
		{"a;b", []string{"a", ";", "b"}},
		// `;;` is two separators, not one doubled token — only & and | double.
		{"a;;b", []string{"a", ";", ";", "b"}},
		{"a & & b", []string{"a", "&", "&", "b"}},
		{"git\tstatus\nls", []string{"git", "status", "ls"}},
		{"", nil},
		{"   ", nil},
		{"&&", []string{"&&"}},
		{"|||", []string{"||", "|"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := tokenize(c.in)
			if len(got) != len(c.want) {
				t.Fatalf("tokenize(%q) = %q, want %q", c.in, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("tokenize(%q) = %q, want %q", c.in, got, c.want)
				}
			}
		})
	}
}

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

		// A git command with global options and NO VERB AT ALL. This reaches the verb
		// search with j == len(tokens), and it was never exercised: a mutation sweep
		// turned `j >= len(t)` into `j > len(t)` — an index-out-of-range on exactly
		// this input — with the suite still green. A PreToolUse hook that panics is a
		// hook whose stderr becomes noise on every single Bash call.
		{"global option with no verb does not panic", live, "git -C repo", false, ""},
		{"bare git does not panic", live, "git", false, ""},
		{"trailing value-flag with no verb does not panic", live, "cd x && git --git-dir=/y -C", false, ""},
		{"git as the last token before a separator", live, "git && ls", false, ""},
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
