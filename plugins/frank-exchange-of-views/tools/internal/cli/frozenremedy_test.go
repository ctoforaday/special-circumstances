package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// OWED BUT FROZEN DOES NOT BLOCK; RED'S OWN TEXT IS ADMITTED (gblock 2026-11-23). A citation on a
// docketed gap's sentence whose source no lens could read owes blue its remedy: the citation's retire,
// or an edit of its sentence. Where red's text on the gap quotes the span the citation's token stands
// in, the freeze refuses both — so the item does not hold blue's sitting, and still holds the PASS.
// Where red's text IS the citation's retire, blue's retire does what red's text does: the freeze
// admits it, and the bench's answer for red applies nothing more.

// frozenRemedyRun is acceptRun with blue's citation on acceptS, a gap of the evidence lens's on it —
// quoting it, or about the citation where aboutCite — docketed at impasse.
func frozenRemedyRun(t *testing.T, aboutCite bool) (runDir, cite, gap, docketing string) {
	t.Helper()
	runDir = acceptRun(t)
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", acceptS+".", "--url", "https://costs/1", "--title", "Source"); err != nil {
		t.Fatalf("blue's cite: %v %s", err, out)
	}
	cite = lastBody(t, runDir, &recordpb.Cite{}).GetLabel()
	var err error
	if aboutCite {
		var out string
		out, err = runMint(t, "--run", runDir, "--seat-id", lensSeat, "--key", "G", "--about-kind", "anchor", "--about", cite,
			"--class", "spec_underspecification", "--problem", "the defect", "--fix", "the fix", "--check-kind", "document",
			"--check", "the acceptance check", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low")
		gap = gapID(out)
	} else {
		gap, err = mintQuote(t, runDir, "G", acceptS+".")
	}
	if err != nil {
		t.Fatalf("the lens's gap: %v", err)
	}
	docketing = docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	return runDir, cite, gap, docketing
}

// unreadAndClosed is the evidence lens's check of cite finding its source unreachable, and blue's
// closing on gap.
func unreadAndClosed(t *testing.T, runDir, cite, gap string) {
	t.Helper()
	if out, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", cite, "--as", "unreachable", "--confidence", "high", "--reason", "the page will not load"); err != nil {
		t.Fatalf("the lens's unreachable check: %v %s", err, out)
	}
	if out, err := blueStates(t, runDir, gap, ""); err != nil {
		t.Fatalf("blue's closing: %v %s", err, out)
	}
}

func TestAnOwedRemedyTheFreezeHoldsDoesNotHoldBluesSitting(t *testing.T) {
	runDir, cite, gap, _ := frozenRemedyRun(t, false)
	report, err := record.RenderRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	words := "Costs rose sharply across the first quarter"
	tail := report[strings.Index(report, words)+len(words):]
	run_ := tail[:anchor.SkipRun(tail, 0)]
	if !strings.Contains(run_, anchor.Token(cite)) {
		t.Fatalf("the citation's token does not end the words: %q", report)
	}
	if out, err := redStates(t, runDir, gap, "--quote", words+run_, "--new", "Costs rose a little"+run_); err != nil {
		t.Fatalf("red's text over the cited words: %v %s", err, out)
	}
	unreadAndClosed(t, runDir, cite, gap)
	out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", cite, "--reason", "no lens can read it")
	heldRefusal(t, "blue's retire of the citation red's text quotes", out, err, "holds until its ruling", words)
	if err != nil && strings.Contains(err.Error()+out, "your replacement") {
		t.Errorf("the freeze's refusal speaks to a seat typing a pair: %v", err)
	}
	refusal, err := record.EnvelopeRefusal(runtest.Open(t, runDir), blueSeat, true)
	if err != nil || refusal != "" {
		t.Errorf("blue's envelope with only a frozen remedy owed: refused %q (%v), want admitted", refusal, err)
	}
	bs, err := record.PassBlockers(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	held := false
	for _, b := range bs {
		held = held || b.Kind == record.BlockerFailedCheck && b.Subject == cite
	}
	if !held {
		t.Errorf("the failed check of %s no longer holds the PASS: %+v", cite, bs)
	}
}

func TestBluesRetireThatIsRedsTextLandsAndTheBenchAppliesNothingMore(t *testing.T) {
	runDir, cite, gap, d := frozenRemedyRun(t, true)
	if out, err := redStates(t, runDir, gap, "--retire"); err != nil {
		t.Fatalf("red's text, the citation's retire: %v %s", err, out)
	}
	unreadAndClosed(t, runDir, cite, gap)
	if out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", cite, "--answers", gap, "--reason", "no lens can read it"); err != nil {
		t.Fatalf("blue's retire of the citation red's text retires: %v %s", err, out)
	}
	conveneBench(t, runDir, gap)
	if out, err := answerDocketing(t, runDir, d, "for-raiser", "--final"); err != nil {
		t.Fatalf("--for-raiser with red's text already done: %v %s", err, out)
	}
	if body := lastBody(t, runDir, &recordpb.Retire{}); len(body.GetAnchors()) != 1 || body.GetAnchors()[0] != cite {
		t.Errorf("the last retire on the record is %+v, want blue's of %s alone", body, cite)
	}
}

