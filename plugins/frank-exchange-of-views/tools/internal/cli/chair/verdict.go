package chair

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The help document renders what holds a PASS from the gate's own table rather than restating
// it: help/verdict.md names {{.PassBlockerKinds}}, so the list a seat reads and the list the gate
// refuses on cannot disagree.
func init() {
	seat.HelpValues["PassBlockerKinds"] = strings.Join(record.PassBlockerGlosses(), ", ")
}

// verdict: the chair's terminal act.
//
// It CHECKPOINTS the record (the append-only event log) to a mirror outside the run, as one
// consistent database snapshot (record.WriteMirror). The run directory is untracked-by-design until
// capture, and the 2026-07-17 incident showed how a stray git operation can
// delete a live blackboard mid-sitting; a mirror keyed by the run path means the
// events survive the working tree. Projections are regenerated on read from the
// mirror, so the frozen snapshot is the source, not a materialized cache.
func newVerdict() *cobra.Command {
	c := seat.New("verdict", verdictResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		v, ok := record.VerdictOf(seat.Str(cmd, flags.As))
		if !ok {
			// THE REFUSAL NAMES WHAT WOULD HAVE WORKED. `%q is not a verdict` tells a seat it
			// was wrong and not what to type — and this is the terminal act of the sitting, so
			// a seat that cannot get past it has nowhere to go. The set is rendered from the
			// same declaration the help renders, so the two cannot say different things.
			return nil, feov.Errorf(feov.Validation,
				"--%s must be one of %s, spelled exactly (got %q) — the verdict is the chair sitting's terminal act and every later reader switches on it",
				flags.As, record.MustEnum("verdict", "verdict").Spelling(), seat.Str(cmd, flags.As))
		}
		// THE VERDICT REACHES THE WORKFLOW ON THE PLAN, AND THIS ACT PRINTS THAT PLAN. The chair's
		// envelope is the plan, and the plan carries the verdict its sitting recorded; so the act
		// records the verdict and asks the dispatch on the record it leaves, in one transaction
		// (record.RecordVerdict, which stores the plan capture compares the relay with), and prints
		// the plan as `dispatch next` does.
		plan, standing, err := record.RecordVerdict(s.Identity(), v)
		if err != nil {
			return nil, err
		}
		open, closed, err := record.BoardCounts(run)
		if err != nil {
			return nil, err
		}
		mirror, err := record.WriteMirror(run)
		if err != nil {
			return nil, err
		}
		return verdictResult{dispatched: dispatchResult{Plan: plan, standing: standing}, verdict: seat.Str(cmd, flags.As), open: open, closed: closed, checkpoint: mirror}, nil
	})

	enumhelp.Flag(c, flags.As, record.MustEnum("verdict", "verdict"), "the seat's terminal act")
	return seat.Require(c, flags.As)
}

// verdictResult is the verdict recorded and the plan the record now prints, which IS the --json
// output (seat.PrintsBare), as `dispatch next`'s is.
type verdictResult struct {
	dispatched dispatchResult
	verdict    string
	open       int
	closed     int
	checkpoint string
}

// Bare is the plan, carrying the verdict just recorded.
func (r verdictResult) Bare() any { return r.dispatched.Bare() }

func (r verdictResult) Human() string {
	return fmt.Sprintf("verdict %s (%d open, %d closed) and checkpointed to %s\n%s", r.verdict, r.open, r.closed, r.checkpoint, r.dispatched.Human())
}
