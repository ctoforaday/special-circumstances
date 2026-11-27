package cli

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// raiserBoard is a cited run with the chair and the minting lens seated and two open gaps of the
// lens's: G, marked on "The grass is green.", and H, marked on "Water is wet.". It returns both ids.
func raiserBoard(t *testing.T, extra ...string) (runDir, g, h string) {
	t.Helper()
	runDir = newRun(t)
	seedCast(t, runDir)
	writeReport(t, runDir, retireBase)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://water/1": []byte("<html>a source</html>")}})
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	mint := func(key, quote string, more ...string) string {
		t.Helper()
		out, err := run(t, append([]string{"mint", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", quote,
			"--class", "spec_underspecification", "--problem", "the claim overreaches", "--fix", "qualify it",
			"--check", "the sentence states its limits", "--severity", "high", "--likelihood", "high", "--impact", "high",
			"--complexity", "low"}, more...)...)
		if err != nil {
			t.Fatalf("mint %s: %v", key, err)
		}
		return gapID(out)
	}
	kind := []string{"--check-kind", "document"}
	if len(extra) > 0 {
		kind = extra
	}
	g = mint("G", "The grass is green.", kind...)
	h = mint("H", "Water is wet.", "--check-kind", "document")
	return runDir, g, h
}

// cutAnswering is blue's pure cut of "The grass is green." with its marker, answering gap ("" answers none).
func cutAnswering(t *testing.T, runDir, g, answers string) {
	t.Helper()
	var extra []string
	if answers != "" {
		extra = []string{"--answers", answers}
	}
	if _, err := cut(t, runDir, "The grass is green"+anchor.Token(g)+".", extra...); err != nil {
		t.Fatalf("the pure cut of %s's sentence answering %q was refused: %v", g, answers, err)
	}
}

// lensItems is the What of every item on lensSeat's work list.
func lensItems(t *testing.T, runDir string) []string {
	t.Helper()
	out, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatal(err)
	}
	var w record.WorkJSON
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatalf("the lens's work list did not parse: %v", err)
	}
	var got []string
	for _, it := range w.Sitting.Open {
		got = append(got, it.What)
	}
	return got
}

func anyHas(items []string, want string) bool {
	return slices.ContainsFunc(items, func(s string) bool { return strings.Contains(s, want) })
}

// lensEngagedOn is whether the plan readies lensSeat engaged on gap g (its turn on g), read without
// writing a dispatch.
func lensEngagedOn(t *testing.T, runDir, g string) bool {
	t.Helper()
	plan, err := record.PlanDispatch(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == lensSeat && slices.Contains(p.GapIDs, g) })
}

