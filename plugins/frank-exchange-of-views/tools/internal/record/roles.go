package record

import (
	"fmt"

	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A SEAT'S ROLE IS ITS ROSTER ROW, and the role selects its surface.
//
// The verb set being the role boundary is the engine's premise: a lens mints its own gaps and, as
// the originator, regrades and closes them; the chair runs the debate and mints and closes
// nothing; blue edits the report and has no board verb; the bench rules and never originates. The
// record is the evidence those boundaries held, so which role a seat id has is answered in one
// place — seatclass.RoleOfSeat, read from internal/seatclass's roster — and every question here is
// a call to it. An id the roster does not carry has no role and selects no surface.
//
// A petition sitting is `judge`, like the bench's other three. Which of the four questions a
// sitting answers rides on the register's `occasion` (see occasion.go), and who filed the petition
// rides on the petition.
//
// OPERATOR IS A SEAT ID, not the absence of one. Setup, capture, dashboard and the hook backends
// are run by a party; that party says so, and the tree is chosen the same way for every caller: by
// the identity, never by its absence. OperatorRole is ALIASED from internal/seatclass, which
// declares it beside the roster it keys.
const OperatorRole = seatclass.OperatorRole

// gavelSeatOf resolves a motion's `ruled_by` role to the one seat that holds the gavel, through
// the roster every other role question reads, so a gavel cannot rest with a seat the roster does
// not give that role. A gavel rests with ONE seat: a role seated as a family of seats (the lenses,
// blue, the lanes) is an error, as is a word that names no role. Printing the word as though it
// were a seat id would send the reader to a seat that does not exist.
func gavelSeatOf(ruler string) (string, error) {
	if ruler == seatclass.Lane.Role {
		return "", fmt.Errorf("its gavel is held by the %s role, which the roster seats as one seat per lane — a gavel rests with one seat", ruler)
	}
	seats := seatclass.SeatsOfRole(ruler)
	if len(seats) == 0 {
		return "", fmt.Errorf("its gavel is held by %q, which is no role in the roster — a motion nobody can rule", ruler)
	}
	if len(seats) == 1 {
		return seats[0], nil
	}
	return "", fmt.Errorf("its gavel is held by the %s role, which the roster seats as more than one seat (%s) — a gavel rests with one seat", ruler, strings.Join(seats, ", "))
}

// PartyOf answers "which party wrote this": the role of the seat the event is attributed to.
//
// It is named for the party and not for the role because `role` is what selects a seat's SURFACE
// (the terms registry's sense), and a lens's AREA is a third fact with its own name (AreaOf). The
// event stores the seat id and no role, so a role can change spelling without a migration.
func PartyOf(e *Event) string { return seatclass.RoleOfSeat(e.GetSeatId()) }

// RequireDispatchedSeat asserts the seat id is one the roster seats, with no claim about which
// role it belongs to.
//
// It is the half of the role guard that holds where there is no role to enforce — the motion
// commands, whose parent is their SUBJECT. Every seat files on the subjects it does not rule
// (rulerfiles.go) and that asymmetry is deliberate; being a seat at all is not optional. A motion
// attributed to an invented identity would be rendered in the report, ruled on, and joined to a
// gap.
func RequireDispatchedSeat(seatID string) error {
	if seatclass.RoleOfSeat(seatID) != "" {
		return nil
	}
	return feov.Errorf(feov.RoleViolation, "seat %q does not belong to any role namespace — "+
		"the engine assigns the seat id; a hand-invented one records under an identity no dispatch created. "+
		"A motion is filed by a seat the engine created", seatID)
}

// roleSample is one seat id per role, for building that role's command tree in the gates and the
// surface walk.
//
// IT IS A CHOICE, NOT A DERIVED FACT, which is why it is written down rather than generated.
// Every seat of a role is handed the same tree, so which one stands for it changes only what the
// goldens render — and picking the lowest id by sort would be a derivation that LOOKS principled
// while quietly answering a question nobody asked.
//
// `lane` is a lane id, which is what makes a surface walk render the generated id shape that
// TierClassOfSeat and dispatchableSeatID both still have to accept. TestEveryRoleHasADispatchableSample
// holds both halves: every value here is an id the engine can produce, and every role on the roster
// has one.
var roleSample = map[string]string{
	"lens":              "red-lens-evidence",
	"chair":             "red-chair",
	"blue":              "blue-respond",
	seatclass.Lane.Role: seatclass.LaneSeatPrefix + "1",
	"bench":             "judge",
	OperatorRole:        OperatorRole,
}

// SampleSeatOf is a seat id of the given role, or "" for a role the roster does not carry.
//
// A seat id carries no epoch, so this is a roster id rather than prefix+"r1" — that composition
// stopped producing an id any dispatch creates.
func SampleSeatOf(role string) string { return roleSample[role] }
