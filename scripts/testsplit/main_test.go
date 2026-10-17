package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/testjson"
	"github.com/ctoforaday/special-circumstances/scripts/internal/testsplit"
)

const fixture = "testdata/fixture"

var fixtureRows = []testsplit.Row{{Module: "splitfixture", Package: "splitfixture/serial"}}

func runFixture(t *testing.T, dir string, rows []testsplit.Row, n int) (int, string) {
	t.Helper()
	t.Setenv("TESTSPLIT_FIXTURE", "1")
	var out strings.Builder
	rc := run(&out, dir, rows, n, "2m")
	return rc, out.String()
}

// The whole scope runs, the split package as several processes, and every listed name answers
// once: ten names listed, ten verdicts, whatever the number of processes.
func TestTheWholeScopeRunsAndEveryListedNameAnswers(t *testing.T) {
	for _, n := range []int{1, 3, 50} {
		rc, out := runFixture(t, fixture, fixtureRows, n)
		if rc != 0 {
			t.Fatalf("n=%d: exit %d\n%s", n, rc, out)
		}
		for _, want := range []string{
			"9 pass, 0 fail, 1 skip",
			"1 pass, 0 fail, 1 skip", // the plain package
			"splitfixture/empty",
			"testsplit: OK — 3 package(s)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("n=%d: the report lacks %q:\n%s", n, want, out)
			}
		}
		// One process splits nothing; otherwise never more processes than the ten names.
		want := fmt.Sprintf("[%d processes]", min(n, 10))
		if got := strings.Contains(out, "processes]"); n == 1 && got {
			t.Errorf("n=1 ran a package as separate processes; one process is plain go test:\n%s", out)
		} else if n > 1 && !strings.Contains(out, want) {
			t.Errorf("n=%d: the report lacks %q:\n%s", n, want, out)
		}
	}
}

// The caller's environment reaches both paths: without it the fixture's own tests fail.
func TestTheCallersEnvironmentReachesBothPaths(t *testing.T) {
	var out strings.Builder
	t.Setenv("TESTSPLIT_FIXTURE", "")
	if rc := run(&out, fixture, fixtureRows, 2, "2m"); rc != 1 {
		t.Fatalf("exit %d without the variable, want 1:\n%s", rc, out.String())
	}
	for _, want := range []string{
		"TESTSPLIT_FIXTURE did not reach a package on the plain path",
		"TESTSPLIT_FIXTURE did not reach a process of a split package",
		"failed: TestSplitSeesTheCallersEnvironment",
		"testsplit: FAILED — 2 package(s)",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the report lacks %q:\n%s", want, out.String())
		}
	}
}

func mutated(t *testing.T, file, old, new string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(fixture)); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, file)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), old) {
		t.Fatalf("%s does not contain %q; the mutation would change nothing", file, old)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(b), old, new, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestTheRunIsRed(t *testing.T) {
	for _, c := range []struct {
		name, file, old, new string
		want                 []string
	}{
		{"on a failing test in a split package",
			"serial/serial_test.go", "func TestB(t *testing.T) { state++ }", "func TestB(t *testing.T) { t.Fatal(\"planted in B\") }",
			[]string{"failed: TestB", "planted in B", "8 pass, 1 fail, 1 skip"}},
		{"on a split process that dies before its tests answer",
			"serial/serial_test.go", "func TestC(t *testing.T) { state++ }", "func TestC(t *testing.T) { os.Exit(0) }",
			[]string{"NOT MEASURED, no verdict came back for:", "TestC"}},
		{"on a failing test in a plain package",
			"plain/plain_test.go", "func TestPlainSkips(t *testing.T) { t.Skip(\"always\") }", "func TestPlainSkips(t *testing.T) { t.Fatal(\"planted in plain\") }",
			[]string{"planted in plain", "FAIL  splitfixture/plain"}},
		{"on a plain package that does not build",
			"plain/plain_test.go", "func TestPlainSkips(t *testing.T) { t.Skip(\"always\") }", "func TestPlainSkips(t *testing.T) { undeclared() }",
			[]string{"undeclared", "testsplit: FAILED"}},
		{"on a split package that does not build",
			"serial/serial_test.go", "func TestA(t *testing.T) { state++ }", "func TestA(t *testing.T) { undeclared() }",
			[]string{"building the test binary", "undeclared"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			rc, out := runFixture(t, mutated(t, c.file, c.old, c.new), fixtureRows, 3)
			if rc != 1 {
				t.Fatalf("exit %d, want 1:\n%s", rc, out)
			}
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("the report lacks %q:\n%s", w, out)
				}
			}
		})
	}
}

// A row naming a package the module does not have splits nothing, and a row for a package
// with no tests splits nothing either. Both are errors: a table that silently stopped matching
// would hand the package back to one core.
func TestARowThatSplitsNothingIsAnError(t *testing.T) {
	rc, out := runFixture(t, fixture, []testsplit.Row{{Module: "splitfixture", Package: "splitfixture/renamed"}}, 2)
	if rc != 1 || !strings.Contains(out, "splitfixture/renamed, which is not a package of this module") {
		t.Errorf("exit %d:\n%s", rc, out)
	}
	// A row is checked even when nothing is split, so a stale table is red on a small box too.
	rc, out = runFixture(t, fixture, []testsplit.Row{{Module: "splitfixture", Package: "splitfixture/renamed"}}, 1)
	if rc != 1 || !strings.Contains(out, "splitfixture/renamed, which is not a package of this module") {
		t.Errorf("n=1: exit %d:\n%s", rc, out)
	}
	rc, out = runFixture(t, fixture, []testsplit.Row{{Module: "splitfixture", Package: "splitfixture/empty"}}, 2)
	if rc != 1 || !strings.Contains(out, "has no tests to split") {
		t.Errorf("exit %d:\n%s", rc, out)
	}
}

