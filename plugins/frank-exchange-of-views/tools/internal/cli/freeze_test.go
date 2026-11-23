package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE FREEZE HOLDS EVERY STATEMENT STANDING ON A GAP UNTIL ITS RULING (#1695 §III.B14, L43, L45, L48),
// driven through the command tree as the seats run it: red's closing's road and a standing supersede
// proposal's quote are re-planned on the report each blue edit leaves; a road is estopped as a red mint
// is, at the closing and at the ruling; a supersede proposal's successor is checked mintable against the
// gap's assignee at the filing, and the acts that move its checks' inputs are refused while it stands.

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

// supersedeOf is blue's supersede proposal of gap whose successor quotes quote, of class, its fix new
// where one is given; it returns the refusal and the motion id where it landed.
func supersedeOf(t *testing.T, runDir, gap, quote, class, new string) (string, error) {
	t.Helper()
	args := []string{"motion", "docket", "file", "supersede", "--run", runDir, "--seat-id", blueSeat, "--id", gap,
		"--reason", "the gap misstates the defect", "--json", "--quote", quote, "--class", class,
		"--problem", "the defect, restated", "--fix", "the fix, restated", "--check-kind", "document",
		"--check", "the acceptance check red runs at re-audit", "--severity", "medium", "--likelihood", "medium", "--impact", "medium"}
	if new != "" {
		args = append(args, "--new", new)
	}
	out, err := runAt(t, args...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runAt(t, args...)
	}
	return out, err
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
		if out, err := answerDocketing(t, runDir, d, "for-red", "--final"); err != nil {
			t.Errorf("--for-red after the freeze held red's text: %v\n%s", err, out)
		}
	})

	// THE PAIR IS LOCATED BY THE GAP'S MARKER (blue.PlanAccept takes the span ending at it byte for byte
	// where its words stand twice), so a copy of red's quote elsewhere leaves the pair planning: the edit
	// lands, and the for-red answer replaces the span at the marker, not the copy.
	t.Run("the freeze: an edit elsewhere copying red's quote leaves the pair planning, and for red applies it at the marker", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing: %v\n%s", err, out)
		}
		if out, err := blueEditOf(t, runDir, "dup", "", freezeLater, freezeLater+" "+acceptS+"."); err != nil {
			t.Fatalf("an edit copying red's quote elsewhere: %v\n%s", err, out)
		}
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "for-red", "--final"); err != nil {
			t.Fatalf("--for-red: %v\n%s", err, out)
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

	t.Run("estoppel at the ruling: a prescription applied after red's closing refuses --for-red, naming it; --for-blue lands", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing: %v\n%s", err, out)
		}
		p := estoppelOver(t, runDir)
		conveneBench(t, runDir, gap)
		out, err := answerDocketing(t, runDir, d, "for-red", "--final")
		heldRefusal(t, "--for-red over an estopped road", out, err, p)
		if out, err := answerDocketing(t, runDir, d, "for-blue", "--final"); err != nil {
			t.Errorf("--for-blue: %v\n%s", err, out)
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

// T-B38's filing and freeze rows: the successor is checked mintable against G's assignee at the
// filing, and while M stands the acts that move its checks' inputs are refused naming M.
func TestASupersedeDocketingLandsAtTheExit(t *testing.T) {
	setup := func(t *testing.T) (runDir, gap string) {
		runDir = acceptRun(t)
		gap, err := mintQuote(t, runDir, "G", acceptS+".")
		if err != nil {
			t.Fatal(err)
		}
		exchangeOn(t, runDir, gap)
		return runDir, gap
	}

	t.Run("a successor the assignee's one-gap rule holds is refused at the filing, and nothing is recorded", func(t *testing.T) {
		runDir, gap := setup(t)
		if _, err := mintQuote(t, runDir, "G2", acceptS+".", "--class", otherClass); err != nil {
			t.Fatal(err)
		}
		before := motionsOnRecord(t, runDir)
		out, err := supersedeOf(t, runDir, gap, acceptS+".", otherClass, "")
		heldRefusal(t, "a successor the one-gap rule holds", out, err, "assignee", "Nothing was recorded")
		if after := motionsOnRecord(t, runDir); after != before {
			t.Errorf("the refused filing recorded a motion (%d → %d)", before, after)
		}
	})

	t.Run("a successor whose quote does not place is refused at the filing", func(t *testing.T) {
		runDir, gap := setup(t)
		out, err := supersedeOf(t, runDir, gap, "words the report never held.", otherClass, "")
		heldRefusal(t, "a successor quoting nothing", out, err, "Nothing was recorded")
	})

	t.Run("a successor whose fix the accept would not apply is refused at the filing", func(t *testing.T) {
		runDir, gap := setup(t)
		out, err := supersedeOf(t, runDir, gap, acceptS+".", otherClass, acceptS+".")
		heldRefusal(t, "a successor whose fix changes nothing", out, err, "Nothing was recorded")
	})

	t.Run("with M standing: the assignee's mint its one-gap rule would then hold, and blue's edit of its quote, are refused naming M; the bench's for-blue lands the successor", func(t *testing.T) {
		runDir, gap := setup(t)
		out, err := supersedeOf(t, runDir, gap, acceptS+".", otherClass, "")
		if err != nil {
			t.Fatalf("the supersede proposal: %v\n%s", err, out)
		}
		m := filedMotionID(t, out)
		_, merr := mintQuote(t, runDir, "G2", acceptS+".", "--class", otherClass)
		heldRefusal(t, "the assignee's mint the successor's one-gap rule would hold", "", merr, m)
		eout, eerr := blueEditOf(t, runDir, "over", gap, acceptS, "Costs rose a little across the first quarter")
		heldRefusal(t, "blue's edit of the successor's quote", eout, eerr, m)
		d := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "for-blue"); err != nil {
			t.Fatalf("--for-blue on the supersede: %v\n%s", err, out)
		}
		if g := gapOf(t, runDir, gap); g == nil || g.Open {
			t.Errorf("gap %s is open after the bench sided with its supersede", gap)
		}
	})
}

// T-B30: no proposal is filed on a docketed gap (L38), by any filer, of any act; after a remand one
// lands.
func TestNoProposalOutlivesARulingOfItsGap(t *testing.T) {
	runDir, d := impasseDocketed(t)
	for _, f := range []struct{ seat, act string }{{"blue-respond", "close"}, {"blue-respond", "regrade"}, {"blue-respond", "supersede"}, {"red-chair", "close"}} {
		args := []string{"motion", "docket", "file", f.act, "--id", boardGap, "--reason", "after the docketing"}
		switch f.act {
		case "close":
			args = append(args, closeFlags("not_a_defect")...)
		case "regrade":
			args = append(args, "--severity", "low")
		case "supersede":
			args = append(args, supersedeFlags()...)
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

// THE BENCH'S APPLIED TEXT IS HELD BY THE FREEZE AS BLUE'S EDIT IS (L45): two gaps on one sentence,
// each docketed at impasse with red's road over it. The bench's for-red answer on one applies its road
// in the ruling's transaction; where that leaves the other's road no longer planning, the answer is
// refused naming the other docketing — every statement standing on a gap holds until its ruling, so
// the bench answers the other side, or rules the other docketing first. An answer that leaves the
// other road planning lands, and that road then applies at its own ruling.
func TestTheBenchsAppliedTextHoldsEveryOtherStandingRoad(t *testing.T) {
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

	t.Run("for red on one gap, whose text leaves the other's road unplannable, is refused naming the other docketing; for blue lands and the other's for red then applies", func(t *testing.T) {
		runDir, g, h, dg, dh := setup(t)
		if out, err := redStates(t, runDir, g); err != nil {
			t.Fatalf("red's closing on G: %v\n%s", err, out)
		}
		if out, err := redStates(t, runDir, h, "--quote", acceptS+".", "--new", steep); err != nil {
			t.Fatalf("red's closing on H: %v\n%s", err, out)
		}
		conveneBench(t, runDir, g, h)
		out, err := answerDocketing(t, runDir, dg, "for-red", "--final")
		heldRefusal(t, "--for-red on G over H's road", out, err, dh, h)
		if rep := readReport(t, runDir); !strings.Contains(rep, acceptS) {
			t.Errorf("the refused answer moved the report:\n%s", rep)
		}
		if out, err := answerDocketing(t, runDir, dg, "for-blue", "--final"); err != nil {
			t.Fatalf("--for-blue on G: %v\n%s", err, out)
		}
		if out, err := answerDocketing(t, runDir, dh, "for-red", "--final"); err != nil {
			t.Fatalf("--for-red on H, its road held: %v\n%s", err, out)
		}
		if rep := readReport(t, runDir); !strings.Contains(rep, strings.TrimSuffix(steep, ".")) {
			t.Errorf("for red on H did not apply H's road:\n%s", rep)
		}
	})

	t.Run("for red whose text leaves the other road planning lands", func(t *testing.T) {
		runDir, g, h, dg, dh := setup(t)
		if out, err := redStates(t, runDir, g); err != nil {
			t.Fatalf("red's closing on G: %v\n%s", err, out)
		}
		// H's road retires nothing and quotes text G's road leaves whole.
		if out, err := redStates(t, runDir, h, "--quote", freezeLater, "--new", "Volume recovered by June."); err != nil {
			t.Fatalf("red's closing on H: %v\n%s", err, out)
		}
		conveneBench(t, runDir, g, h)
		if out, err := answerDocketing(t, runDir, dg, "for-red", "--final"); err != nil {
			t.Fatalf("--for-red on G, H's road untouched: %v\n%s", err, out)
		}
		_ = dh
	})
}
