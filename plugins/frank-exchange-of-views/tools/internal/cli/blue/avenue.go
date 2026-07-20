package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// avenue: a line of inquiry, and what became of it.
//
// think-around-problem mandates exploring genuinely distinct alternatives before
// a significant decision. terse-communication forbids narrating options. So the
// exploration was required, invisible and unverifiable — a dead letter by the
// standard scorecards.md sets out, and the same shape as "confidence
// self-graded", which was mandated and practised five times in 1,892 lines.
//
// This is what instruments it. Exploration becomes a RECORD and the record
// becomes a report section, so the rule stops being self-attested and starts
// leaving artifacts.
//
// THREE STATUSES, each answering a different question:
//
//	declined   considered, not taken     — was the rejection REASONED, or decoration?
//	abandoned  pursued, then died        — what killed it? (the negative result)
//	pursued    became the report's spine — does the report reflect it honestly?
//
// `abandoned` is the most valuable and the most routinely lost: dead ends are
// exactly what a future run needs so it does not re-walk them, and they are
// direct feedstock for the sleeper service's subject mining. Nothing in the
// engine preserved them before this verb — they died in a seat's context.
func newAvenue() *cobra.Command {
	c := seat.Prose(seat.New(role, "avenue",
		`record a line of inquiry and its fate: --line "<the question or approach>" --status declined|abandoned|pursued --reason "<why not taken, or what killed it>" [--method "<the source class or technique>"]`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Text(cmd)
			if err != nil {
				return nil, err
			}
			p := seat.SetSame(cmd, record.NewPayload(), flags.Line, flags.Status, flags.Reason, flags.Method)
			if text != "" {
				p.Set("detail", text)
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "avenue", p); err != nil {
				return nil, err
			}
			return avenueResult{Status: seat.Str(cmd, flags.Status), Line: seat.Str(cmd, flags.Line)}, nil
		}))

	c.Flags().String(flags.Line, "", "the question or approach — what you were going to try")
	c.Flags().String(flags.Status, "", "declined (considered, not taken) | abandoned (pursued, then died) | pursued (became the report's spine)")
	c.Flags().String(flags.Reason, "", "why it was declined, or what killed it — the part a future run actually needs")
	c.Flags().String(flags.Method, "", "the source class or technique it belonged to, when that is what distinguishes it")
	return c
}

type avenueResult struct {
	Status string `json:"status"`
	Line   string `json:"line"`
}

func (r avenueResult) Human() string { return "avenue recorded (" + r.Status + "): " + r.Line }
