package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// THE M31 SEQUENCE THROUGH THE VERBS: a seat's first act registers it silently, then the seat runs
// `register` itself, as its surface tells it to before it writes. The fixture's own register (no
// agent id) opened sitting 1 and the silent one opened sitting 2, so each explicit register answers
// dispatch 2 and writes nothing.
func TestRegisterAfterTheFirstActRegisteredItWritesNothing(t *testing.T) {
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
			if e.GetSeatId() == "red-lens-evidence" && e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
				registers++
			}
		}
		return len(fam.Events), registers
	}
	before, regsBefore := events()
	first, err := run(t, "register", "--json", "--run", runDir, "--seat-id", "red-lens-evidence")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	again, err := run(t, "register", "--json", "--run", runDir, "--seat-id", "red-lens-evidence")
	if err != nil {
		t.Fatalf("register again: %v", err)
	}
	if first != again {
		t.Errorf("two registers in one dispatch answered differently:\n%s\n---\n%s", first, again)
	}
	if !strings.Contains(strings.Join(strings.Fields(first), ""), `"dispatch":2,`) {
		t.Errorf("the explicit register after an auto-register answered %s; want dispatch 2", first)
	}
	if after, regs := events(); after != before || regs != regsBefore {
		t.Errorf("the explicit registers wrote %d event(s) (%d -> %d registers)", after-before, regsBefore, regs)
	}
}
