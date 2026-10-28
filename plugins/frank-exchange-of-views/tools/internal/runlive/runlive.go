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
	// LeadSession is the session that leads the run: the one whose `setup` wrote this row, or the
	// one that took the run over since (TakeOver). It is the run's MEMBERSHIP KEY — a hook payload
	// belongs to this run only when its session_id is this one (InferRunDir). Every subagent a
	// session dispatches carries that session's id on its hook payloads, the Workflow tool's seats
	// included, so the run's seats match and an unrelated session's agents — a dev session in a
	// worktree nested under this checkout — match nothing.
	LeadSession string `json:"leadSession"`
}

// RunLive is the marker FILE: the runs `setup` opened in this project and nothing has closed.
//
// A LIST, AND AT MOST ONE OF ITS ROWS IS LIVE. The rows are what was written; which of them still
// counts is RowState's to say, because only `setup` writes a row and only `capture` removes one, and
// capture is optional — a run killed or never captured leaves its row behind. The payload a hook
// receives names no run, but it names its session, so a hook's run is the one live row its session
// leads, or none. `setup` refuses a run while another is live in the project, whatever session leads
// it (#1255): two rows one session leads cannot be told apart, and two runs led by different sessions
// still share the project's run directories and memory.
type RunLive struct {
	Runs []RunLiveMarker `json:"runs"`
}

// OutcomeMark is the file `bench outcome` writes into the run directory once the run's outcome is
// on the record. A row whose run holds it is live only for the bench's assembly sitting
// (StateOf).
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

// CapturedMark is the file `capture` writes into the run directory at its close: the run was CAPTURED.
// It is the one source of that fact — the record refuses every write to a run that holds it
// (record.RunCaptured), the liveness verdict reads CAPTURED from it (record.Assess), and no row whose run
// holds it is live for anyone (StateOf). An outcome is a different fact: the bench has decided the
// run, and its assembly sitting may still be at work.
const CapturedMark = ".run-captured"

// Captured reports whether the run in runDir was captured. An unreadable mark is still a mark: its
// existence is the fact, and the error says it could not be asked.
func Captured(runDir string) (bool, error) {
	_, err := os.Stat(filepath.Join(runDir, CapturedMark))
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	}
	return false, err
}

// BenchAgentType is the agent configuration the engine dispatches the bench under, as the harness
// attests it on a hook payload. The roster names the bench's configuration with it
// (seatclass.Seats), so the configuration is written once.
const BenchAgentType = "frank-exchange-of-views:lead-judge"

// Caller is the agent a hook or verb asks on behalf of: the harness's handle for it and the
// configuration it was dispatched under, both attested on the payload. The zero Caller is a reader
// that is no seat — a pruner, a watcher, an operator.
//
// Session is the session the agent belongs to — the payload's session_id, which every subagent
// shares with the session that dispatched it — and decides which run, if any, is its
// (InferRunDir). A verb has no payload, and takes it from the client's environment (seatenv.Session).
type Caller struct {
	ID      string
	Type    string
	Session string
}

// Leads reports whether the caller's session leads the row's run. A caller with no session leads
// nothing, and neither does a row naming none.
func (c Caller) Leads(m RunLiveMarker) bool {
	return c.Session != "" && c.Session == m.LeadSession
}

// outcomeMark is the mark's content: the agent whose sitting recorded the outcome.
//
// THE RECORDING SEAT KEEPS ITS RUN UNTIL ITS SITTING ENDS. The assembly seat records the outcome
// first and then assembles in the same sitting; if the mark closed the run to it too, its next tool
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
	// ATOMIC, because StateOf reads it on every hook call: a half-written mark parses as no agent,
	// and the recording seat would lose its run mid-sitting.
	return replaceFile(filepath.Join(runDir, OutcomeMark), b)
}

// RowState is whether a marker row is a live run, and if not, why.
type RowState int

const (
	RowLive     RowState = iota // its directory exists, it was not captured, and its outcome is not recorded — or the caller is its assembly sitting's
	RowGone                     // its run directory is not there
	RowDecided                  // its run's outcome is recorded (OutcomeMark), and the caller is no bench agent
	RowCaptured                 // its run was captured (CapturedMark)
)

// Why names a row's state in the words a pruning report needs.
func (s RowState) Why() string {
	switch s {
	case RowGone:
		return "its run directory is gone"
	case RowDecided:
		return "its outcome is recorded"
	case RowCaptured:
		return "it was captured"
	}
	return "it is live"
}

