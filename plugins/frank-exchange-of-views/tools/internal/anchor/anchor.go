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
