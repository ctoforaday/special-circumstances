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
//	bench  judge-r<N>, judge-petition, judge-terminal, assemble
var roleSeats = map[string][]string{
	"lens":  {"red-lens-"},
	"merge": {"red-merge-"},
	"blue":  {"blue-", "frontier"},
	"bench": {"judge-", "assemble"},
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

// CheckSeatRole enforces the binding at the CLI boundary.
func CheckSeatRole(role, seatID string) error {
	prefixes, known := roleSeats[role]
	if !known {
		return nil
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
