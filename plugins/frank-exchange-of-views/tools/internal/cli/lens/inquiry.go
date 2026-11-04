package lens

import (
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// inquiry check: the owning lens's judgement that a settled inquiry is sound (ruled by gblock
// 2026-11-01: an inquiry's quality judgement lives with a lens, and a shortfall is an ordinary gap).
//
// ONE CHECK PER SETTLED INQUIRY, WITH ITS OWN REASON. One reason across unlike fates is a prose read
// of the report with ids on it; the record's unit is the inquiry, and a seat reads the report once
// and records many checks in one turn.
//
// A POSITIVE JUDGEMENT ONLY. An inquiry the lens finds wanting gets the gap about it, which reaches
// blue with a grade and a fix; there is no `unsound` word to set beside it. The record refuses the
// check of an inquiry that is open, wanting, already checked, or whose fate landed after the seat's
// sitting opened, and any seat but the owner's (record/inquirycheck.go).
//
// newInquiry is the lens's half of the one group the concept has: blue's surface carries `inquiry
// propose | record`, and this carries `inquiry check`, so a seat asking what it may do with an
// inquiry looks in one place.
func newInquiry() *cobra.Command {
	c := &cobra.Command{
		Use:          "inquiry",
		Short:        "judge an inquiry blue settled — `check` records that its latest fate stands as the report's account of it",
		SilenceUsage: true,
	}
	c.AddCommand(newInquiryCheck())
	return c
}

func newInquiryCheck() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("check", inquiryCheckResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		id := seat.Str(cmd, flags.ID)
		why, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// A RETRY OF THE CHECK THAT STANDS records nothing and says so: the same seat, the same
		// inquiry, the same reason. A different reason while a check stands is refused by the write,
		// which names what makes the inquiry owed a check again.
		standing, err := record.StandingInquiryCheck(run, id)
		if err != nil {
			return nil, err
		}
		if standing != nil && standing.SeatID == s.SeatID && standing.Reason == strings.TrimRight(why, "\n") {
			return inquiryCheckResult{ID: id, Repeated: true}, nil
		}
		if _, err := record.Append(s.Identity(), &recordpb.InquiryCheck{InquiryId: proto.String(id), Reason: proto.String(why)}); err != nil {
			return nil, err
		}
		return inquiryCheckResult{ID: id}, nil
	}), "inquiry_check"))
	c.Flags().Var(flags.InquiryID().WithCheck(record.InquiryExists), flags.ID, "the inquiry you judge — its `Q-` id, which `show inquiries` lists with its question, its hypothesis and its latest fate")
	seat.Require(c, flags.ID)
	// The write refuses a check that says nothing, in its words for what the reason is for.
	return seat.SaysRequired(c, flags.Reason)
}

type inquiryCheckResult struct {
	ID string `json:"inquiry_id"`
	// Repeated says this call recorded NOTHING: the same check already stands on the inquiry.
	Repeated bool `json:"repeated,omitempty"`
}

func (r inquiryCheckResult) Human() string {
	if r.Repeated {
		return seat.Repeated("the check of inquiry " + r.ID)
	}
	return "inquiry " + r.ID + " checked"
}
