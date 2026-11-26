package hookcmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// THE CALL THAT RETURNED A SEAT'S ENVELOPE, HAVING SUCCEEDED, IS HANDED OVER AS ITS ADMISSION — from
// the whole PostToolUse path, for the seat's run and agent, and for that call only. A call from no
// seat writes nothing, and an admission the writer could not record goes on the hook record.
func TestASucceededEnvelopeCallIsAdmitted(t *testing.T) {
	cwd, runDir := hookRun(t)
	var admitted []string
	var writeErr error
	prev := recordAdmission
	recordAdmission = func(dir, agent string) error {
		admitted = append(admitted, dir+" | "+agent)
		return writeErr
	}
	t.Cleanup(func() { recordAdmission = prev })
	prevNext := handNextItem
	handNextItem = func(string, string, string, string) (string, error) { return "", nil }
	t.Cleanup(func() { handNextItem = prevNext })
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

// EVERY CALL OF A SEAT'S HANDS IT WHAT IT IS HANDED NEXT, AS additionalContext, whichever tool it
// used — a Read and a WebSearch as a call to the record — and a call from no seat spawns nothing. A
// writer that fails is said to the lens, never silence.
func TestEverySeatCallIsHandedItsNextItem(t *testing.T) {
	cwd, runDir := hookRun(t)
	var asked []string
	text, writeErr := "YOUR NEXT ITEM: …", error(nil)
	prev := handNextItem
	handNextItem = func(dir, agent, typ, _ string) (string, error) {
		asked = append(asked, dir+" | "+agent+" | "+typ)
		return text, writeErr
	}
	t.Cleanup(func() { handNextItem = prev })
	prevAdmit := recordAdmission
	recordAdmission = func(string, string) error { return nil }
	t.Cleanup(func() { recordAdmission = prevAdmit })
	const lens = "frank-exchange-of-views:red-lens-logic"
	call := func(agentID, agentType, tool string, input map[string]string) (string, *bytes.Buffer) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"tool_name": tool, "tool_input": input, "cwd": cwd,
			"session_id": testLead, "agent_id": agentID, "agent_type": agentType, "tool_response": map[string]string{}})
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
	post := func(agentType, command string) (string, *bytes.Buffer) {
		t.Helper()
		return call("a7cc47571ab8b015d", agentType, "Bash", map[string]string{"command": command})
	}

	got, _ := post(lens, `"/plugin/bin/feov-record" lens close --gap G-1`)
	if len(asked) != 1 || asked[0] != runDir+" | a7cc47571ab8b015d | "+lens || got != text {
		t.Fatalf("asked %q, handed %q — want the lens's run, agent and type asked once, and the answer handed on", asked, got)
	}
	// THE CHAIN IS EVERY SEAT'S: a chair's call to the record is handed its next item as a lens's is.
	const chair = "frank-exchange-of-views:red-chair"
	if got, _ := post(chair, "feov-record show work"); len(asked) != 2 || asked[1] != runDir+" | a7cc47571ab8b015d | "+chair || got != text {
		t.Errorf("a chair's call asked %q and was handed %q — want the chair's own next item", asked, got)
	}
	// ANY TOOL ADVANCES THE DELIVERY: a Bash call that is not the record's, a Read, a WebSearch.
	for i, c := range []struct {
		tool  string
		input map[string]string
	}{
		{"Bash", map[string]string{"command": "ls -la"}},
		{"Read", map[string]string{"file_path": cwd + "/notes.md"}},
		{"WebSearch", map[string]string{"query": "is 91 prime"}},
	} {
		if got, _ := call("a7cc47571ab8b015d", lens, c.tool, c.input); len(asked) != 3+i || asked[2+i] != runDir+" | a7cc47571ab8b015d | "+lens || got != text {
			t.Errorf("a lens's %s call asked %q and was handed %q — want its next item, as a call to the record is", c.tool, asked, got)
		}
	}
	// THE ENVELOPE'S CALL HANDS NOTHING: it ends the seat's work. And a call from no seat spawns nothing.
	if _, out := call("a7cc47571ab8b015d", lens, envelopeTool, map[string]string{}); len(asked) != 5 || out.Len() != 0 {
		t.Errorf("the envelope's call spawned the writer or spoke: %q %q", asked, out.String())
	}
	if _, out := call("", "", "Read", map[string]string{"file_path": cwd + "/notes.md"}); len(asked) != 5 || out.Len() != 0 {
		t.Errorf("a call from no seat spawned the writer or spoke: %q %q", asked, out.String())
	}
	text = ""
	if _, out := post(lens, "feov-record show report"); len(asked) != 6 || out.Len() != 0 {
		t.Errorf("nothing to hand was said anyway: %q", out.String())
	}
	writeErr = errors.New("the record is locked")
	if got, _ := post(lens, "feov-record show report"); !strings.Contains(got, "COULD NOT BE READ") || !strings.Contains(got, "the record is locked") {
		t.Errorf("a writer that failed was handed to the lens as %q — want the statement that nothing was determined, and why", got)
	}
}

