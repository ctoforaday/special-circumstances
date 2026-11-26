package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A RULING IS FINAL FOR WHAT IT DECIDED (#1695 §III.B12, §III.B14; T-B19, T-B36), driven through the
// command tree as the seats run it: a proposal filed, its window ended, the dispatch's docketing, the
// bench's sitting and its `motion docket rule`, and every act under test by its own verb.

// quoteBoardBelowImpasse is quoteImpasseBoard with k-max raised past its two exchanges: a proposal on
// the gap is docketed for its window, never at impasse.
func quoteBoardBelowImpasse(t *testing.T, runDir string) {
	t.Helper()
	quoteImpasseBoard(t, runDir)
	writeTerms(t, runDir, 4, 0)
}

// blueProposes is blue's proposal of act on boardGap (its flags after the verb's), returning the motion.
func blueProposes(t *testing.T, runDir, act string, flags ...string) string {
	t.Helper()
	asSeat(t, runDir, blueSeat, "register")
	args := append([]string{"motion", "docket", "file", act, "--id", boardGap, "--json", "--run", runDir, "--seat-id", blueSeat}, flags...)
	out, err := runAt(t, args...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runAt(t, args...)
	}
	if err != nil {
		t.Fatalf("blue's %s proposal on %s: %v\n%s", act, boardGap, err, out)
	}
	return filedMotionID(t, out)
}

// benchAnswers convenes the bench on boardGap's standing docketing, as the dispatch does, and answers it.
func benchAnswers(t *testing.T, runDir, docketing, answer string, extra ...string) {
	t.Helper()
	partiesState(t, runDir)
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	asSeat(t, runDir, "judge", "register", "--occasion", "docket")
	if out, err := answerDocketing(t, runDir, docketing, answer, extra...); err != nil {
		t.Fatalf("--%s on %s: %v\n%s", answer, docketing, err, out)
	}
}

// partiesState files the closings a docketing on boardGap owes and are not yet in: blue's, and the
// assignee's stating no text where it filed none.
func partiesState(t *testing.T, runDir string) {
	t.Helper()
	for _, c := range []struct {
		seat string
		args []string
	}{
		{blueSeat, []string{"closing", "--id", boardGap, "--reason", "blue's position"}},
		{"red-lens-evidence", []string{"closing", "--id", boardGap, "--reason", "red's position", "--no-text"}},
	} {
		asSeat(t, runDir, c.seat, "register")
		out, err := runAt(t, append(c.args, "--run", runDir, "--seat-id", c.seat)...)
		if err != nil && !strings.Contains(err.Error()+out, "already") && !strings.Contains(err.Error()+out, "not owed") {
			t.Logf("%s's closing: %v %.200s", c.seat, err, out)
		}
	}
}

// rewordRuled is blue's reword of ruledSentence as the report now holds it, carrying its tokens: the
// sentence's reopened lists every gap whose token stands on it.
func rewordRuled(t *testing.T, runDir, key, from, to string) {
	t.Helper()
	asSeat(t, runDir, blueSeat, "register")
	md := renders(t, runDir)
	base := strings.TrimSuffix(from, ".")
	at := strings.Index(md, base)
	if at < 0 {
		t.Fatalf("%q is not in the report:\n%s", base, md)
	}
	at += len(base)
	tokens := md[at:anchor.SkipRun(md, at)]
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key, "--quote", base+tokens+".",
		"--new", strings.TrimSuffix(to, ".")+tokens+".", "--reason", "blue's wording"); err != nil {
		t.Fatalf("blue's reword %s: %v %s", key, err, out)
	}
}

// standingRuling is the docketing of the ruling tombstone standing on gap, "" for none.
func standingRuling(t *testing.T, runDir, gap string) string {
	t.Helper()
	w, err := record.WorkJSONOfRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range w.Estopped {
		if e.ID == gap {
			return e.Ruling.Docketing
		}
	}
	return ""
}

// lensMovesSeverity is the assignee's regrade of boardGap's severity to grade, in a sitting of its own.
func lensMovesSeverity(t *testing.T, runDir, grade string) (string, error) {
	t.Helper()
	asSeat(t, runDir, "red-lens-evidence", "register")
	return runAt(t, "regrade", "--run", runDir, "--seat-id", "red-lens-evidence", "--id", boardGap, "--severity", grade, "--reason", "the lens's move")
}

