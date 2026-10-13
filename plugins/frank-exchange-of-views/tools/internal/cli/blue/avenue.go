package blue

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// The pages of the two acts that write an avenue print whose the report's avenue sections are, from
// the assembler's own sentence.
func init() { seat.HelpValues["AvenueSectionsRule"] = report.AvenueSectionsRule }

// avenue: an avenue, and what became of it.
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
// THE STATUSES, each answering a different question:
//
//	declined   considered, never run          — was setting it aside REASONED, or decoration?
//	abandoned  started, and no answer came    — what stopped it?
//	refuted    ran, and its hypothesis did not hold — what did it find? (the negative result)
//	concluded  ran, and the hypothesis held   — what did it find?
//	pursued    became the report's spine      — does the report reflect it honestly?
//
// `refuted` and `abandoned` are the most valuable and the most routinely lost: negative
// results and dead ends are exactly what a future run needs so it does not re-walk them, and
// they are direct feedstock for the sleeper service's subject mining. Nothing in the
// engine preserved them before this verb — they died in a seat's context.
//
// TWO FORMS, AND THEY ARE TWO VERBS (#246).
//
//	propose  --reason "<the approach>" --hypothesis "<what would be true if it pays off>"
//	move     --id A1 --as pursued|concluded|refuted|declined|abandoned|deferred --reason "<what changed>"
//
// The move form is the whole point. Measured over 86 avenue events in six runs: ZERO
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
// --hypothesis is what makes a later move checkable: an avenue that says what would be true if it
// paid off is concluded or refuted against its own claim rather than set down on a shrug. It is
// the ONLY place a predicted result goes: the line is the question, and it is never rewritten.
func newAvenue() *cobra.Command {
	c := &cobra.Command{
		Use:          "avenue",
		Short:        "an approach the research could take, and what became of it — `propose` one, `move` one you already proposed",
		SilenceUsage: true,
	}
	c.AddCommand(newAvenuePropose(), newAvenueMove())
	return c
}

// propose is open at any point in a run because an avenue that occurs to a seat in its third sitting
// is exactly the one a later run needs. --hypothesis exists so the fate is judged against a stated
// prediction: checkable, where a fate with none is a shrug.
func newAvenuePropose() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("propose", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// CRASH-RETRY IDEMPOTENCY, before an id is minted. A seat whose call returned nothing retries
		// it, and without this the retry is a second avenue saying the same thing — ten for
		// five on universe-m12.
		prior, err := record.ExistingProposalByText(run, s.SeatID, why, seat.Str(cmd, flags.Hypothesis))
		if err != nil {
			return nil, err
		}
		if prior != "" {
			return avenueResult{ID: prior, Status: "proposed", Line: why, Idempotent: true}, nil
		}
		id := record.NewID("avenue")
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
		// The record keeps its own word: the payload key is `line`, the flag is --reason.
		// Flag words are not payload keys (see internal/flags).
		//
		// BOTH FIELDS TAKE THE RESOLVED PROSE. `Line` read the FLAG while `Reason` took the
		// channel, so the same act wrote two different values whenever the seat used
		// the file spelling that flag used to have: the line went in empty and the reason went in
		// whole. Both read seat.Reason, the one resolver.
		body.Line = proto.String(why)
		body.Reason = proto.String(why)
		// THE ADVISORY, over exactly what `avenues()` composes into report.md from this act: the
		// line (which is also the reason — propose writes the same resolved prose to both), the
		// hypothesis, printed under its own label on the avenue's row, and the method, rendered
		// there as _(method)_.
		//
		// Measured before this existed: arm A of the #861 smoke carried a process-voice tell in
		// Alternatives considered that came from a blue lane's line, through a verb no advisory saw.
		tells := spanVoiceTells(why, seat.Str(cmd, flags.Hypothesis), seat.Str(cmd, flags.Method))
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return avenueResult{ID: id, Status: "proposed", Line: body.GetLine(), VoiceTells: tells}, nil
	}), "avenue"))

	seat.Supplies(c, "status", "a proposal starts at `proposed` — the state the field exists to express, and the one a seat would not think to type. A MOVE requires it")
	// `line` IS UNCONDITIONAL FOR THIS VERB, and conditional for the message. The field carries no
	// `required` annotation because a MOVE must not need it — but a PROPOSAL always does, and that
	// is a per-VERB fact no field annotation can state.
	//
	// Marked here so cobra refuses before an event exists and the refusal carries the help, which
	// is where a seat actually learns the verb. Without it the only refusal was validate's, which
	// arrives later and teaches less.
	seat.Require(c, flags.Reason)
	flags.Text(c, flags.Hypothesis, "your PREDICTION: what would be TRUE if this avenue pays off. The report prints it as the avenue's hypothesis, and the fate is judged against it — concluded where it held, refuted where it did not")
	flags.Text(c, flags.Method, "the source class or technique it belonged to, when that is what distinguishes it")
	return c
}

