package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// editHistory renders every edit blue recorded to the report, in record order, grouped by the
// sitting that made it: the gap it answers, blue's reason, and the span it replaced beside the text
// that replaced it. "" when the record holds no edit.
//
// THE CHANGELOG IS THE REPORT'S PROVENANCE, and it listed only what left the report. Blue's reason
// for an edit was read by `show changes` alone: across three archived runs, 62 reasons and none in
// any document (the reach audit, 2026-10-20). A superseded wording of a sentence the report still
// carries was nowhere either — the report prints the latest.
//
// THE SITTING IS THE ONE THE RECORD HOLDS THE EDIT IN (fam.At), as the retirements above it name.
func editHistory(fam record.Family) string {
	var b strings.Builder
	group, prevReason, n := "", "", 0
	for _, e := range fam.Events {
		ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok {
			continue
		}
		n++
		w := fam.At.Of(e)
		if g := fmt.Sprintf("### `%s` #%d (epoch %d)", e.GetSeatId(), w.Sitting, w.Epoch); g != group {
			group, prevReason = g, ""
			b.WriteString("\n" + g + "\n\n")
		}
		answers := "blue's own, answering no gap"
		if id := ed.GetAnswers(); id != "" {
			answers = "answers **" + id + "**"
		}
		if ed.GetAppliedVerbatim() {
			answers += ", applying red's proposed text verbatim"
		}
		reason := plain(ed.GetText())
		switch {
		case reason == "":
			reason = "_(no reason recorded)_"
		case reason == prevReason:
			reason = "the same reason as the edit above"
		default:
			prevReason = reason
		}
		fmt.Fprintf(&b, "- %s — %s\n\n", answers, reason)
		b.WriteString("  ```diff\n")
		if old := plain(ed.GetOld()); old != "" {
			b.WriteString("  - " + old + "\n")
		}
		if nw := plain(ed.GetNew()); nw != "" {
			b.WriteString("  + " + nw + "\n")
		}
		b.WriteString("  ```\n")
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("_Every edit blue recorded to this report (%d), in the order made, grouped by the sitting that made it (`seat #N` is that seat's Nth sitting): the gap each answers, blue's reason for it, and the span it replaced (`-`) beside the text that replaced it (`+`). The report carries the last wording; a span replaced again later is history._\n", n) +
		strings.TrimRight(b.String(), "\n")
}
