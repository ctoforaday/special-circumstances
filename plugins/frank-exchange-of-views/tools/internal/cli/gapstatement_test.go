package cli

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// gapOnGapDocketed is a gap G minted by lensSeat, otherLens's gap G2 about it, and G2 docketed at
// impasse: G2's raiser is otherLens, its author lensSeat.
func gapOnGapDocketed(t *testing.T) (runDir, g, g2, d string) {
	t.Helper()
	runDir = seatRun(t)
	g = mintGap(t, runDir, "under", "spec_underspecification")
	asSeat(t, runDir, otherLens, "register")
	g2, err := mintOnGap(t, runDir, otherLens, "on-gap", g)
	if err != nil {
		t.Fatal(err)
	}
	d = docketNow(t, runDir, g2, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	return runDir, g, g2, d
}

// closingOn is seat's closing on gap with extra flags.
func closingOn(t *testing.T, runDir, seat, gap string, extra ...string) (string, error) {
	t.Helper()
	return runAt(t, append([]string{"closing", "--run", runDir, "--seat-id", seat, "--id", gap, "--reason", "the case"}, extra...)...)
}

// T-B44 (the statements): ON A GAP ABOUT A GAP THE RAISER STATES THAT GAP'S WORDING, GRADES OR BOTH,
// AND THE AUTHOR — A LENS — ITS WORD. The raiser's closing on G2 states G's wording, its grades, both,
// or --no-text, exactly one of a statement and --no-text; a pair or a retire is refused (no report text
// is its subject). The author's closing states its word with --as and no statement. A wording or
// grades on a gap about a sentence is refused.
func TestTheRaiserStatesTheGapsWordingOrGrades(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
	}{
		{"a wording", []string{"--problem", "the gap states the check it means"}},
		{"grades", []string{"--severity", "low", "--likelihood", "low"}},
		{"both", []string{"--acceptance-check", "the check names an observable outcome", "--impact", "low"}},
		{"no statement", []string{"--no-text"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir, _, g2, _ := gapOnGapDocketed(t)
			if out, err := closingOn(t, runDir, otherLens, g2, c.args...); err != nil {
				t.Fatalf("the raiser's closing stating %s: %v\n%s", c.name, err, out)
			}
		})
	}
	runDir, g, g2, _ := gapOnGapDocketed(t)
	for _, c := range []struct {
		name, seat string
		args       []string
		want       string
	}{
		{"a pair", otherLens, []string{"--quote", "x", "--new", "y"}, "state report text"},
		{"nothing", otherLens, nil, "exactly one of"},
		{"a statement and --no-text", otherLens, []string{"--problem", "p", "--no-text"}, "exactly one of"},
		{"the author's statement", lensSeat, []string{"--problem", "p"}, "the author's closing states its word"},
	} {
		closings := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSING)
		out, err := closingOn(t, runDir, c.seat, g2, c.args...)
		heldRefusal(t, c.name, out, err, c.want)
		if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSING); got != closings {
			t.Errorf("%s: closings on the record %d, want %d", c.name, got, closings)
		}
	}
	if out, err := closingOn(t, runDir, lensSeat, g2, "--as", "not_a_defect"); err != nil {
		t.Errorf("the author lens's closing word: %v\n%s", err, out)
	}
	// A WORDING ON A GAP ABOUT A SENTENCE states nothing a ruling there applies.
	other := mintGap(t, runDir, "other", "unsourced_claim")
	docketNow(t, runDir, other, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	out, err := closingOn(t, runDir, lensSeat, other, "--problem", "p")
	heldRefusal(t, "a wording on a gap about a section", out, err, "a wording or grades state what a gap about a gap should make of that gap")
	_ = g
}

