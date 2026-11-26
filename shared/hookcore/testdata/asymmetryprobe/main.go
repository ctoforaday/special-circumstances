// asymmetryprobe logs raw and normalized hook payloads across Claude Code and Antigravity so
// live_asymmetry*_test.go can pin both positive capabilities and negative assumptions on each
// harness. When an upstream harness fixes a bug or adds a capability we worked around, the
// corresponding asymmetry test fails loudly.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

type probeEvent struct {
	Event                 string            `json:"event"`
	HookEventName         string            `json:"hook_event_name"`
	Engine                string            `json:"engine"`
	SessionID             string            `json:"session_id"`
	AgentID               string            `json:"agent_id"`
	AgentType             string            `json:"agent_type"`
	ToolName              string            `json:"tool_name"`
	RawToolName           string            `json:"raw_tool_name"`
	Command               string            `json:"command"`
	Source                string            `json:"source"`
	Error                 string            `json:"error"`
	HasResultField        bool              `json:"has_result_field"`
	InvocationNum         *int              `json:"invocation_num,omitempty"`
	EnvNames              []string          `json:"env_names,omitempty"`
	EnvValues             map[string]string `json:"env_values,omitempty"`
	ClaudeSessionFiles    []string          `json:"claude_session_files,omitempty"`
	ClaudeSessionPayloads map[string]string `json:"claude_session_payloads,omitempty"`
	CWD                   string            `json:"cwd,omitempty"`
	Raw                   string            `json:"raw"`
}

