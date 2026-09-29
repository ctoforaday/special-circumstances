package hookcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
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
// and a user pressing stop is not friction.
func PostFailure(stdin io.Reader, _ io.Writer, rec *hookfailures.Recorder) error {
	var in struct {
		ToolName    string          `json:"tool_name"`
		ToolInput   json.RawMessage `json:"tool_input"`
		Error       string          `json:"error"`
		IsInterrupt bool            `json:"is_interrupt"`
		AgentID     string          `json:"agent_id"`
		Cwd         string          `json:"cwd"`
	}
	b, err := io.ReadAll(stdin)
	if err != nil || json.Unmarshal(b, &in) != nil {
		rec.Fail(StageFailureLog, "cannot read or parse the PostToolUseFailure payload — a seat's failed call was not logged")
		return nil
	}
	if in.IsInterrupt {
		return nil
	}
	agent := in.AgentID
	if agent == "" {
		agent = seatenv.AgentID()
	}
	if agent == "" {
		return nil
	}
	runDir := seatRunDir(in.Cwd, rec)
	if runDir == "" {
		return nil
	}
	text, ok := failureEntry(in.ToolName, in.ToolInput, in.Error)
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

// failureEntry is the log line for a failed call: the tool, what it was aimed at, and the error's
// first line of substance. It reports false for a failure the record tool already logged — its own
// refusal, which it writes as a `refusal` entry at the moment it refuses.
//
// WHAT IT WAS AIMED AT, because the guess is the signal: `jq '.[-1]'` against a projection whose
// help says JSONL, a template path the prompt never gave. A Bash call is named by its command's
// first line, a file tool by its path.
func failureEntry(tool string, input json.RawMessage, errText string) (string, bool) {
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(errText), "\n") {
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
	if strings.HasPrefix(first, hookgate.RecordBin+": ") || strings.HasPrefix(first, `{"ok":false`) {
		return "", false
	}
	var ti struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
	}
	_ = json.Unmarshal(input, &ti)
	target := ti.FilePath
	if ti.Command != "" {
		target, _, _ = strings.Cut(strings.TrimSpace(ti.Command), "\n")
	}
	aimed := ""
	if target = clip(target, 160); target != "" {
		aimed = " on `" + target + "`"
	}
	if first == "" {
		first = "no error text"
	}
	return fmt.Sprintf("`%s` failed%s: %s", tool, aimed, clip(first, 300)), true
}

func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
