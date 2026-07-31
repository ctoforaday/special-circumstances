package checkpoint

import (
	"strings"
	"testing"
)

// The fixture is this repository's own note as it stood on 2026-07-31, because
// the defect was found there and a fixture written from the schema would not
// have contained it. Both violations are present: a loop numbered from `0.`,
// and two lettered sub-entries that open no check in either parser.
const offSpecLoop = `0. ` + "`go test -C plugins/gray-area/tools ./...`" + `  · re-armed by: plugins/gray-area/tools/
   last run: pass
0b. ` + "`gray-area checkpoint`" + ` → adjudicates  · re-armed by: plugins/gray-area/tools/internal/claims/
1. ` + "`go test -C plugins/frank-exchange-of-views/tools ./...`" + `  · re-armed by: plugins/frank-exchange-of-views/tools/
2. ` + "`go test -C plugins/prosthetic-conscience/tools ./...`" + `  · re-armed by: plugins/prosthetic-conscience/tools/
7b. ` + "`gray-area checkpoint`" + ` on a new session id → exits 1  · re-armed by: plugins/gray-area/tools/cmd/
`

func TestTheLoopThatCausedThisIsReportedRatherThanRepaired(t *testing.T) {
	got := LoopProblems(offSpecLoop)
	if len(got) == 0 {
		t.Fatal("the note that produced #192 and #193 reported no problems")
	}

	joined := strings.Join(got, "\n")
	for _, want := range []string{`"0b."`, `"7b."`} {
		if !strings.Contains(joined, want) {
			t.Errorf("lettered entry %s not reported; it opens no check and nothing else says so:\n%s", want, joined)
		}
	}
	// `0.` sits in position 1, `1.` in position 2, `2.` in position 3 — every
	// digit label disagrees with its ordinal, which is exactly the off-by-one
	// that made rearmed.json's key 2 mean the note's `1.`
	for _, want := range []string{`"0."`, `"1."`, `"2."`} {
		if !strings.Contains(joined, want) {
			t.Errorf("label/ordinal mismatch for %s not reported:\n%s", want, joined)
		}
	}

	// Reported, NOT repaired: the parser still opens exactly the checks it
	// always did. Renumbering silently would restore the ambiguity.
	if n := len(ParseValidationLoop(offSpecLoop)); n != 3 {
		t.Errorf("ParseValidationLoop opened %d checks; want 3 unchanged — LoopProblems must not repair", n)
	}
}

func TestAConformingLoopReportsNothing(t *testing.T) {
	conforming := "1. `go test ./a`  · re-armed by: a/\n   last run: pass\n2. `go test ./b`  · re-armed by: b/\n3. `go test ./c`  · re-armed by: c/\n"
	if got := LoopProblems(conforming); len(got) != 0 {
		t.Errorf("a conforming loop reported %v; a check that cries wolf gets ignored when it matters", got)
	}
}

func TestProseIsNotMistakenForAnEntry(t *testing.T) {
	// Continuation lines carry version numbers, dates and issue references. A
	// checker that flagged those would be noise, and noise is how the real
	// report gets skipped.
	body := "1. `go test ./a`  · re-armed by: a/\n   last run: pass 2026-07-30T17:24Z on 274ffac\n   see #165 and 4.  the fourth reason\n"
	if got := LoopProblems(body); len(got) != 0 {
		t.Errorf("prose inside a check was read as an entry: %v", got)
	}
}

func TestAGapIsReportedAtTheFirstEntryThatMoved(t *testing.T) {
	// 1, 2, 4 — deleting a check without renumbering is the likeliest way a
	// conforming note goes off-spec later, and it silently re-points every
	// existing rearmed.json key from that point on.
	body := "1. `a`  · re-armed by: a/\n2. `b`  · re-armed by: b/\n4. `d`  · re-armed by: d/\n"
	got := LoopProblems(body)
	if len(got) != 1 || !strings.Contains(got[0], `"4."`) {
		t.Errorf("want exactly one problem naming \"4.\"; got %v", got)
	}
	if !strings.Contains(got[0], "position 3") {
		t.Errorf("the report must name the position it actually occupies, since that is the key state is stored under; got %q", got[0])
	}
}
