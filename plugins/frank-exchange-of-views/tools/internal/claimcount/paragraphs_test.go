package claimcount

import "testing"

// The paragraph rule, pinned per exclusion: each case isolates one shape and says what it counts.
func TestParagraphsCountsBlankLineBlocksThatCarryProse(t *testing.T) {
	for _, c := range []struct {
		name, md string
		want     int
	}{
		{"empty", "", 0},
		{"one line", "Water is wet.", 1},
		{"two blocks", "One.\n\nTwo.", 2},
		{"a wrapped paragraph is one", "One sentence\nwrapped over two lines.", 1},
		{"a whitespace-only line separates", "One.\n   \t\nTwo.", 2},
		{"runs of blank lines are one separator", "One.\n\n\n\nTwo.\n\n", 2},
		{"a heading is not a paragraph", "# Title\n\n## Section\n\nBody.", 1},
		{"a heading over its text counts the text once", "## Section\nBody.", 1},
		{"an anchor-only line is not prose", "<!--cite:c-1-->\n\n- <!--fx:f-2-->\n\nBody<!--proof:p-3-->.", 1},
		{"a footnote definition is not prose", "[^L1]: https://example.org\n\nBody.", 1},
		{"a fence is not prose, blank lines inside it included", "```\ncode\n\nmore code\n```\n\nBody.", 1},
		{"a fence under prose joins its paragraph", "Body:\n```\nx\n\ny\n```", 1},
		{"a tilde fence too", "~~~\nx\n\ny\n~~~", 0},
		{"a tight list is one block", "- one\n- two\n- three", 1},
		{"a loose list counts per item", "- one\n\n- two\n\n- three", 3},
		{"a table is one block", "| a | b |\n|---|---|\n| 1 | 2 |", 1},
		{"a rule is not prose", "One.\n\n---\n\nTwo.", 2},
		{"digits are prose", "42", 1},
	} {
		if got := Paragraphs(c.md); got != c.want {
			t.Errorf("%s: Paragraphs(%q) = %d, want %d", c.name, c.md, got, c.want)
		}
	}
}
