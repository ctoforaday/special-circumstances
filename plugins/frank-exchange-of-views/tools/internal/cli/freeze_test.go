package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE FREEZE HOLDS EVERY STATEMENT STANDING ON A GAP UNTIL ITS RULING (#1695 §III.B14, L43, L45, L48),
// driven through the command tree as the seats run it: red's closing's road is re-planned on the report
// each blue edit leaves, and a road is estopped as a red mint is, at the closing and at the ruling.

const (
	// freezeElse is the sentence before acceptS in acceptRun's report.
	freezeElse = "The year opened badly."
	// freezeLater is the sentence after it.
	freezeLater = "Volume recovered later."
)

// blueEditOf is blue's edit of old into new, answering gap, under key; it returns the refusal.
func blueEditOf(t *testing.T, runDir, key, gap, old, new string) (string, error) {
	t.Helper()
	args := []string{"edit", "--run", runDir, "--seat-id", blueSeat, "--key", key, "--quote", old, "--new", new, "--reason", "blue's change"}
	if gap != "" {
		args = append(args, "--answers", gap)
	}
	return runAt(t, args...)
}

// heldRefusal fails unless err is a refusal whose text carries every want.
func heldRefusal(t *testing.T, what, out string, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s landed; want it refused naming %v\n%s", what, want, out)
		return
	}
	for _, w := range want {
		if !strings.Contains(err.Error()+out, w) {
			t.Errorf("%s: the refusal does not name %q:\n%v\n%s", what, w, err, out)
		}
	}
}

// T-B29's freeze and estoppel rows.
func TestRedsTextIsAppliedWhenTheBenchSidesWithRed(t *testing.T) {
	t.Run("the freeze: blue's edit over red's span is refused naming the docketing; an edit elsewhere lands", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing: %v\n%s", err, out)
		}
		out, err := blueEditOf(t, runDir, "over", gap, acceptS, "Costs rose a little across the first quarter")
		heldRefusal(t, "blue's edit over red's span", out, err, d, gap)
		if rep := readReport(t, runDir); !strings.Contains(rep, acceptS) {
			t.Errorf("the refused edit moved the report:\n%s", rep)
		}
		if out, err := blueEditOf(t, runDir, "else", "", freezeElse, "The year opened poorly."); err != nil {
			t.Errorf("an edit elsewhere, leaving red's pair planning, was refused: %v\n%s", err, out)
		}
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "for-raiser", "--final"); err != nil {
			t.Errorf("--for-raiser after the freeze held red's text: %v\n%s", err, out)
		}
	})

	// THE PAIR IS LOCATED BY THE GAP'S MARKER (blue.PlanAccept takes the span ending at it byte for byte
	// where its words stand twice), so a copy of red's quote elsewhere leaves the pair planning: the edit
	// lands, and the for-raiser answer replaces the span at the marker, not the copy.
	t.Run("the freeze: an edit elsewhere copying red's quote leaves the pair planning, and for red applies it at the marker", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing: %v\n%s", err, out)
		}
		if out, err := blueEditOf(t, runDir, "dup", "", freezeLater, freezeLater+" "+acceptS+"."); err != nil {
			t.Fatalf("an edit copying red's quote elsewhere: %v\n%s", err, out)
		}
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "for-raiser", "--final"); err != nil {
			t.Fatalf("--for-raiser: %v\n%s", err, out)
		}
		if rep := readReport(t, runDir); strings.Count(rep, acceptS) != 1 || !strings.Contains(rep, acceptN) {
			t.Errorf("for red did not replace exactly the span at the gap's marker:\n%s", rep)
		}
	})

	t.Run("the freeze in the assignee's window: red's closing for blue's close proposal holds its span, naming the proposal", func(t *testing.T) {
		runDir := acceptRun(t)
		gap, err := mintQuote(t, runDir, "G", acceptS+".")
		if err != nil {
			t.Fatal(err)
		}
		exchangeOn(t, runDir, gap)
		m := proposeClose(t, runDir, blueSeat, gap, "defect_owed_elsewhere", "the engine owns this")
		if out, err := blueEditOf(t, runDir, "before", gap, freezeLater, "Volume recovered by June."); err != nil {
			t.Fatalf("an edit before red's closing stands: %v\n%s", err, out)
		}
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing in the window: %v\n%s", err, out)
		}
		out, err := blueEditOf(t, runDir, "over", gap, acceptS, "Costs rose a little across the first quarter")
		heldRefusal(t, "blue's edit over red's span in the window", out, err, m)
	})

	t.Run("estoppel: a road quoting text another lens prescribed and blue applied is refused at the closing, naming it", func(t *testing.T) {
		runDir := acceptRun(t)
		gap, err := mintQuote(t, runDir, "G", acceptS+".")
		if err != nil {
			t.Fatal(err)
		}
		docketing := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
		p := estoppelOver(t, runDir)
		out, err := redStates(t, runDir, gap)
		heldRefusal(t, "red's closing quoting estopped text", out, err, p)
		_ = docketing
	})

	t.Run("estoppel at the ruling: a prescription applied after red's closing refuses --for-raiser, naming it; --for-author lands", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing: %v\n%s", err, out)
		}
		p := estoppelOver(t, runDir)
		conveneBench(t, runDir, gap)
		out, err := answerDocketing(t, runDir, d, "for-raiser", "--final")
		heldRefusal(t, "--for-raiser over an estopped road", out, err, p)
		if out, err := answerDocketing(t, runDir, d, "for-author", "--final"); err != nil {
			t.Errorf("--for-author: %v\n%s", err, out)
		}
	})
}

