package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/verify"
)

// newVerify is the operator's read-only cross-check over a run's record. It is NOT a seat verb
// — no debating role runs it; it is the human's and CI's instrument, so it sits at the root
// rather than inside a role namespace. It replays the board once, asserts every invariant, and
// prints the authoritative tally. A failed invariant exits non-zero so CI can gate on it.
func newVerify() *cobra.Command {
	c := &cobra.Command{
		Use:           "verify",
		Short:         "cross-check a run's record against its invariants, and tally it (read-only)",
		Long:          "verify replays <run>'s event record and asserts the invariants that must hold if it is sound — gaps disposed, refs resolve, PASS closed everything, seats registered first — then prints the authoritative counts. It writes nothing. A violated invariant exits non-zero.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Resolved so the injected run reaches reads too, not only writes.
			sc := seat.Of(cmd)
			// seat.Of ALREADY inferred if nothing was supplied — it ends its own resolution with
			// InferRunDir. What used to sit here was a SECOND inference, reached only when the
			// first had produced nothing, which by then meant the run had been REFUSED rather
			// than omitted. Inferring there resolved quietly to a different run than the one the
			// seat was told it could not have. Run() returns that refusal instead.
			run, err := sc.Run()
			if err != nil {
				return err
			}
			board, err := record.BoardState(run)
			if err != nil {
				return fmt.Errorf("verify: %w", err)
			}
			checks := verify.Run(board)
			stats := verify.Compute(board)

			jsonMode, _ := cmd.Flags().GetBool(flags.JSON)
			if jsonMode {
				out := struct {
					OK     bool           `json:"ok"`
					Checks []verify.Check `json:"checks"`
					Stats  verify.Stats   `json:"stats"`
				}{OK: len(verify.Failed(checks)) == 0, Checks: checks, Stats: stats}
				_ = json.NewEncoder(cmd.OutOrStdout()).Encode(out)
			} else {
				printReport(cmd, checks, stats)
			}

			if failed := verify.Failed(checks); len(failed) > 0 {
				return feov.Errorf(feov.Validation, "verify: %d invariant(s) failed", len(failed))
			}
			return nil
		},
	}
	return c
}

func printReport(cmd *cobra.Command, checks []verify.Check, s verify.Stats) {
	w := cmd.OutOrStdout()
	fmt.Fprintln(w, "invariants:")
	// THREE STATES, NOT TWO (#411). A check that did not apply is not a check that held, and
	// printing both as `[ok  ]` is how `pass-closes-all-gaps` sat inapplicable on every run
	// ever recorded while its line read like a considered judgement. The status word comes
	// from the Check itself so this renderer and the report section cannot disagree.
	for _, c := range checks {
		fmt.Fprintf(w, "  [%-4s] %s — %s\n", c.Status(), c.Name, c.Detail)
		for _, v := range c.Violations {
			fmt.Fprintf(w, "         · %s\n", v)
		}
	}
	if na := verify.NotApplicable(checks); len(na) > 0 {
		// Named again, together, because one `n/a` in a list of seven is easy to read past —
		// and a gate that is n/a on EVERY run is the thing worth noticing.
		fmt.Fprintf(w, "\n  %d invariant(s) did not apply to this run: %s\n",
			len(na), strings.Join(namesOf(na), ", "))
	}
	fmt.Fprintf(w, "\nrecord (verdict %s):\n", nonEmpty(s.Verdict, "unrecorded"))
	fmt.Fprintf(w, "  gaps: %d total · %d open · %d closed\n", s.GapsTotal, s.GapsOpen, s.GapsClosed)
	fmt.Fprintf(w, "  findings: %d (%d minted, %d un-minted)\n", s.Findings, s.FindingsMinted, s.FindingsUnminted)
	fmt.Fprintf(w, "  citations checked: %d\n", s.Citations)
	fmt.Fprintf(w, "  dialectic on the record: %d gap(s) with a closing · %d with a dispute · %d with an opinion\n",
		s.GapsWithClosing, s.GapsWithDispute, s.GapsWithOpinion)
	fmt.Fprintf(w, "  events: %s\n", eventTally(s.Events))
}

// eventTally renders the type→count map in a stable, most-frequent-first order.
func eventTally(m map[string]int) string {
	type kv struct {
		k string
		v int
	}
	var kvs []kv
	for k, v := range m {
		kvs = append(kvs, kv{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool {
		if kvs[i].v != kvs[j].v {
			return kvs[i].v > kvs[j].v
		}
		return kvs[i].k < kvs[j].k
	})
	out := ""
	for i, e := range kvs {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%s=%d", e.k, e.v)
	}
	return out
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// namesOf is the check names, for a one-line summary.
func namesOf(cs []verify.Check) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Name
	}
	return out
}
