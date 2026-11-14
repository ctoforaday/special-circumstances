package main

import (
	"encoding/json"
	"flag"
	"maps"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

type probeRecord struct {
	Event               string         `json:"event"`
	Engine              string         `json:"engine"`
	SessionID           string         `json:"session_id"`
	TranscriptPath      string         `json:"transcript_path"`
	AgentID             string         `json:"agent_id"`
	AgentType           string         `json:"agent_type"`
	AgentTranscriptPath string         `json:"agent_transcript_path"`
	ToolName            string         `json:"tool_name"`
	RawToolName         string         `json:"raw_tool_name"`
	Command             string         `json:"command"`
	RawToolInput        map[string]any `json:"raw_tool_input,omitempty"`
}

func appendLog(path string, rec probeRecord) {
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	b, _ := json.Marshal(rec)
	_, _ = f.Write(append(b, '\n'))
}

func main() {
	eventFlag := flag.String("event", "", "hook event name")
	flag.Parse()

	req := hookcore.Read(*eventFlag, os.Stdin)
	appendLog(os.Getenv("SUBAGENTPROBE_LOG"), probeRecord{
		Event:               req.Event,
		Engine:              string(req.Engine),
		SessionID:           req.SessionID,
		TranscriptPath:      req.TranscriptPath,
		AgentID:             req.AgentID,
		AgentType:           req.AgentType,
		AgentTranscriptPath: req.AgentTranscriptPath,
		ToolName:            req.ToolName,
		RawToolName:         req.RawToolName,
		Command:             req.Command,
		RawToolInput:        req.RawToolInput,
	})

	var resp hookcore.Response
	switch req.Event {
	case "SessionStart":
		if req.AgentID != "" {
			resp.Say = "CHILD_SAY_OK"
		}
	case "SubagentStart":
		resp.AdditionalContext = "CHILD_SAY_OK"
	case "PreToolUse", "PreTool":
		switch req.RawToolName {
		case "invoke_subagent":
			updated := make(map[string]any, len(req.RawToolInput)+1)
			maps.Copy(updated, req.RawToolInput)
			updated["Model"] = "flash_lite"
			resp.UpdatedInput = updated
			resp.Say = "HOOK_MODEL_OVERWRITE_AGY"
		case "Agent":
			updated := make(map[string]any, len(req.RawToolInput)+1)
			maps.Copy(updated, req.RawToolInput)
			updated["model"] = "haiku"
			resp.UpdatedInput = updated
			resp.Say = "HOOK_MODEL_OVERWRITE_CLAUDE"
		}
	case "PostToolUse", "PostTool":
		if strings.Contains(req.Command, "CHILD_FEEDBACK_CANARY") {
			resp.Feedback = "CHILD_QUEUE_OK"
		}
	}

	os.Exit(hookcore.Respond(req, resp, os.Stdout, os.Stderr, nil))
}
