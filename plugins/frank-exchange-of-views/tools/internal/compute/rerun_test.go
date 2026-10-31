package compute

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recorded runs a program as `compute` does and stores it as act id: its bundle under dir/ and its
// first run's output and out/ under acts/<id>/.
func recorded(t *testing.T, runDir, id string, p *Program) *Result {
	t.Helper()
	res := run(t, p)
	if err := WriteBundle(runDir, res.Bundle, res.Argv, p.Files); err != nil {
		t.Fatal(err)
	}
	if err := WriteAct(runDir, res.Bundle, id, []byte(res.Output), res.Out); err != nil {
		t.Fatal(err)
	}
	return res
}

func rerun(t *testing.T, runDir string, res *Result) Execution {
	t.Helper()
	e, err := Rerun(runDir, res.Bundle, res.Argv)
	if err != nil {
		t.Fatalf("Rerun: %v", err)
	}
	return e
}

func compare(t *testing.T, runDir string, res *Result, id string, exit int, e Execution) Comparison {
	t.Helper()
	c, err := Compare(runDir, res.Bundle, id, exit, e)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	return c
}

// RED'S AUDIT IS RE-EXECUTION, COMPARED AGAINST THE ACT'S OWN RESULT. One execution of the stored
// bundle, set against this act's output, exit and out/: an out/ file the act recorded otherwise is a
// difference named by its file, with the output equal and so not restated.
func TestAReRunIsComparedAgainstTheActsOwnResult(t *testing.T) {
	runDir := t.TempDir()
	p := program(t, bundle(t, map[string]string{"a": "a"}), "out", "t.csv", "n,7\n")
	res := recorded(t, runDir, "K-00000001", p)
	e := rerun(t, runDir, res)
	if c := compare(t, runDir, res, "K-00000001", res.Exit, e); !c.Matches || c.Drift != "" || c.Recorded != res.Output {
		t.Errorf("a reproducible computation's re-run reads %+v", c)
	}
	if err := os.WriteFile(filepath.Join(ActPath(runDir, res.Bundle, "K-00000001"), "out", "t.csv"), []byte("n,8\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c := compare(t, runDir, res, "K-00000001", res.Exit, e); c.Matches || c.Drift != "out/t.csv differs" {
		t.Errorf("an out/ file the act recorded otherwise, output equal: %+v, want drift naming out/t.csv alone", c)
	}
	// A recorded exit the execution did not give is a difference.
	if c := compare(t, runDir, res, "K-00000001", 3, e); c.Matches || !strings.Contains(c.Drift, "exit moved 3 -> 0") {
		t.Errorf("an exit the execution moved: %+v", c)
	}
	// A file the re-run alone wrote is named as the re-run's.
	extra := e
	extra.Out = append(append([]File{}, e.Out...), File{Path: "u.csv", Body: []byte("x")})
	if c := compare(t, runDir, res, "K-00000001", res.Exit, extra); !strings.Contains(c.Drift, "out/u.csv written by the re-run only") {
		t.Errorf("a file the re-run alone wrote: %q", c.Drift)
	}
	// What the execution wrote is in the drift where the output differs.
	moved := e
	moved.Output = "something else\n"
	if c := compare(t, runDir, res, "K-00000001", res.Exit, moved); c.Matches || !strings.Contains(c.Drift, "on re-run the program wrote: something else") {
		t.Errorf("an output the execution moved: %+v", c)
	}
}

// AN ACT THAT WROTE NO out/ EQUALS AN EXECUTION WHOSE out/ IS EMPTY: acts/<id>/ holds no out/ when the
// run wrote nothing there, and every execution's copy holds an empty one.
func TestAnAbsentRecordedOutEqualsAnEmptyOne(t *testing.T) {
	runDir := t.TempDir()
	res := recorded(t, runDir, "K-00000001", program(t, bundle(t, map[string]string{"a": "a"}), "print", "7"))
	if _, err := os.Stat(filepath.Join(ActPath(runDir, res.Bundle, "K-00000001"), "out")); !os.IsNotExist(err) {
		t.Fatalf("an act that wrote no out/ holds one: %v", err)
	}
	e := rerun(t, runDir, res)
	e.Out = []File{}
	if c := compare(t, runDir, res, "K-00000001", res.Exit, e); !c.Matches {
		t.Errorf("an absent recorded out/ against an empty one: %+v", c)
	}
}

// A BUNDLE THAT CALLS THE TOOL NEVER MATCHES, and the execution names the call.
func TestAnExecutionThatCallsTheToolNeverMatches(t *testing.T) {
	runDir := t.TempDir()
	p := program(t, bundle(t, map[string]string{"a": "a"}), "call")
	if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
		t.Fatal(err)
	}
	if err := WriteAct(runDir, p.Sha, "K-00000001", []byte("0\n"), nil); err != nil {
		t.Fatal(err)
	}
	e, err := Rerun(runDir, p.Sha, p.Argv)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Compare(runDir, p.Sha, "K-00000001", 0, e)
	if err != nil {
		t.Fatal(err)
	}
	if c.Matches || len(e.Calls) != 1 {
		t.Errorf("the re-run of a program that called the tool reads matches=%v calls=%q — its output %q equals the recorded %q, and that is the hiding this exists to end", c.Matches, e.Calls, e.Output, c.Recorded)
	}
}

// Re-running or comparing something never recorded fails loudly, never agreement with nothing.
func TestARerunOfAnUnknownComputationIsRefused(t *testing.T) {
	if _, err := Rerun(t.TempDir(), "deadbeef", []string{"node", "script.js"}); err == nil {
		t.Error("re-running a bundle that was never stored reported success")
	}
	if _, err := Compare(t.TempDir(), "deadbeef", "K-00000001", 0, Execution{}); err == nil {
		t.Error("comparing against an act that was never stored reported success")
	}
}

