// Package bench is the bench's contract: it rules, and never originates.
//
// There is no mint verb here. The bench disposes what the parties bring and
// publishes opinions; a bench that could originate findings would be a third
// adversary wearing a judge's robes.
package bench

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

const role = "bench"

func NewCommand() *cobra.Command {
	return seat.Role(role,
		"the bench — opinions, petition rulings, halt, certification. Never originates.",
		seat.Register("FIRST ACTION at the sitting: register --run <runDir> --seat-id <SEAT_ID from your prompt>"),
		newOpinion(),
		newPetitionRule(),
		newHalt(),
		newCertify(),
		newOutcome(),
		newAssemble(),
		seat.Friction("attributed friction (the bench seats had no event path before — the abort-surviving copy): --text|--file"),
	)
}
