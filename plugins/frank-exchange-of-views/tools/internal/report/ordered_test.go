package report

import "testing"

// ORDERED MOVES WHOLE SECTIONS TO THEIR ROWS' PLACES AND NOTHING ELSE: the title line first, the
// sections of one rank in the order written, text in order returned as it is.
func TestOrderedMovesWholeSectionsToTheirPlaces(t *testing.T) {
	for name, tc := range map[string][2]string{
		"in order":         {title + tl + cat + own1 + own2 + open, title + tl + cat + own1 + own2 + open},
		"no heading":       {"One paragraph.\n\n\n\nAnother.\n", "One paragraph.\n\n\n\nAnother.\n"},
		"no text":          {"", ""},
		"universe-m20":     {title + tl + open + own1 + own2 + cat, title + tl + cat + own1 + own2 + open},
		"universe-m21":     {title + tl + own1 + own2 + cat + open, title + tl + cat + own1 + own2 + open},
		"b4":               {title + tl + cat + own1 + open + own2, title + tl + cat + own1 + own2 + open},
		"text before":      {title + "**Question:** is 91 prime?\n\n" + tl + cat + own1, title + tl + cat + "**Question:** is 91 prime?\n\n" + own1},
		"a late title":     {tl + own1 + "# Is 91 prime?\nafter the title.\n\n" + open, "# Is 91 prime?\n\n" + tl + "## Is 91 divisible by a small prime?\n\nYes: 7 divides it.\n\nafter the title.\n\n" + open},
		"no final newline": {title + own1 + "## TL;DR\n\nNo.", title + "## TL;DR\n\nNo.\n\n## Is 91 divisible by a small prime?\n\nYes: 7 divides it."},
		// A second section under a row's name has no place of its own: it follows the first.
		"a second TL;DR": {title + tl + own1 + "## tl;dr\n\nAgain.\n\n" + open, title + tl + "## tl;dr\n\nAgain.\n\n" + own1 + open},
		"a heading in a fence stays in its section": {title + own1 + "```\n## TL;DR\n```\n\n" + tl, title + tl + own1 + "```\n## TL;DR\n```\n\n"},
	} {
		got := Ordered(tc[0])
		if got != tc[1] {
			t.Errorf("%s: Ordered gave\n%q\nwant\n%q", name, got, tc[1])
		}
		if again := Ordered(got); again != got {
			t.Errorf("%s: Ordered is not settled after one pass:\n%q\nthen\n%q", name, got, again)
		}
	}
}
