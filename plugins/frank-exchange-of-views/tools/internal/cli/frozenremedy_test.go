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
	refusal, err := record.EnvelopeRefusal(runtest.Open(t, runDir), blueSeat)
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
	if out, err := answerDocketing(t, runDir, d, "for-red", "--final"); err != nil {
		t.Fatalf("--for-red with red's text already done: %v %s", err, out)
	}
	if body := lastBody(t, runDir, &recordpb.Retire{}); len(body.GetAnchors()) != 1 || body.GetAnchors()[0] != cite {
		t.Errorf("the last retire on the record is %+v, want blue's of %s alone", body, cite)
	}
}
