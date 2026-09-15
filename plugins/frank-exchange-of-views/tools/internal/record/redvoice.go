package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// RED'S TEXT IN THE REPORT IS WRITTEN TO THE READER (gblock, 2026-09-11: "if notes are going into
// the report, red has to follow the rules on voice. no inside baseball"). Red's text reaches
// report.md three ways, and each is held here at the write where red authors it:
//
//   - the first sentence of a gap's `problem` and of its `required_fix` is a row of the risk matrix
//     while the gap is open;
//   - a `fix_new` prescription becomes the report's text when blue accepts it;
//   - a labelled corroboration's `title` is the source's Bibliography entry.
//
// Only the UNAMBIGUOUS tells are refused — a seat or lens id, a finding label, a gap id joined to a
// process word, a lane tag — because none of them has a reading as subject prose; the rest ("this
// run", "the red team", "epoch 3") are advice, returned by the verb (gblock's fork ruling, D9; the
// title and fix_new rulings, gblock 2026-09-15, plans/feov-lens-bar.md §III.9). Blue's edit advisory
// on accepted text stays as a second net under fix_new: the refusal here is the first.
//
// `mint_reason` and a corroboration's reason ARE NEVER CHECKED. They are red's argument to the other
// seats and stay on the record; they are where the run's part of a finding goes, so refusing process
// words there would close the one channel the refusal points to.
//
// The callers skip both checks under Migrating: an archived mint or corroboration is what a seat DID,
// and migration does not re-judge it under a rule it predates.

// reportSpan is one field of red's text that the report prints, named by the flag that sets it.
type reportSpan struct{ flag, text string }

// refusedTells lists every refused tell in the spans, one line each, naming the flag and the term.
func refusedTells(spans ...reportSpan) []string {
	var hits []string
	for _, s := range spans {
		for _, f := range reportvoice.Refused(s.text) {
			hits = append(hits, fmt.Sprintf("%s carries %q: %s", s.flag, f.Match, f.Redirect))
		}
	}
	return hits
}

// refuseMintReportVoice holds the three mint fields the report prints to the report's voice.
func refuseMintReportVoice(m *recordpb.Mint) error {
	hits := refusedTells(
		reportSpan{"--problem", m.GetProblem()},
		reportSpan{"--fix", m.GetRequiredFix()},
		reportSpan{"--new", m.GetFixNew()},
	)
	if len(hits) == 0 {
		return nil
	}
	return fmt.Errorf("record: mint refused — the first sentence of the problem and of the fix prints in the report's risk matrix while the gap is open, a --new replacement becomes the report's text when blue accepts it, and the report is written to a reader of the subject:\n  - %s\nWrite all three to that reader. The run's part — which seat, which gap, which lane — goes in --reason, beside a --problem that says what is wrong",
		strings.Join(hits, "\n  - "))
}

// refuseCorroborationTitleVoice holds a corroboration's title to the report's voice when the title
// prints: a LABELLED corroboration is a cited source, and its title is its Bibliography entry. An
// unlabelled one — a refutation, an absence, an unreachable source — is no footnote, so its title
// stays on the record and is not checked.
func refuseCorroborationTitleVoice(v *recordpb.Verify) error {
	if v.GetLabel() == "" {
		return nil
	}
	hits := refusedTells(reportSpan{"--title", v.GetTitle()})
	if len(hits) == 0 {
		return nil
	}
	return fmt.Errorf("record: corroborate refused — a corroboration that backs the claim is a footnote, its title prints as the source's Bibliography entry, and the report is written to a reader of the subject:\n  - %s\nGive the source's own title — author, work, publisher. Which seat found it, and how, goes in --reason",
		strings.Join(hits, "\n  - "))
}
