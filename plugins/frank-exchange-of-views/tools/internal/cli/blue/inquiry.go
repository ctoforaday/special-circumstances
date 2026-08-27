package blue

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		id, err := record.MintInquiryID(run.Dir())
		if err != nil {
			return nil, err
		}
		proposed := recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED
		body := &recordpb.Avenue{
			AvenueId:   proto.String(id),
			Method:     proto.String(seat.Str(cmd, flags.Method)),
			Hypothesis: proto.String(seat.Str(cmd, flags.Hypothesis)),
			Status:     &proposed,
		}
		// A fresh proposal with no stated fate is `proposed` — the state the old shape could
		// not express, which forced blue to declare a fate before it had one. It is set on the
		// body above, where the enum makes it a value rather than a spelling.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// The record keeps its own word: the payload key is `line`, the flag is --reason.
		// Flag words are not payload keys (see internal/flags).
		//
		// BOTH FIELDS TAKE THE RESOLVED PROSE. `Line` read the FLAG while `Reason` took the
		// channel, so the same act wrote two different values whenever the seat used
		// --reason-file: the line went in empty and the reason went in whole. seat.Reason
		// resolves --reason, --reason-file and `--reason-file -` alike.
		body.Line = proto.String(why)
		body.Reason = proto.String(why)
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return inquiryResult{ID: id, Status: "proposed", Line: body.GetLine()}, nil
	}), "avenue"))

	seat.Supplies(c, "status", "a proposal starts at `proposed` — the state the field exists to express, and the one a seat would not think to type. A MOVE requires it")
	// `line` IS UNCONDITIONAL FOR THIS VERB, and conditional for the message. The field carries no
	// `required` annotation because a MOVE must not need it — but a PROPOSAL always does, and that
	// is a per-VERB fact no field annotation can state.
	//
	// Marked here so cobra refuses before an event exists and the refusal carries the help, which
	// is where a seat actually learns the verb. Without it the only refusal was validate's, which
	// arrives later and teaches less.
	c.MarkFlagsOneRequired(flags.Reason, flags.ReasonFile)
	c.Flags().String(flags.Hypothesis, "", "what would be TRUE if this line pays off — the claim a later abandonment is judged against, so the fate is checkable rather than a shrug")
	c.Flags().String(flags.Method, "", "the source class or technique it belonged to, when that is what distinguishes it")
	return c
}

func newInquiryMove() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("move", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		id := seat.Str(cmd, flags.ID)
		if err := record.RequireInquiryRef(run.Dir(), id); err != nil {
			return nil, err
		}
		st, known := record.AvenueStatusOf(seat.Str(cmd, flags.As))
		if !known {
			return nil, feov.Errorf(feov.Validation, "blue line-of-inquiry: %q is not a fate a line can take", seat.Str(cmd, flags.As))
		}
		body := &recordpb.Avenue{
			AvenueId:         proto.String(id),
			SupersedesStatus: proto.String("1"),
			Status:           &st,
		}
		// THE CONTEST IS `motion inquiry appeal` (#344), NOT A FIELD HERE. `contests_ruling`
		// was set as a side effect of moving a line to `pursued` against an adverse ruling,
		// and that coupling can only record disagreement that WINS: in one real record the
		// merge ruled a line too thin, blue argued the reasoning at the leaf and then
		// declined the line anyway — the ordinary outcome of an argument — and the field
		// recorded nothing. It appears zero times in the whole record.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Reason = proto.String(why)
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return inquiryResult{ID: id, Status: recordpb.Word(body.GetStatus()), Moved: true}, nil
	}), "avenue"))

	c.Flags().Var(flags.InquiryID().WithCheck(record.InquiryExists), flags.ID, "`inquiry-id` — the line of inquiry whose fate you are moving (A1, A2 …); the lines-of-inquiry projection lists every one")
	_ = c.MarkFlagRequired(flags.ID)
	// THE VALUES ARE NOT RE-LISTED HERE. The hand-written line this replaced carried FOUR of the
	// five statuses — `deferred` had been added to InquiryStatuses and never to the string — and
	// glossed the four it did carry differently from the enum. enumhelp renders every value with
	// its own meaning from the record, so the usage line's job is to say what the FIELD is for.
	enumhelp.Flag(c, flags.As, record.MustEnum("avenue", "status"), "the fate of this line of inquiry")
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
