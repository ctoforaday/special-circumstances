package hookcore

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FailedToolStep describes a tool call recorded in an Antigravity transcript whose execution
// failed (for example, a run_command that exited non-zero, which Antigravity reports on
// PostToolUse with error:"" and writes only to transcript_full.jsonl).
type FailedToolStep struct {
	StepIndex int
	ToolName  string
	Command   string
	ExitCode  int
	Output    string
}

var reCommandExitCode = regexp.MustCompile(`(?m)^Command exited with code (\d+)\.$`)

type pendingToolCall struct {
	ToolName string `json:"tool_name"`
	Command  string `json:"command"`
}

type drainWatermark struct {
	Step    int               `json:"step"`
	Offset  int64             `json:"offset"`
	Pending []pendingToolCall `json:"pending,omitempty"`
}

// DrainFailedToolSteps scans req.TranscriptPath from the persisted byte-offset watermark for newly
// completed Antigravity run_command steps that exited non-zero, updating the watermark so each
// failed step is returned at most once without rescanning earlier turns.
//
// Only complete, newline-terminated lines advance the byte offset, so a half-written final line is
// re-read on the next invocation. Pending PLANNER_RESPONSE tool calls are persisted in the
// watermark so a call and its GENERIC result straddling a drain boundary retain the command text.
// Backgrounded run_command calls (WaitMsBeforeAsync) yield no exit code in their GENERIC step and
// are not reported here. A completed run_command GENERIC step matching no recognized form returns a
// non-nil error so callers can record the unmeasured format loudly instead of returning a silent zero.
func DrainFailedToolSteps(req Request) ([]FailedToolStep, error) {
	if req.Engine != EngineAntigravity || req.TranscriptPath == "" {
		return nil, nil
	}
	key := pendingQueueKey(req)
	wm := drainWatermark{Step: -1}
	var wmPath string
	if req.PluginDataDir != "" && key != "" {
		wmPath = filepath.Join(req.PluginDataDir, "last_failed_step_"+key+".json")
		if b, err := os.ReadFile(wmPath); err == nil {
			var loaded drainWatermark
			if json.Unmarshal(b, &loaded) == nil {
				wm = loaded
			}
		}
	}

	f, err := os.Open(req.TranscriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open transcript %s: %w", req.TranscriptPath, err)
	}
	defer f.Close()

	if wm.Offset > 0 {
		if fi, err := f.Stat(); err == nil && fi.Size() >= wm.Offset {
			if _, err := f.Seek(wm.Offset, io.SeekStart); err != nil {
				return nil, fmt.Errorf("seek transcript %s: %w", req.TranscriptPath, err)
			}
		} else {
			wm = drainWatermark{Step: -1}
		}
	}

	pending := append([]pendingToolCall(nil), wm.Pending...)
	var failed []FailedToolStep
	var parseWarns []string
	maxStep := wm.Step
	curOffset := wm.Offset

	reader := bufio.NewReaderSize(f, 64*1024)
	for {
		rawLine, rerr := reader.ReadBytes('\n')
		if rerr != nil && !errors.Is(rerr, io.EOF) {
			return failed, fmt.Errorf("read transcript %s: %w", req.TranscriptPath, rerr)
		}
		// Never advance past an incomplete (non-newline-terminated) trailing line.
		if len(rawLine) == 0 || rawLine[len(rawLine)-1] != '\n' {
			break
		}
		curOffset += int64(len(rawLine))
		line := bytes.TrimSpace(rawLine)
		if len(line) == 0 {
			if errors.Is(rerr, io.EOF) {
				break
			}
			continue
		}
		var step struct {
			StepIndex int    `json:"step_index"`
			Type      string `json:"type"`
			Status    string `json:"status"`
			Content   string `json:"content"`
			ToolCalls []struct {
				Name string         `json:"name"`
				Args map[string]any `json:"args"`
			} `json:"tool_calls"`
		}
		if json.Unmarshal(line, &step) != nil {
			if errors.Is(rerr, io.EOF) {
				break
			}
			continue
		}
		if step.StepIndex > maxStep {
			maxStep = step.StepIndex
		}
		switch step.Type {
		case "PLANNER_RESPONSE":
			pending = pending[:0]
			for _, tc := range step.ToolCalls {
				cmd := ""
				if raw, ok := tc.Args["CommandLine"].(string); ok {
					cmd = unquoteJSONString(raw)
				} else if raw, ok := tc.Args["command"].(string); ok {
					cmd = unquoteJSONString(raw)
				}
				pending = append(pending, pendingToolCall{
					ToolName: CanonicalToolName(tc.Name),
					Command:  cmd,
				})
			}
		case "GENERIC":
			var call pendingToolCall
			if len(pending) > 0 {
				call = pending[0]
				pending = pending[1:]
			}
			if step.StepIndex <= wm.Step {
				break
			}
			if m := reCommandExitCode.FindStringSubmatch(step.Content); len(m) == 2 {
				var code int
				_, _ = fmt.Sscanf(m[1], "%d", &code)
				if code != 0 {
					tool := call.ToolName
					if tool == "" {
						tool = "Bash"
					}
					failed = append(failed, FailedToolStep{
						StepIndex: step.StepIndex,
						ToolName:  tool,
						Command:   call.Command,
						ExitCode:  code,
						Output:    extractCommandOutput(step.Content),
					})
				}
			} else if call.ToolName == "Bash" && !isKnownNonExitRunCommand(step.Content) {
				parseWarns = append(parseWarns, fmt.Sprintf("step %d: run_command GENERIC result matched no exit-code line (not measured)", step.StepIndex))
			}
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
	}

	if wmPath != "" && (maxStep > wm.Step || curOffset > wm.Offset) {
		_ = os.MkdirAll(req.PluginDataDir, 0o755)
		nextWM := drainWatermark{Step: maxStep, Offset: curOffset, Pending: pending}
		if b, err := json.Marshal(nextWM); err == nil {
			tmp := wmPath + ".tmp"
			if os.WriteFile(tmp, b, 0o644) == nil {
				_ = os.Rename(tmp, wmPath)
			}
		}
	}
	if len(parseWarns) > 0 {
		return failed, errors.New(strings.Join(parseWarns, "; "))
	}
	return failed, nil
}

func isKnownNonExitRunCommand(content string) bool {
	return strings.Contains(content, "Tool is running as a background task") ||
		strings.Contains(content, "tool call denied by pre-tool hook") ||
		strings.Contains(content, "Command was canceled")
}

func unquoteJSONString(s string) string {
	var unq string
	if json.Unmarshal([]byte(s), &unq) == nil {
		return unq
	}
	return s
}

func extractCommandOutput(content string) string {
	const marker = "Command output:\n"
	if _, after, ok := strings.Cut(content, marker); ok {
		return strings.TrimSpace(after)
	}
	return strings.TrimSpace(content)
}
