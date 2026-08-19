// Package lens is the red lens seat's contract.
//
// A lens surfaces findings, notes observations, and verifies citations. It has
// no mint verb and no close verb, and that absence IS the role boundary: the
// merge disposes what a lens surfaces, so a lens structurally cannot put
// anything on the board itself.
package lens

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

const role = "lens"

// Verbs is this seat's surface, mounted at the ROOT of its own tree. See seat.RoleVerbs.
func Verbs() []*cobra.Command {
	return seat.RoleVerbs(role,
		seat.Register("FIRST ACTION at the seat: `register`, no flags of its own — the hook injects your run, so --seat-id is the one you pass, and a value that disagrees with the run you were dispatched into is refused rather than obeyed"),
		newFinding(),
		newVerify(),
		newCorroborate(),
		newReproduce(),
		seat.Friction("a capability gap or protocol misfit, as an event that survives aborts: --reason. CLOSE THIS CHANNEL EVERY SITTING: --none --reason \"<what you reached for and found>\" says nothing blocked you, which silence cannot say — an empty friction log reads the same whether the sitting was clean or the channel went unused"),
	)
}
