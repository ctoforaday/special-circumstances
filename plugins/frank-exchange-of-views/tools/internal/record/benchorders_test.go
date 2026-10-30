package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The bench's orders on the record (gblock, 2026-10-30): a granted petition's relief, a docket
// ruling's answers on the work list, the bench's own regrade and its lock, and reconsideration on
// new grounds. Each case is driven through the write path a seat's act takes.

const orderedRelief = "Restate the section's claim as conditional on the search it names."

func petitionFiled(t *testing.T) Run {
	t.Helper()
	run := corrRun(t)
	mustAppend(t, sit(t, run, "blue-respond"), &recordpb.Motion{MotionId: proto.String("M1"),
		Basis: proto.String("the required fix asks me to assert a search nobody ran"), Relief: proto.String("strike the requirement"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
			Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_INTEGRITY)}}})
	return run
}

func petitionOrder(word recordpb.PetitionRuling, relief *string, binds *recordpb.RulingBinds) *recordpb.MotionRule {
	return &recordpb.MotionRule{MotionId: proto.String("M1"), Opinion: proto.String("on the papers"),
		Ruling: &recordpb.MotionRule_Petition{Petition: word}, Relief: relief, Binds: binds}
}

func TestAGrantStatesItsOrderAndADenialOrdersNothing(t *testing.T) {
	granted, denied := recordpb.PetitionRuling_PETITION_RULING_GRANTED, recordpb.PetitionRuling_PETITION_RULING_DENIED
	blue := recordpb.RulingBinds_RULING_BINDS_BLUE.Enum()
	for _, c := range []struct {
		name string
		rule *recordpb.MotionRule
		want string
	}{
		{"a grant without relief", petitionOrder(granted, nil, blue), "--as granted requires --relief and --binds"},
		{"a grant with blank relief", petitionOrder(granted, proto.String("  "), blue), "--as granted requires --relief and --binds"},
		{"a grant naming nobody", petitionOrder(granted, proto.String(orderedRelief), nil), "--as granted requires --relief and --binds"},
		{"a denial with relief", petitionOrder(denied, proto.String(orderedRelief), nil), "--as denied orders nothing"},
		{"a denial naming whom it binds", petitionOrder(denied, nil, blue), "--as denied orders nothing"},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := petitionFiled(t)
			_, err := Append(sit(t, run, "judge"), c.rule)
			mustRefuse(t, err, c.want)
		})
	}

	t.Run("a docket ruling orders nothing beyond its ruling", func(t *testing.T) {
		s := newClosureStage(t)
		m := s.docketed()
		r := docketRuling(m, recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED)
		r.Relief = proto.String(orderedRelief)
		_, err := Append(sit(t, s.run, "judge"), r)
		mustRefuse(t, err, "a granted petition's order", "a docket ruling orders nothing beyond its ruling")
	})

	t.Run("a grant reaches every seat's work list", func(t *testing.T) {
		run := petitionFiled(t)
		mustAppend(t, sit(t, run, "judge"), petitionOrder(granted, proto.String(orderedRelief), blue))
		for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"chair", "red-chair"}, {"blue", "blue-respond"}, {"bench", "judge"}} {
			w, err := WorkOfSeat(run, seat.role, seat.id)
			if err != nil {
				t.Fatal(err)
			}
			want := ReliefJSON{Motion: "M1", Petitioner: "blue-respond", Binds: "blue", Relief: orderedRelief, Epoch: w.Relief[0].Epoch}
			if len(w.Relief) != 1 || w.Relief[0] != want {
				t.Errorf("%s's work list carries relief %+v, want [%+v]", seat.id, w.Relief, want)
			}
		}
	})
}

