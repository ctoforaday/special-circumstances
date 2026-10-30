package blue

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// The pages of the two acts that write an inquiry print which of the report's inquiry sections are
// whose, from the assembler's own sentence.
func init() { seat.HelpValues["InquirySectionsRule"] = report.InquirySectionsRule }

// inquiry: a question the research could take up, and what became of it.
//
// think-around-problem mandates exploring genuinely distinct alternatives before
// a significant decision. terse-communication forbids narrating options. So the
// exploration is required, invisible and unverifiable unless something records it.
//
// This is what instruments it. Exploration becomes a RECORD and the record
// becomes a report section, so the rule stops being self-attested and starts
// leaving artifacts.
//
// THE FATES, each answering a different question:
//
//	declined   considered, never run          — was setting it aside REASONED, or decoration?
//	abandoned  started, and no answer came    — what stopped it?
//	refuted    ran, and its hypothesis did not hold — what did it find? (the negative result)
//	confirmed  ran, and the hypothesis held   — what did it find?
//	underway   still being worked             — what has it found so far?
//
// `refuted` and `abandoned` are the most valuable and the most routinely lost: negative
// results and dead ends are exactly what a future run needs so it does not re-walk them, and
// they are direct feedstock for the sleeper service's subject mining.
//
// TWO ACTS, AND THEY ARE TWO VERBS.
//
//	propose  --question "<the question>" --hypothesis "<what would be true if it pays off>"
//	record   --id Q-… --as underway|confirmed|refuted|declined|abandoned|deferred --reason "<what settled it>" [--method "<what was done>"]
//
// A proposal needs a question and starts `proposed`; a record needs the id, the fate and the
// account of it. Split, each verb requires what it requires and cobra can mark it.
//
// THE METHOD IS THE FATE'S, NOT THE PROPOSAL'S (owner ruling, 2026-10-17). Written at proposal it was
// written before any research and printed beside the question as if it were what was consulted,
// checked by nothing and changed by no act. On the fate it is report text blue changes the way it
// changes the reason: by recording the fate again, and the latest record prints.
//
// --hypothesis is what makes a later fate checkable: an inquiry that says what would be true if it
// paid off is confirmed or refuted against its own claim rather than set down on a shrug. It is
// the ONLY place a predicted result goes: the question holds none, and it is never rewritten.
func newInquiry() *cobra.Command {
	c := &cobra.Command{
		Use:          "inquiry",
		Short:        "propose a question the research could take up, and say what became of it — `propose` one, `record` the fate of one you proposed",
		SilenceUsage: true,
	}
	c.AddCommand(newInquiryPropose(), newInquiryRecord())
	return c
}

// propose is open at any point in a run because an inquiry that occurs to a seat in its third
// sitting is exactly the one a later run needs. --hypothesis exists so the fate is judged against a
// stated prediction: checkable, where a fate with none is a shrug.
func newInquiryPropose() *cobra.Command {
	c := seat.Records(seat.New("propose", inquiryResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// The question takes the prose channel's trim: a value written through a heredoc ends in a
		// newline the seat did not mean as text.
		question := strings.TrimRight(seat.Str(cmd, flags.Question), "\n")
		// CRASH-RETRY IDEMPOTENCY, before an id is minted. A seat whose call returned nothing retries
		// it, and without this the retry is a second inquiry asking the same question.
		prior, err := record.ExistingProposalByText(run, s.SeatID, question, seat.Str(cmd, flags.Hypothesis))
		if err != nil {
			return nil, err
		}
		if prior != "" {
			return inquiryResult{ID: prior, Status: "proposed", Question: question, Idempotent: true}, nil
		}
		id := record.NewID("inquiry")
		proposed := recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED
		// A fresh proposal is `proposed`, set on the body where the enum makes it a value rather
		// than a spelling. It holds a question and NO reason: a reason is the account of a fate,
		// and a proposal has none yet.
		body := &recordpb.Inquiry{
			InquiryId:  proto.String(id),
			Question:   proto.String(question),
			Hypothesis: proto.String(seat.Str(cmd, flags.Hypothesis)),
			Status:     &proposed,
		}
		// THE ADVISORY, over exactly what `inquiries()` composes into report.md from this act: the
		// question and the hypothesis, each printed under its own label on the inquiry's row.
		tells := spanVoiceTells(question, seat.Str(cmd, flags.Hypothesis))
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return inquiryResult{ID: id, Status: "proposed", Question: body.GetQuestion(), VoiceTells: tells}, nil
	}), "inquiry")

	seat.Supplies(c, "status", "a proposal starts at `proposed` — the state the field exists to express, and the one a seat would not think to type. `inquiry record` takes the fate")
	// THE QUESTION IS UNCONDITIONAL FOR THIS VERB, and conditional for the message: the field carries
	// no `required` annotation because a record of a fate does not repeat it — but a PROPOSAL always
	// needs one, and that is a per-VERB fact no field annotation can state. Marked here so cobra
	// refuses before an event exists and the refusal carries the help.
	flags.Text(c, flags.Question, "the QUESTION this inquiry asks, holding no predicted result — research prose for a reader of the SUBJECT, printed in the report word for word as the inquiry's question; a result you expect goes in --hypothesis")
	seat.Require(c, flags.Question)
	flags.Text(c, flags.Hypothesis, "your PREDICTION: what would be TRUE if this inquiry pays off. The report prints it as the inquiry's hypothesis, and the fate is judged against it — confirmed where it held, refuted where it did not")
	return c
}

