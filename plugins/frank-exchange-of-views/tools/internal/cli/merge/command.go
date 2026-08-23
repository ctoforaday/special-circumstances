// Package merge is the red merge seat's contract: the board's only writer.
//
// mint, close, dispose and regrade live here and nowhere else. No other role has
// them, which is what makes "the board has one writer" a property of the tool
// rather than a rule someone has to follow.
package merge

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

const role = "merge"

// Verbs is this seat's surface, mounted at the ROOT of its own tree. See seat.RoleVerbs.
func Verbs() []*cobra.Command {
	return seat.RoleVerbs(role,
		seat.Register(),
		newMint(),
		newClass(),
		newClose(),
		newCarry(),
		newRegrade(),
		newSpotCheck(),
		newInquirySupport(),
		newNearMatch(),
		seat.Position("position-red"),
		seat.Closing("closing-red"),
		newVerdict(),
		seat.Friction(),
	)
}
