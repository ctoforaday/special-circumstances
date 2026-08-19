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
		seat.Register("FIRST ACTION at the seat: `register`, no flags of its own — just your --run and --seat-id, and a value that disagrees with the run you were dispatched into is refused rather than obeyed"),
		newMint(),
		newClass(),
		newClose(),
		newCarry(),
		newRegrade(),
		newSpotCheck(),
		newInquirySupport(),
		newNearMatch(),
		seat.Position("the round's ### RED section (prose via --reason)"),
		seat.Closing("a ### RED CLOSING entry per docketed gap: --id <gap> --reason <prose>"),
		newVerdict(),
		seat.Friction("a capability gap or protocol misfit, as an event that survives aborts: --reason. CLOSE THIS CHANNEL EVERY SITTING: --none --reason \"<what you reached for and found>\" says nothing blocked you, which silence cannot say — an empty friction log reads the same whether the sitting was clean or the channel went unused"),
	)
}
