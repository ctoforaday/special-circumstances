package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
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
