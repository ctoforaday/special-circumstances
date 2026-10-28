package hookcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
)

// PostFailure is the PostToolUseFailure entry point: a tool call a SEAT made that failed is logged as
// the tool's `failure` entry for that seat.
//
// THE SEATS DO NOT LOG THESE, and the tool cannot see them from inside a verb. On universe-m13 and
// m14 the friction nobody filed was mostly here: a jq that misread a projection's shape, a Read of a
// template at a guessed path, a Read of a directory, a path typed with the wrong universe in it. The
// record tool's own refusals are logged where it refuses them; this is everything else.
//
// NOT A SEAT, NOT A RUN, OR AN INTERRUPT: nothing. The main session's failures are the operator's,
// and a user pressing stop is not a failed call.
func PostFailure(stdin io.Reader, _ io.Writer, rec *hookfailures.Recorder) error {
	var in struct {
		ToolName    string          `json:"tool_name"`
		ToolInput   json.RawMessage `json:"tool_input"`
		Error       string          `json:"error"`
		IsInterrupt bool            `json:"is_interrupt"`
		AgentID     string          `json:"agent_id"`
		AgentType   string          `json:"agent_type"`
		Cwd         string          `json:"cwd"`
		SessionID   string          `json:"session_id"`
	}
	b, err := io.ReadAll(stdin)
	if err != nil || json.Unmarshal(b, &in) != nil {
		rec.Fail(StageFailureLog, "cannot read or parse the PostToolUseFailure payload — a seat's failed call was not logged")
		return nil
	}
	if in.IsInterrupt {
		return nil
	}
	caller := callerOf(in.AgentID, in.AgentType, in.SessionID)
	agent := caller.ID
	if agent == "" {
		return nil
	}
	runDir := seatRunDir(in.Cwd, caller, rec)
	if runDir == "" {
		return nil
	}
	text, ok := failureEntry(runDir, in.ToolName, in.ToolInput, in.Error)
	if !ok {
		return nil
	}
	if err := recordFailure(runDir, agent, text); err != nil {
		rec.Fail(StageFailureLog, "a seat's call failed and the failure could not be logged: "+err.Error())
		return nil
	}
	rec.OK(StageFailureLog)
	return nil
}

// failureEntry is the log line for a failed call: the tool, what it was aimed at, and what it
// printed — its one line of substance, or the first and the last of several. It reports false for
// a failure the record tool already logged — its own refusal, which it writes as a `refusal` entry
// at the moment it refuses.
//
// WHAT IT WAS AIMED AT, because the guess is the signal: `jq '.[-1]'` against a projection whose
// help says JSONL, a template path the prompt never gave. A Bash call is named by its whole
// command, a file tool by its path.
//
// THE RUN DIRECTORY IS WRITTEN `.`, in the command, the path and the output alike. The entry lands in
// the run's log, which run.md and report.html ship, and the directory is the operator's: universe-m25
// shipped seven of its seats' commands with `/home/<operator>/.claude/scratch/...` in them. A path in
// a seat's call that is the run's reads as clearly from the run directory, which is where a reader of
// the shipped set stands; a path outside it is printed as the seat typed it.
func failureEntry(runDir, tool string, input json.RawMessage, errText string) (string, bool) {
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(inRun(runDir, errText)), "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "Exit code ") {
			lines = append(lines, l)
		}
	}
	first := ""
	if len(lines) > 0 {
		first = lines[0]
	}
	// THE RECORD TOOL'S OWN REFUSAL, in either of its forms: the human line it prints under its own
	// name, or the JSON envelope --json gets. It is already on the log as a `refusal`.
	//
	// THE ENVELOPE IS DECODED, by the declaration the tool writes it from — never matched on its
	// opening bytes. A prefix recognises whatever happens to start that way and stops recognising
	// the tool's own refusal the day a key is added in front.
	if _, own := feov.DecodeRefusal(first); own || strings.HasPrefix(first, hookgate.RecordBin+": ") {
		return "", false
	}
	var ti struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
	}
	_ = json.Unmarshal(input, &ti)
	target := clip(ti.FilePath, 160)
	if ti.Command != "" {
		// The input arrives as the PreToolUse hook REWROTE it; the seat's own command is what it
		// reached for.
		//
		// THE WHOLE COMMAND, not its head. A shell call's status is its LAST command's — or, behind
		// `&&`, whichever stopped the chain — and the hook is told the status and the output, never
		// which part produced them. The head alone names the first command, and a head cut at 160
		// characters behind a long path names nothing else: `"$R/.bin/feov-record" … register; echo
		// …` was logged for a call whose `ls`, past the cut, was what exited 2.
		target = clip(strings.TrimSpace(hookgate.WithoutInjection(ti.Command)), commandShown)
	}
	target = inRun(runDir, target)
	aimed := ""
	if target != "" {
		aimed = " on `" + target + "`"
	}
	switch len(lines) {
	case 0:
		return fmt.Sprintf("`%s` failed%s: no error text", tool, aimed), true
	case 1:
		return fmt.Sprintf("`%s` failed%s: %s", tool, aimed, clip(first, 300)), true
	}
	// SEVERAL LINES OF OUTPUT, AND NONE OF THEM IS CALLED THE ERROR. The first line is the first
	// thing the call printed, which for a chain is its first command SUCCEEDING — "registered judge"
	// stood as the failure of a call that registered. What failed usually speaks last, so both ends
	// are given, as what they are.
	return fmt.Sprintf("`%s` failed%s — the status is the whole call's and does not say which part of it failed. Its output opens: %s — and ends: %s",
		tool, aimed, clip(first, 300), clip(lines[len(lines)-1], 300)), true
}

// inRun is s with the run directory written `.` wherever it stands as a whole path: followed by a
// separator, or by a character no path name continues with. A sibling directory whose name begins
// with the run's (`…_is-91-prime-b2` beside `…_is-91-prime`) is another directory, and keeps its path.
//
// The run directory is rooted — `/`-rooted, or a Windows volume path — or nothing is rewritten: a
// relative one would match inside unrelated words. On Windows the hook is handed the run in the OS's
// form (`C:\…\run`) and a seat may type it either way, so its slashed form is written `.` as well.
func inRun(runDir, s string) string {
	runDir = strings.TrimRight(runDir, "/"+string(filepath.Separator))
	if !strings.HasPrefix(runDir, "/") && !filepath.IsAbs(runDir) {
		return s
	}
	s = wholePathAsDot(runDir, s)
	if slashed := filepath.ToSlash(runDir); slashed != runDir {
		s = wholePathAsDot(slashed, s)
	}
	return s
}

// wholePathAsDot is s with dir written `.` wherever it stands as a whole path (see inRun).
func wholePathAsDot(runDir, s string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, runDir)
		if i < 0 {
			return b.String() + s
		}
		end := i + len(runDir)
		if end < len(s) && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-", rune(s[end])) {
			b.WriteString(s[:end])
		} else {
			b.WriteString(s[:i] + ".")
		}
		s = s[end:]
	}
}

// commandShown bounds the command a failure entry quotes. It is far above any chain a seat types
// by hand — the cut is for a script or a long --reason passed in a heredoc — and a command past it
// is marked as cut.
const commandShown = 2000

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
