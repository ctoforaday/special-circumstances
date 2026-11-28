package hookcore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures are agy 1.5.0's own payloads, captured by testdata/eventlog and scrubbed of the home
// directory: what the client actually sends, not what its documentation says it sends.
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "agy-1.5.0", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestReadAgySessionStartAndPreInvocationPayloads(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	for _, c := range []struct{ event, file string }{
		{"SessionStart", "session_start.json"},
		{"PreInvocation", "pre_invocation.json"},
	} {
		raw := readFixture(t, c.file)
		var p struct {
			ConversationID string   `json:"conversationId"`
			TranscriptPath string   `json:"transcriptPath"`
			WorkspacePaths []string `json:"workspacePaths"`
		}
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatal(err)
		}
		req := Read(c.event, bytes.NewReader(raw))
		if req.Engine != EngineAntigravity {
			t.Errorf("%s: engine %q, want antigravity", c.event, req.Engine)
		}
		if req.SessionID != p.ConversationID || req.SessionID == "" {
			t.Errorf("%s: SessionID %q, want the conversationId %q", c.event, req.SessionID, p.ConversationID)
		}
		if req.TranscriptPath != p.TranscriptPath {
			t.Errorf("%s: TranscriptPath %q, want %q", c.event, req.TranscriptPath, p.TranscriptPath)
		}
		if req.Event != c.event {
			t.Errorf("Event %q, want the flag's %q", req.Event, c.event)
		}
	}
}

// agyStepKinds is the step schema agy documents for injectSteps (SessionStart, PreInvocation and
// PostInvocation accept the same set).
var agyStepKinds = map[string]func(any) bool{
	"systemMessage": func(v any) bool {
		m, ok := v.(map[string]any)
		_, isString := m["systemMessage"].(string)
		return ok && isString // nested object, never a plain string
	},
	"userMessage":      func(v any) bool { _, ok := v.(string); return ok },
	"ephemeralMessage": func(v any) bool { _, ok := v.(string); return ok },
	"errorMessage":     func(v any) bool { _, ok := v.(map[string]any); return ok },
}

// checkAgyInvocationReply fails a reply agy's SessionStart/PreInvocation contract does not define:
// top-level keys other than injectSteps and terminationBehavior, a step that is not exactly one
// known kind of the right shape, or a terminationBehavior other than "terminate".
func checkAgyInvocationReply(t *testing.T, label, out string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("%s: reply %q is not one JSON object: %v", label, out, err)
	}
	for k := range doc {
		if k != "injectSteps" && k != "terminationBehavior" {
			t.Errorf("%s: unknown top-level key %q in %s", label, k, out)
		}
	}
	if tb, ok := doc["terminationBehavior"]; ok && tb != "terminate" {
		t.Errorf("%s: terminationBehavior %v; PreInvocation accepts only \"terminate\"", label, tb)
	}
	steps, _ := doc["injectSteps"].([]any)
	for i, s := range steps {
		m, ok := s.(map[string]any)
		if !ok || len(m) != 1 {
			t.Errorf("%s: step %d is not exactly one step kind: %v", label, i, s)
			continue
		}
		for kind, v := range m {
			valid, known := agyStepKinds[kind]
			if !known || !valid(v) {
				t.Errorf("%s: step %d %q is not a valid agy step: %v", label, i, kind, v)
			}
		}
	}
	return doc
}

