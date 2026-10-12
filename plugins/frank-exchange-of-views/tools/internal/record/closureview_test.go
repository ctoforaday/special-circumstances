package record

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// closureStage is one gap, G1, minted by the evidence lens in epoch 1, and the closing acts a row
// writes on it — every one through the write path, as a seat (or migrate's replay) writes it.
type closureStage struct {
	t                  *testing.T
	run                Run
	chair, lens, bench Identity
	motions            int
}

func newClosureStage(t *testing.T) *closureStage {
	t.Helper()
	s := &closureStage{t: t, run: corrRun(t)}
	s.chair = sit(t, s.run, "red-chair")
	s.lens = sit(t, s.run, evLens)
	mustAppend(t, s.lens, corrMint("G1"))
	return s
}

// epoch opens the chair's next sitting: everything written after it is in the next epoch.
func (s *closureStage) epoch() *closureStage {
	s.t.Helper()
	sit(s.t, s.run, "red-chair")
	return s
}

// close is the lens's own close of G1, ending it as word; prose tells two closes apart.
func (s *closureStage) close(word recordpb.Disposition, prose string) *Event {
	s.t.Helper()
	return mustAppend(s.t, s.lens, closureOf(word, prose))
}

func closureOf(word recordpb.Disposition, prose string) *recordpb.Close {
	return &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(word),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String(prose)}
}

// carry is the chair restating in this epoch the closure G1 was given in epoch from.
func (s *closureStage) carry(from string, word recordpb.Disposition) *Event {
	s.t.Helper()
	return mustAppend(s.t, s.chair, &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(word),
		CarriedFrom: proto.String(from), Prose: proto.String("carried")})
}

