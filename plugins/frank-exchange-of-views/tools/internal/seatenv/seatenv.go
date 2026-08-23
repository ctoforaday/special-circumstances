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
	// RunFromInference — the TOOL found the marker itself. The seat's call worked, and the hook
	// did not supply it, which is the state worth seeing.
	RunFromInference RunSource = "inferred"
	// RunUnresolved — nothing supplied one. The caller raises its own "--run is required".
	RunUnresolved RunSource = "unresolved"
)

// ResolveWithSource is Resolve, and says which path won.
func ResolveWithSource(flagRun string, infer func() string) (string, RunSource, error) {
	env := os.Getenv(Var)
	if env != "" && flagRun != "" && !sameDir(env, flagRun) {
		// Conflict, not Validation: the value is well-formed, it simply contradicts the
		// dispatch. A consumer branching on the code should treat it as "two sources
		// disagree", which is what it is.
		return "", RunUnresolved, feov.Errorf(feov.Conflict,
			"--run %q disagrees with the run this seat was dispatched into (%q). "+
				"The engine injects the run directory; you do not type it. If the flag is a typo, drop it — "+
				"omitting --run is correct and always right. If you believe the injected value is wrong, "+
				"record it with the friction verb rather than working around it",
			flagRun, env)
	}
	if env != "" {
		return env, RunFromEnv, nil
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