// T-B36 `TestNoLaterDocketingOverturnsARuling`, the rows B3's rows let the verbs reach.
func TestNoLaterDocketingOverturnsARuling(t *testing.T) {
	const r2Reword = "The index rose by nine points in the producers' spring survey."
	const r3Reword = "The index rose by nine points in the producers' spring survey, revised."

	// R1 → R2 (amendment 6; the lock span 2026-11-21): R1 for blue on blue's regrade proposal on
	// severity writes it, locks it and places its tombstone; the assignee's regrade of severity is
	// refused by the lock naming R1; blue's reword ends the tombstone and the lock with it, and the
	// regrade lands; blue proposes severity again, the window ends, R2 writes the new grade and re-locks
	// it; the assignee's regrade is refused naming R2 while R2's tombstone stands, no grade-movement
	// docketing follows, and after blue's next reword it lands. Run with R2 for red too.
	for _, r2 := range []string{"for-blue", "for-red"} {
		t.Run("R1 then R2 "+r2, func(t *testing.T) {
			runDir := newRun(t)
			quoteBoardBelowImpasse(t, runDir)
			blueProposes(t, runDir, "regrade", "--severity", "low", "--reason", "the harm needs two failures")
			d1 := docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap).GetDocketingId()
			benchAnswers(t, runDir, d1, "for-blue")
			if g := gapOf(t, runDir, boardGap); recordpb.Word(g.Severity) != "low" || !g.Open {
				t.Fatalf("after R1 the gap is open %v at severity %v, want open at low", g.Open, g.Severity)
			}
			if got := standingRuling(t, runDir, boardGap); got != d1 {
				t.Fatalf("after R1 the standing ruling tombstone names %q, want %s", got, d1)
			}
			out, err := lensMovesSeverity(t, runDir, "certain")
			heldRefusal(t, "the assignee's regrade of the axis R1 locked", out, err, "ruling at event")
			rewordRuled(t, runDir, "RW1", ruledSentence, r2Reword)
			if got := standingRuling(t, runDir, boardGap); got != "" {
				t.Fatalf("after blue's reword the ruling tombstone %s still stands", got)
			}
			if out, err := lensMovesSeverity(t, runDir, "medium"); err != nil {
				t.Fatalf("the assignee's regrade after the reword ended R1's tombstone and lock: %v\n%s", err, out)
			}
			blueProposes(t, runDir, "regrade", "--severity", "low", "--reason", "the harm needs two failures, again")
			d2 := docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap).GetDocketingId()
			benchAnswers(t, runDir, d2, r2)
			want := map[string]string{"for-blue": "low", "for-red": "medium"}[r2]
			if g := gapOf(t, runDir, boardGap); recordpb.Word(g.Severity) != want {
				t.Fatalf("after R2 %s the severity = %v, want %s", r2, g.Severity, want)
			}
			if got := standingRuling(t, runDir, boardGap); got != d2 {
				t.Fatalf("after R2 the standing ruling tombstone names %q, want %s", got, d2)
			}
			out, err = lensMovesSeverity(t, runDir, "certain")
			heldRefusal(t, "the assignee's regrade of the axis R2 re-locked", out, err, "ruling at event")
			asSeat(t, runDir, "red-chair", "register")
			asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
			for _, d := range func() []*recordpb.Docketing { ds, _ := docketingsOn(t, runDir, boardGap); return ds }() {
				if d.GetCause() == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT {
					t.Errorf("a grade-movement docketing %s followed R2 on an axis no seat moved", d.GetDocketingId())
				}
			}
			rewordRuled(t, runDir, "RW2", r2Reword, r3Reword)
			if out, err := lensMovesSeverity(t, runDir, "certain"); err != nil {
				t.Errorf("the assignee's regrade after blue's next reword: %v\n%s", err, out)
			}
		})
	}

	// --keep and --restore each place a ruling tombstone, so a proposal on G after either is refused by it.
	for _, answer := range []string{"keep", "restore"} {
		t.Run(answer+" places a ruling tombstone", func(t *testing.T) {
			runDir := newRun(t)
			quoteBoardBelowImpasse(t, runDir)
			for _, g := range []string{"certain", "low"} {
				if out, err := lensMovesSeverity(t, runDir, g); err != nil {
					t.Fatalf("the lens's move to %s: %v\n%s", g, err, out)
				}
			}
			asSeat(t, runDir, "red-chair", "register")
			asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
			d := lastDocketing(t, runDir, boardGap)
			if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT {
				t.Fatalf("the plan after the lens's moves docketed %s, want grade movement", recordpb.Word(d.GetCause()))
			}
			benchAnswers(t, runDir, d.GetDocketingId(), answer)
			if got := standingRuling(t, runDir, boardGap); got != d.GetDocketingId() {
				t.Fatalf("after --%s the standing ruling tombstone names %q, want %s", answer, got, d.GetDocketingId())
			}
			asSeat(t, runDir, blueSeat, "register")
			actRefusedNaming(t, runDir, "blue's proposal on G after --"+answer, blueSeat, []string{"docketing " + d.GetDocketingId(), "tombstone"},
				"motion", "docket", "file", "regrade", "--id", boardGap, "--severity", "medium", "--reason", "the harm is ordinary")
		})
	}

	// Audit p8's trace: blue proposes close W, red's closing missing — --for-red refused naming the
	// missing text; with red's text, --for-red closes G repaired, and no later plan dockets G.
	t.Run("audit p8: blue's close, for red", func(t *testing.T) {
		runDir := newRun(t)
		quoteBoardBelowImpasse(t, runDir)
		blueProposes(t, runDir, "close", append([]string{"--reason", "the survey is the source"}, closeFlags("not_a_defect")...)...)
		d := docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap).GetDocketingId()
		asSeat(t, runDir, "red-chair", "register")
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		asSeat(t, runDir, "judge", "register", "--occasion", "docket")
		out, err := answerDocketing(t, runDir, d, "for-red", "--final")
		heldRefusal(t, "--for-red with red's closing missing", out, err, "filed no closing")
		asSeat(t, runDir, "red-lens-evidence", "register")
		asSeat(t, runDir, "red-lens-evidence", "closing", "--id", boardGap, "--reason", "the sentence still misleads",
			"--quote", ruledSentence, "--new", redRoad)
		benchAnswers(t, runDir, d, "for-red", "--final")
		r := lastRuling(t, runDir)
		if r.GetSidedWith() != recordpb.SidedWith_SIDED_WITH_RED || r.GetDisposition() != recordpb.Disposition_DISPOSITION_REPAIRED || gapOf(t, runDir, boardGap).Open {
			t.Fatalf("the ruling = %v %v, want for red, repaired, G closed", r.GetSidedWith(), r.GetDisposition())
		}
		before, _ := docketingsOn(t, runDir, boardGap)
		asSeat(t, runDir, "red-chair", "register")
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		if after, _ := docketingsOn(t, runDir, boardGap); len(after) != len(before) {
			t.Errorf("a plan after the ruling docketed the closed G again: %d → %d", len(before), len(after))
		}
	})
}

