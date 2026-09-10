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
// reason, a line's method. Each of these reaches report.md: edit through the diff-stack, the
// line-of-inquiry fields through `inquiries()`, which composes Research areas, Future research
// directions and Alternatives considered. A second hand-written copy of this line in each verb
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