// BLUE'S EDIT THAT IS RED'S REPLACEMENT LANDS, and the bench's answer for red applies nothing more: the
// report already holds red's text at the gap's place (record.RedTextStands).
func TestBluesEditThatIsRedsReplacementLandsAndTheBenchAppliesNothingMore(t *testing.T) {
	runDir, _, gap, d := frozenRemedyRun(t, false)
	report, err := record.RenderRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	words := "Costs rose sharply across the first quarter"
	tail := report[strings.Index(report, words)+len(words):]
	run_ := tail[:anchor.SkipRun(tail, 0)]
	if out, err := redStates(t, runDir, gap, "--quote", words+run_, "--new", "Costs rose a little"+run_); err != nil {
		t.Fatalf("red's text over the cited words: %v %s", err, out)
	}
	if out, err := blueStates(t, runDir, gap, ""); err != nil {
		t.Fatalf("blue's closing: %v %s", err, out)
	}
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", words, "--new", "Costs rose a little", "--answers", gap, "--reason", "red's text"); err != nil {
		t.Fatalf("blue's edit applying red's replacement: %v %s", err, out)
	}
	edits := func() int {
		fam, err := record.FamilyOf(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, e := range fam.Events {
			if _, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
				n++
			}
		}
		return n
	}
	before := edits()
	conveneBench(t, runDir, gap)
	if out, err := answerDocketing(t, runDir, d, "for-raiser", "--final"); err != nil {
		t.Fatalf("--for-raiser with red's text already done: %v %s", err, out)
	}
	if after := edits(); after != before {
		t.Errorf("the answer for red wrote %d edit(s) over red's text already standing, want none", after-before)
	}
}

// BLUE'S OWN REWORD OF RED'S SPAN IS STILL HELD: an edit of the words red's text quotes that is not
// exactly red's replacement — other words, or red's with words added — is refused by the freeze.
func TestBluesOwnRewordOfRedsSpanIsHeld(t *testing.T) {
	runDir, _, gap, _ := frozenRemedyRun(t, false)
	report, err := record.RenderRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	words := "Costs rose sharply across the first quarter"
	tail := report[strings.Index(report, words)+len(words):]
	run_ := tail[:anchor.SkipRun(tail, 0)]
	if out, err := redStates(t, runDir, gap, "--quote", words+run_, "--new", "Costs rose a little"+run_); err != nil {
		t.Fatalf("red's text over the cited words: %v %s", err, out)
	}
	for _, own := range []string{"Costs rose somewhat", "In fact, Costs rose a little"} {
		out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", words, "--new", own, "--answers", gap, "--reason", "blue's own")
		heldRefusal(t, "blue's own reword of red's span to "+own, out, err, "holds until its ruling", words)
	}
}

