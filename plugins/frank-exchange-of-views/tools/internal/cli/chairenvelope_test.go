package cli

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// THE PLAN THE VERB PRINTS IS THE PLAN THE CHAIR'S CALL CARRIES, through the process the hook spawns:
// the writer reads what the chair stated off its stdin and prints the answer. A chair that states the
// plan, trimmed as m37's epoch 4 trimmed it, is refused naming what it carries.
func TestTheWriterBinaryWritesThePlanIntoTheChairsCall(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a binary")
	}
	writer := testbuild.Binary(t, "feov-sitting-write")
	runDir := newRun(t)
	stageEvents(t, runDir)
	const agent = "agent_chair"
	if out, err := exec.Command(writer, "-run", runDir, "-phase", "open", "-agent-id", agent,
		"-agent-type", "frank-exchange-of-views:red-chair").CombinedOutput(); err != nil {
		t.Fatalf("opening the chair's sitting: %v: %s", err, out)
	}
	printed, err := run(t, "dispatch", "next", "--run", runDir, "--seat-id", "red-chair", "--json")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, printed)
	}
	answer := func(t *testing.T, stated string) (refusal string, envelope map[string]any) {
		t.Helper()
		cmd := exec.Command(writer, "-run", runDir, "-phase", "envelope", "-agent-id", agent)
		cmd.Stdin = strings.NewReader(stated)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("the writer: %v: %s", err, stderr.String())
		}
		var a struct {
			Refusal  string         `json:"refusal"`
			Envelope map[string]any `json:"envelope"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &a); err != nil {
			t.Fatalf("the writer printed no answer (%v): %q", err, stdout.String())
		}
		return a.Refusal, a.Envelope
	}

	var plan map[string]any
	if err := json.Unmarshal([]byte(printed), &plan); err != nil {
		t.Fatal(err)
	}
	refusal, envelope := answer(t, `{"notes":"a line"}`)
	if refusal != "" || envelope["notes"] != "a line" {
		t.Fatalf("the call's notes did not reach its envelope: refusal %q, envelope %v", refusal, envelope)
	}
	for k, v := range plan {
		if got, _ := json.Marshal(envelope[k]); string(got) != string(mustMarshal(t, v)) {
			t.Errorf("the envelope's %s is %s, and the verb printed %s", k, got, mustMarshal(t, v))
		}
	}
	if len(envelope) != len(plan)+1 {
		t.Errorf("the envelope holds %d fields; the plan and the notes are %d", len(envelope), len(plan)+1)
	}
	refusal, envelope = answer(t, `{"key":"x","blockers":[]}`)
	if envelope != nil || !strings.Contains(refusal, "`blockers`") || !strings.Contains(refusal, "`key`") {
		t.Errorf("a chair stating the plan was not refused naming what it carries: refusal %q, envelope %v", refusal, envelope)
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
