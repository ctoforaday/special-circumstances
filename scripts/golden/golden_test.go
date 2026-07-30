package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// scratchRepo builds a git repo, because changeReport's whole job is asking git what moved.
func scratchRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		full := append([]string{
			"-c", "user.email=t@example.com", "-c", "user.name=t", "-c", "commit.gpgsign=false",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "main")
	write(t, dir, "testdata/one.golden", "recorded\n")
	git("add", "-A")
	git("commit", "-q", "-m", "seed")
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

// An unmoved tree must say so, in words. "no output" and "nothing changed" are the same
// bytes on a terminal and different facts to a reader about to commit.
func TestChangeReportOnAnUnmovedTree(t *testing.T) {
	got := changeReport(scratchRepo(t))
	if !strings.Contains(got, "no golden changed") {
		t.Errorf("report = %q; an unmoved tree must state that it is unmoved", got)
	}
}

func TestChangeReportSeesModifiedGoldens(t *testing.T) {
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "recorded differently\n")

	got := changeReport(dir)
	if !strings.Contains(got, "one.golden") {
		t.Errorf("report = %q; a modified golden must appear", got)
	}
	if !strings.Contains(got, "ON ITS OWN") {
		t.Errorf("report = %q; a changed golden must carry the commit-separately guidance", got)
	}
}

// THE case the .mjs original called out and the reason it asks git twice: `git diff` cannot
// see a golden that did not exist before, and a NEW golden is exactly what a new seat class
// or scenario produces. A report that misses those under-reports on the runs that matter
// most.
func TestChangeReportSeesNewGoldens(t *testing.T) {
	dir := scratchRepo(t)
	write(t, dir, "testdata/two.golden", "brand new\n")

	got := changeReport(dir)
	if !strings.Contains(got, "1 new golden") {
		t.Errorf("report = %q; an untracked golden must be counted", got)
	}
	if !strings.Contains(got, "+ testdata/two.golden") {
		t.Errorf("report = %q; the new golden must be named", got)
	}
}

func TestChangeReportSeesBothAtOnce(t *testing.T) {
	dir := scratchRepo(t)
	write(t, dir, "testdata/one.golden", "changed\n")
	write(t, dir, "testdata/two.golden", "new\n")

	got := changeReport(dir)
	if !strings.Contains(got, "one.golden") || !strings.Contains(got, "+ testdata/two.golden") {
		t.Errorf("report = %q; modified and new goldens must both appear", got)
	}
}

// A non-golden file must not be reported. The report is the review surface for BEHAVIOUR;
// padding it with unrelated churn is how a real drift hides in noise.
func TestChangeReportIgnoresNonGoldens(t *testing.T) {
	dir := scratchRepo(t)
	write(t, dir, "src/main.go", "package main\n")
	write(t, dir, "notes.txt", "scratch\n")

	got := changeReport(dir)
	if strings.Contains(got, "main.go") || strings.Contains(got, "notes.txt") {
		t.Errorf("report = %q; only *.golden belongs in the change report", got)
	}
	if !strings.Contains(got, "no golden changed") {
		t.Errorf("report = %q; with no golden touched the tree is unmoved", got)
	}
}

// Losing the SUMMARY must not look like losing the RECORDING. By the time the report runs,
// the goldens are already written; a git failure here is a reporting problem and has to
// read as one.
func TestChangeReportDegradesWithoutGit(t *testing.T) {
	got := changeReport(t.TempDir()) // not a git repository
	if !strings.Contains(got, "ARE recorded") {
		t.Errorf("report = %q; a git failure must not imply the goldens were lost", got)
	}
}

// The enumerated lists are the thing that rots (a suite deleted, the list not updated), so
// every path named here must exist. `node --test` exits 0 on a missing path, which is how
// this class stays invisible.
func TestEnumeratedTargetsExist(t *testing.T) {
	root, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("not in a git repository")
	}
	repo := strings.TrimSpace(string(root))

	for _, s := range mjsSuites {
		if _, err := os.Stat(filepath.Join(repo, s)); err != nil {
			t.Errorf("mjsSuites names %s, which does not exist — `node --test` exits 0 on a missing path, so this would silently verify nothing", s)
		}
	}
	for _, m := range goModules {
		if _, err := os.Stat(filepath.Join(repo, m, "go.mod")); err != nil {
			t.Errorf("goModules names %s, which has no go.mod", m)
		}
	}
}
