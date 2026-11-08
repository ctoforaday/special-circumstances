package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/jsonshape"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// newShowAssembled is the OPERATOR's read of the assembled report set, mounted under `ops`.
//
// # Why it is not a seat projection
//
// A seat reads the report's text with `show report`, and the content of each section the tool
// composes through the record view that serves its source: `show board`, `show inquiries`, `show
// evidence`, `show changes`, and for the seats that adjudicate `inquest debate` and `inquest
// motions`. The assembled set is a rendering of those sources: a seat reading it audits the
// rendering in place of the record, and a lens or blue reading it reads the debate transcript and
// the motions, which the `inquest` group keeps from the parties being judged. A rendering defect is a defect of
// the assembler's code, found by its tests, and not a finding for a seat to file. Blue's `ingest`
// and `edit` name the composed sections in their own refusal (report.RefusedHeadings).
func newShowAssembled() *cobra.Command {
	c := &cobra.Command{
		Use:   "assembled",
		Short: "read the report set as assembly ships it (operator; every section in shipping order, who composes it, from what, and how content reaches it)",
		Long: "ops assembled assembles the report set from the record as it stands and lists every section in shipping order: its document, " +
			"its heading or the line it ships as, who composes it and from what, how content reaches it, and whether the set carries it as the record stands. " +
			"The bare form lists the sections WITHOUT their text. " +
			"--json gives each section's `text` as assembly composes it — the record's text as written, unescaped, with anchors resolved and notes numbered: pipe it to `jq` and select on `heading`. " +
			"--quote selects sections, printed in the --json form with or without --json. " +
			"The shipped .md differs from that text in three ways: every line ends in a `\\n`; every `<` outside code but a computation's anchor is written `&lt;`, which a markdown reader prints as `<` — so a `<url>` autolink or a `[text](<url>)` destination in the text ships as text, not a link, and a `<` straight after a bare URL prints on GitHub as `&lt;`, inside the link GitHub makes of the URL; and a `<` at the head of a line inside a code span that runs across lines is written `&lt;` too, which the span prints as `&lt;` — at such a `<` GitHub's reader can open an HTML block before it reads the span.\n\n" +
			"It is yours, not a seat's: a seat reads each composed section's content through the view that serves its source.\n\n" +
			"OUTPUT WITH --json: " + jsonshape.Tree(report.LayoutJSON{}) +
			"\n\nEVERY SECTION CARRIES `text`, a string: a section that does not ship carries `\"\"` and `no_text`, which says why it has none.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			run, err := seat.Of(cmd).Run()
			if err != nil {
				return err
			}
			layout, err := report.Layout(run)
			if err != nil {
				return err
			}
			sel, err := seat.SelectorOf(cmd)
			if err != nil {
				return err
			}
			// A SELECTION OF SECTIONS IS THE --json FORM'S, with or without --json: the bare form
			// lists the sections without their text, so it holds nothing a selector could match.
			if asJSON, _ := cmd.Flags().GetBool(flags.JSON); asJSON || sel.Active() {
				out, err := json.MarshalIndent(layout, "", "  ")
				if err != nil {
					return err
				}
				return seat.WriteSelected(cmd, append(out, '\n'))
			}
			if _, werr := fmt.Fprint(cmd.OutOrStdout(), layout.Markdown()); werr != nil {
				return werr
			}
			_, werr := fmt.Fprintln(cmd.ErrOrStderr(), report.LayoutHasNoText)
			return werr
		},
	}
	seat.AddSelectorFlags(c, "sections")
	return c
}
