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

// EDIT BOUNDARIES SIT AT WHITESPACE — the precondition that makes splice damage impossible
// rather than merely repairable.
//
// tidySeam above REPAIRS a doubled mark after the fact. This prevents it: if the --old span both
// begins and ends at a boundary, a --new ending in "." cannot abut another "." because the next
// character is a space. It also forces blue to OWN the punctuation it is editing — quoting "the
// value is stable." rather than "the value is stable" — which makes the diff-stack legible
// (the smoke produced three ONE-BYTE edits, one of them deleting "ly", which read as noise).
//
// An annotation edge counts as a boundary. Anchors are spliced flush against the last content
// character ("Eratosthenes<!--fx:f-…-->"), so demanding literal whitespace after a span would
// reject every edit of an anchored word — the common case, since red anchors exactly the
// sentences blue is asked to repair.

func isSpaceByte(b byte) bool { return b == ' ' || b == '\t' || b == '\n' || b == '\r' }

// annotationEndsAt reports whether an invisible annotation token ends exactly at offset i.
func annotationEndsAt(s string, i int) bool {
	return (i >= 3 && s[i-3:i] == "-->") || (i >= 1 && s[i-1] == ']')
}

// annotationStartsAt reports whether an invisible annotation token begins exactly at offset i.
func annotationStartsAt(s string, i int) bool {
	return strings.HasPrefix(s[i:], "<!--") || strings.HasPrefix(s[i:], "[^")
}

// spanBoundaryOK reports whether [start,end) splits a WORD. It rejects only that: a boundary
// with an alphanumeric character on both sides.
//
// It deliberately does NOT require whitespace. LocateSpan normalises a quote by TRIMMING trailing
// punctuation (#248, so a quote may include or omit a terminal period and still match), which
// means a matched span ALWAYS ends before the period — a whitespace requirement would reject every
// sentence-final edit, including a correctly-quoted one. That same trim is the doubling mechanism
// tidySeam exists for: blue quotes "rising fast.", the matcher matches "rising fast", the verbatim
// --new "climbing fast." lands against the surviving period and yields "climbing fast..". The seam
// tidy handles that; this predicate handles the other half — a span starting or ending inside a
// word, which produces edits like the smoke's one-byte "delete ly" that are unreadable in the
// diff-stack and mean blue is editing letters rather than language.
func spanBoundaryOK(s string, start, end int) bool {
	word := func(b byte) bool {
		return b == '_' || (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}
	splits := func(i int) bool { return i > 0 && i < len(s) && word(s[i-1]) && word(s[i]) }
	return !splits(start) && !splits(end)
}
