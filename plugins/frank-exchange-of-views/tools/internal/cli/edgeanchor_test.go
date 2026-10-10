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
// place, mints, with or without the terminator: the pair an acceptance sends carries the run.
//
// A REPLACEMENT THAT ENDS ITS OWN SENTENCE IS REFUSED while the run stays put (gblock, 2026-10-10;
// universe-m19's ev358, ev434, ev435, ev436): the refusal names every anchor of the run, the report
// and the record are untouched, and the same replacement applies once the quote carries the run.
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
			before := readReport(t, runDir)
			err = acceptEdit(t, runDir, "E", gap, "--quote", acceptS+term, "--new", acceptN+term)
			if term != "" {
				for _, want := range []string{anchor.Label(gap), anchor.Label(cite), "ends its own sentence", "leave its closing punctuation off --new", "quote it as `show report` prints it"} {
					if err == nil || !strings.Contains(err.Error(), want) {
						t.Fatalf("a replacement ending its own sentence before a run that stays put: got %v, want a refusal holding %q", err, want)
					}
				}
				if before != readReport(t, runDir) {
					t.Fatal("the refused edit reached the record: the report it renders moved")
				}
				if err := acceptEdit(t, runDir, "E", gap, "--quote", acceptS+run+term, "--new", acceptN+run+term); err != nil {
					t.Fatalf("the same replacement, its quote carrying the run, was refused: %v", err)
				}
				if rep := readReport(t, runDir); !strings.Contains(rep, "The year opened badly. "+acceptN+run+". Volume recovered later.") {
					t.Errorf("the carried run does not stand before the terminator:\n%s", rep)
				}
				return
			}
			if err != nil {
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