func stepTexts(doc map[string]any, kind string) []string {
	var out []string
	steps, _ := doc["injectSteps"].([]any)
	for _, s := range steps {
		m, _ := s.(map[string]any)
		switch v := m[kind].(type) {
		case string:
			out = append(out, v)
		case map[string]any:
			if s, ok := v["systemMessage"].(string); ok {
				out = append(out, s)
			} else if s, ok := v["modelMessage"].(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

// EVERY REPLY SHAPE hookcore can give an Antigravity SessionStart or PreInvocation, against the
// documented schema. A request to the model rides on PreInvocation + AdditionalContext.
func TestRespondAntigravityInvocationReplyShapes(t *testing.T) {
	for _, c := range []struct {
		name      string
		event     string
		resp      Response
		silent    bool     // no output at all
		ephemeral []string // expected ephemeralMessage texts, in order
		system    []string // expected systemMessage texts, in order
		terminate bool
	}{
		{name: "PreInvocation, nothing to say", event: "PreInvocation", resp: Response{}},
		{name: "PreInvocation, a request", event: "PreInvocation", resp: Response{AdditionalContext: "an injected request"},
			ephemeral: []string{"an injected request"}},
		{name: "PreInvocation, a failure line", event: "PreInvocation", resp: Response{Say: "hook failed"},
			system: []string{"hook failed"}},
		{name: "PreInvocation, feedback then request then failure", event: "PreInvocation",
			resp:      Response{Feedback: "fb", AdditionalContext: "ctx", Say: "say"},
			ephemeral: []string{"fb", "ctx"}, system: []string{"say"}},
		{name: "PreInvocation, terminate with a reason", event: "PreInvocation",
			resp: Response{TerminateTurn: true, Reason: "blocked"}, terminate: true},
		{name: "SessionStart, nothing to say", event: "SessionStart", resp: Response{}, silent: true},
		{name: "SessionStart, a request", event: "SessionStart", resp: Response{AdditionalContext: "ctx"},
			ephemeral: []string{"ctx"}},
		{name: "SessionStart, a failure line", event: "SessionStart", resp: Response{Say: "say"}, system: []string{"say"}},
		{name: "SessionStart ignores TerminateTurn", event: "SessionStart", resp: Response{TerminateTurn: true, Reason: "x"}, silent: true},
		{name: "SubagentStart, work list as ephemeralMessage", event: "SubagentStart", resp: Response{AdditionalContext: "work list"},
			ephemeral: []string{"work list"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			req := Request{Engine: EngineAntigravity, Event: c.event, SessionID: "conv-shape", PluginDataDir: t.TempDir()}
			var stdout, stderr bytes.Buffer
			if code := Respond(req, c.resp, &stdout, &stderr, nil); code != 0 {
				t.Fatalf("exit %d, stderr %s", code, stderr.String())
			}
			out := strings.TrimSpace(stdout.String())
			if c.silent {
				if out != "" {
					t.Errorf("want no output, got %s", out)
				}
				return
			}
			doc := checkAgyInvocationReply(t, c.name, out)
			if got := stepTexts(doc, "ephemeralMessage"); strings.Join(got, "|") != strings.Join(c.ephemeral, "|") {
				t.Errorf("ephemeral steps %q, want %q", got, c.ephemeral)
			}
			if got := stepTexts(doc, "systemMessage"); strings.Join(got, "|") != strings.Join(c.system, "|") {
				t.Errorf("system steps %q, want %q", got, c.system)
			}
			if _, has := doc["terminationBehavior"]; has != c.terminate {
				t.Errorf("terminationBehavior present = %v, want %v", has, c.terminate)
			}
			if c.terminate && strings.Join(stepTexts(doc, "errorMessage"), "") != c.resp.Reason {
				t.Errorf("a terminated turn does not carry its reason as an errorMessage: %s", out)
			}
		})
	}
}

// A queued line from an earlier PostTool reaches the NEXT PreInvocation, ahead of the reply's own
// request, and the queue is then empty: a hook's own request must not displace a queued hook
// failure, nor be repeated by a stale queue.
func TestPreInvocationDeliversTheQueueBeforeItsOwnRequest(t *testing.T) {
	dir := t.TempDir()
	post := Request{Engine: EngineAntigravity, Event: "PostTool", SessionID: "conv-q", PluginDataDir: dir}
	var sink bytes.Buffer
	Respond(post, Response{Say: "queued failure"}, &sink, &sink, nil)

	pre := Request{Engine: EngineAntigravity, Event: "PreInvocation", SessionID: "conv-q", PluginDataDir: dir}
	var out bytes.Buffer
	Respond(pre, Response{AdditionalContext: "an injected request"}, &out, &sink, nil)
	doc := checkAgyInvocationReply(t, "flush", strings.TrimSpace(out.String()))
	if got := stepTexts(doc, "systemMessage"); len(got) != 1 || got[0] != "queued failure" {
		t.Errorf("queued line not delivered: %v", got)
	}
	if got := stepTexts(doc, "ephemeralMessage"); len(got) != 1 || got[0] != "an injected request" {
		t.Errorf("own request not delivered: %v", got)
	}
	out.Reset()
	Respond(pre, Response{}, &out, &sink, nil)
	if got := strings.TrimSpace(out.String()); got != "{}" {
		t.Errorf("second PreInvocation replayed the queue: %s", got)
	}
}

func TestAntigravityWorkflowFinishToolAndSubagentSessionStartContract(t *testing.T) {
	subPayload := `{
		"hookEventName": "SessionStart",
		"conversationId": "conv-child-1",
		"parentConversationId": "conv-parent-1",
		"agentName": "red-lens-logic",
		"transcriptPath": "/tmp/child/transcript.jsonl",
		"workspacePaths": ["/tmp/work"]
	}`
	req := Read("SubagentStart", strings.NewReader(subPayload))
	if req.Engine != EngineAntigravity {
		t.Fatalf("engine = %q, want antigravity", req.Engine)
	}
	if req.SessionID != "conv-parent-1" || req.AgentID != "conv-child-1" || req.AgentType != "red-lens-logic" || req.AgentTranscriptPath != "/tmp/child/transcript.jsonl" {
		t.Fatalf("unexpected subagent normalization: %+v", req)
	}

	finishPayload := `{
		"hookEventName": "PreToolUse",
		"conversationId": "conv-child-1",
		"parentConversationId": "conv-parent-1",
		"cwd": "/tmp/work",
		"toolCall": {
			"id": "call-finish-1",
			"name": "finish",
			"args": {"result": {"verdict": "PASS", "summary": "ok"}}
		}
	}`
	freq := Read("PreToolUse", strings.NewReader(finishPayload))
	if freq.Engine != EngineAntigravity || freq.ToolName != "finish" || freq.RawToolName != "finish" {
		t.Fatalf("unexpected finish request: %+v", freq)
	}
	var stdout, stderr bytes.Buffer
	code := Respond(freq, Response{
		Decision:     DecisionAllow,
		Reason:       "record_read attached",
		UpdatedInput: map[string]any{"result": map[string]any{"verdict": "PASS", "summary": "ok", "record_read": map[string]any{"ok": true}}},
	}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("allow exit %d, stderr %s", code, stderr.String())
	}
	var allowDoc struct {
		Decision  string `json:"decision"`
		Reason    string `json:"reason"`
		Overwrite struct {
			Result map[string]any `json:"result"`
		} `json:"overwrite"`
	}
	dec := json.NewDecoder(&stdout)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&allowDoc); err != nil {
		t.Fatalf("finish allow response failed strict decode: %v\n%s", err, stdout.String())
	}
	if allowDoc.Decision != "allow" || allowDoc.Overwrite.Result["verdict"] != "PASS" || allowDoc.Overwrite.Result["record_read"] == nil {
		t.Fatalf("unexpected finish allow overwrite: %+v", allowDoc)
	}

	stdout.Reset()
	stderr.Reset()
	code = Respond(freq, Response{Decision: DecisionDeny, Reason: "open blocker"}, &stdout, &stderr, nil)
	if code != 0 {
		t.Fatalf("deny exit %d, stderr %s", code, stderr.String())
	}
	var denyDoc struct {
		Decision   string `json:"decision"`
		Reason     string `json:"reason"`
		DenyReason string `json:"deny_reason"`
	}
	dec = json.NewDecoder(&stdout)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&denyDoc); err != nil {
		t.Fatalf("finish deny response failed strict decode: %v\n%s", err, stdout.String())
	}
	if denyDoc.Decision != "deny" || denyDoc.DenyReason != "open blocker" {
		t.Fatalf("unexpected finish deny doc: %+v", denyDoc)
	}
}

func TestClaudeStrictHookContractShapes(t *testing.T) {
	t.Run("PreToolUse StructuredOutput updatedInput", func(t *testing.T) {
		req := Read("PreToolUse", strings.NewReader(`{"session_id":"s1","agent_id":"a1","tool_name":"StructuredOutput","tool_input":{"verdict":"PASS"}}`))
		var out, errb bytes.Buffer
		if code := Respond(req, Response{
			Decision:     DecisionAllow,
			Reason:       "record_read",
			UpdatedInput: map[string]any{"verdict": "PASS", "record_read": map[string]any{"ok": true}},
		}, &out, &errb, nil); code != 0 {
			t.Fatalf("exit %d: %s", code, errb.String())
		}
		var doc struct {
			HookSpecificOutput struct {
				HookEventName            string         `json:"hookEventName"`
				PermissionDecision       string         `json:"permissionDecision"`
				PermissionDecisionReason string         `json:"permissionDecisionReason"`
				UpdatedInput             map[string]any `json:"updatedInput"`
			} `json:"hookSpecificOutput"`
		}
		dec := json.NewDecoder(&out)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&doc); err != nil {
			t.Fatalf("strict Claude PreToolUse decode: %v\n%s", err, out.String())
		}
		if doc.HookSpecificOutput.HookEventName != "PreToolUse" || doc.HookSpecificOutput.PermissionDecision != "allow" || doc.HookSpecificOutput.UpdatedInput["record_read"] == nil {
			t.Fatalf("unexpected Claude PreToolUse doc: %+v", doc)
		}
	})

	for _, ev := range []string{"PostToolUse", "PostToolUseFailure", "SubagentStart"} {
		t.Run(ev+" additionalContext", func(t *testing.T) {
			req := Request{Engine: EngineClaude, Event: ev, SessionID: "s1"}
			var out, errb bytes.Buffer
			if code := Respond(req, Response{AdditionalContext: "next item"}, &out, &errb, nil); code != 0 {
				t.Fatalf("exit %d: %s", code, errb.String())
			}
			var doc struct {
				HookSpecificOutput struct {
					HookEventName     string `json:"hookEventName"`
					AdditionalContext string `json:"additionalContext"`
				} `json:"hookSpecificOutput"`
			}
			dec := json.NewDecoder(&out)
			dec.DisallowUnknownFields()
			if err := dec.Decode(&doc); err != nil {
				t.Fatalf("strict Claude %s decode: %v\n%s", ev, err, out.String())
			}
			if doc.HookSpecificOutput.HookEventName != ev || doc.HookSpecificOutput.AdditionalContext != "next item" {
				t.Fatalf("unexpected Claude %s doc: %+v", ev, doc)
			}
		})
	}

	t.Run("SubagentStop silence", func(t *testing.T) {
		req := Request{Engine: EngineClaude, Event: "SubagentStop", SessionID: "s1"}
		var out, errb bytes.Buffer
		if code := Respond(req, Response{}, &out, &errb, nil); code != 0 {
			t.Fatalf("exit %d: %s", code, errb.String())
		}
		if strings.TrimSpace(out.String()) != "" {
			t.Fatalf("expected silent SubagentStop, got %q", out.String())
		}
	})
}
