package testjson

import (
	"strings"
	"testing"
)

// A VERDICT IS THE EVENT, NOT A LINE. TestA prints a FAIL line and passes; TestB's output has
// no trailing newline, so the `--- PASS` text that follows it is glued to it; the events still
// name both. A subtest's verdict is not its parent's.
func TestReadTakesVerdictsFromEventsPerPackage(t *testing.T) {
	stream := strings.Join([]string{
		`{"Action":"output","Package":"p","Test":"TestA","Output":"--- FAIL: TestA (0.00s)\n"}`,
		`{"Action":"pass","Package":"p","Test":"TestA"}`,
		`{"Action":"output","Package":"p","Test":"TestB","Output":"no newline--- PASS: TestB (0.00s)\n"}`,
		`{"Action":"pass","Package":"p","Test":"TestB"}`,
		`{"Action":"fail","Package":"p","Test":"TestB/sub"}`,
		`{"Action":"pass","Package":"p","Test":"TestE/sub"}`,
		`{"Action":"skip","Package":"p","Test":"TestC"}`,
		`{"Action":"output","Package":"q","Output":"q says\n"}`,
		`{"Action":"fail","Package":"q","Test":"TestD"}`,
		`{"Action":"fail","Package":"q","Elapsed":1.5}`,
		`{"Action":"pass","Package":"p","Elapsed":0.25}`,
	}, "\n")
	s := Read([]byte(stream))
	p, q := s.Packages["p"], s.Packages["q"]
	if p == nil || q == nil || len(s.Packages) != 2 {
		t.Fatalf("packages = %v", s.Packages)
	}
	want := map[string]string{"TestA": "pass", "TestB": "pass", "TestC": "skip"}
	if len(p.Tests) != len(want) {
		t.Errorf("p.Tests = %v, want %v — a subtest is not a top-level test", p.Tests, want)
	}
	for name, v := range want {
		if p.Tests[name] != v {
			t.Errorf("p.Tests[%s] = %q, want %q — a printed line is not a verdict, and a subtest's is not its parent's", name, p.Tests[name], v)
		}
	}
	if p.Verdict != "pass" || p.Elapsed.Seconds() != 0.25 {
		t.Errorf("p: verdict %q elapsed %v", p.Verdict, p.Elapsed)
	}
	if q.Verdict != "fail" || q.Tests["TestD"] != "fail" || q.Elapsed.Seconds() != 1.5 {
		t.Errorf("q = %+v", q)
	}
	if !strings.Contains(p.Output, "no newline--- PASS") || strings.Contains(p.Output, "q says") || q.Output != "q says\n" {
		t.Errorf("output is not kept per package: p=%q q=%q", p.Output, q.Output)
	}
	if s.Stray != "" {
		t.Errorf("stray = %q from a stream of events only", s.Stray)
	}
}

// Everything worth printing that is not a package's own output is kept as stray: a line that is
// not an event, the compiler's build-output, and output no package claims.
func TestReadKeepsWhatIsNotAPackagesOutputAsStray(t *testing.T) {
	stream := strings.Join([]string{
		`go: downloading something`,
		`{"ImportPath":"q [q.test]","Action":"build-output","Output":"q/q_test.go:3:2: undefined: undeclared\n"}`,
		`{"ImportPath":"q [q.test]","Action":"build-fail"}`,
		`{"Action":"output","Output":"WARNING: DATA RACE\n"}`,
		`{"Action":"output","Package":"q","Output":"FAIL\tq [build failed]\n"}`,
		`{"Action":"fail","Package":"q"}`,
		`# q [build failed]`,
	}, "\n")
	s := Read([]byte(stream))
	for _, want := range []string{"go: downloading something", "undefined: undeclared", "WARNING: DATA RACE", "# q [build failed]"} {
		if !strings.Contains(s.Stray, want) {
			t.Errorf("stray lacks %q: %q", want, s.Stray)
		}
	}
	if q := s.Packages["q"]; q == nil || q.Verdict != "fail" || strings.Contains(s.Stray, "FAIL\tq") {
		t.Errorf("q = %+v; stray = %q", q, s.Stray)
	}
}

// EVERY ASKED-FOR NAME ANSWERS, or it is Missing. A package that sent nothing answers nothing.
func TestAnswerSortsEveryNameAndMissesWhatDidNotAnswer(t *testing.T) {
	stream := strings.Join([]string{
		`{"Action":"pass","Package":"p","Test":"TestA"}`,
		`{"Action":"pass","Package":"p","Test":"TestE/sub"}`,
		`{"Action":"skip","Package":"p","Test":"TestB"}`,
		`{"Action":"fail","Package":"p","Test":"TestC"}`,
		`{"Action":"pass","Package":"other","Test":"TestD"}`,
	}, "\n")
	s := Read([]byte(stream))
	a := s.Answer("p", []string{"TestA", "TestB", "TestC", "TestD", "TestE"})
	if got := strings.Join(a.Pass, ","); got != "TestA" {
		t.Errorf("pass = %s", got)
	}
	if strings.Join(a.Skip, ",") != "TestB" || strings.Join(a.Fail, ",") != "TestC" {
		t.Errorf("skip = %v fail = %v", a.Skip, a.Fail)
	}
	// TestD answered in another package; TestE only through a subtest.
	if got := strings.Join(a.Missing, ","); got != "TestD,TestE" {
		t.Errorf("missing = %s", got)
	}
	if a := s.Answer("absent", []string{"TestA"}); strings.Join(a.Missing, ",") != "TestA" || len(a.Pass) != 0 {
		t.Errorf("a package with no events answered: %+v", a)
	}
}