// estoppelOver mints P on freezeLater prescribing text that holds acceptS's sentence whole, and has
// blue accept it verbatim: P's prescription then estops a quote of acceptS. It returns P.
func estoppelOver(t *testing.T, runDir string) string {
	t.Helper()
	p, err := mintQuote(t, runDir, "P", freezeLater, "--class", otherClass, "--new", freezeLater+" "+acceptS+".")
	if err != nil {
		t.Fatalf("minting P: %v", err)
	}
	if err := acceptEdit(t, runDir, "accept-p", p, "--accept"); err != nil {
		t.Fatalf("blue's accept of P: %v", err)
	}
	return p
}

// T-B30: no proposal is filed on a docketed gap (L38), of any act — only the author files one (X6); after a remand one
// lands.
func TestNoProposalOutlivesARulingOfItsGap(t *testing.T) {
	runDir, d := impasseDocketed(t)
	for _, f := range []struct{ seat, act string }{{"blue-respond", "close"}, {"blue-respond", "regrade"}} {
		args := []string{"motion", "docket", "file", f.act, "--id", boardGap, "--reason", "after the docketing"}
		switch f.act {
		case "close":
			args = append(args, closeFlags("not_a_defect")...)
		case "regrade":
			args = append(args, "--severity", "low")
		}
		actRefusedNaming(t, runDir, f.seat+"'s "+f.act+" proposal on a docketed gap", f.seat, []string{"docketing " + d}, args...)
	}
	conveneBench(t, runDir, boardGap)
	if out, err := runAt(t, "motion", "docket", "rule", "--run", runDir, "--seat-id", "judge", "--id", d, "--remand", "--direction", "one more exchange",
		"--principle", "correctness", "--tension", "t", "--review-flag", "none", "--settled", "s", "--reason", "remanded"); err != nil {
		t.Fatalf("the remand: %v\n%s", err, out)
	}
	out, err := runAt(t, append([]string{"motion", "docket", "file", "close", "--run", runDir, "--seat-id", "blue-respond", "--id", boardGap, "--reason", "after the remand"}, closeFlags("not_a_defect")...)...)
	if err != nil && strings.Contains(err.Error()+out, "docketing "+d) {
		t.Errorf("a proposal after the remand is refused naming the remanded docketing: %v\n%s", err, out)
	}
}