// A CUT OF A GAP'S MARKER IS ITS RAISER'S TURN (ruling 4, F21). The record writes no Close: G stays
// open, reading gone, and its minting lens holds a turn on it, whatever gap the cut answered. The lens
// is served the cut (its stale delivery refuses a close, its next delivery passes the cut's old
// paragraph) and closes moot, or keeps the gap with a regrade.
func TestACutReadiesItsRaiser(t *testing.T) {
	const gone = "is open and its anchor is no longer in the report — judge whether the report as it stands meets your acceptance check"
	for _, by := range []string{"another gap", "no gap"} {
		answering := func(h string) string {
			if by == "another gap" {
				return h
			}
			return ""
		}
		t.Run("the turn/"+by, func(t *testing.T) {
			runDir, g, h := raiserBoard(t)
			closes := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSE)
			if lensEngagedOn(t, runDir, g) {
				t.Fatalf("the lens is engaged on %s before any act on it", g)
			}
			cutAnswering(t, runDir, g, answering(h))
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSE); n != closes {
				t.Errorf("the cut wrote %d close(s): the record writes none, the lens closes", n-closes)
			}
			if gp := boardGap(t, runDir, g); gp == nil || gp.LocationState != record.LocationGone {
				t.Errorf("%s after its marker's cut: %+v, want open and gone", g, gp)
			}
			if !lensEngagedOn(t, runDir, g) {
				t.Errorf("the cut of %s's marker gave its minting lens no turn on it", g)
			}
		})
		t.Run("the turn of a retired minter/"+by, func(t *testing.T) {
			runDir, g, h := raiserBoard(t)
			for range 3 {
				lensRereads(t, runDir)
			}
			registerChairOnce(t, runDir)
			cutAnswering(t, runDir, g, answering(h))
			if !lensEngagedOn(t, runDir, g) {
				t.Errorf("the cut of %s's marker did not ready its retired minter on it", g)
			}
		})
		t.Run("the gone item/"+by, func(t *testing.T) {
			runDir, g, h := raiserBoard(t)
			cutAnswering(t, runDir, g, answering(h))
			if items := lensItems(t, runDir); !anyHas(items, "gap "+g+" "+gone) {
				t.Errorf("no gone item on %s; the lens's list:\n  %s", g, strings.Join(items, "\n  "))
			}
		})
		t.Run("the stale close, then moot/"+by, func(t *testing.T) {
			runDir, g, h := raiserBoard(t)
			lensRereads(t, runDir) // its whole re-read, so the sitting after the cut is handed the change
			cutAnswering(t, runDir, g, answering(h))
			err := lensClose(t, runDir, lensSeat, g, "moot")
			if err == nil || !strings.Contains(err.Error(), "the text around "+g+" changed after your delivery was cut") {
				t.Fatalf("a close on a delivery that predates the cut was not refused by the delivered-text rule: %v", err)
			}
			lensRereads(t, runDir) // its next sitting is handed the cut, whose change carries G's old paragraph
			if err := lensClose(t, runDir, lensSeat, g, "moot"); err != nil {
				t.Fatalf("the lens's moot close after its delivery of the cut was refused: %v", err)
			}
			// THE PASSAGE ALONE: at a later sitting the head has not moved past its pin, so no change
			// carries the paragraph, and G's own passage is the cutting edit's old paragraph.
			if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
				t.Fatal(err)
			}
			w, err := record.WorkOfSeat(runtest.Open(t, runDir), seatclass.RoleOfSeat(lensSeat), lensSeat)
			if err != nil {
				t.Fatal(err)
			}
			cutD, err := record.CutDelivery(runtest.Open(t, runDir), lensSeat, record.ReportTextSubjects(w), func(string, string) bool { return true })
			if err != nil {
				t.Fatal(err)
			}
			if !slices.ContainsFunc(cutD.Pieces, func(p record.DeliveryPiece) bool {
				return strings.Contains(p.Label, "THE PASSAGE OF GAP "+g+", AS IT STOOD BEFORE THE LATEST EDIT THAT CHANGED OR CUT ITS SENTENCE") &&
					strings.Contains(p.Text, "The grass is green"+anchor.Token(g))
			}) {
				t.Errorf("%s's passage is not the cut's old paragraph: %+v", g, cutD.Pieces)
			}
		})
	}
	t.Run("a cut answering the gap itself reads as an answer", func(t *testing.T) {
		runDir, g, _ := raiserBoard(t)
		dispatchBlueOnto(t, runDir, g)
		registerBlue(t, runDir)
		cutAnswering(t, runDir, g, g)
		closeBlueSitting(t, runDir)
		items := lensItems(t, runDir)
		if !anyHas(items, "gap "+g+" has been ANSWERED by a change to the report") {
			t.Errorf("no editedAt item on %s:\n  %s", g, strings.Join(items, "\n  "))
		}
		if anyHas(items, "gap "+g+" "+gone) {
			t.Errorf("the gone item on %s, which the cut answered:\n  %s", g, strings.Join(items, "\n  "))
		}
	})
	t.Run("a --quote computation gap's demand is met by its marker's cut", func(t *testing.T) {
		runDir, g, _ := raiserBoard(t, "--check-kind", "computation")
		cutAnswering(t, runDir, g, "")
		lensRereads(t, runDir)
		if err := lensClose(t, runDir, lensSeat, g, "moot"); err != nil {
			t.Fatalf("the moot close of a --quote computation gap whose marker was cut was refused: %v", err)
		}
	})
	t.Run("a failed check on a citation the cut took clears", func(t *testing.T) {
		runDir, _, _ := raiserBoard(t)
		c := citeSentence(t, runDir, "The sky is blue.", "https://water/1")
		if _, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", c,
			"--as", "unreachable", "--confidence", "high", "--reason", "it would not load"); err != nil {
			t.Fatalf("the unreachable verify of %s: %v", c, err)
		}
		g := materialGapAbout(t, runDir, c, "document")
		if err := lensClose(t, runDir, lensSeat, g, "moot"); err == nil || !strings.Contains(err.Error(), "carries citation "+c) {
			t.Fatalf("the close of %s over its standing failed check was not refused: %v", g, err)
		}
		if _, err := cut(t, runDir, "The sky is blue"+anchor.Token(c)+"."); err != nil {
			t.Fatalf("the cut of %s's sentence: %v", c, err)
		}
		lensRereads(t, runDir)
		if err := lensClose(t, runDir, lensSeat, g, "moot"); err != nil {
			t.Errorf("the moot close of %s after the cut took %s was refused: %v", g, c, err)
		}
	})
	t.Run("the lens keeps it", func(t *testing.T) {
		runDir, g, _ := raiserBoard(t)
		cutAnswering(t, runDir, g, "")
		lensRereads(t, runDir)
		if _, err := run(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--severity", "medium",
			"--reason", "the overreach survives in the section's opening"); err != nil {
			t.Fatalf("the lens's regrade of %s after the cut: %v", g, err)
		}
		if gp := boardGap(t, runDir, g); gp == nil {
			t.Errorf("%s is not open after the lens kept it", g)
		}
	})
}

// boardGap is gap g's open row on the board, nil where it is not open.
func boardGap(t *testing.T, runDir, g string) *record.GapJSON {
	t.Helper()
	b := boardOf(t, runDir)
	for i := range b.Open {
		if b.Open[i].ID == g {
			return &b.Open[i]
		}
	}
	return nil
}