// A RETIRE ROAD L6 WOULD REFUSE IS REFUSED WHEN STATED. Red's statement is checked as its application
// would be: while a computation still in the report applies the citation's method, the bench's retire
// for red and blue's `retire --accept` are both refused, so red's closing stating that retire is refused
// naming the computation, and nothing is recorded (seed 5, CI run 46094143286). Each remedy the refusal
// names is then statable: a pair over the citation's sentence, or --no-text; and the retire itself once
// the computation is retired.
func TestARetireRoadL6WouldRefuseIsRefusedWhenStated(t *testing.T) {
	setup := func(t *testing.T) (runDir, cite, gap, k string) {
		t.Helper()
		runDir, cite, gap, _ = frozenRemedyRun(t, true)
		s := script(t, runDir, "k.js", "console.log('k');")
		if out, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "Volume recovered later", "--dir", s,
			"--reason", "the method applied", "--cites", cite, "--", "node", "k.js"); err != nil {
			t.Fatalf("compute --cites %s: %v %s", cite, err, out)
		}
		k = lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
		closings := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSING)
		out, err := redStates(t, runDir, gap, "--retire")
		if err == nil {
			t.Fatalf("red's retire of %s while %s applies it was admitted: %s", cite, k, out)
		}
		for _, want := range []string{k, cite, "L6", "--no-text"} {
			if !strings.Contains(err.Error()+out, want) {
				t.Errorf("the refusal does not name %q: %v %s", want, err, out)
			}
		}
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSING); got != closings {
			t.Errorf("closings on the record %d, want %d: the refused statement recorded one", got, closings)
		}
		return runDir, cite, gap, k
	}
	t.Run("the retire, once the computation is retired", func(t *testing.T) {
		runDir, cite, gap, k := setup(t)
		if out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", k, "--reason", "the method leaves"); err != nil {
			t.Fatalf("blue's retire of %s: %v %s", k, err, out)
		}
		if out, err := redStates(t, runDir, gap, "--retire"); err != nil {
			t.Errorf("red's retire of %s once no computation applies it: %v %s", cite, err, out)
		}
	})
	t.Run("a pair over the citation's sentence", func(t *testing.T) {
		runDir, _, gap, _ := setup(t)
		report, err := record.RenderRun(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		words := "Costs rose sharply across the first quarter"
		tail := report[strings.Index(report, words)+len(words):]
		run_ := tail[:anchor.SkipRun(tail, 0)]
		if out, err := redStates(t, runDir, gap, "--quote", words+run_, "--new", "Costs rose a little"+run_); err != nil {
			t.Errorf("red's pair after the refused retire: %v %s", err, out)
		}
	})
	t.Run("--no-text", func(t *testing.T) {
		runDir, _, gap, _ := setup(t)
		if out, err := redStates(t, runDir, gap, "--no-text"); err != nil {
			t.Errorf("red's --no-text after the refused retire: %v %s", err, out)
		}
	})
}

// EVERY ACT OF BLUE'S THAT CAN BREAK A STANDING STATEMENT IS HELD BY THE SAME GUARD (the freeze's
// roadsHold): a computation applying a citation's method makes red's stated retire of that citation one
// L6 refuses, so `compute --cites C` after the statement is refused naming the docketing that holds it,
// and nothing is recorded; a computation citing nothing lands.
func TestAComputationThatBreaksAStandingRetireIsRefused(t *testing.T) {
	runDir, cite, gap, d := frozenRemedyRun(t, true)
	if out, err := redStates(t, runDir, gap, "--retire"); err != nil {
		t.Fatalf("red's statement, the citation's retire: %v %s", err, out)
	}
	s := script(t, runDir, "k.js", "console.log('k');")
	computations := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION)
	out, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "Volume recovered later", "--dir", s,
		"--reason", "the method applied", "--cites", cite, "--", "node", "k.js")
	heldRefusal(t, "compute --cites "+cite+" under red's retire of it", out, err, d, gap, cite, "L6", "holds until its ruling")
	if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION); got != computations {
		t.Errorf("computations on the record %d, want %d: the refused act recorded one", got, computations)
	}
	if out, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "Volume recovered later", "--dir", s,
		"--reason", "a computation of its own", "--", "node", "k.js"); err != nil {
		t.Errorf("a computation citing no method under the same statement: %v %s", err, out)
	}
}

// A CITATION THAT BREAKS A STANDING PAIR IS REFUSED by the same guard: its token lands inside the span
// red's stated pair quotes, so the pair no longer plans on the report it leaves; a citation elsewhere
// lands.
func TestACitationThatBreaksAStandingPairIsRefused(t *testing.T) {
	runDir, gap, d := positionsRun(t)
	if out, err := redStates(t, runDir, gap); err != nil {
		t.Fatalf("red's pair: %v %s", err, out)
	}
	cites := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE)
	out, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", acceptS+".", "--url", "https://costs/1", "--title", "Source")
	heldRefusal(t, "a cite inside red's quoted span", out, err, d, gap, "holds until its ruling")
	if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE); got != cites {
		t.Errorf("citations on the record %d, want %d: the refused act recorded one", got, cites)
	}
	if out, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", "Volume recovered later.", "--url", "https://costs/1", "--title", "Source"); err != nil {
		t.Errorf("a cite outside red's span: %v %s", err, out)
	}
}
