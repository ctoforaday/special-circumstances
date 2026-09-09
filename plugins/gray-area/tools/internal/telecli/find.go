package telecli

import (
	"bytes"
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
	var asRegex bool
	c := &cobra.Command{
		Use:   "find <term>",
		Short: "search every local transcript (ripgrep, no index)",
		Long: `Searches the raw transcripts the store names, and joins each hit back to whose
session it was.

No full-text index is built. A whole-corpus search measures tens of milliseconds
against a corpus of hundreds of megabytes, where an index would cost gigabytes and
need invalidating on every append — the filesystem is already the index, and it
cannot go stale.

The term is a LITERAL by default, and is passed to ripgrep as a single argument
that never reaches a shell. Pass --regex to treat it as a pattern — which is what
you want for a word boundary, and searching for 'roving' rather than '\broving\b'
is how you get every occurrence of "proving" instead.

Anything that stops this from being a completed search REFUSES rather than
printing "no matches": ripgrep absent, a pattern that does not compile, or a
corpus it could not finish reading. A search that could not run must never be
reported as a search that found nothing.`,
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
			// LITERAL UNLESS ASKED. `find 'Mint('` is a reasonable thing to type and is not a
			// valid regex; as a pattern it made ripgrep exit 2, and the branch below used to read
			// that as "no matches". -F removes the whole class for the common case, and a caller
			// who wants a pattern says so.
			rgArgs := []string{"--count-matches"}
			if !asRegex {
				rgArgs = append(rgArgs, "--fixed-strings")
			}
			rgArgs = append(rgArgs, "--", args[0])
			rgArgs = append(rgArgs, names...)
			out := cmd.OutOrStdout()
			// RIPGREP'S THREE EXIT CODES ARE THREE DIFFERENT ANSWERS, and collapsing them is how
			// this verb produced the exact zero it was written to refuse:
			//
			//	0  matched
			//	1  ran to completion, matched nothing        <- the honest zero
			//	2  did not run, or did not finish            <- NOT a zero
			//
			// `find 'Mint('` took the third path, printed "no transcript contains", and exited 0.
			// --no-messages used to hide the reason as well, so nothing anywhere said the search
			// had failed. It is gone: a refusal that cannot say why is barely better than silence.
			var stderr bytes.Buffer
			run := exec.CommandContext(cmd.Context(), rg, rgArgs...)
			run.Stderr = &stderr
			raw, runErr := run.Output()
			code := 0
			var ee *exec.ExitError
			if errors.As(runErr, &ee) {
				code = ee.ExitCode()
			} else if runErr != nil {
				return fmt.Errorf("could not run ripgrep: %w", runErr)
			}
			if code >= 2 {
				why := strings.TrimSpace(stderr.String())
				if why == "" {
					why = "ripgrep exited " + fmt.Sprint(code) + " without saying why"
				}
				// PARTIAL RESULTS ARE STILL REPORTED, because exit 2 also covers one transcript
				// vanishing mid-walk, and throwing away a good answer over a file that was
				// deleted while we read it would be its own kind of wrong. What must not happen
				// is printing them as if the search were complete.
				if len(raw) == 0 {
					return fmt.Errorf("the search did not run: %s", why)
				}
				fmt.Fprintf(cmd.ErrOrStderr(),
					"telepathy find: INCOMPLETE — ripgrep could not finish, so these hits are a "+
						"floor and not the answer: %s\n", why)
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
	c.Flags().BoolVar(&asRegex, "regex", false,
		"treat the term as a regular expression rather than a literal (needed for \\b word boundaries)")
	return c
}