func docketRuling(motion string, word recordpb.Disposition) *recordpb.MotionRule {
	d := &recordpb.DocketRuling{Disposition: recordtest.P(word),
		Principle: proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"), Settled: proto.String("st")}
	if word == recordpb.Disposition_DISPOSITION_REMANDED {
		d.ReopensOn = proto.String("new evidence")
	} else {
		d.Final = proto.Bool(true)
	}
	return &recordpb.MotionRule{MotionId: proto.String(motion), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Opinion: proto.String("because"), Ruling: &recordpb.MotionRule_Docket{Docket: d}}
}

// ruled is the pair a bench disposition is: the chair dockets G1, the bench sits and rules it word.
func (s *closureStage) ruled(word recordpb.Disposition) *Event {
	s.t.Helper()
	s.motions++
	id := "M" + string(rune('0'+s.motions))
	mustAppend(s.t, s.chair, docketMotion(id, "G1"))
	s.bench = sit(s.t, s.run, "judge")
	return mustAppend(s.t, s.bench, docketRuling(id, word))
}

// strike replaces the act keyed key with body, as a correction the write path does not admit: the
// replacement and its Correction are inserted as rows, past every guard. The two rows that need it
// are acts the write refuses even from migrate's replay —
//
//	(f) "a motion_rule correction changes only the seat's own wording (…); this one changes --as,
//	    --final" (validateCorrection's tier check), and
//	(g) "another seat has acted since this close (…); a correction strikes an act only before any
//	    other seat acts" (validateCorrection's interleaving check)
//
// — so no admissible record holds either, and the rows pin what the readers make of one all the same.
func (s *closureStage) strike(seat, key string, body proto.Message) {
	s.t.Helper()
	db := seatWriter(s.t, filepath.Join(s.run.Records(), "record.db"))
	for _, ev := range []*Event{
		recordtest.At(s.t, seat, key+"~1", body),
		recordtest.At(s.t, seat, correctionKey(seat, key), &recordpb.Correction{
			Corrects: proto.String(key), Replacement: proto.String(key + "~1"), Why: proto.String("w")}),
	} {
		if _, err := recordsql.Insert(db, ev); err != nil {
			s.t.Fatal(err)
		}
	}
}

// closureWant is what every reader must say of G1.
type closureWant struct {
	open          bool
	byBench       bool
	epoch         int
	fate          string // the last closer's word
	bodyClass     string // the embedded board body's closure_class, as the body spells it ("" when the body is the bench's)
	bodyCarried   string // … and its carried_from
	bodyDispo     string // … or the bench's disposition, when no red close stands
	benchDispo    string // FamilyOf's BenchClosure word ("" when no closing ruling stands)
	closureProse  string // FamilyOf's Closure prose ("" when no red close stands)
	estopped      bool
	closureAbsent bool
}

// A GAP'S CLOSURE IS ITS LAST STANDING CLOSING ACT, on each arm, and whichever arm acted last is the
// one that closed it. The board, the family and the work list each say so of the same record: a
// carried or re-ruled gap reads its carry or its second ruling, a struck act closes nothing, and a
// corrected act keeps the place of the act it replaced.
func TestAGapsClosureIsItsLastStandingClosingAct(t *testing.T) {
	const (
		repaired   = recordpb.Disposition_DISPOSITION_REPAIRED
		accepted   = recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED
		notADefect = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT
		remanded   = recordpb.Disposition_DISPOSITION_REMANDED
		elsewhere  = recordpb.Disposition_DISPOSITION_DEFECT_OWED_ELSEWHERE
	)
	rows := []struct {
		name  string
		build func(s *closureStage)
		want  closureWant
	}{
		{"a: red closes, the bench rules it closed in a later sitting", func(s *closureStage) {
			s.close(repaired, "red's close")
			s.epoch().ruled(notADefect)
		}, closureWant{byBench: true, epoch: 2, fate: "not_a_defect", bodyClass: "DISPOSITION_REPAIRED", benchDispo: "not_a_defect",
			closureProse: "red's close", estopped: true}},
		{"b: the bench remands", func(s *closureStage) {
			s.ruled(remanded)
		}, closureWant{open: true, closureAbsent: true}},
		{"c: the bench rules it closed twice", func(s *closureStage) {
			s.ruled(notADefect)
			s.epoch().ruled(elsewhere)
		}, closureWant{byBench: true, epoch: 2, fate: "defect_owed_elsewhere", bodyDispo: "DISPOSITION_DEFECT_OWED_ELSEWHERE",
			benchDispo: "defect_owed_elsewhere", estopped: true}},
		{"d: red closes, the chair carries it under another word", func(s *closureStage) {
			s.close(accepted, "red's close")
			s.epoch().carry("1", repaired)
		}, closureWant{epoch: 2, fate: "repaired", bodyClass: "DISPOSITION_REPAIRED", bodyCarried: "1", closureProse: "carried"}},
		{"e: red closes, the bench rules, the chair carries", func(s *closureStage) {
			s.close(accepted, "red's close")
			s.epoch().ruled(notADefect)
			s.epoch().carry("1", repaired)
		}, closureWant{epoch: 3, fate: "repaired", bodyClass: "DISPOSITION_REPAIRED", bodyCarried: "1", benchDispo: "not_a_defect",
			closureProse: "carried"}},
		{"f: red closes, the bench's closing ruling is corrected to a remand", func(s *closureStage) {
			s.close(repaired, "red's close")
			k := s.epoch().ruled(notADefect).GetKey()
			s.strike("judge", k, docketRuling("M1", remanded))
		}, closureWant{epoch: 1, fate: "repaired", bodyClass: "DISPOSITION_REPAIRED", closureProse: "red's close"}},
		{"g: red closes, the bench rules, red corrects its close", func(s *closureStage) {
			k := s.close(repaired, "red's close").GetKey()
			s.ruled(notADefect)
			s.strike(evLens, k, closureOf(repaired, "red's close, reworded"))
		}, closureWant{byBench: true, epoch: 1, fate: "not_a_defect", bodyClass: "DISPOSITION_REPAIRED", benchDispo: "not_a_defect",
			closureProse: "red's close, reworded", estopped: true}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			s := newClosureStage(t)
			row.build(s)
			w := row.want

			// The board.
			bj, err := BoardJSONOfRun(s.run)
			if err != nil {
				t.Fatal(err)
			}
			if len(bj.Anomalies) > 0 {
				t.Errorf("board anomalies = %q, want none", bj.Anomalies)
			}
			var gj *GapJSON
			for _, g := range append(append([]GapJSON{}, bj.Open...), bj.Closed...) {
				if g.ID == "G1" {
					gj = &g
				}
			}
			if gj == nil {
				t.Fatal("G1 is not on the board")
			}
			if gj.Open != w.open || gj.ClosedByBench != w.byBench || gj.ClosedEpoch != w.epoch {
				t.Errorf("board: open=%v closed_by_bench=%v closed_epoch=%d, want %v/%v/%d", gj.Open, gj.ClosedByBench, gj.ClosedEpoch, w.open, w.byBench, w.epoch)
			}
			word := func(v any) string { s, _ := v.(string); return s }
			if got := [3]string{word(gj.Closure["closure_class"]), word(gj.Closure["carried_from"]), word(gj.Closure["disposition"])}; got != [3]string{w.bodyClass, w.bodyCarried, w.bodyDispo} {
				t.Errorf("board: embedded closure (closure_class, carried_from, disposition) = %q, want %q", got, [3]string{w.bodyClass, w.bodyCarried, w.bodyDispo})
			}
			if (gj.Closure == nil) != w.closureAbsent {
				t.Errorf("board: embedded closure = %v, want absent=%v", gj.Closure, w.closureAbsent)
			}
			if want := map[bool]int{true: 1}[w.byBench]; bj.Counts.ClosedByBench != want {
				t.Errorf("board: counts.closed_by_bench = %d, want %d", bj.Counts.ClosedByBench, want)
			}

			// The family.
			fam, err := FamilyOf(s.run)
			if err != nil {
				t.Fatal(err)
			}
			g := fam.Gap("G1")
			if g == nil {
				t.Fatal("G1 is not in the family")
			}
			if g.Open != w.open || g.HasClosed == w.open || g.ClosedByBench != w.byBench || g.ClosedEpoch != w.epoch {
				t.Errorf("family: open=%v has_closed=%v closed_by_bench=%v closed_epoch=%d, want %v/%v/%v/%d",
					g.Open, g.HasClosed, g.ClosedByBench, g.ClosedEpoch, w.open, !w.open, w.byBench, w.epoch)
			}
			if got := g.ClosureReason(); got != w.fate {
				t.Errorf("family: ClosureReason() = %q, want %q", got, w.fate)
			}
			if got := g.Closure.GetProse(); got != w.closureProse {
				t.Errorf("family: Closure prose = %q, want %q", got, w.closureProse)
			}
			if got := g.Closure.GetCarriedFrom(); got != w.bodyCarried {
				t.Errorf("family: Closure carried_from = %q, want %q", got, w.bodyCarried)
			}
			benchWord := ""
			if g.BenchClosure != nil {
				benchWord = recordpb.Word(g.BenchClosure.GetDisposition())
			}
			if benchWord != w.benchDispo {
				t.Errorf("family: BenchClosure disposition = %q, want %q", benchWord, w.benchDispo)
			}

			// The work list.
			r, err := workView(s.run)
			if err != nil {
				t.Fatal(err)
			}
			states, err := workGapStatesOf(s.run, r)
			if err != nil {
				t.Fatal(err)
			}
			if len(states) != 1 || states[0].Open != w.open || states[0].ClosedByBench != w.byBench || states[0].Fate != w.fate {
				t.Errorf("work: %d gaps, open=%v closed_by_bench=%v fate=%q, want 1, %v/%v/%q",
					len(states), states[0].Open, states[0].ClosedByBench, states[0].Fate, w.open, w.byBench, w.fate)
			}
			wj, err := WorkJSONOfRun(s.run)
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case w.estopped && (len(wj.Estopped) != 1 || wj.Estopped[0].ID != "G1" || wj.Estopped[0].Fate != w.fate):
				t.Errorf("work: estopped = %+v, want G1 with fate %q", wj.Estopped, w.fate)
			case !w.estopped && len(wj.Estopped) != 0:
				t.Errorf("work: estopped = %+v, want none — the bench's is not the act that closed G1", wj.Estopped)
			}
		})
	}
}

