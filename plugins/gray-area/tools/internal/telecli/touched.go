package telecli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

func newTouchedCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "touched <path>",
		Short: "which sessions acted on a path, and when",
		Long: `Answers "is anyone else editing this file".

Matching is on a suffix, so a repo-relative path finds the act whatever absolute
path the other agent used. A path nothing names produces a SENTENCE, not an empty
table: silence here would be a claim about the file, when it is only a claim about
what has been ingested.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := env.openRead(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer db.Close()
			hits, err := catalogue.Touched(cmd.Context(), db, args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(hits) == 0 {
				fmt.Fprintf(out, "no recorded act names %s\n", args[0])
				fmt.Fprintln(out, "  ^ that is a statement about the CATALOGUE, not about the file: a session whose "+
					"transcript was never ingested contributes nothing here. `telepathy backfill` reads what exists.")
				return nil
			}
			fmt.Fprintf(out, "%-10s %-18s %-8s %-10s %5s  %s\n", "SESSION", "AGENT", "TOOL", "OUTCOME", "N", "WHEN")
			for _, h := range hits {
				agent := h.AgentID
				if agent == "" {
					agent = "(main)"
				}
				fmt.Fprintf(out, "%-10s %-18s %-8s %-10s %5d  %s\n",
					catalogue.Short(h.SessionID), agent, h.Tool, h.Outcome, h.N, env.ago(h.TS))
			}
			return nil
		},
	}
}
