package lens

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// RED BRINGS THE TEXT (#1695 §III.B14, L37, L49). Red's closing on a gap docketed at impasse, or for a
// close or a supersede proposed on it, states the text red holds the gap's sentence should carry — one
// road: the span of the report as it stands, ending at the gap's marker, with what it becomes; or, on a
// gap about a citation or a computation, that anchor's retire — or --no-text where every road it can
// state is refused. The pair is planned as blue's `edit --accept` applies it (blue.PlanAccept), on the
// report the closing lands on, so a pair the accept would refuse is refused here and nothing is
// recorded; a for-raiser ruling applies it as the bench's edit. Which questions take text is the write's
// to say (record's requireClosingStatement).

// closingRed is the assignee's closing on its docketed gap.
func closingRed() *cobra.Command {
	return seat.Closing("closing-red", func(c *cobra.Command) {
		flags.Text(c, flags.Quote, "red's text, the span: the words of the report as it stands, ending at the gap's marker — the span a for-raiser ruling replaces. With --new")
		flags.Text(c, flags.New, "red's text, what the span becomes — applied verbatim, as blue applies a prescription it accepts, when the bench sides with the raiser. With --quote")
		c.Flags().Bool(flags.Retire, false, "red's text, on a gap about a citation or a computation: that anchor's retire, its sentence kept word for word, applied when the bench sides with the raiser")
		c.Flags().Bool(flags.NoText, false, "red states no text on this docketing, where every road it can state is refused — blue's accept of a prescription refuses every pair (an anchor stands inside the span), or no retire of the anchor lands. The bench then sides with the author or remands")
	}, func(s seat.Context, cmd *cobra.Command, body *recordpb.Closing) ([]record.Guard, error) {
		gapID := body.GetGapId()
		if seat.Given(cmd, flags.Quote) {
			body.Location = proto.String(seat.Str(cmd, flags.Quote))
		}
		if seat.Given(cmd, flags.New) {
			body.FixNew = proto.String(seat.Str(cmd, flags.New))
		}
		if on, _ := cmd.Flags().GetBool(flags.Retire); on {
			body.Retire = proto.Bool(true)
		}
		if on, _ := cmd.Flags().GetBool(flags.NoText); on {
			body.NoText = proto.Bool(true)
		}
		// THE ROAD IS PLANNED ON THE REPORT THE CLOSING LANDS ON, as the for-raiser ruling will plan it.
		var guards []record.Guard
		switch {
		case body.Location != nil && body.FixNew != nil:
			loc, fix := body.GetLocation(), body.GetFixNew()
			guards = append(guards, record.Check(func(l *record.Landing) error {
				report, err := l.Report()
				if err != nil {
					return err
				}
				if _, err := blue.PlanAccept(report, gapID, loc, fix); err != nil {
					return feov.Errorf(feov.Validation,
						"--quote and --new are no pair blue's `edit --accept` applies to the report as it stands: %v. Quote the span ending at %s's marker, word for word, and what it becomes; where an anchor stands inside every span you can quote, no pair can be stated — file --no-text. Nothing was recorded", err, gapID)
				}
				return nil
			}))
		case body.GetRetire():
			guards = append(guards, record.Check(func(l *record.Landing) error {
				ref, err := l.AboutAnchor(gapID)
				if err != nil || ref == "" {
					return err // a gap about no anchor: the write refuses --retire in its own words
				}
				return blue.RetireAnchor(&recordpb.Retire{}, ref)(l)
			}))
		}
		return guards, nil
	})
}
