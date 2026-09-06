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
		seat.Register(),
		newFinding(),
		newVerify(),
		newCorroborate(),
		newReproduce(),
		seat.Log(),
	)
}
