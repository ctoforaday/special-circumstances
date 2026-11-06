package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE SENTENCE LINE COMPARES IN THE FORM A CHECK RECORDS ITS SENTENCE, for every kind's check. A
// token in the live sentence — a gap placed at its end after the check — is no change of wording; a
// check that recorded no sentence (a migrated reproduce, a verify with no claim) prints no line,
// whatever the sentence now; a check of other words prints them.
func TestTheSentenceLineComparesTheRecordedForm(t *testing.T) {
	const judged = "Alpha holds."
	kinds := map[string]func(claim string) []evidenceCheck{
		"citation": func(claim string) []evidenceCheck {
			return readingChecks([]record.EvidenceVerificationJSON{{SeatID: "red-lens-evidence", Outcome: "supports", Claim: claim}})
		},
		"computation": func(claim string) []evidenceCheck {
			return judgementChecks([]record.EvidenceReproductionJSON{{SeatID: "red-lens-computation", Soundness: "sound", Claim: claim}})
		},
	}
	for kind, checks := range kinds {
		for _, c := range []struct {
			name, claim, live, want string
		}{
			{"a gap token at the sentence's end", judged, "Alpha  holds.<!--gap:G-00000001-->", ""},
			{"no recorded sentence", "", "Alpha still holds.", ""},
			{"an earlier wording", judged, "Alpha still holds.", "  - of the sentence: " + judged + "\n"},
		} {
			got := checkLines(checks(c.claim), c.live)
			line := ""
			if _, after, ok := strings.Cut(got, "\n"); ok {
				line = after
			}
			if line != c.want {
				t.Errorf("%s, %s: the sentence line is %q, want %q:\n%s", kind, c.name, line, c.want, got)
			}
		}
	}
}
