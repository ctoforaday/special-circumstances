package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The bench's docket ruling, as the bench files it (#1201).
func docketRule(id, opinion string) *recordpb.MotionRule {
	return &recordpb.MotionRule{MotionId: proto.String(id),
		Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Opinion: proto.String(opinion),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: recordpb.Disposition_DISPOSITION_REMANDED.Enum(),
			Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
			Settled: proto.String("st"), ReopensOn: proto.String("new evidence")}}}
}

const (
	benchReady   = "G1: docketed and unruled — the bench is ready"
	benchSatIdle = "G1: docket motion %s stands unruled and the bench has sat since it was filed — one bench sitting per docketing, so this gap is not re-readied"
)

// docketBoard is the plan over one open gap and the parties it engages, by seat.
func docketBoard(t *testing.T, run Run) (Plan, map[string][]string) {
	t.Helper()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	engaged := map[string][]string{}
	for _, p := range plan.Parties {
		engaged[p.SeatID] = p.GapIDs
	}
	return plan, engaged
}

// A DOCKET STANDS UNRULED PER MOTION, AND A NEW FILING IS A NEW DOCKETING. Blue dockets G1; the
// chair dispatches the bench for it; what the bench does in its sitting and what blue files next
// varies by arm. Every arm drives the real `dispatch judge [G1]` step before the bench sits — the
// step the first fixture skipped, which is why `last == 0` hid the bench-sat branch.
//
// Three readings of "a docket stands unruled", and three keys for the bench's sitting, are told
// apart, so a regression in each fails its own arm:
//
//   - a count of ruling ROWS against a count of dockets: a ruling corrected in its sitting is a
//     second `motion_rule` row for one live act, so dockets=2 rulings=2 took M2 as ruled and readied
//     the minting lens and blue while the bench was never dispatched (#1201) — "corrected";
//   - a count of LIVE rulings against a count of dockets: two live rulings on one motion cover an
//     unruled sibling — "ruled twice" (the record write refuses the second ruling, #1205, so the arm
//     holds that refusal and that the refused attempt leaves the sibling standing unruled);
//   - the bench's sitting keyed on the newest FILING rather than the newest UNRULED filing: a
//     motion blue files into the bench's open sitting and the bench rules there is newer than the
//     dispatch, so the bench reads as not having sat for the older motion it left alone and is
//     readied again for it — "ruled the newer one in the sitting";
//   - the bench's sitting keyed on the DISPATCH's place against the filing rather than the
//     REGISTER's: blue files M2 between the chair's dispatch and the bench's register, the bench
//     sits with M1 and M2 both on the record and rules nothing, and the plan readies it again for
//     the docketing it just sat for — "filed before the bench registered";
//   - the bench's sitting keyed on the last dispatch alone: a judge register after the dispatch
//     for M1 read as a sitting for M2 too, and nobody was engaged (#1201) — every arm that readies
//     the bench after it sat.
//
// "left as written" is the control for the first, and "sat for both" the control for the key: the
// bench dispatched twice with M1 standing has sat for it, under any key. Blue and the lens are
// engaged in no arm: the gap is the bench's while a docket stands.
func TestADocketStandsUnruledPerMotionAndReadiesTheBenchOncePerDocketing(t *testing.T) {
	type arm struct {
		name    string
		sitting func(t *testing.T, run Run) // the bench's sitting after `dispatch judge [G1]` with M1 standing, and what blue files
		judge   []string                    // the gaps the bench is engaged on
		reason  string                      // the plan's line for G1
	}
	ruleM1 := func(t *testing.T, judge Identity, correct bool) {
		k := mustAppend(t, judge, docketRule("M1", "typo")).GetKey()
		if correct {
			if _, err := Append(correcting(judge, recordpb.EventType_EVENT_TYPE_MOTION_RULE, k, "typo"), docketRule("M1", "fixed")); err != nil {
				t.Fatalf("correcting the docket ruling refused: %v", err)
			}
		}
	}
	fileM2 := func(t *testing.T, run Run) { mustAppend(t, sit(t, run, "blue-respond"), docketMotion("M2", "G1")) }
	for _, tc := range []arm{
		{"corrected", func(t *testing.T, run Run) {
			ruleM1(t, sit(t, run, "judge"), true)
			fileM2(t, run)
		}, []string{"G1"}, benchReady},
		{"left as written", func(t *testing.T, run Run) {
			ruleM1(t, sit(t, run, "judge"), false)
			fileM2(t, run)
		}, []string{"G1"}, benchReady},
		{"another seat dispatched after the filing", func(t *testing.T, run Run) {
			ruleM1(t, sit(t, run, "judge"), false)
			fileM2(t, run)
			mustAppend(t, sit(t, run, "red-chair"), &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(evLens), GapIds: []string{"G1"}})
			sit(t, run, "judge")
		}, []string{"G1"}, benchReady},
		{"ruled twice", func(t *testing.T, run Run) {
			judge := sit(t, run, "judge")
			mustAppend(t, judge, docketRule("M1", "first"))
			if _, err := Append(judge, docketRule("M1", "again")); err == nil {
				t.Fatal("the write admitted a second ruling on M1")
			}
			fileM2(t, run)
		}, []string{"G1"}, benchReady},
		{"ruled the newer one in the sitting", func(t *testing.T, run Run) {
			judge := sit(t, run, "judge")
			fileM2(t, run)
			mustAppend(t, judge, docketRule("M2", "ruled the newer"))
		}, nil, fmt.Sprintf(benchSatIdle, "M1")},
		{"filed before the bench registered", func(t *testing.T, run Run) {
			fileM2(t, run)
			sit(t, run, "judge")
		}, nil, fmt.Sprintf(benchSatIdle, "M2")},
		{"sat for both", func(t *testing.T, run Run) {
			sit(t, run, "judge")
			fileM2(t, run)
			mustAppend(t, sit(t, run, "red-chair"), &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"}, Occasions: []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}})
			mustAppend(t, sit(t, run, "judge"), docketRule("M2", "ruled the newer"))
		}, nil, fmt.Sprintf(benchSatIdle, "M1")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
				register("red-chair").dispatch(2, evLens).register(evLens).
				mint(evLens, "G1", "high").seed()
			mustAppend(t, sit(t, run, "blue-respond"), docketMotion("M1", "G1"))
			mustAppend(t, sit(t, run, "red-chair"), &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"}, Occasions: []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}})
			tc.sitting(t, run)
			sit(t, run, "red-chair")

			plan, engaged := docketBoard(t, run)
			if got := engaged["judge"]; strings.Join(got, ",") != strings.Join(tc.judge, ",") {
				t.Errorf("the bench is engaged on %v, want %v; parties %v", got, tc.judge, engaged)
			}
			if len(engaged["blue-respond"]) != 0 || len(engaged[evLens]) != 0 {
				t.Errorf("blue or the lens is engaged on a gap that is the bench's: %v", engaged)
			}
			if plan.PassPermitted {
				t.Error("PASS is not permitted while a docket motion stands unruled")
			}
			if plan.Ceiling {
				t.Error("a docket standing unruled is not the run at its ceiling")
			}
			why := strings.Join(plan.Why, "\n")
			if !strings.Contains(why, tc.reason) {
				t.Errorf("the plan's line for G1 must read %q:\n%s", tc.reason, why)
			}
		})
	}
}

