package record

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// closureStage is one gap, G1, minted by the evidence lens in epoch 1 on an ingested report, and the
// closing acts a row writes on it — every one through the write path, as a seat (or migrate's
// replay) writes it, each closing act by a seat whose sitting holds its delivery.
type closureStage struct {
	t                  *testing.T
	run                Run
	chair, lens, bench Identity
	motions            int
	docketing          string // the latest docketing docketed wrote
}

func newClosureStage(t *testing.T) *closureStage {
	t.Helper()
	s := &closureStage{t: t, run: withReport(t, corrRun(t))}
	// The cast seats the bench, so the dispatch may convene it (convene).
	mustAppend(t, Identity{Run: s.run, SeatID: HarnessSeat}, castBody([]string{evLens, "red-chair", "blue-respond", "judge"}, nil))
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
	delivered(s.t, s.lens)
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

// bluesWord is blue's closing on gap stating word, which a ruling for blue writes: blue sits and
// files it. A remand or not_a_defect owes none — for blue without a word writes not_a_defect.
func bluesWord(t *testing.T, run Run, gap string, word recordpb.Disposition) {
	t.Helper()
	if word == recordpb.Disposition_DISPOSITION_REMANDED || word == recordpb.Disposition_DISPOSITION_NOT_A_DEFECT {
		return
	}
	mustAppend(t, sit(t, run, "blue-respond"), &recordpb.Closing{GapId: proto.String(gap), ClosureClass: recordtest.P(word),
		Text: proto.String("blue's position")})
}

// docketRuling is the bench's ruling of docketing, holding no proposal: a remand, or for blue with
// word, which bluesWord has blue state first.
func docketRuling(docketing string, word recordpb.Disposition) *recordpb.MotionRule {
	d := &recordpb.DocketRuling{DocketingId: proto.String(docketing), Disposition: recordtest.P(word),
		Principle: proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"), Settled: proto.String("st")}
	if word == recordpb.Disposition_DISPOSITION_REMANDED {
		d.Direction = proto.String("new evidence")
	} else {
		d.Final = proto.Bool(true)
		d.SidedWith = recordpb.SidedWith_SIDED_WITH_BLUE.Enum()
	}
	return &recordpb.MotionRule{Opinion: proto.String("because"), Ruling: &recordpb.MotionRule_Docket{Docket: d}}
}

// ruled is the pair a bench disposition is: the chair dockets G1, the bench sits and rules it word.
func (s *closureStage) ruled(word recordpb.Disposition) *Event {
	s.t.Helper()
	return s.rule(s.docketed(), word)
}

// docketed is the dispatch putting G1 before the bench, and rule the bench, convened on G1, ruling
// that docketing — apart, for the record where the gap closes between the two.
func (s *closureStage) docketed() string {
	s.t.Helper()
	s.docketing = mustDocketByDispatch(s.t, s.run, "G1")
	return s.docketing
}

// docketedArchived is a docket motion on G1 that an archived record holds where the live write
// refuses it: on G1 once it is closed, where no act but a successor moves it (F8). The readers still
// read such a record, so the stage seeds it through the replay writer.
func (s *closureStage) docketedArchived() string {
	s.t.Helper()
	s.motions++
	id := fmt.Sprintf("D-archived-%d", s.motions)
	if _, err := replayed(s.t, s.chair, &recordpb.Docketing{DocketingId: proto.String(id), GapId: proto.String("G1"),
		Cause: recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE.Enum(), Basis: proto.String(impasseBasis("G1"))}); err != nil {
		s.t.Fatal(err)
	}
	return id
}

func (s *closureStage) rule(motion string, word recordpb.Disposition) *Event {
	s.t.Helper()
	bluesWord(s.t, s.run, "G1", word)
	s.bench = convene(s.t, s.run, "G1")
	return mustAppend(s.t, s.bench, docketRuling(motion, word))
}

// closureWant is what every reader must say of G1.
type closureWant struct {
	open          bool
	byBench       bool // the act that ended G1 is the bench's
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

// A GAP'S CLOSURE IS ITS LAST STANDING CLOSING ACT, on each arm: whichever arm acted last gives the
// word and the epoch, and the arm whose act came FIRST is the one that closed it. The board, the
// family and the work list each say so of the same record: a carried or re-ruled gap reads its
// carry or its second ruling, and a gap its lens closed before the bench ruled is the lens's.
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
			s.epoch().rule(s.docketedArchived(), notADefect)
		}, closureWant{epoch: 2, fate: "not_a_defect", bodyClass: "repaired", benchDispo: "not_a_defect",
			closureProse: "red's close", estopped: true}},
		// universe-m25 G-e205e589: the chair dockets the gap at its exchange limit (ev391) and the lens
		// closes it repaired (ev440). The write refuses that close now, and an archived record holds it.
		// A close of the gap written after its docketing ends the docketing unruled, so the bench has
		// nothing left to rule on G1, and the lens's close is the gap's closure.
		{"f: an archived lens closed a gap a docketing stands on, which ends the docketing unruled", func(s *closureStage) {
			d := s.docketed()
			delivered(s.t, s.lens)
			if _, err := replayed(s.t, s.lens, closureOf(repaired, "red's close")); err != nil {
				s.t.Fatal(err)
			}
			s.bench = convene(s.t, s.run, "G1")
			if _, err := Append(s.bench, docketRuling(d, notADefect)); err == nil || !strings.Contains(err.Error(), "docketing "+d+" is not before this sitting") {
				s.t.Errorf("the bench ruling docketing %s, which the lens's close of G1 after it ended = %v, want refused as before no sitting", d, err)
			}
		}, closureWant{epoch: 1, fate: "repaired", bodyClass: "repaired", closureProse: "red's close"}},
		{"b: the bench remands", func(s *closureStage) {
			s.ruled(remanded)
		}, closureWant{open: true, closureAbsent: true}},
		{"c: the bench rules it closed twice", func(s *closureStage) {
			s.ruled(notADefect)
			s.epoch().rule(s.docketedArchived(), elsewhere)
		}, closureWant{byBench: true, epoch: 2, fate: "defect_owed_elsewhere", bodyDispo: "defect_owed_elsewhere",
			benchDispo: "defect_owed_elsewhere", estopped: true}},
		{"d: red closes, the chair carries it under another word", func(s *closureStage) {
			s.close(accepted, "red's close")
			s.epoch().carry("1", repaired)
		}, closureWant{epoch: 2, fate: "repaired", bodyClass: "repaired", bodyCarried: "1", closureProse: "carried"}},
		// THE RULING STILL BARS: its tombstone stands until the act that ends it, whatever closing act
		// follows it (#1695 §III.B12), so the estopped row stands under the carry's fate.
		{"e: red closes, the bench rules, the chair carries", func(s *closureStage) {
			s.close(accepted, "red's close")
			s.epoch().rule(s.docketedArchived(), notADefect)
			s.epoch().carry("1", repaired)
		}, closureWant{epoch: 3, fate: "repaired", bodyClass: "repaired", bodyCarried: "1", benchDispo: "not_a_defect",
			closureProse: "carried", estopped: true}},
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
				t.Errorf("work: estopped = %+v, want none — no ruling tombstone stands on G1", wj.Estopped)
			}
		})
	}
}

