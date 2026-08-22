package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"strings"
)

// Seat identity is BOUND to a role namespace.
//
// Without this the role boundary was a naming convention, not a boundary. A lens
// seat could not run `feov-record lens mint` — the lens namespace has no mint
// verb — but nothing stopped it running `feov-record merge mint --seat-id
// red-lens-r1-L1`, and it minted a board gap. Verified before this file existed;
// the tool said "minted R1-1".
//
// That mattered more than a missing guard usually does, because the verb set
// being the role boundary is the engine's PREMISE, not a convenience: blue is
// additive-only and never touches the ledger, a lens surfaces observations the
// merge disposes, and the bench rules without originating. Those are load-bearing
// claims about who can do what, and the record is the evidence they held. A
// boundary enforced only by which word a seat happens to type is evidence of
// nothing.
//
// The check is a PREFIX match on the seat id the engine assigns, taken from
// debate.js's own recordClause dispatch so the two cannot drift silently:
//
//	lens   red-lens-r<N>-L<M>
//	merge  red-merge-r<N>
//	blue   blue-lane-<N>, blue-respond-r<N>, blue-synthesize, frontier
//	bench  judge-r<N>, judge-petition-<petitioner>, judge-terminal, assemble
//
// The petition sitting carries its PETITIONER because one id must name one sitting: it used to
// be the bare `judge-petition` for all of them, and replay kept one shard per seat id, so every
// earlier sitting's rulings were dropped (#394). The prefix match is unaffected — `judge-` is
// still position 0 — and `judge-petition-red-merge-r1` now reads as round 1 rather than 0.
// OPERATOR IS A SEAT ID, not the absence of one.
//
// The operator surface used to be what you got when NOTHING identified you, which made "no
// identity" a MODE — and a mode nobody selects is one nobody can be refused from. Setup, capture,
// dashboard and the hook backends are run by a party; that party now says so, and the tree is
// chosen the same way for every caller: by the identity, never by its absence.
const OperatorRole = "operator"

var roleSeats = map[string][]string{
	OperatorRole: {OperatorRole},
	"lens":       {"red-lens-"},
	"merge":      {"red-merge-"},
	"blue":       {"blue-", "frontier"},
	"bench":      {"judge-", "assemble"},
}

// chairOfRole maps a seat's ROLE to the CHAIR whose scorecard measures it.
//
// A chair is a side of the debate; a role is a seat's verb set. They are not the same axis —
// `lens` and `merge` are two roles sitting in ONE chair, because a scorecard grades how RED is
// doing on this question, not how one of red's two seats is. Only `operator` has no chair: it is
// not a party to the debate, which is why the operator command takes an explicit --chair and a
// seat's own read takes nothing at all.
//
// ONE COPY. debate.js carried a second, keyed on tool name (`{'red-lens':'red','red-merge':'red',
// blue:'blue', bench:'bench'}`), whose only use was deciding whether to emit the scorecard clause
// at all — a question that has one answer, since every role but operator has a chair. The engine
// no longer needs to know: the seat asks the tool.
var chairOfRole = map[string]string{
	"lens":  "red",
	"merge": "red",
	"blue":  "blue",
	"bench": "bench",
}

// ChairOf reports the chair a role sits in, and whether it has one at all.
//
// The two answers are kept apart rather than collapsed to "": operator having NO chair is a fact
// about the run's structure, and a caller that cannot tell it from an unrecognised role would
// print an empty scorecard for both.
func ChairOf(role string) (string, bool) {
	c, ok := chairOfRole[role]
	return c, ok
}

// roleOfSeat reports which role a seat id belongs to, for the error message.
func roleOfSeat(seatID string) string {
	for role, prefixes := range roleSeats {
		for _, p := range prefixes {
			if strings.HasPrefix(seatID, p) {
				return role
			}
		}
	}
	return ""
}