// THE RULING'S ANSWERS ARE ON THE ESTOPPED ROW, read from the ruling the bench recorded.
func TestAnEstoppedRowStatesWhatTheRulingSettled(t *testing.T) {
	for _, c := range []struct {
		name      string
		final     bool
		reopensOn string
	}{{"final", true, ""}, {"reopens on a condition", false, "a primary source dated after 2020"}} {
		t.Run(c.name, func(t *testing.T) {
			s := newClosureStage(t)
			m := s.docketed()
			r := docketRuling(m, recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED)
			r.GetDocket().Settled = proto.String("the figure is the source's, not a derivation")
			if !c.final {
				r.GetDocket().Final = nil
				r.GetDocket().ReopensOn = proto.String(c.reopensOn)
			}
			mustAppend(t, sit(t, s.run, "judge"), r)
			w, err := WorkOfSeat(s.run, "blue", "blue-respond")
			if err != nil {
				t.Fatal(err)
			}
			if len(w.Estopped) != 1 {
				t.Fatalf("estopped %+v, want G1", w.Estopped)
			}
			e := w.Estopped[0]
			if e.Settled != "the figure is the source's, not a derivation" || e.Final != c.final || e.ReopensOn != c.reopensOn {
				t.Errorf("estopped row settled %q final %v reopens_on %q, want %q %v %q",
					e.Settled, e.Final, e.ReopensOn, "the figure is the source's, not a derivation", c.final, c.reopensOn)
			}
		})
	}
}

func regradeOf(gap string, d recordpb.GradeDimension, g recordpb.Grade) *recordpb.Regrade {
	r := &recordpb.Regrade{GapId: proto.String(gap), Basis: proto.String("the grade as I hold it")}
	switch d {
	case recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY:
		r.Severity = g.Enum()
	case recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD:
		r.Likelihood = g.Enum()
	case recordpb.GradeDimension_GRADE_DIMENSION_IMPACT:
		r.Impact = g.Enum()
	}
	return r
}

