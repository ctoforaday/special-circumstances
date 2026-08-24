package main

import (
	"strings"
	"testing"
)

// codeMask is the correctness core of the whole tool. If it under-masks, operators inside
// strings get mutated, the test fails on changed message text rather than on a branch, and
// the mutant is scored "killed" — silently INFLATING the kill rate with mutants that never
// tested anything. If it over-masks, real code stops being swept and the tool reports a
// perfect score for work it did not do. Both directions are flattering, which is why this
// is tested at the unit rather than trusted through -selftest alone.
func TestCodeMask(t *testing.T) {
	// Asserted as SURVIVING OPERATOR COUNTS rather than as exact mask strings. That is
	// the contract candidates() actually consumes, and hand-counted space padding in an
	// expectation is a test that fails for reasons about the test.
	cases := []struct {
		name    string
		line    string
		inBlock bool
		// wantOps: how many of each operator remain visible as real code.
		wantOps   map[string]int
		wantBlock bool
	}{
		{"plain code untouched", "if a == b {", false, map[string]int{"==": 1}, false},
		{"trailing comment blanked", "if a == b { // x != y", false, map[string]int{"==": 1, "!=": 0}, false},
		{"whole-line comment blanked", "// a == b", false, map[string]int{"==": 0}, false},
		{"string contents blanked", `s := "a != b"`, false, map[string]int{"!=": 0}, false},
		{"code around a string survives", `if s == "x != y" && b {`, false, map[string]int{"==": 1, "&&": 1, "!=": 0}, false},
		{"escaped quote does not end the string early", `s := "a \" b != c"`, false, map[string]int{"!=": 0}, false},
		{"raw string blanked", "s := `a != b`", false, map[string]int{"!=": 0}, false},
		{"rune literal blanked", "if c == '>' && d {", false, map[string]int{"==": 1, "&&": 1}, false},
		{"block comment opens and swallows the rest", "a := 1 /* x != y", false, map[string]int{"!=": 0}, true},
		{"inside a block comment", "still != commented", true, map[string]int{"!=": 0}, true},
		{"block comment closes then code", "x */ if a == b {", true, map[string]int{"==": 1}, false},
		{"two operators on one line", "if a == b && c != d {", false, map[string]int{"==": 1, "&&": 1, "!=": 1}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, block := codeMask(c.line, c.inBlock)
			// Length parity is load-bearing: candidates() indexes the MASK and then
			// mutates the LINE at that index.
			if len(got) != len(c.line) {
				t.Fatalf("mask length %d != line length %d — an index into the mask must address the same byte in the line", len(got), len(c.line))
			}
			for op, want := range c.wantOps {
				if n := strings.Count(got, op); n != want {
					t.Errorf("codeMask(%q) = %q: %q appears %d time(s), want %d", c.line, got, op, n, want)
				}
			}
			if block != c.wantBlock {
				t.Errorf("block state = %v, want %v", block, c.wantBlock)
			}
		})
	}
}

// The bug that OOM-killed the first version: when strings.Index returns -1 the obvious
// index arithmetic leaves the offset unchanged and the loop appends forever. A line with an
// operator that appears once is the minimal reproducer, so it gets a test rather than a
// comment.
func TestCandidatesTerminatesAndFindsEveryOccurrence(t *testing.T) {
	got := candidates("if a == b && c == d {\n")
	// two `==` and one `&&`
	if len(got) != 3 {
		t.Fatalf("expected 3 candidates, got %d: %+v", len(got), got)
	}
	seen := map[string]int{}
	for _, c := range got {
		seen[c.from]++
		if c.line != 0 {
			t.Errorf("candidate on the wrong line: %+v", c)
		}
	}
	if seen["=="] != 2 || seen["&&"] != 1 {
		t.Errorf("occurrence counts wrong: %v", seen)
	}
}

func TestCandidatesIgnoresStringsAndComments(t *testing.T) {
	src := "// if a == b\n" +
		"s := \"x != y\"\n" +
		"t := `p >= q`\n" +
		"/* r <= s */\n" +
		"if real == code {\n"
	got := candidates(src)
	if len(got) != 1 {
		t.Fatalf("only the real comparison may be a candidate, got %d: %+v", len(got), got)
	}
	if got[0].line != 4 || got[0].from != "==" {
		t.Errorf("wrong candidate: %+v", got[0])
	}
}

// suiteFor names the package that OWNS a file — the first suite a mutant is tried against.
//
// It used to return `./...` for anything outside cmd/, on the argument that an internal/ mutant
// can affect anything. That argument is true and it made the tool unusable: the whole module ran
// per mutant. The widening did not disappear — it moved behind `-confirm`, so a SURVIVOR can be
// re-tested against the rest of the module on request, and the common case (a mutant its own
// package kills) costs milliseconds.
func TestSuiteFor(t *testing.T) {
	for in, want := range map[string]string{
		"cmd/sc-doctor/main.go":       "./cmd/sc-doctor/",
		"cmd/sc-doctor/sub/helper.go": "./cmd/sc-doctor/sub/",
		"internal/secrets/secrets.go": "./internal/secrets/",
		"internal/a/b/c.go":           "./internal/a/b/",
		"main.go":                     "./",
		"cmd":                         "./",
	} {
		if got := suiteFor(in); got != want {
			t.Errorf("suiteFor(%q) = %q, want %q", in, got, want)
		}
	}
}
