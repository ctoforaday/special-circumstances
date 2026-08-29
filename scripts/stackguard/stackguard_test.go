package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeContains answers from a table, so the reporting logic is tested without a repository.
func fakeContains(merged map[string]bool, fail map[string]bool) containsFn {
	return func(ref string) (bool, error) {
		if fail[ref] {
			return false, os.ErrNotExist
		}
		return merged[ref], nil
	}
}

func TestABaseAlreadyInMainIsRefused(t *testing.T) {
	got := reportOne(os.Stdout, "record-run-read-leaves", fakeContains(map[string]bool{"record-run-read-leaves": true}, nil))
	if got != exitFound {
		t.Fatalf("exit = %d, want %d — a stacked base already in main must be refused", got, exitFound)
	}
}

func TestALiveBasePasses(t *testing.T) {
	if got := reportOne(os.Stdout, "some-live-branch", fakeContains(nil, nil)); got != exitClean {
		t.Fatalf("exit = %d, want %d — a base that is not yet in main is the normal stacked case", got, exitClean)
	}
}

func TestABaseOfMainIsNotEvenAsked(t *testing.T) {
	asked := false
	probe := func(string) (bool, error) { asked = true; return true, nil }
	if got := reportOne(os.Stdout, "main", probe); got != exitClean {
		t.Fatalf("exit = %d, want %d", got, exitClean)
	}
	if asked {
		t.Error("a pull request based on main asked git anything; it is in main by construction")
	}
}

// AN UNRESOLVABLE BASE IS AN ERROR, NEVER A PASS. This is the whole reason the tool has an
// exit 2: `merge-base --is-ancestor` exits 1 for "not an ancestor" AND for a ref git cannot
// parse, so a typo'd or unfetched base would otherwise read exactly like a healthy branch.
func TestAnUnresolvableBaseIsLoudRatherThanClean(t *testing.T) {
	got := reportOne(os.Stdout, "never-fetched", fakeContains(nil, map[string]bool{"never-fetched": true}))
	if got != exitError {
		t.Fatalf("exit = %d, want %d — a base the tool could not resolve must not report clean", got, exitError)
	}
}

func TestTheSweepReportsEveryStrandedPullRequestNotJustTheFirst(t *testing.T) {
	prs := []pullRequest{
		{Number: 633, Base: "record-run-read-leaves", Title: "record itself holds a run"},
		{Number: 700, Base: "still-live", Title: "a healthy stack"},
		{Number: 701, Base: "also-merged", Title: "another stranded one"},
	}
	contains := fakeContains(map[string]bool{"record-run-read-leaves": true, "also-merged": true}, nil)
	if got := reportSweep(os.Stdout, prs, contains); got != exitFound {
		t.Fatalf("exit = %d, want %d", got, exitFound)
	}
}

// One unresolvable base must not be absorbed by the healthy ones around it.
func TestTheSweepDoesNotLetOneUnmeasurableBaseReadAsClean(t *testing.T) {
	prs := []pullRequest{
		{Number: 1, Base: "live-one", Title: "fine"},
		{Number: 2, Base: "vanished", Title: "cannot measure"},
	}
	got := reportSweep(os.Stdout, prs, fakeContains(nil, map[string]bool{"vanished": true}))
	if got != exitError {
		t.Fatalf("exit = %d, want %d — an unmeasured base outranks the clean ones", got, exitError)
	}
}

func TestAnEmptySweepSaysSoRatherThanImplyingItChecked(t *testing.T) {
	if got := reportSweep(os.Stdout, nil, fakeContains(nil, nil)); got != exitClean {
		t.Fatalf("exit = %d, want %d", got, exitClean)
	}
}

