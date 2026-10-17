package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/racescope"
)

const (
	fixture       = "testdata/fixture"
	fixtureTarget = "./cmd/app"
)

// rowFixtures names, for each row of the table, the fixture test that only that row selects.
// Hand-written and NOT derived from the table: a list built from the table would lose a row's
// fixture in the same edit that lost the row, and the deletion would read as a smaller, green set.
var rowFixtures = map[string]string{
	"(*testing.T).Parallel":                  "TestRowParallel",
	"net/http/httptest.NewServer":            "TestRowHTTPTestNewServer",
	"net/http/httptest.NewTLSServer":         "TestRowHTTPTestNewTLSServer",
	"net/http/httptest.NewUnstartedServer":   "TestRowHTTPTestNewUnstartedServer",
	"(*net/http.Server).Serve":               "TestRowServerServe",
	"(*net/http.Server).ServeTLS":            "TestRowServerServeTLS",
	"(*net/http.Server).ListenAndServe":      "TestRowServerListenAndServe",
	"(*net/http.Server).ListenAndServeTLS":   "TestRowServerListenAndServeTLS",
	"net/http.Serve":                         "TestRowHTTPServe",
	"net/http.ServeTLS":                      "TestRowHTTPServeTLS",
	"net/http.ListenAndServe":                "TestRowHTTPListenAndServe",
	"net/http.ListenAndServeTLS":             "TestRowHTTPListenAndServeTLS",
	"time.AfterFunc":                         "TestRowTimeAfterFunc",
	"os/signal.Notify":                       "TestRowSignalNotify",
	"context.AfterFunc":                      "TestRowContextAfterFunc",
	"(*golang.org/x/sync/errgroup.Group).Go": "TestRowErrgroupGo",
	"(*sync.WaitGroup).Go":                   "TestRowWaitGroupGo",
}

// reachFixtures are the fixture tests the reachability rule selects, one per edge it follows.
var reachFixtures = []string{
	"ExampleSpawn",             // an Example is a root
	"FuzzSpawns",               // a Fuzz function is a root
	"TestCallsASupportPackage", // a direct call into a non-test file outside the scope
	"TestCallsATestHelper",     // a direct call into a test file
	"TestDirectGo",             // a go statement in the test
	"TestGoInsideALiteral",     // a go statement inside a function literal
	"TestOwnsSpawn",            // a directive
	"TestOwnsStart",            // a directive naming a method
}

func deriveFixture(t *testing.T, dir string) *derived {
	t.Helper()
	d, err := derive(dir, fixtureTarget)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func namesOf(d *derived, pkg string) []string {
	names := setNames(d.sets[pkg])
	sort.Strings(names)
	return names
}

// THE SET IS EXACT. Every fixture test the rule selects is in it and nothing else is — the four
// that reach a start only through shipped source, a function value or an interface stay out,
// and so does the one that starts nothing.
func TestTheRaceSetIsExactlyWhatTheRuleSelects(t *testing.T) {
	d := deriveFixture(t, fixture)
	if len(d.problems) > 0 {
		t.Fatalf("the fixture is refused as it stands: %v", d.problems)
	}
	want := append([]string{}, reachFixtures...)
	for _, name := range rowFixtures {
		want = append(want, name)
	}
	sort.Strings(want)
	if got := namesOf(d, "fixture/shipped"); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("race set:\n got %v\nwant %v", got, want)
	}
	if len(d.sets) != 1 {
		t.Errorf("tests were selected in %d packages, want only fixture/shipped", len(d.sets))
	}
	// Named, because each is a rule a later edit could quietly reverse.
	got := strings.Join(namesOf(d, "fixture/shipped"), " ")
	for name, why := range map[string]string{
		"TestCallsAShippedMethod":   "a call to a shipped-source METHOD that starts a goroutine must not select its caller — followed into source, command construction selects a whole package",
		"TestCallsAShippedFunction": "a call to a shipped-source function that starts a goroutine must not select its caller",
		"TestThroughAFunctionValue": "a function value is not an edge",
		"TestThroughAnInterface":    "an interface call is not an edge",
		"TestPlain":                 "a test that starts nothing is not raced",
	} {
		if strings.Contains(" "+got+" ", " "+name+" ") {
			t.Errorf("%s is in the race set: %s", name, why)
		}
	}
}

