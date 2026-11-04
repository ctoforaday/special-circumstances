package trajectory

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	OutcomeOK         = "ok"
	OutcomeError      = "error"
	OutcomeUnresolved = "unresolved"
	SkipThinkingEmpty = "thinking-empty"
)

type Act struct {
	Seq     int
	TS      int64
	Tool    string
	Target  string
	Outcome string
}

type Word struct {
	PromptID   string
	BlockSeq   int
	TS         int64
	Role       string
	Text       string
	OriginKind string
	Meta       bool
}

type Thought struct {
	PromptID string
	BlockSeq int
	TS       int64
	Text     string
}

type Skip struct {
	PromptID string
	Reason   string
	At       int64
}

// Projection is one file's worth of extracted catalogue tiers.
type Projection struct {
	Acts         []Act
	Words        []Word
	Thoughts     []Thought
	Skips        []Skip
	Unparsed     int
	ExplicitFlag map[string]int
	ResultCount  map[string]int
	LastOffset   int64
	CWDs         []string
}

// Project reads a transcript from r and extracts acts, words, thoughts, and skips across Claude and Antigravity.
func Project(r io.Reader, startSeq int) Projection {
	p := Projection{
		ExplicitFlag: map[string]int{},
		ResultCount:  map[string]int{},
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 256*1024), 64*1024*1024)

	var lines [][]byte
	isAntigravity := false

	for sc.Scan() {
		b := bytes.TrimSpace(sc.Bytes())
		if len(b) == 0 {
			continue
		}
		if bytes.Contains(b, []byte(`"step_index"`)) {
			isAntigravity = true
		}
		lineCopy := make([]byte, len(b))
		copy(lineCopy, b)
		lines = append(lines, lineCopy)
	}

	if isAntigravity {
		projectAntigravity(&p, lines, startSeq)
	} else {
		projectClaude(&p, lines, startSeq)
	}

	return p
}

