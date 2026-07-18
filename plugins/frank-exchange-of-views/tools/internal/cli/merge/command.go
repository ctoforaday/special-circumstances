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

func NewCommand() *cobra.Command {
	return seat.Role(role,
		"the red merge seat — the board's only writer.",
		seat.Register(role, "FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID>"),
		newMint(),
		newClose(),
		newDispose(),
		newRegrade(),
		newDisputeRespond(),
		newSpotCheck(),
		seat.Position(role, "the round's ### RED section (prose via --file)"),
		seat.Closing(role, "a ### RED CLOSING entry per docketed gap: --id <gap> --file <prose>"),
		newVerdict(),
		seat.Friction(role, "attributed friction: --text|--file"),
		seat.Petition(role, `petition the bench: --class ethical|safety|integrity|constitutional --basis "..." --relief "..."`, ""),
	)
}
