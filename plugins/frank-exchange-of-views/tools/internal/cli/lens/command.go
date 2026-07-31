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

// NewCommand assembles the role. Each verb builds itself in its own file.
func NewCommand() *cobra.Command {
	return seat.Role(role,
		"red lens seats — findings, observations, citations. Cannot mint or close.",
		seat.Register("FIRST ACTION at the seat: register --run <runDir> --seat-id <SEAT_ID from your prompt>"),
		newFinding(),
		newObserve(),
		newCite(),
		seat.Friction("attributed friction (survives aborts as an event): --reason"),
	)
}
