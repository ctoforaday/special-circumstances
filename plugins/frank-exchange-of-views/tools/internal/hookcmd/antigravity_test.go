package hookcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
)

func TestAntigravityFinishEnvelopeUnwrapsAndRewrapsResult(t *testing.T) {
	cwd, _ := hookRun(t)
	var seenInput string
	prev := envelopeAnswer
	const read = `{"event":7,"bench_owed":[],"halted":false}`
	envelopeAnswer = func(_, _ string, input []byte) (hookgate.EnvelopeAnswer, error) {
		seenInput = string(input)
		return hookgate.EnvelopeAnswer{Read: json.RawMessage(read)}, nil
	}
	t.Cleanup(func() { envelopeAnswer = prev })

	payload, _ := json.Marshal(map[string]any{
		"hookEventName":        "PreToolUse",
		"conversationId":       "conv-seat-1",
		"parentConversationId": testLead,
		"agentName":            "red-lens-logic",
		"cwd":                  cwd,
		"toolCall": map[string]any{
			"id":   "call-finish-1",
			"name": "finish",
			"args": map[string]any{
				"result": map[string]any{
					"verdict":     "PASS",
					"record_read": map[string]any{"spoofed": true},
				},
			},
		},
	})
	var out bytes.Buffer
	if err := Pre(bytes.NewReader(payload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	var unwrapped map[string]any
	if err := json.Unmarshal([]byte(seenInput), &unwrapped); err != nil {
		t.Fatalf("envelopeAnswer input was not valid JSON: %v (%s)", err, seenInput)
	}
	if _, stillWrapped := unwrapped["result"]; stillWrapped || unwrapped["verdict"] != "PASS" {
		t.Fatalf("envelopeAnswer received wrapped or wrong input: %s", seenInput)
	}

	var resp struct {
		Decision  string `json:"decision"`
		Reason    string `json:"reason"`
		Overwrite struct {
			Result map[string]json.RawMessage `json:"result"`
		} `json:"overwrite"`
	}
	dec := json.NewDecoder(&out)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("PreToolUse finish response failed strict decode: %v\n%s", err, out.String())
	}
	if resp.Decision != "allow" || string(resp.Overwrite.Result[hookgate.EnvelopeReadField]) != read {
		t.Fatalf("unexpected finish overwrite: %+v (read=%s)", resp, string(resp.Overwrite.Result[hookgate.EnvelopeReadField]))
	}
}

func TestAntigravityFinishEnvelopeDeny(t *testing.T) {
	cwd, _ := hookRun(t)
	prev := envelopeAnswer
	envelopeAnswer = func(_, _ string, _ []byte) (hookgate.EnvelopeAnswer, error) {
		return hookgate.EnvelopeAnswer{Refusal: "open blocker stands"}, nil
	}
	t.Cleanup(func() { envelopeAnswer = prev })

	payload, _ := json.Marshal(map[string]any{
		"hookEventName":        "PreToolUse",
		"conversationId":       "conv-seat-2",
		"parentConversationId": testLead,
		"agentName":            "red-lens-logic",
		"cwd":                  cwd,
		"toolCall": map[string]any{
			"id":   "call-finish-2",
			"name": "finish",
			"args": map[string]any{"result": map[string]any{"verdict": "PASS"}},
		},
	})
	var out bytes.Buffer
	if err := Pre(bytes.NewReader(payload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	var resp struct {
		Decision   string `json:"decision"`
		Reason     string `json:"reason"`
		DenyReason string `json:"deny_reason"`
	}
	dec := json.NewDecoder(&out)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("PreToolUse finish deny failed strict decode: %v\n%s", err, out.String())
	}
	if resp.Decision != "deny" || resp.DenyReason != "open blocker stands" {
		t.Fatalf("unexpected finish deny: %+v", resp)
	}
}

func TestAntigravityPostFinishAdmitsOrLogsFailure(t *testing.T) {
	cwd, _ := hookRun(t)
	var admitted, failed int
	prevAdmit := recordAdmission
	prevFail := recordFailure
	recordAdmission = func(_, _ string) error { admitted++; return nil }
	recordFailure = func(_, _, _ string) error { failed++; return nil }
	t.Cleanup(func() {
		recordAdmission = prevAdmit
		recordFailure = prevFail
	})

	okPayload, _ := json.Marshal(map[string]any{
		"hookEventName":        "PostToolUse",
		"conversationId":       "conv-seat-3",
		"parentConversationId": testLead,
		"agentName":            "red-lens-logic",
		"cwd":                  cwd,
		"toolCall": map[string]any{
			"id":   "call-finish-3",
			"name": "finish",
			"args": map[string]any{"result": map[string]any{"verdict": "PASS"}},
		},
	})
	var out bytes.Buffer
	if err := Post(bytes.NewReader(okPayload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if admitted != 1 || failed != 0 {
		t.Fatalf("succeeded finish: admitted=%d failed=%d, want 1, 0", admitted, failed)
	}

	errPayload, _ := json.Marshal(map[string]any{
		"hookEventName":        "PostToolUse",
		"conversationId":       "conv-seat-3",
		"parentConversationId": testLead,
		"agentName":            "red-lens-logic",
		"cwd":                  cwd,
		"error":                "schema validation failed",
		"toolCall": map[string]any{
			"id":   "call-finish-4",
			"name": "finish",
			"args": map[string]any{"result": map[string]any{}},
		},
	})
	out.Reset()
	if err := Post(bytes.NewReader(errPayload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if admitted != 1 || failed != 1 {
		t.Fatalf("failed finish: admitted=%d failed=%d, want 1, 1", admitted, failed)
	}
}

func TestAntigravityPostRunCommandQueuesNextItemForPreInvocation(t *testing.T) {
	t.Setenv("PLUGIN_DATA", t.TempDir())
	cwd, _ := hookRun(t)
	prevNext := handNextItem
	handNextItem = func(_, _, _ string) (string, error) {
		return "NEXT ITEM: gap-42", nil
	}
	t.Cleanup(func() { handNextItem = prevNext })

	postPayload, _ := json.Marshal(map[string]any{
		"hookEventName":        "PostToolUse",
		"conversationId":       "conv-seat-4",
		"parentConversationId": testLead,
		"agentName":            "red-lens-logic",
		"cwd":                  cwd,
		"toolCall": map[string]any{
			"id":   "call-cmd-1",
			"name": "run_command",
			"args": map[string]any{"CommandLine": "feov-record red gap close --id gap-42"},
		},
	})
	var postOut bytes.Buffer
	if err := Post(bytes.NewReader(postPayload), &postOut, testRecorder()); err != nil {
		t.Fatal(err)
	}

	preInvPayload, _ := json.Marshal(map[string]any{
		"hookEventName":        "PreInvocation",
		"conversationId":       "conv-seat-4",
		"parentConversationId": testLead,
		"agentName":            "red-lens-logic",
		"cwd":                  cwd,
		"invocationNum":        1,
	})
	var preOut bytes.Buffer
	if err := Pre(bytes.NewReader(preInvPayload), &preOut, testRecorder()); err != nil {
		t.Fatal(err)
	}
	var res struct {
		InjectSteps []struct {
			EphemeralMessage string `json:"ephemeralMessage"`
		} `json:"injectSteps"`
	}
	dec := json.NewDecoder(&preOut)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&res); err != nil {
		t.Fatalf("PreInvocation response failed strict decode: %v\n%s", err, preOut.String())
	}
	if len(res.InjectSteps) != 1 || res.InjectSteps[0].EphemeralMessage != "NEXT ITEM: gap-42" {
		t.Fatalf("expected queued next item on PreInvocation, got %+v", res)
	}
}

func TestAntigravityRunWorkflowSittingBindingInjectsRunAndAgentIdentity(t *testing.T) {
	cwd, runDir := hookRun(t)
	sdir := filepath.Join(runDir, agySittingsDir)
	if err := os.MkdirAll(sdir, 0o700); err != nil {
		t.Fatal(err)
	}
	seatRec, _ := json.Marshal(agySitting{
		AgentID:     "agy-0123456789abcdef",
		AgentType:   "frank-exchange-of-views:red-chair",
		SeatID:      "red-chair",
		RunDir:      runDir,
		LeadSession: testLead,
	})
	if err := os.WriteFile(filepath.Join(sdir, "seat-red-chair.json"), seatRec, 0o600); err != nil {
		t.Fatal(err)
	}

	// First call from run_workflow child conversation: NO parentConversationId and NO agentName.
	// Carrying --seat-id red-chair binds conv-<conversationId>.json and injects FEOV_RUN/FEOV_AGENT_ID/FEOV_AGENT_TYPE.
	const childConv = "child-wf-conv-1"
	const childTranscript = "/tmp/brain/child-wf-conv-1/transcript_full.jsonl"
	payload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": childConv,
		"cwd":            cwd,
		"transcriptPath": childTranscript,
		"toolCall": map[string]any{
			"id":   "call-1",
			"name": "run_command",
			"args": map[string]any{
				"CommandLine": "/bin/feov-record register --run " + runDir + " --seat-id red-chair",
			},
		},
	})
	var out bytes.Buffer
	if err := Pre(bytes.NewReader(payload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"FEOV_RUN=", "FEOV_AGENT_ID='agy-0123456789abcdef'", "FEOV_AGENT_TYPE='frank-exchange-of-views:red-chair'"} {
		if !strings.Contains(got, want) {
			t.Fatalf("first run_command PreToolUse missing %q:\n%s", want, got)
		}
	}
	if tr, err := os.ReadFile(filepath.Join(sdir, "seat-red-chair.transcript")); err != nil || string(tr) != childTranscript {
		t.Fatalf("transcript file = %q err=%v, want %q", string(tr), err, childTranscript)
	}

	// Subsequent call in the same conversation WITHOUT --seat-id still resolves via conv-<id>.json.
	payload2, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": childConv,
		"cwd":            cwd,
		"toolCall": map[string]any{
			"id":   "call-2",
			"name": "run_command",
			"args": map[string]any{
				"CommandLine": "echo hello",
			},
		},
	})
	out.Reset()
	if err := Pre(bytes.NewReader(payload2), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "FEOV_AGENT_ID='agy-0123456789abcdef'") {
		t.Fatalf("subsequent call did not resolve via bound conv file:\n%s", out.String())
	}
}
