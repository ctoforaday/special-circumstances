package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// The bench's docket ruling, as the bench files it (#1201).
func docketRule(docketing, opinion string) *recordpb.MotionRule {
	return &recordpb.MotionRule{
		Opinion: proto.String(opinion),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{DocketingId: proto.String(docketing),
			Disposition: recordpb.Disposition_DISPOSITION_REMANDED.Enum(),
			Principle:   proto.String("pr"), Tension: proto.String("tn"), ReviewFlag: proto.String("rf"),
			Settled: proto.String("st"), Direction: proto.String("new evidence")}}}
}

const (
	benchReady   = "G1: docketed and unruled — the bench is ready"
	benchSatIdle = "G1: docketing %s stands unruled and the bench has sat since it was written — one bench sitting per docketing, so this gap is not re-readied"
)

// argued files G1's two closings: blue's in its sitting, and the evidence lens's in one of its own.
func argued(t *testing.T, run Run, blue Identity) {
	t.Helper()
	mustAppend(t, blue, &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("blue's case")})
	mustAppend(t, sit(t, run, evLens), &recordpb.Closing{GapId: proto.String("G1"), Text: proto.String("the lens's case")})
}

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

// A DOCKETING STANDS UNRULED UNTIL THE BENCH RULES IT, AND THE BENCH SITS ONCE PER DOCKETING. The
// dispatch dockets G1 as D1; the chair dispatches the bench for it; what the bench does in its sitting
// and what the dispatch writes next varies by arm. Every arm drives the real `dispatch judge [G1]` step
// before the bench sits — the step the first fixture skipped, which is why `last == 0` hid the
// bench-sat branch.
//
// Two readings of "a docketing stands unruled", and two keys for the bench's sitting, are told apart,
// so a regression in each fails its own arm:
//
//   - a ruling read as covering the gap rather than its docketing: the bench rules D1, and D2, a new
//     docketing of G1, stands unruled — "left as written"; and a second ruling of D1, which the write
//     refuses (#1205), leaves D2 standing too — "ruled twice";
//   - the bench's sitting keyed on the REGISTER's place against the docketing alone: D2 is written
//     after the chair's dispatch row that convened the bench's sitting, so that sitting is no sitting
//     for D2 and the plan readies the bench again — "written after the row that convened the sitting";
//     a judge register after a dispatch that named another seat reads as no sitting for D2 either
//     (#1201) — "another seat dispatched after the docketing".
//
// "sat for it" is the control for the key: the bench dispatched after D2 has sat for it, under any
// key. Blue and the lens are engaged in no arm: the gap is the bench's while a docketing stands.
func TestADocketStandsUnruledPerMotionAndReadiesTheBenchOncePerDocketing(t *testing.T) {
	type arm struct {
		name    string
		sitting func(t *testing.T, run Run) // the bench's sitting after `dispatch judge [G1]` with D1 standing, and what the dispatch writes
		judge   []string                    // the gaps the bench is engaged on
		reason  string                      // the plan's line for G1
	}
	ruleD1 := func(t *testing.T, judge Identity) {
		mustAppend(t, judge, docketRule("D1", "the ruling"))
	}
	// EACH DOCKETING IS ARGUED BY BOTH PARTIES before the bench is convened for it (closingowed.go),
	// so every docketing here is followed by the two closings: what is under test is the bench's
	// sitting per docketing, not the closings.
	docketD2 := func(t *testing.T, run Run) {
		if _, err := docketByDispatchAs(t, run, "D2", "G1", recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, ""); err != nil {
			t.Fatal(err)
		}
		argued(t, run, sit(t, run, "blue-respond"))
	}
	for _, tc := range []arm{
		{"left as written", func(t *testing.T, run Run) {
			ruleD1(t, sit(t, run, "judge"))
			docketD2(t, run)
		}, []string{"G1"}, benchReady},
		{"another seat dispatched after the docketing", func(t *testing.T, run Run) {
			ruleD1(t, sit(t, run, "judge"))
			docketD2(t, run)
			mustAppend(t, sit(t, run, "red-chair"), &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(evLens), GapIds: []string{"G1"}})
			sit(t, run, "judge")
		}, []string{"G1"}, benchReady},
		{"ruled twice", func(t *testing.T, run Run) {
			judge := sit(t, run, "judge")
			mustAppend(t, judge, docketRule("D1", "first"))
			if _, err := Append(judge, docketRule("D1", "again")); err == nil {
				t.Fatal("the write admitted a second ruling of D1")
			}
			docketD2(t, run)
		}, []string{"G1"}, benchReady},
		// THE ROW CONVENES THE SITTING FOR A DOCKETING IT FOLLOWS (benchSatFor, docketHeld): D2, written
		// after the dispatch that readied the bench, is not held by the sitting that row convened, which
		// neither rules it nor counts as its sitting — the bench is readied again for it.
		{"written after the row that convened the sitting", func(t *testing.T, run Run) {
			ruleD1(t, sit(t, run, "judge"))
			docketD2(t, run)
			sit(t, run, "judge")
		}, []string{"G1"}, benchReady},
		{"sat for it", func(t *testing.T, run Run) {
			ruleD1(t, sit(t, run, "judge"))
			docketD2(t, run)
			mustAppend(t, sit(t, run, "red-chair"), &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"}})
			sit(t, run, "judge")
		}, nil, fmt.Sprintf(benchSatIdle, "D2")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
				register("red-chair").sit(2, evLens).
				mint(evLens, "G1", "high").seed()
			exchangeOn(t, run, "G1")
			if _, err := docketByDispatchAs(t, run, "D1", "G1", recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, ""); err != nil {
				t.Fatal(err)
			}
			argued(t, run, sit(t, run, "blue-respond"))
			mustAppend(t, sit(t, run, "red-chair"), &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"}})
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
				t.Error("PASS is not permitted while a docketing stands unruled")
			}
			if plan.Ceiling {
				t.Error("a docketing standing unruled is not the run at its ceiling")
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
			o := docketRule("D1", "the ruling")
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
		o := docketRule("D1", "the ruling")
		o.GetDocket().Settled = proto.String("")
		if _, err := Append(sit(t, run, "judge"), o); err != nil {
			t.Errorf("a ruling that bars no proposition was refused: %v", err)
		}
	})
}

// EVERY REMAND STATES ITS DIRECTION (gblock, 2026-10-04), AND ONLY A REMAND DOES. A remand sends the
// gap back for one more exchange, and the dispatch hands blue and the minting lens the ruling's
// direction as what that exchange owes. --reopens-on and --final say what would reopen a closed gap,
// which a remand is not, so either on a remand is refused rather than given a meaning; a direction on
// a ruling that closes the gap is research no exchange is dispatched to do.
func TestARemandStatesItsDirection(t *testing.T) {
	for _, c := range []struct {
		name string
		set  func(*recordpb.DocketRuling)
	}{
		{"remanded --final", func(d *recordpb.DocketRuling) { d.Direction, d.Final = nil, proto.Bool(true) }},
		{"remanded with neither", func(d *recordpb.DocketRuling) { d.Direction = nil }},
		{"remanded with a direction and --final", func(d *recordpb.DocketRuling) { d.Final = proto.Bool(true) }},
		{"remanded with a direction and --reopens-on", func(d *recordpb.DocketRuling) { d.ReopensOn = proto.String("new evidence") }},
		{"remanded --reopens-on and no direction", func(d *recordpb.DocketRuling) { d.Direction, d.ReopensOn = nil, proto.String("new evidence") }},
		{"remanded with a blank direction", func(d *recordpb.DocketRuling) { d.Direction = proto.String("") }},
		{"remanded with a whitespace direction", func(d *recordpb.DocketRuling) { d.Direction = proto.String(" \t\n") }},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := mustRun(t, docketRunDir(t))
			o := docketRule("D1", "the ruling")
			c.set(o.GetDocket())
			_, err := Append(sit(t, run, "judge"), o)
			if err == nil {
				t.Fatal("a remand stating no direction reached the record — blue and the minting lens would be dispatched owing nothing")
			}
			if !strings.Contains(err.Error(), "a remand requires --direction") {
				t.Errorf("the refusal does not name --direction, so the bench cannot tell what the remand owes: %v", err)
			}
		})
	}
	t.Run("remanded --direction", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		if _, err := Append(sit(t, run, "judge"), docketRule("D1", "the ruling")); err != nil {
			t.Errorf("a remand stating its direction was refused: %v", err)
		}
	})
	t.Run("a disposition that ends the gap still takes --final", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		o := docketRule("D1", "the ruling")
		d := o.GetDocket()
		d.Disposition, d.Direction, d.Final = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(), nil, proto.Bool(true)
		d.SidedWith = recordpb.SidedWith_SIDED_WITH_BLUE.Enum()
		if _, err := Append(sit(t, run, "judge"), o); err != nil {
			t.Errorf("a final not_a_defect ruling was refused: %v", err)
		}
	})
	t.Run("a disposition that ends the gap states no direction", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		o := docketRule("D1", "the ruling")
		d := o.GetDocket()
		d.Disposition, d.Final = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(), proto.Bool(true)
		d.SidedWith = recordpb.SidedWith_SIDED_WITH_BLUE.Enum()
		_, err := Append(sit(t, run, "judge"), o)
		if err == nil || !strings.Contains(err.Error(), "--direction is the research a remand's one more exchange owes") {
			t.Errorf("a not_a_defect ruling stating a direction was not refused for it: %v", err)
		}
	})
}
