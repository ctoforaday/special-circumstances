package trajectory

import (
	"encoding/json"
	"strings"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// NormalizeToolName returns the canonical tool name.
func NormalizeToolName(raw string) string {
	return hookcore.CanonicalToolName(raw)
}

// NormalizeToolArgs normalizes tool argument keys across Claude Code and Antigravity.
func NormalizeToolArgs(toolName string, in map[string]any) json.RawMessage {
	norm := make(map[string]any, len(in)+4)
	for k, v := range in {
		norm[k] = v
	}

	// Command normalization
	if cmd, ok := in["command"].(string); ok {
		norm["CommandLine"] = cmd
	} else if cmd, ok := in["CommandLine"].(string); ok {
		norm["command"] = cmd
	}

	// Path normalization
	var path string
	for _, k := range []string{"file_path", "Path", "TargetFile", "AbsolutePath", "path"} {
		if s, ok := in[k].(string); ok && s != "" {
			path = s
			break
		}
	}
	if path != "" {
		norm["file_path"] = path
		norm["path"] = path
		norm["Path"] = path
	}

	// Content normalization
	if c, ok := in["content"].(string); ok {
		norm["Content"] = c
		norm["CodeContent"] = c
	} else if c, ok := in["Content"].(string); ok {
		norm["content"] = c
		norm["CodeContent"] = c
	}

	b, _ := json.Marshal(norm)
	return b
}

// TargetOf extracts a bounded target from a tool input payload (matching catalogue convention).
func TargetOf(tool string, input json.RawMessage) string {
	if len(input) == 0 {
		return ""
	}
	var m map[string]any
	if json.Unmarshal(input, &m) != nil {
		return ""
	}
	for _, k := range []string{"file_path", "Path", "path", "notebook_path", "pattern", "command", "CommandLine", "url", "query"} {
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

// WriteTarget returns the path a file-modifying tool acted on, or "".
func WriteTarget(e Event) string {
	switch e.ToolName {
	case "Write", "Edit", "NotebookEdit", "MultiEdit":
	default:
		return ""
	}
	if len(e.ToolInput) == 0 {
		return ""
	}
	var in struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		PathCap  string `json:"Path"`
	}
	if json.Unmarshal(e.ToolInput, &in) != nil {
		return ""
	}
	if in.FilePath != "" {
		return in.FilePath
	}
	if in.Path != "" {
		return in.Path
	}
	return in.PathCap
}

// BashCommand returns a Bash event's command string, or "".
func BashCommand(e Event) string {
	if e.ToolName != "Bash" || len(e.ToolInput) == 0 {
		return ""
	}
	var in struct {
		Command     string `json:"command"`
		CommandLine string `json:"CommandLine"`
	}
	if json.Unmarshal(e.ToolInput, &in) != nil {
		return ""
	}
	if in.Command != "" {
		return in.Command
	}
	return in.CommandLine
}
