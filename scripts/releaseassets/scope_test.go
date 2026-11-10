package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// prRepo is a repository whose HEAD is a pull request's merge commit: main's tip merged with a
// branch that changes the file named.
func prRepo(t *testing.T, changed string) string {
	t.Helper()
	wf, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "hooks.yml"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	writeFile(t, filepath.Join(dir, ".github", "workflows", "hooks.yml"), string(wf))
	writeFile(t, filepath.Join(dir, "plugins", "frank-exchange-of-views", "native", "feov-ocr", "ASSETS"), "")
	writeFile(t, filepath.Join(dir, "plugins", "frank-exchange-of-views", "native", "feov-ocr", "main.cpp"), "int main() {}\n")
	writeFile(t, filepath.Join(dir, "README.md"), "base\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "base")
	git(t, dir, "checkout", "-q", "-b", "pr")
	writeFile(t, filepath.Join(dir, changed), "changed\n")
	git(t, dir, "commit", "-q", "-am", "the pull request")
	git(t, dir, "checkout", "-q", "main")
	writeFile(t, filepath.Join(dir, "other.txt"), "main moved\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "main moves on")
	git(t, dir, "merge", "-q", "--no-ff", "-m", "merge", "pr")
	return dir
}

func scopeTargets(t *testing.T, root string) []Target {
	t.Helper()
	pairs, err := ReadReleasePairs(filepath.Join(root, ".github", "workflows", "hooks.yml"))
	if err != nil {
		t.Fatal(err)
	}
	targets, err := ParseTargets(pairs)
	if err != nil {
		t.Fatal(err)
	}
	return targets
}

// TestAPullRequestScopeReadsItsMergeCommit drives Scope, the path native-scope runs, over a real
// merge commit: a native-set change runs everything, any other change does not, and the change
// list is the merge's diff against the base tip, never main's own later commits.
func TestAPullRequestScopeReadsItsMergeCommit(t *testing.T) {
	for _, tc := range []struct {
		changed string
		full    string
	}{
		{"plugins/frank-exchange-of-views/native/feov-ocr/main.cpp", "full: true"},
		{"README.md", "full: false"},
	} {
		t.Run(tc.changed, func(t *testing.T) {
			root := prRepo(t, tc.changed)
			var out bytes.Buffer
			ghOut := filepath.Join(t.TempDir(), "out")
			if err := Scope(root, "pull_request", "refs/pull/1/merge", repoNativeSet(t), scopeTargets(t, root), &out, ghOut); err != nil {
				t.Fatalf("Scope: %v", err)
			}
			if !strings.Contains(out.String(), tc.full+"\n") || !strings.Contains(out.String(), "changed files: 1,") {
				t.Fatalf("a pull request changing %s decided:\n%s\nwant %q from one changed file", tc.changed, out.String(), tc.full)
			}
			b, err := os.ReadFile(ghOut)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(b), strings.Replace(tc.full, ": ", "=", 1)+"\n") {
				t.Errorf("GITHUB_OUTPUT holds:\n%s", b)
			}
		})
	}
}

// TestAnUnreadablePullRequestIsRefused: a HEAD that is not the merge commit, and a merge commit
// whose base was not fetched, fail loudly instead of deciding from an empty change list.
func TestAnUnreadablePullRequestIsRefused(t *testing.T) {
	root := prRepo(t, "plugins/frank-exchange-of-views/native/feov-ocr/main.cpp")
	targets := scopeTargets(t, root)

	shallow := filepath.Join(t.TempDir(), "shallow")
	git(t, filepath.Dir(shallow), "clone", "-q", "--depth", "1", "file://"+root, shallow)
	if err := Scope(shallow, "pull_request", "refs/pull/1/merge", repoNativeSet(t), targets, &bytes.Buffer{}, ""); err == nil || !strings.Contains(err.Error(), "merge commit") {
		t.Errorf("a merge commit without its parents was scoped: %v", err)
	}

	git(t, root, "checkout", "-q", "pr")
	if err := Scope(root, "pull_request", "refs/pull/1/merge", repoNativeSet(t), targets, &bytes.Buffer{}, ""); err == nil || !strings.Contains(err.Error(), "merge commit") {
		t.Errorf("a HEAD that is not a merge commit was scoped: %v", err)
	}
}
