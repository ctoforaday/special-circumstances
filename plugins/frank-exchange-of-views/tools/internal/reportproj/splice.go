package reportproj

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

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
// intend), anything INSIDE an anchor token — a "<!--fx:…-->" is never itself in range.
//
// BUT AN ANCHOR BETWEEN TWO MARKS IS NOT CONTENT BETWEEN THEM, and reading it as content is
// how this pass came to be a no-op in the only shape the corpus has. `normalizeQuote` trims a
// quote's trailing punctuation, so the located span stops before the sentence's period and
// that period survives at report[end:]. In 40 of the 43 anchors across the four archived runs
// the anchor sits immediately after the last WORD — before that period — so what a replacement
// ending in "." abuts is "<!--fx:…-->." and the two marks are never adjacent. Measured through
// the real command: "The cost is falling.<!--fx:f-abc123-->. Volume grows steadily." — a
// doubled period AND red's marker displaced out of the sentence it annotates. The comment at
// the foot of this file already names why that is the common case rather than a corner:
// red anchors exactly the sentences blue is asked to repair.
//
// So the seam is checked positionally STILL, and the position is found by stepping over the
// anchor layer rather than stopping at it.

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
	// "M<!--fx:…-->M" → "<!--fx:…-->M" : the same doubled mark as above, with the anchor layer
	// between the two. The mark from --new goes and the report's stays, which is what puts the
	// anchor back INSIDE its sentence instead of stranding it after the new sentence's period.
	if j := anchor.SkipRun(string(b), at); j > at && j < len(b) && at > 0 {
		if c := b[at-1]; strings.IndexByte(dupPunct, c) >= 0 && b[j] == c {
			b = append(b[:at-1], b[at:]...)
			changed = true
		}
	}
	// A TERMINAL MARK LEFT BY A DELETION, with nothing in front of it. The quote's trailing
	// punctuation is trimmed before the span is located, so it is never part of what a deletion
	// removes: replacing a whole sentence with "" leaves its period behind, at a line start or
	// against the preceding space. Measured with NO anchor present, so this half is independent
	// of the rule above — it is the existing " ." rule, which encodes the same principle and
	// only ever handled a literal space.
	if k := anchor.SkipRun(string(b), at); k < len(b) && strings.IndexByte(dupPunct, b[k]) >= 0 {
		if at == 0 || b[at-1] == '\n' || b[at-1] == ' ' || b[at-1] == '\t' {
			cut := k + 1
			// And the space the mark was holding open — but ONLY when no anchor was stepped
			// over. An anchor is spliced flush against content, so if one now sits where the
			// deleted sentence began, that space is the only thing keeping it off the next
			// word: eating it yields "<!--fx:f-abc123-->Volume".
			if k == at && cut < len(b) && b[cut] == ' ' {
				cut++
			}
			b = append(b[:k], b[cut:]...)
			changed = true
		}
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
