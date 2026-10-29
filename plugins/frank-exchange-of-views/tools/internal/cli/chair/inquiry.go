package chair

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// inquiry review: red's per-epoch read of the report's account of its inquiries.
//
// # The one claim in the document nothing else checks
//
// An inquiry reaches the report as a row `assemble` GENERATES from the record. It carries no
// citation anchor, so `lens verify` cannot reach it and no gap can be minted against it by the
// ordinary route. "We pursued X", "we deferred Y", "we abandoned Z and here is why" are assertions
// the adversarial half of this system would otherwise have no channel to answer — a fact written
// where nothing can refuse it.
//
// # Why it is red's, and why every epoch
//
// The check is a read of the artifact — does the text back each inquiry as STATED — which is red's
// discipline, not blue's self-report. And it is per-epoch because the report changes every epoch: a
// read made before this epoch's edits answers a question about a document that no longer exists.
//
// # ONE READ PER SITTING, NOT ONE READ PER INQUIRY
//
// Read the report once this sitting and answer every inquiry against that one pass, the way anyone
// checks a document against a list. A read per inquiry is a dozen full reads of the same artifact
// on the most expensive seat in the run, and a duty that costs that much is one seats route around
// — which is worse than no duty, because the record then says checked.
//
// # The read is NOT the ruling
//
// `motion inquiry rule` answers "is this inquiry worth the run's time" — red's judgement about the
// RESEARCH. This answers "does the report's body deliver what its inquiries claim" — red's read of
// the ARTIFACT. Two questions, two records.
//
// newInquiry is the chair's half of the one group the concept has. Blue's surface carries
// `inquiry propose | record`; this carries `inquiry review` — the same group, scoped by role the way
// `motion`'s subjects are, so a seat asking what it may do with an inquiry looks in one place.
func newInquiry() *cobra.Command {
	c := &cobra.Command{
		Use:          "inquiry",
		Short:        "read the report against every inquiry on the record — `review` states what that reading found",
		SilenceUsage: true,
	}
	c.AddCommand(newInquiryReview())
	return c
}

func newInquiryReview() *cobra.Command {
	c := seat.Prose(seat.New("review", inquiryReviewResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// InquiryReview carries `reason` alone: an inquiry treated thinly is a DEFECT IN THE
		// REPORT, which belongs on the board as a minted gap with the lifecycle, the blue duty, the
		// grade and the PASS gate every other gap has — not a second vocabulary for the same fact.
		if _, err := record.Append(s.Identity(), &recordpb.InquiryReview{Reason: proto.String(text)}); err != nil {
			return nil, err
		}
		return inquiryReviewResult{}, nil
	}))

	// NO --id AND NO --as. Inquiries reach the page GENERATED from the record — blue cannot cut one —
	// so presence is not a question, and whether the body delivered the research is an ordinary gap
	// with the grade vocabulary it already has. One read of the document per sitting, recorded as
	// prose.
	//
	// The verb's name is not its event's word — it records an inquiry_review — so the event is
	// declared.
	seat.Records(c, "inquiry_review")
	// validate refuses a review that says nothing — silence cannot clear the duty.
	return seat.SaysRequired(c, flags.Reason)
}

type inquiryReviewResult struct{}

func (r inquiryReviewResult) Human() string { return "inquiry review recorded" }
