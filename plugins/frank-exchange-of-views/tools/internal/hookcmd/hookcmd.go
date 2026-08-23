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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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

// readReportFile is the production report reader injected into hookgate.PostDropped.
func readReportFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Pre denies a non-author raw write to a run's blue/report.md, and otherwise injects the run
// directory and agent id into a Bash call that invokes the record tool.
func Pre(stdin io.Reader, stdout io.Writer) error {
	in, raw, ok := readInput(stdin)
	if !ok {
		// Fail CLOSED, but narrowly: an unparseable payload that references a report path is
		// denied conservatively; anything else gets no opinion (the matcher already scoped this
		// to Write|Edit|Bash, so an odd payload is anomalous).
		if strings.Contains(string(raw), "blue/report.md") || strings.Contains(string(raw), "blue\\\\report.md") {
			emitPreDeny(stdout, hookgate.PreDenyReason()+" (hook input unparseable; denied conservatively)")
		}
		return nil
	}
	// The run directory is resolved from the payload's `cwd` — the SEAT's working directory,
	// which is wire-supplied and documented, never this hook process's os.Getwd(). Absent or
	// unusable marker → empty → no rewrite, matching InferRunDir's "say nothing rather than
	// guess".
	runDir := seat.InferRunDir(cwdOf(raw))
	switch outcome, payload := hookgate.PreOutcome(in, runDir); outcome {
	case hookgate.OutcomeDeny:
		// A DENIAL THIS GATE CANNOT EXPLAIN IS THE ONE IT MUST RECORD. When the refusal is "I
		// could not read your tool_input", the seat did nothing wrong and the cause is in the
		// tool — so it goes on the record as friction, not just into a reason string the run
		// forgets. hookgate stays free of I/O; the write lives here.
		if hookgate.InputUnreadable(in) {
			fileToolFriction(runDir, in.ToolName)
		}
		emitPreDeny(stdout, payload)
	case hookgate.OutcomeRewrite:
		emitPreRewrite(stdout, in.ToolInput, payload)
	}
	return nil
}

// Post blocks and forces a repair when a non-author call dropped an immortal anchor from the
// report by any mechanism — the backstop for the exotic Bash the Pre verb-set cannot enumerate.
func Post(stdin io.Reader, stdout io.Writer) error {
	in, _, ok := readInput(stdin)
	if !ok {
		return nil // best-effort backstop; PreToolUse is the real gate
	}
	missing, err := hookgate.PostDropped(in, hookgate.DefaultAnchorIDs, readReportFile)
	if err != nil || len(missing) == 0 {
		return nil
	}
	emitPostBlock(stdout, missing)
	return nil
}

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

// fileToolFriction records a tool fault on the run's record.
//
// BEST EFFORT, AND SILENT ONLY WHERE THERE IS NOTHING TO WRITE TO. A hook can fire outside a run
// (no marker, no run directory), and there is no record to carry the event then — the deny reason
// still reaches the model, which is the channel that always exists. Where a run IS resolvable the
// fault becomes durable, so a refusal nobody understood at the time is answerable afterwards.
//
// The write must never turn a hook into a failure: this gate's job is to answer, and a seat is
// not blocked because the bookkeeping failed.
func fileToolFriction(runDir, toolName string) {
	if runDir == "" {
		return
	}
	f := &recordpb.Friction{
		Text: proto.String(fmt.Sprintf("hookgate could not parse tool_input for %s, so the blue-report "+
			"lockdown refused the call rather than risk bypassing it. The seat did nothing wrong; "+
			"the tool_input shape and this gate's parser have diverged.", toolName)),
		Kind: recordpb.FrictionKind_FRICTION_KIND_TOOL_ERROR.Enum(),
	}
	// Round -1 is UNKNOWN, and record.Identity is explicit that this is not round 0 — a hook
	// fires outside any seat's round, and conflating the two produced the phantom-archive bug
	// this field was added to prevent.
	_, _ = record.Append(record.Identity{RunDir: runDir, SeatID: "hookgate", Round: -1}, f)
}

// emitPreDeny writes the PreToolUse deny document (exit stays 0).
func emitPreDeny(stdout io.Writer, reason string) {
	emit(stdout, map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	})
}

// emitPostBlock writes the PostToolUse block document (decision:block + reason), which forces the
// seat to repair before proceeding.
func emitPostBlock(stdout io.Writer, missing []string) {
	reason := fmt.Sprintf("you removed immortal anchor(s) %s — blue/report.md carries red's finding-markers and blue's tool-managed citation anchors, and both are immortal. "+
		"Restore the removed anchor(s) exactly, and make every change through `feov-record blue edit` (never a raw write).",
		strings.Join(missing, ", "))
	emit(stdout, map[string]any{"decision": "block", "reason": reason})
}

func emit(stdout io.Writer, doc map[string]any) {
	_ = json.NewEncoder(stdout).Encode(doc)
}
