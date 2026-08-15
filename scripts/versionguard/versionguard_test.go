package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.22.0", "0.21.0", 1},
		{"0.21.0", "0.22.0", -1},
		{"0.21.0", "0.21.0", 0},
		// The case that matters and that lexical comparison gets wrong.
		{"0.10.0", "0.9.0", 1},
		{"0.9.0", "0.10.0", -1},
		{"1.0.0", "0.99.99", 1},
		// Missing parts are zero, so these are the same version, not an ordering surprise.
		{"0.22", "0.22.0", 0},
		{"0.22.1", "0.22", 1},
	}
	for _, c := range cases {
		if got := semver(c.a, c.b); got != c.want {
			t.Errorf("semver(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// repo builds a scratch repository with one plugin at a known version.
func repo(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		full := append([]string{"-c", "user.email=t@example.com", "-c", "user.name=t", "-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", `{"name":"demo","version":"`+version+`"}`)
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x\n")
	git("init", "-q", "-b", "main")
	git("add", "-A")
	git("commit", "-q", "-m", "seed")
	git("checkout", "-q", "-b", "feature")
	return dir
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", msg}} {
		full := append([]string{"-c", "user.email=t@example.com", "-c", "user.name=t", "-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// THE FAILURE THIS GUARD EXISTS FOR, reproduced. #198 set 0.22.0; #195, branched earlier,
// carried 0.21.0 and merged on top; git resolved the version like any other line and main
// kept both features' content at a version already shipped.
func TestCatchesAVersionGoingBackwards(t *testing.T) {
	dir := repo(t, "0.22.0")
	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", `{"name":"demo","version":"0.21.0"}`)
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")
	commit(t, dir, "a branch that carries an older version")

	problems, _, checked, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if checked != 1 {
		t.Fatalf("checked = %d, want 1", checked)
	}
	if len(problems) != 1 {
		t.Fatalf("expected one problem, got %v", problems)
	}
	if !strings.Contains(problems[0], "BACKWARDS") || !strings.Contains(problems[0], "receives NOTHING") {
		t.Errorf("the message must name the consequence: %q", problems[0])
	}
}

// The rule CLAUDE.md already states: content changed, version did not.
func TestCatchesContentChangedWithoutABump(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")
	commit(t, dir, "content change, no bump")

	problems, _, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "ships nothing without a bump") {
		t.Fatalf("expected the no-bump problem, got %v", problems)
	}
}

func TestPassesWhenTheVersionMovesForward(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", `{"name":"demo","version":"0.22.0"}`)
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")
	commit(t, dir, "content change with a bump")

	problems, _, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Errorf("a correct bump must pass: %v", problems)
	}
}

// A branch that touches nothing under plugins/ needs no bump — the guard must not demand
// one for a CI or docs change, or it becomes the thing people work around.
func TestIgnoresBranchesThatTouchNoPlugin(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "README.md", "docs only\n")
	commit(t, dir, "docs only")

	problems, _, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Errorf("a non-plugin change must not require a bump: %v", problems)
	}
}

// A guard that cannot read its own input must fail loudly rather than report success.
func TestFailsLoudlyWithoutManifests(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-q", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	if _, _, _, err := check(dir, "main"); err == nil {
		t.Error("no manifests at all must be an error, not a pass")
	}
}

// THE FALSE GREEN (#423). `git diff <merge-base>...HEAD` is committed-only, while the manifest
// version is read from disk. Edit a plugin, run the local loop before committing, and `touched`
// is false for every plugin — so the version-must-move arm cannot fire and the gate prints
// "plugin versions move forward". Then CI, on a clean checkout, fails it.
//
// This is rule-sweep's defect (#409) in the second diff-based gate.
func TestUncommittedContentIsNotMeasured(t *testing.T) {
	dir := repo(t, "0.21.0")
	// Written but NOT committed — the exact moment the local loop runs.
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed in the working tree\n")

	problems, unmeasured, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Errorf("an uncommitted change is not yet a violation: %v", problems)
	}
	if len(unmeasured) != 1 {
		t.Fatalf("want the plugin reported as unmeasurable, got %v", unmeasured)
	}
	if !strings.Contains(unmeasured[0], "demo") || !strings.Contains(unmeasured[0], "UNCOMMITTED") {
		t.Errorf("the report must name the plugin and why: %q", unmeasured[0])
	}
}

// A committed violation OUTRANKS "could not measure": a dirty tree must never mask a real
// failure that the committed history already proves.
func TestCommittedViolationBeatsADirtyTree(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")
	commit(t, dir, "content change, no bump")
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed again, uncommitted\n")

	problems, unmeasured, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "version stayed at") {
		t.Fatalf("the committed violation must still FAIL: %v", problems)
	}
	if len(unmeasured) != 0 {
		t.Errorf("a plugin with a real, committed problem is measured — it must not also be reported unmeasurable: %v", unmeasured)
	}
}

// A backwards version is tree-based on both sides, so it fires whether or not anything is
// committed. Uncommitted content must not downgrade it to "not measured".
func TestBackwardsVersionFiresEvenUncommitted(t *testing.T) {
	dir := repo(t, "0.22.0")
	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", `{"name":"demo","version":"0.21.0"}`)
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")

	problems, unmeasured, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "BACKWARDS") {
		t.Fatalf("a backwards version is answerable from the tree and must FAIL: %v", problems)
	}
	if len(unmeasured) != 0 {
		t.Errorf("it was measured — do not also report it unmeasurable: %v", unmeasured)
	}
}

// A clean tree with nothing changed stays silent: the warning must not cry wolf on every run.
func TestCleanTreeIsNotReportedAsUnmeasured(t *testing.T) {
	dir := repo(t, "0.21.0")
	problems, unmeasured, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 || len(unmeasured) != 0 {
		t.Errorf("nothing changed: want silence, got problems=%v unmeasured=%v", problems, unmeasured)
	}
}
