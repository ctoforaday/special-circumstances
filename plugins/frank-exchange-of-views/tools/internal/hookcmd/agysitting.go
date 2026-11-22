package hookcmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// agySitting is a seat sitting registered by debate.py under <runDir>/.agy-sittings/.
//
// Antigravity's run_workflow dispatches each agent() as a top-level conversation whose hook
// payloads carry its own conversationId and neither parentConversationId nor agentName. debate.py
// writes seat-<seat_id>.json before launching a seat and removes it on close; on the seat's first
// run_command carrying --seat-id <seat_id>, this hook binds conv-<conversationId>.json so every
// subsequent call in that conversation (run_command, read_file, PostToolUse) resolves its seat
// identity and lead session directly.
type agySitting struct {
	AgentID     string `json:"agent_id"`
	AgentType   string `json:"agent_type"`
	SeatID      string `json:"seat_id"`
	RunDir      string `json:"run_dir"`
	LeadSession string `json:"lead_session"`
}

const agySittingsDir = ".agy-sittings"

func resolveAgySitting(req hookcore.Request, cwd string) (agySitting, bool) {
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
			return s, true
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
		if b, err := json.Marshal(s); err == nil {
			_ = os.WriteFile(filepath.Join(sdir, "conv-"+req.SessionID+".json"), b, 0o600)
		}
		if req.TranscriptPath != "" {
			_ = os.WriteFile(filepath.Join(sdir, "seat-"+seatID+".transcript"), []byte(req.TranscriptPath), 0o600)
		}
		return s, true
	}
	return agySitting{}, false
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
