// sc-strike-counter is a prosthetic-conscience PostToolUseFailure hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER the same tool fails at the same target Threshold times inside the window, it
//	MUST put the count in front of the model, naming what has been tried.
//	During a user INTERRUPT it MUST NOT count anything — anti-spinning also says honor
//	the cancel, and a rule whose two halves fight each other is worse than neither.
//	It MUST NOT block: every path exits 0. The rule says STOP AND ESCALATE, which is a
//	thing the model does, not a thing a hook can do on its behalf.
//
// # What this is, and what it deliberately is not
//
// anti-spinning's rule has two halves. "Re-running the same action is not a retry" requires
// judging whether the APPROACH changed — that stays in the skill, where judgement belongs.
// "Same tool, same target, third failure" is a COUNTER, and a counter is exactly the part a
// model asked to notice its own impaired loop-detection cannot be trusted with. This is the
// counter only.
//
// # Measured basis
//
// plans/hook-surface-spike.md §2 recorded the payload rather than assuming it:
//
//	tool_name     Bash
//	tool_input    {"command": "ls /definitely/not/a/real/path/xyz", …}
//	error         Exit code 2\nls: cannot access '…': No such file or directory
//	is_interrupt  false
//
// `is_interrupt` was NOT anticipated by #127 and is load-bearing: without it a user
// pressing stop three times reads as a spin, and the counter would nag about the very act
// of cancelling.
//
// # Two channels, both now verified (#234, measured 2026-08-03)
//
// #127 proposed emitting `additionalContext` so the count reaches the MODEL. The spike had
// verified that this event CARRIES what a counter needs; it had NOT verified that this event
// can INJECT — and §3a of the same spike records `PostCompact` carrying the compaction
// summary and being unable to inject it, which was the identical assumption one event over.
// So this shipped emitting both channels and saying which one was a guess.
//
// It is no longer a guess. Measured end-to-end under headless `claude -p` — the case the
// issue worried about most, since that is where no operator is watching stderr:
//
//	{"attachment":{"type":"hook_additional_context","hookEvent":"PostToolUseFailure",
//	 "hookName":"PostToolUseFailure:Read","content":["MARKER-PTUF-SUBJECT-9K2M"]}}
//
// Same attachment type that verified `SessionStart` injection, in the same turn as the
// failure, with a `SessionStart` marker as the positive control so a NOT-SEEN could not be
// confused with a run where the hook never fired. The model then reported both markers
// verbatim. See plans/hook-surface-spike.md §2a.
//
// BOTH channels stay, now as a choice rather than a hedge. stderr is what an operator reads
// while watching a session, and it is the channel that survives a model deciding an injected
// token looks like a prompt-injection attempt — behaviour the spike observed twice (§3b).
package strikecounter

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/strikes"
)

type hookInput struct {
	ToolName    string          `json:"tool_name"`
	ToolInput   json.RawMessage `json:"tool_input"`
	ToolUseID   string          `json:"tool_use_id"`
	Error       string          `json:"error"`
	DurationMS  int             `json:"duration_ms"`
	IsInterrupt bool            `json:"is_interrupt"`
	SessionID   string          `json:"session_id"`
	CWD         string          `json:"cwd"`
}

// hookOutput is the documented PostToolUse-family response shape.
type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// maxErrExcerpt bounds how much of the failure text is quoted back. The point is to
// identify WHICH failure keeps recurring, not to re-print it.
const maxErrExcerpt = 200

// excerpt reduces an error to its first meaningful line, bounded.
func excerpt(err string) string {
	for _, line := range strings.Split(err, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			if len(t) > maxErrExcerpt {
				return t[:maxErrExcerpt] + "…"
			}
			return t
		}
	}
	return ""
}

// target renders the key's subject for a human. The key is tool\x00target.
func target(key string) string {
	if i := strings.IndexByte(key, 0); i >= 0 {
		return key[i+1:]
	}
	return ""
}

// message is what the model (and the operator) are handed on the third strike.
//
// It states the COUNT and the WINDOW rather than asserting a verdict: the hook knows the
// same call failed three times in fifteen minutes, and does not know whether the approach
// changed between them. Naming what it actually observed leaves the judgement where
// anti-spinning puts it, and makes the claim falsifiable.
func message(in hookInput, key string, count int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "anti-spinning: %s has now failed %d times in %s at the same target",
		in.ToolName, count, strikes.Window)
	if t := target(key); t != "" {
		fmt.Fprintf(&b, " (%s)", t)
	}
	b.WriteString(". ")
	if e := excerpt(in.Error); e != "" {
		fmt.Fprintf(&b, "Last error: %s. ", e)
	}
	// It never blocks, and a message arriving at a failure reads like the cause of it.
	b.WriteString("This hook blocked nothing — the failure above is the tool's own. The rule: STOP, summarize what was tried, and escalate — and re-running the same action is not a retry, so a fourth attempt needs a DIFFERENT approach, not another go. " +
		"This hook counts tool failures only: it cannot see a command that exits 0 while the work still failed, so its silence is not evidence the loop was broken.")
	return b.String()
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time) int {
	fs := flag.NewFlagSet("sc-strike-counter", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag is never worth disturbing the session over
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line("sc-strike-counter"))
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)

	// HONOR THE CANCEL. anti-spinning's other clause: a user interrupt is not a failed
	// attempt, and counting it would have the rule nag about the act of stopping.
	if in.IsInterrupt {
		return 0
	}
	if in.ToolName == "" {
		return 0
	}

	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	if !hookenv.Explain(projectDir, stderr, "sc-strike-counter") {
		return 0
	}

	path := strikes.Path(projectDir)
	state := strikes.Load(os.ReadFile, path)
	key := strikes.Key(in.ToolName, in.ToolInput)
	count, tripped := strikes.Record(&state, key, now)
	strikes.Save(path, state)

	if !tripped {
		return 0
	}

	msg := message(in, key, count)
	// Both channels, deliberately — see the package comment. Injection is measured (#234);
	// stderr is for the operator watching the session, not a fallback for a doubted one.
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "PostToolUseFailure"
	out.HookSpecificOutput.AdditionalContext = msg
	_ = json.NewEncoder(stdout).Encode(out)
	fmt.Fprintln(stderr, msg)
	return 0
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now())
}
