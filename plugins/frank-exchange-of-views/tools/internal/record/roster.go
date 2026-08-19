package record

import (
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// THE ROSTER: the seat ids the engine can actually produce.
//
// `RequireDispatchedSeat` checks a PREFIX, so `red-lens-` admits `red-lens-banana` and
// `red-lens-r99-L99-oops`. That was the whole guard on a seat id's legitimacy, and it is the half
// of identity the binding could not close: register is the one call that takes a seat's word for
// who it is, so what it accepts had better be an id a dispatch could have created.
//
// THIS IS THE `roleSeats` COMMENT, PROMOTED FROM PROSE TO A PATTERN. That comment has always
// carried the real vocabulary —
//
//	lens   red-lens-r<N>-L<M>
//	merge  red-merge-r<N>
//	blue   blue-lane-<N>, blue-respond-r<N>, blue-synthesize, frontier
//	bench  judge-r<N>, judge-petition-<petitioner>, judge-terminal, assemble
//
// — and nothing could refuse anything against it, because a comment is not a check. Every id below
// is verified against debate.js's own dispatch sites; see TestTheRosterMatchesWhatTheEngineActuallyDispatches.
//
// WHAT THIS DOES NOT DO, stated because a gate that seems to prove more than it does is worse than
// none. It bounds the SHAPE, never the membership: `red-lens-r99-L4` is well formed and no run will
// ever dispatch it. Bounding the round against the run's declared maxRounds was considered and
// REJECTED — a resume legitimately reduces that ceiling (the standing stop-and-resume practice), so
// the bound would refuse seats from the run's own earlier rounds. The shape is what can be checked
// without a second copy of the engine's dispatch logic living over here and drifting.
type seatShape struct {
	role   string
	re     *regexp.Regexp
	sample string // drives the drift test; never used at runtime
}

var seatShapes = []seatShape{
	{"lens", regexp.MustCompile(`^red-lens-r\d+-L\d+$`), "red-lens-r1-L1"},
	{"merge", regexp.MustCompile(`^red-merge-r\d+$`), "red-merge-r1"},
	{"blue", regexp.MustCompile(`^blue-lane-\d+$`), "blue-lane-1"},
	{"blue", regexp.MustCompile(`^blue-respond-r\d+$`), "blue-respond-r1"},
	{"blue", regexp.MustCompile(`^blue-synthesize$`), "blue-synthesize"},
	{"blue", regexp.MustCompile(`^frontier$`), "frontier"},
	{"bench", regexp.MustCompile(`^judge-r\d+$`), "judge-r1"},
	{"bench", regexp.MustCompile(`^judge-terminal$`), "judge-terminal"},
	{"bench", regexp.MustCompile(`^assemble$`), "assemble"},
	{OperatorRole, regexp.MustCompile(`^` + OperatorRole + `$`), OperatorRole},
}

// petitionPrefix is handled apart from the table because its tail is ITSELF a seat id — the
// sitting is named for the seat that petitioned, so one id names one sitting (#394). Folding it
// into the table would need a pattern matching every other pattern, which is the point at which a
// regex stops being a schema and becomes a guess.
const petitionPrefix = "judge-petition-"

// DispatchableSeatID reports whether an id is one the engine's naming scheme can produce.
func DispatchableSeatID(seatID string) bool {
	if petitioner, ok := strings.CutPrefix(seatID, petitionPrefix); ok {
		// A petition sitting is named for a real seat, and never for another petition sitting:
		// there is no sitting about a sitting.
		return !strings.HasPrefix(petitioner, petitionPrefix) && DispatchableSeatID(petitioner)
	}
	for _, s := range seatShapes {
		if s.re.MatchString(seatID) {
			return true
		}
	}
	return false
}

// RequireDispatchableSeat refuses an id no dispatch could have produced.
//
// It fires at `register` and nowhere else, deliberately. Register is where a seat asserts who it
// is — every later call reads the binding that assertion created — so this is the one door worth
// standing at, and putting it on every verb would only re-check a value the record now supplies.
func RequireDispatchableSeat(seatID string) error {
	if DispatchableSeatID(seatID) {
		return nil
	}
	return feov.Errorf(feov.RoleViolation,
		"seat %q is not an id the engine dispatches. Registering binds this id to you for the whole run, so it "+
			"must be one a dispatch could have created — not a near miss and not one you composed. Your id is "+
			"stated in your prompt as SEAT_ID; copy it exactly. If it IS what your prompt says, that is a defect "+
			"in the dispatch rather than in your call, and the friction verb is the channel for it",
		seatID)
}
