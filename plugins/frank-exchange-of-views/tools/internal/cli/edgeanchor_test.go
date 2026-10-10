package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN ANCHOR RUN AT THE END OF BLUE'S QUOTE STAYS WHERE IT STANDS, through the verbs themselves.
//
// The shape is smoke m18's: a sentence holding a citation and a gap before its terminator, blue
// quoting and rewriting its prose alone. The edit applies, the run's bytes stand after the new text
// and before the terminator, the event records the replacement as written and every anchor of the
// run as reopened — and a lens prescribing a fix on that sentence, its quote stopping at the same
// place, mints.
func TestAnAnchorRunAtTheQuotesEndStaysWhereItStands(t *testing.T) {
	for _, term := range []string{"", "."} {
		t.Run("edit"+term, func(t *testing.T) {
			runDir := acceptRun(t)
			cite := citeSentence(t, runDir, acceptS+".", "https://costs/1")
			gap, err := mintQuote(t, runDir, "G", acceptS+".")
			if err != nil {
				t.Fatal(err)
			}
			run := anchor.Token(gap) + anchor.Token(cite)
			if rep := readReport(t, runDir); !strings.Contains(rep, acceptS+run+".") {
				t.Fatalf("the fixture does not hold the run before the terminator:\n%s", rep)
			}
			if err := acceptEdit(t, runDir, "E", gap, "--quote", acceptS+term, "--new", acceptN+term); err != nil {
				t.Fatalf("an edit whose quote stops before the run was refused: %v", err)
			}
			if rep := readReport(t, runDir); !strings.Contains(rep, "The year opened badly. "+acceptN+run+". Volume recovered later.") {
				t.Errorf("the run does not stand after the replacement, before the terminator:\n%s", rep)
			}
			ed := lastBody(t, runDir, &recordpb.BlueEdit{})
			if ed.GetOld() != acceptS+term || ed.GetNew() != acceptN+term {
				t.Errorf("the event records %q → %q, want the pair as blue wrote it", ed.GetOld(), ed.GetNew())
			}
			for _, id := range []string{gap, cite} {
				if !slices.Contains(ed.GetReopened(), id) {
					t.Errorf("reopened = %v, want %s: the sentence under it changed", ed.GetReopened(), id)
				}
			}
		})

		t.Run("mint"+term, func(t *testing.T) {
			runDir := acceptRun(t)
			cite := citeSentence(t, runDir, acceptS+".", "https://costs/1")
			gap, err := mintQuote(t, runDir, "G", acceptS+term, "--new", acceptN+term)
			if err != nil {
				t.Fatalf("a prescription whose quote stops before the citation on its sentence was refused: %v", err)
			}
			if rep := readReport(t, runDir); !strings.Contains(rep, acceptS+anchor.Token(gap)+anchor.Token(cite)+".") {
				t.Errorf("the gap's anchor is not at the quote's end, before the citation already there:\n%s", rep)
			}
		})
	}
}
