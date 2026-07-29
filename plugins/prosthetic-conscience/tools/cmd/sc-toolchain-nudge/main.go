// sc-toolchain-nudge is a prosthetic-conscience SessionStart hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER a session starts, the operator MUST see ONE non-blocking line when a
//	recommended tool is missing, and ONE more when a research run is live — and
//	nothing at all when the toolchain is healthy and no run is live. It never blocks
//	and never errors the session; each nudge is exactly one line.
//
// (The "at most one line" wording predates the live-run nudge, which is a second,
// independent line; the invariant that survived is one line PER nudge.)
package main

import (
	"encoding/json"
	"fmt"
	"io"
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
		// NotApplicable tools are absent by design in this environment (gh in a
		// cloud session). Nudging about them would move sc-doctor's false alarm
		// into the one line every session starts with — the most expensive place
		// in the suite for a warning nobody can act on.
		if s.NotApplicable {
			continue
		}
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

// run is the hook with its process boundary passed in. probe is a parameter rather
// than a direct toolchain.Probe call so a test's verdict does not depend on what
// happens to be installed on the machine running it.
func run(args []string, stdout io.Writer, pluginRoot, projectDir string, probe func([]toolchain.Tool) []toolchain.Status) int {
	if len(args) > 0 && args[0] == "-version" {
		fmt.Fprintln(stdout, "sc-toolchain-nudge", version)
		return 0
	}

	// A missing/unreadable/malformed manifest degrades to silence, never to an error:
	// a SessionStart hook that fails is a session that fails.
	if pluginRoot != "" {
		if raw, err := os.ReadFile(filepath.Join(pluginRoot, "requirements.json")); err == nil {
			var req requirements
			if err := json.Unmarshal(raw, &req); err == nil {
				if line := nudge(probe(req.Tools)); line != "" {
					fmt.Fprintln(stdout, line) // SessionStart stdout reaches the session
				}
			}
		}
	}

	// The live-run nudge is INDEPENDENT of the manifest: a live research run must be
	// announced even when there is no plugin root to read requirements from.
	if line := liveNudge(runlive.Read(projectDir)); line != "" {
		fmt.Fprintln(stdout, line)
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout,
		os.Getenv("CLAUDE_PLUGIN_ROOT"), os.Getenv("CLAUDE_PROJECT_DIR"), toolchain.Probe))
}
