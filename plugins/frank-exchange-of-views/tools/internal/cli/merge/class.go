package merge

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// class: the gap-class registry, and coining an entry in it.
//
// IT WAS FOUR FLAGS ON `mint`, and that is what a verb looks like when it is wearing another
// verb's clothes. `--class-new` was a boolean whose whole meaning was "I also passed
// --definition, --neighbor and --distinguisher"; the three of them were optional to cobra and
// policed in the mint handler; and a mint that coined a class wrote TWO events, one of which had
// nothing to do with putting a gap on the board.
//
// A cluster of flags that must travel together, changes what else is required, and writes its
// own event, is a verb. As one, cobra refuses an incomplete coining at parse and `--help` can
// say what a class IS without that competing for room with what a gap is.
func newClass() *cobra.Command {
	c := &cobra.Command{
		Use:          "class",
		Short:        "the gap-class registry: what KINDS of defect this run recognises",
		SilenceUsage: true,
	}
	c.AddCommand(newClassNew())
	return c
}

func newClassNew() *cobra.Command {
	c := seat.Records(seat.New("new", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		body := &recordpb.ClassNew{
			Slug:          proto.String(seat.Str(cmd, flags.Class)),
			Definition:    proto.String(seat.Str(cmd, flags.Definition)),
			Neighbor:      proto.String(seat.Str(cmd, flags.Neighbor)),
			Distinguisher: proto.String(seat.Str(cmd, flags.Distinguisher)),
		}
		if _, err := record.Append(s.Identity(), body); err != nil {
			return nil, err
		}
		return classResult{Slug: seat.Str(cmd, flags.Class)}, nil
	}), "class-new")

	c.Flags().String(flags.Class, "", "the slug you are coining — lowercase, hyphenated, and NOT one the registry already has")
	c.Flags().String(flags.Definition, "", "what this class is, in one line")
	c.Flags().String(flags.Neighbor, "", "the existing class it sits closest to")
	c.Flags().String(flags.Distinguisher, "", "the tie-break question that tells the two apart — without it a new class is a synonym, and the registry stops discriminating")
	_ = c.MarkFlagRequired(flags.Class)
	// COBRA SAYS THE TRIO TRAVELS TOGETHER. It was a boolean plus three optional flags checked
	// in a handler, so a coining missing its distinguisher was composed, sent, and refused after
	// the fact — if it was refused at all.
	c.MarkFlagsRequiredTogether(flags.Definition, flags.Neighbor, flags.Distinguisher)
	_ = c.MarkFlagRequired(flags.Definition)
	return c
}

type classResult struct {
	Slug string `json:"slug"`
}

func (r classResult) Human() string {
	return "class " + r.Slug + " coined — mint against it with `merge mint --class " + r.Slug + "`"
}