// THE BOARD BESIDE ITS CLOSURES, on one record: a gap closed by BOTH arms (attribution follows the
// LAST closing act while the embedded body prefers the red close), a non-closing docket ruling that
// must close nothing, regrade overlays, credited and uncredited findings, and the verify/cite count
// split.
func TestTheBoardEmbedsItsClosuresRegradesAndCredit(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-chair"}
	blue := Identity{Run: run, SeatID: "blue-respond"}
	lens := Identity{Run: run, SeatID: "red-lens-evidence"}
	judge2 := Identity{Run: run, SeatID: "judge"}
	app := func(id Identity, body proto.Message) {
		t.Helper()
		if _, err := Append(id, body); err != nil {
			t.Fatal(err)
		}
	}

	app(lens, &recordpb.Finding{Id: proto.String("F-0a0a0a0a"), Location: proto.String("¶3"), Text: proto.String("overclaims"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH)})
	app(lens, &recordpb.Finding{Id: proto.String("F-0b0b0b0b"), Text: proto.String("uncredited")})

	mint := func(id string, extra func(m *recordpb.Mint)) {
		m := &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), Class: proto.String("self-attestation"),
			Problem: proto.String("p " + id), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
			CheckKind:  recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
		if extra != nil {
			extra(m)
		}
		app(red, m)
	}
	// The chair sits: its mints, regrade and close below are epoch 1.
	if _, _, err := RegisterSeat(red, "", ""); err != nil {
		t.Fatal(err)
	}
	mint("G1", func(m *recordpb.Mint) {
		m.FoundBy = []string{"F-0a0a0a0a"}
		m.MintReason = proto.String("the argument")
		m.Location = proto.String("the claimed span")
		m.FixBasis = proto.String("proposed")
		m.FixNew = proto.String("the fix")
		m.CheckKind = recordpb.CheckKind_CHECK_KIND_COMPUTATION.Enum()
	})
	mint("G2", func(m *recordpb.Mint) { m.Supersedes = []string{"G1"} })
	mint("G3", nil)

	// A regrade the projection must both OVERLAY (grades) and EMBED (the regrades list).
	app(red, &recordpb.Regrade{GapId: proto.String("G3"),
		Impact: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("moved")})
	// G3: closed by red in round 1, then ruled by the bench in round 2 — BOTH arms.
	// Attribution must follow the bench (last), the embedded body must stay red's.
	app(red, &recordpb.Close{GapId: proto.String("G3"),
		ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
		AnchorSeat:   proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"),
		Prose: proto.String("verified at the leaf")})
	// THE BENCH'S DISPOSITION IS A DOCKET MOTION'S RULING (#681 Scope 2), so each of these is a
	// PAIR: red files the gap to the bench, the bench rules. The gap rides the FILING, which is
	// why both halves have to be on the record for either projection to attribute the closure.
	docket := func(motionID, gapID string, rule *recordpb.DocketRuling) {
		t.Helper()
		app(red, &recordpb.Motion{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:   proto.String("red cannot settle " + gapID),
			Filing:  &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gapID)}}})
		// The chair sits again before the bench rules, so the ruling — and the closure attributed to
		// it — lands in epoch 2, which is what "ruled by the bench in round 2" now means.
		if _, _, err := RegisterSeat(red, "", ""); err != nil {
			t.Fatal(err)
		}
		app(judge2, &recordpb.MotionRule{MotionId: proto.String(motionID),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String("ra"),
			Ruling:  &recordpb.MotionRule_Docket{Docket: rule}})
	}
	docket("M1", "G3", &recordpb.DocketRuling{
		Disposition: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Settled: proto.String("st"), Final: proto.Bool(true)})
	// A CARRIED ruling on an open gap closes NOTHING — the vocabulary's own facet decides.
	docket("M2", "G2", &recordpb.DocketRuling{
		Disposition: recordpb.Disposition_DISPOSITION_REMANDED.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Settled: proto.String("st"), ReopensOn: proto.String("new evidence")})

	// The count split: one verify (red's read), one cite (blue's authoring).
	app(red, &recordpb.Verify{Url: proto.String("https://example.org"), Claim: proto.String("c"),
		Label: proto.String("v-1"), Title: proto.String("e"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("the source states it plainly")})
	app(blue, &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), Label: proto.String("c-1"), Url: proto.String("https://example.org"),
		Title: proto.String("t"), CiteKey: proto.String("k1")})

	bj, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]GapJSON{}
	for _, g := range append(append([]GapJSON{}, bj.Open...), bj.Closed...) {
		byID[g.ID] = g
	}
	// G3: closed by red (r1), then ruled by the bench (r2). Attribution follows the bench;
	// the embedded body stays red's close (its prose proves which body rendered).
	g3 := byID["G3"]
	if g3.Open || !g3.ClosedByBench || g3.ClosedEpoch != 2 {
		t.Errorf("G3 attribution = open=%v bench=%v round=%d, want closed/bench/2", g3.Open, g3.ClosedByBench, g3.ClosedEpoch)
	}
	if g3.Closure == nil || g3.Closure["prose"] != "verified at the leaf" {
		t.Errorf("G3 embedded closure = %v, want red's close body (red's close is preferred)", g3.Closure)
	}
	if len(g3.Regrades) != 1 {
		t.Errorf("G3 regrades = %v, want the one recorded regrade embedded", g3.Regrades)
	}
	if g3.Impact != "high" {
		t.Errorf("G3 impact = %v, want the regrade overlay", g3.Impact)
	}
	// G2: a carried docket ruling closes nothing.
	if g2 := byID["G2"]; !g2.Open {
		t.Error("a carried ruling closed G2 — the vocabulary's own facet says it must not")
	}
	// G1 (computation, unproved) and G2 (carried) stay open; G3 closed by both arms.
	if bj.Counts.Open != 2 || bj.Counts.Closed != 1 || bj.Counts.ClosedByBench != 1 {
		t.Errorf("counts = %+v", bj.Counts)
	}
	if bj.Counts.Citations != 1 || bj.Counts.CitationsAuthored != 1 {
		t.Errorf("citation split = %d/%d, want 1/1", bj.Counts.Citations, bj.Counts.CitationsAuthored)
	}
	credited, uncredited := 0, 0
	for _, o := range bj.Observations {
		if o.Credited {
			credited++
		} else {
			uncredited++
		}
	}
	if credited != 1 || uncredited != 1 || bj.Counts.UncreditedFindings != 1 {
		t.Errorf("credit split = %d/%d (counter %d), want 1/1/1", credited, uncredited, bj.Counts.UncreditedFindings)
	}
}

