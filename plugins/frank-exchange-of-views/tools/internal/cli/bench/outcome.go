package bench

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// outcome: the run's terminal verdict, recorded as a fact.
//
// VERIFIED | CEILING | HALTED | UNVERIFIED is the orchestrator's read of the EXIT STATE,
// and it is not recoverable from any single event: the per-round `verdict` events are
// red-merge's PASS/FAIL gate, and "hit the round ceiling" needs the round cap the log
// never carries. So the value originates outside the record — but it BELONGS in it. The
// report is assembled from the log; a verdict passed as an ephemeral --inputs field would
// be the one fact in the report that nothing recorded, and an unrecorded fact is one a
// future reader cannot audit. This verb writes it; `bench assemble` reads it back.
//
// deadlocked / exhausted say HOW a non-pass ended (judged deadlock vs safety ceiling);
// they decorate the stamp and are recorded alongside the verdict, never inferred later.
func newOutcome() *cobra.Command {
	c := seat.New("outcome",
		"record the run's terminal verdict as a fact: --as "+record.MustEnum("outcome", "verdict").Spelling()+" [--deadlocked] [--exhausted]",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			verdict := seat.Str(cmd, flags.As)
			if verdict == "" {
				return nil, fmt.Errorf("outcome: --as <verdict> is required (%s)", record.MustEnum("outcome", "verdict").Usage("the run's terminal verdict"))
			}
			// THE TOOL DECIDES; THE FLAG IS A CROSS-CHECK (#308).
			//
			// --as used to be the SOURCE of the run's terminal verdict: debate.js computed it in
			// JS, told the assembler in its prompt, and the seat typed it back. Nothing compared
			// the two, so a seat could record VERIFIED over a board with open gaps and the
			// report's stamp, capture's audits and every scorecard would believe it. Same
			// posture as seatenv's --run: where the tool can decide, a flag that disagrees is
			// REFUSED naming both, rather than obeyed.
			basis := record.VerdictAsserted
			if derived, why, ok := record.DeriveVerdict(s.RunDir); ok {
				if verdict != derived {
					return nil, feov.Errorf(feov.Conflict,
						"outcome: --as %s contradicts the record, which says %s (%s). The verdict is DERIVED, not claimed — if the record is wrong the fix is on the record, not in this flag",
						verdict, derived, why)
				}
				basis = record.VerdictDerived
			}
			deadlocked, _ := cmd.Flags().GetBool(flags.Deadlocked)
			exhausted, _ := cmd.Flags().GetBool(flags.Exhausted)
			p := record.NewPayload().
				Set("verdict", verdict).
				Set("verdict_basis", basis).
				Set("deadlocked", deadlocked).
				Set("exhausted", exhausted)
			if _, err := record.Append(s.RunDir, s.SeatID, "outcome", p); err != nil {
				return nil, err
			}
			return outcomeResult{Verdict: verdict, Deadlocked: deadlocked, Exhausted: exhausted}, nil
		})

	c.Flags().String(flags.As, "", record.MustEnum("outcome", "verdict").Usage("the run's terminal verdict"))
	c.Flags().Bool(flags.Deadlocked, false, "the run ended by judged deadlock")
	c.Flags().Bool(flags.Exhausted, false, "the run ended by safety/round ceiling")
	return c
}

type outcomeResult struct {
	Verdict    string `json:"verdict"`
	Deadlocked bool   `json:"deadlocked"`
	Exhausted  bool   `json:"exhausted"`
}

func (r outcomeResult) Human() string {
	by := ""
	switch {
	case r.Deadlocked:
		by = " (by judged deadlock)"
	case r.Exhausted:
		by = " (by safety ceiling)"
	}
	return fmt.Sprintf("outcome recorded: %s%s", r.Verdict, by)
}