// The real table: every row names its module, a package under it, and a reason.
func TestEveryRowOfTheTableIsWhole(t *testing.T) {
	if len(testsplit.Table) == 0 {
		t.Fatal("the table is empty; the tool would be plain go test with more steps")
	}
	for _, r := range testsplit.Table {
		if !strings.HasPrefix(r.Package, r.Module+"/") || r.SerialSeconds <= 0 || strings.TrimSpace(r.Why) == "" {
			t.Errorf("row %+v: want a package under its module, its measured serial seconds, and a reason", r)
		}
		// Dir and Module name one module twice; the module's own go.mod is what holds them together.
		mod, err := os.ReadFile(filepath.Join("..", "..", r.Dir, "go.mod"))
		if err != nil || !strings.Contains(string(mod), "module "+r.Module+"\n") {
			t.Errorf("row %+v: %s/go.mod does not declare module %s (%v)", r, r.Dir, r.Module, err)
		}
		if !testsplit.Splits(r.Dir) {
			t.Errorf("Splits(%q) is false for a directory a row names", r.Dir)
		}
		if len(testsplit.For(r.Module)) == 0 {
			t.Errorf("For(%q) does not return the row that names it", r.Module)
		}
	}
	if testsplit.Splits("scripts") {
		t.Error("Splits is true for a module with no row")
	}
	if rows := testsplit.For("no/such/module"); len(rows) != 0 {
		t.Errorf("For returned %d rows for a module the table does not name", len(rows))
	}
}

func TestDealDropsNoNameAndRepeatsNone(t *testing.T) {
	names := []string{"a", "b", "c", "d", "e", "f", "g"}
	for _, n := range []int{1, 2, 3, 7, 20} {
		seen := map[string]int{}
		shares := deal(names, n)
		for _, s := range shares {
			if len(s) == 0 {
				t.Errorf("n=%d: an empty share would run a process whose -test.run matches nothing", n)
			}
			for _, name := range s {
				seen[name]++
			}
		}
		for _, name := range names {
			if seen[name] != 1 {
				t.Errorf("n=%d: %s was dealt %d times", n, name, seen[name])
			}
		}
		if want := min(n, len(names)); len(shares) != want {
			t.Errorf("n=%d: %d shares, want %d", n, len(shares), want)
		}
	}
}

func TestTestNamesKeepsTestsFuzzAndExamplesOnly(t *testing.T) {
	got := testNames("TestA\nsome line a TestMain printed\nFuzzB\nExampleC\n\nok\n")
	if strings.Join(got, ",") != "TestA,FuzzB,ExampleC" {
		t.Errorf("names = %v", got)
	}
}

// The plain path's tally counts every top-level verdict; reading them is testjson's.
func TestTallyCountsEveryVerdict(t *testing.T) {
	p := &testjson.Package{Tests: map[string]string{"TestA": "pass", "TestB": "pass", "TestC": "skip", "TestD": "fail"}}
	if got := tallyOf(p); got != (tally{pass: 2, skip: 1, fail: 1}) {
		t.Errorf("tally = %+v", got)
	}
}

// A process that died before it printed anything the stream claims still shows what it said.
func TestOutputFallsBackToTheRawBytes(t *testing.T) {
	if got := outputOr(nil, []byte("raw")); got != "raw" {
		t.Errorf("no package: %q", got)
	}
	if got := outputOr(&testjson.Package{Output: " \n"}, []byte("raw")); got != "raw" {
		t.Errorf("blank output: %q", got)
	}
	if got := outputOr(&testjson.Package{Output: "said"}, []byte("raw")); got != "said" {
		t.Errorf("package output: %q", got)
	}
}

// -count=1 ON BOTH PATHS. The gate's own arguments no longer show it, so this is where a
// cached PASS is refused.
func TestBothPathsDefeatTheTestCache(t *testing.T) {
	if args := testsplit.UnsplitArgs("10m", 4, []string{"./a"}); !has(args, "-count=1") || !has(args, "-json") || !has(args, "-timeout=10m") || !has(args, "./a") {
		t.Errorf("the plain path runs %v", args)
	}
	if args := testsplit.SplitArgs("10m", []string{"TestA", "TestB"}); !has(args, "-test.count=1") || !has(args, "-test.timeout=10m") || !has(args, "^(TestA|TestB)$") {
		t.Errorf("a split process runs %v", args)
	}
}

func has(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}

func TestDefaultProcesses(t *testing.T) {
	for cpus, want := range map[int]int{1: 1, 2: 1, 4: 1, 6: 3, 16: 8, 64: 8} {
		if got := defaultProcesses(cpus); got != want {
			t.Errorf("%d cores: %d processes, want %d", cpus, got, want)
		}
	}
}
