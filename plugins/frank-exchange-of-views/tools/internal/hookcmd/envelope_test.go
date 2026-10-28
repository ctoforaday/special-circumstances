package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// THE CALL THAT RETURNS A SEAT'S ENVELOPE IS DENIED WITH THE RECORD'S REFUSAL, from the whole
// PreToolUse path, and only that call: a seat's sitting cannot end while a blocking item on its own
// work list stands, and the harness re-asks a seat whose envelope call is refused.
func TestTheHookDeniesAnEnvelopeTheRecordRefuses(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	refusal, checkErr := "", error(nil)
	var asked []string
	prev := envelopeRefusal
	envelopeRefusal = func(dir, agent string) (string, error) {
		asked = append(asked, dir+" | "+agent)
		return refusal, checkErr
	}
	t.Cleanup(func() { envelopeRefusal = prev })
	pre := func(tool string) (decision, reason string) {
		b, _ := json.Marshal(map[string]any{"tool_name": tool, "tool_input": map[string]string{}, "cwd": cwd,
			"session_id": testLead, "agent_id": "a7cc47571ab8b015d", "agent_type": "frank-exchange-of-views:lead-judge"})
		var out bytes.Buffer
		if err := Pre(bytes.NewReader(b), &out, testRecorder()); err != nil {
			t.Fatal(err)
		}
		if out.Len() == 0 {
			return "", ""
		}
		var doc struct {
			H struct {
				Decision string `json:"permissionDecision"`
				Reason   string `json:"permissionDecisionReason"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
			t.Fatalf("not JSON: %v\n%s", err, out.String())
		}
		return doc.H.Decision, doc.H.Reason
	}

	refusal = "your envelope is refused: M2 (on G1)"
	if d, r := pre(envelopeTool); d != "deny" || r != refusal {
		t.Fatalf("an envelope the record refuses went through: decision %q, reason %q", d, r)
	}
	if len(asked) != 1 || asked[0] != runDir+" | a7cc47571ab8b015d" {
		t.Errorf("the record was asked about %q, want the seat's run and agent", asked)
	}
	if d, _ := pre("Glob"); d == "deny" {
		t.Error("a call that returns no envelope was refused with the envelope's refusal")
	}
	if len(asked) != 1 {
		t.Errorf("a call that returns no envelope was checked as one: %q", asked)
	}
	refusal = ""
	if d, _ := pre(envelopeTool); d == "deny" {
		t.Error("an envelope the record does not refuse was denied")
	}
	// A CHECK THAT CANNOT RUN REFUSES, naming the error (an unreadable record refuses).
	refusal, checkErr = "", errors.New("the run's record is not readable at /x: no such file")
	if d, r := pre(envelopeTool); d != "deny" || !strings.Contains(r, "the run's record is not readable at /x: no such file") {
		t.Errorf("an envelope whose check could not run was not refused naming the error: decision %q, reason %q", d, r)
	}
	// A refusal printed before the check failed (the log write) is the refusal that stands.
	refusal, checkErr = "your envelope is refused: M2 (on G1)", errors.New("the log write failed")
	if d, r := pre(envelopeTool); d != "deny" || r != refusal {
		t.Errorf("a refusal beside a failed log was not the denial: decision %q, reason %q", d, r)
	}
}
