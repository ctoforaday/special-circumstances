package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE M31 SEQUENCE THROUGH THE VERBS: a seat's first act registers it silently, then the seat runs
// `register` itself. The fixture's own register (no agent id) opened sitting 1 and the silent one
// opened sitting 2, so each explicit register is refused naming dispatch 2, and writes nothing.
func TestRegisterAfterTheFirstActRegisteredItIsRefused(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_lens_first_act_registered")
	if _, err := run(t, "motion", "petition", "file", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--class", "safety", "--relief", "r", "--reason", "x"); err != nil {
		t.Fatalf("the first act: %v", err)
	}
	events := func() (all, registers int) {
		t.Helper()
		fam, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range fam.Events {
			switch {
			case e.GetType() == recordpb.EventType_EVENT_TYPE_LOG:
				// a refusal is logged; it is not the register
			case e.GetSeatId() == "red-lens-evidence" && e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER:
				registers++
				all++
			default:
				all++
			}
		}
		return all, registers
	}
	before, regsBefore := events()
	for range 2 {
		_, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence")
		if err == nil || feov.CodeOf(err) != string(feov.Conflict) || !strings.Contains(err.Error(), "already registered in this sitting (dispatch 2)") {
			t.Errorf("the explicit register after an auto-register = %v; want it refused naming dispatch 2", err)
		}
	}
	if after, regs := events(); after != before || regs != regsBefore {
		t.Errorf("the explicit registers wrote %d event(s) (%d -> %d registers)", after-before, regsBefore, regs)
	}
}
