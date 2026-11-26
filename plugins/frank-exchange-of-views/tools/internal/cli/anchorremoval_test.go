package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// ONE VIEW OF AN ANCHOR LEAVING THE REPORT (P12b). A pure cut and a retire are the two acts that take
// an anchor out, and every reader of "this anchor is out of the report" reads both: the SQL readers
// through anchor_removal, the Go folds through the edit's removed_anchors beside the retire's anchors,
// and the readers of `reopened` through a union with the edit arm. One subtest per reader, each
// driving a pure cut through the command tree.
func TestAnchorRemovalIsOneView(t *testing.T) {
	// THE `reopened` READERS THAT UNION THE EDIT ARM.
	t.Run("reopened/lensTurnActsOf", func(t *testing.T) {
		runDir, g, h := raiserBoard(t)
		cutAnswering(t, runDir, g, h)
		if !lensEngagedOn(t, runDir, g) {
			t.Errorf("the cut of %s's marker gave its minting lens no turn on it", g)
		}
	})
	t.Run("reopened/changedSinceSQL", func(t *testing.T) {
		runDir, g, h := raiserBoard(t)
		cutAnswering(t, runDir, g, h)
		if err := lensClose(t, runDir, lensSeat, g, "moot"); err == nil || !strings.Contains(err.Error(), "the text around "+g+" changed after your delivery was cut") {
			t.Errorf("a close on a delivery that predates the cut of %s was not refused by the delivered-text rule: %v", g, err)
		}
	})
	t.Run("reopened/reopeningEditsAt", func(t *testing.T) {
		runDir, g, h := raiserBoard(t)
		cutAnswering(t, runDir, g, h)
		out, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
		if err != nil {
			t.Fatal(err)
		}
		var w record.WorkJSON
		if err := json.Unmarshal([]byte(out), &w); err != nil {
			t.Fatal(err)
		}
		for _, row := range w.Open {
			if row.ID == g {
				if len(row.EditedSince) != 1 || !strings.Contains(row.EditedSince[0].Old, "The grass is green"+anchor.Token(g)) {
					t.Errorf("%s's edited_since after the cut of its marker = %+v, want the cut once", g, row.EditedSince)
				}
				return
			}
		}
		t.Errorf("%s is not on the lens's work list after the cut", g)
	})
	t.Run("reopened/estoppingPrescriptions", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n\nAnother sentence stays.\n")
		g := mintWithProposal(t, runDir, "G1", "Five independent verification approaches agree", prescribedText)
		h := mintWithProposal(t, runDir, "H1", "Another sentence stays", "Another sentence stays, as written.")
		registerBlue(t, runDir)
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "A1", "--answers", g, "--accept",
			"--reason", "red's wording"); err != nil {
			t.Fatalf("accept: %v", err)
		}
		estops := func() bool {
			t.Helper()
			fam, err := record.FamilyOf(runtest.Open(t, runDir))
			if err != nil {
				t.Fatal(err)
			}
			_, ok := record.EstoppelConflict(fam, prescribedText)
			return ok
		}
		if !estops() {
			t.Fatalf("the verbatim application of %s's prescription does not estop red", g)
		}
		applied := strings.Replace(prescribedText, "primality.", "primality"+anchor.Token(g)+".", 1)
		if _, err := cut(t, runDir, applied, "--answers", h); err != nil {
			t.Fatalf("the pure cut of the applied sentence: %v\n%s", err, readReport(t, runDir))
		}
		if estops() {
			t.Errorf("%s's prescription still estops red after a pure cut took its sentence and marker", g)
		}
	})
}