func gradeMotionOn(id, gap string, d recordpb.GradeDimension, g recordpb.Grade) *recordpb.Motion {
	return &recordpb.Motion{MotionId: proto.String(id), Basis: proto.String("the axis is graded wrong"),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"), Dimension: d.Enum(), Proposed: g.Enum()}}}
}

const (
	sev    = recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY
	impact = recordpb.GradeDimension_GRADE_DIMENSION_IMPACT
	low    = recordpb.Grade_GRADE_LOW
	medium = recordpb.Grade_GRADE_MEDIUM
	high   = recordpb.Grade_GRADE_HIGH
	certn  = recordpb.Grade_GRADE_CERTAIN
)

// benchSetSeverity is G1 (minted medium on every axis) docketed, and the bench hearing it setting its
// severity high.
func benchSetSeverity(t *testing.T) *closureStage {
	t.Helper()
	s := newClosureStage(t)
	s.docketed()
	s.bench = sit(t, s.run, "judge")
	mustAppend(t, s.bench, regradeOf("G1", sev, high))
	return s
}

func TestTheBenchSetsAGradeWhileItHearsTheGap(t *testing.T) {
	t.Run("not before a docket motion stands", func(t *testing.T) {
		s := newClosureStage(t)
		_, err := Append(sit(t, s.run, "judge"), regradeOf("G1", sev, high))
		mustRefuse(t, err, "no docket motion on gap G1 stands before you unruled")
	})
	t.Run("not after the motion is ruled", func(t *testing.T) {
		s := newClosureStage(t)
		s.epoch().ruled(recordpb.Disposition_DISPOSITION_REMANDED)
		_, err := Append(sit(t, s.run, "judge"), regradeOf("G1", sev, high))
		mustRefuse(t, err, "no docket motion on gap G1 stands before you unruled")
	})
	t.Run("not to keep a grade", func(t *testing.T) {
		s := newClosureStage(t)
		s.docketed()
		_, err := Append(sit(t, s.run, "judge"), regradeOf("G1", sev, medium))
		mustRefuse(t, err, "this regrade moves no axis of gap G1")
	})
	t.Run("a move is admitted, and is the grade every reader reads", func(t *testing.T) {
		s := benchSetSeverity(t)
		w, err := WorkOfSeat(s.run, "lens", evLens)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range w.Open {
			if g.ID == "G1" && g.Severity != "high" {
				t.Errorf("G1's severity on the lens's list is %q, want the bench's high", g.Severity)
			}
		}
	})
}

func TestAnAxisTheBenchSetIsTheBenchsToMove(t *testing.T) {
	t.Run("a lens regrade moving it is refused", func(t *testing.T) {
		s := benchSetSeverity(t)
		_, err := Append(sit(t, s.run, evLens), regradeOf("G1", sev, low))
		mustRefuse(t, err, "gap G1's severity is high, which the bench set at event", "only the bench moves it again")
	})
	t.Run("a lens regrade keeping it, or moving another axis, is admitted", func(t *testing.T) {
		s := benchSetSeverity(t)
		lens := sit(t, s.run, evLens)
		mustAppend(t, lens, regradeOf("G1", sev, high))
		mustAppend(t, sit(t, s.run, evLens), regradeOf("G1", impact, low))
	})
	t.Run("a grade motion on it is refused, and on another axis admitted", func(t *testing.T) {
		s := benchSetSeverity(t)
		blue := sit(t, s.run, "blue-respond")
		_, err := Append(blue, gradeMotionOn("M7", "G1", sev, low))
		mustRefuse(t, err, "gap G1's severity is high, which the bench set at event", "a grade motion on it asks the chair for a grade no lens may then file")
		mustAppend(t, blue, gradeMotionOn("M8", "G1", impact, low))
	})
	t.Run("a successor carries the bench's grade", func(t *testing.T) {
		s := benchSetSeverity(t)
		lens := sit(t, s.run, evLens)
		successor := func(g recordpb.Grade) *recordpb.Mint {
			m := corrMint("G2")
			m.Supersedes, m.Severity = []string{"G1"}, g.Enum()
			return m
		}
		_, err := Append(lens, successor(medium))
		mustRefuse(t, err, "--supersedes names G1", "mint it at --severity high")
		mustAppend(t, lens, successor(high))
		// THE LOCK FOLLOWS THE LINEAGE: the successor's severity is the bench's too.
		_, err = Append(sit(t, s.run, evLens), regradeOf("G2", sev, low))
		mustRefuse(t, err, "gap G2's severity is high, which the bench set at event")
	})
	t.Run("the bench moves it again, and the lock holds the new grade", func(t *testing.T) {
		s := benchSetSeverity(t)
		mustAppend(t, sit(t, s.run, "judge"), regradeOf("G1", sev, certn))
		lens := sit(t, s.run, evLens)
		_, err := Append(lens, regradeOf("G1", sev, high))
		mustRefuse(t, err, "gap G1's severity is certain, which the bench set at event")
		mustAppend(t, lens, regradeOf("G1", sev, certn))
	})
}

// A BENCH REGRADE IS NO PARTY'S MOVEMENT. Inside the remand's exchange it would read as the exchange
// moving the gap, and a second remand at impasse would be admitted past the limit.
func TestABenchRegradeInARemandExchangeDoesNotSpendTheLimit(t *testing.T) {
	st := remandedAtImpasse(t).lensActs("G1")
	agent := fmt.Sprintf("blue-proves-%d", st.n)
	st.registerAs("blue-respond", agent).blueProves("G1")
	st.docketMotion("blue-respond", "M2", "G1")
	sevCertain := certn
	st.add("judge", &recordpb.Regrade{GapId: proto.String("G1"), Severity: &sevCertain, Basis: proto.String("the bench's grade")})
	st.stop(agent)
	run := st.register("red-chair").docketMotion("red-chair", "M3", "G1").
		dispatch(2, "judge", "G1").register("judge").seed()
	err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, remandRuling("M3", "a second direction"))
	mustRefuse(t, err, RemandLimitRule)
}

// …nor the board's grade movement: the bench's grade is what its ruling rules on.
func TestABenchRegradeIsNoGradeMovement(t *testing.T) {
	b := movedBoard(t)
	b.regrade("judge", "G1", "low").regrade("judge", "G2", "low") // 2.0 + 1.0 were it a party's
	openChairSitting(b)
	run := b.seed()
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range kindsOf(bs) {
		if strings.HasPrefix(k, string(BlockerGradeMovement)) {
			t.Fatalf("the bench's regrades hold a PASS as grade movement: %v", kindsOf(bs))
		}
	}
}

