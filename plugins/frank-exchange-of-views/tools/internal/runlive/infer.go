package runlive

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
)

// InferRunDir answers "which run am I in?" from the live-run marker instead of
// requiring every call to say so.
//
// The first live run measured 55 tool-call errors in 534 executions, and TEN of them
// were this one flag: a seat copies the engine's `register --run <dir> --seat-id <id>`
// line, then improvises later verbs and drops the flags. Shell state does not persist
// between tool calls, so the seat cannot export it once; there is no per-agent
// environment variable to carry it; and the engine is a sandboxed script that cannot
// set one. But the answer is already on disk — setup writes .claude/run-live.json with
// the runDir, and the hook guards already read it.
//
// An explicit --run always wins: inference is a fallback for the seat that forgot, not
// a new source of truth. The marker's runDir is project-relative, so it resolves
// against the directory holding .claude/, and an inferred directory that does not
// exist is discarded rather than passed on — a wrong answer here would attach a seat's
// events to the wrong run, which is worse than the error it replaces.
//
// # Why it lives in runlive
//
// It was in internal/cli/seat, which is a cobra command tree. Nothing about walking up to a
// run-live marker needs one, and the cost was paid by an unrelated caller: the PreToolUse hook
// imported this function and nothing else from that package, and inherited cobra, protobuf and
// a SQLite driver behind it. Measured on an idle 4-core box — feov-pretooluse at 3.99 ms and
// 13.3 MB against a 1.15 ms empty-binary floor, with `internal/record` accounting for the whole
// excess, on a binary that fires once per Bash call (#684 F2). The comment above already said
// "the hook guards already read it"; this puts the function where they can.
// Inference says WHY InferRunDir found the directory it did, or none.
//
// A bare "" used to answer every one of these, and they are not the same: two of them are the
// ordinary shape of a session (no run; no live run), and three are FAULTS — a marker that cannot be
// read, a marker with more than one live run, and nowhere to look. A caller that must tell a human
// about a fault cannot do it from a zero value it shares with the healthy cases.
type Inference int

const (
	NoMarker    Inference = iota // no marker between here and the root: not in a run
	Resolved                     // exactly one live run
	NoRunOpen                    // a marker, naming no live run (StateOf)
	Ambiguous                    // more than one live run: no seat's run can be told — a fault
	Unreadable                   // a marker that cannot be read or parsed — a fault
	Unlocatable                  // no cwd in the payload, no CLAUDE_PROJECT_DIR, and no working directory — a fault
)

// Fault reports whether the inference failed for a reason a human should hear about.
//
// TWO LIVE RUNS IS A FAULT. A hook's payload names no run, so with two live the hooks bracket no
// sitting, count no call and inject no run for either one — every seat of both runs acts unbracketed
// and uncapped. `setup` refuses a second live run, so this is reached only by a marker written
// around it, and it is said rather than absorbed.
func (i Inference) Fault() bool { return i == Unreadable || i == Ambiguous || i == Unlocatable }

// Inferred is InferRunDir's answer: the directory, why, and where the marker was — the project a
// fault belongs to, which is its scope on the failure record. Live names the live runs when there
// is more than one, for the fault's own words.
type Inferred struct {
	Dir       string
	Why       Inference
	MarkerDir string
	Live      []string
}

// StageUnusable is the failure-record stage for a marker that is a fault (Inference.Fault). It is
// declared here, beside the marker, because two unrelated readers hit it — the PreToolUse hook and
// the subagent sitting hooks — and the stage names one fact about one file.
const StageUnusable hookfailures.Stage = "run-dir-unusable"

// caller is the agent asking (the zero Caller for none) — StateOf's exceptions between a run's
// outcome and its capture, the bench's assembly sitting, are answered through it.
func InferRunDir(start string, caller Caller) Inferred {
	dir := start
	if dir == "" {
		if p := os.Getenv("CLAUDE_PROJECT_DIR"); p != "" {
			dir = p
		} else if wd, err := os.Getwd(); err == nil {
			dir = wd
		} else {
			// Nowhere to start the search from. Falling through made this read as NoMarker — "not in
			// a run" — which is the one answer it cannot honestly give.
			return Inferred{Why: Unlocatable}
		}
	}
	for i := 0; dir != "" && i < 12; i++ {
		marker := filepath.Join(dir, ".claude", "run-live.json")
		if _, err := os.Stat(marker); err == nil {
			// THROUGH THE PACKAGE THAT OWNS THE FILE, never a private decode of it.
			//
			// This used to unmarshal its own `struct{ RunDir string }` — a second reader of a
			// shape stated elsewhere, which is the defect RunLiveMarker's own comment names.
			// It broke the moment the marker became a list (#529): the private decoder found no
			// `runDir` key, inference stopped resolving anything, and the LIVE symptom was a
			// verb asking for --run — which is exactly what it asks when no run is open, so the
			// regression read as correct behaviour. Five tests caught it; driving it by hand did
			// not, and would not have.
			runs, readable := readRunLive(dir)
			if !readable {
				return Inferred{Why: Unreadable, MarkerDir: dir}
			}
			// ONLY LIVE ROWS COUNT. A row whose outcome is recorded or whose directory is gone is
			// residue of a run nothing captured; letting it resolve attributed every later session in this project
			// to a dead run, and letting it count made a live run beside it ambiguous.
			var live []string
			for _, m := range runs {
				if StateOf(dir, m, caller) == RowLive {
					live = append(live, Resolve(dir, m.RunDir))
				}
			}
			switch len(live) {
			case 0:
				return Inferred{Why: NoRunOpen, MarkerDir: dir}
			case 1:
				return Inferred{Dir: live[0], Why: Resolved, MarkerDir: dir}
			}
			return Inferred{Why: Ambiguous, MarkerDir: dir, Live: live}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return Inferred{Why: NoMarker}
}

// FaultDetail names which fault an inference hit, in the words its fix needs. ONE wording, because
// the PreToolUse hook and the sitting hooks both report it, and a human reading the record should
// not have to reconcile two descriptions of one broken file.
func FaultDetail(i Inferred) string {
	switch i.Why {
	case Unreadable:
		return "the live-run marker in " + i.MarkerDir + " cannot be read"
	case Ambiguous:
		return "the live-run marker in " + i.MarkerDir + " names " + strconv.Itoa(len(i.Live)) + " live runs (" +
			strings.Join(i.Live, ", ") + "), so no seat's run can be told apart and none is used — close all but one " +
			"with `feov-record capture <runDir> <transcript dir> --seat-id operator`, or run the others from a separate checkout"
	case Unlocatable:
		return "there is nowhere to look for a live-run marker (no cwd in the payload, no CLAUDE_PROJECT_DIR, and no working directory)"
	}
	return ""
}
