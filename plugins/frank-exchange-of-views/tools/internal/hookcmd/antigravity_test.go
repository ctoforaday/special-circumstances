package hookcmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

func TestAntigravityFinishEnvelopeUnwrapsAndRewrapsResult(t *testing.T) {
	cwd, _ := hookRun(t)
	var seenInput string
	prev := envelopeAnswer
	const read = `{"event":7,"bench_owed":[],"halted":false}`
	envelopeAnswer = func(_, _, _ string, input []byte) (hookgate.EnvelopeAnswer, error) {
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
	envelopeAnswer = func(_, _, _ string, _ []byte) (hookgate.EnvelopeAnswer, error) {
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
	handNextItem = func(_, _, _, _ string) (string, error) {
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

	// PreInvocation drains a failed run_command step from transcript_full.jsonl and logs it via
	// failureEntry/recordFailure, skipping feov-record refusals already logged as refusals.
	t.Setenv("PLUGIN_DATA", t.TempDir())
	tp := filepath.Join(t.TempDir(), "transcript_full.jsonl")
	trData := strings.Join([]string{
		`{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"\"/bin/feov-record cite --help\""}}]}`,
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 2.\n\nCommand output:\nfeov-record: \"cite\" is not on your surface\n"}`,
		`{"step_index":3,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"\"/missing/script.py\""}}]}`,
		`{"step_index":4,"source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 127.\n\nCommand output:\nbash: /missing/script.py: No such file or directory\n"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(tp, []byte(trData), 0o600); err != nil {
		t.Fatal(err)
	}
	var logged []string
	prevFail := recordFailure
	recordFailure = func(_, agentID, text string) error {
		logged = append(logged, agentID+"|"+text)
		return nil
	}
	t.Cleanup(func() { recordFailure = prevFail })

	preInvPayload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreInvocation",
		"conversationId": childConv,
		"cwd":            cwd,
		"transcriptPath": tp,
		"invocationNum":  1,
	})
	out.Reset()
	if err := Pre(bytes.NewReader(preInvPayload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if len(logged) != 1 || !strings.HasPrefix(logged[0], "agy-0123456789abcdef|") || !strings.Contains(logged[0], "/missing/script.py") {
		t.Fatalf("PreInvocation failed-step logs = %v, want 1 non-refusal entry for /missing/script.py", logged)
	}

	// Bound agy seat PostToolUse carrying a tool-level error (e.g. view_file error) routes through
	// PostFailure with the bound agent id and redacts runDir to ".".
	logged = nil
	postErrPayload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PostToolUse",
		"conversationId": childConv,
		"cwd":            cwd,
		"error":          "open " + filepath.Join(runDir, "missing.md") + ": no such file or directory",
		"toolCall": map[string]any{
			"id":   "call-read-err",
			"name": "read_file",
			"args": map[string]any{"Path": filepath.Join(runDir, "missing.md")},
		},
	})
	out.Reset()
	if err := Post(bytes.NewReader(postErrPayload), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if len(logged) != 1 || !strings.HasPrefix(logged[0], "agy-0123456789abcdef|") {
		t.Fatalf("PostToolUse tool error logs = %v, want 1 entry for agy-0123456789abcdef", logged)
	}
	if strings.Contains(logged[0], runDir) {
		t.Fatalf("PostToolUse failure log leaked runDir %q: %s", runDir, logged[0])
	}

	// Second conversation trying to claim the already-bound seat-red-chair is refused (O_EXCL)
	// and records a StageAgyBind failure on the recorder.
	hijackPayload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": "other-conv-hijack",
		"cwd":            cwd,
		"toolCall": map[string]any{
			"id":   "call-hijack",
			"name": "run_command",
			"args": map[string]any{"CommandLine": "/bin/feov-record register --seat-id red-chair"},
		},
	})
	out.Reset()
	hijackRec := testRecorder()
	if err := Pre(bytes.NewReader(hijackPayload), &out, hijackRec); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "FEOV_AGENT_ID='agy-0123456789abcdef'") {
		t.Fatalf("second conversation hijacked already-bound seat: %s", out.String())
	}
	if !hijackRec.Failed() {
		t.Fatalf("expected O_EXCL conflict to record a failure on recorder")
	}
}

