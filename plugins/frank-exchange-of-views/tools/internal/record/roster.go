package record

import (
	"regexp"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"

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
//	lens   red-lens-r<N>-<area>
//	merge  red-merge-r<N>
//	blue   blue-lane-<N>, blue-respond-r<N>, blue-synthesize, frontier
//	bench  judge-r<N>, judge-petition-<petitioner>, judge-terminal, assemble
//
// — and nothing could refuse anything against it, because a comment is not a check. Every id below
// is verified against debate.js's own dispatch sites by TestTheRosterMatchesWhatTheEngineActuallyDispatches,
// in both directions: a dispatched seat no shape admits, and a shape no dispatch produces.
//
// THAT SENTENCE WAS ITSELF AN UNCHECKED COMMENT until #675 — it named a test that did not exist,
// in the paragraph arguing that a comment is not a check. What stood in its place reconciled these
// shapes against `roleSeats`: two hand-written tables over here stating one vocabulary, neither of
// them the engine. The bind reads the seat ids debate.js hands `recordClause` — the string a seat
// is told to type at `register` — rather than the dispatch LABELS, which are a dashboard's text and
// do not carry a seat id's shape at all.
//
// WHAT THIS DOES NOT DO, stated because a gate that seems to prove more than it does is worse than
// none. It bounds the SHAPE, never the membership: `red-lens-r99-L4` is well formed and no run will
// ever dispatch it. Bounding the round against the run's declared maxRounds was considered and
// REJECTED — a resume legitimately reduces that ceiling (the standing stop-and-resume practice), so
// the bound would refuse seats from the run's own earlier rounds. The shape is what can be checked
// without a second copy of the engine's dispatch logic living over here and drifting.
type seatShape struct {
	role string
	re   *regexp.Regexp
	// base is the seat name internal/seatclass keys its tier class on. The shape table already
	// owns the ID grammar, so this is where an id becomes a class — the alternative was a second
	// regex ladder over seat ids living in seatclass, which is the same fact with two authors.
	//
	// EMPTY FOR THE OPERATOR, which is not a debating seat and rides no tier.
	base   string
	sample string // drives the drift test; never used at runtime
}

var seatShapes = []seatShape{
	{"lens", regexp.MustCompile(`^red-lens-r\d+-[a-z]+(?:-[a-z]+)*$`), "red-lens", "red-lens-r1-evidence"},
	{"merge", regexp.MustCompile(`^red-merge-r\d+$`), "red-merge", "red-merge-r1"},
	{"blue", regexp.MustCompile(`^blue-lane-\d+$`), "blue-lane", "blue-lane-1"},
	{"blue", regexp.MustCompile(`^blue-respond-r\d+$`), "blue-respond", "blue-respond-r1"},
	{"blue", regexp.MustCompile(`^blue-synthesize$`), "blue-synthesize", "blue-synthesize"},
	{"blue", regexp.MustCompile(`^frontier$`), "frontier", "frontier"},
	{"bench", regexp.MustCompile(`^judge-r\d+$`), "judge", "judge-r1"},
	{"bench", regexp.MustCompile(`^judge-terminal$`), "judge-terminal", "judge-terminal"},
	{"bench", regexp.MustCompile(`^assemble$`), "assemble", "assemble"},
	{OperatorRole, regexp.MustCompile(`^` + OperatorRole + `$`), "", OperatorRole},
}

// TierClassOfSeat maps a dispatched seat id to the tier class the engine dispatched it on, or ""
// for a seat that rides no tier (the operator) and for an id no dispatch could have produced.
//
// "" IS TWO ANSWERS AND THAT IS DELIBERATE HERE, because both mean the same thing to the one
// caller: there is no configured tier to hold this seat to. An unrecognised id is already refused
// at the door by requireDispatchableSeat, so it cannot reach the tier check as a silent pass.
func TierClassOfSeat(seatID string) string {
	if petitioner, ok := strings.CutPrefix(seatID, petitionPrefix); ok && dispatchableSeatID(petitioner) {
		return seatclass.ClassOf("judge-petition")
	}
	for _, s := range seatShapes {
		if s.re.MatchString(seatID) {
			return seatclass.ClassOf(s.base)
		}
	}
	return ""
}

// petitionPrefix is handled apart from the table because its tail is ITSELF a seat id — the
// sitting is named for the seat that petitioned, so one id names one sitting (#394). Folding it
// into the table would need a pattern matching every other pattern, which is the point at which a
// regex stops being a schema and becomes a guess.
//
// Bound to the head debate.js actually composes by TestThePetitionPrefixMatchesTheOneDebateComposes:
// Go CUTS this prefix where the engine COMPOSES it, so a rename over there refuses every petition
// seat over here, in a message about an id the engine had just dispatched.
// LensAreas are the strategic areas a lens seat can be dispatched for — one seat each, and the
// name IS the identity: a finding filed by the adversary lens is labelled `adversary-F1`, and
// `found_by` reads back as what found it rather than as a number needing a lookup table.
//
// MEMBERSHIP, NOT JUST SHAPE. The pattern above bounds a lens id to a hyphenated word, which
// would admit `red-lens-r1-evidence-oops`. This list is what makes the id refusable: an area the
// engine does not dispatch is not an area. TestTheRosterMatchesWhatTheEngineActuallyDispatches
// holds it against debate.js's own RED_AREAS, so adding an area there and not here fails.
var LensAreas = []string{"evidence", "logic", "dark-side", "voice", "computation", "adversary", "architecture"}

// lensAreaRe lifts the area off a lens seat id. It is a SECOND read of the shape the seat pattern
// already matched, and deliberately so: the pattern answers "is this a lens id", this answers
// "which area", and collapsing them would make the shape table carry seven alternatives that
// skeletonOfPattern could not compare against debate.js.
var lensAreaRe = regexp.MustCompile(`^red-lens-r\d+-(.+)$`)

func isLensArea(s string) bool {
	for _, a := range LensAreas {
		if a == s {
			return true
		}
	}
	return false
}

const petitionPrefix = "judge-petition-"

// dispatchableSeatID reports whether an id is one the engine's naming scheme can produce.
func dispatchableSeatID(seatID string) bool {
	if petitioner, ok := strings.CutPrefix(seatID, petitionPrefix); ok {
		// A petition sitting is named for a real seat, and never for another petition sitting:
		// there is no sitting about a sitting.
		return !strings.HasPrefix(petitioner, petitionPrefix) && dispatchableSeatID(petitioner)
	}
	for _, s := range seatShapes {
		if !s.re.MatchString(seatID) {
			continue
		}
		// A lens id's tail is an AREA, and the shape alone cannot say whether it is one.
		if s.role == "lens" {
			m := lensAreaRe.FindStringSubmatch(seatID)
			return m != nil && isLensArea(m[1])
		}
		return true
	}
	return false
}

// requireDispatchableSeat refuses an id no dispatch could have produced.
//
// It fires at `register` and nowhere else, deliberately. Register is where a seat asserts who it
// is — every later call reads the binding that assertion created — so this is the one door worth
// standing at, and putting it on every verb would only re-check a value the record now supplies.
func requireDispatchableSeat(seatID string) error {
	if dispatchableSeatID(seatID) {
		return nil
	}
	return feov.Errorf(feov.RoleViolation,
		"seat %q is not an id the engine dispatches. Registering binds this id to you for the whole run, so it "+
			"must be one a dispatch could have created — not a near miss and not one you composed. Your id is "+
			"stated in your prompt as SEAT_ID; copy it exactly. If it IS what your prompt says, that is a defect "+
			"in the dispatch rather than in your call, and the friction verb is the channel for it",
		seatID)
}
