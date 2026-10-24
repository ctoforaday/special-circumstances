package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

// THE CALL THAT RETURNS A SEAT'S ENVELOPE IS DENIED WITH THE RECORD'S REFUSAL, from the whole
// PreToolUse path, and only that call: the bench's docket sitting cannot end while a docket motion
// on a gap it is convened on stands unruled, and the harness re-asks a seat whose envelope call is
// refused.
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
			"agent_id": "a7cc47571ab8b015d", "agent_type": "frank-exchange-of-views:lead-judge"})
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
	refusal, checkErr = "", errors.New("feov-sitting-write is not beside this hook")
	if d, _ := pre(envelopeTool); d == "deny" {
		t.Error("an envelope whose check could not run was denied — every sitting of the run would end in a null return")
	}
}
