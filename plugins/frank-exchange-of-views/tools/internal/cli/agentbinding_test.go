package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

	if _, err := run(t, "register", "--seat-id", "red-lens-r1-evidence"); err != nil {
		t.Fatalf("register: %v", err)
	}
	got := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_REGISTER).GetRegister().GetAgentId()
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

	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-evidence"); err != nil {
		t.Fatalf("register: %v", err)
	}
	ev := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_REGISTER)
	// PRESENCE, not the empty string: agent_id is an optional field, so "never measured" is the
	// field being ABSENT. Reading it with GetAgentId() would collapse the two cases this test exists
	// to keep apart.
	if ev.GetRegister().AgentId != nil {
		t.Errorf("an unhooked run wrote an agent_id field anyway: %q", ev.GetRegister().GetAgentId())
	}
	// And the register itself must still work. A seat whose harness has no hook is not a seat
	// that cannot record.
	if ev.GetSeatId() != "red-lens-r1-evidence" {
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
	seats := map[string]string{"red-lens-r1-evidence": "agent_aaa", "red-lens-r1-adversary": "agent_bbb"}

	for seat, agent := range seats {
		t.Setenv(seatenv.AgentVar, agent)
		if _, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil {
			t.Fatalf("register %s: %v", seat, err)
		}
	}
	// THE LATEST REGISTER PER SEAT, which is what the binding IS. seatRun already registered
	// red-lens-r1-evidence with no agent, and this loop asserted over EVERY register event — so it
	// failed on the fixture's own earlier sitting rather than on the binding under test.
	//
	// Taking the last one is not a way around that: it is the rule agentbinding.go states. A
	// re-dispatch writes a fresh register, both stay on the record because it is append-only, and
	// the binding is the most recent claim. Asserting over all of them would refuse every resume.
	bound := map[string]string{}
	for _, ev := range events(t, runDir) {
		if ev.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			continue
		}
		bound[ev.GetSeatId()] = ev.GetRegister().GetAgentId()
	}
	for seat, want := range seats {
		if bound[seat] != want {
			t.Errorf("seat %s bound to %q, want %q — the handles crossed", seat, bound[seat], want)
		}
	}
}

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

// AN AGENT THAT NEVER REGISTERED CANNOT ACT, and that refusal is what makes the binding a
// mechanism rather than a note.
//
// Without it a seat could skip `register`, keep typing --seat-id, and file events under any id
// the tree will build for it — which is exactly the self-asserted identity this replaces. The
// binding would be advisory, and an advisory guarantee is one the audit reads as enforced.
func TestAnUnregisteredAgentIsRefused(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_never_registered")

	_, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-r1-evidence",
		"--reason", "acting without an identity", "--type", "defect")
	if err == nil {
		t.Fatal("an agent with no binding on the record filed an event anyway — the seat id was taken on trust, which is the thing this replaces")
	}
	// THE REFUSAL MUST NAME THE ONE REMEDY. A seat handed an unexplained refusal logs friction
	// and works around it, losing the capability for the run — measured, and the reason every
	// refusal in this tree carries its own way out.
	if !strings.Contains(err.Error(), "register") {
		t.Errorf("the refusal does not name the act that fixes it: %v", err)
	}
}

// AND REGISTER ITSELF MUST RUN UNBOUND, or the mechanism cannot start. This is the bootstrap and
// the one asymmetry in it: register is what CREATES the binding, so it is the single verb that
// may act on a claim.
func TestRegisterIsTheOneVerbThatMayRunUnbound(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_bootstrapping")

	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-evidence"); err != nil {
		t.Fatalf("register was refused for want of the binding it exists to write — the mechanism cannot start: %v", err)
	}
	// And now the SAME agent may act, with no --seat-id at all. That is the whole point: the id
	// is typed once, at register, and never again.
	if _, err := run(t, "log", "--run", runDir, "--reason", "the tool has no path for X", "--type", "defect"); err != nil {
		t.Fatalf("a registered agent still could not act without retyping its seat id: %v", err)
	}
	if got := lastOfType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG).GetSeatId(); got != "red-lens-r1-evidence" {
		t.Errorf("the event was filed under %q; the binding did not carry the identity", got)
	}
}

// NO HANDLE MEANS NO DEMAND. A test, an operator at a shell, or any harness without a PreToolUse
// hook cannot bind — refusing them would be requiring a mechanism their environment does not
// have. The check is keyed on the handle's PRESENCE, and this is the arm that says so.
func TestWithoutAHandleTheFlagStillWorks(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "")

	if _, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-r1-evidence",
		"--reason", "no hook in this environment", "--type", "defect"); err != nil {
		t.Fatalf("an unhooked caller was held to a binding it has no way to create: %v", err)
	}
}

// A --seat-id CONTRADICTING THE BINDING IS REFUSED, naming both — the guarantee that used to be
// checked against FEOV_SEAT, a variable no run could set, and is now checked against the record.
func TestAFlagContradictingTheBindingIsRefused(t *testing.T) {
	runDir := seatRun(t)
	t.Setenv(seatenv.AgentVar, "agent_lens_one")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-evidence"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "log", "--run", runDir, "--seat-id", "red-lens-r1-adversary", "--reason", "x", "--type", "defect")
	if err == nil {
		t.Fatal("an event was filed under a seat this agent did not register as; every found_by and estoppel downstream reads that attribution")
	}
	for _, want := range []string{"red-lens-r1-evidence", "red-lens-r1-adversary"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name BOTH ids so the seat can see which is which; missing %q in %v", want, err)
		}
	}
}
