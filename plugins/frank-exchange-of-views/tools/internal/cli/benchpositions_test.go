package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE BENCH CHOOSES BETWEEN THE PARTIES' POSITIONS (#1695 §III.B14). Each docketing puts two positions
// before the bench, each stated by its party in the closing it owes there; the bench names a side,
// and the tool applies that side's statement in the ruling's transaction.

// docketNow writes the tool's docketing of gap — cause, and the proposal a PROPOSAL docketing holds —
// as the dispatch writes it, and returns its id.
func docketNow(t *testing.T, runDir, gap string, cause recordpb.DocketingCause, motion string) string {
	t.Helper()
	ensureCast(t, runDir)
	d := &recordpb.Docketing{DocketingId: proto.String(record.NewID("docketing")), GapId: proto.String(handleText(t, runDir, gap)),
		Cause: cause.Enum(), Basis: proto.String("before the bench")}
	if motion != "" {
		d.MotionId = proto.String(motion)
	}
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"}, d); err != nil {
		t.Fatalf("docketing %s: %v", gap, err)
	}
	return d.GetDocketingId()
}

// answerDocketing is the bench's `motion docket rule` of docketing with answer (a flag, without its
// dashes) and the extra flags given, in the bench's sitting already open; it returns the refusal.
func answerDocketing(t *testing.T, runDir, docketing, answer string, extra ...string) (string, error) {
	t.Helper()
	args := append([]string{"motion", "docket", "rule", "--run", runDir, "--seat-id", "judge", "--id", docketing, "--" + answer,
		"--principle", "correctness", "--tension", "thoroughness against ceremony",
		"--review-flag", "no — the ruling is mechanical", "--settled", "the proposition this ruling bars",
		"--reason", "the ruling as reasoned"}, extra...)
	return runAt(t, args...)
}

// positionsRun is acceptRun's report with a quote gap on acceptS, minted by lensSeat, at impasse —
// docketed — with the bench convened on it. It returns the gap and the docketing.
func positionsRun(t *testing.T) (runDir, gap, docketing string) {
	t.Helper()
	runDir = acceptRun(t)
	gap, err := mintQuote(t, runDir, "G", acceptS+".")
	if err != nil {
		t.Fatal(err)
	}
	docketing = docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	return runDir, gap, docketing
}