// StateOf is THE liveness predicate: every reader that asks which runs are live asks it here.
//
// caller is the agent asking, the zero Caller for a reader that is no seat. It matters between the
// outcome and the capture only, and there it decides:
//
//   - the agent that recorded the outcome still holds the run, so its own sitting can finish;
//   - so does any agent dispatched under the bench's configuration (BenchAgentType). The one bench
//     sitting the engine dispatches after the outcome is the assembly sitting, and a resume
//     dispatches it again as a FRESH agent the mark does not name — which the record admits
//     (record.requireOpenToSeats takes its assemble-occasion register, its declare, certify and
//     assemble). Without this its hooks bind it to no run: no sitting span, no injected run or
//     identity, and every act it makes fails to find the run it was dispatched into;
//   - both exceptions hold only for an agent of the session that leads the run (Caller.Leads);
//   - every other agent, and every reader that is no seat, sees no live run. That is what the
//     outcome mark is for (#1255): a decided run's row resolving as THE run for every later session
//     in the project, and making a live run beside it ambiguous. A stray agent of another
//     configuration in the project is attributed to nothing, as before; the residue is a bench
//     agent dispatched in the project for some other purpose in that window, whose acts the record
//     refuses all the same unless it sits to assemble.
//
// A captured run (CapturedMark) is live for no one, the recording agent included. A mark that cannot
// be asked about counts as captured, as the record's own refusal does (record.RunCaptured).
//
// A ROW WHOSE RECORD CANNOT BE OPENED STILL COUNTS AS LIVE, deliberately. Its directory is there
// and something may be writing into it; alone, it resolves and the writers that open it fail
// loudly, and beside another live run it is named in the two-live-runs fault. Calling it not live
// would make the hooks skip a run on a read error, which is the silence this predicate replaces.
func StateOf(projectDir string, m RunLiveMarker, caller Caller) RowState {
	dir := Resolve(projectDir, m.RunDir)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return RowGone
	}
	if captured, err := Captured(dir); captured || err != nil {
		return RowCaptured
	}
	b, err := readSharingDelete(filepath.Join(dir, OutcomeMark))
	if errors.Is(err, fs.ErrNotExist) {
		return RowLive
	}
	// BOTH EXCEPTIONS ARE THE RUN'S OWN SESSION'S. A bench agent of another session — a dev session
	// dispatching a lead-judge in a nested worktree — is no assembly sitting of this run.
	if !caller.Leads(m) {
		return RowDecided
	}
	if caller.ID != "" && caller.Type == BenchAgentType {
		return RowLive
	}
	// A mark that is there and cannot be read or parsed still says the outcome is recorded — its
	// existence is the fact, and its content only grants the recording agent's exception.
	var mark outcomeMark
	if err == nil && json.Unmarshal(b, &mark) == nil && caller.ID != "" && mark.AgentID == caller.ID {
		return RowLive
	}
	return RowDecided
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

// PruneRunLive removes every row that is live for no one — its directory gone, its outcome
// recorded, or its run captured — and returns what it removed, so the caller can say so. The file goes when no row is
// left. A marker it cannot read is left alone: nothing in it can be judged. An error means the rows
// it names as pruned are STILL in the file.
func PruneRunLive(projectDir string) (_ []Pruned, err error) {
	release, err := lockMarker(projectDir)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, release()) }()
	runs, readable := readMarkerFile(projectDir)
	if !readable {
		return nil, nil
	}
	var pruned []Pruned
	kept := make([]RunLiveMarker, 0, len(runs))
	for _, m := range runs {
		if s := StateOf(projectDir, m, Caller{}); s != RowLive {
			pruned = append(pruned, Pruned{Row: m, State: s})
			continue
		}
		kept = append(kept, m)
	}
	if len(pruned) > 0 {
		if err := writeRunLive(projectDir, kept); err != nil {
			return pruned, err
		}
	}
	return pruned, nil
}

