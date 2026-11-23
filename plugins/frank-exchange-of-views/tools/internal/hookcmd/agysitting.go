package hookcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// agySitting is a seat sitting registered by debate.py under <runDir>/.agy-sittings/.
//
// Antigravity's run_workflow dispatches each agent() as a top-level conversation whose hook
// payloads carry its own conversationId and neither parentConversationId nor agentName. debate.py
// writes seat-<seat_id>.json before launching a seat and removes it on close; on the seat's first
// run_command carrying --seat-id <seat_id>, this hook binds conv-<conversationId>.json atomically
// (guarded by an O_EXCL seat-<seat_id>.bound marker so a second conversation cannot claim the same
// sitting) so every subsequent call in that conversation (run_command, read_file, PostToolUse,
// PostFailure, PreInvocation) resolves its seat identity and lead session directly.
type agySitting struct {
	AgentID     string `json:"agent_id"`
	AgentType   string `json:"agent_type"`
	SeatID      string `json:"seat_id"`
	RunDir      string `json:"run_dir"`
	LeadSession string `json:"lead_session"`
}

const agySittingsDir = ".agy-sittings"

// StageAgyBind is recorded when a second Antigravity conversation tries to bind an already-claimed
// seat sitting, instead of silently ignoring the conflict. Cleared on the next successful bind.
const StageAgyBind hookfailures.Stage = "agy-sitting-bind"

var resolveAgySittingFn = resolveAgySitting

// applyAgySitting populates req's AgentID, AgentType and SessionID from an active Antigravity
// run_workflow sitting when applicable. Used uniformly by Pre, Post, PostFailure, and PreInvocation.
func applyAgySitting(req *hookcore.Request, cwd string, rec *hookfailures.Recorder) (agySitting, bool) {
	s, ok := resolveAgySittingFn(*req, cwd, rec)
	if !ok {
		return agySitting{}, false
	}
	req.AgentID = s.AgentID
	req.AgentType = s.AgentType
	req.SessionID = s.LeadSession
	return s, true
}

// applyBoundAgySitting resolves an already-bound Antigravity sitting (fast-path conv-<id>.json
// only, never touching O_EXCL seat-<id>.bound) so Run can populate req.AgentID before f runs
// without racing with f's first-bind path.
func applyBoundAgySitting(req *hookcore.Request, cwd string) bool {
	probe := *req
	probe.ToolName = ""
	probe.Command = ""
	s, ok := resolveAgySittingFn(probe, cwd, nil)
	if !ok {
		return false
	}
	req.AgentID = s.AgentID
	req.AgentType = s.AgentType
	req.SessionID = s.LeadSession
	return true
}

func resolveAgySitting(req hookcore.Request, cwd string, rec *hookfailures.Recorder) (agySitting, bool) {
	if req.Engine != hookcore.EngineAntigravity || req.AgentID != "" || req.SessionID == "" {
		return agySitting{}, false
	}
	markerDir := findRunLiveDir(cwd)
	if markerDir == "" {
		return agySitting{}, false
	}
	runs := runlive.ReadRunLive(markerDir)
	for _, m := range runs {
		if m.LeadSession == "" {
			continue
		}
		dir := runlive.Resolve(markerDir, m.RunDir)
		sdir := filepath.Join(dir, agySittingsDir)

		// Fast path: this conversation is already bound to an active sitting.
		if s, ok := readAgySitting(filepath.Join(sdir, "conv-"+req.SessionID+".json")); ok && s.LeadSession == m.LeadSession {
			caller := runlive.Caller{ID: s.AgentID, Type: s.AgentType, Session: s.LeadSession}
			if runlive.StateOf(markerDir, m, caller) == runlive.RowLive {
				return s, true
			}
			continue
		}

		// First-call bind: a run_command calling feov-record --seat-id <id>.
		if req.ToolName != "Bash" {
			continue
		}
		seatID := extractSeatIDFlag(req.Command)
		if seatID == "" {
			continue
		}
		s, ok := readAgySitting(filepath.Join(sdir, "seat-"+seatID+".json"))
		if !ok || s.LeadSession != m.LeadSession {
			continue
		}
		caller := runlive.Caller{ID: s.AgentID, Type: s.AgentType, Session: s.LeadSession}
		if runlive.StateOf(markerDir, m, caller) != runlive.RowLive {
			continue
		}
		// First-bind exclusivity: O_EXCL prevents a second conversation from hijacking an active seat.
		boundPath := filepath.Join(sdir, "seat-"+seatID+".bound")
		bf, err := os.OpenFile(boundPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			if rec != nil {
				rec.FailIn(StageAgyBind, markerDir, fmt.Sprintf("conversation %s tried to bind already-claimed seat %q (%v)", req.SessionID, seatID, err))
			}
			continue
		}
		_, _ = bf.WriteString(req.SessionID + "\n")
		_ = bf.Close()
		if rec != nil {
			rec.OKIn(StageAgyBind, markerDir)
		}

		if b, err := json.Marshal(s); err == nil {
			_ = writeFileAtomic(filepath.Join(sdir, "conv-"+req.SessionID+".json"), b, 0o600)
		}
		_ = writeFileAtomic(filepath.Join(sdir, "seat-"+seatID+".conv"), []byte(req.SessionID), 0o600)
		if req.TranscriptPath != "" {
			_ = writeFileAtomic(filepath.Join(sdir, "seat-"+seatID+".transcript"), []byte(req.TranscriptPath), 0o600)
		}
		return s, true
	}
	return agySitting{}, false
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func findRunLiveDir(start string) string {
	dir := start
	if dir == "" {
		dir, _ = os.Getwd()
	}
	for i := 0; dir != "" && i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "run-live.json")); err == nil {
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

func readAgySitting(path string) (agySitting, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return agySitting{}, false
	}
	var s agySitting
	if err := json.Unmarshal(b, &s); err != nil || s.AgentID == "" || s.AgentType == "" || s.LeadSession == "" {
		return agySitting{}, false
	}
	return s, true
}

