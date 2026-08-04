package blue

import "strings"

// SPLICE HYGIENE — the punctuation a tool should own, not a debate round.
//
// Measured on the 2026-08-04 smoke: of blue-respond-r2's 17 edits, SIX were pure punctuation
// repair ("remove double colon", "remove double period", "remove double semicolon", "remove
// double period and extra period"). None answered a red finding. They were repairing SPLICE
// DAMAGE from blue's own earlier edits — a --new ending in "." landing against text that already
// began with one. A third of a paid round spent on artifacts a deterministic pass can remove,
// the same argument that ended indentation debates in code review once gofmt existed.
//
// The normalization is deliberately NARROW, applied ONLY at the two seams an edit creates and
// only to doubles a splice can manufacture. Prose is not code: a rule that reflows sentences
// would change meaning, so this touches nothing but adjacent duplicate terminal marks and a
// space that a splice orphaned in front of one.
//
// NOT collapsed, on purpose: "..." (an ellipsis is content), "!!"/"??" (emphasis a human may
// intend), anything inside an anchor token — the seams are checked positionally, so a
// "<!--fx:…-->" is never in range.

// dupPunct are the marks a splice can duplicate where the duplicate is always an artifact.
const dupPunct = ".,:;"

// tidySeam removes splice artifacts around the byte offset `at` in s — the join between spliced
// and pre-existing text. Returns the repaired string and whether anything changed.
func tidySeam(s string, at int) (string, bool) {
	if at <= 0 || at >= len(s) {
		return s, false
	}
	b := []byte(s)
	changed := false

	// " ." → "." : a splice left a space stranded before terminal punctuation.
	if i := at; i > 0 && i < len(b) && b[i-1] == ' ' && strings.IndexByte(dupPunct, b[i]) >= 0 {
		b = append(b[:i-1], b[i:]...)
		changed = true
	}
	// ".." → "." : only an EXACT pair. A third of the same mark means an ellipsis or deliberate
	// emphasis, which is content and is left alone.
	for i := 1; i < len(b); i++ {
		if i != at && i != at-1 && i != at+1 { // only at the seam
			continue
		}
		c := b[i]
		if strings.IndexByte(dupPunct, c) < 0 || b[i-1] != c {
			continue
		}
		if i+1 < len(b) && b[i+1] == c { // three or more — content, not an artifact
			continue
		}
		if i-2 >= 0 && b[i-2] == c {
			continue
		}
		b = append(b[:i], b[i+1:]...)
		changed = true
		i--
	}
	return string(b), changed
}
