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
		"the bench — opinions, halt, certification. Never originates. It rules petitions through `motion petition rule`.",
		seat.Register("FIRST ACTION at the sitting: `register`, with no flags — the engine injects your run and your identity, and a --run or --seat-id that disagrees with the dispatch is refused rather than obeyed"),
		newOpinion(),
		newHalt(),
		newCertify(),
		newDeclare(),
		newOutcome(),
		newAssemble(),
		seat.Friction("a capability gap or protocol misfit, as an event that survives aborts: --reason. CLOSE THIS CHANNEL EVERY SITTING: --none --reason \"<what you reached for and found>\" says nothing blocked you, which silence cannot say — an empty friction log reads the same whether the sitting was clean or the channel went unused"),
	)
}
