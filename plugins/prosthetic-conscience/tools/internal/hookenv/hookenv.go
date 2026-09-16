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
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
)

// ProjectDir returns the project root, preferring the environment and falling back to the
// payload the harness already sent. Empty means genuinely unknown.
func ProjectDir(env, payloadCWD string) string {
	if env != "" {
		return env
	}
	return payloadCWD
}

// StageProjectRoot is recorded when no project root resolves: every hook that calls Explain then
// does NOTHING for the whole session.
const StageProjectRoot hookfailures.Stage = "project-root"

// Explain records that the root could not be resolved, and reports whether the caller should carry
// on.
//
// A hook that no-ops silently is indistinguishable from a hook with nothing to do, which
// is how this survived: the restore hook returned 0 and printed nothing, and the session
// read that as "no checkpoint exists".
//
// THIS COMMENT USED TO SAY STDERR WAS THE RIGHT CHANNEL — "it does not reach the transcript, so it
// costs the session no context, and it is where a `claude --debug` run looks". The first half is
// true and the conclusion does not follow: measured 2026-09-16, a hook's stderr at exit 0 reaches
// the debug log and NOBODY ELSE, so seven binaries announced their own total silence into a file
// nobody opens. The line still goes to stderr, and now the failure is also on the record, which the
// next displaying event reads out.
func Explain(dir string, rec *hookfailures.Recorder, hook string) bool {
	Note(dir, rec)
	return dir != ""
}

// Note records whether the root resolved, and decides NOTHING. It is for the events whose units must
// each run regardless — PreToolUse's secrets gate fails closed with or without a project, and must
// not be told to give up by a lookup it does not depend on. Those two binaries never called Explain,
// so an unresolved root was recorded by every hook except the two that fire on every tool call.
func Note(dir string, rec *hookfailures.Recorder) {
	if dir != "" {
		rec.OK(StageProjectRoot)
		return
	}
	rec.Fail(StageProjectRoot, "no project root (CLAUDE_PROJECT_DIR unset and the hook payload carried no cwd) — doing nothing rather than guessing from the working directory")
}
