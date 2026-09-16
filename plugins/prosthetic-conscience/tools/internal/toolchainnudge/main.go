// sc-toolchain-nudge is a prosthetic-conscience SessionStart nudge, shipped as a UNIT of
// the merged sc-sessionstart binary — not as a binary of its own (#201).
//
// Its standalone run/Main pair survived that merge with a second copy of the manifest
// read and the live-run check, built by nothing, and the package's tests drove it rather
// than the unit that actually runs. Both are gone; the tests drive Unit.
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
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

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

// The stages this unit records.
const (
	// StageManifest: the tool manifest cannot be read, does not parse, or has no plugin root to be
	// read from. Each leaves the nudge silent about which tools are missing.
	StageManifest hookfailures.Stage = "toolchain-manifest"
	// StageMarkerUnparsable is the same stage pushfreezeguard records, named once per package that
	// can hit it: SessionStart reads the marker too, and an unparsable one silenced this nudge.
	StageMarkerUnparsable hookfailures.Stage = "run-live-parse"
)

// liveNudge warns when a research run is live — plugin updates and pushes to pinned
// paths are frozen (marker-and-hook: the commitment is state, not memory).
func liveNudge(st runlive.State) string {
	// Describe is empty exactly when nothing is live, and names the runs — or names the
	// fact that the marker's shape is not understood — when something is. The phrasing
	// lives in runlive so this guard and pushfreezeguard cannot drift apart in what they
	// call the same state, which is the smaller cousin of the drift that emptied both.
	desc := st.Describe()
	if desc == "" {
		return ""
	}
	return fmt.Sprintf("prosthetic-conscience: %s — do NOT update plugins and do NOT push to pinned paths until run-capture completes.", desc)
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
			// error: a SessionStart hook that fails is a session that fails. THE SILENCE IS NOW
			// RECORDED, because it was total — the nudge reported no tools AND no reason, which
			// reads exactly like a machine with every tool installed.
			if pluginRoot == "" {
				// No plugin root at all: the manifest cannot even be located, and the nudge was
				// silent about exactly the tools it exists to name.
				c.Rec.Fail(StageManifest, "no plugin root (CLAUDE_PLUGIN_ROOT unset) — the toolchain nudge cannot read requirements.json, so it cannot say which tools are missing")
			} else {
				manifest := filepath.Join(pluginRoot, "requirements.json")
				switch raw, err := os.ReadFile(manifest); {
				case err != nil:
					c.Rec.Fail(StageManifest, "cannot read "+manifest+": "+err.Error()+
						" — the toolchain nudge cannot say which tools are missing")
				default:
					var req requirements
					if err := json.Unmarshal(raw, &req); err != nil {
						c.Rec.Fail(StageManifest, manifest+" does not parse: "+err.Error()+
							" — the toolchain nudge cannot say which tools are missing")
						break
					}
					c.Rec.OK(StageManifest)
					if line := nudge(probe(req.Tools)); line != "" {
						lines = append(lines, line)
					}
				}
			}
			// INDEPENDENT of the manifest: a live research run must be announced even when
			// there is no plugin root to read requirements from.
			st := runlive.Read(c.ProjectDir)
			if st.Unparsable {
				// An unparsable marker reads as "nothing live" to Describe, so this nudge said
				// nothing about a run that may well be live. SessionStart displays; say why.
				c.Rec.FailIn(StageMarkerUnparsable, c.ProjectDir, "the live-run marker exists and does not parse — "+
					"whether a research run is live (and a freeze in force) cannot be told")
			} else {
				c.Rec.OKIn(StageMarkerUnparsable, c.ProjectDir)
			}
			if line := liveNudge(st); line != "" {
				lines = append(lines, line)
			}
			return hookunit.Result{Name: "sc-toolchain-nudge", Stdout: strings.Join(lines, "\n")}
		},
	}
}
