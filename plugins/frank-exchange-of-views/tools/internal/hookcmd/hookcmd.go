// Package hookcmd is the PreToolUse/PostToolUse backend, as its own program rather than a verb
// on the seat tool.
//
// # Why this is not a subcommand any more
//
// It was `feov-record hook pretooluse`, and hooks.json invoked it with no --seat-id because a
// hook HAS no seat id — Claude Code calls it, not a seat and not an operator. That worked only
// while "nobody said who is asking" fell through to the operator surface. When identity started
// SELECTING the surface, no-identity stopped resolving any command: the root refused the shipped
// hook's own invocation, RefuseAndTeach exits 2, and exit 2 from a PreToolUse hook DENIES the
// tool call. Matcher Write|Edit|MultiEdit|NotebookEdit|Bash — so every mutating tool call in the
// session, for anyone whose plugin bin/ held the binary. The suite stayed green because the
// change taught eight test files to pass --seat-id, and hooks.json is the one caller that cannot
// be taught.
//
// Registering the verb outside the scoped surface fixed that instance and left the shape wrong.
// A hook backend living inside an identity-scoped command tree has to be exempted from identity,
// hidden from help, and then exempted AGAIN from the surface census — CommandPaths walks hidden
// commands, so `blue hook pretooluse` became a seat capability and three gates demanded a trigger
// row and a board that drives it. Every one of those exemptions is a symptom of the same category
// error: this is not a verb, it is a program.
//
// The suite's other eleven hooks are dedicated single-purpose binaries — prosthetic-conscience's
// nine and gray-area's one — and not one of them broke. This joins them.
//
// # The exit-code contract, enforced here rather than documented downstream
//
// EVERY ENTRY POINT RETURNS 0, whatever it decides. The decision travels in the stdout JSON; the
// exit status is not a channel. That invariant used to live in a doc comment on a leaf command,
// two levels below the argument parsing that broke it. Run() enforces it at the process boundary,
// including through a panic, because a hook that dies is a session that cannot edit files.
package hookcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittinghook"
)

// Run invokes a hook entry point and guarantees the process exit code is 0.
//
// A panic here is a bug worth fixing, and it is NOT worth denying every tool call in the session
// to report. It is swallowed deliberately: stderr still carries it for anyone reading hook logs,
// and the caller proceeds.
func Run(f func(io.Reader, io.Writer) error, stdin io.Reader, stdout io.Writer) (code int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "feov hook: recovered from %v (exiting 0: a hook must never block on its own fault)\n", r)
			code = 0
		}
	}()
	if err := f(stdin, stdout); err != nil {
		fmt.Fprintf(os.Stderr, "feov hook: %v (exiting 0)\n", err)
	}
	return 0
}

// readInput returns the parsed Input, the raw bytes, and whether the parse succeeded.
func readInput(stdin io.Reader) (hookgate.Input, []byte, bool) {
	var in hookgate.Input
	raw, err := io.ReadAll(stdin)
	if err != nil || len(raw) == 0 {
		return in, raw, false
	}
	ok := json.Unmarshal(raw, &in) == nil
	return in, raw, ok
}

// Pre injects the run directory and the calling agent's id into a Bash call made inside a live
// run, so a seat never mistypes the path and never loses the identity that binds it to its seat on
// the record — and it refuses two things: any tool call past the seat's per-sitting call limit
// (enforceLimit), and a tool command carrying a backtick the shell would run, which would rewrite
// the seat's prose before the tool saw it (hookgate/substitution.go).
func Pre(stdin io.Reader, stdout io.Writer) error {
	in, raw, ok := readInput(stdin)
	if !ok {
		return nil // no payload to inject into
	}
	// THE TURN LIMIT COMES FIRST, and covers every tool: a refused call runs nothing, so it needs
	// nothing injected. A counting fault goes to stderr and the call proceeds.
	denied, err := enforceLimit(in, cwdOf(raw), stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "feov hook: turn limit: %v\n", err)
	}
	if denied {
		return nil
	}
	// Injection is for Bash alone, so every other tool stops here, before the marker search.
	if in.ToolName != "Bash" {
		return nil
	}
	// The run directory is resolved from the payload's `cwd` — the SEAT's working directory,
	// which is wire-supplied and documented, never this hook process's os.Getwd(). Absent or
	// unusable marker → empty → no rewrite, matching InferRunDir's "say nothing rather than guess".
	runDir := runlive.InferRunDir(cwdOf(raw))
	switch outcome, payload := hookgate.PreOutcome(in, runDir); outcome {
	case hookgate.OutcomeRewrite:
		emitPreRewrite(stdout, in.ToolInput, payload)
	case hookgate.OutcomeDeny:
		emitPreDeny(stdout, payload)
	}
	return nil
}

