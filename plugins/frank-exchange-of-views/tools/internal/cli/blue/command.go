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

// Verbs is this seat's surface, mounted at the ROOT of its own tree. See seat.RoleVerbs.
func Verbs() []*cobra.Command {
	return seat.RoleVerbs(role,
		seat.Register(),
		newEdit(),
		newCite(),
		newProve(),
		newRevision(),
		newRetire(),
		newInquiry(),
		newManifestRow(),
		newClaimIndex(),
		seat.Position("position-blue"),
		seat.Closing("closing-blue"),
		seat.Friction(),
	)
}
