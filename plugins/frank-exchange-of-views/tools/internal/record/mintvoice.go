package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// refuseMintReportVoice holds the two mint fields the report prints to the report's voice.
//
// RED'S TEXT IN THE REPORT IS WRITTEN TO THE READER (gblock, 2026-09-11: "if notes are going into
// the report, red has to follow the rules on voice. no inside baseball"). The first sentence of a
// gap's `problem` and of its `required_fix` is a row of report.md's risk matrix while the gap is
// open. Only the UNAMBIGUOUS tells are refused — a seat or lens id, a finding label, a gap id joined
// to a process word, a lane tag — because none of them has a reading as subject prose; the rest
// ("this run", "the red team", "epoch 3") are advice, returned by the verb (gblock's fork ruling,
// plans/feov-lens-bar.md §III.9 and D9).
//
// `mint_reason` IS NEVER CHECKED. It is red's argument to the other seats and stays on the record;
// it is where the run's part of a finding goes, so refusing process words there would close the one
// channel the refusal points to.
//
// The caller skips this under Migrating: an archived mint is what a seat DID, and migration does not
// re-judge it under a rule it predates.
func refuseMintReportVoice(m *recordpb.Mint) error {
	var hits []string
	for _, field := range []struct{ flag, text string }{
		{"--problem", m.GetProblem()},
		{"--fix", m.GetRequiredFix()},
	} {
		for _, f := range reportvoice.Refused(field.text) {
			hits = append(hits, fmt.Sprintf("%s carries %q: %s", field.flag, f.Match, f.Redirect))
		}
	}
	if len(hits) == 0 {
		return nil
	}
	return fmt.Errorf("record: mint refused — the first sentence of the problem and of the fix prints in the report's risk matrix while the gap is open, and the report is written to a reader of the subject:\n  - %s\nWrite both to that reader. The run's part — which seat, which gap, which lane — goes in --reason, beside a --problem that says what is wrong",
		strings.Join(hits, "\n  - "))
}