// move: the fate is not the point; the account of what settled it is. An avenue given its fate
// against its own stated hypothesis is evidence of choosing, and one set down on a shrug is not —
// which is why the reason is required where the record requires it and printed in the report.
func newAvenueMove() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("move", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		id := seat.Str(cmd, flags.ID)
		if err := record.RequireAvenueRef(run, id); err != nil {
			return nil, err
		}
		st, known := record.AvenueStatusOf(seat.Str(cmd, flags.As))
		if !known {
			return nil, feov.Errorf(feov.Validation, "%q is not a fate an avenue can take", seat.Str(cmd, flags.As))
		}
		body := &recordpb.Avenue{
			AvenueId:         proto.String(id),
			SupersedesStatus: proto.String("1"),
			Status:           &st,
		}
		// RED'S RULING ON THE AVENUE BINDS NOTHING HERE. It is an argument beside the line; the
		// move records what blue decided — any status, against the ruling or with it — and no
		// field marks a move as made against one.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Reason = proto.String(why)
		// A move's reason is composed into the same report.md row as the line it moves, so it
		// meets the same advisory.
		tells := spanVoiceTells(why)
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return avenueResult{ID: id, Status: recordpb.Word(body.GetStatus()), Moved: true, VoiceTells: tells}, nil
	}), "avenue"))

	c.Flags().Var(flags.AvenueID().WithCheck(record.AvenueExists), flags.ID, "the avenue whose fate you are moving; the avenues projection lists every one")
	seat.Require(c, flags.ID)
	// THE VALUES ARE NOT RE-LISTED HERE. The hand-written line this replaced carried FOUR of the
	// five statuses — `deferred` had been added to AvenueStatuses and never to the string — and
	// glossed the four it did carry differently from the enum. enumhelp renders every value with
	// its own meaning from the record, so the usage line's job is to say what the FIELD is for.
	enumhelp.Flag(c, flags.As, record.MustEnum("avenue", "status"), "the fate of this avenue")
	return c
}

type avenueResult struct {
	ID     string `json:"avenue_id"`
	Status string `json:"status"`
	Line   string `json:"line,omitempty"`
	Moved  bool   `json:"moved,omitempty"`
	// Idempotent says this call recorded NOTHING and returned the line it already had — a retry of a
	// propose whose first attempt the seat never saw the answer to.
	Idempotent bool `json:"idempotent,omitempty"`
	// VoiceTells is ADVICE, and the act is already recorded by the time it renders. An avenue
	// is composed into report.md, which is written for a reader of the SUBJECT; this names
	// where the text sounds like the run talking about itself. It refuses nothing — a pattern cannot
	// tell a line narrating its own process from a line naming a source that does.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r avenueResult) Human() string {
	head := "avenue " + r.ID + " recorded (" + r.Status + "): " + r.Line
	if r.Idempotent {
		head = "avenue " + r.ID + " already records this avenue (idempotent retry — nothing written)"
	}
	if r.Moved {
		head = "avenue " + r.ID + " moved to " + r.Status
	}
	if len(r.VoiceTells) == 0 {
		return head
	}
	return head + "\n\nNOTE — this text is composed into report.md (Research areas, Future research\n" +
		"directions, Alternatives considered) and in places sounds like the run rather than\n" +
		"the subject. It is recorded; this is not a refusal, and it may be wrong:\n  - " +
		strings.Join(r.VoiceTells, "\n  - ") +
		"\n\nSEPARATION, NEVER DELETION: a real limit on the CONCLUSION stays, re-voiced as a\n" +
		"limit on the subject; only the fact about the run goes. Red's voice lens holds that\n" +
		"judgement — these are only the literal tells."
}
