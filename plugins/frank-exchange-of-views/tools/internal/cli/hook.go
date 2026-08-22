package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// readReportFile is the production report reader injected into hookgate.PostDropped.
func readReportFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// newHook is the operator command group backing the blue-report lockdown. It is invoked by
// the plugin's hooks.json, NOT as a seat verb: `hook pretooluse` reads a PreToolUse event
// on stdin and DENIES a non-author raw write to blue/report.md; `hook posttooluse` reads a
// PostToolUse event and BLOCKS (forcing repair) if a non-author call dropped a
// finding-marker via any mechanism. Both ALWAYS exit 0 — the decision travels in the stdout
// JSON, never the exit status (mirrors prosthetic-conscience's sc-pretooluse). The verb set
// under `hook` is machine-facing; a seat never types it.
// hookVerb is the ONE command name that carries no seat identity by design. It is a const so the
// registration and the identity pre-check in root.go cannot drift into disagreeing about which
// name is exempt — the failure mode of two hand-kept copies is that the exemption stops covering
// the command it exists for, and the symptom is a hook that blocks every tool call.
const hookVerb = "hook"

func newHook() *cobra.Command {
	c := &cobra.Command{
		Use:           hookVerb,
		Short:         "PreToolUse/PostToolUse backends for the blue-report lockdown (invoked by hooks.json, not a seat verb)",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	c.AddCommand(newHookPre(), newHookPost())
	return c
}

// readHookInput returns the parsed Input, the raw bytes, and whether the parse succeeded.
func readHookInput(cmd *cobra.Command) (hookgate.Input, []byte, bool) {
	var in hookgate.Input
	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil || len(raw) == 0 {
		return in, raw, false
	}
	ok := json.Unmarshal(raw, &in) == nil
	return in, raw, ok
}

func newHookPre() *cobra.Command {
	return &cobra.Command{
		Use:           "pretooluse",
		Short:         "deny a non-author raw write to blue/report.md (reads a PreToolUse event on stdin)",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			in, raw, ok := readHookInput(cmd)
			if !ok {
				// Fail CLOSED, but narrowly: an unparseable payload that references a report
				// path is denied conservatively; anything else gets no opinion (the matcher
				// already scoped this to Write|Edit|Bash, so an odd payload is anomalous).
				if strings.Contains(string(raw), "blue/report.md") || strings.Contains(string(raw), "blue\\\\report.md") {
					emitPreDeny(cmd, hookgate.PreDenyReason()+" (hook input unparseable; denied conservatively)")
				}
				return nil
			}
			// The run directory is resolved from the payload's `cwd` — the SEAT's working
			// directory, which is wire-supplied and documented, never this hook process's
			// os.Getwd(). Absent or unusable marker → empty → no rewrite, matching
			// InferRunDir's "say nothing rather than guess".
			runDir := seat.InferRunDir(cwdOf(raw))
			switch outcome, payload := hookgate.PreOutcome(in, runDir); outcome {
			case hookgate.OutcomeDeny:
				// A DENIAL THIS GATE CANNOT EXPLAIN IS THE ONE IT MUST RECORD. When the refusal is
				// "I could not read your tool_input", the seat did nothing wrong and the cause is
				// in the tool — so it goes on the record as friction, not just into a reason
				// string the run forgets. hookgate stays free of I/O; the write lives here.
				if hookgate.InputUnreadable(in) {
					fileToolFriction(runDir, in.ToolName)
				}
				emitPreDeny(cmd, payload)
			case hookgate.OutcomeRewrite:
				emitPreRewrite(cmd, in.ToolInput, payload)
			}
			return nil
		},
	}
}

