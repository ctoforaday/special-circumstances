package cli

import (
	"io"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

// A HOOK BRACKET OPENS A SITTING (#1262). The bracket here is the SubagentStart writer's own write,
// for a seat seatRun's build never registered.
const (
	bracketedSeat   = "red-lens-logic"
	bracketedConfig = "frank-exchange-of-views:red-lens-logic"
)

func bracketSeat(t *testing.T, runDir, agent string) {
	t.Helper()
	if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: runDir, Phase: sittingwrite.Open, AgentID: agent, AgentType: bracketedConfig}, io.Discard); err != nil {
		t.Fatalf("the hook's bracket for %s: %v", agent, err)
	}
}

// `ops diagnostics` reads the seats that SAT: a seat whose only sitting its bracket opened is one.
func TestDiagnosticsReadsASeatSeatedByItsBracketAlone(t *testing.T) {
	runDir := seatRun(t)
	bracketSeat(t, runDir, "agent_bracketed")
	trajectoryFor(t, runDir, "nothing")
	_, err := run(t, "ops", "diagnostics", "--seat-id", "operator", "--run", runDir, "--sitting", bracketedSeat)
	if err != nil && strings.Contains(err.Error(), bracketedSeat) && strings.Contains(err.Error(), "in this run") {
		t.Fatalf("a seat its bracket seated was refused as one that never sat: %v", err)
	}
}

// THE CALL COUNT IS KEYED BY THE STORED SITTING, so the limit the hook hands the writer is one the
// record holds. A register under the agent its bracket seated joins that sitting rather than
// opening one, a second is refused, and the count stays the sitting's.
func TestASecondRegisterInABracketedSittingCountsInThatSitting(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.Var, runDir)
	t.Setenv(seatenv.AgentVar, "agent_bracketed")
	t.Setenv(seatenv.TypeVar, bracketedConfig)
	bracketSeat(t, runDir, "agent_bracketed")
	if _, err := run(t, "register", "--seat-id", bracketedSeat); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := run(t, "register", "--seat-id", bracketedSeat); err == nil || feov.CodeOf(err) != string(feov.Conflict) {
		t.Fatalf("the second register in the sitting = %v; want it refused", err)
	}
	r, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	sittings, err := record.SittingsOf(r, bracketedSeat)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := header(t, runDir, "agent_bracketed")
	if !ok {
		t.Fatal("register opened no count")
	}
	if sittings != 1 || h.Sitting != sittings {
		t.Errorf("the count is for sitting %d and the seat has %d stored sitting(s), want both 1 — the bracket opened the sitting and the register joined it", h.Sitting, sittings)
	}
	if err := sittingwrite.WriteLimit(runDir, "agent_bracketed", h.Sitting, 150); err != nil {
		t.Errorf("the limit of the sitting the count was keyed to was refused: %v", err)
	}
}