// A CLOSING ACT THE READ DID NOT LOAD IS AN ERROR, never a gap that did not close. The gap view
// names the act; a read that lacks it — the stream and the gap rows off two snapshots, or a
// declaration that leaves the family out — says which gap and which event, and the board, which has
// a channel for it, reports the closure dropped.
func TestAClosingActTheReadDidNotLoadIsLoud(t *testing.T) {
	s := newClosureStage(t)
	s.close(recordpb.Disposition_DISPOSITION_REPAIRED, "red's close")
	s.epoch().ruled(recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)

	absent := sql.NullInt64{Int64: 4242, Valid: true}
	for _, c := range []struct {
		name         string
		close, bench sql.NullInt64
		want         string
	}{
		{"the red close", absent, sql.NullInt64{}, "names event 4242 as its close_event"},
		{"the bench's ruling", sql.NullInt64{}, absent, "names event 4242 as its bench_close_event"},
	} {
		got, err := closeStateOf("G1", c.close, c.bench, c.bench.Valid, map[int64]*Event{}, WindowIndex{})
		if got != nil || err == nil || !strings.Contains(err.Error(), "gap G1") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s the read did not load: closeStateOf = (%v, %v), want an error naming gap G1 and %q", c.name, got, err, c.want)
		}
	}
	if got, err := closeStateOf("G1", sql.NullInt64{}, sql.NullInt64{}, false, nil, WindowIndex{}); got != nil || err != nil {
		t.Errorf("a gap no act closes: closeStateOf = (%v, %v), want (nil, nil)", got, err)
	}

	// The call, on each reader: a read whose events leave the closing acts out.
	err := readSnapshot(s.run, func(q recordsql.Querier) error {
		evs, ws, err := recordsql.EventsOfTypes(q, "regrade")
		if err != nil {
			return err
		}
		bj, err := boardJSONOfRecord(q, evs, windowIndexOf(evs, ws))
		if err != nil {
			return err
		}
		if len(bj.Anomalies) != 1 || !strings.Contains(bj.Anomalies[0], "gap G1") || !strings.Contains(bj.Anomalies[0], "DROPPED") {
			t.Errorf("board anomalies = %q, want the one dropped closure of G1", bj.Anomalies)
		}
		if len(bj.Closed) != 1 || bj.Closed[0].Closure != nil {
			t.Errorf("board closed = %+v, want G1 closed with no embedded closure", bj.Closed)
		}
		r, err := workReadsAt(q)
		if err != nil {
			return err
		}
		r.evs = evs
		if _, err := workGapStatesOf(s.run, r); err == nil || !strings.Contains(err.Error(), "gap G1") {
			t.Errorf("the work list over a read without the closing acts: %v, want an error naming gap G1", err)
		}
		fr, err := familyReadsAt(q)
		if err != nil {
			return err
		}
		fr.merged.Events = evs
		if _, err := gapStatesOf(fr); err == nil || !strings.Contains(err.Error(), "gap G1") {
			t.Errorf("the family over a read without the closing acts: %v, want an error naming gap G1", err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
