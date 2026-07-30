package claims

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// SessionRow is the manifest entry that answers "where is THIS session's
// transcript" — written at SessionStart from what the harness handed over.
type SessionRow struct {
	Kind           string `json:"kind"`
	CapturedAt     string `json:"captured_at"`
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Resolved       bool   `json:"resolved"`
	CaptureError   string `json:"capture_error,omitempty"`

	// Where this row was read from, so "the tool picked a transcript for you" is a
	// claim the reader can check like any other.
	Manifest string `json:"manifest"`
	Line     int    `json:"manifest_line"`
}

// ResolveSession returns the newest usable session row under the project's own
// manifest directory.
//
// IT READS ONLY `.claude/gray-area/`, WHICH THIS PLUGIN WRITES. The obvious
// alternative — globbing ~/.claude/projects/ for the newest file — is ruled out
// by plans/gray-area.md §3 and Phase 1's acceptance criterion: the design is that
// the harness hands the path over deterministically, and guessing inside a store
// this plugin does not own is the attribution failure the manifest exists to
// remove. Listing our own output directory is a different act from guessing
// inside somebody else's.
//
// An unresolved row (one whose transcript did not stat) is still RETURNED rather
// than skipped, so the caller can say WHY it is unusable instead of reporting the
// same "nothing found" it would report for an empty directory.
func ResolveSession(dir string, glob func(string) ([]string, error), open func(string) ([]byte, error)) (SessionRow, error) {
	paths, err := glob(filepath.Join(dir, "trajectories-*.jsonl"))
	if err != nil {
		return SessionRow{}, fmt.Errorf("claims: reading manifests in %s: %w", dir, err)
	}
	if len(paths) == 0 {
		return SessionRow{}, fmt.Errorf("claims: no trajectory manifest under %s — gray-area's SessionStart hook has not run in this project, so there is no recorded path to this session's transcript. Pass one explicitly, or install the hook (see the plugin README)", dir)
	}
	sort.Strings(paths)

	var best SessionRow
	for _, p := range paths {
		body, err := open(p)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(strings.NewReader(string(body)))
		sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		n := 0
		for sc.Scan() {
			n++
			var r SessionRow
			if json.Unmarshal(sc.Bytes(), &r) != nil || r.Kind != "session" {
				continue
			}
			r.Manifest, r.Line = p, n
			// Newest wins, by the harness's own capture stamp rather than by file
			// order: manifests are per-session, so file order says nothing about time.
			if best.CapturedAt == "" || r.CapturedAt > best.CapturedAt {
				best = r
			}
		}
	}
	if best.TranscriptPath == "" {
		return SessionRow{}, fmt.Errorf("claims: manifests exist under %s but none carries a session row — only subagent rows were written, which means gray-area's SessionStart hook is not wired (SubagentStop alone never records the MAIN session's transcript)", dir)
	}
	return best, nil
}

// ManifestDir is where this plugin's own manifests live, relative to a project.
func ManifestDir(projectDir string) string {
	return filepath.Join(projectDir, ".claude", "gray-area")
}

// GlobFS adapts filepath.Glob for callers that want the real filesystem.
var GlobFS = filepath.Glob
