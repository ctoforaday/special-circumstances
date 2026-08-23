// Package seatenv answers "which run am I in?" for a seat verb, from the environment first.
//
// THE MEASURED FAILURE. The first live run recorded 55 tool-call errors in 534 executions and
// TEN of them were one flag: a seat copies the engine's `register --run <dir> --seat-id <id>`
// line and then improvises later verbs, dropping the flag. `InferRunDir` was added for that
// and NEVER FIRES on the real path, because the prompt hands the seat an absolute `--run` at
// every call site and an explicit flag always wins.
//
// Worse than absence is a WRONG value. In the 2026-08-05 smoke `blue-respond-r1` typed
// `special circumstances` — a space where the path has a hyphen — and the tool answered
// "names gap R1-2, which no mint event created". The seat believed the tool, filed friction
// blaming a dangling-reference rule, and abandoned the manifest receipts for R1-3…R1-7. One
// typo cost five receipts and produced a false bug report, because a hand-typed path is an
// unmediated fact and nothing could refuse it (see [[facts-are-fields]]).
//
// So the run directory is INJECTED: the PreToolUse hook rewrites a seat's command to carry
// `export FEOV_RUN='<the live marker's runDir>';` ahead of it. This package is the reading
// half — and the point is not that the environment is another source, it is that a `--run`
// which DISAGREES with the injected value is REFUSED rather than obeyed. A typo now fails
// loudly at the one place that can see both values.
package seatenv

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// Var is the injected variable's name. It is deliberately absent from every --help surface:
// a seat never types it, the hook sets it, and documenting it would invite a seat to pass it
// by hand — which is the hand-typed path this exists to remove.
const Var = "FEOV_RUN"

// VarWrapper is the run directory as the PER-RUN WRAPPER supplies it.
//
// `setup` writes <runDir>/.bin/feov-record, which execs the real binary with this variable
// set, and prints that directory as the `binDir` to hand the workflow. The prompt already
// interpolates "${binDir}/feov-record", so a seat runs the wrapper without knowing it exists
// and without retyping anything.
//
// WHY A SEPARATE VARIABLE FROM Var, which would have been one line shorter. Two reasons, and
// the second is the one that matters:
//
//  1. The hook sets Var. If the wrapper set it too, `run_via` would read "injected" on every
//     register and the hook-absence signal built for #512 would go blind — the exact defect
//     that issue exists to end, reintroduced by the fix for its sibling. Distinct variables
//     keep "wrapper" a PRECISE hook-absence detector on its own: the hook, when it fires,
//     sets Var, and Var wins. No identity pairing needed.
//  2. Two independent carriers of one fact, with a gate between them. #500 wanted that and
//     could only get it by making seats type --run again — reviving the dropped-flag class
//     (10 of 55 tool-call errors on the first live run). The wrapper types it instead, so the
//     gate below costs nothing and finally fires on the real bug: a marker that moved
//     mid-run, which is how run 3's `assemble` wrote a foreign VERIFIED into a live run.
//
// It is absent from every --help surface for the same reason Var is: a seat never types it.
const VarWrapper = "FEOV_RUN_FROM_WRAPPER"

// Resolve returns the run directory for a seat verb.
//
// Order: FEOV_RUN → --run → inference → empty (the caller's own "--run is required").
//
// The DISAGREEMENT is the whole point. When both are present and differ, neither is trusted:
// obeying the flag reinstates the typo, and silently overriding it would make a seat's own
// argument disappear without a word — the failure mode where a seat spends a round arguing
// with a run directory it never chose. Both values are named, so the operator reading the
// error can see which is the typo.
//
// Trailing separators are tolerated on the comparison only: `…/run` and `…/run/` are the same
// directory, and refusing that pair would be pedantry rather than protection. Nothing else is
// normalised — case and separator style are NOT smoothed over, because on the platforms this
// runs on those can be genuinely different paths, and a guess here attaches a seat's events
// to the wrong run, which is the outcome this whole mechanism exists to prevent.
func Resolve(flagRun string, infer func() string) (string, error) {
	dir, _, err := ResolveWithSource(flagRun, infer)
	return dir, err
}

