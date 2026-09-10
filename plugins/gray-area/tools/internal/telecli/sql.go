package telecli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// viewReference renders the published schema FROM the schema, rather than restating it.
//
// A hand-typed copy here would be a second roster of the same fact, kept in step by nobody: the
// help would go on describing a column after it was renamed, and it is precisely the caller
// reading this text who would then write the query that fails. catalogue.ViewColumns is already
// checked against the real database by TestViewColumnsAreTheContract, so generating from it means
// this text cannot be wrong without that test going red first.
func viewReference() string {
	names := make([]string, 0, len(catalogue.ViewColumns))
	width := 0
	for v := range catalogue.ViewColumns {
		names = append(names, v)
		if len(v) > width {
			width = len(v)
		}
	}
	sort.Strings(names)
	var b strings.Builder
	for _, v := range names {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, v, strings.Join(catalogue.ViewColumns[v], ", "))
	}
	return b.String()
}

func newSQLCmd(env *Env) *cobra.Command {
	var limit int
	c := &cobra.Command{
		Use:   "sql <query>",
		Short: "read-only SQL over the published views",
		Long: `Runs a query against the store and prints the result as tab-separated columns.

The questions worth asking of a trajectory corpus are not knowable in advance, so
the schema is published and callers write their own. These views are the contract:

` + viewReference() + `
The connection is opened read-only, with query_only and ATTACH both refused, so a
query cannot write and cannot reach a second database. Underlying tables are not
the contract: query the views.

NULL prints as the word NULL, because "" and NULL are different answers.`,
		Example: `  telepathy sql "SELECT tool, count(*) n FROM v_action GROUP BY 1 ORDER BY 2 DESC"
  telepathy sql "SELECT text FROM v_thought WHERE session_id = 'abc' ORDER BY block_seq"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit < 0 {
				return usagef("--limit cannot be negative (0 means no limit)")
			}
			db, err := env.openRead(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer db.Close()
			cols, rows, err := catalogue.RawQuery(cmd.Context(), db, args[0], limit)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, strings.Join(cols, "\t"))
			for _, r := range rows {
				fmt.Fprintln(out, strings.Join(r, "\t"))
			}
			if limit > 0 && len(rows) == limit {
				// TRUNCATION IS SAID OUT LOUD. A capped result that looks complete is the same
				// defect as a search that reports zero because it could not run.
				fmt.Fprintf(out, "\n(%d rows shown — the limit, so there may be more)\n", limit)
			}
			return nil
		},
	}
	c.Flags().IntVar(&limit, "limit", 200, "stop after this many rows (0 for no limit); truncation is reported")
	return c
}
