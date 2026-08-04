// Package hookenv resolves the one piece of context every hook in this suite needs and
// none of them owned: WHERE THE PROJECT IS.
//
// THE CLASS. Seven binaries read `os.Getenv("CLAUDE_PROJECT_DIR")` and, finding it empty,
// each did something quiet and different: sc-checkpoint-restore returned 0 with no output
// — indistinguishable from "this project has no checkpoint" — while sc-recall-index (since
// retired) once
// resolved its state file RELATIVE, writing debounce state into whatever directory the
// process happened to start in. That second one was found and fixed IN PLACE, with a
// comment noting that "the log writer in this same binary already guarded against exactly
// that; the stamp did not". The instance was fixed; the class was not, and the other six
// call sites kept the hole.
//
// Every one of those binaries is handed the project root a SECOND time, in the hook
// payload's `cwd` field, and every one of them ignored it. So the resolution is:
//
//	CLAUDE_PROJECT_DIR  ->  the payload's cwd  ->  nothing
//
// NOTHING, deliberately — never os.Getwd(). Falling back to the working directory is the
// defect sc-recall-index had: it turns "I do not know where the project is" into
// "I will write somewhere", and the somewhere is unrelated to the session. A hook that
// cannot locate the project must do nothing, and SAY it did nothing.
package hookenv

import (
	"fmt"
	"io"
)

// ProjectDir returns the project root, preferring the environment and falling back to the
// payload the harness already sent. Empty means genuinely unknown.
func ProjectDir(env, payloadCWD string) string {
	if env != "" {
		return env
	}
	return payloadCWD
}

// Explain writes one line to stderr when the root could not be resolved, and reports
// whether the caller should carry on.
//
// A hook that no-ops silently is indistinguishable from a hook with nothing to do, which
// is how this survived: the restore hook returned 0 and printed nothing, and the session
// read that as "no checkpoint exists". Stderr is the right channel — it does not reach the
// transcript, so it costs the session no context, and it is where a `claude --debug` run
// looks.
func Explain(dir string, stderr io.Writer, hook string) bool {
	if dir != "" {
		return true
	}
	fmt.Fprintf(stderr, "%s: no project root (CLAUDE_PROJECT_DIR unset and the hook payload carried no cwd) — doing nothing rather than guessing from the working directory\n", hook)
	return false
}
