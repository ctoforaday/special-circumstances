package record

import (
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// THE ATTESTATION TABLE: which roles an agent configuration is allowed to be seated as.
//
// The roster gate (roster.go) bounds the SHAPE of a seat id and says so itself — "it bounds the
// SHAPE, never the membership". `red-lens-r9-L9` is well formed and no run will dispatch it, and
// nothing could tell that a `lead-judge` agent had registered as `red-merge-r1`. This table is the
// membership half, and it is the first fact in this system about a seat that the seat did not
// supply: `agent_type` comes off the PreToolUse payload, and a seat cannot state it or withhold it.
//
// A SET PER TYPE, NOT A ROLE PER TYPE, because the mapping genuinely is not one-to-one and
// pretending otherwise would put a lie in a table. debate.js dispatches BOTH the lenses (:845) and
// the merge (:871) as `frank-exchange-of-views:red-auditor`, so an attestation of red-auditor
// admits either and can refuse neither. That is a real limit on what this can catch and it is
// written here rather than discovered later:
//
//	red-auditor       lens, merge     <- the one ambiguous row
//	blue-researcher   blue
//	blue-synthesizer  blue
//	lead-judge        bench
//
// It narrows to exact if red-merge is ever given its own agent configuration: red-auditor's set
// loses `merge` and a red-merge row appears. Nothing else changes, which is why the shape is a set
// today rather than a scalar with a special case bolted on later.
//
// KEYED ON THE FULL PREFIXED STRING, exactly as the harness delivers it and as
// hookgate.AuthorAgentType already spells it. A bare-keyed table would map every real seat to
// unattested, which is the silent pass this exists to remove.
var agentTypeRoles = map[string][]string{
	"frank-exchange-of-views:red-auditor":      {"lens", "merge"},
	"frank-exchange-of-views:blue-researcher":  {"blue"},
	"frank-exchange-of-views:blue-synthesizer": {"blue"},
	"frank-exchange-of-views:lead-judge":       {"bench"},
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
	allowed, known := agentTypeRoles[agentType]
	if !known {
		return nil
	}
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
			"prompt says, the dispatch is wrong rather than your call, and the friction verb is the channel for it",
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
