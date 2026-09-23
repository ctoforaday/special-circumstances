package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// THE ROSTER: the seat ids the engine can actually produce.
//
// `RequireDispatchedSeat` checks a PREFIX, so `red-lens-` admits `red-lens-banana` and
// `red-lens-L99-oops`. That was the whole guard on a seat id's legitimacy, and it is the half
// of identity the binding could not close: register is the one call that takes a seat's word for
// who it is, so what it accepts had better be an id a dispatch could have created.
//
// THIS IS THE `roleSeats` COMMENT, PROMOTED FROM PROSE TO A PATTERN. The vocabulary is —
//
//	lens   red-lens-<area>
//	chair  red-chair
//	blue   blue-lane-<N>, blue-respond, blue-synthesize, frontier
//	bench  judge
//
// ONE BENCH SEAT, BECAUSE THE SITTINGS DIFFERED ONLY IN THE QUESTION. `judge-terminal`,
// `assemble` and `judge-petition-<petitioner>` were separate ids whose verb surfaces were
// IDENTICAL to `judge`'s — same twelve verbs, same `judgment` tier, same `bench` role, and no
// refusal anywhere keyed on which of them a seat claimed to be. What actually differed was the
// question the engine asked: rule this docket, rule what still stands at the exit, hear this
// petition, assemble the report. That is an instruction for a sitting, not an identity.
//
// The petitioner suffix existed for a reason that no longer holds. Replay kept one shard per seat
// id, so a bare `judge-petition` lost every earlier sitting's rulings (#394) and the id had to
// carry who filed. Under the store there is no losing shard — see DiscardedForSeat in
// agentbinding.go, where that loss is now UNREPRESENTABLE — so the id stopped needing to carry it
// and nobody went back to collapse it.
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
// none. A lens id is bounded twice — the pattern for its shape, `LensAreas` for its area (see
// MEMBERSHIP, NOT JUST SHAPE below) — so `red-lens-evidence-oops` is refused, not admitted. What
// stays unbounded is the SITTING: a seat from any epoch passes. Bounding it against a declared
// ceiling was considered and
// REJECTED — a resume legitimately reduces that ceiling (the standing stop-and-resume practice), so
// the bound would refuse seats from the run's own earlier epochs. The shape is what can be checked
// without a second copy of the engine's dispatch logic living over here and drifting.
// THE ROSTER ITSELF LIVES IN internal/seatclass, one level down, because internal/reportvoice
// needs the same list and this package imports reportvoice. The reason is stated there, at the
// list. What stays here is what a seat may DO with an id: whether the engine could have produced
// it, which tier it rides, and which area it audits.

// TierClassOfSeat maps a dispatched seat id to the tier class the engine dispatched it on, or ""
// for a seat that rides no tier (the operator) and for an id no dispatch could have produced.
//
// "" IS TWO ANSWERS AND THAT IS DELIBERATE HERE, because both mean the same thing to the one
// caller: there is no configured tier to hold this seat to. An unrecognised id is already refused
// at the door by requireDispatchableSeat, so it cannot reach the tier check as a silent pass.
func TierClassOfSeat(seatID string) string {
	if s, ok := seatclass.Seats[seatID]; ok {
		return seatclass.ClassOf(s.Base)
	}
	if seatclass.LaneSeat.MatchString(seatID) {
		return seatclass.ClassOf("blue-lane")
	}
	return ""
}

// AreaOf is the strategic area a lens seat audits, or "" for a seat that audits none.
//
// IT WAS CALLED RoleOf AND RETURNED AN AREA. `role` is a governed noun in the terms registry —
// it is what selects a seat's SURFACE (lens, chair, blue, bench, operator) — so a function named
// for it that hands back `evidence` was the same word in two senses, which reads as confirmation
// to anyone checking. The area is a field now and the name says which fact it is.
func AreaOf(seatID string) string { return seatclass.Seats[seatID].Area }

// LensAreas are the strategic areas a lens seat can be dispatched for — one seat each, and the
// name IS the identity (#791). ALIASED from internal/flags, which declares it: internal/record
// imports that package, so a second copy here is the one thing this arrangement exists to stop.
//
// It is what GENERATES the lens rows above, so membership and shape are no longer two questions
// with two answers: an id is a lens seat exactly when this list named its area.
// TestTheLensAreasMatchWhatTheEngineDeclares holds it against debate.js's own RED_AREAS, so
// adding an area there and not here fails.
var LensAreas = flags.LensAreas

// dispatchableSeatID reports whether an id is one the engine's naming scheme can produce.
//
// A PETITION IS NO LONGER A SEAT ID. It was `judge-petition-<petitioner>`, which made who filed
// part of the bench's identity; it is now a question put to `judge` for one sitting, and who
// filed is on the petition it rules. See the bench note above the roster.
func dispatchableSeatID(seatID string) bool {
	if _, ok := seatclass.Seats[seatID]; ok {
		return true
	}
	return seatclass.LaneSeat.MatchString(seatID)
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
			"in the dispatch rather than in your call, and the log is where it goes",
		seatID)
}

// EVERY ROLE HAS A SAMPLE AND EVERY SAMPLE IS A REAL ID.
//
// roleSample is the one hand-written table left in this file, kept because which seat stands for
// a role is a choice rather than a fact (see roles.go). A hand-written table with no gate is the
// defect one level up, so this is the gate: a role on the roster with no sample builds no
// command tree, and a sample no dispatch could produce walks a surface that does not exist.
func rolesAndSamplesAgree() error {
	for _, s := range seatclass.Seats {
		if _, ok := roleSample[s.Role]; !ok {
			return fmt.Errorf("role %q is on the roster and has no sample seat, so nothing can build its command tree", s.Role)
		}
	}
	for role, id := range roleSample {
		if !dispatchableSeatID(id) {
			return fmt.Errorf("role %q samples %q, which is not an id the engine dispatches", role, id)
		}
		if r := roleOfSeatID(id); r != role {
			return fmt.Errorf("role %q samples %q, whose own role is %q", role, id, r)
		}
	}
	return nil
}

// roleOfSeatID is the role a seat id is dispatched under, or "" for an id that is not one.
func roleOfSeatID(seatID string) string {
	if s, ok := seatclass.Seats[seatID]; ok {
		return s.Role
	}
	if seatclass.LaneSeat.MatchString(seatID) {
		return "blue"
	}
	return ""
}
