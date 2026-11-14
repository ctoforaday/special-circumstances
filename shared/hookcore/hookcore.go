// Package hookcore provides a pure-Go-standard-library wire adapter normalizing
// process boundaries (env, stdin, stdout, stderr, exit code) between Claude Code
// (claude 2.1.309) and Google Antigravity (agy 1.4.12 exa.hooks_pb).
package hookcore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Engine represents the executing AI assistant harness.
type Engine string

const (
	EngineClaude      Engine = "claude"
	EngineAntigravity Engine = "antigravity"
)

// Decision represents a pre-tool execution gate decision.
type Decision string

const (
	DecisionNone  Decision = ""
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

// Settler is satisfied by hook failure recorders without requiring a circular package dependency.
type Settler interface {
	Settle() string
	Persist()
}

// Request holds normalized hook input parameters across both harnesses.
type Request struct {
	Engine              Engine
	Event               string
	Raw                 []byte           // Exact stdin bytes (preserved for raw scanning)
	Parsed              bool             // Whether stdin parsed as valid JSON
	ParseErr            error            // JSON decode error when Parsed is false
	ProjectDir          string           // CLAUDE_PROJECT_DIR -> cwd -> workspacePaths[0] -> tool-arg git root
	PluginRoot          string           // CLAUDE_PLUGIN_ROOT -> PLUGIN_ROOT -> os.Executable() grandparent -> os.Getwd()
	PluginDataDir       string           // CLAUDE_PLUGIN_DATA -> PLUGIN_DATA -> <appDataDir>/plugin-data/<plugin> -> <ProjectDir>/.claude
	SessionID           string           // session_id | conversationId
	TranscriptPath      string           // transcript_path | transcriptPath
	ArtifactDirPath     string           // artifactDirectoryPath (Antigravity)
	AgentID             string           // agent_id | (conversationId when parentConversationId != "")
	AgentType           string           // agent_type | agentName
	AgentTranscriptPath string           // agent_transcript_path | (transcriptPath when subagent)
	ToolName            string           // Canonical ("Bash", "Write", "Edit", "Read", "WebSearch", "WebFetch", "Glob", "Grep")
	RawToolName         string           // Original wire name ("run_command", "write_to_file", "write_file", ...)
	RawToolInput        map[string]any   // Original wire tool input map without synthetic keys
	ToolInput           json.RawMessage  // Normalized JSON carrying canonical snake_case keys + original fields
	ToolResult          json.RawMessage  // tool_response (Claude) | PostToolHookArgs.result (Antigravity 1.4.12+)
	FilePath            string           // Extracted file path (file_path | path | Path | TargetFile | AbsolutePath)
	Command             string           // Extracted shell command (command | CommandLine)
	ToolError           string           // error (PostToolUseFailure on Claude; PostTool.error on Antigravity)
	LastUserInput       string           // prompt (Claude) | lastUserInput (Antigravity 1.4.7+)
	IsInterrupt         bool             // is_interrupt | (terminationReason == "USER_CANCELED")
	StopHookActive      bool             // stop_hook_active | (executionNum > 0)
	TerminationReason   string           // PostInvocation.terminationReason on Antigravity
	Trigger             string           // PreCompact/PostCompact trigger
	CustomInstructions  string           // PreCompact custom_instructions
	CompactSummary      string           // PostCompact compact_summary
	CWD                 string           // cwd
	HookEventName       string           // hook_event_name
	Reason              string           // reason (SessionEnd: clear|logout|prompt_input_exit|other)
	BackgroundTasks     *json.RawMessage // background_tasks (SubagentStop)
	SessionCrons        *json.RawMessage // session_crons (SubagentStop)
	FileEvent           string           // event (FileChanged: change|add|unlink)
}

// Response holds hook return parameters to be emitted to the harness.
type Response struct {
	Decision          Decision
	Reason            string
	UpdatedCommand    string
	UpdatedInput      map[string]any
	Feedback          string
	AdditionalContext string
	WatchPaths        []string
	PreCompactSteer   string
	Say               string
	Warnings          string
	TerminateTurn     bool // PreInvocation: emits terminationBehavior="terminate" + errorMessage on Antigravity
}

// CanonicalToolName maps raw wire tool names from either harness to canonical names.
func CanonicalToolName(raw string) string {
	switch raw {
	case "Bash", "run_command":
		return "Bash"
	case "Write", "write_to_file", "write_file":
		return "Write"
	case "Edit", "replace_file_content", "multi_replace_file_content", "edit_file":
		return "Edit"
	case "Read", "view_file", "read_file":
		return "Read"
	case "WebSearch", "search_web":
		return "WebSearch"
	case "WebFetch", "read_url_content", "read_browser_page":
		return "WebFetch"
	case "Glob", "find_by_name":
		return "Glob"
	case "Grep", "grep_search":
		return "Grep"
	default:
		return raw
	}
}

// DetectEngine sniffs the harness environment from environment variables and payload signatures.
func DetectEngine(raw []byte) Engine {
	if len(raw) > 0 {
		var probe map[string]any
		if err := json.Unmarshal(raw, &probe); err == nil {
			// Definitive Antigravity keys
			if _, ok := probe["conversationId"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["workspacePaths"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["toolCall"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["artifactDirectoryPath"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["stepIdx"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["invocationNum"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["initialNumSteps"]; ok {
				return EngineAntigravity
			}
			if _, ok := probe["parentConversationId"]; ok {
				return EngineAntigravity
			}

			// Definitive Claude Code keys
			if _, ok := probe["session_id"]; ok {
				return EngineClaude
			}
			if _, ok := probe["transcript_path"]; ok {
				return EngineClaude
			}
			if _, ok := probe["permission_mode"]; ok {
				return EngineClaude
			}
			if _, ok := probe["hook_event_name"]; ok {
				return EngineClaude
			}
			if _, ok := probe["source"]; ok {
				return EngineClaude
			}
			if _, ok := probe["custom_instructions"]; ok {
				return EngineClaude
			}
			if _, ok := probe["tool_name"]; ok {
				return EngineClaude
			}
			if _, ok := probe["tool_input"]; ok {
				return EngineClaude
			}
			if _, ok := probe["tool_response"]; ok {
				return EngineClaude
			}
		}
		if bytes.Contains(raw, []byte(`"conversationId"`)) ||
			bytes.Contains(raw, []byte(`"workspacePaths"`)) ||
			bytes.Contains(raw, []byte(`"toolCall"`)) ||
			bytes.Contains(raw, []byte(`"stepIdx"`)) {
			return EngineAntigravity
		}
		if bytes.Contains(raw, []byte(`"tool_name"`)) ||
			bytes.Contains(raw, []byte(`"tool_input"`)) ||
			bytes.Contains(raw, []byte(`"session_id"`)) ||
			bytes.Contains(raw, []byte(`"hook_event_name"`)) {
			return EngineClaude
		}
		return EngineClaude
	}
	return EngineClaude
}

// ProjectDir resolves the project root directory from environment or payload.
// It never falls back to os.Getwd() to prevent silent location errors.
func ProjectDir(envProjectDir, payloadCWD string, workspacePaths []string, toolTarget string) string {
	if envProjectDir != "" {
		return envProjectDir
	}
	if payloadCWD != "" {
		return payloadCWD
	}
	if len(workspacePaths) > 0 && workspacePaths[0] != "" {
		return workspacePaths[0]
	}
	if toolTarget != "" {
		if root := findGitRoot(toolTarget); root != "" {
			return root
		}
	}
	return ""
}

func findGitRoot(start string) string {
	dir := start
	if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && (fi.IsDir() || !fi.IsDir()) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "." || parent == "/" {
			break
		}
		dir = parent
	}
	return ""
}

// PluginRoot resolves the plugin's root directory.
func PluginRoot(envClaude, envPluginRoot, exePath string) string {
	if envClaude != "" {
		return envClaude
	}
	if envPluginRoot != "" {
		return envPluginRoot
	}
	if exePath != "" {
		grandparent := filepath.Dir(filepath.Dir(exePath))
		if isPluginDir(grandparent) {
			return grandparent
		}
	}
	if exe, err := os.Executable(); err == nil {
		grandparent := filepath.Dir(filepath.Dir(exe))
		if isPluginDir(grandparent) {
			return grandparent
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		if isPluginDir(cwd) {
			return cwd
		}
	}
	return ""
}

func isPluginDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".antigravity-plugin", "plugin.json")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude-plugin", "plugin.json")); err == nil {
		return true
	}
	return false
}

// PluginDataDir resolves the plugin's private persistent state directory.
func PluginDataDir(envClaude, envPluginData, artifactDirPath, pluginName, projectDir string) string {
	if envClaude != "" {
		return envClaude
	}
	if envPluginData != "" {
		return envPluginData
	}
	if artifactDirPath != "" {
		// artifactDirPath is typically <appDataDir>/brain/<conversationId>
		base := filepath.Dir(filepath.Dir(artifactDirPath))
		if pluginName != "" {
			return filepath.Join(base, "plugin-data", pluginName)
		}
		return filepath.Join(base, "plugin-data")
	}
	if projectDir != "" {
		return filepath.Join(projectDir, ".claude")
	}
	return ""
}

// Read decodes standard input into a normalized Request.
func Read(event string, stdin io.Reader) Request {
	raw, _ := io.ReadAll(stdin)
	req := Request{
		Event:  event,
		Raw:    raw,
		Engine: DetectEngine(raw),
	}

	req.PluginRoot = PluginRoot(os.Getenv("CLAUDE_PLUGIN_ROOT"), os.Getenv("PLUGIN_ROOT"), "")
	req.ProjectDir = ProjectDir(os.Getenv("CLAUDE_PROJECT_DIR"), "", nil, "")
	req.PluginDataDir = PluginDataDir(os.Getenv("CLAUDE_PLUGIN_DATA"), os.Getenv("PLUGIN_DATA"), "", filepath.Base(req.PluginRoot), req.ProjectDir)

	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		req.Parsed = false
		req.ParseErr = err
		return req
	}
	req.Parsed = true

	// Envelope fields
	if s, ok := env["session_id"].(string); ok {
		req.SessionID = s
	} else if s, ok := env["conversationId"].(string); ok {
		req.SessionID = s
	}

	if s, ok := env["transcript_path"].(string); ok {
		req.TranscriptPath = s
	} else if s, ok := env["transcriptPath"].(string); ok {
		req.TranscriptPath = s
	}

	if s, ok := env["artifactDirectoryPath"].(string); ok {
		req.ArtifactDirPath = s
	}

	if req.TranscriptPath == "" && req.ArtifactDirPath != "" {
		req.TranscriptPath = filepath.Join(req.ArtifactDirPath, ".system_generated", "logs", "transcript.jsonl")
	}

	// Subagent fields
	if s, ok := env["agent_id"].(string); ok {
		req.AgentID = s
	}
	if s, ok := env["agent_type"].(string); ok {
		req.AgentType = s
	}
	if s, ok := env["agent_transcript_path"].(string); ok {
		req.AgentTranscriptPath = s
	}
	if s, ok := env["agentName"].(string); ok && s != "" && req.AgentType == "" {
		req.AgentType = s
		if req.Engine == EngineAntigravity {
			if req.AgentID == "" {
				req.AgentID = req.SessionID
			}
			if req.AgentTranscriptPath == "" {
				req.AgentTranscriptPath = req.TranscriptPath
			}
		}
	}
	if parent, ok := env["parentConversationId"].(string); ok && parent != "" {
		if req.AgentID == "" {
			req.AgentID = req.SessionID
		}
		if req.AgentTranscriptPath == "" {
			req.AgentTranscriptPath = req.TranscriptPath
		}
		req.SessionID = parent
	}

	// Compact & user input
	if p, ok := env["prompt"].(string); ok {
		req.LastUserInput = p
	} else if u, ok := env["lastUserInput"].(string); ok {
		req.LastUserInput = u
	}

	if tr, ok := env["trigger"].(string); ok {
		req.Trigger = tr
	}
	if ci, ok := env["custom_instructions"].(string); ok {
		req.CustomInstructions = ci
	}
	if cs, ok := env["compact_summary"].(string); ok {
		req.CompactSummary = cs
	}

	if isInt, ok := env["is_interrupt"].(bool); ok {
		req.IsInterrupt = isInt
	}
	if termReason, ok := env["terminationReason"].(string); ok {
		req.TerminationReason = termReason
		if termReason == "USER_CANCELED" {
			req.IsInterrupt = true
		}
	}

	if s, ok := env["hook_event_name"].(string); ok {
		req.HookEventName = s
	}
	if s, ok := env["reason"].(string); ok {
		req.Reason = s
	}
	if raw, ok := env["background_tasks"]; ok && raw != nil {
		if b, err := json.Marshal(raw); err == nil {
			msg := json.RawMessage(b)
			req.BackgroundTasks = &msg
		}
	}
	if raw, ok := env["session_crons"]; ok && raw != nil {
		if b, err := json.Marshal(raw); err == nil {
			msg := json.RawMessage(b)
			req.SessionCrons = &msg
		}
	}
	if s, ok := env["event"].(string); ok {
		req.FileEvent = s
	}

	if stopActive, ok := env["stop_hook_active"].(bool); ok {
		req.StopHookActive = stopActive
	}
	if execNum, ok := env["executionNum"].(float64); ok && execNum > 0 {
		req.StopHookActive = true
	}

	// Tool call extraction & normalization
	var toolInputMap map[string]any
	if tn, ok := env["tool_name"].(string); ok {
		req.RawToolName = tn
		req.ToolName = CanonicalToolName(tn)
		if ti, ok := env["tool_input"].(map[string]any); ok {
			toolInputMap = ti
		}
	} else if tc, ok := env["toolCall"].(map[string]any); ok {
		if tn, ok := tc["name"].(string); ok {
			req.RawToolName = tn
			req.ToolName = CanonicalToolName(tn)
		}
		if args, ok := tc["args"].(map[string]any); ok {
			toolInputMap = args
		}
	}

	// Normalize tool arguments and populate canonical keys
	if toolInputMap != nil {
		req.RawToolInput = toolInputMap
		req.ToolInput = normalizeToolArgs(req.ToolName, toolInputMap, &req)
	}

	// Tool response & errors
	if tr, ok := env["tool_response"]; ok {
		if b, err := json.Marshal(tr); err == nil {
			req.ToolResult = b
		}
	} else if res, ok := env["result"]; ok {
		if b, err := json.Marshal(res); err == nil {
			req.ToolResult = b
		}
	}

	if errVal, ok := env["error"].(string); ok {
		req.ToolError = errVal
	}

	// Path resolutions
	var payloadCWD string
	if cwd, ok := env["cwd"].(string); ok {
		req.CWD = cwd
		payloadCWD = cwd
	} else if toolInputMap != nil {
		if cwd, ok := toolInputMap["Cwd"].(string); ok && cwd != "" {
			req.CWD = cwd
			payloadCWD = cwd
		} else if cwd, ok := toolInputMap["cwd"].(string); ok && cwd != "" {
			req.CWD = cwd
			payloadCWD = cwd
		}
	}
	var workspacePaths []string
	if wsp, ok := env["workspacePaths"].([]any); ok {
		for _, w := range wsp {
			if s, ok := w.(string); ok {
				workspacePaths = append(workspacePaths, s)
			}
		}
	}

	if req.FilePath == "" {
		if fp, ok := env["file_path"].(string); ok && fp != "" {
			req.FilePath = fp
		} else if fp, ok := env["path"].(string); ok && fp != "" {
			req.FilePath = fp
		} else if fp, ok := env["Path"].(string); ok && fp != "" {
			req.FilePath = fp
		}
	}

	toolTarget := req.FilePath
	if toolTarget == "" {
		toolTarget = req.Command
	}
	req.ProjectDir = ProjectDir(os.Getenv("CLAUDE_PROJECT_DIR"), payloadCWD, workspacePaths, toolTarget)
	req.PluginRoot = PluginRoot(os.Getenv("CLAUDE_PLUGIN_ROOT"), os.Getenv("PLUGIN_ROOT"), "")
	req.PluginDataDir = PluginDataDir(os.Getenv("CLAUDE_PLUGIN_DATA"), os.Getenv("PLUGIN_DATA"), req.ArtifactDirPath, filepath.Base(req.PluginRoot), req.ProjectDir)

	return req
}

func normalizeToolArgs(canonicalTool string, in map[string]any, req *Request) json.RawMessage {
	norm := make(map[string]any, len(in)+4)
	for k, v := range in {
		norm[k] = v
	}

	// Command extraction
	if cmd, ok := in["command"].(string); ok {
		req.Command = cmd
		norm["CommandLine"] = cmd
	} else if cmd, ok := in["CommandLine"].(string); ok {
		req.Command = cmd
		norm["command"] = cmd
	}

	if c, ok := in["Cwd"].(string); ok {
		norm["cwd"] = c
	} else if c, ok := in["cwd"].(string); ok {
		norm["Cwd"] = c
	}

	if a, ok := in["toolAction"].(string); ok {
		norm["description"] = a
	} else if s, ok := in["toolSummary"].(string); ok {
		norm["description"] = s
	} else if d, ok := in["description"].(string); ok {
		norm["toolAction"] = d
		norm["toolSummary"] = d
	}

	// File path extraction
	var path string
	for _, k := range []string{"file_path", "Path", "TargetFile", "AbsolutePath", "path"} {
		if s, ok := in[k].(string); ok && s != "" {
			path = s
			break
		}
	}
	if path != "" {
		req.FilePath = path
		norm["file_path"] = path
		norm["Path"] = path
		norm["TargetFile"] = path
		norm["AbsolutePath"] = path
	}

	// Content extraction
	if c, ok := in["content"].(string); ok {
		norm["CodeContent"] = c
		norm["Content"] = c
	} else if c, ok := in["CodeContent"].(string); ok {
		norm["content"] = c
	} else if c, ok := in["Content"].(string); ok {
		norm["content"] = c
	}

	// Edit strings
	if s, ok := in["old_string"].(string); ok {
		norm["TargetContent"] = s
		norm["Target"] = s
	} else if s, ok := in["TargetContent"].(string); ok {
		norm["old_string"] = s
	} else if s, ok := in["Target"].(string); ok {
		norm["old_string"] = s
	}

	if s, ok := in["new_string"].(string); ok {
		norm["ReplacementContent"] = s
		norm["Replacement"] = s
	} else if s, ok := in["ReplacementContent"].(string); ok {
		norm["new_string"] = s
	} else if s, ok := in["Replacement"].(string); ok {
		norm["new_string"] = s
	}

	if am, ok := in["AllowMultiple"].(bool); ok {
		norm["replace_all"] = am
	} else if ra, ok := in["replace_all"].(bool); ok {
		norm["AllowMultiple"] = ra
	}

	// Queries & URLs
	if q, ok := in["query"].(string); ok {
		norm["Query"] = q
	} else if q, ok := in["Query"].(string); ok {
		norm["query"] = q
	}

	if u, ok := in["url"].(string); ok {
		norm["Url"] = u
	} else if u, ok := in["Url"].(string); ok {
		norm["url"] = u
	}

	b, _ := json.Marshal(norm)
	return b
}

type pendingItem struct {
	Feedback          string `json:"feedback,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
	Say               string `json:"say,omitempty"`
}

// pendingQueueKey identifies which Antigravity conversation owns a queued PreInvocation step.
// Subagent hooks promote parentConversationId to req.SessionID while storing the child's own
// conversationId in req.AgentID; keying on AgentID keeps concurrent subagents isolated rather
// than colliding in pending_<parent>.jsonl.
func pendingQueueKey(req Request) string {
	if req.AgentID != "" {
		return req.AgentID
	}
	return req.SessionID
}

func queuePending(pluginDataDir, sessionID string, item pendingItem) {
	if pluginDataDir == "" || sessionID == "" {
		return
	}
	_ = os.MkdirAll(pluginDataDir, 0755)
	queueFile := filepath.Join(pluginDataDir, "pending_"+sessionID+".jsonl")
	f, err := os.OpenFile(queueFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	b, _ := json.Marshal(item)
	_, _ = f.Write(append(b, '\n'))
}

// QueueSay enqueues an advisory or failure message for delivery on the next Antigravity PreInvocation event.
func QueueSay(pluginDataDir, sessionID, say string) {
	if say == "" {
		return
	}
	queuePending(pluginDataDir, sessionID, pendingItem{Say: say})
}

func flushPending(pluginDataDir, sessionID string) []pendingItem {
	if pluginDataDir == "" || sessionID == "" {
		return nil
	}
	queueFile := filepath.Join(pluginDataDir, "pending_"+sessionID+".jsonl")
	b, err := os.ReadFile(queueFile)
	if err != nil {
		return nil
	}
	_ = os.Remove(queueFile)

	var items []pendingItem
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var it pendingItem
		if err := json.Unmarshal(line, &it); err == nil {
			items = append(items, it)
		}
	}
	return items
}

// Respond emits the wire format for the request's event and engine, returning the process exit code.
func Respond(req Request, resp Response, stdout, stderr io.Writer, rec Settler) int {
	if rec != nil {
		defer rec.Persist()
		settled := rec.Settle()
		if settled != "" {
			if resp.Say != "" {
				resp.Say = resp.Say + "\n\n" + settled
			} else {
				resp.Say = settled
			}
		}
	}

	if resp.Warnings != "" {
		fmt.Fprintln(stderr, strings.TrimRight(resp.Warnings, "\n"))
	}

	if !req.Parsed && len(req.Raw) > 0 {
		errStr := "syntax error"
		if req.ParseErr != nil {
			errStr = req.ParseErr.Error()
		}
		fmt.Fprintf(stderr, "[hookcore] unparseable payload for event %q (%d bytes): %s\n", req.Event, len(req.Raw), errStr)
	}

	if req.Engine == EngineAntigravity {
		return respondAntigravity(req, resp, stdout, stderr)
	}
	return respondClaude(req, resp, stdout, stderr)
}

func respondClaude(req Request, resp Response, stdout, stderr io.Writer) int {
	switch req.Event {
	case "PreToolUse":
		if resp.Decision == DecisionNone && resp.UpdatedCommand == "" && len(resp.UpdatedInput) == 0 && resp.Say == "" {
			return 0
		}
		if resp.Decision == DecisionNone && resp.UpdatedCommand == "" && len(resp.UpdatedInput) == 0 && resp.Say != "" {
			out := map[string]any{"systemMessage": resp.Say}
			b, _ := json.Marshal(out)
			fmt.Fprintln(stdout, string(b))
			return 0
		}
		var out struct {
			HookSpecificOutput struct {
				HookEventName            string         `json:"hookEventName"`
				PermissionDecision       string         `json:"permissionDecision,omitempty"`
				PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"`
				UpdatedInput             map[string]any `json:"updatedInput,omitempty"`
			} `json:"hookSpecificOutput"`
			SystemMessage string `json:"systemMessage,omitempty"`
		}
		out.HookSpecificOutput.HookEventName = "PreToolUse"
		if resp.Decision != DecisionNone {
			out.HookSpecificOutput.PermissionDecision = string(resp.Decision)
			out.HookSpecificOutput.PermissionDecisionReason = resp.Reason
		}
		if len(resp.UpdatedInput) > 0 {
			out.HookSpecificOutput.UpdatedInput = resp.UpdatedInput
		} else if resp.UpdatedCommand != "" {
			updated := make(map[string]any, len(req.RawToolInput)+1)
			for k, v := range req.RawToolInput {
				updated[k] = v
			}
			updated["command"] = resp.UpdatedCommand
			out.HookSpecificOutput.UpdatedInput = updated
		}
		if len(out.HookSpecificOutput.UpdatedInput) > 0 && out.HookSpecificOutput.PermissionDecision == "" {
			out.HookSpecificOutput.PermissionDecision = string(DecisionAllow)
		}
		out.SystemMessage = resp.Say
		b, _ := json.Marshal(out)
		fmt.Fprintln(stdout, string(b))
		return 0

	case "PostToolUse", "PostToolUseFailure":
		if resp.Feedback != "" {
			fmt.Fprintln(stderr, resp.Feedback)
			if resp.Say != "" {
				out := map[string]any{"systemMessage": resp.Say}
				b, _ := json.Marshal(out)
				fmt.Fprintln(stdout, string(b))
			}
			return 2
		}
		if resp.AdditionalContext != "" || resp.Say != "" {
			out := map[string]any{}
			if resp.AdditionalContext != "" {
				out["hookSpecificOutput"] = map[string]any{
					"hookEventName":     req.Event,
					"additionalContext": resp.AdditionalContext,
				}
			}
			if resp.Say != "" {
				out["systemMessage"] = resp.Say
			}
			b, _ := json.Marshal(out)
			fmt.Fprintln(stdout, string(b))
		}
		return 0

	case "SessionStart":
		if resp.AdditionalContext == "" && len(resp.WatchPaths) == 0 && resp.Say == "" {
			return 0
		}
		out := map[string]any{}
		if resp.AdditionalContext != "" || len(resp.WatchPaths) > 0 {
			hso := map[string]any{
				"hookEventName": "SessionStart",
			}
			if resp.AdditionalContext != "" {
				hso["additionalContext"] = resp.AdditionalContext
			}
			if len(resp.WatchPaths) > 0 {
				hso["watchPaths"] = resp.WatchPaths
			}
			out["hookSpecificOutput"] = hso
		}
		if resp.Say != "" {
			out["systemMessage"] = resp.Say
		}
		b, _ := json.Marshal(out)
		fmt.Fprintln(stdout, string(b))
		return 0

	case "Stop":
		if resp.AdditionalContext == "" && resp.Say == "" {
			return 0
		}
		out := map[string]any{}
		if resp.AdditionalContext != "" {
			out["hookSpecificOutput"] = map[string]any{
				"hookEventName":     "Stop",
				"additionalContext": resp.AdditionalContext,
			}
		}
		if resp.Say != "" {
			out["systemMessage"] = resp.Say
		}
		b, _ := json.Marshal(out)
		fmt.Fprintln(stdout, string(b))
		return 0

	case "PreCompact":
		// On Claude Code, PreCompact's stdout becomes the custom compact instructions (plain text).
		if resp.PreCompactSteer != "" {
			fmt.Fprintln(stdout, resp.PreCompactSteer)
		} else if resp.AdditionalContext != "" {
			fmt.Fprintln(stdout, resp.AdditionalContext)
		}
		return 0

	case "SubagentStop", "SessionEnd", "FileChanged", "PostCompact":
		// SubagentStop, SessionEnd, FileChanged, and PostCompact are silent on stdout.
		return 0

	default:
		if resp.AdditionalContext != "" || resp.Say != "" {
			out := map[string]any{}
			if resp.AdditionalContext != "" {
				out["hookSpecificOutput"] = map[string]any{
					"hookEventName":     req.Event,
					"additionalContext": resp.AdditionalContext,
				}
			}
			if resp.Say != "" {
				out["systemMessage"] = resp.Say
			}
			b, _ := json.Marshal(out)
			fmt.Fprintln(stdout, string(b))
		}
		return 0
	}
}

