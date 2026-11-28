package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// A RULING ON A SECTION BARS BY THE SECTION'S TOKEN (§III.B6's section arm, T-B42's section cell): a
// renamed heading carrying the token is still the ruled section, a new heading of the ruled one's old
// words is another section, and a section gap whose token the report does not hold (an archived one
// about a section missing at its mint) is read by its heading's key.
func TestASectionRulingBarsByTheSectionsToken(t *testing.T) {
	g := heldGap{id: "G-0000000a", aboutKind: "section", aboutRef: "Method"}
	renamed := "# T\n\n## Approach" + anchor.Token(g.id) + "\n\nWe read.\n\n## Method\n\nA new section.\n"
	if !sameSection(renamed, g, "Approach") {
		t.Error("the renamed heading carrying the ruled gap's token is not read as the ruled section")
	}
	if sameSection(renamed, g, "Method") {
		t.Error("a new heading of the ruled section's old words is read as the ruled section")
	}
	tokenless := "# T\n\n## Method\n\nWe read.\n"
	if !sameSection(tokenless, g, "method") || sameSection(tokenless, g, "Results") {
		t.Error("a tokenless section gap is not read by its heading's key")
	}
}