// THE BENCH RULES THE DOCKET IN ORDER, AND EACH RULING STANDS ON THE ONE BEFORE (#1695 §III.B4, the
// docket order; X21). G and H, two gaps on one sentence, are docketed in that order and the bench is
// convened on both. Only G, the docket's first, is ruled first: H is refused naming G. For red on G
// applies G's text over the span H's road quotes — no other statement holds the head's application —
// and H, now first, finds its road broken: for red on H is refused naming G's ruling, no closing on H is
// owed again, and for blue on H lands (the residue). Where G's text leaves H's road planning, both are
// ruled for red in the one sitting, in order.
func TestTheDocketsFirstIsRuledAndALaterBrokenStatementLosesNamingTheRuling(t *testing.T) {
	const steep = "Costs rose steeply across the first quarter."
	setup := func(t *testing.T) (runDir, g, h, dg, dh string) {
		t.Helper()
		runDir = acceptRun(t)
		var err error
		if g, err = mintQuote(t, runDir, "G", acceptS+"."); err != nil {
			t.Fatal(err)
		}
		if h, err = mintQuote(t, runDir, "H", acceptS+".", "--class", otherClass); err != nil {
			t.Fatal(err)
		}
		dg = docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
		dh = docketNow(t, runDir, h, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
		return runDir, g, h, dg, dh
	}

	t.Run("the head's text applies over a later road, which then loses for red naming the ruling; for blue lands", func(t *testing.T) {
		runDir, g, h, dg, dh := setup(t)
		if out, err := redStates(t, runDir, g); err != nil {
			t.Fatalf("red's closing on G: %v\n%s", err, out)
		}
		if out, err := redStates(t, runDir, h, "--quote", acceptS+".", "--new", steep); err != nil {
			t.Fatalf("red's closing on H: %v\n%s", err, out)
		}
		conveneBench(t, runDir, g, h)
		out, err := answerDocketing(t, runDir, dh, "for-author", "--final")
		heldRefusal(t, "a ruling of H behind G", out, err, dg, "not the docket's first")
		if out, err := answerDocketing(t, runDir, dg, "for-raiser", "--final"); err != nil {
			t.Fatalf("--for-raiser on G, the docket's first: %v\n%s", err, out)
		}
		if rep := readReport(t, runDir); !strings.Contains(rep, acceptN) {
			t.Errorf("for red on G did not apply G's text:\n%s", rep)
		}
		out, err = answerDocketing(t, runDir, dh, "for-raiser", "--final")
		heldRefusal(t, "--for-raiser on H, its road broken by G's ruling", out, err, dg, "an earlier ruling broke")
		all, err := record.MergedEvents(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		for _, duty := range record.ClosingDuties(all.Events, all.At) {
			if duty.Docketing == dh && duty.Party == lensSeat && duty.State != record.ClosingFiled {
				t.Errorf("red's closing on %s is owed again after G's ruling broke its statement: %+v", dh, duty)
			}
		}
		if out, err := answerDocketing(t, runDir, dh, "for-author", "--final"); err != nil {
			t.Fatalf("--for-author on H: %v\n%s", err, out)
		}
	})

	// A ROAD THE BENCH'S APPLICATION BROKE FREEZES NOTHING (seed 28 of the release fuzz): blue's act
	// cannot make inapplicable a statement that no longer applies before it, so an edit elsewhere lands
	// while H's docketing stands, as it would with no road on H at all.
	t.Run("a road an earlier ruling broke does not freeze blue's edit elsewhere", func(t *testing.T) {
		runDir, g, h, dg, _ := setup(t)
		if out, err := redStates(t, runDir, g); err != nil {
			t.Fatalf("red's closing on G: %v\n%s", err, out)
		}
		if out, err := redStates(t, runDir, h, "--quote", acceptS+".", "--new", steep); err != nil {
			t.Fatalf("red's closing on H: %v\n%s", err, out)
		}
		conveneBench(t, runDir, g, h)
		if out, err := answerDocketing(t, runDir, dg, "for-raiser", "--final"); err != nil {
			t.Fatalf("--for-raiser on G, the docket's first: %v\n%s", err, out)
		}
		if out, err := blueEditOf(t, runDir, "elsewhere", "", freezeElse, "The year opened well."); err != nil {
			t.Fatalf("blue's edit elsewhere, H's road already broken by G's ruling: %v\n%s", err, out)
		}
	})

	t.Run("the head's text leaves the later road planning: both ruled for red in one sitting, in order", func(t *testing.T) {
		runDir, g, h, dg, dh := setup(t)
		if out, err := redStates(t, runDir, g); err != nil {
			t.Fatalf("red's closing on G: %v\n%s", err, out)
		}
		// H's road quotes text G's road leaves whole.
		if out, err := redStates(t, runDir, h, "--quote", freezeLater, "--new", "Volume recovered by June."); err != nil {
			t.Fatalf("red's closing on H: %v\n%s", err, out)
		}
		conveneBench(t, runDir, g, h)
		if out, err := answerDocketing(t, runDir, dg, "for-raiser", "--final"); err != nil {
			t.Fatalf("--for-raiser on G: %v\n%s", err, out)
		}
		if out, err := answerDocketing(t, runDir, dh, "for-raiser", "--final"); err != nil {
			t.Fatalf("--for-raiser on H, next in the same sitting: %v\n%s", err, out)
		}
		if rep := readReport(t, runDir); !strings.Contains(rep, acceptN) || !strings.Contains(rep, "Volume recovered by June") {
			t.Errorf("the two rulings did not each apply their text:\n%s", rep)
		}
	})
}
