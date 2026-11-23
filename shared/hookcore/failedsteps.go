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
	"time"

	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// StageFailedStepDrain is recorded when Antigravity's PreInvocation failed-step transcript drain
// encounters a corrupt watermark, shrunken transcript, misaligned rewrite, or unrecognized Bash
// GENERIC format. Registered with ConversationStageTTL so ended conversations age out once shown.
const StageFailedStepDrain hookfailures.Stage = "failed-step-drain"

// StageHookUnmarshal is recorded when an Antigravity transcript step failed because a hook emitted
// output the client could not unmarshal. Registered with ConversationStageTTL.
const StageHookUnmarshal hookfailures.Stage = "hook-result-unmarshal"

// ConversationStageTTL is how long a conversation-scoped drain/unmarshal failure persists after
// Last once shown at least once, before any hook merge prunes it.
const ConversationStageTTL = 30 * time.Minute

func init() {
	hookfailures.RegisterTTL(StageFailedStepDrain, ConversationStageTTL)
	hookfailures.RegisterTTL(StageHookUnmarshal, ConversationStageTTL)
}

// FailedToolStep describes a tool call recorded in an Antigravity transcript whose execution
// failed (for example, a run_command that exited non-zero, which Antigravity reports on
// PostToolUse with error:"" and writes only to transcript_full.jsonl).
type FailedToolStep struct {
	StepIndex int
	ToolName  string
	Command   string
	ExitCode  int
	Output    string
	HookError bool
}

var reCommandExitCode = regexp.MustCompile(`(?m)^(?:The )?[Cc]ommand exited with code (\d+)\.$`)

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
	var corruptErr error
	if req.PluginDataDir != "" && key != "" {
		wmPath = filepath.Join(req.PluginDataDir, "last_failed_step_"+key+".json")
		b, err := os.ReadFile(wmPath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("read drain watermark %s: %w", wmPath, err)
			}
		} else {
			var loaded drainWatermark
			if err := json.Unmarshal(b, &loaded); err != nil {
				corruptErr = fmt.Errorf("corrupt drain watermark %s: %w", wmPath, err)
			} else {
				wm = loaded
			}
		}
	}

	f, err := os.Open(req.TranscriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			if corruptErr != nil && wmPath != "" {
				_ = writeDrainWatermark(wmPath, req.PluginDataDir, drainWatermark{Step: -1, Offset: 0})
				return nil, fmt.Errorf("%v; rebased watermark to 0 with no replay", corruptErr)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("open transcript %s: %w", req.TranscriptPath, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat transcript %s: %w", req.TranscriptPath, err)
	}

	// Rebase helper for corrupt watermark, shrunken transcript, or rewritten/misaligned offset:
	// report the condition once, advance watermark to the last newline boundary at or before EOF
	// with Step: -1 (since renumbered transcripts start at step 0 again while Offset already skips
	// all old bytes, and landing on '\n' avoids misaligning on a half-written trailing line).
	rebaseEOF := func(reason string) ([]FailedToolStep, error) {
		targetOffset := lastNewlineOffset(f, fi.Size())
		if wmPath != "" {
			if werr := writeDrainWatermark(wmPath, req.PluginDataDir, drainWatermark{Step: -1, Offset: targetOffset}); werr != nil {
				return nil, fmt.Errorf("%s; rebase write failed: %w", reason, werr)
			}
		}
		return nil, fmt.Errorf("%s; rebased watermark to offset %d with no replay", reason, targetOffset)
	}

	if corruptErr != nil {
		return rebaseEOF(corruptErr.Error())
	}

	if wm.Offset > 0 {
		if fi.Size() < wm.Offset {
			return rebaseEOF(fmt.Sprintf("transcript %s shrank (%d < watermark offset %d)", req.TranscriptPath, fi.Size(), wm.Offset))
		}
		// Verify the byte immediately before wm.Offset is still '\n'; if the file was rewritten
		// to >= old size, seeking to wm.Offset would land mid-line and silently drop a step.
		var prev [1]byte
		if _, err := f.ReadAt(prev[:], wm.Offset-1); err != nil {
			return nil, fmt.Errorf("check watermark boundary in %s: %w", req.TranscriptPath, err)
		}
		if prev[0] != '\n' {
			return rebaseEOF(fmt.Sprintf("transcript %s misaligned at watermark offset %d (expected newline boundary)", req.TranscriptPath, wm.Offset))
		}
		if _, err := f.Seek(wm.Offset, io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek transcript %s: %w", req.TranscriptPath, err)
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
			} else if step.Status == "ERROR" && strings.Contains(step.Content, "Encountered error in step execution: failed to unmarshal result from hook") {
				tool := call.ToolName
				if tool == "" {
					tool = "Bash"
				}
				failed = append(failed, FailedToolStep{
					StepIndex: step.StepIndex,
					ToolName:  tool,
					Command:   call.Command,
					Output:    strings.TrimSpace(step.Content),
					HookError: true,
				})
			} else if call.ToolName == "Bash" && !isKnownNonExitRunCommand(step.Content) {
				parseWarns = append(parseWarns, fmt.Sprintf("step %d: run_command GENERIC result matched no exit-code line (not measured)", step.StepIndex))
			}
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
	}

	var errs []string
	if wmPath != "" && (maxStep > wm.Step || curOffset > wm.Offset) {
		if err := writeDrainWatermark(wmPath, req.PluginDataDir, drainWatermark{Step: maxStep, Offset: curOffset, Pending: pending}); err != nil {
			// Refuse to emit steps whose watermark could not be persisted, so a broken
			// watermark path never replays the same failure on every subsequent drain.
			return nil, err
		}
	}
	if len(parseWarns) > 0 {
		const maxSamples = 3
		shown := parseWarns
		suffix := ""
		if len(parseWarns) > maxSamples {
			shown = parseWarns[:maxSamples]
			suffix = fmt.Sprintf(" (+%d more, %d unmatched total)", len(parseWarns)-maxSamples, len(parseWarns))
		}
		errs = append(errs, strings.Join(shown, "; ")+suffix)
	}
	if len(errs) > 0 {
		return failed, errors.New(strings.Join(errs, "; "))
	}
	return failed, nil
}

// lastNewlineOffset returns the byte offset immediately after the last '\n' in f (up to size),
// or 0 if no newline is present. Keeps rebased watermarks aligned to a line boundary.
func lastNewlineOffset(f *os.File, size int64) int64 {
	const chunk = 64 * 1024
	buf := make([]byte, chunk)
	for end := size; end > 0; {
		start := end - chunk
		if start < 0 {
			start = 0
		}
		n, err := f.ReadAt(buf[:end-start], start)
		if n > 0 {
			if idx := bytes.LastIndexByte(buf[:n], '\n'); idx >= 0 {
				return start + int64(idx) + 1
			}
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return 0
		}
		end = start
	}
	return 0
}

// DrainScopeKey returns the conversation-scoped key used to isolate failed-step watermarks and
// hook-failure stage entries per conversation so a clean drain on one seat never clears another's.
func DrainScopeKey(req Request) string {
	return pendingQueueKey(req)
}

func writeDrainWatermark(wmPath, dir string, wm drainWatermark) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create drain watermark dir %s: %w", dir, err)
	}
	b, err := json.Marshal(wm)
	if err != nil {
		return fmt.Errorf("marshal drain watermark: %w", err)
	}
	tmp := wmPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("write drain watermark %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, wmPath); err != nil {
		return fmt.Errorf("rename drain watermark %s: %w", wmPath, err)
	}
	return nil
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
