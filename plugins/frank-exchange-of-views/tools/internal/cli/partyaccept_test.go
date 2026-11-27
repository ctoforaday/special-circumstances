package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// partyRecord is the record as the readers fold it: every motion and every docketing, by id.
func partyRecord(t *testing.T, runDir string) (map[string]*record.Motion, map[string]*record.Docketing, []*record.Event) {
	t.Helper()
	all, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	ms, ds := map[string]*record.Motion{}, map[string]*record.Docketing{}
	for _, m := range record.MotionsOf(all.Events, all.At) {
		ms[m.ID] = m
	}
	for _, d := range record.DocketingsOf(all.Events) {
		ds[d.ID] = d
	}
	return ms, ds, all.Events
}

// motionsPageAnswer is whether the motions page a seat reads (`show motions`, record.MotionsJSONBytes)
// shows motion m answered, and with what word.
func motionsPageAnswer(t *testing.T, runDir, m string) (bool, string) {
	t.Helper()
	b, err := record.MotionsJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var page record.MotionsJSON
	if err := json.Unmarshal(b, &page); err != nil {
		t.Fatal(err)
	}
	for _, mj := range page.Motions {
		if mj.ID == m {
			return mj.Ruled, mj.Ruling
		}
	}
	t.Fatalf("the motions page lists no %s", m)
	return false, ""
}

// standingCols is gap_state's standing proposal and docketing on gap, "" for none.
func standingCols(t *testing.T, runDir, gap string) (proposal, docketing string) {
	t.Helper()
	p, d, err := record.StandingOnGap(runtest.Open(t, runDir), handleText(t, runDir, gap))
	if err != nil {
		t.Fatal(err)
	}
	return p, d
}

// acceptArgs is the raiser's verb taking motion m: `regrade` or `close`, with --accept and --reason.
func acceptArgs(runDir, verb, m string) []string {
	return []string{verb, "--run", runDir, "--seat-id", lensSeat, "--accept", m, "--reason", "the author is right", "--json"}
}

