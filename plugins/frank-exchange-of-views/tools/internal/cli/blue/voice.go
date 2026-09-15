package blue

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// spanVoiceTells runs the PRESENCE advisory over the short spans a verb is about to record, and
// renders each hit the one way every advising verb renders it (reportvoice.Found.String).
//
// EVERY SPAN-SIZED WRITER — `blue edit`'s --new, a line of inquiry, a move's reason, a line's
// method, a proof's note, a citation's title. Each of these reaches report.md: edit through the
// diff-stack, the line-of-inquiry fields through `inquiries()`, which composes Research areas,
// Future research directions and Alternatives considered, and the proof note and source title
// through its note and the Bibliography. Blue is advised on the WHOLE list, the tells red's mint
// refuses included: blue's writes are never refused, and a seat id in blue's prose is the same leak.
//
// Presence (`Find`) is the right shape here and the wrong one for a whole document: a span is where
// "this carries a lane tag" is the entire signal. `blue ingest` hands over a document, and uses the
// census instead (#873).
//
// A verb whose text the report prints OUTSIDE the body — the proof note (a [^PN] footnote and its
// evidence.md heading) and the source title (its Bibliography entry) — renders the advice with
// reportvoice.Note, naming where it lands. The 2026-09-11 is-91-prime run's lane drafts and frozen
// base carried no tell, and its report.md carried four, every one in a proof footnote: the census
// that exempted the Bibliography counted it as the tool's text, and the note inside each entry is
// the seat's.
func spanVoiceTells(spans ...string) []string {
	var out []string
	for _, s := range spans {
		if strings.TrimSpace(s) == "" {
			continue
		}
		for _, f := range reportvoice.Find(s) {
			out = append(out, f.String())
		}
	}
	return out
}