// T-B19 (11): on an open G ruled for red on blue's regrade proposal, a grade-movement docketing on an
// axis the proposal did not name lands through `dispatch next`, and the chair's own regrade proposal
// on G is refused naming the tombstone.
func TestARulingProtectsTheGapItLeavesStanding(t *testing.T) {
	t.Run("(11) an open G ruled for red", func(t *testing.T) {
		runDir := newRun(t)
		quoteBoardBelowImpasse(t, runDir)
		blueProposes(t, runDir, "regrade", "--severity", "low", "--reason", "the harm needs two failures")
		d := docketsUnaccepted(t, runDir, "red-lens-evidence", boardGap).GetDocketingId()
		benchAnswers(t, runDir, d, "for-red")
		if !gapOf(t, runDir, boardGap).Open || standingRuling(t, runDir, boardGap) != d {
			t.Fatalf("after the for-red ruling the gap is open %v under %q, want open under %s", gapOf(t, runDir, boardGap).Open, standingRuling(t, runDir, boardGap), d)
		}
		actRefusedNaming(t, runDir, "the chair's own regrade proposal on G", "red-chair", []string{"docketing " + d, "tombstone"},
			"motion", "docket", "file", "regrade", "--id", boardGap, "--impact", "low", "--reason", "the chair's view")
		asSeat(t, runDir, "red-lens-evidence", "register")
		for _, g := range []string{"certain", "low"} {
			asSeat(t, runDir, "red-lens-evidence", "register")
			if out, err := runAt(t, "regrade", "--run", runDir, "--seat-id", "red-lens-evidence", "--id", boardGap, "--impact", g, "--reason", "the lens's move"); err != nil {
				t.Fatalf("the lens's impact move to %s on an axis the ruling did not decide: %v\n%s", g, err, out)
			}
		}
		asSeat(t, runDir, "red-chair", "register")
		asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
		if got := lastDocketing(t, runDir, boardGap); got.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT || got.GetDocketingId() == d {
			t.Errorf("the plan after the impact moves wrote %+v, want a grade-movement docketing on the ruled G", got)
		}
	})
}

