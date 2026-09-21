package record

import (
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// THE ATTESTATION TABLE: which roles an agent configuration is allowed to be seated as.
//
// The roster gate (roster.go) bounds a seat id's shape and, for a lens, its area. What it cannot
// see is WHO registered: nothing there could tell that a `lead-judge` agent had registered as
// `red-chair`, because both are ids the engine really does dispatch. This table is the
// agent-to-role half, and it is the first fact in this system about a seat that the seat did not
// supply: `agent_type` comes off the PreToolUse payload, and a seat cannot state it or withhold it.
//
// A SET PER TYPE IS NO LONGER NEEDED FOR RED, AND THAT IS THE POINT OF THE SPLIT. This table
// carried `red-auditor -> {lens, merge}` and said so as a real limit: debate.js dispatched both
// the lenses and the chair as one configuration, so an attestation of red-auditor admitted either
// and could refuse neither. Every red seat now has its OWN configuration, so `agent_type` names
// WHICH lens — not merely that it is one — and the ambiguous row is gone.
//
// The sets remain sets because blue's are genuinely one-to-many and because a scalar here would
// have to be widened back the first time any configuration seats two roles.
//
// KEYED ON THE FULL PREFIXED STRING, exactly as the harness delivers it and as
// hookgate.AuthorAgentType already spells it. A bare-keyed table would map every real seat to
// unattested, which is the silent pass this exists to remove.
// attestation is what one agent configuration may be seated as.
//
// SEAT IS A FIELD, NOT A SUBSTRING OF THE TYPE. `frank-exchange-of-views:red-lens-voice` contains
// its seat id, and recovering it with a prefix-strip is the shape this repository keeps removing:
// a fact composed into a name at one end and pulled back out at the other, whose miss is a
// plausible zero. Written here, a configuration that stops being one-to-one says so by having no
// seat rather than by a regex quietly matching something else.
//
// EMPTY MEANS THE CONFIGURATION SEATS SEVERAL, which is a real answer and not a gap in the table:
// `blue-researcher` dispatches blue-lane-N, blue-respond and frontier, so nothing but the seat's
// own word can say which one registered. TestEveryDispatchedAgentTypeIsAttestable keeps the rows
// honest against debate.js.
type attestation struct {
	roles []string
	// seat is the ONE seat id this configuration is ever dispatched as, or "" where it seats more
	// than one. Where it is set, a sitting is knowable from the hook alone.
	seat string
}

var agentTypeRoles = map[string]attestation{
	"frank-exchange-of-views:red-lens-evidence":     {[]string{"lens"}, "red-lens-evidence"},
	"frank-exchange-of-views:red-lens-logic":        {[]string{"lens"}, "red-lens-logic"},
	"frank-exchange-of-views:red-lens-dark-side":    {[]string{"lens"}, "red-lens-dark-side"},
	"frank-exchange-of-views:red-lens-voice":        {[]string{"lens"}, "red-lens-voice"},
	"frank-exchange-of-views:red-lens-computation":  {[]string{"lens"}, "red-lens-computation"},
	"frank-exchange-of-views:red-lens-adversary":    {[]string{"lens"}, "red-lens-adversary"},
	"frank-exchange-of-views:red-lens-architecture": {[]string{"lens"}, "red-lens-architecture"},
	"frank-exchange-of-views:red-chair":             {[]string{"chair"}, "red-chair"},
	"frank-exchange-of-views:blue-synthesizer":      {[]string{"blue"}, "blue-synthesize"},
	"frank-exchange-of-views:lead-judge":            {[]string{"bench"}, "judge"},
	// ONE CONFIGURATION, THREE SEATS — so no seat id, and blue keeps paying for its own register.
	"frank-exchange-of-views:blue-researcher": {[]string{"blue"}, ""},
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
// turning an incomplete table into a broken run. TestEveryDispatchedAgentTypeIsAttestable is what
// keeps the table complete instead; the runtime stays permissive on purpose.
func CheckAttestedRole(agentType, seatID string) error {
	if agentType == "" {
		return nil
	}
	a, known := agentTypeRoles[agentType]
	if !known {
		return nil
	}
	allowed := a.roles
	role := roleOfSeat(seatID)
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
