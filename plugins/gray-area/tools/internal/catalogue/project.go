package catalogue

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"
)

// record is the subset of a transcript line this reads. The format is vendor-internal and
// version-unstable, so everything here is optional and a missing field is data rather than an
// error — the projection fails toward UNDER-counting, and §V asserts that an unparsed line is
// reported as unparsed rather than as a session with zero acts.
type record struct {
	UUID       string          `json:"uuid"`
	ParentUUID string          `json:"parentUuid"`
	PromptID   string          `json:"promptId"`
	SessionID  string          `json:"sessionId"`
	CWD        string          `json:"cwd"`
	Type       string          `json:"type"`
	Timestamp  string          `json:"timestamp"`
	Message    json.RawMessage `json:"message"`
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

// Projection is one file's worth of extracted rows.
type Projection struct {
	Acts     []Act
	Words    []Word
	Thoughts []Thought
	// Unparsed counts lines that are not JSON. Reported, never skipped in silence: a torn line
	// and an absent line produce the same missing row, and only this count separates them.
	Unparsed int
	// ExplicitFlag counts tool results carrying an is_error key, by tool. R2's detector: Bash
	// states it on 100.00% of calls that produced a result, so a rename drives that to 0.
	ExplicitFlag map[string]int
	ResultCount  map[string]int
	LastOffset   int64
}

type Act struct {
	Seq     int
	TS      int64
	Tool    string
	Target  string
	Outcome string
}

type Word struct {
	PromptID string
	BlockSeq int
	TS       int64
	Role     string
	Text     string
}

type Thought struct {
	PromptID string
	BlockSeq int
	TS       int64
	Text     string
}

// Project reads a transcript from r and extracts the three tiers.
//
// It resolves each assistant record's prompt_id by walking parentUuid to the nearest ancestor
// carrying one, because assistant records never carry it themselves and `user` records do —
// measured, 46,043 of 46,046 assistant records resolve, 99.9935%. The 3 that do not are record 0
// of a resumed file whose parent is in another file, and one subagent record with no parent at
// all; they are stored with an empty prompt_id rather than dropped.
func Project(r io.Reader, startSeq int) Projection {
	p := Projection{ExplicitFlag: map[string]int{}, ResultCount: map[string]int{}}
	byUUID := map[string]*record{}
	var recs []*record

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 256*1024), 64*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec record
		if json.Unmarshal([]byte(line), &rec) != nil {
			p.Unparsed++
			continue
		}
		if rec.UUID != "" {
			byUUID[rec.UUID] = &rec
		}
		recs = append(recs, &rec)
	}

	// Outcomes are resolved by pairing tool_use to tool_result across the whole file first: a
	// result may be many lines after its call.
	type useInfo struct {
		tool, target string
		ts           int64
	}
	uses := map[string]useInfo{}
	// THE ORDER THE CALLS WERE MADE IN, kept alongside the map because a Go map does not have
	// one. Ranging the map to assign `seq` gave every projection of the same file a DIFFERENT
	// ordering — the column that exists to say what an agent did first held a fresh shuffle on
	// every run, and reprojecting a file produced a store that disagreed with the last one. It
	// read as fine: the values were dense, unique and plausible, and every count over them was
	// right. Caught by a golden test whose two runs disagreed, never by a count.
	useOrder := []string{}
	results := map[string]*bool{}
	seen := map[string]bool{}

	seq := startSeq
	for _, rec := range recs {
		m := decodeMessage(rec.Message)
		if m == nil {
			continue
		}
		blocks := decodeBlocks(m.Content)
		ts := parseTS(rec.Timestamp)
		pid := resolvePromptID(rec, byUUID)

		blockSeq := 0
		for _, b := range blocks {
			switch b.Type {
			case "tool_use":
				if _, dup := uses[b.ID]; !dup {
					useOrder = append(useOrder, b.ID)
				}
				uses[b.ID] = useInfo{tool: b.Name, target: targetOf(b.Name, b.Input), ts: ts}
			case "tool_result":
				results[b.ToolUseID] = b.IsError
				seen[b.ToolUseID] = true
			case "text":
				if strings.TrimSpace(b.Text) == "" {
					continue
				}
				p.Words = append(p.Words, Word{PromptID: pid, BlockSeq: blockSeq, TS: ts, Role: m.Role, Text: b.Text})
				blockSeq++
			case "thinking":
				if strings.TrimSpace(b.Thinking) == "" {
					continue // an empty block with a signature is capture being off, not a thought
				}
				p.Thoughts = append(p.Thoughts, Thought{PromptID: pid, BlockSeq: blockSeq, TS: ts, Text: b.Thinking})
				blockSeq++
			}
		}
	}

	for _, id := range useOrder {
		u := uses[id]
		out := OutcomeUnresolved
		if seen[id] {
			p.ResultCount[u.tool]++
			flag := results[id]
			if flag != nil {
				p.ExplicitFlag[u.tool]++
			}
			// ABSENT IS SUCCESS. Only an explicit true is an error.
			if flag != nil && *flag {
				out = OutcomeError
			} else {
				out = OutcomeOK
			}
		}
		p.Acts = append(p.Acts, Act{Seq: seq, TS: u.ts, Tool: u.tool, Target: u.target, Outcome: out})
		seq++
	}
	return p
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

// resolvePromptID walks parentUuid to the nearest ancestor carrying a promptId.
func resolvePromptID(rec *record, byUUID map[string]*record) string {
	if rec.PromptID != "" {
		return rec.PromptID
	}
	cur := rec.ParentUUID
	for n := 0; cur != "" && n < 500; n++ {
		p := byUUID[cur]
		if p == nil {
			return "" // parent lives in another file: a resumed or forked transcript
		}
		if p.PromptID != "" {
			return p.PromptID
		}
		cur = p.ParentUUID
	}
	return ""
}

// targetOf extracts a BOUNDED target from a tool's input — the first path-like or command-head
// argument, truncated. Never the whole input: the row is a summary, and tool inputs are the bulk
// this store exists not to copy.
func targetOf(tool string, input json.RawMessage) string {
	if len(input) == 0 {
		return ""
	}
	var m map[string]any
	if json.Unmarshal(input, &m) != nil {
		return ""
	}
	for _, k := range []string{"file_path", "path", "notebook_path", "pattern", "command", "url", "query"} {
		if v, ok := m[k].(string); ok && v != "" {
			return truncate(strings.TrimSpace(v), 200)
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
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
