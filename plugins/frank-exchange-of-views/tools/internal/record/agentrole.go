package record

import (
	"slices"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE ATTESTATION TABLE: which roles an agent configuration is allowed to be seated as.
//
// The roster gate (roster.go) bounds a seat id to the ids the engine dispatches. What it cannot see
// is WHO registered: nothing there could tell that a `lead-judge` agent had registered as
// `red-chair`, because both are ids the engine really does dispatch. This table is the
// agent-to-role half, and it is the first fact in this system about a seat that the seat did not
// supply: `agent_type` comes off the hook payload, and a seat cannot state it or withhold it.
//
// IT IS COMPUTED FROM THE ROSTER, NOT WRITTEN BESIDE IT. seatclass.Seats names the configuration
// each seat is dispatched under, and this is that column read the other way: per configuration,
// the roles of the seats it seats, and the seat where there is exactly one. A configuration is a
// row here exactly when a seat names it, keyed on the full prefixed string the harness delivers.
//
// EVERY RED SEAT HAS ITS OWN CONFIGURATION, so `agent_type` names WHICH lens sat and not merely
// that one did. The roles remain a set because blue's configuration is one-to-many.
type attestation struct {
	roles []string
	// seat is the ONE seat id this configuration is ever dispatched as, or "" where it seats more
	// than one: `blue-researcher` seats the lanes, blue-respond and the frontier, so nothing but
	// the seat's own word can say which one registered. Where it is set, a sitting is knowable
	// from the hook alone.
	seat string
}

var agentTypeRoles = attestationsOf(seatclass.Seats, seatclass.Lane)

// attestationsOf reads a roster by configuration. lane stands for every lane seat: its ids are
// generated per run, so a configuration that seats a lane seats no ONE id, whatever else it seats.
func attestationsOf(seats map[string]seatclass.Seat, lane seatclass.Seat) map[string]attestation {
	out := map[string]attestation{}
	seated := map[string]int{}
	add := func(id string, s seatclass.Seat) {
		if s.AgentType == "" {
			return // the operator: nothing dispatches it, so nothing attests it
		}
		a := out[s.AgentType]
		if !slices.Contains(a.roles, s.Role) {
			a.roles = append(a.roles, s.Role)
		}
		seated[s.AgentType]++
		a.seat = ""
		if seated[s.AgentType] == 1 {
			a.seat = id
		}
		out[s.AgentType] = a
	}
	for id, s := range seats {
		add(id, s)
	}
	add("", lane)
	return out
}

// SeatOfAgentType is the seat a configuration is always dispatched as, where there is exactly one.
//
// This is what makes a no-op sitting free: the SubagentStart hook records `agent_type` with no
// command from the seat, so for these configurations the record already knows WHO sat before the
// seat has done anything at all. ok is false where the configuration seats several — the honest
// answer, and the one that keeps blue's register load-bearing instead of guessed at.
func SeatOfAgentType(agentType string) (string, bool) {
	a, known := agentTypeRoles[agentType]
	if !known || a.seat == "" {
		return "", false
	}
	return a.seat, true
}

// CheckAttestedRole refuses a seat id whose role the attested agent configuration cannot hold.
//
// UNATTESTED IS NOT A VIOLATION. An empty agentType means nothing attested anything — an operator
// at a shell, CI, a test, the bootstrap window before the hook binaries exist — and refusing those
// would be demanding a mechanism their environment does not have. They keep exactly the behaviour
// they had before this existed.
//
// AN UNKNOWN TYPE IS ALSO NOT A VIOLATION, and this is the deliberate half. A type absent from the
// table is one this build has not been taught, which happens the moment a new agent configuration
// is added — and failing closed there would refuse every seat of the new kind at its first act,
// turning an incomplete table into a broken run. The table holds every configuration the roster
// names, and integration/surface's TestEveryDispatchSeatsItsSeatUnderItsOwnConfigurationAndTier
// holds every dispatch the engine makes to the roster's configuration for its seat; the runtime
// stays permissive on purpose.
func CheckAttestedRole(agentType, seatID string) error {
	if agentType == "" {
		return nil
	}
	a, known := agentTypeRoles[agentType]
	if !known {
		return nil
	}
	allowed := a.roles
	role := seatclass.RoleOfSeat(seatID)
	if role == "" {
		// An id matching no role cannot be checked here. requireDispatchableSeat is what refuses
		// it, and it has a better message for the case than this would.
		return nil
	}
	for _, a := range allowed {
		if a == role {
			return nil
		}
	}
	return feov.Errorf(feov.RoleViolation,
		"seat %q is a %s seat, and this agent is dispatched as %s, which seats %s. Your seat id and the "+
			"configuration you are running under disagree, and the configuration is the half you did not type — "+
			"so the id is what to check. Copy it exactly as SEAT_ID states it in your prompt. If it IS what your "+
			"prompt says, the dispatch is wrong rather than your call, and the log is where it goes",
		seatID, role, agentType, prose(allowed))
}

// prose renders an allowed-role set for the refusal, so the message names what the type CAN seat
// rather than only what it cannot.
func prose(roles []string) string {
	r := append([]string(nil), roles...)
	sort.Strings(r)
	switch len(r) {
	case 0:
		return "no seats"
	case 1:
		return r[0] + " seats"
	default:
		return strings.Join(r[:len(r)-1], ", ") + " and " + r[len(r)-1] + " seats"
	}
}