// redStates files the lens's closing on gap stating red's text: acceptS becomes acceptN.
func redStates(t *testing.T, runDir, gap string, extra ...string) (string, error) {
	t.Helper()
	if len(extra) == 0 {
		extra = []string{"--quote", acceptS + ".", "--new", acceptN + "."}
	}
	return runAt(t, append([]string{"closing", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--reason", "red's case"}, extra...)...)
}

// blueStates files blue's closing on gap, stating word where one is given.
func blueStates(t *testing.T, runDir, gap, word string) (string, error) {
	t.Helper()
	args := []string{"closing", "--run", runDir, "--seat-id", blueSeat, "--id", gap, "--reason", "blue's case"}
	if word != "" {
		args = append(args, "--as", word)
	}
	return runAt(t, args...)
}

// lastRuling is the newest docket ruling on the record.
func lastRuling(t *testing.T, runDir string) *recordpb.DocketRuling {
	t.Helper()
	return lastBody(t, runDir, &recordpb.MotionRule{}).GetDocket()
}

// benchActs is every act the bench wrote of the given type.
func benchActs(t *testing.T, runDir string, typ recordpb.EventType) []*record.Event {
	t.Helper()
	evs, _, err := record.EventsOf(runtest.Open(t, runDir), typ)
	if err != nil {
		t.Fatal(err)
	}
	var out []*record.Event
	for _, e := range evs {
		if e.GetSeatId() == "judge" {
			out = append(out, e)
		}
	}
	return out
}

func TestTheBenchChoosesBetweenThePartiesPositions(t *testing.T) {
	t.Run("impasse, for red: red's text applied as the bench's edit, then repaired", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing with a planned pair: %v\n%s", err, out)
		}
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "for-red", "--final"); err != nil {
			t.Fatalf("--for-red: %v\n%s", err, out)
		}
		r := lastRuling(t, runDir)
		if r.GetSidedWith() != recordpb.SidedWith_SIDED_WITH_RED || r.GetDisposition() != recordpb.Disposition_DISPOSITION_REPAIRED {
			t.Errorf("the ruling = %v %v, want for red, repaired", r.GetSidedWith(), r.GetDisposition())
		}
		edits := benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
		if len(edits) != 1 {
			t.Fatalf("the bench wrote %d edits, want red's text as one", len(edits))
		}
		if e, _ := recordpb.BodyAs[*recordpb.BlueEdit](edits[0]); e.GetAnswers() != gap || !e.GetAppliedVerbatim() {
			t.Errorf("the bench's edit = %+v, want red's text answering %s verbatim", e, gap)
		}
		if rep := readReport(t, runDir); !strings.Contains(rep, acceptN) || strings.Contains(rep, acceptS+".") {
			t.Errorf("the report does not carry red's text:\n%s", rep)
		}
		if g := gapOf(t, runDir, gap); g == nil || g.Open {
			t.Errorf("gap %s is open after a ruling for red", gap)
		}
	})

	for _, word := range []string{"", "defect_accepted", "moot"} {
		t.Run("impasse, for blue, blue states "+word, func(t *testing.T) {
			runDir, gap, d := positionsRun(t)
			if _, err := blueStates(t, runDir, gap, word); err != nil {
				t.Fatalf("blue's closing: %v", err)
			}
			before := readReport(t, runDir)
			conveneBench(t, runDir, gap)
			if out, err := answerDocketing(t, runDir, d, "for-blue", "--final"); err != nil {
				t.Fatalf("--for-blue: %v\n%s", err, out)
			}
			want := word
			if want == "" {
				want = "not_a_defect"
			}
			if r := lastRuling(t, runDir); r.GetSidedWith() != recordpb.SidedWith_SIDED_WITH_BLUE || recordpb.Word(r.GetDisposition()) != want {
				t.Errorf("the ruling = %v %v, want for blue, %s", r.GetSidedWith(), r.GetDisposition(), want)
			}
			if readReport(t, runDir) != before {
				t.Error("a ruling for blue changed the report")
			}
			if g := gapOf(t, runDir, gap); g == nil || g.Open {
				t.Errorf("gap %s is open after a ruling for blue", gap)
			}
		})
	}

	t.Run("impasse, neither side filed: for red refused naming the missing text, for blue lands not_a_defect", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		conveneBench(t, runDir, gap)
		out, err := answerDocketing(t, runDir, d, "for-red", "--final")
		if err == nil || !strings.Contains(err.Error()+out, "filed no closing") {
			t.Fatalf("--for-red with no red closing: %v\n%s", err, out)
		}
		if out, err := answerDocketing(t, runDir, d, "for-blue", "--final"); err != nil {
			t.Fatalf("--for-blue: %v\n%s", err, out)
		}
		if r := lastRuling(t, runDir); r.GetDisposition() != recordpb.Disposition_DISPOSITION_NOT_A_DEFECT {
			t.Errorf("the ruling's word = %v, want not_a_defect", r.GetDisposition())
		}
	})

	t.Run("impasse: a flag outside the pair, two answers, and no answer are refused naming the pair", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		conveneBench(t, runDir, gap)
		for _, args := range [][]string{{"for-filer"}, {"keep"}, {"for-red", "--for-blue"}} {
			out, err := answerDocketing(t, runDir, d, args[0], append(args[1:], "--final")...)
			if err == nil || !strings.Contains(err.Error()+out, "--for-red | --for-blue") {
				t.Errorf("%v: %v\n%s", args, err, out)
			}
		}
		for _, gone := range []string{"--as", "--accept", "--reject"} {
			out, err := runAt(t, "motion", "docket", "rule", "--run", runDir, "--seat-id", "judge", "--id", d, gone, "repaired", "--reason", "r")
			if err == nil || !strings.Contains(err.Error()+out, "unknown flag") {
				t.Errorf("%s on the docket ruling: %v\n%s", gone, err, out)
			}
		}
	})

	t.Run("impasse, remand: remanded, no side", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "remand", "--direction", "a reproduction"); err != nil {
			t.Fatalf("--remand: %v\n%s", err, out)
		}
		if r := lastRuling(t, runDir); r.GetDisposition() != recordpb.Disposition_DISPOSITION_REMANDED || r.SidedWith != nil || r.MovedGrades != nil {
			t.Errorf("the remand = %+v", r)
		}
	})

	t.Run("blue's close proposal: for blue writes its word, for red applies red's text", func(t *testing.T) {
		for _, side := range []string{"for-blue", "for-red"} {
			runDir := acceptRun(t)
			gap, err := mintQuote(t, runDir, "G", acceptS+".")
			if err != nil {
				t.Fatal(err)
			}
			exchangeOn(t, runDir, gap)
			m := proposeClose(t, runDir, blueSeat, gap, "defect_owed_elsewhere", "the engine owns this")
			if _, err := redStates(t, runDir, gap); err != nil {
				t.Fatalf("red's closing in the window: %v", err)
			}
			d := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
			conveneBench(t, runDir, gap)
			if out, err := answerDocketing(t, runDir, d, side, "--final"); err != nil {
				t.Fatalf("%s: %v\n%s", side, err, out)
			}
			r := lastRuling(t, runDir)
			want := map[string]recordpb.Disposition{"for-blue": recordpb.Disposition_DISPOSITION_DEFECT_OWED_ELSEWHERE, "for-red": recordpb.Disposition_DISPOSITION_REPAIRED}[side]
			if r.GetDisposition() != want {
				t.Errorf("%s: the word = %v, want %v", side, r.GetDisposition(), want)
			}
			if n := len(benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)); (side == "for-red") != (n == 1) {
				t.Errorf("%s: the bench wrote %d edits", side, n)
			}
		}
	})

	t.Run("blue's regrade proposal: for blue writes the stored grades, for red none; both lock the axis", func(t *testing.T) {
		for _, side := range []string{"for-blue", "for-red"} {
			runDir := acceptRun(t)
			gap, err := mintQuote(t, runDir, "G", acceptS+".")
			if err != nil {
				t.Fatal(err)
			}
			exchangeOn(t, runDir, gap)
			if err := proposeRegradeOf(t, runDir, blueSeat, gap, "low"); err != nil {
				t.Fatal(err)
			}
			m := lastBody(t, runDir, &recordpb.Motion{}).GetMotionId()
			d := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
			conveneBench(t, runDir, gap)
			if out, err := answerDocketing(t, runDir, d, side); err != nil {
				t.Fatalf("%s: %v\n%s", side, err, out)
			}
			want := map[string]string{"for-blue": "low", "for-red": "medium"}[side]
			if g := gapOf(t, runDir, gap); g == nil || !g.Open || recordpb.Word(g.Severity) != want {
				t.Errorf("%s: the gap = %+v, want open at severity %s", side, g, want)
			}
			if n := len(benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE)); (side == "for-blue") != (n == 1) {
				t.Errorf("%s: the bench wrote %d regrades", side, n)
			}
			out, err := lensRegrade(t, runDir, gap, "severity", "high")
			if err == nil || !strings.Contains(err.Error()+out, "ruling at event") {
				t.Errorf("%s: the lens moved a ruled axis: %v\n%s", side, err, out)
			}
		}
	})
}

