package cli

import (
	"encoding/json"
	"math"
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

	// THE GO FOLDS OVER THE TWO REMOVING ACTS: each reads a pure cut's removed_anchors as it reads a
	// retire's anchors.
	t.Run("fold/countsat", func(t *testing.T) {
		runDir := citedRun(t, "https://grass/1")
		citeSentence(t, runDir, "The grass is green.", "https://grass/1")
		if _, err := cut(t, runDir, "The grass is green."); err != nil {
			t.Fatal(err)
		}
		fam, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		if c := fam.CountsAt(math.MaxInt64); c.Citations != 1 || c.CitationsRetired != 1 {
			t.Errorf("CountsAt after the cut of the citation's sentence = %+v, want 1 citation and 1 out", c)
		}
	})
	t.Run("fold/answersretired", func(t *testing.T) {
		runDir, g, _ := raiserBoard(t)
		c := placeAnswer(t, runDir, "citation", "The sky is blue.", "https://water/1", g)
		lensRereads(t, runDir)
		if err := lensClose(t, runDir, lensSeat, g, "repaired"); err != nil {
			t.Fatalf("the lens's close of %s on its answer %s: %v", g, c, err)
		}
		if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(c)+"."); err != nil {
			t.Fatalf("the cut of the answer's sentence: %v", err)
		}
		if items := lensItems(t, runDir); !anyHas(items, "the answers to gap "+g+", which you closed") || !anyHas(items, c) {
			t.Errorf("no news that %s's answer %s left with its sentence:\n  %s", g, c, strings.Join(items, "\n  "))
		}
	})
	t.Run("fold/evidenceview retiredBy", func(t *testing.T) {
		runDir := citedRun(t, "https://grass/1")
		c := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
		if _, err := cut(t, runDir, "The grass is green"+anchor.Token(c)+"."); err != nil {
			t.Fatal(err)
		}
		out, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--anchor", c)
		if err != nil {
			t.Fatal(err)
		}
		var at struct {
			Retired *record.EvidenceRetiredJSON `json:"retired"`
		}
		if err := json.Unmarshal([]byte(out), &at); err != nil {
			t.Fatalf("show evidence --anchor %s did not parse: %v\n%s", c, err, out)
		}
		if at.Retired == nil || at.Retired.By != record.RemovedByEdit || at.Retired.Reason != "the claim cannot stand" {
			t.Errorf("%s's removal as show evidence states it = %+v, want the cut with its reason", c, at.Retired)
		}
	})
	t.Run("fold/anchorchecks", func(t *testing.T) {
		runDir := citedRun(t, "https://grass/1")
		c := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
		if _, err := cut(t, runDir, "The grass is green"+anchor.Token(c)+"."); err != nil {
			t.Fatal(err)
		}
		m, err := record.MergedEvents(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		if !record.RetiredAnchors(m.Events, m.At)[c] {
			t.Errorf("the ledger does not hold %s out of the report after the cut of its sentence", c)
		}
	})
	// A CUT REACHES THE GAPS ABOUT WHAT IT TOOK (WindowIndex.Reaches, gapActsOf): the lens of a gap about
	// the cut citation holds a turn on it, and blue's list owes no evidence for it (its demand is met).
	t.Run("fold/gapacts Reaches", func(t *testing.T) {
		runDir, _, _ := raiserBoard(t)
		c := citeSentence(t, runDir, "The sky is blue.", "https://water/1")
		g := materialGapAbout(t, runDir, c)
		if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(c)+"."); err != nil {
			t.Fatal(err)
		}
		if !lensEngagedOn(t, runDir, g) {
			t.Errorf("the cut of %s, which %s is about, gave its lens no turn on it", c, g)
		}
	})
	// The cut meets the gap's demand (met arm 2), so blue's list owes no evidence for it: no sitting arm
	// is needed to excuse it.
	t.Run("fold/awaiting evidence met by the cut", func(t *testing.T) {
		runDir, _, _ := raiserBoard(t)
		c := citeSentence(t, runDir, "The sky is blue.", "https://water/1")
		g := materialGapAbout(t, runDir, c)
		dispatchBlueOnto(t, runDir, g)
		registerBlue(t, runDir)
		if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(c)+"."); err != nil {
			t.Fatal(err)
		}
		out, err := run(t, "show", "work", "--run", runDir, "--seat-id", blueSeat)
		if err != nil {
			t.Fatal(err)
		}
		var w record.WorkJSON
		if err := json.Unmarshal([]byte(out), &w); err != nil {
			t.Fatal(err)
		}
		for _, it := range w.Sitting.Open {
			if it.Kind == record.ItemAwaitingEvidence && strings.Contains(it.What, g) {
				t.Errorf("blue's list owes evidence for %s after its cut took the anchor %s is about: %q", g, g, it.What)
			}
		}
	})
}

// materialGapAbout is a material gap of lensSeat's about the anchor c, of check kind kind ("source"
// where none is given).
func materialGapAbout(t *testing.T, runDir, c string, kind ...string) string {
	t.Helper()
	k := "source"
	if len(kind) > 0 {
		k = kind[0]
	}
	out, err := runMint(t, "--run", runDir, "--seat-id", lensSeat, "--key", "A", "--class", "spec_underspecification",
		"--problem", "the source does not say it", "--fix", "cite one that does", "--check-kind", k,
		"--check", "a source that says it", "--severity", "high", "--likelihood", "high", "--impact", "high",
		"--about-kind", "anchor", "--about", c)
	if err != nil {
		t.Fatalf("mint --about %s: %v", c, err)
	}
	return gapID(out)
}
