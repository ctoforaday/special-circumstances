package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"

	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// Seat identity is BOUND to a role namespace.
//
// Without this the role boundary was a naming convention, not a boundary. A lens
// seat could not run `feov-record lens mint` — the lens namespace has no mint
// verb — but nothing stopped it running `feov-record merge mint --seat-id
// red-lens-evidence`, and it minted a board gap. Verified before this file existed;
// the tool said "minted G1".
//
// That mattered more than a missing guard usually does, because the verb set
// being the role boundary is the engine's PREMISE, not a convenience: blue is
// additive-only and never touches the ledger, a lens surfaces observations the
// chair disposes, and the bench rules without originating. Those are load-bearing
// claims about who can do what, and the record is the evidence they held. A
// boundary enforced only by which word a seat happens to type is evidence of
// nothing.
//
// The check is a PREFIX match on the seat id the engine assigns, taken from
// debate.js's own recordClause dispatch so the two cannot drift silently:
//
// The grammar lives ONCE, in roster.go's seatShapes: that table is what admits a seat id, and its
// drift test is what holds it against debate.js's own dispatch sites. A second copy here would be
// a hand-written restatement of the same vocabulary with no check under it, and a stale copy reads
// exactly like a current one.
//
// A petition sitting is `judge`, like the other three. Which of the four questions a sitting
// answers rides on the register's `occasion` (see occasion.go) rather than on the id, and who filed
// the petition rides on the petition. The prefix match is unaffected — `judge` is still position 0.
// OPERATOR IS A SEAT ID, not the absence of one.
//
// The operator surface used to be what you got when NOTHING identified you, which made "no
// identity" a MODE — and a mode nobody selects is one nobody can be refused from. Setup, capture,
// dashboard and the hook backends are run by a party; that party now says so, and the tree is
// chosen the same way for every caller: by the identity, never by its absence.
// OperatorRole is ALIASED from internal/seatclass, which declares it beside the roster it keys.
// A second const here would be the same string with two authors.
const OperatorRole = seatclass.OperatorRole

var roleSeats = map[string][]string{
	OperatorRole: {OperatorRole},
	"lens":       {"red-lens-"},
	// The seat is `red-chair-r<n>`; `red-merge-r<n>` is what archived runs hold, and migrate
	// maps it. The role and the seat share the word: the scorecard is named for itself
	// (ScorecardOf, `--card`), so "chair" means the seat and its role and nothing else, and
	// "merge" means only blue's union of the lane drafts.
	"chair": {"red-chair"},
	"blue":  {"blue-", "frontier"},
	// THE BENCH IS ONE PREFIX, BECAUSE IT IS ONE SEAT. `assemble` stayed here after the bench
	// collapsed, and this table is the GRANT: RoleOfSeat scopes the command surface from it, while
	// seatShapes is what REFUSES at register. Removing every refusal and leaving the grant meant
	// `--seat-id assemble-anything` still selected the whole bench tree — and `halt`, the safety
	// stop capture relays verbatim to a human, recorded under an identity no dispatch can create.
	benchRole: {"judge"},
}

// scorecardOfRole maps a seat's ROLE to the CARD — the scorecard — that measures it.
//
// A card is one of three scorecards, red, blue or bench; a role is a seat's verb set. They are
// not the same axis — `lens` and `chair` are two roles measured on ONE card, because a scorecard
// grades how RED is doing on this question, not how one of red's two seats is. Only `operator`
// has no card: it is not a party to the debate, which is why the operator command prints every
// card (or one, with --card) and a seat's own read takes nothing at all.
//
// ONE COPY. The engine does not keep a second: whether a seat gets the scorecard clause has one
// answer, since every role but operator has a card, and the seat asks the tool which card is its.
var scorecardOfRole = map[string]string{
	"lens":  "red",
	"chair": "red",
	"blue":  "blue",
	"bench": "bench",
}

// ScorecardOf reports the card a role is measured on, and whether it has one at all.
//
// The two answers are kept apart rather than collapsed to "": operator having NO card is a fact
// about the run's structure, and a caller that cannot tell it from an unrecognised role would
// print an empty scorecard for both.
func ScorecardOf(role string) (card string, ok bool) {
	c, ok := scorecardOfRole[role]
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
// NOT named RoleOf, and that name is now free because nothing should take it. What used to hold
// it read a lens's AREA off its seat id and is called AreaOf; `role` is what selects a seat's
// SURFACE, which is this file's sense and the terms registry's. Two things were sharing one
// word, which is the collision this whole exercise exists to prevent — and a seat's area is a
// namespace whose collapse once made 39 of 60 disposals ambiguous, so it kept a name of its own
// rather than losing one.
//
// The field is stamped at the write. The fallback re-derives from the id only for records
// written before the field existed — a real corpus this tool still reads — and it is the
// fallback precisely because it is the thing being retired: `strings.HasPrefix(e.SeatID,
// "red-merge")` decided whether a position rendered as RED or BLUE, so an id that failed to
// match its expected prefix rendered as the wrong party with nothing to notice.
// The fallback is guarded on the STAMPED VALUE being usable, not merely present. `role` is
// optional in the schema, so it now has three states where it had two: stamped, stamped-empty,
// and absent. The old code's single `!= ""` collapsed the last two, and that collapse is the
// correct one HERE — both mean "no role was stamped", and both must reach the id fallback. This
// is deliberately not `e.Role != nil`: a record carrying an explicit empty role would otherwise
// return "" as if that were the party, which is the wrong-party-silently failure the field was
// added to end.
func PartyOf(e *Event) string {
	if role := e.GetRole(); role != "" {
		return role
	}
	return roleOfSeat(e.GetSeatId())
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

// roleSample is one seat id per role, for building that role's command tree in the gates and the
// surface walk.
//
// IT IS A CHOICE, NOT A DERIVED FACT, which is why it is written down rather than generated.
// Every seat of a role is handed the same tree, so which one stands for it changes only what the
// goldens render — and picking the lowest id by sort would be a derivation that LOOKS principled
// while quietly answering a question nobody asked.
//
// `blue` is a LANE on purpose. It is the one role whose sample can carry a lane id, and a surface
// walk that never renders one stops covering the id shape that TierClassOfSeat and
// dispatchableSeatID both still have to accept. TestEveryRoleHasADispatchableSample holds both
// halves: every value here is an id the engine can produce, and every role on the roster has one.
var roleSample = map[string]string{
	"lens":       "red-lens-evidence",
	"chair":      "red-chair",
	"blue":       "blue-lane-1",
	"bench":      "judge",
	OperatorRole: OperatorRole,
}

// SampleSeatOf is a seat id of the given role, or "" for a role the roster does not carry.
//
// A seat id carries no epoch, so this is a roster id rather than prefix+"r1" — that composition
// stopped producing an id any dispatch creates.
func SampleSeatOf(role string) string { return roleSample[role] }
