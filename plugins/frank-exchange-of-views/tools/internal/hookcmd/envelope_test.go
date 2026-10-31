package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
)

// THE CALL THAT RETURNS A SEAT'S ENVELOPE IS DENIED WITH THE RECORD'S REFUSAL, from the whole
// PreToolUse path, and only that call: a seat's sitting cannot end while a blocking item on its own
// work list stands, and the harness re-asks a seat whose envelope call is refused.
func TestTheHookDeniesAnEnvelopeTheRecordRefuses(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	refusal, checkErr := "", error(nil)
	var asked []string
	prev := envelopeAnswer
	envelopeAnswer = func(dir, agent string) (hookgate.EnvelopeAnswer, error) {
		asked = append(asked, dir+" | "+agent)
		return hookgate.EnvelopeAnswer{Refusal: refusal}, checkErr
	}
	t.Cleanup(func() { envelopeAnswer = prev })
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

// THE READ IS THE HOOK'S, OVER ANYTHING THE SEAT TYPED. The call that returns an envelope is allowed
// with the input it carried and the record's read in record_read: written where the seat left it out,
// overwritten where it typed one, and a seat typing `halted` changes nothing, so it cannot stop the
// run or keep it going. An input that does not parse cannot carry a read, and is refused rather than
// returned with none; a call the record has no seat for is left alone.
func TestTheHookWritesTheRecordsReadOverAnythingTyped(t *testing.T) {
	cwd, _, _ := limitRun(t, 1000)
	const read = `{"event":7,"bench_owed":["petition"],"halted":false}`
	answer := hookgate.EnvelopeAnswer{Read: json.RawMessage(read)}
	prev := envelopeAnswer
	envelopeAnswer = func(string, string) (hookgate.EnvelopeAnswer, error) { return answer, nil }
	t.Cleanup(func() { envelopeAnswer = prev })
	type doc struct {
		H struct {
			Decision string          `json:"permissionDecision"`
			Reason   string          `json:"permissionDecisionReason"`
			Updated  json.RawMessage `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	pre := func(toolInput string) (doc, bool) {
		b, _ := json.Marshal(map[string]any{"tool_name": envelopeTool, "tool_input": json.RawMessage(toolInput), "cwd": cwd,
			"session_id": testLead, "agent_id": "a7cc47571ab8b015d", "agent_type": "frank-exchange-of-views:lead-judge"})
		var out bytes.Buffer
		if err := Pre(bytes.NewReader(b), &out, testRecorder()); err != nil {
			t.Fatal(err)
		}
		var d doc
		if out.Len() == 0 {
			return d, false
		}
		if err := json.Unmarshal(out.Bytes(), &d); err != nil {
			t.Fatalf("not JSON: %v\n%s", err, out.String())
		}
		return d, true
	}
	for _, tc := range []struct{ name, input, keeps string }{
		{"omitted", `{"notes":"mine"}`, "mine"},
		{"typed", `{"notes":"mine","record_read":{"event":999,"bench_owed":[],"halted":false}}`, "mine"},
		{"typed halted", `{"notes":"mine","record_read":{"halted":true}}`, "mine"},
		{"typed nonsense", `{"notes":"mine","record_read":"Not halted"}`, "mine"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, answered := pre(tc.input)
			if !answered || d.H.Decision != "allow" {
				t.Fatalf("answered %v, decision %q (%s), want allow", answered, d.H.Decision, d.H.Reason)
			}
			var updated map[string]json.RawMessage
			if err := json.Unmarshal(d.H.Updated, &updated); err != nil {
				t.Fatalf("updatedInput: %v\n%s", err, d.H.Updated)
			}
			if string(updated["record_read"]) != read {
				t.Errorf("record_read = %s, want the record's %s", updated["record_read"], read)
			}
			if !strings.Contains(string(updated["notes"]), tc.keeps) {
				t.Errorf("the seat's own field did not survive: %s", d.H.Updated)
			}
		})
	}
	t.Run("input that does not parse", func(t *testing.T) {
		if d, _ := pre(`[1,2]`); d.H.Decision != "deny" {
			t.Errorf("an input that cannot carry a read = decision %q, want deny", d.H.Decision)
		}
	})
	t.Run("a call the record has no seat for", func(t *testing.T) {
		answer = hookgate.EnvelopeAnswer{}
		if d, answered := pre(`{"notes":"mine"}`); answered && len(d.H.Updated) != 0 {
			t.Errorf("an unbound call was rewritten: %s", d.H.Updated)
		}
	})
}
