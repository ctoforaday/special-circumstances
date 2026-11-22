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
// ruling's answers on the work list, and the regrade a ruling applies with its lock. Each case is driven through the write path a seat's act takes.

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
		r := docketRuling(m, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)
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
			bluesWord(t, s.run, "G1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED)
			mustAppend(t, convene(t, s.run, "G1"), r)
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

const (
	sev    = recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY
	impact = recordpb.GradeDimension_GRADE_DIMENSION_IMPACT
	low    = recordpb.Grade_GRADE_LOW
	medium = recordpb.Grade_GRADE_MEDIUM
	high   = recordpb.Grade_GRADE_HIGH
	certn  = recordpb.Grade_GRADE_CERTAIN
)

// benchRuledSeverity is G1 (minted medium on every axis) with the chair's proposal M1 to regrade its
// severity low docketed, and the bench ruling for the filer: the tool writes the proposal's regrade as
// the bench's, and the ruling after it decides the severity.
func benchRuledSeverity(t *testing.T) *closureStage {
	t.Helper()
	s := newClosureStage(t)
	fileProposal(t, s.run, regradeMotion("M1", "G1", "severity is overstated"))
	s.docketing = docketProposal(t, s.run, "G1", "M1")
	s.bench = convene(t, s.run, "G1")
	mustAppend(t, s.bench, regradeOf("G1", sev, low))
	mustAppend(t, s.bench, ruling(s.docketing, "M1", func(d *recordpb.DocketRuling) {
		d.SidedWith = recordpb.SidedWith_SIDED_WITH_FILER.Enum()
	}))
	return s
}

// THE BENCH'S REGRADE IS A RULING'S STATEMENT: it stands on a docketing standing unruled, moves an
// axis, and the ruling after it applies it or is refused.
func TestTheBenchsRegradeIsARulingsStatement(t *testing.T) {
	t.Run("not before a docketing stands", func(t *testing.T) {
		s := newClosureStage(t)
		_, err := Append(sit(t, s.run, "judge"), regradeOf("G1", sev, high))
		mustRefuse(t, err, "no docketing on gap G1 stands unruled")
	})
	t.Run("not after the docketing is ruled", func(t *testing.T) {
		s := newClosureStage(t)
		s.epoch().ruled(recordpb.Disposition_DISPOSITION_REMANDED)
		_, err := Append(sit(t, s.run, "judge"), regradeOf("G1", sev, high))
		mustRefuse(t, err, "no docketing on gap G1 stands unruled")
	})
	t.Run("not to keep a grade", func(t *testing.T) {
		s := newClosureStage(t)
		s.docketed()
		_, err := Append(sit(t, s.run, "judge"), regradeOf("G1", sev, medium))
		mustRefuse(t, err, "this regrade moves no axis of gap G1")
	})
	t.Run("a ruling that applies none of it is refused", func(t *testing.T) {
		s := newClosureStage(t)
		s.docketed()
		s.bench = convene(t, s.run, "G1")
		mustAppend(t, s.bench, regradeOf("G1", sev, high))
		_, err := Append(s.bench, docketRuling(s.docketing, recordpb.Disposition_DISPOSITION_REMANDED))
		mustRefuse(t, err, "the bench writes no grade of its own")
	})
	t.Run("the applied grade is the grade every reader reads", func(t *testing.T) {
		s := benchRuledSeverity(t)
		w, err := WorkOfSeat(s.run, "lens", evLens)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range w.Open {
			if g.ID == "G1" && g.Severity != "low" {
				t.Errorf("G1's severity on the lens's list is %q, want the ruled low", g.Severity)
			}
		}
	})
}

// AN AXIS A RULING DECIDED IS NOT MOVED BACK: the lens may not move it, a successor carries it, and
// another axis stays the lens's.
func TestAnAxisARulingDecidedStaysDecided(t *testing.T) {
	t.Run("a lens regrade moving it is refused", func(t *testing.T) {
		s := benchRuledSeverity(t)
		_, err := Append(sit(t, s.run, evLens), regradeOf("G1", sev, high))
		mustRefuse(t, err, "gap G1's severity is low, which the bench's ruling at event", "a ruled grade is not moved back")
	})
	t.Run("a lens regrade keeping it, or moving another axis, is admitted", func(t *testing.T) {
		s := benchRuledSeverity(t)
		mustAppend(t, sit(t, s.run, evLens), regradeOf("G1", sev, low))
		mustAppend(t, sit(t, s.run, evLens), regradeOf("G1", impact, low))
	})
	t.Run("a successor carries the ruled grade", func(t *testing.T) {
		s := benchRuledSeverity(t)
		lens := sit(t, s.run, evLens)
		successor := func(g recordpb.Grade) *recordpb.Mint {
			m := corrMint("G2")
			m.Supersedes, m.Severity = proto.String("G1"), g.Enum()
			return m
		}
		_, err := Append(lens, successor(medium))
		mustRefuse(t, err, "--supersedes names G1", "mint it at --severity low")
		mustAppend(t, lens, successor(low))
		// THE LOCK FOLLOWS THE LINEAGE: the successor's severity is the ruling's too.
		_, err = Append(sit(t, s.run, evLens), regradeOf("G2", sev, high))
		mustRefuse(t, err, "gap G2's severity is low, which the bench's ruling at event")
	})
}

// A BENCH REGRADE IS NO PARTY'S MOVEMENT. Inside the remand's exchange it would read as the exchange
// moving the gap, and a second remand at impasse would be admitted past the limit.
func TestABenchRegradeInARemandExchangeDoesNotSpendTheLimit(t *testing.T) {
	st := remandedAtImpasse(t).lensActs("G1")
	agent := fmt.Sprintf("blue-computes-%d", st.n)
	st.registerAs("blue-respond", agent).blueComputes("G1")
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