// extractSeatIDFlag returns the value of --seat-id <val> or --seat-id=<val> from a shell command.
func extractSeatIDFlag(cmd string) string {
	fields := strings.Fields(cmd)
	for i := 0; i < len(fields); i++ {
		f := strings.Trim(fields[i], `"'`)
		if f == "--seat-id" && i+1 < len(fields) {
			return sanitizeSeatID(fields[i+1])
		}
		if after, ok := strings.CutPrefix(f, "--seat-id="); ok {
			return sanitizeSeatID(after)
		}
	}
	return ""
}

func sanitizeSeatID(raw string) string {
	s := strings.Trim(raw, `"'`+";&|)")
	if s == "" || strings.ContainsAny(s, "/\\..") {
		return ""
	}
	return s
}

// logAgyFailedToolSteps drains newly completed non-zero run_command steps from a bound Antigravity
// seat's transcript on PreInvocation and logs each via failureEntry/recordFailure. The caller check
// runs first so unbound conversations perform zero transcript I/O.
func logAgyFailedToolSteps(req hookcore.Request, in hookgate.Input, cwd string, rec *hookfailures.Recorder) {
	if req.Engine != hookcore.EngineAntigravity {
		return
	}
	caller := callerAgent(in)
	if caller.ID == "" {
		return
	}
	runDir := seatRunDir(cwd, caller, rec)
	if runDir == "" {
		return
	}
	scope := hookcore.DrainScopeKey(req)
	failed, drainErr := hookcore.DrainFailedToolSteps(req)
	if drainErr != nil {
		rec.FailBounded(hookcore.StageFailedStepDrain, scope, "failed-step drain: "+drainErr.Error())
	} else {
		rec.OKIn(hookcore.StageFailedStepDrain.Stage, scope)
	}
	sawHookErr := false
	for _, st := range failed {
		if st.HookError {
			sawHookErr = true
			rec.FailBounded(hookcore.StageHookUnmarshal, scope, "hook result failed to parse: "+strings.TrimSpace(st.Output))
		}
		rawInput, _ := json.Marshal(map[string]any{"command": strings.TrimSpace(st.Command)})
		errText := strings.TrimSpace(st.Output)
		if errText == "" {
			errText = fmt.Sprintf("Command exited with code %d.", st.ExitCode)
		}
		text, ok := failureEntry(runDir, st.ToolName, rawInput, errText)
		if !ok {
			continue
		}
		if err := recordFailure(runDir, caller.ID, text); err != nil {
			rec.Fail(StageFailureLog, "a seat's call failed and could not be logged: "+err.Error())
		} else {
			rec.OK(StageFailureLog)
		}
	}
	if !sawHookErr {
		rec.OKIn(hookcore.StageHookUnmarshal.Stage, scope)
	}
}
