package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
)

func header(t *testing.T, runDir, agent string) (sittingcap.Header, bool) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(runDir, sittingcap.Dir, agent+".json"))
	if os.IsNotExist(err) {
		return sittingcap.Header{}, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var h sittingcap.Header
	if err := json.Unmarshal(b, &h); err != nil {
		t.Fatal(err)
	}
	return h, true
}

// REGISTER OPENS THE SITTING'S COUNT, numbered as the record numbers the sitting. Each register
// opens the next one — which is what resets the count for a warm session's second sitting.
func TestRegisterOpensEachSittingsCount(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.Var, runDir) // what the hook injects
	t.Setenv(seatenv.AgentVar, "agent_warm")
	r, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	prev := 0
	for i := 1; i <= 2; i++ {
		if _, err := run(t, "register", "--seat-id", "red-lens-evidence"); err != nil {
			t.Fatalf("register %d: %v", i, err)
		}
		h, ok := header(t, runDir, "agent_warm")
		if !ok {
			t.Fatalf("register %d opened no count; the hook would never limit this sitting", i)
		}
		// The fixture may already have registered this seat, so the number is the record's.
		want, err := record.SittingsOf(r, "red-lens-evidence")
		if err != nil {
			t.Fatal(err)
		}
		if h.SeatID != "red-lens-evidence" || h.Sitting != want {
			t.Errorf("after register %d the count is for %s sitting %d, want red-lens-evidence sitting %d (the record's)",
				i, h.SeatID, h.Sitting, want)
		}
		if prev != 0 && h.Sitting != prev+1 {
			t.Errorf("the second register opened sitting %d after %d — the count did not move to the next sitting", h.Sitting, prev)
		}
		prev = h.Sitting
	}
}

// NO AGENT ID, NO COUNT: there is nothing to key it on, and the hook could not find it.
func TestRegisterWithNoAgentOpensNoCount(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.Var, runDir)
	t.Setenv(seatenv.AgentVar, "")
	if _, err := run(t, "register", "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Join(runDir, sittingcap.Dir))
	if len(entries) != 0 {
		t.Errorf("a register with no agent id opened %d count file(s)", len(entries))
	}
}
