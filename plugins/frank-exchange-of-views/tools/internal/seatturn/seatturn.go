// Package seatturn reads a seat's transcript into per-turn rows: what each API round trip cost and
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

// Turn is one API round trip: the usage it reported and the shape of what came back.
type Turn struct {
	Index         int   // position among the transcript's requests, from 0, in the order each first appears
	TSMillis      int64 // the timestamp of the request's LAST line, when its response was complete; 0 when it carried none
	Model         string
	Input         int
	Output        int
	CacheRead     int
	CacheCreation int
	// CacheCreation1h is the part of CacheCreation written with the one-hour lifetime, which is
	// priced at twice the five-minute write. It is a parse output for pricing and is not a seat_turn
	// column: the table's shape is the record's, and the price is computed from the transcript.
	CacheCreation1h int
	// Thinking and Tool are read off the CONTENT BLOCKS, not inferred from the numbers. The F11
	// decomposition inferred a thinking turn from a low output_tokens with a long span, which is
	// a correlation that happens to hold; the block is the fact, and it is right there.
	Thinking bool
	Tool     bool
}

// Prompt is what the request sent: its uncached input, what it read from the cache and what it
// wrote to it. A price card that changes with prompt length is chosen on this.
func (t Turn) Prompt() int { return t.Input + t.CacheRead + t.CacheCreation }

// Parse reads a transcript into its turns and reports the agent it belongs to.
//
// ONE TURN PER REQUEST, NOT PER LINE. The client writes one transcript line per CONTENT BLOCK of a
// response — a thinking block, then each tool_use, then text — and every one of those lines repeats
// the request's usage. Counting lines counted each request once per block: on the 2026-10-17 smoke
// a dark-side seat had 32 lines for 14 requests, and the run's cache reads summed to 13.97M against
// the 6,521,112 the client itself billed. The request is named by `message.id`, the API's own id for
// the response, so the lines sharing one are folded into one turn:
//   - usage is the LAST line's. The client streams, and the early lines carry the output count as it
//     stood when that block closed (8, then 8, then 563); the last carries the response's total. The
//     input and cache counts are the same on every line of a request.
//   - the timestamp is the last line's, when the response was complete, so a span between two turns
//     runs from one finished response to the next.
//   - Thinking and Tool are true when ANY of the request's blocks was one.
//
// A usage line with no message.id is its own turn: nothing names the request it belongs to, and
// joining it to a neighbour would be a guess.
//
// THE AGENT ID COMES OFF THE LINE, NOT THE FILENAME. Every record in these files carries
// `agentId`, and the file is also called agent-<id>.jsonl — recovering it from the name would be
// a fact composed into a path at one end and matched out at the other, whose miss (a renamed or
// copied file) is silent and produces rows filed against nothing. The field is free.
//
// A line that does not parse, or carries no usage, is skipped: transcripts interleave user
// records, tool results and meta lines, and only the assistant's API round trips report usage.
func Parse(transcript string) (agentID string, turns []Turn) {
	byID := map[string]int{}
	for _, line := range strings.Split(transcript, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec struct {
			AgentID   string `json:"agentId"`
			Timestamp string `json:"timestamp"`
			Message   *struct {
				ID      string `json:"id"`
				Model   string `json:"model"`
				Content []struct {
					Type string `json:"type"`
				} `json:"content"`
				Usage *struct {
					Input         int `json:"input_tokens"`
					Output        int `json:"output_tokens"`
					CacheRead     int `json:"cache_read_input_tokens"`
					CacheCreation int `json:"cache_creation_input_tokens"`
					ByLifetime    *struct {
						OneHour int `json:"ephemeral_1h_input_tokens"`
					} `json:"cache_creation"`
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
		u := rec.Message.Usage
		t := Turn{
			Index:         len(turns),
			TSMillis:      millis(rec.Timestamp),
			Model:         rec.Message.Model,
			Input:         u.Input,
			Output:        u.Output,
			CacheRead:     u.CacheRead,
			CacheCreation: u.CacheCreation,
		}
		if u.ByLifetime != nil {
			t.CacheCreation1h = u.ByLifetime.OneHour
		}
		for _, b := range rec.Message.Content {
			switch b.Type {
			case "thinking":
				t.Thinking = true
			case "tool_use":
				t.Tool = true
			}
		}
		if id := rec.Message.ID; id != "" {
			if i, seen := byID[id]; seen {
				prev := turns[i]
				t.Index = prev.Index
				t.Thinking = t.Thinking || prev.Thinking
				t.Tool = t.Tool || prev.Tool
				if t.TSMillis == 0 {
					t.TSMillis = prev.TSMillis
				}
				turns[i] = t
				continue
			}
			byID[id] = len(turns)
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
