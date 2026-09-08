// Package lens is the red lens seat's contract.
//
// A lens audits the report against its area and puts what it finds on the board itself: it
// records findings, verifies citations, screens the board with near-match, and MINTS its own gaps
// against its own budget. A gap belongs to the lens that minted it for its whole life — regrade
// and close are the originator's acts, refused at the record for any other seat
// (plans/roundless.md §III.B.3). The chair dispatches the lens when its gap needs acting on; the
// bench disposes of a docketed gap through its ruling.
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
		newMint(),
		newClass(),
		newClose(),
		newRegrade(),
		newNearMatch(),
		newFinding(),
		newVerify(),
		newCorroborate(),
		newReproduce(),
		seat.Log(),
	)
}
