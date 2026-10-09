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
		{"attached", "The sky is blue<!--cite:C-00000001-->.\n", 1, nil},
		{"alone on its line", "<!--cite:C-00000001-->\n", 0, []string{"C-00000001"}},
		{"an emptied bullet", "- <!--fx:F-00000001-->?\n", 0, []string{"F-00000001"}},
		{"at the head of the next sentence", "<!--cite:C-00000001--> The grass is green.\n", 0, []string{"C-00000001"}},
		{"mid-line, its own segment", "One. <!--cite:C-00000001-->. Two.\n", 0, []string{"C-00000001"}},
		{"two anchors after prose are both attached", "Prose<!--cite:C-00000001--><!--cite:C-00000002-->.\n", 2, nil},
		{"the next sentence's own cite still counts", "<!--cite:C-00000001--> Grass is green<!--cite:C-00000002-->.\n", 1, []string{"C-00000001"}},
		// MARKDOWN AROUND THE TERMINATOR. `blue cite` on "**Water is wet.**" places the anchor
		// after the closing "**" (a "*" is not trailing punctuation); the closer and an anchor flush
		// after it belong to the sentence the terminator ends, so it is attached, not bare.
		{"bold closing a sentence", "**Water is wet.**<!--cite:C-00000001-->\n", 1, nil},
		{"italic closing a sentence", "_Water is wet._<!--cite:C-00000001-->\n", 1, nil},
		{"a parenthetical closing a sentence", "Seen (p. 3.)<!--cite:C-00000001-->\n", 1, nil},
		{"bold in a bullet", "- **Water is wet.**<!--cite:C-00000001-->\n", 1, nil},
		{"a second anchor after bold is attached too", "**Wet.**<!--cite:C-00000001--><!--cite:C-00000002-->\n", 2, nil},
		{"a plain bullet", "- Water is wet<!--cite:C-00000001-->.\n", 1, nil},
		// ...but whitespace, or nothing at all, between a terminator and the anchor is the gutted
		// shape, and a list marker's digit is not prose.
		{"a space after the closer is the gutted shape", "One. **<!--cite:C-00000001-->**\n", 0, []string{"C-00000001"}},
		{"so is a soft wrap after the closer", "One.**\n<!--cite:C-00000001-->\n", 0, []string{"C-00000001"}},
		{"flush after a bare terminator is the gutted shape", "One.<!--cite:C-00000001--> Two.\n", 0, []string{"C-00000001"}},
		{"an emptied numbered item", "1) <!--cite:C-00000001-->\n", 0, []string{"C-00000001"}},
		{"an emptied bold bullet", "- ****<!--cite:C-00000001-->\n", 0, []string{"C-00000001"}},
		// A bullet whose first sentence was cut down to its anchor: "- X<c>. Y." gutted leaves the
		// anchor at the head of Y — bare by the same rule as a paragraph's head.
		{"a gutted bullet head", "- <!--cite:C-00000001--> Y stays.\n", 0, []string{"C-00000001"}},
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