// writeRunLive replaces the marker with these rows, and removes the file when there are none.
//
// THE ERROR IS THE CALLER'S. A removal or write that failed leaves the rows it meant to take still
// live — every hook keeps attributing to a run its capture closed — so it is never dropped here.
// A marker already absent is the outcome asked for, not a failure.
func writeRunLive(projectDir string, runs []RunLiveMarker) error {
	p := filepath.Join(projectDir, ".claude", "run-live.json")
	if len(runs) == 0 {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	b, err := marshalJSON(RunLive{Runs: runs})
	if err != nil {
		return err
	}
	return replaceFile(p, b)
}

// replaceFile writes b to path ATOMICALLY: a temporary file in the same directory, renamed over
// path. Readers poll these files on every hook call and every dashboard tick; an in-place write
// truncates first, and a read landing between the truncate and the write sees an empty or partial
// file — which readRunLive reports as unreadable, the dashboard watcher as the run's row gone, and
// InferRunDir as a broken marker. A rename shows a reader the old file or the new one, never a mix.
//
// THE RENAME GOES THROUGH os.Root, NOT os.Rename. On Windows os.Rename is MoveFileEx, which refuses
// to replace a file any reader holds open — "Access is denied" even though every reader shares
// delete (openSharingDelete), measured on the Windows runner. Root.Rename asks for POSIX rename
// semantics, under which a target held open with FILE_SHARE_DELETE is replaced and its readers keep
// the old file; TestAnOpenReaderDoesNotBlockReplacement pins that. On Unix both are rename(2).
func replaceFile(path string, b []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, werr := tmp.Write(b)
	err = errors.Join(werr, tmp.Close())
	if err == nil {
		err = os.Chmod(name, 0o644)
	}
	if err == nil {
		err = renameInDir(dir, filepath.Base(name), filepath.Base(path))
	}
	if err != nil {
		// The temporary file is litter beside the marker; a failure to remove it is said too.
		if rerr := os.Remove(name); rerr != nil && !errors.Is(rerr, fs.ErrNotExist) {
			err = errors.Join(err, rerr)
		}
	}
	return err
}

// renameInDir renames one entry of dir over another through an os.Root on dir.
func renameInDir(dir, from, to string) error {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	return errors.Join(root.Rename(from, to), root.Close())
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
//
// It holds the marker's lock shared (MarkerLock), so its open never lands inside a writer's rename;
// a lock it cannot take is a marker it cannot read.
func readRunLive(projectDir string) (runs []RunLiveMarker, readable bool) {
	release, err := rlockMarker(projectDir)
	if err != nil {
		return nil, false
	}
	defer release()
	return readMarkerFile(projectDir)
}

// readMarkerFile is readRunLive without the lock, for a writer that already holds it exclusive.
func readMarkerFile(projectDir string) (runs []RunLiveMarker, readable bool) {
	b, err := readSharingDelete(filepath.Join(projectDir, ".claude", "run-live.json"))
	if err != nil {
		return nil, false
	}
	return decodeRunLive(b)
}

// decodeRunLive is readRunLive's decoder: the rows with a run directory, and whether b parsed.
func decodeRunLive(b []byte) (runs []RunLiveMarker, readable bool) {
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

// WriteRunLiveMarker writes projectDir/.claude/run-live.json (commitment-as-state) and returns its
// path. `now` is injected so the sole non-deterministic field is controllable in tests. An error
// means the run is NOT live: the hooks cannot find it.
//
// leadSession is the session launching the run, and is required: a row naming no session is a run
// no hook payload can belong to.
func WriteRunLiveMarker(projectDir, runDir string, pinnedPaths []string, now time.Time, runID, scriptPath, leadSession string) (_ string, err error) {
	if strings.TrimSpace(leadSession) == "" {
		return "", errors.New("no lead session: a live-run row names the session that leads its run, and the hooks attribute nothing to a row without one")
	}
	dir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, "run-live.json")
	release, err := lockMarker(projectDir)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, release()) }()
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
		LeadSession: leadSession,
	}
	// AN UPSERT, NOT AN OVERWRITE. Overwriting is what made a stale marker self-healing and
	// SILENT; with a list it would also be what makes a second run erase the first one's row,
	// which is the cross-run erasure `capture` already carries a guard against.
	runs, _ := readMarkerFile(projectDir)
	replaced := false
	for i := range runs {
		if SameRun(projectDir, runs[i].RunDir, runDir) {
			// A RE-SETUP IS ITS LEAD SESSION'S OWN. Another session re-running setup on a run that is
			// still live would otherwise replace the run's lead session silently — the take-over
			// without its refusal.
			if lead := runs[i].LeadSession; lead != leadSession && StateOf(projectDir, runs[i], Caller{}) == RowLive {
				return "", &LedElsewhere{Lead: lead}
			}
			runs[i] = marker
			replaced = true
			break
		}
	}
	if !replaced {
		runs = append(runs, marker)
	}
	b, err := marshalJSON(RunLive{Runs: runs})
	if err != nil {
		return "", err
	}
	if err := replaceFile(p, b); err != nil {
		return "", err
	}
	return p, nil
}

