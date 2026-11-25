// asymmetryprobe logs raw and normalized hook payloads across Claude Code and Antigravity so
// live_asymmetry_test.go can pin both positive capabilities and negative assumptions on each
// harness. When an upstream harness fixes a bug or adds a capability we worked around, the
// corresponding asymmetry test fails loudly.
package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

type probeEvent struct {
	Event          string `json:"event"`
	HookEventName  string `json:"hook_event_name"`
	Engine         string `json:"engine"`
	SessionID      string `json:"session_id"`
	AgentID        string `json:"agent_id"`
	AgentType      string `json:"agent_type"`
	ToolName       string `json:"tool_name"`
	RawToolName    string `json:"raw_tool_name"`
	Command        string `json:"command"`
	Source         string `json:"source"`
	Error          string `json:"error"`
	HasResultField bool   `json:"has_result_field"`
	Raw            string `json:"raw"`
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

	ev := probeEvent{
		Event:          *evFlag,
		HookEventName:  req.HookEventName,
		Engine:         string(req.Engine),
		SessionID:      req.SessionID,
		AgentID:        req.AgentID,
		AgentType:      req.AgentType,
		ToolName:       req.ToolName,
		RawToolName:    req.RawToolName,
		Command:        req.Command,
		Source:         srcStr,
		Error:          errStr,
		HasResultField: hasResult,
		Raw:            string(req.Raw),
	}
	if logPath := os.Getenv("ASYMMETRY_LOG"); logPath != "" {
		if b, err := json.Marshal(ev); err == nil {
			if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
				_, _ = f.Write(append(b, '\n'))
				_ = f.Close()
			}
		}
	}

	resp := hookcore.Response{}
	if (*evFlag == "PreToolUse" || *evFlag == "PreTool") && os.Getenv("ASYMMETRY_OVERWRITE_FINISH") == "1" {
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
