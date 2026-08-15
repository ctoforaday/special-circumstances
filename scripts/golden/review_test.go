package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The mode exists because `-update` could not fail. These tests hold the properties that
// make -review different: it never keeps what it was not told to keep, it never touches a
// file it did not write, and a mistyped accept is an error rather than a silent revert.

func props(paths ...string) []proposal {
	out := make([]proposal, len(paths))
	for i, p := range paths {
		out[i] = proposal{path: p}
	}
	return out
}

// ACCEPT NOTHING BY DEFAULT. If this ever inverts, -review becomes -update with extra steps.
func TestReviewPlanKeepsNothingWithoutAnAccept(t *testing.T) {
	p := props("a/testdata/x.golden", "b/testdata/y.golden")
	accepted, reverted := reviewPlan(p, map[string]bool{}, false)
	if len(accepted) != 0 {
		t.Fatalf("accepted %d with no -accept; the default must keep NOTHING", len(accepted))
	}
	if len(reverted) != 2 {
		t.Fatalf("reverted %d, want 2", len(reverted))
	}
}

func TestReviewPlanKeepsOnlyTheNamedGolden(t *testing.T) {
	p := props("a/testdata/x.golden", "b/testdata/y.golden")
	keep, err := acceptSet("x.golden", p)
	if err != nil {
		t.Fatal(err)
	}
	accepted, reverted := reviewPlan(p, keep, false)
	if len(accepted) != 1 || accepted[0].path != "a/testdata/x.golden" {
		t.Fatalf("accepted = %v, want only a/testdata/x.golden", accepted)
	}
	if len(reverted) != 1 || reverted[0].path != "b/testdata/y.golden" {
		t.Fatalf("reverted = %v, want only b/testdata/y.golden", reverted)
	}
}

func TestAcceptAllKeepsEverything(t *testing.T) {
	p := props("a/testdata/x.golden", "b/testdata/y.golden")
	accepted, reverted := reviewPlan(p, map[string]bool{}, true)
	if len(accepted) != 2 || len(reverted) != 0 {
		t.Fatalf("accept-all kept %d and reverted %d, want 2 and 0", len(accepted), len(reverted))
	}
}

// A name that matches nothing must be an ERROR. Treating it as a no-op would revert the
// golden the author believed they had just accepted — the exact silent-wrong-answer this
// mode exists to remove.
func TestAcceptSetRefusesANameThatMatchesNothing(t *testing.T) {
	_, err := acceptSet("typo.golden", props("a/testdata/x.golden"))
	if err == nil {
		t.Fatal("a non-matching -accept must be an error, not a silent no-op")
	}
	if !strings.Contains(err.Error(), "matched none") {
		t.Errorf("error should say the name matched nothing, got: %v", err)
	}
}

// A bare filename is enough; a full path also works.
func TestAcceptSetMatchesBareNameAndFullPath(t *testing.T) {
	p := props("deep/nested/testdata/x.golden")
	for _, spec := range []string{"x.golden", "deep/nested/testdata/x.golden"} {
		keep, err := acceptSet(spec, p)
		if err != nil {
			t.Fatalf("%s: %v", spec, err)
		}
		if !keep["deep/nested/testdata/x.golden"] {
			t.Errorf("%s did not match", spec)
		}
	}
}

// THE FENCE. Anything dirty before the run belongs to the human, and must never appear as
// a proposal — a proposal is a revert candidate.
func TestProposalsExcludeWhatWasAlreadyDirty(t *testing.T) {
	before := goldenState{
		modified:  map[string]bool{"mine.golden": true},
		untracked: map[string]bool{"also-mine.golden": true},
	}
	after := goldenState{
		modified:  map[string]bool{"mine.golden": true, "tool-changed.golden": true},
		untracked: map[string]bool{"also-mine.golden": true, "tool-made.golden": true},
	}
	got := proposalsBetween(before, after)
	if len(got) != 2 {
		t.Fatalf("proposals = %v, want exactly the two the run produced", got)
	}
	for _, p := range got {
		if strings.Contains(p.path, "mine") {
			t.Errorf("%s was dirty before the run and must never be a proposal", p.path)
		}
	}
	if !got[0].isNew && got[1].isNew {
		return // tool-changed (modified) then tool-made (new), sorted
	}
}