func TestAntigravitySittingRefusesDecidedRun(t *testing.T) {
	cwd, runDir := hookRun(t)
	sdir := filepath.Join(runDir, agySittingsDir)
	if err := os.MkdirAll(sdir, 0o700); err != nil {
		t.Fatal(err)
	}
	seatRec, _ := json.Marshal(agySitting{
		AgentID:     "agy-decided00000001",
		AgentType:   "frank-exchange-of-views:red-lens-logic",
		SeatID:      "red-lens-logic",
		RunDir:      runDir,
		LeadSession: testLead,
	})
	if err := os.WriteFile(filepath.Join(sdir, "seat-red-lens-logic.json"), seatRec, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sdir, "conv-already-bound.json"), seatRec, 0o600); err != nil {
		t.Fatal(err)
	}
	// Mark the run decided so runlive.StateOf returns RowDecided for red-lens-logic.
	if err := os.WriteFile(filepath.Join(runDir, runlive.OutcomeMark), []byte(`{"agent_id":"other-agent"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// Direct resolver check: both fast-path (conv-*.json) and first-bind (--seat-id) must return
	// false on a decided run (fails immediately if either StateOf check in resolveAgySitting is reverted).
	if s, ok := resolveAgySitting(hookcore.Request{Engine: hookcore.EngineAntigravity, SessionID: "already-bound"}, cwd, testRecorder()); ok {
		t.Fatalf("resolveAgySitting fast-path resolved a decided run: %+v", s)
	}
	if s, ok := resolveAgySitting(hookcore.Request{Engine: hookcore.EngineAntigravity, SessionID: "fresh-conv", ToolName: "Bash", Command: "feov-record register --seat-id red-lens-logic"}, cwd, testRecorder()); ok {
		t.Fatalf("resolveAgySitting first-bind resolved a decided run: %+v", s)
	}

	// Both fast-path (conv-*.json) and first-bind (--seat-id) must refuse to bind a decided run,
	// without creating .bound / conv-*.json and without PostToolUse logging under the decided agent.
	var logged []string
	prevFail := recordFailure
	recordFailure = func(_, agentID, text string) error {
		logged = append(logged, agentID+"|"+text)
		return nil
	}
	t.Cleanup(func() { recordFailure = prevFail })

	for _, tc := range []struct {
		name string
		conv string
		cmd  string
	}{
		{"fast-path", "already-bound", "echo hi"},
		{"first-bind", "fresh-conv", "/bin/feov-record register --seat-id red-lens-logic"},
	} {
		payload, _ := json.Marshal(map[string]any{
			"hookEventName":  "PreToolUse",
			"conversationId": tc.conv,
			"cwd":            cwd,
			"toolCall": map[string]any{
				"id":   "call-" + tc.name,
				"name": "run_command",
				"args": map[string]any{"CommandLine": tc.cmd},
			},
		})
		var out bytes.Buffer
		if err := Pre(bytes.NewReader(payload), &out, testRecorder()); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if strings.Contains(out.String(), "FEOV_AGENT_ID='agy-decided00000001'") {
			t.Fatalf("%s bound a decided run: %s", tc.name, out.String())
		}
	}
	if _, err := os.Stat(filepath.Join(sdir, "seat-red-lens-logic.bound")); err == nil {
		t.Fatalf("first-bind on a decided run created seat-red-lens-logic.bound")
	}
	if _, err := os.Stat(filepath.Join(sdir, "conv-fresh-conv.json")); err == nil {
		t.Fatalf("first-bind on a decided run created conv-fresh-conv.json")
	}
	postErr, _ := json.Marshal(map[string]any{
		"hookEventName":  "PostToolUse",
		"conversationId": "already-bound",
		"cwd":            cwd,
		"error":          "boom",
		"toolCall": map[string]any{
			"id":   "call-decided-post",
			"name": "read_file",
			"args": map[string]any{"Path": "x.md"},
		},
	})
	var postOut bytes.Buffer
	if err := Post(bytes.NewReader(postErr), &postOut, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if len(logged) != 0 {
		t.Fatalf("PostToolUse on decided run logged under decided agent: %v", logged)
	}
}

func TestAntigravityPreInvocationDrainErrorScopedAndCleared(t *testing.T) {
	t.Setenv("PLUGIN_DATA", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, runDir := hookRun(t)
	sdir := filepath.Join(runDir, agySittingsDir)
	if err := os.MkdirAll(sdir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeConv := func(conv, agentID string) {
		b, _ := json.Marshal(agySitting{
			AgentID:     agentID,
			AgentType:   "frank-exchange-of-views:red-lens-logic",
			SeatID:      "red-lens-logic",
			RunDir:      runDir,
			LeadSession: testLead,
		})
		if err := os.WriteFile(filepath.Join(sdir, "conv-"+conv+".json"), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeConv("conv-A", "agy-drainerrA00001")
	writeConv("conv-B", "agy-drainerrB00001")

	tpA := filepath.Join(t.TempDir(), "a.jsonl")
	trBad := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"false"}}]}` + "\n" +
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"DONE","content":"unrecognized bash exit format\n"}` + "\n"
	if err := os.WriteFile(tpA, []byte(trBad), 0o600); err != nil {
		t.Fatal(err)
	}
	tpB := filepath.Join(t.TempDir(), "b.jsonl")
	if err := os.WriteFile(tpB, []byte(`{"step_index":1,"source":"MODEL","type":"GENERIC","status":"DONE","content":"ok\n"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	readFailures := func() string {
		p, err := hookfailures.Path("frank-exchange-of-views")
		if err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(p)
		return string(b)
	}
	firePreInv := func(conv, tp string) {
		payload, _ := json.Marshal(map[string]any{
			"hookEventName":  "PreInvocation",
			"conversationId": conv,
			"cwd":            cwd,
			"transcriptPath": tp,
			"invocationNum":  1,
		})
		rec := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "PreInvocation", time.Now(), &bytes.Buffer{})
		var out bytes.Buffer
		if err := Pre(bytes.NewReader(payload), &out, rec); err != nil {
			t.Fatal(err)
		}
		rec.Persist()
	}

	// 1. conv-A fails drain -> recorded under hookcore.StageFailedStepDrain() with TTL scoped to its agent id.
	firePreInv("conv-A", tpA)
	if f := readFailures(); !strings.Contains(f, string(hookcore.StageFailedStepDrain().Stage)) || !strings.Contains(f, "agy-drainerrA00001") || !strings.Contains(f, `"expires_after"`) {
		t.Fatalf("failures.json missing TTL-scoped StageFailedStepDrain for agy-drainerrA00001: %s", f)
	}
	// 2. Peer conv-B clean drain must NOT clear conv-A's entry.
	firePreInv("conv-B", tpB)
	if f := readFailures(); !strings.Contains(f, "agy-drainerrA00001") {
		t.Fatalf("peer conv-B clean drain cleared agy-drainerrA00001's failure entry: %s", f)
	}
	// 3. Displaying event shows conv-A's entry (stamps Notified), then conv-A's clean drain clears it.
	showRec := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "Stop", time.Now(), &bytes.Buffer{})
	if msg := showRec.Settle(); !strings.Contains(msg, "agy-drainerrA00001") {
		t.Fatalf("Stop Settle did not show conv-A drain error: %q", msg)
	}
	firePreInv("conv-A", tpA)
	if f := readFailures(); strings.Contains(f, "agy-drainerrA00001") {
		t.Fatalf("conv-A clean drain did not clear its StageFailedStepDrain entry after being shown: %s", f)
	}

	// 4. Hook-unmarshal ERROR step records hookcore.StageHookUnmarshal() with TTL, and clears on next clean drain after shown.
	prevFail := recordFailure
	recordFailure = func(_, _, _ string) error { return nil }
	t.Cleanup(func() { recordFailure = prevFail })
	writeConv("conv-H", "agy-hookerrH00001")
	tpHook := filepath.Join(t.TempDir(), "hookerr.jsonl")
	trHook := `{"step_index":1,"source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","tool_calls":[{"name":"run_command","args":{"CommandLine":"echo hi"}}]}` + "\n" +
		`{"step_index":2,"source":"MODEL","type":"GENERIC","status":"ERROR","error":"unmarshal","content":"Encountered error in step execution: failed to unmarshal result from hook jsonhook__feov\n"}` + "\n"
	if err := os.WriteFile(tpHook, []byte(trHook), 0o600); err != nil {
		t.Fatal(err)
	}
	firePreInv("conv-H", tpHook)
	if f := readFailures(); !strings.Contains(f, string(hookcore.StageHookUnmarshal().Stage)) || !strings.Contains(f, `"expires_after"`) {
		t.Fatalf("failures.json missing TTL StageHookUnmarshal: %s", f)
	}
	showRec2 := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "Stop", time.Now(), &bytes.Buffer{})
	_ = showRec2.Settle()
	firePreInv("conv-H", tpHook)
	if f := readFailures(); strings.Contains(f, string(hookcore.StageHookUnmarshal().Stage)) {
		t.Fatalf("clean drain did not clear StageHookUnmarshal after being shown: %s", f)
	}

	// 5. F6: Run() on a bound agy seat queues Settle's Say under the bound AgentID so the next
	// PreInvocation flushes it (instead of stranding it under the raw conversationId).
	seedRec := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "PreInvocation", time.Now(), &bytes.Buffer{})
	seedRec.FailTTL(hookcore.StageFailedStepDrain(), "agy-drainerrA00001", "F6_BOUND_DELIVERY_CANARY")
	seedRec.Persist()
	preToolPayload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": "conv-A",
		"cwd":            cwd,
		"toolCall": map[string]any{
			"id":   "call-f6",
			"name": "run_command",
			"args": map[string]any{"CommandLine": "echo ok"},
		},
	})
	var runOut bytes.Buffer
	Run("feov-pretooluse", "PreToolUse", Pre, bytes.NewReader(preToolPayload), &runOut)
	preInvF6, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreInvocation",
		"conversationId": "conv-A",
		"cwd":            cwd,
		"invocationNum":  1,
	})
	var preInvOut bytes.Buffer
	if err := Pre(bytes.NewReader(preInvF6), &preInvOut, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preInvOut.String(), "F6_BOUND_DELIVERY_CANARY") {
		t.Fatalf("bound agy seat PreInvocation did not flush Run() Settle message (stranded under conversationId?): %s", preInvOut.String())
	}

	// 6. StageAgyBind conflict is recorded and then cleared on the next successful bind.
	seatB, _ := json.Marshal(agySitting{
		AgentID: "agy-bindclear0001", AgentType: "frank-exchange-of-views:red-chair",
		SeatID: "red-chair", RunDir: runDir, LeadSession: testLead,
	})
	if err := os.WriteFile(filepath.Join(sdir, "seat-red-chair.json"), seatB, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sdir, "seat-red-chair.bound"), []byte("other\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rec1 := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "PreToolUse", time.Now(), &bytes.Buffer{})
	resolveAgySitting(hookcore.Request{Engine: hookcore.EngineAntigravity, SessionID: "c-conflict", ToolName: "Bash", Command: "feov-record register --seat-id red-chair"}, cwd, rec1)
	rec1.Persist()
	if f := readFailures(); !strings.Contains(f, string(StageAgyBind)) {
		t.Fatalf("failures.json missing StageAgyBind after conflict: %s", f)
	}
	_ = os.Remove(filepath.Join(sdir, "seat-red-chair.bound"))
	rec2 := hookfailures.New("frank-exchange-of-views", "feov-pretooluse", "PreToolUse", time.Now(), &bytes.Buffer{})
	if _, ok := resolveAgySitting(hookcore.Request{Engine: hookcore.EngineAntigravity, SessionID: "c-ok", ToolName: "Bash", Command: "feov-record register --seat-id red-chair"}, cwd, rec2); !ok {
		t.Fatal("expected clean bind to succeed")
	}
	rec2.Persist()
	if f := readFailures(); strings.Contains(f, string(StageAgyBind)) {
		t.Fatalf("successful bind did not clear StageAgyBind: %s", f)
	}
}

