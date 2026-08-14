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
package toolchainnudge

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
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
			// The tool's own purpose, so the line says WHICH capability is absent rather
			// than only which binary is. It comes from requirements.json, so it stays
			// true as the manifest changes instead of drifting into a second description.
			if s.Purpose != "" {
				missing = append(missing, s.Name+" ("+s.Purpose+")")
				continue
			}
			missing = append(missing, s.Name)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	// Name WHICH capability is lost, not just that something degraded: "quality hooks
	// degrade" tells an agent nothing it can weigh. Each tool's own purpose comes from
	// requirements.json, so the line stays true as the manifest changes.
	return fmt.Sprintf("prosthetic-conscience: missing tool(s) — %s. Those checks are SKIPPED, not failing: work continues and nothing is blocked, but their coverage is absent, so do not read a clean run as proof they passed. Install with /prosthetic-conscience:doctor --fix, or carry on and say in your summary which coverage was missing.",
		strings.Join(missing, "; "))
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
// stdin is read for ONE field: the payload's cwd, the second place the harness states
// where the project is. This hook took the project root from the environment alone, and
// with CLAUDE_PROJECT_DIR unset the live-run nudge went silent — a research run in
// progress and nothing said so, which is the one job that nudge has.
func run(args []string, stdin io.Reader, stdout io.Writer, pluginRoot, projectDir string, probe func([]toolchain.Tool) []toolchain.Status) int {
	if len(args) > 0 && args[0] == "-version" {
		fmt.Fprintln(stdout, "sc-toolchain-nudge", version)
		return 0
	}

	if stdin != nil {
		var in struct {
			CWD string `json:"cwd"`
		}
		raw, _ := io.ReadAll(stdin)
		_ = json.Unmarshal(raw, &in)
		projectDir = hookenv.ProjectDir(projectDir, in.CWD)
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

// Unit exposes the nudge to the merged SessionStart binary.
//
// It returns TEXT rather than writing it. The standalone binary printed the line straight to
// stdout on the strength of "SessionStart stdout reaches the session" — a claim that is NOT
// in plans/hook-surface-spike.md §5's verified list. What IS verified there is
// additionalContext ("hook_additional_context attachment, transcript line 42"), and one
// process cannot emit both a bare line and a JSON document on stdout without corrupting the
// document. So the merge puts this on the channel that was actually measured.
func Unit(pluginRoot string, probe func([]toolchain.Tool) []toolchain.Status) hookunit.Unit {
	return hookunit.Unit{
		Name: "sc-toolchain-nudge",
		Run: func(c *hookunit.Ctx) hookunit.Result {
			var lines []string
			// A missing/unreadable/malformed manifest degrades to silence, never to an
			// error: a SessionStart hook that fails is a session that fails.
			if pluginRoot != "" {
				if raw, err := os.ReadFile(filepath.Join(pluginRoot, "requirements.json")); err == nil {
					var req requirements
					if err := json.Unmarshal(raw, &req); err == nil {
						if line := nudge(probe(req.Tools)); line != "" {
							lines = append(lines, line)
						}
					}
				}
			}
			// INDEPENDENT of the manifest: a live research run must be announced even when
			// there is no plugin root to read requirements from.
			if line := liveNudge(runlive.Read(c.ProjectDir)); line != "" {
				lines = append(lines, line)
			}
			return hookunit.Result{Name: "sc-toolchain-nudge", Stdout: strings.Join(lines, "\n")}
		},
	}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout,
		os.Getenv("CLAUDE_PLUGIN_ROOT"), os.Getenv("CLAUDE_PROJECT_DIR"), toolchain.Probe)
}
