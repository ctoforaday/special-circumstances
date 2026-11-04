package trajectory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// TranscriptFile represents one discovered session or subagent transcript.
type TranscriptFile struct {
	Path      string `json:"path"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"` // empty for root session
	Engine    string `json:"engine"`   // "claude" | "antigravity"
}

// DefaultClaudeProjectsDir returns the default Claude Code projects directory.
func DefaultClaudeProjectsDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".claude", "projects")
	}
	return ""
}

// DefaultAntigravityBrainDirs returns the default Antigravity brain directories.
func DefaultAntigravityBrainDirs() []string {
	var dirs []string
	if home, err := os.UserHomeDir(); err == nil {
		candidates := []string{
			filepath.Join(home, ".gemini", "antigravity-cli", "brain"),
			filepath.Join(home, ".gemini", "antigravity", "brain"),
			filepath.Join(home, ".gemini", "antigravity-ide", "brain"),
		}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && st.IsDir() {
				dirs = append(dirs, c)
			}
		}
	}
	return dirs
}

// DiscoverAll finds all transcript files across both Claude Code and Antigravity roots.
func DiscoverAll(claudeProjectsDir string, agyBrainDirs []string) ([]TranscriptFile, error) {
	var out []TranscriptFile

	if claudeProjectsDir != "" {
		cFiles, err := DiscoverClaude(claudeProjectsDir)
		if err == nil {
			out = append(out, cFiles...)
		}
	}

	for _, bDir := range agyBrainDirs {
		aFiles, err := DiscoverAntigravity(bDir)
		if err == nil {
			out = append(out, aFiles...)
		}
	}

	return out, nil
}

// DiscoverClaude finds all Claude Code transcript files under projectsRoot.
func DiscoverClaude(projectsRoot string) ([]TranscriptFile, error) {
	var out []TranscriptFile
	err := filepath.WalkDir(projectsRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if tf, ok := AttributeClaudeFile(projectsRoot, p); ok {
			out = append(out, tf)
		}
		return nil
	})
	return out, err
}

// AttributeClaudeFile attributes one file by the Claude 3-tier rule.
func AttributeClaudeFile(projectsRoot, p string) (TranscriptFile, bool) {
	if !strings.HasSuffix(p, ".jsonl") {
		return TranscriptFile{}, false
	}
	rel, err := filepath.Rel(projectsRoot, p)
	if err != nil {
		return TranscriptFile{}, false
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	switch {
	case len(parts) == 2:
		return TranscriptFile{
			Path:      p,
			SessionID: strings.TrimSuffix(parts[1], ".jsonl"),
			Engine:    "claude",
		}, true
	case len(parts) == 4 && parts[2] == "subagents":
		return TranscriptFile{
			Path:      p,
			SessionID: parts[1],
			AgentID:   agentIDOf(parts[3]),
			Engine:    "claude",
		}, true
	case len(parts) == 6 && parts[2] == "subagents" && parts[3] == "workflows":
		return TranscriptFile{
			Path:      p,
			SessionID: parts[1],
			AgentID:   agentIDOf(parts[5]),
			Engine:    "claude",
		}, true
	}
	return TranscriptFile{}, false
}

func agentIDOf(base string) string {
	return strings.TrimSuffix(strings.TrimPrefix(base, "agent-"), ".jsonl")
}

// DiscoverAntigravity finds all transcript files under an Antigravity brain directory.
func DiscoverAntigravity(brainRoot string) ([]TranscriptFile, error) {
	entries, err := os.ReadDir(brainRoot)
	if err != nil {
		return nil, err
	}

	// Map of subagent conversation ID -> parent session ID
	subagentParents := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		parentID := e.Name()
		subagentDir := filepath.Join(brainRoot, parentID, ".system_generated", "subagents")
		if subEntries, err := os.ReadDir(subagentDir); err == nil {
			for _, se := range subEntries {
				if strings.HasSuffix(se.Name(), ".json") {
					subID := strings.TrimSuffix(se.Name(), ".json")
					// Read descriptor to ensure valid ID
					b, err := os.ReadFile(filepath.Join(subagentDir, se.Name()))
					if err == nil {
						var desc struct {
							ConversationID string `json:"conversationId"`
						}
						if json.Unmarshal(b, &desc) == nil && desc.ConversationID != "" {
							subID = desc.ConversationID
						}
					}
					subagentParents[subID] = parentID
				}
			}
		}
	}

	var out []TranscriptFile
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		convID := e.Name()
		tPath := filepath.Join(brainRoot, convID, ".system_generated", "logs", "transcript.jsonl")
		if _, err := os.Stat(tPath); err != nil {
			tPath = filepath.Join(brainRoot, convID, "transcript.jsonl")
			if _, err := os.Stat(tPath); err != nil {
				continue
			}
		}

		sessionID := convID
		agentID := ""
		if parent, ok := subagentParents[convID]; ok {
			sessionID = parent
			agentID = convID
		}

		out = append(out, TranscriptFile{
			Path:      tPath,
			SessionID: sessionID,
			AgentID:   agentID,
			Engine:    "antigravity",
		})
	}

	return out, nil
}

// DiscoverSession finds transcripts for one specific session across Claude and Antigravity roots.
func DiscoverSession(claudeProjectsDir string, agyBrainDirs []string, sessionID string) ([]TranscriptFile, error) {
	all, err := DiscoverAll(claudeProjectsDir, agyBrainDirs)
	if err != nil {
		return nil, err
	}
	var out []TranscriptFile
	for _, f := range all {
		if f.SessionID == sessionID {
			out = append(out, f)
		}
	}
	return out, nil
}
