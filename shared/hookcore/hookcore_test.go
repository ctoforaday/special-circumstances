package hookcore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockSettler struct {
	settleMsg string
	persisted bool
}

func (m *mockSettler) Settle() string {
	return m.settleMsg
}

func (m *mockSettler) Persist() {
	m.persisted = true
}

func TestCanonicalToolName(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"run_command", "Bash"},
		{"Bash", "Bash"},
		{"write_to_file", "Write"},
		{"write_file", "Write"},
		{"Write", "Write"},
		{"replace_file_content", "Edit"},
		{"multi_replace_file_content", "Edit"},
		{"edit_file", "Edit"},
		{"Edit", "Edit"},
		{"view_file", "Read"},
		{"read_file", "Read"},
		{"Read", "Read"},
		{"search_web", "WebSearch"},
		{"WebSearch", "WebSearch"},
		{"read_url_content", "WebFetch"},
		{"read_browser_page", "WebFetch"},
		{"WebFetch", "WebFetch"},
		{"find_by_name", "Glob"},
		{"Glob", "Glob"},
		{"grep_search", "Grep"},
		{"Grep", "Grep"},
		{"custom_mcp_tool", "custom_mcp_tool"},
	}

	for _, c := range cases {
		if got := CanonicalToolName(c.raw); got != c.want {
			t.Errorf("CanonicalToolName(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

func TestDetectEngine(t *testing.T) {
	// 1. Env precedence
	t.Setenv("ANTIGRAVITY_CONVERSATION_ID", "conv-123")
	if got := DetectEngine(nil); got != EngineAntigravity {
		t.Errorf("DetectEngine with ANTIGRAVITY_CONVERSATION_ID = %v, want %v", got, EngineAntigravity)
	}
	_ = os.Unsetenv("ANTIGRAVITY_CONVERSATION_ID")

	// 2. Claude payload sniffing
	claudePayload := []byte(`{"session_id":"sess-1","tool_name":"Bash","cwd":"/repo"}`)
	if got := DetectEngine(claudePayload); got != EngineClaude {
		t.Errorf("DetectEngine(claudePayload) = %v, want %v", got, EngineClaude)
	}

	// 3. Antigravity payload sniffing
	agyPayload := []byte(`{"conversationId":"conv-1","workspacePaths":["/repo"],"toolCall":{"name":"run_command"}}`)
	if got := DetectEngine(agyPayload); got != EngineAntigravity {
		t.Errorf("DetectEngine(agyPayload) = %v, want %v", got, EngineAntigravity)
	}

	// 4. Default fallback without env
	_ = os.Unsetenv("CLAUDE_PLUGIN_ROOT")
	_ = os.Unsetenv("CLAUDE_PROJECT_DIR")
	if got := DetectEngine([]byte(`{}`)); got != EngineClaude {
		t.Errorf("DetectEngine(empty) = %v, want %v", got, EngineClaude)
	}
}

func TestProjectDir(t *testing.T) {
	// 1. Environment variable takes highest priority
	if got := ProjectDir("/env/dir", "/payload/cwd", []string{"/workspace/dir"}, ""); got != "/env/dir" {
		t.Errorf("ProjectDir env = %q, want /env/dir", got)
	}

	// 2. Payload CWD
	if got := ProjectDir("", "/payload/cwd", []string{"/workspace/dir"}, ""); got != "/payload/cwd" {
		t.Errorf("ProjectDir cwd = %q, want /payload/cwd", got)
	}

	// 3. WorkspacePaths[0]
	if got := ProjectDir("", "", []string{"/workspace/dir"}, ""); got != "/workspace/dir" {
		t.Errorf("ProjectDir workspacePaths = %q, want /workspace/dir", got)
	}

	// 4. Fallback to nothing (never os.Getwd())
	if got := ProjectDir("", "", nil, ""); got != "" {
		t.Errorf("ProjectDir with all empty = %q, want empty", got)
	}
}

func TestReadClaudePayloads(t *testing.T) {
	// 1. PreToolUse
	input := `{"session_id":"s-100","cwd":"/project","hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"git status"}}`
	req := Read("PreToolUse", strings.NewReader(input))

	if req.Engine != EngineClaude {
		t.Errorf("req.Engine = %v, want %v", req.Engine, EngineClaude)
	}
	if !req.Parsed {
		t.Errorf("req.Parsed = false, want true")
	}
	if req.SessionID != "s-100" {
		t.Errorf("req.SessionID = %q, want s-100", req.SessionID)
	}
	if req.ToolName != "Bash" {
		t.Errorf("req.ToolName = %q, want Bash", req.ToolName)
	}
	if req.Command != "git status" {
		t.Errorf("req.Command = %q, want git status", req.Command)
	}

	// Check normalized tool input has both command and CommandLine
	var ti map[string]any
	if err := json.Unmarshal(req.ToolInput, &ti); err != nil {
		t.Fatalf("unmarshal ToolInput: %v", err)
	}
	if ti["CommandLine"] != "git status" {
		t.Errorf("ToolInput CommandLine = %v, want git status", ti["CommandLine"])
	}

	// 2. PostToolUse with tool_response
	input2 := `{"session_id":"s-100","cwd":"/project","hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"main.go","content":"package main"},"tool_response":{"output":"ok"}}`
	req2 := Read("PostToolUse", strings.NewReader(input2))
	if req2.ToolName != "Write" {
		t.Errorf("req2.ToolName = %q, want Write", req2.ToolName)
	}
	if req2.FilePath != "main.go" {
		t.Errorf("req2.FilePath = %q, want main.go", req2.FilePath)
	}
	if len(req2.ToolResult) == 0 {
		t.Errorf("req2.ToolResult is empty, want JSON object")
	}

	// 3. PostToolUseFailure
	input3 := `{"session_id":"s-100","cwd":"/project","hook_event_name":"PostToolUseFailure","error":"process killed"}`
	req3 := Read("PostToolUseFailure", strings.NewReader(input3))
	if req3.ToolError != "process killed" {
		t.Errorf("req3.ToolError = %q, want process killed", req3.ToolError)
	}

	// 4. Raw invalid JSON preservation (secretsgate contract #211)
	malformed := `{"session_id": "malformed", "tool_input": BAD_JSON`
	req4 := Read("PreToolUse", strings.NewReader(malformed))
	if req4.Parsed {
		t.Errorf("req4.Parsed = true, want false")
	}
	if string(req4.Raw) != malformed {
		t.Errorf("req4.Raw = %q, want %q", string(req4.Raw), malformed)
	}
}

func TestReadAntigravityPayloads(t *testing.T) {
	// 1. PreTool run_command
	input := `{
		"conversationId": "agy-conv-1",
		"workspacePaths": ["/agy/workspace"],
		"artifactDirectoryPath": "/home/user/.gemini/antigravity-cli/brain/agy-conv-1",
		"toolCall": {
			"name": "run_command",
			"args": {
				"CommandLine": "ls -l",
				"Cwd": "/agy/workspace",
				"toolAction": "list files"
			}
		}
	}`
	req := Read("PreTool", strings.NewReader(input))
	if req.Engine != EngineAntigravity {
		t.Errorf("req.Engine = %v, want %v", req.Engine, EngineAntigravity)
	}
	if req.SessionID != "agy-conv-1" {
		t.Errorf("req.SessionID = %q, want agy-conv-1", req.SessionID)
	}
	if req.ToolName != "Bash" {
		t.Errorf("req.ToolName = %q, want Bash", req.ToolName)
	}
	if req.Command != "ls -l" {
		t.Errorf("req.Command = %q, want ls -l", req.Command)
	}
	if req.ProjectDir != "/agy/workspace" {
		t.Errorf("req.ProjectDir = %q, want /agy/workspace", req.ProjectDir)
	}

	// Check synthesized snake_case keys in ToolInput
	var ti map[string]any
	if err := json.Unmarshal(req.ToolInput, &ti); err != nil {
		t.Fatalf("unmarshal ToolInput: %v", err)
	}
	if ti["command"] != "ls -l" {
		t.Errorf("ToolInput command = %v, want ls -l", ti["command"])
	}
	if ti["description"] != "list files" {
		t.Errorf("ToolInput description = %v, want list files", ti["description"])
	}

	// 2. PreTool write_to_file & oneharness write_file
	inputWrite := `{
		"conversationId": "agy-conv-1",
		"toolCall": {
			"name": "write_to_file",
			"args": {
				"TargetFile": "/path/to/foo.go",
				"CodeContent": "package foo"
			}
		}
	}`
	reqWrite := Read("PreTool", strings.NewReader(inputWrite))
	if reqWrite.ToolName != "Write" {
		t.Errorf("reqWrite.ToolName = %q, want Write", reqWrite.ToolName)
	}
	if reqWrite.FilePath != "/path/to/foo.go" {
		t.Errorf("reqWrite.FilePath = %q, want /path/to/foo.go", reqWrite.FilePath)
	}

	// 3. PreTool replace_file_content & edit_file
	inputEdit := `{
		"conversationId": "agy-conv-1",
		"toolCall": {
			"name": "edit_file",
			"args": {
				"Path": "/path/to/foo.go",
				"Target": "old",
				"Replacement": "new",
				"AllowMultiple": true
			}
		}
	}`
	reqEdit := Read("PreTool", strings.NewReader(inputEdit))
	if reqEdit.ToolName != "Edit" {
		t.Errorf("reqEdit.ToolName = %q, want Edit", reqEdit.ToolName)
	}
	if reqEdit.FilePath != "/path/to/foo.go" {
		t.Errorf("reqEdit.FilePath = %q, want /path/to/foo.go", reqEdit.FilePath)
	}
	var tiEdit map[string]any
	_ = json.Unmarshal(reqEdit.ToolInput, &tiEdit)
	if tiEdit["old_string"] != "old" || tiEdit["new_string"] != "new" || tiEdit["replace_all"] != true {
		t.Errorf("Edit args normalization failed: %v", tiEdit)
	}

	// 4. Subagent conversation identification
	inputSub := `{
		"conversationId": "sub-conv-2",
		"parentConversationId": "root-conv-1",
		"agentName": "researcher",
		"transcriptPath": "/path/to/sub/transcript.jsonl"
	}`
	reqSub := Read("SessionStart", strings.NewReader(inputSub))
	if reqSub.AgentID != "sub-conv-2" {
		t.Errorf("reqSub.AgentID = %q, want sub-conv-2", reqSub.AgentID)
	}
	if reqSub.AgentType != "researcher" {
		t.Errorf("reqSub.AgentType = %q, want researcher", reqSub.AgentType)
	}
	if reqSub.AgentTranscriptPath != "/path/to/sub/transcript.jsonl" {
		t.Errorf("reqSub.AgentTranscriptPath = %q, want /path/to/sub/transcript.jsonl", reqSub.AgentTranscriptPath)
	}
}

func TestRespondClaudePreToolUse(t *testing.T) {
	req := Request{
		Engine: EngineClaude,
		Event:  "PreToolUse",
	}

	// 1. Deny decision
	resp := Response{
		Decision: DecisionDeny,
		Reason:   "secrets detected: gitleaks rule generic-api-key",
		Say:      "Security violation",
	}
	var stdout, stderr bytes.Buffer
	rec := &mockSettler{settleMsg: "failure settlement info"}

	code := Respond(req, resp, &stdout, &stderr, rec)
	if code != 0 {
		t.Errorf("Respond exit code = %d, want 0", code)
	}
	if !rec.persisted {
		t.Errorf("rec.Persist() was not called")
	}

	var out struct {
		HookSpecificOutput struct {
			HookEventName            string `json:"hookEventName"`
			PermissionDecision       string `json:"permissionDecision"`
			PermissionDecisionReason string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
		SystemMessage string `json:"systemMessage"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal Claude PreToolUse output: %v", err)
	}
	if out.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want deny", out.HookSpecificOutput.PermissionDecision)
	}
	if out.HookSpecificOutput.PermissionDecisionReason != resp.Reason {
		t.Errorf("PermissionDecisionReason = %q, want %q", out.HookSpecificOutput.PermissionDecisionReason, resp.Reason)
	}
	if !strings.Contains(out.SystemMessage, "Security violation") || !strings.Contains(out.SystemMessage, "failure settlement info") {
		t.Errorf("SystemMessage = %q, missing expected content", out.SystemMessage)
	}

	// 2. Allow with command rewrite
	respAllow := Response{
		Decision:       DecisionAllow,
		UpdatedCommand: "safe_command",
	}
	stdout.Reset()
	Respond(req, respAllow, &stdout, &stderr, nil)
	var outAllow struct {
		HookSpecificOutput struct {
			UpdatedInput map[string]string `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &outAllow); err != nil {
		t.Fatalf("unmarshal Claude UpdatedInput: %v", err)
	}
	if outAllow.HookSpecificOutput.UpdatedInput["command"] != "safe_command" {
		t.Errorf("updatedInput[command] = %q, want safe_command", outAllow.HookSpecificOutput.UpdatedInput["command"])
	}
}

func TestRespondClaudePostToolUse(t *testing.T) {
	req := Request{
		Engine: EngineClaude,
		Event:  "PostToolUse",
	}

	// 1. Feedback triggers exit 2 and writes to stderr
	respFeedback := Response{
		Feedback: "qlty: unformatted code in scratch/bad.go",
		Say:      "Lint warnings recorded",
	}
	var stdout, stderr bytes.Buffer
	code := Respond(req, respFeedback, &stdout, &stderr, nil)
	if code != 2 {
		t.Errorf("Respond Claude PostToolUse with Feedback exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unformatted code") {
		t.Errorf("stderr = %q, want feedback message", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Lint warnings recorded") {
		t.Errorf("stdout = %q, want systemMessage", stdout.String())
	}

	// 2. AdditionalContext without Feedback exits 0
	stdout.Reset()
	stderr.Reset()
	respContext := Response{
		AdditionalContext: "Strike 2 on tool error",
	}
	code = Respond(req, respContext, &stdout, &stderr, nil)
	if code != 0 {
		t.Errorf("Respond Claude PostToolUse with AdditionalContext exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "Strike 2 on tool error") {
		t.Errorf("stdout = %q, want additionalContext", stdout.String())
	}
}

func TestRespondAntigravityPreTool(t *testing.T) {
	tmpDir := t.TempDir()
	req := Request{
		Engine:        EngineAntigravity,
		Event:         "PreTool",
		SessionID:     "conv-test-1",
		PluginDataDir: tmpDir,
	}

	// 1. Deny decision
	respDeny := Response{
		Decision: DecisionDeny,
		Reason:   "access denied",
		Say:      "Advisory note",
	}
	var stdout, stderr bytes.Buffer
	code := Respond(req, respDeny, &stdout, &stderr, nil)
	if code != 0 {
		t.Errorf("Antigravity PreTool exit code = %d, want 0", code)
	}

	var out struct {
		Decision   string `json:"decision"`
		Reason     string `json:"reason"`
		DenyReason string `json:"deny_reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal Antigravity PreTool output: %v", err)
	}
	if out.Decision != "deny" || out.Reason != "access denied" || out.DenyReason != "access denied" {
		t.Errorf("PreTool output mismatch: %+v", out)
	}

	// Verify Say was enqueued to pending
	items := flushPending(tmpDir, "conv-test-1")
	if len(items) != 1 || items[0].Say != "Advisory note" {
		t.Errorf("flushed pending items = %+v, want 1 item with Say", items)
	}

	// 2. Command overwrite
	stdout.Reset()
	respOverwrite := Response{
		Decision:       DecisionAllow,
		UpdatedCommand: "echo rewritten",
	}
	Respond(req, respOverwrite, &stdout, &stderr, nil)
	var outOverwrite struct {
		Overwrite map[string]string `json:"overwrite"`
	}
	_ = json.Unmarshal(stdout.Bytes(), &outOverwrite)
	if outOverwrite.Overwrite["CommandLine"] != "echo rewritten" {
		t.Errorf("overwrite[CommandLine] = %q, want echo rewritten", outOverwrite.Overwrite["CommandLine"])
	}
}

func TestRespondAntigravityPostToolAndPreInvocation(t *testing.T) {
	tmpDir := t.TempDir()
	reqPost := Request{
		Engine:        EngineAntigravity,
		Event:         "PostTool",
		SessionID:     "conv-queue-1",
		PluginDataDir: tmpDir,
	}

	// 1. PostTool emits {} and queues feedback
	respPost := Response{
		Feedback:          "Fix syntax in foo.go",
		AdditionalContext: "Strike count 1",
	}
	var stdout, stderr bytes.Buffer
	code := Respond(reqPost, respPost, &stdout, &stderr, nil)
	if code != 0 {
		t.Errorf("PostTool exit code = %d, want 0", code)
	}
	if strings.TrimSpace(stdout.String()) != "{}" {
		t.Errorf("PostTool stdout = %q, want {}", stdout.String())
	}

	// Verify file was written
	queueFile := filepath.Join(tmpDir, "pending_conv-queue-1.jsonl")
	if _, err := os.Stat(queueFile); err != nil {
		t.Fatalf("pending queue file was not created: %v", err)
	}

	// 2. PreInvocation flushes queued items and deletes queue file
	reqPreInvoc := Request{
		Engine:        EngineAntigravity,
		Event:         "PreInvocation",
		SessionID:     "conv-queue-1",
		PluginDataDir: tmpDir,
	}
	stdout.Reset()
	respPreInvoc := Response{
		Say: "System is ready",
	}
	code = Respond(reqPreInvoc, respPreInvoc, &stdout, &stderr, nil)
	if code != 0 {
		t.Errorf("PreInvocation exit code = %d, want 0", code)
	}

	// Verify queue file was deleted
	if _, err := os.Stat(queueFile); !os.IsNotExist(err) {
		t.Errorf("pending queue file was not deleted after PreInvocation flush")
	}

	var outPre struct {
		InjectSteps         []map[string]any `json:"injectSteps"`
		TerminationBehavior string           `json:"terminationBehavior"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &outPre); err != nil {
		t.Fatalf("unmarshal PreInvocation output: %v", err)
	}
	if outPre.TerminationBehavior != "force_continue" {
		t.Errorf("TerminationBehavior = %q, want force_continue", outPre.TerminationBehavior)
	}
	if len(outPre.InjectSteps) != 3 {
		t.Fatalf("len(InjectSteps) = %d, want 3 (Feedback, AdditionalContext, Say)", len(outPre.InjectSteps))
	}
}

func TestRespondAntigravityTerminateTurn(t *testing.T) {
	req := Request{
		Engine: EngineAntigravity,
		Event:  "PreInvocation",
	}
	resp := Response{
		TerminateTurn: true,
		Reason:        "Max turn budget exceeded",
	}
	var stdout, stderr bytes.Buffer
	code := Respond(req, resp, &stdout, &stderr, nil)
	if code != 0 {
		t.Errorf("TerminateTurn exit code = %d, want 0", code)
	}

	var out struct {
		InjectSteps         []map[string]any `json:"injectSteps"`
		TerminationBehavior string           `json:"terminationBehavior"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if out.TerminationBehavior != "terminate" {
		t.Errorf("TerminationBehavior = %q, want terminate", out.TerminationBehavior)
	}
	if len(out.InjectSteps) != 1 {
		t.Fatalf("len(InjectSteps) = %d, want 1 error message step", len(out.InjectSteps))
	}
	errMsg, ok := out.InjectSteps[0]["errorMessage"].(map[string]any)
	if !ok || errMsg["modelMessage"] != resp.Reason {
		t.Errorf("errorMessage step invalid: %v", out.InjectSteps[0])
	}
}

func TestContractPreCompactClaudeAndAntigravity(t *testing.T) {
	// Claude: Must emit plain text prose verbatim on stdout, NEVER a JSON object.
	reqClaude := Request{Engine: EngineClaude, Event: "PreCompact"}
	respClaude := Response{PreCompactSteer: "## Instructions\nPreserve all validation loops."}
	var stdoutClaude, stderrClaude bytes.Buffer
	Respond(reqClaude, respClaude, &stdoutClaude, &stderrClaude, nil)

	sClaude := strings.TrimSpace(stdoutClaude.String())
	if strings.HasPrefix(sClaude, "{") {
		t.Errorf("Claude PreCompact wrote JSON to stdout, want plain text: %s", sClaude)
	}
	if !strings.Contains(sClaude, "Preserve all validation loops") {
		t.Errorf("Claude PreCompact output missing steer text: %s", sClaude)
	}

	// Antigravity: Emits injectSteps with systemMessage.
	reqAgy := Request{Engine: EngineAntigravity, Event: "PreCompact"}
	respAgy := Response{PreCompactSteer: "## Instructions\nPreserve all validation loops."}
	var stdoutAgy, stderrAgy bytes.Buffer
	Respond(reqAgy, respAgy, &stdoutAgy, &stderrAgy, nil)

	var outAgy struct {
		InjectSteps []map[string]any `json:"injectSteps"`
	}
	if err := json.Unmarshal(stdoutAgy.Bytes(), &outAgy); err != nil {
		t.Fatalf("unmarshal Antigravity PreCompact output: %v", err)
	}
	if len(outAgy.InjectSteps) != 1 {
		t.Fatalf("len(InjectSteps) = %d, want 1", len(outAgy.InjectSteps))
	}
	sysMsg, ok := outAgy.InjectSteps[0]["systemMessage"].(map[string]any)
	if !ok || !strings.Contains(sysMsg["systemMessage"].(string), "Preserve all validation loops") {
		t.Errorf("Antigravity PreCompact missing steer message: %v", outAgy.InjectSteps[0])
	}
}

func TestContractSubagentStopAndSessionEnd(t *testing.T) {
	// Claude: SubagentStop and SessionEnd seal silently and must never write to stdout.
	for _, ev := range []string{"SubagentStop", "SessionEnd"} {
		req := Request{Engine: EngineClaude, Event: ev}
		var stdout, stderr bytes.Buffer
		code := Respond(req, Response{AdditionalContext: "unsolicited"}, &stdout, &stderr, nil)
		if code != 0 {
			t.Errorf("Claude %s exit code = %d, want 0", ev, code)
		}
		if len(stdout.Bytes()) != 0 {
			t.Errorf("Claude %s wrote to stdout: %q, want zero bytes", ev, stdout.String())
		}
	}

	// Antigravity: SubagentStop and SessionEnd emit {} on stdout.
	for _, ev := range []string{"SubagentStop", "SessionEnd"} {
		req := Request{Engine: EngineAntigravity, Event: ev}
		var stdout, stderr bytes.Buffer
		code := Respond(req, Response{AdditionalContext: "unsolicited"}, &stdout, &stderr, nil)
		if code != 0 {
			t.Errorf("Antigravity %s exit code = %d, want 0", ev, code)
		}
		if strings.TrimSpace(stdout.String()) != "{}" {
			t.Errorf("Antigravity %s stdout = %q, want {}", ev, stdout.String())
		}
	}
}

func TestContractToolMappingMatrix(t *testing.T) {
	cases := []struct {
		name          string
		payload       string
		wantCanonical string
		wantFile      string
		wantCommand   string
	}{
		{
			name:          "Claude Bash command",
			payload:       `{"session_id":"s","tool_name":"Bash","tool_input":{"command":"git diff"}}`,
			wantCanonical: "Bash",
			wantCommand:   "git diff",
		},
		{
			name:          "Antigravity run_command CommandLine",
			payload:       `{"conversationId":"c","toolCall":{"name":"run_command","args":{"CommandLine":"git diff"}}}`,
			wantCanonical: "Bash",
			wantCommand:   "git diff",
		},
		{
			name:          "Claude Write file_path",
			payload:       `{"session_id":"s","tool_name":"Write","tool_input":{"file_path":"foo.go","content":"code"}}`,
			wantCanonical: "Write",
			wantFile:      "foo.go",
		},
		{
			name:          "Antigravity write_to_file TargetFile",
			payload:       `{"conversationId":"c","toolCall":{"name":"write_to_file","args":{"TargetFile":"foo.go","CodeContent":"code"}}}`,
			wantCanonical: "Write",
			wantFile:      "foo.go",
		},
		{
			name:          "Antigravity oneharness write_file Path",
			payload:       `{"conversationId":"c","toolCall":{"name":"write_file","args":{"Path":"foo.go","Content":"code"}}}`,
			wantCanonical: "Write",
			wantFile:      "foo.go",
		},
		{
			name:          "Claude Edit file_path",
			payload:       `{"session_id":"s","tool_name":"Edit","tool_input":{"file_path":"bar.go","old_string":"a","new_string":"b"}}`,
			wantCanonical: "Edit",
			wantFile:      "bar.go",
		},
		{
			name:          "Antigravity replace_file_content TargetContent",
			payload:       `{"conversationId":"c","toolCall":{"name":"replace_file_content","args":{"TargetFile":"bar.go","TargetContent":"a","ReplacementContent":"b"}}}`,
			wantCanonical: "Edit",
			wantFile:      "bar.go",
		},
		{
			name:          "Antigravity edit_file Target",
			payload:       `{"conversationId":"c","toolCall":{"name":"edit_file","args":{"Path":"bar.go","Target":"a","Replacement":"b"}}}`,
			wantCanonical: "Edit",
			wantFile:      "bar.go",
		},
		{
			name:          "Claude Read file_path",
			payload:       `{"session_id":"s","tool_name":"Read","tool_input":{"file_path":"doc.md"}}`,
			wantCanonical: "Read",
			wantFile:      "doc.md",
		},
		{
			name:          "Antigravity view_file AbsolutePath",
			payload:       `{"conversationId":"c","toolCall":{"name":"view_file","args":{"AbsolutePath":"/abs/doc.md"}}}`,
			wantCanonical: "Read",
			wantFile:      "/abs/doc.md",
		},
		{
			name:          "Antigravity read_file Path",
			payload:       `{"conversationId":"c","toolCall":{"name":"read_file","args":{"Path":"/abs/doc.md"}}}`,
			wantCanonical: "Read",
			wantFile:      "/abs/doc.md",
		},
		{
			name:          "Claude WebSearch query",
			payload:       `{"session_id":"s","tool_name":"WebSearch","tool_input":{"query":"go test"}}`,
			wantCanonical: "WebSearch",
		},
		{
			name:          "Antigravity search_web Query",
			payload:       `{"conversationId":"c","toolCall":{"name":"search_web","args":{"Query":"go test"}}}`,
			wantCanonical: "WebSearch",
		},
		{
			name:          "Claude WebFetch url",
			payload:       `{"session_id":"s","tool_name":"WebFetch","tool_input":{"url":"https://example.com"}}`,
			wantCanonical: "WebFetch",
		},
		{
			name:          "Antigravity read_url_content Url",
			payload:       `{"conversationId":"c","toolCall":{"name":"read_url_content","args":{"Url":"https://example.com"}}}`,
			wantCanonical: "WebFetch",
		},
		{
			name:          "Antigravity find_by_name SearchDirectory",
			payload:       `{"conversationId":"c","toolCall":{"name":"find_by_name","args":{"SearchDirectory":"/src","Pattern":"*.go"}}}`,
			wantCanonical: "Glob",
		},
		{
			name:          "Antigravity grep_search SearchPath",
			payload:       `{"conversationId":"c","toolCall":{"name":"grep_search","args":{"SearchPath":"/src","Query":"func"}}}`,
			wantCanonical: "Grep",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := Read("PreTool", strings.NewReader(c.payload))
			if req.ToolName != c.wantCanonical {
				t.Errorf("ToolName = %q, want %q", req.ToolName, c.wantCanonical)
			}
			if c.wantFile != "" && req.FilePath != c.wantFile {
				t.Errorf("FilePath = %q, want %q", req.FilePath, c.wantFile)
			}
			if c.wantCommand != "" && req.Command != c.wantCommand {
				t.Errorf("Command = %q, want %q", req.Command, c.wantCommand)
			}
		})
	}
}

func TestContractMultiTurnPendingFeedbackQueue(t *testing.T) {
	dataDir := t.TempDir()
	sessionID := "turn-cycle-1"

	// Turn 1, Tool 1: PostTool records linter feedback
	req1 := Request{
		Engine:        EngineAntigravity,
		Event:         "PostTool",
		SessionID:     sessionID,
		PluginDataDir: dataDir,
	}
	Respond(req1, Response{Feedback: "qlty: unformatted code in a.go"}, &bytes.Buffer{}, &bytes.Buffer{}, nil)

	// Turn 1, Tool 2: PostTool records strike counter context
	req2 := Request{
		Engine:        EngineAntigravity,
		Event:         "PostTool",
		SessionID:     sessionID,
		PluginDataDir: dataDir,
	}
	Respond(req2, Response{AdditionalContext: "Strike count 2 on test error"}, &bytes.Buffer{}, &bytes.Buffer{}, nil)

	// Turn 1, Tool 3: PreTool records push freeze advisory Say
	req3 := Request{
		Engine:        EngineAntigravity,
		Event:         "PreTool",
		SessionID:     sessionID,
		PluginDataDir: dataDir,
	}
	Respond(req3, Response{Decision: DecisionAllow, Say: "[push-freeze] pushing directly to main"}, &bytes.Buffer{}, &bytes.Buffer{}, nil)

	// Turn 2: PreInvocation flushes all 3 items in order
	reqTurn2 := Request{
		Engine:        EngineAntigravity,
		Event:         "PreInvocation",
		SessionID:     sessionID,
		PluginDataDir: dataDir,
	}
	var stdoutTurn2, stderrTurn2 bytes.Buffer
	Respond(reqTurn2, Response{}, &stdoutTurn2, &stderrTurn2, nil)

	var outTurn2 struct {
		InjectSteps []map[string]any `json:"injectSteps"`
	}
	if err := json.Unmarshal(stdoutTurn2.Bytes(), &outTurn2); err != nil {
		t.Fatalf("unmarshal Turn 2 output: %v", err)
	}
	if len(outTurn2.InjectSteps) != 3 {
		t.Fatalf("len(InjectSteps) = %d, want 3", len(outTurn2.InjectSteps))
	}
	if outTurn2.InjectSteps[0]["ephemeralMessage"] != "qlty: unformatted code in a.go" {
		t.Errorf("Step 0 = %v, want qlty feedback", outTurn2.InjectSteps[0])
	}
	if outTurn2.InjectSteps[1]["ephemeralMessage"] != "Strike count 2 on test error" {
		t.Errorf("Step 1 = %v, want strike context", outTurn2.InjectSteps[1])
	}
	step2Sys := outTurn2.InjectSteps[2]["systemMessage"].(map[string]any)["systemMessage"]
	if !strings.Contains(step2Sys.(string), "push-freeze") {
		t.Errorf("Step 2 = %v, want push freeze systemMessage", outTurn2.InjectSteps[2])
	}

	// Queue file must now be deleted
	queueFile := filepath.Join(dataDir, "pending_"+sessionID+".jsonl")
	if _, err := os.Stat(queueFile); !os.IsNotExist(err) {
		t.Errorf("queue file still exists after Turn 2 flush")
	}

	// Turn 3: Subsequent PreInvocation with no pending items produces empty injectSteps
	var stdoutTurn3, stderrTurn3 bytes.Buffer
	Respond(reqTurn2, Response{}, &stdoutTurn3, &stderrTurn3, nil)
	var outTurn3 struct {
		InjectSteps []map[string]any `json:"injectSteps"`
	}
	if err := json.Unmarshal(stdoutTurn3.Bytes(), &outTurn3); err != nil {
		t.Fatalf("unmarshal Turn 3 output: %v", err)
	}
	if len(outTurn3.InjectSteps) != 0 {
		t.Errorf("len(InjectSteps) = %d, want 0 on empty queue", len(outTurn3.InjectSteps))
	}
}

func TestContractSettlerPersistence(t *testing.T) {
	rec := &mockSettler{settleMsg: "failure settlement recorded"}
	req := Request{Engine: EngineClaude, Event: "SessionStart"}
	var stdout, stderr bytes.Buffer

	Respond(req, Response{Say: "Welcome"}, &stdout, &stderr, rec)

	if !rec.persisted {
		t.Errorf("rec.Persist() was not called")
	}
	var out struct {
		SystemMessage string `json:"systemMessage"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !strings.Contains(out.SystemMessage, "Welcome") || !strings.Contains(out.SystemMessage, "failure settlement recorded") {
		t.Errorf("SystemMessage = %q, missing settlement text", out.SystemMessage)
	}
}

func TestContractMalformedStdinSecretsScanning(t *testing.T) {
	// Exact raw byte preservation for secretsgate (#211)
	malformed := "{\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"echo ghp_SECRET\""
	req := Read("PreToolUse", strings.NewReader(malformed))

	if req.Parsed {
		t.Errorf("req.Parsed = true, want false")
	}
	if !strings.Contains(string(req.Raw), "ghp_SECRET") {
		t.Errorf("req.Raw does not preserve raw secret token: %q", string(req.Raw))
	}
}