// THE BOARD BESIDE ITS CLOSURES, on one record: a gap closed by BOTH arms (attribution follows the
// act that ended it, the epoch the LAST closing act, and the embedded body prefers the red close), a
// non-closing docket ruling that
// must close nothing, regrade overlays, and the verify/cite count split.
func TestTheBoardEmbedsItsClosuresRegradesAndMints(t *testing.T) {
	runDir := newRun(t)
	run := withReport(t, mustRun(t, runDir))
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

	mint := func(id string, extra func(m *recordpb.Mint)) {
		m := &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
			Problem: proto.String("p " + id), RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("a"),
			CheckKind:  recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
		if extra != nil {
			extra(m)
		}
		app(lens, m)
	}
	// The chair sits, then the lens: its mints, regrade and close below are epoch 1.
	if _, _, err := RegisterSeat(red, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(lens, "", ""); err != nil {
		t.Fatal(err)
	}
	mint("G1", func(m *recordpb.Mint) {
		m.MintReason = proto.String("the argument")
		m.Location = proto.String("the claimed span")
		m.FixBasis = proto.String("proposed")
		m.FixNew = proto.String("the fix")
		m.CheckKind = recordpb.CheckKind_CHECK_KIND_COMPUTATION.Enum()
	})
	mint("G2", func(m *recordpb.Mint) {
		m.CheckKind = recordpb.CheckKind_CHECK_KIND_COMPUTATION.Enum()
	})
	mint("G3", nil)

	// A regrade the projection must both OVERLAY (grades) and EMBED (the regrades list).
	app(lens, &recordpb.Regrade{GapId: proto.String("G3"),
		Impact: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("moved")})
	// G3: closed by red in round 1, then ruled by the bench in round 2 — BOTH arms.
	// Attribution must follow the bench (last), the embedded body must stay red's.
	delivered(t, lens)
	app(lens, &recordpb.Close{GapId: proto.String("G3"),
		ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
		AnchorSeat:   proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"),
		Prose: proto.String("verified at the leaf")})
	// THE BENCH'S DISPOSITION IS A DOCKETING'S RULING, so each of these is a PAIR: the dispatch
	// dockets the gap for the bench, the bench rules. The gap rides the DOCKETING, which is why both
	// halves have to be on the record for either projection to attribute the closure. G3 is docketed
	// after red closed it, and the bench rules it in a later epoch; the stage seeds both pairs through
	// the replay writer, and the board reads what such a record holds.
	replay := func(id Identity, body proto.Message) {
		t.Helper()
		if _, err := replayed(t, id, body); err != nil {
			t.Fatal(err)
		}
	}
	docket := func(docketingID, gapID string, rule *recordpb.DocketRuling) {
		t.Helper()
		replay(red, docketing(docketingID, gapID))
		// The chair sits again before the bench rules, so the ruling — and the closure attributed to
		// it — lands in epoch 2, which is what "ruled by the bench in round 2" now means.
		if _, _, err := RegisterSeat(red, "", ""); err != nil {
			t.Fatal(err)
		}
		rule.DocketingId = proto.String(docketingID)
		replay(judge2, &recordpb.MotionRule{Opinion: proto.String("ra"),
			Ruling: &recordpb.MotionRule_Docket{Docket: rule}})
	}
	docket("D1", "G3", &recordpb.DocketRuling{
		SidedWith:   recordpb.SidedWith_SIDED_WITH_BLUE.Enum(),
		Disposition: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Settled: proto.String("st"), Final: proto.Bool(true)})
	// A CARRIED ruling on an open gap closes NOTHING — the vocabulary's own facet decides.
	docket("D2", "G2", &recordpb.DocketRuling{
		Disposition: recordpb.Disposition_DISPOSITION_REMANDED.Enum(),
		Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
		Settled: proto.String("st"), Direction: proto.String("new evidence")})

	// The count split: one verify (red's read), one cite (blue's authoring).
	app(red, &recordpb.Verify{CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(), Url: proto.String("https://example.org"), Claim: proto.String("c"),
		Title:      proto.String("e"),
		Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("the source says nothing of it")})
	app(blue, &recordpb.Cite{FetchedAt: proto.String("2026-11-12T00:00:00Z"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), SourceSpan: proto.String("s"), SpanStart: proto.Int32(0), SpanEnd: proto.Int32(1), SpanTextSha: proto.String("t"), SourceMatched: proto.String("s"), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(), Label: proto.String("c-1"), Url: proto.String("https://example.org"),
		Title: proto.String("t"), CiteKey: proto.String("k1")})

	bj, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]GapJSON{}
	for _, g := range append(append([]GapJSON{}, bj.Open...), bj.Closed...) {
		byID[g.ID] = g
	}
	// G3: closed by red (r1), then ruled by the bench (r2). Red's close ended it, so attribution is
	// red's; the epoch is the bench's later ruling, and the embedded body stays red's close (its
	// prose proves which body rendered).
	g3 := byID["G3"]
	if g3.Open || g3.ClosedByBench || g3.ClosedEpoch != 2 {
		t.Errorf("G3 attribution = open=%v bench=%v round=%d, want closed/red/2", g3.Open, g3.ClosedByBench, g3.ClosedEpoch)
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
	// G1 and G2 (carried) stay open; G3 is closed by both arms.
	if bj.Counts.Open != 2 || bj.Counts.Closed != 1 || bj.Counts.ClosedByBench != 0 {
		t.Errorf("counts = %+v, anomalies %v", bj.Counts, bj.Anomalies)
	}
	if bj.Counts.Citations != 1 || bj.Counts.CitationsAuthored != 1 {
		t.Errorf("citation split = %d/%d, want 1/1", bj.Counts.Citations, bj.Counts.CitationsAuthored)
	}
}

// A CLOSING ACT THE READ DID NOT LOAD IS AN ERROR, never a gap that did not close. The gap view
// names the act; a read that lacks it — the stream and the gap rows off two snapshots, or a
// declaration that leaves the family out — says which gap and which event, and the board, which has
// a channel for it, reports the closure dropped.
func TestAClosingActTheReadDidNotLoadIsLoud(t *testing.T) {
	s := newClosureStage(t)
	s.close(recordpb.Disposition_DISPOSITION_REPAIRED, "red's close")
	s.epoch().rule(s.docketedArchived(), recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)

	absent := sql.NullInt64{Int64: 4242, Valid: true}
	for _, c := range []struct {
		name         string
		close, bench sql.NullInt64
		want         string
	}{
		{"the red close", absent, sql.NullInt64{}, "names event 4242 as its close_event"},
		{"the bench's ruling", sql.NullInt64{}, absent, "names event 4242 as its bench_close_event"},
	} {
		got, err := closeStateOf("G1", gapClosure{closeEvent: c.close, benchEvent: c.bench, closedByBench: c.bench.Valid, benchClosedLast: c.bench.Valid}, map[int64]*Event{}, WindowIndex{})
		if got != nil || err == nil || !strings.Contains(err.Error(), "gap G1") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s the read did not load: closeStateOf = (%v, %v), want an error naming gap G1 and %q", c.name, got, err, c.want)
		}
	}
	if got, err := closeStateOf("G1", gapClosure{}, nil, WindowIndex{}); got != nil || err != nil {
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
