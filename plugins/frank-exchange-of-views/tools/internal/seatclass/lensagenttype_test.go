package seatclass

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// THE HOOKS' LENS TEST IS THE ROSTER'S. The PostToolUse hook hands a lens its next item, and it
// decides "is this a lens" from the payload's agent type without linking the roster
// (runlive.IsLensAgentType). Every roster row must answer it as its role does: a lens configuration
// that failed it would never be handed its next item, and any other seat that passed it would spawn
// the writer on every call.
func TestTheHooksLensTestAgreesWithTheRoster(t *testing.T) {
	for id, s := range Seats {
		if s.AgentType == "" {
			continue
		}
		if got, want := runlive.IsLensAgentType(s.AgentType), s.Role == "lens"; got != want {
			t.Errorf("seat %s (%s, role %s): IsLensAgentType = %v, want %v", id, s.AgentType, s.Role, got, want)
		}
	}
	if runlive.IsLensAgentType(Lane.AgentType) {
		t.Errorf("a lane's configuration %s reads as a lens's", Lane.AgentType)
	}
}
