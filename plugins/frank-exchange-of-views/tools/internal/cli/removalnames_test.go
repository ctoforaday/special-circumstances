package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY SEAT-FACING STRING BUILT FROM A REMOVAL NAMES ITS ACT (lead call 11-26, plan-audit r6 gap 1).
// A retire keeps its sentence and a cut takes it, so a string calling every removal a retire tells a
// seat a cut sentence still stands. One subtest per site of the plan's wording table; each drives a
// pure cut of a cited sentence, and, where the site has a retire arm, an anchor-only retire.
func TestARemovalNamesItsAct(t *testing.T) {
	// removed is a board with the citation C on "The sky is blue.", taken out by a cut (cut true) or an
	// anchor-only retire.
	removed := func(t *testing.T, byCut bool, before func(runDir, c string)) (runDir, c string) {
		t.Helper()
		runDir, _, _ = raiserBoard(t)
		c = citeSentence(t, runDir, "The sky is blue.", "https://water/1")
		if before != nil {
			before(runDir, c)
		}
		if byCut {
			if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(c)+"."); err != nil {
				t.Fatalf("the cut of %s's sentence: %v", c, err)
			}
		} else {
			retireAnchor(t, runDir, c)
		}
		return runDir, c
	}
	const cutSaid, retireSaid = "cut its sentence, and the anchor left with it", "blue retired it at event"
	acts := []struct {
		name  string
		byCut bool
	}{{"cut", true}, {"retire", false}}
	for _, a := range acts {
		t.Run("1-2 compute --cites/"+a.name, func(t *testing.T) {
			runDir, c := removed(t, a.byCut, nil)
			s := script(t, runDir, "k.js", "console.log('k');")
			_, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "Water is wet.", "--dir", s,
				"--reason", "the method applied", "--cites", c, "--", "node", "k.js")
			want := retireSaid
			if a.byCut {
				want = cutSaid
			}
			if err == nil || !strings.Contains(err.Error(), "names a citation that is out of the report") || !strings.Contains(err.Error(), want) {
				t.Errorf("compute --cites %s after its %s: %v, want the removal named (%q)", c, a.name, err, want)
			}
		})
		t.Run("3 mint --about/"+a.name, func(t *testing.T) {
			runDir, c := removed(t, a.byCut, nil)
			_, err := runMint(t, "--run", runDir, "--seat-id", lensSeat, "--key", "Z", "--class", "spec_underspecification",
				"--problem", "p", "--fix", "f", "--check-kind", "document", "--check", "c", "--severity", "high", "--likelihood", "high",
				"--impact", "high", "--about-kind", "anchor", "--about", c)
			want := "A defect still in the sentence it backed"
			if a.byCut {
				want = "Its sentence left with it"
			}
			if err == nil || !strings.Contains(err.Error(), "which is out of the report") || !strings.Contains(err.Error(), want) {
				t.Errorf("mint --about %s after its %s: %v, want the advice %q", c, a.name, err, want)
			}
		})
		t.Run("5 show report --anchor/"+a.name, func(t *testing.T) {
			runDir, c := removed(t, a.byCut, nil)
			_, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--anchor", c)
			want := "because blue retired it"
			if a.byCut {
				want = "because blue's edit cut its sentence"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("show report --anchor %s after its %s: %v, want %q", c, a.name, err, want)
			}
		})
		t.Run("6 the metAt item/"+a.name, func(t *testing.T) {
			var g string
			runDir, c := removed(t, a.byCut, func(runDir, c string) { g = materialGapAbout(t, runDir, c) })
			want := "gap " + g + ": blue retired " + c + ", which " + g + " is about"
			if a.byCut {
				want = "gap " + g + ": blue's edit cut the sentence holding " + c + ", which " + g + " is about"
			}
			if items := lensItems(t, runDir); !anyHas(items, want) {
				t.Errorf("no item says %q:\n  %s", want, strings.Join(items, "\n  "))
			}
		})
		t.Run("7-8 the stale line/"+a.name, func(t *testing.T) {
			var g string
			runDir, c := removed(t, a.byCut, func(runDir, c string) {
				g = materialGapAbout(t, runDir, c)
				lensRereads(t, runDir) // the sitting the close is made in opens before the removal
			})
			// A REGRADE, not a close: the stale guard (record.Fresh) is a regrade's, a finding's and a
			// motion's; a close stands on the delivered-text rule instead.
			_, err := run(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--severity", "medium",
				"--reason", "the defect stands")
			want := "blue's retire of " + c + ", which " + g + " is about"
			if a.byCut {
				want = "blue's cut of " + c + "'s sentence, which " + g + " is about"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("the regrade of %s in a sitting opened before the %s: %v, want %q", g, a.name, err, want)
			}
		})
	}
	t.Run("4 a check of a cut citation", func(t *testing.T) {
		runDir, c := removed(t, true, nil)
		_, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", c, "--as", "supports", "--confidence", "high", "--reason", "read it")
		if err == nil || !strings.Contains(err.Error(), "is out of the report") || !strings.Contains(err.Error(), cutSaid) {
			t.Errorf("a check of %s after its cut: %v, want the cut named", c, err)
		}
	})
	t.Run("9 the answers news", func(t *testing.T) {
		runDir, g, _ := raiserBoard(t)
		c := placeAnswer(t, runDir, "citation", "The sky is blue.", "https://water/1", g)
		lensRereads(t, runDir)
		if err := lensClose(t, runDir, lensSeat, g, "repaired"); err != nil {
			t.Fatal(err)
		}
		if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(c)+"."); err != nil {
			t.Fatal(err)
		}
		if items := lensItems(t, runDir); !anyHas(items, "the answers to gap "+g+", which you closed, have all left the report, the last by blue's cut at event") {
			t.Errorf("the news does not name the cut:\n  %s", strings.Join(items, "\n  "))
		}
	})
	t.Run("10 show evidence", func(t *testing.T) {
		for _, a := range acts {
			runDir, c := removed(t, a.byCut, nil)
			out, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--anchor", c)
			if err != nil {
				t.Fatal(err)
			}
			var at struct {
				Retired *record.EvidenceRetiredJSON `json:"retired"`
			}
			if err := json.Unmarshal([]byte(out), &at); err != nil {
				t.Fatal(err)
			}
			want := record.RemovedByRetire
			if a.byCut {
				want = record.RemovedByEdit
			}
			if at.Retired == nil || at.Retired.By != want {
				t.Errorf("show evidence --anchor %s after its %s: %+v, want by %q", c, a.name, at.Retired, want)
			}
		}
	})
}
