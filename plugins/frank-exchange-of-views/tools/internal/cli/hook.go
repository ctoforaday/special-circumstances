package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
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
func newHook() *cobra.Command {
	c := &cobra.Command{
		Use:           "hook",
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
			if deny, reason := hookgate.PreDecision(in); deny {
				emitPreDeny(cmd, reason)
			}
			return nil
		},
	}
}

func newHookPost() *cobra.Command {
	return &cobra.Command{
		Use:           "posttooluse",
		Short:         "block + force repair if a non-author call dropped a finding-marker (reads a PostToolUse event on stdin)",
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
	reason := fmt.Sprintf("you removed finding-marker(s) %s — blue/report.md is red's record and markers are immortal. "+
		"Restore the removed marker(s) exactly, and make every change through `feov-record blue edit` (never a raw write).",
		strings.Join(missing, ", "))
	out := map[string]any{"decision": "block", "reason": reason}
	_ = json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}
