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
		got, err := semver(c.a, c.b)
		if err != nil {
			t.Errorf("semver(%q,%q) errored on a well-formed pair: %v", c.a, c.b, err)
			continue
		}
		if got != c.want {
			t.Errorf("semver(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// A VERSION THIS GATE CANNOT READ MUST NOT COMPARE AS ZERO.
//
// strconv.Atoi's discarded error made "0.2x.0" parse as 0.0.0, and the direction that
// matters is a malformed BASE: zero is lower than every head, so every branch looked like a
// forward move and the backwards-guard passed without comparing anything. The gate reported
// the same clean board whether it had checked or not.
func TestAnUnreadableVersionIsRefusedNotCoercedToZero(t *testing.T) {
	for _, bad := range []string{"0.2x.0", "v0.22.0", "0.22.0-rc1", "abc", "", "0.22.0.1", "0..0", "-1.0.0"} {
		if _, err := semver(bad, "0.22.0"); err == nil {
			t.Errorf("semver(%q, …) was accepted; an unreadable version must be refused, not read as zero", bad)
		}
		if _, err := semver("0.22.0", bad); err == nil {
			t.Errorf("semver(…, %q) was accepted; the BASE side is the dangerous one — it reads as lower than everything", bad)
		}
	}
	// The deliberate tolerance survives: a SHORT version is not a malformed one.
	if _, err := semver("0.22", "0.22.0"); err != nil {
		t.Errorf("missing parts must still count as zero: %v", err)
	}
}

// The refusal has to name the version and the part, or it sends someone reading three
// manifests to find which one this gate could not order.
func TestTheRefusalNamesTheOffendingPart(t *testing.T) {
	_, err := semver("0.2x.0", "0.22.0")
	if err == nil {
		t.Fatal("want a refusal")
	}
	for _, want := range []string{"0.2x.0", "part 2", "2x"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q is missing %q", err, want)
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

	problems, checked, err := check(dir, "main")
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
// THE RULE THIS REPLACED, INVERTED (#405). Content changing without a bump used to FAIL. It
// is now the ordinary case: versions move at a release boundary, which is a human call, not on
// every PR. The per-PR rule made the version a commit counter — 1.45 to 1.57 in one afternoon
// — while the tags never moved, which is why no plugin had a tag matching its own version.
func TestContentMayChangeWithoutABump(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")
	commit(t, dir, "content change, no bump — an ordinary PR")

	problems, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("an ordinary PR must not be failed for leaving the version alone: %v", problems)
	}
}

func TestPassesWhenTheVersionMovesForward(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", `{"name":"demo","version":"0.22.0"}`)
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")
	commit(t, dir, "content change with a bump")

	problems, _, err := check(dir, "main")
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

	problems, _, err := check(dir, "main")
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
	if _, _, err := check(dir, "main"); err == nil {
		t.Error("no manifests at all must be an error, not a pass")
	}
}

// THE "NOT MEASURED" STATE IS GONE, AND ITS ABSENCE IS CHECKED (#405).
//
// It existed because the retired must-bump rule read COMMITTED history while the version was
// read from disk, so an uncommitted change made the question unanswerable. The only surviving
// check reads the manifest on disk against the merge-base — tree-based on both sides — so it
// always answers. A branch that can never fire, reporting a state that can never occur, is the
// dead check this suite spent the week removing; it is not left behind as plumbing.
func TestUncommittedContentIsStillMeasured(t *testing.T) {
	dir := repo(t, "0.21.0")
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed, uncommitted\n")

	problems, checked, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("nothing was compared, so this test would pass on any behaviour")
	}
	if len(problems) != 0 {
		t.Errorf("uncommitted content is not a violation now that versions move at a release: %v", problems)
	}
}

// A committed violation OUTRANKS "could not measure": a dirty tree must never mask a real
// failure that the committed history already proves.
// A backwards version is tree-based on both sides, so it fires whether or not anything is
// committed. Uncommitted content must not downgrade it to "not measured".
func TestBackwardsVersionFiresEvenUncommitted(t *testing.T) {
	dir := repo(t, "0.22.0")
	write(t, dir, "plugins/demo/.claude-plugin/plugin.json", `{"name":"demo","version":"0.21.0"}`)
	write(t, dir, "plugins/demo/skills/x/SKILL.md", "# x, changed\n")

	problems, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "BACKWARDS") {
		t.Fatalf("a backwards version is answerable from the tree and must FAIL: %v", problems)
	}
}

// A clean tree with nothing changed stays silent: the gate must not cry wolf on every run.
func TestCleanTreeIsSilent(t *testing.T) {
	dir := repo(t, "0.21.0")
	problems, _, err := check(dir, "main")
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Errorf("nothing changed: want silence, got %v", problems)
	}
}

// THE RELEASE BOUNDARY, WHICH IS WHAT REPLACES THE PER-PR BUMP (#405).
//
// A tag cannot be derived from the manifest, nor the manifest from the tag: the plugin system
// reads plugin.json out of the checkout, so the version has to be IN the tagged commit. The
// non-circular answer is that one release act writes both — and this is what makes them unable
// to disagree. Measured 2026-08-15: NO plugin had a tag matching its own version, so
// `sc-doctor -fix` pinned every download to a release that does not exist.
func TestReleaseTagMustMatchTheManifest(t *testing.T) {
	dir := repo(t, "0.21.0")

	if err := releaseTagMatchesManifest(dir, "demo--v0.21.0"); err != nil {
		t.Errorf("a tag matching the manifest must be accepted: %v", err)
	}

	err := releaseTagMatchesManifest(dir, "demo--v0.22.0")
	if err == nil {
		t.Fatal("a tag ahead of the manifest publishes assets no `/plugin update` will ask for; it must be refused")
	}
	for _, want := range []string{"0.22.0", "0.21.0", "ONE act"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name both numbers and the remedy, missing %q: %v", want, err)
		}
	}
}

// `-tag` WITH NO VALUE MUST NOT EXIT 0. This is the argument-shaped version of the plausible
// zero: the flag used to fall through to the ordinary branch check, which passes on a clean
// tree — so a release job whose GITHUB_REF_NAME came through empty would have printed a green
// gate for a tag nobody ever examined. Exercised through the real binary because the defect
// lived in argument dispatch, which is the part a direct call to releaseTagMatchesManifest
// cannot see.
func TestTagModeWithNoTagIsRefused(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-tag")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("`-tag` with no value exited 0; it must refuse rather than measure something else:\n%s", out)
	}
	if !strings.Contains(string(out), "-tag needs the tag") {
		t.Errorf("the refusal must say what is missing: %s", out)
	}
}

// A malformed tag builds the wrong plugin or nothing, and the release job resolves the plugin
// FROM the tag — so the shape is refused before anything is published.
func TestReleaseTagMustNameAPluginAndVersion(t *testing.T) {
	dir := repo(t, "0.21.0")
	for _, tag := range []string{"demo", "demo--v", "--v0.21.0", "v0.21.0"} {
		if err := releaseTagMatchesManifest(dir, tag); err == nil {
			t.Errorf("tag %q is malformed and must be refused", tag)
		}
	}
	if err := releaseTagMatchesManifest(dir, "nosuchplugin--v1.0.0"); err == nil {
		t.Error("a tag naming a plugin with no manifest must be refused, not silently pass")
	}
}
