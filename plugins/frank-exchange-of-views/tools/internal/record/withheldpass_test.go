package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// permittingBoard is a record whose chair sitting, just opened, finds a PASS permitted: the report
// ingested at head 2, the one lens served it there twice, nothing open.
func permittingBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).sit(2, evLens).register("red-chair")
}

// permittedPlan is the plan the chair's `dispatch next` stores over a board that permits a PASS.
func (b *stage) permittedPlan() *stage {
	p := emptyPlan()
	p.PassPermitted, p.Head = true, b.head()
	return b.add(chairSeat, &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())})
}

// withhold is the chair holding back the PASS its plan permits: the plan stored, then the position.
func (b *stage) withhold(id string) *stage {
	return b.permittedPlan().add(chairSeat, &recordpb.Position{PositionId: proto.String(id), Text: proto.String("the PASS is withheld: the report's core claim rests on one source")})
}

// positionRuling is the bench's ruling on a position: the word, and on an upheld one the work.
func positionRuling(id string, word recordpb.PositionRuling, lens string, gaps []string, direction string) *recordpb.MotionRule {
	h := &recordpb.PositionHearing{Ruling: word.Enum(), GapIds: gaps}
	if lens != "" {
		h.Lens = proto.String(lens)
	}
	if direction != "" {
		h.Direction = proto.String(direction)
	}
	return &recordpb.MotionRule{MotionId: proto.String(id), Opinion: proto.String("the bench's reasoning on the withheld PASS"),
		Ruling: &recordpb.MotionRule_Position{Position: h}}
}

const (
	heardID   = "A-0000000a"
	direction = "find a second, independent source for the core claim, or state that none exists"
	upheld    = recordpb.PositionRuling_POSITION_RULING_UPHELD
	overruled = recordpb.PositionRuling_POSITION_RULING_OVERRULED
)

// recordReset points the stage at a fresh run directory, so the events it holds seed a record again.
func recordReset(t *testing.T, b *stage) { b.runDir = newRun(t) }

