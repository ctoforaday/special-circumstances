package bench

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// outcome: the run's terminal verdict, recorded as a fact.
//
// VERIFIED | CEILING | HALTED | UNVERIFIED is the orchestrator's read of the EXIT STATE,
// and it is not recoverable from any single event: the per-epoch `gate` events are
// red-merge's PASS/FAIL verdict, and "hit the ceiling" needs the run terms the log
// never carries. So the value originates outside the record — but it BELONGS in it. The
// report is assembled from the log; a verdict passed as an ephemeral --inputs field would
// be the one fact in the report that nothing recorded, and an unrecorded fact is one a
// future reader cannot audit. This verb writes it; `bench assemble` reads it back.
//
// HOW the run ended is the verdict's to say: CEILING is derived from the board, a halt is on the
// record, and deadlock is per gap. The `--ended` modifier that restated it is retired.
func newOutcome() *cobra.Command {
	c := seat.New("outcome", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// REQUIRED BY THE RECORD AND MARKED BY THE MECHANISM. This was a hand-rolled refusal in
		// RunE, so `--help` could not say the flag was required and the message arrived only after
		// the seat had composed the whole call.
		verdict := seat.Str(cmd, flags.As)
		// THE TOOL DECIDES; THE FLAG IS A CROSS-CHECK (#308).
		//
		// --as used to be the SOURCE of the run's terminal verdict: debate.js computed it in
		// JS, told the assembler in its prompt, and the seat typed it back. Nothing compared
		// the two, so a seat could record VERIFIED over a board with open gaps and the
		// report's stamp, capture's audits and every scorecard would believe it. Same
		// posture as seatenv's --run: where the tool can decide, a flag that disagrees is
		// REFUSED naming both, rather than obeyed.
		basis, basisWhy := record.VerdictAsserted, ""
		derived, why, ok := record.DeriveVerdict(run)
		switch {
		case ok && verdict != derived:
			return nil, feov.Errorf(feov.Conflict,
				"outcome: --as %s contradicts the record, which says %s (%s). The verdict is DERIVED, not claimed — if the record is wrong the fix is on the record, not in this flag",
				verdict, derived, why)
		case ok:
			basis, basisWhy = record.VerdictDerived, why
		case !strings.EqualFold(verdict, "UNVERIFIED"):
			// THE ONLY WORD THE BENCH MAY ASSERT. When the record holds no terminal state — no
			// halt, no PASS, no board at its ceiling — the run ended before it reached one, and
			// UNVERIFIED is the name of that. Any other word here is a claim the record would
			// refuse if it could decide, arriving through the one gap where it cannot.
			return nil, feov.Errorf(feov.Conflict,
				"outcome: --as %s, but the record holds no terminal state to derive it from (%s). Only UNVERIFIED may be asserted over such a record — a run that ended before a halt, a PASS or the ceiling",
				verdict, why)
		}

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
		// THE CHANNEL, NOT THE FLAG. This read --reason directly and registered its own copy of
		// it, so the run's terminal account — which this flag's own help calls "the only evidence
		// of why the run stopped" on an UNVERIFIED run — had no file form and no stdin form.
		// A bench with a paragraph had to fight the shell for the most consequential prose field
		// in the run.
		prose, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		reason := strings.TrimSpace(prose)

		// THE WORD IS REFUSED HERE, not recorded as the zero. RunOutcome's zero is
		// UNSPECIFIED, so a verdict the schema does not know would land as a run with no
		// verdict at all — which reads downstream exactly like a run that never reached one.
		v, ok := record.RunOutcomeOf(verdict)
		if !ok {
			return nil, feov.Errorf(feov.Validation,
				"bench outcome: %q is not a verdict this record can carry", verdict)
		}
		body := &recordpb.Outcome{Verdict: &v, VerdictBasis: proto.String(basis)}
		if reason != "" {
			body.Prose = proto.String(reason)
		}
		// THE DERIVATION'S OWN REASONING, recorded rather than discarded. It was computed on
		// every call and used only to phrase an error, so the report stamped a verdict and
		// could never say why it was that one.
		if basisWhy != "" {
			body.VerdictWhy = proto.String(basisWhy)
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return outcomeResult{Verdict: verdict, Basis: basis}, nil
	})

	seat.Prose(c)
	c.Flags().Lookup(flags.Reason).Usage = "how this run ended, in your words — the bench's account of the sitting, and on an UNVERIFIED run the only evidence of why it stopped"
	enumhelp.Flag(c, flags.As, record.MustEnum("outcome", "verdict"), ("the run's terminal verdict"))
	return c
}

type outcomeResult struct {
	Verdict string `json:"verdict"`
	Basis   string `json:"verdict_basis"`
}

func (r outcomeResult) Human() string {
	return fmt.Sprintf("outcome recorded: %s (%s)", r.Verdict, r.Basis)
}
