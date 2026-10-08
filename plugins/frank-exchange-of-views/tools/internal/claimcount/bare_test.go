package claimcount

import (
	"slices"
	"testing"
)

// A BARE ANCHOR IS NOT A CLAIM. An anchor backs the prose before it; one with nothing before it
// in its segment is what an edit leaves when it cuts the anchored sentence away. Each shape below
// is one the real verbs produce — including the one where the bare anchor sits at the head of the
// NEXT sentence, because the splice tidy removed the cut sentence's terminator.
func TestABareAnchorIsNotAClaim(t *testing.T) {
	for _, tc := range []struct {
		name, md string
		count    int
		bare     []string
	}{
		{"attached", "The sky is blue<!--cite:c-1-->.\n", 1, nil},
		{"alone on its line", "<!--cite:c-1-->\n", 0, []string{"c-1"}},
		{"an emptied bullet", "- <!--fx:f-1-->?\n", 0, []string{"f-1"}},
		{"at the head of the next sentence", "<!--cite:c-1--> The grass is green.\n", 0, []string{"c-1"}},
		{"mid-line, its own segment", "One. <!--cite:c-1-->. Two.\n", 0, []string{"c-1"}},
		{"two anchors after prose are both attached", "Prose<!--cite:c-1--><!--cite:c-2-->.\n", 2, nil},
		{"the next sentence's own cite still counts", "<!--cite:c-1--> Grass is green<!--cite:c-2-->.\n", 1, []string{"c-1"}},
		// MARKDOWN AROUND THE TERMINATOR. `blue cite` on "**Water is wet.**" places the anchor
		// after the closing "**" (a "*" is not trailing punctuation); the closer and an anchor flush
		// after it belong to the sentence the terminator ends, so it is attached, not bare.
		{"bold closing a sentence", "**Water is wet.**<!--cite:c-1-->\n", 1, nil},
		{"italic closing a sentence", "_Water is wet._<!--cite:c-1-->\n", 1, nil},
		{"a parenthetical closing a sentence", "Seen (p. 3.)<!--cite:c-1-->\n", 1, nil},
		{"bold in a bullet", "- **Water is wet.**<!--cite:c-1-->\n", 1, nil},
		{"a second anchor after bold is attached too", "**Wet.**<!--cite:c-1--><!--cite:c-2-->\n", 2, nil},
		{"a plain bullet", "- Water is wet<!--cite:c-1-->.\n", 1, nil},
		// ...but whitespace, or nothing at all, between a terminator and the anchor is the gutted
		// shape, and a list marker's digit is not prose.
		{"a space after the closer is the gutted shape", "One. **<!--cite:c-1-->**\n", 0, []string{"c-1"}},
		{"so is a soft wrap after the closer", "One.**\n<!--cite:c-1-->\n", 0, []string{"c-1"}},
		{"flush after a bare terminator is the gutted shape", "One.<!--cite:c-1--> Two.\n", 0, []string{"c-1"}},
		{"an emptied numbered item", "1) <!--cite:c-1-->\n", 0, []string{"c-1"}},
		{"an emptied bold bullet", "- ****<!--cite:c-1-->\n", 0, []string{"c-1"}},
		// A bullet whose first sentence was cut down to its anchor: "- X<c>. Y." gutted leaves the
		// anchor at the head of Y — bare by the same rule as a paragraph's head.
		{"a gutted bullet head", "- <!--cite:c-1--> Y stays.\n", 0, []string{"c-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Count(tc.md); got != tc.count {
				t.Errorf("Count(%q) = %d, want %d", tc.md, got, tc.count)
			}
			if got := BareAnchorIDs(tc.md); !slices.Equal(got, tc.bare) {
				t.Errorf("BareAnchorIDs(%q) = %v, want %v", tc.md, got, tc.bare)
			}
		})
	}
}