func TestProposalsMarkNewFilesAsNew(t *testing.T) {
	before := goldenState{modified: map[string]bool{}, untracked: map[string]bool{}}
	after := goldenState{
		modified:  map[string]bool{"old.golden": true},
		untracked: map[string]bool{"brand-new.golden": true},
	}
	got := proposalsBetween(before, after)
	byPath := map[string]proposal{}
	for _, p := range got {
		byPath[p.path] = p
	}
	if byPath["brand-new.golden"].isNew != true {
		t.Error("a golden git has never seen must be marked NEW — git diff cannot show it")
	}
	if byPath["old.golden"].isNew != false {
		t.Error("a tracked golden must not be marked NEW; reverting it would DELETE it")
	}
}

// A new golden reverts by deletion, a tracked one by checkout. Getting these the wrong way
// round either deletes a recorded contract or leaves an unreviewed file behind.
func TestRevertDeletesANewGoldenAndRestoresATrackedOne(t *testing.T) {
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "changed\n")
	write(t, dir, "testdata/new.golden", "brand new\n")

	if err := revert(dir, proposal{path: "testdata/new.golden", isNew: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "testdata/new.golden")); !os.IsNotExist(err) {
		t.Error("a new golden must be removed on revert")
	}

	if err := revert(dir, proposal{path: "testdata/one.golden"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "recorded\n" {
		t.Errorf("tracked golden = %q, want the committed content restored", b)
	}
}

// The precondition that makes the revert safe.
func TestReviewRefusesToStartOnADirtyGoldenTree(t *testing.T) {
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "a change the human made\n")

	st, err := readGoldenState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.dirty()) != 1 {
		t.Fatalf("dirty = %v, want the human's edit to be seen", st.dirty())
	}
	if code := reviewMode(dir, "", false, ""); code != 2 {
		t.Errorf("exit = %d, want 2 — review must refuse a tree it cannot tell apart from its own writes", code)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if string(b) != "a change the human made\n" {
		t.Errorf("the human's edit was touched: %q", b)
	}
}

// -accept-all without a reason is refused, and refused BEFORE anything runs.
func TestAcceptAllWithoutAReasonIsRefusedUpFront(t *testing.T) {
	if code := reviewMode(scratchRepo(t), "", true, "   "); code != 2 {
		t.Errorf("exit = %d, want 2 — -accept-all must cost a stated reason", code)
	}
}

// readGoldenState must ERROR outside a repo rather than report a clean tree; a clean read
// is what would let the dirty-tree fence pass on a tree it cannot see.
func TestReadGoldenStateErrorsOutsideARepo(t *testing.T) {
	if _, err := readGoldenState(t.TempDir()); err == nil {
		t.Error("state read outside a git repo must be an error, never an empty (clean) result")
	}
}

// The diff budget must announce what it withheld. A silently truncated diff is a review
// that looks complete and is not.
func TestBudgetSaysWhatItWithheld(t *testing.T) {
	var lines []string
	for i := 0; i < diffBudget+40; i++ {
		lines = append(lines, "line")
	}
	got := budget(strings.Join(lines, "\n"))
	if !strings.Contains(got, "40 more line(s)") {
		t.Errorf("truncation must state the withheld count, got tail: %q", got[len(got)-160:])
	}
	if !strings.Contains(got, "git diff") {
		t.Error("truncation must point at the command that shows the rest")
	}
}

// A verify failure with no golden movement is a REAL test failure and must not be reported
// as staleness — that would send the author to regenerate a contract that is not the problem.
//
// NOTE FOR ANYONE READING THIS SUITE'S OUTPUT: reviewMode streams its child `go test` to
// stdout, so this case prints a genuine "--- FAIL:" line from the FIXTURE below while this
// test passes. The fixture is named so that line cannot be mistaken for a failure of this
// suite — a passing run that prints FAIL is precisely the misleading signal this repo keeps
// finding, and the fix is to make the name say so rather than to hide the output.
func TestReviewSeparatesARealFailureFromStaleness(t *testing.T) {
	dir := scratchRepo(t)
	// A go module whose test fails without touching any golden.
	write(t, dir, "mod/go.mod", "module scratch\n\ngo 1.25\n")
	write(t, dir, "mod/x_test.go", "package scratch\n\nimport \"testing\"\n\n"+
		"func TestFixtureFailsOnPurpose_NotAFailureOfTheGoldenSuite(t *testing.T) { t.Fatal(\"deliberate\") }\n")

	saveGo, saveMjs := goModules, mjsSuites
	goModules, mjsSuites = []string{"mod"}, nil
	t.Cleanup(func() { goModules, mjsSuites = saveGo, saveMjs })

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	if code := reviewMode(dir, "", false, ""); code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "testdata/one.golden"))
	if string(b) != "recorded\n" {
		t.Errorf("a non-golden failure must leave goldens untouched, got %q", b)
	}
}