// groundsBoard is G1 minted, docketed and ruled closed (final); more is recorded after the ruling.
func groundsBoard(t *testing.T, after func(*stage)) Run {
	t.Helper()
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high").mint(evLens, "G2", "high").
		add(evLens, findingAbout("F-00000001", "G1")).
		register("red-chair").docketMotion("red-chair", "M1", "G1").
		dispatch(2, "judge", "G1").register("judge").add("judge", closingRuling("M1"))
	if after != nil {
		after(st)
	}
	return st.register("blue-respond").seed()
}

func findingAbout(id, gap string) *recordpb.Finding {
	return &recordpb.Finding{Id: proto.String(id), Text: proto.String("what the ruling did not have"),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String(gap)}
}

func redocket(id string, grounds, quote *string) *recordpb.Motion {
	m := docketMotion(id, "G1")
	m.GetDocket().Grounds, m.GetDocket().GroundsQuote = grounds, quote
	return m
}

func TestAReconsiderationIsAdmittedOnNewGroundsOnly(t *testing.T) {
	file := func(run Run, m *recordpb.Motion) error {
		return validateRun(run, "blue-respond", recordpb.EventType_EVENT_TYPE_MOTION, m)
	}
	newFinding := func(st *stage) { st.add(evLens, findingAbout("F-00000002", "G1")) }
	for _, c := range []struct {
		name    string
		after   func(*stage)
		grounds *string
		quote   *string
		want    []string // nil: admitted
	}{
		{name: "no grounds", want: []string{"the bench ruled gap G1 defect_accepted on motion M1", "it is final", "admitted on NEW GROUNDS only"}},
		{name: "grounds the ruling had", grounds: proto.String("F-00000001"), want: []string{"--grounds names finding F-00000001, recorded at event", "before the ruling it would reopen"}},
		{name: "grounds about another gap", after: func(st *stage) { st.add(evLens, findingAbout("F-00000003", "G2")) },
			grounds: proto.String("F-00000003"), want: []string{"--grounds names finding F-00000003, which is not about gap G1"}},
		{name: "grounds no act placed", grounds: proto.String("F-99999999"), want: []string{"--grounds names F-99999999, which no act on this record placed"}},
		{name: "both locators", after: newFinding, grounds: proto.String("F-00000002"), quote: proto.String("x"), want: []string{"--grounds and --quote each locate the new grounds; pass one"}},
		{name: "a quote no edit since wrote", quote: proto.String("never written"), want: []string{`--quote "never written" is not new grounds on gap G1`}},
		{name: "a finding about the gap since the ruling", after: newFinding, grounds: proto.String("F-00000002")},
		{name: "a proof answering the gap since the ruling", after: func(st *stage) { st.blueProves("G1") }, grounds: proto.String("P-")},
		{name: "a quote of text an edit answering the gap wrote since", after: func(st *stage) { st.edit("G1", "the old sentence", "the sentence the search now backs") },
			quote: proto.String("the search now backs")},
	} {
		t.Run(c.name, func(t *testing.T) {
			var proofID string
			after := c.after
			if c.grounds != nil && *c.grounds == "P-" {
				after = func(st *stage) { proofID = fmt.Sprintf("P-%d", st.n+1); st.blueProves("G1") }
			}
			run := groundsBoard(t, after)
			g := c.grounds
			if proofID != "" {
				g = proto.String(proofID)
			}
			err := file(run, redocket("M2", g, c.quote))
			if c.want == nil {
				if err != nil {
					t.Fatalf("refused: %v", err)
				}
				return
			}
			mustRefuse(t, err, c.want...)
		})
	}

	t.Run("grounds on a gap the bench has not ruled closed", func(t *testing.T) {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high").
			add(evLens, findingAbout("F-00000002", "G1")).register("blue-respond")
		err := file(st.seed(), redocket("M2", proto.String("F-00000002"), nil))
		mustRefuse(t, err, "the bench has not ruled gap G1 closed")
	})
	t.Run("a re-docket after a remand needs none", func(t *testing.T) {
		run := remandedAtImpasse(t).register("blue-respond").seed()
		if err := file(run, redocket("M2", nil, nil)); err != nil {
			t.Fatalf("a docket motion on a remanded gap was refused: %v", err)
		}
	})
}
