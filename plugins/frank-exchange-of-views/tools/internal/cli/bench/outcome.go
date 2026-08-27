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
		if derived, why, ok := record.DeriveVerdict(run.Dir()); ok {
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
		// THE CHANNEL, NOT THE FLAG. This read --reason directly and registered its own copy of
		// it, so the run's terminal account — which this flag's own help calls "the only evidence
		// the determination ever had" on a judged deadlock — had no file form and no stdin form.
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
		if e := seat.Str(cmd, flags.Ended); e != "" {
			body.Ended = proto.String(e)
		}
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
		return outcomeResult{Verdict: verdict, Ended: ended}, nil
	})

	seat.Prose(c)
	c.Flags().Lookup(flags.Reason).Usage = "how this run ended, in your words. The verdict itself is derived from the record; this is the bench's account of the sitting, and on a judged deadlock it is the only evidence the determination ever had"
	enumhelp.Flag(c, flags.As, record.MustEnum("outcome", "verdict"), ("the run's terminal verdict"))
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
