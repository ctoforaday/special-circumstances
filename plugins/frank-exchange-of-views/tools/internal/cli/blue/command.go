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
		seat.Register("FIRST ACTION at the seat: `register`, no flags of its own — the hook injects your run, so --seat-id is the one you pass, and a value that disagrees with the run you were dispatched into is refused rather than obeyed"),
		newEdit(),
		newCite(),
		newProve(),
		newRevision(),
		newRetire(),
		newInquiry(),
		newManifestRow(),
		newClaimIndex(),
		seat.Position("the round's ### BLUE section (prose via --reason)"),
		seat.Closing("a ### BLUE CLOSING entry per docketed gap: --id <gap> --reason <prose>"),
		seat.Friction("a capability gap or protocol misfit, as an event that survives aborts: --reason. CLOSE THIS CHANNEL EVERY SITTING: --none --reason \"<what you reached for and found>\" says nothing blocked you, which silence cannot say — an empty friction log reads the same whether the sitting was clean or the channel went unused"),
	)
}
