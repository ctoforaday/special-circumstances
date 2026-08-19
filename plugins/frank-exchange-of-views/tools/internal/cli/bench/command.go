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

// Verbs is this seat's surface, mounted at the ROOT of its own tree. See seat.RoleVerbs.
func Verbs() []*cobra.Command {
	return seat.RoleVerbs(role,
		seat.Register("FIRST ACTION at the sitting: `register`, no flags of its own — just your --run and --seat-id, and a value that disagrees with the run you were dispatched into is refused rather than obeyed"),
		newOpinion(),
		newHalt(),
		newCertify(),
		newDeclare(),
		newOutcome(),
		newAssemble(),
		seat.Friction("a capability gap or protocol misfit, as an event that survives aborts: --reason. CLOSE THIS CHANNEL EVERY SITTING: --none --reason \"<what you reached for and found>\" says nothing blocked you, which silence cannot say — an empty friction log reads the same whether the sitting was clean or the channel went unused"),
	)
}