func mustFamily(t *testing.T, run Run) Family {
	t.Helper()
	f, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func runPlan(t *testing.T, run Run) Plan {
	t.Helper()
	p, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func partySeats(p Plan) []string {
	var out []string
	for _, x := range p.Parties {
		out = append(out, x.SeatID)
	}
	return out
}

// HOLDING BACK A PERMITTED PASS OWES THE POSITION A HEARING (gblock, 2026-10-30). The position the
// chair files where its plan permitted the PASS is a motion the bench rules: the record's read reports
// a position sitting owed, the PASS is refused while it stands, and the position is the sitting's duty
// discharged. A position at a sitting whose plan permitted none, or after the PASS, readies
// nobody.
func TestAWithheldPassReadiesTheBenchOnThePosition(t *testing.T) {
	b := permittingBoard(t)
	if p := runPlan(t, b.withhold(heardID).seed()); true {
		if p.PassPermitted || len(p.Parties) != 0 {
			t.Fatalf("plan after the withheld PASS = %+v, want nobody ready and no PASS: the hearing is on no plan", p)
		}
		if !slices.Contains(p.Blockers, PlanBlocker{Kind: BlockerUnruledMotion, Subject: heardID, Owner: benchSeat}) {
			t.Errorf("blockers = %+v, want the position unruled, the bench's to rule", p.Blockers)
		}
	}
	run := mustRun(t, b.runDir)
	if err := validateRun(run, chairSeat, recordpb.EventType_EVENT_TYPE_VERDICT, passGate); err == nil || !strings.Contains(err.Error(), heardID) {
		t.Errorf("a PASS over the unheard position = %v, want it refused naming %s", err, heardID)
	}
	fam := mustFamily(t, run)
	rows := PositionSittings(fam.Events, fam.At, AfterTheRun)
	if got := rows[len(rows)-1].State; got != PositionFiled {
		t.Errorf("the withholding sitting reads %s, want %s", got, PositionFiled)
	}

	// NOT HEARD: the plan permitted no PASS, or the sitting recorded one.
	notPermitted := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").mint(evLens, "G1", "high").
		add(chairSeat, &recordpb.Position{PositionId: proto.String(heardID), Text: proto.String("an argument about no one gap")}).seed()
	if p := runPlan(t, notPermitted); slices.Contains(partySeats(p), benchSeat) {
		t.Errorf("a position under a plan that permitted no PASS readied the bench: %+v", p)
	}
	passed := permittingBoard(t).permittedPlan().add(chairSeat, passGate).
		add(chairSeat, &recordpb.Position{PositionId: proto.String(heardID), Text: proto.String("after the PASS")}).seed()
	if p := runPlan(t, passed); slices.Contains(partySeats(p), benchSeat) {
		t.Errorf("a position after the PASS readied the bench: %+v", p)
	}
}

// UPHELD NAMES THE WORK, AND THE DEBATE CONTINUES ON IT (gblock, 2026-10-30): a lens, readied to
// audit its area until it has been served the report since the ruling; or open gaps, each sent back
// for one more exchange on the ruling's direction. Either way the PASS is held until the work is had.
func TestAnUpheldPositionReadiesTheWorkItNames(t *testing.T) {
	t.Run("a lens", func(t *testing.T) {
		b := permittingBoard(t).withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, upheld, evLens, nil, direction)).register("red-chair")
		p := runPlan(t, b.seed())
		if p.PassPermitted || !slices.Equal(partySeats(p), []string{evLens}) {
			t.Fatalf("plan after the upheld ruling = %+v, want the named lens alone", p)
		}
		if !slices.Contains(p.Blockers, PlanBlocker{Kind: BlockerUpheldPosition, Subject: heardID, Owner: evLens}) {
			t.Errorf("blockers = %+v, want the lens's audit", p.Blockers)
		}
		if why := strings.Join(p.Why, "\n"); !strings.Contains(why, direction) {
			t.Errorf("the plan's reasons do not carry the ruling's direction:\n%s", why)
		}
		// The lens sits after the ruling and is served the report: its audit is had.
		after := b.sit(b.head(), evLens).register("red-chair")
		recordReset(t, after)
		if p := runPlan(t, after.seed()); !p.PassPermitted || len(p.Parties) != 0 {
			t.Errorf("plan after the lens's audit = %+v, want PASS permitted again", p)
		}
	})
	t.Run("open gaps", func(t *testing.T) {
		// G1 is open and not material, so the PASS was permitted over it.
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").sit(2, evLens).sit(2, evLens).mint(evLens, "G1", "low").register("red-chair").
			withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, upheld, "", []string{"G1"}, direction)).register("red-chair")
		p := runPlan(t, b.seed())
		if p.PassPermitted || !slices.Equal(p.RemandOwed, []RemandOwed{{GapID: "G1", Direction: direction}}) {
			t.Fatalf("plan after the upheld ruling = %+v, want G1's one more exchange owed on the ruling's direction", p)
		}
		// BLUE ACTS FIRST, THEN THE LENS'S TURN, as on any remand: the ruling readies blue on G1, and
		// its minting lens only once blue has acted on it — on G1's news, blue's answer, since a gap
		// that is not material is routed to nobody once its remand's exchange has begun.
		if !slices.Contains(gapsOf(p, blueRespondSeat), "G1") || slices.Contains(gapsOf(p, evLens), "G1") {
			t.Errorf("parties = %+v, want blue on G1 and not its minting lens", p.Parties)
		}
		acted := b.blueSitsOn("G1", true).register("red-chair")
		recordReset(t, acted)
		if p := runPlan(t, acted.seed()); !slices.Contains(partySeats(p), evLens) || slices.Contains(p.RemandOwed, RemandOwed{GapID: "G1", Direction: direction}) {
			t.Errorf("plan after blue's act on G1 = %+v (remands owed %+v), want its minting lens's turn and the remand's exchange begun", p.Parties, p.RemandOwed)
		}
	})
}

