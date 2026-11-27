package blue

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// BLUE STATES ITS WORD (#1695 §III.B14, L34). Blue's closing on a gap docketed at impasse may state the
// closing word a for-author ruling writes; with none stated, that ruling writes `not_a_defect` — blue's
// text stands. On a proposal blue's statement is the proposal, and at grade movement the recorded
// grades, so the word is refused there (record's requireClosingStatement).

// closingBlue is blue-respond's closing on a docketed gap.
func closingBlue() *cobra.Command {
	return seat.Closing("closing-blue", func(c *cobra.Command) {
		enumhelp.Flag(c, flags.As, record.MustEnum("closing", "closure_class"),
			"on a gap docketed at impasse: the closing word a ruling for blue writes — omit it and that ruling writes not_a_defect, your text standing. Not repaired_with_regression or amends_prior")
	}, func(_ seat.Context, cmd *cobra.Command, body *recordpb.Closing) ([]record.Guard, error) {
		if w := seat.Str(cmd, flags.As); w != "" {
			d, ok := record.DispositionOf(w)
			if !ok {
				return nil, feov.Errorf(feov.Validation, "--as %q is not a closing word", w)
			}
			body.ClosureClass = &d
		}
		return nil, nil
	})
}
