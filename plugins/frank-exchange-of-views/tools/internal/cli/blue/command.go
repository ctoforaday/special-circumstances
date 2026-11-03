// Package blue is the blue side's contract: the lanes, synthesis, and responses. A lane is its own
// role with its own surface (LaneVerbs); the others share blue's.
//
// Blue has NO board verbs at all — no mint, no close, no dispose, no regrade.
// Additive-only and never-touch-the-ledger are topology here rather than
// obedience: blue cannot subtract from the board because it has no verb that
// could.
package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

const role = "blue"

// Verbs is this seat's surface, mounted at the ROOT of its own tree. See seat.RoleVerbs.
func Verbs() []*cobra.Command {
	return seat.RoleVerbs(role,
		seat.Register(),
		newEdit(),
		newCite(),
		newCompute(),
		newRetire(),
		newInquiry(),
		newManifestRow(),
		newEngage(),
		seat.Closing("closing-blue"),
		seat.Log(),
	)
}

// LaneVerbs is a lane's surface: what it needs to research and record its part of the report, and
// nothing that changes the report once a part is on it. Edit, cite, compute and retire are blue's,
// and a lane that types one is refused as the wrong seat — the report's first change closes the
// door on every lane still writing (record.PartStanding), so a lane acting on its own part would
// cost the others theirs. `ingest` is here and nowhere else: a part is a lane's.
func LaneVerbs() []*cobra.Command {
	return seat.RoleVerbs(seatclass.Lane.Role,
		seat.Register(),
		newIngest(),
		newInquiry(),
		seat.Log(),
	)
}