// Every row has a fixture, and every row's fixture is selected. Delete a row and its fixture
// leaves the set; add one without a fixture and this says so.
func TestEveryRowOfTheTableSelectsItsFixture(t *testing.T) {
	d := deriveFixture(t, fixture)
	set := map[string]string{}
	for _, s := range d.sets["fixture/shipped"] {
		set[s.name] = s.why
	}
	for callee, name := range rowFixtures {
		if _, ok := racescope.Lookup(callee); !ok {
			t.Errorf("the table has no row for %s, and %s exists to exercise one", callee, name)
		}
		if why, ok := set[name]; !ok {
			t.Errorf("%s is not in the race set, so a call to %s selects nothing", name, callee)
		} else if !strings.Contains(why, callee) {
			t.Errorf("%s was selected for %q, not for its own row %s — the fixture proves nothing about that row", name, why, callee)
		}
	}
	for _, row := range racescope.Table {
		if rowFixtures[row.Callee] == "" {
			t.Errorf("row %s has no fixture test; a row nothing exercises can stop matching unnoticed", row.Callee)
		}
		if strings.TrimSpace(row.Why) == "" {
			t.Errorf("row %s carries no reason", row.Callee)
		}
	}
}

// The starts a raced test must execute: both in shipped source, and the support package's,
// because a raced binary links it.
func TestTheOwedStartsAreTheNonTestOnes(t *testing.T) {
	d := deriveFixture(t, fixture)
	var got []string
	for _, s := range d.owed {
		got = append(got, s.owner())
	}
	want := "shipped/shipped.go:Spawn shipped/shipped.go:Svc.Start support/support.go:Launch"
	if strings.Join(got, " ") != want {
		t.Errorf("owed starts = %v, want %s", got, want)
	}
	if links := strings.Join(d.links["fixture/shipped"], " "); links != "fixture/shipped fixture/support" {
		t.Errorf("the evidence run would cover %q, want the package and the support package it links", links)
	}
}

// mutated copies the fixture and applies one edit to one file.
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

func TestWhatIsRefusedBeforeAnythingRuns(t *testing.T) {
	for _, c := range []struct {
		name, file, old, new string
		want                 []string
	}{
		{"a shipped start no test declares",
			"shipped/shipped_test.go", "//racescope:covers shipped/shipped.go:Spawn\n", "",
			[]string{"shipped/shipped.go:7", "no test declares"}},
		{"a new go statement in shipped source",
			"shipped/shipped.go", "func Plain() int { return 1 }", "func Plain() int { go func() {}(); return 1 }",
			[]string{"shipped/shipped.go:21", "in Plain", "no test declares"}},
		{"a directive naming a function with no start",
			"shipped/shipped_test.go", "//racescope:covers shipped/shipped.go:Spawn\n", "//racescope:covers shipped/shipped.go:Plain\n",
			[]string{"TestOwnsSpawn declares it covers shipped/shipped.go:Plain", "no goroutine starts there"}},
		{"a directive on something that is not a test",
			"shipped/shipped_test.go", "func helperSpawns()", "//racescope:covers shipped/shipped.go:Spawn\nfunc helperSpawns()",
			[]string{"helperSpawns carries a racescope:covers directive and is not a test"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			d := deriveFixture(t, mutated(t, c.file, c.old, c.new))
			all := strings.Join(d.problems, "\n")
			if all == "" {
				t.Fatal("nothing was refused")
			}
			for _, w := range c.want {
				if !strings.Contains(all, w) {
					t.Errorf("the refusal does not say %q:\n%s", w, all)
				}
			}
		})
	}
}

// A go statement added to a test outside the set puts that test in the next set, wherever in
// the body it sits.
func TestAStartAddedToATestSelectsIt(t *testing.T) {
	for name, body := range map[string]string{
		"at the top":       "func TestPlain(t *testing.T) {\n\tgo func() {}()\n",
		"inside a literal": "func TestPlain(t *testing.T) {\n\tt.Cleanup(func() { go func() {}() })\n",
	} {
		t.Run(name, func(t *testing.T) {
			d := deriveFixture(t, mutated(t, "shipped/shipped_test.go", "func TestPlain(t *testing.T) {\n", body))
			if !strings.Contains(" "+strings.Join(namesOf(d, "fixture/shipped"), " ")+" ", " TestPlain ") {
				t.Error("TestPlain starts a goroutine now and is not in the race set")
			}
		})
	}
}

// A module that does not type-check resolves some calls to nothing. That is an error, never a
// smaller set.
func TestAModuleThatDoesNotLoadIsNotMeasured(t *testing.T) {
	dir := mutated(t, "shipped/shipped_test.go", "func TestPlain(t *testing.T) {\n", "func TestPlain(t *testing.T) {\n\tundeclared()\n")
	if _, err := derive(dir, fixtureTarget); err == nil || !strings.Contains(err.Error(), "not measured") {
		t.Errorf("derive = %v, want a refusal that says the set is not measured", err)
	}
}

