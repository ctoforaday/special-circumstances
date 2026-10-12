package record

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EVERY ROLE A TYPE CLAIMS IS A ROLE THAT EXISTS. A typo'd role in the table is not a compile
// error, and it would make the check pass every seat of that family forever.
func TestEveryAttestedRoleIsARealRole(t *testing.T) {
	for at, a := range agentTypeRoles {
		for _, r := range a.roles {
			if _, ok := roleSeats[r]; !ok {
				t.Errorf("agentTypeRoles[%q] admits role %q, which roleSeats does not define", at, r)
			}
		}
	}
	// And every debating role is reachable from some attestation — otherwise a whole family
	// registers unattested and nothing says so.
	reachable := map[string]bool{}
	for _, a := range agentTypeRoles {
		for _, r := range a.roles {
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
	err := CheckAttestedRole("frank-exchange-of-views:lead-judge", "red-chair")
	if err == nil {
		t.Fatal("a lead-judge agent registered as red-chair and nothing refused it")
	}
	for _, want := range []string{"red-chair", "chair", "lead-judge", "bench seats"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal does not name %q — it must name both sides for the seat to act on it:\n%s", want, err)
		}
	}
}

// UNATTESTED AND UNKNOWN BOTH PASS, and both are load-bearing: an operator at a shell has no hook,
// and a build that has not learned a new type must not refuse every seat of that kind.
func TestUnattestedAndUnknownTypesArePermitted(t *testing.T) {
	if err := CheckAttestedRole("", "red-chair"); err != nil {
		t.Errorf("an unattested caller was refused, which demands a mechanism its environment lacks: %v", err)
	}
	if err := CheckAttestedRole("frank-exchange-of-views:some-future-seat", "red-chair"); err != nil {
		t.Errorf("an unknown agent type was refused, which would break every run adding one: %v", err)
	}
}

// THE TABLE IS THE ROSTER READ BY CONFIGURATION. A configuration one seat names is that seat, from
// the hook alone; one several seats name — or a lane, whose id no roster row can hold — is no one
// seat, and admits every role those seats hold. A seat with no configuration attests nothing.
func TestTheAttestationIsDerivedFromTheRoster(t *testing.T) {
	got := attestationsOf(map[string]seatclass.Seat{
		"solo":     {Role: "lens", AgentType: "t:solo"},
		"shared-a": {Role: "blue", AgentType: "t:shared"},
		"shared-b": {Role: "bench", AgentType: "t:shared"},
		"beside":   {Role: "blue", AgentType: "t:lane"},
		"nobody":   {Role: "operator"},
	}, seatclass.Seat{Role: "blue", AgentType: "t:lane"})
	for _, c := range []struct {
		typ, seat string
		roles     []string
	}{
		{"t:solo", "solo", []string{"lens"}},
		{"t:shared", "", []string{"bench", "blue"}},
		{"t:lane", "", []string{"blue"}},
	} {
		a, ok := got[c.typ]
		slices.Sort(a.roles)
		if !ok || a.seat != c.seat || !slices.Equal(a.roles, c.roles) {
			t.Errorf("%s attests seat %q, roles %v (row present: %v); want seat %q, roles %v", c.typ, a.seat, a.roles, ok, c.seat, c.roles)
		}
	}
	if len(got) != 3 {
		t.Errorf("%d configuration(s) attested, want 3 — a seat with no configuration is no row: %v", len(got), got)
	}
	// A lane ALONE under a configuration is still no one seat: its id is generated per run.
	if a := attestationsOf(nil, seatclass.Seat{Role: "blue", AgentType: "t:lane"})["t:lane"]; a.seat != "" || !slices.Equal(a.roles, []string{"blue"}) {
		t.Errorf("a configuration seating only lanes attests seat %q, roles %v; want no seat and blue", a.seat, a.roles)
	}

	// AND THE LIVE TABLE IS THAT READING OF THE LIVE ROSTER: every seat's configuration admits the
	// seat's own role, and names the seat exactly where no other seat shares it.
	for id, s := range seatclass.Seats {
		if s.AgentType == "" {
			continue
		}
		if err := CheckAttestedRole(s.AgentType, id); err != nil {
			t.Errorf("seat %s is refused under its own configuration %s: %v", id, s.AgentType, err)
		}
		sole := s.AgentType != seatclass.Lane.AgentType
		for other, o := range seatclass.Seats {
			sole = sole && (other == id || o.AgentType != s.AgentType)
		}
		if seat, ok := SeatOfAgentType(s.AgentType); ok != sole || (sole && seat != id) {
			t.Errorf("SeatOfAgentType(%s) = %q, %v; seat %s names it and is its only seat: %v", s.AgentType, seat, ok, id, sole)
		}
	}
}
