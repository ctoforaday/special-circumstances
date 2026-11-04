package catalogue

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	sharedtraj "github.com/ctoforaday/special-circumstances/shared/trajectory"
)

type Act = sharedtraj.Act
type Word = sharedtraj.Word
type Thought = sharedtraj.Thought
type Skip = sharedtraj.Skip
type Projection = sharedtraj.Projection

// Project reads a transcript from r and extracts acts, words, thoughts, and skips across Claude and Antigravity.
func Project(r io.Reader, startSeq int) Projection {
	return sharedtraj.Project(r, startSeq)
}

// record is the subset of a transcript line this reads. Retained for hit.go and legacy helpers.
type record struct {
	UUID             string          `json:"uuid"`
	ParentUUID       string          `json:"parentUuid"`
	PromptID         string          `json:"promptId"`
	SessionID        string          `json:"sessionId"`
	CWD              string          `json:"cwd"`
	Type             string          `json:"type"`
	Timestamp        string          `json:"timestamp"`
	Message          json.RawMessage `json:"message"`
	Origin           *origin         `json:"origin"`
	IsMeta           bool            `json:"isMeta"`
	IsCompactSummary bool            `json:"isCompactSummary"`
	Attachment       *attachment     `json:"attachment"`
}

type origin struct {
	Kind string `json:"kind"`
}

// attachment is a record with no message. The only kind read is `queued_command` — a prompt,
// a peer's message or a notification delivered MID-TURN — which carries its own origin and mode.
type attachment struct {
	Type        string          `json:"type"`
	Prompt      json.RawMessage `json:"prompt"`
	CommandMode string          `json:"commandMode"`
	Origin      *origin         `json:"origin"`
}

func (o *origin) kind() string {
	if o == nil {
		return ""
	}
	return o.Kind
}

type message struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type block struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	Input     json.RawMessage `json:"input"`
	// IsError is a POINTER because its absence is the normal success shape for every tool but
	// Bash — measured, 18.4% of all calls and ~99% of non-Bash successes. A plain bool would make
	// "absent" and "false" the same value, which is fine for the outcome but hides the drift
	// detector: if the vendor renames this field, Bash's explicit rate falls from 100% to 0 and
	// that is the only signal the failure rate did not really go to zero.
	IsError *bool `json:"is_error"`
}

func decodeMessage(raw json.RawMessage) *message {
	if len(raw) == 0 {
		return nil
	}
	var m message
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	return &m
}

func decodeBlocks(raw json.RawMessage) []block {
	if len(raw) == 0 {
		return nil
	}
	var bs []block
	if json.Unmarshal(raw, &bs) == nil {
		return bs
	}
	// content may be a bare string on user records; that is a word, handled by the caller via a
	// synthetic text block.
	var s string
	if json.Unmarshal(raw, &s) == nil && strings.TrimSpace(s) != "" {
		return []block{{Type: "text", Text: s}}
	}
	return nil
}

func parseTS(s string) int64 {
	if s == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0
	}
	return t.Unix()
}
