package migrate_test

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// AN EDIT REOPENS THE SAME ANCHORS WHETHER EACH ANCHOR'S SENTENCE IS READ OFF ONE SPLIT OF THE
// REPORT OR OFF A SPLIT OF ITS OWN. bluedoc.ReopenedAnchors splits each document once and reads
// every anchor's sentence off that split; reopenedSplitPerAnchor below splits the document again
// for every anchor, which is what it did before. Over an archived run's report, every anchor it
// holds is read on each side of four edits: none, every sentence's words changed, every other
// anchored sentence changed, and every anchor stripped bare of its sentence.
func TestEveryArchivedAnchorReopensAsWhenItsSentenceWasSplitAlone(t *testing.T) {
	for _, name := range []string{"2026-10-10_is-91-prime-m19.tar.gz", "2026-09-11_is-91-prime-b9.tar.gz", "2026-10-13_is-91-prime-m23.tar.gz"} {
		t.Run(name, func(t *testing.T) {
			_, run := migrateArchive(t, name)
			report, err := reportproj.RenderFromRecord(run)
			if err != nil {
				t.Fatal(err)
			}
			ids := claimcount.ProtectedAnchorIDs(report)
			if len(ids) < 10 {
				t.Fatalf("the report holds %d anchor(s): too few for the comparison to read anything", len(ids))
			}
			word := regexp.MustCompile(`\bthe\b`)
			everyOther, n := report, 0
			for _, sp := range slices.Backward(anchor.Sentences(report)) {
				if s := report[sp[0]:sp[1]]; len(anchor.IDs(s)) > 0 {
					if n++; n%2 == 0 {
						everyOther = everyOther[:sp[0]] + "Indeed, " + everyOther[sp[0]:]
					}
				}
			}
			bare := report
			for _, id := range ids {
				tok := anchor.Token(id)
				if i := strings.Index(bare, tok); i >= 0 {
					bare = bare[:i] + "\n\n" + tok + "\n\n" + bare[i+len(tok):]
				}
			}
			for _, c := range []struct {
				edit, after string
				some        bool
			}{
				{"none", report, false},
				{"every sentence's words", word.ReplaceAllString(report, "the very"), true},
				{"every other anchored sentence", everyOther, true},
				{"every anchor stripped bare", bare, false},
			} {
				got, want := bluedoc.ReopenedAnchors(report, c.after), reopenedSplitPerAnchor(report, c.after)
				if !slices.Equal(got, want) {
					t.Errorf("%s: ReopenedAnchors = %d anchor(s), the split per anchor %d:\n  got  %v\n  want %v", c.edit, len(got), len(want), got, want)
				}
				if c.some && len(want) == 0 {
					t.Errorf("%s: no anchor reopened on either side — the edit changed no anchored sentence, so it compares nothing", c.edit)
				}
				t.Logf("%s: %d of %d anchor(s) reopened", c.edit, len(want), len(ids))
			}
		})
	}
}

// reopenedSplitPerAnchor is ReopenedAnchors with each anchor's sentence read off a split of its own.
func reopenedSplitPerAnchor(before, after string) []string {
	bare := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(after) {
		bare[id] = true
	}
	var out []string
	for _, id := range claimcount.ProtectedAnchorIDs(before) {
		b, okB := sentenceSplitAlone(before, anchor.Token(id))
		a, okA := sentenceSplitAlone(after, anchor.Token(id))
		if !okB || !okA || bare[id] {
			continue
		}
		if b != a {
			out = append(out, id)
		}
	}
	return out
}

func sentenceSplitAlone(doc, tok string) (string, bool) {
	i := strings.Index(doc, tok)
	for _, sp := range anchor.Sentences(doc) {
		if sp[0] <= i && i+len(tok) <= sp[1] {
			return strings.Join(strings.Fields(claimcount.StripAnchors(doc[sp[0]:sp[1]])), " "), true
		}
	}
	return "", false
}
