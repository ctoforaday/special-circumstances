package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// accept runs the assignee's `motion docket accept --id m` as seatID, repeating it once where the
// gap moved under the seat's sitting (the stale refusal moves its read point).
func accept(t *testing.T, runDir, seatID, m string) (string, error) {
	t.Helper()
	args := []string{"motion", "docket", "accept", "--run", runDir, "--seat-id", seatID, "--id", m, "--reason", "the proposal is right"}
	out, err := runAt(t, args...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runAt(t, args...)
	}
	return out, err
}

// motionStands reports whether motion m stands unruled.
func motionStands(t *testing.T, runDir, m string) bool {
	t.Helper()
	b, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range record.MotionsOf(b.Events, b.At) {
		if x.ID == m {
			return !x.Ruled()
		}
	}
	t.Fatalf("no motion %s on the record", m)
	return false
}

// eventsFromRuling is the run's events from the ruling on m onward, by type word and seat.
func eventsFromRuling(t *testing.T, runDir, m string) []string {
	t.Helper()
	all, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	from := -1
	for i, e := range all.Events {
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetMotionId() == m {
			from = i
		}
		if from >= 0 {
			out = append(out, recordpb.Word(e.GetType())+"@"+e.GetSeatId())
		}
	}
	return out
}

// proposed files blue's proposal of act on a fresh gap minted by the lens, after the gap's exchange,
// and opens the lens's next sitting; it returns the run, the gap and the motion.
func proposed(t *testing.T, key, act string, flagsOf ...string) (runDir, gap, motion string) {
	t.Helper()
	runDir = seatRun(t)
	gap = mintGap(t, runDir, key, "spec_underspecification")
	exchangeOn(t, runDir, gap)
	out, err := propose(t, runDir, blueSeat, act, gap, flagsOf...)
	if err != nil {
		t.Fatalf("blue's %s proposal: %v", act, err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	return runDir, gap, filedMotionID(t, out)
}

// T-B5: THE ASSIGNEE'S ACCEPT IS ITS RULING AND THEN THE ACT, under the assignee, in one write — each
// act's own effects landing with it.
func TestTheAssigneesAcceptAppliesTheActAsFiled(t *testing.T) {
	t.Run("regrade", func(t *testing.T) {
		runDir, g, m := proposed(t, "accept-regrade", "regrade", "--severity", "low")
		if _, err := accept(t, runDir, lensSeat, m); err != nil {
			t.Fatalf("the accept: %v", err)
		}
		if got := eventsFromRuling(t, runDir, m); len(got) < 2 || got[0] != "motion_rule@"+lensSeat || got[1] != "regrade@"+lensSeat {
			t.Errorf("the accept wrote %v, want the lens's ruling and then its regrade", got)
		}
		if gp := gapOf(t, runDir, g); gp == nil || grade(gp.Severity) != "GRADE_LOW" {
			t.Errorf("after the accept G's severity is %v, want low", gp)
		}
	})
	t.Run("close", func(t *testing.T) {
		runDir, g, m := proposed(t, "accept-close", "close", closeFlags("not_a_defect")...)
		if _, err := accept(t, runDir, lensSeat, m); err != nil {
			t.Fatalf("the accept: %v", err)
		}
		if got := eventsFromRuling(t, runDir, m); len(got) < 2 || got[0] != "motion_rule@"+lensSeat || got[1] != "close@"+lensSeat {
			t.Errorf("the accept wrote %v, want the lens's ruling and then its close", got)
		}
		if gp := gapOf(t, runDir, g); gp == nil || gp.Open || dispo(gp.Closure.GetClosureClass()) != "DISPOSITION_NOT_A_DEFECT" {
			t.Errorf("after the accept G is %+v, want closed not_a_defect", gp)
		}
	})
	t.Run("supersede, anchored at its quote", func(t *testing.T) {
		runDir, g, m := proposed(t, "accept-supersede", "supersede", "--quote", quoteTwo, "--class", "spec_underspecification",
			"--problem", "the defect, restated", "--fix", "the fix, restated", "--check-kind", "document",
			"--check", "the check", "--severity", "medium", "--likelihood", "medium", "--impact", "medium")
		_, successor, _ := proposalRow(t, runDir, m)
		if _, err := accept(t, runDir, lensSeat, m); err != nil {
			t.Fatalf("the accept: %v", err)
		}
		got := eventsFromRuling(t, runDir, m)
		if len(got) < 3 || got[0] != "motion_rule@"+lensSeat || got[1] != "mint@"+lensSeat || got[2] != "anchor@"+lensSeat {
			t.Errorf("the accept wrote %v, want the lens's ruling, its mint and the mint's anchor", got)
		}
		if s := gapOf(t, runDir, successor); s == nil || !s.Open {
			t.Errorf("the successor %s is %+v, want open", successor, s)
		}
		if gp := gapOf(t, runDir, g); gp == nil || gp.Open {
			t.Errorf("G after the accept is %+v, want closed by its successor", gp)
		}
	})
}

// T-B6: AN ACCEPT IS REFUSED for a seat that is not the assignee, and where the act's own check
// refuses it; the proposal stands after each.
func TestAnAcceptIsRefusedWhereTheActCannotLand(t *testing.T) {
	t.Run("another lens", func(t *testing.T) {
		runDir := seatRun(t)
		seedCastWith(t, runDir, "red-lens-logic")
		g := mintGap(t, runDir, "accept-other", "spec_underspecification")
		exchangeOn(t, runDir, g)
		out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
		if err != nil {
			t.Fatal(err)
		}
		m := filedMotionID(t, out)
		if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-logic"); err != nil {
			t.Fatal(err)
		}
		if out, err := accept(t, runDir, "red-lens-logic", m); err == nil || !strings.Contains(err.Error()+out, "you are red-lens-logic") {
			t.Errorf("another lens's accept = %v %q, want refused naming the assignee", err, out)
		}
		if !motionStands(t, runDir, m) {
			t.Errorf("%s does not stand after the refused accept", m)
		}
	})
	t.Run("a close the gap's demand refuses", func(t *testing.T) {
		runDir := seatRun(t)
		registerChairOnce(t, runDir)
		registerLensOnce(t, runDir)
		out, err := run(t, "mint", "--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", lensSeat,
			"--key", "accept-demand", "--class", "spec_underspecification", "--problem", "the defect", "--fix", "the fix",
			"--check-kind", "source", "--check", "a cited source settles it",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low")
		if err != nil {
			t.Fatalf("mint: %v", err)
		}
		g := gapID(out)
		exchangeOn(t, runDir, g)
		pout, err := propose(t, runDir, blueSeat, "close", g, closeFlags("not_a_defect")...)
		if err != nil {
			t.Fatalf("blue's close proposal: %v", err)
		}
		m := filedMotionID(t, pout)
		if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
			t.Fatal(err)
		}
		if out, err := accept(t, runDir, lensSeat, m); err == nil || !strings.Contains(err.Error()+out, "live-answer demand") {
			t.Errorf("the accept of a close the demand refuses = %v %q, want the close's own refusal", err, out)
		}
		if !motionStands(t, runDir, m) {
			t.Errorf("%s does not stand after the refused accept", m)
		}
		if gp := gapOf(t, runDir, g); gp == nil || !gp.Open {
			t.Errorf("G after the refused accept is %+v, want open", gp)
		}
	})
}

// THE ACT'S GUARDS RIDE THE ACT, not the first body of the write: a regrade accept on a gap that moved
// since the lens's sitting opened is refused by the regrade's own Fresh guard, writing neither the
// ruling nor the act, and the proposal stands.
func TestAnAcceptsActKeepsItsOwnGuards(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "accept-stale", "spec_underspecification")
	exchangeOn(t, runDir, g) // blue's sittings on G land after the lens's sitting opened
	out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
	if err != nil {
		t.Fatal(err)
	}
	m := filedMotionID(t, out)
	before := len(eventsOfWord(t, runDir, "motion_rule"))
	out, err = runAt(t, "motion", "docket", "accept", "--run", runDir, "--seat-id", lensSeat, "--id", m, "--reason", "r")
	if err == nil || !strings.Contains(err.Error()+out, "changed after your sitting opened") {
		t.Fatalf("an accept on a gap that moved under the lens's sitting = %v %q, want the regrade's stale refusal", err, out)
	}
	if n := len(eventsOfWord(t, runDir, "motion_rule")); n != before || !motionStands(t, runDir, m) {
		t.Errorf("the refused accept wrote %d ruling(s) and left %s standing %v; want none and standing", n-before, m, motionStands(t, runDir, m))
	}
}
