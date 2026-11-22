package hookcore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// DrainFailedToolSteps scans req.TranscriptPath from the persisted byte-offset watermark for newly
// completed Antigravity run_command steps that exited non-zero, updating the watermark so each
// failed step is returned at most once without rescanning earlier turns.
//
// Antigravity's PostToolUse hook fires with error:"" and no result field when run_command exits
// non-zero, and the GENERIC result step is not yet flushed to transcript_full.jsonl until the
// following PreInvocation. Backgrounded run_command calls (WaitMsBeforeAsync) yield no exit code
// in their GENERIC step and are not reported here.
func DrainFailedToolSteps(req Request) []FailedToolStep {
	if req.Engine != EngineAntigravity || req.TranscriptPath == "" {
		return nil
	}
	key := pendingQueueKey(req)
	watermarkStep := -1
	var watermarkOffset int64
	var wmPath string
	if req.PluginDataDir != "" && key != "" {
		wmPath = filepath.Join(req.PluginDataDir, "last_failed_step_"+key+".txt")
		if b, err := os.ReadFile(wmPath); err == nil {
			fields := strings.Fields(strings.TrimSpace(string(b)))
			if len(fields) >= 1 {
				if n, err := strconv.Atoi(fields[0]); err == nil {
					watermarkStep = n
				}
			}
			if len(fields) >= 2 {
				if off, err := strconv.ParseInt(fields[1], 10, 64); err == nil && off > 0 {
					watermarkOffset = off
				}
			}
		}
	}

	f, err := os.Open(req.TranscriptPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	if watermarkOffset > 0 {
		if fi, err := f.Stat(); err == nil && fi.Size() >= watermarkOffset {
			_, _ = f.Seek(watermarkOffset, io.SeekStart)
		} else {
			watermarkOffset = 0
		}
	}

	type pendingCall struct {
		toolName string
		command  string
	}
	var pending []pendingCall
	var failed []FailedToolStep
	maxStep := watermarkStep
	curOffset := watermarkOffset

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		rawLine := sc.Bytes()
		curOffset += int64(len(rawLine)) + 1 // +1 for '\n'
		line := strings.TrimSpace(string(rawLine))
		if line == "" {
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
		if json.Unmarshal([]byte(line), &step) != nil {
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
				pending = append(pending, pendingCall{
					toolName: CanonicalToolName(tc.Name),
					command:  cmd,
				})
			}
		case "GENERIC":
			var call pendingCall
			if len(pending) > 0 {
				call = pending[0]
				pending = pending[1:]
			}
			if step.StepIndex <= watermarkStep {
				continue
			}
			if m := reCommandExitCode.FindStringSubmatch(step.Content); len(m) == 2 {
				code, _ := strconv.Atoi(m[1])
				if code != 0 {
					tool := call.toolName
					if tool == "" {
						tool = "Bash"
					}
					failed = append(failed, FailedToolStep{
						StepIndex: step.StepIndex,
						ToolName:  tool,
						Command:   call.command,
						ExitCode:  code,
						Output:    extractCommandOutput(step.Content),
					})
				}
			}
		}
	}
	if sc.Err() != nil {
		return failed
	}

	if wmPath != "" && (maxStep > watermarkStep || curOffset > watermarkOffset) {
		_ = os.MkdirAll(req.PluginDataDir, 0o755)
		tmp := wmPath + ".tmp"
		if os.WriteFile(tmp, []byte(fmt.Sprintf("%d %d\n", maxStep, curOffset)), 0o644) == nil {
			_ = os.Rename(tmp, wmPath)
		}
	}
	return failed
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