// record: the fate is not the point; the account of what settled it is. An inquiry given its fate
// against its own stated hypothesis is evidence of choosing, and one set down on a shrug is not —
// which is why the reason is required where the record requires it and printed in the report.
func newInquiryRecord() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("record", inquiryResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		id := seat.Str(cmd, flags.ID)
		if err := record.RequireInquiryRef(run, id); err != nil {
			return nil, err
		}
		st, known := record.InquiryStatusOf(seat.Str(cmd, flags.As))
		if !known {
			return nil, feov.Errorf(feov.Validation, "%q is not a fate an inquiry can take", seat.Str(cmd, flags.As))
		}
		body := &recordpb.Inquiry{
			InquiryId:        proto.String(id),
			SupersedesStatus: proto.String("1"),
			Status:           &st,
		}
		// RED'S RULING ON THE INQUIRY BINDS NOTHING HERE. It is an argument beside the question;
		// this records what blue decided — any fate, against the ruling or with it — and no field
		// marks a record as made against one.
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		body.Reason = proto.String(why)
		method := strings.TrimRight(seat.Str(cmd, flags.Method), "\n")
		if method != "" {
			body.Method = proto.String(method)
		}
		// A fate's reason and method are composed into the same report.md row as the question, so
		// they meet the same advisory.
		tells := spanVoiceTells(why, method)
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return inquiryResult{ID: id, Status: recordpb.Word(body.GetStatus()), Recorded: true, VoiceTells: tells}, nil
	}), "inquiry"))

	c.Flags().Var(flags.InquiryID().WithCheck(record.InquiryExists), flags.ID, "the inquiry whose fate you are recording — the id its proposal returned, which show inquiries lists with every inquiry's fate")
	seat.Require(c, flags.ID)
	// THE VALUES ARE NOT RE-LISTED HERE: enumhelp renders every value with its own meaning from the
	// record, so the usage line's job is to say what the FIELD is for.
	enumhelp.Flag(c, flags.As, record.MustEnum("inquiry", "status"), "the fate of this inquiry")
	flags.Text(c, flags.Method, "optional: the method this fate is the account of, in the SUBJECT's terms — what was actually done for an approach that ran, what a later run should do for one deferred or declined, what was tried for one abandoned. Printed on the inquiry's row beside the fate")
	return c
}

type inquiryResult struct {
	ID       string `json:"inquiry_id"`
	Status   string `json:"status"`
	Question string `json:"question,omitempty"`
	Recorded bool   `json:"recorded,omitempty"`
	// Idempotent says this call recorded NOTHING and returned the question it already had — a retry
	// of a propose whose first attempt the seat never saw the answer to.
	Idempotent bool `json:"idempotent,omitempty"`
	// VoiceTells is ADVICE, and the act is already recorded by the time it renders. An inquiry
	// is composed into report.md, which is written for a reader of the SUBJECT; this names
	// where the text sounds like the run talking about itself. It refuses nothing — a pattern cannot
	// tell a question narrating its own process from a question naming a source that does.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r inquiryResult) Human() string {
	head := "inquiry " + r.ID + " recorded (" + r.Status + "): " + r.Question
	if r.Idempotent {
		head = seat.Repeated("inquiry " + r.ID)
	}
	if r.Recorded {
		head = "inquiry " + r.ID + " recorded as " + r.Status
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
