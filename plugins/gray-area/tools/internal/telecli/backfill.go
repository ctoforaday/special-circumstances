package telecli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

func newBackfillCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "backfill",
		Short: "read the existing corpus into the store",
		Long: `Projects every transcript under --projects into the store, once.

This is the only WRITING verb, and it is explicit rather than something a hook
does: the hooks keep the store current with a bounded per-invocation cap, and this
is the cold read that cap exists to avoid facing at the start of a session.

Re-running is safe and cheap. Each file's consumed prefix is fingerprinted, so an
unchanged file is skipped and a file that GREW is resumed from where it stopped —
a file that was REWRITTEN is detected and re-read from the beginning.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env.Store == "" {
				return fmt.Errorf("no catalogue path: neither --store nor a home directory could be resolved")
			}
			db, err := catalogue.Open(env.Store)
			if err != nil {
				return err
			}
			defer db.Close()
			files, err := catalogue.TranscriptFiles(env.ProjectsDir)
			if err != nil {
				return err
			}
			start := env.Now()
			var acts, words, thoughts, skipped, unparsed int
			var read int64
			for _, f := range files {
				r, err := catalogue.IngestFile(db, f)
				if err != nil {
					return err
				}
				acts += r.Acts
				words += r.Words
				thoughts += r.Thoughts
				skipped += r.Skipped
				unparsed += r.Unparsed
				read += r.BytesRead
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "backfilled %d transcripts, %.0f MB in %.1fs\n  %d acts, %d words, %d thoughts\n",
				len(files), float64(read)/1e6, env.Now().Sub(start).Seconds(), acts, words, thoughts)
			if skipped > 0 {
				// SAID AT THE POINT OF INGEST, not left to whoever queries v_skip later. This is
				// the single largest population the projection drops, and on the measured corpus
				// it is two orders of magnitude larger than what it keeps: 15,503 withheld against
				// 248 stored. A backfill that reported only the 248 would be describing a corpus
				// nobody has.
				fmt.Fprintf(out, "  %d thinking block(s) carried no text and are recorded as skips, not thoughts — "+
					"reasoning withheld by the client is not reasoning the agent did not do\n", skipped)
			}
			if unparsed > 0 {
				// Reported, never swallowed: a torn line and an absent line leave the same missing
				// row, and only one of them is anybody's fault.
				fmt.Fprintf(out, "  %d line(s) did not parse — those records are ABSENT from the store, "+
					"which is a different fault from a session that did less\n", unparsed)
			}
			fmt.Fprintln(out, "  re-running is safe: unchanged files are skipped by their stored offset")
			return nil
		},
	}
}