func parseRFC3339(s string) int64 {
	if s == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func findRepoRoot(path string) string {
	cleaned := filepath.Clean(path)
	dir := cleaned
	if fi, err := os.Stat(cleaned); err == nil && !fi.IsDir() {
		dir = filepath.Dir(cleaned)
	} else if err != nil && filepath.Ext(cleaned) != "" {
		dir = filepath.Dir(cleaned)
	}
	for len(dir) > 1 {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func projectAntigravity(p *Projection, lines [][]byte, startSeq int) {
	type pendingCall struct {
		id     string
		tool   string
		target string
		ts     int64
	}
	var pending []pendingCall
	var useOrder []string
	callMap := map[string]pendingCall{}
	results := map[string]bool{} // true = error, false = ok
	seen := map[string]bool{}

	seq := startSeq
	curPromptID := ""
	seenCWD := map[string]bool{}

	for _, line := range lines {
		var al agyRawLine
		if json.Unmarshal(line, &al) != nil {
			p.Unparsed++
			continue
		}

		ts := parseRFC3339(al.CreatedAt)
		stepPID := fmt.Sprintf("step-%d", al.StepIndex)

		switch al.Type {
		case "USER_INPUT":
			curPromptID = stepPID
			if strings.TrimSpace(al.Content) != "" {
				p.Words = append(p.Words, Word{
					PromptID: curPromptID,
					BlockSeq: 0,
					TS:       ts,
					Role:     "user",
					Text:     al.Content,
				})
			}

		case "PLANNER_RESPONSE":
			pid := curPromptID
			if pid == "" {
				pid = stepPID
			}
			blockSeq := 0
			if strings.TrimSpace(al.Thinking) != "" {
				p.Thoughts = append(p.Thoughts, Thought{
					PromptID: pid,
					BlockSeq: blockSeq,
					TS:       ts,
					Text:     al.Thinking,
				})
				blockSeq++
			} else if al.Thinking != "" {
				p.Skips = append(p.Skips, Skip{
					PromptID: pid,
					Reason:   SkipThinkingEmpty,
					At:       ts,
				})
			}
			if strings.TrimSpace(al.Content) != "" {
				p.Words = append(p.Words, Word{
					PromptID: pid,
					BlockSeq: blockSeq,
					TS:       ts,
					Role:     "assistant",
					Text:     al.Content,
				})
				blockSeq++
			}

			for idx, tc := range al.ToolCalls {
				callID := fmt.Sprintf("call_%d_%d", al.StepIndex, idx)
				cTool := NormalizeToolName(tc.Name)
				cTarget := TargetOf(cTool, tc.Args)
				if clean := strings.Trim(cTarget, "\"'"); strings.HasPrefix(clean, "/") {
					if repo := findRepoRoot(clean); repo != "" && !seenCWD[repo] {
						seenCWD[repo] = true
						p.CWDs = append(p.CWDs, repo)
					}
				}
				pc := pendingCall{
					id:     callID,
					tool:   cTool,
					target: cTarget,
					ts:     ts,
				}
				pending = append(pending, pc)
				callMap[callID] = pc
				useOrder = append(useOrder, callID)
			}

		case "GENERIC":
			if len(pending) > 0 {
				call := pending[0]
				pending = pending[1:]
				isErr := al.Status == "ERROR" || al.Error != ""
				seen[call.id] = true
				results[call.id] = isErr
			}

		case "SYSTEM_MESSAGE":
			if strings.TrimSpace(al.Content) != "" {
				pid := curPromptID
				if pid == "" {
					pid = stepPID
				}
				p.Words = append(p.Words, Word{
					PromptID:   pid,
					BlockSeq:   0,
					TS:         ts,
					Role:       "system",
					Text:       al.Content,
					OriginKind: "system",
				})
			}
		}
	}

	for _, id := range useOrder {
		call := callMap[id]
		out := OutcomeUnresolved
		if seen[id] {
			p.ResultCount[call.tool]++
			p.ExplicitFlag[call.tool]++
			if results[id] {
				out = OutcomeError
			} else {
				out = OutcomeOK
			}
		}
		p.Acts = append(p.Acts, Act{
			Seq:     seq,
			TS:      call.ts,
			Tool:    call.tool,
			Target:  call.target,
			Outcome: out,
		})
		seq++
	}
}

type claudeRecord struct {
	UUID             string                 `json:"uuid"`
	ParentUUID       string                 `json:"parentUuid"`
	PromptID         string                 `json:"promptId"`
	SessionID        string                 `json:"sessionId"`
	CWD              string                 `json:"cwd"`
	Type             string                 `json:"type"`
	Timestamp        string                 `json:"timestamp"`
	Message          json.RawMessage        `json:"message"`
	Origin           *struct{ Kind string } `json:"origin"`
	IsMeta           bool                   `json:"isMeta"`
	IsCompactSummary bool                   `json:"isCompactSummary"`
}

type claudeMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type claudeBlk struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	Input     json.RawMessage `json:"input"`
	IsError   *bool           `json:"is_error"`
}

func projectClaude(p *Projection, lines [][]byte, startSeq int) {
	byUUID := map[string]*claudeRecord{}
	var recs []*claudeRecord
	seenCWD := map[string]bool{}

	for _, line := range lines {
		var rec claudeRecord
		if json.Unmarshal(line, &rec) != nil {
			p.Unparsed++
			continue
		}
		if rec.UUID != "" {
			byUUID[rec.UUID] = &rec
		}
		if rec.CWD != "" && !seenCWD[rec.CWD] {
			seenCWD[rec.CWD] = true
			p.CWDs = append(p.CWDs, rec.CWD)
		}
		recs = append(recs, &rec)
	}

	type useInfo struct {
		tool, target string
		ts           int64
	}
	uses := map[string]useInfo{}
	useOrder := []string{}
	results := map[string]*bool{}
	seen := map[string]bool{}

	seq := startSeq
	for _, rec := range recs {
		if len(rec.Message) == 0 {
			continue
		}
		var m claudeMsg
		if json.Unmarshal(rec.Message, &m) != nil {
			continue
		}
		var blocks []claudeBlk
		if json.Unmarshal(m.Content, &blocks) != nil {
			var s string
			if json.Unmarshal(m.Content, &s) == nil && strings.TrimSpace(s) != "" {
				blocks = []claudeBlk{{Type: "text", Text: s}}
			}
		}
		ts := parseRFC3339(rec.Timestamp)
		pid := resolveClaudePromptID(rec, byUUID)

		blockSeq := 0
		for _, b := range blocks {
			switch b.Type {
			case "tool_use":
				if _, dup := uses[b.ID]; !dup {
					useOrder = append(useOrder, b.ID)
				}
				uses[b.ID] = useInfo{tool: b.Name, target: TargetOf(b.Name, b.Input), ts: ts}
			case "tool_result":
				results[b.ToolUseID] = b.IsError
				seen[b.ToolUseID] = true
			case "text":
				if strings.TrimSpace(b.Text) == "" {
					continue
				}
				origKind := ""
				if rec.Origin != nil {
					origKind = rec.Origin.Kind
				}
				p.Words = append(p.Words, Word{
					PromptID:   pid,
					BlockSeq:   blockSeq,
					TS:         ts,
					Role:       m.Role,
					Text:       b.Text,
					OriginKind: origKind,
					Meta:       rec.IsMeta || rec.IsCompactSummary,
				})
				blockSeq++
			case "thinking":
				if strings.TrimSpace(b.Thinking) == "" {
					p.Skips = append(p.Skips, Skip{PromptID: pid, Reason: SkipThinkingEmpty, At: ts})
					continue
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
			if flag != nil && *flag {
				out = OutcomeError
			} else {
				out = OutcomeOK
			}
		}
		p.Acts = append(p.Acts, Act{Seq: seq, TS: u.ts, Tool: u.tool, Target: u.target, Outcome: out})
		seq++
	}
}

func resolveClaudePromptID(rec *claudeRecord, byUUID map[string]*claudeRecord) string {
	if rec.PromptID != "" {
		return rec.PromptID
	}
	cur := rec.ParentUUID
	for n := 0; cur != "" && n < 500; n++ {
		p := byUUID[cur]
		if p == nil {
			return ""
		}
		if p.PromptID != "" {
			return p.PromptID
		}
		cur = p.ParentUUID
	}
	return ""
}