func TestDecodePRsReadsGhsLineDelimitedObjects(t *testing.T) {
	in := "{\"Number\":633,\"Base\":\"record-run-read-leaves\",\"Title\":\"x\"}\n\n{\"Number\":55,\"Base\":\"feat/x\",\"Title\":\"y\"}\n"
	prs, err := decodePRs(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 2 || prs[0].Number != 633 || prs[1].Base != "feat/x" {
		t.Fatalf("decodePRs = %+v", prs)
	}
}

func TestMalformedGhOutputIsAnErrorNotAnEmptySweep(t *testing.T) {
	if _, err := decodePRs("{not json}"); err == nil {
		t.Fatal("decodePRs accepted malformed input — an empty list here reads as a clean board")
	}
}

// THE REAL THING, against a real repository: #633's timeline, reconstructed.
//
// A table-driven test of reportOne proves the reporting. It does not prove that
// merge-base --is-ancestor answers the question this tool thinks it answers, which is where
// the actual bug lived. This builds the branch shape and asks git.
func TestGitContainsAnswersTheSixThirtyThreeTimeline(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	commit := func(msg string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, "f"), []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", "f")
		run("commit", "-m", msg)
	}

	run("init", "-q", "-b", "main")
	commit("base")

	// The stack: a feature branch off main, which is #631's branch.
	run("checkout", "-q", "-b", "record-run-read-leaves")
	commit("the read leaves")

	// Local branches stand in for origin/*, which is what the tool resolves.
	run("update-ref", "refs/remotes/origin/record-run-read-leaves", "record-run-read-leaves")
	run("checkout", "-q", "main")
	run("update-ref", "refs/remotes/origin/main", "main")

	contains := gitContains(dir)

	// BEFORE #631 merges: the base is live, and a stacked pull request onto it is fine.
	// This is the state #633's own checks saw, and why they passed.
	merged, err := contains("record-run-read-leaves")
	if err != nil {
		t.Fatal(err)
	}
	if merged {
		t.Fatal("the base read as already-in-main BEFORE it was merged")
	}
	if got := reportOne(os.Stdout, "record-run-read-leaves", contains); got != exitClean {
		t.Fatalf("pre-merge exit = %d, want %d", got, exitClean)
	}

	// 17:00 — #631 merges the base into main, and the branch is NOT deleted.
	run("merge", "--no-ff", "-m", "Merge pull request #631", "record-run-read-leaves")
	run("update-ref", "refs/remotes/origin/main", "main")

	// 19:47 — this is the moment #633 merged. The guard must now refuse.
	merged, err = contains("record-run-read-leaves")
	if err != nil {
		t.Fatal(err)
	}
	if !merged {
		t.Fatal("the base did NOT read as in-main after being merged — the guard would have stayed silent, which is the #633 defect")
	}
	if got := reportOne(os.Stdout, "record-run-read-leaves", contains); got != exitFound {
		t.Fatalf("post-merge exit = %d, want %d — this is the exact case the tool exists for", got, exitFound)
	}
}

// The stated blind spot, asserted rather than left to a comment: a SQUASH-merged base leaves
// no ancestry, so this guard reports it live. A future reader who changes the containment test
// should see this fail and decide deliberately, rather than discover the limit in production.
func TestASquashMergedBaseIsNotDetectedAndThatIsDocumented(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-q", "-b", "main")
	write("f", "base")
	run("add", "f")
	run("commit", "-m", "base")
	run("checkout", "-q", "-b", "squashed-base")
	write("g", "feature")
	run("add", "g")
	run("commit", "-m", "feature")
	run("update-ref", "refs/remotes/origin/squashed-base", "squashed-base")
	run("checkout", "-q", "main")
	run("merge", "--squash", "squashed-base")
	run("commit", "-m", "squashed feature")
	run("update-ref", "refs/remotes/origin/main", "main")

	merged, err := gitContains(dir)("squashed-base")
	if err != nil {
		t.Fatal(err)
	}
	if merged {
		t.Fatal("a squash-merged base now reads as contained — the tool's stated blind spot has closed; " +
			"update the package comment, which currently tells readers this case is NOT caught")
	}
}

// Every caller spells the base differently, and all three must mean the same branch.
// `github.base_ref` is bare; scripts/check defaults the flag to `origin/main`; the sibling
// gates in this directory are invoked as `-base "origin/$BASE_REF"`.
func TestTheDefaultBranchIsRecognisedInEverySpelling(t *testing.T) {
	for _, ref := range []string{"main", "origin/main", "refs/heads/main"} {
		asked := false
		probe := func(string) (bool, error) { asked = true; return true, nil }
		if got := reportOne(os.Stdout, ref, probe); got != exitClean {
			t.Errorf("reportOne(%q) = %d, want %d — the local gate set passes origin/main and would refuse itself", ref, got, exitClean)
		}
		if asked {
			t.Errorf("reportOne(%q) consulted git about the default branch", ref)
		}
	}
}