// aboutMintArgs is the lens's mint flags of class about subject (--about-kind and --about).
func aboutMintArgs(runDir, seat, class, kind, ref string) []string {
	return []string{"--run", runDir, "--seat-id", seat, "--about-kind", kind, "--about", ref, "--class", class,
		"--problem", "the defect", "--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low"}
}

// mintArgs is runMint of args, asked once more where the gap moved under the seat's sitting.
func mintArgs(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, err := runMint(t, args...)
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runMint(t, args...)
	}
	return out, err
}

// ruledForBlueAtImpasse takes gap, the lens's, to impasse through the dispatch (k-max 1, one exchange),
// files both closings — blue's not_a_defect, the lens's stating no text where noText (a gap about an
// inquiry takes no flag for it) — and has the bench, convened
// by the dispatch, rule for blue. It returns the docketing.
func ruledForBlueAtImpasse(t *testing.T, runDir, gap string, noText bool) string {
	t.Helper()
	writeTerms(t, runDir, 1, 0)
	exchangeOn(t, runDir, gap)
	exchangeOn(t, runDir, gap)
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	d := lastDocketing(t, runDir, gap)
	if d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE {
		t.Fatalf("the plan at k-max wrote %+v, want an impasse docketing", d)
	}
	asSeat(t, runDir, blueSeat, "register")
	asSeat(t, runDir, blueSeat, "closing", "--id", gap, "--as", "not_a_defect", "--reason", "the source holds")
	asSeat(t, runDir, lensSeat, "register")
	red := []string{"closing", "--id", gap, "--reason", "it still misleads"}
	if noText {
		red = append(red, "--no-text")
	}
	asSeat(t, runDir, lensSeat, red...)
	asSeat(t, runDir, "red-chair", "register")
	asSeat(t, runDir, "red-chair", "dispatch", "next", "--json")
	asSeat(t, runDir, "judge", "register", "--occasion", "docket")
	if out, err := answerDocketing(t, runDir, d.GetDocketingId(), "for-blue", "--final"); err != nil {
		t.Fatalf("--for-blue: %v\n%s", err, out)
	}
	return d.GetDocketingId()
}