// PartyOf answers "which party wrote this" from the EVENT, not from its seat id (#348).
//
// NOT named RoleOf: that name is taken by the LENS-INDEX reader (findinglabel.go), which
// extracts "L2" from "red-lens-r3-L2" and is the concurrency namespace this change must not
// touch — collapsing it once made 39 of 60 disposals ambiguous. Two different things were about
// to share one name, which is the collision this whole exercise exists to prevent.
//
// The field is stamped at the write. The fallback re-derives from the id only for records
// written before the field existed — a real corpus this tool still reads — and it is the
// fallback precisely because it is the thing being retired: `strings.HasPrefix(e.SeatID,
// "red-merge")` decided whether a position rendered as RED or BLUE, so an id that failed to
// match its expected prefix rendered as the wrong party with nothing to notice.
func PartyOf(e Event) string {
	if e.Role != "" {
		return e.Role
	}
	return roleOfSeat(e.SeatID)
}

// RequireDispatchedSeat asserts the seat id is one the engine created, with no claim about which
// role it belongs to.
//
// It is the half of the role guard that survives where there is no role to enforce — the motion
// commands, whose parent is their SUBJECT. Every seat may FILE any motion and that asymmetry is
// deliberate; being a seat at all is not optional, and it was not being checked.
func RequireDispatchedSeat(seatID string) error {
	if roleOfSeat(seatID) != "" {
		return nil
	}
	return feov.Errorf(feov.RoleViolation, "seat %q does not belong to any role namespace — "+
		"the engine assigns the seat id; a hand-invented one records under an identity no dispatch created. "+
		"Every seat may file a motion, but only a seat the engine created", seatID)
}

// CheckSeatRole enforces the binding at the CLI boundary.
//
// AN UNKNOWN ROLE STILL CHECKS THAT THE SEAT EXISTS, and it did not.
//
// `if !known { return nil }` was written for the commands that have no role — and it also caught
// every MOTION, because a motion's command parent is its SUBJECT. `motion grade file` arrives with
// role "grade", which is in no role table, so the whole guard returned nil.
//
// Measured 2026-08-15, on a real board:
//
//	blue position      --seat-id totally-invented   REFUSED  ("does not belong to any role namespace")
//	motion grade file  --seat-id totally-invented   ACCEPTED ("motion M2 filed (grade)")
//
// Every seat may FILE any motion — that asymmetry is deliberate and this does not touch it. What
// was lost is the other half of the same guard: that the seat id is one the engine created. A
// motion attributed to an invented identity is rendered in the report, ruled on, and joined to a
// gap, which is exactly the property the refusal's own message claims to protect ("a hand-invented
// one records under an identity no dispatch created").
//
// The shape is this repository's recurring one, in a guard: a lookup MISS returned "fine". The
// unknown-role case and the everything-is-in-order case were the same nil.
func CheckSeatRole(role, seatID string) error {
	prefixes, known := roleSeats[role]
	if !known {
		// No role to enforce, so enforce what is still true: the seat must belong to SOME
		// namespace. This is the check `motion` needs and the one it silently skipped.
		return RequireDispatchedSeat(seatID)
	}
	for _, p := range prefixes {
		if strings.HasPrefix(seatID, p) {
			return nil
		}
	}
	// The message names the seat's ACTUAL role when it has one. A seat that
	// reached for another role's verbs is usually mis-scripted rather than
	// malicious, and telling it where its own capability lives is the difference
	// between a refusal it can act on and one it will retry.
	if own := roleOfSeat(seatID); own != "" {
		return feov.Errorf(feov.RoleViolation, "seat %q belongs to the %s role and may not write through %s — "+
			"the verb set is the role boundary, and a seat that could cross it would make the record evidence of nothing "+
			"(use: feov-record %s <verb>)", seatID, own, role, own)
	}
	return feov.Errorf(feov.RoleViolation, "seat %q does not belong to any role namespace (expected one of %s) — "+
		"the engine assigns the seat id; a hand-invented one records under an identity no dispatch created",
		seatID, strings.Join(prefixes, ", "))
}

// SampleSeatOf is a seat id of the given role, for building that role's command tree in the gates
// and the surface walk. Derived from roleSeats so a new role or a renamed prefix cannot leave a
// hand-written sample pointing at a namespace that no longer exists.
func SampleSeatOf(role string) string {
	if p := roleSeats[role]; len(p) > 0 {
		return p[0] + "r1"
	}
	return ""
}
