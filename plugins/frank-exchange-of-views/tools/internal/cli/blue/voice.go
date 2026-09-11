package blue

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// spanVoiceTells runs the PRESENCE advisory over the short spans a verb is about to record, and
// renders each hit the one way every advising verb renders it.
//
// ONE FORMATTER FOR EVERY SPAN-SIZED WRITER — `blue edit`'s --new, a line of inquiry, a move's
// reason, a line's method, a proof's note, a citation's title. Each of these reaches report.md:
// edit through the diff-stack, the line-of-inquiry fields through `inquiries()`, which composes
// Research areas, Future research directions and Alternatives considered, and the proof note and
// source title through the Bibliography. A second hand-written copy of this line in each verb
// would drift the first time one was reworded, and a seat hearing the same advice two ways learns
// that the wording is not load-bearing.
//
// Presence (`Find`) is the right shape here and the wrong one for a whole document: a span is where
// "this carries a lane tag" is the entire signal. `blue ingest` hands over a document, and uses the
// census instead (#873).
func spanVoiceTells(spans ...string) []string {
	var out []string
	for _, s := range spans {
		if strings.TrimSpace(s) == "" {
			continue
		}
		for _, f := range reportvoice.Find(s) {
			out = append(out, fmt.Sprintf("%q reads as %s — %s", f.Match, f.Class, f.Redirect))
		}
	}
	return out
}

// voiceNote renders the advisory for a verb whose text the report prints OUTSIDE the body — the
// proof note (`blue prove --reason`, a [^PN] footnote and its evidence.md heading) and the source
// title (`blue cite --title`, its Bibliography entry). The 2026-09-11 is-91-prime run's lane drafts
// and frozen base carried no tell, and its report.md carried four, every one in a proof footnote:
// the census that exempted the Bibliography counted it as the tool's text, and the note inside each
// entry is the seat's. `where` names the place it lands, because a seat told only "the report" looks
// for it in the body and does not find it.
//
// Empty when nothing matched: a clean act says nothing extra.
func voiceNote(where string, tells []string) string {
	if len(tells) == 0 {
		return ""
	}
	return "\n\nNOTE — this text is printed in the report (" + where + ") and in places sounds\n" +
		"like the run rather than the subject. It is recorded; this is not a refusal, and it may be wrong:\n  - " +
		strings.Join(tells, "\n  - ") +
		"\n\nSEPARATION, NEVER DELETION: what the evidence establishes stays, re-voiced in the\n" +
		"subject's terms; only the fact about the run goes. Red's voice lens holds that\n" +
		"judgement — these are only the literal tells."
}