// G1: A panic in resolveAgySitting during Run's post-entry sitting lookup MUST be caught by Run's
// recover(), exit 0, and record StagePanic (never escape Run with exit 2).
func TestAntigravityRunRecoversPanicInSittingLookup(t *testing.T) {
	t.Setenv("PLUGIN_DATA", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	prev := resolveAgySittingFn
	resolveAgySittingFn = func(hookcore.Request, string, *hookfailures.Recorder) (agySitting, bool) {
		panic("injected sitting lookup panic")
	}
	t.Cleanup(func() { resolveAgySittingFn = prev })

	payload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": "conv-panic-test",
		"cwd":            t.TempDir(),
		"toolCall": map[string]any{
			"id":   "call-panic",
			"name": "run_command",
			"args": map[string]any{"CommandLine": "echo hi"},
		},
	})
	var out bytes.Buffer
	code := Run("feov-pretooluse", "PreToolUse", func(io.Reader, io.Writer, *hookfailures.Recorder) error {
		return nil
	}, bytes.NewReader(payload), &out)
	if code != 0 {
		t.Fatalf("Run returned exit %d on resolveAgySitting panic, want 0", code)
	}
	p, err := hookfailures.Path("frank-exchange-of-views")
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); !strings.Contains(string(b), string(StagePanic)) {
		t.Fatalf("failures.json missing StagePanic after resolveAgySitting panic: %s", string(b))
	}
}

