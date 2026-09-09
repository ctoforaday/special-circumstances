package telecli

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// errNoRipgrep is the refusal, held as a value so a test can assert THIS failure rather than
// matching prose that a reword would silently detach the test from.
var errNoRipgrep = errors.New(
	"ripgrep is not installed, so this search cannot run. Refusing rather than reporting zero " +
		"matches, which would be indistinguishable from a real result. Install it (see " +
		"gray-area's requirements.json) or use `telepathy sql`")

func newFindCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "find <term>",
		Short: "search every local transcript (ripgrep, no index)",
		Long: `Searches the raw transcripts the store names, and joins each hit back to whose
session it was.

No full-text index is built. A whole-corpus search measures tens of milliseconds
against a corpus of hundreds of megabytes, where an index would cost gigabytes and
need invalidating on every append — the filesystem is already the index, and it
cannot go stale.

The term is passed to ripgrep as a single argument and never reaches a shell.

If ripgrep is missing this REFUSES rather than printing "no matches": a search that
could not run must never be reported as a search that found nothing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := env.openRead()
			if err != nil {
				return err
			}
			defer db.Close()
			paths, err := catalogue.PathsFor(cmd.Context(), db)
			if err != nil {
				return err
			}
			if len(paths) == 0 {
				return errors.New("the catalogue names no transcripts yet — run `telepathy backfill`")
			}
			rg, err := exec.LookPath("rg")
			if err != nil {
				return errNoRipgrep
			}
			// THE ARGV IS SORTED, and that is where the determinism comes from. `paths` is a map,
			// so building the argv by ranging it hands ripgrep a different command line every
			// run — and then the output order, and anything a caller diffs against yesterday's,
			// is an accident of Go's map seed. Sorting here fixes it at the source rather than
			// hoping to sort it back afterwards.
			names := make([]string, 0, len(paths))
			for p := range paths {
				names = append(names, p)
			}
			sort.Strings(names)
			rgArgs := append([]string{"--count-matches", "--no-messages", "--", args[0]}, names...)
			out := cmd.OutOrStdout()
			// Exit 1 from ripgrep means "no matches" and is not an error here; exit 2 means it
			// could not read something, and --no-messages already means we are not relaying why.
			raw, runErr := exec.CommandContext(cmd.Context(), rg, rgArgs...).Output()
			if runErr != nil && len(raw) == 0 {
				fmt.Fprintf(out, "no transcript contains %q\n", args[0])
				return nil
			}
			type hit struct {
				f catalogue.TranscriptFile
				n string
			}
			var hits []hit
			for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
				i := strings.LastIndex(line, ":")
				if i < 0 {
					continue
				}
				if f, ok := paths[line[:i]]; ok {
					hits = append(hits, hit{f, line[i+1:]})
				}
			}
			if len(hits) == 0 {
				fmt.Fprintf(out, "no transcript contains %q\n", args[0])
				return nil
			}
			// Belt and braces: the argv is already sorted, so ripgrep has no reason to emit these
			// out of order — but "no reason to" is a claim about a tool we do not own, and the
			// cost of not depending on it is one line.
			sort.Slice(hits, func(i, j int) bool { return hits[i].f.Path < hits[j].f.Path })
			fmt.Fprintf(out, "%-10s %-18s %6s  %s\n", "SESSION", "AGENT", "HITS", "PATH")
			for _, h := range hits {
				agent := h.f.AgentID
				if agent == "" {
					agent = "(main)"
				}
				fmt.Fprintf(out, "%-10s %-18s %6s  %s\n", catalogue.Short(h.f.SessionID), agent, h.n, h.f.Path)
			}
			return nil
		},
	}
}
