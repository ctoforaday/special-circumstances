package record

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// agentTypeDispatch captures every agentType debate.js dispatches under.
var agentTypeDispatch = regexp.MustCompile(`agentType:\s*'([^']+)'`)

// EVERY TYPE THE ENGINE DISPATCHES IS IN THE TABLE, because a missing row is SILENT.
//
// CheckAttestedRole passes an unknown type deliberately: failing closed on a type this build has
// not been taught would refuse every seat of a newly added kind at its first act, turning an
// incomplete table into a broken run. That choice is only safe while something else keeps the
// table complete — otherwise the permissive branch is exactly the plausible zero this package
// exists to refuse, and the day someone adds an agent configuration the attestation quietly stops
// attesting anything for it.
//
// This is that something else, and it reads the engine rather than a second list of its own.
func TestEveryDispatchedAgentTypeIsAttestable(t *testing.T) {
	src, err := os.ReadFile(debateSource)
	if err != nil {
		t.Fatalf("cannot read debate.js — the dispatch this table is derived from: %v", err)
	}

	ms := agentTypeDispatch.FindAllStringSubmatch(string(src), -1)
	if len(ms) < 10 {
		t.Fatalf("found %d agentType dispatch sites in debate.js, expected at least 10 — this bind is "+
			"reading the wrong thing and would pass on an empty set", len(ms))
	}

	seen := map[string]bool{}
	for _, m := range ms {
		at := m[1]
		seen[at] = true
		if _, known := agentTypeRoles[at]; !known {
			t.Errorf("debate.js dispatches agent type %q and agentTypeRoles has no row for it, so every "+
				"seat under it registers with its role unattested", at)
		}
	}
	for at := range agentTypeRoles {
		if !seen[at] {
			t.Errorf("agentTypeRoles has a row for %q and debate.js dispatches no seat under it", at)
		}
	}
}

// EVERY ROLE A TYPE CLAIMS IS A ROLE THAT EXISTS. A typo'd role in the table is not a compile
// error, and it would make the check pass every seat of that family forever.
func TestEveryAttestedRoleIsARealRole(t *testing.T) {
	for at, roles := range agentTypeRoles {
		for _, r := range roles {
			if _, ok := roleSeats[r]; !ok {
				t.Errorf("agentTypeRoles[%q] admits role %q, which roleSeats does not define", at, r)
			}
		}
	}
	// And every debating role is reachable from some attestation — otherwise a whole family
	// registers unattested and nothing says so.
	reachable := map[string]bool{}
	for _, roles := range agentTypeRoles {
		for _, r := range roles {
			reachable[r] = true
		}
	}
	for r := range roleSeats {
		if r == OperatorRole {
			continue // not dispatched by the engine and attested by nothing; see roster.go.
		}
		if !reachable[r] {
			t.Errorf("role %q is seated by no agent type, so every %s seat registers unattested", r, r)
		}
	}
}

func TestTheAttestationRefusesASeatFromTheWrongFamily(t *testing.T) {
	err := CheckAttestedRole("frank-exchange-of-views:lead-judge", "red-merge-r1")
	if err == nil {
		t.Fatal("a lead-judge agent registered as red-merge-r1 and nothing refused it")
	}
	for _, want := range []string{"red-merge-r1", "merge", "lead-judge", "bench seats"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal does not name %q — it must name both sides for the seat to act on it:\n%s", want, err)
		}
	}
}

// THE AMBIGUOUS ROW IS ASSERTED, not left to be rediscovered. red-auditor seats the lenses AND the
// merge, so the attestation cannot tell them apart and must admit both. If this ever starts
// failing it is because red-merge got its own configuration — at which point the table narrows and
// this test states the new truth.
func TestRedAuditorSeatsBothLensAndMerge(t *testing.T) {
	for _, seat := range []string{"red-lens-r1-evidence", "red-merge-r1"} {
		if err := CheckAttestedRole("frank-exchange-of-views:red-auditor", seat); err != nil {
			t.Errorf("red-auditor cannot seat %s, but debate.js dispatches it: %v", seat, err)
		}
	}
}

// UNATTESTED AND UNKNOWN BOTH PASS, and both are load-bearing: an operator at a shell has no hook,
// and a build that has not learned a new type must not refuse every seat of that kind.
func TestUnattestedAndUnknownTypesArePermitted(t *testing.T) {
	if err := CheckAttestedRole("", "red-merge-r1"); err != nil {
		t.Errorf("an unattested caller was refused, which demands a mechanism its environment lacks: %v", err)
	}
	if err := CheckAttestedRole("frank-exchange-of-views:some-future-seat", "red-merge-r1"); err != nil {
		t.Errorf("an unknown agent type was refused, which would break every run adding one: %v", err)
	}
}
