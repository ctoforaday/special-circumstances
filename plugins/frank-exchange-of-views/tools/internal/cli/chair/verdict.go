package chair

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The help document renders what holds a PASS from the gate's own table rather than restating
// it: help/verdict.md names {{.PassBlockerKinds}}, so the list a seat reads and the list the gate
// refuses on cannot disagree.
func init() {
	seat.HelpValues["PassBlockerKinds"] = strings.Join(record.PassBlockerGlosses(), ", ")
}

// verdict: the chair's terminal act, the PASS.
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
		// THE VERDICT REACHES THE WORKFLOW ON THE PLAN, AND THIS ACT PRINTS THAT PLAN. The chair's
		// envelope is the plan, and the plan carries the PASS its sitting recorded; so the act
		// records the PASS and asks the dispatch on the record it leaves, in one transaction
		// (record.RecordPass, which stores the plan the hook writes into the envelope), and prints
		// the plan as `dispatch next` does.
		plan, standing, err := record.RecordPass(s.Identity())
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
		return verdictResult{dispatched: dispatchResult{Plan: plan, standing: standing}, open: open, closed: closed, checkpoint: mirror}, nil
	})
	return c
}

// verdictResult is the PASS recorded and the plan the record now prints, which IS the --json
// output (seat.PrintsBare), as `dispatch next`'s is.
type verdictResult struct {
	dispatched dispatchResult
	open       int
	closed     int
	checkpoint string
}

// Bare is the plan, carrying the PASS just recorded.
func (r verdictResult) Bare() any { return r.dispatched.Bare() }

func (r verdictResult) Human() string {
	return fmt.Sprintf("PASS (%d open, %d closed) and checkpointed to %s\n%s", r.open, r.closed, r.checkpoint, r.dispatched.Human())
}
