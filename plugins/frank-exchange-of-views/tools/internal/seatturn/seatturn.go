// Package seatturn reads a seat's transcript into per-turn rows: what each API round cost and
// what kind of turn it was.
//
// # What this is for
//
// Three separate readers scan these transcripts today and each answers a different slice of one
// question — cost.md sums tokens and price, the dashboard sums them again for its header, and the
// #684 F11 wall-clock decomposition was done by hand, once, in a session that is now gone. None
// of them can answer the next question without another scan.
//
// So the turn is read ONCE into rows, and everything else becomes a view over them (#684 F16).
// This package does the reading and nothing else: no aggregation, no pricing, no judgement about
// which turns matter. A row here is what the transcript said, and the questions are asked in SQL
// where they can be re-asked.
package seatturn

import (
	"encoding/json"
	"strings"
	"time"
)

// Turn is one API round: the usage it reported and the shape of what came back.
type Turn struct {
	Index         int   // position among the usage-bearing lines, from 0
	TSMillis      int64 // the line's own timestamp; 0 when it carried none
	Model         string
	Input         int
	Output        int
	CacheRead     int
	CacheCreation int
	// Thinking and Tool are read off the CONTENT BLOCKS, not inferred from the numbers. The F11
	// decomposition inferred a thinking turn from a low output_tokens with a long span, which is
	// a correlation that happens to hold; the block is the fact, and it is right there.
	Thinking bool
	Tool     bool
}

// Parse reads a transcript into its turns and reports the agent it belongs to.
//
// THE AGENT ID COMES OFF THE LINE, NOT THE FILENAME. Every record in these files carries
// `agentId`, and the file is also called agent-<id>.jsonl — recovering it from the name would be
// a fact composed into a path at one end and matched out at the other, whose miss (a renamed or
// copied file) is silent and produces rows filed against nothing. The field is free.
//
// A line that does not parse, or carries no usage, is skipped: transcripts interleave user
// records, tool results and meta lines, and only the assistant's API rounds report usage.
func Parse(transcript string) (agentID string, turns []Turn) {
	for _, line := range strings.Split(transcript, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec struct {
			AgentID   string `json:"agentId"`
			Timestamp string `json:"timestamp"`
			Message   *struct {
				Model   string `json:"model"`
				Content []struct {
					Type string `json:"type"`
				} `json:"content"`
				Usage *struct {
					Input         int `json:"input_tokens"`
					Output        int `json:"output_tokens"`
					CacheRead     int `json:"cache_read_input_tokens"`
					CacheCreation int `json:"cache_creation_input_tokens"`
				} `json:"usage"`
			} `json:"message"`
		}
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if agentID == "" {
			agentID = rec.AgentID
		}
		if rec.Message == nil || rec.Message.Usage == nil {
			continue
		}
		t := Turn{
			Index:         len(turns),
			TSMillis:      millis(rec.Timestamp),
			Model:         rec.Message.Model,
			Input:         rec.Message.Usage.Input,
			Output:        rec.Message.Usage.Output,
			CacheRead:     rec.Message.Usage.CacheRead,
			CacheCreation: rec.Message.Usage.CacheCreation,
		}
		for _, b := range rec.Message.Content {
			switch b.Type {
			case "thinking":
				t.Thinking = true
			case "tool_use":
				t.Tool = true
			}
		}
		turns = append(turns, t)
	}
	return agentID, turns
}

// millis is the line's timestamp in unix milliseconds, or 0 when it is absent or unparseable.
//
// ZERO IS A STATED ABSENCE, not a date. A turn with no timestamp cannot contribute to a wall-clock
// figure, and the views that compute spans exclude it rather than treating it as the epoch — which
// would put every such seat's start in 1970 and make its duration the age of the universe.
func millis(ts string) int64 {
	if ts == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}