// OVERRULED SENDS THE PASS BACK TO THE CHAIR'S NEXT SITTING (gblock, 2026-10-30). Its plan still
// permits the PASS; its work list holds the verdict as a blocking item; a position there is refused;
// and a sitting that closes with no verdict is short.
func TestAnOverruledPositionSendsThePassBackToTheChair(t *testing.T) {
	ruled := func(t *testing.T) *stage {
		return permittingBoard(t).withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
			add(benchSeat, positionRuling(heardID, overruled, "", nil, "")).register("red-chair").permittedPlan()
	}
	b := ruled(t)
	run := b.seed()
	if p := runPlan(t, run); !p.PassPermitted || len(p.Parties) != 0 {
		t.Fatalf("plan after the overruling = %+v, want PASS permitted and nobody ready", p)
	}
	fam := mustFamily(t, run)
	if id, owed := OverruledVerdictOwed(fam.Events, fam.At); !owed || id != heardID {
		t.Fatalf("OverruledVerdictOwed = (%q, %v), want (%q, true)", id, owed, heardID)
	}
	if items := blockingItems(sittingOfRunT(t, run, "chair", chairSeat)); !slices.Contains(items, overruledVerdictItem(heardID)) {
		t.Errorf("the chair's blocking items = %q, want the verdict the overruling left owed", items)
	}
	_, err := Append(Identity{Run: run, SeatID: chairSeat}, &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("the same objection")})
	mustRefuse(t, err, "overruled your position "+heardID, "Nothing was recorded")
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, passGate)
	fam = mustFamily(t, run)
	if _, owed := OverruledVerdictOwed(fam.Events, fam.At); owed {
		t.Error("the verdict is recorded and the overruling still reads as owing one")
	}
	// THE VERDICT DOES NOT REOPEN THE OBJECTION: a position after it, in the same sitting, would be
	// heard again on the same withholding, and is refused all the same.
	_, err = Append(Identity{Run: run, SeatID: chairSeat}, &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("the same objection, after the verdict")})
	mustRefuse(t, err, "overruled your position "+heardID, "Nothing was recorded")
	if got, _, ok, err := DeriveVerdict(run); err != nil || !ok || got != "VERIFIED" {
		t.Errorf("derived (%q, %v, %v), want VERIFIED", got, ok, err)
	}

	// A SITTING THAT CLOSES WITH NO VERDICT IS SHORT.
	short := ruled(t)
	sf := mustFamily(t, short.seed())
	rows := PositionSittings(sf.Events, sf.At, AfterTheRun)
	if got := rows[len(rows)-1]; got.State != PositionVerdictMissing || got.Overruled != heardID {
		t.Errorf("the sitting after the overruling, closed with no verdict, reads %+v; want %s naming %s", got, PositionVerdictMissing, heardID)
	}
}

// AT THE EPOCH LIMIT THE TERMINAL BENCH HEARS THE POSITION AT THE EXIT, AND THE RUN ENDS CEILING
// however it rules: the chair's final sitting dispatches nobody, so the bench it readies is the
// limit's, as a petition's is (epochLimitHeldTheBench).
func TestAPositionAtTheFinalSittingEndsTheRunAtItsCeiling(t *testing.T) {
	for _, word := range []recordpb.PositionRuling{upheld, overruled} {
		t.Run(recordpb.Word(word), func(t *testing.T) {
			b := permittingBoard(t).withhold(heardID)
			run := b.seed()
			writeRunConfig(t, b.runDir, `{"maxEpochs":2}`)
			p := runPlan(t, run)
			if !p.EpochLimitReached || !p.Ceiling || len(p.Parties) != 0 {
				t.Fatalf("plan at the final sitting after the withheld PASS = %+v, want the limit reached over the bench", p)
			}
			lens := ""
			if word == upheld {
				lens = evLens
			}
			d := ""
			if word == upheld {
				d = direction
			}
			b.benchRegister(recordpb.Occasion_OCCASION_TERMINAL).add(benchSeat, positionRuling(heardID, word, lens, nil, d))
			recordReset(t, b)
			run = b.seed()
			writeRunConfig(t, b.runDir, `{"maxEpochs":2}`)
			if got, why, ok, err := DeriveVerdict(run); err != nil || !ok || got != "CEILING" {
				t.Errorf("derived (%q, %q, %v, %v), want CEILING: the limit held the bench", got, why, ok, err)
			}
		})
	}
}