// T-B44 (the ruling): FOR THE RAISER, THE BENCH APPLIES THE STATEMENT AS ITS AMEND AND REGRADE OF THE
// GAP. A for-raiser ruling on D(G2) writes the bench's amend of G from the wording stated and its
// regrade from the grades stated, each answering G2, in the ruling's transaction; G keeps its id and
// its minter, open, and G2 closes repaired. A ruling that sides with a raiser who stated nothing is
// refused.
func TestForTheRaiserTheBenchAmendsAndRegradesTheGap(t *testing.T) {
	runDir, g, g2, d := gapOnGapDocketed(t)
	if out, err := closingOn(t, runDir, otherLens, g2, "--problem", "the gap states the check it means", "--severity", "low"); err != nil {
		t.Fatalf("the raiser's statement: %v\n%s", err, out)
	}
	if out, err := closingOn(t, runDir, lensSeat, g2, "--as", "not_a_defect"); err != nil {
		t.Fatalf("the author's word: %v\n%s", err, out)
	}
	conveneBench(t, runDir, g2)
	if out, err := answerDocketing(t, runDir, d, "for-raiser", "--final"); err != nil {
		t.Fatalf("--for-raiser: %v\n%s", err, out)
	}
	if r := lastRuling(t, runDir); r.GetSidedWith() != recordpb.SidedWith_SIDED_WITH_RAISER || r.GetDisposition() != recordpb.Disposition_DISPOSITION_REPAIRED {
		t.Errorf("the ruling sides %v with %v, want the raiser and repaired", r.GetSidedWith(), r.GetDisposition())
	}
	amends, regrades := benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_AMEND), benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE)
	if len(amends) != 1 || len(regrades) != 1 {
		t.Fatalf("the bench wrote %d amend(s) and %d regrade(s), want one of each", len(amends), len(regrades))
	}
	a := lastBody(t, runDir, &recordpb.Amend{})
	if a.GetGapId() != g || a.GetAnswers() != g2 || a.GetResultProblem() != "the gap states the check it means" {
		t.Errorf("the bench's amend = %v, want G %s's problem restated, answering %s", a, g, g2)
	}
	rg := lastBody(t, runDir, &recordpb.Regrade{})
	if rg.GetGapId() != g || rg.GetAnswers() != g2 || rg.GetSeverity() != recordpb.Grade_GRADE_LOW {
		t.Errorf("the bench's regrade = %v, want G %s's severity low, answering %s", rg, g, g2)
	}
	under := openGap(t, runDir, g)
	if under.Problem != "the gap states the check it means" || under.Severity != "low" || under.Assignee != lensSeat {
		t.Errorf("G %s reads problem %q severity %v assignee %q, want the stated problem and grade, and its minter", g, under.Problem, under.Severity, under.Assignee)
	}
	for _, o := range boardOf(t, runDir).Open {
		if o.ID == g2 {
			t.Errorf("G2 %s is still open after the ruling for its raiser", g2)
		}
	}

	runDir, _, g2, d = gapOnGapDocketed(t)
	if out, err := closingOn(t, runDir, otherLens, g2, "--no-text"); err != nil {
		t.Fatalf("the raiser's --no-text: %v\n%s", err, out)
	}
	conveneBench(t, runDir, g2)
	out, err := answerDocketing(t, runDir, d, "for-raiser", "--final")
	heldRefusal(t, "for the raiser who stated nothing", out, err, "states nothing")
}

// T-B44/T-B49 (the author's accept): THE AUTHOR CONCEDES D(G2) BY TAKING ITS RAISER'S STATEMENT. While
// D(G2) stands, A's own amend of G is refused (G is locked); A's `amend --accept G2` — or `regrade
// --accept G2`, the same offer — writes the stated wording and grades on G, each answering G2 and
// accepted, and ends D(G2) with no ruling; G2 stays open for its raiser.
func TestTheAuthorTakesTheStatementOnAGapAboutItsGap(t *testing.T) {
	for _, verb := range []string{"amend", "regrade"} {
		t.Run(verb, func(t *testing.T) {
			runDir, g, g2, _ := gapOnGapDocketed(t)
			if out, err := closingOn(t, runDir, otherLens, g2, "--problem", "the gap states the check it means", "--likelihood", "low"); err != nil {
				t.Fatalf("the raiser's statement: %v\n%s", err, out)
			}
			out, err := amendGap(t, runDir, lensSeat, g, "--problem", "my own restatement", "--reason", "r")
			heldRefusal(t, "the author's own amend while D(G2) stands", out, err, "is locked", g2)
			out, err = runAt(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--severity", "high", "--reason", "r")
			heldRefusal(t, "the author's own regrade while D(G2) stands", out, err, "is locked", g2)
			out, err = runAt(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--as", "not_a_defect",
				"--verified-by", lensSeat, "--verified-with", "a re-read", "--verified-against", "the report", "--reason", "r")
			heldRefusal(t, "the author's own close while D(G2) stands", out, err, "is locked", g2)
			if out, err := runAt(t, verb, "--run", runDir, "--seat-id", lensSeat, "--accept", g2, "--reason", "the raiser is right"); err != nil {
				t.Fatalf("%s --accept %s: %v\n%s", verb, g2, err, out)
			}
			a, rg := lastBody(t, runDir, &recordpb.Amend{}), lastBody(t, runDir, &recordpb.Regrade{})
			if a.GetGapId() != g || a.GetAnswers() != g2 || !a.GetAccepted() || a.GetResultProblem() != "the gap states the check it means" {
				t.Errorf("the accept's amend = %v, want the stated problem on %s, answering %s, accepted", a, g, g2)
			}
			if rg.GetGapId() != g || rg.GetAnswers() != g2 || !rg.GetAccepted() || rg.GetLikelihood() != recordpb.Grade_GRADE_LOW || rg.Severity != nil {
				t.Errorf("the accept's regrade = %v, want likelihood low alone on %s, answering %s, accepted", rg, g, g2)
			}
			if _, docketing, err := record.StandingOnGap(runtest.Open(t, runDir), g2); err != nil || docketing != "" {
				t.Errorf("D(G2) after the author's concession: %q (%v), want none standing", docketing, err)
			}
			if !slices.ContainsFunc(boardOf(t, runDir).Open, func(o record.GapJSON) bool { return o.ID == g2 }) {
				t.Errorf("G2 %s closed on the author's concession; it stays open for its raiser", g2)
			}
		})
	}
}