// T-B48: THE RAISER ACCEPTS BY ITS OWN ACT. `regrade --accept M` and `close --accept M` write the act
// M stores, carrying accepts = M, and no ruling: M reads answered — "accepted", by the raiser — in the
// Go fold and the SQL view alike, no proposal stands, and the next filing is admitted. A second answer
// is refused, and so is a ruling of it. Under the docketing of M, the accept is the raiser's
// concession: it ends the docketing and M, alike in docketingsAt and gap_state, and a ruling of the
// docketing is refused after it.
func TestTheRaiserAcceptsByItsOwnAct(t *testing.T) {
	t.Run("regrade --accept M in the window", func(t *testing.T) {
		runDir := seatRun(t)
		g := mintGap(t, runDir, "accept-regrade", "spec_underspecification")
		exchangeOn(t, runDir, g)
		out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
		if err != nil {
			t.Fatalf("the author's proposal: %v", err)
		}
		m := filedMotionID(t, out)
		// The filing moved the gap after the raiser's sitting opened: the first act refuses once,
		// naming what changed, and moves the read point; the same accept lands when made again.
		if out, err := runAt(t, acceptArgs(runDir, "regrade", m)...); err != nil && !strings.Contains(err.Error()+out, "changed after your sitting opened") {
			t.Fatalf("regrade --accept %s: %v %s", m, err, out)
		}
		if out, err := runAt(t, acceptArgs(runDir, "regrade", m)...); err != nil && !strings.Contains(err.Error()+out, "already answered") {
			t.Fatalf("regrade --accept %s, read: %v %s", m, err, out)
		}
		ms, _, evs := partyRecord(t, runDir)
		if got := ms[m]; got == nil || got.Ruling != record.AcceptedWord || got.RulingBy != lensSeat {
			t.Fatalf("motion %s reads %+v; want answered %q by %s", m, got, record.AcceptedWord, lensSeat)
		}
		if ruled, word := motionsPageAnswer(t, runDir, m); !ruled || word != record.AcceptedWord {
			t.Errorf("the motions page reads %s ruled %v %q; want ruled %q", m, ruled, word, record.AcceptedWord)
		}
		var regrades, rulings int
		for _, e := range evs {
			if rg, ok := recordpb.BodyAs[*recordpb.Regrade](e); ok && rg.GetAccepts() == m {
				regrades++
				if rg.GetSeverity() != recordpb.Grade_GRADE_LOW || e.GetSeatId() != lensSeat {
					t.Errorf("the accepting regrade is %v by %s; want M's grade by the raiser", rg, e.GetSeatId())
				}
			}
			if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetMotionId() == m {
				rulings++
			}
		}
		if regrades != 1 || rulings != 0 {
			t.Errorf("the accept wrote %d regrade(s) carrying accepts and %d ruling(s); want 1 and 0", regrades, rulings)
		}
		if p, d := standingCols(t, runDir, g); p != "" || d != "" {
			t.Errorf("gap_state reads proposal %q and docketing %q standing after the accept; want none", p, d)
		}
		if out, err := runAt(t, acceptArgs(runDir, "regrade", m)...); err == nil || !strings.Contains(err.Error()+out, "already answered") {
			t.Errorf("a second accept of %s = %v %q; want refused, answered once", m, err, out)
		}
		if out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "medium"); err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
			_, err = propose(t, runDir, blueSeat, "regrade", g, "--severity", "medium")
			if err != nil {
				t.Errorf("the next proposal after the accept answered %s: %v; want admitted", m, err)
			}
		} else if err != nil {
			t.Errorf("the next proposal after the accept answered %s: %v; want admitted", m, err)
		}
	})

	t.Run("close --accept M under its docketing concedes", func(t *testing.T) {
		runDir := seatRun(t)
		g := mintGap(t, runDir, "accept-close", "spec_underspecification")
		m := docketFile(t, runDir, blueSeat, g, "the section says what the gap asks")
		d := docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
		if out, err := runAt(t, acceptArgs(runDir, "close", m)...); err != nil {
			t.Fatalf("close --accept %s under %s: %v %s", m, d, err, out)
		}
		ms, ds, _ := partyRecord(t, runDir)
		if got := ms[m]; got == nil || got.Ruling != record.AcceptedWord {
			t.Errorf("motion %s reads %+v; want %q", m, got, record.AcceptedWord)
		}
		if ruled, word := motionsPageAnswer(t, runDir, m); !ruled || word != record.AcceptedWord {
			t.Errorf("the motions page reads %s ruled %v %q; want ruled %q", m, ruled, word, record.AcceptedWord)
		}
		if k := ds[d]; k == nil || k.Standing() || k.Ruled() {
			t.Errorf("docketing %s reads %+v after the raiser's concession; want ended, unruled", d, k)
		}
		if p, sd := standingCols(t, runDir, g); p != "" || sd != "" {
			t.Errorf("gap_state reads proposal %q and docketing %q standing after the concession; want none", p, sd)
		}
		conveneBench(t, runDir, g)
		if _, err := answerDocketing(t, runDir, d, "for-author"); err == nil {
			t.Errorf("the bench's ruling of %s after the raiser conceded it was admitted; want refused", d)
		}
	})

	t.Run("the accept names its kind and its proposal", func(t *testing.T) {
		runDir := seatRun(t)
		g := mintGap(t, runDir, "accept-kinds", "spec_underspecification")
		exchangeOn(t, runDir, g)
		out, err := propose(t, runDir, blueSeat, "regrade", g, "--severity", "low")
		if err != nil {
			t.Fatal(err)
		}
		m := filedMotionID(t, out)
		for name, c := range map[string]struct {
			args []string
			want string
		}{
			"a regrade proposal taken by close": {acceptArgs(runDir, "close", m), "which `regrade --accept"},
			"a gap id where a motion is taken":  {acceptArgs(runDir, "regrade", g), "is not motion-id"},
			"a grade beside the accept":         {append(acceptArgs(runDir, "regrade", m), "--severity", "high"), "second source"},
			"an unknown motion":                 {acceptArgs(runDir, "regrade", "M-00000000"), "names no proposal"},
		} {
			if out, err := runAt(t, c.args...); err == nil || !strings.Contains(err.Error()+out, c.want) {
				t.Errorf("%s = %v %q; want refused saying %q", name, err, out, c.want)
			}
		}
		if p, _ := standingCols(t, runDir, g); p != m {
			t.Errorf("after the refused accepts gap_state reads proposal %q standing; want %s", p, m)
		}
	})
}