// T-B19 (5), (8), (10), (15): what ends a ruling tombstone follows the ruled gap's anchor (L17).
func TestARulingTombstoneEndsWithItsAnchor(t *testing.T) {
	const class = "spec_underspecification"

	// (5) A CITATION PLACED AS G'S ANSWER ENDS NOTHING.
	t.Run("(5) a citation answering G", func(t *testing.T) {
		runDir := newRun(t)
		quoteImpasseBoard(t, runDir)
		withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://survey/1": []byte("<html>the survey</html>")}})
		d := ruleAtImpasse(t, runDir, false)
		asSeat(t, runDir, blueSeat, "register")
		asSeat(t, runDir, blueSeat, "cite", "--quote", strings.TrimSuffix(ruledSentence, "."), "--url", "https://survey/1",
			"--title", "Survey", "--answers", boardGap, "--reason", "the survey is the source")
		if rows, _ := tombstoneRows(t, runDir); len(rows) != 1 || standingRuling(t, runDir, boardGap) != d {
			t.Errorf("the citation answering G wrote rows or ended G's tombstone: %v", rows)
		}
		actRefusedNaming(t, runDir, "a proposal on G after the citation", otherLens, []string{"docketing " + d, "tombstone"},
			"motion", "docket", "file", "regrade", "--id", boardGap, "--severity", "low", "--reason", "the harm needs two failures")
	})

	// (8) AN ANCHOR GAP ABOUT CITATION C: a mint about C of the class, and a --quote mint of the class
	// on C's sentence (the cross-arm), refused; an edit whose reopened lists C ends it; on a second run,
	// blue's retire --anchor C ends it. (10) A seat's owed failed check on the ruled gap's anchor still
	// admits its gap.
	citedRuling := func(t *testing.T) (runDir, c, gap, d string) {
		runDir, c = citedClause(t)
		seedCastWith(t, runDir, otherLens)
		out, err := mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "anchor", c)...)
		if err != nil {
			t.Fatalf("the lens's gap about %s: %v\n%s", c, err, out)
		}
		gap = gapID(out)
		return runDir, c, gap, ruledForBlueAtImpasse(t, runDir, gap, true)
	}
	t.Run("(8) about a citation: a reword of its sentence ends it", func(t *testing.T) {
		runDir, c, gap, d := citedRuling(t)
		names := []string{"the bench ruled gap " + gap, "docketing " + d, "tombstone"}
		asSeat(t, runDir, lensSeat, "register")
		out, err := mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "anchor", c)...)
		heldRefusal(t, "a mint of the class about the ruled citation", out, err, names...)
		out, err = tombstoneMint(t, runDir, lensSeat, class, "the sky is blue")
		heldRefusal(t, "a --quote mint of the class on the cited sentence", out, err, names...)
		asSeat(t, runDir, blueSeat, "register")
		edit := []string{"edit", "--run", runDir, "--seat-id", blueSeat, "--key", "RW", "--quote", "The data shows the sky is blue",
			"--new", "The survey data shows the sky is blue", "--reason", "blue's wording"}
		if out, err := run(t, edit...); err != nil {
			t.Fatalf("blue's reword of the cited sentence: %v\n%s", err, out)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetReopened(); !slices.Contains(got, c) {
			t.Fatalf("the reword's reopened %v does not list %s", got, c)
		}
		if standingRuling(t, runDir, gap) != "" {
			t.Fatalf("the reword of %s's sentence left the ruling tombstone on %s standing", c, gap)
		}
		asSeat(t, runDir, lensSeat, "register")
		if out, err := mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "anchor", c)...); err != nil {
			t.Errorf("a mint of the class about %s after the end: %v\n%s", c, err, out)
		}
	})
	t.Run("(8) about a citation: its retire ends it", func(t *testing.T) {
		runDir, c, gap, _ := citedRuling(t)
		asSeat(t, runDir, blueSeat, "register")
		asSeat(t, runDir, blueSeat, "retire", "--anchor", c, "--reason", "the source does not hold")
		if standingRuling(t, runDir, gap) != "" {
			t.Fatalf("the retire of %s left the ruling tombstone on %s standing", c, gap)
		}
		if rows, _ := tombstoneRows(t, runDir); len(rows) != 2 || !rows[1].GetEnds() {
			t.Errorf("tombstone rows %v, want the ruling's and the retire's end row", rows)
		}
	})
	t.Run("(8) about a computation: a mint about it refused, its retire ends it", func(t *testing.T) {
		runDir := newRun(t)
		const sentence = "Nine is composite, so the protocol rejects a false claim."
		seedCastWith(t, runDir, "red-lens-computation")
		seat := computeSeat(t, runDir, "# H\n\n"+sentence+"\n")
		dir := script(t, runDir, "nine.js", "let d=[];for(let i=2;i<9;i++)if(9%i===0)d.push(i);console.log('divisors:',d.join(','));")
		if out, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", sentence, "--dir", dir,
			"--reason", "trial division settles it", "--", "node", "nine.js"); err != nil {
			t.Fatalf("compute: %v\n%s", err, out)
		}
		k := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
		registerLensOnce(t, runDir)
		out, err := mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "anchor", k)...)
		if err != nil {
			t.Fatalf("the lens's gap about %s: %v\n%s", k, err, out)
		}
		gap := gapID(out)
		d := ruledForBlueAtImpasse(t, runDir, gap, true)
		names := []string{"the bench ruled gap " + gap, "docketing " + d, "tombstone"}
		asSeat(t, runDir, lensSeat, "register")
		out, err = mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "anchor", k)...)
		heldRefusal(t, "a mint of the class about the ruled computation", out, err, names...)
		// A computation's token pairs with nothing (onegapperclass.go): the cross-arm is a citation's.
		if out, err := tombstoneMint(t, runDir, lensSeat, class, sentence); err != nil {
			t.Errorf("a --quote mint of the class on the computed sentence, which no cross-arm reaches: %v\n%s", err, out)
		}
		asSeat(t, runDir, blueSeat, "register")
		asSeat(t, runDir, blueSeat, "retire", "--anchor", k, "--reason", "the script is superseded")
		if standingRuling(t, runDir, gap) != "" {
			t.Fatalf("the retire of %s left the ruling tombstone on %s standing", k, gap)
		}
	})
	t.Run("(10) an owed failed check on the ruled gap's anchor admits its gap", func(t *testing.T) {
		runDir, c, _, _ := citedRuling(t)
		asSeat(t, runDir, otherLens, "register")
		if _, err := runAt(t, "verify", "--run", runDir, "--seat-id", otherLens, "--anchor", c, "--as", "unreachable", "--confidence", "high", "--reason", "the page will not load"); err != nil {
			t.Fatalf("%s's reading of %s: %v", otherLens, c, err)
		}
		if out, err := mintArgs(t, aboutMintArgs(runDir, otherLens, class, "anchor", c)...); err != nil {
			t.Errorf("the gap %s's failed check owes, about the ruled gap's anchor: %v\n%s", otherLens, err, out)
		}
	})

	// (15) AN INQUIRY GAP ABOUT Q: a mint about Q of the class and a proposal on the ruled gap refused;
	// blue's next fate on Q ends it, then both land; the estopped row names Q.
	t.Run("(15) about an inquiry", func(t *testing.T) {
		runDir := oneGapRun(t)
		q := proposeInquiry(t, runDir, "whether costs rose")
		out, err := mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "inquiry", q)...)
		if err != nil {
			t.Fatalf("the lens's gap about %s: %v\n%s", q, err, out)
		}
		gap := gapID(out)
		d := ruledForBlueAtImpasse(t, runDir, gap, false)
		w, err := record.WorkJSONOfRun(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		if len(w.Estopped) != 1 || w.Estopped[0].ID != gap || !strings.Contains(w.Estopped[0].Ends, q) {
			t.Errorf("estopped = %+v, want %s ended by %s's next fate", w.Estopped, gap, q)
		}
		names := []string{"the bench ruled gap " + gap, "docketing " + d, "tombstone"}
		asSeat(t, runDir, lensSeat, "register")
		out, err = mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "inquiry", q)...)
		heldRefusal(t, "a mint of the class about the ruled inquiry", out, err, names...)
		proposal := []string{"motion", "docket", "file", "regrade", "--id", gap, "--severity", "low", "--reason", "the harm needs two failures"}
		asSeat(t, runDir, blueSeat, "register")
		actRefusedNaming(t, runDir, "a proposal on the ruled inquiry gap", blueSeat, names, proposal...)
		asSeat(t, runDir, blueSeat, "inquiry", "record", "--id", q, "--as", "confirmed", "--reason", "costs rose", "--method", "the ledger read")
		if standingRuling(t, runDir, gap) != "" {
			t.Fatalf("%s's next fate left the ruling tombstone on %s standing", q, gap)
		}
		asSeat(t, runDir, lensSeat, "register")
		if out, err := mintArgs(t, aboutMintArgs(runDir, lensSeat, class, "inquiry", q)...); err != nil {
			t.Errorf("a mint of the class about %s after its fate: %v\n%s", q, err, out)
		}
	})
}
