package telecli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

func newSessionCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "session <id>",
		Short: "one session's shape: calls and errors by tool",
		Long: `Summarises what one session did, by tool, with its error count beside each.

The id may be the full session id. Run 'telepathy agents' for the ones running now,
or query v_session for the ones that have ended.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := env.openRead()
			if err != nil {
				return err
			}
			defer db.Close()
			s, err := catalogue.Shape(cmd.Context(), db, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			state := "open"
			if s.ClosedAt.Valid {
				state = "closed " + env.ago(s.ClosedAt.Int64)
			}
			fmt.Fprintf(out, "session %s  (%s)\n  %s\n\n", catalogue.Short(s.SessionID), state, s.ProjectDir)
			tools := make([]string, 0, len(s.ByTool))
			for k := range s.ByTool {
				tools = append(tools, k)
			}
			sort.Slice(tools, func(i, j int) bool {
				if s.ByTool[tools[i]] != s.ByTool[tools[j]] {
					return s.ByTool[tools[i]] > s.ByTool[tools[j]]
				}
				return tools[i] < tools[j] // ties sort by name, so the output is reproducible
			})
			fmt.Fprintf(out, "  %-16s %7s %7s\n", "TOOL", "CALLS", "ERRORS")
			for _, tl := range tools {
				fmt.Fprintf(out, "  %-16s %7d %7d\n", tl, s.ByTool[tl], s.Failures[tl])
			}
			fmt.Fprintf(out, "\n  %d words, %d thoughts\n", s.Words, s.Thoughts)
			// THE SKIPS ARE THE HALF THAT USED TO BE INVISIBLE. "0 thoughts" and "0 thoughts, 300
			// of them arrived empty" are different facts about different subjects — the second is
			// about the CLIENT — and only one of them says anything about the agent.
			if n := s.Skipped[catalogue.SkipThinkingEmpty]; n > 0 {
				fmt.Fprintf(out, "  %d thinking block(s) carried NO TEXT and were not stored — "+
					"reasoning was withheld from the transcript, not absent from the agent\n", n)
			}
			if s.Thoughts == 0 && s.Skipped[catalogue.SkipThinkingEmpty] == 0 && s.Words > 0 {
				// NAME THE CAUSE THAT IS ACTUALLY LIKELY. This used to send the reader to check
				// `showThinkingSummaries`, and then measurement said otherwise: on this box the
				// setting is ON and reasoning text still arrives EMPTY — 911 thinking blocks in
				// one transcript, 0 carrying text, and none anywhere since 2026-09-08. Pointing a
				// reader at a setting that is already correct costs them the trip and teaches them
				// the wrong model of why the column is empty.
				fmt.Fprintln(out, "  ^ no thinking blocks at all in this session's transcript — not even "+
					"empty ones, which is a different state from reasoning that was withheld. Check "+
					"`showThinkingSummaries`; and do not read this zero as a fact about the agent.")
			}
			return nil
		},
	}
}