func main() {
	evFlag := flag.String("event", "", "hook event name")
	flag.Parse()

	req := hookcore.Read(*evFlag, os.Stdin)

	var rawMap map[string]any
	_ = json.Unmarshal(req.Raw, &rawMap)
	hasResult := false
	for _, k := range []string{"result", "tool_response", "toolResponse"} {
		if v, ok := rawMap[k]; ok && v != nil {
			if s, ok := v.(string); !ok || s != "" {
				hasResult = true
				break
			}
		}
	}
	errStr, _ := rawMap["error"].(string)
	srcStr, _ := rawMap["source"].(string)
	var invNum *int
	if n, ok := rawMap["invocationNum"].(float64); ok {
		v := int(n)
		invNum = &v
	}

	var envNames []string
	envValues := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(k, "ANTIGRAVITY_") || strings.HasPrefix(k, "CLAUDE_") || strings.HasPrefix(k, "PLUGIN_") {
			envNames = append(envNames, k)
		}
		if k == "CLAUDE_PLUGIN_ROOT" || k == "PLUGIN_ROOT" {
			envValues[k] = v
		}
	}
	sort.Strings(envNames)
	if len(envValues) == 0 {
		envValues = nil
	}

	var sessFiles []string
	var sessPayloads map[string]string
	if home := os.Getenv("HOME"); home != "" {
		sessDir := filepath.Join(home, ".claude", "sessions")
		if entries, err := os.ReadDir(sessDir); err == nil {
			sessPayloads = make(map[string]string, len(entries))
			for _, e := range entries {
				sessFiles = append(sessFiles, e.Name())
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					if data, rerr := os.ReadFile(filepath.Join(sessDir, e.Name())); rerr == nil {
						sessPayloads[e.Name()] = string(data)
					}
				}
			}
			sort.Strings(sessFiles)
		}
	}

	wd, _ := os.Getwd()
	ev := probeEvent{
		Event:                 *evFlag,
		HookEventName:         req.HookEventName,
		Engine:                string(req.Engine),
		SessionID:             req.SessionID,
		AgentID:               req.AgentID,
		AgentType:             req.AgentType,
		ToolName:              req.ToolName,
		RawToolName:           req.RawToolName,
		Command:               req.Command,
		Source:                srcStr,
		Error:                 errStr,
		HasResultField:        hasResult,
		InvocationNum:         invNum,
		EnvNames:              envNames,
		EnvValues:             envValues,
		ClaudeSessionFiles:    sessFiles,
		ClaudeSessionPayloads: sessPayloads,
		CWD:                   wd,
		Raw:                   string(req.Raw),
	}
	if logPath := os.Getenv("ASYMMETRY_LOG"); logPath != "" {
		if b, err := json.Marshal(ev); err == nil {
			if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
				_, _ = f.Write(append(b, '\n'))
				_ = f.Close()
			}
		}
	}

	// Generated large-payload helper so tests can exercise delivery-size caps without stuffing
	// megabytes into environment variables.
	if bigStr := os.Getenv("ASYMMETRY_EMIT_BIG_LEN"); bigStr != "" {
		targetEv := os.Getenv("ASYMMETRY_EMIT_EVENT")
		if targetEv == "" || targetEv == *evFlag {
			if n, err := strconv.Atoi(bigStr); err == nil && n > 0 {
				marker := os.Getenv("ASYMMETRY_EMIT_BIG_MARKER")
				if marker == "" {
					marker = "BIG_TAIL_MARKER_9911"
				}
				headMarker := "BIG_HEAD_9911"
				midMarker := "BIG_MID_9911"
				padTotal := max(n-len(headMarker)-len(midMarker)-len(marker)-4, 0)
				half := padTotal / 2
				payload := headMarker + " " + strings.Repeat("a", half) + " " + midMarker + " " + strings.Repeat("b", padTotal-half) + " " + marker
				var out []byte
				if req.Engine == hookcore.EngineAntigravity {
					out, _ = json.Marshal(map[string]any{
						"injectSteps": []any{
							map[string]any{
								"systemMessage": map[string]any{
									"systemMessage": payload,
								},
							},
						},
					})
				} else {
					out, _ = json.Marshal(map[string]any{
						"hookSpecificOutput": map[string]any{
							"hookEventName":     *evFlag,
							"additionalContext": payload,
						},
					})
				}
				_, _ = fmt.Fprintln(os.Stdout, string(out))
				os.Exit(0)
			}
		}
	}

	// Raw response override for tests that intentionally emit non-schema or edge-case hook wire JSON.
	if rawOut := os.Getenv("ASYMMETRY_EMIT_RAW"); rawOut != "" {
		targetEv := os.Getenv("ASYMMETRY_EMIT_EVENT")
		targetSrc := os.Getenv("ASYMMETRY_EMIT_SOURCE")
		if (targetEv == "" || targetEv == *evFlag) && (targetSrc == "" || targetSrc == srcStr) {
			_, _ = fmt.Fprintln(os.Stdout, rawOut)
			os.Exit(0)
		}
	}
	if rawOut2 := os.Getenv("ASYMMETRY_EMIT_RAW2"); rawOut2 != "" {
		targetEv2 := os.Getenv("ASYMMETRY_EMIT_EVENT2")
		if targetEv2 == "" || targetEv2 == *evFlag {
			_, _ = fmt.Fprintln(os.Stdout, rawOut2)
			os.Exit(0)
		}
	}

	resp := hookcore.Response{}
	if (*evFlag == "PreToolUse" || *evFlag == "PreTool") && os.Getenv("ASYMMETRY_DENY_REASON") != "" {
		resp.Decision = hookcore.DecisionDeny
		resp.Reason = os.Getenv("ASYMMETRY_DENY_REASON")
	} else if (*evFlag == "PreToolUse" || *evFlag == "PreTool") && os.Getenv("ASYMMETRY_OVERWRITE_FINISH") == "1" {
		switch req.RawToolName {
		case "finish":
			resp.Decision = hookcore.DecisionAllow
			resp.UpdatedInput = map[string]any{
				"result": map[string]any{
					"answer":          "OVERWRITTEN_BY_PRETOOLUSE",
					"stamped_by_hook": "OVERWRITTEN_BY_PRETOOLUSE",
				},
			}
		case "StructuredOutput":
			resp.Decision = hookcore.DecisionAllow
			resp.UpdatedInput = map[string]any{
				"answer":          "OVERWRITTEN_BY_PRETOOLUSE",
				"stamped_by_hook": "OVERWRITTEN_BY_PRETOOLUSE",
			}
		default:
			resp.Decision = hookcore.DecisionAllow
		}
	} else if *evFlag == "PreToolUse" || *evFlag == "PreTool" {
		resp.Decision = hookcore.DecisionAllow
	}
	os.Exit(hookcore.Respond(req, resp, os.Stdout, os.Stderr, nil))
}