// THE WRITE REFUSES A DOCKET RULING THAT OMITS --tension, --review-flag OR --settled (#1234), and
// takes an empty --settled: a ruling may bar no proposition, so for that field alone the empty
// answer is an answer and what the write can refuse is only the field never said — which is why
// the verb must leave an omitted flag ABSENT rather than write it as "". An empty --tension or
// --review-flag is refused (TestTheBenchStatesItsRuleItsTensionAndItsReviewFlag).
func TestTheDocketRulingWriteRefusesAnOmittedPresenceField(t *testing.T) {
	for _, c := range []struct {
		flag  string
		clear func(*recordpb.DocketRuling)
	}{
		{"--tension", func(d *recordpb.DocketRuling) { d.Tension = nil }},
		{"--review-flag", func(d *recordpb.DocketRuling) { d.ReviewFlag = nil }},
		{"--settled", func(d *recordpb.DocketRuling) { d.Settled = nil }},
	} {
		t.Run("omitted "+c.flag, func(t *testing.T) {
			run := mustRun(t, docketRunDir(t))
			o := docketRule("M1", "the ruling")
			c.clear(o.GetDocket())
			_, err := Append(sit(t, run, "judge"), o)
			if err == nil {
				t.Fatalf("a docket ruling with no %s reached the record", c.flag)
			}
			if !strings.Contains(err.Error(), c.flag) {
				t.Errorf("the refusal does not name %s: %v", c.flag, err)
			}
		})
	}
	t.Run("settled present and empty", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		o := docketRule("M1", "the ruling")
		o.GetDocket().Settled = proto.String("")
		if _, err := Append(sit(t, run, "judge"), o); err != nil {
			t.Errorf("a ruling that bars no proposition was refused: %v", err)
		}
	})
}

// EVERY REMAND STATES ITS DIRECTION (gblock, 2026-10-04). A remand sends the gap back for one more
// exchange, and the dispatch hands blue and the minting lens the ruling's reopens_on as what that
// exchange owes; a remand that said --final would hand them nothing. --final says nothing would
// reopen the gap, which a remand contradicts, so the pair is refused rather than given a meaning.
func TestARemandStatesItsDirection(t *testing.T) {
	for _, c := range []struct {
		name string
		set  func(*recordpb.DocketRuling)
	}{
		{"remanded --final", func(d *recordpb.DocketRuling) { d.ReopensOn, d.Final = nil, proto.Bool(true) }},
		{"remanded with neither", func(d *recordpb.DocketRuling) { d.ReopensOn = nil }},
		{"remanded with a direction and --final", func(d *recordpb.DocketRuling) { d.Final = proto.Bool(true) }},
		{"remanded with a blank direction", func(d *recordpb.DocketRuling) { d.ReopensOn = proto.String("") }},
		{"remanded with a whitespace direction", func(d *recordpb.DocketRuling) { d.ReopensOn = proto.String(" \t\n") }},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := mustRun(t, docketRunDir(t))
			o := docketRule("M1", "the ruling")
			c.set(o.GetDocket())
			_, err := Append(sit(t, run, "judge"), o)
			if err == nil {
				t.Fatal("a remand stating no direction reached the record — blue and the minting lens would be dispatched owing nothing")
			}
			if !strings.Contains(err.Error(), "--as remanded requires --reopens-on") {
				t.Errorf("the refusal does not name --reopens-on, so the bench cannot tell what the remand owes: %v", err)
			}
		})
	}
	t.Run("remanded --reopens-on", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		if _, err := Append(sit(t, run, "judge"), docketRule("M1", "the ruling")); err != nil {
			t.Errorf("a remand stating its direction was refused: %v", err)
		}
	})
	t.Run("a disposition that ends the gap still takes --final", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		o := docketRule("M1", "the ruling")
		d := o.GetDocket()
		d.Disposition, d.ReopensOn, d.Final = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(), nil, proto.Bool(true)
		if _, err := Append(sit(t, run, "judge"), o); err != nil {
			t.Errorf("a final not_a_defect ruling was refused: %v", err)
		}
	})
}