// AN EXECUTION'S NAME IS THE SHA OF WHAT IT WROTE, framed so no file name forges another listing.
func TestAnExecutionsShaIsItsContent(t *testing.T) {
	base := Execution{Output: "7\n", Exit: 0, Out: []File{{Path: "a", Body: []byte("x")}}}
	if base.Sha() != (Execution{Output: "7\n", Out: []File{{Path: "a", Body: []byte("x")}}, Calls: []string{"c"}}).Sha() {
		t.Error("the sha moved with something the directory does not hold")
	}
	for name, e := range map[string]Execution{
		"exit":     {Output: "7\n", Exit: 1, Out: base.Out},
		"output":   {Output: "8\n", Out: base.Out},
		"out body": {Output: "7\n", Out: []File{{Path: "a", Body: []byte("y")}}},
		"out name": {Output: "7\n", Out: []File{{Path: "b", Body: []byte("x")}}},
		"no out":   {Output: "7\n"},
	} {
		if e.Sha() == base.Sha() {
			t.Errorf("a different %s has the same sha", name)
		}
	}
	split := Execution{Out: []File{{Path: "a\n1:b", Body: nil}}}
	joined := Execution{Out: []File{{Path: "a", Body: nil}, {Path: "b", Body: nil}}}
	if split.Sha() == joined.Sha() {
		t.Error("a file name holding a newline forged a two-file listing")
	}
}

// THE RE-RUN STORE IS STAGED AND RENAMED. A discarded staging leaves nothing; a promoted one is
// rerun/<sha>/ holding the output and out/; a second promotion of the same bytes finds the first in
// place and leaves no staging beside it.
func TestTheReRunStoreIsStagedThenRenamed(t *testing.T) {
	runDir := t.TempDir()
	res := recorded(t, runDir, "K-00000001", program(t, bundle(t, map[string]string{"a": "a"}), "out", "t.csv", "n,7\n"))
	e := rerun(t, runDir, res)
	bundleDir := filepath.Join(runDir, StoreDir, res.Bundle)
	staged := func() []string {
		t.Helper()
		m, _ := filepath.Glob(filepath.Join(bundleDir, ".tmp-*"))
		return m
	}

	s, err := StageRerun(runDir, res.Bundle, e)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(bundleDir, "rerun")); !os.IsNotExist(err) {
		t.Fatalf("staging wrote under rerun/: %v", err)
	}
	if err := s.Discard(); err != nil {
		t.Fatal(err)
	}
	if m := staged(); len(m) != 0 {
		t.Errorf("a discarded staging left %q", m)
	}

	for i := 0; i < 2; i++ {
		s, err := StageRerun(runDir, res.Bundle, e)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Promote(); err != nil {
			t.Fatalf("promotion %d: %v", i+1, err)
		}
		if s.Sha != e.Sha() {
			t.Errorf("the staged sha %s is not the execution's %s", s.Sha, e.Sha())
		}
		if err := s.Discard(); err != nil {
			t.Errorf("discarding a promoted staging: %v", err)
		}
	}
	if m := staged(); len(m) != 0 {
		t.Errorf("a promotion left staging %q", m)
	}
	entries, _ := os.ReadDir(filepath.Join(bundleDir, "rerun"))
	if len(entries) != 1 || entries[0].Name() != e.Sha() {
		t.Fatalf("rerun/ holds %v, want the one execution %s", entries, e.Sha())
	}
	got, err := ReadRerun(runDir, res.Bundle, e.Sha(), e.Exit)
	if err != nil {
		t.Fatal(err)
	}
	if got.Sha() != e.Sha() {
		t.Errorf("the stored execution reads back as %+v, want %+v", got, e)
	}
	if _, err := ReadRerun(runDir, res.Bundle, strings.Repeat("0", 64), 0); err == nil {
		t.Error("a re-run the store does not hold read as an execution")
	}
}

// TWO LANES WRITING ONE BUNDLE (V1a.5 (e)): both stage and rename at once, one dir/ stands and both
// acts are recorded, and no staging is left. An act's directory already present is refused, never
// written over.
func TestTwoLanesWriteOneBundleOnce(t *testing.T) {
	runDir := t.TempDir()
	p := program(t, bundle(t, map[string]string{"a": "a", "b/c": "c"}), "print", "7")
	errs := make(chan error, 2)
	for _, id := range []string{"K-00000001", "K-00000002"} {
		go func() {
			if err := WriteBundle(runDir, p.Sha, p.Argv, p.Files); err != nil {
				errs <- err
				return
			}
			errs <- WriteAct(runDir, p.Sha, id, []byte("7\n"), nil)
		}()
	}
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatalf("a lane's write: %v", err)
		}
	}
	files, err := Bundle(runDir, p.Sha)
	if err != nil || BundleSha(p.Argv, Sums(files)) != p.Sha {
		t.Fatalf("dir/ does not hash to its name: %v", err)
	}
	for _, id := range []string{"K-00000001", "K-00000002"} {
		if _, err := ActOutput(runDir, p.Sha, id); err != nil {
			t.Errorf("act %s: %v", id, err)
		}
	}
	if m, _ := filepath.Glob(filepath.Join(runDir, StoreDir, p.Sha, ".tmp-*")); len(m) != 0 {
		t.Errorf("staging left: %q", m)
	}
	if err := WriteAct(runDir, p.Sha, "K-00000001", []byte("other\n"), nil); err == nil {
		t.Error("an act's directory already present was written over")
	}
	if b, _ := ActOutput(runDir, p.Sha, "K-00000001"); string(b) != "7\n" {
		t.Errorf("the refused write changed the act: %q", b)
	}
}
