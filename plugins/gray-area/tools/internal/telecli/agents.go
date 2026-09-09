package telecli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

func newAgentsCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "agents",
		Short: "who is running, in which worktree, doing what",
		Long: `Lists every session the client currently advertises, with its liveness, how much it
has done and what it did last.

Liveness comes from the client's session files and /proc, never from the store: a
transcript cannot say whether the process that wrote it is still there. 'unknown'
is a third answer and not a synonym for 'ended'.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			db, err := env.openRead()
			if err != nil {
				return err
			}
			defer db.Close()
			rows, err := catalogue.Agents(cmd.Context(), db, env.SessionsDir)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(rows) == 0 {
				// NOT an empty table. The client writes a session file per live session, so none
				// means nothing is running — a different fact from a store with no rows, and said
				// in words.
				fmt.Fprintf(out, "no sessions are advertised in %s — nothing is running on this box\n", env.SessionsDir)
				return nil
			}
			sort.Slice(rows, func(i, j int) bool {
				if rows[i].LastAt != rows[j].LastAt {
					return rows[i].LastAt > rows[j].LastAt
				}
				return rows[i].SessionID < rows[j].SessionID // a stable order for equal clocks
			})
			fmt.Fprintf(out, "%-10s %-8s %7s  %-14s %s\n", "SESSION", "STATE", "ACTS", "LAST", "CWD")
			for _, a := range rows {
				last := "—"
				if a.LastTool != "" {
					last = a.LastTool
					if a.LastAt > 0 {
						last += " " + env.ago(a.LastAt)
					}
				}
				fmt.Fprintf(out, "%-10s %-8s %7d  %-14s %s\n",
					catalogue.Short(a.SessionID), a.Liveness, a.Acts, last, a.CWD)
			}
			for _, a := range rows {
				if a.Liveness == catalogue.Unknown {
					fmt.Fprintln(out, "\n  `unknown` means liveness could not be established — a foreign pid namespace, "+
						"or a platform where this is not implemented. It is not `ended`.")
					break
				}
			}
			return nil
		},
	}
}
