// Package trajectory provides citable events and cross-platform streaming parsers
// for both Anthropic Claude Code and Google Antigravity session transcripts.
//
// Contract (Design by Contract):
//
//	Every event MUST carry its provenance — file, 1-based line, and uuid — because
//	a finding cites the trajectory, never the reader's opinion of it.
//	A line that does not parse MUST be counted and reported, never silently
//	skipped: a reader that quietly drops input answers confidently from a subset.
package trajectory

import (
	"encoding/json"
)

// Event is one citable act or message in a session trajectory.
type Event struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	UUID       string `json:"uuid"`
	ParentUUID string `json:"parent_uuid,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`

	Kind      string `json:"kind"` // "tool_use" | "tool_result" | "user_text" | "assistant_text"
	AgentType string `json:"agent_type,omitempty"`

	ToolName    string          `json:"tool_name,omitempty"`     // Canonical tool name: "Bash", "Write", "Edit", "Read", etc.
	RawToolName string          `json:"raw_tool_name,omitempty"` // Raw wire name: "run_command", etc.
	ToolUseID   string          `json:"tool_use_id,omitempty"`
	ToolInput   json.RawMessage `json:"tool_input,omitempty"`
	ToolResult  json.RawMessage `json:"tool_result,omitempty"`
	IsError     bool            `json:"is_error,omitempty"`
	Text        string          `json:"text,omitempty"`
	Thinking    string          `json:"thinking,omitempty"`
}

// Cited reports whether the event can back a finding. An event without
// provenance is an opinion, and callers should refuse to emit it.
func (e Event) Cited() bool {
	return e.File != "" && e.Line > 0 && e.UUID != ""
}

// Result is the parse outcome across a whole transcript.
type Result struct {
	Events    []Event
	Lines     int
	BadLines  int
	BadSample []int // 1-based line numbers, capped at 10
}

const badSampleCap = 10

// ToolUses filters to invocations.
func (r Result) ToolUses() []Event {
	var out []Event
	for _, e := range r.Events {
		if e.Kind == "tool_use" {
			out = append(out, e)
		}
	}
	return out
}