// RunSource says WHICH of Resolve's three paths supplied the run directory.
//
// The order alone cannot answer it afterwards, and the difference is diagnostic. A run whose
// seats all resolve by INFERENCE is a run the PreToolUse hook is not reaching — it injects
// FEOV_RUN on every Bash call in a live run, so inference winning means the hook declined, was
// never invoked, or found no marker from the payload's cwd. Those are three states the hook
// records nowhere, and 2026-08-22_is-7-prime spent an entire run in one of them: fourteen seats,
// zero agent_id, no stderr, and no way to tell a day later which (#512).
//
// Recording the source at `register` is what makes the next one diagnosable. It is a field on
// the record, refusable at the write, rather than a fact recovered afterwards from what is
// missing.
type RunSource string

const (
	// RunFromEnv — the hook injected it. The healthy path.
	RunFromEnv RunSource = "injected"
	// RunFromFlag — the seat typed --run. Legitimate, and the operator path.
	RunFromFlag RunSource = "flag"
	// RunFromWrapper — the per-run wrapper supplied it. The run directory is right (the
	// dispatcher baked it in at setup) and the HOOK IS NOT REACHING THIS RUN, because the hook
	// sets Var and Var wins. A run whose seats all report this is healthy and unhooked.
	RunFromWrapper RunSource = "wrapper"
	// RunFromInference — the TOOL found the marker itself. The seat's call worked, and neither
	// the hook nor a wrapper supplied it, which is the state worth seeing.
	RunFromInference RunSource = "inferred"
	// RunUnresolved — nothing supplied one. The caller raises its own "--run is required".
	RunUnresolved RunSource = "unresolved"
)

// ResolveWithSource is Resolve, and says which path won.
func ResolveWithSource(flagRun string, infer func() string) (string, RunSource, error) {
	env, wrapper := os.Getenv(Var), os.Getenv(VarWrapper)

	// THE HOOK AND THE WRAPPER DISAGREEING IS THE FAILURE THIS PAIR EXISTS TO CATCH, and it is
	// checked FIRST because neither is the seat's doing — a seat cannot cause or fix it, so
	// reporting it as though the seat typed something wrong would send it looking in the wrong
	// place. The wrapper is baked at setup and cannot move; Var is re-derived per call from a
	// marker that can. They differ exactly when the marker moved under a live run, which is how
	// run 3's `assemble` came to write a foreign VERIFIED into a different run's record (#500).
	if env != "" && wrapper != "" && !sameDir(env, wrapper) {
		return "", RunUnresolved, feov.Errorf(feov.Conflict,
			"the run directory injected into this call (%q) disagrees with the run this seat was "+
				"dispatched into (%q). You did not cause this and cannot fix it by changing your command: "+
				"the engine's live-run marker has moved since this run started, which means another run "+
				"claimed it. STOP and record it with the friction verb — continuing would file this seat's "+
				"work against the wrong run",
			env, wrapper)
	}

	// The dispatched run, whichever carrier delivered it. Below here a --run that disagrees is
	// the seat's own typo, and is refused on the same terms either way.
	dispatched, via := env, RunFromEnv
	if dispatched == "" {
		dispatched, via = wrapper, RunFromWrapper
	}
	if dispatched != "" && flagRun != "" && !sameDir(dispatched, flagRun) {
		// Conflict, not Validation: the value is well-formed, it simply contradicts the
		// dispatch. A consumer branching on the code should treat it as "two sources
		// disagree", which is what it is.
		return "", RunUnresolved, feov.Errorf(feov.Conflict,
			"--run %q disagrees with the run this seat was dispatched into (%q). "+
				"The engine supplies the run directory; you do not type it. If the flag is a typo, drop it — "+
				"omitting --run is correct and always right. If you believe the dispatched value is wrong, "+
				"record it with the friction verb rather than working around it",
			flagRun, dispatched)
	}
	if dispatched != "" {
		return dispatched, via, nil
	}
	if flagRun != "" {
		return flagRun, RunFromFlag, nil
	}
	if infer != nil {
		if d := infer(); d != "" {
			return d, RunFromInference, nil
		}
	}
	return "", RunUnresolved, nil
}

// sameDir compares two directory paths, ignoring only a trailing separator.
func sameDir(a, b string) bool { return trimSep(a) == trimSep(b) }

func trimSep(p string) string {
	for len(p) > 1 && (p[len(p)-1] == '/' || p[len(p)-1] == '\\') {
		p = p[:len(p)-1]
	}
	return p
}
