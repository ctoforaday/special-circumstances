package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// AN EDIT THAT MOVES CITED TEXT REOPENS THE CITATION, END TO END.
//
// The anchor is never lost — that promise is enforced by AnchorsTransitUnchanged and backstopped
// by droppedMarker, and it holds. This is the half that was missing: an anchor SURVIVING onto
// rewritten prose is a citation backing a sentence nobody read, and until now nothing said so.
// Measured before the fix: a cite placed on "The sky is blue and the grass is green" followed the
// text to "The sky is green and the grass is on fire", silently.
func TestAnEditThatMovesCitedTextReopensTheCitation(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, id := range []string{"red-lens-evidence", "blue-respond"} {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", id); err != nil {
			t.Fatal(err)
		}
	}
	seedBlueReport(t, runDir)
	const claim = "§2 the finding prose lands in a quoted sentence."
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/red", "--title", "T", "--quote", claim,
		"--as", "supports", "--confidence", "high", "--reason", "read it at the leaf"); err != nil {
		t.Fatal(err)
	}

	evidence := func() record.EvidenceJSON {
		b, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		return record.EvidenceJSONOf(b.Events, b.At)
	}
	if got := evidence().Reopened; len(got) != 0 {
		t.Fatalf("reopened = %v before any edit, want none", got)
	}

	// AN EDIT ELSEWHERE does not reopen it. Reopening on every edit is the same as reopening on
	// none — a reader learns to skip the field.
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", "the parser accepts an empty body in this line.",
		"--new", "the parser rejects an empty body in this line.", "--reason", "unrelated"); err != nil {
		t.Fatal(err)
	}
	if got := evidence().Reopened; len(got) != 0 {
		t.Errorf("reopened = %v after an edit elsewhere in the document, want none", got)
	}

	// THE EDIT THAT MOVES THE CITED SENTENCE reopens it — quoted AS PRINTED, token carried.
	printed, tok := anchoredLine(t, runDir)
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-respond",
		"--quote", printed, "--new", "§2 an entirely different assertion now"+tok+".",
		"--reason", "rewrote the sentence red corroborated"); err != nil {
		t.Fatal(err)
	}
	got := evidence().Reopened
	if len(got) != 1 || !strings.HasPrefix(got[0].Anchor, "C-") || got[0].Edit <= got[0].LastCheck {
		t.Fatalf("reopened = %+v, want the corroboration's citation, its edit after its check — the footnote now backs a sentence nobody read", got)
	}
	anchor := got[0].Anchor
	// The anchor is still THERE: reopening is not losing, and the no-loss promise still holds.
	md, err := reportproj.RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, anchor) {
		t.Error("the anchor was lost rather than reopened — an anchor may transit an edit but never be dropped by one")
	}

	// IT STANDS UNTIL A SITTING OPENED AFTER THE EDIT CHECKS IT: the lens sits again and reads the
	// sentence as it now stands. (Its open sitting holds one check per anchor, so the write path
	// already refuses a second one from the sitting the edit landed in.)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, "verify", "--run", runDir, "--seat-id", "red-lens-evidence", "--anchor", anchor,
		"--quote", "§2 an entirely different assertion now", "--as", "supports", "--confidence", "high", "--reason", "read again at the leaf"); err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	if got := evidence().Reopened; len(got) != 0 {
		t.Errorf("reopened = %+v after a later sitting verified it, want none", got)
	}
}

// anchoredLine returns the report line carrying a citation anchor and the token itself — what a
// seat sees in `show report` and copies into its quote.
func anchoredLine(t *testing.T, runDir string) (string, string) {
	t.Helper()
	md, err := reportproj.RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(md, "\n") {
		i := strings.Index(l, "<!--cite:")
		if i < 0 {
			continue
		}
		j := strings.Index(l[i:], "-->")
		return l, l[i : i+j+3]
	}
	t.Fatal("no anchored line in the report")
	return "", ""
}

// THE PAGE THE PASS REFUSAL SENDS A SEAT TO SAYS WHAT IT IS LOOKING AT. A `reopened` row is three
// values and two of them are bare numbers: the page says they are event ids, which one is the
// edit, and what removes the row — the lens's whole instruction for clearing it.
func TestShowEvidenceHelpSaysWhatAReopenedRowMeans(t *testing.T) {
	h := help(t, "show", "evidence", "--help", "--seat-id", "red-lens-evidence")
	for _, want := range []string{
		"A `reopened` ROW IS A CITATION OR PROOF WHOSE SENTENCE AN EDIT CHANGED and that no sitting opened since has checked",
		"`edit` and `last_check` are event ids",
		"from a sitting opened after `edit` removes the row",
		"`--as unreachable` does not",
	} {
		if !strings.Contains(h, want) {
			t.Errorf("`show evidence --help` does not say %q", want)
		}
	}
	// The shape is on the page as the fields the binary prints, at the table and at one anchor.
	for _, tree := range []string{"reopened:[{anchor,edit,last_check}]", "reopened:{anchor,edit,last_check}"} {
		if !strings.Contains(h, tree) {
			t.Errorf("`show evidence --help` does not print the shape %s", tree)
		}
	}
}