func TestTheGateRunsGreenOnTheFixture(t *testing.T) {
	var out strings.Builder
	if rc := run(&out, fixture, fixtureTarget, "5m", 1, false); rc != 0 {
		t.Fatalf("exit %d:\n%s", rc, out.String())
	}
	for _, want := range []string{
		"shipped — 25 selected: 25 pass, 0 fail, 0 skip",
		"executed: shipped/shipped.go:7 (go statement, in Spawn)",
		"executed: shipped/shipped.go:16 (go statement, in Svc.Start)",
		"executed: support/support.go:7 (go statement, in Launch)",
		"racescope: OK — 25 test(s) raced, 3 start(s) executed",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the report lacks %q:\n%s", want, out.String())
		}
	}
}

// The gate's reds, each driven through the real runs on a mutated copy of the fixture.
func TestTheGateIsRed(t *testing.T) {
	for _, c := range []struct {
		name, file, old, new string
		want                 []string
	}{
		// THE CHECK THAT THE VERDICT RUN CARRIES NO COVERAGE: with a profile in the same run
		// the detector does not report this.
		{"on a data race in a raced test",
			"shipped/shipped_test.go",
			"func TestDirectGo(t *testing.T) {\n\tc := make(chan int)\n\tgo func() { c <- 1 }()\n\t<-c\n}",
			"func TestDirectGo(t *testing.T) {\n\tn := 0\n\tc := make(chan int)\n\tgo func() { n++; c <- 1 }()\n\tn++\n\t<-c\n}",
			[]string{"WARNING: DATA RACE", "failed: TestDirectGo"}},
		{"when the declaring test stops executing its start",
			"shipped/shipped_test.go", "func TestOwnsSpawn(t *testing.T) { <-Spawn() }", "func TestOwnsSpawn(t *testing.T) {}",
			[]string{"shipped/shipped.go:7 (go statement, in Spawn): no raced test executed this start"}},
		{"on a selected Example that cannot run",
			"shipped/shipped_test.go", "\t// Output: 1\n", "",
			[]string{"NOT MEASURED, no verdict came back for: ExampleSpawn"}},
		{"on a selected test that fails",
			"shipped/shipped_test.go", "func TestCallsATestHelper(t *testing.T) { <-helperSpawns() }",
			"func TestCallsATestHelper(t *testing.T) { <-helperSpawns(); t.Fatal(\"planted\") }",
			[]string{"failed: TestCallsATestHelper", "planted"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			var out strings.Builder
			if rc := run(&out, mutated(t, c.file, c.old, c.new), fixtureTarget, "5m", 1, false); rc != 1 {
				t.Fatalf("exit %d, want 1:\n%s", rc, out.String())
			}
			for _, w := range c.want {
				if !strings.Contains(out.String(), w) {
					t.Errorf("the report lacks %q:\n%s", w, out.String())
				}
			}
		})
	}
}

// A skipped concurrency test was not raced. It does not fail the gate; it is printed by name.
func TestASkippedSelectedTestIsNamed(t *testing.T) {
	dir := mutated(t, "shipped/shipped_test.go", "func TestDirectGo(t *testing.T) {\n", "func TestDirectGo(t *testing.T) {\n\tt.Skip(\"planted\")\n")
	var out strings.Builder
	if rc := run(&out, dir, fixtureTarget, "5m", 1, false); rc != 0 {
		t.Fatalf("exit %d:\n%s", rc, out.String())
	}
	if want := "24 pass, 0 fail, 1 skip (skipped, so NOT raced: TestDirectGo)"; !strings.Contains(out.String(), want) {
		t.Errorf("the report lacks %q:\n%s", want, out.String())
	}
}