func respondAntigravity(req Request, resp Response, stdout, stderr io.Writer) int {
	switch req.Event {
	case "PreTool", "PreToolUse":
		var out struct {
			Decision   string         `json:"decision,omitempty"`
			Reason     string         `json:"reason,omitempty"`
			DenyReason string         `json:"deny_reason,omitempty"`
			Overwrite  map[string]any `json:"overwrite,omitempty"`
		}
		if resp.Decision != DecisionNone {
			out.Decision = string(resp.Decision)
			out.Reason = resp.Reason
			if resp.Decision == DecisionDeny {
				out.DenyReason = resp.Reason
			}
		}
		if len(resp.UpdatedInput) > 0 {
			out.Overwrite = resp.UpdatedInput
		} else if resp.UpdatedCommand != "" {
			out.Overwrite = map[string]any{
				"CommandLine": resp.UpdatedCommand,
			}
		}
		if len(out.Overwrite) > 0 && out.Decision == "" {
			out.Decision = string(DecisionAllow)
		}
		if resp.Say != "" {
			queuePending(req.PluginDataDir, pendingQueueKey(req), pendingItem{Say: resp.Say})
		}
		if resp.Decision == DecisionNone && resp.UpdatedCommand == "" && len(resp.UpdatedInput) == 0 {
			return 0
		}
		b, _ := json.Marshal(out)
		fmt.Fprintln(stdout, string(b))
		return 0

	case "PostTool", "PostToolUse", "PostToolUseFailure":
		if resp.Feedback != "" || resp.AdditionalContext != "" || resp.Say != "" {
			queuePending(req.PluginDataDir, pendingQueueKey(req), pendingItem{
				Feedback:          resp.Feedback,
				AdditionalContext: resp.AdditionalContext,
				Say:               resp.Say,
			})
		}
		fmt.Fprintln(stdout, "{}")
		return 0

	case "PreInvocation", "SessionStart":
		pending := flushPending(req.PluginDataDir, pendingQueueKey(req))
		var injectSteps []map[string]any

		for _, p := range pending {
			if p.Feedback != "" {
				injectSteps = append(injectSteps, map[string]any{
					"ephemeralMessage": p.Feedback,
				})
			}
			if p.AdditionalContext != "" {
				injectSteps = append(injectSteps, map[string]any{
					"ephemeralMessage": p.AdditionalContext,
				})
			}
			if p.Say != "" {
				injectSteps = append(injectSteps, map[string]any{
					"systemMessage": map[string]any{
						"systemMessage": p.Say,
					},
				})
			}
		}

		if resp.Feedback != "" {
			injectSteps = append(injectSteps, map[string]any{
				"ephemeralMessage": resp.Feedback,
			})
		}
		if resp.AdditionalContext != "" {
			injectSteps = append(injectSteps, map[string]any{
				"ephemeralMessage": resp.AdditionalContext,
			})
		}
		if resp.Say != "" {
			injectSteps = append(injectSteps, map[string]any{
				"systemMessage": map[string]any{
					"systemMessage": resp.Say,
				},
			})
		}

		if req.Event == "SessionStart" && len(injectSteps) == 0 {
			return 0
		}

		out := map[string]any{}
		if len(injectSteps) > 0 {
			out["injectSteps"] = injectSteps
		}
		// PreInvocation accepts only "terminate" or an absent terminationBehavior;
		// the CLI fails the whole turn on any other value ("force_continue" is
		// PostInvocation's).
		if req.Event == "PreInvocation" && resp.TerminateTurn {
			if resp.Reason != "" {
				out["injectSteps"] = append(injectSteps, map[string]any{
					"errorMessage": map[string]any{
						"modelMessage": resp.Reason,
						"userMessage":  resp.Reason,
					},
				})
			}
			out["terminationBehavior"] = "terminate"
		}

		b, _ := json.Marshal(out)
		fmt.Fprintln(stdout, string(b))
		return 0

	case "PostInvocation", "Stop":
		if resp.AdditionalContext != "" {
			out := map[string]any{
				"decision": "continue",
				"reason":   resp.AdditionalContext,
			}
			b, _ := json.Marshal(out)
			fmt.Fprintln(stdout, string(b))
		}
		return 0

	case "PreCompact":
		text := resp.PreCompactSteer
		if text == "" {
			text = resp.AdditionalContext
		}
		if text != "" {
			out := map[string]any{
				"injectSteps": []map[string]any{
					{
						"systemMessage": map[string]any{
							"systemMessage": text,
						},
					},
				},
			}
			b, _ := json.Marshal(out)
			fmt.Fprintln(stdout, string(b))
		} else {
			fmt.Fprintln(stdout, "{}")
		}
		return 0

	case "SubagentStop", "SessionEnd", "FileChanged", "PostCompact":
		fmt.Fprintln(stdout, "{}")
		return 0

	default:
		fmt.Fprintln(stdout, "{}")
		return 0
	}
}