func newHookPost() *cobra.Command {
	return &cobra.Command{
		Use:           "posttooluse",
		Short:         "block + force repair if a non-author call dropped a finding-marker or citation anchor (reads a PostToolUse event on stdin)",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			in, _, ok := readHookInput(cmd)
			if !ok {
				return nil // best-effort backstop; PreToolUse is the real gate
			}
			missing, err := hookgate.PostDropped(in, hookgate.DefaultAnchorIDs, readReportFile)
			if err != nil || len(missing) == 0 {
				return nil
			}
			emitPostBlock(cmd, missing)
			return nil
		},
	}
}

// cwdOf pulls the seat's working directory out of the raw payload.
//
// It is read from the RAW bytes rather than added to hookgate.Input because Input is the
// gate's contract and cwd is the CLI's business — the gate takes a resolved directory as a
// parameter precisely so it never has to know how one is found.
func cwdOf(raw []byte) string {
	var p struct {
		Cwd string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &p)
	return p.Cwd
}

// emitPreRewrite replaces the tool's arguments with the same call plus the injected run
// directory. It emits `permissionDecision: allow` with `updatedInput` — ONE document, the
// same slot a deny would have used, which is why the ordering in PreOutcome is structural.
//
// The whole tool_input is preserved and only `command` is replaced: a Bash call may carry
// other fields (description, timeout, run_in_background) and dropping them would silently
// change how the seat's command runs.
func emitPreRewrite(cmd *cobra.Command, toolInput json.RawMessage, command string) {
	updated := map[string]any{}
	if err := json.Unmarshal(toolInput, &updated); err != nil {
		// SAYING NOTHING NAMES THE WRONG CAUSE. Without the injection the seat hits the "no run
		// directory" refusal from a different layer entirely, and nobody learns the real reason.
		// Declining to send a shape the client cannot use is right; `ask` IS a shape it can use.
		emitPreAsk(cmd, fmt.Sprintf("feov-record: tool_input did not parse (%v), so the run "+
			"directory could not be injected. Nothing is blocked — this is surfaced rather than "+
			"swallowed, because the refusal you would otherwise hit names the wrong cause.", err))
		return
	}
	updated["command"] = command
	out := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "allow",
			"permissionDecisionReason": "feov-record: the run directory is injected by the engine, not typed by the seat (#281)",
			"updatedInput":             updated,
		},
	}
	_ = json.NewEncoder(cmd.OutOrStdout()).Encode(out)
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
func emitPreAsk(cmd *cobra.Command, reason string) {
	out := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "ask",
			"permissionDecisionReason": reason,
		},
	}
	_ = json.NewEncoder(cmd.OutOrStdout()).Encode(out)
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
	p := record.NewPayload().
		Set("text", fmt.Sprintf("hookgate could not parse tool_input for %s, so the blue-report "+
			"lockdown refused the call rather than risk bypassing it. The seat did nothing wrong; "+
			"the tool_input shape and this gate's parser have diverged.", toolName)).
		Set(record.FrictionKindKey, record.FrictionKindToolError)
	// Round -1 is UNKNOWN, and record.Identity is explicit that this is not round 0 — a hook
	// fires outside any seat's round, and conflating the two produced the phantom-archive bug
	// this field was added to prevent.
	_, _ = record.Append(record.Identity{RunDir: runDir, SeatID: "hookgate", Round: -1}, "friction", p)
}

// emitPreDeny writes the PreToolUse deny document (exit stays 0).
func emitPreDeny(cmd *cobra.Command, reason string) {
	out := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": reason,
		},
	}
	_ = json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}

// emitPostBlock writes the PostToolUse block document (decision:block + reason), which
// forces the seat to repair before proceeding.
func emitPostBlock(cmd *cobra.Command, missing []string) {
	reason := fmt.Sprintf("you removed immortal anchor(s) %s — blue/report.md carries red's finding-markers and blue's tool-managed citation anchors, and both are immortal. "+
		"Restore the removed anchor(s) exactly, and make every change through `feov-record blue edit` (never a raw write).",
		strings.Join(missing, ", "))
	out := map[string]any{"decision": "block", "reason": reason}
	_ = json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}
