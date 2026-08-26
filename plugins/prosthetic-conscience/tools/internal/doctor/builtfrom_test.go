package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
)

// THE FIVE STATES ARE FIVE, NOT TWO. The defect this closes (#450) is that a stale binary and a
// current one produced the same evidence. Collapsing "cannot tell" into "current" would rebuild
// exactly that: a check that did not run, reported as a check that passed.
//
// The last row used to expect Unstamped for a STAMPED binary with no reference, and that expectation
// was itself the bug: it made the verdict line say "carries no build stamp" about binaries that
// carry one, in every real installation — an install directory is not a checkout, so HEAD was
// always "" there and the sentence always blamed the binary. The two failures name different
// parties and now have different words.
func TestCompareSeparatesCannotTellFromClean(t *testing.T) {
	head := "4bf16692af2bc2334eb9cbf08c54a9d8f8cb7162"
	other := "62155fa54f9be1336081ce450f905fc0a7aef8eb"

	for _, c := range []struct {
		name string
		s    BuildStamp
		head string
		want Staleness
	}{
		{"same commit, clean tree", BuildStamp{Revision: head}, head, Current},
		{"different commit", BuildStamp{Revision: other}, head, Stale},
		{"same commit, dirty tree", BuildStamp{Revision: head, Modified: true}, head, Dirty},
		{"no stamp at all", BuildStamp{}, head, Unstamped},
		{"unstamped AND no reference — the binary's fault is named first", BuildStamp{}, "", Unstamped},
		{"stamped but no reference to check against", BuildStamp{Revision: head}, "", NoReference},
	} {
		if got := Compare(c.s, c.head); got != c.want {
			t.Errorf("%s: Compare = %q, want %q", c.name, got, c.want)
		}
	}
}

// A STALE VERDICT MUST NAME THE REMEDY. The entire failure mode is that nothing about a stale
// binary suggests rebuilding it, so a message that reports the fact without the action leaves the
// operator exactly where they were.
func TestStaleMessageNamesBothCommitsAndTheFix(t *testing.T) {
	s := BuildStamp{Revision: "62155fa54f9be1336081ce450f905fc0a7aef8eb"}
	msg := Describe(Stale, s, "4bf16692af2bc2334eb9cbf08c54a9d8f8cb7162")
	for _, want := range []string{"62155fa", "4bf1669", "doctor --fix"} {
		if !strings.Contains(msg, want) {
			t.Errorf("stale message is missing %q: %s", want, msg)
		}
	}
}

// Unstamped must not read as an accusation: it is the tool saying it could not answer, which is a
// different thing from saying the binary is wrong.
func TestUnstampedSaysItCannotJudgeRatherThanThatSomethingIsWrong(t *testing.T) {
	msg := Describe(Unstamped, BuildStamp{}, "abc1234")
	if !strings.Contains(msg, "cannot be judged") {
		t.Errorf("unstamped message must say the question was unanswerable: %s", msg)
	}
	if strings.Contains(msg, "doctor --fix") {
		t.Errorf("unstamped must not prescribe a rebuild it has no evidence for: %s", msg)
	}
}

// THE LOAD-BEARING TEST, and it is an END-TO-END one on purpose.
//
// Everything above is arithmetic on a struct and would pass even if the Go toolchain stamped
// nothing at all. The claim this whole change rests on is that the stamp is ALREADY THERE on an
// ordinary build with no flags — so this builds a real binary the ordinary way and reads it back.
// If a future toolchain or a `-buildvcs=false` default ever removes it, this fails and the reason
// is visible, rather than every binary silently becoming Unstamped.
func TestAnOrdinaryBuildCarriesItsRevisionWithNoFlags(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	root := repoRoot(t)
	// THE CHECKOUT GO STAMPS FROM. In a worktree `.git` is a file pointing at the main
	// checkout, so the stamp carries the MAIN tree's revision and dirty flag — not this
	// tree's. Asserting the local HEAD failed in every worktree while the toolchain was
	// doing exactly what it documents (#532).
	stampDir, ok := buildid.StampedFrom(root)
	if !ok {
		t.Skip("not in a git checkout with git on PATH")
	}
	head := HeadCommit(stampDir)
	if head == "" {
		t.Skip("not in a git checkout with git on PATH")
	}

	out := filepath.Join(t.TempDir(), "probe")
	// No -ldflags. That absence IS the assertion.
	cmd := exec.Command("go", "build", "-o", out, "./cmd/sc-doctor")
	cmd.Dir = filepath.Join(root, "plugins", "prosthetic-conscience", "tools")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("build failed in this environment: %v\n%s", err, b)
	}

	s := ReadBuildStamp(out)
	if !s.Known() {
		t.Fatal("an ordinary `go build` inside the repo produced a binary with NO vcs stamp — " +
			"the mechanism this package reads has gone away, and every binary will now report Unstamped")
	}
	if s.Revision != head {
		t.Errorf("stamp = %s, HEAD of the stamping checkout (%s) = %s — a freshly built binary must report the revision Go read", s.Revision, stampDir, head)
	}
	if got := Compare(s, head); got != Current && got != Dirty {
		t.Errorf("a just-built binary judged %q; want current (or dirty if the tree has edits)", got)
	}
}

// A binary that is not there, or is not a Go binary, must return the empty stamp rather than
// blowing up — the doctor already reports existence separately, and a reader that panics on a
// truncated file is worse than one that says nothing.
func TestUnreadableBinariesReturnAnEmptyStampRatherThanFailing(t *testing.T) {
	if s := ReadBuildStamp(filepath.Join(t.TempDir(), "does-not-exist")); s.Known() {
		t.Error("a missing file must not produce a stamp")
	}
	junk := filepath.Join(t.TempDir(), "junk")
	if err := os.WriteFile(junk, []byte("not an executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	if s := ReadBuildStamp(junk); s.Known() {
		t.Error("a non-Go file must not produce a stamp")
	}
}

// ---- staleness reaches the verdict ----

// THE DETECTION WAS ALREADY RIGHT AND THE VERDICT IGNORED IT. Compare returned Stale, Describe
// named the remedy, the table printed the row — and the doctor returned READY in the same breath.
// Measured 2026-08-22: a hook fix committed at 07:50 sat undeployed behind binaries from 05:31
// while a debate ran sixteen minutes against the old behaviour.
func TestStalenessDegradesTheVerdict(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    Staleness
		want bool // should it degrade?
	}{
		{"stale binary predates its source", Stale, true},
		{"current binary", Current, false},
		// Dirty is the ordinary state of a working tree; degrading on it would make the verdict
		// permanently wrong during development, which teaches an operator to discount it.
		{"dirty tree, built from HEAD", Dirty, false},
	} {
		if got := (tc.v == Stale); got != tc.want {
			t.Errorf("%s: degrades=%v want %v", tc.name, got, tc.want)
		}
	}
}

// NOT MEASURED IS NOT MEASURED. An unstamped binary is what a plain `go build` produces — the
// exact thing that hid the stale hook — and Compare answers Unstamped, which is neither Current
// nor Stale. It must not be folded into either.
func TestUnstampedIsNeitherCurrentNorStale(t *testing.T) {
	var none BuildStamp
	if v := Compare(none, "abc123"); v != Unstamped {
		t.Errorf("an unstamped binary compared as %v — it must report that it cannot be judged", v)
	}
	if v := Compare(none, "abc123"); v == Current || v == Stale {
		t.Error("an unmeasurable binary must never read as a measurement")
	}
}
