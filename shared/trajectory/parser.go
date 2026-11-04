package trajectory

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type claudeRawLine struct {
	UUID        string `json:"uuid"`
	ParentUUID  string `json:"parentUuid"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
	Attribution string `json:"attributionAgent"`
	Message     *struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type claudeBlock struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	Input     json.RawMessage `json:"input"`
	Content   json.RawMessage `json:"content"`
	IsError   *bool           `json:"is_error"`
}

type agyRawLine struct {
	StepIndex int           `json:"step_index"`
	Source    string        `json:"source"`
	Type      string        `json:"type"`
	Status    string        `json:"status"`
	Error     string        `json:"error"`
	CreatedAt string        `json:"created_at"`
	Content   string        `json:"content"`
	Thinking  string        `json:"thinking"`
	ToolCalls []agyToolCall `json:"tool_calls"`
}

type agyToolCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

// Parse reads one transcript file (Claude Code or Antigravity) into citable acts (tool_use and tool_result).
func Parse(r io.Reader, file string) (Result, error) {
	return parseInternal(r, file, false)
}

// ParseAll reads one transcript file into all citable events including user and assistant text.
func ParseAll(r io.Reader, file string) ([]Event, error) {
	res, err := parseInternal(r, file, true)
	return res.Events, err
}

func parseInternal(r io.Reader, file string, includeAll bool) (Result, error) {
	var res Result
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 256*1024), 64*1024*1024)

	var pendingAgyCalls []string
	var lastAgyStepUUID string

	for sc.Scan() {
		res.Lines++
		text := bytes.TrimSpace(sc.Bytes())
		if len(text) == 0 {
			continue
		}

		// Detect format: Antigravity has "step_index", Claude has "uuid" or "message"
		if bytes.Contains(text, []byte(`"step_index"`)) {
			var al agyRawLine
			if err := json.Unmarshal(text, &al); err != nil {
				res.BadLines++
				if len(res.BadSample) < badSampleCap {
					res.BadSample = append(res.BadSample, res.Lines)
				}
				continue
			}

			stepUUID := fmt.Sprintf("step-%d", al.StepIndex)
			parentUUID := lastAgyStepUUID
			lastAgyStepUUID = stepUUID

			switch al.Type {
			case "USER_INPUT":
				if includeAll && al.Content != "" {
					res.Events = append(res.Events, Event{
						File:       file,
						Line:       res.Lines,
						UUID:       stepUUID,
						ParentUUID: parentUUID,
						Timestamp:  al.CreatedAt,
						Kind:       "user_text",
						Text:       al.Content,
					})
				}

			case "PLANNER_RESPONSE":
				if includeAll {
					if al.Thinking != "" {
						res.Events = append(res.Events, Event{
							File:       file,
							Line:       res.Lines,
							UUID:       stepUUID + "-think",
							ParentUUID: parentUUID,
							Timestamp:  al.CreatedAt,
							Kind:       "assistant_text",
							Thinking:   al.Thinking,
						})
					}
					if al.Content != "" {
						res.Events = append(res.Events, Event{
							File:       file,
							Line:       res.Lines,
							UUID:       stepUUID + "-text",
							ParentUUID: parentUUID,
							Timestamp:  al.CreatedAt,
							Kind:       "assistant_text",
							Text:       al.Content,
						})
					}
				}

				for idx, tc := range al.ToolCalls {
					callID := fmt.Sprintf("call_%d_%d", al.StepIndex, idx)
					pendingAgyCalls = append(pendingAgyCalls, callID)

					canonicalName := NormalizeToolName(tc.Name)
					var normInput json.RawMessage
					var argMap map[string]any
					if err := json.Unmarshal(tc.Args, &argMap); err == nil && argMap != nil {
						normInput = NormalizeToolArgs(canonicalName, argMap)
					} else {
						normInput = tc.Args
					}

					res.Events = append(res.Events, Event{
						File:        file,
						Line:        res.Lines,
						UUID:        stepUUID,
						ParentUUID:  parentUUID,
						Timestamp:   al.CreatedAt,
						Kind:        "tool_use",
						ToolName:    canonicalName,
						RawToolName: tc.Name,
						ToolUseID:   callID,
						ToolInput:   normInput,
					})
				}

			case "GENERIC":
				if len(pendingAgyCalls) > 0 {
					callID := pendingAgyCalls[0]
					pendingAgyCalls = pendingAgyCalls[1:]

					isErr := al.Status == "ERROR" || al.Error != ""
					res.Events = append(res.Events, Event{
						File:       file,
						Line:       res.Lines,
						UUID:       stepUUID,
						ParentUUID: parentUUID,
						Timestamp:  al.CreatedAt,
						Kind:       "tool_result",
						ToolUseID:  callID,
						ToolResult: json.RawMessage(fmt.Sprintf("%q", al.Content)),
						IsError:    isErr,
					})
				}
			}
			continue
		}

		// Claude Code format
		var rl claudeRawLine
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

		var blocks []claudeBlock
		if err := json.Unmarshal(rl.Message.Content, &blocks); err != nil {
			// Plain string content is normal for user lines, not a parse failure.
			if includeAll {
				var s string
				if err := json.Unmarshal(rl.Message.Content, &s); err == nil && s != "" {
					res.Events = append(res.Events, Event{
						File:       file,
						Line:       res.Lines,
						UUID:       rl.UUID,
						ParentUUID: rl.ParentUUID,
						Timestamp:  rl.Timestamp,
						Kind:       "user_text",
						Text:       s,
					})
				}
			}
			continue
		}

		for _, b := range blocks {
			ev := Event{
				File:       file,
				Line:       res.Lines,
				UUID:       rl.UUID,
				ParentUUID: rl.ParentUUID,
				Timestamp:  rl.Timestamp,
				AgentType:  rl.Attribution,
			}
			switch b.Type {
			case "tool_use":
				ev.Kind = "tool_use"
				ev.ToolName = NormalizeToolName(b.Name)
				ev.RawToolName = b.Name
				ev.ToolUseID = b.ID
				ev.ToolInput = b.Input
			case "tool_result":
				ev.Kind = "tool_result"
				ev.ToolUseID = b.ToolUseID
				ev.ToolResult = b.Content
				if b.IsError != nil {
					ev.IsError = *b.IsError
				}
			case "text":
				if !includeAll {
					continue
				}
				ev.Kind = "assistant_text"
				ev.Text = string(b.Content)
			case "thinking":
				if !includeAll {
					continue
				}
				ev.Kind = "assistant_text"
				ev.Thinking = string(b.Content)
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
