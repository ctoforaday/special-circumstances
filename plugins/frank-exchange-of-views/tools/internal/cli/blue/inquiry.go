package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// line of inquiry: a line of inquiry, and what became of it.
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
//
// TWO FORMS, AND THEY ARE TWO VERBS (#246).
//
//	propose  --reason "<the approach>" --hypothesis "<what would be true if it pays off>"
//	move     --id A1 --as pursued|declined|abandoned|deferred --reason "<what changed>"
//
// The move form is the whole point. Measured over 86 line-of-inquiry events in six runs: ZERO
// were ever recorded twice and ZERO statuses ever changed, because there was no id and no update
// path. 83 of 86 landed in round 0 — so "pursued" meant "I intend to", nothing could ever
// falsify it, and a direction that died mid-run had no way to say so.
//
// They were ONE verb, and --id silently chose which contract applied: a proposal needs a
// hypothesis and defaults its status, a move needs the id, the new status and the reason. Nothing
// cobra could mark, so both forms ran with every flag optional and the seat learned which was
// which by being refused. The whole `conditionallyRequired` mechanism in the seat package existed
// to describe this one verb, and it is gone with the split.
//
// --hypothesis is what makes a later move checkable: a line that says what would be true if it
// paid off can be abandoned against its own claim rather than on a shrug.
func newInquiry() *cobra.Command {
	c := &cobra.Command{
		Use:          "line-of-inquiry",
		Short:        "a direction the report could take, and what became of it — `propose` one, `move` one you already proposed",
		SilenceUsage: true,
	}
	c.AddCommand(newInquiryPropose(), newInquiryMove())
	return c
}

func newInquiryPropose() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("propose", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		id, err := record.MintInquiryID(s.RunDir)
		if err != nil {
			return nil, err
		}
		p := record.NewPayload().Set("inquiry_id", id)
		seat.SetSame(cmd, p, flags.Method)
		seat.Set(cmd, p, "hypothesis", flags.Hypothesis)
		// The record keeps its own word: the payload key is `line`, the flag is --reason.
		// Flag words are not payload keys (see internal/flags).
		//
		// READ THE PROSE CHANNEL ONCE, AND FILL BOTH KEYS FROM IT. This read `--reason` as a raw
		// FLAG for `line` and the resolved CHANNEL for `reason`, so the two were not the same
		// argument: `--reason` filled both, and `--reason-file` filled only `reason`, leaving
		// `line` empty and the write refused. The flag's own help says --reason-file is "the same
		// field as --reason, for anything long or that would fight shell quoting", and for this
		// one verb it was not.
		//
		// FOUND BY A SEAT, 2026-08-21, through the friction channel. A blue seat wrote its
		// proposal as a heredoc into `--reason-file -` — exactly what that flag is for, and the
		// only sane way to pass a paragraph — was refused, tried twice more, and filed a friction
		// report. It could not name the cause (the refusal pointed at a flag that does not exist)
		// but it was right that the tool was wrong.
		//
		// ONCE, because `--reason-file -` is STDIN: a second read of the channel returns nothing,
		// so resolving separately for each key would leave whichever key read second empty.
		prose, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		p.Set("line", prose)
		if prose != "" {
			p.Set("reason", prose)
		}
		// A fresh proposal with no stated fate is `proposed` — the state the old shape could
		// not express, which forced blue to declare a fate before it had one.
		p.Set("status", "proposed")
		if _, err := record.Append(s.Identity(), "line-of-inquiry", p); err != nil {
			return nil, err
		}
		return inquiryResult{ID: id, Status: "proposed", Line: p.Str("line")}, nil
	}), "line-of-inquiry"))

	seat.Supplies(c, "status", "a proposal starts at `proposed` — the state the field exists to express, and the one a seat would not think to type. A MOVE requires it")
	c.Flags().String(flags.Hypothesis, "", "what would be TRUE if this line pays off — the claim a later abandonment is judged against, so the fate is checkable rather than a shrug")
	c.Flags().String(flags.Method, "", "the source class or technique it belonged to, when that is what distinguishes it")
	return c
}

func newInquiryMove() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("move", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		id := seat.Str(cmd, flags.ID)
		if err := record.RequireInquiryRef(s.RunDir, id); err != nil {
			return nil, err
		}
		p := record.NewPayload().
			Set("inquiry_id", id).
			Set("supersedes_status", "1").
			Set("status", seat.Str(cmd, flags.As))
		// THE CONTEST IS `motion inquiry appeal` (#344), NOT A FIELD HERE. `contests_ruling`
		// was set as a side effect of moving a line to `pursued` against an adverse ruling,
		// and that coupling can only record disagreement that WINS: in one real record the
		// merge ruled a line too thin, blue argued the reasoning at the leaf and then
		// declined the line anyway — the ordinary outcome of an argument — and the field
		// recorded nothing. It appears zero times in the whole record.
		if err := seat.SetReason(cmd, p, "reason"); err != nil {
			return nil, err
		}
		if _, err := record.Append(s.Identity(), "line-of-inquiry", p); err != nil {
			return nil, err
		}
		return inquiryResult{ID: id, Status: p.Str("status"), Moved: true}, nil
	}), "line-of-inquiry"))

	c.Flags().Var(flags.InquiryID().WithCheck(record.InquiryExists), flags.ID, "`inquiry-id` — the line of inquiry whose fate you are moving (A1, A2 …); the lines-of-inquiry projection lists every one")
	_ = c.MarkFlagRequired(flags.ID)
	// THE VALUES ARE NOT RE-LISTED HERE. The hand-written line this replaced carried FOUR of the
	// five statuses — `deferred` had been added to InquiryStatuses and never to the string — and
	// glossed the four it did carry differently from the enum. enumhelp renders every value with
	// its own meaning from the record, so the usage line's job is to say what the FIELD is for.
	enumhelp.Flag(c, flags.As, record.MustEnum("line-of-inquiry", "status"), "the fate of this line of inquiry")
	return c
}

type inquiryResult struct {
	ID     string `json:"inquiry_id"`
	Status string `json:"status"`
	Line   string `json:"line,omitempty"`
	Moved  bool   `json:"moved,omitempty"`
}

func (r inquiryResult) Human() string {
	if r.Moved {
		return "line of inquiry " + r.ID + " moved to " + r.Status
	}
	return "line of inquiry " + r.ID + " recorded (" + r.Status + "): " + r.Line
}
