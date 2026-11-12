package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// THE CALL THAT RETURNED A SEAT'S ENVELOPE, HAVING SUCCEEDED, IS HANDED OVER AS ITS ADMISSION — from
// the whole PostToolUse path, for the seat's run and agent, and for that call only. A call from no
// seat writes nothing, and an admission the writer could not record goes on the hook record.
func TestASucceededEnvelopeCallIsAdmitted(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	var admitted []string
	var writeErr error
	prev := recordAdmission
	recordAdmission = func(dir, agent string) error {
		admitted = append(admitted, dir+" | "+agent)
		return writeErr
	}
	t.Cleanup(func() { recordAdmission = prev })
	post := func(tool, agent string) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"tool_name": tool, "tool_input": map[string]string{}, "cwd": cwd,
			"session_id": testLead, "agent_id": agent, "agent_type": "frank-exchange-of-views:lead-judge",
			"tool_response": map[string]string{}})
		var out bytes.Buffer
		if err := Post(bytes.NewReader(b), &out, testRecorder()); err != nil {
			t.Fatal(err)
		}
		if out.Len() != 0 {
			t.Errorf("the admission hook spoke: %s", out.String())
		}
	}

	post(envelopeTool, "a7cc47571ab8b015d")
	if len(admitted) != 1 || admitted[0] != runDir+" | a7cc47571ab8b015d" {
		t.Fatalf("admissions handed over = %q, want the seat's run and agent once", admitted)
	}
	post("Bash", "a7cc47571ab8b015d")
	if len(admitted) != 1 {
		t.Errorf("a call that returns no envelope was admitted: %q", admitted)
	}
	t.Setenv("FEOV_AGENT_ID", "")
	post(envelopeTool, "")
	if len(admitted) != 1 {
		t.Errorf("a call from no seat was admitted: %q", admitted)
	}
	// AN ADMISSION THE WRITER COULD NOT RECORD is the hook's failure, and still exits clean.
	writeErr = errors.New("the record is captured")
	rec := testRecorder()
	b, _ := json.Marshal(map[string]any{"tool_name": envelopeTool, "cwd": cwd, "session_id": testLead,
		"agent_id": "a7cc47571ab8b015d", "agent_type": "frank-exchange-of-views:lead-judge"})
	if err := Post(bytes.NewReader(b), &bytes.Buffer{}, rec); err != nil {
		t.Fatal(err)
	}
	if !rec.Failed() {
		t.Error("an admission the writer could not record did not reach the hook record")
	}
}

// A LENS'S CALL TO THE RECORD HANDS IT WHAT IT IS HANDED NEXT, AS additionalContext — and nothing
// else spawns the writer: another seat's call, a lens's call that does not reach the record, a call
// from no seat. A writer that fails is said to the lens, never silence.
func TestALensCallToTheRecordIsHandedItsNextItem(t *testing.T) {
	cwd, runDir, _ := limitRun(t, 1000)
	var asked []string
	text, writeErr := "YOUR NEXT ITEM: …", error(nil)
	prev := handNextItem
	handNextItem = func(dir, agent, typ string) (string, error) {
		asked = append(asked, dir+" | "+agent+" | "+typ)
		return text, writeErr
	}
	t.Cleanup(func() { handNextItem = prev })
	const lens = "frank-exchange-of-views:red-lens-logic"
	post := func(agentType, command string) (string, *bytes.Buffer) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": command}, "cwd": cwd,
			"session_id": testLead, "agent_id": "a7cc47571ab8b015d", "agent_type": agentType, "tool_response": map[string]string{}})
		var out bytes.Buffer
		if err := Post(bytes.NewReader(b), &out, testRecorder()); err != nil {
			t.Fatal(err)
		}
		var doc struct {
			HookSpecificOutput struct {
				HookEventName     string `json:"hookEventName"`
				AdditionalContext string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if out.Len() > 0 {
			if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
				t.Fatalf("the hook printed %q, not a PostToolUse document: %v", out.String(), err)
			}
		}
		return doc.HookSpecificOutput.AdditionalContext, &out
	}

	got, _ := post(lens, `"/plugin/bin/feov-record" lens close --gap G-1`)
	if len(asked) != 1 || asked[0] != runDir+" | a7cc47571ab8b015d | "+lens || got != text {
		t.Fatalf("asked %q, handed %q — want the lens's run, agent and type asked once, and the answer handed on", asked, got)
	}
	if _, out := post("frank-exchange-of-views:red-chair", "feov-record show work"); len(asked) != 1 || out.Len() != 0 {
		t.Errorf("a chair's call spawned the writer or spoke: %q %q", asked, out.String())
	}
	if _, out := post(lens, "ls -la"); len(asked) != 1 || out.Len() != 0 {
		t.Errorf("a lens's call that does not reach the record spawned the writer or spoke: %q %q", asked, out.String())
	}
	text = ""
	if _, out := post(lens, "feov-record show report"); len(asked) != 2 || out.Len() != 0 {
		t.Errorf("nothing to hand was said anyway: %q", out.String())
	}
	writeErr = errors.New("the record is locked")
	if got, _ := post(lens, "feov-record show report"); !strings.Contains(got, "COULD NOT BE READ") || !strings.Contains(got, "the record is locked") {
		t.Errorf("a writer that failed was handed to the lens as %q — want the statement that nothing was determined, and why", got)
	}
}