// THE RULING NAMES THE WORK ONE WAY, OR NONE (gblock, 2026-10-30): the write refuses an upheld ruling
// naming neither a lens nor gaps, or both, or no direction; an overruled one naming any; a lens not in
// the cast; a closed gap; and a position the bench does not hear.
func TestThePositionRulingIsRefusedWrongWork(t *testing.T) {
	board := func(t *testing.T) Run {
		return permittingBoard(t).mint(evLens, "G1", "low").closeGap(evLens, "G1").mint(evLens, "G2", "low").
			register("red-chair").withhold(heardID).
			register("red-chair").add(chairSeat, &recordpb.Position{PositionId: proto.String("A-0000000b"), Text: proto.String("x")}). // stored no plan: not heard
			seed()
	}
	for _, c := range []struct {
		name string
		rule *recordpb.MotionRule
		want string
	}{
		{"upheld, no work", positionRuling(heardID, upheld, "", nil, direction), "not both, and not neither"},
		{"upheld, both", positionRuling(heardID, upheld, evLens, []string{"G2"}, direction), "not both, and not neither"},
		{"upheld, no direction", positionRuling(heardID, upheld, evLens, nil, ""), "requires --direction"},
		{"overruled, a lens", positionRuling(heardID, overruled, evLens, nil, ""), "names no work"},
		{"upheld, no lens of the cast", positionRuling(heardID, upheld, "red-lens-voice", nil, direction), "no lens of this run's cast"},
		{"upheld, a closed gap", positionRuling(heardID, upheld, "", []string{"G1"}, direction), "G1"},
		{"a position the bench does not hear", positionRuling("A-0000000b", overruled, "", nil, ""), "withheld no PASS"},
		{"no position at all", positionRuling("A-0000000c", overruled, "", nil, ""), "no position on the record"},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := board(t)
			_, err := Append(sit(t, run, benchSeat), c.rule)
			mustRefuse(t, err, c.want)
		})
	}
	run := board(t)
	mustAppend(t, sit(t, run, benchSeat), positionRuling(heardID, upheld, "", []string{"G2"}, direction))
}

// AN UPHELD POSITION NAMING A GAP IS THAT GAP'S REMAND, to every reader of a remand: blue's answer to
// it is read after the ruling (remandFloorsOf), and the ruling is a change to the gap a seat that sat
// before it has not read (staleOf).
func TestAnUpheldPositionNamingAGapIsItsRemand(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).sit(2, evLens).mint(evLens, "G1", "low").register("red-chair").
		blueSitsOn("G1", true).register("red-chair").
		withhold(heardID).benchRegister(recordpb.Occasion_OCCASION_POSITION).
		add(benchSeat, positionRuling(heardID, upheld, "", []string{"G1"}, direction))
	evs, win := mustEvents(t, b.seed())
	ids := win.IDs(evs)
	if a, ok := blueAnswersOf(evs, win, map[string]bool{"G1": true}, motionsAt(evs, ids, win))["G1"]; ok {
		t.Errorf("blue's answer to G1 from before the bench remanded it is read as an answer to the remand: %+v", a)
	}
	ch := staleOf(evs, win, evLens, "G1")
	if !slices.ContainsFunc(ch, func(c StaleChange) bool { return strings.Contains(c.What, "upholding position "+heardID) }) {
		t.Errorf("changes to G1 since %s sat = %+v, want the bench's remand of it upholding %s", evLens, ch, heardID)
	}
}
