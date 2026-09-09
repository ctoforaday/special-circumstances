// Package merge is the red chair's contract: the seat that RUNS the debate.
//
// The chair reads the board through `dispatch next` and the record says who sits; it issues the
// verdict when the board permits one, files closing arguments on docketed gaps, spot-checks the
// archive and carries closures the archive already holds. It mints nothing and closes nothing:
// a gap is its lens's from mint to close (plans/roundless.md §III.B.3), and the chair is party to
// no exchange — which is what makes per-dispute dispatch save chair sittings rather than add them.
package merge

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

const role = "merge"

// Verbs is this seat's surface, mounted at the ROOT of its own tree. See seat.RoleVerbs.
func Verbs() []*cobra.Command {
	return seat.RoleVerbs(role,
		seat.Register(),
		newCarry(),
		newSpotCheck(),
		newInquirySupport(),
		seat.Position("position-red"),
		seat.Closing("closing-red"),
		newVerdict(),
		newDispatch(),
		seat.Log(),
	)
}
