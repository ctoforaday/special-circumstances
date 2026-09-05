// Package runlive owns .claude/run-live.json — WHICH RUNS ARE OPEN, and nothing else.
//
// It is a leaf on purpose. The marker's shape used to live in `setup`, which is where it is
// written; but `InferRunDir` also has to READ it, and seat cannot import setup (setup
// imports internal/cli, of which seat is a subpackage). So seat carried its own inline decode
// of the file — the exact defect RunLiveMarker's own comment complains about, that "every
// reader re-declared its own subset", reproduced one package away from the type stating it.
//
// InferRunDir itself now lives HERE rather than in seat (see infer.go), which is the end of that
// story: the reader and the shape it reads are one package, and a caller who wants only "which
// run am I in?" — the PreToolUse hook is the one that matters — no longer takes a cobra command
// tree and a SQLite driver to get it.
//
// MEASURED: making the file a list (#529) broke seat's private decoder and nothing said so.
// Inference silently stopped resolving any run at all, and the only symptom was a verb asking
// for --run — which is what it asks when no run is open, and therefore reads as correct.
package runlive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RunLiveMarker is the marker as a READ value. The writer's struct was anonymous, so every reader
// re-declared its own subset — InferRunDir reads only RunDir, and the shape has never been
// stated in one place a new reader can find (#270).
type RunLiveMarker struct {
	RunDir      string   `json:"runDir"`
	PinnedPaths []string `json:"pinnedPaths"`
	Started     string   `json:"started"`
	// RunID and ScriptPath make a STALE marker actionable rather than litter.
	//
	// A workflow killed by an idle SIGTERM never lifts this marker — it cannot, it is gone — so
	// the marker outlives the run that wrote it and says "live" forever. Naming only a directory
	// tells a reader that something WAS running here; naming the run id and script tells them
	// how to continue it. The difference decides whether a discovered stale marker is a shrug or
	// a resume. Empty when the launcher did not supply them: absent stays absent.
	RunID      string `json:"runId,omitempty"`
	ScriptPath string `json:"scriptPath,omitempty"`
}

// RunLive is the marker FILE: the set of runs currently open.
//
// IT IS A LIST BECAUSE CONCURRENT RUNS ARE SAFE NOW. `setup` bakes the run directory into
// <runDir>/.bin/feov-record (#526), so a seat carries its run in its own environment and a
// second run claiming this file can no longer retarget the first run's seats. A singleton was
// the right shape while the file was the only thing that knew which run a seat belonged to; it
// stopped being the right shape the moment that stopped being true (#529).
type RunLive struct {
	Runs []RunLiveMarker `json:"runs"`
}

// ReadRunLive reports every run the marker says is open, in the order it holds them.
//
// A file it cannot read is not evidence of an open run — it is evidence of a broken file, and
// treating that as "a run is live" would trade a silent hazard for a stuck operator. So an
// unreadable or unparseable marker reports NO runs, and the next write replaces it.
func ReadRunLive(projectDir string) []RunLiveMarker {
	b, err := os.ReadFile(filepath.Join(projectDir, ".claude", "run-live.json"))
	if err != nil {
		return nil
	}
	var f RunLive
	if json.Unmarshal(b, &f) != nil {
		return nil
	}
	out := make([]RunLiveMarker, 0, len(f.Runs))
	for _, m := range f.Runs {
		if strings.TrimSpace(m.RunDir) != "" {
			out = append(out, m)
		}
	}
	return out
}

// ReadRunLiveMarker reports THE open run, if the marker names exactly one usably.
//
// ok=false covers absent, unreadable, unparseable, empty-runDir — and now MORE THAN ONE. That
// last case is the whole point: this is what InferRunDir asks when a verb arrives with no
// run directory, and with two runs open there is no answer, only a guess. Its own rule for an
// unusable marker already says what to do with that — "say nothing rather than guess" — and
// saying nothing is safe now because a wrapped seat never has to ask (#526, #529).
func ReadRunLiveMarker(projectDir string) (RunLiveMarker, bool) {
	runs := ReadRunLive(projectDir)
	if len(runs) != 1 {
		return RunLiveMarker{}, false
	}
	return runs[0], true
}

// sameRun compares two run directories as PATHS, not as strings. The marker stores whatever it
// was given — `research/x` from one invocation, an absolute path from another — so a string
// compare would refuse a re-setup of the very run in progress, which is ordinary and must stay
// idempotent.
func SameRun(projectDir, a, b string) bool {
	abs := func(p string) string {
		if !filepath.IsAbs(p) {
			p = filepath.Join(projectDir, p)
		}
		if r, err := filepath.Abs(filepath.Clean(p)); err == nil {
			return r
		}
		return filepath.Clean(p)
	}
	return strings.EqualFold(abs(a), abs(b))
}

// WriteRunLiveMarker writes projectDir/.claude/run-live.json (commitment-as-state).
// `now` is injected so the sole non-deterministic field is controllable in tests.
func WriteRunLiveMarker(projectDir, runDir string, pinnedPaths []string, now time.Time, runID, scriptPath string) string {
	dir := filepath.Join(projectDir, ".claude")
	os.MkdirAll(dir, 0o755)
	p := filepath.Join(dir, "run-live.json")
	if pinnedPaths == nil {
		pinnedPaths = []string{}
	}
	// THE WRITER USES THE READ TYPE, because it used to re-declare the shape three lines from
	// where RunLiveMarker defines it — two copies of one fact, one file apart, and the comment on
	// that type already complained that readers re-declare their own subsets (#270). A field
	// added to one copy reaches nobody through the other.
	marker := RunLiveMarker{
		RunDir:      runDir,
		PinnedPaths: pinnedPaths,
		Started:     now.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		RunID:       runID,
		ScriptPath:  scriptPath,
	}
	// AN UPSERT, NOT AN OVERWRITE. Overwriting is what made a stale marker self-healing and
	// SILENT; with a list it would also be what makes a second run erase the first one's row,
	// which is the cross-run erasure `capture` already carries a guard against.
	runs := ReadRunLive(projectDir)
	replaced := false
	for i := range runs {
		if SameRun(projectDir, runs[i].RunDir, runDir) {
			runs[i] = marker
			replaced = true
			break
		}
	}
	if !replaced {
		runs = append(runs, marker)
	}
	b, _ := marshalJSON(RunLive{Runs: runs})
	os.WriteFile(p, b, 0o644)
	return p
}

// RemoveRunLiveEntry closes ONE run and reports whether it was there and how many remain.
//
// The file survives while any other run is open. Removing it wholesale is what let capturing
// run A lift run B's marker — guarded since 2026-08-22 by comparing paths, which is a guard
// where a per-run row is a removal of the class.
func RemoveRunLiveEntry(projectDir, runDir string) (found bool, remaining int) {
	runs := ReadRunLive(projectDir)
	kept := runs[:0]
	for _, m := range runs {
		if SameRun(projectDir, m.RunDir, runDir) {
			found = true
			continue
		}
		kept = append(kept, m)
	}
	p := filepath.Join(projectDir, ".claude", "run-live.json")
	if len(kept) == 0 {
		os.Remove(p)
		return found, 0
	}
	b, _ := marshalJSON(RunLive{Runs: kept})
	os.WriteFile(p, b, 0o644)
	return found, len(kept)
}

// marshalJSON is runlive's own indented encoder: the marker is read by humans as often as by
// the tool, and a one-line blob is the shape that gets deleted rather than understood.
func marshalJSON(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
