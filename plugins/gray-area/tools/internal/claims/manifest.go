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

	// Recovered marks a row whose recorded `resolved: false` was contradicted by a
	// fresh stat: the transcript is there now. See ResolveSession.
	Recovered bool `json:"recovered,omitempty"`
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
//
// # THE ROW'S `resolved` IS A ONE-TIME STAT, AND IS RE-CHECKED HERE
//
// Measured (plans/gray-area.md §11.9): on the FIRST SessionStart of a newly
// minted session id, the harness hands over a transcript_path whose file does not
// exist yet — captured 03:43:31Z, present by 03:45:18Z. The capture hook records
// `resolved: false` and nothing ever revisits it, so a single early stat disabled
// `gray-area checkpoint` for that session's entire life. The refusal was correct
// about the row and wrong about the world.
//
// So `resolved` is treated as what it is — an observation made at capture time,
// not a property of the transcript — and re-checked with a fresh stat. A row that
// has since resolved is Recovered, and usable. A row recorded as resolved whose
// file has since vanished is demoted, for the same reason in the other direction.
//
// Selection prefers a row that resolves NOW over a newer one that does not, which
// is the same principle: the question is which transcript can actually be read,
// not which claim about a transcript was written last.
//
// stat reports whether a path is readable; it is injected so the selection logic
// is testable without a filesystem.
func ResolveSession(dir string, glob func(string) ([]string, error), open func(string) ([]byte, error), stat func(string) error) (SessionRow, error) {
	paths, err := glob(filepath.Join(dir, "trajectories-*.jsonl"))
	if err != nil {
		return SessionRow{}, fmt.Errorf("claims: reading manifests in %s: %w", dir, err)
	}
	if len(paths) == 0 {
		return SessionRow{}, fmt.Errorf("claims: no trajectory manifest under %s — gray-area's SessionStart hook has not run in this project, so there is no recorded path to this session's transcript. Pass one explicitly, or install the hook (see the plugin README)", dir)
	}
	sort.Strings(paths)

	// Two candidates: the newest row that reads NOW, and the newest row overall.
	// The second exists only so a total failure can be explained with a real row
	// rather than a bare "nothing found".
	var usable, newest SessionRow
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

			// The recorded flag is an observation; this is the fact.
			readable := r.TranscriptPath != "" && stat(r.TranscriptPath) == nil
			r.Recovered = readable && !r.Resolved
			if !readable && r.Resolved {
				// Recorded as resolved, gone now. Say so rather than hand back a
				// path that will fail on open with a less legible message.
				r.CaptureError = "recorded as resolved at " + r.CapturedAt +
					", but the transcript is not readable now: " + r.TranscriptPath
			}
			r.Resolved = readable

			// Newest wins, by the harness's own capture stamp rather than by file
			// order: manifests are per-session, so file order says nothing about time.
			if newest.CapturedAt == "" || r.CapturedAt > newest.CapturedAt {
				newest = r
			}
			if readable && (usable.CapturedAt == "" || r.CapturedAt > usable.CapturedAt) {
				usable = r
			}
		}
	}
	if usable.TranscriptPath != "" {
		return usable, nil
	}
	best := newest
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