func TestTheBenchHasNoGradeOfItsOwn(t *testing.T) {
	runDir, gap, d := positionsRun(t)
	conveneBench(t, runDir, gap)
	if out, err := runAt(t, "regrade", "--run", runDir, "--seat-id", "judge", "--id", gap, "--severity", "low", "--reason", "r"); err == nil {
		t.Fatalf("regrade is on the bench's surface: %s", out)
	}
	// A bench regrade appended below the verbs, before a ruling that applies none, is refused at
	// that ruling.
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"},
		&recordpb.Regrade{GapId: proto.String(gap), Severity: recordpb.Grade_GRADE_LOW.Enum(), Basis: proto.String("raw")}); err != nil {
		t.Fatalf("the raw bench regrade: %v", err)
	}
	out, err := answerDocketing(t, runDir, d, "for-blue", "--final")
	if err == nil || !strings.Contains(err.Error()+out, "no grade of its own") {
		t.Fatalf("a ruling after a bench regrade it applies none of: %v\n%s", err, out)
	}
}

func TestEachSideStatesItsPositionAtTheDocketing(t *testing.T) {
	t.Run("blue's word: each of the five lands, the two no ruling carries are refused", func(t *testing.T) {
		for _, word := range []string{"repaired", "not_a_defect", "defect_accepted", "defect_owed_elsewhere", "moot", "repaired_with_regression", "amends_prior"} {
			runDir, gap, _ := positionsRun(t)
			_, err := blueStates(t, runDir, gap, word)
			if refused := word == "repaired_with_regression" || word == "amends_prior"; refused != (err != nil) {
				t.Errorf("blue's closing --as %s: %v", word, err)
			}
		}
	})
	t.Run("a word on red's closing, and on blue's at a proposal, is refused", func(t *testing.T) {
		runDir, gap, _ := positionsRun(t)
		if _, err := runAt(t, "closing", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--reason", "r", "--as", "moot"); err == nil {
			t.Error("red's closing took a word")
		}
		runDir = acceptRun(t)
		gap, _ = mintQuote(t, runDir, "G", acceptS+".")
		exchangeOn(t, runDir, gap)
		m := proposeClose(t, runDir, blueSeat, gap, "moot", "moot")
		docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
		if out, err := blueStates(t, runDir, gap, "moot"); err == nil || !strings.Contains(err.Error()+out, "impasse") {
			t.Errorf("blue's word on its own proposal's docketing: %v\n%s", err, out)
		}
	})
	t.Run("red's closing at impasse states its text: no road refused, --no-text lands, a road with --no-text refused", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := runAt(t, "closing", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--reason", "r"); err == nil || !strings.Contains(err.Error()+out, "--no-text") {
			t.Errorf("a road-less red closing: %v\n%s", err, out)
		}
		if out, err := redStates(t, runDir, gap, "--quote", acceptS+".", "--new", acceptN+".", "--no-text"); err == nil {
			t.Errorf("a road with --no-text landed: %s", out)
		}
		if out, err := redStates(t, runDir, gap, "--quote", "words the report does not hold", "--new", "x"); err == nil {
			t.Errorf("a pair the accept refuses landed: %s", out)
		}
		if out, err := redStates(t, runDir, gap, "--no-text"); err != nil {
			t.Fatalf("--no-text: %v\n%s", err, out)
		}
		conveneBench(t, runDir, gap)
		if out, err := answerDocketing(t, runDir, d, "for-red", "--final"); err == nil || !strings.Contains(err.Error()+out, "--no-text") {
			t.Errorf("--for-red after --no-text: %v\n%s", err, out)
		}
		if out, err := answerDocketing(t, runDir, d, "for-blue", "--final"); err != nil {
			t.Errorf("--for-blue after --no-text: %v\n%s", err, out)
		}
	})
}

