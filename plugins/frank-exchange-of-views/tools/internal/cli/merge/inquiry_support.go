package merge

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// inquiry-support: red's per-round verdict that the REPORT still carries a line of inquiry.
//
// # The one claim in the document nothing could check
//
// A line of inquiry reaches the report as a row `assemble` GENERATES from the record. It carries no
// citation anchor, so `lens verify` cannot reach it and no gap can be minted against it by the
// ordinary route. "We pursued X", "we deferred Y", "we abandoned Z and here is why" were assertions
// the adversarial half of this system had no channel to answer — a fact written where nothing can
// refuse it, which is the defect class this repository keeps finding in its own surfaces.
//
// Measured over six runs: 83 of 86 lines were declared in round 0 and never moved. That number was
// read as blue neglecting to revisit them. It is equally consistent with nobody ever ASKING, and
// nothing in the tool asked.
//
// # Why it is red's, and why every round
//
// The check is a read of the artifact — is this line in the report, and does the text still back it
// as STATED — which is red's discipline, not blue's self-report. And it is per-round because the
// report changes every round: a verdict cast before this round's edits answers a question about a
// document that no longer exists. `record.UnvotedInquiries` keys on the round for exactly that
// reason, and the merge's sitting is not complete while any line is unvoted.
//
// # ONE READ PER ROUND, NOT ONE READ PER LINE
//
// The votes are recorded per line because each line gets its own verdict and its own quote — that
// is the record's shape and it is right. The READING is not per line: you read the report once this
// round and answer every line against that one pass, the way anyone checks a document against a
// list. A dozen lines over four rounds would otherwise be forty-eight full reads of the same
// artifact on the most expensive seat in the run, and a duty that costs that much is one seats
// route around — which is worse than no duty, because the record then says checked.
//
// # The verdict is NOT the ruling
//
// `motion inquiry rule` answers "is this direction worth the run's time" — red's judgement about
// the RESEARCH. This answers "does the report carry it" — red's read of the ARTIFACT. A line can be
// `endorsed` and `absent` at once: red agreed it was worth taking and the section that took it has
// since been cut. Two questions, two records, and collapsing them would lose the one that says the
// document drifted from its own account of itself.
func newInquirySupport() *cobra.Command {
	c := seat.Prose(seat.New("inquiry-support",
		`your per-round verdict that the REPORT still carries a line of inquiry: --id Q1 --as supported|weakened|unsupported|absent --reason "<what the report says for it, quoted>". Read the report ONCE this round and vote every line against that read`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			// THE SCHEMA COLLAPSED THIS EVENT AND THE VERB HAS NOT CAUGHT UP. InquiryReview
			// carries `reason` alone: the per-line grade is retired because "a line is treated
			// thinly" is a DEFECT IN THE REPORT, which the schema says belongs on the board as a
			// minted gap with the lifecycle, the blue duty, the grade and the PASS gate every
			// other gap has — "a second vocabulary for the same fact is exactly the aliasing this
			// schema exists to remove".
			//
			// So --id and --as no longer reach the record. That is the schema's decision, not a
			// conversion slip, and it is written here rather than left as a silent drop: the
			// follow-through this branch still owes is removing the two flags from the surface,
			// because a flag a seat can pass and the record ignores is worse than one that does
			// not exist.
			if _, err := record.Append(s.Identity(), &recordpb.InquiryReview{Reason: proto.String(text)}); err != nil {
				return nil, err
			}
			return inquirySupportResult{ID: seat.Str(cmd, flags.ID), As: seat.Str(cmd, flags.As)}, nil
		}))

	c.Flags().Var(flags.InquiryID().WithCheck(record.InquiryExists), flags.ID,
		"the line of inquiry you are voting on — `show lines-of-inquiry` lists every one")
	enumhelp.Flag(c, flags.As, record.MustEnum("inquiry-support", "as"),
		"REQUIRED — what the report does for this line RIGHT NOW, from this round's read of the document rather than from the record you already have. `unsupported` and `absent` put it on blue's work list; `weakened` is a flag blue may answer or accept")
	return c
}

type inquirySupportResult struct {
	ID string `json:"id"`
	As string `json:"as"`
}

func (r inquirySupportResult) Human() string {
	return "line of inquiry " + r.ID + " voted " + r.As
}
