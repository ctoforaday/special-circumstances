package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// IDENTITY IS DETECTED, NOT ASSERTED — end to end, across the seam where it used to be dropped.
//
// The two halves were each fine on their own and had never been driven together. FEOV_SEAT had
// READERS AND NO WRITER: the tool resolved an injected seat id that nothing on earth produced,
// because nothing upstream of `register` knows which agent holds which seat — the dispatcher does
// not learn the id (Workflow's agent() returns a result, not a handle), and recovering one from
// the seat's own command would be a fact read back out of prose.
//
// So the hook exports the fact it HAS — agent_id, the only payload field that discriminates one
// seat from another (session_id and prompt_id are byte-identical across every concurrent
// subagent; plans/hook-surface-spike.md §7) — and `register` writes the join.
//
// This test is the one the feature needed and the unit tests could not be: it takes a real hook
// payload, applies the REWRITE the hook would emit, puts the exported variables into the process
// the way a shell would, and then checks the record. Everything before this proved the injection
// was well formed. Nothing proved it ARRIVED.
func TestTheHookInjectsAnIdentityThatRegisterRecords(t *testing.T) {
	runDir := seatRun(t)
	const agent = "agent_01H9XKQ7"

	out, payload := hookgate.PreOutcome(hookgate.Input{
		AgentID:   agent,
		ToolName:  "Bash",
		ToolInput: []byte(`{"command":"\"/c/bin/feov-record\" register"}`),
	}, runDir)
	if out != hookgate.OutcomeRewrite {
		t.Fatalf("the hook had no opinion on a feov-record call; nothing would be injected (outcome %v)", out)
	}
	applyExports(t, payload)

	if _, err := run(t, "register", "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	got := lastOfType(t, runDir, "register").Payload.Str("agent_id")
	if got != agent {
		t.Fatalf("the register event does not carry the agent that made the call: got %q, want %q.\n"+
			"The binding is the whole mechanism — without it every later call is back to trusting whatever --seat-id it is handed.", got, agent)
	}
}

// A RUN WHOSE HOOK NEVER FIRED SAYS SO, rather than recording an agent whose handle is "".
//
// This is the distinction the whole suite keeps losing: the absent case and the healthy case must
// not be the same bytes. An empty agent_id written as a field would make "identity was never
// measured here" read exactly like "an agent with an empty id registered", and every audit
// downstream would count the second.
func TestNoHookMeansNoAgentField(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "")

	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L1"); err != nil {
		t.Fatalf("register: %v", err)
	}
	ev := lastOfType(t, runDir, "register")
	if ev.Payload.Has(recordAgentKey) {
		t.Errorf("an unhooked run wrote an agent_id field anyway: %v", ev.Payload.Keys())
	}
	// And the register itself must still work. A seat whose harness has no hook is not a seat
	// that cannot record.
	if ev.SeatID != "red-lens-r1-L1" {
		t.Errorf("register did not record the seat: %+v", ev)
	}
}

// THE MAIN SESSION IS NOT A SEAT, and its payload carries no agent_id at all — measured, §7. The
// hook must not manufacture one, or the operator's own calls start looking like a seat's.
func TestAMainSessionCallGetsNoIdentity(t *testing.T) {
	_, payload := hookgate.PreOutcome(hookgate.Input{
		ToolName:  "Bash",
		ToolInput: []byte(`{"command":"feov-record scorecard"}`),
	}, "/runs/x")
	if strings.Contains(payload, seatenv.AgentVar) {
		t.Errorf("a main-session call was given an identity:\n%s", payload)
	}
	if !strings.Contains(payload, seatenv.Var) {
		t.Errorf("the run directory stopped being injected when the identity was absent:\n%s", payload)
	}
}

// TWO CONCURRENT SEATS KEEP THEIR OWN HANDLES. The ids are distinct across concurrent agents
// (§7), and the record has to preserve that or the binding cannot answer the question it exists
// for — which agent held this seat.
func TestConcurrentSeatsRecordDistinctAgents(t *testing.T) {
	runDir := seatRun(t)
	seats := map[string]string{"red-lens-r1-L1": "agent_aaa", "red-lens-r1-L2": "agent_bbb"}

	for seat, agent := range seats {
		t.Setenv(seatenv.AgentVar, agent)
		if _, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil {
			t.Fatalf("register %s: %v", seat, err)
		}
	}
	for _, ev := range events(t, runDir) {
		if ev.Type != "register" {
			continue
		}
		if want, ok := seats[ev.SeatID]; ok && ev.Payload.Str("agent_id") != want {
			t.Errorf("seat %s bound to %q, want %q — the handles crossed", ev.SeatID, ev.Payload.Str("agent_id"), want)
		}
	}
}

const recordAgentKey = "agent_id"

// applyExports puts the variables from a hook rewrite into this process, the way the shell that
// receives the rewritten command would.
//
// It reads them out of the EMITTED STRING rather than setting what the test knows it asked for —
// so a rewrite that silently dropped a variable fails here instead of passing on values the test
// supplied to itself.
func applyExports(t *testing.T, payload string) {
	t.Helper()
	var found int
	for _, part := range strings.Split(payload, "; ") {
		rest, ok := strings.CutPrefix(part, "export ")
		if !ok {
			break // the assignments lead; the first non-assignment is the seat's command
		}
		name, value, ok := strings.Cut(rest, "=")
		if !ok {
			continue
		}
		t.Setenv(name, strings.ReplaceAll(strings.Trim(value, "'"), `'\''`, `'`))
		found++
	}
	if found == 0 {
		t.Fatalf("no exports to apply — the rewrite carried nothing:\n%s", payload)
	}
}