// TestRestoringIsTheFirstFromAgainstTheLastToThroughRestore drives `--restore` at grade movement: per
// axis the window moved, the bench's regrade restores the first move's from (gradeMovement.restoring);
// an axis moved down and back up is not written, and with no axis moved net the ruling lands alone.
// `--keep` writes no grade; both lock each moved axis.
func TestRestoringIsTheFirstFromAgainstTheLastToThroughRestore(t *testing.T) {
	moved := func(t *testing.T, moves ...[2]string) (string, string, string) {
		t.Helper()
		runDir := acceptRun(t)
		gap, err := mintQuote(t, runDir, "G", acceptS+".")
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range moves {
			if _, err := lensRegrade(t, runDir, gap, m[0], m[1]); err != nil {
				t.Fatalf("the lens's move %v: %v", m, err)
			}
		}
		d := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, "")
		conveneBench(t, runDir, gap)
		return runDir, gap, d
	}
	t.Run("restore: severity back to its first from, impact moved and back not written", func(t *testing.T) {
		runDir, gap, d := moved(t, [2]string{"severity", "certain"}, [2]string{"impact", "low"}, [2]string{"severity", "high"}, [2]string{"impact", "medium"})
		if out, err := answerDocketing(t, runDir, d, "restore"); err != nil {
			t.Fatalf("--restore: %v\n%s", err, out)
		}
		rs := benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE)
		if len(rs) != 1 {
			t.Fatalf("the bench wrote %d regrades, want the restore", len(rs))
		}
		if r, _ := recordpb.BodyAs[*recordpb.Regrade](rs[0]); r.GetSeverity() != recordpb.Grade_GRADE_MEDIUM || r.Impact != nil {
			t.Errorf("the restore = %v, want severity medium alone", r)
		}
		if r := lastRuling(t, runDir); r.GetMovedGrades() != recordpb.MovedGrades_MOVED_GRADES_RESTORED || r.SidedWith != nil || r.Disposition != nil {
			t.Errorf("the ruling = %+v, want grades restored and nothing else", r)
		}
		for _, axis := range []string{"severity", "impact"} {
			out, err := lensRegrade(t, runDir, gap, axis, "certain")
			if err == nil || !strings.Contains(err.Error()+out, "ruling at event") {
				t.Errorf("the lens moved the restored %s: %v", axis, err)
			}
		}
	})
	t.Run("restore with no axis moved net: the ruling alone", func(t *testing.T) {
		runDir, _, d := moved(t, [2]string{"severity", "certain"}, [2]string{"severity", "medium"})
		if out, err := answerDocketing(t, runDir, d, "restore"); err != nil {
			t.Fatalf("--restore: %v\n%s", err, out)
		}
		if n := len(benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE)); n != 0 {
			t.Errorf("the bench wrote %d regrades, want none", n)
		}
	})
	t.Run("keep: no grade written, the moved axis locked; a side and a remand refused", func(t *testing.T) {
		runDir, gap, d := moved(t, [2]string{"severity", "certain"}, [2]string{"severity", "low"})
		for _, refused := range []string{"for-red", "for-blue", "remand"} {
			if out, err := answerDocketing(t, runDir, d, refused, "--direction", "x"); err == nil {
				t.Errorf("--%s at grade movement landed: %s", refused, out)
			}
		}
		if out, err := answerDocketing(t, runDir, d, "keep"); err != nil {
			t.Fatalf("--keep: %v\n%s", err, out)
		}
		if n := len(benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_REGRADE)); n != 0 {
			t.Errorf("--keep wrote %d bench regrades", n)
		}
		if g := gapOf(t, runDir, gap); recordpb.Word(g.Severity) != "low" {
			t.Errorf("the kept severity = %v, want low", g.Severity)
		}
		out, err := lensRegrade(t, runDir, gap, "severity", "high")
		if err == nil || !strings.Contains(err.Error()+out, "ruling at event") {
			t.Errorf("the lens moved the kept axis: %v", err)
		}
	})
}

