// Package runlive reads the .run-live marker — the commitment-as-state pattern from the
// design-by-contract rule: promises about future behavior (push freeze on pinned paths,
// no plugin updates mid-run) are recorded as observable state by setup-research-run.mjs, consulted
// by deterministic guards here, and removed by capture-research-run.mjs. Memory is not an
// enforcement mechanism; this file is.
package runlive

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Marker is the parsed .claude/run-live.json.
type Marker struct {
	RunDir      string   `json:"runDir"`
	PinnedPaths []string `json:"pinnedPaths"`
	Started     string   `json:"started"`
}

// Read returns the live marker under projectDir, or nil when no run is live
// (missing or malformed marker both mean "not live" — guards fail open).
func Read(projectDir string) *Marker {
	if projectDir == "" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(projectDir, ".claude", "run-live.json"))
	if err != nil {
		return nil
	}
	var m Marker
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return &m
}