// H1 + J1: When f panics on an already-bound Antigravity seat, Run must still queue the hook-panic
// message under pending_<AgentID>.jsonl (NOT pending_<conversationId>.jsonl) so the bound seat's
// next PreInvocation delivers it.
func TestAntigravityRunBoundSeatEntryPanicDeliveredOnPreInvocation(t *testing.T) {
	pluginData := t.TempDir()
	t.Setenv("PLUGIN_DATA", pluginData)
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, runDir := hookRun(t)
	sdir := filepath.Join(runDir, agySittingsDir)
	if err := os.MkdirAll(sdir, 0o700); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(agySitting{
		AgentID:     "agy-panicbound0001",
		AgentType:   "frank-exchange-of-views:red-lens-logic",
		SeatID:      "red-lens-logic",
		RunDir:      runDir,
		LeadSession: testLead,
	})
	// Lookup reads "conv-" + SessionID + ".json", so conversationId "conv-P" binds via conv-conv-P.json.
	if err := os.WriteFile(filepath.Join(sdir, "conv-conv-P.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": "conv-P",
		"cwd":            cwd,
		"toolCall": map[string]any{
			"id":   "call-p",
			"name": "run_command",
			"args": map[string]any{"CommandLine": "echo hi"},
		},
	})
	var runOut bytes.Buffer
	code := Run("feov-pretooluse", "PreToolUse", func(io.Reader, io.Writer, *hookfailures.Recorder) error {
		panic("simulated entry panic on bound seat")
	}, bytes.NewReader(payload), &runOut)
	if code != 0 {
		t.Fatalf("Run returned exit %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(pluginData, "pending_agy-panicbound0001.jsonl")); err != nil {
		t.Fatalf("expected hook-panic queued in pending_agy-panicbound0001.jsonl: %v", err)
	}
	if matches, _ := filepath.Glob(filepath.Join(pluginData, "pending_conv-*.jsonl")); len(matches) != 0 {
		t.Fatalf("hook-panic stranded in unbound pending_conv-* queue: %v", matches)
	}

	preInvPayload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreInvocation",
		"conversationId": "conv-P",
		"cwd":            cwd,
		"invocationNum":  1,
	})
	var preInvOut bytes.Buffer
	if err := Pre(bytes.NewReader(preInvPayload), &preInvOut, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preInvOut.String(), "hook-panic") || !strings.Contains(preInvOut.String(), "simulated entry panic on bound seat") {
		t.Fatalf("bound seat PreInvocation did not deliver hook-panic after f panicked: %s", preInvOut.String())
	}
}

