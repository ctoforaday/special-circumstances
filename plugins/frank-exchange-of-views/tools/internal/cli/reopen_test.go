package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// AN EDIT THAT MOVES CITED TEXT REOPENS THE CITATION, END TO END.
//
// The anchor is never lost — that promise is enforced by AnchorsTransitUnchanged and backstopped
// by droppedMarker, and it holds. This is the half that was missing: an anchor SURVIVING onto
// rewritten prose is a citation backing a sentence nobody read, and until now nothing said so.
// Measured before the fix: a cite placed on "The sky is blue and the grass is green" followed the
// text to "The sky is green and the grass is on fire", silently.
func TestAnEditThatMovesCitedTextReopensTheCitation(t *testing.T) {
	runDir := t.TempDir()
	for _, s := range []struct{ role, id string }{{"lens", "red-lens-r1-L1"}, {"blue", "blue-respond-r1"}} {
		if _, err := run(t, s.role, "register", "--run", runDir, "--seat-id", s.id); err != nil {
			t.Fatal(err)
		}
	}
	seedBlueReport(t, runDir)
	const claim = "§2 the finding prose lands in a quoted sentence."
	if _, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--url", "https://example.org/red", "--title", "T", "--quote", claim,
		"--as", "supports", "--confidence", "high", "--reason", "read it at the leaf"); err != nil {
		t.Fatal(err)
	}

	evidence := func() record.EvidenceJSON {
		b, err := record.BoardState(runDir)
		if err != nil {
			t.Fatal(err)
		}
		return record.EvidenceJSONOf(b)
	}
	if got := evidence().Reopened; len(got) != 0 {
		t.Fatalf("reopened = %v before any edit, want none", got)
	}

	// AN EDIT ELSEWHERE does not reopen it. Reopening on every edit is the same as reopening on
	// none — a reader learns to skip the field.
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", "the parser accepts an empty body in this line.",
		"--new", "the parser rejects an empty body in this line.", "--reason", "unrelated"); err != nil {
		t.Fatal(err)
	}
	if got := evidence().Reopened; len(got) != 0 {
		t.Errorf("reopened = %v after an edit elsewhere in the document, want none", got)
	}

	// THE EDIT THAT MOVES THE CITED SENTENCE reopens it.
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--quote", claim, "--new", "§2 an entirely different assertion now.",
		"--reason", "rewrote the sentence red corroborated"); err != nil {
		t.Fatal(err)
	}
	got := evidence().Reopened
	if len(got) != 1 || !strings.HasPrefix(got[0], "c-") {
		t.Fatalf("reopened = %v, want the corroboration's citation — the footnote now backs a sentence nobody read", got)
	}
	// The anchor is still THERE: reopening is not losing, and the no-loss promise still holds.
	md, err := record.ReadBlueReport(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(md), got[0]) {
		t.Error("the anchor was lost rather than reopened — an anchor may transit an edit but never be dropped by one")
	}
}
