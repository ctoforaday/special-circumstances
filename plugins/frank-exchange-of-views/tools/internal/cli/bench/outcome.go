package bench

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
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
// --ended says HOW a non-pass ended (a judged deadlock, or the safety ceiling). It decorates the
// stamp and is recorded alongside the verdict, never inferred later. It was two booleans that
// every reader took apart in a switch, which is an enum with a fourth state nobody defined.
func newOutcome() *cobra.Command {
	c := seat.New("outcome",
		"record the run's terminal verdict as a fact: --as "+record.MustEnum("outcome", "verdict").Spelling()+" --reason \"<how the run ended, in your words>\" [--ended "+record.MustEnum("outcome", "ended").Spelling()+"]",
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
			basis, basisWhy := record.VerdictAsserted, ""
			if derived, why, ok := record.DeriveVerdict(s.RunDir); ok {
				basisWhy = why
				if verdict != derived {
					return nil, feov.Errorf(feov.Conflict,
						"outcome: --as %s contradicts the record, which says %s (%s). The verdict is DERIVED, not claimed — if the record is wrong the fix is on the record, not in this flag",
						verdict, derived, why)
				}
				basis = record.VerdictDerived
			}
			ended := seat.Str(cmd, flags.Ended)

			// --reason IS THE RULE, NOT AN EXCEPTION, and the rule is enforced in `validate`
			// rather than here — one write path, one enforcer. The 2026-07-20 vocabulary
			// collapse made every claim or judgment act carry its prose, because "a ruling, a
			// closure, a removal or a dispute with no stated reasoning is indistinguishable from
			// a default". `outcome` — the run's TERMINAL act — was simply missing from that
			// list, which is how a bench seat came to reach for --reason and file its absence
			// as friction (#375).
			//
			// A first pass made the requirement conditional on --deadlocked, on the argument
			// that a derived verdict needs no defence. Wrong twice: it invents a per-flag
			// contract on a verb that had one rule, and "the record derived it" explains the
			// VERDICT while saying nothing about how the sitting ended.
			reason := strings.TrimSpace(seat.Str(cmd, flags.Reason))

			p := record.NewPayload().
				Set("verdict", verdict).
				Set("verdict_basis", basis)
			seat.Set(cmd, p, "ended", flags.Ended)
			if reason != "" {
				p.Set("reason", reason)
			}
			// THE DERIVATION'S OWN REASONING, recorded rather than discarded. It was computed on
			// every call and used only to phrase an error, so the report stamped a verdict and
			// could never say why it was that one.
			if basisWhy != "" {
				p.Set("verdict_why", basisWhy)
			}
			if _, err := record.Append(s.Identity(), "outcome", p); err != nil {
				return nil, err
			}
			return outcomeResult{Verdict: verdict, Ended: ended}, nil
		})

	enumhelp.Flag(c, flags.As, record.MustEnum("outcome", "verdict"), ("the run's terminal verdict"))
	c.Flags().String(flags.Reason, "", "REQUIRED — how this run ended, in your words. The verdict itself is derived from the record; this is the bench's account of the sitting, and on a judged deadlock it is the only evidence the determination ever had")
	enumhelp.Flag(c, flags.Ended, record.MustEnum("outcome", "ended"), "what STOPPED the run, when it was not a pass — a judgement or a ceiling")
	return c
}

type outcomeResult struct {
	Verdict string `json:"verdict"`
	Ended   string `json:"ended,omitempty"`
}

func (r outcomeResult) Human() string {
	by := ""
	switch r.Ended {
	case "deadlock":
		by = " (by judged deadlock)"
	case "ceiling":
		by = " (by safety ceiling)"
	}
	return fmt.Sprintf("outcome recorded: %s%s", r.Verdict, by)
}
