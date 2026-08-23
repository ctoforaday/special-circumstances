// Package anchor is the immortal-anchor vocabulary: how an anchor is spelled in the report, what
// KIND of thing each id names, and how to read the report at one.
//
// # Why it is a leaf
//
// Three anchor classes are minted by three different verbs — `lens finding` (`<!--fx:…-->`),
// `blue cite` (`<!--cite:…-->`) and `blue prove` (`<!--proof:…-->`) — and read back by the edit
// guard, the board and every seat that wants the live text at one. That is a vocabulary several
// layers share, so it depends on NOTHING but the standard library. It was in `internal/bluedoc`,
// which reaches up into `internal/cli/lens` for its span locator; anything importing it inherited a
// command package, and `internal/cli/seat` could not import it at all without a cycle.
//
// The enforcement stayed where it belongs: `bluedoc` still owns the rule that an edit may carry an
// anchor but never drop, duplicate or invent one. This package owns only what an anchor IS.
package anchor

import "strings"

// Token rebuilds the literal token for an anchor id, so a message can quote what must be
// reproduced verbatim.
//
// THE CLASS IS CARRIED BY THE ID's PREFIX, which is the one string-encoded fact here that is
// load-bearing on purpose: the minting verbs choose the prefix, and every reader — this function,
// Label, the edit guard's sweep — recovers the class from it rather than from a field. It is
// tolerable because the id and its token are minted together and never travel apart, and because
// an unknown prefix falls to the finding class rather than to a plausible zero.
func Token(id string) string {
	switch {
	case strings.HasPrefix(id, "c-"):
		return "<!--cite:" + id + "-->"
	case strings.HasPrefix(id, "p-"):
		return "<!--proof:" + id + "-->"
	default:
		return "<!--fx:" + id + "-->"
	}
}

// Label describes an anchor id by its class, so a seat is told which KIND of anchor its edit would
// have disturbed. A generic name is passed through unchanged.
func Label(id string) string {
	switch {
	case strings.HasPrefix(id, "c-"):
		return "citation anchor " + id + " (citations are tool-managed — remove one with the tool, never a raw edit)"
	case strings.HasPrefix(id, "p-"):
		return "proof anchor " + id + " (a computation backs this sentence — the script and its output are cached; remove one with the tool, never a raw edit)"
	case strings.HasPrefix(id, "f-"):
		return "finding-marker " + id
	default:
		return id
	}
}

// SkipRun returns the index just past a maximal run of anchor tokens starting at i in s, or i
// itself when no token starts there.
//
// WHY A RUN AND NOT A TOKEN. Anchors abut: `verification<!--fx:f-e4bc25ec--><!--fx:f-73a56bd3-->`
// is a real span from the 2026-08-22 corpus, where two lenses anchored the same sentence. A
// caller stepping over "the anchor" would step over one of two and land inside the layer it
// was trying to see past.
func SkipRun(s string, i int) int {
	for i >= 0 && i < len(s) {
		n := tokenLenAt(s, i)
		if n == 0 {
			break
		}
		i += n
	}
	return i
}

// tokenLenAt is the length of the anchor token beginning at i, or 0. It recognises the same
// three classes Token spells, and nothing else: a stray HTML comment is not an anchor.
func tokenLenAt(s string, i int) int {
	if i < 0 || i >= len(s) {
		return 0
	}
	const open, close = "<!--", "-->"
	rest := s[i:]
	if !strings.HasPrefix(rest, open) {
		return 0
	}
	body := rest[len(open):]
	id, ok := "", false
	for _, c := range []struct{ prefix, idPrefix string }{{"fx:", "f-"}, {"cite:", "c-"}, {"proof:", "p-"}} {
		if strings.HasPrefix(body, c.prefix) {
			id, ok = body[len(c.prefix):], true
			if !strings.HasPrefix(id, c.idPrefix) {
				return 0
			}
			id = id[len(c.idPrefix):]
			break
		}
	}
	if !ok {
		return 0
	}
	end := strings.Index(id, close)
	// THE ID IS HEX, and this reader must agree with the ones that PROTECT an anchor —
	// claimcount's findingMarkerRe / citationMarkerRe / proofMarkerRe, which each require
	// `[0-9a-f]+`. A looser reader here would step over a token those refuse to see, so a
	// splice would tidy punctuation around something nothing else treats as an anchor.
	// Pinned by TestSkipRunAgreesWithTheProtectionSweep rather than by this comment; the
	// pattern is not unified across the two packages because anchor is a stdlib-only leaf
	// and claimcount reaches for regexp, and inverting that dependency to share one pattern
	// is a larger move than this guard is worth.
	if end <= 0 {
		return 0
	}
	for i := 0; i < end; i++ {
		if c := id[i]; !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return 0
		}
	}
	return len(rest) - len(id) + end + len(close)
}