// enforceLimit counts this call against the calling seat's sitting and refuses it when the sitting
// is past the run's per-sitting tool-call limit (internal/sittingcap).
//
// THE SEAT IS FOUND THE SAME WAY UNDER EITHER ENGINE. A Workflow seat is a subagent, and its
// payload carries agent_id. A seat run as the main session of a headless `claude -p` process has
// no agent_id on the payload; its identity and run are in the process environment this hook
// inherits, so those are read when the payload has none. An agent that never registered as a
// seat has no sitting open and is never counted — the main session, an operator, anything else.
//
// The one call a limited sitting lets through is a register (hookgate.OpensASitting), because
// that is what opens the next sitting's count.
func enforceLimit(in hookgate.Input, cwd string, stdout io.Writer) (denied bool, err error) {
	agentID, agentType := in.AgentID, in.AgentType
	if agentID == "" {
		agentID, agentType = seatenv.AgentID(), seatenv.AgentType()
	}
	if agentID == "" {
		return false, nil
	}
	runDir := runlive.InferRunDir(cwd)
	if runDir == "" {
		runDir = os.Getenv(seatenv.Var)
	}
	if runDir == "" {
		runDir = os.Getenv(seatenv.VarWrapper)
	}
	d, counted, err := sittingcap.Count(runDir, agentID)
	if err != nil || !counted || !d.Over || hookgate.OpensASitting(in) {
		return false, err
	}
	if d.First {
		err = recordLimit(runDir, agentID, agentType, d.Sitting, d.Limit)
	}
	emitPreDeny(stdout, hookgate.LimitReason(d.Count, d.Limit))
	return true, err
}

// recordLimit is a variable so the hook's handoff can be tested without a built writer on disk.
var recordLimit = sittinghook.Limit

// cwdOf pulls the seat's working directory out of the raw payload.
//
// It is read from the RAW bytes rather than added to hookgate.Input because Input is the gate's
// contract and cwd is this program's business — the gate takes a resolved directory as a
// parameter precisely so it never has to know how one is found.
func cwdOf(raw []byte) string {
	var p struct {
		Cwd string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &p)
	return p.Cwd
}

// emitPreRewrite replaces the tool's arguments with the same call plus the injected run
// directory. It emits `permissionDecision: allow` with `updatedInput` — ONE document, the same
// slot a deny would have used, which is why the ordering in PreOutcome is structural.
//
// The whole tool_input is preserved and only `command` is replaced: a Bash call may carry other
// fields (description, timeout, run_in_background) and dropping them would silently change how
// the seat's command runs.
func emitPreRewrite(stdout io.Writer, toolInput json.RawMessage, command string) {
	updated := map[string]any{}
	if err := json.Unmarshal(toolInput, &updated); err != nil {
		// SAYING NOTHING NAMES THE WRONG CAUSE. Without the injection the seat hits the "no run
		// directory" refusal from a different layer entirely, and nobody learns the real reason.
		// Declining to send a shape the client cannot use is right; `ask` IS a shape it can use.
		emitPreAsk(stdout, fmt.Sprintf("feov-record: tool_input did not parse (%v), so the run "+
			"directory could not be injected. Nothing is blocked — this is surfaced rather than "+
			"swallowed, because the refusal you would otherwise hit names the wrong cause.", err))
		return
	}
	updated["command"] = command
	emit(stdout, map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "allow",
			"permissionDecisionReason": "feov-record: the run directory is injected by the engine, not typed by the seat (#281)",
			"updatedInput":             updated,
		},
	})
}

// emitPreAsk writes the PreToolUse ask document: hand the call to the human WITH THE REAL CAUSE.
//
// FAIL-OPEN VS FAIL-CLOSED IS A FALSE BINARY: `ask` is the third option. A hook that cannot
// understand its input need not guess a direction — `permissionDecisionReason` is carried back to
// the model, so the agent sees the actual error and can correct or escalate instead of being
// silently allowed or silently blocked.
//
// This is the SYNCHRONOUS half of the answer; a friction event of kind FRICTION_KIND_TOOL_ERROR
// is the durable half. The reason string reaches the agent NOW; the friction event is what a
// later reader finds. Neither replaces the other.
func emitPreAsk(stdout io.Writer, reason string) {
	emit(stdout, map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "ask",
			"permissionDecisionReason": reason,
		},
	})
}

// emitPreDeny writes the PreToolUse deny document. `permissionDecisionReason` is carried back to
// the model, so the seat reads what the shell would have run and the form that works, in the same
// turn — a refusal that teaches rather than one the seat has to reverse-engineer.
func emitPreDeny(stdout io.Writer, reason string) {
	emit(stdout, map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	})
}

func emit(stdout io.Writer, doc map[string]any) {
	_ = json.NewEncoder(stdout).Encode(doc)
}