// The verdict is red on a failed or missing name, a reported race, or a failed `go test`. How
// each name answered is testjson's to assert; this is what racescope makes of it.
func TestReadVerdictNeedsAnAnswerForEveryName(t *testing.T) {
	stream := strings.Join([]string{
		`{"Action":"run","Package":"p","Test":"TestA"}`,
		`{"Action":"output","Package":"p","Test":"TestA","Output":"hello\n"}`,
		`{"Action":"pass","Package":"p","Test":"TestA"}`,
		`{"Action":"skip","Package":"p","Test":"TestB"}`,
		`{"Action":"fail","Package":"p","Test":"TestC"}`,
		`{"Action":"fail","Package":"p"}`,
		`not json at all`,
	}, "\n")
	v := readVerdict("p", []string{"TestA", "TestB", "TestC", "TestD"}, []byte(stream), nil)
	if v.pass != 1 || strings.Join(v.skipped, ",") != "TestB" || strings.Join(v.failed, ",") != "TestC" || strings.Join(v.missing, ",") != "TestD" {
		t.Errorf("verdict = %+v", v)
	}
	if !v.red() || v.race {
		t.Errorf("red=%v race=%v, want red without a race", v.red(), v.race)
	}
	// The printed output is the package's AND what was not its events.
	if !strings.Contains(v.output, "hello") || !strings.Contains(v.output, "not json at all") {
		t.Errorf("output dropped a line: %q", v.output)
	}
	if v := readVerdict("p", []string{"TestA"}, []byte(`{"Action":"pass","Package":"other","Test":"TestA"}`), nil); len(v.missing) != 1 || !v.red() {
		t.Errorf("another package's pass answered for the raced one: %+v", v)
	}
	// All answered and passing, and the detector still spoke: red — whether its report arrived
	// as the package's output or as a line that was not an event.
	for _, race := range []string{`{"Action":"output","Package":"p","Output":"WARNING: DATA RACE\n"}`, `WARNING: DATA RACE`} {
		stream := race + "\n" + `{"Action":"pass","Package":"p","Test":"TestA"}`
		if v := readVerdict("p", []string{"TestA"}, []byte(stream), nil); !v.race || !v.red() {
			t.Errorf("a reported race did not turn the verdict red: %+v", v)
		}
	}
	if v := readVerdict("p", []string{"TestA"}, []byte(`{"Action":"pass","Package":"p","Test":"TestA"}`), errors.New("exit status 1")); !v.red() {
		t.Error("a failed go test with every name passing read as green")
	}
	if v := readVerdict("p", []string{"TestA"}, []byte(`{"Action":"pass","Package":"p","Test":"TestA"}`), nil); v.red() {
		t.Errorf("a clean run read as red: %+v", v)
	}
}

func TestAProfileThatCannotBeReadIsAnError(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	good := write("good", "mode: set\nfixture/shipped/shipped.go:6.25,9.2 3 1\n")
	if blocks, err := readProfiles([]string{good}); err != nil || len(blocks["fixture/shipped/shipped.go"]) != 1 {
		t.Fatalf("a well-formed profile: blocks=%v err=%v", blocks, err)
	}
	for name, p := range map[string]string{
		"missing":        filepath.Join(dir, "absent"),
		"empty":          write("empty", ""),
		"header only":    write("header", "mode: set\n"),
		"no header":      write("noheader", "fixture/shipped/shipped.go:6.25,9.2 3 1\n"),
		"a broken line":  write("broken", "mode: set\nfixture/shipped/shipped.go:6.25 3 1\n"),
		"a broken count": write("count", "mode: set\nfixture/shipped/shipped.go:6.25,9.2 3 many\n"),
	} {
		if _, err := readProfiles([]string{good, p}); err == nil {
			t.Errorf("%s: read without error — the starts would be judged against nothing", name)
		}
	}
}

func TestProveTellsUnexecutedFromUnmeasured(t *testing.T) {
	blocks := map[string][]block{
		"m/p/a.go": {{sl: 5, sc: 10, el: 9, ec: 2, count: 1}, {sl: 12, sc: 10, el: 15, ec: 2, count: 0}},
	}
	at := func(line, col int) start {
		return start{pkg: "m/p", file: "p/a.go", line: line, col: col, fn: "F", what: "go statement"}
	}
	proved, unproven := prove([]start{at(7, 2), at(13, 2), at(20, 2), at(5, 9), at(9, 2)}, blocks)
	if len(proved) != 1 || proved[0].line != 7 {
		t.Errorf("proved = %v, want only line 7", proved)
	}
	all := strings.Join(unproven, "\n")
	for _, want := range []string{
		"p/a.go:13 (go statement, in F): no raced test executed this start",
		"p/a.go:20 (go statement, in F): not measured",
		"p/a.go:5 (go statement, in F): not measured", // one column before the block opens
		"p/a.go:9 (go statement, in F): not measured", // the block's end is exclusive
	} {
		if !strings.Contains(all, want) {
			t.Errorf("unproven lacks %q:\n%s", want, all)
		}
	}
	// The same block counted in two profiles: executed in either is executed.
	blocks["m/p/a.go"] = append(blocks["m/p/a.go"], block{sl: 12, sc: 10, el: 15, ec: 2, count: 1})
	if proved, _ := prove([]start{at(13, 2)}, blocks); len(proved) != 1 {
		t.Error("a start executed by one of two binaries read as unexecuted")
	}
}
