package checkpoint

import (
	"strings"
	"testing"
)

// The same fixture gray-area's claims suite holds (conformance_test.go there),
// restated on purpose: the heading rule is restated across the module boundary,
// so the fixture that proves the two readers agree is restated with it.
var loopHeadings = map[string]HeadingMatch{
	"Validation loop": Named,
	"Validation loop (all last run at ed3bc3ae)": Named,
	"validation LOOP — per branch":               Named,
	"Validation loop        ← load-bearing":      Named,
	"Validation loop:":                           Named,
	"Validation Loops":                           NearlyNamed,
	"validation-loop":                            NearlyNamed,
	"ValidationLoop":                             NearlyNamed,
	"Validation  loop":                           NearlyNamed,
	"Committed; running the validation loop":     NotNamed,
	"Validator notes":                            NotNamed,
}

func TestMatchHeadingAgreesWithTheSharedFixture(t *testing.T) {
	for h, want := range loopHeadings {
		if got := MatchHeading(h, "Validation loop"); got != want {
			t.Errorf("MatchHeading(%q) = %d, want %d", h, got, want)
		}
	}
}

// A suffixed heading is the section. Measured on the development box: 7 of 58
// real loop headings carry a suffix, and the exact-match lookup read every one
// of those notes as having no loop — restored nothing, watched nothing.
func TestASuffixedHeadingIsTheSection(t *testing.T) {
	n := Parse("---\nschema: 3\n---\n## Validation loop (all last run at ed3bc3ae)\n1. `go test ./...` · re-armed by: tools/\n\n## Next intended steps ← each item\n1. do x\n")
	if body, ok := n.NonEmptySection("Validation loop"); !ok || !strings.Contains(body, "go test") {
		t.Errorf("loop = %q, %v", body, ok)
	}
	if body, ok := n.NonEmptySection("Next intended steps"); !ok || !strings.Contains(body, "do x") {
		t.Errorf("next steps = %q, %v", body, ok)
	}
}

// A near miss is said, by every reader that would otherwise say "no loop".
func TestANearMissLoopHeadingIsAProblem(t *testing.T) {
	note := "---\nschema: 3\n---\n## Validation Loops\n1. `go test ./...` · re-armed by: tools/\n"
	p := NoteLoopProblems(note)
	if len(p) != 1 || !strings.Contains(p[0], `"## Validation Loops"`) || !strings.Contains(p[0], "ABSENT") {
		t.Errorf("problems = %q", p)
	}
	if LoopHeadingProblem(Parse("## Validation loop\n1. x\n## Validation loops (old)\n")) != "" {
		t.Error("a near miss beside the real loop heading was reported as the loop's absence")
	}
}

// Spellings the label pattern did not see: an uppercase letter, and ")". Each
// opens no check in ParseValidationLoop, so each is a problem, as gray-area's
// reader refuses each.
func TestLoopProblemsSeeEverySpellingNoCheckOpensFor(t *testing.T) {
	p := LoopProblems("1. `a`\n7B. `b`\n2) `c`\n2. `d`\n")
	if len(p) != 2 || !strings.Contains(p[0], `"7B."`) || !strings.Contains(p[1], `"2)"`) {
		t.Errorf("problems = %q", p)
	}
}
