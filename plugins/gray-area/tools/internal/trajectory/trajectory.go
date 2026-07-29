// Package trajectory reads Claude Code session transcripts into citable events.
//
// Contract (Design by Contract):
//
//	Every event MUST carry its provenance — file, 1-based line, and uuid — because
//	a finding cites the trajectory, never the reader's opinion of it.
//	A line that does not parse MUST be counted and reported, never silently
//	skipped: a reader that quietly drops input answers confidently from a subset.
//
// The transcript format is vendor-internal and version-unstable. This package
// reads only the fields it needs and treats everything else as opaque, so a
// shape it does not recognise degrades to "no tool calls on this line" rather
// than to a crash or a wrong answer.
package trajectory

import (
	"bufio"
	"encoding/json"
	"io"
)

// Event is one citable act. Provenance is not optional: File, Line and UUID are
// what a consumer quotes when a finding rests on this event.
type Event struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	UUID       string `json:"uuid"`
	ParentUUID string `json:"parent_uuid,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`

	Kind      string `json:"kind"` // "tool_use" | "tool_result" | "user_text" | "assistant_text"
	AgentType string `json:"agent_type,omitempty"`

	ToolName  string          `json:"tool_name,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
}

// Cited reports whether the event can back a finding. An event without
// provenance is an opinion, and callers should refuse to emit it.
func (e Event) Cited() bool { return e.File != "" && e.Line > 0 && e.UUID != "" }

// Result is the parse outcome. BadLines is surfaced rather than swallowed: a
// non-zero count means the reader saw input it could not account for, and any
// answer derived from this parse is over a subset of the record.
type Result struct {
	Events    []Event
	Lines     int
	BadLines  int
	BadSample []int // 1-based line numbers, capped
}

const badSampleCap = 10

type rawLine struct {
	UUID        string `json:"uuid"`
	ParentUUID  string `json:"parentUuid"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
	Attribution string `json:"attributionAgent"`
	Message     *struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type contentBlock struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	Input     json.RawMessage `json:"input"`
}

// Parse reads one transcript. file is recorded on every event as its provenance,
// so it should be the path the caller would quote — not a temporary copy.
func Parse(r io.Reader, file string) (Result, error) {
	var res Result
	sc := bufio.NewScanner(r)
	// Transcript lines carry whole tool inputs and results; the default 64KB
	// token limit truncates them, which would silently corrupt the parse.
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)

	for sc.Scan() {
		res.Lines++
		text := sc.Bytes()
		if len(text) == 0 {
			continue
		}
		var rl rawLine
		if err := json.Unmarshal(text, &rl); err != nil {
			res.BadLines++
			if len(res.BadSample) < badSampleCap {
				res.BadSample = append(res.BadSample, res.Lines)
			}
			continue
		}
		if rl.Message == nil || len(rl.Message.Content) == 0 {
			continue
		}
		var blocks []contentBlock
		if err := json.Unmarshal(rl.Message.Content, &blocks); err != nil {
			// A string content body is normal for user lines, not a parse failure.
			continue
		}
		for _, b := range blocks {
			ev := Event{
				File: file, Line: res.Lines, UUID: rl.UUID, ParentUUID: rl.ParentUUID,
				Timestamp: rl.Timestamp, AgentType: rl.Attribution,
			}
			switch b.Type {
			case "tool_use":
				ev.Kind = "tool_use"
				ev.ToolName = b.Name
				ev.ToolUseID = b.ID
				ev.ToolInput = b.Input
			case "tool_result":
				ev.Kind = "tool_result"
				ev.ToolUseID = b.ToolUseID
			default:
				continue
			}
			res.Events = append(res.Events, ev)
		}
	}
	if err := sc.Err(); err != nil {
		return res, err
	}
	return res, nil
}

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
