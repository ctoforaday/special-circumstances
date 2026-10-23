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
	"errors"
	"io/fs"
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

// RunLive is the marker FILE: the runs `setup` opened in this project and nothing has closed.
//
// A LIST, AND AT MOST ONE OF ITS ROWS IS LIVE. The rows are what was written; which of them still
// counts is RowState's to say, because only `setup` writes a row and only `capture` removes one, and
// capture is optional — a run killed or never captured leaves its row behind. `setup` refuses a run
// while another is live (#1255): the hooks find a seat's run through this file, and with two live
// runs they cannot tell which a seat belongs to. The payload a hook receives names no run, so the
// answer has to be the one live row or none.
type RunLive struct {
	Runs []RunLiveMarker `json:"runs"`
}

// OutcomeMark is the file `bench outcome` writes into the run directory once the run's outcome is
// on the record. A row whose run holds it no longer counts as live.
//
// A FILE, BECAUSE THE READERS THAT NEED IT CANNOT OPEN THE RECORD. The hooks resolve a run on every
// tool call a seat makes and link no SQLite driver (integration/surface/hookgraph_test.go); a stat
// is what they can afford. The record stays the authority on WHAT the outcome is — this file says
// only that one was recorded, and is written by the verb that records it, after the write succeeds.
//
// WHAT A MISS RETURNS: an outcome recorded with no mark beside it (the write failed after the
// append, which the verb reports) reads as a LIVE run. That is the loud direction — beside another
// run it is the two-live-runs fault, and `setup` refuses naming it — never the silent one.
const OutcomeMark = ".outcome-recorded"

// outcomeMark is the mark's content: the agent whose sitting recorded the outcome.
//
// THE RECORDING SEAT KEEPS ITS RUN UNTIL ITS SITTING ENDS. The assembly seat records the outcome
// first and then assembles in the same sitting; if the mark ended the run for it too, its next tool
// calls would carry no agent identity and its sitting would never close on the record. Agent ids
// are unique per dispatch, so the exception expires with that one sitting.
type outcomeMark struct {
	AgentID string `json:"agentId"`
}

// WriteOutcomeMark records, beside the run, that its outcome is on the record — named for the agent
// that recorded it ("" when no agent identity reached the verb).
func WriteOutcomeMark(runDir, agentID string) error {
	b, err := marshalJSON(outcomeMark{AgentID: agentID})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runDir, OutcomeMark), b, 0o644)
}

// RowState is whether a marker row is a live run, and if not, why.
type RowState int

const (
	RowLive  RowState = iota // its directory exists and its outcome is not recorded
	RowGone                  // its run directory is not there
	RowEnded                 // its run's outcome is recorded (OutcomeMark)
)

// Why names a row's state in the words a pruning report needs.
func (s RowState) Why() string {
	switch s {
	case RowGone:
		return "its run directory is gone"
	case RowEnded:
		return "its outcome is recorded"
	}
	return "it is live"
}

// StateOf is THE liveness predicate: every reader that asks which runs are live asks it here.
//
// agentID is the agent asking, "" for a reader that is no seat. It matters in one case only: the
// agent that recorded the run's outcome still holds the run, so its own sitting can finish.
//
// A ROW WHOSE RECORD CANNOT BE OPENED STILL COUNTS AS LIVE, deliberately. Its directory is there
// and something may be writing into it; alone, it resolves and the writers that open it fail
// loudly, and beside another live run it is named in the two-live-runs fault. Calling it not live
// would make the hooks skip a run on a read error, which is the silence this predicate replaces.
func StateOf(projectDir string, m RunLiveMarker, agentID string) RowState {
	dir := Resolve(projectDir, m.RunDir)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return RowGone
	}
	b, err := readSharingDelete(filepath.Join(dir, OutcomeMark))
	if errors.Is(err, fs.ErrNotExist) {
		return RowLive
	}
	// A mark that is there and cannot be read or parsed still says the outcome is recorded — its
	// existence is the fact, and its content only grants the recording agent's exception.
	var mark outcomeMark
	if err == nil && json.Unmarshal(b, &mark) == nil && agentID != "" && mark.AgentID == agentID {
		return RowLive
	}
	return RowEnded
}

// Resolve is a row's run directory as a path: a relative runDir is relative to the project the
// marker is in.
func Resolve(projectDir, runDir string) string {
	if filepath.IsAbs(runDir) {
		return runDir
	}
	return filepath.Join(projectDir, runDir)
}

// RowOf is this run's row in the project's marker, if it has one.
func RowOf(projectDir, runDir string) (RunLiveMarker, bool) {
	for _, m := range ReadRunLive(projectDir) {
		if SameRun(projectDir, m.RunDir, runDir) {
			return m, true
		}
	}
	return RunLiveMarker{}, false
}

// Pruned is a row PruneRunLive removed, and why it was no longer live.
type Pruned struct {
	Row   RunLiveMarker
	State RowState
}

// PruneRunLive removes every row that is live for no one — its directory gone, or its outcome
// recorded — and returns what it removed, so the caller can say so. The file goes when no row is
// left. A marker it cannot read is left alone: nothing in it can be judged.
func PruneRunLive(projectDir string) []Pruned {
	runs, readable := readRunLive(projectDir)
	if !readable {
		return nil
	}
	var pruned []Pruned
	kept := make([]RunLiveMarker, 0, len(runs))
	for _, m := range runs {
		if s := StateOf(projectDir, m, ""); s != RowLive {
			pruned = append(pruned, Pruned{Row: m, State: s})
			continue
		}
		kept = append(kept, m)
	}
	if len(pruned) > 0 {
		writeRunLive(projectDir, kept)
	}
	return pruned
}

// writeRunLive replaces the marker with these rows, and removes the file when there are none.
func writeRunLive(projectDir string, runs []RunLiveMarker) {
	p := filepath.Join(projectDir, ".claude", "run-live.json")
	if len(runs) == 0 {
		os.Remove(p)
		return
	}
	b, _ := marshalJSON(RunLive{Runs: runs})
	os.WriteFile(p, b, 0o644)
}

// ReadRunLive reports every row the marker holds, in order — live or not; StateOf says which are.
//
// A file it cannot read is not evidence of an open run — it is evidence of a broken file, and
// treating that as "a run is live" would trade a silent hazard for a stuck operator. So an
// unreadable or unparseable marker reports NO runs, and the next write replaces it.
func ReadRunLive(projectDir string) []RunLiveMarker {
	runs, _ := readRunLive(projectDir)
	return runs
}

// readRunLive is ReadRunLive plus whether the file could be READ at all. ReadRunLive's callers want
// the runs and treat a broken file as no runs, which is right for them; InferRunDir must also be
// able to say that a marker is BROKEN rather than empty, because only one of those is a fault a
// human should be told about. One decoder serves both, so the two can never disagree about the
// shape of the file.
func readRunLive(projectDir string) (runs []RunLiveMarker, readable bool) {
	b, err := readSharingDelete(filepath.Join(projectDir, ".claude", "run-live.json"))
	if err != nil {
		return nil, false
	}
	var f RunLive
	if json.Unmarshal(b, &f) != nil {
		return nil, false
	}
	out := make([]RunLiveMarker, 0, len(f.Runs))
	for _, m := range f.Runs {
		if strings.TrimSpace(m.RunDir) != "" {
			out = append(out, m)
		}
	}
	return out, true
}

// SameRun compares two run directories as PATHS, not as strings. The marker stores whatever it
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
	writeRunLive(projectDir, kept)
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