// A DELIVERY ANOTHER CALL HOLDS HANDS NOTHING AND IS NO FAILURE: that call is handing the item. The
// lens is told nothing — not the unreadable statement — and the hook record holds no failure.
func TestAHeldDeliveryHandsNothingAndRecordsNoFailure(t *testing.T) {
	cwd, _ := hookRun(t)
	prev := handNextItem
	handNextItem = func(string, string, string, string) (string, error) {
		return "", fmt.Errorf("feov-sitting-write: %w: (x.lock) past 5s", hookgate.ErrDeliveryHeld)
	}
	t.Cleanup(func() { handNextItem = prev })
	b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": "feov-record show report"},
		"cwd": cwd, "session_id": testLead, "agent_id": "a7cc47571ab8b015d", "agent_type": "frank-exchange-of-views:red-lens-logic",
		"tool_response": map[string]string{}})
	var out bytes.Buffer
	rec := testRecorder()
	if err := Post(bytes.NewReader(b), &out, rec); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Errorf("a held delivery said something to the lens: %s", out.String())
	}
	if rec.Failed() {
		t.Error("a held delivery was recorded as the hook's failure")
	}
}

// A STUCK DELIVERY ENDS THE SEAT'S TURN, AND ONLY A SEAT'S: a subagent's or Workflow seat's call is
// answered "continue": false with the writer's reason, and the hook record holds the failure; a call
// from the lead session itself is never stopped, and is told the delivery could not be read.
func TestAStuckDeliveryEndsOnlyTheSeatsTurn(t *testing.T) {
	cwd, _ := hookRun(t)
	prev := handNextItem
	const reason = "item 1 of red-lens-logic's delivery renders to 9100 bytes, over the item budget of 9000"
	handNextItem = func(string, string, string, string) (string, error) {
		return "", fmt.Errorf("feov-sitting-write: %w: %s", hookgate.ErrDeliveryStuck, reason)
	}
	t.Cleanup(func() { handNextItem = prev })
	post := func(agentID string) (map[string]any, *hookfailures.Recorder) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]string{"command": "feov-record show work"},
			"cwd": cwd, "session_id": testLead, "agent_id": agentID, "agent_type": "frank-exchange-of-views:red-lens-logic",
			"tool_response": map[string]string{}})
		var out bytes.Buffer
		rec := testRecorder()
		if err := Post(bytes.NewReader(b), &out, rec); err != nil {
			t.Fatal(err)
		}
		doc := map[string]any{}
		if out.Len() > 0 {
			if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
				t.Fatalf("the hook printed %q: %v", out.String(), err)
			}
		}
		return doc, rec
	}
	doc, rec := post("a7cc47571ab8b015d")
	if doc["continue"] != false || !strings.Contains(fmt.Sprint(doc["stopReason"]), reason) {
		t.Errorf("a seat's stuck delivery answered %v — want continue false and the writer's reason as stopReason", doc)
	}
	if !rec.Failed() {
		t.Error("a stuck delivery was not recorded as the hook's failure")
	}
	doc, _ = post(testLead)
	if _, stopped := doc["continue"]; stopped {
		t.Errorf("the lead session's own call was stopped: %v", doc)
	}
}