// RemoveRunLiveEntry closes ONE run and reports whether it was there and how many remain. An error
// means the file still holds what it held: the run is still live.
//
// The file survives while any other run is open. Removing it wholesale is what let capturing
// run A lift run B's marker — guarded since 2026-08-22 by comparing paths, which is a guard
// where a per-run row is a removal of the class.
func RemoveRunLiveEntry(projectDir, runDir string) (found bool, remaining int, err error) {
	release, err := lockMarker(projectDir)
	if err != nil {
		return false, 0, err
	}
	defer func() { err = errors.Join(err, release()) }()
	runs, _ := readMarkerFile(projectDir)
	kept := runs[:0]
	for _, m := range runs {
		if SameRun(projectDir, m.RunDir, runDir) {
			found = true
			continue
		}
		kept = append(kept, m)
	}
	if err := writeRunLive(projectDir, kept); err != nil {
		return found, len(runs), err
	}
	return found, len(kept), nil
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

// TakeOver makes session `to` lead the run in runDir, in place of session `from`.
//
// A RESUME MAY RUN IN A NEW SESSION — the run's lead session crashed or restarted — and its seats
// then carry the new session's id, which the row does not name: every sitting of the resumed run
// would go unbracketed. The take-over is how the new session says so, and it is REFUSABLE: `from`
// must be the session the row names, so a session cannot take a run from a lead session it did not
// name — the operator reads it off the refusal or the marker and states it. A row that names no lead
// session is taken over from "" (FromSessionArg). A take-over by the session already leading is a
// no-op.
//
// It refuses a run with no row, or whose row is no longer live: there is nothing to lead. It refuses
// a session that already leads ANOTHER live row: two live rows one session leads are Ambiguous, and
// the hooks would attribute neither run.
func TakeOver(projectDir, runDir, from, to string) (prior string, err error) {
	if strings.TrimSpace(to) == "" {
		return "", errors.New("no session to take the run over: the caller's session id is unknown")
	}
	release, err := lockMarker(projectDir)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, release()) }()
	runs, readable := readMarkerFile(projectDir)
	if !readable {
		return "", errors.New("the live-run marker in " + projectDir + " cannot be read")
	}
	target := -1
	for i := range runs {
		if SameRun(projectDir, runs[i].RunDir, runDir) {
			target = i
			break
		}
	}
	if target < 0 {
		return "", errors.New("no live-run row for " + runDir + " in " + filepath.Join(projectDir, ".claude", "run-live.json"))
	}
	prior = runs[target].LeadSession
	if s := StateOf(projectDir, runs[target], Caller{}); s != RowLive && s != RowDecided {
		return prior, errors.New("its row is no longer live — " + s.Why())
	}
	if prior == to {
		return prior, nil
	}
	if from != prior {
		return prior, &LedElsewhere{Lead: prior}
	}
	for i := range runs {
		if i != target && runs[i].LeadSession == to && StateOf(projectDir, runs[i], Caller{}) == RowLive {
			return prior, &LeadsAnother{Run: runs[i].RunDir}
		}
	}
	runs[target].LeadSession = to
	return prior, writeRunLive(projectDir, runs)
}

// LedElsewhere is TakeOver's refusal: the row's lead session is not the session the caller named.
type LedElsewhere struct{ Lead string }

func (e *LedElsewhere) Error() string {
	return "the run's lead session is " + FromSessionArg(e.Lead) + ", not the session named"
}

// LeadsAnother is TakeOver's refusal: the calling session already leads another live run here.
type LeadsAnother struct{ Run string }

func (e *LeadsAnother) Error() string {
	return "the calling session already leads the live run " + e.Run + ", and a session leads one live run"
}

// FromSessionArg is a lead session as `take-over --from-session` takes it: the id, or "" — quoted,
// so it survives the shell — for a row that names none.
func FromSessionArg(lead string) string {
	if lead == "" {
		return `""`
	}
	return lead
}
