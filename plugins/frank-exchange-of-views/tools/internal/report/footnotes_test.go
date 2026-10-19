package report

import (
	"slices"
	"testing"
)

// A DANGLING REFERENCE IN PROSE IS FOUND WHEREVER CODE AROUND IT IS SHAPED. Code is no footnote,
// and the audit must not read the prose between two code blocks as code: a four-backtick fence
// that holds a three-backtick line, or a lone backtick, paired wrongly would blank the prose after
// it, and the dangling reference there would read as a clean board.
func TestFootnotesInReadsMarkdownNotAPattern(t *testing.T) {
	for _, c := range []struct {
		name, md string
		dangling []string
	}{
		{"defined and resolved", "A claim[^1].\n\n[^1]: a source\n", nil},
		{"undefined", "A claim[^1].\n", []string{"1"}},
		{"after a fence that holds a shorter fence", "````\n```\n````\n\nA claim[^P1].\n", []string{"P1"}},
		// Markdown pairs a backtick with the next run of the same length, so the span is the first
		// two and the reference after them is prose.
		{"backticks pair as markdown pairs them", "a ` tick, then a claim[^P2] and `[^code]`.\n", []string{"code"}},
		{"in a tilde fence", "~~~\n[^x]\n~~~\n", nil},
		{"in an indented block", "Para.\n\n    [^x]\n", nil},
		{"in a code span", "see `[^x]` here\n", nil},
		{"escaped", "a class \\[^x] in prose\n", nil},
		{"inside emphasis", "*claim[^e]*\n", []string{"e"}},
		{"followed by a colon mid-line", "artifacts exist[^P1]: `skills/`\n", []string{"P1"}},
		{"in a heading", "### [^P3] proof\n", []string{"P3"}},
		{"a definition indented in a list is not one", "- item\n\n  [^n]: a note\n\nsee[^n]\n", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, _, got := FootnotesIn(c.md)
			if !slices.Equal(got, c.dangling) {
				t.Errorf("dangling = %v, want %v", got, c.dangling)
			}
		})
	}
}

// The counts are of distinct labels: two references to one note are one reference.
func TestFootnotesInCountsLabels(t *testing.T) {
	refs, defs, dangling := FootnotesIn("a[^1] b[^1] c[^x]\n\n[^1]: one\n")
	if refs != 2 || defs != 1 || !slices.Equal(dangling, []string{"x"}) {
		t.Errorf("refs, defs, dangling = %d, %d, %v; want 2, 1, [x]", refs, defs, dangling)
	}
}