// lensRegrade is the lens's regrade of one axis of gap, in a sitting of its own: a sitting holds one
// regrade per gap.
func lensRegrade(t *testing.T, runDir, gap, axis, grade string) (string, error) {
	t.Helper()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	return run(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", gap, "--"+axis, grade, "--reason", "the lens's move")
}

// Lens against lens: a red seat's close proposal on another lens's gap puts the filer's proposal
// against the assignee's text; --for-filer applies the close with its stored word, --for-assignee the
// assignee's text, and red's pair is refused there.
func TestALensProposalIsRuledForTheFilerOrTheAssignee(t *testing.T) {
	for _, side := range []string{"for-filer", "for-assignee"} {
		t.Run(side, func(t *testing.T) {
			runDir := newRun(t)
			seedCastWith(t, runDir, "red-lens-logic")
			writeReport(t, runDir, "# Costs\n\nThe year opened badly. "+acceptS+". Volume recovered later.\n")
			registerBlue(t, runDir)
			gap, err := mintQuote(t, runDir, "G", acceptS+".")
			if err != nil {
				t.Fatal(err)
			}
			exchangeOn(t, runDir, gap)
			if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-logic"); err != nil {
				t.Fatal(err)
			}
			m := proposeClose(t, runDir, "red-lens-logic", gap, "moot", "the sentence no longer matters")
			if _, err := redStates(t, runDir, gap); err != nil {
				t.Fatalf("the assignee's closing: %v", err)
			}
			d := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, m)
			conveneBench(t, runDir, gap)
			for _, refused := range []string{"for-red", "for-blue"} {
				if out, err := answerDocketing(t, runDir, d, refused, "--final"); err == nil || !strings.Contains(err.Error()+out, "--for-filer | --for-assignee") {
					t.Errorf("--%s on a lens's proposal: %v\n%s", refused, err, out)
				}
			}
			if out, err := answerDocketing(t, runDir, d, side, "--final"); err != nil {
				t.Fatalf("--%s: %v\n%s", side, err, out)
			}
			r := lastRuling(t, runDir)
			want := map[string][2]string{"for-filer": {"filer", "moot"}, "for-assignee": {"assignee", "repaired"}}[side]
			if recordpb.Word(r.GetSidedWith()) != want[0] || recordpb.Word(r.GetDisposition()) != want[1] {
				t.Errorf("the ruling = %v %v, want %v", r.GetSidedWith(), r.GetDisposition(), want)
			}
			if g := gapOf(t, runDir, gap); g == nil || g.Open {
				t.Errorf("gap %s stays open", gap)
			}
		})
	}
}
