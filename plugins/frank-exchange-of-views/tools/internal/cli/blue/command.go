// Package blue is the blue seats' contract: synthesis, lanes, and responses.
//
// Blue has NO board verbs at all — no mint, no close, no dispose, no regrade.
// Additive-only and never-touch-the-ledger are topology here rather than
// obedience: blue cannot subtract from the board because it has no verb that
// could.
package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

const role = "blue"

func NewCommand() *cobra.Command {
	return seat.Role(role,
		"blue seats — revisions, manifest rows, disputes. No board verbs at all.",
		seat.Register("FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID from your prompt>"),
		newRevision(),
		newRetire(),
		newAvenue(),
		newManifestRow(),
		newDispute(),
		newConfidence(),
		seat.Position("the round's ### BLUE section (prose via --reason)"),
		seat.Closing("a ### BLUE CLOSING entry per docketed gap: --id <gap> --reason <prose>"),
		seat.Friction("attributed friction (survives aborts as an event): --reason"),
		seat.Petition(
			`petition the bench (never sanctioned; does not pause your duties): --petition-class ethical|safety|integrity|constitutional --reason "..." --relief "..."`,
			""),
	)
}
