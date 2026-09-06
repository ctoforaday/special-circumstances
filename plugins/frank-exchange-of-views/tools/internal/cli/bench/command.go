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
		seat.Register(),
		newOpinion(),
		newHalt(),
		newCertify(),
		newDeclare(),
		newOutcome(),
		newAssemble(),
		seat.Log(),
	)
}
