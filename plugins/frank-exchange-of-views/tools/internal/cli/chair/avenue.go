package chair

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// avenue review: red's per-epoch verdict that the REPORT still carries an avenue.
//
// # The one claim in the document nothing could check
//
// An avenue reaches the report as a row `assemble` GENERATES from the record. It carries no
// citation anchor, so `lens verify` cannot reach it and no gap can be minted against it by the
// ordinary route. "We pursued X", "we deferred Y", "we abandoned Z and here is why" were assertions
// the adversarial half of this system had no channel to answer — a fact written where nothing can
// refuse it, which is the defect class this repository keeps finding in its own surfaces.
//
// Measured over six runs: 83 of 86 lines were declared in round 0 and never moved. That number was
// read as blue neglecting to revisit them. It is equally consistent with nobody ever ASKING, and
// nothing in the tool asked.
//
// # Why it is red's, and why every epoch
//
// The check is a read of the artifact — is this line in the report, and does the text still back it
// as STATED — which is red's discipline, not blue's self-report. And it is per-epoch because the
// report changes every epoch: a verdict cast before this epoch's edits answers a question about a
// document that no longer exists. The unvoted-avenue check keys on the epoch for exactly that
// reason, and the chair's sitting is not complete while any line is unvoted.
//
// # ONE READ PER SITTING, NOT ONE READ PER LINE
//
// The votes are recorded per line because each line gets its own verdict and its own quote — that
// is the record's shape and it is right. The READING is not per line: you read the report once this
// sitting and answer every line against that one pass, the way anyone checks a document against a
// list. A dozen lines over four epochs would otherwise be forty-eight full reads of the same
// artifact on the most expensive seat in the run, and a duty that costs that much is one seats
// route around — which is worse than no duty, because the record then says checked.
//
// # The verdict is NOT the ruling
//
// `motion avenue rule` answers "is this direction worth the run's time" — red's judgement about
// the RESEARCH. This answers "does the report carry it" — red's read of the ARTIFACT. A line can be
// `endorsed` and `absent` at once: red agreed it was worth taking and the section that took it has
// since been cut. Two questions, two records, and collapsing them would lose the one that says the
// document drifted from its own account of itself.
// newAvenue is the chair's half of the one verb the concept has. Blue's surface carries
// `avenue propose | move`; this carries `avenue review` — the same group, scoped by role the way
// `motion`'s subjects are, so a seat asking what it may do with an avenue looks in one place.
func newAvenue() *cobra.Command {
	c := &cobra.Command{
		Use:          "avenue",
		Short:        "read the report against every avenue on the record — `review` states what that reading found",
		SilenceUsage: true,
	}
	c.AddCommand(newAvenueReview())
	return c
}

func newAvenueReview() *cobra.Command {
	c := seat.Prose(seat.New("review", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// THE SCHEMA COLLAPSED THIS EVENT AND THE VERB HAS NOT CAUGHT UP. AvenueReview
		// carries `reason` alone: the per-line grade is retired because "a line is treated
		// thinly" is a DEFECT IN THE REPORT, which the schema says belongs on the board as a
		// minted gap with the lifecycle, the blue duty, the grade and the PASS gate every
		// other gap has — "a second vocabulary for the same fact is exactly the aliasing this
		// schema exists to remove".
		//
		// So --id and --as no longer reach the record. That is the schema's decision, not a
		// conversion slip: the flags are gone from the surface and from the help,
		if _, err := record.Append(s.Identity(), &recordpb.AvenueReview{Reason: proto.String(text)}); err != nil {
			return nil, err
		}
		return avenueReviewResult{}, nil
	}))

	// NO --id AND NO --as, AND THE ABSENCE IS THE RULING.
	//
	// THE HELP ABOVE ADVERTISED BOTH UNTIL NOW — `--id Q1 --as supported|weakened|unsupported|absent`
	// on a command that registers neither, so a seat following its own help got a cobra refusal for
	// a flag the schema deliberately retired. The comment below already called removing them "the
	// follow-through this branch still owes"; the flags went and the sentence describing them
	// stayed, which is the half-state that reads as done. Third instance of that family this
	// branch: `--as supports-with-bridge` advertised and refused, and the fuzz's hyphenated
	// ruling words.
	//
	// The four-value `--as` answered "does the report still carry this line". Presence stopped
	// being a question when the lines became GENERATED onto the page from the record — blue cannot
	// cut one — and the surviving question, whether the body delivered the research, is an ORDINARY
	// GAP with the grade vocabulary it already has. record/enums.go states the same ruling from the
	// other side by declining to declare a set: "a second vocabulary for the same fact is exactly
	// the aliasing this schema exists to remove".
	//
	// A flag a seat can pass and the record ignores is worse than one that does not exist, so both
	// go rather than lingering as accepted-and-discarded. One read of the document per sitting,
	// recorded as prose.
	//
	// The verb's name is not its event's word — it records an avenue_review — so the event is
	// declared.
	seat.Records(c, "avenue_review")
	// validate refuses a review that says nothing — silence cannot clear the duty.
	return seat.SaysRequired(c, flags.Reason)
}

type avenueReviewResult struct{}

func (r avenueReviewResult) Human() string { return "avenue review recorded" }
