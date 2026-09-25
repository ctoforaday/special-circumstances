package sittingwrite

import (
	"fmt"
	"io"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE WORK LIST ARRIVES WITH THE SEAT (#1122).
//
// A seat could not learn that it owed nothing without asking, so the cheapest empty sitting was one
// call and never zero: measured on the 2026-09-22 smoke, 0 of 13 empty sittings took the affordance
// #1089 gave them, and on universe-m9 the floor across 37 sittings was two calls, five no-work lens
// sittings sitting exactly on it.
//
// The record knows the answer at dispatch. SubagentStart fires once and delivers to the SEAT's
// context (plans/hook-surface-spike.md §10, re-measured 2026-09-25), the configuration names the seat
// wherever it is one-to-one, and the sitting_open this process has just written is what satisfies the
// dispatch — so by the time this runs the list is both computable and correct.
//
// IT IS THE PROJECTION'S OWN BYTES, NOT A SECOND AUTHOR OF THE SEAT'S DUTIES. Everything the seat is
// told about what it owes, what it may do, and whether ending its turn now is the correct act is
// written by record.SittingOf and reaches every other reader identically. What the preamble adds is
// the ONE fact only this channel can state — that the seat is already holding the list — because a
// prompt cannot say it (blue's configuration seats three seats and gets no injection) and the help
// page cannot say it (a resumed dispatch fires no SubagentStart, #1111, and that seat must still ask).
const (
	// arrived is the preamble on a list that was read. It names no command: the surface owns the
	// verbs, and a seat that wants the list again finds it where it already found everything else.
	arrived = "YOUR WORK LIST, read from this run's record at the moment you were dispatched. It is the " +
		"same projection your own surface returns, delivered instead of fetched — you are holding it, " +
		"and you do not need to ask for it before you act. Everything open to you is in `sitting.open`, " +
		"each item with whether it blocks you finishing; `sitting.complete` is the answer to whether " +
		"anything does. Act on it, and read it again from your surface after you have recorded anything, " +
		"because your own acts move it."

	// unreadable is the preamble on a list that could NOT be read, and it exists so that this case
	// and an empty list are not the same bytes. A seat handed "nothing is open to you" when the truth
	// is "nothing was determined" ends its turn on a measurement that never happened.
	unreadable = "YOUR WORK LIST COULD NOT BE READ, and that is NOT the same as nothing being open to " +
		"you: nothing was determined. Read it from your own surface before you act, and treat what " +
		"follows as the reason this delivery failed rather than as your work."
)

// seatRoleOf is the surface role a seat id is dispatched under. A variable so the one branch nothing
// can reach through the real tables — a seat the attestation table names and the roster does not —
// is still driveable, because that branch is the one whose silence would be indistinguishable from a
// clean board.
var seatRoleOf = func(seat string) string { return seatclass.Seats[seat].Role }

// injectWorkList renders the seat's work list for the SubagentStart hook to hand to the seat, or the
// loud statement that it could not be rendered.
//
// IT RETURNS NOTHING RATHER THAN A PLACEHOLDER WHERE THE SEAT IS UNKNOWABLE. blue-researcher seats
// three seats — the frontier, the lanes, blue-respond — so nothing but that seat's own register can
// say which one sat; an injection there would have to guess. Silence is the honest answer and costs
// blue exactly the call it already paid.
//
// A WRITE FAULT IS NOT THIS FUNCTION'S TO REPORT. The caller has already appended the span and
// returns its own error; this one either produces text for the seat or produces nothing.
func injectWorkList(run record.Run, agentType string, out io.Writer) {
	seat, known := record.SeatOfAgentType(agentType)
	if !known {
		return
	}
	role := seatRoleOf(seat)
	if role == "" {
		// The attestation table and the roster disagree about a seat id. Loud, because the seat is
		// about to work with no list at all and the cause is a code defect rather than a run's state.
		fmt.Fprintf(out, "%s\n\n%s is not a seat the roster gives a role, so no work list could be projected for it.\n",
			unreadable, seat)
		return
	}
	b, err := record.WorkJSONBytes(run, role, seat)
	if err != nil {
		fmt.Fprintf(out, "%s\n\n%v\n", unreadable, err)
		return
	}
	fmt.Fprintf(out, "%s\n\n%s", arrived, b)
}