// T-B49 (a pair): THE AUTHOR ACCEPTS THE RAISER'S STATEMENT. `edit --accept G` writes the stated pair
// with answers = G and accepted; under a docketing it is the author's concession, ending the
// docketing and the author's proposal; under grade movement it is refused, and the raiser's close
// alone concedes it. A raiser's closing stating no text withdraws the pair, and the accept is then
// refused.
func TestTheAuthorAcceptsTheRaisersStatement(t *testing.T) {
	t.Run("edit --accept G under an impasse docketing concedes", func(t *testing.T) {
		runDir, g := prescribedRun(t)
		exchangeOn(t, runDir, g)
		d := docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
		if out, err := runAt(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--accept", g, "--reason", "red is right", "--json"); err != nil {
			t.Fatalf("edit --accept %s under %s: %v %s", g, d, err, out)
		}
		_, ds, evs := partyRecord(t, runDir)
		var accepted int
		for _, e := range evs {
			if ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok && ed.GetAccepted() && ed.GetAnswers() == handleText(t, runDir, g) {
				accepted++
			}
		}
		if accepted != 1 {
			t.Errorf("%d accepted edit(s) answer %s; want 1", accepted, g)
		}
		if k := ds[d]; k == nil || k.Standing() || k.Ruled() {
			t.Errorf("docketing %s reads %+v after blue's concession; want ended, unruled", d, k)
		}
		if _, sd := standingCols(t, runDir, g); sd != "" {
			t.Errorf("gap_state reads docketing %q standing after blue's concession", sd)
		}
	})

	t.Run("refused under grade movement", func(t *testing.T) {
		runDir, g := prescribedRun(t)
		d := docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, "")
		out, err := runAt(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--accept", g, "--reason", "red is right", "--json")
		if err == nil || !strings.Contains(err.Error()+out, "docketed for grade movement") {
			t.Errorf("edit --accept %s under grade movement %s = %v %q; want refused", g, d, err, out)
		}
		if _, sd := standingCols(t, runDir, g); sd != d {
			t.Errorf("gap_state reads docketing %q; want %s standing", sd, d)
		}
	})

	t.Run("a raiser's closing stating no text withdraws the pair", func(t *testing.T) {
		runDir, g := prescribedRun(t)
		exchangeOn(t, runDir, g)
		docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
		asSeat(t, runDir, lensSeat, "register")
		if out, err := runAt(t, "closing", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--reason", "no pair lands", "--no-text"); err != nil {
			t.Fatalf("the raiser's closing stating no text: %v %s", err, out)
		}
		out, err := runAt(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--accept", g, "--reason", "red is right", "--json")
		if err == nil || !strings.Contains(err.Error()+out, "withdrawing every pair") {
			t.Errorf("edit --accept %s after the raiser withdrew its text = %v %q; want refused", g, err, out)
		}
	})

	t.Run("the raiser's close under grade movement concedes", func(t *testing.T) {
		runDir, g := prescribedRun(t)
		d := docketNow(t, runDir, g, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, "")
		asSeat(t, runDir, lensSeat, "register")
		if out, err := runAt(t, "close", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--as", "not_a_defect", "--verified-by", lensSeat,
			"--verified-with", "the report read", "--verified-against", "the report's sentence", "--reason", "it is not a defect"); err != nil {
			t.Fatalf("the raiser's close under grade movement %s: %v %s", d, err, out)
		}
		_, ds, _ := partyRecord(t, runDir)
		if k := ds[d]; k == nil || k.Standing() || k.Ruled() {
			t.Errorf("docketing %s reads %+v after the raiser's close; want ended, unruled", d, k)
		}
	})

	t.Run("--accept and --answers together are refused", func(t *testing.T) {
		runDir, g := prescribedRun(t)
		out, err := runAt(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--accept", g, "--answers", g, "--reason", "r", "--json")
		if err == nil || !strings.Contains(err.Error()+out, "second source") {
			t.Errorf("edit --accept with --answers = %v %q; want refused", err, out)
		}
	})
}

// prescribedRun is a run whose report holds acceptS, and a gap on it prescribing acceptN.
func prescribedRun(t *testing.T) (runDir, gap string) {
	t.Helper()
	runDir = acceptRun(t)
	gap, err := mintQuote(t, runDir, "P", acceptS, "--new", acceptN)
	if err != nil {
		t.Fatalf("mint a prescribed gap: %v", err)
	}
	return runDir, gap
}