// J2 (probe I8): When a first-bind PreToolUse (--seat-id) creates the binding inside f and then f
// panics, Run's defer still resolves the newly-written conv-<id>.json under its own recover and
// queues StagePanic to pending_<AgentID>.jsonl.
func TestAntigravityRunFirstBindThenPanicDeliveredOnPreInvocation(t *testing.T) {
	pluginData := t.TempDir()
	t.Setenv("PLUGIN_DATA", pluginData)
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, runDir := hookRun(t)
	sdir := filepath.Join(runDir, agySittingsDir)
	if err := os.MkdirAll(sdir, 0o700); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(agySitting{
		AgentID:     "agy-firstbindpanic01",
		AgentType:   "frank-exchange-of-views:red-lens-logic",
		SeatID:      "red-lens-logic",
		RunDir:      runDir,
		LeadSession: testLead,
	})
	if err := os.WriteFile(filepath.Join(sdir, "seat-red-lens-logic.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreToolUse",
		"conversationId": "conv-FBP",
		"cwd":            cwd,
		"toolCall": map[string]any{
			"id":   "call-fbp",
			"name": "run_command",
			"args": map[string]any{"CommandLine": "/bin/feov-record register --seat-id red-lens-logic"},
		},
	})
	var runOut bytes.Buffer
	code := Run("feov-pretooluse", "PreToolUse", func(in io.Reader, out io.Writer, rec *hookfailures.Recorder) error {
		_ = Pre(in, out, rec)
		panic("panic right after first bind")
	}, bytes.NewReader(payload), &runOut)
	if code != 0 {
		t.Fatalf("Run returned exit %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(pluginData, "pending_agy-firstbindpanic01.jsonl")); err != nil {
		t.Fatalf("expected hook-panic queued in pending_agy-firstbindpanic01.jsonl after first-bind panic: %v", err)
	}
	if matches, _ := filepath.Glob(filepath.Join(pluginData, "pending_conv-*.jsonl")); len(matches) != 0 {
		t.Fatalf("hook-panic stranded in unbound pending_conv-* queue: %v", matches)
	}

	preInvPayload, _ := json.Marshal(map[string]any{
		"hookEventName":  "PreInvocation",
		"conversationId": "conv-FBP",
		"cwd":            cwd,
		"invocationNum":  1,
	})
	var preInvOut bytes.Buffer
	if err := Pre(bytes.NewReader(preInvPayload), &preInvOut, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preInvOut.String(), "hook-panic") || !strings.Contains(preInvOut.String(), "panic right after first bind") {
		t.Fatalf("PreInvocation did not deliver hook-panic after first-bind panic: %s", preInvOut.String())
	}
}

// J3: applyBoundAgySitting blanks ToolName/Command so it NEVER claims O_EXCL seat-<id>.bound
// (which would otherwise block Pre's real bind from clearing a prior StageAgyBind conflict).
func TestApplyBoundAgySittingNeverCreatesBoundMarker(t *testing.T) {
	cwd, runDir := hookRun(t)
	sdir := filepath.Join(runDir, agySittingsDir)
	if err := os.MkdirAll(sdir, 0o700); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(agySitting{
		AgentID:     "agy-readonlybind001",
		AgentType:   "frank-exchange-of-views:red-chair",
		SeatID:      "red-chair",
		RunDir:      runDir,
		LeadSession: testLead,
	})
	if err := os.WriteFile(filepath.Join(sdir, "seat-red-chair.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	req := hookcore.Request{
		Engine:    hookcore.EngineAntigravity,
		SessionID: "conv-fresh-seat",
		ToolName:  "Bash",
		Command:   "/bin/feov-record register --seat-id red-chair",
	}
	if applyBoundAgySitting(&req, cwd) {
		t.Fatalf("applyBoundAgySitting unexpectedly bound an unbound --seat-id request: %+v", req)
	}
	if _, err := os.Stat(filepath.Join(sdir, "seat-red-chair.bound")); err == nil {
		t.Fatalf("applyBoundAgySitting created seat-red-chair.bound despite read-only contract")
	}
}
