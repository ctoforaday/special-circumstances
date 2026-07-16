// sc-toolchain-nudge is a prosthetic-conscience SessionStart hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER a session starts, the operator MUST see ONE non-blocking line when a
//	recommended tool is missing — and nothing at all when the toolchain is healthy.
//	It never blocks, never errors the session, and prints at most one line.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

const version = "0.1.0"

type requirements struct {
	Tools []toolchain.Tool `json:"tools"`
}

// nudge formats the single-line warning for missing recommended/required tools.
// Empty string means healthy — print nothing.
func nudge(statuses []toolchain.Status) string {
	var missing []string
	for _, s := range statuses {
		if !s.Found && (s.Tier == "required" || s.Tier == "recommended") {
			missing = append(missing, s.Name)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("prosthetic-conscience: missing tool(s): %s — quality hooks degrade until installed; run /prosthetic-conscience:doctor for install commands.",
		strings.Join(missing, ", "))
}

// liveNudge warns when a research run is live — plugin updates and pushes to pinned
// paths are frozen (marker-and-hook: the commitment is state, not memory).
func liveNudge(m *runlive.Marker) string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("prosthetic-conscience: a research run is LIVE (%s, started %s) — do NOT update plugins and do NOT push to pinned paths until run-capture completes.",
		m.RunDir, m.Started)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-version" {
		fmt.Println("sc-toolchain-nudge", version)
		return
	}

	root := os.Getenv("CLAUDE_PLUGIN_ROOT")
	if root == "" {
		os.Exit(0) // no plugin context — stay silent
	}
	raw, err := os.ReadFile(filepath.Join(root, "requirements.json"))
	if err != nil {
		os.Exit(0)
	}
	var req requirements
	if err := json.Unmarshal(raw, &req); err != nil {
		os.Exit(0)
	}

	if line := nudge(toolchain.Probe(req.Tools)); line != "" {
		fmt.Println(line) // SessionStart stdout is surfaced to the session
	}
	if line := liveNudge(runlive.Read(os.Getenv("CLAUDE_PROJECT_DIR"))); line != "" {
		fmt.Println(line)
	}
	os.Exit(0)
}
